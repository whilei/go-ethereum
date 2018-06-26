// Copyright 2017 The go-ethereum Authors
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

package core

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereumproject/ethash"
	"github.com/ethereumproject/go-ethereum/common"
	"github.com/ethereumproject/go-ethereum/ethdb"
	"github.com/ethereumproject/go-ethereum/event"
	"github.com/ethereumproject/go-ethereum/params"
)

func TestDefaultGenesisBlock(t *testing.T) {
	df := params.DefaultGenesisBlock()
	block := GenesisToBlock(nil, df)
	if block.Nonce() != 66 {
		t.Error("nonce mismatch", block.Nonce(), 42)
	}
	if block.Hash() != params.MainnetGenesisHash {
		spew.Println(block)
		spew.Println(params.DefaultGenesisBlock())
		t.Log(df.Nonce, block.Nonce())
		t.Log(df.Coinbase.Hex(), block.Coinbase().Hex())
		t.Log(df.Mixhash.Hex(), block.MixDigest().Hex())
		t.Log(df.GasLimit, block.GasLimit())
		t.Log(df.GasUsed, block.GasUsed())
		t.Log(string(df.ExtraData), string(block.Extra()))
		t.Log(df.Timestamp, block.Time().String())
		t.Log(df.Difficulty.String(), block.Difficulty().String())
		t.Log(df.Number, block.Number().String())
		t.Log(df.ParentHash.Hex(), block.ParentHash().Hex())
		t.Log(block.Root().Hex())
		t.Log("extra", block.Extra())
		t.Errorf("wrong mainnet genesis hash, got %v, want %v", block.Hash().Hex(), params.MainnetGenesisHash.Hex())
	}
	df = params.DefaultTestnetGenesisBlock()
	block = GenesisToBlock(nil, df)
	if block.Nonce() != 120325427979630 {
		t.Error("nonce mismatch", block.Nonce(), 42)
	}
	if block.Hash() != params.TestnetGenesisHash {
		// spew.Println(block)
		// for _, diff := range pretty.Diff(block, params.DefaultConfigMorden.Genesis) {
		// 	t.Log("diff", diff)
		// }
		t.Errorf("wrong testnet genesis hash, got %v, want %v", block.Hash().Hex(), params.TestnetGenesisHash.Hex())
	}
}

func TestSetupGenesis(t *testing.T) {
	var (
		customghash = common.HexToHash("0x89c99d90b79719238d2645c7642f2c9295246e80775b38cfd162b696817fbd50")
		customg     = params.Genesis{
			Config: &params.ChainConfig{HomesteadBlock: big.NewInt(3)},
			Alloc: params.GenesisAlloc{
				{1}: {Balance: big.NewInt(1), Storage: map[common.Hash]common.Hash{{1}: {1}}},
			},
		}
		oldcustomg = customg
	)
	oldcustomg.Config = &params.ChainConfig{HomesteadBlock: big.NewInt(2)}
	tests := []struct {
		name       string
		fn         func(ethdb.Database) (*params.ChainConfig, common.Hash, error)
		wantConfig *params.ChainConfig
		wantHash   common.Hash
		wantErr    error
	}{
		{
			name: "genesis without ChainConfig",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				return SetupGenesisBlock(db, new(params.Genesis))
			},
			wantErr:    errGenesisNoConfig,
			wantConfig: params.AllEthashProtocolChanges,
		},
		{
			name: "no block in DB, genesis == nil",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				return SetupGenesisBlock(db, nil)
			},
			wantHash:   params.MainnetGenesisHash,
			wantConfig: params.DefaultConfigMainnet.ChainConfig,
		},
		{
			name: "mainnet block in DB, genesis == nil",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				MustCommitGenesis(db, params.DefaultGenesisBlock())
				return SetupGenesisBlock(db, nil)
			},
			wantHash:   params.MainnetGenesisHash,
			wantConfig: params.DefaultConfigMainnet.ChainConfig,
		},
		{
			name: "custom block in DB, genesis == nil",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				MustCommitGenesis(db, &customg)
				return SetupGenesisBlock(db, nil)
			},
			wantHash:   customghash,
			wantConfig: customg.Config,
		},
		{
			name: "custom block in DB, genesis == testnet",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				MustCommitGenesis(db, &customg)
				return SetupGenesisBlock(db, params.DefaultTestnetGenesisBlock())
			},
			wantErr:    &GenesisMismatchError{Stored: customghash, New: params.TestnetGenesisHash},
			wantHash:   params.TestnetGenesisHash,
			wantConfig: params.DefaultConfigMorden.ChainConfig,
		},
		{
			name: "compatible config in DB",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				MustCommitGenesis(db, &oldcustomg)
				return SetupGenesisBlock(db, &customg)
			},
			wantHash:   customghash,
			wantConfig: customg.Config,
		},
		{
			name: "incompatible config in DB",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				// CommitGenesis the 'old' genesis block with Homestead transition at #2.
				// Advance to block #4, past the homestead transition block of customg.
				// genesis := oldcustomg.MustCommit(db)
				genesis := MustCommitGenesis(db, &oldcustomg)

				fakepow, err := ethash.NewForTesting()
				if err != nil {
					return nil, common.Hash{}, err
				}
				bc, _ := NewBlockChain(db, oldcustomg.Config, fakepow, new(event.TypeMux))
				defer bc.Stop()

				blocks, _ := GenerateChain(oldcustomg.Config, genesis, db, 4, nil)
				bc.InsertChain(blocks)
				bc.CurrentBlock()
				// This should return a compatibility error.
				return SetupGenesisBlock(db, &customg)
			},
			wantHash:   customghash,
			wantConfig: customg.Config,
			wantErr: &params.ConfigCompatError{
				What:         "Homestead fork block",
				StoredConfig: big.NewInt(2),
				NewConfig:    big.NewInt(3),
				RewindTo:     1,
			},
		},
	}

	for _, test := range tests {
		db, _ := ethdb.NewMemDatabase()
		config, hash, err := test.fn(db)
		// Check the return values.
		if !reflect.DeepEqual(err, test.wantErr) {
			spew := spew.ConfigState{DisablePointerAddresses: true, DisableCapacities: true}
			t.Errorf("%s: returned error %#v, want %#v", test.name, spew.NewFormatter(err), spew.NewFormatter(test.wantErr))
		}
		if !reflect.DeepEqual(config, test.wantConfig) {
			t.Errorf("%s:\nreturned %v\nwant     %v", test.name, config, test.wantConfig)
		}
		if hash != test.wantHash {
			t.Errorf("%s: returned hash %s, want %s", test.name, hash.Hex(), test.wantHash.Hex())
		} else if err == nil {
			// Check database content.
			stored := GetBlock(db, test.wantHash)
			if stored.Hash() != test.wantHash {
				t.Errorf("%s: block in DB has hash %s, want %s", test.name, stored.Hash(), test.wantHash)
			}
		}
	}
}
