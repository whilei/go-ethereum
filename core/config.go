// Copyright 2016 The go-ethereum Authors
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

package core

import (
	hexlib "encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"sort"
	"sync"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/ethereumproject/go-ethereum/core/state"
	"github.com/ethereumproject/go-ethereum/core/types"
	"github.com/ethereumproject/go-ethereum/core/vm"
	"github.com/ethereumproject/go-ethereum/ethdb"
	"github.com/ethereumproject/go-ethereum/logger"
	"github.com/ethereumproject/go-ethereum/logger/glog"
)

var (
	ErrChainConfigNotFound     = errors.New("ChainConfig not found")
	ErrChainConfigForkNotFound = errors.New("ChainConfig fork not found")

	ErrHashKnownBad  = errors.New("known bad hash")
	ErrHashKnownFork = validateError("known fork hash mismatch")
)

// SufficientChainConfig holds necessary data for externalizing a given blockchain configuration.
type SufficientChainConfig struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Genesis     *GenesisDump `json:"genesis"`
	ChainConfig *ChainConfig `json:"chainConfig"`
	Bootstrap   []string     `json:"bootstrap"`
}

// GenesisDump is the geth JSON format.
// https://github.com/ethereumproject/wiki/wiki/Ethereum-Chain-Spec-Format#subformat-genesis
type GenesisDump struct {
	Nonce      prefixedHex `json:"nonce"`
	Timestamp  prefixedHex `json:"timestamp"`
	ParentHash prefixedHex `json:"parentHash"`
	ExtraData  prefixedHex `json:"extraData"`
	GasLimit   prefixedHex `json:"gasLimit"`
	Difficulty prefixedHex `json:"difficulty"`
	Mixhash    prefixedHex `json:"mixhash"`
	Coinbase   prefixedHex `json:"coinbase"`

	// Alloc maps accounts by their address.
	Alloc map[hex]*GenesisDumpAlloc `json:"alloc"`
}

// GenesisDumpAlloc is a GenesisDump.Alloc entry.
type GenesisDumpAlloc struct {
	Code    prefixedHex `json:"-"` // skip field for json encode
	Storage map[hex]hex `json:"-"`
	Balance string      `json:"balance"` // decimal string
}

type GenesisAccount struct {
	Address common.Address `json:"address"`
	Balance *big.Int       `json:"balance"`
}

// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	// Forks holds fork block requirements. See ErrHashKnownFork.
	Forks Forks `json:"forks"`
	// BadHashes holds well known blocks with consensus issues. See ErrHashKnownBad.
	BadHashes []*BadHash `json:"badHashes"`
}

type Fork struct {
	Name string `json:"name"`
	// Block is the block number where the hard-fork commences on
	// the Ethereum network.
	Block *big.Int `json:"block"`
	// Used to improve sync for a known network split
	RequiredHash common.Hash `json:"-"`
	// Configurable features.
	Features []*ForkFeature `json:"features"`
}

// Forks implements sort interface, sorting by block number
type Forks []*Fork

func (fs Forks) Len() int { return len(fs) }
func (fs Forks) Less(i, j int) bool {
	iF := fs[i]
	jF := fs[j]
	return iF.Block.Cmp(jF.Block) < 0
}
func (fs Forks) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}

// ForkFeatures are designed to decouple the implementation feature upgrades from Forks themselves.
// For example, there are several 'set-gasprice' features, each using a different gastable,
// as well as protocol upgrades including 'eip155', 'ecip1010', ... etc.
type ForkFeature struct {
	ID                string                    `json:"id"`
	Options           ChainFeatureConfigOptions `json:"options"` // no * because they have to be iterable(?)
	optionsLock       sync.RWMutex
	ParsedOptions     map[string]interface{} `json:"-"` // don't include in JSON dumps, since its for holding parsed JSON in mem
	parsedOptionsLock sync.RWMutex
	// TODO Derive Oracle contracts from fork struct (Version, Registrar, Release)
}

