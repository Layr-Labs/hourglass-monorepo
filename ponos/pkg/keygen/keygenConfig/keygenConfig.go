package keygenConfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

const (
	EnvPrefix = "KEYGEN_"

	// Debug enables debug logging
	Debug = "debug"
	// CurveType is the curve type (bls381 or bn254)
	CurveType = "curve-type"
	// OutputDir is the directory to write output files to (Private Key, Public Key, Mnemonic)
	OutputDir = "output-dir"
	// FilePrefix is the string that file names will start with (e.g. bn254-0)
	FilePrefix = "file-prefix"
	// KeyFile to use for deriving private keys (e.g. a .pkey file containing a private key)
	KeyFile = "key-file"
	// Seed value to use for random number generation (defaults to system entropy)
	Seed = "seed"
	// Path for deriving keys from seed
	Path = "path"
	// Password to use for encrypting/decrypting private key files
	Password = "password"
	// UseKeystore enables keystore format for private key storage
	UseKeystore = "use-keystore"
	// UseRandomPassword generates a random password for keystore encryption
	UseRandomPassword = "use-random-password"
	// LightEncryption uses lighter encryption parameters for keystore (faster but less secure)
	LightEncryption = "light-encryption"
)

// KeygenConfig encapsulates the configuration for the key generation utility
type KeygenConfig struct {
	Debug             bool   `mapstructure:"debug" yaml:"debug" json:"debug"`
	CurveType         string `mapstructure:"curve_type" yaml:"curveType" json:"curveType"`
	OutputDir         string `mapstructure:"output_dir" yaml:"outputDir" json:"outputDir"`
	FilePrefix        string `mapstructure:"file_prefix" yaml:"filePrefix" json:"filePrefix"`
	KeyFile           string `mapstructure:"key_file" yaml:"keyFile" json:"keyFile"`
	Seed              string `mapstructure:"seed" yaml:"seed" json:"seed"`
	Path              string `mapstructure:"path" yaml:"path" json:"path"`
	Password          string `mapstructure:"password" yaml:"password" json:"password"`
	UseKeystore       bool   `mapstructure:"use_keystore" yaml:"useKeystore" json:"useKeystore"`
	UseRandomPassword bool   `mapstructure:"use_random_password" yaml:"useRandomPassword" json:"useRandomPassword"`
	LightEncryption   bool   `mapstructure:"light_encryption" yaml:"lightEncryption" json:"lightEncryption"`
}

// Validate validates the config required for key generation
func (c *KeygenConfig) Validate() error {
	// For info command, we only need a key file path
	if c.KeyFile != "" {
		return nil
	}

	if c.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	if c.CurveType == "" {
		return fmt.Errorf("curve type is required")
	}

	curveType := strings.ToLower(c.CurveType)
	if curveType != "bls381" && curveType != "bn254" {
		return fmt.Errorf("invalid curve type: must be either 'bls381' or 'bn254'")
	}

	// If using keystore but not using random password, then a password must be provided
	if c.UseKeystore && !c.UseRandomPassword && c.Password == "" {
		return fmt.Errorf("password is required when using keystore without random password generation")
	}

	return nil
}

// LoadConfigFromJSON loads configuration from a JSON file
func LoadConfigFromJSON(filePath string) (*KeygenConfig, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config KeygenConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &config, nil
}

// LoadConfigFromYAML loads configuration from a YAML file
func LoadConfigFromYAML(filePath string) (*KeygenConfig, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config KeygenConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

// NewKeygenConfig creates a new KeygenConfig with values from viper
func NewKeygenConfig() *KeygenConfig {
	return &KeygenConfig{
		Debug:             viper.GetBool(config.NormalizeFlagName(Debug)),
		CurveType:         viper.GetString(config.NormalizeFlagName(CurveType)),
		OutputDir:         viper.GetString(config.NormalizeFlagName(OutputDir)),
		FilePrefix:        viper.GetString(config.NormalizeFlagName(FilePrefix)),
		KeyFile:           viper.GetString(config.NormalizeFlagName(KeyFile)),
		Seed:              viper.GetString(config.NormalizeFlagName(Seed)),
		Path:              viper.GetString(config.NormalizeFlagName(Path)),
		Password:          viper.GetString(config.NormalizeFlagName(Password)),
		UseKeystore:       viper.GetBool(config.NormalizeFlagName(UseKeystore)),
		UseRandomPassword: viper.GetBool(config.NormalizeFlagName(UseRandomPassword)),
		LightEncryption:   viper.GetBool(config.NormalizeFlagName(LightEncryption)),
	}
}

// NewKeygenConfigFromYamlBytes creates a KeygenConfig from YAML bytes
func NewKeygenConfigFromYamlBytes(data []byte) (*KeygenConfig, error) {
	var kc KeygenConfig
	if err := yaml.Unmarshal(data, &kc); err != nil {
		return nil, err
	}
	return &kc, nil
}

// NewKeygenConfigFromJsonBytes creates a KeygenConfig from JSON bytes
func NewKeygenConfigFromJsonBytes(data []byte) (*KeygenConfig, error) {
	var kc KeygenConfig
	if err := json.Unmarshal(data, &kc); err != nil {
		return nil, err
	}
	return &kc, nil
}
