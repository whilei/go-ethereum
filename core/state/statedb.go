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

// Package state provides a caching layer atop the Ethereum state trie.
package state

import (
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/ethereumproject/go-ethereum/crypto"
	"github.com/ethereumproject/go-ethereum/ethdb"
	"github.com/ethereumproject/go-ethereum/logger"
	"github.com/ethereumproject/go-ethereum/logger/glog"
	"github.com/ethereumproject/go-ethereum/rlp"
	"github.com/ethereumproject/go-ethereum/trie"
	lru "github.com/hashicorp/golang-lru"
)

// The starting nonce determines the default nonce when new accounts are being
// created.
var StartingNonce uint64

const (
	// Number of past tries to keep. The arbitrarily chosen value here
	// is max uncle depth + 1.
	maxPastTries = 8

	// Trie cache generation limit.
	maxTrieCacheGen = 100

	// Number of codehash->size associations to keep.
	codeSizeCacheSize = 100000

	// Default StartingNonce for Morden Testnet
	DefaultTestnetStartingNonce = uint64(1048576)
)

type revision struct {
	id           int
	journalIndex int
}

// StateDBs within the ethereum protocol are used to store anything
// within the merkle trie. StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * Contracts
// * Accounts
type StateDB struct {
	db            ethdb.Database
	trie          *trie.SecureTrie
	pastTries     []*trie.SecureTrie
	codeSizeCache *lru.Cache

	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects      map[common.Address]*StateObject
	stateObjectsDirty map[common.Address]struct{}

	// The refund counter, also used by state transitioning.
	refund *big.Int

	txHash, blockHash common.Hash
	txIndex           int
	logs              map[common.Hash]Logs
	logSize           uint

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        journal
	validRevisions []revision
	nextRevisionId int

	lock sync.Mutex
}

// Create a new state from a given trie
func New(root common.Hash, db ethdb.Database) (*StateDB, error) {
	tr, err := trie.NewSecure(root, db, maxTrieCacheGen)
	if err != nil {
		return nil, err
	}
	csc, _ := lru.New(codeSizeCacheSize)
	return &StateDB{
		db:                db,
		trie:              tr,
		codeSizeCache:     csc,
		stateObjects:      make(map[common.Address]*StateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		refund:            new(big.Int),
		logs:              make(map[common.Hash]Logs),
	}, nil
}

// New creates a new statedb by reusing any journalled tries to avoid costly
// disk io.
func (sdb *StateDB) New(root common.Hash) (*StateDB, error) {
	sdb.lock.Lock()
	defer sdb.lock.Unlock()

	tr, err := sdb.openTrie(root)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		db:                sdb.db,
		trie:              tr,
		codeSizeCache:     sdb.codeSizeCache,
		stateObjects:      make(map[common.Address]*StateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		refund:            new(big.Int),
		logs:              make(map[common.Hash]Logs),
	}, nil
}

// Reset clears out all emphemeral state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
func (sdb *StateDB) Reset(root common.Hash) error {
	sdb.lock.Lock()
	defer sdb.lock.Unlock()

	tr, err := sdb.openTrie(root)
	if err != nil {
		return err
	}
	sdb.trie = tr
	sdb.stateObjects = make(map[common.Address]*StateObject)
	sdb.stateObjectsDirty = make(map[common.Address]struct{})
	sdb.txHash = common.Hash{}
	sdb.blockHash = common.Hash{}
	sdb.txIndex = 0
	sdb.logs = make(map[common.Hash]Logs)
	sdb.logSize = 0
	sdb.clearJournalAndRefund()

	return nil
}

// openTrie creates a trie. It uses an existing trie if one is available
// from the journal if available.
func (sdb *StateDB) openTrie(root common.Hash) (*trie.SecureTrie, error) {
	for i := len(sdb.pastTries) - 1; i >= 0; i-- {
		if sdb.pastTries[i].Hash() == root {
			tr := *sdb.pastTries[i]
			return &tr, nil
		}
	}
	return trie.NewSecure(root, sdb.db, maxTrieCacheGen)
}

func (sdb *StateDB) pushTrie(t *trie.SecureTrie) {
	sdb.lock.Lock()
	defer sdb.lock.Unlock()

	if len(sdb.pastTries) >= maxPastTries {
		copy(sdb.pastTries, sdb.pastTries[1:])
		sdb.pastTries[len(sdb.pastTries)-1] = t
	} else {
		sdb.pastTries = append(sdb.pastTries, t)
	}
}

func (sdb *StateDB) StartRecord(thash, bhash common.Hash, ti int) {
	sdb.txHash = thash
	sdb.blockHash = bhash
	sdb.txIndex = ti
}

func (sdb *StateDB) AddLog(log *Log) {
	sdb.journal = append(sdb.journal, addLogChange{txhash: sdb.txHash})

	log.TxHash = sdb.txHash
	log.BlockHash = sdb.blockHash
	log.TxIndex = uint(sdb.txIndex)
	log.Index = sdb.logSize
	sdb.logs[sdb.txHash] = append(sdb.logs[sdb.txHash], log)
	sdb.logSize++
}

func (sdb *StateDB) GetLogs(hash common.Hash) Logs {
	return sdb.logs[hash]
}

func (sdb *StateDB) Logs() Logs {
	var logs Logs
	for _, lgs := range sdb.logs {
		logs = append(logs, lgs...)
	}
	return logs
}

func (sdb *StateDB) AddRefund(gas *big.Int) {
	sdb.journal = append(sdb.journal, refundChange{prev: new(big.Int).Set(sdb.refund)})
	sdb.refund.Add(sdb.refund, gas)
}

// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (sdb *StateDB) Exist(addr common.Address) bool {
	return sdb.GetStateObject(addr) != nil
}

