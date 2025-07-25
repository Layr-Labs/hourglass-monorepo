package commands

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/executor"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

func deployArtifactAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.ShowSubcommandHelp(c)
	}

	avsAddress := c.Args().Get(0)
	operatorSetID := uint32(c.Uint64("operator-set-id"))
	version := c.String("version")
	legacyDigest := c.String("legacy-digest")
	registryURL := c.String("registry-url")

	// Get context
	currentCtx := c.Context.Value("currentContext").(*config.Context)
	log := logger.FromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	// Legacy mode
	if legacyDigest != "" {
		log.Info("Deploying artifact in legacy mode",
			zap.String("avs", avsAddress),
			zap.String("digest", legacyDigest),
			zap.String("registry", registryURL))

		// Create executor client
		executorClient, err := executor.NewClient(currentCtx.ExecutorAddress, log)
		if err != nil {
			return fmt.Errorf("failed to create executor client: %w", err)
		}
		defer executorClient.Close()

		// Deploy using legacy method
		deploymentID, err := executorClient.DeployArtifact(c.Context, avsAddress, legacyDigest, registryURL)
		if err != nil {
			return fmt.Errorf("failed to deploy artifact: %w", err)
		}

		log.Info("Artifact deployed successfully",
			zap.String("deploymentID", deploymentID))

		fmt.Printf("Deployment ID: %s\n", deploymentID)
		return nil
	}

	// EigenRuntime mode
	if currentCtx.RPCUrl == "" {
		return fmt.Errorf("RPC URL not configured")
	}

	if currentCtx.ReleaseManagerAddress == "" {
		return fmt.Errorf("release manager address not configured")
	}

	log.Info("Fetching release from ReleaseManager",
		zap.String("avs", avsAddress),
		zap.Uint32("operatorSetID", operatorSetID))

	// Create contract client
	contractClient, err := client.NewContractClient(currentCtx.RPCUrl, log)
	if err != nil {
		return fmt.Errorf("failed to create contract client: %w", err)
	}
	defer contractClient.Close()

	var release *client.ReleaseManagerRelease

	if version == "" {
		// Get latest release
		nextReleaseId, err := contractClient.GetNextReleaseId(
			c.Context,
			common.HexToAddress(currentCtx.ReleaseManagerAddress),
			common.HexToAddress(avsAddress),
			operatorSetID,
		)
		if err != nil {
			return fmt.Errorf("failed to get next release ID: %w", err)
		}

		if nextReleaseId.Uint64() == 0 {
			return fmt.Errorf("no releases found for operator set %d", operatorSetID)
		}

		latestId := new(big.Int).Sub(nextReleaseId, big.NewInt(1))
		release, err = contractClient.GetRelease(
			c.Context,
			common.HexToAddress(currentCtx.ReleaseManagerAddress),
			common.HexToAddress(avsAddress),
			operatorSetID,
			latestId,
		)
		if err != nil {
			return fmt.Errorf("failed to get latest release: %w", err)
		}

		log.Info("Using latest release", zap.String("releaseID", latestId.String()))
	} else {
		// Get specific version
		versionBig := new(big.Int)
		versionBig.SetString(version, 10)

		release, err = contractClient.GetRelease(
			c.Context,
			common.HexToAddress(currentCtx.ReleaseManagerAddress),
			common.HexToAddress(avsAddress),
			operatorSetID,
			versionBig,
		)
		if err != nil {
			return fmt.Errorf("failed to get release %s: %w", version, err)
		}

		log.Info("Using release", zap.String("releaseID", version))
	}

	if len(release.Artifacts) == 0 {
		return fmt.Errorf("no artifacts found in release")
	}

	artifact := release.Artifacts[0]

	log.Info("Found release artifact",
		zap.String("digest", fmt.Sprintf("0x%x", artifact.Digest)),
		zap.String("registry", artifact.RegistryName))

	// Create OCI client
	ociClient := client.NewOCIClient(log)

	// Pull runtime spec
	log.Info("Pulling runtime spec from OCI registry...")
	spec, err := ociClient.PullRuntimeSpec(c.Context, artifact.RegistryName, fmt.Sprintf("0x%x", artifact.Digest))
	if err != nil {
		return fmt.Errorf("failed to pull runtime spec: %w", err)
	}

	// Find performer component in spec
	var performerImage string
	var performerDigest string

	for name, component := range spec.Spec {
		if name == "performer" {
			performerImage = component.Registry
			performerDigest = component.Digest
			break
		}
	}

	if performerImage == "" || performerDigest == "" {
		return fmt.Errorf("performer component not found in runtime spec")
	}

	log.Info("Deploying performer",
		zap.String("image", performerImage),
		zap.String("digest", performerDigest))

	// Create executor client
	executorClient, err := executor.NewClient(currentCtx.ExecutorAddress, log)
	if err != nil {
		return fmt.Errorf("failed to create executor client: %w", err)
	}
	defer executorClient.Close()

	// Deploy performer
	deploymentID, err := executorClient.DeployArtifact(c.Context, avsAddress, performerDigest, performerImage)
	if err != nil {
		return fmt.Errorf("failed to deploy performer: %w", err)
	}

	log.Info("Performer deployed successfully",
		zap.String("deploymentID", deploymentID))

	fmt.Printf("Deployment ID: %s\n", deploymentID)
	return nil
}
