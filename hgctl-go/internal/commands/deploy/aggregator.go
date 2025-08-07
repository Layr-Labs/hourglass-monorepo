package deploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
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
			&cli.Uint64Flag{
				Name:  "operator-set-id",
				Usage: "Operator set ID",
				Value: 0,
			},
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
	// Get context and validate
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	log := logger.FromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	if currentCtx.AVSAddress == "" {
		return fmt.Errorf("AVS address not configured. Run 'hgctl context set --avs-address <address>' first")
	}

	// Get contract client
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Create platform deployer
	platform := NewPlatformDeployer(
		currentCtx,
		log,
		contractClient,
		uint32(c.Uint64("operator-set-id")),
		c.String("release-id"),
		c.String("env-file"),
		c.StringSlice("env"),
	)

	// Create aggregator deployer
	deployer := &AggregatorDeployer{
		PlatformDeployer: platform,
		networkMode:      c.String("network"),
		dryRun:           c.Bool("dry-run"),
	}

	return deployer.Deploy(c.Context)
}

// Deploy executes the aggregator deployment
func (d *AggregatorDeployer) Deploy(ctx context.Context) error {
	// Step 1: Fetch runtime specification
	spec, err := d.FetchRuntimeSpec(ctx)
	if err != nil {
		return err
	}

	// Step 2: Extract aggregator component
	component, err := d.ExtractComponent(spec, "aggregator")
	if err != nil {
		return err
	}

	// Step 3: Prepare environment configuration
	cfg := d.PrepareEnvironmentConfig()

	// Step 4: Validate configuration
	if err := ValidateAggregator(cfg.FinalEnvMap); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Step 5: Validate keystore
	keystoreName := cfg.FinalEnvMap["KEYSTORE_NAME"]
	if _, err := d.ValidateKeystore(keystoreName); err != nil {
		return err
	}

	// Step 6: Handle dry-run
	if d.dryRun {
		return d.handleDryRun(cfg, component.Registry, component.Digest)
	}

	// Step 7: Create temp directories
	tempCfg, err := d.CreateTempDirectories("aggregator")
	if err != nil {
		return err
	}
	// Merge configs
	cfg.TempDir = tempCfg.TempDir
	cfg.ConfigDir = tempCfg.ConfigDir
	cfg.KeystoreDir = tempCfg.KeystoreDir

	// Step 8: Generate configuration files
	if err := d.generateConfiguration(cfg); err != nil {
		return err
	}

	// Step 9: Deploy container
	return d.deployContainer(component, cfg)
}

// handleDryRun handles the dry-run scenario
func (d *AggregatorDeployer) handleDryRun(cfg *DeploymentConfig, registry string, digest string) error {
	d.Log.Info("✅ Dry run successful - aggregator configuration is valid")

	// Display configuration summary
	d.Log.Info("Configuration:",
		zap.String("keystoreName", cfg.FinalEnvMap["KEYSTORE_NAME"]),
		zap.String("operatorAddress", cfg.FinalEnvMap["OPERATOR_ADDRESS"]),
		zap.String("avsAddress", cfg.FinalEnvMap["AVS_ADDRESS"]),
		zap.String("registry", registry),
		zap.String("digest", digest),
		zap.String("network", d.networkMode))

	fmt.Println("\n✅ Configuration is valid. Run without --dry-run to deploy.")
	return nil
}

// generateConfiguration generates aggregator configuration files
func (d *AggregatorDeployer) generateConfiguration(cfg *DeploymentConfig) error {
	// Generate aggregator configuration using ConfigBuilder
	configBuilder := templates.NewConfigBuilder()
	aggregatorConfig, err := configBuilder.BuildAggregatorConfig(nil, cfg.FinalEnvMap)
	if err != nil {
		return fmt.Errorf("failed to build aggregator config: %w", err)
	}

	// Write aggregator config
	cfg.ConfigPath = filepath.Join(cfg.ConfigDir, "aggregator.yaml")
	if err := os.WriteFile(cfg.ConfigPath, aggregatorConfig, 0600); err != nil {
		return fmt.Errorf("failed to write aggregator config: %w", err)
	}

	d.Log.Info("Configuration written to", zap.String("path", cfg.ConfigPath))

	// Verify config file
	if stat, err := os.Stat(cfg.ConfigPath); err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	} else {
		d.Log.Info("Config file verified", zap.Int64("size", stat.Size()))
	}

	return nil
}

// deployContainer handles the aggregator-specific container deployment
func (d *AggregatorDeployer) deployContainer(component *runtime.ComponentSpec, cfg *DeploymentConfig) error {
	// Substitute environment variables in component
	runtime.SubstituteComponentEnv(component, cfg.EnvMap)

	// Prepare container configuration
	containerName := fmt.Sprintf("hgctl-aggregator-%s", cfg.EnvMap["AVS_ADDRESS"])
	imageRef := fmt.Sprintf("%s@%s", component.Registry, component.Digest)

	// Pull the image
	if err := d.PullDockerImage(imageRef, "aggregator"); err != nil {
		return err
	}

	// Stop and remove existing container
	d.CleanupExistingContainer(containerName)

	// Build docker arguments
	dockerArgs := d.BuildDockerArgs(containerName, component, cfg)

	// Add aggregator-specific options
	if d.networkMode != "" {
		// Insert network mode after "run" but before other flags
		dockerArgs = append(dockerArgs[:2], append([]string{"--network", d.networkMode}, dockerArgs[2:]...)...)
	}

	// Mount keystores
	if err := d.MountKeystores(&dockerArgs, cfg.FinalEnvMap["KEYSTORE_NAME"]); err != nil {
		return err
	}

	// Inject file contents as environment variables (for legacy support)
	dockerArgs = d.InjectFileContentsAsEnvVars(dockerArgs)

	// Add image
	dockerArgs = append(dockerArgs, imageRef)

	// Override command to use our config file
	dockerArgs = append(dockerArgs, "aggregator", "run", "--config", "/config/aggregator.yaml")

	// Run the container
	containerID, err := d.RunDockerContainer(dockerArgs, "aggregator")
	if err != nil {
		return err
	}

	// Print success message
	d.printSuccessMessage(containerName, containerID, cfg)
	return nil
}

// printSuccessMessage prints a user-friendly success message
func (d *AggregatorDeployer) printSuccessMessage(containerName, containerID string, cfg *DeploymentConfig) {
	d.Log.Info("✅ Aggregator deployed successfully",
		zap.String("container", containerName),
		zap.String("containerID", containerID),
		zap.String("config", cfg.ConfigDir),
		zap.String("tempDir", cfg.TempDir))

	fmt.Printf("\n✅ Aggregator deployed successfully\n")
	fmt.Printf("Container Name: %s\n", containerName)
	fmt.Printf("Container ID: %s\n", containerID[:12])
	fmt.Printf("Network Mode: %s\n", d.networkMode)
	fmt.Printf("Config Path: %s\n", cfg.ConfigPath)
	fmt.Printf("\nUseful commands:\n")
	fmt.Printf("  View logs:    docker logs -f %s\n", containerName)
	fmt.Printf("  Stop:         docker stop %s\n", containerName)
	fmt.Printf("  Restart:      docker restart %s\n", containerName)
	fmt.Printf("  Inspect:      docker inspect %s\n", containerName)
}
