***

### Personal
The `personal` api exposes method for personal  the methods to manage, control or monitor your node. It allows for limited file system access.

***

#### personal.listAccounts
    personal.listAccounts

List all accounts

#### Return
collection with accounts

#### Example
` personal.listAccounts`

***

#### personal.newAccount

    personal.newAccount(passwd)

Create a new password protected account

#### Return
`string` address of the new account

#### Example
` personal.newAccount("mypasswd")`

` personal.newAccount() # will prompt for the password`

***


#### personal.unlockAccount
    personal.unlockAccount(addr, passwd, duration)

Unlock the account with the given address, password and an optional duration (in seconds). If password is not given you will be prompted for it.

#### Return
`boolean` indication if the account was unlocked

#### Example
` personal.unlockAccount(eth.coinbase, "mypasswd", 300)`

***

### Geth

***

#### geth_getAddressTransactions

Returns transactions for an address.

Usage requires address-transaction indexes using `geth --atxi` to enable and create indexes during chain sync/import, optionally  using `geth atxi-build` to index pre-existing chain data.


##### Parameters
1. `DATA`, 20 Bytes - address to check for transactions
2. `QUANTITY` - integer block number to filter transactions floor
3. `QUANTITY` - integer block number to filter transactions ceiling
4. `STRING` - `[t|f|tf|]`, use `t` for transactions _to_ the address, `f` for _from_, or `tf`/`''` for both
5. `STRING` - `[s|c|sc|]`, use `s` for _standard_ transactions, `c` for _contracts_, or `sc`/``''` for both
6. `QUANTITY` - integer of index to begin pagination. Using `-1` equivalent to `0`.
7. `QUANTITY` - integer of index to end pagination. Using `-1` equivalent to last transaction _n_.
8. `BOOL` - whether to return transactions in order of oldest first. By default `false` returns transaction hashes ordered by newest transactions first.

```
params: [
   '0x407d73d8a49eeb85d32cf465507dd71d507100c1',
   123, // earliest block
   456, // latest block, use 0 for "undefined", ie. eth.blockNumber
   't', // only transactions to this address
   '', // both standard and contract transactions
   -1, // do not trim transactions for pagination (start)
   -1, // do not trim transactions for pagination (end)
   false // do not reverse order ('true' will reverse order to be oldest first)
]
```

##### Returns

`Array` - Array of transaction hashes, or an empty array if no transactions found

##### Example
```js
> geth.getAddressTransactions("0xf4e6FeA8C10C05fe9E2C2FA7545e4c9dd3993a26", 0, 0, "tf", "sc", -1, -1, true)

