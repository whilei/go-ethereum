package params

import (
	"bytes"
	enchex "encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/ethereumproject/go-ethereum/common/hexutil"
	"github.com/ethereumproject/go-ethereum/common/math"
	"github.com/ethereumproject/go-ethereum/rlp"
)

// Genesis is the geth JSON format.
// https://github.com/ethereumproject/wiki/wiki/Ethereum-Chain-Spec-Format#subformat-genesis
type Genesis struct {
	Config     *ChainConfig   `json:"-"`
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

func (ga *GenesisAlloc) UnmarshalJSON(data []byte) error {
	m := make(map[common.UnprefixedAddress]GenesisAccount)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*ga = make(GenesisAlloc)
	for addr, a := range m {
		(*ga)[common.Address(addr)] = a
	}
	return nil
}

// field type overrides for gencodec
type genesisSpecMarshaling struct {
	Nonce      math.HexOrDecimal64
	Timestamp  math.HexOrDecimal64
	ExtraData  hexutil.Bytes
	GasLimit   math.HexOrDecimal64
	GasUsed    math.HexOrDecimal64
	Number     math.HexOrDecimal64
	Difficulty *math.HexOrDecimal256
	Alloc      map[common.UnprefixedAddress]GenesisAccount
}

type genesisAccountMarshaling struct {
	Code       hexutil.Bytes
	Balance    *math.HexOrDecimal256
	Nonce      math.HexOrDecimal64
	Storage    map[storageJSON]storageJSON
	PrivateKey hexutil.Bytes
}

// storageJSON represents a 256 bit byte array, but allows less than 256 bits when
// unmarshaling from hex.
type storageJSON common.Hash

func (h *storageJSON) UnmarshalText(text []byte) error {
	text = bytes.TrimPrefix(text, []byte("0x"))
	if len(text) > 64 {
		return fmt.Errorf("too many hex characters in storage key/value %q", text)
	}
	offset := len(h) - len(text)/2 // pad on the left
	if _, err := enchex.Decode(h[offset:], text); err != nil {
		fmt.Println(err)
		return fmt.Errorf("invalid hex storage key/value %q", text)
	}
	return nil
}

func (h storageJSON) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

func decodePrealloc(data string) GenesisAlloc {
	var p []struct{ Addr, Balance *big.Int }
	if err := rlp.NewStream(strings.NewReader(data), 0).Decode(&p); err != nil {
		panic(err)
	}
	ga := make(GenesisAlloc, len(p))
	for _, account := range p {
		ga[common.BigToAddress(account.Addr)] = GenesisAccount{Balance: account.Balance}
	}
	return ga
}

// DefaultGenesisBlock returns the Ethereum main net genesis block.
func DefaultGenesisBlock() *Genesis {
	return DefaultConfigMainnet.Genesis
}

// DefaultTestnetGenesisBlock returns the Ropsten network genesis block.
func DefaultTestnetGenesisBlock() *Genesis {
	return DefaultConfigMorden.Genesis
}

func (g *Genesis) ConfigOrDefault(ghash common.Hash) *ChainConfig {
	switch {
	case g != nil:
		return g.Config
	case ghash == MainnetGenesisHash:
		return DefaultConfigMainnet.ChainConfig
	case ghash == TestnetGenesisHash:
		return DefaultConfigMorden.ChainConfig
	default:
		return AllEthashProtocolChanges
	}
}
