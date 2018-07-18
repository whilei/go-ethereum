package params

import (
	"github.com/ethereumproject/go-ethereum/logger/glog"
)

var (
	DefaultConfigMainnet *SufficientChainConfig
	DefaultConfigMorden  *SufficientChainConfig
	// TestChainConfig      *ChainConfig
)

func init() {

	var err error

	DefaultConfigMainnet, err = parseExternalChainConfig("/params/config/mainnet.json", assetsOpen)
	if err != nil {
		glog.Fatal("Error parsing mainnet defaults from JSON:", err)
	}
	DefaultConfigMainnet.ChainConfig = DefaultConfigMainnet.ChainConfig.SetForkBlockVals()

	DefaultConfigMorden, err = parseExternalChainConfig("/params/config/morden.json", assetsOpen)
	if err != nil {
		glog.Fatal("Error parsing morden defaults from JSON:", err)
	}
	DefaultConfigMorden.ChainConfig = DefaultConfigMorden.ChainConfig.SetForkBlockVals()

	// TestChainConfig = DefaultConfigMorden.ChainConfig
}
