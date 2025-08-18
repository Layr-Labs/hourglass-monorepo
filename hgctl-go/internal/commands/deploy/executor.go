package deploy

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/templates"
)

func executorCommand() *cli.Command {
	return &cli.Command{
		Name:      "executor",
		Usage:     "Deploy the executor component",
		ArgsUsage: "",
		Description: `Deploy the executor component from a release.

The AVS address must be configured in the context before running this command.`,
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:  "operator-set-id",
				Usage: "Operator set ID",
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
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Validate configuration without starting the container",
			},
		},
		Action: deployExecutorAction,
	}
}

// ExecutorDeployer handles executor-specific deployment logic
type ExecutorDeployer struct {
	*PlatformDeployer
	dryRun bool
}

func deployExecutorAction(c *cli.Context) error {
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	log := config.LoggerFromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	if currentCtx.AVSAddress == "" {
		return fmt.Errorf("AVS address not configured. Run 'hgctl context set --avs-address <address>' first")
	}

	opSetId := uint32(c.Uint64("operator-set-id"))
	if opSetId == 0 {
		opSetId = currentCtx.OperatorSetID
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
		opSetId,
		c.String("release-id"),
		c.String("env-file"),
		c.StringSlice("env"),
	)

	deployer := &ExecutorDeployer{
		PlatformDeployer: platform,
		dryRun:           c.Bool("dry-run"),
	}

	return deployer.Deploy(c.Context)
}

// Deploy executes the executor deployment
func (d *ExecutorDeployer) Deploy(ctx context.Context) error {
	spec, err := d.FetchRuntimeSpec(ctx)
	if err != nil {
		return err
	}

	component, err := d.ExtractComponent(spec, "executor")
	if err != nil {
		return err
	}

	cfg := &DeploymentConfig{
		Env: d.LoadEnvironmentVariables(),
	}

	if err := ValidateComponentSpec(component, cfg.Env); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	keystoreName := cfg.Env[config.KeystoreName]
	keystorePassword := cfg.Env[config.KeystorePassword]
	var keystore *signer.KeystoreReference
	if keystore, err = d.ValidateKeystore(keystoreName, keystorePassword); err != nil {
		return err
	}

	if d.dryRun {
		return d.handleDryRun(cfg, component.Registry, component.Digest)
	}

	tempCfg, err := d.CreateTempDirectories("executor")
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

	return d.deployContainer(component, keystore, cfg)
}

// generateConfiguration generates executor configuration files
func (d *ExecutorDeployer) generateConfiguration(cfg *DeploymentConfig) error {
	executorConfig, err := templates.BuildExecutorConfig(cfg.Env)
	if err != nil {
		return fmt.Errorf("failed to build executor config: %w", err)
	}

	cfg.ConfigPath = filepath.Join(cfg.ConfigDir, "executor.yaml")
	if err := os.WriteFile(cfg.ConfigPath, executorConfig, 0600); err != nil {
		return fmt.Errorf("failed to write executor config: %w", err)
	}

	d.Log.Info("Configuration written to", zap.String("path", cfg.ConfigPath))

	if stat, err := os.Stat(cfg.ConfigPath); err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	} else {
		d.Log.Info("Config file verified", zap.Int64("size", stat.Size()))
	}

	return nil
}

