package commands

import (
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/executor"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
)

func getPerformerAction(c *cli.Context) error {
	var avsAddress string
	if c.NArg() > 0 {
		avsAddress = c.Args().Get(0)
	}

	// Get context
	currentCtx := c.Context.Value("currentContext").(*config.Context)
	log := logger.FromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	if currentCtx.ExecutorAddress == "" {
		return fmt.Errorf("executor address not configured")
	}

	log.Info("Getting performers",
		zap.String("avs", avsAddress))

	// Create executor client
	executorClient, err := executor.NewClient(currentCtx.ExecutorAddress, log)
	if err != nil {
		return fmt.Errorf("failed to create executor client: %w", err)
	}
	defer executorClient.Close()

	// Get performers
	performers, err := executorClient.GetPerformers(c.Context, avsAddress)
	if err != nil {
		return fmt.Errorf("failed to get performers: %w", err)
	}

	if len(performers) == 0 {
		log.Info("No performers found")
		return nil
	}

	log.Info("Found performers", zap.Int("count", len(performers)))

	// Format output
	outputFormat := c.String("output")

	switch outputFormat {
	case "json":
		formatter := output.NewFormatter(outputFormat)
		return formatter.PrintJSON(performers)
	case "yaml":
		formatter := output.NewFormatter(outputFormat)
		return formatter.PrintYAML(performers)
	default:
		// Table output
		table := tablewriter.NewWriter(c.App.Writer)
		table.SetHeader([]string{"ID", "AVS ADDRESS", "DIGEST", "STATUS", "LAST HEALTH CHECK"})

		for _, p := range performers {
			digest := p.ArtifactDigest
			if len(digest) > 12 {
				digest = digest[:12] + "..."
			}

			table.Append([]string{
				p.PerformerId,
				p.AvsAddress,
				digest,
				p.Status,
				p.LastHealthCheck,
			})
		}

		table.Render()
	}

	return nil
}
