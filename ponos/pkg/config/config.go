package config

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"slices"
)

type ChainId uint

const (
	ChainId_EthereumMainnet ChainId = 1
	ChainId_EthereumHolesky ChainId = 17000
	ChainId_EthereumHoodi   ChainId = 560048
	ChainId_EthereumAnvil   ChainId = 31337
)

const (
	ContractName_AllocationManager = "AllocationManager"
	ContractName_TaskMailbox       = "TaskMailbox"
)

const (
	AVSRegistrarSimulationAddress = "0xf4c5c29b14f0237131f7510a51684c8191f98e06"
)

var EthereumSimulationContracts = CoreContractAddresses{
	AllocationManager: "0x948a420b8cc1d6bfd0b6087c2e7c344a2cd0bc39",
	TaskMailbox:       "0x7306a649b451ae08781108445425bd4e8acf1e00",
}

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
	DelegationManager string
	TaskMailbox       string
}

var (
	CoreContracts = map[ChainId]*CoreContractAddresses{
		ChainId_EthereumMainnet: {
			AllocationManager: "0x948a420b8cc1d6bfd0b6087c2e7c344a2cd0bc39",
			DelegationManager: "0x39053d51b77dc0d36036fc1fcc8cb819df8ef37a",
			TaskMailbox:       "0x7306a649b451ae08781108445425bd4e8acf1e00",
		},
		ChainId_EthereumHolesky: {
			AllocationManager: "0x78469728304326cbc65f8f95fa756b0b73164462",
			DelegationManager: "0xa44151489861fe9e3055d95adc98fbd462b948e7",
			TaskMailbox:       "0xtaskMailbox",
		},
		ChainId_EthereumHoodi: {
			AllocationManager: "",
			DelegationManager: "",
			TaskMailbox:       "0xtaskMailbox",
		},
		ChainId_EthereumAnvil: {
			AllocationManager: "0x948a420b8cc1d6bfd0b6087c2e7c344a2cd0bc39",
			DelegationManager: "0x39053d51b77dc0d36036fc1fcc8cb819df8ef37a",
			TaskMailbox:       "0x7306a649b451ae08781108445425bd4e8acf1e00",
		},
	}
)

func GetCoreContractsForChainId(chainId ChainId) (*CoreContractAddresses, error) {
	contracts, ok := CoreContracts[chainId]
	if !ok {
		return nil, fmt.Errorf("unsupported chain ID: %d", chainId)
	}
	return contracts, nil
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

func GetContractsMapForChain(chainId ChainId) *CoreContractAddresses {
	contracts, ok := CoreContracts[chainId]
	if !ok {
		return nil
	}
	return contracts
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
