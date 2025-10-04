package register

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
)

// DeregisterAVSCommand returns the command for deregistering an operator from an AVS
func DeregisterAVSCommand() *cli.Command {
	return &cli.Command{
		Name:  "deregister-avs",
		Usage: "Deregister operator from an AVS",
		Description: `Deregister an operator from an AVS (Actively Validated Service).

This command deregisters the operator from the operator set configured in the context.

Prerequisites:
- AVS address must be configured: hgctl context set --avs-address <address>
- Operator set ID must be configured: hgctl context set --operator-set-id <id>
- Operator address must be configured: hgctl context set --operator-address <address>

Example:
  hgctl deregister-avs`,
		Action: deregisterAVSAction,
	}
}

func deregisterAVSAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	// Get context
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	if currentCtx == nil {
		return fmt.Errorf("no context configured. Run: hgctl context use <name>")
	}

	if currentCtx.AVSAddress == "" {
		return fmt.Errorf("AVS address not configured. Run: hgctl context set --avs-address <address>")
	}

	if currentCtx.OperatorAddress == "" {
		return fmt.Errorf("operator address not configured. Run: hgctl context set --operator-address <address>")
	}

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Get operator set ID from context (as a single ID, converted to slice for compatibility)
	operatorSetIDs := []uint32{currentCtx.OperatorSetID}

	log.Info("Deregistering operator from AVS",
		zap.Uint32("operatorSetId", currentCtx.OperatorSetID),
		zap.String("avsAddress", currentCtx.AVSAddress),
		zap.String("operatorAddress", currentCtx.OperatorAddress),
	)

	// Deregister operator from AVS
	if err := contractClient.DeregisterOperatorFromAVS(c.Context, operatorSetIDs); err != nil {
		return fmt.Errorf("failed to deregister operator from AVS: %w", err)
	}

	log.Info("Successfully deregistered operator from AVS")
	return nil
}