["0x148809a063efc39e66e35a27a72a82747905071bc2c3b7fc12370dd979eac650", "0x5649e7346ed868bde9ef3a532f8140aeb4171392d278da8e030b26540e248f8a", "0x11bc379dd4f42db7bce759e89dbfda8420fa489e785b1989374f719dac1923dd", "0xdc983ced410b96a95d9a27c9ed88c20ddf735c45797237a34c8f103bbad30caa", "0x58532fa1492a77df622fac57e5e4853f417dcbc3c92940c9df0a2bde72b303f9", "0x223c8589024914b293b92b1fbda08636dd1d3a121fc75532aaa046d579ff641d", "0xd5332daa2e8cb8621912ada4ce09bb1ed8d5831844f8260c4bd07e39677f1201", "0x15bced756880910783272beafc644b9d755291e9fa643ae7c305c6cae961fa26", "0x099e6323f5f9a09197fa5032a546f0f0706b3b8e31404e297e80fcea89210ccd", "0xb5c8b065561e1ee144c2999786accdb5626c10b31c691e01b3c94f22380c0143", "0x743156d73d92595d6005124b6130e4ed85e52312af08aa1303e7ea53741c8cff", "0x38d4128c12c1c4b7a0aab8ab028f95860e2a5a5deb4cd9c992d6cea5f3c45c2b", "0xf87d4e67aa21fda8e749fbedf2e6a6f9bb499d9e4f94e2faae65b473718d6905", "0x4aa8bb43108488e247d52e57ae50ba115b5e95452b89aa4cee92458cb2c9e148", "0x13dc8baa1f4bc0076095e7d73a9aa22e049a30e064e7ea13b34d1498f108730c", "0x351f388bd8271feef0b3b81dcbc500b1f8a0b16064722fca76612a6d04e37378", "0x02ecebe4cc15179991c202b249f358bbc81c26613c66d16bdcc201a550557a7b", "0xd5a50c70909b9f494495449df6cd3e3f5621de41ff5ab4174b066b61468ddbcc"]
```

***

### TxPool

***

#### txpool.status
    txpool.status

Number of pending/queued transactions

#### Return
`pending` all processable transactions

`queued` all non-processable transactions

#### Example
` txpool.status`

***

### Admin

The `admin` exposes the methods to manage, control or monitor your node. It allows for limited file system access.

***

#### admin.nodeInfo

    admin.nodeInfo

##### Returns

information on the node.

##### Example

```
> admin.nodeInfo
{
   Name: 'Ethereum(G)/v0.9.36/darwin/go1.4.1',
   NodeUrl: 'enode://c32e13952965e5f7ebc85b02a2eb54b09d55f553161c6729695ea34482af933d0a4b035efb5600fc5c3ea9306724a8cbd83845bb8caaabe0b599fc444e36db7e@89.42.0.12:30303',
   NodeID: '0xc32e13952965e5f7ebc85b02a2eb54b09d55f553161c6729695ea34482af933d0a4b035efb5600fc5c3ea9306724a8cbd83845bb8caaabe0b599fc444e36db7e',
   IP: '89.42.0.12',
   DiscPort: 30303,
   TCPPort: 30303,
   Td: '0',
   ListenAddr: '[::]:30303'
}
```

To connect to a node, use the [enode-format](https://github.com/ethereumproject/wiki/wiki/enode-url-format) nodeUrl as an argument to [addPeer](#adminaddpeer) or with CLI param `bootnodes`.

***

#### admin.addPeer

    admin.addPeer(nodeURL)

Pass a `nodeURL` to connect a to a peer on the network. The `nodeURL` needs to be in [enode URL format](https://github.com/ethereumproject/wiki/wiki/enode-url-format). geth will maintain the connection until it
shuts down and attempt to reconnect if the connection drops intermittently.

You can find out your own node URL by using [nodeInfo](#adminnodeinfo) or looking at the logs when the node boots up e.g.:

```
[P2P Discovery] Listening, enode://6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec012937307647bd3b9a82abe2974e1407241d54947bbb39763a4cac9f77166ad92a0@54.169.166.226:30303
```

##### Returns

`true` on success.

##### Example

```javascript
> admin.addPeer('enode://6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec012937307647bd3b9a82abe2974e1407241d54947bbb39763a4cac9f77166ad92a0@54.169.166.226:30303')
```

***

#### admin.peers

    admin.peers

##### Returns

an array of objects with information about connected peers.

##### Example

```
> admin.peers
[ { ID: '0x6cdd090303f394a1cac34ecc9f7cda18127eafa2a3a06de39f6d920b0e583e062a7362097c7c65ee490a758b442acd5c80c6fce4b148c6a391e946b45131365b', Name: 'Ethereum(G)/v0.9.0/linux/go1.4.1', Caps: 'eth/56, shh/2', RemoteAddress: '54.169.166.226:30303', LocalAddress: '10.1.4.216:58888' } { ID: '0x4f06e802d994aaea9b9623308729cf7e4da61090ffb3615bc7124c5abbf46694c4334e304be4314392fafcee46779e506c6e00f2d31371498db35d28adf85f35', Name: 'Mist/v0.9.0/linux/go1.4.2', Caps: 'eth/58, shh/2', RemoteAddress: '37.142.103.9:30303', LocalAddress: '10.1.4.216:62393' } ]
```
***

#### admin.importChain

    admin.importChain(file)

Imports the blockchain from a marshalled binary format.
**Note** that the blockchain is reset (to genesis) before the imported blocks are inserted to the chain.


##### Returns

`true` on success, otherwise `false`.

##### Example

```javascript
admin.importChain('path/to/file')
// true
```

***

#### admin.exportChain

    admin.exportChain(file)

Exports the blockchain to the given file in binary format.

##### Returns

`true` on success, otherwise `false`.

##### Example

```javascript
admin.exportChain('path/to/file')
```

***

#### admin.startRPC

     admin.startRPC(host, portNumber, corsheader, modules)

Starts the HTTP server for the [JSON-RPC](https://github.com/ethereumproject/wiki/wiki/JSON-RPC).

##### Returns

`true` on success, otherwise `false`.

##### Example

```javascript
admin.startRPC("127.0.0.1", 8545, "*", "web3,net,eth")
// true
```

***

#### admin.stopRPC

    admin.stopRPC()

Stops the HTTP server for the [JSON-RPC](https://github.com/ethereumproject/wiki/wiki/JSON-RPC).

##### Returns

`true` on success, otherwise `false`.

##### Example

```javascript
admin.stopRPC()
// true
```

***

#### admin.startWS

     admin.startWS(host, portNumber, allowedOrigins, modules)

Starts the websocket server for the [JSON-RPC](https://github.com/ethereumproject/wiki/wiki/JSON-RPC).

##### Returns

`true` on success, otherwise `false`.

##### Example

```javascript
admin.startWS("127.0.0.1", 8546, "*", "web3,net,eth")
// true
```

***

#### admin.stopWS

    admin.stopWS()

Stops the websocket server for the [JSON-RPC](https://github.com/ethereumproject/wiki/wiki/JSON-RPC).

##### Returns

`true` on success, otherwise `false`.

##### Example

```javascript
admin.stopWS()
// true
```

***

#### admin.sleep

    admin.sleep(s)

Sleeps for s seconds.

***

#### admin.sleepBlocks

    admin.sleepBlocks(n)

Sleeps for n blocks.

***

#### admin.datadir

    admin.datadir

the directory this nodes stores its data

##### Returns

directory on success

##### Example

```javascript
admin.datadir
'/Users/username/Library/Ethereum'
```

***

#### admin.setSolc

    admin.setSolc(path2solc)

Set the solidity compiler

##### Returns

a string describing the compiler version when path was valid, otherwise an error

##### Example

```javascript
admin.setSolc('/some/path/solc')
'solc v0.9.29
Solidity Compiler: /some/path/solc
'
```

***

#### admin.startNatSpec

     admin.startNatSpec()

activate NatSpec: when sending a transaction to a contract,
Registry lookup and url fetching is used to retrieve authentic contract Info for it. It allows for prompting a user with authentic contract-specific confirmation messages.

***

#### admin.stopNatSpec

     admin.stopNatSpec()

deactivate NatSpec: when sending a transaction, the user  will be prompted with a generic confirmation message, no contract info is fetched

***

#### admin.getContractInfo

     admin.getContractInfo(address)

this will retrieve the [contract info json](./Contracts-and-Transactions#contract-info-metadata) for a contract on the address

##### Returns

returns the contract info object

##### Examples

```js
> info = admin.getContractInfo(contractaddress)
> source = info.source
> abi = info.abiDefinition
```

***
#### admin.saveInfo

    admin.saveInfo(contract.info, filename);

will write [contract info json](./Contracts-and-Transactions#contract-info-metadata) into the target file, calculates its content hash. This content hash then can used to associate a public url with where the contract info is publicly available and verifiable. If you register the codehash (hash of the code of the contract on contractaddress).

##### Returns

`contenthash` on success, otherwise `undefined`.

##### Examples

```js
source = "contract test {\n" +
"   /// @notice will multiply `a` by 7.\n" +
"   function multiply(uint a) returns(uint d) {\n" +
"      return a * 7;\n" +
"   }\n" +
"} ";
contract = eth.compile.solidity(source).test;
txhash = eth.sendTransaction({from: primary, data: contract.code });
// after it is uncluded
contractaddress = eth.getTransactionReceipt(txhash);
filename = "/tmp/info.json";
contenthash = admin.saveInfo(contract.info, filename);
```

***
#### admin.register

    admin.register(address, contractaddress, contenthash);

will register content hash to the codehash (hash of the code of the contract on contractaddress). The register transaction is sent from the address in the first parameter. The transaction needs to be processed and confirmed on the canonical chain for the registration to take effect.

##### Returns

`true` on success, otherwise `false`.

##### Examples

```js
source = "contract test {\n" +
"   /// @notice will multiply `a` by 7.\n" +
"   function multiply(uint a) returns(uint d) {\n" +
"      return a * 7;\n" +
"   }\n" +
"} ";
contract = eth.compile.solidity(source).test;
txhash = eth.sendTransaction({from: primary, data: contract.code });
// after it is uncluded
contractaddress = eth.getTransactionReceipt(txhash);
filename = "/tmp/info.json";
contenthash = admin.saveInfo(contract.info, filename);
admin.register(primary, contractaddress, contenthash);
```

***

#### admin.registerUrl

    admin.registerUrl(address, codehash, contenthash);

this will register a contant hash to the contract' codehash. This will be used to locate [contract info json](./Contracts-and-Transactions#contract-info-metadata)
files. Address in the first parameter will be used to send the transaction.

##### Returns

`true` on success, otherwise `false`.

##### Examples

```js
source = "contract test {\n" +
"   /// @notice will multiply `a` by 7.\n" +
"   function multiply(uint a) returns(uint d) {\n" +
"      return a * 7;\n" +
"   }\n" +
"} ";
contract = eth.compile.solidity(source).test;
txhash = eth.sendTransaction({from: primary, data: contract.code });
// after it is uncluded
contractaddress = eth.getTransactionReceipt(txhash);
filename = "/tmp/info.json";
contenthash = admin.saveInfo(contract.info, filename);
admin.register(primary, contractaddress, contenthash);
admin.registerUrl(primary, contenthash, "file://"+filename);
```

***

### Miner

***

#### miner.start

    miner.start(threadCount)

Starts [mining](see ./Mining) on with the given `threadNumber` of parallel threads. This is an optional argument.

##### Returns

`true` on success, otherwise `false`.

##### Example

```javascript
miner.start()
// true
```

***

#### miner.stop

    miner.stop()

##### Returns

`true` on success, otherwise `false`.

##### Example

```javascript
miner.stop()
// true
```

***

#### miner.startAutoDAG

    miner.startAutoDAG()

Starts automatic pregeneration of the [ethash DAG](https://github.com/ethereumproject/wiki/wiki/Ethash-DAG). This process make sure that the DAG for the subsequent epoch is available allowing mining right after the new epoch starts. If this is used by most network nodes, then blocktimes are expected to be normal at epoch transition. Auto DAG is switched on automatically when mining is started and switched off when the miner stops.

##### Returns

`true` on success, otherwise `false`.

***

#### miner.stopAutoDAG

    miner.stopAutoDAG()

Stops automatic pregeneration of the [ethash DAG](https://github.com/ethereumproject/wiki/wiki/Ethash-DAG). Auto DAG is switched off automatically when mining is stops.

##### Returns

`true` on success, otherwise `false`.

***

#### miner.makeDAG

    miner.makeDAG(blockNumber, dir)

Generates the DAG for epoch `blockNumber/epochLength`. dir specifies a target directory,
If `dir` is the empty string, then ethash will use the default directories `~/.ethash` on Linux and MacOS, and `~\AppData\Ethash` on Windows. The DAG file's name is `full-<revision-number>R-<seedhash>`

##### Returns

`true` on success, otherwise `false`.

***

#### miner.setExtra

    miner.setExtra("extra data")

**Sets** the extra data for the block when finding a block. Limited to 32 bytes.

***

#### miner.setGasPrice

    miner.setGasPrice(gasPrice)

**Sets** the the gasprice for the miner

***

#### miner.setEtherbase

    miner.setEtherbase(account)

**Sets** the the ether base, the address that will receive mining rewards.

***

### Debug

***

#### debug.setHead

    debug.setHead(blockNumber)

**Sets** the current head of the blockchain to the block referred to by _blockNumber_.
See [web3.eth.getBlock](https://github.com/ethereumproject/wiki/wiki/JavaScript-API#web3ethgetblock) for more details on block fields and lookup by number or hash.

##### Returns

`true` on success, otherwise `false`.

##### Example

    debug.setHead(eth.blockNumber-1000)

***

#### debug.seedHash

    debug.seedHash(blockNumber)

Returns the hash for the epoch the given block is in.

##### Returns

hash in hex format

##### Example

    > debug.seedHash(eth.blockNumber)
    '0xf2e59013a0a379837166b59f871b20a8a0d101d1c355ea85d35329360e69c000'

***

#### debug.getBlockRlp

    debug.getBlockRlp(blockNumber)

Returns the hexadecimal representation of the RLP encoding of the block.
See [web3.eth.getBlock](https://github.com/ethereumproject/wiki/wiki/JavaScript-API#web3ethgetblock) for more details on block fields and lookup by number or hash.

##### Returns

The hex representation of the RLP encoding of the block.

##### Example

```
> debug.getBlockRlp(131805)    'f90210f9020ba0ea4dcb53fe575e23742aa30266722a15429b7ba3d33ba8c87012881d7a77e81ea01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d4934794a4d8e9cae4d04b093aac82e6cd355b6b963fb7ffa01f892bfd6f8fb2ec69f30c8799e371c24ebc5a9d55558640de1fb7ca8787d26da056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b901000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000083bb9266830202dd832fefd880845534406d91ce9e5448ce9ed0af535048ce9ed0afce9ea04cf6d2c4022dfab72af44e9a58d7ac9f7238ffce31d4da72ed6ec9eda60e1850883f9e9ce6a261381cc0c0'
```

***

#### debug.printBlock

    debug.printBlock(blockNumber)

Prints information about the block such as size, total difficulty, as well as header fields properly formatted.

See [web3.eth.getBlock](https://github.com/ethereumproject/wiki/wiki/JavaScript-API#web3ethgetblock) for more details on block fields and lookup by number or hash.

##### Returns

formatted string representation of the block

##### Example

```
> debug.printBlock(131805)
BLOCK(be465b020fdbedc4063756f0912b5a89bbb4735bd1d1df84363e05ade0195cb1): Size: 531.00 B TD: 643485290485 {
NoNonce: ee48752c3a0bfe3d85339451a5f3f411c21c8170353e450985e1faab0a9ac4cc
Header:
[

        ParentHash:         ea4dcb53fe575e23742aa30266722a15429b7ba3d33ba8c87012881d7a77e81e
        UncleHash:          1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347
        Coinbase:           a4d8e9cae4d04b093aac82e6cd355b6b963fb7ff
        Root:               1f892bfd6f8fb2ec69f30c8799e371c24ebc5a9d55558640de1fb7ca8787d26d
        TxSha               56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421
        ReceiptSha:         56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421
        Bloom:              00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
        Difficulty:         12292710
        Number:             131805
        GasLimit:           3141592
        GasUsed:            0
        Time:               1429487725
        Extra:              ΞTHΞЯSPHΞЯΞ
        MixDigest:          4cf6d2c4022dfab72af44e9a58d7ac9f7238ffce31d4da72ed6ec9eda60e1850
        Nonce:              3f9e9ce6a261381c
]
Transactions:
[]
Uncles:
[]
}


