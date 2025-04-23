package keystore

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bls381"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/google/uuid"
)

// ErrInvalidKeystoreFile is returned when a keystore file is not valid or is corrupted
var ErrInvalidKeystoreFile = errors.New("invalid keystore file")

// Keystore represents a private key encrypted using keystore V4 format
type Keystore struct {
	PublicKey string              `json:"publicKey"`
	Crypto    keystore.CryptoJSON `json:"crypto"`
	UUID      string              `json:"uuid"`
	Version   int                 `json:"version"`
	CurveType string              `json:"curveType"` // Either "bls381" or "bn254"
}

// Options provides configuration options for keystore operations
type Options struct {
	// ScryptN is the N parameter of scrypt encryption algorithm
	ScryptN int
	// ScryptP is the P parameter of scrypt encryption algorithm
	ScryptP int
}

// Default returns the default options for keystore operations
func Default() *Options {
	return &Options{
		ScryptN: keystore.StandardScryptN,
		ScryptP: keystore.StandardScryptP,
	}
}

// Light returns light options for keystore operations (faster but less secure)
func Light() *Options {
	return &Options{
		ScryptN: keystore.LightScryptN,
		ScryptP: keystore.LightScryptP,
	}
}

// ParseKeystoreJSON takes a string representation of the keystore JSON and returns the Keystore struct
func ParseKeystoreJSON(keystoreJSON string) (*Keystore, error) {
	var keystore Keystore
	if err := json.Unmarshal([]byte(keystoreJSON), &keystore); err != nil {
		return nil, fmt.Errorf("failed to parse keystore JSON: %w", err)
	}

	// Verify it's a valid keystore by checking required fields
	if keystore.PublicKey == "" {
		return nil, ErrInvalidKeystoreFile
	}

	// Verify crypto field has required components
	if keystore.Crypto.Cipher == "" || keystore.Crypto.CipherText == "" ||
		keystore.Crypto.KDF == "" || len(keystore.Crypto.KDFParams) == 0 {
		return nil, fmt.Errorf("%w: missing required crypto fields", ErrInvalidKeystoreFile)
	}

	return &keystore, nil
}

// DetermineCurveType attempts to determine the curve type based on the private key
// This is a best-effort function that uses the curveStr path in the keygen operation
func DetermineCurveType(curveStr string) string {
	switch strings.ToLower(curveStr) {
	case "bls381":
		return "bls381"
	case "bn254":
		return "bn254"
	default:
		// Default to empty if we can't determine
		return ""
	}
}

// SaveToKeystore saves a private key to a keystore file using the Web3 Secret Storage format
func SaveToKeystore(privateKey signing.PrivateKey, filePath, password string, opts *Options) error {
	return SaveToKeystoreWithCurveType(privateKey, filePath, password, "", opts)
}

// SaveToKeystoreWithCurveType saves a private key to a keystore file using the Web3 Secret Storage format
// and includes the curve type in the keystore file
func SaveToKeystoreWithCurveType(privateKey signing.PrivateKey, filePath, password, curveType string, opts *Options) error {
	if opts == nil {
		opts = Default()
	}

	// Generate UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}

	// Get the public key
	publicKey := privateKey.Public()

	// Create the directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Encrypt the private key
	cryptoStruct, err := keystore.EncryptDataV3(
		privateKey.Bytes(),
		[]byte(password),
		opts.ScryptN,
		opts.ScryptP,
	)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Validate the curve type
	curveType = DetermineCurveType(curveType)

	// Create the keystore structure
	encryptedKey := Keystore{
		PublicKey: fmt.Sprintf("%x", publicKey.Bytes()),
		Crypto:    cryptoStruct,
		UUID:      id.String(),
		Version:   4,
		CurveType: curveType,
	}

	// Marshal to JSON
	content, err := json.MarshalIndent(encryptedKey, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keystore: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, content, 0600); err != nil {
		return fmt.Errorf("failed to write keystore file: %w", err)
	}

	return nil
}

