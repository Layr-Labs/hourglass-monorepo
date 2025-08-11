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
	ChainId_EthereumSepolia  ChainId = 11155111
	ChainId_BaseSepolia      ChainId = 84532
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
		ChainId_EthereumSepolia,
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
	ethereumSepoliaCoreContracts = &CoreContractAddresses{
		AllocationManager:        "0x42583067658071247ec8ce0a516a58f682002d07",
		DelegationManager:        "0xd4a7e1bd8015057293f0d0a557088c286942e84b",
		TaskMailbox:              "0xb99cc53e8db7018f557606c2a5b066527bf96b26",
		KeyRegistrar:             "0xa4db30d08d8bbca00d40600bee9f029984db162a",
		CrossChainRegistry:       "0x287381b1570d9048c4b4c7ec94d21ddb8aa1352a",
		ECDSACertificateVerifier: "0xb3cd1a457dea9a9a6f6406c6419b1c326670a96f",
		BN254CertificateVerifier: "0xff58a373c18268f483c1f5ca03cf885c0c43373a",
	}
	baseSepoliaCoreContracts = &CoreContractAddresses{
		TaskMailbox:              "0xb99cc53e8db7018f557606c2a5b066527bf96b26",
		ECDSACertificateVerifier: "0xb3cd1a457dea9a9a6f6406c6419b1c326670a96f",
		BN254CertificateVerifier: "0xff58a373c18268f483c1f5ca03cf885c0c43373a",
	}

	CoreContracts = map[ChainId]*CoreContractAddresses{
		ChainId_EthereumMainnet: {
			AllocationManager: "0x42583067658071247ec8CE0A516A58f682002d07",
			DelegationManager: "0xD4A7E1Bd8015057293f0D0A557088c286942e84b",
		},
		ChainId_EthereumHolesky: {
			AllocationManager: "0x78469728304326cbc65f8f95fa756b0b73164462",
			DelegationManager: "0xa44151489861fe9e3055d95adc98fbd462b948e7",
		},
		ChainId_EthereumHoodi: {
			AllocationManager: "",
			DelegationManager: "",
		},
		ChainId_EthereumSepolia:  ethereumSepoliaCoreContracts,
		ChainId_BaseSepolia:      baseSepoliaCoreContracts,
		ChainId_EthereumAnvil:    ethereumSepoliaCoreContracts, // fork of ethereum sepolia
		ChainId_BaseSepoliaAnvil: baseSepoliaCoreContracts,     // fork of base sepolia
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
	// OperatorPrivateKey is the private key of the operator used for signing transactions.
	OperatorPrivateKey *ECDSAKeyConfig `json:"operatorPrivateKey" yaml:"operatorPrivateKey"`
	// SigningKeys contains the signing keys for the operator (BLS or ECDSA)
	SigningKeys SigningKeys `json:"signingKeys" yaml:"signingKeys"`
}

func (oc *OperatorConfig) Validate() error {
	var allErrors field.ErrorList
	if oc.Address == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("address"), "address is required"))
	}
	if oc.OperatorPrivateKey == nil {
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

type RemoteSignerConfig struct {
	Url         string `json:"url" yaml:"url"`
	CACert      string `json:"caCert" yaml:"caCert"`
	Cert        string `json:"cert" yaml:"cert"`
	Key         string `json:"key" yaml:"key"`
	FromAddress string `json:"fromAddress" yaml:"fromAddress"`
	PublicKey   string `json:"publicKey" yaml:"publicKey"`
}

func (rsc *RemoteSignerConfig) Validate() error {
	var allErrors field.ErrorList
	if rsc.FromAddress == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("fromAddress"), "fromAddress is required"))
	}
	if rsc.PublicKey == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("publicKey"), "publicKey is required"))
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

type ECDSAKeyConfig struct {
	UseRemoteSigner    bool                `json:"remoteSigner" yaml:"remoteSigner"`
	RemoteSignerConfig *RemoteSignerConfig `json:"remoteSignerConfig" yaml:"remoteSignerConfig"`
	PrivateKey         string              `json:"privateKey" yaml:"privateKey"`
}

func (ekc *ECDSAKeyConfig) Validate() error {
	var allErrors field.ErrorList
	if ekc.UseRemoteSigner {
		if ekc.RemoteSignerConfig == nil {
			allErrors = append(allErrors, field.Required(field.NewPath("remoteSignerConfig"), "remoteSignerConfig is required when UseRemoteSigner is true"))
		} else if err := ekc.RemoteSignerConfig.Validate(); err != nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("remoteSignerConfig"), ekc.RemoteSignerConfig, err.Error()))
		}
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

type SigningKeys struct {
	BLS   *SigningKey     `json:"bls"`
	ECDSA *ECDSAKeyConfig `json:"ecdsa"`
}

func (sk *SigningKeys) Validate() error {
	var allErrors field.ErrorList
	if sk.BLS == nil && sk.ECDSA == nil {
		allErrors = append(allErrors, field.Required(field.NewPath("bls"), "at least one signing key (BLS or ECDSA) is required"))
	}
	if sk.BLS != nil {
		if err := sk.BLS.Validate(); err != nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("bls"), sk.BLS, err.Error()))
		}
	}
	if sk.ECDSA != nil {
		if err := sk.ECDSA.Validate(); err != nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("ecdsa"), sk.ECDSA, err.Error()))
		}
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

type KubernetesEnv struct {
	ValueFrom struct {
		SecretKeyRef struct {
			Name string `json:"name" yaml:"name"` // Name of the secret
			Key  string `json:"key" yaml:"key"`   // Key within the secret
		} `json:"secretKeyRef" yaml:"secretKeyRef"` // SecretKeyRef is used to reference a key in a Kubernetes secret
		ConfigMapKeyRef struct {
			Name string `json:"name" yaml:"name"` // Name of the config map
			Key  string `json:"key" yaml:"key"`   // Key within the config map
		} `json:"configMapKeyRef" yaml:"configMapKeyRef"` // ConfigMapKeyRef is used to reference a key in a Kubernetes config map

	} `json:"valueFrom" yaml:"valueFrom"` // ValueFrom is used to reference a Kubernetes secret
}

type AVSPerformerEnv struct {
	Name          string
	Value         string // Value is a direct value, passed to the Executor and forwarded
	ValueFromEnv  string // ValueFromEnv is the name of an environment variable that should be forwarded to the Performer
	KubernetesEnv *KubernetesEnv
}

func (a *AVSPerformerEnv) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}
