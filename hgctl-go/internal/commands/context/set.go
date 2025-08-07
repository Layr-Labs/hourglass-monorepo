package context

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

func setCommand() *cli.Command {
	return &cli.Command{
		Name:  "set",
		Usage: "Set context properties",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "executor-address",
				Usage: "Set the executor gRPC address",
			},
			&cli.StringFlag{
				Name:  "avs-address",
				Usage: "Set the AVS contract address",
			},
			&cli.StringFlag{
				Name:  "operator-address",
				Usage: "Set the operator address",
			},
			&cli.Uint64Flag{
				Name:  "operator-set-id",
				Usage: "Set the operator set ID",
			},
			&cli.StringFlag{
				Name:  "rpc-url",
				Usage: "Set the Ethereum RPC URL",
			},
			&cli.StringFlag{
				Name:  "release-manager",
				Usage: "Set the release manager contract address",
			},
			&cli.StringFlag{
				Name:  "delegation-manager",
				Usage: "Set the delegation manager contract address",
			},
			&cli.StringFlag{
				Name:  "allocation-manager",
				Usage: "Set the allocation manager contract address",
			},
			&cli.StringFlag{
				Name:  "strategy-manager",
				Usage: "Set the strategy manager contract address",
			},
			&cli.StringFlag{
				Name:  "key-registrar",
				Usage: "Set the key registrar contract address",
			},
			&cli.StringSliceFlag{
				Name:  "env",
				Usage: "Set plain environment variables (KEY=VALUE)",
			},
			&cli.StringFlag{
				Name:  "env-secrets-path",
				Usage: "Set the path to environment secrets file",
			},
		},
		Action: contextSetAction,
	}
}

func contextSetAction(c *cli.Context) error {
	log := logger.FromContext(c.Context)

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, exists := cfg.Contexts[cfg.CurrentContext]
	if !exists {
		return fmt.Errorf("current context '%s' not found", cfg.CurrentContext)
	}

	updated := false

	if addr := c.String("executor-address"); addr != "" {
		ctx.ExecutorAddress = addr
		updated = true
		log.Info("Updated executor address", zap.String("address", addr))
	}

	if addr := c.String("avs-address"); addr != "" {
		ctx.AVSAddress = addr
		updated = true
		log.Info("Updated AVS address", zap.String("address", addr))
	}

	if addr := c.String("operator-address"); addr != "" {
		ctx.OperatorAddress = addr
		updated = true
		log.Info("Updated operator address", zap.String("address", addr))
	}

	if id := c.Uint64("operator-set-id"); c.IsSet("operator-set-id") {
		ctx.OperatorSetID = uint32(id)
		updated = true
		log.Info("Updated operator set ID", zap.Uint32("id", uint32(id)))
	}

	if url := c.String("rpc-url"); url != "" {
		ctx.L1RPCUrl = url
		updated = true
		log.Info("Updated RPC URL", zap.String("url", url))
	}

	// Handle environment variables
	envFlags := c.StringSlice("env")
	if len(envFlags) > 0 {
		if ctx.EnvironmentVars == nil {
			ctx.EnvironmentVars = make(map[string]string)
		}

		for _, env := range envFlags {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid env format: %s (expected KEY=VALUE)", env)
			}

			key := parts[0]
			value := parts[1]

			// Check if it's a secret variable
			if config.IsSecretVariable(key) {
				log.Warn("Skipping secret variable (use runtime flags instead)",
					zap.String("variable", key))
				continue
			}

			ctx.EnvironmentVars[key] = value
			log.Info("Set environment variable", zap.String("key", key))
			updated = true
		}
	}

	// Handle contract addresses
	contractsUpdated := false

	if addr := c.String("delegation-manager"); addr != "" {
		if ctx.ContractOverrides == nil {
			ctx.ContractOverrides = &config.ContractOverrides{}
		}
		ctx.ContractOverrides.DelegationManager = addr
		contractsUpdated = true
		log.Debug("Updated delegation manager address", zap.String("address", addr))
	}

	if addr := c.String("allocation-manager"); addr != "" {
		if ctx.ContractOverrides == nil {
			ctx.ContractOverrides = &config.ContractOverrides{}
		}
		ctx.ContractOverrides.AllocationManager = addr
		contractsUpdated = true
		log.Debug("Updated allocation manager address", zap.String("address", addr))
	}

	if addr := c.String("strategy-manager"); addr != "" {
		if ctx.ContractOverrides == nil {
			ctx.ContractOverrides = &config.ContractOverrides{}
		}
		ctx.ContractOverrides.StrategyManager = addr
		contractsUpdated = true
		log.Debug("Updated strategy manager address", zap.String("address", addr))
	}

	if addr := c.String("key-registrar"); addr != "" {
		if ctx.ContractOverrides == nil {
			ctx.ContractOverrides = &config.ContractOverrides{}
		}
		ctx.ContractOverrides.KeyRegistrar = addr
		contractsUpdated = true
		log.Debug("Updated key registrar override address", zap.String("address", addr))
	}

	if addr := c.String("release-manager"); addr != "" {
		if ctx.ContractOverrides == nil {
			ctx.ContractOverrides = &config.ContractOverrides{}
		}
		ctx.ContractOverrides.ReleaseManager = addr
		contractsUpdated = true
		log.Debug("Updated release manager address", zap.String("address", addr))
	}

	if contractsUpdated {
		updated = true
	}

	if path := c.String("env-secrets-path"); path != "" {
		ctx.EnvSecretsPath = path
		updated = true
		log.Info("Updated env secrets path", zap.String("path", path))
	}

	if !updated {
		return fmt.Errorf("no values provided to update")
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Context '%s' updated\n", cfg.CurrentContext)
	return nil
}
