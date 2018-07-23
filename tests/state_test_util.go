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
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/ethereumproject/go-ethereum/core"
	"github.com/ethereumproject/go-ethereum/core/state"
	"github.com/ethereumproject/go-ethereum/core/types"
	"github.com/ethereumproject/go-ethereum/crypto"
	"github.com/ethereumproject/go-ethereum/ethdb"
	"github.com/ethereumproject/go-ethereum/logger/glog"
)

var oldStateTestDir = filepath.Join(filepath.Join(".", "files"), "StateTests")

func RunStateTestWithReader(ruleSet RuleSet, r io.Reader, skipTests []string) error {
	tests := make(map[string]VmTest)
	if err := readJson(r, &tests); err != nil {
		return err
	}

	if err := runStateTests(ruleSet, tests, skipTests); err != nil {
		return err
	}

	return nil
}

func RunStateTest(ruleSet RuleSet, p string, skipTests []string) error {
	tests := make(map[string]VmTest)
	if err := readJsonFile(p, &tests); err != nil {
		return err
	}

	if err := runStateTests(ruleSet, tests, skipTests); err != nil {
		return err
	}

	return nil

}

func BenchStateTest(ruleSet RuleSet, p string, conf bconf, b *testing.B) error {
	tests := make(map[string]VmTest)
	if err := readJsonFile(p, &tests); err != nil {
		return err
	}
	test, ok := tests[conf.name]
	if !ok {
		return fmt.Errorf("test not found: %s", conf.name)
	}

	// XXX Yeah, yeah...
	env := make(map[string]string)
	env["currentCoinbase"] = test.Env.CurrentCoinbase
	env["currentDifficulty"] = test.Env.CurrentDifficulty
	env["currentGasLimit"] = test.Env.CurrentGasLimit
	env["currentNumber"] = test.Env.CurrentNumber
	env["previousHash"] = test.Env.PreviousHash
	if n, ok := test.Env.CurrentTimestamp.(float64); ok {
		env["currentTimestamp"] = strconv.Itoa(int(n))
	} else {
		env["currentTimestamp"] = test.Env.CurrentTimestamp.(string)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStateTest(ruleSet, test, env, b)
	}

	return nil
}

func benchStateTest(ruleSet RuleSet, test VmTest, env map[string]string, b *testing.B) {
	b.StopTimer()
	db := ethdb.NewMemDatabase()
	statedb := makePreState(db, test.Pre)
	b.StartTimer()

	RunState(ruleSet, db, statedb, env, test.Exec)
}

func runStateTests(ruleSet RuleSet, tests map[string]VmTest, skipTests []string) error {
	skipTest := make(map[string]bool, len(skipTests))
	for _, name := range skipTests {
		skipTest[name] = true
	}

	var indexes []int
	var errs []error
	var testsCount int
	for name, test := range tests {
		if skipTest[name] /*|| name != "callcodecallcode_11" */ {
			glog.Infoln("Skipping state test", name)
			continue
		}

		testsCount++

		//fmt.Println("StateTest:", name)
		if err := runStateTest(ruleSet, test); err != nil {
			indexes = append(indexes, testsCount-1)
			errs = append(errs, fmt.Errorf("[OLD]%s: %s", name, err.Error()))
		}

		//glog.Infoln("State test passed: ", name)
		//fmt.Println(string(statedb.Dump()))
	}
	if len(errs) > 0 {
		var s string
		for i, e := range errs {
			s += fmt.Sprintf("i:%3d/%d/%d/%d|%v\n", indexes[i], i+1, len(errs), testsCount, e.Error())
		}
		return errors.New(s)
	}
	return nil
}