// These are the raw key-value configuration options made available
// by an external JSON file.
type ChainFeatureConfigOptions map[string]interface{}

type BadHash struct {
	Block *big.Int
	Hash  common.Hash
}

// Header returns the mapping.
func (g *GenesisDump) Header() (*types.Header, error) {
	var h types.Header

	var err error
	if err = g.Nonce.Decode(h.Nonce[:]); err != nil {
		return nil, fmt.Errorf("malformed nonce: %s", err)
	}
	if h.Time, err = g.Timestamp.Int(); err != nil {
		return nil, fmt.Errorf("malformed timestamp: %s", err)
	}
	if err = g.ParentHash.Decode(h.ParentHash[:]); err != nil {
		return nil, fmt.Errorf("malformed parentHash: %s", err)
	}
	if h.Extra, err = g.ExtraData.Bytes(); err != nil {
		return nil, fmt.Errorf("malformed extraData: %s", err)
	}
	if h.GasLimit, err = g.GasLimit.Int(); err != nil {
		return nil, fmt.Errorf("malformed gasLimit: %s", err)
	}
	if h.Difficulty, err = g.Difficulty.Int(); err != nil {
		return nil, fmt.Errorf("malformed difficulty: %s", err)
	}
	if err = g.Mixhash.Decode(h.MixDigest[:]); err != nil {
		return nil, fmt.Errorf("malformed mixhash: %s", err)
	}
	if err := g.Coinbase.Decode(h.Coinbase[:]); err != nil {
		return nil, fmt.Errorf("malformed coinbase: %s", err)
	}

	return &h, nil
}

// SortForks sorts a ChainConfiguration's forks by block number smallest to bigget (chronologically).
// This should need be called only once after construction
func (c *ChainConfig) SortForks() *ChainConfig {
	sort.Sort(c.Forks)
	return c
}

// IsHomestead returns whether num is either equal to the homestead block or greater.
func (c *ChainConfig) IsHomestead(num *big.Int) bool {
	if c.ForkByName("Homestead").Block == nil || num == nil {
		return false
	}
	return num.Cmp(c.ForkByName("Homestead").Block) >= 0
}

// IsDiehard returns whether num is greater than or equal to the Diehard block, but less than explosion.
func (c *ChainConfig) IsDiehard(num *big.Int) bool {
	fork := c.ForkByName("Diehard")
	if fork.Block == nil || num == nil {
		return false
	}
	return num.Cmp(fork.Block) >= 0
}

// IsExplosion returns whether num is either equal to the explosion block or greater.
func (c *ChainConfig) IsExplosion(num *big.Int) bool {
	feat, fork, configured := c.GetFeature(num, "difficulty")

	if configured {
		//name, exists := feat.GetStringOptions("type")
		if name, exists := feat.GetStringOptions("type"); exists && name == "ecip1010" {
			block := big.NewInt(0)
			if length, ok := feat.GetBigInt("length"); ok {
				block = block.Add(fork.Block, length)
			} else {
				panic("Fork feature ecip1010 requires length value.")
			}
			return num.Cmp(block) >= 0
		}
	}
	return false
}

// ForkByName looks up a Fork by its name, assumed to be unique
func (c *ChainConfig) ForkByName(name string) *Fork {
	for i := range c.Forks {
		if c.Forks[i].Name == name {
			return c.Forks[i]
		}
	}
	return &Fork{}
}

// GetFeature looks up fork features by id, where id can (currently) be [difficulty, gastable, eip155].
// GetFeature returns the feature|nil, the latest fork configuring a given id, and if the given feature id was found at all
// If queried feature is not found, returns *empty* ForkFeature, Fork, false
func (c *ChainConfig) GetFeature(num *big.Int, id string) (*ForkFeature, *Fork, bool) {
	var okForkFeature = &ForkFeature{}
	var okFork = &Fork{}
	var found = false
	for _, f := range c.Forks {
		if f.Block.Cmp(num) > 0 {
			break // NOTE: break assumes chronological
		}
		for _, ff := range f.Features {
			if ff.ID == id {
				okForkFeature = ff
				okFork = f
				found = true
			}
		}
	}
	return okForkFeature, okFork, found
}

