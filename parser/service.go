package parser

import (
	"context"
	"fmt"

	"github.com/EliasManj/tx-parser/rpcclient"
)

var (
	db *MyParser
)

func GetParser() *MyParser {
	return db
}

func Init(ctx context.Context, endpoint string, filename string, optionalStartFrom ...int64) *MyParser {
	startingBlockNumber, err := rpcclient.GetLatestBlockNumber(endpoint)
	if err != nil {
		panic("Error getting latest block number")
	}
	var startFrom int64
	if len(optionalStartFrom) > 0 {
		startFrom = optionalStartFrom[0]
	} else {
		startFrom = startingBlockNumber.Int64()
	}
	db = &MyParser{
		latestProcessedBlockNumber: startFrom,
		subscribedAddresses:        make(map[string]*AddressTransactions),
	}
	err = db.LoadFromFile(filename)
	if err != nil {
		fmt.Println("Error loading data:", err)
	}
	go db.Loop(ctx, endpoint, filename)
	return db
}
