package bn254

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/google/uuid"
)

// ErrInvalidKeystoreFile is returned when a keystore file is not valid or is corrupted
var ErrInvalidKeystoreFile = errors.New("invalid keystore file")

// encryptedBN254KeyV4 represents a BN254 private key encrypted using keystore V4 format
type encryptedBN254KeyV4 struct {
	PublicKey string              `json:"publicKey"`
	Crypto    keystore.CryptoJSON `json:"crypto"`
	UUID      string              `json:"uuid"`
	Version   int                 `json:"version"`
}

// KeystoreOptions provides configuration options for keystore operations
type KeystoreOptions struct {
	// ScryptN is the N parameter of scrypt encryption algorithm
	ScryptN int
	// ScryptP is the P parameter of scrypt encryption algorithm
	ScryptP int
}

// DefaultKeystoreOptions returns the default options for keystore operations
func DefaultKeystoreOptions() *KeystoreOptions {
	return &KeystoreOptions{
		ScryptN: keystore.StandardScryptN,
		ScryptP: keystore.StandardScryptP,
	}
}

// LightKeystoreOptions returns light options for keystore operations (faster but less secure)
func LightKeystoreOptions() *KeystoreOptions {
	return &KeystoreOptions{
		ScryptN: keystore.LightScryptN,
		ScryptP: keystore.LightScryptP,
	}
}

// SaveToKeystore saves a private key to a keystore file
func SaveToKeystore(privateKey *PrivateKey, filePath, password string, opts *KeystoreOptions) error {
	if opts == nil {
		opts = DefaultKeystoreOptions()
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

	// Create the keystore structure
	encryptedKey := encryptedBN254KeyV4{
		PublicKey: fmt.Sprintf("%x", publicKey.Bytes()),
		Crypto:    cryptoStruct,
		UUID:      id.String(),
		Version:   4,
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

// LoadFromKeystore loads a private key from a keystore file
func LoadFromKeystore(filePath, password string) (*PrivateKey, error) {
	// Read keystore file
	content, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %w", err)
	}

	// Unmarshal JSON
	var encryptedKey encryptedBN254KeyV4
	if err := json.Unmarshal(content, &encryptedKey); err != nil {
		return nil, fmt.Errorf("failed to parse keystore file: %w", err)
	}

	// Verify it's a BN254 keystore
	if encryptedKey.PublicKey == "" {
		return nil, ErrInvalidKeystoreFile
	}

	// Decrypt the private key
	keyBytes, err := keystore.DecryptDataV3(encryptedKey.Crypto, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// Recreate the private key
	privateKey, err := NewPrivateKeyFromBytes(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key from decrypted data: %w", err)
	}

	return privateKey, nil
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
