package parser

import (
	"encoding/json"
	"fmt"
	"os"
)

type Storage interface {
	Save(subscribedAddresses map[string]*AddressTransactions, latestBlockNumber int64) error
	Load() (map[string]*AddressTransactions, int64, error)
	Display() string
}

type JsonFileStorage struct {
	FilePath string
}

var _ Storage = &JsonFileStorage{}

func (s *JsonFileStorage) Display() string {
	return fmt.Sprintf("Json File Storage - %s", s.FilePath)
}

func (s *JsonFileStorage) Save(subscribedAddresses map[string]*AddressTransactions, latestBlockNumber int64) error {
	data := map[string]interface{}{
		"latestBlockNumber":   latestBlockNumber,
		"subscribedAddresses": subscribedAddresses,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	err = os.WriteFile(s.FilePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	fmt.Println("Data saved to file:", s.FilePath)
	return nil
}

func (s *JsonFileStorage) Load() (map[string]*AddressTransactions, int64, error) {
	data, err := os.ReadFile(s.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File does not exist, starting fresh")
			return make(map[string]*AddressTransactions), 0, nil
		}
		return nil, -1, fmt.Errorf("failed to read file: %v", err)
	}

	var result struct {
		LatestBlockNumber   int64                           `json:"latestBlockNumber"`
		SubscribedAddresses map[string]*AddressTransactions `json:"subscribedAddresses"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal data: %v", err)
	}

	fmt.Println("Data loaded from file:", s.FilePath)
	return result.SubscribedAddresses, result.LatestBlockNumber, nil
}
