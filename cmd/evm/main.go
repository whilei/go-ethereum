// Copyright 2014 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// evm executes EVM code snippets.
package main

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/ETCDEVTeam/sputnikvm-ffi/go/sputnikvm"
	"github.com/ethereumproject/go-ethereum/common"
	"github.com/ethereumproject/go-ethereum/core"
	"github.com/ethereumproject/go-ethereum/core/state"
	"github.com/ethereumproject/go-ethereum/core/types"
	"github.com/ethereumproject/go-ethereum/core/vm"
	"github.com/ethereumproject/go-ethereum/crypto"
	"github.com/ethereumproject/go-ethereum/ethdb"
	"github.com/ethereumproject/go-ethereum/logger/glog"
)

// Version is the application revision identifier. It can be set with the linker
// as in: go build -ldflags "-X main.Version="`git describe --tags`
var Version = "unknown"

var (
	DebugFlag = cli.BoolFlag{
		Name:  "debug",
		Usage: "output full trace logs",
	}
	ForceJitFlag = cli.BoolFlag{
		Name:  "forcejit",
		Usage: "forces jit compilation",
	}
	DisableJitFlag = cli.BoolFlag{
		Name:  "nojit",
		Usage: "disabled jit compilation",
	}
	CodeFlag = cli.StringFlag{
		Name:  "code",
		Usage: "EVM code",
	}
	GasFlag = cli.StringFlag{
		Name:  "gas",
		Usage: "gas limit for the evm",
		Value: "10000000000",
	}
	PriceFlag = cli.StringFlag{
		Name:  "price",
		Usage: "price set for the evm",
		Value: "0",
	}
	ValueFlag = cli.StringFlag{
		Name:  "value",
		Usage: "value set for the evm",
		Value: "0",
	}
	DumpFlag = cli.BoolFlag{
		Name:  "dump",
		Usage: "dumps the state after the run",
	}
	InputFlag = cli.StringFlag{
		Name:  "input",
		Usage: "input for the EVM",
	}
	SysStatFlag = cli.BoolFlag{
		Name:  "sysstat",
		Usage: "display system stats",
	}
	VerbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Usage: "sets the verbosity level",
	}
	CreateFlag = cli.BoolFlag{
		Name:  "create",
		Usage: "indicates the action should be create rather than call",
	}
	SVMFlag = cli.BoolFlag{
		Name:  "sputnikvm",
		Usage: "use SputnikEVM instead of standard EVM",
	}
	EVM2Flag = cli.BoolFlag{
		Name:  "evm2",
		Usage: "use ApplyTransaction fn instead of handrolled evm Call/Create",
	}
)

var app *cli.App

func init() {
	app = cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Version = Version
	app.Usage = "the evm command line interface"
	app.Action = run
	app.Flags = []cli.Flag{
		CreateFlag,
		DebugFlag,
		VerbosityFlag,
		ForceJitFlag,
		DisableJitFlag,
		SysStatFlag,
		CodeFlag,
		GasFlag,
		PriceFlag,
		ValueFlag,
		DumpFlag,
		InputFlag,
		SVMFlag,
		EVM2Flag,
	}
}

// callmsg is the message type used for call transactions.
type callmsg struct {
	from          *state.StateObject
	to            *common.Address
	gas, gasPrice *big.Int
	value         *big.Int
	data          []byte
}

// accessor boilerplate to implement core.Message
func (m callmsg) From() (common.Address, error)         { return m.from.Address(), nil }
func (m callmsg) FromFrontier() (common.Address, error) { return m.from.Address(), nil }
func (m callmsg) Nonce() uint64                         { return m.from.Nonce() }
func (m callmsg) To() *common.Address                   { return m.to }
func (m callmsg) GasPrice() *big.Int                    { return m.gasPrice }
func (m callmsg) Gas() *big.Int                         { return m.gas }
func (m callmsg) Value() *big.Int                       { return m.value }
func (m callmsg) Data() []byte                          { return m.data }