func (c *ChainConfig) HeaderCheck(h *types.Header) error {
	for _, fork := range c.Forks {
		if fork.Block.Cmp(h.Number) != 0 {
			continue
		}
		if !fork.RequiredHash.IsEmpty() && fork.RequiredHash != h.Hash() {
			return ErrHashKnownFork
		}
	}

	for _, bad := range c.BadHashes {
		if bad.Block.Cmp(h.Number) != 0 {
			continue
		}
		if bad.Hash == h.Hash() {
			return ErrHashKnownBad
		}
	}

	return nil
}

func (c *ChainConfig) GetSigner(blockNumber *big.Int) types.Signer {
	feature, _, configured := c.GetFeature(blockNumber, "eip155")
	if configured {
		if chainId, ok := feature.GetBigInt("chainID"); ok {
			return types.NewChainIdSigner(chainId)
		} else {
			panic(fmt.Errorf("chainID is not set for EIP-155 at %v", blockNumber))
		}
	}
	return types.BasicSigner{}
}

// GasTable returns the gas table corresponding to the current fork
// The returned GasTable's fields shouldn't, under any circumstances, be changed.
func (c *ChainConfig) GasTable(num *big.Int) *vm.GasTable {
	defaultTable := &vm.GasTable{
		ExtcodeSize:     big.NewInt(20),
		ExtcodeCopy:     big.NewInt(20),
		Balance:         big.NewInt(20),
		SLoad:           big.NewInt(50),
		Calls:           big.NewInt(40),
		Suicide:         big.NewInt(0),
		ExpByte:         big.NewInt(10),
		CreateBySuicide: nil,
	}

	f, _, configured := c.GetFeature(num, "gastable")
	if !configured {
		return defaultTable
	}
	name, ok := f.GetStringOptions("type")
	if !ok {
		name = ""
	} // will wall to default panic
	switch name {
	case "homestead":
		return DefaultHomeSteadGasTable
	case "eip150":
		return DefaultGasRepriceGasTable
	case "eip160":
		return DefaultDiehardGasTable
	default:
		panic(fmt.Errorf("Unsupported gastable value '%v' at block: %v", name, num))
	}
}

// WriteToJSONFile writes a given config to a specified file path.
// It doesn't run any checks on the file path so make sure that's already squeaky clean.
func (c *SufficientChainConfig) WriteToJSONFile(path string) error {
	jsonConfig, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return fmt.Errorf("Could not marshal json from chain config: %v", err)
	}

	if err := ioutil.WriteFile(path, jsonConfig, 0644); err != nil {
		return fmt.Errorf("Could not write external chain config file: %v", err)
	}
	return nil
}

// ReadChainConfigFromJSONFile reads a given json file into a *ChainConfig.
// Again, no checks are made on the file path.
func ReadChainConfigFromJSONFile(path string) (*SufficientChainConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read external chain configuration file: %s", err)
	}
	defer f.Close()

	var config = &SufficientChainConfig{}
	if json.NewDecoder(f).Decode(config); err != nil {
		return nil, fmt.Errorf("%s: %s", path, err)
	}
	return config, nil
}

