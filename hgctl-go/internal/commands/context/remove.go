package context

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/urfave/cli/v2"
)

func removeCommand() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Remove context properties",
		Description: `Remove specific properties from the current context.
This command allows you to unset values for various context properties.

Examples:
  # Remove executor address
  hgctl context remove --executor-address

  # Remove contract overrides
  hgctl context remove --delegation-manager --allocation-manager`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "executor-address",
				Usage: "Remove the executor gRPC address",
			},
			&cli.BoolFlag{
				Name:  "avs-address",
				Usage: "Remove the AVS contract address",
			},
			&cli.BoolFlag{
				Name:  "operator-address",
				Usage: "Remove the operator address",
			},
			&cli.BoolFlag{
				Name:  "operator-set-id",
				Usage: "Remove the operator set ID",
			},
			&cli.BoolFlag{
				Name:  "rpc-url",
				Usage: "Remove the Ethereum RPC URL",
			},
			&cli.StringSliceFlag{
				Name:  "env",
				Usage: "Remove environment variables (specify keys)",
			},
			&cli.BoolFlag{
				Name:  "env-secrets-path",
				Usage: "Remove the path to environment secrets file",
			},
		},
		Action: contextRemoveAction,
	}
}

func contextRemoveAction(c *cli.Context) error {
	log := config.LoggerFromContext(c.Context)

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, exists := cfg.Contexts[cfg.CurrentContext]
	if !exists {
		return fmt.Errorf("current context '%s' not found", cfg.CurrentContext)
	}

	removed := false

	// Remove simple string fields
	if c.Bool("executor-endpoint") {
		if ctx.ExecutorEndpoint != "" {
			ctx.ExecutorEndpoint = ""
			removed = true
			log.Info("Removed executor endpoint")
		}
	}

	if c.Bool("avs-address") {
		if ctx.AVSAddress != "" {
			ctx.AVSAddress = ""
			removed = true
			log.Info("Removed AVS address")
		}
	}

	if c.Bool("operator-address") {
		if ctx.OperatorAddress != "" {
			ctx.OperatorAddress = ""
			removed = true
			log.Info("Removed operator address")
		}
	}

	if c.Bool("operator-set-id") {
		if ctx.OperatorSetID != 0 {
			ctx.OperatorSetID = 0
			removed = true
			log.Info("Removed operator set ID")
		}
	}

	if c.Bool("rpc-url") {
		if ctx.L1RPCUrl != "" {
			ctx.L1RPCUrl = ""
			removed = true
			log.Info("Removed RPC URL")
		}
	}

	if c.Bool("env-secrets-path") {
		if ctx.EnvSecretsPath != "" {
			ctx.EnvSecretsPath = ""
			removed = true
			log.Info("Removed env secrets path")
		}
	}

	if !removed {
		return fmt.Errorf("no values were removed (either not set or not specified)")
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Context '%s' updated\n", cfg.CurrentContext)
	return nil
}
