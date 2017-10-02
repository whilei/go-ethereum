package classic

import (
	"fmt"
	"math/big"

	"github.com/ethereumproject/go-ethereum/core/state"
	"github.com/ethereumproject/go-ethereum/core/vm"
	"github.com/ethereumproject/go-ethereum/common"
	"github.com/ethereumproject/go-ethereum/ethdb"
)

type command struct {
	code uint8
	value interface{}
}

type account struct {
	address common.Address
	nonce   uint64
	balance *big.Int
}

const (
	cmdStepID = iota
	cmdFireID
	cmdAccountID
	cmdHashID
	cmdCodeID
	cmdRuleID
)

var (
	cmdStep = &command{cmdStepID, nil}
	cmdFire = &command{cmdFireID, nil}
)

type vmHash struct {
	number uint64
	hash   common.Hash
}

type vmAccount struct {
	onModify func()
	acc state.AccountObject
}

func (vacc *vmAccount) SubBalance(amount *big.Int) {
	vacc.acc.SubBalance(amount)
	vacc.onModify()
}

func (vacc *vmAccount) AddBalance(amount *big.Int) {
	vacc.acc.AddBalance(amount)
	vacc.onModify()
}

func (vacc *vmAccount) SetBalance(amount *big.Int) {
	vacc.acc.SetBalance(amount)
	vacc.onModify()
}

func (vacc *vmAccount) SetNonce(nonce uint64) {
	vacc.acc.SetNonce(nonce)
	vacc.onModify()
}

func (vacc *vmAccount) ReturnGas(gas, price *big.Int) {
	vacc.acc.ReturnGas(gas, price)
	vacc.onModify()
}

func (vacc *vmAccount) SetCode(hash common.Hash, code []byte) {
	vacc.acc.SetCode(hash,code)
	vacc.onModify()
}

func (vacc *vmAccount) Nonce() uint64                                       { return vacc.acc.Nonce() }
func (vacc *vmAccount) Balance() *big.Int                                   { return vacc.acc.Balance() }
func (vacc *vmAccount) Address() common.Address                             { return vacc.acc.Address() }
func (vacc *vmAccount) ForEachStorage(cb func(key, value common.Hash) bool) { panic("unimplemented")}

type vmCode struct {
	address common.Address
	code []byte
	size int
	hash common.Hash
}

func (vmc *vmCode) Address() common.Address {
	return vmc.address
}

func (vmc *vmCode) Code() []byte {
	return vmc.code
}

func (vmc *vmCode) Hash() common.Hash {
	return vmc.hash
}

func (vmc *vmCode) Size() int {
	return vmc.size
}

type vmRule struct {
	table 		*vm.GasTable
	fork  		vm.Fork
	difficulty 	*big.Int
	gasLimit 	*big.Int
	time		*big.Int
}

type machine struct {}

type context struct {
	env    vmEnv
}

type ChangeLevel byte
const (
	None ChangeLevel = iota
	Absent
	Committed
	Modified
	Removed
)

type vmEnv struct {
	cmdc    	chan *command
	rqc			chan *vm.Require
	rules   	*vmRule
	evm     	*EVM
	depth   	int
	contract 	*Contract
	output      []byte
	stepByStep	bool

	db          *state.StateDB
	origin      common.Address
	coinbase    common.Address
	address     common.Address
	account     map[common.Address]ChangeLevel
	code        map[common.Address]ChangeLevel
	hash        map[uint64]*vmHash
	blockNumber uint64

	status      vm.Status
	err 		error
}

func NewMachine() vm.Machine {
	return &machine{}
}

func mkEnv(env *vmEnv, number uint64) *vmEnv {
	env.rqc = make(chan *vm.Require)
	env.cmdc = make(chan *command)
	env.blockNumber = number
	db, _ := ethdb.NewMemDatabase()
	env.db, _ = state.New(common.Hash{}, db)
	env.account = make(map[common.Address]ChangeLevel)
	env.code = make(map[common.Address]ChangeLevel)
	env.hash = make(map[uint64]*vmHash)
	env.status = vm.Inactive
	return env
}

func (m *machine) Name() string {
	return "CLASSIC VM"
}

func (m *machine) Type() vm.Type {
	return vm.ClassicVm
}

