package keystore

import (
	"crypto/aes"
	cryptoCipher "crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Layr-Labs/crypto-libs/pkg/keystore/legacy"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/text/unicode/norm"

	"github.com/Layr-Labs/crypto-libs/pkg/bls381"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/google/uuid"
)

// Package keystore implements an EIP-2335 compliant keystore for BLS private keys.
// It provides support for both BLS12-381 and BN254 curve types with the following features:
//
// 1. EIP-2335 compliance for standardized keystore format
// 2. Backward compatibility with legacy keystore format
// 3. Multiple KDF support (scrypt and pbkdf2)
// 4. AES-128-CTR encryption
// 5. Password processing according to EIP-2335 spec (NFKD normalization, control code stripping)
//
// The keystore format follows the EIP-2335 specification with crypto modules for KDF, checksum,
// and cipher operations while adding a custom "curveType" field to support both BLS12-381 and
// BN254 curve types.

// ErrInvalidKeystoreFile is returned when a keystore file is not valid or is corrupted
var ErrInvalidKeystoreFile = errors.New("invalid keystore file")

// Module represents a cryptographic module in EIP-2335
type Module struct {
	Function string                 `json:"function"`
	Params   map[string]interface{} `json:"params"`
	Message  string                 `json:"message"`
}

// EIP2335Keystore represents a BLS private key encrypted using EIP-2335 format
type EIP2335Keystore struct {
	Crypto struct {
		KDF      Module `json:"kdf"`
		Checksum Module `json:"checksum"`
		Cipher   Module `json:"cipher"`
	} `json:"crypto"`
	Description string `json:"description,omitempty"`
	Pubkey      string `json:"pubkey"`
	Path        string `json:"path"`
	UUID        string `json:"uuid"`
	Version     int    `json:"version"`
	CurveType   string `json:"curveType,omitempty"` // Custom field, either "bls381" or "bn254"
}

// LegacyKeystore represents the old keystore format
// type LegacyKeystore struct {
// 	PublicKey string              `json:"publicKey"`
// 	Crypto    keystore.CryptoJSON `json:"crypto"`
// 	UUID      string              `json:"uuid"`
// 	Version   int                 `json:"version"`
// 	CurveType string              `json:"curveType"`
// }

// processPassword prepares a password according to EIP-2335:
// 1. Convert to NFKD representation
// 2. Strip control codes (C0, C1, and Delete)
// 3. UTF-8 encode (handled by Go strings)
func processPassword(password string) []byte {
	// Step 1: Convert to NFKD representation
	normalized := norm.NFKD.String(password)

	// Step 2: Strip control codes
	var cleaned []rune
	for _, r := range normalized {
		// Skip C0 (0x00-0x1F), C1 (0x80-0x9F), and Delete (0x7F)
		if (r >= 0x00 && r <= 0x1F) || (r >= 0x80 && r <= 0x9F) || r == 0x7F {
			continue
		}
		cleaned = append(cleaned, r)
	}

	// Return the UTF-8 encoded string
	return []byte(string(cleaned))
}

