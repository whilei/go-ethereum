#!/usr/bin/env bats

load api_env

@test "personal_sign1" {
    run $GETH_CMD $GETH_OPTS \
				--password=<(echo $tesetacc_pass) \
        --exec="personal.sign('0xdeadbeef', '"$testacc"', '"$tesetacc_pass"');" console 2> /dev/null
		echo "$output"
		[ "$status" -eq 0 ]
    [[ "$output" =~ $regex_signature_success ]]
}

@test "personal_sign2" {
    run $GETH_CMD $GETH_OPTS \
				--password=<(echo $tesetacc_pass) \
        --exec="personal.sign(web3.fromAscii('Schoolbus'), '"$testacc"', '"$tesetacc_pass"');" console 2> /dev/null
		echo "$output"
		[ "$status" -eq 0 ]
    [[ "$output" =~ $regex_signature_success ]]
}

@test "personal_listAccounts" {
    run $GETH_CMD $GETH_OPTS \
				--password=<(echo $tesetacc_pass) \
        --exec="personal.listAccounts;" console 2> /dev/null
		echo "$output"
		[ "$status" -eq 0 ]
    [[ "$output" =~ $testacc ]]
}

@test "personal_lockAccount" {
    run $GETH_CMD $GETH_OPTS \
				--password=<(echo $tesetacc_pass) \
        --exec="personal.lockAccount('"$testacc"');" console 2> /dev/null
		echo "$output"
		[ "$status" -eq 0 ]
    [[ "$output" =~ 'true' ]]
}

@test "personal_unlockAccount" {
    run $GETH_CMD $GETH_OPTS \
				--password=<(echo $tesetacc_pass) \
        --exec="personal.lockAccount('"$testacc"') && personal.unlockAccount('"$testacc"', '"$tesetacc_pass"', 0);" console 2> /dev/null
		echo "$output"
		[ "$status" -eq 0 ]
    [[ "$output" =~ 'true' ]]
}