```


***


#### debug.dumpBlock

    debug.dumpBlock(blockNumber)

##### Returns

the raw dump of a block referred to by block number or block hash or undefined if the block is not found.
see [web3.eth.getBlock](https://github.com/ethereumproject/wiki/wiki/JavaScript-API#web3ethgetblock) for more details on block fields and lookup by number or hash.

##### Example


```js
> debug.dumpBlock(eth.blockNumber)
```


***

#### debug.metrics

    debug.metrics(raw)

##### Returns

Collection of metrics, see for more information [this](./Metrics-and-Monitoring) wiki page.

##### Example


```js
> metrics(true)
```


***

#### debug.accountExist

```
debug.accountExist(address, blockNumber)
```

##### Returns

Returns `BOOL` if a given account exists at a given block. Whether an account
exists affects the gas cost of a transaction.


##### Example
```js
debug.accountExist("0x102e61f5d8f9bc71d0ad4a084df4e65e05ce0e1c", 1000)
> true
```


***

### Additional interfaces

***

#### loadScript

     loadScript('/path/to/myfile.js');

Loads a JavaScript file and executes it. Relative paths are interpreted as relative to `jspath` which is specified as a command line flag, see [Command Line Options](./Command-Line-Options).

#### setInterval

    setInterval(s, func() {})

#### clearInterval
#### setTimeout
#### clearTimeout

***

#### web3
The `web3` exposes all methods of the [JavaScript API](https://github.com/ethereumproject/wiki/wiki/JavaScript-API).

***

#### net
The `net` is a shortcut for [web3.net](https://github.com/ethereumproject/wiki/wiki/JavaScript-API#web3net).

***

#### eth
The `eth` is a shortcut for [web3.eth](https://github.com/ethereumproject/wiki/wiki/JavaScript-API#web3eth). In addition to the `web3` and `eth` interfaces exposed by [web3.js](https://github.com/ethereumproject/web3.js) a few additional calls are exposed.

***

#### eth.sign

    eth.sign(signer, data)

#### eth.pendingTransactions

    eth.pendingTransactions

Returns pending transactions that belong to one of the users `eth.accounts`.

***

#### eth.resend

    eth.resend(tx, <optional gas price>, <optional gas limit>)

Resends the given transaction returned by `pendingTransactions()` and allows you to overwrite the gas price and gas limit of the transaction.

##### Example

```javascript
eth.sendTransaction({from: eth.accounts[0], to: "...", gasPrice: "1000"})
var tx = eth.pendingTransactions[0]
eth.resend(tx, web3.toWei(10, "szabo"))
```

***


#### shh
The `shh` is a shortcut for [web3.shh](https://github.com/ethereumproject/wiki/wiki/JavaScript-API#web3shh).

***


#### inspect
The `inspect` method pretty prints the given value (supports colours)

***
