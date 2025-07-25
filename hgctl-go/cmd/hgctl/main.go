package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
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
			&cli.StringFlag{
				Name:  "executor-address",
				Usage: "Override executor address",
			},
			&cli.StringFlag{
				Name:  "rpc-url",
				Usage: "Override RPC URL",
			},
			&cli.StringFlag{
				Name:  "release-manager",
				Usage: "Override release manager address",
			},
			&cli.Uint64Flag{
				Name:  "operator-set-id",
				Usage: "Override operator set ID",
				Value: 0,
			},
		},
		Before: func(c *cli.Context) error {
			// Initialize logger
			verbose := c.Bool("verbose")
			logger.InitGlobalLogger(verbose)

			// Load config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get current context
			currentCtx, err := config.GetCurrentContext()
			if err != nil && !isConfigCommand(c) {
				logger.GetLogger().Warn("No context configured, using defaults")
			}

			// Apply command-line overrides
			if currentCtx != nil {
				if addr := c.String("executor-address"); addr != "" {
					currentCtx.ExecutorAddress = addr
				}
				if url := c.String("rpc-url"); url != "" {
					currentCtx.RPCUrl = url
				}
				if addr := c.String("release-manager"); addr != "" {
					currentCtx.ReleaseManagerAddress = addr
				}
				if id := c.Uint64("operator-set-id"); id > 0 {
					currentCtx.OperatorSetID = uint32(id)
				}
			}

			// Store in context
			c.Context = context.WithValue(c.Context, "config", cfg)
			c.Context = context.WithValue(c.Context, "currentContext", currentCtx)
			c.Context = logger.WithLogger(c.Context, logger.GetLogger())

			return nil
		},
		Commands: []*cli.Command{
			commands.GetCommand(),
			commands.DescribeCommand(),
			commands.DeployCommand(),
			commands.TranslateCommand(),
			commands.RemoveCommand(),
			commands.ContextCommand(),
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		logger.GetLogger().Error("Command failed", zap.Error(err))
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
