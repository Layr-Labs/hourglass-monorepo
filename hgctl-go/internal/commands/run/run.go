package run

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Run hourglass components",
		Description: `Automatically discovers and run the appropriate component (aggregator or executor)
based on the operator-set-id and AVS address in your context.

The command queries the AVS to determine which operator sets have which roles:
- If your operator-set-id matches the aggregator operator set, it deploys an aggregator
- If your operator-set-id matches an executor operator set, it deploys an executor

This removes the need to manually specify which component to deploy.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "release-id",
				Usage: "Release ID to deploy (defaults to latest)",
			},
			&cli.StringSliceFlag{
				Name:  "env",
				Usage: "Set environment variables (can be used multiple times)",
			},
			&cli.StringFlag{
				Name:  "env-file",
				Usage: "Load environment variables from file",
			},
			&cli.StringFlag{
				Name:  "network",
				Usage: "Docker network mode for aggregator (e.g., host, bridge)",
				Value: "bridge",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Validate configuration without deploying",
			},
		},
		Action: Run,
	}
}

func Run(c *cli.Context) error {
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	log := config.LoggerFromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	if currentCtx.AVSAddress == "" {
		return fmt.Errorf("AVS address not configured. Run 'hgctl context set --avs-address <address>' first")
	}

	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	log.Info("Discovering operator role for AVS",
		zap.String("avs", currentCtx.AVSAddress),
		zap.Uint32("operatorSetId", currentCtx.OperatorSetID))

	role, err := discoverOperatorRole(contractClient, currentCtx, log)
	if err != nil {
		return err
	}

	log.Info("Discovered operator role", zap.String("role", role))

	switch role {
	case "aggregator":
		fmt.Printf("Discovered role: aggregator\n")
		fmt.Printf("   Deploying aggregator for operator-set-id %d...\n\n", currentCtx.OperatorSetID)
		return deployAggregatorAction(c)
	case "executor":
		fmt.Printf("Discovered role: executor\n")
		fmt.Printf("   Deploying executor for operator-set-id %d...\n\n", currentCtx.OperatorSetID)
		return deployExecutorAction(c)
	default:
		return fmt.Errorf("unknown role: %s", role)
	}
}

func discoverOperatorRole(contractClient *client.ContractClient, ctx *config.Context, log logger.Logger) (string, error) {
	aggregatorOpSetId, err := contractClient.GetAvsAggregatorOperatorSetId(ctx.AVSAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get aggregator operator set ID: %w", err)
	}

	log.Debug("Retrieved aggregator operator set ID",
		zap.Uint32("aggregatorOperatorSetId", aggregatorOpSetId))

	if ctx.OperatorSetID == aggregatorOpSetId {
		return "aggregator", nil
	}

	executorOpSetIds, err := contractClient.GetAvsExecutorOperatorSetIds(ctx.AVSAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get executor operator set IDs: %w", err)
	}

	log.Debug("Retrieved executor operator set IDs",
		zap.Any("executorOperatorSetIds", executorOpSetIds))

	for _, execOpSetId := range executorOpSetIds {
		if ctx.OperatorSetID == execOpSetId {
			return "executor", nil
		}
	}

	return "", fmt.Errorf("operator-set-id %d does not match any known role for AVS %s (aggregator: %d, executors: %v)",
		ctx.OperatorSetID, ctx.AVSAddress, aggregatorOpSetId, executorOpSetIds)
}
