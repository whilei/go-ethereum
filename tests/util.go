// Copyright 2015 The go-ethereum Authors
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

package tests

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/ethereumproject/go-ethereum/core"
	"github.com/ethereumproject/go-ethereum/core/state"
	"github.com/ethereumproject/go-ethereum/core/types"
	"github.com/ethereumproject/go-ethereum/core/vm"
	"github.com/ethereumproject/go-ethereum/crypto"
	"github.com/ethereumproject/go-ethereum/ethdb"
	"github.com/ethereumproject/go-ethereum/logger/glog"
	"github.com/ethereumproject/go-ethereum/params"
)

func init() {
	glog.SetV(0)
}

func checkLogs(tlog []Log, logs []*types.Log) error {

	if len(tlog) != len(logs) {
		return fmt.Errorf("log length mismatch. Expected %d, got %d", len(tlog), len(logs))
	} else {
		for i, log := range tlog {
			if common.HexToAddress(log.AddressF) != logs[i].Address {
				return fmt.Errorf("log address expected %v got %x", log.AddressF, logs[i].Address)
			}

			if !bytes.Equal(logs[i].Data, common.FromHex(log.DataF)) {
				return fmt.Errorf("log data expected %v got %x", log.DataF, logs[i].Data)
			}

			if len(log.TopicsF) != len(logs[i].Topics) {
				return fmt.Errorf("log topics length expected %d got %d", len(log.TopicsF), logs[i].Topics)
			} else {
				for j, topic := range log.TopicsF {
					if common.HexToHash(topic) != logs[i].Topics[j] {
						return fmt.Errorf("log topic[%d] expected %v got %x", j, topic, logs[i].Topics[j])
					}
				}
			}
			genBloom := common.LeftPadBytes(types.LogsBloom([]*types.Log{logs[i]}).Bytes(), 256)

			if !bytes.Equal(genBloom, common.Hex2Bytes(log.BloomF)) {
				return fmt.Errorf("bloom mismatch")
			}
		}
	}
	return nil
}

type Account struct {
	Balance string
	Code    string
	Nonce   string
	Storage map[string]string
}

type Log struct {
	AddressF string   `json:"address"`
	DataF    string   `json:"data"`
	TopicsF  []string `json:"topics"`
	BloomF   string   `json:"bloom"`
}

func (self Log) Address() []byte      { return common.Hex2Bytes(self.AddressF) }
func (self Log) Data() []byte         { return common.Hex2Bytes(self.DataF) }
func (self Log) RlpData() interface{} { return nil }
func (self Log) Topics() [][]byte {
	t := make([][]byte, len(self.TopicsF))
	for i, topic := range self.TopicsF {
		t[i] = common.Hex2Bytes(topic)
	}
	return t
}

func makePreState(db ethdb.Database, accounts map[string]Account) *state.StateDB {
	sdb := state.NewDatabase(db)
	statedb, _ := state.New(common.Hash{}, sdb)
	for addr, account := range accounts {
		insertAccount(statedb, addr, account)
	}
	return statedb
}

func insertAccount(state *state.StateDB, saddr string, account Account) {
	if common.IsHex(account.Code) {
		account.Code = account.Code[2:]
	}
	addr := common.HexToAddress(saddr)
	state.SetCode(addr, common.Hex2Bytes(account.Code))
	if i, err := strconv.ParseUint(account.Nonce, 0, 64); err != nil {
		panic(err)
	} else {
		state.SetNonce(addr, i)
	}
	if i, ok := new(big.Int).SetString(account.Balance, 0); !ok {
		panic("malformed account balance")
	} else {
		state.SetBalance(addr, i)
	}
	for a, v := range account.Storage {
		state.SetState(addr, common.HexToHash(a), common.HexToHash(v))
	}
}

type VmEnv struct {
	CurrentCoinbase   string
	CurrentDifficulty string
	CurrentGasLimit   string
	CurrentNumber     string
	CurrentTimestamp  interface{}
	PreviousHash      string
}

type VmTest struct {
	Callcreates interface{}
	//Env         map[string]string
	Env           VmEnv
	Exec          map[string]string
	Transaction   map[string]string
	Logs          []Log
	Gas           string
	Out           string
	Post          map[string]Account
	Pre           map[string]Account
	PostStateRoot string
}

type RuleSet struct {
	HomesteadBlock           *big.Int
	HomesteadGasRepriceBlock *big.Int
	DiehardBlock             *big.Int
	ExplosionBlock           *big.Int
}

func (r RuleSet) GasTable(num *big.Int) *params.GasTable {
	if r.HomesteadGasRepriceBlock == nil || num == nil || num.Cmp(r.HomesteadGasRepriceBlock) < 0 {
		return &params.GasTable{
			ExtcodeSize:     uint64(20),
			ExtcodeCopy:     uint64(20),
			Balance:         uint64(20),
			SLoad:           uint64(50),
			Calls:           uint64(40),
			Suicide:         uint64(0),
			ExpByte:         uint64(10),
			CreateBySuicide: uint64(0),
		}
	}
	if r.DiehardBlock == nil || num == nil || num.Cmp(r.DiehardBlock) < 0 {
		return &params.GasTable{
			ExtcodeSize:     uint64(700),
			ExtcodeCopy:     uint64(700),
			Balance:         uint64(400),
			SLoad:           uint64(200),
			Calls:           uint64(700),
			Suicide:         uint64(5000),
			ExpByte:         uint64(10),
			CreateBySuicide: uint64(25000),
		}
	}

	return &params.GasTable{
		ExtcodeSize:     uint64(700),
		ExtcodeCopy:     uint64(700),
		Balance:         uint64(400),
		SLoad:           uint64(200),
		Calls:           uint64(700),
		Suicide:         uint64(5000),
		ExpByte:         uint64(50),
		CreateBySuicide: uint64(25000),
	}
}

