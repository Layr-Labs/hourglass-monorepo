package allocate

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
)

// SetAllocationDelayCommand returns the command for setting allocation delay
func SetAllocationDelayCommand() *cli.Command {
	return &cli.Command{
		Name:  "set-allocation-delay",
		Usage: "Set allocation delay for an operator",
		Description: `Set the allocation delay for an operator.
The allocation delay determines how long an operator must wait before allocation changes take effect.

The operator address must be configured in the context before running this command.

Example:
  hgctl set-allocation-delay --delay 86400  # 24 hours in seconds`,
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:     "delay",
				Usage:    "Allocation delay in seconds",
				Required: true,
			},
		},
		Action: setAllocationDelayAction,
	}
}

func setAllocationDelayAction(c *cli.Context) error {
	log := middleware.GetLogger(c)
	// Get parameters
	delay := uint32(c.Uint64("delay"))

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}
	log.Info("Setting allocation delay", zap.Uint32("delay", delay))

	// Set allocation delay
	if err := contractClient.SetAllocationDelay(c.Context, delay); err != nil {
		log.Error("failed to set allocation delay", zap.Error(err))
		return fmt.Errorf("failed to set allocation delay: %w", err)
	}

	log.Info("Successfully set allocation delay")
	return nil
}
