package contracts

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type ContractStore struct {
	addresses map[string]string
}

func NewContractStore() (*ContractStore, error) {
	cs := &ContractStore{
		addresses: make(map[string]string),
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			if value != "" && len(value) >= 40 {
				cs.addresses[key] = value
			}
		}
	}

	return cs, nil
}

func (cs *ContractStore) GetContract(envVarName string) (common.Address, error) {
	upperName := strings.ToUpper(envVarName)
	return cs.getAddress(upperName)
}

func (cs *ContractStore) GetTaskAVSRegistrar() (common.Address, error) {
	return cs.getAddress("TASKAVSREGISTRAR")
}

func (cs *ContractStore) GetTaskMailbox() (common.Address, error) {
	return cs.getAddress("TASKMAILBOX")
}

func (cs *ContractStore) ListContracts() []string {
	var contracts []string
	for key, value := range cs.addresses {
		if value != "" && common.IsHexAddress(value) {
			contracts = append(contracts, key)
		}
	}
	return contracts
}

func (cs *ContractStore) getAddress(envVarName string) (common.Address, error) {
	value, exists := cs.addresses[envVarName]
	if !exists || value == "" {
		return common.Address{}, fmt.Errorf("contract %s not found in environment", envVarName)
	}

	if !common.IsHexAddress(value) {
		return common.Address{}, fmt.Errorf("invalid address for %s: %s", envVarName, value)
	}

	return common.HexToAddress(value), nil
}