package eigenlayer

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
)

//go:embed coreContracts
var CoreContracts embed.FS

type CoreContractsData struct {
	Contracts []*contracts.Contract `json:"contracts"`
}

func LoadCoreContractsForL1Chain(chainId config.ChainId) (map[string]*contracts.Contract, error) {
	var data []byte
	var err error
	switch chainId {
	case config.ChainId_EthereumHolesky:
		data, err = CoreContracts.ReadFile("coreContracts/holesky.json")
	case config.ChainId_EthereumHoodi:
		return nil, fmt.Errorf("chainId %d not supported", chainId)

	case config.ChainId_EthereumMainnet:
		data, err = CoreContracts.ReadFile("coreContracts/mainnet.json")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load core contracts for chainId %d: %w", chainId, err)
	}

	var coreContractsData *CoreContractsData
	if err = json.Unmarshal(data, &coreContractsData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal core contracts data: %w", err)
	}

	mappedContracts := make(map[string]*contracts.Contract)
	for _, contractData := range coreContractsData.Contracts {
		mappedContracts[contractData.Address] = contractData
	}

	return mappedContracts, nil
}
