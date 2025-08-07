package keystore

import (
	"encoding/json"
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
		Usage: "Import an existing keystore file reference",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Name for the keystore reference",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "path",
				Usage:    "Path to the keystore file",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			log := logger.FromContext(c.Context)

			name := c.String("name")
			path := c.String("path")

			// Resolve to absolute path
			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("failed to resolve absolute path: %w", err)
			}

			// Validate file exists
			if _, err := os.Stat(absPath); err != nil {
				return fmt.Errorf("keystore file not found: %w", err)
			}

			// Validate it's a valid keystore file
			fileContent, err := os.ReadFile(absPath)
			if err != nil {
				return fmt.Errorf("failed to read keystore file: %w", err)
			}

			var jsonData map[string]interface{}
			if err := json.Unmarshal(fileContent, &jsonData); err != nil {
				return fmt.Errorf("invalid keystore file format: %w", err)
			}

			// Determine keystore type
			keystoreType := "unknown"
			if _, hasAddress := jsonData["address"]; hasAddress {
				keystoreType = "ecdsa"
			} else if _, hasPubkey := jsonData["pubkey"]; hasPubkey {
				keystoreType = "bn254"
			}

			if keystoreType == "unknown" {
				return fmt.Errorf("unable to determine keystore type")
			}

			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get context
			ctx, exists := cfg.Contexts[cfg.CurrentContext]
			if !exists {
				return fmt.Errorf("context %s not found", cfg.CurrentContext)
			}

			// Check if keystore with same name already exists
			for _, ks := range ctx.Keystores {
				if ks.Name == name {
					return fmt.Errorf("keystore with name %s already exists in context %s", name, cfg.CurrentContext)
				}
			}

			// Add keystore reference
			ctx.Keystores = append(ctx.Keystores, config.KeystoreReference{
				Name: name,
				Path: absPath,
				Type: keystoreType,
			})

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			log.Info("✅ Keystore registered successfully",
				zap.String("context", cfg.CurrentContext),
				zap.String("name", name),
				zap.String("type", keystoreType),
				zap.String("path", absPath))

			return nil
		},
	}
}
