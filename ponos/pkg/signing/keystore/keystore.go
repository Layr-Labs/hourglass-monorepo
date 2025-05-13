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

// GetPrivateKey decrypts and returns the private key from the keystore
func (k *Keystore) GetPrivateKey(password string, scheme signing.SigningScheme) (signing.PrivateKey, error) {
	if k == nil {
		return nil, fmt.Errorf("keystore data cannot be nil")
	}

	// Decrypt the private key
	keyBytes, err := keystore.DecryptDataV3(k.Crypto, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// If scheme is nil, try to determine the scheme from the curve type in the keystore
	if scheme == nil && k.CurveType != "" {
		scheme, err = GetSigningSchemeForCurveType(k.CurveType)
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

func (k *Keystore) GetBN254PrivateKey(password string) (*bn254.PrivateKey, error) {
	if k == nil {
		return nil, fmt.Errorf("keystore data cannot be nil")
	}

	// Decrypt the private key
	keyBytes, err := keystore.DecryptDataV3(k.Crypto, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	if k.CurveType != "bn254" {
		return nil, fmt.Errorf("keystore curve type is not bn254")
	}

	// Recreate the private key using the provided scheme
	privateKey, err := bn254.NewPrivateKeyFromBytes(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key from decrypted data: %w", err)
	}

	return privateKey, nil
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
	var ks Keystore
	if err := json.Unmarshal([]byte(keystoreJSON), &ks); err != nil {
		return nil, fmt.Errorf("failed to parse ks JSON: %w", err)
	}

	// Verify it's a valid ks by checking required fields
	if ks.PublicKey == "" {
		return nil, ErrInvalidKeystoreFile
	}

	// Verify crypto field has required components
	if ks.Crypto.Cipher == "" || ks.Crypto.CipherText == "" ||
		ks.Crypto.KDF == "" || len(ks.Crypto.KDFParams) == 0 {
		return nil, fmt.Errorf("%w: missing required crypto fields", ErrInvalidKeystoreFile)
	}

	return &ks, nil
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

// GetSigningSchemeForCurveType returns the appropriate signing scheme based on curve type
func GetSigningSchemeForCurveType(curveType string) (signing.SigningScheme, error) {
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

// TestKeystore tests a keystore by signing a test message
func TestKeystore(filePath, password string, scheme signing.SigningScheme) error {
	// Load the keystore file
	keystoreData, err := LoadKeystoreFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load keystore file: %w", err)
	}

	// Load the private key from keystore
	privateKey, err := keystoreData.GetPrivateKey(password, scheme)
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
