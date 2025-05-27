package keystore

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bls381"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeystoreFormat(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "key.json")

	password := "testpassword123!"

	// Test with BN254
	testKeystoreWithScheme(t, bn254.NewScheme(), "bn254", keyPath, password)

	// Test with BLS381
	keyPath = filepath.Join(tempDir, "key_bls381.json")
	testKeystoreWithScheme(t, bls381.NewScheme(), "bls381", keyPath, password)
}

func testKeystoreWithScheme(t *testing.T, scheme signing.SigningScheme, curveType, keyPath, password string) {
	// Generate key
	privKey, pubKey, err := scheme.GenerateKeyPair()
	require.NoError(t, err)

	// Save key to keystore
	err = SaveToKeystoreWithCurveType(privKey, keyPath, password, curveType, Light())
	require.NoError(t, err)

	// Load keystore
	ks, err := LoadKeystoreFile(keyPath)
	require.NoError(t, err)

	// Validate keystore format is EIP-2335
	assert.Equal(t, 4, ks.Version)
	assert.NotEmpty(t, ks.Pubkey)
	assert.NotEmpty(t, ks.UUID)
	assert.NotEmpty(t, ks.Path)
	assert.Equal(t, curveType, ks.CurveType)

	// Validate crypto modules
	assert.NotEmpty(t, ks.Crypto.KDF.Function)
	assert.NotEmpty(t, ks.Crypto.Checksum.Function)
	assert.NotEmpty(t, ks.Crypto.Cipher.Function)

	// Load private key from keystore
	loadedPrivKey, err := ks.GetPrivateKey(password, scheme)
	require.NoError(t, err)

	// Compare keys
	assert.Equal(t, hex.EncodeToString(privKey.Bytes()), hex.EncodeToString(loadedPrivKey.Bytes()))

	// Verify we can sign with loaded key
	message := []byte("test message")
	sig, err := loadedPrivKey.Sign(message)
	require.NoError(t, err)

	valid, err := sig.Verify(pubKey, message)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestLegacyKeystoreBackwardCompatibility(t *testing.T) {
	// This test validates that we can still load legacy keystores
	// We'll create a legacy format keystore by temporarily setting up legacy
	// format creation, then load it with the new format loader

	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "legacy_key.json")

	// Generate a BN254 key
	scheme := bn254.NewScheme()
	privKey, pubKey, err := scheme.GenerateKeyPair()
	require.NoError(t, err)

	// Create a mock keystore in legacy format
	origKey := privKey.Bytes()
	pubKeyHex := hex.EncodeToString(pubKey.Bytes())

	// Generate test salt and IV
	salt := make([]byte, 32)
	iv := make([]byte, 16)

	// Create a mock MAC (we're not doing actual encryption)
	mac := sha256.Sum256(append(origKey, salt...))
	macHex := hex.EncodeToString(mac[:])

	// Create a temporary test file in the old format
	legacyFormat := `{
		"publicKey": "` + pubKeyHex + `",
		"crypto": {
			"cipher": "aes-128-ctr",
			"ciphertext": "` + hex.EncodeToString(origKey) + `",
			"cipherparams": {
				"iv": "` + hex.EncodeToString(iv) + `"
			},
			"kdf": "scrypt",
			"kdfparams": {
				"dklen": 32,
				"n": 4096,
				"p": 1,
				"r": 8,
				"salt": "` + hex.EncodeToString(salt) + `"
			},
			"mac": "` + macHex + `"
		},
		"uuid": "00000000-0000-0000-0000-000000000000",
		"version": 4,
		"curveType": "bn254"
	}`

	err = os.WriteFile(keyPath, []byte(legacyFormat), 0600)
	require.NoError(t, err)

	// Load the legacy keystore
	ks, err := LoadKeystoreFile(keyPath)
	require.NoError(t, err)

	// Validate it got converted to the new format
	assert.Equal(t, 4, ks.Version)
	assert.Equal(t, "bn254", ks.CurveType)
	assert.Equal(t, pubKeyHex, ks.Pubkey)

	// Verify the legacy indicator fields are set
	assert.Equal(t, "legacy", ks.Crypto.KDF.Function)
	assert.Equal(t, "legacy", ks.Crypto.Checksum.Function)
	assert.Equal(t, "legacy", ks.Crypto.Cipher.Function)

	// Verify a description was added
	assert.Contains(t, ks.Description, "legacy")
}

func TestPasswordProcessing(t *testing.T) {
	// Test password processing according to EIP-2335

	// Test with control characters that should be stripped
	rawPassword := "test\u0000password\u0008with\u001Fcontrol\u007Fchars"
	processed := processPassword(rawPassword)

	// Control characters should be stripped
	expectedProcessed := []byte("testpasswordwithcontrolchars")
	assert.Equal(t, expectedProcessed, processed)

	// Test normalization (NFKD)
	// Using a precomposed character (é) vs decomposed (e + ´)
	precomposed := "café"      // é as a single code point
	decomposed := "cafe\u0301" // e + combining acute accent

	processedPrecomposed := processPassword(precomposed)
	processedDecomposed := processPassword(decomposed)

	// Both should normalize to the same result (we don't compare the exact bytes
	// but rather that they're equivalent after processing)
	assert.Equal(t, processedPrecomposed, processedDecomposed)
}

func TestGenerateRandomPassword(t *testing.T) {
	password, err := GenerateRandomPassword(20)
	require.NoError(t, err)
	assert.Len(t, password, 20)

	password2, err := GenerateRandomPassword(20)
	require.NoError(t, err)

	// Two generated passwords should be different
	assert.NotEqual(t, password, password2)
}

