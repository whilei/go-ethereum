package main

import (
	"gopkg.in/urfave/cli.v1"
	"strconv"
	"github.com/ethereumproject/go-ethereum/logger/glog"
	"errors"
	"github.com/ethereumproject/go-ethereum/core/types"
	"github.com/ethereumproject/go-ethereum/core"
	"github.com/ethereumproject/go-ethereum/logger"
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

	bc, chainDB := MakeChain(ctx)
	defer chainDB.Close()


	indexDb := MakeIndexDatabase(ctx)

	var block *types.Block
	block = bc.GetBlockByNumber(blockIndex)

	for block != nil {
		for _, tx := range block.Transactions() {
			var err error
			from, err := tx.From()
			if err != nil {
				return err
			}
			err = core.AddTxA(indexDb, from.Hash(), tx.Hash())
			if err != nil {
				return err
			}

			to := tx.To()
			if to == nil {
				continue
			}
			err = core.AddTxA(indexDb, to.Hash(), tx.Hash())
			if err != nil {
				return err
			}
		}
		glog.V(logger.Error).Infoln("Store tx/addr indexes for block %d/%d with %d txs", block.NumberU64(), bc.CurrentBlock().NumberU64(), block.Transactions().Len())
	}
	return nil
}


