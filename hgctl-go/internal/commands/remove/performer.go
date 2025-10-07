package remove

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
)

func removePerformerAction(c *cli.Context) error {
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	log := config.LoggerFromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	if currentCtx.ExecutorEndpoint == "" {
		return fmt.Errorf("executor endpoint not configured in context")
	}

	if currentCtx.AVSAddress == "" {
		return fmt.Errorf("AVS address not configured in context")
	}

	executorClient, err := client.NewExecutorClient(currentCtx.ExecutorEndpoint, log)
	if err != nil {
		return fmt.Errorf("failed to create executor client: %w", err)
	}
	defer executorClient.Close()

	log.Info("Querying executor for performers", zap.String("avsAddress", currentCtx.AVSAddress))

	performers, err := executorClient.GetPerformers(c.Context, currentCtx.AVSAddress)
	if err != nil {
		return fmt.Errorf("failed to get performers: %w", err)
	}

	if len(performers) == 0 {
		return fmt.Errorf("no performers found for AVS %s", currentCtx.AVSAddress)
	}

	log.Info("Found performers", zap.Int("count", len(performers)))

	for _, performer := range performers {
		log.Info("Removing performer",
			zap.String("performerID", performer.PerformerId),
			zap.String("avsAddress", performer.AvsAddress))

		err = executorClient.RemovePerformer(c.Context, performer.PerformerId)
		if err != nil {
			return fmt.Errorf("failed to remove performer %s: %w", performer.PerformerId, err)
		}

		fmt.Printf("✅ Performer removed successfully\n")
		fmt.Printf("   Performer ID: %s\n", performer.PerformerId)
		fmt.Printf("   AVS Address: %s\n", performer.AvsAddress)
	}

	fmt.Printf("\nRemoved %d performer(s) from executor %s\n\n", len(performers), currentCtx.ExecutorEndpoint)

	return nil
}