func runevm2(ctx *cli.Context) error {
	db, _ := ethdb.NewMemDatabase()
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db))
	sender := statedb.CreateAccount(common.StringToAddress("sender"))

	valueFlag, _ := new(big.Int).SetString(ctx.GlobalString(ValueFlag.Name), 0)
	if valueFlag == nil {
		log.Fatalf("malformed %s flag value %q", ValueFlag.Name, ctx.GlobalString(ValueFlag.Name))
	}
	vmenv := NewEnv(statedb, common.StringToAddress("evmuser"), valueFlag)
	gasFlag, _ := new(big.Int).SetString(ctx.GlobalString(GasFlag.Name), 0)
	if gasFlag == nil {
		log.Fatalf("malformed %s flag value %q", GasFlag.Name, ctx.GlobalString(GasFlag.Name))
	}
	priceFlag, _ := new(big.Int).SetString(ctx.GlobalString(PriceFlag.Name), 0)
	if priceFlag == nil {
		log.Fatalf("malformed %s flag value %q", PriceFlag.Name, ctx.GlobalString(PriceFlag.Name))
	}

	tstart := time.Now()

	statedb.SetBalance(sender.Address(), common.MaxBig)

	msg := callmsg{
		from:     statedb.GetOrNewStateObject(sender.Address()),
		to:       nil,
		gas:      gasFlag,
		gasPrice: priceFlag,
		value:    valueFlag,
		data:     append(common.Hex2Bytes(ctx.GlobalString(CodeFlag.Name)), common.Hex2Bytes(ctx.GlobalString(InputFlag.Name))...),
	}
	if msg.gas.Sign() == 0 {
		msg.gas = big.NewInt(50000000)
	}
	if msg.gasPrice.Sign() == 0 {
		msg.gasPrice = new(big.Int).Mul(big.NewInt(50), common.Shannon)
	}

	gp := new(core.GasPool).AddGas(common.MaxBig)
	ret, _, _, err := core.ApplyMessage(vmenv, msg, gp)

	vmdone := time.Since(tstart)

	if ctx.GlobalBool(DumpFlag.Name) {
		statedb.CommitTo(db, false)
		fmt.Println(string(statedb.Dump([]common.Address{})))
	}

	if ctx.GlobalBool(SysStatFlag.Name) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		fmt.Printf("vm took %v\n", vmdone)
		fmt.Printf(`alloc:      %d
tot alloc:  %d
no. malloc: %d
heap alloc: %d
heap objs:  %d
num gc:     %d
`, mem.Alloc, mem.TotalAlloc, mem.Mallocs, mem.HeapAlloc, mem.HeapObjects, mem.NumGC)
	}

	fmt.Printf("OUT: 0x%x", ret)
	if err != nil {
		fmt.Printf(" error: %v", err)
	}
	fmt.Println()
	return nil

}

func runevm(ctx *cli.Context) error {
	db, _ := ethdb.NewMemDatabase()
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db))
	sender := statedb.CreateAccount(common.StringToAddress("sender"))

	valueFlag, _ := new(big.Int).SetString(ctx.GlobalString(ValueFlag.Name), 0)
	if valueFlag == nil {
		log.Fatalf("malformed %s flag value %q", ValueFlag.Name, ctx.GlobalString(ValueFlag.Name))
	}
	vmenv := NewEnv(statedb, common.StringToAddress("evmuser"), valueFlag)

	tstart := time.Now()

	var (
		ret []byte
		err error
	)

	gasFlag, _ := new(big.Int).SetString(ctx.GlobalString(GasFlag.Name), 0)
	if gasFlag == nil {
		log.Fatalf("malformed %s flag value %q", GasFlag.Name, ctx.GlobalString(GasFlag.Name))
	}
	priceFlag, _ := new(big.Int).SetString(ctx.GlobalString(PriceFlag.Name), 0)
	if priceFlag == nil {
		log.Fatalf("malformed %s flag value %q", PriceFlag.Name, ctx.GlobalString(PriceFlag.Name))
	}

	if ctx.GlobalBool(CreateFlag.Name) {
		input := append(common.Hex2Bytes(ctx.GlobalString(CodeFlag.Name)), common.Hex2Bytes(ctx.GlobalString(InputFlag.Name))...)
		ret, _, err = vmenv.Create(sender, input, gasFlag, priceFlag, valueFlag)
	} else {
		receiver := statedb.CreateAccount(common.StringToAddress("receiver"))

		code := common.Hex2Bytes(ctx.GlobalString(CodeFlag.Name))
		receiver.SetCode(crypto.Keccak256Hash(code), code)
		ret, err = vmenv.Call(sender, receiver.Address(), common.Hex2Bytes(ctx.GlobalString(InputFlag.Name)), gasFlag, priceFlag, valueFlag)
	}
	vmdone := time.Since(tstart)

	if ctx.GlobalBool(DumpFlag.Name) {
		statedb.CommitTo(db, false)
		fmt.Println(string(statedb.Dump([]common.Address{})))
	}

	if ctx.GlobalBool(SysStatFlag.Name) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		fmt.Printf("vm took %v\n", vmdone)
		fmt.Printf(`alloc:      %d
tot alloc:  %d
no. malloc: %d
heap alloc: %d
heap objs:  %d
num gc:     %d
`, mem.Alloc, mem.TotalAlloc, mem.Mallocs, mem.HeapAlloc, mem.HeapObjects, mem.NumGC)
	}

	fmt.Printf("OUT: 0x%x", ret)
	if err != nil {
		fmt.Printf(" error: %v", err)
	}
	fmt.Println()
	return nil

}

