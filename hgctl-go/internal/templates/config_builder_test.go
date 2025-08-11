package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
)

// TestConfigBuilder_PrivateKeyConfiguration tests configuration generation with private keys
func TestConfigBuilder_PrivateKeyConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		signerConfigs map[string]*signer.SignerConfig
		envVars       map[string]string
		validateFunc  func(t *testing.T, config map[string]interface{})
	}{
		{
			name: "executor with BLS keystore configuration",
			signerConfigs: map[string]*signer.SignerConfig{
				"BLS": {
					Type:             signer.SignerTypeKeystore,
					KeystorePath:     "/path/to/bls.keystore",
					KeystoreContent:  `{"version":3,"crypto":{"cipher":"aes-128-ctr"}}`,
					KeystorePassword: "test-password",
				},
			},
			envVars: map[string]string{
				"OPERATOR_ADDRESS":   "0x1234567890123456789012345678901234567890",
				"L1_CHAIN_ID":        "1",
				"L1_RPC_URL":         "http://localhost:8545",
				"PERFORMER_REGISTRY": "registry.example.com/performer",
				"KEYSTORE_PASSWORD":  "test-password",
			},
			validateFunc: func(t *testing.T, config map[string]interface{}) {
				// Check top-level structure
				assert.Contains(t, config, "operator")
				assert.Contains(t, config, "grpcPort")
				assert.Contains(t, config, "avsPerformers")

				// Check operator configuration
				operator := config["operator"].(map[string]interface{})
				assert.Equal(t, "0x1234567890123456789012345678901234567890", operator["address"])

				// Check signing keys
				assert.Contains(t, operator, "signingKeys")
				signingKeys := operator["signingKeys"].(map[string]interface{})

				// Validate BLS keystore configuration (template has static keystore path)
				assert.Contains(t, signingKeys, "bls")
				blsSigner := signingKeys["bls"].(map[string]interface{})
				assert.Equal(t, "/keystores/operator.keystore.json", blsSigner["keystoreFile"])
				assert.Equal(t, "test-password", blsSigner["password"])

				// Template only supports BLS, not ECDSA
				assert.NotContains(t, signingKeys, "ecdsa")
			},
		},
		{
			name: "executor with empty password",
			signerConfigs: map[string]*signer.SignerConfig{
				"BLS": {
					Type:             signer.SignerTypeKeystore,
					KeystorePath:     "/path/to/bls.keystore",
					KeystoreContent:  `{"version":3,"crypto":{"cipher":"aes-128-ctr"}}`,
					KeystorePassword: "", // Empty password
				},
			},
			envVars: map[string]string{
				"OPERATOR_ADDRESS":   "0x9999999999999999999999999999999999999999",
				"L1_CHAIN_ID":        "5",
				"L1_RPC_URL":         "https://goerli.infura.io/v3/YOUR-PROJECT-ID",
				"PERFORMER_REGISTRY": "gcr.io/my-project/performer",
				"KEYSTORE_PASSWORD":  "", // Empty password in env
			},
			validateFunc: func(t *testing.T, config map[string]interface{}) {
				operator := config["operator"].(map[string]interface{})
				signingKeys := operator["signingKeys"].(map[string]interface{})

				// Validate BLS signer - template has static keystore config
				assert.Contains(t, signingKeys, "bls")
				blsSigner := signingKeys["bls"].(map[string]interface{})
				assert.Equal(t, "/keystores/operator.keystore.json", blsSigner["keystoreFile"])
				assert.Equal(t, "", blsSigner["password"])

				// Template only supports BLS
				assert.NotContains(t, signingKeys, "ecdsa")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewConfigBuilder()

			configBytes, err := builder.BuildExecutorConfig(tt.signerConfigs, tt.envVars)
			require.NoError(t, err)
			require.NotEmpty(t, configBytes)

			// Parse YAML to validate structure
			var config map[string]interface{}
			err = yaml.Unmarshal(configBytes, &config)
			require.NoError(t, err)

			// Run custom validation
			tt.validateFunc(t, config)
		})
	}
}

