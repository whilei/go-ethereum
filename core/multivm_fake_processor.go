// +build !sputnikvm

package core

import (
	"math/big"

	"github.com/ethereumproject/go-ethereum/core/state"
	"github.com/ethereumproject/go-ethereum/core/types"
)

const SputnikVMExists = false

var UseSputnikVM = false

func ApplyMultiVmTransaction(config *params.ChainConfig, bc *BlockChain, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, totalUsedGas *big.Int) (*types.Receipt, []*types.Log, *big.Int, error) {
	panic("not implemented")
}
