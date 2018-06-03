package main

import (
	"strings"
	"sync"

	"github.com/ethereumproject/go-ethereum/eth"
	"github.com/ethereumproject/go-ethereum/eth/downloader"
	"github.com/ethereumproject/go-ethereum/eth/fetcher"
	"github.com/ethereumproject/go-ethereum/logger/glog"
	"gopkg.in/cheggaaa/pb.v1"
	"gopkg.in/urfave/cli.v1"

	"time"
)

var headersBar *pb.ProgressBar
var receiptsBar *pb.ProgressBar
var blocksBar *pb.ProgressBar
var peersBar *pb.ProgressBar
var barPool *pb.Pool

// greenDisplaySystem is "spec'd" in PR #423 and is a little fancier/more detailed and colorful than basic.
var progressDisplaySystem = displayEventHandlers{
	{
		eventT: logEventDownloaderInsertChain,
		ev:     downloader.InsertChainEvent{},
		handlers: displayEventHandlerFns{
			func(ctx *cli.Context, e *eth.Ethereum, evData interface{}, tickerInterval time.Duration) {
				switch d := evData.(type) {
				case downloader.InsertChainEvent:
					_, _, height, _, _ := e.Downloader().Progress()
					for _, b := range []*pb.ProgressBar{blocksBar, receiptsBar, headersBar} {
						b.Set64(int64(d.LastNumber)).SetTotal64(int64(height))
					}
				}
			},
		},
	},
	{
		eventT: logEventDownloaderInsertHeaderChain,
		ev:     downloader.InsertHeaderChainEvent{},
		handlers: displayEventHandlerFns{
			func(ctx *cli.Context, e *eth.Ethereum, evData interface{}, tickerInterval time.Duration) {
				switch d := evData.(type) {
				case downloader.InsertHeaderChainEvent:
					headersBar.Set64(int64(d.LastNumber))
					_, _, height, _, _ := e.Downloader().Progress()
					for _, b := range []*pb.ProgressBar{blocksBar, receiptsBar, headersBar} {
						b.SetTotal64(int64(height))
					}
				}
			},
		},
	},
	{
		eventT: logEventDownloaderInsertReceiptChain,
		ev:     downloader.InsertReceiptChainEvent{},
		handlers: displayEventHandlerFns{
			func(ctx *cli.Context, e *eth.Ethereum, evData interface{}, tickerInterval time.Duration) {
				switch d := evData.(type) {
				case downloader.InsertReceiptChainEvent:
					_, _, height, _, _ := e.Downloader().Progress()
					receiptsBar.Set64(int64(d.LastNumber))
					for _, b := range []*pb.ProgressBar{blocksBar, receiptsBar, headersBar} {
						b.SetTotal64(int64(height))
					}
				}
			},
		},
	},
	{
		eventT: logEventFetcherInsert,
		ev:     fetcher.FetcherInsertBlockEvent{},
		handlers: displayEventHandlerFns{
			func(ctx *cli.Context, e *eth.Ethereum, evData interface{}, tickerInterval time.Duration) {
				switch d := evData.(type) {
				case fetcher.FetcherInsertBlockEvent:
					for _, b := range []*pb.ProgressBar{blocksBar, receiptsBar, headersBar} {
						b.SetTotal64(d.Block.Number().Int64()).Set64(d.Block.Number().Int64())
					}
				}
			},
		},
	},
	{
		eventT: logEventPMHandlerRemove,
		ev:     eth.PMHandlerRemoveEvent{},
		handlers: displayEventHandlerFns{
			func(ctx *cli.Context, e *eth.Ethereum, evData interface{}, tickerInterval time.Duration) {
				switch d := evData.(type) {
				case eth.PMHandlerRemoveEvent:
					peersBar.Set(d.PMPeersLen).Postfix(" -removed " + d.Peer.ID().String()[:9] + "@" + strings.Split(d.Peer.Name(), "/")[0])
				}
			},
		},
	},
	{
		eventT: logEventPMHandlerAdd,
		ev:     eth.PMHandlerAddEvent{},
		handlers: displayEventHandlerFns{
			func(ctx *cli.Context, e *eth.Ethereum, evData interface{}, tickerInterval time.Duration) {
				switch d := evData.(type) {
				case eth.PMHandlerAddEvent:
					peersBar.Set(d.PMPeersLen).Postfix("  +added " + d.Peer.ID().String()[:9] + "@" + strings.Split(d.Peer.Name(), "/")[0])
				}
			},
		},
	},
	{
		eventT: logEventInterval,
		handlers: displayEventHandlerFns{
			func(ctx *cli.Context, e *eth.Ethereum, evData interface{}, tickerInterval time.Duration) {
			},
		},
	},
	{
		eventT: logEventBefore,
		handlers: displayEventHandlerFns{
			func(ctx *cli.Context, e *eth.Ethereum, evData interface{}, tickerInterval time.Duration) {
				currentBlockNumber = e.BlockChain().CurrentFastBlock().NumberU64()
				go func() {
					headersBar = pb.New(int(currentBlockNumber)).Prefix("headers ").Set(int(currentBlockNumber))
					receiptsBar = pb.New(int(currentBlockNumber)).Prefix("receipts").Set(int(currentBlockNumber))
					blocksBar = pb.New(int(currentBlockNumber)).Prefix("blocks")
					if e.Downloader().GetMode() == downloader.FullSync {
						blocksBar.Set(int(e.BlockChain().CurrentBlock().NumberU64()))
					}
					peersBar = pb.New(ctx.GlobalInt(aliasableName(MaxPeersFlag.Name, ctx))).Prefix("peers")
					bars := []*pb.ProgressBar{headersBar, receiptsBar, blocksBar, peersBar}
					for i := range bars {
						bars[i].ManualUpdate = true
						bars[i].ShowElapsedTime = false
						bars[i].ShowFinalTime = false
						bars[i].AutoStat = false
						bars[i].ShowSpeed = false
						bars[i].ShowTimeLeft = false
						bars[i].ShowPercent = true
						bars[i].Format("[*|-]")
					}
					peersBar.ShowPercent = false
					peersBar.Format("[=|-]")
					wg := new(sync.WaitGroup)
					wg.Add(1)
					pool, err := pb.StartPool(headersBar, receiptsBar, blocksBar, peersBar)
					if err != nil {
						glog.Fatal(err)
					}
					barPool = pool
					wg.Wait()
					barPool.Stop()
				}()
			},
		},
	},
}
