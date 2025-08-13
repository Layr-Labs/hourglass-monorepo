package keystore

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"os"
	"path/filepath"
	"strings"

	blskeystore "github.com/Layr-Labs/crypto-libs/pkg/keystore"
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
		Name:  "show",
		Usage: "Used to show existing keys from local keystore",
		Description: `Used to export ecdsa and bls key from local keystore

keyname - This will be the name of the key to be imported. If the path of keys is
different from default path created by "create"/"import" command, then provide the
full path using --key-path flag.

If both keyname is provided and --key-path flag is provided, then keyname will be used. 

use --key-type ecdsa/bn254 to export ecdsa/bn254 key. 
- ecdsa - exported key will be plaintext hex encoded private key
- bn254 (or bls) - exported key will be plaintext bn254 private key

It will prompt for password to encrypt the key.

This command will import keys from $HOME/.eigenlayer/operator_keys/ location

But if you want it to export from a different location, use --key-path flag`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "key-type",
				Required: true,
				Usage:    "Type of key you want to export. Currently supports 'ecdsa' and 'bn254' (or 'bls')",
			},
			&cli.StringFlag{
				Name:     "key-path",
				Required: true,
				Usage:    "Use this flag to specify the path of the key",
			},
		},
		Action: func(c *cli.Context) error {
			keyType := strings.ToLower(c.String("key-type"))

			keyName := c.Args().Get(0)

			keyPath := c.String("key-path")
			if len(keyPath) == 0 && len(keyName) == 0 {
				return errors.New("one of keyname or --key-path is required")
			}

			if len(keyPath) > 0 && len(keyName) > 0 {
				return errors.New("keyname and --key-path both are provided. Please provide only one")
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
			fmt.Println("exporting key from: ", keyPath)

			privateKey, err := getPrivateKey(keyType, keyPath, password)
			if err != nil {
				return err
			}
			fmt.Println("Private key: ", privateKey)
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
		return hex.EncodeToString(key.PrivateKey.D.Bytes()), nil

	case KeyTypeBN254:
		// Read the keystore file
		keystoreData, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read keystore file: %w", err)
		}

		// Check if it's a BLS keystore
		var testKeystore map[string]interface{}
		if err := json.Unmarshal(keystoreData, &testKeystore); err == nil {
			var keystoreFile blskeystore.EIP2335Keystore
			if err := json.Unmarshal(keystoreData, &keystoreFile); err != nil {
				return "", fmt.Errorf("failed to unmarshal BLS keystore: %w", err)
			}

			scheme, err := blskeystore.GetSigningSchemeForCurveType("bn254")
			if err != nil {
				return "", fmt.Errorf("failed to get bn254 scheme: %w", err)
			}

			privateKey, err := keystoreFile.GetPrivateKey(password, scheme)
			if err != nil {
				return "", fmt.Errorf("failed to decrypt BLS private key: %w", err)
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
	// Check if it's a BLS keystore
	var testKeystore map[string]interface{}
	if err := json.Unmarshal(keystoreData, &testKeystore); err == nil {
		var keystoreFile blskeystore.EIP2335Keystore
		if err := json.Unmarshal(keystoreData, &keystoreFile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal BLS keystore: %w", err)
		}
		scheme, err := blskeystore.GetSigningSchemeForCurveType("bn254")
		if err != nil {
			return nil, fmt.Errorf("failed to get bn254 scheme: %w", err)
		}
		privateKey, err := keystoreFile.GetPrivateKey(password, scheme)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt BLS private key: %w", err)
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
