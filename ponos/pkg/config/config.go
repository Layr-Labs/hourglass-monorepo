package config

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"slices"
)

type CurveType string

func (c CurveType) String() string {
	return string(c)
}
func (c CurveType) Uint8() (uint8, error) {
	return ConvertCurveTypeToSolidityEnum(c)
}

const (
	CurveTypeUnknown CurveType = "unknown"
	CurveTypeECDSA   CurveType = "ecdsa"
	CurveTypeBN254   CurveType = "bn254" // BN254 is the only supported curve type for now
)

func ConvertCurveTypeToSolidityEnum(curveType CurveType) (uint8, error) {
	switch curveType {
	case CurveTypeUnknown:
		return 0, nil
	case CurveTypeECDSA:
		return 1, nil
	case CurveTypeBN254:
		return 2, nil
	default:
		return 0, fmt.Errorf("unsupported curve type: %s", curveType)
	}
}

func ConvertSolidityEnumToCurveType(enumValue uint8) (CurveType, error) {
	switch enumValue {
	case 0:
		return CurveTypeUnknown, nil
	case 1:
		return CurveTypeECDSA, nil
	case 2:
		return CurveTypeBN254, nil
	default:
		return "", fmt.Errorf("unsupported curve type enum value: %d", enumValue)
	}
}

type ChainId uint

const (
	ChainId_EthereumMainnet  ChainId = 1
	ChainId_EthereumHolesky  ChainId = 17000
	ChainId_EthereumHoodi    ChainId = 560048
	ChainId_EthereumAnvil    ChainId = 31337
	ChainId_BaseSepoliaAnvil ChainId = 31338
)

const (
	ContractName_AllocationManager  = "AllocationManager"
	ContractName_TaskMailbox        = "TaskMailbox"
	ContractName_KeyRegistrar       = "KeyRegistrar"
	ContractName_CrossChainRegistry = "CrossChainRegistry"
)

const (
	AVSRegistrarSimulationAddress = "0xf4c5c29b14f0237131f7510a51684c8191f98e06"
)

var EthereumSimulationContracts = CoreContractAddresses{
	AllocationManager: "0x42583067658071247ec8CE0A516A58f682002d07",
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
	AllocationManager        string
	DelegationManager        string
	TaskMailbox              string
	KeyRegistrar             string
	CrossChainRegistry       string
	ECDSACertificateVerifier string
	BN254CertificateVerifier string
}

var (
	CoreContracts = map[ChainId]*CoreContractAddresses{
		ChainId_EthereumMainnet: {
			AllocationManager: "0x42583067658071247ec8CE0A516A58f682002d07",
			DelegationManager: "0xD4A7E1Bd8015057293f0D0A557088c286942e84b",
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
		// fork of ethereum sepolia
		ChainId_EthereumAnvil: {
			AllocationManager:        "0x42583067658071247ec8ce0a516a58f682002d07",
			DelegationManager:        "0xd4a7e1bd8015057293f0d0a557088c286942e84b",
			TaskMailbox:              "0x203ca0e6b9bce319937ab44660f3854c41f3331f",
			KeyRegistrar:             "0x78de554ac8dff368e3caa73b3df8acccfd92928a",
			CrossChainRegistry:       "0xe850d8a178777b483d37fd492a476e3e6004c816",
			ECDSACertificateVerifier: "0xad2f58a551bd0e77fa20b5531da96ef440c392bf",
			BN254CertificateVerifier: "0x998535833f3fee44ce720440e735554699f728a5",
		},
		ChainId_BaseSepoliaAnvil: {
			TaskMailbox:              "0x203ca0e6b9bce319937ab44660f3854c41f3331f",
			ECDSACertificateVerifier: "0xad2f58a551bd0e77fa20b5531da96ef440c392bf",
			BN254CertificateVerifier: "0x998535833f3fee44ce720440e735554699f728a5",
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
		ChainId_BaseSepoliaAnvil,
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
	Address string `json:"address" yaml:"address"`
	// OperatorPrivateKey is the private key of the operator used for signing transactions
	OperatorPrivateKey string `json:"operatorPrivateKey" yaml:"operatorPrivateKey"`
	// SigningKeys contains the signing keys for the operator (BLS or ECDSA)
	SigningKeys SigningKeys `json:"signingKeys" yaml:"signingKeys"`
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

// SigningKey represents the signing key configuration for the operator.
// Order of precedence for signing keys: keystore string, keystore file
type SigningKey struct {
	Keystore     string `json:"keystore"`
	KeystoreFile string `json:"keystoreFile"`
	Password     string `json:"password"`
}

func (sk *SigningKey) Validate() error {
	var allErrors field.ErrorList
	if sk.Keystore == "" && sk.KeystoreFile == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("keystore"), "keystore or keystoreFile is required"))
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

type SigningKeys struct {
	BLS   *SigningKey `json:"bls"`
	ECDSA string      `json:"ecdsa"`
}

func (sk *SigningKeys) Validate() error {
	var allErrors field.ErrorList
	if sk.BLS == nil {
		allErrors = append(allErrors, field.Required(field.NewPath("bls"), "bls is required"))
	}
	if err := sk.BLS.Validate(); err != nil {
		allErrors = append(allErrors, field.Invalid(field.NewPath("bls"), sk.BLS, err.Error()))
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

type OverrideContract struct {
	Contract string    `json:"contract" yaml:"contract"`
	ChainIds []ChainId `json:"chainIds" yaml:"chainIds"`
}

type OverrideContracts struct {
	TaskMailbox *OverrideContract `json:"taskMailbox" yaml:"taskMailbox"`
}
