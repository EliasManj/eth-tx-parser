package rpcclient

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/EliasManj/tx-parser/utils"
	"github.com/stretchr/testify/require"
)

var (
	URL      string
	Accounts []interface{}
	StdGas   int64 = 21000
)

func TestMain(m *testing.M) {

	var err error
	URL = "http://127.0.0.1:8545"
	fmt.Println("URL:", URL)

	Accounts, err = AnvilGetAccounts(URL)
	if err != nil {
		log.Fatalf("Error getting accounts: %v", err)
	}
	code := m.Run()
	os.Exit(code)
}

func TestGetBalance(t *testing.T) {
	acc1 := Accounts[0].(string)
	balance, err := GetBalance(acc1, URL)
	require.NoError(t, err)
	require.NotEmpty(t, balance)

	wei, err := utils.HexToDec(balance)
	require.NoError(t, err)
	require.NotEmpty(t, wei)
}

func TestAnvilSendWei(t *testing.T) {
	acc1 := Accounts[0].(string)
	acc2 := Accounts[1].(string)
	acc1Balance, err := GetBalance(acc1, URL)
	require.NoError(t, err)
	acc1wei, err := utils.HexToDec(acc1Balance)
	require.NoError(t, err)

	toSend := 100
	txHash, err := AnvilSendWei(acc1, acc2, toSend, URL)
	require.NoError(t, err)
	require.NotEmpty(t, txHash)
	time.Sleep(1 * time.Second)

	newAcc1Balance, err := GetBalance(acc1, URL)
	require.NoError(t, err)
	newacc1wei, err := utils.HexToDec(newAcc1Balance)
	require.NoError(t, err)

	require.NotEqual(t, acc1wei, newacc1wei)
}

func TestGetBlockNumber(t *testing.T) {
	blockNumber, err := GetLatestBlockNumber(URL)
	require.NoError(t, err)
	require.NotEmpty(t, blockNumber)
}

func TestGetTransactions(t *testing.T) {
	acc1 := Accounts[0].(string)
	acc2 := Accounts[1].(string)

	txHash, err := AnvilSendWei(acc1, acc2, 100, URL)
	require.NoError(t, err)
	require.NotEmpty(t, txHash)
	time.Sleep(1 * time.Second)
	recipt, err := GetTransactionReceipt(txHash, URL)
	require.NoError(t, err)
	require.NotEmpty(t, recipt)
	block := recipt["blockNumber"].(string)
	require.NotEmpty(t, block)

	txs, err := GetTransactionsByBlockNumber(block, acc1, URL)
	require.NoError(t, err)
	require.NotEmpty(t, txs)

	txhasinblocktxs := false
	for _, tx := range txs {
		if tx["hash"].(string) == txHash {
			txhasinblocktxs = true
		}
	}
	require.True(t, txhasinblocktxs)
}

func TestGetAddresTxHistory(t *testing.T) {
	acc1 := Accounts[0].(string)
	acc2 := Accounts[1].(string)
	acc3 := Accounts[2].(string)

	txHash1, err := AnvilSendWei(acc1, acc2, 100, URL)
	require.NoError(t, err)
	require.NotEmpty(t, txHash1)
	txHash2, err := AnvilSendWei(acc1, acc3, 100, URL)
	require.NoError(t, err)
	require.NotEmpty(t, txHash2)
	txHash3, err := AnvilSendWei(acc3, acc1, 100, URL)
	require.NoError(t, err)
	require.NotEmpty(t, txHash3)

	time.Sleep(1 * time.Second)

	transactions, err := GetAddressTxHistory(acc1, URL)
	require.NoError(t, err)
	require.NotEmpty(t, transactions)

	tx1Found := false
	tx2Found := false
	tx3Found := false

	for _, tx := range transactions {
		if txHash, ok := tx["hash"].(string); ok {
			if txHash == txHash1 {
				tx1Found = true
			}
			if txHash == txHash2 {
				tx2Found = true
			}
			if txHash == txHash3 {
				tx3Found = true
			}
		}
	}

	require.True(t, tx1Found, "txHash1 not found in transactions")
	require.True(t, tx2Found, "txHash2 not found in transactions")
	require.True(t, tx3Found, "txHash2 not found in transactions")
}
