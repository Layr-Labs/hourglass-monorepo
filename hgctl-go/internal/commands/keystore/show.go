package keystore

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"os"
	"path/filepath"
	"strings"

	blskeystore "github.com/Layr-Labs/crypto-libs/pkg/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
	"github.com/urfave/cli/v2"
)

// Define constants
const (
	KeyTypeECDSA = "ecdsa"
	KeyTypeBN254 = "bn254"
)

var (
	ErrInvalidKeyType = errors.New("invalid key type")
)

func showCommand() *cli.Command {
	return &cli.Command{
		Name:      "show",
		Usage:     "Show the private key of a keystore from your context",
		ArgsUsage: "[key-name]",
		Description: `Shows the private key of a keystore that has been added to your context.
		
You can specify the key name either as a positional argument or using the --name flag.
The keystore must have been previously added to your context using 'keystore add'.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Required: false,
				Usage:    "Name of the key you want to show.",
			},
		},
		Action: func(c *cli.Context) error {
			// Get key name from flag or positional argument
			keyName := c.String("name")
			if keyName == "" && c.NArg() > 0 {
				keyName = c.Args().Get(0)
			}

			if keyName == "" {
				return errors.New("key name is required (use --name flag or provide as argument)")
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			currentCtx, exists := cfg.Contexts[cfg.CurrentContext]
			if !exists {
				return fmt.Errorf("current context '%s' not found", cfg.CurrentContext)
			}

			var keystoreRef *signer.KeystoreReference
			for _, ks := range currentCtx.Keystores {
				if ks.Name == keyName {
					keystoreRef = &ks
					break
				}
			}

			if keystoreRef == nil {
				return fmt.Errorf("keystore '%s' not found in current context", keyName)
			}

			if _, err := os.Stat(keystoreRef.Path); os.IsNotExist(err) {
				return fmt.Errorf("keystore file does not exist at path: %s", keystoreRef.Path)
			}

			confirm, err := output.Confirm("This will show your private key. Are you sure you want to export?")
			if err != nil {
				return err
			}
			if !confirm {
				return nil
			}

			password, err := output.InputHiddenString("Enter password to decrypt the key", "", func(s string) error {
				return nil
			})
			if err != nil {
				return err
			}
			fmt.Printf("Showing the key '%s' from: %s\n", keyName, keystoreRef.Path)

			privateKey, err := getPrivateKey(strings.ToLower(keystoreRef.Type), keystoreRef.Path, password)
			if err != nil {
				return err
			}
			fmt.Println("Private key:", privateKey)
			return nil
		},
	}
}

func getPrivateKey(keyType string, filePath string, password string) (string, error) {
	switch keyType {
	case KeyTypeECDSA:
		keyStoreContents, err := os.ReadFile(filepath.Clean(filePath))
		if err != nil {
			return "", err
		}

		key, err := keystore.DecryptKey(keyStoreContents, password)
		if err != nil {
			return "", err
		}
		fmt.Println("Address:", key.Address.Hex())
		return hex.EncodeToString(key.PrivateKey.D.Bytes()), nil

	case KeyTypeBN254:
		// Read the keystore file
		keystoreData, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read keystore file: %w", err)
		}

		// Check if it's a BN254 keystore
		var testKeystore map[string]interface{}
		if err := json.Unmarshal(keystoreData, &testKeystore); err == nil {
			var keystoreFile blskeystore.EIP2335Keystore
			if err := json.Unmarshal(keystoreData, &keystoreFile); err != nil {
				return "", fmt.Errorf("failed to unmarshal BN254 keystore: %w", err)
			}

			scheme, err := blskeystore.GetSigningSchemeForCurveType("bn254")
			if err != nil {
				return "", fmt.Errorf("failed to get bn254 scheme: %w", err)
			}

			privateKey, err := keystoreFile.GetPrivateKey(password, scheme)
			if err != nil {
				return "", fmt.Errorf("failed to decrypt BN254 private key: %w", err)
			}

			return hex.EncodeToString(privateKey.Bytes()), nil
		}

		// Try as a bn254 private key directly
		privateKeyBytes, err := decryptBn254Keystore(keystoreData, password)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt BN254 keystore: %w", err)
		}

		return hex.EncodeToString(privateKeyBytes), nil

	default:
		return "", ErrInvalidKeyType
	}
}

// decryptBn254Keystore decrypts a keystore file to get the private key
func decryptBn254Keystore(keystoreData []byte, password string) ([]byte, error) {
	// Check if it's a BN254 keystore
	var testKeystore map[string]interface{}
	if err := json.Unmarshal(keystoreData, &testKeystore); err == nil {
		var keystoreFile blskeystore.EIP2335Keystore
		if err := json.Unmarshal(keystoreData, &keystoreFile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal BN254 keystore: %w", err)
		}
		scheme, err := blskeystore.GetSigningSchemeForCurveType("bn254")
		if err != nil {
			return nil, fmt.Errorf("failed to get bn254 scheme: %w", err)
		}
		privateKey, err := keystoreFile.GetPrivateKey(password, scheme)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt BN254 private key: %w", err)
		}

		return privateKey.Bytes(), nil
	}

	// Try standard Ethereum keystore
	key, err := keystore.DecryptKey(keystoreData, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore: %w", err)
	}

	return key.PrivateKey.D.Bytes(), nil
}
