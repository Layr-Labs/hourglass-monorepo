package describe

import (
	"fmt"
	"math/big"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
)

func releaseCommand() *cli.Command {
	return &cli.Command{
		Name:      "release",
		Usage:     "Describe a specific release",
		ArgsUsage: "[release-id]",
		Description: `Describe a specific release by fetching its details from the release manager
and pulling the runtime specification from the OCI registry.

The AVS address must be configured in the context before running this command.`,
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:  "operator-set-id",
				Usage: "Operator set ID",
				Value: 0,
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Output format (table, json, yaml)",
				Value: "table",
			},
		},
		Action: describeReleaseAction,
	}
}

func describeReleaseAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.ShowSubcommandHelp(c)
	}

	releaseID := c.Args().Get(0)
	operatorSetID := uint32(c.Uint64("operator-set-id"))

	log := middleware.GetLogger(c)

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	log.Info("Fetching release details",
		zap.String("releaseID", releaseID),
		zap.Uint32("operatorSetID", operatorSetID))

	// Parse release ID
	releaseIdBig := new(big.Int)
	if _, ok := releaseIdBig.SetString(releaseID, 10); !ok {
		return fmt.Errorf("invalid release ID: %s", releaseID)
	}

	// Get release from contract
	release, err := contractClient.GetRelease(
		c.Context,
		operatorSetID,
		releaseIdBig,
	)
	if err != nil {
		return fmt.Errorf("failed to get release: %w", err)
	}

	if len(release.Artifacts) == 0 {
		return fmt.Errorf("no artifacts found in release")
	}

	// Convert to internal release type
	internalRelease := &client.Release{
		ID: releaseID,
		OperatorSetReleases: map[string]client.OperatorSetRelease{
			fmt.Sprintf("%d", operatorSetID): {
				Digest:   fmt.Sprintf("0x%x", release.Artifacts[0].Digest),
				Registry: release.Artifacts[0].RegistryName,
			},
		},
		UpgradeByTime: release.UpgradeByTime,
	}

	log.Info("Found release",
		zap.String("digest", fmt.Sprintf("0x%x", release.Artifacts[0].Digest)),
		zap.String("registry", release.Artifacts[0].RegistryName))

	// Create OCI client
	ociClient := client.NewOCIClient(log)

	// Pull runtime spec from registry
	log.Info("Pulling runtime spec from OCI registry...")

	digest := fmt.Sprintf("0x%x", release.Artifacts[0].Digest)
	spec, specRaw, err := ociClient.PullRuntimeSpec(c.Context, release.Artifacts[0].RegistryName, digest)
	if err != nil {
		log.Error("Failed to pull runtime spec", zap.Error(err))
		return fmt.Errorf("failed to pull runtime spec: %w", err)
	}

	log.Info("Successfully pulled runtime spec", zap.Int("specSize", len(specRaw)))

	// Create release with spec for output
	releaseWithSpec := &output.ReleaseWithSpec{
		Release:        internalRelease,
		RuntimeSpec:    spec,
		RuntimeSpecRaw: specRaw,
	}

	// Output the result
	outputFormat := c.String("output")
	formatter := output.NewFormatter(outputFormat)
	return formatter.PrintReleaseWithSpec(releaseWithSpec)
}