func (m *machine) Call(blockNumber uint64, caller common.Address, to common.Address, data []byte, gas, price, value *big.Int) (vm.Context, error) {
	ctx := &context{}

	go func (e *vmEnv) {
		cmd, ok := <-e.cmdc
		if !ok {
			close(e.rqc)
			return
		}
		if cmd.code == cmdStepID || cmd.code == cmdFireID {
			e.status = vm.Running
			e.stepByStep = cmd.code == cmdStepID
			callerRef := e.queryAccount(caller)
			e.address = to
			e.evm = NewVM(e)
			e.output, e.err = e.Call(callerRef,to,data,gas,price,value)
			e.rqc <- nil
			close(e.rqc)
			return
		} else {
			e.handleCmdc(cmd)
		}
	}(mkEnv(&ctx.env,blockNumber))

	return ctx, nil
}

func (m *machine) Create(blockNumber uint64, caller common.Address, code []byte, gas, price, value *big.Int) (vm.Context, error) {
	ctx := &context{}

	go func (e *vmEnv) {
		cmd, ok := <-e.cmdc
		if !ok {
			return
		}
		if cmd.code == cmdStepID || cmd.code == cmdFireID {
			e.status = vm.Running
			e.stepByStep = cmd.code == cmdStepID
			callerRef := e.queryAccount(caller)
			e.evm = NewVM(e)
			e.output, e.address, e.err = e.Create(callerRef,code,gas,price,value)
			e.rqc <- nil
			return
		} else {
			e.handleCmdc(cmd)
		}
	}(mkEnv(&ctx.env,blockNumber))

	return ctx, nil
}

func (vmenv *vmEnv) handleCmdc(cmd *command) bool {
	fmt.Println("handleCmd", *cmd)
	switch cmd.code {
	case cmdHashID:
		h := cmd.value.(*vmHash)
		vmenv.hash[h.number] = h
	case cmdCodeID:
		c := cmd.value.(*vmCode)
		fmt.Println("handleCmd", *c)
		if c.code == nil {
			vmenv.code[c.address] = Absent
		} else {
			var acc state.AccountObject
			if !vmenv.db.Exist(c.address) {
				acc = vmenv.db.CreateAccount(c.address)
			} else {
				acc = vmenv.db.GetAccount(c.address)
			}
			acc.SetCode(c.hash, c.code)
			vmenv.code[c.address] = Committed
		}
	case cmdAccountID:
		a := cmd.value.(*account)
		if a.balance == nil {
			vmenv.account[a.address] = Absent
		} else {
			var acc state.AccountObject
			if !vmenv.db.Exist(a.address) {
				acc = vmenv.db.CreateAccount(a.address)
			} else {
				acc = vmenv.db.GetAccount(a.address)
			}
			acc.SetNonce(a.nonce)
			acc.SetBalance(a.balance)
			vmenv.account[a.address] = Committed
		}
	case cmdRuleID:
		vmenv.rules = cmd.value.(*vmRule)
	case cmdStepID:
		vmenv.stepByStep = true
		return true
	case cmdFireID:
		vmenv.stepByStep = false
		return true
	default:
		panic("invalid command")
	}
	return false
}

func (vmenv *vmEnv) handleRequire(rq *vm.Require) {
	fmt.Println("handleRequire", *rq)
	vmenv.status = vm.RequireErr
	vmenv.rqc <- rq
	for {
		cmd := <-vmenv.cmdc
		if vmenv.handleCmdc(cmd) {
			vmenv.status = vm.Running
			return
		}
	}
}

func (c *context) Address() (common.Address, error) {
	return c.env.address, nil
}

func (c *context) CommitRules(table *vm.GasTable, fork vm.Fork, difficulty, gasLimit, time *big.Int) (err error) {
	rule := vmRule{table,fork, difficulty, gasLimit, time}
	c.env.cmdc <- &command{cmdRuleID, &rule}
	return
}

func (c *context) CommitAccount(address common.Address, nonce uint64, balance *big.Int) (err error) {
	account := account{address:address,nonce:nonce,balance:balance}
	c.env.cmdc <- &command{cmdAccountID, &account}
	return
}

func (c *context) CommitBlockHash(number uint64, hash common.Hash) (err error) {
	value := vmHash{number, hash}
	c.env.cmdc <- &command{cmdHashID, &value}
	return
}

