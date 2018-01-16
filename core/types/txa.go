package types

import (
	"github.com/ethereumproject/go-ethereum/common"
)

type TxHashList []common.Hash

func Has(list TxHashList, hash common.Hash) bool {
	for _, h := range list {
		if h == hash {
			return true
		}
	}
	return false
}