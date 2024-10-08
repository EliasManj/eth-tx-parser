package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/EliasManj/tx-parser/parser"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, the server is running!")
}

func GetCurrentBlockHandler(w http.ResponseWriter, r *http.Request) {
	myparser := parser.GetParser()
	currentBlock := myparser.GetCurrentBlock()
	response := struct {
		CurrentBlock int `json:"currentBlock"`
	}{
		CurrentBlock: currentBlock,
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Missing address parameter", http.StatusBadRequest)
		return
	}
	myparser := parser.GetParser()
	success := myparser.Subscribe(strings.ToLower(address))
	if !success {
		http.Error(w, "Address already subscribed", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Address %s subscribed successfully", address)
}

func GetSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	myparser := parser.GetParser()
	subscriptions := myparser.GetSubscriptions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscriptions)
}

func GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Missing address parameter", http.StatusBadRequest)
		return
	}
	address = strings.ToLower(address)

	myparser := parser.GetParser()

	transactions := myparser.GetTransactions(address)
	if transactions == nil {
		http.Error(w, "No transactions found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(transactions)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}
