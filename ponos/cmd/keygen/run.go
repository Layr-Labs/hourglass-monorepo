package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bls381"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/keystore"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

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
			password, err = keystore.GenerateRandomPassword(32)
			if err != nil {
				return fmt.Errorf("failed to generate random password: %w", err)
			}
			l.Sugar().Infow("Generated random password", "password", password)
		} else {
			password = Config.Password
		}

		// Determine keystore options
		var keystoreOpts *keystore.Options
		if Config.LightEncryption {
			keystoreOpts = keystore.Light()
			l.Sugar().Warn("Using light encryption - this is less secure but faster")
		} else {
			keystoreOpts = keystore.Default()
		}

		// Curve type determines the file naming
		curveStr := strings.ToLower(Config.CurveType)

		// Save the keys in the appropriate format
		if Config.UseKeystore {
			// Save using Web3 Secret Storage format with curve type information
			keystorePath := filepath.Join(Config.OutputDir, fmt.Sprintf("%s_%s.json", Config.FilePrefix, curveStr))
			if err := keystore.SaveToKeystoreWithCurveType(privateKey, keystorePath, password, curveStr, keystoreOpts); err != nil {
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
		if keyFile == "" {
			return fmt.Errorf("key file path is required")
		}

		l.Sugar().Infow("Reading key file", "file", keyFile)

		// Check if the file might be a keystore
		if strings.HasSuffix(keyFile, ".json") {
			// Try to parse as a keystore (without decrypting)
			storedKeys, err := keystore.LoadKeystoreFile(keyFile)
			if err != nil {
				return fmt.Errorf("failed to parse keystore JSON: %w", err)
			}
			// Get curve type from keystore if available, otherwise use the one from config
			curveType := Config.CurveType
			if storedKeys.CurveType != "" {
				curveType = storedKeys.CurveType
			}

			// This appears to be a valid keystore
			l.Sugar().Infow("Key Information",
				"type", "keystore",
				"curve", curveType,
				"publicKey", storedKeys.PublicKey,
				"uuid", storedKeys.UUID,
				"version", storedKeys.Version,
			)

			keyScheme, err := keystore.GetSigningSchemeForCurveType(storedKeys.CurveType)
			if err != nil {
				return fmt.Errorf("failed to get signing scheme: %w", err)
			}

			privateSigningKey, err := storedKeys.GetPrivateKey(Config.Password, keyScheme)
			if err != nil {
				return fmt.Errorf("failed to get private key: %w", err)
			}

			pubKey := privateSigningKey.Public()

			if hex.EncodeToString(pubKey.Bytes()) == storedKeys.PublicKey {
				l.Sugar().Infow("Public key matches keystore public key")
			} else {
				l.Sugar().Infow("Public key does not match keystore public key")
			}
			return nil
		}

		// If not a keystore (or couldn't parse as one), try as raw key
		keyData, err := os.ReadFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}

		// Determine the curve type from config or from filename
		curveType := Config.CurveType
		if curveType == "" {
			// Try to infer from filename
			fileName := filepath.Base(keyFile)
			if strings.Contains(fileName, "bls381") {
				curveType = "bls381"
			} else if strings.Contains(fileName, "bn254") {
				curveType = "bn254"
			} else {
				return fmt.Errorf("curve type not provided and could not be inferred from filename")
			}
		}

		var scheme signing.SigningScheme
		switch strings.ToLower(curveType) {
		case "bls381":
			scheme = bls381.NewScheme()
		case "bn254":
			scheme = bn254.NewScheme()
		default:
			return fmt.Errorf("unsupported curve type: %s", curveType)
		}

		// Try to load as private key
		privateKey, err := scheme.NewPrivateKeyFromBytes(keyData)
		if err != nil {
			publicKey := privateKey.Public()
			l.Sugar().Infow("Key Information",
				"type", "private key",
				"curve", curveType,
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
				"curve", curveType,
				"publicKey", hex.EncodeToString(publicKey.Bytes()),
			)
			return nil
		}

		return fmt.Errorf("could not parse key as either private or public key or keystore for curve %s", curveType)
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

		// First, try to get the keystore from the file
		var scheme signing.SigningScheme
		keystoreData, err := keystore.LoadKeystoreFile(keyFile)
		if err == nil {
			// If the keystore has a curveType field, use it
			if keystoreData.CurveType != "" {
				scheme, err = keystore.GetSigningSchemeForCurveType(keystoreData.CurveType)
				if err != nil {
					// If we can't get a scheme from the stored curve type, fall back to config
					l.Sugar().Warnw("Failed to get signing scheme from stored curve type, using config value",
						"storedCurveType", keystoreData.CurveType,
						"error", err)
				}
			}
		}

		// If we couldn't determine the scheme from the keystore, use the config value
		if scheme == nil {
			switch strings.ToLower(Config.CurveType) {
			case "bls381":
				scheme = bls381.NewScheme()
			case "bn254":
				scheme = bn254.NewScheme()
			default:
				return fmt.Errorf("unsupported curve type: %s and keystore does not contain curve type information", Config.CurveType)
			}
		}

		// Test the keystore
		if err := keystore.TestKeystore(keyFile, password, scheme); err != nil {
			return fmt.Errorf("keystore test failed: %w", err)
		}

		// Get the curve type string for display
		curveTypeStr := "unknown"
		if keystoreData != nil && keystoreData.CurveType != "" {
			curveTypeStr = keystoreData.CurveType
		} else if Config.CurveType != "" {
			curveTypeStr = Config.CurveType
		}

		l.Sugar().Infow("Keystore test successful",
			"curve", curveTypeStr,
			"file", keyFile,
		)

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
