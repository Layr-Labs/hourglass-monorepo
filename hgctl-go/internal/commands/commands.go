package commands

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/allocate"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/delegate"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/deposit"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/operatorset"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/version"
	"github.com/urfave/cli/v2"

	contextcmd "github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/context"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/deploy"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/describe"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/get"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/remove"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/translate"
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
			GetCommand(),
			DescribeCommand(),
			DeployCommand(),
			TranslateCommand(),
			RemoveCommand(),
			ContextCommand(),
			KeystoreCommand(),
			Web3SignerCommand(),
			// Operator management commands
			RegisterCommand(),
			RegisterAVSCommand(),
			RegisterKeyCommand(),
			DepositCommand(),
			DelegateCommand(),
			AllocateCommand(),
			SetAllocationDelayCommand(),
			operatorset.Command(),
		},
	}
}

func isConfigCommand(c *cli.Context) bool {
	if c.NArg() == 0 {
		return false
	}
	cmd := c.Args().Get(0)
	return cmd == "context" || cmd == "completion"
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

// TranslateCommand returns the translate command
func TranslateCommand() *cli.Command {
	return translate.Command()
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
