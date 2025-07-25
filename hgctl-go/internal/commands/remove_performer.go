package commands

import (
    "fmt"
    
    "github.com/urfave/cli/v2"
    "go.uber.org/zap"
    
    "github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
    "github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/executor"
    "github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

func removePerformerAction(c *cli.Context) error {
    if c.NArg() != 1 {
        return cli.ShowSubcommandHelp(c)
    }
    
    performerID := c.Args().Get(0)
    
    // Get context
    currentCtx := c.Context.Value("currentContext").(*config.Context)
    log := logger.FromContext(c.Context)
    
    if currentCtx == nil {
        return fmt.Errorf("no context configured")
    }
    
    if currentCtx.ExecutorAddress == "" {
        return fmt.Errorf("executor address not configured")
    }
    
    log.Info("Removing performer",
        zap.String("performerID", performerID))
    
    // Create executor client
    executorClient, err := executor.NewClient(currentCtx.ExecutorAddress, log)
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
