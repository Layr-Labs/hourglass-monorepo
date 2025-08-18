package config

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
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

type ContractOverrides struct {
	DelegationManager string `yaml:"delegationManager,omitempty"`
	AllocationManager string `yaml:"allocationManager,omitempty"`
	StrategyManager   string `yaml:"strategyManager,omitempty"`
	KeyRegistrar      string `yaml:"keyRegistrar,omitempty"`
	ReleaseManager    string `yaml:"releaseManager,omitempty"`
}

type Context struct {
	Name            string `yaml:"-"`
	ExecutorAddress string `yaml:"executorAddress,omitempty"`
	AVSAddress      string `yaml:"avsAddress,omitempty"`
	OperatorAddress string `yaml:"operatorAddress,omitempty"`
	OperatorSetID   uint32 `yaml:"operatorSetId,omitempty"`
	L1ChainID       uint32 `yaml:"l1ChainId,omitempty"`
	L1RPCUrl        string `yaml:"l1RpcUrl,omitempty"`
	L2ChainID       uint32 `yaml:"l2ChainId,omitempty"`
	L2RPCUrl        string `yaml:"l2RpcUrl,omitempty"`

	// Private key for transactions (should be provided via env var or flag)
	PrivateKey string `yaml:"-"`

	// Path to secrets file (e.g., .env.secrets)
	EnvSecretsPath string `yaml:"envSecretsPath"` // Remove omitempty to preserve field

	// Keystore and Web3 Signer references
	Keystores []signer.KeystoreReference `yaml:"keystores,omitempty"`

	// Signing keys
	SystemSignerKeys *signer.SigningKeys `yaml:"systemSigner,omitempty"`

	// Operator Keys
	OperatorKeys *signer.ECDSAKeyConfig `yaml:"operatorSigner,omitempty"`

	// EigenLayer contract addresses (optional - overrides chainId-based lookup)
	ContractOverrides *ContractOverrides `yaml:"contractOverrides,omitempty"`

	// Experimental features flag
	Experimental bool `yaml:"experimental,omitempty"`
}

type Config struct {
	CurrentContext string              `yaml:"currentContext"`
	Contexts       map[string]*Context `yaml:"contexts"`
}

// OperatorSignerFromContext loads the operator key signer from context
func OperatorSignerFromContext(ctx *Context, l logger.Logger) (signer.ISigner, error) {
	if ctx == nil || ctx.OperatorKeys == nil {
		return nil, fmt.Errorf("no operator signing keys configured -- please use `hgctl signer` and follow the wizard to setup")
	}

	opKeys := ctx.OperatorKeys
	if opKeys.Keystore != nil {
		return signer.LoadKeystoreSigner(opKeys.Keystore)
	}

	if opKeys.RemoteSignerConfig != nil {
		web3SignerConfig, err := signer.LoadWeb3SignerConfig(opKeys.RemoteSignerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load web3 signer config: %w", err)
		}
		return client.LoadWeb3Signer(web3SignerConfig, l)
	}

	if opKeys.PrivateKey != false {
		return signer.LoadPrivateKeySigner()
	}

	return nil, fmt.Errorf("operator signing keys not found in context")
}

// SystemSignerFromContext loads the operator key signer from context
func SystemSignerFromContext(ctx *Context, l logger.Logger) (signer.ISigner, error) {
	if ctx == nil || ctx.OperatorKeys == nil {
		return nil, fmt.Errorf("no operator signing keys found in context")
	}

	opKeys := ctx.OperatorKeys
	if opKeys.Keystore != nil {
		return signer.LoadKeystoreSigner(opKeys.Keystore)
	}

	if opKeys.RemoteSignerConfig != nil {
		web3SignerConfig, err := signer.LoadWeb3SignerConfig(opKeys.RemoteSignerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load web3 signer config: %w", err)
		}
		return client.LoadWeb3Signer(web3SignerConfig, l)
	}

	if opKeys.PrivateKey != false {
		return signer.LoadPrivateKeySigner()
	}

	return nil, fmt.Errorf("operator signing keys not found in context")
}

// LoggerFromContext retrieves the logger from the context
func LoggerFromContext(ctx context.Context) logger.Logger {
	if l, ok := ctx.Value(LoggerKey).(logger.Logger); ok {
		return l
	}
	return logger.GetLogger()
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
		result["l1-rpc-url"] = c.L1RPCUrl
	}

	if c.L2ChainID != 0 {
		result["l2-chain-id"] = c.L2ChainID
	}

	if c.L2RPCUrl != "" {
		result["l2-rpc-url"] = c.L2RPCUrl
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

	// Add signer key if set
	if c.OperatorKeys != nil {
		result["operator-key"] = c.OperatorKeys
	}

	// Add signer key if set
	if c.SystemSignerKeys != nil {
		result["system-key"] = c.SystemSignerKeys
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
