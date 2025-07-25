package commands

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
)

func describeReleaseAction(c *cli.Context) error {
	if c.NArg() != 2 {
		return cli.ShowSubcommandHelp(c)
	}

	avsAddress := c.Args().Get(0)
	releaseID := c.Args().Get(1)
	operatorSetID := uint32(c.Uint64("operator-set-id"))

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

	log.Info("Fetching release details",
		zap.String("avs", avsAddress),
		zap.String("releaseID", releaseID),
		zap.Uint32("operatorSetID", operatorSetID))

	// Create contract client
	contractClient, err := client.NewContractClient(currentCtx.RPCUrl, log)
	if err != nil {
		return fmt.Errorf("failed to create contract client: %w", err)
	}
	defer contractClient.Close()

	// Parse release ID
	releaseIdBig := new(big.Int)
	releaseIdBig.SetString(releaseID, 10)

	// Get release from contract
	release, err := contractClient.GetRelease(
		c.Context,
		common.HexToAddress(currentCtx.ReleaseManagerAddress),
		common.HexToAddress(avsAddress),
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
	spec, err := ociClient.PullRuntimeSpec(c.Context, release.Artifacts[0].RegistryName, digest)
	if err != nil {
		return fmt.Errorf("failed to pull runtime spec: %w", err)
	}

	log.Info("Successfully pulled runtime spec")

	// Create release with spec for output
	releaseWithSpec := &output.ReleaseWithSpec{
		Release:     internalRelease,
		RuntimeSpec: spec,
	}

	// Output the result
	outputFormat := c.String("output")
	formatter := output.NewFormatter(outputFormat)
	return formatter.PrintReleaseWithSpec(releaseWithSpec)
}
