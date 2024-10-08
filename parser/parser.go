package parser

import (
	"context"
	"fmt"
	"strconv"
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
	storage                    Storage
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

func NewParser(storage Storage, startFrom int64) *MyParser {

	var addresses = make(map[string]*AddressTransactions)
	var latestBlockNumber int64
	var err error

	if storage != nil {
		addresses, latestBlockNumber, err = storage.Load()
		if err != nil {
			fmt.Printf("Error loading from storage: %v\n", err)
			latestBlockNumber = startFrom
		}
		if latestBlockNumber == -1 {
			latestBlockNumber = startFrom
		}
	}

	return &MyParser{
		latestProcessedBlockNumber: latestBlockNumber,
		subscribedAddresses:        addresses,
		storage:                    storage,
	}
}

func (s *MyParser) PollLatestBlock(endpoint string) (int64, error) {
	blockNumber, err := rpcclient.GetLatestBlockNumber(endpoint)
	if err != nil {
		return 0, err
	}
	return blockNumber.Int64(), nil
}

func (s *MyParser) ProcessBlock(blockNumber int64, endpoint string) bool {
	txfound := false
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
				txfound = true
				details.Transactions = append(details.Transactions, txDetails)
				fmt.Printf("Transaction found for address: %s; Hash: %s; Block: %s\n", address, txHash, strconv.FormatInt(blockNumber, 10))
			}
		}
		s.mu.Unlock()
	}
	return txfound
}

func (s *MyParser) transactionExists(transactions []Transaction, txHash string) bool {
	for _, tx := range transactions {
		if tx.Txhash == txHash {
			return true
		}
	}
	return false
}

func (s *MyParser) Save() {
	if s.storage != nil {
		s.storage.Save(s.subscribedAddresses, s.latestProcessedBlockNumber)
	}
}

func (s *MyParser) Loop(ctx context.Context, endpoint string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Loop stopped, saving data...")
			s.Save()
			return
		case <-ticker.C:
			txfound := false
			//fmt.Println("Looping")
			latestBlockNumber, err := s.PollLatestBlock(endpoint)
			if err != nil {
				fmt.Println("Error polling latest block:", err)
				continue
			}
			if latestBlockNumber > s.latestProcessedBlockNumber {
				for blockNumber := s.latestProcessedBlockNumber + 1; blockNumber <= latestBlockNumber; blockNumber++ {
					fmt.Println("Processing block number:", blockNumber)
					s.latestProcessedBlockNumber = blockNumber
					txfound = s.ProcessBlock(blockNumber, endpoint)
				}
				if txfound {
					s.Save()
				}
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
