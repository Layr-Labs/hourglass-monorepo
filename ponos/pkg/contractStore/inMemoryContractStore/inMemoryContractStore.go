package inMemoryContractStore

import (
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"go.uber.org/zap"
	"strings"
)

type InMemoryContractStore struct {
	contracts []*contracts.Contract
	logger    *zap.Logger
}

func NewInMemoryContractStore(contracts []*contracts.Contract, logger *zap.Logger) *InMemoryContractStore {
	fmt.Printf("contracts: %+v\n", contracts)
	return &InMemoryContractStore{
		contracts: contracts,
		logger:    logger,
	}
}

// TODO(seanmcgary): take a chain ID as an argument to increase specificity
func (ics *InMemoryContractStore) GetContractByAddress(address string) (*contracts.Contract, error) {
	address = strings.ToLower(address)

	contract := util.Find(ics.contracts, func(c *contracts.Contract) bool {
		return strings.EqualFold(c.Address, address)
	})

	if contract == nil {
		ics.logger.Error("Contract not found", zap.String("address", address))
		return nil, nil
	}
	return contract, nil
}

func (ics *InMemoryContractStore) ListContractAddresses() []string {
	return util.Map(ics.contracts, func(c *contracts.Contract, i uint64) string {
		return strings.ToLower(c.Address)
	})
}

func (ics *InMemoryContractStore) ListContracts() []*contracts.Contract {
	return ics.contracts
}
