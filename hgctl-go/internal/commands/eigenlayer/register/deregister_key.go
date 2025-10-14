package register

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func DeregisterKeyCommand() *cli.Command {
	return &cli.Command{
		Name:  "deregister-key",
		Usage: "Deregister system signing key from AVS",
		Description: `Deregister system signing keys from an AVS operator set.

Prerequisites: AVS address, operator set ID, and operator signer must be configured.
Note: Keys remain in the global key registry to prevent reuse.`,
		Action: deregisterKeyAction,
	}
}

func deregisterKeyAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)

	if currentCtx.AVSAddress == "" {
		return fmt.Errorf("AVS address not configured. Run: hgctl context set --avs-address <address>")
	}

	if currentCtx.OperatorKeys == nil {
		return fmt.Errorf("operator signer not configured. Run: hgctl signer operator")
	}

	operatorSetID := currentCtx.OperatorSetID

	log.Info("Deregistering key",
		zap.Uint32("operatorSetId", operatorSetID),
		zap.String("avsAddress", currentCtx.AVSAddress))

	if err := contractClient.DeregisterKey(
		c.Context,
		operatorSetID,
	); err != nil {
		log.Error("Failed to deregister key", zap.Error(err))
		return fmt.Errorf("failed to deregister key: %w", err)
	}

	log.Info("Successfully deregistered key")
	return nil
}
