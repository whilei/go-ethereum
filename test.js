
var a = {
	a: function() {personal.unlockAccount(eth.accounts[0])},
	b: miner.start,
	c: function() {eth.sendTransaction({from: eth.accounts[0], to: eth.accounts[1], value: web3.toWei(1.33, "ether")}); }
};

a.a()
a.b()

a.c()
a.c()
a.c()



