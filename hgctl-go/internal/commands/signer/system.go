package signer

import (
	"fmt"
	"os"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
	"github.com/urfave/cli/v2"
)

// systemCommand returns the system signer configuration command
func systemCommand() *cli.Command {
	return &cli.Command{
		Name:  "system",
		Usage: "Configure system signing keys",
		Subcommands: []*cli.Command{
			systemPrivateKeyCommand(),
			systemKeystoreCommand(),
			systemRemoveCommand(),
		},
	}
}

// systemPrivateKeyCommand configures system ECDSA with private key from env var
func systemPrivateKeyCommand() *cli.Command {
	return &cli.Command{
		Name:  "privatekey",
		Usage: "Configure system ECDSA with private key from SYSTEM_PRIVATE_KEY env var",
		Action: func(c *cli.Context) error {
			// Check that SYSTEM_PRIVATE_KEY env var exists
			privateKey := os.Getenv("SYSTEM_PRIVATE_KEY")
			if privateKey == "" {
				return fmt.Errorf("SYSTEM_PRIVATE_KEY environment variable not set")
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

			// Check that operator keys are configured first
			if ctx.OperatorKeys == nil {
				return fmt.Errorf("operator keys must be configured before system keys. Run 'hgctl signer operator' first")
			}

			// Initialize SystemSignerKeys if nil
			if ctx.SystemSignerKeys == nil {
				ctx.SystemSignerKeys = &signer.SigningKeys{}
			}

			// Configure system ECDSA with private key
			ctx.SystemSignerKeys.ECDSA = &signer.ECDSAKeyConfig{
				PrivateKey: true,
			}

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✅ System ECDSA configured with private key for context '%s'\n", contextName)
			return nil
		},
	}
}

// systemKeystoreCommand configures system signer with keystore
func systemKeystoreCommand() *cli.Command {
	return &cli.Command{
		Name:  "keystore",
		Usage: "Configure system signer with keystore",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Required: true,
				Usage:    "Name of the keystore to use",
			},
			&cli.StringFlag{
				Name:     "type",
				Required: true,
				Usage:    "Key type (only 'bn254' is supported)",
			},
		},
		Action: func(c *cli.Context) error {
			keystoreName := c.String("name")
			keyType := strings.ToLower(c.String("type"))

			// Validate key type
			if keyType != "bn254" && keyType != "ecdsa" {
				return fmt.Errorf("invalid key type: %s (must be 'ecdsa' or 'bn254')", keyType)
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

			// Check that operator keys are configured first
			if ctx.OperatorKeys == nil {
				return fmt.Errorf("operator keys must be configured before system keys. Run 'hgctl signer operator' first")
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
					if ks.Type == keyType {
						available = append(available, fmt.Sprintf("%s (%s)", ks.Name, ks.Type))
					}
				}
				if len(available) > 0 {
					return fmt.Errorf("keystore '%s' not found. Available %s keystores: %v",
						keystoreName, strings.ToUpper(keyType), available)
				}
				return fmt.Errorf("keystore '%s' not found. No %s keystores available in context",
					keystoreName, strings.ToUpper(keyType))
			}

			// Validate keystore type matches requested type
			if foundKeystore.Type != keyType {
				return fmt.Errorf("keystore '%s' is type '%s', but requested type '%s'",
					keystoreName, foundKeystore.Type, keyType)
			}

			// Initialize SystemSignerKeys if nil
			if ctx.SystemSignerKeys == nil {
				ctx.SystemSignerKeys = &signer.SigningKeys{}
			}

			// Configure system keystore based on type
			keystoreRef := &signer.KeystoreReference{
				Name: foundKeystore.Name,
				Type: foundKeystore.Type,
				Path: foundKeystore.Path,
			}

			if keyType == "ecdsa" {
				ctx.SystemSignerKeys.ECDSA = &signer.ECDSAKeyConfig{
					Keystore: keystoreRef,
				}
				fmt.Printf("✅ System ECDSA configured with keystore '%s' for context '%s'\n",
					keystoreName, contextName)
			} else {
				ctx.SystemSignerKeys.BN254 = keystoreRef
				fmt.Printf("✅ System BN254 configured with keystore '%s' for context '%s'\n",
					keystoreName, contextName)
			}

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			return nil
		},
	}
}

// systemRemoveCommand removes system signer configuration from context
func systemRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Remove system signer configuration from current context",
		Action: func(c *cli.Context) error {
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

			// Check if system signer is configured
			if ctx.SystemSignerKeys == nil {
				fmt.Println("No system signer configuration found for this context")
				return nil
			}

			// Remove system signer configuration
			ctx.SystemSignerKeys = nil

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✅ System signer configuration removed from context '%s'\n", contextName)
			return nil
		},
	}
}
