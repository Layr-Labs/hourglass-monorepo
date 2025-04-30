package config

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	"slices"
	"fmt"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type ChainId uint

const (
	ChainId_EthereumMainnet ChainId = 1
	ChainId_EthereumHolesky ChainId = 17000
	ChainId_EthereumHoodi   ChainId = 560048
	ChainId_EthereumAnvil   ChainId = 31337
)

func IsL1Chain(chainId ChainId) bool {
	return slices.Contains([]ChainId{
		ChainId_EthereumMainnet,
		ChainId_EthereumHolesky,
		ChainId_EthereumHoodi,
		ChainId_EthereumAnvil,
	}, chainId)
}

type CoreContractAddresses struct {
	AllocationManager string
}

var (
	CoreContracts = map[ChainId]CoreContractAddresses{
		ChainId_EthereumMainnet: {
			AllocationManager: "0x948a420b8cc1d6bfd0b6087c2e7c344a2cd0bc39",
		},
		ChainId_EthereumHolesky: {
			AllocationManager: "0x78469728304326cbc65f8f95fa756b0b73164462",
		},
		ChainId_EthereumHoodi: {
			AllocationManager: "",
		},
	}
)

func GetCoreContractsForChainId(chainId ChainId) (*CoreContractAddresses, error) {
	contracts, ok := CoreContracts[chainId]
	if !ok {
		return nil, fmt.Errorf("unsupported chain ID: %d", chainId)
	}
	return &contracts, nil
}

var (
	SupportedChainIds = []ChainId{
		ChainId_EthereumMainnet,
		ChainId_EthereumHolesky,
		ChainId_EthereumHoodi,
		ChainId_EthereumAnvil,
	}
)

type ContractAddresses struct {
	AllocationManager string
	TaskMailbox       string
}

func GetContractsMapForChain(chainId ChainId) *ContractAddresses {
	switch chainId {
	case ChainId_EthereumHolesky:
		return &ContractAddresses{
			AllocationManager: "0x78469728304326cbc65f8f95fa756b0b73164462",
			TaskMailbox:       "0xTaskMailbox",
		}
	case ChainId_EthereumHoodi:
		// TODO(seanmcgary): Add hoodi contracts
		return nil
	case ChainId_EthereumMainnet:
		return &ContractAddresses{
			AllocationManager: "0x948a420b8cc1d6bfd0b6087c2e7c344a2cd0bc39",
			TaskMailbox:       "0xTaskMailbox",
		}
	}
	return nil
}

type OperatorConfig struct {
	Address            string      `json:"address" yaml:"address"`
	OperatorPrivateKey string      `json:"operatorPrivateKey" yaml:"operatorPrivateKey"`
	SigningKeys        SigningKeys `json:"signingKeys" yaml:"signingKeys"`
}

func (oc *OperatorConfig) Validate() error {
	var allErrors field.ErrorList
	if oc.Address == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("address"), "address is required"))
	}
	if oc.OperatorPrivateKey == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("operatorPrivateKey"), "operatorPrivateKey is required"))
	}
	if err := oc.SigningKeys.Validate(); err != nil {
		allErrors = append(allErrors, field.Invalid(field.NewPath("signingKeys"), oc.SigningKeys, err.Error()))
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

type SigningKey struct {
	Keystore string `json:"keystore"`
	Password string `json:"password"`
}

type SigningKeys struct {
	BLS *SigningKey `json:"bls"`
}

func (sk *SigningKeys) Validate() error {
	var allErrors field.ErrorList
	if sk.BLS == nil {
		allErrors = append(allErrors, field.Required(field.NewPath("bls"), "bls is required"))
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

type SimulatedPeer struct {
	NetworkAddress  string `json:"networkAddress" yaml:"networkAddress"`
	PublicKey       string `json:"publicKey" yaml:"publicKey"`
	OperatorAddress string `json:"operatorAddress" yaml:"operatorAddress"`
	OperatorSetId   uint32 `json:"operatorSetId" yaml:"operatorSetId"`
}

type SimulatedPeeringConfig struct {
	Enabled         bool            `json:"enabled" yaml:"enabled"`
	AggregatorPeers []SimulatedPeer `json:"aggregatorPeers" yaml:"aggregatorPeers"`
	OperatorPeers   []SimulatedPeer `json:"operatorPeers" yaml:"operatorPeers"`
}
