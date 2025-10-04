package remove

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
)

func performerCommand() *cli.Command {
	return &cli.Command{
		Name:      "performer",
		Usage:     "Remove a performer",
		ArgsUsage: "[performer-id]",
		Action:    removePerformerAction,
	}
}

func removePerformerAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.ShowSubcommandHelp(c)
	}

	performerID := c.Args().Get(0)

	// Get context
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	log := config.LoggerFromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	if currentCtx.ExecutorEndpoint == "" {
		return fmt.Errorf("executor address not configured")
	}

	log.Info("Removing performer",
		zap.String("performerID", performerID))

	// Create executor client
	executorClient, err := client.NewExecutorClient(currentCtx.ExecutorEndpoint, log)
	if err != nil {
		return fmt.Errorf("failed to create executor client: %w", err)
	}
	defer executorClient.Close()

	// Remove performer
	err = executorClient.RemovePerformer(c.Context, performerID)
	if err != nil {
		return fmt.Errorf("failed to remove performer: %w", err)
	}

	log.Info("Performer removed successfully",
		zap.String("performerID", performerID))

	fmt.Printf("Performer %s removed successfully\n", performerID)
	return nil
}
