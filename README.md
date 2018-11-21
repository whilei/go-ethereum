[![MacOS Build Status](https://circleci.com/gh/ethereumproject/go-ethereum/tree/master.svg?style=shield)](https://circleci.com/gh/ethereumproject/go-ethereum/tree/master)
[![Windows Build Status](https://ci.appveyor.com/api/projects/status/github/ethereumproject/go-ethereum?svg=true)](https://ci.appveyor.com/project/splix/go-ethereum)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethereumproject/go-ethereum)](https://goreportcard.com/report/github.com/ethereumproject/go-ethereum)
[![API Reference](https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://godoc.org/github.com/ethereumproject/go-ethereum)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ethereumproject/go-ethereum?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

Start geth with:
- `rm -rf` ing old development files w/ each rerun
- new `--ezdev` flag
- `--sputnikvm` for demonstration
- `--keystore` for demonstrating a custom keystore dir. This is the directory in which keyfiles live. If there are NO keyfiles in this dir, then  EZDev :registered: will generate 10 keys, each with password `foo`. If there ARE ANY key files in this dir, then geth will not generate any new files, and will endow those accounts with substantial premine balances (`10000000000000000000000000000000wei`) in the genesis block. This directory can be anywhere.
```
$ cd go-ethereum
$ make cmd/geth && rm -rf ./keys && rm -rf ~/.ethereum-classic/ezdev && ./bin/geth --ezdev --sputnikvm --keystore ./keys
```

Of note:
- if you smell hack, it's because there is hack
- the genesis and chain config setup "bootstraps itself" by reading a default config 

