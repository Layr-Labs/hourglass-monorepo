package keygenConfig

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestKeygenConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *KeygenConfig
		wantErr bool
	}{
		{
			name: "valid config - bls381",
			config: &KeygenConfig{
				CurveType:  "bls381",
				OutputDir:  "./keys",
				FilePrefix: "test",
			},
			wantErr: false,
		},
		{
			name: "valid config - bn254",
			config: &KeygenConfig{
				CurveType:  "bn254",
				OutputDir:  "./keys",
				FilePrefix: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid curve type",
			config: &KeygenConfig{
				CurveType:  "invalid",
				OutputDir:  "./keys",
				FilePrefix: "test",
			},
			wantErr: true,
		},
		{
			name: "missing output dir for key generation",
			config: &KeygenConfig{
				CurveType:  "bls381",
				FilePrefix: "test",
			},
			wantErr: true,
		},
		{
			name: "valid config with key file for info command",
			config: &KeygenConfig{
				CurveType: "bls381",
				KeyFile:   "./keys/test.key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewKeygenConfig(t *testing.T) {
	// Setup viper with test values
	viper.Reset()
	viper.Set("debug", true)
	viper.Set("curve_type", "bls381")
	viper.Set("output_dir", "./test-keys")
	viper.Set("file_prefix", "test-key")
	viper.Set("key_file", "test.key")
	viper.Set("seed", "1234567890abcdef")
	viper.Set("path", "m/12381/60/0/0")
	viper.Set("password", "test-password")

	// Create config from viper
	config := NewKeygenConfig()

	// Verify values
	assert.True(t, config.Debug)
	assert.Equal(t, "bls381", config.CurveType)
	assert.Equal(t, "./test-keys", config.OutputDir)
	assert.Equal(t, "test-key", config.FilePrefix)
	assert.Equal(t, "test.key", config.KeyFile)
	assert.Equal(t, "1234567890abcdef", config.Seed)
	assert.Equal(t, "m/12381/60/0/0", config.Path)
	assert.Equal(t, "test-password", config.Password)
}

func TestNewKeygenConfigFromYamlBytes(t *testing.T) {
	yamlData := []byte(`
debug: true
curveType: bls381
outputDir: ./yaml-keys
filePrefix: yaml-key
keyFile: yaml.key
seed: yaml-seed
path: m/12381/60/0/0
password: yaml-password
`)

	config, err := NewKeygenConfigFromYamlBytes(yamlData)
	assert.NoError(t, err)
	assert.NotNil(t, config)

	assert.True(t, config.Debug)
	assert.Equal(t, "bls381", config.CurveType)
	assert.Equal(t, "./yaml-keys", config.OutputDir)
	assert.Equal(t, "yaml-key", config.FilePrefix)
	assert.Equal(t, "yaml.key", config.KeyFile)
	assert.Equal(t, "yaml-seed", config.Seed)
	assert.Equal(t, "m/12381/60/0/0", config.Path)
	assert.Equal(t, "yaml-password", config.Password)
}

func TestNewKeygenConfigFromJsonBytes(t *testing.T) {
	jsonData := []byte(`{
  "debug": true,
  "curveType": "bn254",
  "outputDir": "./json-keys",
  "filePrefix": "json-key",
  "keyFile": "json.key",
  "seed": "json-seed",
  "path": "m/12381/60/0/0",
  "password": "json-password"
}`)

	config, err := NewKeygenConfigFromJsonBytes(jsonData)
	assert.NoError(t, err)
	assert.NotNil(t, config)

	assert.True(t, config.Debug)
	assert.Equal(t, "bn254", config.CurveType)
	assert.Equal(t, "./json-keys", config.OutputDir)
	assert.Equal(t, "json-key", config.FilePrefix)
	assert.Equal(t, "json.key", config.KeyFile)
	assert.Equal(t, "json-seed", config.Seed)
	assert.Equal(t, "m/12381/60/0/0", config.Path)
	assert.Equal(t, "json-password", config.Password)
}