func (c *context) CommitCode(address common.Address, hash common.Hash, code []byte) (err error) {
	value := vmCode{address, code, len(code), hash}
	c.env.cmdc <- &command{cmdCodeID, &value}
	return
}

func (c *context) Status() vm.Status {
	return c.env.status
}

func (c *context) Finish() (err error) {
	return nil
}

func (c *context) Fire() (*vm.Require) {
	c.env.cmdc <- cmdFire
	rq := <-c.env.rqc
	return rq
}

func (c *context) Code(addr common.Address) (common.Hash, []byte, error) {
	if c.env.account[addr] == Modified {
		return c.env.db.GetCodeHash(addr), c.env.db.GetCode(addr), nil
	} else {
		return common.Hash{}, nil, nil
	}
}

func (c *context) Modified() (accounts []vm.Account, err error) {
	accounts = []vm.Account{}
	for addr, level := range c.env.account {
		if level == Modified {
			acc := c.env.db.GetAccount(addr)
			accounts = append(accounts, acc.(vm.Account))
		}
	}
	return
}

func (c *context) Removed() (addresses []common.Address, err error) {
	addresses = []common.Address{}
	for addr, level := range c.env.account {
		if level == Removed {
			addresses = append(addresses, addr)
		}
	}
	return
}

func (c *context) Committed() (addresses []common.Address, err error) {
	addresses = []common.Address{}
	for addr, level := range c.env.account {
		if level == Committed {
			addresses = append(addresses, addr)
		}
	}
	return
}

func (c *context) Out() ([]byte,error) {
	return c.env.output, c.env.err
}

// TODO: either able to return an error or remove
func (c *context) Logs() (logs state.Logs, err error) {
	logs = c.env.db.Logs()
	return
}

func (c *context) Err() error {
	return c.env.err
}

func (vmenv *vmEnv) queryRule() *vmRule {
	for {
		if vmenv.rules != nil {
			return vmenv.rules
		} else {
			vmenv.handleRequire(&vm.Require{ID:vm.RequireRules,Number: vmenv.blockNumber})
		}
	}
}

func (vmenv *vmEnv) GasTable(block *big.Int) *vm.GasTable{
	number := block.Uint64()
	if number != vmenv.blockNumber {
		vmenv.status = vm.Broken
		panic("invalid block number")
	}
	return vmenv.queryRule().table
}

func (vmenv *vmEnv) IsHomestead(block *big.Int) bool{
	number := block.Uint64()
	if number != vmenv.blockNumber {
		vmenv.status = vm.Broken
		panic("invalid block number")
	}
	return vmenv.queryRule().fork >= vm.Homestead
}

func (vmenv *vmEnv) RuleSet() vm.RuleSet {
	return vmenv
}

func (vmenv *vmEnv) Origin() common.Address {
	return vmenv.origin
}

func (vmenv *vmEnv) BlockNumber() *big.Int {
	return new(big.Int).SetUint64(vmenv.blockNumber)
}

func (vmenv *vmEnv) Coinbase() common.Address {
	return vmenv.coinbase
}

func (vmenv *vmEnv) Time() *big.Int {
	return vmenv.queryRule().time
}

func (vmenv *vmEnv) Difficulty() *big.Int {
	return vmenv.queryRule().difficulty
}

func (vmenv *vmEnv) GasLimit() *big.Int {
	return vmenv.queryRule().gasLimit
}

func (vmenv *vmEnv) Db() Database {
	// wrap database activity
	return vmenv
}

func (vmenv *vmEnv) Depth() int {
	return vmenv.depth
}

func (vmenv *vmEnv) SetDepth(i int) {
	vmenv.depth = i
}

func (vmenv *vmEnv) GetHash(n uint64) common.Hash {
	for {
		if h, exists := vmenv.hash[n]; exists {
			return h.hash
		} else {
			vmenv.handleRequire(&vm.Require{ID:vm.RequireHash, Number:n})
		}
	}
}

func (vmenv *vmEnv) AddLog(log *state.Log) {
	vmenv.db.AddLog(log)
}

func (vmenv *vmEnv) CanTransfer(from common.Address, balance *big.Int) bool {
	return vmenv.GetBalance(from).Cmp(balance) >= 0
}

