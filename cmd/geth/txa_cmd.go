package main

import (
	"gopkg.in/urfave/cli.v1"
	"strconv"
	"github.com/ethereumproject/go-ethereum/logger/glog"
	"github.com/ethereumproject/go-ethereum/core/types"
	"github.com/ethereumproject/go-ethereum/core"
	"github.com/cheggaaa/pb"
	"os"
	"path/filepath"
	"io/ioutil"
)

func buildTxAIndex(ctx *cli.Context) error {
	startIndex := uint64(ctx.Int("start"))
	var stopIndex uint64

	// Use persistent placeholder in case start not spec'd
	placeholderFilename := filepath.Join(MustMakeChainDataDir(ctx), "index.at")
	if !ctx.IsSet("start") {
		bs, err := ioutil.ReadFile(placeholderFilename)
		if err == nil { // ignore errors for now
			startIndex, _ = strconv.ParseUint(string(bs), 10, 64)
		}
	}

	bc, chainDB := MakeChain(ctx)
	if bc == nil || chainDB == nil {
		panic("bc or cdb is nil")
	}
	defer chainDB.Close()

	stopIndex = uint64(ctx.Int("stop"))
	if stopIndex == 0 {
		stopIndex = bc.CurrentHeader().Number.Uint64()
	}

	if stopIndex < startIndex {
		glog.Fatal("start must be prior to (smaller than) or equal to stop, got start=", startIndex, "stop=", stopIndex)
	}

	indexDb := MakeIndexDatabase(ctx)
	if indexDb == nil {
		panic("indexdb is nil")
	}
	defer indexDb.Close()

	var block *types.Block
	blockIndex := startIndex
	block = bc.GetBlockByNumber(blockIndex)
	if block == nil {
		glog.Fatal(blockIndex, "block is nil")
	}

	bar := pb.StartNew(int(stopIndex)) // progress bar
	for block != nil && block.NumberU64() <= stopIndex {
		txs := block.Transactions()
		if txs == nil {
			panic("txs were nil")
		}
		for _, tx := range txs {
			var err error
			from, err := tx.From()
			if err != nil {
				return err
			}
			err = core.PutAddrTxs(indexDb, block, false, from, tx.Hash())
			if err != nil {
				return err
			}

			to := tx.To()
			if to == nil {
				continue
			}
			err = core.PutAddrTxs(indexDb, block,true, *to, tx.Hash())
			if err != nil {
				return err
			}

		}
		bar.Set(int(block.NumberU64()))
		blockIndex++
		if blockIndex % 1000 == 0 {
			ioutil.WriteFile(placeholderFilename, []byte(strconv.Itoa(int(blockIndex))), os.ModePerm)
		}
		block = bc.GetBlockByNumber(blockIndex)
	}
	bar.Finish()
	return nil
}