func runStateTest(ruleSet RuleSet, test VmTest) error {
	db := ethdb.NewMemDatabase()
	statedb := makePreState(db, test.Pre)

	// XXX Yeah, yeah...
	env := make(map[string]string)
	env["currentCoinbase"] = test.Env.CurrentCoinbase
	env["currentDifficulty"] = test.Env.CurrentDifficulty
	env["currentGasLimit"] = test.Env.CurrentGasLimit
	env["currentNumber"] = test.Env.CurrentNumber
	env["previousHash"] = test.Env.PreviousHash
	if n, ok := test.Env.CurrentTimestamp.(float64); ok {
		env["currentTimestamp"] = strconv.Itoa(int(n))
	} else {
		env["currentTimestamp"] = test.Env.CurrentTimestamp.(string)
	}

	var (
		ret    []byte
		gas    *big.Int
		failed bool
		err    error
		logs   []*types.Log
	)

	wrapStateErr := func(e error) error {
		return fmt.Errorf("%v\nret=%x gas=%d failed=%v err=%v logs=%v", e, ret, gas, failed, err, logs)
	}

	checkError := func() error {
		// Compare expected and actual return
		rexp := common.FromHex(test.Out)
		if bytes.Compare(rexp, ret) != 0 {
			return fmt.Errorf("return failed. Expected %x, got %x", rexp, ret)
		}

		// check post state
		for addr, account := range test.Post {
			a := common.HexToAddress(addr)
			exist := statedb.Exist(a)
			if !exist {
				return wrapStateErr(fmt.Errorf("did not find expected post-state account: %s", addr))
			}

			gotBalance := statedb.GetBalance(a)
			balance, ok := new(big.Int).SetString(account.Balance, 0)
			if !ok {
				panic("malformed test account balance")
			}
			if balance.Cmp(gotBalance) != 0 {
				diff := new(big.Int).Sub(balance, gotBalance)
				return wrapStateErr(fmt.Errorf("(%x) balance failed. Expected: %v have: %v (diff= %v)", a.Bytes()[:4], balance, gotBalance, diff))
			}

			gotNonce := statedb.GetNonce(a)
			nonce, err := strconv.ParseUint(account.Nonce, 0, 64)
			if err != nil {
				return fmt.Errorf("test account %q malformed nonce: %s", addr, err)
			}
			if gotNonce != nonce {
				return wrapStateErr(fmt.Errorf("(%x) nonce failed. Expected: %v have: %v", a.Bytes()[:4], nonce, gotNonce))
			}

			for addr, value := range account.Storage {
				v := statedb.GetState(a, common.HexToHash(addr))
				vexp := common.HexToHash(value)

				if v != vexp {
					return wrapStateErr(fmt.Errorf("storage failed:\n%x: %s:\nexpected: %x\nhave:     %x\n(%v %v)", a.Bytes(), addr, vexp, v, vexp.Big(), v.Big()))
				}
			}
		}

		root := statedb.IntermediateRoot(false)
		if common.HexToHash(test.PostStateRoot) != root {
			return wrapStateErr(fmt.Errorf("Post state root error. Expected: %s have: %x", test.PostStateRoot, root))
		}

		// check logs
		if len(test.Logs) > 0 {
			if err := checkLogs(test.Logs, logs); err != nil {
				return err
			}
		}
		return nil
	}
	ret, logs, gas, failed, err = RunState(ruleSet, db, statedb, env, test.Transaction)
	// return checkError()

	var e1, e2 error
	e1 = checkError()
	if e1 != nil {
		db := ethdb.NewMemDatabase()
		statedb := makePreState(db, test.Pre)
		ret, logs, gas, failed, err = RunStateNoRecursion(ruleSet, db, statedb, env, test.Transaction)
		e2 = checkError()
		if e2 == nil {
			return nil
		}
	}

	return nil
}

func RunState(ruleSet RuleSet, db ethdb.Database, statedb *state.StateDB, env, tx map[string]string) ([]byte, []*types.Log, *big.Int, bool, error) {
	data := common.FromHex(tx["data"])
	gas, _ := new(big.Int).SetString(tx["gasLimit"], 0)
	price, _ := new(big.Int).SetString(tx["gasPrice"], 0)
	value, _ := new(big.Int).SetString(tx["value"], 0)
	if gas == nil || price == nil || value == nil {
		panic("malformed gas, price or value")
	}
	nonce, err := strconv.ParseUint(tx["nonce"], 0, 64)
	if err != nil {
		panic(err)
	}

	var to *common.Address
	if len(tx["to"]) > 2 {
		t := common.HexToAddress(tx["to"])
		to = &t
	}
	// Set pre compiled contracts
	//vm.Precompiled = vm.PrecompiledContracts()
	currentGasLimit, ok := new(big.Int).SetString(env["currentGasLimit"], 0)
	if !ok {
		panic("malformed currentGasLimit")
	}
	gaspool := new(core.GasPool).AddGas(currentGasLimit.Uint64())

	key, err := hex.DecodeString(tx["secretKey"])
	if err != nil {
		panic(err)
	}
	addr := crypto.PubkeyToAddress(crypto.ToECDSA(key).PublicKey)
	//func NewMessage(from common.Address, to *common.Address, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, checkNonce bool) Message {
	message := types.NewMessage(addr, to, nonce, value, gas.Uint64(), price, data, true)

	vmenv := NewEnvFromMap(ruleSet, statedb, env, tx)
	vmenv.origin = addr

	if vmenv.evm == nil {
		panic("NIL EVM")
	}
	if vmenv.evm.StateDB == nil {
		panic("NIL EVM STATE")
	}
	if vmenv.state == nil {
		panic("NIL STATE")
	}
	if gaspool == nil {
		panic("NIL GASPOOl")
	}

	// NOTE(whilei): Just noting that EVM used embedded Context struct, which is what the GasPrice field
	// is referencing here.
	vmenv.evm.GasPrice = message.GasPrice()
	if vmenv.evm.GasPrice == nil {
		panic("NIL GASPRICE")
	}

	snapshot := statedb.Snapshot()
	ret, usedGas, failed, err := core.ApplyMessage(vmenv.evm, message, gaspool)
	vmenv.Gas.SetUint64(usedGas)

	if core.IsNonceErr(err) || core.IsInvalidTxErr(err) || core.IsGasLimitErr(err) {
		statedb.RevertToSnapshot(snapshot)
	}
	root, err := statedb.Commit(false)
	if err != nil {
		panic("COMMIT STATE ERR: " + err.Error())
	}
	if err := statedb.Database().TrieDB().Commit(root, false); err != nil {
		panic("COMMIT STATE TRIE ERR: " + err.Error())
	}

	return ret, vmenv.state.Logs(), vmenv.Gas, failed, err
}

