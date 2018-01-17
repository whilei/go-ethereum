package main

import (
	"gopkg.in/urfave/cli.v1"
	"strconv"
	"github.com/ethereumproject/go-ethereum/logger/glog"
	"errors"
	"github.com/ethereumproject/go-ethereum/core/types"
	"github.com/ethereumproject/go-ethereum/core"
	"github.com/cheggaaa/pb"
	"os"
	"path/filepath"
	"io/ioutil"
)

func buildTxAIndex(ctx *cli.Context) error {
	startIndex := ctx.Args().First()
	var stopIndex string
	filename := filepath.Join(MustMakeChainDataDir(ctx), "startIndex.at")
	if len(startIndex) == 0 {
		bs, err := ioutil.ReadFile(filename)
		if err != nil { // ignore errors for now
			startIndex = "0"
		} else {
			startIndex = string(bs)
		}
	} else {
		if len(ctx.Args()) > 1 {
			stopIndex = ctx.Args()[1]
		}
	}

	blockIndex, err := strconv.ParseUint(startIndex, 10, 64)
	if err != nil {
		glog.Fatalf("FIXME: this message is wrong > invalid argument: use `build-txa 12345`, were '12345' is a required number specifying which block number to roll back to")
		return errors.New("invalid flag usage")
	}
	
	bc, chainDB := MakeChain(ctx)
	if bc == nil || chainDB == nil {
		panic("bc or cdb is nil")
	}
	defer chainDB.Close()

	var stopIndexI uint64
	// If no argument for stop index given ($2), then use bc header height
	if len(stopIndex) == 0 {
		stopIndexI = bc.CurrentHeader().Number.Uint64()
	} else {
		stopIndexI, _ = strconv.ParseUint(stopIndex, 10, 64)
	}

	indexDb := MakeIndexDatabase(ctx)
	if indexDb == nil {
		panic("indexdb is nil")
	}
	defer indexDb.Close()

	var block *types.Block
	block = bc.GetBlockByNumber(blockIndex)
	if block == nil {
		glog.Fatal("block is nil")
	}

	// FIXME: able to differentiate a fast sync from full chain
	bar := pb.StartNew(int(stopIndexI))
	for block != nil && block.NumberU64() <= stopIndexI {
		txs := block.Transactions()
		if txs == nil {
			panic("txs were nil")
		}
		for _, tx := range txs {
			//glog.D(logger.Error).Infoln("got here2")
			var err error
			from, err := tx.From()
			if err != nil {
				return err
			}
			//glog.D(logger.Error).Infoln("got here3")
			err = core.PutAddrTxs(indexDb, block, false, from.Hash(), tx.Hash())
			if err != nil {
				return err
			}

			to := tx.To()
			if to == nil {
				continue
			}
			err = core.PutAddrTxs(indexDb, block,true, to.Hash(), tx.Hash())
			if err != nil {
				return err
			}

		}
		bar.Set(int(block.NumberU64()))
		blockIndex++
		if blockIndex % 1000 == 0 {
			ioutil.WriteFile(filename, []byte(strconv.Itoa(int(blockIndex))), os.ModePerm)
		}
		block = bc.GetBlockByNumber(blockIndex)
	}
	bar.Finish()
	return nil
}


