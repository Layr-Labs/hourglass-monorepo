package config

import (
	"fmt"
	"slices"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

type TableCalculatorAddresses struct {
	BN254 string
	ECDSA string
}

var TableCalculatorsByChain = map[ChainId]*TableCalculatorAddresses{
	ChainId_EthereumMainnet: {
		BN254: "0x55F4b21681977F412B318eCB204cB933bD1dF57c",
		ECDSA: "0xA933CB4cbD0C4C208305917f56e0C3f51ad713Fa",
	},
	ChainId_BaseMainnet: {
		BN254: "0x55F4b21681977F412B318eCB204cB933bD1dF57c",
		ECDSA: "0xA933CB4cbD0C4C208305917f56e0C3f51ad713Fa",
	},
	ChainId_EthereumAnvil: {
		BN254: "0x55F4b21681977F412B318eCB204cB933bD1dF57c",
		ECDSA: "0xA933CB4cbD0C4C208305917f56e0C3f51ad713Fa",
	},
	ChainId_BaseAnvil: {
		BN254: "0x55F4b21681977F412B318eCB204cB933bD1dF57c",
		ECDSA: "0xA933CB4cbD0C4C208305917f56e0C3f51ad713Fa",
	},
	ChainId_EthereumSepolia: {
		BN254: "0x797d076aB96a5d4104062C15c727447fD8b71eB0",
		ECDSA: "0xbcff2Cb40eD4A80e3A9EB095840986F9c8395a38",
	},
	ChainId_BaseSepolia: {
		BN254: "0x797d076aB96a5d4104062C15c727447fD8b71eB0",
		ECDSA: "0xbcff2Cb40eD4A80e3A9EB095840986F9c8395a38",
	},
}

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
	ChainId_EthereumMainnet ChainId = 1
	ChainId_EthereumHolesky ChainId = 17000
	ChainId_EthereumHoodi   ChainId = 560048
	ChainId_EthereumSepolia ChainId = 11155111
	ChainId_BaseSepolia     ChainId = 84532
	ChainId_BaseMainnet     ChainId = 8453
	ChainId_EthereumAnvil   ChainId = 31337
	ChainId_BaseAnvil       ChainId = 31338
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
	ReleaseManager           string
	TaskMailbox              string
	KeyRegistrar             string
	CrossChainRegistry       string
	ECDSACertificateVerifier string
	BN254CertificateVerifier string
}

var (
	ethereumMainnetCoreContracts = &CoreContractAddresses{
		AllocationManager:        "0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39",
		DelegationManager:        "0x39053D51B77DC0d36036Fc1fCc8Cb819df8Ef37A",
		ReleaseManager:           "0xeDA3CAd031c0cf367cF3f517Ee0DC98F9bA80C8F",
		TaskMailbox:              "0x132b466d9d5723531F68797519DfED701aC2C749",
		KeyRegistrar:             "0x54f4bC6bDEbe479173a2bbDc31dD7178408A57A4",
		CrossChainRegistry:       "0x9376A5863F2193cdE13e1aB7c678F22554E2Ea2b",
		ECDSACertificateVerifier: "0xd0930ee96D07de4F9d493c259232222e46B6EC25",
		BN254CertificateVerifier: "0x3F55654b2b2b86bB11bE2f72657f9C33bf88120A",
	}
	baseMainnetCoreContracts = &CoreContractAddresses{
		TaskMailbox:              "0x132b466d9d5723531F68797519DfED701aC2C749",
		ECDSACertificateVerifier: "0xd0930ee96D07de4F9d493c259232222e46B6EC25",
		BN254CertificateVerifier: "0x3F55654b2b2b86bB11bE2f72657f9C33bf88120A",
	}
	ethereumSepoliaCoreContracts = &CoreContractAddresses{
		AllocationManager:        "0x42583067658071247ec8ce0a516a58f682002d07",
		DelegationManager:        "0xd4a7e1bd8015057293f0d0a557088c286942e84b",
		ReleaseManager:           "0x59c8D715DCa616e032B744a753C017c9f3E16bf4",
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
		ChainId_EthereumMainnet: ethereumMainnetCoreContracts,
		ChainId_BaseMainnet:     baseMainnetCoreContracts,
		ChainId_EthereumHolesky: {
			AllocationManager: "0x78469728304326cbc65f8f95fa756b0b73164462",
			DelegationManager: "0xa44151489861fe9e3055d95adc98fbd462b948e7",
		},
		ChainId_EthereumHoodi: {
			AllocationManager: "",
			DelegationManager: "",
		},
		ChainId_EthereumSepolia: ethereumSepoliaCoreContracts,
		ChainId_BaseSepolia:     baseSepoliaCoreContracts,
		ChainId_EthereumAnvil:   ethereumMainnetCoreContracts,
		ChainId_BaseAnvil:       baseMainnetCoreContracts,
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
		ChainId_BaseMainnet,
		ChainId_EthereumHolesky,
		ChainId_EthereumHoodi,
		ChainId_EthereumSepolia,
		ChainId_BaseSepolia,
		ChainId_EthereumAnvil,
		ChainId_BaseAnvil,
	}
)

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
