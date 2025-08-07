package web3signer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func importCommand() *cli.Command {
	return &cli.Command{
		Name:  "import",
		Usage: "Import web3 signer configuration file paths",
		Flags: addContextFlag([]cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Name for the web3 signer configuration",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "config",
				Usage: "Path to the web3 signer config file",
			},
			&cli.StringFlag{
				Name:  "ca-cert",
				Usage: "Path to the CA certificate file",
			},
			&cli.StringFlag{
				Name:  "client-cert",
				Usage: "Path to the client certificate file",
			},
			&cli.StringFlag{
				Name:  "client-key",
				Usage: "Path to the client key file",
			},
		}),
		Action: func(c *cli.Context) error {
			log := logger.FromContext(c.Context)

			contextName := getContextName(c)
			name := c.String("name")

			// At least one file must be provided
			configPath := c.String("config")
			caCertPath := c.String("ca-cert")
			clientCertPath := c.String("client-cert")
			clientKeyPath := c.String("client-key")

			if configPath == "" && caCertPath == "" && clientCertPath == "" && clientKeyPath == "" {
				return fmt.Errorf("at least one file must be provided")
			}

			// Create web3signer reference
			ref := config.Web3SignerReference{
				Name: name,
			}

			// Validate and resolve paths
			if configPath != "" {
				absPath, err := filepath.Abs(configPath)
				if err != nil {
					return fmt.Errorf("failed to resolve config path: %w", err)
				}
				if _, err := os.Stat(absPath); err != nil {
					return fmt.Errorf("config file not found: %w", err)
				}
				ref.ConfigPath = absPath
			}

			if caCertPath != "" {
				absPath, err := filepath.Abs(caCertPath)
				if err != nil {
					return fmt.Errorf("failed to resolve CA cert path: %w", err)
				}
				if _, err := os.Stat(absPath); err != nil {
					return fmt.Errorf("CA cert file not found: %w", err)
				}
				ref.CACertPath = absPath
			}

			if clientCertPath != "" {
				absPath, err := filepath.Abs(clientCertPath)
				if err != nil {
					return fmt.Errorf("failed to resolve client cert path: %w", err)
				}
				if _, err := os.Stat(absPath); err != nil {
					return fmt.Errorf("client cert file not found: %w", err)
				}
				ref.ClientCertPath = absPath
			}

			if clientKeyPath != "" {
				absPath, err := filepath.Abs(clientKeyPath)
				if err != nil {
					return fmt.Errorf("failed to resolve client key path: %w", err)
				}
				if _, err := os.Stat(absPath); err != nil {
					return fmt.Errorf("client key file not found: %w", err)
				}
				ref.ClientKeyPath = absPath
			}

			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get context
			ctx, exists := cfg.Contexts[contextName]
			if !exists {
				return fmt.Errorf("context %s not found", contextName)
			}

			// Check if web3signer with same name already exists
			for _, ws := range ctx.Web3Signers {
				if ws.Name == name {
					return fmt.Errorf("web3signer with name %s already exists in context %s", name, contextName)
				}
			}

			// Add web3signer reference
			ctx.Web3Signers = append(ctx.Web3Signers, ref)

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			registeredFiles := []string{}
			if ref.ConfigPath != "" {
				registeredFiles = append(registeredFiles, "config")
			}
			if ref.CACertPath != "" {
				registeredFiles = append(registeredFiles, "ca-cert")
			}
			if ref.ClientCertPath != "" {
				registeredFiles = append(registeredFiles, "client-cert")
			}
			if ref.ClientKeyPath != "" {
				registeredFiles = append(registeredFiles, "client-key")
			}

			log.Info("✅ Web3 signer configuration registered successfully",
				zap.String("context", contextName),
				zap.String("name", name),
				zap.Strings("files", registeredFiles))

			return nil
		},
	}
}
