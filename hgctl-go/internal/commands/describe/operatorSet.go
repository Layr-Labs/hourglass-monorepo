package describe

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
)

func operatorSetCommand() *cli.Command {
	return &cli.Command{
		Name:      "operator-set",
		Usage:     "Describe a specific operator set",
		ArgsUsage: "[operator-set-id]",
		Description: `Describe a specific operator set including its type and metadata.

The AVS address must be configured in the context before running this command.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "output",
				Usage: "Output format (table, json, yaml)",
				Value: "table",
			},
		},
		Action: describeOperatorSetAction,
	}
}

func describeOperatorSetAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.ShowSubcommandHelp(c)
	}

	// Parse operator set ID from args
	operatorSetIDStr := c.Args().Get(0)
	var operatorSetID uint32
	if _, err := fmt.Sscanf(operatorSetIDStr, "%d", &operatorSetID); err != nil {
		return fmt.Errorf("invalid operator set ID: %s", operatorSetIDStr)
	}

	log := middleware.GetLogger(c)

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	log.Info("Describing operator set",
		zap.Uint32("operatorSetId", operatorSetID))

	// Get operator set metadata
	metadataURI, err := contractClient.GetOperatorSetMetadataURI(c.Context, operatorSetID)
	if err != nil {
		return fmt.Errorf("failed to get operator set metadata: %w", err)
	}

	// Get AVS config to determine operator set type
	avsConfig, err := contractClient.GetAVSConfig()
	if err != nil {
		return fmt.Errorf("failed to get AVS config: %w", err)
	}

	// Determine operator set type
	operatorSetType := "unknown"
	if operatorSetID == avsConfig.AggregatorOperatorSetID {
		operatorSetType = "aggregator"
	} else {
		for _, id := range avsConfig.ExecutorOperatorSetIDs {
			if operatorSetID == id {
				operatorSetType = "executor"
				break
			}
		}
	}

	// Prepare output data
	type OperatorSetDetails struct {
		ID          uint32 `json:"id" yaml:"id"`
		Type        string `json:"type" yaml:"type"`
		MetadataURI string `json:"metadataUri" yaml:"metadataUri"`
	}

	details := OperatorSetDetails{
		ID:          operatorSetID,
		Type:        operatorSetType,
		MetadataURI: metadataURI,
	}

	log.Info("Operator set details",
		zap.Uint32("id", operatorSetID),
		zap.String("type", operatorSetType),
		zap.String("metadataUri", metadataURI))

	// Output the results
	outputFormat := c.String("output")
	formatter := output.NewFormatter(outputFormat)

	// Use a generic print method for structured data
	return formatter.Print(details)
}
