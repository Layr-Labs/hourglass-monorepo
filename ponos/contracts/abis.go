package contracts

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

//go:embed *
var abis embed.FS

const (
	relativeChainContractPath = "abi/chain-contracts.json"
	relativeAbiPathFormat     = "abi/%s/%s.abi.json"
)

type AbiEntry struct {
	Abi     *abi.ABI
	Address common.Address
}

type chainEntry struct {
	ChainId   config.ChainId               `json:"chainId"`
	Contracts map[string]map[string]string `json:"contracts"`
}

func GetChainAbis(chainId config.ChainId, contractName string) ([]AbiEntry, error) {
	bytes, err := abis.ReadFile(relativeChainContractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read chain-contracts.json: %w", err)
	}

	var entries []chainEntry
	if err := json.Unmarshal(bytes, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse chain-contracts.json: %w", err)
	}

	var result []AbiEntry
	for _, entry := range entries {
		if entry.ChainId != chainId {
			continue
		}

		for version, contracts := range entry.Contracts {
			addrStr, ok := contracts[contractName]
			if !ok || addrStr == "" {
				continue
			}

			addr := common.HexToAddress(addrStr)
			parsedABI, err := getContractAbi(version, contractName)
			if err != nil {
				return nil, fmt.Errorf("failed to load ABI for %s: %w", contractName, err)
			}

			result = append(result, AbiEntry{
				Abi:     parsedABI,
				Address: addr,
			})
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no contract %q found on chain %d", contractName, chainId)
	}
	return result, nil
}

func getContractAbi(version string, contractName string) (*abi.ABI, error) {
	abiFile := fmt.Sprintf(relativeAbiPathFormat, version, contractName)
	abiBytes, err := abis.ReadFile(abiFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded ABI file %s: %w", abiFile, err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}
	return &parsedABI, nil
}
