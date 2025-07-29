package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
)

// RegisterCommand returns the register command for registering an operator with EigenLayer
func RegisterCommand() *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "Register operator with EigenLayer",
		Description: `Register an operator with EigenLayer's delegation manager.
This command handles the operator registration process including setting up
the operator's details and metadata URI.

The operator address must be configured in the context before running this command.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "metadata-uri",
				Usage: "Metadata URI for operator details",
				Value: "",
			},
			&cli.Uint64Flag{
				Name:  "allocation-delay",
				Usage: "Allocation delay in seconds",
				Value: 0,
			},
		},
		Action: registerAction,
	}
}

func registerAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Get metadata URI
	metadataURI := c.String("metadata-uri")
	if metadataURI == "" {
		return fmt.Errorf("metadata-uri flag is required")
	}

	// Get allocation delay
	allocationDelay := uint32(c.Uint64("allocation-delay"))

	log.Info("Registering operator with EigenLayer",
		zap.String("metadataURI", metadataURI),
		zap.Uint32("allocationDelay", allocationDelay),
	)

	// Register operator
	if err := contractClient.RegisterAsOperator(c.Context, allocationDelay, metadataURI); err != nil {
		log.Error("Failed to register operator with EigenLayer", zap.Error(err))
		return fmt.Errorf("failed to register operator: %w", err)
	}

	return nil
}
