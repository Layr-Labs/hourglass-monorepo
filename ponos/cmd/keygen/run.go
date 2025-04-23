package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bls381"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// encryptedKeyV4 represents a private key encrypted using keystore V4 format
type encryptedKeyV4 struct {
	PublicKey string              `json:"publicKey"`
	Crypto    keystore.CryptoJSON `json:"crypto"`
	UUID      string              `json:"uuid"`
	Version   int                 `json:"version"`
}

// keystoreOptions provides configuration options for keystore operations
type keystoreOptions struct {
	ScryptN int
	ScryptP int
}

// Default keystore options
func defaultKeystoreOptions() *keystoreOptions {
	return &keystoreOptions{
		ScryptN: keystore.StandardScryptN,
		ScryptP: keystore.StandardScryptP,
	}
}

// Light keystore options (faster but less secure)
func lightKeystoreOptions() *keystoreOptions {
	return &keystoreOptions{
		ScryptN: keystore.LightScryptN,
		ScryptP: keystore.LightScryptP,
	}
}

// generateRandomPassword generates a cryptographically secure random password
func generateRandomPassword(length int) (string, error) {
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

// saveToKeystore saves a private key to a keystore file using the Web3 Secret Storage format
func saveToKeystore(privateKey signing.PrivateKey, publicKey signing.PublicKey, filePath, password string, opts *keystoreOptions) error {
	// Generate UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}

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
	encryptedKey := encryptedKeyV4{
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

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new BLS key pair",
	RunE: func(cmd *cobra.Command, args []string) error {
		initRunCmd(cmd)

		l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: Config.Debug})

		l.Sugar().Infow("Generating key pair", "curve", Config.CurveType)

		// Create the output directory if it doesn't exist
		if err := os.MkdirAll(Config.OutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		var scheme signing.SigningScheme
		switch strings.ToLower(Config.CurveType) {
		case "bls381":
			scheme = bls381.NewScheme()
		case "bn254":
			scheme = bn254.NewScheme()
		default:
			return fmt.Errorf("unsupported curve type: %s", Config.CurveType)
		}

		var (
			privateKey signing.PrivateKey
			publicKey  signing.PublicKey
			err        error
		)

		// Check if a seed is provided
		if Config.Seed != "" {
			seedBytes, err := hex.DecodeString(Config.Seed)
			if err != nil {
				return fmt.Errorf("invalid seed format: %w", err)
			}

			// Check if a path is provided for EIP-2333
			if Config.Path != "" && strings.ToLower(Config.CurveType) == "bls381" {
				var path []uint32
				for _, segment := range strings.Split(Config.Path, "/") {
					if segment == "" || segment == "m" {
						continue
					}
					var value uint32
					if _, err := fmt.Sscanf(segment, "%d", &value); err != nil {
						return fmt.Errorf("invalid path segment '%s': %w", segment, err)
					}
					path = append(path, value)
				}
				privateKey, publicKey, err = scheme.GenerateKeyPairEIP2333(seedBytes, path)
				if err != nil {
					return fmt.Errorf("failed to generate key pair with EIP-2333: %w", err)
				}
			} else {
				privateKey, publicKey, err = scheme.GenerateKeyPairFromSeed(seedBytes)
				if err != nil {
					return fmt.Errorf("failed to generate key pair from seed: %w", err)
				}
			}
		} else {
			// Generate a random key pair
			privateKey, publicKey, err = scheme.GenerateKeyPair()
			if err != nil {
				return fmt.Errorf("failed to generate key pair: %w", err)
			}
		}

		// Get the password to use for keystore
		var password string
		if Config.UseRandomPassword {
			// Generate a random password if requested
			var err error
			password, err = generateRandomPassword(32)
			if err != nil {
				return fmt.Errorf("failed to generate random password: %w", err)
			}
			l.Sugar().Infow("Generated random password", "password", password)
		} else {
			password = Config.Password
		}

		// Determine keystore options
		var keystoreOpts *keystoreOptions
		if Config.LightEncryption {
			keystoreOpts = lightKeystoreOptions()
			l.Sugar().Warn("Using light encryption - this is less secure but faster")
		} else {
			keystoreOpts = defaultKeystoreOptions()
		}

		// Curve type determines the file naming
		curveStr := strings.ToLower(Config.CurveType)

		// Save the keys in the appropriate format
		if Config.UseKeystore {
			// Save using Web3 Secret Storage format
			keystorePath := filepath.Join(Config.OutputDir, fmt.Sprintf("%s_%s.json", Config.FilePrefix, curveStr))
			if err := saveToKeystore(privateKey, publicKey, keystorePath, password, keystoreOpts); err != nil {
				return fmt.Errorf("failed to save keystore: %w", err)
			}
			l.Sugar().Infow(fmt.Sprintf("Generated %s keys in keystore format", strings.ToUpper(curveStr)),
				"keystoreFile", keystorePath,
				"publicKey", hex.EncodeToString(publicKey.Bytes()))
		} else {
			// Save in raw format
			privFilePath := filepath.Join(Config.OutputDir, fmt.Sprintf("%s_%s.pri", Config.FilePrefix, curveStr))
			pubFilePath := filepath.Join(Config.OutputDir, fmt.Sprintf("%s_%s.pub", Config.FilePrefix, curveStr))

			if err := os.WriteFile(privFilePath, privateKey.Bytes(), 0600); err != nil {
				return fmt.Errorf("failed to write private key: %w", err)
			}

			if err := os.WriteFile(pubFilePath, publicKey.Bytes(), 0644); err != nil {
				return fmt.Errorf("failed to write public key: %w", err)
			}

			l.Sugar().Infow(fmt.Sprintf("Generated %s keys in raw format", strings.ToUpper(curveStr)),
				"privateKeyFile", privFilePath,
				"publicKeyFile", pubFilePath)
		}

		return nil
	},
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display information about a BLS key",
	RunE: func(cmd *cobra.Command, args []string) error {
		initRunCmd(cmd)

		l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: Config.Debug})

		keyFile := Config.KeyFile
		fmt.Printf("cofnig: %+v\n", Config)
		if keyFile == "" {
			return fmt.Errorf("key file path is required")
		}

		l.Sugar().Infow("Reading key file", "file", keyFile)

		// Check if the file might be a keystore
		if strings.HasSuffix(keyFile, ".json") {
			// Try to parse as a keystore (without decrypting)
			content, err := os.ReadFile(keyFile)
			if err != nil {
				return fmt.Errorf("failed to read key file: %w", err)
			}

			var keyData struct {
				PublicKey string `json:"publicKey"`
				UUID      string `json:"uuid"`
				Version   int    `json:"version"`
			}

			if err := json.Unmarshal(content, &keyData); err == nil && keyData.PublicKey != "" {
				// This appears to be a valid keystore
				l.Sugar().Infow("Key Information",
					"type", "keystore",
					"curve", Config.CurveType,
					"publicKey", keyData.PublicKey,
					"uuid", keyData.UUID,
					"version", keyData.Version,
				)
				return nil
			}
		}

		// If not a keystore (or couldn't parse as one), try as raw key
		keyData, err := os.ReadFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}

		var scheme signing.SigningScheme
		switch strings.ToLower(Config.CurveType) {
		case "bls381":
			scheme = bls381.NewScheme()
		case "bn254":
			scheme = bn254.NewScheme()
		default:
			return fmt.Errorf("unsupported curve type: %s", Config.CurveType)
		}

		// Try to load as private key
		privateKey, err := scheme.NewPrivateKeyFromBytes(keyData)
		if err == nil {
			publicKey := privateKey.Public()
			l.Sugar().Infow("Key Information",
				"type", "private key",
				"curve", Config.CurveType,
				"publicKey", hex.EncodeToString(publicKey.Bytes()),
				"privateKey", hex.EncodeToString(privateKey.Bytes()),
			)
			return nil
		}

		// Try to load as public key
		publicKey, err := scheme.NewPublicKeyFromBytes(keyData)
		if err == nil {
			l.Sugar().Infow("Key Information",
				"type", "public key",
				"curve", Config.CurveType,
				"publicKey", hex.EncodeToString(publicKey.Bytes()),
			)
			return nil
		}

		return fmt.Errorf("could not parse key as either private or public key or keystore for curve %s", Config.CurveType)
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test a keystore by signing a test message",
	RunE: func(cmd *cobra.Command, args []string) error {
		initRunCmd(cmd)

		l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: Config.Debug})

		keyFile := Config.KeyFile
		if keyFile == "" {
			return fmt.Errorf("key file path is required")
		}

		password := Config.Password
		if password == "" {
			return fmt.Errorf("password is required to decrypt the keystore")
		}

		l.Sugar().Infow("Testing keystore", "file", keyFile)

		// Check if the file is a keystore
		if !strings.HasSuffix(keyFile, ".json") {
			return fmt.Errorf("file must be a keystore JSON file")
		}

		var scheme signing.SigningScheme
		switch strings.ToLower(Config.CurveType) {
		case "bls381":
			scheme = bls381.NewScheme()
		case "bn254":
			scheme = bn254.NewScheme()
		default:
			return fmt.Errorf("unsupported curve type: %s", Config.CurveType)
		}

		// Read keystore file
		content, err := os.ReadFile(filepath.Clean(keyFile))
		if err != nil {
			return fmt.Errorf("failed to read keystore file: %w", err)
		}

		// Parse keystore to get public key (without decrypting)
		var keyData struct {
			PublicKey string              `json:"publicKey"`
			Crypto    keystore.CryptoJSON `json:"crypto"`
			UUID      string              `json:"uuid"`
			Version   int                 `json:"version"`
		}

		if err := json.Unmarshal(content, &keyData); err != nil {
			return fmt.Errorf("failed to parse keystore file: %w", err)
		}

		// Decrypt the private key
		keyBytes, err := keystore.DecryptDataV3(keyData.Crypto, password)
		if err != nil {
			return fmt.Errorf("failed to decrypt private key: %w", err)
		}

		// Recreate the private key
		privateKey, err := scheme.NewPrivateKeyFromBytes(keyBytes)
		if err != nil {
			return fmt.Errorf("failed to create private key from decrypted data: %w", err)
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

		if valid {
			l.Sugar().Infow("Keystore test successful",
				"curve", Config.CurveType,
				"publicKey", hex.EncodeToString(publicKey.Bytes()),
				"signature", hex.EncodeToString(sig.Bytes()),
			)
		} else {
			return fmt.Errorf("keystore verification failed: signature is invalid")
		}

		return nil
	},
}

func initRunCmd(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err := viper.BindPFlag(config.KebabToSnakeCase(f.Name), f); err != nil {
			fmt.Printf("Failed to bind flag '%s' - %+v\n", f.Name, err)
		}
		if err := viper.BindEnv(f.Name); err != nil {
			fmt.Printf("Failed to bind env '%s' - %+v\n", f.Name, err)
		}
	})
}
