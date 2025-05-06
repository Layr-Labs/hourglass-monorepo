package contractStore

import "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"

type IContractStore interface {
	GetContractByAddress(address string) (*contracts.Contract, error)
	ListContractAddresses() []string
	ListContracts() []*contracts.Contract
}