func (sdb *StateDB) GetAccount(addr common.Address) AccountObject {
	return sdb.GetStateObject(addr)
}

// Retrieve the balance from the given address or 0 if object not found
func (sdb *StateDB) GetBalance(addr common.Address) *big.Int {
	stateObject := sdb.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return new(big.Int)
}

func (sdb *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := sdb.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return StartingNonce
}

func (sdb *StateDB) GetCode(addr common.Address) []byte {
	stateObject := sdb.GetStateObject(addr)
	if stateObject != nil {
		code := stateObject.Code(sdb.db)
		key := common.BytesToHash(stateObject.CodeHash())
		sdb.codeSizeCache.Add(key, len(code))
		return code
	}
	return nil
}

func (sdb *StateDB) GetCodeSize(addr common.Address) int {
	stateObject := sdb.GetStateObject(addr)
	if stateObject == nil {
		return 0
	}
	key := common.BytesToHash(stateObject.CodeHash())
	if cached, ok := sdb.codeSizeCache.Get(key); ok {
		return cached.(int)
	}
	size := len(stateObject.Code(sdb.db))
	if stateObject.dbErr == nil {
		sdb.codeSizeCache.Add(key, size)
	}
	return size
}

func (sdb *StateDB) GetCodeHash(addr common.Address) common.Hash {
	stateObject := sdb.GetStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}

func (sdb *StateDB) GetState(a common.Address, b common.Hash) common.Hash {
	stateObject := sdb.GetStateObject(a)
	if stateObject != nil {
		return stateObject.GetState(sdb.db, b)
	}
	return common.Hash{}
}

func (sdb *StateDB) HasSuicided(addr common.Address) bool {
	stateObject := sdb.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.suicided
	}
	return false
}

/*
 * SETTERS
 */

func (sdb *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObject := sdb.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}

func (sdb *StateDB) SetBalance(addr common.Address, amount *big.Int) {
	stateObject := sdb.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(amount)
	}
}

func (sdb *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := sdb.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}

func (sdb *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := sdb.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(crypto.Keccak256Hash(code), code)
	}
}

func (sdb *StateDB) SetState(addr common.Address, key common.Hash, value common.Hash) {
	stateObject := sdb.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetState(sdb.db, key, value)
	}
}

// Suicide marks the given account as suicided.
// This clears the account balance.
//
// The account's state object is still available until the state is committed,
// GetStateObject will return a non-nil account after Suicide.
func (sdb *StateDB) Suicide(addr common.Address) bool {
	stateObject := sdb.GetStateObject(addr)
	if stateObject == nil {
		return false
	}
	sdb.journal = append(sdb.journal, suicideChange{
		account:     &addr,
		prev:        stateObject.suicided,
		prevbalance: new(big.Int).Set(stateObject.Balance()),
	})
	stateObject.markSuicided()
	stateObject.data.Balance = new(big.Int)
	return true
}

//
// Setting, updating & deleting state object methods
//

// updateStateObject writes the given object to the trie.
func (sdb *StateDB) updateStateObject(stateObject *StateObject) {
	addr := stateObject.Address()
	data, err := rlp.EncodeToBytes(stateObject)
	if err != nil {
		panic(fmt.Errorf("can't encode object at %x: %v", addr[:], err))
	}
	sdb.trie.Update(addr[:], data)
}

