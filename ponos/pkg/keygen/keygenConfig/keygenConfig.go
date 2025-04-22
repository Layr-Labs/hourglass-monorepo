package keygenConfig

import (
	"encoding/json"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
)

const (
	EnvPrefix = "KEYGEN_"

	Debug      = "debug"
	CurveType  = "curve-type"
	OutputDir  = "output-dir"
	FilePrefix = "file-prefix"
	KeyFile    = "key-file"
	Seed       = "seed"
	Path       = "path"
	Password   = "password"
)

// KeygenConfig represents the configuration for the key generation utility
type KeygenConfig struct {
	Debug      bool   `json:"debug" yaml:"debug"`
	CurveType  string `json:"curveType" yaml:"curveType"`
	OutputDir  string `json:"outputDir" yaml:"outputDir"`
	FilePrefix string `json:"filePrefix" yaml:"filePrefix"`
	KeyFile    string `json:"keyFile" yaml:"keyFile"`
	Seed       string `json:"seed" yaml:"seed"`
	Path       string `json:"path" yaml:"path"`
	Password   string `json:"password" yaml:"password"`
}

// Validate ensures that all required fields are set
func (kc *KeygenConfig) Validate() error {
	var allErrors field.ErrorList

	// CurveType must be bls381 or bn254
	if kc.CurveType != "bls381" && kc.CurveType != "bn254" {
		allErrors = append(allErrors, field.Invalid(field.NewPath("curveType"), kc.CurveType, "must be either 'bls381' or 'bn254'"))
	}

	// If we're generating keys, output directory is required
	if kc.KeyFile == "" && kc.OutputDir == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("outputDir"), "outputDir is required for key generation"))
	}

	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

// NewKeygenConfig creates a new KeygenConfig with values from viper
func NewKeygenConfig() *KeygenConfig {
	return &KeygenConfig{
		Debug:      viper.GetBool(config.NormalizeFlagName(Debug)),
		CurveType:  viper.GetString(config.NormalizeFlagName(CurveType)),
		OutputDir:  viper.GetString(config.NormalizeFlagName(OutputDir)),
		FilePrefix: viper.GetString(config.NormalizeFlagName(FilePrefix)),
		KeyFile:    viper.GetString(config.NormalizeFlagName(KeyFile)),
		Seed:       viper.GetString(config.NormalizeFlagName(Seed)),
		Path:       viper.GetString(config.NormalizeFlagName(Path)),
		Password:   viper.GetString(config.NormalizeFlagName(Password)),
	}
}

// NewKeygenConfigFromYamlBytes creates a KeygenConfig from YAML bytes
func NewKeygenConfigFromYamlBytes(data []byte) (*KeygenConfig, error) {
	var kc *KeygenConfig
	if err := yaml.Unmarshal(data, &kc); err != nil {
		return nil, err
	}
	return kc, nil
}

// NewKeygenConfigFromJsonBytes creates a KeygenConfig from JSON bytes
func NewKeygenConfigFromJsonBytes(data []byte) (*KeygenConfig, error) {
	var kc *KeygenConfig
	if err := json.Unmarshal(data, &kc); err != nil {
		return nil, err
	}
	return kc, nil
}
