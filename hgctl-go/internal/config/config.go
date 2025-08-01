package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ContextKey = "context"
	EnvKey     = "env"
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

type Context struct {
	Name                  string `yaml:"-"`
	ExecutorAddress       string `yaml:"executorAddress"`
	AVSAddress            string `yaml:"avsAddress,omitempty"`
	OperatorAddress       string `yaml:"operatorAddress,omitempty"`
	OperatorSetID         uint32 `yaml:"operatorSetId,omitempty"`
	NetworkID             uint64 `yaml:"networkId,omitempty"`
	RPCUrl                string `yaml:"rpcUrl,omitempty"`

	// Private key for transactions (should be provided via env var or flag)
	PrivateKey string `yaml:"-"`

	// Environment variables for deployments (non-secret values only)
	// Secrets should be provided at runtime via flags or environment variables
	EnvironmentVars map[string]string `yaml:"environmentVars,omitempty"`

	// Keystore and Web3 Signer references
	Keystores   []KeystoreReference   `yaml:"keystores,omitempty"`
	Web3Signers []Web3SignerReference `yaml:"web3signers,omitempty"`

	// EigenLayer contract addresses (optional - overrides chainId-based lookup)
	ContractOverrides *ContractOverrides `yaml:"contractOverrides,omitempty"`
}

type Config struct {
	CurrentContext string              `yaml:"context"`
	Contexts       map[string]*Context `yaml:"contexts"`
}

// Global config instance
var globalConfig *Config

func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := defaultConfig()
			globalConfig = cfg
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set context names
	for name, ctx := range config.Contexts {
		ctx.Name = name
	}

	globalConfig = &config
	return &config, nil
}

func SaveConfig(config *Config) error {
	configPath := getConfigPath()

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	globalConfig = config
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
	// Check for environment variable first
	if configDir := os.Getenv("HGCTL_CONFIG_DIR"); configDir != "" {
		return configDir
	}
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

	// Add basic fields
	result["executor-address"] = c.ExecutorAddress

	// Add optional fields only if they have values
	if c.AVSAddress != "" {
		result["avs-address"] = c.AVSAddress
	}
	if c.OperatorAddress != "" {
		result["operator-address"] = c.OperatorAddress
	}
	if c.OperatorSetID != 0 {
		result["operator-set-id"] = c.OperatorSetID
	}
	if c.NetworkID != 0 {
		result["network-id"] = c.NetworkID
	}
	if c.RPCUrl != "" {
		result["rpc-url"] = c.RPCUrl
	}

	// Add environment variables if any
	if len(c.EnvironmentVars) > 0 {
		result[EnvKey] = c.EnvironmentVars
	}

	// Add keystore references if any
	if len(c.Keystores) > 0 {
		result["keystores"] = c.Keystores
	}

	// Add web3signer references if any
	if len(c.Web3Signers) > 0 {
		result["web3signers"] = c.Web3Signers
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

// IsSecretVariable checks if an environment variable name indicates it contains a secret
func IsSecretVariable(name string) bool {
	// List of patterns that indicate secret variables
	secretPatterns := []string{
		"PRIVATE_KEY",
		"PASSWORD",
		"SECRET",
		"TOKEN",
		"API_KEY",
		"MNEMONIC",
		"SEED",
	}

	upperName := strings.ToUpper(name)
	for _, pattern := range secretPatterns {
		if strings.Contains(upperName, pattern) {
			return true
		}
	}

	return false
}
