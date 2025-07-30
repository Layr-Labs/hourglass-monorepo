package get

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
)

func releaseCommand() *cli.Command {
	return &cli.Command{
		Name:      "release",
		Usage:     "Get releases for an AVS",
		ArgsUsage: "<avs-address>",
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
		},
		Action: getReleaseAction,
	}
}

func getReleaseAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.ShowSubcommandHelp(c)
	}

	avsAddress := c.Args().Get(0)
	limit := c.Uint64("limit")

	// Get context
	currentCtx := c.Context.Value("currentContext").(*config.Context)
	log := logger.FromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	if currentCtx.RPCUrl == "" {
		return fmt.Errorf("RPC URL not configured")
	}

	if currentCtx.ReleaseManagerAddress == "" {
		return fmt.Errorf("release manager address not configured")
	}

	// Get all releases for all operator sets
	operatorSets := []uint32{0, 1} // TODO: Get actual operator sets from contract

	log.Info("Listing releases for AVS",
		zap.String("avs", avsAddress),
		zap.Any("operatorSets", operatorSets),
		zap.Uint64("limit", limit))

	// Create contract client
	contractClient, err := client.NewContractClient(currentCtx.RPCUrl, log)
	if err != nil {
		return fmt.Errorf("failed to create contract client: %w", err)
	}
	defer contractClient.Close()

	releases, err := contractClient.GetReleases(
		c.Context,
		common.HexToAddress(currentCtx.ReleaseManagerAddress),
		common.HexToAddress(avsAddress),
		operatorSets,
		limit,
	)
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