// deriveKeyFromPassword derives a key from the password using the specified KDF
func deriveKeyFromPassword(password string, kdf Module) ([]byte, error) {
	processedPassword := processPassword(password)

	switch kdf.Function {
	case "pbkdf2":
		// Extract parameters
		salt, err := hex.DecodeString(kdf.Params["salt"].(string))
		if err != nil {
			return nil, fmt.Errorf("invalid salt: %w", err)
		}

		c, ok := kdf.Params["c"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid iterations count")
		}

		dklen, ok := kdf.Params["dklen"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid dklen")
		}

		prf, ok := kdf.Params["prf"].(string)
		if !ok || prf != "hmac-sha256" {
			return nil, fmt.Errorf("unsupported PRF: %v", prf)
		}

		// EIP-2335 parameter validation for PBKDF2
		if len(salt) < 16 {
			return nil, fmt.Errorf("salt too short: must be at least 16 bytes, got %d", len(salt))
		}
		if len(salt) > 64 {
			return nil, fmt.Errorf("salt too long: must be at most 64 bytes, got %d", len(salt))
		}

		// Iteration count validation - EIP-2335 reference value is 262144
		if int(c) < 1000 {
			return nil, fmt.Errorf("iteration count too low: must be at least 1000, got %d", int(c))
		}
		if int(c) > 10000000 {
			return nil, fmt.Errorf("iteration count too high: must be at most 10000000, got %d", int(c))
		}

		// Derived key length validation
		if int(dklen) != 32 {
			return nil, fmt.Errorf("invalid dklen: EIP-2335 requires 32 bytes, got %d", int(dklen))
		}

		return pbkdf2.Key(processedPassword, salt, int(c), int(dklen), sha256.New), nil

	case "scrypt":
		// Extract parameters
		salt, err := hex.DecodeString(kdf.Params["salt"].(string))
		if err != nil {
			return nil, fmt.Errorf("invalid salt: %w", err)
		}

		n, ok := kdf.Params["n"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid N parameter")
		}

		r, ok := kdf.Params["r"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid r parameter")
		}

		p, ok := kdf.Params["p"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid p parameter")
		}

		dklen, ok := kdf.Params["dklen"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid dklen")
		}

		// EIP-2335 parameter validation for scrypt
		if len(salt) < 16 {
			return nil, fmt.Errorf("salt too short: must be at least 16 bytes, got %d", len(salt))
		}
		if len(salt) > 64 {
			return nil, fmt.Errorf("salt too long: must be at most 64 bytes, got %d", len(salt))
		}

		// N parameter validation - must be a power of 2
		nInt := int(n)
		if nInt < 1024 {
			return nil, fmt.Errorf("N parameter too low: must be at least 1024, got %d", nInt)
		}
		if nInt > 1048576 { // 2^20, reasonable upper bound
			return nil, fmt.Errorf("N parameter too high: must be at most 1048576, got %d", nInt)
		}
		if nInt&(nInt-1) != 0 {
			return nil, fmt.Errorf("N parameter must be a power of 2, got %d", nInt)
		}

		// r parameter validation - EIP-2335 reference value is 8
		rInt := int(r)
		if rInt < 1 {
			return nil, fmt.Errorf("r parameter too low: must be at least 1, got %d", rInt)
		}
		if rInt > 32 {
			return nil, fmt.Errorf("r parameter too high: must be at most 32, got %d", rInt)
		}

		// p parameter validation - EIP-2335 reference value is 1
		pInt := int(p)
		if pInt < 1 {
			return nil, fmt.Errorf("p parameter too low: must be at least 1, got %d", pInt)
		}
		if pInt > 16 {
			return nil, fmt.Errorf("p parameter too high: must be at most 16, got %d", pInt)
		}

		// Derived key length validation
		if int(dklen) != 32 {
			return nil, fmt.Errorf("invalid dklen: EIP-2335 requires 32 bytes, got %d", int(dklen))
		}

		// Memory usage validation - prevent excessive memory consumption
		// Memory usage is approximately 128 * N * r bytes
		memoryUsage := 128 * nInt * rInt
		if memoryUsage > 1024*1024*1024 { // 1GB limit
			return nil, fmt.Errorf("scrypt parameters would require too much memory: %d bytes (max 1GB)", memoryUsage)
		}

		return scrypt.Key(processedPassword, salt, int(n), int(r), int(p), int(dklen))

	default:
		return nil, fmt.Errorf("unsupported KDF function: %s", kdf.Function)
	}
}

// verifyPassword checks if the provided password is correct
func verifyPassword(decryptionKey []byte, checksum Module, cipherMessage string) (bool, error) {
	if checksum.Function != "sha256" {
		return false, fmt.Errorf("unsupported checksum function: %s", checksum.Function)
	}

	// Get the second 16 bytes of the decryption key
	dkSlice := decryptionKey[16:32]

	// Decode the cipher message
	cipherBytes, err := hex.DecodeString(cipherMessage)
	if err != nil {
		return false, fmt.Errorf("invalid cipher message: %w", err)
	}

	// Create the pre-image: DK_slice | cipher_message
	preImage := append(dkSlice, cipherBytes...)

	// Calculate the checksum
	calculatedChecksum := sha256.Sum256(preImage)
	checksumHex := hex.EncodeToString(calculatedChecksum[:])

	// Compare with the stored checksum
	return checksumHex == checksum.Message, nil
}

