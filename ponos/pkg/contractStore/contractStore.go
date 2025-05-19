package contractStore

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
)

type IContractStore interface {
	GetContractByAddress(address string) (*contracts.Contract, error)
	ListContractAddressesForChain(chainId config.ChainId) []string
	ListContracts() []*contracts.Contract
	OverrideContract(contractName string, chainIds []config.ChainId, contract *contracts.Contract) error
}
