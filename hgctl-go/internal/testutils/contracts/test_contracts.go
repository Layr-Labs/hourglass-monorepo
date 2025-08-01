package contracts

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
)

// GetTestChainCoreContracts returns the core contract addresses for test chains
// These addresses are either deployed by our test setup or are mock addresses
func GetTestChainCoreContracts(chainID config.ChainId) *config.CoreContractAddresses {
	switch chainID {
	case 31337: // L1 test chain
		return &config.CoreContractAddresses{
			// These are addresses from our test deployment
			// We'll use the KeyRegistrar address from our chain config
			KeyRegistrar: "0xa4db30d08d8bbca00d40600bee9f029984db162a",
			// For now, we'll leave the EigenLayer contracts empty
			// since they're not deployed in our test environment
			AllocationManager: "",
			DelegationManager: "",
			// We can add TaskMailbox and other contracts as needed
			TaskMailbox: "",
		}
	case 31338: // L2 test chain
		return &config.CoreContractAddresses{
			// L2 specific contracts
			TaskMailbox: "",
		}
	default:
		return nil
	}
}