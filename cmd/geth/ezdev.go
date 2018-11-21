package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereumproject/go-ethereum/core"
	"github.com/ethereumproject/go-ethereum/logger"
	"github.com/ethereumproject/go-ethereum/logger/glog"
	"gopkg.in/urfave/cli.v1"
)

func setCTXDefault(ctx *cli.Context, name, val string) {
	if !ctx.GlobalIsSet(aliasableName(name, ctx)) {
		ctx.GlobalSet(name, val)
	}
}

func setEZDevFlags(ctx *cli.Context) {
	setCTXDefault(ctx, NoDiscoverFlag.Name, "true")
	setCTXDefault(ctx, LightKDFFlag.Name, "true")
}

func setupEZDev(ctx *cli.Context, config *core.SufficientChainConfig) error {
	glog.Errorln("Initializing EZDEV...")

	// set flag config defaults
	// copy dev.json, dev_genesis.json, dev_genesis_allow.csv to datadir/ezdev/{ chain, dev_genesis, dev_genesis_alloc }.json/csv,
	// init 10 accounts with password 'foo'
	// add these accounts to genesis alloc csv

	config.Include = []string{"dev_genesis.json"}

	cg := config.Genesis

	// Set original genesis to nil so no conflict between GenesisAlloc field and present Genesis obj.
	config.Genesis = nil
	cg.AllocFile = "dev_genesis_alloc.csv"

	// because this cli ctx is weird and accounts seem slower to generate if this isn't here
	setEZDevFlags(ctx)

	// cc.Genesis = cg
	// make some accounts
	accman := MakeAccountManager(ctx)
	data := []byte{}
	bal := "10000000000000000000000000000000"
	if len(accman.Accounts()) == 0 {
		glog.D(logger.Warn).Infoln("No existing EZDEV accounts found, creating 10")
		password := ""
		// accounts := []accounts.Account{}
		for i := 0; i < 10; i++ {
			acc, err := accman.NewAccount(password)
			if err != nil {
				return err
			}
			// accounts = append(accounts, acc)
			// a := acc.Address.Hex()
			a := strings.Replace(acc.Address.Hex(), "0x", "", -1)
			d := fmt.Sprintf(`"%s","%v"
`, a, bal)
			glog.D(logger.Warn).Infoln(acc.Address.Hex(), acc.File)
			// b, ok := new(big.Int).SetString(bal, 10)
			// if !ok {
			// 	panic("not ok set string", b, bal)
			// }
			data = append(data, []byte(d)...)
		}
	} else {
		glog.D(logger.Warn).Infoln("Found existing keyfiles, using: ")
		for _, acc := range accman.Accounts() {
			d := fmt.Sprintf("%s,%v\n", acc.Address.Hex(), bal)
			glog.D(logger.Warn).Infoln(acc.Address.Hex(), acc.File)
			data = append(data, []byte(d)...)
		}
	}

	// marshal and write config json IFF it doesn't already exist
	if _, err := os.Stat(filepath.Join(MustMakeChainDataDir(ctx), "chain.json")); err != nil && os.IsNotExist(err) {

		if err := config.WriteToJSONFile(filepath.Join(MustMakeChainDataDir(ctx), "chain.json")); err != nil {
			return err
		}
	}

	// marshal and write dev_genesis.json
	genC, err := json.MarshalIndent(struct {
		Genesis *core.GenesisDump `json:"genesis"`
	}{cg}, "", "    ")
	if err != nil {
		return fmt.Errorf("Could not marshal json from chain config: %v", err)
	}
	if err := ioutil.WriteFile(filepath.Join(MustMakeChainDataDir(ctx), "dev_genesis.json"), genC, 0644); err != nil {
		return err
	}

	// write alloc file, ALWAYS, because these never change and it's just extra logic, even though it would seem more right to care if the file already exists or not
	ioutil.WriteFile(filepath.Join(MustMakeChainDataDir(ctx), "dev_genesis_alloc.csv"), data, os.ModePerm)

	// again.. hacky. maybe unnecessary.
	cc, err := core.ReadExternalChainConfigFromFile(filepath.Join(MustMakeChainDataDir(ctx), "chain.json"))
	if err != nil {
		panic(err)
	}
	config.Genesis = cc.Genesis
	config.ChainConfig.Automine = true

	return nil
}
