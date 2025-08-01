package deploy

import (
	"bytes"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
			&cli.Uint64Flag{
				Name:  "operator-set-id",
				Usage: "Operator set ID",
				Value: 0,
			},
			&cli.StringFlag{
				Name:  "release-id",
				Usage: "Release ID to deploy (defaults to latest)",
			},
			&cli.StringFlag{
				Name:  "registry",
				Usage: "OCI registry URL (use with --digest to bypass release lookup)",
			},
			&cli.StringFlag{
				Name:  "digest",
				Usage: "Artifact digest (use with --registry to bypass release lookup)",
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
		},
		Action: deployAggregatorAction,
	}
}

func deployAggregatorAction(c *cli.Context) error {
	if c.NArg() != 0 {
		return cli.ShowSubcommandHelp(c)
	}

	operatorSetID := uint32(c.Uint64("operator-set-id"))
	releaseID := c.String("release-id")
	registry := c.String("registry")
	digest := c.String("digest")

	// Get logger and context from middleware
	log := middleware.GetLogger(c)
	currentCtx := c.Context.Value("currentContext").(*config.Context)
	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	// Validate flag combinations
	if (registry != "" && digest == "") || (registry == "" && digest != "") {
		return fmt.Errorf("--registry and --digest must be used together")
	}

	var spec *runtime.Spec
	var artifactRegistry, artifactDigest string

	if registry != "" && digest != "" {
		// Direct registry/digest mode
		log.Info("Using direct registry and digest",
			zap.String("registry", registry),
			zap.String("digest", digest))

		artifactRegistry = registry
		artifactDigest = digest

		// Create OCI client
		ociClient := client.NewOCIClient(log)

		// Pull runtime spec directly
		log.Info("Pulling runtime spec from OCI registry...")
		var err error
		spec, err = ociClient.PullRuntimeSpec(c.Context, artifactRegistry, artifactDigest)
		if err != nil {
			return fmt.Errorf("failed to pull runtime spec: %w", err)
		}
	} else {
		// Release-based mode
		// Get contract client from middleware
		contractClient, err := middleware.GetContractClient(c)
		if err != nil {
			return fmt.Errorf("failed to get contract client: %w", err)
		}

		// Get AVS address from contract client
		avsAddress := contractClient.GetAVSAddress()

		log.Info("Fetching release from ReleaseManager",
			zap.String("avs", avsAddress.Hex()),
			zap.Uint32("operatorSetID", operatorSetID))

		var release *client.ReleaseManagerRelease

		if releaseID == "" {
			// Get latest release
			nextReleaseId, err := contractClient.GetNextReleaseId(c.Context, operatorSetID)
			if err != nil {
				return fmt.Errorf("failed to get next release ID: %w", err)
			}

			if nextReleaseId.Uint64() == 0 {
				return fmt.Errorf("no releases found for operator set %d", operatorSetID)
			}

			latestId := new(big.Int).Sub(nextReleaseId, big.NewInt(1))
			release, err = contractClient.GetRelease(c.Context, operatorSetID, latestId)
			if err != nil {
				return fmt.Errorf("failed to get latest release: %w", err)
			}

			log.Info("Using latest release", zap.String("releaseID", latestId.String()))
		} else {
			// Get specific release
			releaseIDBig := new(big.Int)
			releaseIDBig.SetString(releaseID, 10)

			release, err = contractClient.GetRelease(c.Context, operatorSetID, releaseIDBig)
			if err != nil {
				return fmt.Errorf("failed to get release %s: %w", releaseID, err)
			}

			log.Info("Using release", zap.String("releaseID", releaseID))
		}

		if len(release.Artifacts) == 0 {
			return fmt.Errorf("no artifacts found in release")
		}

		artifact := release.Artifacts[0]
		artifactRegistry = artifact.RegistryName
		artifactDigest = fmt.Sprintf("0x%x", artifact.Digest)

		log.Info("Found release artifact",
			zap.String("digest", artifactDigest),
			zap.String("registry", artifactRegistry))

		// Create OCI client
		ociClient := client.NewOCIClient(log)

		// Pull runtime spec
		log.Info("Pulling runtime spec from OCI registry...")
		spec, err = ociClient.PullRuntimeSpec(c.Context, artifactRegistry, artifactDigest)
		if err != nil {
			return fmt.Errorf("failed to pull runtime spec: %w", err)
		}
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
	var avsAddressStr string
	if registry != "" && digest != "" {
		// In direct mode, get AVS address from context or fail
		if currentCtx.AVSAddress == "" {
			return fmt.Errorf("AVS address not configured in context for direct deployment mode")
		}
		avsAddressStr = currentCtx.AVSAddress
	} else {
		// In release mode, we already have it from contractClient
		contractClient, _ := middleware.GetContractClient(c)
		avsAddressStr = contractClient.GetAVSAddress().Hex()
	}
	envMap["AVS_ADDRESS"] = avsAddressStr

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
	containerName := fmt.Sprintf("hgctl-aggregator-%s", avsAddressStr)

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

	// Add network mode if specified
	networkMode := c.String("network")
	if networkMode != "" {
		dockerArgs = append(dockerArgs, "--network", networkMode)
	}

	// Add volume mount for config
	dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/config:ro", configDir))

	// Mount registered keystores directly from their original locations
	if currentCtx.Keystores != nil {
		for _, ks := range currentCtx.Keystores {
			if ks.Type == "bn254" || ks.Type == "bls" {
				// Mount BLS keystore directly at the expected location
				dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/keystores/operator.bls.keystore.json:ro", ks.Path))
				log.Info("Mounting registered BLS keystore",
					zap.String("name", ks.Name),
					zap.String("source", ks.Path),
					zap.String("container_path", "/keystores/operator.bls.keystore.json"))
				break // Only mount the first BLS keystore
			}
		}
	} else {
		// Fall back to mounting the keystores directory if no registered keystores
		dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/keystores:ro", keystoreDir))
	}

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
