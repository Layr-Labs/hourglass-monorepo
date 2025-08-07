package allocate

import (
	"fmt"
	"math/big"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
)

// Command returns the command for modifying operator allocations
func Command() *cli.Command {
	return &cli.Command{
		Name:  "allocate",
		Usage: "Modify operator allocations to AVS operator sets",
		Description: `Modify operator allocations to specific operator sets within an AVS.
This command allows operators to allocate their stake to AVS operator sets.

The operator address and AVS address must be configured in the context before running this command.

Example:
  hgctl allocate --operator-set-id 0 --strategy 0x789... --magnitude 1e18`,
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:     "operator-set-id",
				Usage:    "Operator set ID to allocate to",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "strategy",
				Usage:    "Strategy contract address",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "magnitude",
				Usage: "Allocation magnitude (e.g., '1e18' for full allocation)",
				Value: "1e18",
			},
		},
		Action: allocateAction,
	}
}

func allocateAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Get parameters
	operatorSetID := uint32(c.Uint64("operator-set-id"))
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
	)

	// Modify allocations
	if err := contractClient.ModifyAllocations(c.Context, operatorSetID, strategyAddress, magnitudeUint64); err != nil {
		return fmt.Errorf("failed to modify allocations: %w", err)
	}

	log.Info("Successfully modified allocations")
	return nil
}