func runsvm(ctx *cli.Context) error {

	db, _ := ethdb.NewMemDatabase()
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db))
	sender := statedb.CreateAccount(common.StringToAddress("sender"))

	statedb.SetBalance(sender.Address(), common.MaxBig)

	valueFlag, _ := new(big.Int).SetString(ctx.GlobalString(ValueFlag.Name), 0)
	if valueFlag == nil {
		log.Fatalf("malformed %s flag value %q", ValueFlag.Name, ctx.GlobalString(ValueFlag.Name))
	}
	vmenv := NewEnv(statedb, common.StringToAddress("evmuser"), valueFlag)
	gasFlag, _ := new(big.Int).SetString(ctx.GlobalString(GasFlag.Name), 0)
	if gasFlag == nil {
		log.Fatalf("malformed %s flag value %q", GasFlag.Name, ctx.GlobalString(GasFlag.Name))
	}
	priceFlag, _ := new(big.Int).SetString(ctx.GlobalString(PriceFlag.Name), 0)
	if priceFlag == nil {
		log.Fatalf("malformed %s flag value %q", PriceFlag.Name, ctx.GlobalString(PriceFlag.Name))
	}

	tstart := time.Now()

	vmtx := sputnikvm.Transaction{
		Caller:   sender.Address(),
		GasPrice: gasFlag,
		GasLimit: priceFlag,
		Address:  nil,
		Value:    valueFlag,
		Input:    append(common.Hex2Bytes(ctx.GlobalString(CodeFlag.Name)), common.Hex2Bytes(ctx.GlobalString(InputFlag.Name))...),
		Nonce:    new(big.Int).SetUint64(statedb.GetOrNewStateObject(sender.Address()).Nonce()),
	}
	vmheader := sputnikvm.HeaderParams{
		Beneficiary: vmenv.Coinbase(),
		Timestamp:   vmenv.Time().Uint64(),
		Number:      vmenv.BlockNumber(),
		Difficulty:  vmenv.Difficulty(),
		GasLimit:    vmenv.GasLimit(),
	}

	if vmtx.GasLimit.Sign() == 0 {
		vmtx.GasLimit = big.NewInt(50000000)
	}
	if vmtx.GasPrice.Sign() == 0 {
		vmtx.GasPrice = new(big.Int).Mul(big.NewInt(50), common.Shannon)
	}

	// always using latest configuration for now.
	var vm *sputnikvm.VM
	if state.StartingNonce == 0 {
		vm = sputnikvm.NewEIP160(&vmtx, &vmheader)
	} else if state.StartingNonce == 1048576 {
		vm = sputnikvm.NewMordenEIP160(&vmtx, &vmheader)
	} else {
		sputnikvm.SetCustomInitialNonce(big.NewInt(int64(state.StartingNonce)))
		vm = sputnikvm.NewCustomEIP160(&vmtx, &vmheader)
	}

