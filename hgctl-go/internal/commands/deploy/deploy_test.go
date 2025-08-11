package deploy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInjectFileContentsAsEnvVars tests the environment variable injection functionality
func TestInjectFileContentsAsEnvVars(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "hgctl-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := map[string]string{
		"operator.bls.keystore.json":   `{"version":3,"crypto":{"cipher":"aes-128-ctr"}}`,
		"operator.ecdsa.keystore.json": `{"version":3,"crypto":{"cipher":"aes-256-ctr"}}`,
		"web3signer-bls-ca.crt": `-----BEGIN CERTIFICATE-----
MIICertificateContentBLS
-----END CERTIFICATE-----`,
		"web3signer-ecdsa-ca.crt": `-----BEGIN CERTIFICATE-----
MIICertificateContentECDSA
-----END CERTIFICATE-----`,
	}

	// Write test files
	for filename, content := range testFiles {
		err := os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0600)
		require.NoError(t, err)
	}

	// Test injection
	log := logger.NewLogger(false)
	dockerArgs := []string{"docker", "run"}

	// Call the injection function
	resultArgs := injectFileContentsAsEnvVars(dockerArgs, tempDir, log)

	// Verify that environment variables were added
	expectedEnvVars := map[string]string{
		"BLS_KEYSTORE_CONTENT":   `{"version":3,"crypto":{"cipher":"aes-128-ctr"}}`,
		"ECDSA_KEYSTORE_CONTENT": `{"version":3,"crypto":{"cipher":"aes-256-ctr"}}`,
		"WEB3_SIGNER_BLS_CA_CERT_CONTENT": `-----BEGIN CERTIFICATE-----
MIICertificateContentBLS
-----END CERTIFICATE-----`,
		"WEB3_SIGNER_ECDSA_CA_CERT_CONTENT": `-----BEGIN CERTIFICATE-----
MIICertificateContentECDSA
-----END CERTIFICATE-----`,
	}

	// Count how many env vars were added
	envCount := 0
	for i, arg := range resultArgs {
		if arg == "-e" && i+1 < len(resultArgs) {
			envCount++
			// Parse the env var
			envVar := resultArgs[i+1]
			for expectedVar, expectedContent := range expectedEnvVars {
				if envVar == expectedVar+"="+expectedContent {
					delete(expectedEnvVars, expectedVar) // Remove from expected once found
				}
			}
		}
	}

	// Should have added 4 environment variables
	assert.Equal(t, 4, envCount)
	// All expected env vars should have been found
	assert.Empty(t, expectedEnvVars)
}

// TestInjectFileContentsAsEnvVars_MissingFiles tests behavior when files don't exist
func TestInjectFileContentsAsEnvVars_MissingFiles(t *testing.T) {
	// Create a temporary directory with no files
	tempDir, err := os.MkdirTemp("", "hgctl-test-empty-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	log := logger.NewLogger(false)
	dockerArgs := []string{"docker", "run"}

	// Call the injection function
	resultArgs := injectFileContentsAsEnvVars(dockerArgs, tempDir, log)

	// Should not add any env vars when files don't exist
	assert.Equal(t, dockerArgs, resultArgs)
}

// TestInjectFileContentsAsEnvVars_PartialFiles tests behavior with only some files present
func TestInjectFileContentsAsEnvVars_PartialFiles(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "hgctl-test-partial-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create only BLS keystore
	blsContent := `{"version":3,"bls":true}`
	err = os.WriteFile(filepath.Join(tempDir, "operator.bls.keystore.json"), []byte(blsContent), 0600)
	require.NoError(t, err)

	log := logger.NewLogger(false)
	dockerArgs := []string{"docker", "run", "-d"}

	// Call the injection function
	resultArgs := injectFileContentsAsEnvVars(dockerArgs, tempDir, log)

	// Should have added only one env var
	envCount := 0
	foundBLS := false
	for i, arg := range resultArgs {
		if arg == "-e" && i+1 < len(resultArgs) {
			envCount++
			if resultArgs[i+1] == "BLS_KEYSTORE_CONTENT="+blsContent {
				foundBLS = true
			}
		}
	}

	assert.Equal(t, 1, envCount)
	assert.True(t, foundBLS)
}

// TestInjectFileContentsAsEnvVars_PreservesExistingArgs tests that existing docker args are preserved
func TestInjectFileContentsAsEnvVars_PreservesExistingArgs(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "hgctl-test-preserve-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test file
	err = os.WriteFile(
		filepath.Join(tempDir, "operator.bls.keystore.json"),
		[]byte(`{"test":"data"}`),
		0600,
	)
	require.NoError(t, err)

	log := logger.NewLogger(false)
	dockerArgs := []string{
		"docker", "run", "-d",
		"--name", "test-container",
		"-e", "EXISTING_VAR=value",
		"-p", "8080:8080",
		"myimage:latest",
	}

	// Call the injection function
	resultArgs := injectFileContentsAsEnvVars(dockerArgs, tempDir, log)

	// Verify original args are preserved
	assert.Contains(t, resultArgs, "docker")
	assert.Contains(t, resultArgs, "run")
	assert.Contains(t, resultArgs, "-d")
	assert.Contains(t, resultArgs, "--name")
	assert.Contains(t, resultArgs, "test-container")
	assert.Contains(t, resultArgs, "EXISTING_VAR=value")
	assert.Contains(t, resultArgs, "-p")
	assert.Contains(t, resultArgs, "8080:8080")
	assert.Contains(t, resultArgs, "myimage:latest")

	// And new env var is added
	assert.Contains(t, resultArgs, `BLS_KEYSTORE_CONTENT={"test":"data"}`)
}