// deployContainer handles the executor-specific container deployment
func (d *ExecutorDeployer) deployContainer(
	component *runtime.ComponentSpec,
	keystore *signer.KeystoreReference,
	cfg *DeploymentConfig,
) error {
	containerName := fmt.Sprintf("hgctl-executor-%s-%s", d.Context.Name, d.Context.AVSAddress)
	imageRef := fmt.Sprintf("%s@%s", component.Registry, component.Digest)

	if err := d.PullDockerImage(imageRef, "executor"); err != nil {
		return err
	}

	d.CleanupExistingContainer(containerName)

	dockerArgs := d.BuildDockerArgs(containerName, component, cfg)

	// Add port mappings for executor service and management ports
	servicePort := cfg.Env["EXECUTOR_SERVICE_PORT"]
	if servicePort == "" {
		servicePort = "9090"
	}
	mgmtPort := cfg.Env["EXECUTOR_MGMT_PORT"]
	if mgmtPort == "" {
		mgmtPort = "9091"
	}
	dockerArgs = append(dockerArgs, "-p", fmt.Sprintf("%s:%s", servicePort, servicePort))
	dockerArgs = append(dockerArgs, "-p", fmt.Sprintf("%s:%s", mgmtPort, mgmtPort))
	d.Log.Info("Exposing executor ports", 
		zap.String("servicePort", servicePort),
		zap.String("managementPort", mgmtPort))

	if err := d.MountKeystores(&dockerArgs, keystore); err != nil {
		return err
	}

	dockerArgs = d.InjectFileContentsAsEnvVars(dockerArgs)

	dockerArgs = append(dockerArgs, "-e", "EXECUTOR_CONFIG_PATH=/config/executor.yaml")

	dockerArgs = append(dockerArgs, imageRef)

	dockerArgs = append(dockerArgs, "executor", "run", "--config", "/config/executor.yaml")

	containerID, err := d.RunDockerContainer(dockerArgs, "executor")
	if err != nil {
		return err
	}

	d.printSuccessMessage(containerName, containerID, cfg)
	
	// Save the executor address to the context
	if err := d.saveExecutorAddress(cfg); err != nil {
		d.Log.Warn("Failed to save executor address to context", zap.Error(err))
		// Don't fail the deployment, just warn
	}
	
	return nil
}

// handleDryRun handles the dry-run scenario
func (d *ExecutorDeployer) handleDryRun(cfg *DeploymentConfig, registry string, digest string) error {
	d.Log.Info("✅ Dry run successful - executor configuration is valid")

	d.Log.Info("Configuration:",
		zap.String("keystoreName", cfg.Env["KEYSTORE_NAME"]),
		zap.String("operatorAddress", d.Context.OperatorAddress),
		zap.String("avsAddress", d.Context.AVSAddress),
		zap.String("registry", registry),
		zap.String("digest", digest))

	fmt.Println("\n✅ Configuration is valid. Run without --dry-run to deploy.")
	return nil
}

// saveExecutorAddress saves the executor management gRPC address to the context configuration
func (d *ExecutorDeployer) saveExecutorAddress(cfg *DeploymentConfig) error {
	// Get the executor management port from config
	executorMgmtPort := cfg.Env["EXECUTOR_MGMT_PORT"]
	if executorMgmtPort == "" {
		executorMgmtPort = "9091"
	}
	
	// Build the executor address (using localhost since it's exposed on the host)
	executorAddress := fmt.Sprintf("localhost:%s", executorMgmtPort)
	
	// Load current config
	configData, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Update the executor address in the current context
	if ctx, exists := configData.Contexts[configData.CurrentContext]; exists {
		ctx.ExecutorAddress = executorAddress
		
		// Save the updated config
		if err := config.SaveConfig(configData); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		
		d.Log.Info("Saved executor management address to context", 
			zap.String("address", executorAddress),
			zap.String("context", configData.CurrentContext))
	}
	
	return nil
}

// printSuccessMessage prints a user-friendly success message
func (d *ExecutorDeployer) printSuccessMessage(containerName, containerID string, cfg *DeploymentConfig) {
	d.Log.Info("✅ Executor deployed successfully",
		zap.String("container", containerName),
		zap.String("containerID", containerID),
		zap.String("config", cfg.ConfigDir),
		zap.String("tempDir", cfg.TempDir))

	// Get the management port for display
	mgmtPort := cfg.Env["EXECUTOR_MGMT_PORT"]
	if mgmtPort == "" {
		mgmtPort = "9091"
	}
	
	fmt.Printf("\n✅ Executor deployed successfully\n")
	fmt.Printf("Container Name: %s\n", containerName)
	fmt.Printf("Container ID: %s\n", containerID[:12])
	fmt.Printf("Management Port: %s\n", mgmtPort)
	fmt.Printf("Config Path: %s\n", cfg.ConfigPath)
	fmt.Printf("\nThe executor management address (localhost:%s) has been saved to the context.\n", mgmtPort)
	fmt.Printf("\nUseful commands:\n")
	fmt.Printf("  View logs:    docker logs -f %s\n", containerName)
	fmt.Printf("  Stop:         docker stop %s\n", containerName)
	fmt.Printf("  Restart:      docker restart %s\n", containerName)
	fmt.Printf("  Inspect:      docker inspect %s\n", containerName)
}
