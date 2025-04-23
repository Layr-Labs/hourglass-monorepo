package keystore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bls381"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeystoreBN254(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "keystore-test-bn254")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test keystore path
	keystorePath := filepath.Join(tempDir, "test-bn254.json")

	// Generate a key pair
	scheme := bn254.NewScheme()
	privateKey, publicKey, err := scheme.GenerateKeyPair()
	require.NoError(t, err)

	// Test password
	password := "test-password"

	// Save to keystore with curve type
	err = SaveToKeystoreWithCurveType(privateKey, keystorePath, password, "bn254", Light())
	require.NoError(t, err)

	// Verify the file exists
	_, err = os.Stat(keystorePath)
	require.NoError(t, err)

	// Get keystore info
	info, err := GetKeystoreInfo(keystorePath)
	require.NoError(t, err)
	assert.NotNil(t, info["publicKey"])
	assert.NotNil(t, info["uuid"])
	assert.NotNil(t, info["version"])
	assert.Equal(t, "bn254", info["curveType"])

	// Load from keystore
	loadedKey, err := LoadFromKeystore(keystorePath, password, scheme)
	require.NoError(t, err)

	// Verify the loaded key matches the original
	assert.Equal(t, privateKey.Bytes(), loadedKey.Bytes())
	assert.Equal(t, publicKey.Bytes(), loadedKey.Public().Bytes())

	// Test the keystore
	err = TestKeystore(keystorePath, password, scheme)
	require.NoError(t, err)

	// Test with incorrect password
	_, err = LoadFromKeystore(keystorePath, "wrong-password", scheme)
	assert.Error(t, err)

	// Test loading without providing a scheme (should use the curve type from the keystore)
	loadedKey2, err := LoadFromKeystore(keystorePath, password, nil)
	require.NoError(t, err)
	assert.Equal(t, privateKey.Bytes(), loadedKey2.Bytes())
}

func TestKeystoreBLS381(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "keystore-test-bls381")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test keystore path
	keystorePath := filepath.Join(tempDir, "test-bls381.json")

	// Generate a key pair
	scheme := bls381.NewScheme()
	privateKey, publicKey, err := scheme.GenerateKeyPair()
	require.NoError(t, err)

	// Test password
	password := "test-password"

	// Save to keystore with curve type
	err = SaveToKeystoreWithCurveType(privateKey, keystorePath, password, "bls381", Light())
	require.NoError(t, err)

	// Verify the file exists
	_, err = os.Stat(keystorePath)
	require.NoError(t, err)

	// Get keystore info
	info, err := GetKeystoreInfo(keystorePath)
	require.NoError(t, err)
	assert.NotNil(t, info["publicKey"])
	assert.NotNil(t, info["uuid"])
	assert.NotNil(t, info["version"])
	assert.Equal(t, "bls381", info["curveType"])

	// Load from keystore
	loadedKey, err := LoadFromKeystore(keystorePath, password, scheme)
	require.NoError(t, err)

	// Verify the loaded key matches the original
	assert.Equal(t, privateKey.Bytes(), loadedKey.Bytes())
	assert.Equal(t, publicKey.Bytes(), loadedKey.Public().Bytes())

	// Test the keystore
	err = TestKeystore(keystorePath, password, scheme)
	require.NoError(t, err)

	// Test with incorrect password
	_, err = LoadFromKeystore(keystorePath, "wrong-password", scheme)
	assert.Error(t, err)

	// Test loading without providing a scheme (should use the curve type from the keystore)
	loadedKey2, err := LoadFromKeystore(keystorePath, password, nil)
	require.NoError(t, err)
	assert.Equal(t, privateKey.Bytes(), loadedKey2.Bytes())
}

