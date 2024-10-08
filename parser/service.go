package parser

import "github.com/EliasManj/tx-parser/rpcclient"

var (
	db *MyParser
)

func GetParser() *MyParser {
	return db
}

func Init(endpoint string, optionalStartFrom ...int64) *MyParser {
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
	go db.Loop(endpoint)
	return db
}
