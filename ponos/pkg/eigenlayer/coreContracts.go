package eigenlayer

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

//go:embed coreContracts
var CoreContracts embed.FS

type CoreContract struct {
	Address     string
	AbiVersions []string
}

func (c *CoreContract) GetCombinedAbis() (string, error) {
	return combineAbis(c.AbiVersions)
}

func (c *CoreContract) GetAbi() (*abi.ABI, error) {
	combinedAbi, err := c.GetCombinedAbis()
	if err != nil {
		return nil, fmt.Errorf("failed to combine ABIs: %w", err)
	}

	parsedAbi, err := abi.JSON(strings.NewReader(combinedAbi))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return &parsedAbi, nil
}

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

func combineAbis(abis []string) (string, error) {
	abisToCombine := make([]string, 0)

	for _, contractAbi := range abis {
		strippedContractAbi := contractAbi[1 : len(contractAbi)-1]
		abisToCombine = append(abisToCombine, strippedContractAbi)
	}

	combinedAbi := fmt.Sprintf("[%s]", strings.Join(abisToCombine, ","))
	return combinedAbi, nil
}

func LoadCoreContractsForL1Chain(chainId config.ChainId) (map[string]*CoreContract, error) {
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

	// address --> CoreContract
	mappedContracts := make(map[string]*CoreContract)

	for _, contract := range coreContractsData.ProxyContracts {
		// front-facing proxy contract
		proxyContractAddr := contract.ContractAddress

		c, ok := mappedContracts[proxyContractAddr]
		if !ok {
			c = &CoreContract{
				Address:     proxyContractAddr,
				AbiVersions: make([]string, 0),
			}
			mappedContracts[proxyContractAddr] = c

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

	return mappedContracts, nil
}
