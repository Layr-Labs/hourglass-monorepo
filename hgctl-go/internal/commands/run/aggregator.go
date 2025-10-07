package run

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/templates"
)

func aggregatorCommand() *cli.Command {
	return &cli.Command{
		Name:      "aggregator",
		Usage:     "Deploy the aggregator component",
		ArgsUsage: "",
		Description: `Deploy the aggregator component from a release.

The AVS address must be configured in the context before running this command.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "release-id",
				Usage: "Release ID to deploy (defaults to latest)",
			},
			&cli.StringSliceFlag{
				Name:  "env",
				Usage: "Set environment variables (can be used multiple times)",
			},
			&cli.StringFlag{
				Name:  "env-file",
				Usage: "Load environment variables from file",
			},
			&cli.StringFlag{
				Name:  "network",
				Usage: "Docker network mode (e.g., host, bridge)",
				Value: "bridge",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Validate configuration without deploying",
			},
		},
		Action: deployAggregatorAction,
	}
}

// AggregatorDeployer handles aggregator-specific deployment logic
type AggregatorDeployer struct {
	*PlatformDeployer
	networkMode string
	dryRun      bool
}

func deployAggregatorAction(c *cli.Context) error {
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	log := config.LoggerFromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	if currentCtx.AVSAddress == "" {
		return fmt.Errorf("AVS address not configured. Run 'hgctl context set --avs-address <address>' first")
	}

	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	platform := NewPlatformDeployer(
		currentCtx,
		log,
		contractClient,
		currentCtx.AVSAddress,
		currentCtx.OperatorSetID,
		c.String("release-id"),
		c.String("env-file"),
		c.StringSlice("env"),
	)

	deployer := &AggregatorDeployer{
		PlatformDeployer: platform,
		networkMode:      c.String("network"),
		dryRun:           c.Bool("dry-run"),
	}

	return deployer.Deploy(c.Context)
}

// Deploy executes the aggregator deployment
func (d *AggregatorDeployer) Deploy(ctx context.Context) error {
	containerName := fmt.Sprintf("hgctl-aggregator-%s-%s", d.Context.Name, d.Context.AVSAddress)

	// Fetch runtime spec first to determine required chain IDs
	spec, err := d.FetchRuntimeSpec(ctx)
	if err != nil {
		return err
	}

	component, err := d.ExtractComponent(spec, "aggregator")
	if err != nil {
		return err
	}

	// Determine which chain IDs are required based on the runtime spec
	chainIDs, err := d.extractChainIDsFromSpec(component)
	if err != nil {
		return err
	}

	// Check if an aggregator container is already running
	isRunning, containerID, err := d.CheckContainerRunning(containerName)
	if err != nil {
		return fmt.Errorf("failed to check aggregator container status: %w", err)
	}

	if isRunning {
		d.Log.Info("Found existing running aggregator container",
			zap.String("container", containerName),
			zap.String("containerID", containerID[:12]))

		if err = d.registerAvsWithAggregator(ctx, containerName, containerID, chainIDs); err != nil {
			return err
		}
		return nil
	}

	cfg := &DeploymentConfig{
		Env: d.LoadEnvironmentVariables(),
	}

	if err := ValidateComponentSpec(component, cfg.Env); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Validate that required passwords are set for system keystores
	if cfg.Env["SYSTEM_BN254_KEYSTORE_PATH"] != "" || cfg.Env["SYSTEM_ECDSA_KEYSTORE_PATH"] != "" {
		if cfg.Env["SYSTEM_KEYSTORE_PASSWORD"] == "" {
			return fmt.Errorf("SYSTEM_KEYSTORE_PASSWORD environment variable required for system keystores")
		}
	}

	if d.dryRun {
		return d.handleDryRun(cfg, component.Registry, component.Digest)
	}

	tempCfg, err := d.CreateTempDirectories("aggregator")
	if err != nil {
		return err
	}
	// Merge configs
	cfg.TempDir = tempCfg.TempDir
	cfg.ConfigDir = tempCfg.ConfigDir
	cfg.KeystoreDir = tempCfg.KeystoreDir

	if err := d.generateConfiguration(cfg); err != nil {
		return err
	}

	if err := d.deployContainer(component, cfg); err != nil {
		return err
	}

	// Save aggregator endpoint to context
	if err := d.saveAggregatorEndpoint(cfg); err != nil {
		d.Log.Warn("Failed to save aggregator endpoint to context", zap.Error(err))
	}

	if err := d.registerAvsWithAggregator(ctx, containerName, "", chainIDs); err != nil {
		d.Log.Warn("Failed to register AVS with newly deployed aggregator", zap.Error(err))
		fmt.Printf("\n⚠️  Warning: Aggregator deployed successfully but AVS registration failed: %v\n", err)
	}

	return nil
}

// extractChainIDsFromSpec determines which chain IDs are required based on the runtime spec
func (d *AggregatorDeployer) extractChainIDsFromSpec(component *runtime.ComponentSpec) ([]uint32, error) {
	hasL2ChainID := false
	for _, envVar := range component.Env {
		if envVar.Name == "L2_CHAIN_ID" {
			hasL2ChainID = true
			break
		}
	}

	// Build chain IDs slice based on what's required
	chainIDs := []uint32{d.Context.L1ChainID}

	if hasL2ChainID {
		if d.Context.L2ChainID == 0 {
			return nil, fmt.Errorf("L2_CHAIN_ID is required by the runtime spec but not configured in context. Run: hgctl context set --l2-chain-id <id>")
		}
		chainIDs = append(chainIDs, d.Context.L2ChainID)
		d.Log.Info("Including both L1 and L2 chain IDs (L2_CHAIN_ID required in runtime spec)",
			zap.Uint32("l1ChainID", d.Context.L1ChainID),
			zap.Uint32("l2ChainID", d.Context.L2ChainID))
	} else {
		d.Log.Info("Using L1 chain ID only (L2_CHAIN_ID not required in runtime spec)",
			zap.Uint32("l1ChainID", d.Context.L1ChainID))
	}

	return chainIDs, nil
}

// registerAvsWithAggregator registers the AVS with an aggregator
func (d *AggregatorDeployer) registerAvsWithAggregator(
	ctx context.Context,
	containerName string,
	containerID string,
	chainIDs []uint32,
) error {
	// Try to get the management port from the running container
	mgmtPort, err := d.GetContainerPort(containerName, "9010")
	if err != nil {
		// Try default port if we can't get it from Docker
		mgmtPort = "9010"
		d.Log.Warn("Could not get management port from container, using default",
			zap.String("port", mgmtPort),
			zap.Error(err))
	}

	// Create aggregator client
	aggregatorAddr := fmt.Sprintf("localhost:%s", mgmtPort)
	aggregatorClient, err := client.NewAggregatorClient(aggregatorAddr, d.Log)
	if err != nil {
		return fmt.Errorf("failed to create aggregator client: %w", err)
	}
	defer func() {
		if err := aggregatorClient.Close(); err != nil {
			d.Log.Warn("Failed to close aggregator client", zap.Error(err))
		}
	}()

	// Register AVS with the aggregator
	if err := aggregatorClient.RegisterAvs(ctx, d.Context.AVSAddress, chainIDs); err != nil {
		return fmt.Errorf("failed to register AVS with aggregator: %w", err)
	}

	fmt.Printf("\n✅ AVS registered with aggregator\n")
	fmt.Printf("   AVS Address: %s\n", d.Context.AVSAddress)
	fmt.Printf("   Chain IDs: %v\n", chainIDs)
	fmt.Printf("   Management Port: %s\n\n", mgmtPort)

	return nil
}

// saveAggregatorEndpoint saves the aggregator management gRPC address to the context configuration
func (d *AggregatorDeployer) saveAggregatorEndpoint(cfg *DeploymentConfig) error {
	aggregatorMgmtPort := cfg.Env["AGGREGATOR_MGMT_PORT"]
	if aggregatorMgmtPort == "" {
		aggregatorMgmtPort = "9010"
	}

	aggregatorEndpoint := fmt.Sprintf("localhost:%s", aggregatorMgmtPort)

	configData, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if ctx, exists := configData.Contexts[configData.CurrentContext]; exists {
		ctx.AggregatorEndpoint = aggregatorEndpoint

		if err := config.SaveConfig(configData); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		d.Log.Info("Saved aggregator management endpoint to context",
			zap.String("endpoint", aggregatorEndpoint),
			zap.String("context", configData.CurrentContext))
	}

	return nil
}

// generateConfiguration generates aggregator configuration files
func (d *AggregatorDeployer) generateConfiguration(cfg *DeploymentConfig) error {
	aggregatorConfig, err := templates.BuildAggregatorConfig(cfg.Env)
	if err != nil {
		return fmt.Errorf("failed to build aggregator config: %w", err)
	}

	cfg.ConfigPath = filepath.Join(cfg.ConfigDir, "aggregator.yaml")
	if err := os.WriteFile(cfg.ConfigPath, aggregatorConfig, 0600); err != nil {
		return fmt.Errorf("failed to write aggregator config: %w", err)
	}

	d.Log.Info("Configuration written to", zap.String("path", cfg.ConfigPath))

	if stat, err := os.Stat(cfg.ConfigPath); err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	} else {
		d.Log.Info("Config file verified", zap.Int64("size", stat.Size()))
	}

	return nil
}

// deployContainer handles the aggregator-specific container deployment
func (d *AggregatorDeployer) deployContainer(
	component *runtime.ComponentSpec,
	cfg *DeploymentConfig,
) error {

	containerName := fmt.Sprintf("hgctl-aggregator-%s-%s", d.Context.Name, d.Context.AVSAddress)
	imageRef := fmt.Sprintf("%s@%s", component.Registry, component.Digest)

	if err := d.PullDockerImage(imageRef, "aggregator"); err != nil {
		return err
	}

	d.CleanupExistingContainer(containerName)

	dockerArgs := d.BuildDockerArgs(containerName, component, cfg)

	if d.networkMode != "" {
		dockerArgs = append(dockerArgs[:2], append([]string{"--network", d.networkMode}, dockerArgs[2:]...)...)
	}

	mgmtPort := cfg.Env["AGGREGATOR_MGMT_PORT"]
	if mgmtPort == "" {
		mgmtPort = "9010"
	}
	dockerArgs = append(dockerArgs, "-p", fmt.Sprintf("%s:%s", mgmtPort, mgmtPort))
	d.Log.Info("Exposing aggregator management port", zap.String("port", mgmtPort))

	// Mount system keystores if configured (operator keystores are converted to private keys)
	d.MountSystemKeystores(&dockerArgs, cfg)

	dockerArgs = d.InjectFileContentsAsEnvVars(dockerArgs)

	dockerArgs = append(dockerArgs, imageRef)

	dockerArgs = append(dockerArgs, "aggregator", "run", "--config", "/config/aggregator.yaml")

	containerID, err := d.RunDockerContainer(dockerArgs, "aggregator")
	if err != nil {
		return err
	}

	d.printSuccessMessage(containerName, containerID, cfg)
	return nil
}

// handleDryRun handles the dry-run scenario
func (d *AggregatorDeployer) handleDryRun(cfg *DeploymentConfig, registry string, digest string) error {
	d.Log.Info("✅ Dry run successful - aggregator configuration is valid")

	d.Log.Info("Configuration:",
		zap.String("keystoreName", cfg.Env[config.KeystoreName]),
		zap.String("operatorAddress", d.Context.OperatorAddress),
		zap.String("avsAddress", d.Context.AVSAddress),
		zap.String("registry", registry),
		zap.String("digest", digest),
		zap.String("network", d.networkMode))

	fmt.Println("\n✅ Configuration is valid. Run without --dry-run to deploy.")
	return nil
}

// printSuccessMessage prints a user-friendly success message
func (d *AggregatorDeployer) printSuccessMessage(containerName, containerID string, cfg *DeploymentConfig) {
	fmt.Printf("\n✅ Aggregator container deployed successfully\n")
	fmt.Printf("   Container ID: %s\n", containerID[:12])
	fmt.Printf("   Config Path: %s\n\n", cfg.ConfigPath)
	fmt.Printf("Useful commands:\n")
	fmt.Printf("  View logs:  docker logs -f %s\n", containerName)
	fmt.Printf("  Stop:       docker stop %s\n", containerName)
	fmt.Printf("  Restart:    docker restart %s\n\n", containerName)
}
