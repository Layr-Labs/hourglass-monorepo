package eigenlayer

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
)

//go:embed coreContracts
var CoreContracts embed.FS

type CoreContractsData struct {
	Contracts []*contracts.Contract `json:"contracts"`
}

func LoadContracts() ([]*contracts.Contract, error) {
	data, err := CoreContracts.ReadFile("coreContracts/contracts.json")

	if err != nil {
		return nil, fmt.Errorf("failed to load core contracts: %w", err)
	}

	return loadCoreContractsFromJsonData(data)
}

func LoadContractsFromRuntime(jsonData string) ([]*contracts.Contract, error) {
	return loadCoreContractsFromJsonData([]byte(jsonData))
}

func LoadOverrideContract(jsonData string) (*contracts.Contract, error) {
	var contract *contracts.Contract
	if err := json.Unmarshal([]byte(jsonData), &contract); err != nil {
		return nil, fmt.Errorf("failed to unmarshal override contract data: %w", err)
	}

	if contract == nil {
		return nil, fmt.Errorf("override contract data is nil")
	}

	return contract, nil
}

func loadCoreContractsFromJsonData(jsonData []byte) ([]*contracts.Contract, error) {
	var coreContractsData *CoreContractsData
	if err := json.Unmarshal(jsonData, &coreContractsData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal core contracts data: %w", err)
	}

	return coreContractsData.Contracts, nil
}