// GetStringOptions gets and option value for an options with key 'name',
// returning value as a string.
func (o *ForkFeature) GetStringOptions(name string) (string, bool) {
	if o.ParsedOptions == nil {
		o.parsedOptionsLock.Lock()
		o.ParsedOptions = make(map[string]interface{})
		o.parsedOptionsLock.Unlock()
	} else {
		o.parsedOptionsLock.RLock()
		val, ok := o.ParsedOptions[name]
		o.parsedOptionsLock.RUnlock()
		if ok {
			return val.(string), ok
		}
	}
	o.optionsLock.RLock()
	val, ok := o.Options[name].(string)
	o.optionsLock.RUnlock()
	o.parsedOptionsLock.Lock()
	o.ParsedOptions[name] = val //expect it as a string in config
	o.parsedOptionsLock.Unlock()
	return val, ok
}

// GetBigInt gets and option value for an options with key 'name',
// returning value as a *big.Int and ok if it exists.
func (o *ForkFeature) GetBigInt(name string) (*big.Int, bool) {

	if o.ParsedOptions == nil {
		o.parsedOptionsLock.Lock()
		o.ParsedOptions = make(map[string]interface{})
		o.parsedOptionsLock.Unlock()
	} else {
		o.parsedOptionsLock.RLock()
		val, ok := o.ParsedOptions[name]
		o.parsedOptionsLock.RUnlock()
		if ok {
			return val.(*big.Int), true
		}
	}
	o.optionsLock.RLock()
	originalValue := o.Options[name]
	o.optionsLock.RUnlock()
	if value, ok := originalValue.(int64); ok {
		i := big.NewInt(value)
		o.parsedOptionsLock.Lock()
		o.ParsedOptions[name] = i
		o.parsedOptionsLock.Unlock()
		return i, true
	}
	if value, ok := originalValue.(int); ok {
		i := big.NewInt(int64(value))
		o.parsedOptionsLock.Lock()
		o.ParsedOptions[name] = i
		o.parsedOptionsLock.Unlock()
		return i, true
	}
	if value, ok := originalValue.(string); ok {
		i, ok := new(big.Int).SetString(value, 0)
		if ok {
			o.parsedOptionsLock.Lock()
			o.ParsedOptions[name] = i
			o.parsedOptionsLock.Unlock()
		}
		return i, ok
	}
	return nil, false
}

// WriteGenesisBlock writes the genesis block to the database as block number 0
func WriteGenesisBlock(chainDb ethdb.Database, genesis *GenesisDump) (*types.Block, error) {
	statedb, err := state.New(common.Hash{}, chainDb)
	if err != nil {
		return nil, err
	}

	for addrHex, account := range genesis.Alloc {
		var addr common.Address
		if err := addrHex.Decode(addr[:]); err != nil {
			return nil, fmt.Errorf("malformed addres %q: %s", addrHex, err)
		}

		balance, ok := new(big.Int).SetString(account.Balance, 0)
		if !ok {
			return nil, fmt.Errorf("malformed account %q balance %q", addrHex, account.Balance)
		}
		statedb.AddBalance(addr, balance)

		code, err := account.Code.Bytes()
		if err != nil {
			return nil, fmt.Errorf("malformed account %q code: %s", addrHex, err)
		}
		statedb.SetCode(addr, code)

		for key, value := range account.Storage {
			var k, v common.Hash
			if err := key.Decode(k[:]); err != nil {
				return nil, fmt.Errorf("malformed account %q key: %s", addrHex, err)
			}
			if err := value.Decode(v[:]); err != nil {
				return nil, fmt.Errorf("malformed account %q value: %s", addrHex, err)
			}
			statedb.SetState(addr, k, v)
		}
	}
	root, stateBatch := statedb.CommitBatch()

	header, err := genesis.Header()
	if err != nil {
		return nil, err
	}
	header.Root = root

	block := types.NewBlock(header, nil, nil, nil)

	if block := GetBlock(chainDb, block.Hash()); block != nil {
		glog.V(logger.Info).Infoln("Genesis block already in chain. Writing canonical number")
		err := WriteCanonicalHash(chainDb, block.Hash(), block.NumberU64())
		if err != nil {
			return nil, err
		}
		return block, nil
	}

	if err := stateBatch.Write(); err != nil {
		return nil, fmt.Errorf("cannot write state: %v", err)
	}
	if err := WriteTd(chainDb, block.Hash(), header.Difficulty); err != nil {
		return nil, err
	}
	if err := WriteBlock(chainDb, block); err != nil {
		return nil, err
	}
	if err := WriteBlockReceipts(chainDb, block.Hash(), nil); err != nil {
		return nil, err
	}
	if err := WriteCanonicalHash(chainDb, block.Hash(), block.NumberU64()); err != nil {
		return nil, err
	}
	if err := WriteHeadBlockHash(chainDb, block.Hash()); err != nil {
		return nil, err
	}

	return block, nil
}

