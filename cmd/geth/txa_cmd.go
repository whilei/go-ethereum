package main

import (
	"gopkg.in/urfave/cli.v1"
	"strconv"
	"github.com/ethereumproject/go-ethereum/logger/glog"
	"errors"
	"github.com/ethereumproject/go-ethereum/core/types"
	"github.com/ethereumproject/go-ethereum/core"
	"github.com/ethereumproject/go-ethereum/logger"
	"github.com/cheggaaa/pb"
)

func buildTxAIndex(ctx *cli.Context) error {
	index := ctx.Args().First()
	if len(index) == 0 {
		glog.Fatal("FIXME: this message is wrong > missing argument: use `build-txa 12345` to specify required block number to roll back to")
		return errors.New("invalid flag usage")
	}

	blockIndex, err := strconv.ParseUint(index, 10, 64)
	if err != nil {
		glog.Fatalf("FIXME: this message is wrong > invalid argument: use `build-txa 12345`, were '12345' is a required number specifying which block number to roll back to")
		return errors.New("invalid flag usage")
	}

	glog.D(logger.Error).Infoln("number", blockIndex)

	bc, chainDB := MakeChain(ctx)
	if bc == nil || chainDB == nil {
		panic("bc or cdb is nil")
	}
	defer chainDB.Close()

	indexDb := MakeIndexDatabase(ctx)
	if indexDb == nil {
		panic("indexdb is nil")
	}
	defer indexDb.Close()

	var block *types.Block
	block = bc.GetBlockByNumber(blockIndex)
	if block == nil {
		panic("init block is nil")
	}
	// FIXME: able to differentiate a fast sync from full chain
	bar := pb.StartNew(int(bc.CurrentBlock().NumberU64()))
	for block != nil {
		//glog.D(logger.Error).Infoln("got here")
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
			err = core.PutAddrTxIdx(indexDb, block, false, from.Hash(), tx.Hash())
			if err != nil {
				return err
			}

			to := tx.To()
			if to == nil {
				continue
			}
			err = core.PutAddrTxIdx(indexDb, block,true, to.Hash(), tx.Hash())
			if err != nil {
				return err
			}

		}
		bar.Set(int(block.NumberU64()))
		blockIndex++
		block = bc.GetBlockByNumber(blockIndex)
	}
	bar.Finish()
	return nil
}


