package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var errInvalidArgument = errors.New("invalid argument")
var errInvalidHeader = errors.New("invalid header")

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if err := validate(scanner.Text()); err != nil {
			log.Fatal(err)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	os.Exit(0)
}

func validate(in string) error {

	args := strings.Split(in, " ")
	if len(args) != 3 {
		return (errInvalidArgument)
	}

	headerJSON, parentJSON, isUncleStr := args[0], args[1], args[2]

	// Initial sanity checks that args are actually present and legit.
	_, err := strconv.ParseBool(isUncleStr)
	if err != nil {
		return (err)
	}
	if !strings.Contains(headerJSON, "Hash") {
		return fmt.Errorf("%v: %s", errInvalidArgument, headerJSON)
	}
	if !strings.Contains(parentJSON, "Hash") {
		return fmt.Errorf("%v: %s", errInvalidArgument, parentJSON)
	}

	var header, parent *struct {
		Number uint64
		Hash   string
	}

	if err := json.Unmarshal([]byte(headerJSON), &header); err != nil {
		return (err)
	}
	if err := json.Unmarshal([]byte(parentJSON), &parent); err != nil {
		return (err)
	}

	// TODO Implement your own header validations.

	if header.Number != parent.Number+1 {
		return fmt.Errorf("%v: %d %d", errInvalidHeader, header.Number, parent.Number)
	}

	// // Here's an example of making an upstream RPC request (to a TRUSTED source),
	// // then using that data to provide supplemental information for block validation.
	// remoteRPCAPI, err := url.Parse("http://localhost:8545")
	// if err != nil {
	// 	return (err)
	// }

	// req := &rpc.JSONRequest{
	// 	// Id
	// 	Version: "2.0",
	// 	Method:  "eth_blockNumber",
	// 	// Payload
	// }

	// b, err := json.Marshal(req)
	// if err != nil {
	// 	return (err)
	// }

	// res, err := http.Post(remoteRPCAPI.String(), "application/json", bytes.NewBuffer(b))
	// if err != nil {
	// 	return (err)
	// }

	// var r *rpc.JSONResponse
	// err = json.NewDecoder(res.Body).Decode(&r)
	// if err != nil {
	// 	return (err)
	// }

	// if r.Error != nil {
	// 	return (r.Error)
	// }

	// // TODO Do some logic with response
	// // ...

	return nil
}
