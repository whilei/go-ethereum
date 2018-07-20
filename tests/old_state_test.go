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
	"math/big"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkStateCall1024(b *testing.B) {
	fn := filepath.Join(oldStateTestDir, "stCallCreateCallCodeTestState.json")
	if err := BenchVmTest(fn, bconf{"Call1024BalanceTooLow", true, os.Getenv("JITVM") == "true"}, b); err != nil {
		b.Error(err)
	}
}

func TestStateStateSystemOperations(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stSystemOperationsTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateStateExample(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stExample.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateStatePreCompiledContracts(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stPreCompiledContracts.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateStateRecursiveCreate(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stRecursiveCreate.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateStateSpecial(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stSpecialTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateStateRefund(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stRefundTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateStateBlockHash(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stBlockHashTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateStateInitCode(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stInitCodeTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateStateLog(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stLogTestStates.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateStateTransaction(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stTransactionTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateStateTransition(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stTransitionTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateCallCreateCallCode(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stCallCreateCallCodeTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateCallCodes(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stCallCodes.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateDelegateCall(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stDelegatecallTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateMemory(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stMemoryTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateMemoryStress(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	if os.Getenv("TEST_VM_COMPLEX") == "" {
		t.Skip()
	}
	fn := filepath.Join(oldStateTestDir, "stMemoryStressTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateQuadraticComplexity(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	if os.Getenv("TEST_VM_COMPLEX") == "" {
		t.Skip()
	}
	fn := filepath.Join(oldStateTestDir, "stQuadraticComplexityTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateSolidity(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stSolidityTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateWallet(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fn := filepath.Join(oldStateTestDir, "stWalletTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateStateTestStatesRandom(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: big.NewInt(1000000),
	}

	fns, _ := filepath.Glob("./files/StateTestStates/RandomTestStates/*")
	for _, fn := range fns {
		if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
			t.Error(err)
		}
	}
}

// homestead tests
func TestStateHomesteadStateSystemOperations(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stSystemOperationsTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadStatePreCompiledContracts(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stPreCompiledContracts.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadStateRecursiveCreate(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stSpecialTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadStateRefund(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stRefundTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadStateInitCode(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stInitCodeTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadStateLog(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stLogTestStates.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadStateTransaction(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stTransactionTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadCallCreateCallCode(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stCallCreateCallCodeTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadCallCodes(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stCallCodes.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadMemory(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stMemoryTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadMemoryStress(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	if os.Getenv("TEST_VM_COMPLEX") == "" {
		t.Skip()
	}
	fn := filepath.Join(oldStateTestDir, "Homestead", "stMemoryStressTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadQuadraticComplexity(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	if os.Getenv("TEST_VM_COMPLEX") == "" {
		t.Skip()
	}
	fn := filepath.Join(oldStateTestDir, "Homestead", "stQuadraticComplexityTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadWallet(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stWalletTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadDelegateCodes(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stCallDelegateCodes.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateHomesteadDelegateCodesCallCode(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock: new(big.Int),
	}

	fn := filepath.Join(oldStateTestDir, "Homestead", "stCallDelegateCodesCallCode.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

// EIP150 tests
func TestStateEIP150Specific(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "stEIPSpecificTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150SingleCodeGasPrice(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "stEIPSingleCodeGasPrices.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150MemExpandingCalls(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "stMemExpandingEIPCalls.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadStateSystemOperations(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stSystemOperationsTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadStatePreCompiledContracts(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(1457000),
		DiehardBlock:             big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stPreCompiledContracts.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadStateRecursiveCreate(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stSpecialTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadStateRefund(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(1457000),
		DiehardBlock:             big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stRefundTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadStateInitCode(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stInitCodeTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadStateLog(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stLogTestStates.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadStateTransaction(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stTransactionTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadCallCreateCallCode(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stCallCreateCallCodeTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadCallCodes(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stCallCodes.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadMemory(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stMemoryTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadMemoryStress(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	if os.Getenv("TEST_VM_COMPLEX") == "" {
		t.Skip()
	}
	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stMemoryStressTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadQuadraticComplexity(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	if os.Getenv("TEST_VM_COMPLEX") == "" {
		t.Skip()
	}
	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stQuadraticComplexityTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadWallet(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(1457000),
		DiehardBlock:             big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stWalletTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadDelegateCodes(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stCallDelegateCodes.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadDelegateCodesCallCode(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stCallDelegateCodesCallCode.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}

func TestStateEIP150HomesteadBounds(t *testing.T) {
	ruleSet := RuleSet{
		HomesteadBlock:           new(big.Int),
		HomesteadGasRepriceBlock: big.NewInt(2457000),
	}

	fn := filepath.Join(oldStateTestDir, "EIP150", "Homestead", "stBoundsTestState.json")
	if err := RunStateTest(ruleSet, fn, StateSkipTests); err != nil {
		t.Error(err)
	}
}
