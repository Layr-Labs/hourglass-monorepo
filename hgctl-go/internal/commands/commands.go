package commands

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/allocate"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/delegate"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/deposit"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/version"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"os"

	contextcmd "github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/context"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/deploy"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/describe"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/get"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/register"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/remove"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/web3signer"
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
		),
		After: func(c *cli.Context) error {
			return middleware.CleanupContractClient(c)
		},
		Commands: []*cli.Command{
			GetCommand(),
			DescribeCommand(),
			DeployCommand(),
			RemoveCommand(),
			ContextCommand(),
			KeystoreCommand(),
			Web3SignerCommand(),
			// Operator management commands
			RegisterOperatorCommand(),
			RegisterAVSCommand(),
			RegisterKeyCommand(),
			DepositCommand(),
			DelegateCommand(),
			AllocateCommand(),
			SetAllocationDelayCommand(),
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

// GetCommand returns the get command
func GetCommand() *cli.Command {
	return get.Command()
}

// DescribeCommand returns the describe command
func DescribeCommand() *cli.Command {
	return describe.Command()
}

// DeployCommand returns the deploy command
func DeployCommand() *cli.Command {
	return deploy.Command()
}

// RemoveCommand returns the remove command
func RemoveCommand() *cli.Command {
	return remove.Command()
}

// ContextCommand returns the context command
func ContextCommand() *cli.Command {
	return contextcmd.Command()
}

// KeystoreCommand returns the keystore command
func KeystoreCommand() *cli.Command {
	return keystore.Command()
}

// Web3SignerCommand returns the web3signer command
func Web3SignerCommand() *cli.Command {
	return web3signer.Command()
}

// DelegateCommand returns the delegate command
func DelegateCommand() *cli.Command {
	return delegate.Command()
}

// DepositCommand returns the delegate command
func DepositCommand() *cli.Command {
	return deposit.Command()
}

// AllocateCommand returns the delegate command
func AllocateCommand() *cli.Command {
	return allocate.Command()
}

// RegisterOperatorCommand returns the register-operator command
func RegisterOperatorCommand() *cli.Command {
	return register.RegisterOperatorCommand()
}

// RegisterAVSCommand returns the register-avs command
func RegisterAVSCommand() *cli.Command {
	return register.RegisterAVSCommand()
}

// RegisterKeyCommand returns the register-key command
func RegisterKeyCommand() *cli.Command {
	return register.RegisterKeyCommand()
}

// SetAllocationDelayCommand returns the set-allocation-delay command
func SetAllocationDelayCommand() *cli.Command {
	return register.SetAllocationDelayCommand()
}
