// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package state

import (
	"bytes"
	"fmt"
	"io"
	"math/big"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/ethereumproject/go-ethereum/crypto"
	"github.com/ethereumproject/go-ethereum/logger"
	"github.com/ethereumproject/go-ethereum/logger/glog"
	"github.com/ethereumproject/go-ethereum/rlp"
	"github.com/ethereumproject/go-ethereum/trie"
)

var emptyCodeHash = crypto.Keccak256(nil)

type Code []byte

func (self Code) String() string {
	return string(self) //strings.Join(Disassemble(self), " ")
}

type Storage map[common.Hash]common.Hash

func (self Storage) String() (str string) {
	for key, value := range self {
		str += fmt.Sprintf("%X : %X\n", key, value)
	}

	return
}

func (self Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range self {
		cpy[key] = value
	}

	return cpy
}

// StateObject represents an Ethereum account which is being modified.
//
// The usage pattern is as follows:
// First you need to obtain a state object.
// Account values can be accessed and modified through the object.
// Finally, call CommitTrie to write the modified storage trie into a database.
type StateObject struct {
	address common.Address // Ethereum address of this account
	data    Account
	db      *StateDB

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// Write caches.
	trie *trie.SecureTrie // storage trie, which becomes non-nil on first access
	code Code             // contract bytecode, which gets set when code is loaded

	cachedStorage Storage // Storage entry cache to avoid duplicate reads
	dirtyStorage  Storage // Storage entries that need to be flushed to disk

	// Cache flags.
	// When an object is marked suicided it will be delete from the trie
	// during the "update" phase of the state transition.
	dirtyCode bool // true if the code was updated
	suicided  bool
	deleted   bool
	onDirty   func(addr common.Address) // Callback method to mark a state object newly dirty
}

// AccountObject is a reduced StateObject interface
type AccountObject interface {
	SubBalance(amount *big.Int)
	AddBalance(amount *big.Int)
	SetBalance(*big.Int)
	SetNonce(uint64)
	SetCode(common.Hash, []byte)
	Nonce() uint64
	Balance() *big.Int
	Address() common.Address
	//	Value() *big.Int
	ReturnGas(*big.Int, *big.Int)
	ForEachStorage(cb func(key, value common.Hash) bool)
}

// Account is the Ethereum consensus representation of accounts.
// These objects are stored in the main account trie.
type Account struct {
	Nonce    uint64
	Balance  *big.Int
	Root     common.Hash // merkle root of the storage trie
	CodeHash []byte
}

// newObject creates a state object.
func newObject(db *StateDB, address common.Address, data Account, onDirty func(addr common.Address)) *StateObject {
	if data.Balance == nil {
		data.Balance = new(big.Int)
	}
	if data.CodeHash == nil {
		data.CodeHash = emptyCodeHash
	}
	return &StateObject{db: db, address: address, data: data, cachedStorage: make(Storage), dirtyStorage: make(Storage), onDirty: onDirty}
}

// EncodeRLP implements rlp.Encoder.
func (sobj *StateObject) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, sobj.data)
}

// setError remembers the first non-nil error it is called with.
func (sobj *StateObject) setError(err error) {
	if sobj.dbErr == nil {
		sobj.dbErr = err
	}
}

func (sobj *StateObject) markSuicided() {
	sobj.suicided = true
	if sobj.onDirty != nil {
		sobj.onDirty(sobj.Address())
		sobj.onDirty = nil
	}
	if glog.V(logger.Core) {
		glog.Infof("%x: #%d %v X\n", sobj.Address(), sobj.Nonce(), sobj.Balance())
	}
}

func (sobj *StateObject) getTrie(db trie.Database) *trie.SecureTrie {
	if sobj.trie == nil {
		var err error
		sobj.trie, err = trie.NewSecure(sobj.data.Root, db, 0)
		if err != nil {
			sobj.trie, _ = trie.NewSecure(common.Hash{}, db, 0)
			sobj.setError(fmt.Errorf("can't create storage trie: %v", err))
		}
	}
	return sobj.trie
}

// GetState returns a value in account storage.
func (sobj *StateObject) GetState(db trie.Database, key common.Hash) common.Hash {
	value, exists := sobj.cachedStorage[key]
	if exists {
		return value
	}
	// Load from DB in case it is missing.
	if enc := sobj.getTrie(db).Get(key[:]); len(enc) > 0 {
		_, content, _, err := rlp.Split(enc)
		if err != nil {
			sobj.setError(err)
		}
		value.SetBytes(content)
	}
	if (value != common.Hash{}) {
		sobj.cachedStorage[key] = value
	}
	return value
}

// SetState updates a value in account storage.
func (sobj *StateObject) SetState(db trie.Database, key, value common.Hash) {
	sobj.db.journal = append(sobj.db.journal, storageChange{
		account:  &sobj.address,
		key:      key,
		prevalue: sobj.GetState(db, key),
	})
	sobj.setState(key, value)
}

func (sobj *StateObject) setState(key, value common.Hash) {
	sobj.cachedStorage[key] = value
	sobj.dirtyStorage[key] = value

	if sobj.onDirty != nil {
		sobj.onDirty(sobj.Address())
		sobj.onDirty = nil
	}
}

// updateTrie writes cached storage modifications into the object's storage trie.
func (sobj *StateObject) updateTrie(db trie.Database) {
	tr := sobj.getTrie(db)
	for key, value := range sobj.dirtyStorage {
		delete(sobj.dirtyStorage, key)
		if (value == common.Hash{}) {
			tr.Delete(key[:])
			continue
		}
		// Encoding []byte cannot fail, ok to ignore the error.
		v, _ := rlp.EncodeToBytes(bytes.TrimLeft(value[:], "\x00"))
		tr.Update(key[:], v)
	}
}

