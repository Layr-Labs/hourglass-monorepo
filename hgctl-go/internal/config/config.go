package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

// Define custom types for context keys to avoid collisions
type contextKey string

var (
	ContextKey        contextKey = "currentContext"
	ConfigKey         contextKey = "config"
	EnvKey            contextKey = "env"
	ContractClientKey contextKey = "contractClient"
	LoggerKey         contextKey = "loggerKey"
	KeystoreName      string     = "KEYSTORE_NAME"
	KeystorePassword  string     = "KEYSTORE_PASSWORD"
)

type KeystoreReference struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
	Type string `yaml:"type"`
}

type Web3SignerReference struct {
	Name           string `yaml:"name"`
	ConfigPath     string `yaml:"configPath,omitempty"`
	CACertPath     string `yaml:"caCertPath,omitempty"`
	ClientCertPath string `yaml:"clientCertPath,omitempty"`
	ClientKeyPath  string `yaml:"clientKeyPath,omitempty"`
}

type ContractOverrides struct {
	DelegationManager string `yaml:"delegationManager,omitempty"`
	AllocationManager string `yaml:"allocationManager,omitempty"`
	StrategyManager   string `yaml:"strategyManager,omitempty"`
	KeyRegistrar      string `yaml:"keyRegistrar,omitempty"`
	ReleaseManager    string `yaml:"releaseManager,omitempty"`
}

type SignerConfig struct {
	Type       string `yaml:"type"` // "keystore", "web3signer", or "privatekey"
	Keystore   string `yaml:"keystore,omitempty"`
	PrivateKey string `yaml:"-"` // Never save to disk
	Web3Signer string `yaml:"web3signer,omitempty"`
}

type Context struct {
	Name            string `yaml:"-"`
	ExecutorAddress string `yaml:"executorAddress,omitempty"`
	AVSAddress      string `yaml:"avsAddress,omitempty"`
	OperatorAddress string `yaml:"operatorAddress,omitempty"`
	OperatorSetID   uint32 `yaml:"operatorSetId,omitempty"`
	L1ChainID       uint64 `yaml:"l1ChainId,omitempty"`
	L1RPCUrl        string `yaml:"rpcUrl,omitempty"`

	// Private key for transactions (should be provided via env var or flag)
	PrivateKey string `yaml:"-"`

	// Path to secrets file (e.g., .env.secrets)
	EnvSecretsPath string `yaml:"envSecretsPath"` // Remove omitempty to preserve field

	// Keystore and Web3 Signer references
	Keystores   []KeystoreReference   `yaml:"keystores,omitempty"`
	Web3Signers []Web3SignerReference `yaml:"web3signers,omitempty"`

	// Signer configuration
	Signer *SignerConfig `yaml:"signer,omitempty"`

	// EigenLayer contract addresses (optional - overrides chainId-based lookup)
	ContractOverrides *ContractOverrides `yaml:"contractOverrides,omitempty"`
}

type Config struct {
	CurrentContext string              `yaml:"currentContext"`
	Contexts       map[string]*Context `yaml:"contexts"`
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

type RemoteSignerConfig struct {
	Url         string `json:"url" yaml:"url"`
	CACert      string `json:"caCert" yaml:"caCert"`
	Cert        string `json:"cert" yaml:"cert"`
	Key         string `json:"key" yaml:"key"`
	FromAddress string `json:"fromAddress" yaml:"fromAddress"`
	PublicKey   string `json:"publicKey" yaml:"publicKey"`
}

func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := defaultConfig()
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set context names and initialize nil pointers
	for name, ctx := range config.Contexts {
		ctx.Name = name

		// Initialize ContractOverrides if nil to prevent loss during save
		if ctx.ContractOverrides == nil {
			ctx.ContractOverrides = &ContractOverrides{}
		}
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	configPath := getConfigPath()

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Ensure all contexts have initialized ContractOverrides before saving
	for _, ctx := range config.Contexts {
		if ctx.ContractOverrides == nil {
			ctx.ContractOverrides = &ContractOverrides{}
		}
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func GetCurrentContext() (*Context, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	ctx, ok := cfg.Contexts[cfg.CurrentContext]
	if !ok {
		return nil, fmt.Errorf("current context '%s' not found", cfg.CurrentContext)
	}

	return ctx, nil
}

func GetConfigDir() string {
	// Default to home directory
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hgctl")
}

func getConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.yaml")
}

func defaultConfig() *Config {
	return &Config{
		CurrentContext: "default",
		Contexts: map[string]*Context{
			"default": {
				Name:            "default",
				ExecutorAddress: "127.0.0.1:9090",
			},
		},
	}
}

// ToMap converts the Context to a map for display purposes
func (c *Context) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	// Add optional fields only if they have values
	if c.AVSAddress != "" {
		result["avs-address"] = c.AVSAddress
	}

	// Always show operator-set-id since 0 is a valid value
	result["operator-set-id"] = c.OperatorSetID

	if c.L1ChainID != 0 {
		result["l1-chain-id"] = c.L1ChainID
	}

	if c.L1RPCUrl != "" {
		result["rpc-url"] = c.L1RPCUrl
	}

	if c.ExecutorAddress != "" {
		result["executor-address"] = c.ExecutorAddress
	}

	if c.OperatorAddress != "" {
		result["operator-address"] = c.OperatorAddress
	}

	// Add env secrets path if set
	if c.EnvSecretsPath != "" {
		result["env-secrets-path"] = c.EnvSecretsPath
	}

	// Add keystore references if any
	if len(c.Keystores) > 0 {
		result["keystores"] = c.Keystores
	}

	// Add web3signer references if any
	if len(c.Web3Signers) > 0 {
		result["web3signers"] = c.Web3Signers
	}

	// Add signer key if set
	if c.Signer != nil {
		result["signer-key"] = c.SignerKey
	}

	// Add contract overrides if any
	if c.ContractOverrides != nil {
		overrides := make(map[string]string)
		if c.ContractOverrides.DelegationManager != "" {
			overrides["delegation-manager"] = c.ContractOverrides.DelegationManager
		}
		if c.ContractOverrides.AllocationManager != "" {
			overrides["allocation-manager"] = c.ContractOverrides.AllocationManager
		}
		if c.ContractOverrides.StrategyManager != "" {
			overrides["strategy-manager"] = c.ContractOverrides.StrategyManager
		}
		if c.ContractOverrides.KeyRegistrar != "" {
			overrides["key-registrar"] = c.ContractOverrides.KeyRegistrar
		}
		if c.ContractOverrides.ReleaseManager != "" {
			overrides["release-manager"] = c.ContractOverrides.ReleaseManager
		}
		if len(overrides) > 0 {
			result["contract-overrides"] = overrides
		}
	}

	return result
}
