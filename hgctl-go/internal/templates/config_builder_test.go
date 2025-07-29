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
			name: "executor with BLS and ECDSA private keys",
			signerConfigs: map[string]*signer.SignerConfig{
				"BLS": {
					Type:       signer.SignerTypePrivateKey,
					PrivateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				},
				"ECDSA": {
					Type:       signer.SignerTypePrivateKey,
					PrivateKey: "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				},
			},
			envVars: map[string]string{
				"OPERATOR_ADDRESS":     "0x1234567890123456789012345678901234567890",
				"L1_CHAIN_ID":         "1",
				"L1_RPC_URL":          "http://localhost:8545",
				"PERFORMER_REGISTRY":  "registry.example.com/performer",
				"AVS_ADDRESS":         "0xabc1234567890123456789012345678901234567",
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

				// Validate BLS private key
				assert.Contains(t, signingKeys, "bls")
				blsSigner := signingKeys["bls"].(map[string]interface{})
				assert.Equal(t, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", blsSigner["privateKey"])

				// Validate ECDSA private key
				assert.Contains(t, signingKeys, "ecdsa")
				ecdsaSigner := signingKeys["ecdsa"].(map[string]interface{})
				assert.Equal(t, "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", ecdsaSigner["privateKey"])
			},
		},
		{
			name: "executor with BLS private key only",
			signerConfigs: map[string]*signer.SignerConfig{
				"BLS": {
					Type:       signer.SignerTypePrivateKey,
					PrivateKey: "0xdeadbeef1234567890abcdef1234567890abcdef1234567890abcdef12345678",
				},
			},
			envVars: map[string]string{
				"OPERATOR_ADDRESS":    "0x9999999999999999999999999999999999999999",
				"L1_CHAIN_ID":        "5",
				"L1_RPC_URL":         "https://goerli.infura.io/v3/YOUR-PROJECT-ID",
				"PERFORMER_REGISTRY": "gcr.io/my-project/performer",
			},
			validateFunc: func(t *testing.T, config map[string]interface{}) {
				operator := config["operator"].(map[string]interface{})
				signingKeys := operator["signingKeys"].(map[string]interface{})

				// Validate BLS signer
				assert.Contains(t, signingKeys, "bls")
				blsSigner := signingKeys["bls"].(map[string]interface{})
				assert.Equal(t, "0xdeadbeef1234567890abcdef1234567890abcdef1234567890abcdef12345678", blsSigner["privateKey"])

				// Ensure ECDSA signer is not present
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
			name: "executor with BLS and ECDSA keystores",
			signerConfigs: map[string]*signer.SignerConfig{
				"BLS": {
					Type:             signer.SignerTypeKeystore,
					KeystorePath:     "/path/to/bls.keystore",
					KeystoreContent:  `{"version":3,"crypto":{"cipher":"aes-128-ctr"}}`,
					KeystorePassword: "bls-password-123",
				},
				"ECDSA": {
					Type:             signer.SignerTypeKeystore,
					KeystorePath:     "/path/to/ecdsa.keystore",
					KeystoreContent:  `{"version":3,"crypto":{"cipher":"aes-128-ctr"}}`,
					KeystorePassword: "ecdsa-password-456",
				},
			},
			envVars: map[string]string{
				"OPERATOR_ADDRESS":    "0x1111111111111111111111111111111111111111",
				"L1_CHAIN_ID":        "1",
				"L1_RPC_URL":         "http://localhost:8545",
				"PERFORMER_REGISTRY": "registry.example.com/performer",
			},
			validateFunc: func(t *testing.T, config map[string]interface{}) {
				operator := config["operator"].(map[string]interface{})
				signingKeys := operator["signingKeys"].(map[string]interface{})

				// Validate BLS keystore
				assert.Contains(t, signingKeys, "bls")
				blsSigner := signingKeys["bls"].(map[string]interface{})
				assert.Contains(t, blsSigner, "keystoreContent")
				assert.Equal(t, "bls-password-123", blsSigner["password"])

				// Validate ECDSA keystore
				assert.Contains(t, signingKeys, "ecdsa")
				ecdsaSigner := signingKeys["ecdsa"].(map[string]interface{})
				assert.Contains(t, ecdsaSigner, "keystoreContent")
				assert.Equal(t, "ecdsa-password-456", ecdsaSigner["password"])
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
				"OPERATOR_ADDRESS":    "0x2222222222222222222222222222222222222222",
				"L1_CHAIN_ID":        "1",
				"L1_RPC_URL":         "http://localhost:8545",
				"PERFORMER_REGISTRY": "registry.example.com/performer",
			},
			validateFunc: func(t *testing.T, config map[string]interface{}) {
				operator := config["operator"].(map[string]interface{})
				signingKeys := operator["signingKeys"].(map[string]interface{})

				// Validate BLS signer
				blsSigner := signingKeys["bls"].(map[string]interface{})
				assert.Contains(t, blsSigner, "keystoreContent")
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
func TestConfigBuilder_Web3SignerConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		signerConfigs map[string]*signer.SignerConfig
		envVars       map[string]string
		validateFunc  func(t *testing.T, config map[string]interface{})
	}{
		{
			name: "executor with BLS and ECDSA web3signers",
			signerConfigs: map[string]*signer.SignerConfig{
				"BLS": {
					Type:                signer.SignerTypeWeb3Signer,
					Web3SignerURL:       "http://web3signer-bls:9000",
					Web3SignerPublicKey: "0xabc123def456",
				},
				"ECDSA": {
					Type:                signer.SignerTypeWeb3Signer,
					Web3SignerURL:       "https://web3signer-ecdsa:9001",
					Web3SignerPublicKey: "0xdef456abc789",
				},
			},
			envVars: map[string]string{
				"OPERATOR_ADDRESS":    "0x3333333333333333333333333333333333333333",
				"L1_CHAIN_ID":        "1",
				"L1_RPC_URL":         "http://localhost:8545",
				"PERFORMER_REGISTRY": "registry.example.com/performer",
			},
			validateFunc: func(t *testing.T, config map[string]interface{}) {
				operator := config["operator"].(map[string]interface{})
				signingKeys := operator["signingKeys"].(map[string]interface{})

				// Validate BLS web3signer
				blsSigner := signingKeys["bls"].(map[string]interface{})
				assert.Equal(t, true, blsSigner["remoteSigner"])
				remoteConfig := blsSigner["remoteSignerConfig"].(map[string]interface{})
				assert.Equal(t, "http://web3signer-bls:9000", remoteConfig["url"])
				assert.Equal(t, "0xabc123def456", remoteConfig["publicKey"])

				// Validate ECDSA web3signer
				ecdsaSigner := signingKeys["ecdsa"].(map[string]interface{})
				assert.Equal(t, true, ecdsaSigner["remoteSigner"])
				remoteConfig = ecdsaSigner["remoteSignerConfig"].(map[string]interface{})
				assert.Equal(t, "https://web3signer-ecdsa:9001", remoteConfig["url"])
				assert.Equal(t, "0xdef456abc789", remoteConfig["publicKey"])
			},
		},
		{
			name: "executor with web3signer and TLS certificates",
			signerConfigs: map[string]*signer.SignerConfig{
				"BLS": {
					Type:                signer.SignerTypeWeb3Signer,
					Web3SignerURL:       "https://secure-web3signer:9000",
					Web3SignerPublicKey: "my-bls-key",
					Web3SignerCA:        "-----BEGIN CERTIFICATE-----\nCA CONTENT\n-----END CERTIFICATE-----",
					Web3SignerCert:      "-----BEGIN CERTIFICATE-----\nCERT CONTENT\n-----END CERTIFICATE-----",
					Web3SignerKey:       "-----BEGIN PRIVATE KEY-----\nKEY CONTENT\n-----END PRIVATE KEY-----",
				},
			},
			envVars: map[string]string{
				"OPERATOR_ADDRESS":    "0x4444444444444444444444444444444444444444",
				"L1_CHAIN_ID":        "1",
				"L1_RPC_URL":         "http://localhost:8545",
				"PERFORMER_REGISTRY": "registry.example.com/performer",
			},
			validateFunc: func(t *testing.T, config map[string]interface{}) {
				operator := config["operator"].(map[string]interface{})
				signingKeys := operator["signingKeys"].(map[string]interface{})

				// Validate BLS web3signer with TLS
				blsSigner := signingKeys["bls"].(map[string]interface{})
				remoteConfig := blsSigner["remoteSignerConfig"].(map[string]interface{})
				assert.Contains(t, remoteConfig, "caCertContent")
				assert.Contains(t, remoteConfig, "certContent")
				assert.Contains(t, remoteConfig, "keyContent")
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

// TestConfigBuilder_MixedConfiguration tests configuration with mixed signer types
func TestConfigBuilder_MixedConfiguration(t *testing.T) {
	builder := NewConfigBuilder()

	// Test executor with BLS keystore and ECDSA web3signer
	signerConfigs := map[string]*signer.SignerConfig{
		"BLS": {
			Type:             signer.SignerTypeKeystore,
			KeystorePath:     "/keystores/bls.json",
			KeystoreContent:  `{"version":3}`,
			KeystorePassword: "secure-password",
		},
		"ECDSA": {
			Type:                signer.SignerTypeWeb3Signer,
			Web3SignerURL:       "http://web3signer:9000",
			Web3SignerPublicKey: "ecdsa-key-001",
		},
	}

	envVars := map[string]string{
		"OPERATOR_ADDRESS":    "0x5555555555555555555555555555555555555555",
		"L1_CHAIN_ID":        "1",
		"L1_RPC_URL":         "http://localhost:8545",
		"PERFORMER_REGISTRY": "registry.example.com/performer",
	}

	configBytes, err := builder.BuildExecutorConfig(signerConfigs, envVars)
	require.NoError(t, err)
	require.NotEmpty(t, configBytes)

	// Parse and validate
	var config map[string]interface{}
	err = yaml.Unmarshal(configBytes, &config)
	require.NoError(t, err)

	operator := config["operator"].(map[string]interface{})
	signingKeys := operator["signingKeys"].(map[string]interface{})

	// Validate BLS keystore signer
	blsSigner := signingKeys["bls"].(map[string]interface{})
	assert.Contains(t, blsSigner, "keystoreContent")
	assert.Contains(t, blsSigner, "password")

	// Validate ECDSA web3signer
	ecdsaSigner := signingKeys["ecdsa"].(map[string]interface{})
	assert.Equal(t, true, ecdsaSigner["remoteSigner"])
	remoteConfig := ecdsaSigner["remoteSignerConfig"].(map[string]interface{})
	assert.Equal(t, "http://web3signer:9000", remoteConfig["url"])
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
		"OPERATOR_ADDRESS":          "0x6666666666666666666666666666666666666666",
		"L1_CHAIN_ID":              "1337",
		"L1_RPC_URL":               "http://custom-rpc.example.com:8545",
		"PERFORMER_REGISTRY":       "custom-registry.io/performer",
		"EXECUTOR_PORT":            "9999",
		"PERFORMER_NETWORK_NAME":   "custom-network",
		"PERFORMER_TAG":            "v1.0.0",
		"PERFORMER_PROCESS_TYPE":   "worker",
		"AVS_ADDRESS":              "0x7777777777777777777777777777777777777777",
		"WORKER_COUNT":             "5",
		"SIGNING_CURVE":            "secp256k1",
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

	performers := config["avsPerformers"].([]interface{})
	performer := performers[0].(map[string]interface{})
	assert.Equal(t, "0x7777777777777777777777777777777777777777", performer["avsAddress"])
	assert.Equal(t, 5, performer["workerCount"])
	assert.Equal(t, "secp256k1", performer["signingCurve"]) 
	assert.Equal(t, "worker", performer["processType"])
}

// TestConfigBuilder_ErrorHandling tests error conditions
func TestConfigBuilder_ErrorHandling(t *testing.T) {
	builder := NewConfigBuilder()

	// Test with nil signer configs
	envVars := map[string]string{
		"OPERATOR_ADDRESS":    "0x8888888888888888888888888888888888888888",
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