package context

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

func createCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new context",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Name of the context",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "rpc-url",
				Usage: "Ethereum RPC URL",
				Value: "http://localhost:8545",
			},
			&cli.StringFlag{
				Name:  "executor-address",
				Usage: "Executor gRPC address",
				Value: "127.0.0.1:9090",
			},
			&cli.IntFlag{
				Name:  "network-id",
				Usage: "Network ID",
				Value: 31337,
			},
			&cli.BoolFlag{
				Name:  "use",
				Usage: "Set as current context",
				Value: true,
			},
		},
		Action: contextCreateAction,
	}
}

func contextCreateAction(c *cli.Context) error {
	log := logger.FromContext(c.Context)

	name := c.String("name")
	rpcURL := c.String("rpc-url")
	executorAddress := c.String("executor-address")
	networkID := c.Int("network-id")
	setCurrent := c.Bool("use")

	// Load existing config or create new one
	cfg, err := config.LoadConfig()
	if err != nil {
		// If config doesn't exist, create a new one
		cfg = &config.Config{
			Contexts: make(map[string]*config.Context),
		}
	}

	// Check if context already exists
	if _, exists := cfg.Contexts[name]; exists {
		return fmt.Errorf("context '%s' already exists", name)
	}

	// Create new context
	ctx := &config.Context{
		RPCUrl:          rpcURL,
		ExecutorAddress: executorAddress,
		NetworkID:       uint64(networkID),
	}

	// Add to config
	cfg.Contexts[name] = ctx

	// Set as current if requested
	if setCurrent {
		cfg.CurrentContext = name
	}

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Info("Context created",
		zap.String("name", name),
		zap.String("rpc-url", rpcURL),
		zap.String("executor-address", executorAddress),
		zap.Int("network-id", networkID),
		zap.Bool("current", setCurrent))

	fmt.Printf("Context '%s' created successfully\n", name)
	if setCurrent {
		fmt.Printf("Current context set to '%s'\n", name)
	}

	return nil
}