#!/usr/bin/env bash

go build -ldflags "-X main.Version=asdf" github.com/ethereumproject/go-ethereum/cmd/geth
if ./geth version | grep --quiet asdf; then
	echo ok
	exit 0
else
	echo notok
	exit 1
fi

