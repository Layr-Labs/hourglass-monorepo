package keystore

import (
	"bufio"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	blskeystore "github.com/Layr-Labs/crypto-libs/pkg/keystore"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"golang.org/x/term"
)

func createCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new BN254 or ECDSA keystore",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Name for the keystore",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "type",
				Usage:    "Keystore type (bn254 for BN254 or ecdsa)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "key",
				Usage: "Private key (BN254 in large number format or ECDSA in hex format). If not provided, a new key will be generated",
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "Password to encrypt the keystore file",
				Value: "",
			},
		},
		Action: func(c *cli.Context) error {
			log := config.LoggerFromContext(c.Context)

			name := c.String("name")
			keystoreType := c.String("type")
			privateKey := c.String("key")
			password := c.String("password")

			// Validate keystore type
			if keystoreType != "bn254" && keystoreType != "ecdsa" {
				log.Error("invalid key store type", zap.String("type", keystoreType))
				return fmt.Errorf("unsupported keystore type: %s (supported: bn254, ecdsa)", keystoreType)
			}

			// Load config first to validate name uniqueness
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get context
			ctx, exists := cfg.Contexts[cfg.CurrentContext]
			if !exists {
				return fmt.Errorf("context %s not found", cfg.CurrentContext)
			}

			// Check if keystore with same name already exists in the context
			for _, ks := range ctx.Keystores {
				if ks.Name == name {
					return fmt.Errorf("keystore with name '%s' already exists in context '%s'", name, cfg.CurrentContext)
				}
			}

			// Prompt for password if not provided
			if password == "" {
				var err error
				password, err = promptForPasswordWithConfirmation("Enter password for keystore: ")
				if err != nil {
					return err
				}
			}

			// Create keystore directory with absolute path
			keystoreDir := filepath.Join(config.GetConfigDir(), cfg.CurrentContext, "keystores", name)
			if err := os.MkdirAll(keystoreDir, 0700); err != nil {
				return fmt.Errorf("failed to create keystore directory: %w", err)
			}

			// Ensure we have an absolute path
			keystorePath, err := filepath.Abs(filepath.Join(keystoreDir, "key.json"))
			if err != nil {
				return fmt.Errorf("failed to get absolute keystore path: %w", err)
			}

			// Check if keystore file already exists on disk
			if _, err := os.Stat(keystorePath); err == nil {
				return fmt.Errorf("keystore file already exists at %s", keystorePath)
			}

			log.Info("Creating keystore",
				zap.String("context", cfg.CurrentContext),
				zap.String("name", name),
				zap.String("type", keystoreType),
				zap.String("path", keystorePath))

			switch keystoreType {
			case "bn254":
				if err := createBLSKeystore(log, privateKey, keystorePath, password); err != nil {
					return err
				}
			case "ecdsa":
				if err := createECDSAKeystore(log, privateKey, keystorePath, password); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported keystore type: %s", keystoreType)
			}

			// Add keystore reference
			ctx.Keystores = append(ctx.Keystores, signer.KeystoreReference{
				Name: name,
				Path: keystorePath,
				Type: keystoreType,
			})

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			log.Info("✅ Keystore created and registered",
				zap.String("context", cfg.CurrentContext),
				zap.String("name", name))

			return nil
		},
	}
}

