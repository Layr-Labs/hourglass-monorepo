package signer

import (
	"fmt"
	"os"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type Web3SignerInfo struct {
	Name    string   `json:"name" yaml:"name"`
	Context string   `json:"context" yaml:"context"`
	Path    string   `json:"path" yaml:"path"`
	Files   []string `json:"files" yaml:"files"`
}

func listCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List web3 signer configurations",
		Flags: addContextFlag([]cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "List web3 signer configurations from all contexts",
			},
		}),
		Action: func(c *cli.Context) error {
			log := logger.FromContext(c.Context)
			listAll := c.Bool("all")

			var contexts []string
			if listAll {
				// List all contexts
				cfg, err := config.LoadConfig()
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				for _, ctx := range cfg.Contexts {
					contexts = append(contexts, ctx.Name)
				}
			} else {
				// Just the current/specified context
				contexts = []string{getContextName(c)}
			}

			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			var web3signers []Web3SignerInfo

			for _, contextName := range contexts {
				ctx, exists := cfg.Contexts[contextName]
				if !exists {
					continue
				}

				// List registered web3signers
				for _, ws := range ctx.Web3Signers {
					var files []string
					var paths []string

					if ws.ConfigPath != "" {
						files = append(files, "config.yaml")
						paths = append(paths, ws.ConfigPath)
					}
					if ws.CACertPath != "" {
						files = append(files, "ca.crt")
						paths = append(paths, ws.CACertPath)
					}
					if ws.ClientCertPath != "" {
						files = append(files, "client.crt")
						paths = append(paths, ws.ClientCertPath)
					}
					if ws.ClientKeyPath != "" {
						files = append(files, "client.key")
						paths = append(paths, ws.ClientKeyPath)
					}

					// Check if files still exist
					allExist := true
					for _, p := range paths {
						if _, err := os.Stat(p); err != nil {
							log.Warn("Web3 signer file not found",
								zap.String("name", ws.Name),
								zap.String("path", p),
								zap.Error(err))
							allExist = false
						}
					}

					if allExist {
						info := Web3SignerInfo{
							Name:    ws.Name,
							Context: contextName,
							Path:    paths[0], // Use first path for display
							Files:   files,
						}
						web3signers = append(web3signers, info)
					}
				}
			}

			if len(web3signers) == 0 {
				log.Info("No web3 signer configurations found")
				return nil
			}

			// Format output
			outputFormat := c.String("output")
			formatter := output.NewFormatter(outputFormat)

			switch outputFormat {
			case "json":
				return formatter.PrintJSON(web3signers)
			case "yaml":
				return formatter.PrintYAML(web3signers)
			default:
				// Table format
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Name", "Context", "Files"})

				for _, ws := range web3signers {
					row := []string{
						ws.Name,
						ws.Context,
						strings.Join(ws.Files, ", "),
					}
					table.Append(row)
				}

				table.Render()
				return nil
			}
		},
	}
}