// deleteStateObject removes the given object from the state trie.
func (sdb *StateDB) deleteStateObject(stateObject *StateObject) {
	stateObject.deleted = true
	addr := stateObject.Address()
	sdb.trie.Delete(addr[:])
}

// Retrieve a state object given my the address. Returns nil if not found.
func (sdb *StateDB) GetStateObject(addr common.Address) (stateObject *StateObject) {
	// Prefer 'live' objects.
	sdb.lock.Lock()
	if obj := sdb.stateObjects[addr]; obj != nil {
		sdb.lock.Unlock()
		if obj.deleted {

			return nil
		}
		return obj
	}
	sdb.lock.Unlock()

	// Load the object from the database.
	enc := sdb.trie.Get(addr[:])
	if len(enc) == 0 {
		return nil
	}
	var data Account
	if err := rlp.DecodeBytes(enc, &data); err != nil {
		glog.Errorf("can't decode object at %x: %v", addr[:], err)
		return nil
	}
	// Insert into the live set.
	obj := newObject(sdb, addr, data, sdb.MarkStateObjectDirty)
	sdb.setStateObject(obj)
	return obj
}

func (sdb *StateDB) setStateObject(object *StateObject) {
	sdb.lock.Lock()
	sdb.stateObjects[object.Address()] = object
	sdb.lock.Unlock()
}

// Retrieve a state object or create a new state object if nil
func (sdb *StateDB) GetOrNewStateObject(addr common.Address) *StateObject {
	stateObject := sdb.GetStateObject(addr)
	if stateObject == nil || stateObject.deleted {
		stateObject, _ = sdb.createObject(addr)
	}
	return stateObject
}