func createBLSKeystore(log logger.Logger, privateKey, path, password string) error {
	scheme := bn254.NewScheme()

	var privKey signing.PrivateKey
	var err error

	if privateKey == "" {
		// Generate new key
		log.Info("Generating new BN254 private key")
		privKey, _, err = scheme.GenerateKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate BN254 private key: %w", err)
		}
	} else {
		// Use provided key
		cleanedKey := strings.TrimPrefix(privateKey, "0x")
		// For BN254, we need to use bytes instead of hex string
		keyBytes, err := hex.DecodeString(cleanedKey)
		if err != nil {
			return fmt.Errorf("failed to decode private key hex: %w", err)
		}
		privKey, err = scheme.NewPrivateKeyFromBytes(keyBytes)
		if err != nil {
			return fmt.Errorf("failed to create private key from bytes: %w", err)
		}
	}

	// Save to keystore
	err = blskeystore.SaveToKeystoreWithCurveType(privKey, path, password, "bn254", blskeystore.Default())
	if err != nil {
		return fmt.Errorf("failed to create keystore: %w", err)
	}

	// Validate by reloading
	keystoreData, err := blskeystore.LoadKeystoreFile(path)
	if err != nil {
		return fmt.Errorf("failed to reload keystore: %w", err)
	}

	privateKeyData, err := keystoreData.GetPrivateKey(password, scheme)
	if err != nil {
		return errors.New("failed to extract the private key from the keystore file")
	}

	log.Info("✅ BN254 keystore created successfully",
		zap.String("path", path))
	log.Info("🔑 Save this BN254 private key in a secure location:",
		zap.String("privateKey", string(privateKeyData.Bytes())))

	return nil
}

func createECDSAKeystore(log logger.Logger, privateKeyHex, path, password string) error {
	var privateKey *ecdsa.PrivateKey
	var err error

	if privateKeyHex == "" {
		// Generate new key
		log.Info("Generating new ECDSA private key")
		privateKey, err = crypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("failed to generate ECDSA private key: %w", err)
		}
	} else {
		// Use provided key
		cleanedKey := strings.TrimPrefix(privateKeyHex, "0x")
		privateKey, err = crypto.HexToECDSA(cleanedKey)
		if err != nil {
			return fmt.Errorf("failed to parse ECDSA private key: %w", err)
		}
	}

	// Get the address from the private key
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Create a new keystore Key structure
	key := &keystore.Key{
		Id:         uuid.New(),
		Address:    address,
		PrivateKey: privateKey,
	}

	// Encrypt the key
	keyjson, err := keystore.EncryptKey(key, password, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return fmt.Errorf("failed to encrypt ECDSA key: %w", err)
	}

	// Write the encrypted key to file
	if err := os.WriteFile(path, keyjson, 0600); err != nil {
		return fmt.Errorf("failed to write keystore file: %w", err)
	}

	// Validate by trying to decrypt
	decryptedKey, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		return fmt.Errorf("failed to validate keystore (decrypt failed): %w", err)
	}

	// Verify the address matches
	if decryptedKey.Address != address {
		return errors.New("keystore validation failed: address mismatch")
	}

	log.Info("✅ ECDSA keystore created successfully",
		zap.String("path", path),
		zap.String("address", address.Hex()))

	return nil
}

// promptForPasswordWithConfirmation prompts the user for a password twice and ensures they match
func promptForPasswordWithConfirmation(prompt string) (string, error) {
	// Check if stdin is a terminal
	if !term.IsTerminal(syscall.Stdin) {
		// Non-terminal mode (e.g., piped input)
		scanner := bufio.NewScanner(os.Stdin)

		fmt.Print(prompt)
		if !scanner.Scan() {
			return "", fmt.Errorf("failed to read password")
		}
		password1 := scanner.Text()

		fmt.Print("Confirm password: ")
		if !scanner.Scan() {
			return "", fmt.Errorf("failed to read password confirmation")
		}
		password2 := scanner.Text()

		if password1 != password2 {
			return "", fmt.Errorf("passwords do not match")
		}

		if password1 == "" {
			return "", fmt.Errorf("password cannot be empty")
		}

		return password1, nil
	}

	// Terminal mode - use secure password reading
	fmt.Print(prompt)
	password1, err := term.ReadPassword(syscall.Stdin)
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	fmt.Print("Confirm password: ")
	password2, err := term.ReadPassword(syscall.Stdin)
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("failed to read password confirmation: %w", err)
	}

	if string(password1) != string(password2) {
		return "", fmt.Errorf("passwords do not match")
	}

	if len(password1) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}

	return string(password1), nil
}
