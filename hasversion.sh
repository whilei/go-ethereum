#!/usr/bin/env bash

go build -ldflags "-X main.Version=%VERSION%" github.com\ethereumproject\go-ethereum\cmd\geth
if .\geth version | grep --quiet v5; then
	exit 0
else
	exit 1
fi