func TestKeystoreBN254(t *testing.T) {
	// Create temp directory for test keystores
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "test_key_bn254.json")

	// Create BN254 private key
	scheme := bn254.NewScheme()
	privateKey, _, err := scheme.GenerateKeyPair()
	require.NoError(t, err)

	// Save the private key to keystore file
	err = SaveToKeystoreWithCurveType(privateKey, keyPath, "testpassword", "bn254", Light())
	require.NoError(t, err)

	// Load the keystore file
	loadedKeystore, err := LoadKeystoreFile(keyPath)
	require.NoError(t, err)

	// Parse the keystore file
	keystoreContent, err := os.ReadFile(keyPath)
	require.NoError(t, err)
	parsedKeystore, err := ParseKeystoreJSON(string(keystoreContent))
	require.NoError(t, err)

	// Verify that parsed and loaded keystores are the same
	assert.Equal(t, loadedKeystore.Pubkey, parsedKeystore.Pubkey)
	assert.Equal(t, loadedKeystore.UUID, parsedKeystore.UUID)
	assert.Equal(t, loadedKeystore.CurveType, parsedKeystore.CurveType)

	// Load the private key from keystore
	loadedKey, err := loadedKeystore.GetPrivateKey("testpassword", scheme)
	require.NoError(t, err)
	assert.Equal(t, privateKey.Bytes(), loadedKey.Bytes())

	// Test GetBN254PrivateKey helper
	loadedKey2, err := loadedKeystore.GetBN254PrivateKey("testpassword")
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

func TestKeystoreBLS381Helper(t *testing.T) {
	// Create temp directory for test keystores
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "test_key_bls381.json")

	// Create BLS381 private key
	scheme := bls381.NewScheme()
	privateKey, _, err := scheme.GenerateKeyPair()
	require.NoError(t, err)

	// Save the private key to keystore file
	err = SaveToKeystoreWithCurveType(privateKey, keyPath, "testpassword", "bls381", Light())
	require.NoError(t, err)

	// Load the keystore file
	loadedKeystore, err := LoadKeystoreFile(keyPath)
	require.NoError(t, err)

	// Test GetBLS381PrivateKey helper
	loadedKey, err := loadedKeystore.GetBLS381PrivateKey("testpassword")
	require.NoError(t, err)

	// Get the bytes from both keys for comparison
	privKeyBytes := privateKey.Bytes()
	loadedKeyBytes := loadedKey.Bytes()

	assert.Equal(t, privKeyBytes, loadedKeyBytes)
}

func TestInvalidKeystore(t *testing.T) {
	// Test with invalid keystore JSON that's not even a valid JSON
	invalidJSON := `{"invalid": "not a valid keystore`
	_, err := ParseKeystoreJSON(invalidJSON)
	assert.Error(t, err)

	// Test with completely invalid crypto structure
	invalidCryptoJSON := `{
		"pubkey": "0123456789abcdef",
		"uuid": "00000000-0000-0000-0000-000000000000",
		"version": 4,
		"crypto": "not an object"
	}`
	_, err = ParseKeystoreJSON(invalidCryptoJSON)
	assert.Error(t, err)

	// Test a clearly invalid keystore (empty JSON)
	emptyJSON := `{}`
	_, err = ParseKeystoreJSON(emptyJSON)
	assert.Error(t, err)

	// Test with nil keystore
	var nilKeystore *EIP2335Keystore
	_, err = nilKeystore.GetPrivateKey("password", bn254.NewScheme())
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

func Test_ParseLegacyKeystoreToEIP2335Keystore(t *testing.T) {
	legacyJson := `
{
    "publicKey": "15de6881d859600f953e1b00fa58a950c65eac7c86860412b269c2a33bac09e51d37f7f962df4041a82808c85140c6186878b4695673ecd3c6fd1b7953d9f77000161e1c998df8e36d9cd89717ec47a5e385220ea4c9fa4bf419a3563fb5c3541425a016e78736ea3568613ff6338ffcec5a40e597b31ed959bb630d22502a70",
    "crypto": {
      "cipher": "aes-128-ctr",
      "ciphertext": "c1a3a27f1c720f683cc2dcef2cc349454d44f86a7a6b50f74d5911a73aa3440be94e39cd76b5b9712e924d313851ad55b7507ac31dda7d502358b4d97154688335493750738de0fd5716c8d013",
      "cipherparams": {
        "iv": "dcb57cec6cd31368eef3940211c2b567"
      },
      "kdf": "scrypt",
      "kdfparams": {
        "dklen": 32,
        "n": 262144,
        "p": 1,
        "r": 8,
        "salt": "b9750dc08899fa40fb0ed858c5d1e29e68beeccd9f837f4504974bb78e3ae0d3"
      },
      "mac": "9644d808d705c9739b79ed6f8c3575883213850faa232003178449ced0a2e266"
    },
    "uuid": "20e14f79-5274-46c4-a081-99f45bf8824d",
    "version": 4,
    "curveType": "bn254"
  }
`
	ks, err := ParseLegacyKeystoreToEIP2335Keystore(legacyJson)
	assert.Nil(t, err)
	assert.NotNil(t, ks)

	pk, err := ks.GetPrivateKey("testpass", bn254.NewScheme())
	assert.Nil(t, err)
	assert.NotNil(t, pk)
}
