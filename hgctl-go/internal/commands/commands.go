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
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/telemetry"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/hooks"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
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
			func(c *cli.Context) error {
				verbose := c.Bool("verbose")
				logger.InitGlobalLoggerWithWriter(verbose, c.App.Writer)
				l := logger.GetLogger()

				for _, arg := range os.Args {
					if arg == "--help" || arg == "-h" || arg == "help" {
						c.Context = context.WithValue(c.Context, config.ConfigKey, &config.Config{})
						c.Context = context.WithValue(c.Context, config.ContextKey, &config.Context{})
						return nil
					}
				}

				cfg, err := config.LoadConfig()
				if err != nil {
					l.Error("Failed to load config", zap.Error(err))
					cfg = &config.Config{
						Contexts: make(map[string]*config.Context),
					}
				}

				currentCtx, err := config.GetCurrentContext()
				if err != nil {
					if isConfigCommand(c) {
						currentCtx = &config.Context{}
					} else {
						l.Error("No context configured. Please create a context.")
						fmt.Fprintf(os.Stderr, "\nError: No context configured\n\n")
						fmt.Fprintf(os.Stderr, "To create a context, run:\n")
						fmt.Fprintf(os.Stderr, "  hgctl context create --name default --use\n\n")
						fmt.Fprintf(os.Stderr, "To list available contexts:\n")
						fmt.Fprintf(os.Stderr, "  hgctl context list\n\n")
						return fmt.Errorf("no context configured")
					}
				}

				c.Context = context.WithValue(c.Context, config.ConfigKey, cfg)
				c.Context = context.WithValue(c.Context, config.ContextKey, currentCtx)

				return hooks.WithCommandMetricsContext(c)
			},
			middleware.MiddlewareBeforeFunc,
		),
		After: middleware.CleanupContractClient,
		Commands: []*cli.Command{
			GetCommand(),
			DescribeCommand(),
			DeployCommand(),
			RemoveCommand(),
			ContextCommand(),
			KeystoreCommand(),
			SignerCommand(),
			EigenLayerCommand(),
			TelemetryCommand(),
		},
		ExitErrHandler: middleware.ExitErrHandler,
	}
}

func isConfigCommand(c *cli.Context) bool {
	if c.NArg() == 0 {
		return false
	}
	cmd := c.Args().Get(0)
	return cmd == "context" || cmd == "telemetry"
}

func GetCommand() *cli.Command {
	cmd := get.Command()
	cmd.Before = middleware.RequireContext
	return cmd
}

func DescribeCommand() *cli.Command {
	cmd := describe.Command()
	cmd.Before = middleware.RequireContext
	return cmd
}

func DeployCommand() *cli.Command {
	cmd := deploy.Command()
	cmd.Before = middleware.RequireContext
	return cmd
}

func RemoveCommand() *cli.Command {
	cmd := remove.Command()
	cmd.Before = middleware.RequireContext
	return cmd
}

func ContextCommand() *cli.Command {
	return contextcmd.Command()
}

func KeystoreCommand() *cli.Command {
	cmd := keystore.Command()
	cmd.Before = middleware.RequireContext
	return cmd
}

func SignerCommand() *cli.Command {
	cmd := signer.Command()
	cmd.Before = middleware.RequireContext
	return cmd
}

func EigenLayerCommand() *cli.Command {
	cmd := eigenlayer.Command()
	cmd.Before = middleware.RequireContext
	return cmd
}

func TelemetryCommand() *cli.Command {
	return telemetry.Command()
}
