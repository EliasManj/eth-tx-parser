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
	Endpoint string
}

type EndpointData struct {
	LatestBlockNumber   int64                           `json:"latestBlockNumber"`
	SubscribedAddresses map[string]*AddressTransactions `json:"subscribedAddresses"`
}

var _ Storage = &JsonFileStorage{}

func (s *JsonFileStorage) Display() string {
	return fmt.Sprintf("Json File Storage - %s", s.FilePath)
}

func (s *JsonFileStorage) Save(subscribedAddresses map[string]*AddressTransactions, latestBlockNumber int64) error {
	existingData, err := s.loadAll()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load existing data: %v", err)
	}

	existingData[s.Endpoint] = EndpointData{
		LatestBlockNumber:   latestBlockNumber,
		SubscribedAddresses: subscribedAddresses,
	}

	jsonData, err := json.MarshalIndent(existingData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	err = os.WriteFile(s.FilePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	fmt.Println("Data saved for endpoint:", s.Endpoint)
	return nil
}

func (s *JsonFileStorage) loadAll() (map[string]EndpointData, error) {
	data := make(map[string]EndpointData)

	fileData, err := os.ReadFile(s.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File does not exist, starting fresh")
			return data, nil
		}
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	err = json.Unmarshal(fileData, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %v", err)
	}

	return data, nil
}

func (s *JsonFileStorage) Load() (map[string]*AddressTransactions, int64, error) {
	// Load the entire file data
	existingData, err := s.loadAll()
	if err != nil {
		return nil, 0, err
	}

	// Filter for the specific endpoint
	if endpointData, ok := existingData[s.Endpoint]; ok {
		return endpointData.SubscribedAddresses, endpointData.LatestBlockNumber, nil
	}

	// If no data exists for this endpoint, return fresh data
	return make(map[string]*AddressTransactions), 0, nil
}