// MarkStateObjectDirty adds the specified object to the dirty map to avoid costly
// state object cache iteration to find a handful of modified ones.
func (sdb *StateDB) MarkStateObjectDirty(addr common.Address) {
	sdb.stateObjectsDirty[addr] = struct{}{}
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (sdb *StateDB) createObject(addr common.Address) (newobj, prev *StateObject) {
	prev = sdb.GetStateObject(addr)
	newobj = newObject(sdb, addr, Account{}, sdb.MarkStateObjectDirty)
	newobj.setNonce(StartingNonce) // sets the object to dirty
	if prev == nil {
		if glog.V(logger.Debug) {
			glog.Infof("(+) %x\n", addr)
		}
		sdb.journal = append(sdb.journal, createObjectChange{account: &addr})
	} else {
		sdb.journal = append(sdb.journal, resetObjectChange{prev: prev})
	}
	sdb.setStateObject(newobj)
	return newobj, prev
}

// CreateAccount explicitly creates a state object. If a state object with the address
// already exists the balance is carried over to the new account.
//
// CreateAccount is called during the EVM CREATE operation. The situation might arise that
// a contract does the following:
//
//   1. sends funds to sha(account ++ (nonce + 1))
//   2. tx_create(sha(account ++ nonce)) (note that this gets the address of 1)
//
// Carrying over the balance ensures that Ether doesn't disappear.
func (sdb *StateDB) CreateAccount(addr common.Address) AccountObject {
	new, prev := sdb.createObject(addr)
	if prev != nil {
		new.setBalance(prev.data.Balance)
	}
	return new
}

// Copy creates a deep, independent copy of the state.
// Snapshots of the copied state cannot be applied to the copy.
func (sdb *StateDB) Copy() *StateDB {
	sdb.lock.Lock()
	defer sdb.lock.Unlock()

	// Copy all the basic fields, initialize the memory ones
	state := &StateDB{
		db:                sdb.db,
		trie:              sdb.trie,
		pastTries:         sdb.pastTries,
		codeSizeCache:     sdb.codeSizeCache,
		stateObjects:      make(map[common.Address]*StateObject, len(sdb.stateObjectsDirty)),
		stateObjectsDirty: make(map[common.Address]struct{}, len(sdb.stateObjectsDirty)),
		refund:            new(big.Int).Set(sdb.refund),
		logs:              make(map[common.Hash]Logs, len(sdb.logs)),
		logSize:           sdb.logSize,
	}
	// Copy the dirty states and logs
	for addr := range sdb.stateObjectsDirty {
		state.stateObjects[addr] = sdb.stateObjects[addr].deepCopy(state, state.MarkStateObjectDirty)
		state.stateObjectsDirty[addr] = struct{}{}
	}
	for hash, logs := range sdb.logs {
		state.logs[hash] = make(Logs, len(logs))
		copy(state.logs[hash], logs)
	}
	return state
}

// Snapshot returns an identifier for the current revision of the state.
func (sdb *StateDB) Snapshot() int {
	id := sdb.nextRevisionId
	sdb.nextRevisionId++
	sdb.validRevisions = append(sdb.validRevisions, revision{id, len(sdb.journal)})
	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (sdb *StateDB) RevertToSnapshot(revid int) {
	// Find the snapshot in the stack of valid snapshots.
	idx := sort.Search(len(sdb.validRevisions), func(i int) bool {
		return sdb.validRevisions[i].id >= revid
	})
	if idx == len(sdb.validRevisions) || sdb.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := sdb.validRevisions[idx].journalIndex

	// Replay the journal to undo changes.
	for i := len(sdb.journal) - 1; i >= snapshot; i-- {
		sdb.journal[i].undo(sdb)
	}
	sdb.journal = sdb.journal[:snapshot]

	// Remove invalidated snapshots from the stack.
	sdb.validRevisions = sdb.validRevisions[:idx]
}

// GetRefund returns the current value of the refund counter.
// The return value must not be modified by the caller and will become
// invalid at the next call to AddRefund.
func (sdb *StateDB) GetRefund() *big.Int {
	return sdb.refund
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (sdb *StateDB) IntermediateRoot() common.Hash {
	for addr := range sdb.stateObjectsDirty {
		stateObject := sdb.stateObjects[addr]
		if stateObject.suicided {
			sdb.deleteStateObject(stateObject)
		} else {
			stateObject.updateRoot(sdb.db)
			sdb.updateStateObject(stateObject)
		}
	}
	// Invalidate journal because reverting across transactions is not allowed.
	sdb.clearJournalAndRefund()
	return sdb.trie.Hash()
}

// DeleteSuicides flags the suicided objects for deletion so that it
// won't be referenced again when called / queried up on.
//
// DeleteSuicides should not be used for consensus related updates
// under any circumstances.
func (sdb *StateDB) DeleteSuicides() {
	// Reset refund so that any used-gas calculations can use this method.
	sdb.clearJournalAndRefund()

	for addr := range sdb.stateObjectsDirty {
		stateObject := sdb.stateObjects[addr]

		// If the object has been removed by a suicide
		// flag the object as deleted.
		if stateObject.suicided {
			stateObject.deleted = true
		}
		delete(sdb.stateObjectsDirty, addr)
	}
}

// Commit commits all state changes to the database.
func (sdb *StateDB) Commit() (root common.Hash, err error) {
	root, batch := sdb.CommitBatch()
	return root, batch.Write()
}

// CommitBatch commits all state changes to a write batch but does not
// execute the batch. It is used to validate state changes against
// the root hash stored in a block.
func (sdb *StateDB) CommitBatch() (root common.Hash, batch ethdb.Batch) {
	batch = sdb.db.NewBatch()
	root, _ = sdb.commit(batch)
	return root, batch
}

func (sdb *StateDB) clearJournalAndRefund() {
	sdb.journal = nil
	sdb.validRevisions = sdb.validRevisions[:0]
	sdb.refund = new(big.Int)
}

func (sdb *StateDB) commit(dbw trie.DatabaseWriter) (root common.Hash, err error) {
	defer sdb.clearJournalAndRefund()

	// Commit objects to the trie.
	for addr, stateObject := range sdb.stateObjects {
		if stateObject.suicided {
			// If the object has been removed, don't bother syncing it
			// and just mark it for deletion in the trie.
			sdb.deleteStateObject(stateObject)
		} else if _, ok := sdb.stateObjectsDirty[addr]; ok {
			// Write any contract code associated with the state object
			if stateObject.code != nil && stateObject.dirtyCode {
				if err := dbw.Put(stateObject.CodeHash(), stateObject.code); err != nil {
					return common.Hash{}, err
				}
				stateObject.dirtyCode = false
			}
			// Write any storage changes in the state object to its storage trie.
			if err := stateObject.CommitTrie(sdb.db, dbw); err != nil {
				return common.Hash{}, err
			}
			// Update the object in the main account trie.
			sdb.updateStateObject(stateObject)
		}
		delete(sdb.stateObjectsDirty, addr)
	}
	// Write trie changes.
	root, err = sdb.trie.CommitTo(dbw)
	if err == nil {
		sdb.pushTrie(sdb.trie)
	}
	return root, err
}
