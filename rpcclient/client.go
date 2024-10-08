package rpcclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/EliasManj/tx-parser/utils"
)

func sendRequest(endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error encoding JSON: %v", err)
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return result, nil
}

func GetBalance(address string, endpoint string) (string, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBalance",
		"params":  []interface{}{address, "latest"},
		"id":      1,
	}

	result, err := sendRequest(endpoint, payload)
	if err != nil {
		return "", err
	}

	if balance, ok := result["result"].(string); ok {
		return balance, nil
	}

	return "", fmt.Errorf("balance not found in response")
}

func GetTransactionsByBlockNumber(blockNumber string, address string, endpoint string) ([]map[string]interface{}, error) {
	address = strings.ToLower(address)
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{blockNumber, true},
		"id":      1,
	}
	result, err := sendRequest(endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	blockData := result["result"].(map[string]interface{})
	transactions := blockData["transactions"].([]interface{})
	var filteredTxs []map[string]interface{}
	for _, tx := range transactions {
		txData := tx.(map[string]interface{})
		fromAddress := strings.ToLower(txData["from"].(string))
		var toAddress string
		if txData["to"] != nil {
			toAddress = strings.ToLower(txData["to"].(string))
		}
		if fromAddress == address || (toAddress != "" && toAddress == address) {
			filteredTxs = append(filteredTxs, txData)
		}
	}
	return filteredTxs, nil
}

func GetLatestBlockNumber(endpoint string) (*big.Int, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}

	result, err := sendRequest(endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	blockNumberStr := result["result"].(string)
	number, err := utils.HexToDec(blockNumberStr)
	if err != nil {
		return nil, fmt.Errorf("error converting block number: %v", err)
	}

	return number, nil
}

func AnvilGetAccounts(endpoint string) ([]interface{}, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_accounts",
		"params":  []interface{}{},
		"id":      1,
	}

	result, err := sendRequest(endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	accounts := result["result"].([]interface{})

	return accounts, nil
}

func AnvilSendWei(from string, to string, amt int, endpoint string) (string, error) {
	value := utils.IntToHex(amt)
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_sendTransaction",
		"params": []interface{}{
			map[string]interface{}{
				"from":  from,
				"to":    to,
				"value": value,
				"gas":   "0x5208",
			},
		},
		"id": 1,
	}

	result, err := sendRequest(endpoint, payload)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}

	if result["error"] != nil {
		errorInfo := result["error"].(map[string]interface{})
		return "", fmt.Errorf("RPC error: %v", errorInfo["message"])
	}

	txHash, ok := result["result"].(string)
	if !ok {
		return "", fmt.Errorf("transaction failed: no result returned")
	}

	return txHash, nil
}

func GetTransactionReceipt(txHash string, endpoint string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionReceipt",
		"params":  []interface{}{txHash},
		"id":      1,
	}

	result, err := sendRequest(endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	if result["error"] != nil {
		errorInfo := result["error"].(map[string]interface{})
		return nil, fmt.Errorf("RPC error: %v", errorInfo["message"])
	}

	receipt := result["result"].(map[string]interface{})
	return receipt, nil
}

func GetAddressTxHistory(address string, endpoint string) ([]map[string]interface{}, error) {
	var transactions []map[string]interface{}
	latestBlock, err := GetLatestBlockNumber(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block number: %v", err)
	}
	maxBlock := latestBlock.Int64()

	for blockNumber := int64(0); blockNumber <= maxBlock; blockNumber++ {
		blockHex := utils.IntToHex(blockNumber)
		txs, err := GetTransactionsByBlockNumber(blockHex, address, endpoint)
		if err != nil {
			return nil, fmt.Errorf("error fetching block %d: %v", blockNumber, err)
		}

		for _, tx := range txs {
			if tx["from"] == address || tx["to"] == address {
				transactions = append(transactions, tx)
			}
		}
	}
	return transactions, nil
}