// GetSigningScheme returns the appropriate signing scheme based on curve type
func GetSigningScheme(curveType string) (signing.SigningScheme, error) {
	switch strings.ToLower(curveType) {
	case "bls381":
		return bls381.NewScheme(), nil
	case "bn254":
		return bn254.NewScheme(), nil
	default:
		return nil, fmt.Errorf("unsupported curve type: %s", curveType)
	}
}

// LoadKeystoreFile loads a keystore from a file and returns the parsed Keystore struct
func LoadKeystoreFile(filePath string) (*Keystore, error) {
	// Read keystore file
	content, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %w", err)
	}

	// Parse and return the keystore
	return ParseKeystoreJSON(string(content))
}

// LoadPrivateKeyFromKeystore loads a private key from a keystore JSON string
func LoadPrivateKeyFromKeystore(keystoreJSON, password string, scheme signing.SigningScheme) (signing.PrivateKey, error) {
	// Parse keystore
	keystoreData, err := ParseKeystoreJSON(keystoreJSON)
	if err != nil {
		return nil, err
	}

	// Decrypt the private key
	keyBytes, err := keystore.DecryptDataV3(keystoreData.Crypto, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// If scheme is nil, try to determine the scheme from the curve type in the keystore
	if scheme == nil && keystoreData.CurveType != "" {
		scheme, err = GetSigningScheme(keystoreData.CurveType)
		if err != nil {
			return nil, fmt.Errorf("failed to determine signing scheme: %w", err)
		}
	}

	// If scheme is still nil, we can't proceed
	if scheme == nil {
		return nil, fmt.Errorf("no signing scheme provided and unable to determine from keystore")
	}

	// Recreate the private key using the provided scheme
	privateKey, err := scheme.NewPrivateKeyFromBytes(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key from decrypted data: %w", err)
	}

	return privateKey, nil
}

// LoadFromKeystore loads a private key from a keystore file (deprecated, use LoadKeystoreFile + LoadPrivateKeyFromKeystore instead)
func LoadFromKeystore(filePath, password string, scheme signing.SigningScheme) (signing.PrivateKey, error) {
	// Read keystore file
	content, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %w", err)
	}

	// Use the new function with the file content
	return LoadPrivateKeyFromKeystore(string(content), password, scheme)
}

// GetKeystoreInfo retrieves basic info from a keystore file without decrypting
func GetKeystoreInfo(filePath string) (*Keystore, error) {
	// Read keystore file
	content, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %w", err)
	}

	// Parse keystore
	return ParseKeystoreJSON(string(content))
}

// TestKeystore tests a keystore by signing a test message
func TestKeystore(filePath, password string, scheme signing.SigningScheme) error {
	// Load the keystore file
	keystoreData, err := LoadKeystoreFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load keystore file: %w", err)
	}

	// Convert the keystore to JSON
	keystoreJSON, err := json.Marshal(keystoreData)
	if err != nil {
		return fmt.Errorf("failed to marshal keystore data: %w", err)
	}

	// Load the private key from keystore
	privateKey, err := LoadPrivateKeyFromKeystore(string(keystoreJSON), password, scheme)
	if err != nil {
		return fmt.Errorf("failed to load private key from keystore: %w", err)
	}

	// Get the public key
	publicKey := privateKey.Public()

	// Test signing a message
	testMessage := []byte("Test message for keystore verification")
	sig, err := privateKey.Sign(testMessage)
	if err != nil {
		return fmt.Errorf("failed to sign test message: %w", err)
	}

	// Verify signature
	valid, err := sig.Verify(publicKey, testMessage)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	if !valid {
		return fmt.Errorf("keystore verification failed: signature is invalid")
	}

	return nil
}

// GenerateRandomPassword generates a cryptographically secure random password
func GenerateRandomPassword(length int) (string, error) {
	if length < 16 {
		length = 16 // Minimum password length for security
	}

	// Create a byte slice to hold the random password
	bytes := make([]byte, length)

	// Fill with random bytes
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Define character set (alphanumeric + special chars)
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"
	charsetLen := len(charset)

	// Convert random bytes to character set
	for i := 0; i < length; i++ {
		bytes[i] = charset[int(bytes[i])%charsetLen]
	}

	return string(bytes), nil
}
