package signer

import (
	"fmt"
	"os"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
	"github.com/urfave/cli/v2"
)

// operatorCommand returns the operator signer configuration command
func operatorCommand() *cli.Command {
	return &cli.Command{
		Name:  "operator",
		Usage: "Configure operator signing keys (non-interactive)",
		Subcommands: []*cli.Command{
			operatorPrivateKeyCommand(),
			operatorKeystoreCommand(),
		},
	}
}

// operatorPrivateKeyCommand configures operator with private key from env var
func operatorPrivateKeyCommand() *cli.Command {
	return &cli.Command{
		Name:  "privatekey",
		Usage: "Configure operator with private key from OPERATOR_PRIVATE_KEY env var",
		Action: func(c *cli.Context) error {
			// Check that OPERATOR_PRIVATE_KEY env var exists
			privateKey := os.Getenv("OPERATOR_PRIVATE_KEY")
			if privateKey == "" {
				return fmt.Errorf("OPERATOR_PRIVATE_KEY environment variable not set")
			}

			// Get context name
			contextName := getContextName()

			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			ctx, ok := cfg.Contexts[contextName]
			if !ok {
				return fmt.Errorf("context '%s' not found", contextName)
			}

			// Configure operator keys with private key
			ctx.OperatorKeys = &signer.ECDSAKeyConfig{
				PrivateKey: true,
			}

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✅ Operator configured with private key for context '%s'\n", contextName)
			return nil
		},
	}
}

// operatorKeystoreCommand configures operator with existing keystore
func operatorKeystoreCommand() *cli.Command {
	return &cli.Command{
		Name:  "keystore",
		Usage: "Configure operator with existing keystore",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Required: true,
				Usage:    "Name of the keystore to use",
			},
		},
		Action: func(c *cli.Context) error {
			keystoreName := c.String("name")

			// Get context name
			contextName := getContextName()

			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			ctx, ok := cfg.Contexts[contextName]
			if !ok {
				return fmt.Errorf("context '%s' not found", contextName)
			}

			// Find keystore in context
			var foundKeystore *signer.KeystoreReference
			for _, ks := range ctx.Keystores {
				if ks.Name == keystoreName {
					foundKeystore = &ks
					break
				}
			}

			if foundKeystore == nil {
				// List available keystores for helpful error message
				var available []string
				for _, ks := range ctx.Keystores {
					if ks.Type == "ecdsa" {
						available = append(available, ks.Name)
					}
				}
				if len(available) > 0 {
					return fmt.Errorf("keystore '%s' not found. Available ECDSA keystores: %v", keystoreName, available)
				}
				return fmt.Errorf("keystore '%s' not found. No ECDSA keystores available in context", keystoreName)
			}

			// Validate keystore type (operator must be ECDSA)
			if foundKeystore.Type != "ecdsa" {
				return fmt.Errorf("operator keys must be ECDSA type, but keystore '%s' is type '%s'", 
					keystoreName, foundKeystore.Type)
			}

			// Configure operator keys with keystore
			ctx.OperatorKeys = &signer.ECDSAKeyConfig{
				Keystore: &signer.KeystoreReference{
					Name: foundKeystore.Name,
					Type: foundKeystore.Type,
					Path: foundKeystore.Path,
				},
			}

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✅ Operator configured with keystore '%s' for context '%s'\n", 
				keystoreName, contextName)
			return nil
		},
	}
}