package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/EliasManj/tx-parser/api"
	"github.com/EliasManj/tx-parser/parser"
)

func main() {
	rpcURL := flag.String("url", "https://ethereum-rpc.publicnode.com", "Ethereum RPC URL")
	startFrom := flag.String("startblock", "", "Optional: Block Number to start parsing from")
	filename := flag.String("file", "data.json", "File to persist the subscribed addresses and transactions")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{Addr: ":8082"}

	go func() {
		<-sigChan
		fmt.Println("Received shutdown signal, stopping loop and HTTP server...")
		cancel()
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()

		if err := server.Shutdown(ctxShutdown); err != nil {
			fmt.Printf("HTTP server Shutdown: %v\n", err)
		} else {
			fmt.Println("HTTP server stopped")
		}
	}()

	// Initialize the parser
	if *startFrom != "" {
		start, err := strconv.ParseInt(*startFrom, 10, 64)
		if err != nil {
			fmt.Println("Error parsing start block:", err)
			return
		}
		parser.Init(ctx, *rpcURL, *filename, start)
	} else {
		parser.Init(ctx, *rpcURL, *filename)
	}

	//parser.Init("http://localhost:8545")
	//parser.Init("https://ethereum-rpc.publicnode.com")
	//parser.Init("https://ethereum-sepolia-rpc.publicnode.com/", 6836867)

	http.HandleFunc("/", api.HelloHandler)
	http.HandleFunc("/getCurrentBlock", api.GetCurrentBlockHandler)
	http.HandleFunc("/subscribe", api.SubscribeHandler)
	http.HandleFunc("/getTransactions", api.GetTransactionsHandler)
	http.HandleFunc("/getSubscriptions", api.GetSubscriptionsHandler)

	fmt.Println("Server is running on port 8082...")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Println("Failed to start server:", err)
	}
}
