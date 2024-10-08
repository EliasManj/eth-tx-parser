package parser

import (
	"context"

	"github.com/EliasManj/tx-parser/rpcclient"
)

var (
	db *MyParser
)

func GetParser() *MyParser {
	return db
}

func Init(ctx context.Context, endpoint string, storage Storage, optionalStartFrom ...int64) *MyParser {
	startingBlockNumber, err := rpcclient.GetLatestBlockNumber(endpoint)
	if err != nil {
		panic("Error getting latest block number")
	}
	var startFrom int64
	if len(optionalStartFrom) > 0 {
		startFrom = optionalStartFrom[0]
	} else {
		startFrom = startingBlockNumber.Int64() - 1
	}
	db = NewParser(storage, startFrom)
	go db.Loop(ctx, endpoint)
	return db
}