func TestKeystoreBackwardCompatibility(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "keystore-test-compat")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test keystore path
	keystorePath := filepath.Join(tempDir, "test-compat.json")

	// Generate a key pair
	scheme := bn254.NewScheme()
	privateKey, publicKey, err := scheme.GenerateKeyPair()
	require.NoError(t, err)

	// Test password
	password := "test-password"

	// Save to keystore without curve type (using the old function)
	err = SaveToKeystore(privateKey, keystorePath, password, Light())
	require.NoError(t, err)

	// Verify the file exists
	_, err = os.Stat(keystorePath)
	require.NoError(t, err)

	// Get keystore info
	info, err := GetKeystoreInfo(keystorePath)
	require.NoError(t, err)
	assert.NotNil(t, info["publicKey"])
	assert.NotNil(t, info["uuid"])
	assert.NotNil(t, info["version"])

	// The curveType field might be missing or empty in backward compatibility mode
	if curveType, ok := info["curveType"]; ok {
		assert.Equal(t, "", curveType)
	}

	// Load from keystore - we need to provide the scheme explicitly
	loadedKey, err := LoadFromKeystore(keystorePath, password, scheme)
	require.NoError(t, err)

	// Verify the loaded key matches the original
	assert.Equal(t, privateKey.Bytes(), loadedKey.Bytes())
	assert.Equal(t, publicKey.Bytes(), loadedKey.Public().Bytes())

	// Test loading without providing a scheme (should fail since there's no curve type in the keystore)
	_, err = LoadFromKeystore(keystorePath, password, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no signing scheme provided and unable to determine from keystore")
}

func TestGenerateRandomPassword(t *testing.T) {
	// Test password generation with default length
	password, err := GenerateRandomPassword(32)
	require.NoError(t, err)
	assert.Len(t, password, 32)

	// Test password generation with short length (should be raised to 16)
	password, err = GenerateRandomPassword(8)
	require.NoError(t, err)
	assert.Len(t, password, 16)

	// Test that two generated passwords are different
	password1, err := GenerateRandomPassword(16)
	require.NoError(t, err)
	password2, err := GenerateRandomPassword(16)
	require.NoError(t, err)
	assert.NotEqual(t, password1, password2)
}

func TestInvalidKeystore(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "keystore-test-invalid")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create an invalid file
	invalidPath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidPath, []byte("{\"not_a_keystore\": true}"), 0600)
	require.NoError(t, err)

	// Try to get info from invalid file
	_, err = GetKeystoreInfo(invalidPath)
	assert.Error(t, err)

	// Try to load an invalid file
	scheme := bn254.NewScheme()
	_, err = LoadFromKeystore(invalidPath, "password", scheme)
	assert.Error(t, err)

	// Test with non-existent file
	_, err = LoadFromKeystore("/nonexistent/file.json", "password", scheme)
	assert.Error(t, err)
}

func TestGetSigningScheme(t *testing.T) {
	// Test getting valid signing schemes
	scheme1, err := GetSigningScheme("bls381")
	require.NoError(t, err)
	assert.NotNil(t, scheme1)
	assert.IsType(t, &bls381.Scheme{}, scheme1)

	scheme2, err := GetSigningScheme("bn254")
	require.NoError(t, err)
	assert.NotNil(t, scheme2)
	assert.IsType(t, &bn254.Scheme{}, scheme2)

	// Test case insensitivity
	scheme3, err := GetSigningScheme("BLS381")
	require.NoError(t, err)
	assert.NotNil(t, scheme3)

	// Test invalid curve type
	_, err = GetSigningScheme("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported curve type")
}

func TestDetermineCurveType(t *testing.T) {
	assert.Equal(t, "bls381", DetermineCurveType("bls381"))
	assert.Equal(t, "bls381", DetermineCurveType("BLS381"))
	assert.Equal(t, "bn254", DetermineCurveType("bn254"))
	assert.Equal(t, "bn254", DetermineCurveType("BN254"))
	assert.Equal(t, "", DetermineCurveType("invalid"))
	assert.Equal(t, "", DetermineCurveType(""))
}
