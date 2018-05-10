
#### eth_accounts

Returns a list of addresses owned by client.


##### Parameters
none

##### Returns

`Array of DATA`, 20 Bytes - addresses owned by the client.

##### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}'

// Result
{
"id":1,
"jsonrpc": "2.0",
"result": ["0x407d73d8a49eeb85d32cf465507dd71d507100c1"]
}
```
