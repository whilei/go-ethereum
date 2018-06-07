#!/usr/bin/env bats

# Current build.
# @ghc Use the given temp data dir from Bats as ephemeral data-dir.
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
# @ghc-#
# it also seems like someday it would be good to extract this kind of testing to it's own
# repo. So that we don't have to bundle hardcoded test cases like this dumb key along with geth archives and binaries
# @ghc-

setup() {
    testacc=f466859ead1932d743d622cb74fc058882e8648a
	tesetacc_pass=foobar
	regex_signature_success='0x[0-9a-f]{130}'
}