package contracts

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

const (
	ContractL1ProxyAdmin         = "L1PROXYADMIN"
	ContractTaskAVSRegistrar     = "TASKAVSREGISTRAR"
	ContractTaskAVSRegistrarImpl = "TASKAVSREGISTRARIMPL"

	ContractL2ProxyAdmin = "L2PROXYADMIN"
	ContractAVSTaskHook  = "AVSTASKHOOK"
	ContractTaskMailbox  = "TASKMAILBOX"
)

type ContractStore struct {
	addresses map[string]string
}

func NewContractStore() (*ContractStore, error) {
	store := &ContractStore{
		addresses: make(map[string]string),
	}

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}

		key := pair[0]
		value := pair[1]

		if common.IsHexAddress(value) {
			store.addresses[key] = value
		}
	}

	return store, nil
}

func (cs *ContractStore) GetL1ProxyAdmin() (common.Address, error) {
	return cs.getAddress(ContractL1ProxyAdmin)
}

func (cs *ContractStore) GetTaskAVSRegistrar() (common.Address, error) {
	return cs.getAddress(ContractTaskAVSRegistrar)
}

func (cs *ContractStore) GetTaskAVSRegistrarImpl() (common.Address, error) {
	return cs.getAddress(ContractTaskAVSRegistrarImpl)
}

func (cs *ContractStore) GetL2ProxyAdmin() (common.Address, error) {
	return cs.getAddress(ContractL2ProxyAdmin)
}

func (cs *ContractStore) GetAVSTaskHook() (common.Address, error) {
	return cs.getAddress(ContractAVSTaskHook)
}

func (cs *ContractStore) GetTaskMailbox() (common.Address, error) {
	return cs.getAddress(ContractTaskMailbox)
}

func (cs *ContractStore) GetContract(envVarName string) (common.Address, error) {
	upperName := strings.ToUpper(envVarName)
	return cs.getAddress(upperName)
}

func (cs *ContractStore) getAddress(name string) (common.Address, error) {
	address, exists := cs.addresses[name]
	if !exists {
		return common.Address{}, fmt.Errorf("contract %s not found in environment", name)
	}
	return common.HexToAddress(address), nil
}

func (cs *ContractStore) ListContracts() []string {
	names := make([]string, 0, len(cs.addresses))
	for name := range cs.addresses {
		names = append(names, name)
	}
	return names
}
