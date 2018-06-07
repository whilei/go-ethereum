#!/usr/bin/env bats

load api_env

@test "eth_sign1" {
		run $GETH_CMD $GETH_OPTS \
				--password=<(echo $tesetacc_pass) \
        --exec="eth.sign('"$testacc"', '"$d"');" console 2> /dev/null
		echo "$output"
		[ "$status" -eq 0 ]
    [[ "$output" =~ $regex_signature_success ]]
}

@test "eth_sign2" {
    run $GETH_CMD $GETH_OPTS \
				--password=<(echo $tesetacc_pass) \
        --exec="eth.sign('"$testacc"', web3.fromAscii('Schoolbus'));" console 2> /dev/null
		echo "$output"
		[ "$status" -eq 0 ]
    [[ "$output" =~ $regex_signature_success ]]
}
