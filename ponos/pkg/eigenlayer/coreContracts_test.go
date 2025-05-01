package eigenlayer

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CoreContracts(t *testing.T) {

	t.Run("holesky", func(t *testing.T) {
		t.Run("Should load core contracts for holesky", func(t *testing.T) {
			contractAddresses := config.GetContractsMapForChain(config.ChainId_EthereumHolesky)
			contracts, err := LoadCoreContractsForL1Chain(config.ChainId_EthereumHolesky)
			if err != nil {
				t.Fatalf("Failed to load core contracts for holesky: %v", err)
			}

			assert.Equal(t, 3, len(contracts[contractAddresses.AllocationManager].AbiVersions))
		})
		t.Run("Should parse the allocation manager contract to an abi.Abi", func(t *testing.T) {
			contractAddresses := config.GetContractsMapForChain(config.ChainId_EthereumHolesky)
			contracts, err := LoadCoreContractsForL1Chain(config.ChainId_EthereumHolesky)
			if err != nil {
				t.Fatalf("Failed to load core contracts for holesky: %v", err)
			}

			allocationManagerContract := contracts[contractAddresses.AllocationManager]

			a, err := allocationManagerContract.GetAbi()
			if err != nil {
				t.Fatalf("Failed to get ABI for allocation manager contract: %v", err)
			}
			assert.NotNil(t, a)
		})
	})
	t.Run("mainnet", func(t *testing.T) {
		t.Run("Should load core contracts for mainnet", func(t *testing.T) {
			contractAddresses := config.GetContractsMapForChain(config.ChainId_EthereumMainnet)
			contracts, err := LoadCoreContractsForL1Chain(config.ChainId_EthereumMainnet)
			if err != nil {
				t.Fatalf("Failed to load core contracts for holesky: %v", err)
			}

			assert.Equal(t, 2, len(contracts[contractAddresses.AllocationManager].AbiVersions))
		})
		t.Run("Should parse the allocation manager contract to an abi.Abi", func(t *testing.T) {
			contractAddresses := config.GetContractsMapForChain(config.ChainId_EthereumMainnet)
			contracts, err := LoadCoreContractsForL1Chain(config.ChainId_EthereumMainnet)
			if err != nil {
				t.Fatalf("Failed to load core contracts for holesky: %v", err)
			}

			allocationManagerContract := contracts[contractAddresses.AllocationManager]

			a, err := allocationManagerContract.GetAbi()
			if err != nil {
				t.Fatalf("Failed to get ABI for allocation manager contract: %v", err)
			}
			assert.NotNil(t, a)
		})
	})
}