type Env struct {
	ruleSet      RuleSet
	depth        int
	state        *state.StateDB
	skipTransfer bool
	initial      bool
	Gas          *big.Int

	origin   common.Address
	parent   common.Hash
	coinbase common.Address

	number     *big.Int
	time       *big.Int
	difficulty *big.Int
	gasLimit   *big.Int

	vmTest bool

	evm *vm.EVM
}

func NewEnv(ruleSet RuleSet, state *state.StateDB) *Env {
	env := &Env{
		ruleSet: ruleSet,
		state:   state,
	}
	return env
}

func (env *Env) VmContext() vm.Context {
	return vm.Context{
		CanTransfer: env.CanTransfer,
		Transfer:    env.Transfer,
		GetHash:     vmTestBlockHash,
		Origin:      env.origin,
		GasPrice:    big.NewInt(0), // new(big.Int).Set(msg.GasPrice()), // this gets set per test via env.Call
		Coinbase:    env.coinbase,
		GasLimit:    env.gasLimit.Uint64(),
		BlockNumber: env.number,
		Time:        env.time,
		Difficulty:  env.difficulty,
	}
}

func NewEnvFromMap(ruleSet RuleSet, state *state.StateDB, envValues map[string]string, exeValues map[string]string) *Env {
	env := NewEnv(ruleSet, state)

	env.origin = common.HexToAddress(exeValues["caller"])
	env.parent = common.HexToHash(envValues["previousHash"])
	env.coinbase = common.HexToAddress(envValues["currentCoinbase"])
	env.number, _ = new(big.Int).SetString(envValues["currentNumber"], 0)
	if env.number == nil {
		panic("malformed current number")
	}
	env.time, _ = new(big.Int).SetString(envValues["currentTimestamp"], 0)
	if env.time == nil {
		panic("malformed current timestamp")
	}
	env.difficulty, _ = new(big.Int).SetString(envValues["currentDifficulty"], 0)
	if env.difficulty == nil {
		panic("malformed current difficulty")
	}
	env.gasLimit, _ = new(big.Int).SetString(envValues["currentGasLimit"], 0)
	if env.gasLimit == nil {
		panic("malformed current gas limit")
	}
	env.Gas = new(big.Int)

	// build a chain config corresponding to given ruleset
	chainConf := params.DefaultConfigMorden.ChainConfig
	chainConf.HomesteadBlock = ruleSet.HomesteadBlock
	chainConf.EIP150Block = ruleSet.HomesteadGasRepriceBlock
	chainConf.EIP155Block = ruleSet.DiehardBlock
	chainConf.EIP160Block = ruleSet.DiehardBlock
	chainConf.ByzantiumBlock = nil

	if chainConf.EIP160Block != nil {
		chainConf.ChainID = big.NewInt(1)
	} else {
		chainConf.ChainID = nil
	}

	// env.evm = vm.NewEVM(env.VmContext(), env.state, params.DefaultConfigMorden.ChainConfig, vm.Config{NoRecursion: true})
	env.evm = vm.NewEVM(env.VmContext(), env.state, chainConf, vm.Config{NoRecursion: true})

	return env
}

func vmTestBlockHash(n uint64) common.Hash {
	return common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(int64(n)).String())))
}

func (self *Env) CanTransfer(statedb vm.StateDB, from common.Address, balance *big.Int) bool {
	if self.skipTransfer {
		if self.initial {
			self.initial = false
			return true
		}
	}

	return statedb.GetBalance(from).Cmp(balance) >= 0
}

func (self *Env) Transfer(db vm.StateDB, from, to common.Address, amount *big.Int) {
	if self.skipTransfer {
		return
	}
	core.Transfer(db, from.Address(), to.Address(), amount)
}

func (self *Env) Call(caller vm.ContractRef, addr common.Address, data []byte, gas uint64, price, value *big.Int) ([]byte, error) {
	self.evm.GasPrice = price
	if self.vmTest && self.depth > 0 {
		self.state.AddBalance(caller.Address(), new(big.Int).Mul(new(big.Int).SetUint64(gas), price))
		return nil, nil
	}

	// original gas in bigInt
	og := big.NewInt(0).SetUint64(gas)
	self.Gas = og

	ret, leftoverGas, err := self.evm.Call(caller, addr, data, gas, value)

	// deduce spent gas from (original gas - leftover gas), and set as test resulting gas
	// leftoverGas in bigInt
	logB := big.NewInt(0).SetUint64(leftoverGas)
	spentGas := new(big.Int).Sub(og, logB)
	self.Gas.Sub(og, spentGas)

	return ret, err

}

func (self *Env) CallCode(caller vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	panic("unimplemented")
}

func (self *Env) DelegateCall(caller vm.ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error) {
	panic("unimplemented")
}

func (self *Env) Create(caller vm.ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error) {
	panic("unimplemented")
}
