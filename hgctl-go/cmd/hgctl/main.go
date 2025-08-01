package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/operatorset"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/telemetry"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/version"
)

func main() {
	// Initialize telemetry
	telemetry.Init()
	defer telemetry.Close()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	app := &cli.App{
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
				// Load config
				cfg, err := config.LoadConfig()
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				// Get current context
				currentCtx, err := config.GetCurrentContext()
				if err != nil {
					// Allow config commands to run without a context
					if isConfigCommand(c) {
						currentCtx = &config.Context{}
					} else {
						return fmt.Errorf("no context configured: %w", err)
					}
				}

				// Store in context
				c.Context = context.WithValue(c.Context, "config", cfg)
				c.Context = context.WithValue(c.Context, "currentContext", currentCtx)

				return nil
			},
			// Then run the middleware
			middleware.MiddlewareBeforeFunc,
		),
		After: func(c *cli.Context) error {
			// Cleanup contract client
			return middleware.CleanupContractClient(c)
		},
		Commands: []*cli.Command{
			commands.GetCommand(),
			commands.DescribeCommand(),
			commands.DeployCommand(),
			commands.TranslateCommand(),
			commands.RemoveCommand(),
			commands.ContextCommand(),
			commands.KeystoreCommand(),
			commands.Web3SignerCommand(),
			// Operator management commands
			commands.RegisterCommand(),
			commands.RegisterAVSCommand(),
			commands.RegisterKeyCommand(),
			commands.DepositCommand(),
			commands.DelegateCommand(),
			commands.AllocateCommand(),
			commands.SetAllocationDelayCommand(),
			operatorset.Command(),
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		// Error already logged by middleware
		os.Exit(1)
	}
}

func isConfigCommand(c *cli.Context) bool {
	if c.NArg() == 0 {
		return false
	}
	cmd := c.Args().Get(0)
	return cmd == "context" || cmd == "completion"
}
