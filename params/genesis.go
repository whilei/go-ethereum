package params

import (
	"math/big"

	"github.com/ethereumproject/go-ethereum/common"
)

// Genesis is the geth JSON format.
// https://github.com/ethereumproject/wiki/wiki/Ethereum-Chain-Spec-Format#subformat-genesis
type Genesis struct {
	Nonce      uint64         `json:"nonce"`
	Timestamp  uint64         `json:"timestamp"`
	ExtraData  []byte         `json:"extraData"`
	GasLimit   uint64         `json:"gasLimit"`
	Difficulty *big.Int       `json:"difficulty"`
	Mixhash    common.Hash    `json:"mixhash"`
	Coinbase   common.Address `json:"coinbase"`

	// Alloc maps accounts by their address.
	Alloc GenesisAlloc `json:"alloc"`
	// Alloc file contains CSV representation of Alloc
	AllocFile string `json:"alloc_file"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	ParentHash common.Hash `json:"parentHash"`
	Number     uint64
	GasUsed    uint64
}

type GenesisAlloc map[common.Address]GenesisAccount

// GenesisDumpAlloc is a Genesis.Alloc entry.
// type GenesisDumpAlloc struct {
// 	Code    PrefixedHex `json:"-"` // skip field for json encode
// 	Storage map[Hex]Hex `json:"-"`
// 	Balance string      `json:"balance"` // decimal string
// }
//
// type GenesisAccount struct {
// 	Address common.Address `json:"address"`
// 	Balance *big.Int       `json:"balance"`
// }

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Code       []byte                      `json:"code,omitempty"`
	Storage    map[common.Hash]common.Hash `json:"storage,omitempty"`
	Balance    *big.Int                    `json:"balance" gencodec:"required"`
	Nonce      uint64                      `json:"nonce,omitempty"`
	PrivateKey []byte                      `json:"secretKey,omitempty"` // for tests
}
