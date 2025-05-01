package contracts

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/ethereum/go-ethereum/common"
)

//go:embed *
var chains embed.FS

type chainEntry struct {
	ChainId   config.ChainId    `json:"chainId"`
	Contracts map[string]string `json:"contracts"`
}

func GetContractAddress(contractName string, version string, chainId config.ChainId) (common.Address, error) {
	mapPath := fmt.Sprintf("abi/%s/chain-contracts.json", version)
	bytes, err := chains.ReadFile(mapPath)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to parse chain-contracts.json: %w", err)
	}

	var entries []chainEntry
	if err := json.Unmarshal(bytes, &entries); err != nil {
		return common.Address{}, fmt.Errorf("failed to parse chain-contracts.json: %w", err)
	}

	for _, entry := range entries {
		if entry.ChainId == chainId {
			addrStr := entry.Contracts[contractName]
			if addrStr == "" {
				return common.Address{}, fmt.Errorf("failed to find contract address")
			}
			return common.HexToAddress(addrStr), nil
		}
	}

	return common.Address{}, fmt.Errorf("no address found for contract %s on chain %d", contractName, chainId)
}
