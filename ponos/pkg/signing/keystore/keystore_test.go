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

	// Load keystore file
	loadedKeystore, err := LoadKeystoreFile(keystorePath)
	require.NoError(t, err)

	// Load private key from keystore object
	loadedKey, err := loadedKeystore.GetPrivateKey(password, scheme)
	require.NoError(t, err)

	// Verify the loaded key matches the original
	assert.Equal(t, privateKey.Bytes(), loadedKey.Bytes())
	assert.Equal(t, publicKey.Bytes(), loadedKey.Public().Bytes())

	// Test the keystore
	err = TestKeystore(keystorePath, password, scheme)
	require.NoError(t, err)

	// Test with incorrect password
	_, err = loadedKeystore.GetPrivateKey("wrong-password", scheme)
	assert.Error(t, err)

	// Test loading without providing a scheme (should use the curve type from the keystore)
	loadedKey2, err := loadedKeystore.GetPrivateKey(password, nil)
	require.NoError(t, err)
	assert.Equal(t, privateKey.Bytes(), loadedKey2.Bytes())

	// Test the ParseKeystoreJSON function
	fileContent, err := os.ReadFile(keystorePath)
	require.NoError(t, err)
	parsedKeystore, err := ParseKeystoreJSON(string(fileContent))
	require.NoError(t, err)
	assert.Equal(t, loadedKeystore.PublicKey, parsedKeystore.PublicKey)
	assert.Equal(t, loadedKeystore.UUID, parsedKeystore.UUID)
	assert.Equal(t, loadedKeystore.Version, parsedKeystore.Version)
	assert.Equal(t, loadedKeystore.CurveType, parsedKeystore.CurveType)

	loadedKey3, err := parsedKeystore.GetPrivateKey(password, scheme)
	require.NoError(t, err)
	assert.Equal(t, privateKey.Bytes(), loadedKey3.Bytes())
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

	// Load keystore file
	loadedKeystore, err := LoadKeystoreFile(keystorePath)
	require.NoError(t, err)

	// Load private key from keystore object
	loadedKey, err := loadedKeystore.GetPrivateKey(password, scheme)
	require.NoError(t, err)

	// Verify the loaded key matches the original
	assert.Equal(t, privateKey.Bytes(), loadedKey.Bytes())
	assert.Equal(t, publicKey.Bytes(), loadedKey.Public().Bytes())

	// Test the keystore
	err = TestKeystore(keystorePath, password, scheme)
	require.NoError(t, err)

	// Test with incorrect password
	_, err = loadedKeystore.GetPrivateKey("wrong-password", scheme)
	assert.Error(t, err)

	// Test loading without providing a scheme (should use the curve type from the keystore)
	loadedKey2, err := loadedKeystore.GetPrivateKey(password, nil)
	require.NoError(t, err)
	assert.Equal(t, privateKey.Bytes(), loadedKey2.Bytes())
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

	// Try to load invalid file
	_, err = LoadKeystoreFile(invalidPath)
	assert.Error(t, err)

	// Test ParseKeystoreJSON with invalid JSON
	_, err = ParseKeystoreJSON("{invalid json")
	assert.Error(t, err)

	// Test ParseKeystoreJSON with valid JSON but invalid keystore format
	_, err = ParseKeystoreJSON("{\"not_a_keystore\": true}")
	assert.Error(t, err)

	// Test with nil keystore
	var nilKeystore *Keystore
	_, err = nilKeystore.GetPrivateKey("password", bn254.NewScheme())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "keystore data cannot be nil")

	// Test with non-existent file
	_, err = LoadKeystoreFile("/nonexistent/file.json")
	assert.Error(t, err)
}

func TestGetSigningScheme(t *testing.T) {
	// Test getting valid signing schemes
	scheme1, err := GetSigningSchemeForCurveType("bls381")
	require.NoError(t, err)
	assert.NotNil(t, scheme1)
	assert.IsType(t, &bls381.Scheme{}, scheme1)

	scheme2, err := GetSigningSchemeForCurveType("bn254")
	require.NoError(t, err)
	assert.NotNil(t, scheme2)
	assert.IsType(t, &bn254.Scheme{}, scheme2)

	// Test case insensitivity
	scheme3, err := GetSigningSchemeForCurveType("BLS381")
	require.NoError(t, err)
	assert.NotNil(t, scheme3)

	// Test invalid curve type
	_, err = GetSigningSchemeForCurveType("invalid")
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
