package deploy

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
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
	containerName := fmt.Sprintf("hgctl-executor-%s-%s", d.Context.Name, d.Context.AVSAddress)

	// Check if an executor container is already running
	isRunning, containerID, err := d.CheckContainerRunning(containerName)
	if err != nil {
		d.Log.Warn("Error checking container status, proceeding with deployment", zap.Error(err))
		// Continue with normal deployment
	}

	if isRunning {
		d.Log.Info("Found existing running executor container",
			zap.String("container", containerName),
			zap.String("containerID", containerID[:12]))

		// Deploy performer to the existing executor
		if err = d.deployPerformerToExistingExecutor(ctx, containerName, containerID); err != nil {
			return err
		}
		return nil
	}

	// No existing container, proceed with normal deployment
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

	// Validate that required passwords are set for system keystores
	if cfg.Env["SYSTEM_BN254_KEYSTORE_PATH"] != "" || cfg.Env["SYSTEM_ECDSA_KEYSTORE_PATH"] != "" {
		if cfg.Env["SYSTEM_KEYSTORE_PASSWORD"] == "" {
			return fmt.Errorf("SYSTEM_KEYSTORE_PASSWORD environment variable required for system keystores")
		}
	}

	keystoreName := cfg.Env[config.KeystoreName]
	keystorePassword := cfg.Env[config.KeystorePassword]
	var keystore *signer.KeystoreReference

	if keystoreName != "" {
		if keystore, err = d.ValidateKeystore(keystoreName, keystorePassword); err != nil {
			return err
		}
		d.Log.Info("Using keystore configuration",
			zap.String("keystoreName", keystoreName))
	} else {
		if cfg.Env["OPERATOR_PRIVATE_KEY"] == "" {
			return fmt.Errorf("neither keystore nor OPERATOR_PRIVATE_KEY is configured")
		}
		d.Log.Info("Using private key configuration")
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

// ensureDockerNetwork checks if a Docker network exists and creates it if it doesn't
func (d *ExecutorDeployer) ensureDockerNetwork(networkName string) error {
	// Check if network exists
	cmd := exec.Command("docker", "network", "inspect", networkName)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Network doesn't exist, create it
		d.Log.Info("Docker network not found, creating it", zap.String("network", networkName))

		createCmd := exec.Command("docker", "network", "create", networkName)
		var createStderr bytes.Buffer
		createCmd.Stderr = &createStderr

		if err := createCmd.Run(); err != nil {
			return fmt.Errorf("failed to create Docker network %s: %w\nstderr: %s", networkName, err, createStderr.String())
		}

		d.Log.Info("Docker network created successfully", zap.String("network", networkName))
	} else {
		d.Log.Info("Docker network already exists", zap.String("network", networkName))
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

	// Mount Docker socket so the executor can create performer containers
	dockerArgs = append(dockerArgs, "-v", "/var/run/docker.sock:/var/run/docker.sock")
	d.Log.Info("Mounting Docker socket for performer container management")

	// Ensure the hourglass-local network exists and connect to it
	networkName := "hourglass-local_hourglass-network"
	if err := d.ensureDockerNetwork(networkName); err != nil {
		return fmt.Errorf("failed to ensure Docker network: %w", err)
	}
	dockerArgs = append(dockerArgs, "--network", networkName)
	d.Log.Info("Connecting to Docker network", zap.String("network", networkName))

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

	if keystore != nil {
		if err := d.MountKeystores(&dockerArgs, keystore); err != nil {
			return err
		}
	}

	// Mount system keystores if configured
	d.MountSystemKeystores(&dockerArgs, cfg)

	dockerArgs = d.InjectFileContentsAsEnvVars(dockerArgs)

	dockerArgs = append(dockerArgs, "-e", "EXECUTOR_CONFIG_PATH=/config/executor.yaml")

	dockerArgs = append(dockerArgs, imageRef)

	dockerArgs = append(dockerArgs, "executor", "run", "--config", "/config/executor.yaml")

	containerID, err := d.RunDockerContainer(dockerArgs, "executor")
	if err != nil {
		return err
	}

	d.printSuccessMessage(containerName, containerID, cfg)

	if err := d.saveExecutorAddress(cfg); err != nil {
		d.Log.Warn("Failed to save executor address to context", zap.Error(err))
	}

	return nil
}

// handleDryRun handles the dry-run scenario
func (d *ExecutorDeployer) handleDryRun(cfg *DeploymentConfig, registry string, digest string) error {
	d.Log.Info("✅ Dry run successful - executor configuration is valid")

	configInfo := []zap.Field{
		zap.String("operatorAddress", d.Context.OperatorAddress),
		zap.String("avsAddress", d.Context.AVSAddress),
		zap.String("registry", registry),
		zap.String("digest", digest),
	}
	if cfg.Env["KEYSTORE_NAME"] != "" {
		configInfo = append(configInfo, zap.String("keystoreName", cfg.Env["KEYSTORE_NAME"]))
	} else {
		configInfo = append(configInfo, zap.String("signerType", "privateKey"))
	}
	d.Log.Info("Configuration:", configInfo...)

	fmt.Println("\n✅ Configuration is valid. Run without --dry-run to deploy.")
	return nil
}

// saveExecutorAddress saves the executor management gRPC address to the context configuration
func (d *ExecutorDeployer) saveExecutorAddress(cfg *DeploymentConfig) error {
	executorMgmtPort := cfg.Env["EXECUTOR_MGMT_PORT"]
	if executorMgmtPort == "" {
		executorMgmtPort = "9091"
	}

	executorEndpoint := fmt.Sprintf("localhost:%s", executorMgmtPort)

	configData, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if ctx, exists := configData.Contexts[configData.CurrentContext]; exists {
		ctx.ExecutorEndpoint = executorEndpoint

		if err := config.SaveConfig(configData); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		d.Log.Info("Saved executor management address to context",
			zap.String("address", executorEndpoint),
			zap.String("context", configData.CurrentContext))
	}

	return nil
}

// deployPerformerToExistingExecutor deploys a performer to an existing executor container via gRPC
func (d *ExecutorDeployer) deployPerformerToExistingExecutor(ctx context.Context, containerName string, containerID string) error {
	// Try to get the management port from the running container
	mgmtPort, err := d.GetContainerPort(containerName, "9091")
	if err != nil {
		// Try default port if we can't get it from Docker
		mgmtPort = "9091"
		d.Log.Warn("Could not get management port from container, using default",
			zap.String("port", mgmtPort),
			zap.Error(err))
	}

	// Create executor client
	executorEndpoint := fmt.Sprintf("localhost:%s", mgmtPort)
	executorClient, err := client.NewExecutorClient(executorEndpoint, d.Log)
	if err != nil {
		return fmt.Errorf("failed to create executor client: %w", err)
	}
	defer executorClient.Close()

	// Fetch the performer spec from the runtime spec
	spec, err := d.FetchRuntimeSpec(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch runtime spec: %w", err)
	}

	performerComponent, err := d.ExtractComponent(spec, "performer")
	if err != nil {
		return fmt.Errorf("failed to extract performer component: %w", err)
	}

	// Load environment variables for the performer
	envVars := d.LoadEnvironmentVariables()

	// Deploy the performer via gRPC
	d.Log.Info("Deploying performer to existing executor via gRPC",
		zap.String("executor", executorEndpoint),
		zap.String("avsAddress", d.Context.AVSAddress),
		zap.String("image", performerComponent.Registry),
		zap.String("digest", performerComponent.Digest))

	deploymentID, err := executorClient.DeployPerformerWithEnv(
		ctx,
		d.Context.AVSAddress,
		performerComponent.Digest,
		performerComponent.Registry,
		envVars,
	)
	if err != nil {
		return fmt.Errorf("failed to deploy performer via gRPC: %w", err)
	}

	// Update executor address in context if needed
	if d.Context.ExecutorEndpoint != executorEndpoint {
		configData, err := config.LoadConfig()
		if err == nil {
			if ctx, exists := configData.Contexts[configData.CurrentContext]; exists {
				ctx.ExecutorEndpoint = executorEndpoint
				if err := config.SaveConfig(configData); err == nil {
					d.Log.Info("Updated executor address in context",
						zap.String("address", executorEndpoint))
				}
			}
		}
	}

	// Success - print confirmation
	d.Log.Info("Successfully deployed performer to existing executor")
	fmt.Printf("\n✅ Performer deployed to existing executor via gRPC\n")
	fmt.Printf("Container Name: %s\n", containerName)
	fmt.Printf("Container ID: %s\n", containerID[:12])
	fmt.Printf("Management Port: %s\n", mgmtPort)
	fmt.Printf("Deployment ID: %s\n", deploymentID)
	fmt.Printf("AVS Address: %s\n", d.Context.AVSAddress)
	fmt.Printf("Performer Image: %s@%s\n", performerComponent.Registry, performerComponent.Digest)
	fmt.Printf("\nThe performer is now running on the executor.\n")
	fmt.Printf("\nUseful commands:\n")
	fmt.Printf("  View executor logs:  docker logs -f %s\n", containerName)
	fmt.Printf("  List performers:     hgctl get performer\n")
	fmt.Printf("  Remove performer:    hgctl remove performer %s\n", deploymentID)

	return nil
}

// printSuccessMessage prints a user-friendly success message
func (d *ExecutorDeployer) printSuccessMessage(containerName, containerID string, cfg *DeploymentConfig) {
	d.Log.Info("✅ Executor deployed successfully",
		zap.String("container", containerName),
		zap.String("containerID", containerID),
		zap.String("config", cfg.ConfigDir),
		zap.String("tempDir", cfg.TempDir))

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
