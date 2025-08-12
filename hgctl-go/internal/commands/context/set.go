package context

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
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
				Name:  "l1-rpc-url",
				Usage: "Set the L1 RPC URL",
			},
			&cli.StringFlag{
				Name:  "l2-rpc-url",
				Usage: "Set the L2 RPC URL",
			},
			&cli.StringSliceFlag{
				Name:  "env",
				Usage: "Set plain environment variables (KEY=VALUE)",
			},
			&cli.StringFlag{
				Name:  "env-secrets-path",
				Usage: "Set the path to environment secrets file",
			},
			&cli.StringFlag{
				Name:  "signer-key",
				Usage: "Set the signer keystore name (must be a keystore in this context)",
			},
			&cli.BoolFlag{
				Name:  "experimental",
				Usage: "Enable experimental features",
			},
		},
		Action: contextSetAction,
	}
}

func contextSetAction(c *cli.Context) error {
	log := config.LoggerFromContext(c.Context)

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, exists := cfg.Contexts[cfg.CurrentContext]
	if !exists {
		return fmt.Errorf("current context '%s' not found", cfg.CurrentContext)
	}

	updated := false

	if addr := c.String("avs-address"); addr != "" {
		ctx.AVSAddress = addr
		updated = true
		log.Info("Updated AVS address", zap.String("address", addr))
	}

	if addr := c.String("executor-address"); addr != "" {
		ctx.ExecutorAddress = addr
		updated = true
		log.Info("Updated Executor address", zap.String("address", addr))
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

	if url := c.String("l1-rpc-url"); url != "" {
		ctx.L1RPCUrl = url
		updated = true
		log.Info("Updated L1 RPC URL", zap.String("url", url))
	}

	if url := c.String("l2-rpc-url"); url != "" {
		ctx.L2RPCUrl = url
		updated = true
		log.Info("Updated L2 RPC URL", zap.String("url", url))
	}

	if path := c.String("env-secrets-path"); path != "" {
		ctx.EnvSecretsPath = path
		updated = true
		log.Info("Updated env secrets path", zap.String("path", path))
	}

	if c.IsSet("experimental") {
		ctx.Experimental = c.Bool("experimental")
		updated = true
		if ctx.Experimental {
			log.Info("Enabled experimental features")
		} else {
			log.Info("Disabled experimental features")
		}
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
