package keystore

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type KeystoreInfo struct {
	Name    string `json:"name" yaml:"name"`
	Type    string `json:"type" yaml:"type"`
	Context string `json:"context" yaml:"context"`
	Path    string `json:"path" yaml:"path"`
	Address string `json:"address,omitempty" yaml:"address,omitempty"`
	Pubkey  string `json:"pubkey,omitempty" yaml:"pubkey,omitempty"`
}

func listCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List keystores",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "List keystores from all contexts",
			},
			&cli.BoolFlag{
				Name:    "full",
				Aliases: []string{"f"},
				Usage:   "Show full public keys without truncation",
			},
		},
		Action: func(c *cli.Context) error {
			log := config.LoggerFromContext(c.Context)
			listAll := c.Bool("all")
			showFull := c.Bool("full")
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			var contexts []string
			if listAll {
				// List all contexts
				for _, ctx := range cfg.Contexts {
					contexts = append(contexts, ctx.Name)
				}
			} else {
				// Just the current/specified context
				contexts = []string{cfg.CurrentContext}
			}

			return listKeystores(c, log, contexts, showFull)
		},
	}
}

func listKeystores(c *cli.Context, log logger.Logger, contexts []string, showFull bool) error {
	var keystores []KeystoreInfo

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	for _, contextName := range contexts {
		ctx, exists := cfg.Contexts[contextName]
		if !exists {
			continue
		}

		// List registered keystores
		for _, ks := range ctx.Keystores {
			// Check if file still exists
			if _, err := os.Stat(ks.Path); err != nil {
				log.Warn("Keystore file not found",
					zap.String("name", ks.Name),
					zap.String("path", ks.Path),
					zap.Error(err))
				continue
			}

			info := KeystoreInfo{
				Name:    ks.Name,
				Type:    ks.Type,
				Context: contextName,
				Path:    ks.Path,
			}

			// Try to read additional info from the keystore file
			if fileContent, err := os.ReadFile(ks.Path); err == nil {
				var jsonData map[string]interface{}
				if err := json.Unmarshal(fileContent, &jsonData); err == nil {
					// Extract address or pubkey
					if ks.Type == "ecdsa" {
						if address, ok := jsonData["address"].(string); ok {
							info.Address = "0x" + address
						}
					} else if ks.Type == "bn254" {
						if pubkey, ok := jsonData["pubkey"].(string); ok {
							info.Pubkey = pubkey
						}
					}
				}
			}

			keystores = append(keystores, info)
		}
	}

	if len(keystores) == 0 {
		log.Info("No keystores found")
		return nil
	}

	// Format output
	outputFormat := c.String("output")
	formatter := output.NewFormatter(outputFormat)

	switch outputFormat {
	case "json":
		return formatter.PrintJSON(keystores)
	case "yaml":
		return formatter.PrintYAML(keystores)
	default:
		// Table format
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Type", "Context", "Key Info"})
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)

		for _, ks := range keystores {
			var keyInfo string
			if ks.Type == "ecdsa" {
				keyInfo = ks.Address
			} else if ks.Type == "bn254" {
				if showFull || len(ks.Pubkey) <= 60 {
					keyInfo = ks.Pubkey
				} else {
					// Truncate long pubkeys: show first 20 + ... + last 20 chars
					keyInfo = ks.Pubkey[:20] + "..." + ks.Pubkey[len(ks.Pubkey)-20:]
				}
			}

			row := []string{
				ks.Name,
				ks.Type,
				ks.Context,
				keyInfo,
			}
			table.Append(row)
		}

		table.Render()
		return nil
	}
}
