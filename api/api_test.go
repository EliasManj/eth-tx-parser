package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/EliasManj/tx-parser/parser"
	"github.com/EliasManj/tx-parser/rpcclient"
	"github.com/stretchr/testify/require"
)

var (
	Accounts []interface{}
	Server   *httptest.Server
	AnvilUrl string = "http://127.0.0.1:8545"
)

func TestMain(m *testing.M) {
	var err error
	url := "http://localhost:8545"
	parser.Init(url)
	Accounts, err = rpcclient.AnvilGetAccounts(url)
	if err != nil {
		log.Fatalf("Error getting accounts: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", HelloHandler)
	mux.HandleFunc("/getCurrentBlock", GetCurrentBlockHandler)
	mux.HandleFunc("/subscribe", SubscribeHandler)
	mux.HandleFunc("/getTransactions", GetTransactionsHandler)
	mux.HandleFunc("/getSubscriptions", GetSubscriptionsHandler)

	Server = httptest.NewServer(mux)
	defer Server.Close()

	code := m.Run()
	os.Exit(code)
}

func TestHelloHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(HelloHandler)
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")
}

func TestAPI(t *testing.T) {

	acc0 := Accounts[0].(string)
	acc1 := Accounts[1].(string)
	acc2 := Accounts[2].(string)

	resp1 := subscribeAddress(t, Server, acc0)
	defer resp1.Body.Close()
	time.Sleep(5 * time.Second)
	resp2 := subscribeAddress(t, Server, acc1)
	defer resp2.Body.Close()
	time.Sleep(5 * time.Second)

	subscriptions := getSubscriptions(t, Server)
	require.Len(t, subscriptions, 2, "Expected 2 subscriptions")
	require.Contains(t, subscriptions, acc0, "Expected address 1 to be subscribed")
	require.Contains(t, subscriptions, acc1, "Expected address 2 to be subscribed")

	// do some txs
	_, err := rpcclient.AnvilSendWei(acc0, acc1, 100, AnvilUrl)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)
	_, err = rpcclient.AnvilSendWei(acc0, acc2, 100, AnvilUrl)
	require.NoError(t, err)
	time.Sleep(10 * time.Second)

	transactions := getTransactions(t, Server, acc0)
	require.GreaterOrEqual(t, len(transactions), 2, "Expected at least 2 transactions")
}

func subscribeAddress(t *testing.T, server *httptest.Server, address string) *http.Response {
	req, err := http.NewRequest("GET", server.URL+"/subscribe?address="+address, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK but got: %v", resp.Status)
	}
	return resp
}

func getSubscriptions(t *testing.T, server *httptest.Server) []string {
	req, err := http.NewRequest("GET", server.URL+"/getSubscriptions", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get subscriptions: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK but got: %v", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	var subscriptions []string
	err = json.Unmarshal(body, &subscriptions)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}
	return subscriptions
}

func getTransactions(t *testing.T, server *httptest.Server, address string) []parser.Transaction {
	req, err := http.NewRequest("GET", server.URL+"/getTransactions?address="+address, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK but got: %v", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	var transactions []parser.Transaction
	err = json.Unmarshal(body, &transactions)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}
	return transactions
}
