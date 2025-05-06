package eigenlayer

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func filterContractsForChainId(contractList []*contracts.Contract, chainId config.ChainId) []*contracts.Contract {
	return util.Filter(contractList, func(c *contracts.Contract) bool {
		return c.ChainId == chainId
	})
}

func getContractByNameAndChainId(contractList []*contracts.Contract, name string, chainId config.ChainId) *contracts.Contract {
	return util.Find(contractList, func(c *contracts.Contract) bool {
		return c.Name == name && c.ChainId == chainId
	})
}

func Test_CoreContracts(t *testing.T) {

	t.Run("holesky", func(t *testing.T) {
		t.Run("Should load core contracts for holesky", func(t *testing.T) {
			loadedContracts, err := LoadContracts()
			if err != nil {
				t.Fatalf("Failed to load core loadedContracts for holesky: %v", err)
			}

			filteredContracts := filterContractsForChainId(loadedContracts, config.ChainId_EthereumHolesky)
			allocationManager := getContractByNameAndChainId(filteredContracts, config.ContractName_AllocationManager, config.ChainId_EthereumHolesky)

			assert.Equal(t, 3, len(allocationManager.AbiVersions))
		})
		t.Run("Should parse the allocation manager contract to an abi.Abi", func(t *testing.T) {
			loadedContracts, err := LoadContracts()
			if err != nil {
				t.Fatalf("Failed to load core loadedContracts for holesky: %v", err)
			}

			filteredContracts := filterContractsForChainId(loadedContracts, config.ChainId_EthereumHolesky)

			allocationManager := getContractByNameAndChainId(filteredContracts, config.ContractName_AllocationManager, config.ChainId_EthereumHolesky)

			a, err := allocationManager.GetAbi()
			if err != nil {
				t.Fatalf("Failed to get ABI for allocation manager contract: %v", err)
			}
			assert.NotNil(t, a)
		})
	})
	t.Run("mainnet", func(t *testing.T) {
		t.Run("Should load core contracts for mainnet", func(t *testing.T) {
			loadedContracts, err := LoadContracts()
			if err != nil {
				t.Fatalf("Failed to load core loadedContracts for mainnet: %v", err)
			}

			filteredContracts := filterContractsForChainId(loadedContracts, config.ChainId_EthereumMainnet)
			allocationManager := getContractByNameAndChainId(filteredContracts, config.ContractName_AllocationManager, config.ChainId_EthereumMainnet)

			assert.Equal(t, 2, len(allocationManager.AbiVersions))
		})
		t.Run("Should parse the allocation manager contract to an abi.Abi", func(t *testing.T) {
			loadedContracts, err := LoadContracts()
			if err != nil {
				t.Fatalf("Failed to load core loadedContracts for holesky: %v", err)
			}

			filteredContracts := filterContractsForChainId(loadedContracts, config.ChainId_EthereumMainnet)
			allocationManager := getContractByNameAndChainId(filteredContracts, config.ContractName_AllocationManager, config.ChainId_EthereumMainnet)

			a, err := allocationManager.GetAbi()
			if err != nil {
				t.Fatalf("Failed to get ABI for allocation manager contract: %v", err)
			}
			assert.NotNil(t, a)
		})
	})
}