// UpdateRoot sets the trie root to the current root hash of
func (sobj *StateObject) updateRoot(db trie.Database) {
	sobj.updateTrie(db)
	sobj.data.Root = sobj.trie.Hash()
}

// CommitTrie the storage trie of the object to dwb.
// This updates the trie root.
func (sobj *StateObject) CommitTrie(db trie.Database, dbw trie.DatabaseWriter) error {
	sobj.updateTrie(db)
	if sobj.dbErr != nil {
		return sobj.dbErr
	}
	root, err := sobj.trie.CommitTo(dbw)
	if err == nil {
		sobj.data.Root = root
	}
	return err
}

func (sobj *StateObject) AddBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	sobj.SetBalance(new(big.Int).Add(sobj.Balance(), amount))

	if glog.V(logger.Debug) {
		glog.Infof("%x: #%d %v (+ %v)\n", sobj.Address(), sobj.Nonce(), sobj.Balance(), amount)
	}
}

func (sobj *StateObject) SubBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	sobj.SetBalance(new(big.Int).Sub(sobj.Balance(), amount))

	if glog.V(logger.Core) {
		glog.Infof("%x: #%d %v (- %v)\n", sobj.Address(), sobj.Nonce(), sobj.Balance(), amount)
	}
}

func (sobj *StateObject) SetBalance(amount *big.Int) {
	sobj.db.journal = append(sobj.db.journal, balanceChange{
		account: &sobj.address,
		prev:    new(big.Int).Set(sobj.data.Balance),
	})
	sobj.setBalance(amount)
}

func (sobj *StateObject) setBalance(amount *big.Int) {
	sobj.data.Balance = amount
	if sobj.onDirty != nil {
		sobj.onDirty(sobj.Address())
		sobj.onDirty = nil
	}
}

// Return the gas back to the origin. Used by the Virtual machine or Closures
func (sobj *StateObject) ReturnGas(gas, price *big.Int) {}

func (sobj *StateObject) deepCopy(db *StateDB, onDirty func(addr common.Address)) *StateObject {
	so := newObject(db, sobj.address, sobj.data, onDirty)
	so.trie = sobj.trie
	// Modified to use bytecode instead of a copy of the bytecode
	so.code = sobj.code
	so.dirtyStorage = sobj.dirtyStorage.Copy()
	so.cachedStorage = sobj.dirtyStorage.Copy()
	so.suicided = sobj.suicided
	so.dirtyCode = sobj.dirtyCode
	so.deleted = sobj.deleted
	return so
}

//
// Attribute accessors
//

// Returns the address of the contract/account
func (sobj *StateObject) Address() common.Address {
	return sobj.address
}

// Code returns the contract code associated with this object, if any.
func (sobj *StateObject) Code(db trie.Database) []byte {
	if sobj.code != nil {
		return sobj.code
	}
	if bytes.Equal(sobj.CodeHash(), emptyCodeHash) {
		return nil
	}
	code, err := db.Get(sobj.CodeHash())
	if err != nil {
		sobj.setError(fmt.Errorf("can't load code hash %x: %v", sobj.CodeHash(), err))
	}
	sobj.code = code
	return code
}

func (sobj *StateObject) SetCode(codeHash common.Hash, code []byte) {
	prevcode := sobj.Code(sobj.db.db)
	sobj.db.journal = append(sobj.db.journal, codeChange{
		account:  &sobj.address,
		prevhash: sobj.CodeHash(),
		prevcode: prevcode,
	})
	sobj.setCode(codeHash, code)
}

func (sobj *StateObject) setCode(codeHash common.Hash, code []byte) {
	sobj.code = code
	sobj.data.CodeHash = codeHash[:]
	sobj.dirtyCode = true
	if sobj.onDirty != nil {
		sobj.onDirty(sobj.Address())
		sobj.onDirty = nil
	}
}

func (sobj *StateObject) SetNonce(nonce uint64) {
	sobj.db.journal = append(sobj.db.journal, nonceChange{
		account: &sobj.address,
		prev:    sobj.data.Nonce,
	})
	sobj.setNonce(nonce)
}

func (sobj *StateObject) setNonce(nonce uint64) {
	sobj.data.Nonce = nonce
	if sobj.onDirty != nil {
		sobj.onDirty(sobj.Address())
		sobj.onDirty = nil
	}
}

func (sobj *StateObject) CodeHash() []byte {
	return sobj.data.CodeHash
}

func (sobj *StateObject) Balance() *big.Int {
	return sobj.data.Balance
}

func (sobj *StateObject) Nonce() uint64 {
	return sobj.data.Nonce
}

// Never called, but must be present to allow StateObject to be used
// as a vm.Account interface that also satisfies the vm.ContractRef
// interface. Interfaces are awesome.
//func (self *StateObject) Value() *big.Int {
//	panic("Value on StateObject should never be called")
//}

func (sobj *StateObject) ForEachStorage(cb func(key, value common.Hash) bool) {
	// When iterating over the storage check the cache first
	for h, value := range sobj.cachedStorage {
		cb(h, value)
	}

	it := sobj.getTrie(sobj.db.db).Iterator()
	for it.Next() {
		// ignore cached values
		key := common.BytesToHash(sobj.trie.GetKey(it.Key))
		if _, ok := sobj.cachedStorage[key]; !ok {
			cb(key, common.BytesToHash(it.Value))
		}
	}
}