func (vmenv *vmEnv) SnapshotDatabase() int {
	return vmenv.db.Snapshot()
}

func (vmenv *vmEnv) RevertToSnapshot(snapshot int) {
	vmenv.db.RevertToSnapshot(snapshot)
}

func (vmenv *vmEnv) Transfer(from, to state.AccountObject, amount *big.Int) {
	Transfer(from, to, amount)
}

func (vmenv *vmEnv) Call(me ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return Call(vmenv, me, addr, data, gas, price, value)
}

func (vmenv *vmEnv) CallCode(me ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return CallCode(vmenv, me, addr, data, gas, price, value)
}

func (vmenv *vmEnv) DelegateCall(me ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error) {
	return DelegateCall(vmenv, me.(*Contract), addr, data, gas, price)
}

func (vmenv *vmEnv) Create(me ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error) {
	return Create(vmenv, me, data, gas, price, value)
}

func (vmenv *vmEnv) Run(contract *Contract, input []byte) (ret []byte, err error) {
	return vmenv.evm.Run(contract,input)
}

func (vmenv *vmEnv) GetAccount(addr common.Address) state.AccountObject {
	return vmenv.queryAccount(addr)
}

func (vmenv *vmEnv) CreateAccount(addr common.Address) state.AccountObject {
	vmenv.account[addr] = Modified
	return vmenv.db.CreateAccount(addr)
}

func (vmenv *vmEnv) queryCode(addr common.Address) {
	for {
		if vmenv.code[addr] == None {
			vmenv.handleRequire(&vm.Require{ID:vm.RequireCode,Address:addr})
		}
	}
}

func (vmenv *vmEnv) queryAccount(addr common.Address) state.AccountObject {
	for {
		switch vmenv.account[addr] {
		case None:
			vmenv.handleRequire(&vm.Require{ID:vm.RequireAccount,Address:addr})
		case Committed, Modified:
			return &vmAccount{ func(){ vmenv.account[addr] = Modified }, vmenv.db.GetAccount(addr) }
		default:
			return nil
		}
	}
}

func (vmenv *vmEnv) AddBalance(addr common.Address, ammount *big.Int) {
	vmenv.queryAccount(addr).AddBalance(ammount)
}

func (vmenv *vmEnv) GetBalance(addr common.Address) *big.Int {
	return vmenv.queryAccount(addr).Balance()
}

func (vmenv *vmEnv) GetNonce(addr common.Address) uint64 {
	return vmenv.queryAccount(addr).Nonce()
}

func (vmenv *vmEnv) SetNonce(addr common.Address, nonce uint64) {
	vmenv.queryAccount(addr).SetNonce(nonce)
}

func (vmenv *vmEnv) GetCodeHash(addr common.Address) common.Hash {
	vmenv.queryCode(addr)
	return vmenv.db.GetCodeHash(addr)
}

func (vmenv *vmEnv) GetCodeSize(addr common.Address) int {
	vmenv.queryCode(addr)
	return vmenv.db.GetCodeSize(addr)
}

func (vmenv *vmEnv) GetCode(addr common.Address) []byte {
	vmenv.queryCode(addr)
	return vmenv.db.GetCode(addr)
}

func (vmenv *vmEnv) SetCode(addr common.Address, code []byte) {
	vmenv.code[addr] = Modified
	vmenv.db.SetCode(addr,code)
}

func (vmenv *vmEnv) AddRefund(gas *big.Int) {
	vmenv.db.AddRefund(gas)
}

func (vmenv *vmEnv) GetRefund() *big.Int {
	return vmenv.db.GetRefund()
}

func (vmenv *vmEnv) GetState(common.Address, common.Hash) common.Hash {
	panic("GetState is unimplemented")
	return common.Hash{}
}

func (vmenv *vmEnv) SetState(common.Address, common.Hash, common.Hash) {
	panic("SetState is unimplemented")
}

func (vmenv *vmEnv) Suicide(addr common.Address) bool {
	vmenv.account[addr] = Removed
	return vmenv.db.Suicide(addr)
}

func (vmenv *vmEnv) HasSuicided(addr common.Address) bool {
	return vmenv.db.HasSuicided(addr)
}

func (vmenv *vmEnv) Exist(addr common.Address) bool {
	vmenv.queryAccount(addr)
	return vmenv.db.Exist(addr)
}
