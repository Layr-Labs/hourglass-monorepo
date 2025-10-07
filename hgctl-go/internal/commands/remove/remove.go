package remove

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
		Name:  "remove",
		Usage: "Remove hourglass components",
		Description: `Automatically discovers and removes the appropriate component
based on the operator-set-id and AVS address in your context.

The command queries the AVS to determine which operator sets have which roles:
- If your operator-set-id matches the aggregator operator set, it deregisters the AVS from the aggregator
- If your operator-set-id matches an executor operator set, it removes the performer

This removes the need to manually specify which component to remove.`,
		Flags: []cli.Flag{},
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

	if currentCtx.OperatorSetID == 0 {
		return fmt.Errorf("operator-set-id not configured. Run 'hgctl context set --operator-set-id <id>' first")
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
		fmt.Printf("   Deregistering AVS from aggregator...\n\n")
		return removeAggregatorAction(c, currentCtx, log)
	case "executor":
		fmt.Printf("Discovered role: executor\n")
		fmt.Printf("   Removing performer...\n\n")
		return removePerformerAction(c)
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

func removeAggregatorAction(c *cli.Context, currentCtx *config.Context, log logger.Logger) error {
	if currentCtx.AggregatorEndpoint == "" {
		return fmt.Errorf("aggregator endpoint not configured in context")
	}

	aggregatorClient, err := client.NewAggregatorClient(currentCtx.AggregatorEndpoint, log)
	if err != nil {
		return fmt.Errorf("failed to create aggregator client: %w", err)
	}
	defer func() {
		if err := aggregatorClient.Close(); err != nil {
			log.Warn("Failed to close aggregator client", zap.Error(err))
		}
	}()

	log.Info("Deregistering AVS from aggregator",
		zap.String("avsAddress", currentCtx.AVSAddress),
		zap.String("aggregatorEndpoint", currentCtx.AggregatorEndpoint))

	err = aggregatorClient.DeRegisterAvs(c.Context, currentCtx.AVSAddress)
	if err != nil {
		return fmt.Errorf("failed to deregister AVS: %w", err)
	}

	fmt.Printf("\n✅ AVS deregistered from aggregator successfully\n")
	fmt.Printf("   AVS Address: %s\n", currentCtx.AVSAddress)
	fmt.Printf("   Aggregator Endpoint: %s\n\n", currentCtx.AggregatorEndpoint)

	return nil
}
