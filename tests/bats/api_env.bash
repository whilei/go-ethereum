#!/usr/bin/env bats

# Current build.
: ${GETH_CMD:=$GOPATH/bin/geth}
: ${GETH_OPTS:=--datadir $BATS_TMPDIR \
               		--lightkdf \
               		--verbosity 0 \
               		--display 0 \
               		--port 33333 \
               		--no-discover \
               		--keystore $GOPATH/src/github.com/ethereumproject/go-ethereum/accounts/testdata/keystore \
               		--unlock "f466859ead1932d743d622cb74fc058882e8648a" \
    }

setup() {
	# GETH_TMP_DATA_DIR=`mktemp -d`
	# mkdir "$BATS_TMPDIR/mainnet"
    testacc=f466859ead1932d743d622cb74fc058882e8648a
	tesetacc_pass=foobar
	regex_signature_success='0x[0-9a-f]{130}'
}