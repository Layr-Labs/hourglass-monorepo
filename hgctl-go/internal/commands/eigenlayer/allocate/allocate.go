package allocate

import (
	"fmt"
	"math/big"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
)

// Command returns the command for modifying operator allocations
func Command() *cli.Command {
	return &cli.Command{
		Name:  "allocate",
		Usage: "Modify operator allocations to AVS operator sets",
		Description: `Modify operator allocations to the operator set configured in the context.
This command allows operators to allocate their stake to AVS operator sets.

Prerequisites:
- AVS address must be configured: hgctl context set --avs-address <address>
- Operator set ID must be configured: hgctl context set --operator-set-id <id>
- Operator address must be configured: hgctl context set --operator-address <address>

Example:
  hgctl allocate --strategy 0x789... --magnitude 1e18`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "strategy",
				Usage:    "Strategy contract address",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "magnitude",
				Usage:    "Allocation magnitude (e.g., '1e18' for full allocation)",
				Required: true,
			},
		},
		Action: allocateAction,
	}
}

func allocateAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	// Get context
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	if currentCtx == nil {
		return fmt.Errorf("no context configured. Run: hgctl context use <name>")
	}

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Get operator set ID from context
	operatorSetID := currentCtx.OperatorSetID
	strategyAddress := c.String("strategy")
	magnitudeStr := c.String("magnitude")

	// Parse magnitude
	magnitude := new(big.Int)
	_, ok := magnitude.SetString(magnitudeStr, 10)
	if !ok {
		// Try parsing as float with exponent (e.g., 1e18)
		magnitudeFloat := new(big.Float)
		_, _, err := magnitudeFloat.Parse(magnitudeStr, 10)
		if err != nil {
			return fmt.Errorf("invalid magnitude format: %s", magnitudeStr)
		}
		magnitude, _ = magnitudeFloat.Int(nil)
	}

	// Convert magnitude to uint64
	if !magnitude.IsUint64() {
		return fmt.Errorf("magnitude too large for uint64")
	}
	magnitudeUint64 := magnitude.Uint64()

	log.Info("Modifying operator allocations",
		zap.Uint32("operatorSetId", operatorSetID),
		zap.String("strategy", strategyAddress),
		zap.Uint64("magnitude", magnitudeUint64),
		zap.String("avsAddress", currentCtx.AVSAddress),
		zap.String("operatorAddress", currentCtx.OperatorAddress),
	)

	// Modify allocations
	if err := contractClient.ModifyAllocations(c.Context, operatorSetID, strategyAddress, magnitudeUint64); err != nil {
		return fmt.Errorf("failed to modify allocations: %w", err)
	}

	log.Info("Successfully modified allocations")
	return nil
}