// decryptSecret decrypts the encrypted private key
func decryptSecret(decryptionKey []byte, cipher Module) ([]byte, error) {
	if cipher.Function != "aes-128-ctr" {
		return nil, fmt.Errorf("unsupported cipher function: %s", cipher.Function)
	}

	// Decode the IV and message
	iv, err := hex.DecodeString(cipher.Params["iv"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid IV: %w", err)
	}

	cipherText, err := hex.DecodeString(cipher.Message)
	if err != nil {
		return nil, fmt.Errorf("invalid cipher text: %w", err)
	}

	// Use only the first 16 bytes of the decryption key for AES-128
	key := decryptionKey[:16]

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create the CTR mode
	ctr := cryptoCipher.NewCTR(block, iv)

	// Decrypt the cipher text
	plainText := make([]byte, len(cipherText))
	ctr.XORKeyStream(plainText, cipherText)

	return plainText, nil
}

// GetPrivateKey decrypts and returns the private key from the keystore
func (k *EIP2335Keystore) GetPrivateKey(password string, scheme signing.SigningScheme) (signing.PrivateKey, error) {
	if k == nil {
		return nil, fmt.Errorf("keystore data cannot be nil")
	}

	// Derive decryption key from password
	decryptionKey, err := deriveKeyFromPassword(password, k.Crypto.KDF)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	// Verify password
	valid, err := verifyPassword(decryptionKey, k.Crypto.Checksum, k.Crypto.Cipher.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("invalid password")
	}

	// Decrypt the private key
	keyBytes, err := decryptSecret(decryptionKey, k.Crypto.Cipher)
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

// GetBN254PrivateKey gets a BN254 private key from the keystore
func (k *EIP2335Keystore) GetBN254PrivateKey(password string) (*bn254.PrivateKey, error) {
	if k == nil {
		return nil, fmt.Errorf("keystore data cannot be nil")
	}

	scheme, err := GetSigningSchemeForCurveType("bn254")
	if err != nil {
		return nil, err
	}

	privKey, err := k.GetPrivateKey(password, scheme)
	if err != nil {
		return nil, err
	}

	// Try to use the UnwrapPrivateKey method if available
	type unwrapper interface {
		UnwrapPrivateKey() *bn254.PrivateKey
	}

	if unwrapper, ok := privKey.(unwrapper); ok {
		return unwrapper.UnwrapPrivateKey(), nil
	}

	// Fall back to recreating from bytes if unwrapper not available
	rawBytes := privKey.Bytes()
	if len(rawBytes) > 0 {
		bn254PrivKey, err := bn254.NewPrivateKeyFromBytes(rawBytes)
		if err == nil {
			return bn254PrivKey, nil
		}
		return nil, fmt.Errorf("failed to create BN254 private key from bytes: %w", err)
	}

	return nil, fmt.Errorf("private key is not of compatible bn254 type or cannot be converted")
}

// GetBLS381PrivateKey gets a BLS381 private key from the keystore
func (k *EIP2335Keystore) GetBLS381PrivateKey(password string) (*bls381.PrivateKey, error) {
	if k == nil {
		return nil, fmt.Errorf("keystore data cannot be nil")
	}

	scheme, err := GetSigningSchemeForCurveType("bls381")
	if err != nil {
		return nil, err
	}

	privKey, err := k.GetPrivateKey(password, scheme)
	if err != nil {
		return nil, err
	}

	// Try to use an unwrapper method if available
	type unwrapper interface {
		UnwrapPrivateKey() *bls381.PrivateKey
	}

	if unwrapper, ok := privKey.(unwrapper); ok {
		return unwrapper.UnwrapPrivateKey(), nil
	}

	// Fall back to recreating from bytes if unwrapper not available
	rawBytes := privKey.Bytes()
	if len(rawBytes) > 0 {
		bls381PrivKey, err := bls381.NewPrivateKeyFromBytes(rawBytes)
		if err == nil {
			return bls381PrivKey, nil
		}
		return nil, fmt.Errorf("failed to create BLS381 private key from bytes: %w", err)
	}

	return nil, fmt.Errorf("private key is not of compatible bls381 type or cannot be converted")
}

// Options provides configuration options for keystore operations
type Options struct {
	// ScryptN is the N parameter of scrypt encryption algorithm
	ScryptN int
	// ScryptP is the P parameter of scrypt encryption algorithm
	ScryptP int
	// ScryptR is the R parameter of scrypt encryption algorithm (added for EIP-2335)
	ScryptR int
	// KDFType selects which KDF to use ("scrypt" or "pbkdf2")
	KDFType string
	// Description is an optional description for the keystore
	Description string
}

// Default returns the default options for keystore operations
func Default() *Options {
	return &Options{
		ScryptN:     262144, // EIP-2335 reference value
		ScryptP:     1,      // EIP-2335 reference value
		ScryptR:     8,      // EIP-2335 reference value
		KDFType:     "scrypt",
		Description: "",
	}
}

func ParseLegacyKeystoreToEIP2335Keystore(legacyJSON string, password string, scheme signing.SigningScheme) (*EIP2335Keystore, error) {
	lks, err := legacy.ParseKeystoreJSON(legacyJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse legacy keystore: %w", err)
	}
	pk, err := lks.GetPrivateKey(password, scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key from legacy keystore: %w", err)
	}

	// Convert legacy format to EIP-2335 format
	return GenerateKeystore(pk, password, lks.CurveType, Default())
}

// ParseKeystoreJSON takes a string representation of the keystore JSON and returns the EIP2335Keystore struct
func ParseKeystoreJSON(keystoreJSON string) (*EIP2335Keystore, error) {
	// Check for empty or whitespace-only input
	if strings.TrimSpace(keystoreJSON) == "" || strings.TrimSpace(keystoreJSON) == "{}" {
		return nil, ErrInvalidKeystoreFile
	}

	var ks EIP2335Keystore
	if err := json.Unmarshal([]byte(keystoreJSON), &ks); err != nil {
		return nil, fmt.Errorf("failed to parse keystore JSON: %w", err)
	}

	// Verify it's a valid keystore by checking required fields
	// An EIP-2335 compliant keystore must have either:
	// 1. A valid pubkey field (non-empty)
	// 2. A valid crypto object with proper KDF function
	if ks.Pubkey == "" || ks.Crypto.KDF.Function == "" {
		return nil, ErrInvalidKeystoreFile
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

// generateRandomIV generates a random IV for AES encryption
func generateRandomIV() ([]byte, error) {
	iv := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	return iv, nil
}

// generateRandomSalt generates a random salt for KDF
func generateRandomSalt() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func GenerateKeystore(privateKey signing.PrivateKey, password, curveType string, opts *Options) (*EIP2335Keystore, error) {
	if opts == nil {
		opts = Default()
	}

	// Generate UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}

	// Get the public key
	publicKey := privateKey.Public()
	pubkeyHex := hex.EncodeToString(publicKey.Bytes())

	// Validate the curve type
	curveType = DetermineCurveType(curveType)
	if curveType == "" {
		return nil, fmt.Errorf("invalid curve type")
	}

	// Generate salt and IV
	salt, err := generateRandomSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	iv, err := generateRandomIV()
	if err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Process password
	processedPassword := processPassword(password)

	// Set up KDF parameters
	var decryptionKey []byte
	var kdfModule Module

	if opts.KDFType == "pbkdf2" {
		// PBKDF2 parameters
		kdfModule = Module{
			Function: "pbkdf2",
			Params: map[string]interface{}{
				"dklen": float64(32),
				"c":     float64(262144), // Iterations
				"prf":   "hmac-sha256",
				"salt":  hex.EncodeToString(salt),
			},
			Message: "",
		}
		decryptionKey = pbkdf2.Key(processedPassword, salt, 262144, 32, sha256.New)
	} else {
		// Default to scrypt
		kdfModule = Module{
			Function: "scrypt",
			Params: map[string]interface{}{
				"dklen": float64(32),
				"n":     float64(opts.ScryptN),
				"r":     float64(opts.ScryptR),
				"p":     float64(opts.ScryptP),
				"salt":  hex.EncodeToString(salt),
			},
			Message: "",
		}
		decryptionKey, err = scrypt.Key(processedPassword, salt, opts.ScryptN, opts.ScryptR, opts.ScryptP, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to derive key: %w", err)
		}
	}

	// Encrypt the private key
	// Use only the first 16 bytes of the decryption key for AES-128
	key := decryptionKey[:16]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Encrypt the private key
	privateKeyBytes := privateKey.Bytes()
	cipherText := make([]byte, len(privateKeyBytes))
	ctrCipher := cryptoCipher.NewCTR(block, iv)
	ctrCipher.XORKeyStream(cipherText, privateKeyBytes)

	// Set up cipher module
	cipherModule := Module{
		Function: "aes-128-ctr",
		Params: map[string]interface{}{
			"iv": hex.EncodeToString(iv),
		},
		Message: hex.EncodeToString(cipherText),
	}

	// Create checksum
	// Get the second 16 bytes of the decryption key
	dkSlice := decryptionKey[16:32]
	preImage := append(dkSlice, cipherText...)
	checksum := sha256.Sum256(preImage)

	// Set up checksum module
	checksumModule := Module{
		Function: "sha256",
		Params:   map[string]interface{}{},
		Message:  hex.EncodeToString(checksum[:]),
	}

	// Create path based on curve type
	var path string
	if curveType == "bls381" {
		path = "m/12381/60/0/0" // Standard path for BLS12-381
	} else {
		path = "m/1/0/0" // Simple path for BN254 (non-standard)
	}

	// Create the ks structure
	ks := &EIP2335Keystore{
		Pubkey:      pubkeyHex,
		UUID:        id.String(),
		Version:     4,
		CurveType:   curveType,
		Path:        path,
		Description: opts.Description,
	}
	ks.Crypto.KDF = kdfModule
	ks.Crypto.Checksum = checksumModule
	ks.Crypto.Cipher = cipherModule

	return ks, nil
}

// SaveToKeystoreWithCurveType saves a private key to a keystore file using the EIP-2335 format
// and includes the curve type in the keystore file
func SaveToKeystoreWithCurveType(privateKey signing.PrivateKey, filePath, password, curveType string, opts *Options) error {
	ks, err := GenerateKeystore(privateKey, password, curveType, opts)
	if err != nil {
		return fmt.Errorf("failed to generate keystore: %w", err)
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON
	content, err := json.MarshalIndent(ks, "", "  ")
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

// LoadKeystoreFile loads a keystore from a file and returns the parsed EIP2335Keystore struct
func LoadKeystoreFile(filePath string) (*EIP2335Keystore, error) {
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
