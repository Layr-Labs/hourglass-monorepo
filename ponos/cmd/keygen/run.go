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

		// Save the keys in the appropriate format
		switch strings.ToLower(Config.CurveType) {
		case "bls381":
			// TODO: Implement EIP-2335 format for BLS12-381
			privFilePath := filepath.Join(Config.OutputDir, fmt.Sprintf("%s_bls381.pri", Config.FilePrefix))
			pubFilePath := filepath.Join(Config.OutputDir, fmt.Sprintf("%s_bls381.pub", Config.FilePrefix))

			if err := os.WriteFile(privFilePath, privateKey.Bytes(), 0600); err != nil {
				return fmt.Errorf("failed to write private key: %w", err)
			}

			if err := os.WriteFile(pubFilePath, publicKey.Bytes(), 0644); err != nil {
				return fmt.Errorf("failed to write public key: %w", err)
			}

			l.Sugar().Infow("Generated BLS12-381 keys",
				"privateKeyFile", privFilePath,
				"publicKeyFile", pubFilePath)

		case "bn254":
			// TODO: Implement Web3 Secret Storage format for BN254
			privFilePath := filepath.Join(Config.OutputDir, fmt.Sprintf("%s_bn254.pri", Config.FilePrefix))
			pubFilePath := filepath.Join(Config.OutputDir, fmt.Sprintf("%s_bn254.pub", Config.FilePrefix))

			if err := os.WriteFile(privFilePath, privateKey.Bytes(), 0600); err != nil {
				return fmt.Errorf("failed to write private key: %w", err)
			}

			if err := os.WriteFile(pubFilePath, publicKey.Bytes(), 0644); err != nil {
				return fmt.Errorf("failed to write public key: %w", err)
			}

			l.Sugar().Infow("Generated BN254 keys",
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

		return fmt.Errorf("could not parse key as either private or public key for curve %s", Config.CurveType)
	},
}

func init() {
	// Generate command flags
	generateCmd.Flags().String("seed", "", "Hex-encoded seed for deterministic key generation")
	generateCmd.Flags().String("path", "", "Derivation path for EIP-2333 (BLS12-381 only), e.g., m/12381/3600/0/0")
	generateCmd.Flags().String("password", "", "Password for encrypting the private key (not used yet)")

	// Info command flags
	infoCmd.Flags().String("key-file", "", "Path to the key file to display information about")

	// Bind the flags to viper
	for _, cmd := range []*cobra.Command{generateCmd, infoCmd} {
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if err := viper.BindPFlag(config.KebabToSnakeCase(f.Name), f); err != nil {
				fmt.Printf("Failed to bind flag '%s' - %+v\n", f.Name, err)
			}
		})
	}
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
