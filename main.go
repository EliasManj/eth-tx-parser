package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/EliasManj/tx-parser/api"
	"github.com/EliasManj/tx-parser/parser"
)

func main() {

	rpcURL := flag.String("url", "https://ethereum-rpc.publicnode.com", "Ethereum RPC URL")
	startFrom := flag.String("startblock", "", "Optional: Block Number to start parsing from")
	flag.Parse()

	if *startFrom != "" {
		start, err := strconv.ParseInt(*startFrom, 10, 64)
		if err != nil {
			fmt.Println("Error parsing start block:", err)
			return
		}
		parser.Init(*rpcURL, start)
	} else {
		parser.Init(*rpcURL)
	}

	//parser.Init("http://localhost:8545")
	//parser.Init("https://ethereum-rpc.publicnode.com")
	//parser.Init("https://ethereum-sepolia-rpc.publicnode.com/", 6836867)

	// Define HTTP handlers
	http.HandleFunc("/", api.HelloHandler)
	http.HandleFunc("/getCurrentBlock", api.GetCurrentBlockHandler)
	http.HandleFunc("/subscribe", api.SubscribeHandler)
	http.HandleFunc("/getTransactions", api.GetTransactionsHandler)
	http.HandleFunc("/getSubscriptions", api.GetSubscriptionsHandler)

	// Start the HTTP server
	fmt.Println("Server is running on port 8082...")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		fmt.Println("Failed to start server:", err)
		return
	}
}
