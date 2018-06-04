/*
2018-06-04 06:56:41 Geth Classic version: whilei.v5.4.0+143-3916a08[branch=frankenstein/sync-pbs,commit=init-working-basic-progress-bar-display,built=Jun/04Mon-06:56:22]
2018-06-04 06:56:41 Blockchain: Ethereum Classic Mainnet
2018-06-04 06:56:41 Chain database: /Users/ia/Library/EthereumClassic/mainnet/chaindata
2018-06-04 06:56:41 Blockchain upgrades configured: 6
2018-06-04 06:56:41  1150000 Homestead (0x584bdb5d4e74fe97f5a5222b533fe1322fd0b6ad3eb03f02c3221984e2c0b430)
2018-06-04 06:56:41  1920000 The DAO Hard Fork (0x94365e3a8c0b35089c1d1195081fe7489b528a84b22199c916180db8b28ade7f)
2018-06-04 06:56:41  2500000 GasReprice (0xca12c63534f565899681965528d536c52cb05b7c48e269c2a6cb77ad864d878a)
2018-06-04 06:56:41  3000000 Diehard (0x20a2817ab5545ae3c7aed17e534cffc813040b15e4e095c1eb687b85ff5e5305)
2018-06-04 06:56:41  5000000 Gotham (0xb2f55d12af971452c3669669380e03040ff01fabb64afec6bfddb3052dbd0117)
2018-06-04 06:56:41  5900000 Defuse Difficulty Bomb (0x52bc7bbcf1d9251f3f3541ef8c138ee4deacf14e5d66f9614a9bf95d19611bd4)
2018-06-04 06:56:41 Using 10 configured bootnodes
2018-06-04 06:56:41 Use Sputnik EVM: true
2018-06-04 06:56:41 Allotted 4000MB cache and 1024 file handles to /Users/ia/Library/EthereumClassic/mainnet/chaindata
2018-06-04 06:56:48 Allotted 16MB cache and 16 file handles to /Users/ia/Library/EthereumClassic/mainnet/dapp
2018-06-04 06:56:48 Protocol Versions: [63 62], Network Id: 1, Chain Id: 61
2018-06-04 06:56:49 Genesis block: 0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3 (mainnet)
2018-06-04 06:56:49 Local head header:     #3067492 [0x7d034b…] TD=49543988623626722722
2018-06-04 06:56:49 Local head full block: #0 [0xd4e567…] TD=17179869184
2018-06-04 06:56:49 Local head fast block: #3057108 [0x04ac2b…] TD=49395841392101262790
2018-06-04 06:56:49 Fast sync mode enabled.
2018-06-04 06:56:49 Starting server...
2018-06-04 06:56:49 UDP listening. Client enode: enode://1ec5f8dd067172f75ebe404d5e278a118e452a913a17ddfbf102070bb72fefda5d3eb4bceccbdd6a45858b5950dd84723ca55e18e79c7d66d8281aafa1a9d2a5@[::]:30303
2018-06-04 06:56:49 HTTP endpoint: http://localhost:8545
2018-06-04 06:56:49 Debug log config: verbosity=0 log-dir=/Users/ia/Library/EthereumClassic/mainnet/log vmodule=*
2018-06-04 06:56:49 Display log config: display=3 status=sync=1m0s
2018-06-04 06:56:49 Machine log config: mlog=kv mlog-dir=/Users/ia/Library/EthereumClassic/mainnet/mlogs
2018-06-04 06:56:49 IPC endpoint opened: /Users/ia/Library/EthereumClassic/mainnet/geth.ipc
headers  3208371 / 5932806 [******************************************************************************************|----------------------------------------------------------------------------]  54.08%
receipts 3196404 / 5932806 [*****************************************************************************************|-----------------------------------------------------------------------------]  53.88%
blocks 0 / 5932806 [-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------]   0.00%
peers 4 / 50 [============|----------------------------------------------------------------------------------------------------------------------------------------------] -removed 4d3b78723@Parity/v1.10.4
*/

package main

import (
	"regexp"
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
					// name := strings.Split(d.Peer.Name(), "/")[0]
					name := d.Peer.Name()
					re := regexp.MustCompile(`v\d*\.\d*\.\d*`)
					matches := re.FindStringSubmatch(name)
					clientName := strings.Split(name, "/")[0]
					if matches != nil {
						name = clientName + "/" + matches[0]
					}
					peersBar.Set(d.PMPeersLen).Postfix(" -removed " + d.Peer.ID().String()[:9] + "@" + name)
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
					name := d.Peer.Name()
					re := regexp.MustCompile(`v\d*\.\d*\.\d*`)
					matches := re.FindStringSubmatch(name)
					clientName := strings.Split(name, "/")[0]
					if matches != nil {
						name = clientName + "/" + matches[0]
					}
					peersBar.Set(d.PMPeersLen).Postfix("  +added " + d.Peer.ID().String()[:9] + "@" + name)
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