Loop:
	for {
		ret := vm.Fire()
		switch ret.Typ() {
		case sputnikvm.RequireNone:
			break Loop
		case sputnikvm.RequireAccount:
			address := ret.Address()
			if statedb.Exist(address) {
				vm.CommitAccount(address, new(big.Int).SetUint64(statedb.GetNonce(address)),
					statedb.GetBalance(address), statedb.GetCode(address))
				break
			}
			vm.CommitNonexist(address)
		case sputnikvm.RequireAccountCode:
			address := ret.Address()
			if statedb.Exist(address) {
				vm.CommitAccountCode(address, statedb.GetCode(address))
				break
			}
			vm.CommitNonexist(address)
		case sputnikvm.RequireAccountStorage:
			address := ret.Address()
			key := common.BigToHash(ret.StorageKey())
			if statedb.Exist(address) {
				value := statedb.GetState(address, key).Big()
				sKey := ret.StorageKey()
				vm.CommitAccountStorage(address, sKey, value)
				break
			}
			vm.CommitNonexist(address)
		case sputnikvm.RequireBlockhash:
			number := ret.BlockNumber()
			hash := common.Hash{}
			// if block := bc.GetBlockByNumber(number.Uint64()); block != nil && block.Number().Cmp(number) == 0 {
			// 	hash = block.Hash()
			// }
			vm.CommitBlockhash(number, hash)
		}
	}

	// VM execution is finished at this point. We apply changes to the statedb.

	for _, account := range vm.AccountChanges() {
		switch account.Typ() {
		case sputnikvm.AccountChangeIncreaseBalance:
			address := account.Address()
			amount := account.ChangedAmount()
			statedb.AddBalance(address, amount)
		case sputnikvm.AccountChangeDecreaseBalance:
			address := account.Address()
			amount := account.ChangedAmount()
			balance := new(big.Int).Sub(statedb.GetBalance(address), amount)
			statedb.SetBalance(address, balance)
		case sputnikvm.AccountChangeRemoved:
			address := account.Address()
			statedb.Suicide(address)
		case sputnikvm.AccountChangeFull:
			address := account.Address()
			code := account.Code()
			nonce := account.Nonce()
			balance := account.Balance()
			statedb.SetBalance(address, balance)
			statedb.SetNonce(address, nonce.Uint64())
			statedb.SetCode(address, code)
			for _, item := range account.ChangedStorage() {
				statedb.SetState(address, common.BigToHash(item.Key), common.BigToHash(item.Value))
			}
		case sputnikvm.AccountChangeCreate:
			address := account.Address()
			code := account.Code()
			nonce := account.Nonce()
			balance := account.Balance()
			statedb.SetBalance(address, balance)
			statedb.SetNonce(address, nonce.Uint64())
			statedb.SetCode(address, code)
			for _, item := range account.Storage() {
				statedb.SetState(address, common.BigToHash(item.Key), common.BigToHash(item.Value))
			}
		default:
			panic("unreachable")
		}
	}

	vmdone := time.Since(tstart)

	for i, log := range vm.Logs() {
		fmt.Println("log", i, log.Address, log.Topics, log.Data)
		statelog := evm.NewLog(log.Address, log.Topics, log.Data, vmheader.Number.Uint64())
		statedb.AddLog(*statelog)
	}
	// for _, log := range vm.Logs() {
	// 	statelog := evm.NewLog(log.Address, log.Topics, log.Data, header.Number.Uint64())
	// 	statedb.AddLog(*statelog)
	// }
	// usedGas := vm.UsedGas()
	// totalUsedGas.Add(totalUsedGas, usedGas)
	fmt.Println("used gas: ", vm.UsedGas())
	fmt.Println("intermediate root: ", statedb.IntermediateRoot(false).Hex())
	fmt.Println("vm failed: ", vm.Failed())
	fmt.Println("took: ", time.Since(tstart))

	if ctx.GlobalBool(SysStatFlag.Name) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		fmt.Printf("vm took %v\n", vmdone)
		fmt.Printf(`alloc:      %d
tot alloc:  %d
no. malloc: %d
heap alloc: %d
heap objs:  %d
num gc:     %d
`, mem.Alloc, mem.TotalAlloc, mem.Mallocs, mem.HeapAlloc, mem.HeapObjects, mem.NumGC)
	}

	if ctx.GlobalBool(DumpFlag.Name) {
		fmt.Println("StateDB dump:")
		statedb.CommitTo(db, false)
		fmt.Println(string(statedb.Dump([]common.Address{})))
	}

	// receipt := types.NewReceipt(statedb.IntermediateRoot(false).Bytes(), totalUsedGas)
	// receipt.TxHash = tx.Hash()
	// receipt.GasUsed = new(big.Int).Set(totalUsedGas)
	// if vm.Failed() {
	// 	receipt.Status = types.TxFailure
	// } else {
	// 	receipt.Status = types.TxSuccess
	// }
	// if MessageCreatesContract(tx) {
	// 	receipt.ContractAddress = crypto.CreateAddress(from, tx.Nonce())
	// }

	// logs := statedb.GetLogs(tx.Hash())
	// receipt.Logs = logs
	// receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	// glog.V(logger.Debug).Infoln(receipt)

	vm.Free()
	// return ret, nil
	return nil
	// return receipt, logs, totalUsedGas, nil
}

