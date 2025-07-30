package deploy

import (
	"bytes"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/templates"
)

func aggregatorCommand() *cli.Command {
	return &cli.Command{
		Name:      "aggregator",
		Usage:     "Deploy the aggregator component",
		ArgsUsage: "<avs-address>",
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:  "operator-set-id",
				Usage: "Operator set ID",
				Value: 0,
			},
			&cli.StringFlag{
				Name:  "version",
				Usage: "Release version (defaults to latest)",
			},
			&cli.StringSliceFlag{
				Name:  "env",
				Usage: "Set environment variables (can be used multiple times)",
			},
			&cli.StringFlag{
				Name:  "env-file",
				Usage: "Load environment variables from file",
			},
		},
		Action: deployAggregatorAction,
	}
}

func deployAggregatorAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.ShowSubcommandHelp(c)
	}

	avsAddress := c.Args().Get(0)
	operatorSetID := uint32(c.Uint64("operator-set-id"))
	version := c.String("version")

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

	// Find aggregator component in spec
	aggregatorComponent, found := spec.Spec["aggregator"]
	if !found {
		return fmt.Errorf("aggregator component not found in runtime spec")
	}

	log.Info("Found aggregator component",
		zap.String("registry", aggregatorComponent.Registry),
		zap.String("digest", aggregatorComponent.Digest))

	// Prepare environment variables from multiple sources
	envMap := make(map[string]string)

	// Start with context environment vars
	if currentCtx.EnvironmentVars != nil {
		for k, v := range currentCtx.EnvironmentVars {
			envMap[k] = v
		}
	}

	// Add AVS address
	envMap["AVS_ADDRESS"] = avsAddress

	// Load from env file if specified
	if envFile := c.String("env-file"); envFile != "" {
		fileEnv, err := runtime.LoadEnvFile(envFile)
		if err != nil {
			return fmt.Errorf("failed to load env file: %w", err)
		}
		for k, v := range fileEnv {
			envMap[k] = v
		}
	}

	// Parse and add individual env flags
	for _, env := range c.StringSlice("env") {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		} else {
			log.Warn("Invalid env flag format", zap.String("env", env))
		}
	}

	// Substitute environment variables
	runtime.SubstituteComponentEnv(&aggregatorComponent, envMap)

	// Create temporary directory for config and keystores
	tmpDir, err := os.MkdirTemp("", "hgctl-aggregator-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	// Don't remove tmpDir immediately - it needs to persist while container runs
	// Log it so user can clean up if needed
	log.Info("Created temporary directory for config", zap.String("path", tmpDir))

	configDir := filepath.Join(tmpDir, "config")
	keystoreDir := filepath.Join(tmpDir, "keystores")

	// Determine context directory
	contextName := currentCtx.Name
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	contextDir := filepath.Join(homeDir, ".hgctl", contextName)

	// Load context-specific environment variables
	contextEnv, err := templates.LoadContextEnv(contextDir)
	if err != nil {
		log.Warn("Failed to load context environment", zap.Error(err))
	}

	// Merge all environment variables (context env, then envMap overrides)
	finalEnvMap := runtime.MergeEnvMaps(contextEnv, envMap)

	// Create config generator
	configGen := templates.NewConfigGenerator(contextDir, finalEnvMap)

	// Generate aggregator config
	log.Info("Generating aggregator configuration...")
	configPath := filepath.Join(configDir, "aggregator.yaml")
	if err := configGen.GenerateAggregatorConfig(configPath); err != nil {
		return fmt.Errorf("failed to generate aggregator config: %w", err)
	}

	// Log the generated config location for debugging
	log.Info("Generated aggregator config", zap.String("path", configPath))

	// Verify the file exists
	if stat, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("config file not found after generation: %w", err)
	} else {
		log.Info("Config file verified", zap.Int64("size", stat.Size()))
	}

	// If debug mode, also log the config content
	if configContent, err := os.ReadFile(configPath); err == nil {
		log.Debug("Generated config content", zap.String("content", string(configContent)))
	}

	// Copy keystore and certificate files
	log.Info("Copying keystore files...")
	if err := configGen.CopyKeystoreFiles(keystoreDir); err != nil {
		return fmt.Errorf("failed to copy keystore files: %w", err)
	}

	// Container name
	containerName := fmt.Sprintf("hgctl-aggregator-%s", avsAddress)

	// Construct image reference
	imageRef := fmt.Sprintf("%s@%s", aggregatorComponent.Registry, aggregatorComponent.Digest)

	log.Info("Pulling aggregator image...", zap.String("image", imageRef))

	pullCmd := exec.Command("docker", "pull", imageRef)
	var pullStdout, pullStderr bytes.Buffer
	pullCmd.Stdout = &pullStdout
	pullCmd.Stderr = &pullStderr

	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("failed to pull aggregator image: %w\nstderr: %s", err, pullStderr.String())
	}

	log.Info("Successfully pulled aggregator image", zap.String("output", pullStdout.String()))

	// Stop existing container if running
	log.Info("Stopping existing aggregator container if running...")
	stopCmd := exec.Command("docker", "stop", containerName)
	if err := stopCmd.Run(); err != nil {
		log.Debug("No existing container to stop", zap.Error(err))
	}

	// Remove existing container
	removeCmd := exec.Command("docker", "rm", "-f", containerName)
	if err := removeCmd.Run(); err != nil {
		log.Debug("No existing container to remove", zap.Error(err))
	}

	// Build docker run command
	dockerArgs := []string{"run", "-d"}
	dockerArgs = append(dockerArgs, "--name", containerName)
	dockerArgs = append(dockerArgs, "--restart", "unless-stopped")

	// Add volume mount for config only
	dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/config:ro", configDir))

	// Add environment variables
	for _, env := range aggregatorComponent.Env {
		dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", env.Name, env.Value))
	}
	
	// Inject keystore and certificate contents as environment variables
	dockerArgs = injectFileContentsAsEnvVars(dockerArgs, contextDir, log)

	// Add ports
	for _, port := range aggregatorComponent.Ports {
		dockerArgs = append(dockerArgs, "-p", fmt.Sprintf("%d:%d", port, port))
	}

	// Add image (use the corrected imageRef)
	dockerArgs = append(dockerArgs, imageRef)

	// Override command to use our config file
	dockerArgs = append(dockerArgs, "aggregator", "run", "--config", "/config/aggregator.yaml")

	log.Info("Starting aggregator container...",
		zap.String("container", containerName),
		zap.String("image", imageRef))

	// Run docker command
	runCmd := exec.CommandContext(c.Context, "docker", dockerArgs...)
	var stdout, stderr bytes.Buffer
	runCmd.Stdout = &stdout
	runCmd.Stderr = &stderr

	if err := runCmd.Run(); err != nil {
		return fmt.Errorf("failed to start aggregator container: %w\nstderr: %s", err, stderr.String())
	}

	containerID := strings.TrimSpace(stdout.String())
	log.Info("Aggregator deployed successfully",
		zap.String("container", containerName),
		zap.String("containerID", containerID))

	fmt.Printf("Aggregator deployed successfully\n")
	fmt.Printf("Container Name: %s\n", containerName)
	fmt.Printf("Container ID: %s\n", containerID)
	fmt.Printf("Config directory: %s\n", configDir)
	fmt.Printf("To check logs: docker logs -f %s\n", containerName)
	fmt.Printf("To inspect mounts: docker inspect %s | grep -A20 Mounts\n", containerName)

	return nil
}

