package commands

import (
	"context"
	"fmt"
	"os"

	contextcmd "github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/context"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/deploy"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/describe"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/get"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/remove"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/signer"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/telemetry"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/version"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func Hgctl() *cli.App {
	return &cli.App{
		Name:    "hgctl",
		Usage:   "CLI for managing Hourglass AVS deployments",
		Version: version.GetFullVersion(),
		Description: `hgctl is a command-line interface for managing Hourglass AVS deployments.
It provides commands to interact with the Executor service, manage contexts,
and deploy AVS artifacts including EigenRuntime specifications.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable verbose logging",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output format (table|json|yaml)",
				Value:   "table",
			},
		},
		Before: middleware.ChainBeforeFuncs(
			// First load config and context
			func(c *cli.Context) error {
				// Initialize logger early for debugging
				verbose := c.Bool("verbose")
				logger.InitGlobalLoggerWithWriter(verbose, c.App.Writer)
				l := logger.GetLogger()

				// Check if user is requesting help
				for _, arg := range os.Args {
					if arg == "--help" || arg == "-h" || arg == "help" {
						// Set empty context for help display
						c.Context = context.WithValue(c.Context, config.ConfigKey, &config.Config{})
						c.Context = context.WithValue(c.Context, config.ContextKey, &config.Context{})
						return nil
					}
				}

				// Load config
				cfg, err := config.LoadConfig()
				if err != nil {
					l.Error("Failed to load config", zap.Error(err))
					// Create empty config if it doesn't exist
					cfg = &config.Config{
						Contexts: make(map[string]*config.Context),
					}
				}

				// Get current context
				currentCtx, err := config.GetCurrentContext()
				if err != nil {
					// Allow config commands to run without a context
					if isConfigCommand(c) {
						currentCtx = &config.Context{}
					} else {
						// Log helpful error message
						l.Error("No context configured. Please create a context.")
						fmt.Fprintf(os.Stderr, "\nError: No context configured\n\n")
						fmt.Fprintf(os.Stderr, "To create a context, run:\n")
						fmt.Fprintf(os.Stderr, "  hgctl context create --name default --use\n\n")
						fmt.Fprintf(os.Stderr, "To list available contexts:\n")
						fmt.Fprintf(os.Stderr, "  hgctl context list\n\n")
						return fmt.Errorf("no context configured")
					}
				}

				// Store in context
				c.Context = context.WithValue(c.Context, config.ConfigKey, cfg)
				c.Context = context.WithValue(c.Context, config.ContextKey, currentCtx)

				return nil
			},
			middleware.MiddlewareBeforeFunc,
			middleware.InitTelemetry, // Initialize telemetry client
		),
		After: func(c *cli.Context) error {
			// Cleanup contract client
			if err := middleware.CleanupContractClient(c); err != nil {
				return err
			}
			// Close telemetry client
			telemetry.Close()
			return nil
		},
		Commands: []*cli.Command{
			GetCommand(),
			DescribeCommand(),
			DeployCommand(),
			RemoveCommand(),
			ContextCommand(),
			KeystoreCommand(),
			SignerCommand(),
			EigenLayerCommand(),
		},
		ExitErrHandler: middleware.ExitErrHandler,
	}
}

func isConfigCommand(c *cli.Context) bool {
	if c.NArg() == 0 {
		return false
	}
	cmd := c.Args().Get(0)
	return cmd == "context"
}

// wrapCommandWithTelemetry recursively wraps command actions with telemetry
func wrapCommandWithTelemetry(cmd *cli.Command) {
	if cmd == nil {
		return
	}

	// Wrap the action if it exists
	if cmd.Action != nil {
		cmd.Action = middleware.WithTelemetry(cmd.Action)
	}

	// Recursively wrap subcommands
	for _, subCmd := range cmd.Subcommands {
		wrapCommandWithTelemetry(subCmd)
	}
}

// GetCommand returns the get command
func GetCommand() *cli.Command {
	cmd := get.Command()
	cmd.Before = middleware.RequireContext
	wrapCommandWithTelemetry(cmd)
	return cmd
}

// DescribeCommand returns the describe command
func DescribeCommand() *cli.Command {
	cmd := describe.Command()
	cmd.Before = middleware.RequireContext
	wrapCommandWithTelemetry(cmd)
	return cmd
}

// DeployCommand returns the deploy command
func DeployCommand() *cli.Command {
	cmd := deploy.Command()
	cmd.Before = middleware.RequireContext
	wrapCommandWithTelemetry(cmd)
	return cmd
}

// RemoveCommand returns the remove command
func RemoveCommand() *cli.Command {
	cmd := remove.Command()
	cmd.Before = middleware.RequireContext
	wrapCommandWithTelemetry(cmd)
	return cmd
}

// ContextCommand returns the context command
func ContextCommand() *cli.Command {
	cmd := contextcmd.Command()
	wrapCommandWithTelemetry(cmd)
	return cmd
}

// KeystoreCommand returns the keystore command
func KeystoreCommand() *cli.Command {
	cmd := keystore.Command()
	cmd.Before = middleware.RequireContext
	wrapCommandWithTelemetry(cmd)
	return cmd
}

// SignerCommand returns the web3signer command
func SignerCommand() *cli.Command {
	cmd := signer.Command()
	cmd.Before = middleware.RequireContext
	wrapCommandWithTelemetry(cmd)
	return cmd
}

// EigenLayerCommand returns the eigenlayer command with all subcommands
func EigenLayerCommand() *cli.Command {
	cmd := eigenlayer.Command()
	cmd.Before = middleware.RequireContext
	wrapCommandWithTelemetry(cmd)
	return cmd
}
