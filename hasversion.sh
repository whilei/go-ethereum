#!/usr/bin/env bash

echo "building geth..."
go build -ldflags "-X main.Version=%VERSION%" github.com\ethereumproject\go-ethereum\cmd\geth
echo "running .\geth version | grep --quiet v (if statement to determine good/bad)..."
.\geth version | grep --quiet v (
    echo "OK!!"
	exit 0
) || (
    echo "FAIL!!"
	exit 1
)