func run(ctx *cli.Context) error {
	glog.SetToStderr(true)
	glog.SetV(ctx.GlobalInt(VerbosityFlag.Name))

	if ctx.Bool(SVMFlag.Name) {
		return runsvm(ctx)
	}
	if ctx.Bool(EVM2Flag.Name) {
		return runevm2(ctx)
	}
	return runevm(ctx)
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type VMEnv struct {
	state *state.StateDB
	block *types.Block

	transactor *common.Address
	value      *big.Int

	depth int
	Gas   *big.Int
	time  *big.Int

	evm *vm.EVM
}

func NewEnv(state *state.StateDB, transactor common.Address, value *big.Int) *VMEnv {
	env := &VMEnv{
		state:      state,
		transactor: &transactor,
		value:      value,
		time:       big.NewInt(time.Now().Unix()),
	}

	env.evm = vm.New(env)
	return env
}

// ruleSet implements vm.RuleSet and will always default to the homestead rule set.
type ruleSet struct{}

func (ruleSet) IsHomestead(*big.Int) bool { return true }

func (ruleSet) GasTable(*big.Int) *vm.GasTable {
	return &vm.GasTable{
		ExtcodeSize:     big.NewInt(700),
		ExtcodeCopy:     big.NewInt(700),
		Balance:         big.NewInt(400),
		SLoad:           big.NewInt(200),
		Calls:           big.NewInt(700),
		Suicide:         big.NewInt(5000),
		ExpByte:         big.NewInt(10),
		CreateBySuicide: big.NewInt(25000),
	}
}

func (self *VMEnv) RuleSet() vm.RuleSet       { return ruleSet{} }
func (self *VMEnv) Vm() vm.Vm                 { return self.evm }
func (self *VMEnv) Db() vm.Database           { return self.state }
func (self *VMEnv) SnapshotDatabase() int     { return self.state.Snapshot() }
func (self *VMEnv) RevertToSnapshot(snap int) { self.state.RevertToSnapshot(snap) }
func (self *VMEnv) Origin() common.Address    { return *self.transactor }
func (self *VMEnv) BlockNumber() *big.Int     { return new(big.Int) }
func (self *VMEnv) Coinbase() common.Address  { return *self.transactor }
func (self *VMEnv) Time() *big.Int            { return self.time }
func (self *VMEnv) Difficulty() *big.Int      { return common.Big1 }
func (self *VMEnv) BlockHash() []byte         { return make([]byte, 32) }
func (self *VMEnv) Value() *big.Int           { return self.value }
func (self *VMEnv) GasLimit() *big.Int        { return big.NewInt(1000000000) }
func (self *VMEnv) VmType() vm.Type           { return vm.StdVmTy }
func (self *VMEnv) Depth() int                { return 0 }
func (self *VMEnv) SetDepth(i int)            { self.depth = i }
func (self *VMEnv) GetHash(n uint64) common.Hash {
	if self.block.Number().Cmp(big.NewInt(int64(n))) == 0 {
		return self.block.Hash()
	}
	return common.Hash{}
}
func (self *VMEnv) AddLog(log *vm.Log) {
	self.state.AddLog(*log)
}
func (self *VMEnv) CanTransfer(from common.Address, balance *big.Int) bool {
	return self.state.GetBalance(from).Cmp(balance) >= 0
}
func (self *VMEnv) Transfer(from, to vm.Account, amount *big.Int) {
	core.Transfer(from, to, amount)
}

func (self *VMEnv) Call(caller vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	self.Gas = gas
	return core.Call(self, caller, addr, data, gas, price, value)
}

func (self *VMEnv) CallCode(caller vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return core.CallCode(self, caller, addr, data, gas, price, value)
}

func (self *VMEnv) DelegateCall(caller vm.ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error) {
	return core.DelegateCall(self, caller, addr, data, gas, price)
}

func (self *VMEnv) Create(caller vm.ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error) {
	return core.Create(self, caller, data, gas, price, value)
}
