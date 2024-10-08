package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/EliasManj/tx-parser/rpcclient"
	"github.com/EliasManj/tx-parser/utils"
)

type Transaction struct {
	Txhash          string `json:"txhash"`
	Blockhash       string `json:"blockhash"`
	BlockNumber     string `json:"blocknumber"`
	From            string `json:"from"`
	To              string `json:"to"`
	Txtype          string `json:"txtype"`
	GasUsed         string `json:"gasUsed"`
	GasPrice        string `json:"blobGasPrice"`
	ContractAddress string `json:"contractAddress"`
	Nonce           string `json:"nonce"`
}

type AddressTransactions struct {
	Transactions []Transaction `json:"transactions"`
}

type MyParser struct {
	latestProcessedBlockNumber int64
	subscribedAddresses        map[string]*AddressTransactions
	mu                         sync.RWMutex
}

type Parser interface {
	// Get the last parsed block
	GetCurrentBlock() int

	// Add an address to the observer
	Subscribe(address string) bool

	// List of inbound or outbound transactions for an address
	GetTransactions(address string) []Transaction
}

var _ Parser = &MyParser{}

func (s *MyParser) SaveToFile(filename string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.subscribedAddresses, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	fmt.Println("Data saved to file:", filename)
	return nil
}

func (s *MyParser) LoadFromFile(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File does not exist, starting fresh")
			return nil
		}
		return fmt.Errorf("failed to read file: %v", err)
	}

	err = json.Unmarshal(data, &s.subscribedAddresses)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %v", err)
	}

	fmt.Println("Data loaded from file:", filename)
	return nil
}

func (s *MyParser) PollLatestBlock(endpoint string) (int64, error) {
	blockNumber, err := rpcclient.GetLatestBlockNumber(endpoint)
	if err != nil {
		return 0, err
	}
	return blockNumber.Int64(), nil
}

func (s *MyParser) ProcessBlock(blockNumber int64, endpoint string) {
	for address, details := range s.subscribedAddresses {
		transactions, err := rpcclient.GetTransactionsByBlockNumber(utils.IntToHex(blockNumber), address, endpoint)
		if err != nil {
			fmt.Println("Error getting transactions for block number", blockNumber, "and address", address)
			continue
		}
		s.mu.Lock()
		for _, tx := range transactions {
			txHash, _ := tx["hash"].(string)
			blockHash, _ := tx["blockHash"].(string)
			from, _ := tx["from"].(string)
			to := ""
			if tx["to"] != nil {
				to = tx["to"].(string)
			}
			txtype := ""
			if tx["type"] != nil {
				txtype = tx["type"].(string)
			}
			txDetails := Transaction{
				Txhash:      txHash,
				Blockhash:   blockHash,
				From:        from,
				To:          to,
				BlockNumber: tx["blockNumber"].(string),
				Txtype:      txtype,
				GasUsed:     tx["gas"].(string),
				GasPrice:    tx["gasPrice"].(string),
				Nonce:       tx["nonce"].(string),
			}
			if !s.transactionExists(details.Transactions, txDetails.Txhash) {
				details.Transactions = append(details.Transactions, txDetails)
			}
			fmt.Printf("Transaction found for address %s: %s\n", address, txHash)
		}
		s.mu.Unlock()
	}
}

func (s *MyParser) transactionExists(transactions []Transaction, txHash string) bool {
	for _, tx := range transactions {
		if tx.Txhash == txHash {
			return true
		}
	}
	return false
}

func (s *MyParser) Loop(ctx context.Context, endpoint string, saveFile string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Loop stopped, saving data...")
			s.SaveToFile(saveFile)
			return
		case <-ticker.C:
			fmt.Println("Looping")
			latestBlockNumber, err := s.PollLatestBlock(endpoint)
			if err != nil {
				fmt.Println("Error polling latest block:", err)
				continue
			}
			if latestBlockNumber > s.latestProcessedBlockNumber {
				for blockNumber := s.latestProcessedBlockNumber; blockNumber <= latestBlockNumber; blockNumber++ {
					s.latestProcessedBlockNumber = blockNumber
					s.ProcessBlock(blockNumber, endpoint)
				}
				s.SaveToFile(saveFile)
			}
		}
	}
}

func (s *MyParser) GetCurrentBlock() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int(s.latestProcessedBlockNumber)
}

func (s *MyParser) Subscribe(address string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.subscribedAddresses[address]; exists {
		return false
	}

	s.subscribedAddresses[address] = &AddressTransactions{
		Transactions: []Transaction{},
	}
	return true
}

func (s *MyParser) GetSubscriptions() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var subscriptions []string
	for address := range s.subscribedAddresses {
		subscriptions = append(subscriptions, address)
	}
	return subscriptions
}

func (s *MyParser) GetTransactions(address string) []Transaction {
	addrTrans, exists := s.subscribedAddresses[address]
	if !exists {
		return nil
	}
	return addrTrans.Transactions
}
