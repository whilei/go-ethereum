
// a function that unlocks all accounts
function unlockall() {
	for (var i = 0; i < eth.accounts.length; i++) {
		var a = eth.accounts[i] 
		personal.unlockAccount(a, "");
		console.log("unlocked", a);
	}
		console.log("unlocked all owned accounts");
}

// a function that makes a lot of transactions
function ntxs(n) {
	for (var i = 0; i < n; i++) {
		var txh = eth.sendTransaction({from: eth.accounts[i%eth.accounts.length], to: eth.accounts[(i%eth.accounts.length)+1], value: web3.toWei(0.33, "ether")}); 
		console.log("tx", i, "pending", eth.pendingTransactions.length, ": ", txh);
	}
}


unlockall();

// note that miner.start() MUST be called prior to processing any txs
miner.start();

ntxs(10000);






