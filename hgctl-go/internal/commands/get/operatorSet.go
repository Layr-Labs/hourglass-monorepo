package get

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
)

func operatorSetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "List operator sets for an AVS",
		ArgsUsage: "",
		Description: `List operator sets for an AVS.

The AVS address must be configured in the context before running this command.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "output",
				Usage: "Output format (table, json, yaml)",
				Value: "table",
			},
		},
		Action: getOperatorSetsAction,
	}
}

func getOperatorSetsAction(c *cli.Context) error {
	if c.NArg() != 0 {
		return cli.ShowSubcommandHelp(c)
	}

	log := middleware.GetLogger(c)

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	log.Info("Getting operator sets for AVS")

	// Get AVS config to discover operator sets
	avsConfig, err := contractClient.GetAVSConfig()
	if err != nil {
		return fmt.Errorf("failed to get AVS config: %w", err)
	}

	// Prepare output data
	type OperatorSetInfo struct {
		ID   uint32 `json:"id" yaml:"id"`
		Type string `json:"type" yaml:"type"`
		Name string `json:"name" yaml:"name"`
	}

	var operatorSets []OperatorSetInfo

	// Add aggregator operator set
	operatorSets = append(operatorSets, OperatorSetInfo{
		ID:   avsConfig.AggregatorOperatorSetID,
		Type: "aggregator",
		Name: fmt.Sprintf("Aggregator Operator Set %d", avsConfig.AggregatorOperatorSetID),
	})

	// Add executor operator sets
	for _, id := range avsConfig.ExecutorOperatorSetIDs {
		operatorSets = append(operatorSets, OperatorSetInfo{
			ID:   id,
			Type: "executor",
			Name: fmt.Sprintf("Executor Operator Set %d", id),
		})
	}

	log.Info("Found operator sets",
		zap.Int("count", len(operatorSets)),
		zap.Uint32("aggregator", avsConfig.AggregatorOperatorSetID),
		zap.Any("executors", avsConfig.ExecutorOperatorSetIDs))

	// Output the results
	outputFormat := c.String("output")
	formatter := output.NewFormatter(outputFormat)

	// Use a generic print method for structured data
	return formatter.Print(operatorSets)
}
