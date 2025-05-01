package inMemoryContractStore

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"go.uber.org/zap"
)

type InMemoryContractStore struct {
	contracts map[string]*contracts.Contract
	logger    *zap.Logger
}

func NewInMemoryContractStore(contracts map[string]*contracts.Contract, logger *zap.Logger) *InMemoryContractStore {
	return &InMemoryContractStore{
		contracts: contracts,
		logger:    logger,
	}
}

func (ics *InMemoryContractStore) GetContractByAddress(address string) (*contracts.Contract, error) {
	contract, ok := ics.contracts[address]
	if !ok {
		ics.logger.Error("Contract not found", zap.String("address", address))
		return nil, nil
	}
	return contract, nil
}

func (ics *InMemoryContractStore) ListContractAddresses() []string {
	addresses := make([]string, 0, len(ics.contracts))
	for address := range ics.contracts {
		addresses = append(addresses, address)
	}
	return addresses
}
