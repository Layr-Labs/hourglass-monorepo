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
			&cli.BoolFlag{
				Name:  "release-manager",
				Usage: "Remove the release manager contract address",
			},
			&cli.BoolFlag{
				Name:  "delegation-manager",
				Usage: "Remove the delegation manager contract address",
			},
			&cli.BoolFlag{
				Name:  "allocation-manager",
				Usage: "Remove the allocation manager contract address",
			},
			&cli.BoolFlag{
				Name:  "strategy-manager",
				Usage: "Remove the strategy manager contract address",
			},
			&cli.BoolFlag{
				Name:  "key-registrar",
				Usage: "Remove the key registrar contract address",
			},
			&cli.StringSliceFlag{
				Name:  "env",
				Usage: "Remove environment variables (specify keys)",
			},
			&cli.BoolFlag{
				Name:  "env-secrets-path",
				Usage: "Remove the path to environment secrets file",
			},
			&cli.BoolFlag{
				Name:  "all-contracts",
				Usage: "Remove all contract overrides",
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
	if c.Bool("executor-address") {
		if ctx.ExecutorAddress != "" {
			ctx.ExecutorAddress = ""
			removed = true
			log.Info("Removed executor address")
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

	// Handle contract overrides removal
	if c.Bool("all-contracts") {
		if ctx.ContractOverrides != nil {
			ctx.ContractOverrides = nil
			removed = true
			log.Info("Removed all contract overrides")
		}
	} else if ctx.ContractOverrides != nil {
		contractsRemoved := false

		if c.Bool("delegation-manager") && ctx.ContractOverrides.DelegationManager != "" {
			ctx.ContractOverrides.DelegationManager = ""
			contractsRemoved = true
			log.Info("Removed delegation manager address")
		}

		if c.Bool("allocation-manager") && ctx.ContractOverrides.AllocationManager != "" {
			ctx.ContractOverrides.AllocationManager = ""
			contractsRemoved = true
			log.Info("Removed allocation manager address")
		}

		if c.Bool("strategy-manager") && ctx.ContractOverrides.StrategyManager != "" {
			ctx.ContractOverrides.StrategyManager = ""
			contractsRemoved = true
			log.Info("Removed strategy manager address")
		}

		if c.Bool("key-registrar") && ctx.ContractOverrides.KeyRegistrar != "" {
			ctx.ContractOverrides.KeyRegistrar = ""
			contractsRemoved = true
			log.Info("Removed key registrar address")
		}

		if c.Bool("release-manager") && ctx.ContractOverrides.ReleaseManager != "" {
			ctx.ContractOverrides.ReleaseManager = ""
			contractsRemoved = true
			log.Info("Removed release manager address")
		}

		if contractsRemoved {
			removed = true
			// Clean up empty contract overrides struct
			if ctx.ContractOverrides.DelegationManager == "" &&
				ctx.ContractOverrides.AllocationManager == "" &&
				ctx.ContractOverrides.StrategyManager == "" &&
				ctx.ContractOverrides.KeyRegistrar == "" &&
				ctx.ContractOverrides.ReleaseManager == "" {
				ctx.ContractOverrides = nil
				log.Debug("Cleaned up empty contract overrides")
			}
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
