// +build !sputnikvm

package core

import (
	"github.com/ethereumproject/go-ethereum/core/state"
	"github.com/ethereumproject/go-ethereum/core/types"
	"github.com/ethereumproject/go-ethereum/params"
)

const SputnikVMExists = false

var UseSputnikVM = false

func ApplyMultiVmTransaction(config *params.ChainConfig, bc *BlockChain, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, totalUsedGas *uint64) (*types.Receipt, []*types.Log, uint64, error) {
	panic("not implemented")
}