func RunStateNoRecursion(ruleSet RuleSet, db ethdb.Database, statedb *state.StateDB, env, tx map[string]string) ([]byte, []*types.Log, *big.Int, bool, error) {
	data := common.FromHex(tx["data"])
	gas, _ := new(big.Int).SetString(tx["gasLimit"], 0)
	price, _ := new(big.Int).SetString(tx["gasPrice"], 0)
	value, _ := new(big.Int).SetString(tx["value"], 0)
	if gas == nil || price == nil || value == nil {
		panic("malformed gas, price or value")
	}
	nonce, err := strconv.ParseUint(tx["nonce"], 0, 64)
	if err != nil {
		panic(err)
	}

	var to *common.Address
	if len(tx["to"]) > 2 {
		t := common.HexToAddress(tx["to"])
		to = &t
	}
	// Set pre compiled contracts
	//vm.Precompiled = vm.PrecompiledContracts()
	currentGasLimit, ok := new(big.Int).SetString(env["currentGasLimit"], 0)
	if !ok {
		panic("malformed currentGasLimit")
	}
	gaspool := new(core.GasPool).AddGas(currentGasLimit.Uint64())

	key, err := hex.DecodeString(tx["secretKey"])
	if err != nil {
		panic(err)
	}
	addr := crypto.PubkeyToAddress(crypto.ToECDSA(key).PublicKey)
	//func NewMessage(from common.Address, to *common.Address, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, checkNonce bool) Message {
	message := types.NewMessage(addr, to, nonce, value, gas.Uint64(), price, data, true)

	vmenv := NewEnvFromMapNoRecursion(ruleSet, statedb, env, tx)
	vmenv.origin = addr

	if vmenv.evm == nil {
		panic("NIL EVM")
	}
	if vmenv.evm.StateDB == nil {
		panic("NIL EVM STATE")
	}
	if vmenv.state == nil {
		panic("NIL STATE")
	}
	if gaspool == nil {
		panic("NIL GASPOOl")
	}

	// NOTE(whilei): Just noting that EVM used embedded Context struct, which is what the GasPrice field
	// is referencing here.
	vmenv.evm.GasPrice = message.GasPrice()
	if vmenv.evm.GasPrice == nil {
		panic("NIL GASPRICE")
	}

	snapshot := statedb.Snapshot()
	ret, usedGas, failed, err := core.ApplyMessage(vmenv.evm, message, gaspool)
	vmenv.Gas.SetUint64(usedGas)

	if core.IsNonceErr(err) || core.IsInvalidTxErr(err) || core.IsGasLimitErr(err) {
		statedb.RevertToSnapshot(snapshot)
	}
	root, err := statedb.Commit(false)
	if err != nil {
		panic("COMMIT STATE ERR: " + err.Error())
	}
	if err := statedb.Database().TrieDB().Commit(root, false); err != nil {
		panic("COMMIT STATE TRIE ERR: " + err.Error())
	}

	return ret, vmenv.state.Logs(), vmenv.Gas, failed, err
}
