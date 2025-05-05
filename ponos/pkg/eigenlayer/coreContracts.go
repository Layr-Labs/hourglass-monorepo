package eigenlayer

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"slices"
)

//go:embed coreContracts
var CoreContracts embed.FS

type CoreContractData struct {
	ContractAddress string `json:"contract_address"`
	ContractAbi     string `json:"contract_abi"`
	BytecodeHash    string `json:"bytecode_hash"`
}

type CoreProxyContractData struct {
	ContractAddress      string `json:"contract_address"`
	ProxyContractAddress string `json:"proxy_contract_address"`
	BlockNumber          int64  `json:"block_number"`
}

type CoreContractsData struct {
	CoreContracts  []*CoreContractData      `json:"core_contracts"`
	ProxyContracts []*CoreProxyContractData `json:"proxy_contracts"`
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

	// address --> Contract
	mappedContracts := make(map[string]*contracts.Contract)

	proxyContractAddresses := make([]string, 0)
	for _, contract := range coreContractsData.ProxyContracts {
		// front-facing proxy contract
		proxyContractAddr := contract.ContractAddress

		c, ok := mappedContracts[proxyContractAddr]
		if !ok {
			c = &contracts.Contract{
				Address:     proxyContractAddr,
				AbiVersions: make([]string, 0),
			}
			mappedContracts[proxyContractAddr] = c
			proxyContractAddresses = append(proxyContractAddresses, proxyContractAddr)

			baseAbi := util.Find(coreContractsData.CoreContracts, func(cc *CoreContractData) bool {
				return cc.ContractAddress == proxyContractAddr
			})
			if baseAbi != nil {
				c.AbiVersions = append(c.AbiVersions, baseAbi.ContractAbi)
			}
		}

		// find the implementations
		foundAbi := util.Find(coreContractsData.CoreContracts, func(cc *CoreContractData) bool {
			return cc.ContractAddress == contract.ProxyContractAddress
		})
		if foundAbi == nil {
			return nil, fmt.Errorf("failed to find ABI for proxy contract %s", contract.ProxyContractAddress)
		}
		c.AbiVersions = append(c.AbiVersions, foundAbi.ContractAbi)
	}
	// find any core contracts that are NOT proxies but need to be added
	for _, contract := range coreContractsData.CoreContracts {
		proxy := util.Find(coreContractsData.ProxyContracts, func(p *CoreProxyContractData) bool {
			return slices.Contains([]string{p.ContractAddress, p.ProxyContractAddress}, contract.ContractAddress)
		})
		if proxy != nil {
			continue
		}
		c, ok := mappedContracts[contract.ContractAddress]
		if !ok {
			c = &contracts.Contract{
				Address:     contract.ContractAddress,
				AbiVersions: make([]string, 0),
			}
			mappedContracts[contract.ContractAddress] = c
		}
		c.AbiVersions = append(c.AbiVersions, contract.ContractAbi)
	}

	return mappedContracts, nil
}
