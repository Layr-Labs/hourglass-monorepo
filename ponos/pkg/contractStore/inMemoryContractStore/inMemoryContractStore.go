package inMemoryContractStore

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
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

func (ics *InMemoryContractStore) ListContractAddressesForChain(chainId config.ChainId) []string {
	// use a map to make sure we're getting unique addresses and no duplicates
	chainContracts := util.Reduce(ics.contracts, func(acc map[string]*contracts.Contract, c *contracts.Contract) map[string]*contracts.Contract {
		if c.ChainId == chainId {
			acc[strings.ToLower(c.Address)] = c
		}
		return acc
	}, make(map[string]*contracts.Contract))
	addresses := make([]string, 0, len(chainContracts))
	for a, _ := range chainContracts {
		addresses = append(addresses, a)
	}
	return addresses
}

func (ics *InMemoryContractStore) ListContracts() []*contracts.Contract {
	return ics.contracts
}