func WriteGenesisBlockForTesting(db ethdb.Database, accounts ...GenesisAccount) *types.Block {
	dump := GenesisDump{
		GasLimit:   "0x47E7C4",
		Difficulty: "0x020000",
		Alloc:      make(map[hex]*GenesisDumpAlloc, len(accounts)),
	}

	for _, a := range accounts {
		dump.Alloc[hex(hexlib.EncodeToString(a.Address[:]))] = &GenesisDumpAlloc{
			Balance: a.Balance.String(),
		}
	}

	block, err := WriteGenesisBlock(db, &dump)
	if err != nil {
		panic(err)
	}
	return block
}

// MakeGenesisDump makes a genesis dump
func MakeGenesisDump(chaindb ethdb.Database) (*GenesisDump, error) {

	genesis := GetBlock(chaindb, GetCanonicalHash(chaindb, 0))
	if genesis == nil {
		return nil, nil
	}

	// Settings.
	genesisHeader := genesis.Header()
	nonce := fmt.Sprintf(`0x%x`, genesisHeader.Nonce)
	time := common.BigToHash(genesisHeader.Time).Hex()
	parentHash := genesisHeader.ParentHash.Hex()
	extra := common.ToHex(genesisHeader.Extra)
	gasLimit := common.BigToHash(genesisHeader.GasLimit).Hex()
	difficulty := common.BigToHash(genesisHeader.Difficulty).Hex()
	mixHash := genesisHeader.MixDigest.Hex()
	coinbase := genesisHeader.Coinbase.Hex()

	var dump = &GenesisDump{
		Nonce:      prefixedHex(nonce), // common.ToHex(n)), // common.ToHex(
		Timestamp:  prefixedHex(time),
		ParentHash: prefixedHex(parentHash),
		ExtraData:  prefixedHex(extra),
		GasLimit:   prefixedHex(gasLimit),
		Difficulty: prefixedHex(difficulty),
		Mixhash:    prefixedHex(mixHash),
		Coinbase:   prefixedHex(coinbase),
		//Alloc: ,
	}

	// State allocations.
	genState, err := state.New(genesis.Root(), chaindb)
	if err != nil {
		return nil, err
	}
	stateDump := genState.RawDump()

	stateAccounts := stateDump.Accounts
	dump.Alloc = make(map[hex]*GenesisDumpAlloc, len(stateAccounts))

	for address, acct := range stateAccounts {
		if common.IsHexAddress(address) {
			dump.Alloc[hex(address)] = &GenesisDumpAlloc{
				Balance: acct.Balance,
			}
		} else {
			return nil, fmt.Errorf("Invalid address in genesis state: %v", address)
		}
	}
	return dump, nil
}

// ReadGenesisFromJSONFile allows the use a genesis file in JSON format.
// Implemented in `init` command via initGenesis method.
func ReadGenesisFromJSONFile(jsonFilePath string) (dump *GenesisDump, err error) {
	f, err := os.Open(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read genesis file: %s", err)
	}
	defer f.Close()

	dump = new(GenesisDump)
	if json.NewDecoder(f).Decode(dump); err != nil {
		return nil, fmt.Errorf("%s: %s", jsonFilePath, err)
	}
	return dump, nil
}
