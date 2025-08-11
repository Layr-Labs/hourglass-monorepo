package get

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
)

func releaseCommand() *cli.Command {
	return &cli.Command{
		Name:      "release",
		Usage:     "Get releases for an AVS",
		ArgsUsage: "",
		Description: `Get releases for an AVS from the release manager contract.

The AVS address must be configured in the context before running this command.

By default, this command will try to auto-detect the operator sets from the AVS
configuration. You can override this by specifying operator sets manually.`,
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:  "limit",
				Usage: "Limit the number of releases to retrieve per operator set",
				Value: 5,
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Output format (table, json, yaml)",
				Value: "table",
			},
			&cli.Uint64SliceFlag{
				Name:  "operator-sets",
				Usage: "Specify operator set IDs manually (overrides auto-detection)",
			},
		},
		Action: getReleaseAction,
	}
}

func getReleaseAction(c *cli.Context) error {
	if c.NArg() != 0 {
		return cli.ShowSubcommandHelp(c)
	}

	limit := c.Uint64("limit")
	log := middleware.GetLogger(c)

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Try to get operator sets
	var operatorSets []uint32

	// Check if operator sets are manually specified
	if manualSets := c.Uint64Slice("operator-sets"); len(manualSets) > 0 {
		// Convert uint64 slice to uint32
		for _, id := range manualSets {
			operatorSets = append(operatorSets, uint32(id))
		}
		log.Info("Using manually specified operator sets", zap.Any("operatorSets", operatorSets))
	} else {
		// Try to get from AVS config
		avsConfig, err := contractClient.GetAVSConfig()
		if err != nil {
			// If the call fails, it might not be an AVS registrar address
			// Fall back to default operator sets
			log.Info("Failed to get AVS config, using default operator sets", zap.Error(err))
			operatorSets = []uint32{0, 1}
		} else {
			// Combine aggregator and executor operator sets
			operatorSets = append([]uint32{avsConfig.AggregatorOperatorSetID}, avsConfig.ExecutorOperatorSetIDs...)
			log.Info("Retrieved AVS config",
				zap.Uint32("aggregatorOperatorSetId", avsConfig.AggregatorOperatorSetID),
				zap.Any("executorOperatorSetIds", avsConfig.ExecutorOperatorSetIDs))
		}
	}

	log.Info("Listing releases for AVS",
		zap.Any("operatorSets", operatorSets),
		zap.Uint64("limit", limit))

	releases, err := contractClient.GetReleases(c.Context, operatorSets, limit)
	if err != nil {
		return fmt.Errorf("failed to get releases: %w", err)
	}

	if len(releases) == 0 {
		log.Info("No releases found")
		return nil
	}

	log.Info("Found releases", zap.Int("count", len(releases)))

	// Output the results
	outputFormat := c.String("output")
	formatter := output.NewFormatter(outputFormat)
	return formatter.PrintReleases(releases)
}
