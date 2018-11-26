// note that miner.start() MUST be called prior to processing any txs
miner.start();

// a function that unlocks all accounts
function unlockall() {
	console.log("Unlocking all accounts", "bn=", eth.blockNumber);
	for (var i = 0; i < eth.accounts.length; i++) {
		var a = eth.accounts[i] 
		personal.unlockAccount(a, "");
		// console.log("unlocked", a, "bal=", eth.getBalance(a), "#txs=", geth.getTransactionsByAddress(a, 0, 'latest', 'tf', 's', 0, -1, false).length);
		console.log("unlocked", a, "bal=", eth.getBalance(a));
	}
		console.log("unlocked all owned accounts");
}

// a function that makes a lot of transactions
function ntxs(n) {
	var d = new Date();
	var nn = 0;
	var txps = 0;
	for (var i = 0; i < n; i++) {
		var bn = eth.blockNumber;
		var txh = eth.sendTransaction({from: eth.accounts[i%eth.accounts.length], to: eth.accounts[(i%eth.accounts.length)+1], value: web3.toWei(0.33, "ether")}); 
		nn++;
		console.log("tx n=", i, "bn=", eth.blockNumber, "tx_pool.pending=", eth.pendingTransactions.length, "tx/s=", txps, txh.substring(0,8));
		var dd = new Date();
		if (dd.getSeconds() > d.getSeconds()) {
			d = dd;
			txps = nn;
			nn = 0;
		}
	}
}


unlockall();


// ntxs(10000);






