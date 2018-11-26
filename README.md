# README

### Instructions

Start geth with:

```shell
$ cd go-ethereum
$ make cmd/geth && rm -rf ./keys && rm -rf ~/.ethereum-classic/ezdev && ./bin/geth --ezdev --sputnikvm --keystore ./keys
```
- `rm -rf` ing old development files w/ each rerun; this is not necessary
- new `--ezdev` flag (note that this overrides the `--chain` flag)
- `--sputnikvm` for demonstration
- `--keystore` for demonstrating a custom keystore dir. This is the directory in which keyfiles live. If there are NO keyfiles in this dir, then  EZDev :registered: will generate 10 keys, each with password `foo`. If there ARE ANY key files in this dir, then geth will not generate any new files, and will endow those accounts with substantial premine balances (`10000000000000000000000000000000wei`) in the genesis block. This directory can be anywhere.

Then, in another session let's test out the automine feature. 

```shell
$ ./bin/geth --chain ezdev --preload test.js attach
$ ntxs(10000000000);
```
> [./test.js](./test.js)

### Of note:
- if you smell hack, it's because there is hack
- the genesis and chain config setup "bootstraps itself" by reading a default config 