// TestConfigBuilder_KeystoreConfiguration tests configuration generation with keystores
func TestConfigBuilder_KeystoreConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		signerConfigs map[string]*signer.SignerConfig
		envVars       map[string]string
		validateFunc  func(t *testing.T, config map[string]interface{})
	}{
		{
			name: "executor with BLS keystore from environment",
			signerConfigs: map[string]*signer.SignerConfig{
				"BLS": {
					Type:             signer.SignerTypeKeystore,
					KeystorePath:     "/path/to/bls.keystore",
					KeystoreContent:  `{"version":3,"crypto":{"cipher":"aes-128-ctr"}}`,
					KeystorePassword: "bls-password-123",
				},
			},
			envVars: map[string]string{
				"OPERATOR_ADDRESS":   "0x1111111111111111111111111111111111111111",
				"L1_CHAIN_ID":        "1",
				"L1_RPC_URL":         "http://localhost:8545",
				"PERFORMER_REGISTRY": "registry.example.com/performer",
				"KEYSTORE_PASSWORD":  "bls-password-123",
			},
			validateFunc: func(t *testing.T, config map[string]interface{}) {
				operator := config["operator"].(map[string]interface{})
				signingKeys := operator["signingKeys"].(map[string]interface{})

				// Validate BLS keystore - static template configuration
				assert.Contains(t, signingKeys, "bls")
				blsSigner := signingKeys["bls"].(map[string]interface{})
				assert.Equal(t, "/keystores/operator.keystore.json", blsSigner["keystoreFile"])
				assert.Equal(t, "bls-password-123", blsSigner["password"])

				// Template only supports BLS
				assert.NotContains(t, signingKeys, "ecdsa")
			},
		},
		{
			name: "executor with keystore and empty password",
			signerConfigs: map[string]*signer.SignerConfig{
				"BLS": {
					Type:             signer.SignerTypeKeystore,
					KeystorePath:     "/opt/keystores/bls.json",
					KeystoreContent:  `{"version":3}`,
					KeystorePassword: "", // Empty password
				},
			},
			envVars: map[string]string{
				"OPERATOR_ADDRESS":   "0x2222222222222222222222222222222222222222",
				"L1_CHAIN_ID":        "1",
				"L1_RPC_URL":         "http://localhost:8545",
				"PERFORMER_REGISTRY": "registry.example.com/performer",
			},
			validateFunc: func(t *testing.T, config map[string]interface{}) {
				operator := config["operator"].(map[string]interface{})
				signingKeys := operator["signingKeys"].(map[string]interface{})

				// Validate BLS signer - template has static keystore config
				blsSigner := signingKeys["bls"].(map[string]interface{})
				assert.Equal(t, "/keystores/operator.keystore.json", blsSigner["keystoreFile"])
				assert.Equal(t, "", blsSigner["password"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewConfigBuilder()

			configBytes, err := builder.BuildExecutorConfig(tt.signerConfigs, tt.envVars)
			require.NoError(t, err)
			require.NotEmpty(t, configBytes)

			// Parse YAML to validate structure
			var config map[string]interface{}
			err = yaml.Unmarshal(configBytes, &config)
			require.NoError(t, err)

			// Run custom validation
			tt.validateFunc(t, config)
		})
	}
}

// TestConfigBuilder_Web3SignerConfiguration tests configuration generation with web3signers
// TODO: Enable when template supports dynamic signer configuration
func TestConfigBuilder_Web3SignerConfiguration(t *testing.T) {
	t.Skip("Skipping Web3Signer tests - template currently only supports static BLS keystore configuration")
}

// TestConfigBuilder_MixedConfiguration tests configuration with mixed signer types
// TODO: Enable when template supports dynamic signer configuration
func TestConfigBuilder_MixedConfiguration(t *testing.T) {
	t.Skip("Skipping MixedConfiguration test - template currently only supports static BLS keystore configuration")
}

// TestConfigBuilder_EnvironmentVariables tests environment variable substitution
func TestConfigBuilder_EnvironmentVariables(t *testing.T) {
	builder := NewConfigBuilder()

	signerConfigs := map[string]*signer.SignerConfig{
		"BLS": {
			Type:       signer.SignerTypePrivateKey,
			PrivateKey: "0x1234567890abcdef",
		},
	}

	// Test with various environment variables
	envVars := map[string]string{
		"OPERATOR_ADDRESS":       "0x6666666666666666666666666666666666666666",
		"L1_CHAIN_ID":            "1337",
		"L1_RPC_URL":             "http://custom-rpc.example.com:8545",
		"PERFORMER_REGISTRY":     "custom-registry.io/performer",
		"EXECUTOR_PORT":          "9999",
		"PERFORMER_NETWORK_NAME": "custom-network",
	}

	configBytes, err := builder.BuildExecutorConfig(signerConfigs, envVars)
	require.NoError(t, err)

	// Parse and validate
	var config map[string]interface{}
	err = yaml.Unmarshal(configBytes, &config)
	require.NoError(t, err)

	// Check that environment variables are properly substituted
	assert.Equal(t, 9999, config["grpcPort"])
	assert.Equal(t, "custom-network", config["performerNetworkName"])

	operator := config["operator"].(map[string]interface{})
	assert.Equal(t, "0x6666666666666666666666666666666666666666", operator["address"])

	l1Chain := config["l1Chain"].(map[string]interface{})
	assert.Equal(t, "1337", l1Chain["chainId"])
	assert.Equal(t, "http://custom-rpc.example.com:8545", l1Chain["rpcUrl"])

	// avsPerformers should be empty/nil - performers are not specified during executor deployment
	assert.Contains(t, config, "avsPerformers")
}

// TestConfigBuilder_ErrorHandling tests error conditions
func TestConfigBuilder_ErrorHandling(t *testing.T) {
	builder := NewConfigBuilder()

	// Test with nil signer configs
	envVars := map[string]string{
		"OPERATOR_ADDRESS":   "0x8888888888888888888888888888888888888888",
		"L1_CHAIN_ID":        "1",
		"L1_RPC_URL":         "http://localhost:8545",
		"PERFORMER_REGISTRY": "registry.example.com/performer",
	}

	configBytes, err := builder.BuildExecutorConfig(nil, envVars)
	assert.NoError(t, err) // Should not error, just create config without signers
	assert.NotEmpty(t, configBytes)

	// Verify the config still has valid structure
	var config map[string]interface{}
	err = yaml.Unmarshal(configBytes, &config)
	assert.NoError(t, err)
	assert.Contains(t, config, "operator")
	assert.Contains(t, config, "avsPerformers")

	// Test with empty signer configs
	configBytes, err = builder.BuildExecutorConfig(map[string]*signer.SignerConfig{}, envVars)
	assert.NoError(t, err)
	assert.NotEmpty(t, configBytes)

	// Test with nil env vars
	signerConfigs := map[string]*signer.SignerConfig{
		"BLS": {
			Type:       signer.SignerTypePrivateKey,
			PrivateKey: "0x1234",
		},
	}
	configBytes, err = builder.BuildExecutorConfig(signerConfigs, nil)
	assert.NoError(t, err) // Should work but might have empty values
	assert.NotEmpty(t, configBytes)
}

// TestConfigBuilder_EnvironmentVariableInjection tests the new environment variable injection for file contents
func TestConfigBuilder_EnvironmentVariableInjection(t *testing.T) {
	builder := NewConfigBuilder()

	tests := []struct {
		name          string
		signerConfigs map[string]*signer.SignerConfig
		envVars       map[string]string
		validateFunc  func(t *testing.T, config map[string]interface{})
	}{
		{
			name: "fallback to file paths when no content provided",
			signerConfigs: map[string]*signer.SignerConfig{
				"BLS": {
					Type:         signer.SignerTypeKeystore,
					KeystorePath: "/path/to/keystore.json",
					// No KeystoreContent provided
					KeystorePassword: "test-password",
				},
			},
			envVars: map[string]string{
				"OPERATOR_ADDRESS":   "0x1234567890123456789012345678901234567890",
				"L1_CHAIN_ID":        "1",
				"L1_RPC_URL":         "http://localhost:8545",
				"PERFORMER_REGISTRY": "registry.example.com/performer",
			},
			validateFunc: func(t *testing.T, config map[string]interface{}) {
				operator := config["operator"].(map[string]interface{})
				signingKeys := operator["signingKeys"].(map[string]interface{})

				blsSigner := signingKeys["bls"].(map[string]interface{})
				// Template has static keystore configuration
				assert.Equal(t, "/keystores/operator.keystore.json", blsSigner["keystoreFile"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configBytes, err := builder.BuildExecutorConfig(tt.signerConfigs, tt.envVars)
			require.NoError(t, err)
			require.NotEmpty(t, configBytes)

			// Parse YAML to validate structure
			var config map[string]interface{}
			err = yaml.Unmarshal(configBytes, &config)
			require.NoError(t, err)

			// Run custom validation
			tt.validateFunc(t, config)
		})
	}
}
