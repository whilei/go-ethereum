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
	"math"
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
	"strings"
)

var oldStateTestDir = filepath.Join(filepath.Join(".", "files"), "StateTests")

func RunStateTestWithReader(ruleSet RuleSet, r io.Reader, skipTests []string) error {
	tests := make(map[string]VmTest)
	if err := readJson(r, &tests); err != nil {
		return err
	}

	if err := runStateTests(ruleSet, "", tests, skipTests); err != nil {
		return err
	}

	return nil
}

func RunStateTest(ruleSet RuleSet, p string, skipTests []string) error {
	tests := make(map[string]VmTest)
	if err := readJsonFile(p, &tests); err != nil {
		return err
	}

	if err := runStateTests(ruleSet, p, tests, skipTests); err != nil {
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

func rulesetToString(rs RuleSet) string {
	switch true {
	case rs.HomesteadBlock == nil:
		return "Frontier"
	case rs.HomesteadGasRepriceBlock == nil:
		return "Homestead"
	case rs.DiehardBlock == nil:
		return "EIP150"
	default:
		return "Diehard"
		// case rs.ExplosionBlock == nil:
		// return "Bomb"
	}
	return ""
	// if rs.HomesteadBlock == nil {
	// 	return "Frontier"
	// } else if rs.HomesteadGasRepriceBlock == nil {
	// 	return "Homestead"
	// } else if rs.DiehardBlock {
	// 	return "EIP150"
	// }
}

func runStateTests(ruleSet RuleSet, path string, tests map[string]VmTest, skipTests []string) error {
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
			errs = append(errs, fmt.Errorf("[OLD]path=%s name=%s ruleset=%s err=%s want_psr=%s", path, name, rulesetToString(ruleSet), err.Error(), test.PostStateRoot))
		}

		//glog.Infoln("State test passed: ", name)
		//fmt.Println(string(statedb.Dump()))
	}
	if len(errs) > 0 {
		var s string
		for i, e := range errs {
			s += fmt.Sprintf("\n- i=%3d/%d/%d/%d err=%v\n", indexes[i], i+1, len(errs), testsCount, e.Error())
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
		ret     []byte
		gas     *big.Int
		failed  bool
		err     error
		logs    []*types.Log
		gotRoot common.Hash
	)

	wrapStateErr := func(e error) error {
		if gotRoot.IsEmpty() {
			gotRoot = statedb.IntermediateRoot(false)
		}
		return fmt.Errorf("%v\nret=%x gas=%d failed=%v err=%v logs=%v root=%s", e, ret, gas, failed, err, logs, gotRoot.Hex())
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
			wantBalance, ok := new(big.Int).SetString(account.Balance, 0)
			if !ok {
				panic("malformed test account balance")
			}
			if wantBalance.Cmp(gotBalance) != 0 {
				if strings.Contains(account.Balance, "ffffffffffffffffffffffffffffff") && gotBalance.Uint64() == math.MaxUint64 {
					// if wantBalance.Cmp(common.MaxBig) == 0 {
					// 	panic("max max")
					// }
					// if strings.Contains(test.Pre[addr].Balance, "ffffffffffffffffffffffffffffff") {
					// 	panic("max pre bal")
					// }
					return wrapStateErr(fmt.Errorf("want/got:maxbig/maxuint64"))
				} else {
					diff := new(big.Int).Sub(wantBalance, gotBalance)
					return wrapStateErr(fmt.Errorf("(%x) balance failed. Expected: %v have: %v (diff= %v)", a.Bytes()[:4], wantBalance, gotBalance, diff))
				}
			}

			gotNonce := statedb.GetNonce(a)
			wantNonce, err := strconv.ParseUint(account.Nonce, 0, 64)
			if err != nil {
				return fmt.Errorf("test account %q malformed nonce: %s", addr, err)
			}
			if gotNonce != wantNonce {
				return wrapStateErr(fmt.Errorf("(%x) nonce failed. Expected: %v have: %v", a.Bytes()[:4], wantNonce, gotNonce))
			}

			for addr, value := range account.Storage {
				v := statedb.GetState(a, common.HexToHash(addr))
				vexp := common.HexToHash(value)

				if v != vexp {
					return wrapStateErr(fmt.Errorf("storage failed:\n%x: %s:\nexpected: %x\nhave:     %x\n(%v %v)", a.Bytes(), addr, vexp, v, vexp.Big(), v.Big()))
				}
			}
		}

		gotRoot = statedb.IntermediateRoot(false)
		if common.HexToHash(test.PostStateRoot) != gotRoot {
			return wrapStateErr(fmt.Errorf("Post state gotRoot error. Expected: %s have: %x", test.PostStateRoot, gotRoot))
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

	// var e1, e2 error
	var e1 error
	e1 = checkError()
	return e1
	// if e1 != nil {
	// 	// run test again with VM NoRecursion
	// 	db := ethdb.NewMemDatabase()
	// 	statedb := makePreState(db, test.Pre)
	// 	ret, logs, gas, failed, err = RunStateNoRecursion(ruleSet, db, statedb, env, test.Transaction)
	// 	e2 = checkError()
	// 	if e2 == nil {
	// 		return fmt.Errorf("requires VM.NoRecursion=true")
	// 		// return nil
	// 	} else if e1 != e2 {
	// 		return fmt.Errorf("[NoRecursion=false]%v\n[NoRecursion=true]%v", e1, e2)
	// 	} else {
	// 		// TODO maybe only return 1
	// 		return fmt.Errorf("[NoRecursion=false]%v\n[NoRecursion=true]%v", e1, e2)
	// 	}
	// }

	// return nil
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

	// debug checks for svm panic
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

	// Here, ret, usedGas, and failed are only assigned for debugging/logging reasons
	ret, usedGas, failed, err := core.ApplyMessage(vmenv.evm, message, gaspool)
	if err != nil {
		if err.Error() == "gas uint64 overflow" {
			panic(err.Error())
		}
		statedb.RevertToSnapshot(snapshot)
	}
	vmenv.Gas.SetUint64(usedGas)
	gotRoot, err := statedb.Commit(false)
	if err != nil {
		panic("COMMIT STATE ERR: " + err.Error())
	}
	if err := statedb.Database().TrieDB().Commit(gotRoot, false); err != nil {
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
	gotRoot, err := statedb.Commit(false)
	if err != nil {
		panic("COMMIT STATE ERR: " + err.Error())
	}
	if err := statedb.Database().TrieDB().Commit(gotRoot, false); err != nil {
		panic("COMMIT STATE TRIE ERR: " + err.Error())
	}

	return ret, vmenv.state.Logs(), vmenv.Gas, failed, err
}
