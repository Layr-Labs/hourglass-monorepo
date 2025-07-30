package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/templates"
)

func executorCommand() *cli.Command {
	return &cli.Command{
		Name:      "executor",
		Usage:     "Deploy the executor component",
		ArgsUsage: "<registry> <digest>",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "env",
				Usage: "Set environment variables (can be used multiple times)",
			},
			&cli.StringFlag{
				Name:  "env-file",
				Usage: "Load environment variables from file",
			},
			&cli.StringFlag{
				Name:  "bls-password",
				Usage: "Password for BLS keystore (or use BLS_PASSWORD env var)",
			},
			&cli.StringFlag{
				Name:  "ecdsa-password",
				Usage: "Password for ECDSA keystore (or use ECDSA_PASSWORD env var)",
			},
			&cli.BoolFlag{
				Name:  "password-stdin",
				Usage: "Read passwords from stdin (format: BLS_PASSWORD=xxx\\nECDSA_PASSWORD=xxx)",
			},
		},
		Action: deployExecutorAction,
	}
}

func deployExecutorAction(c *cli.Context) error {
	if c.NArg() != 2 {
		return cli.ShowSubcommandHelp(c)
	}

	registryURL := c.Args().Get(0)
	digest := c.Args().Get(1)

	// Get context
	currentCtx := c.Context.Value("currentContext").(*config.Context)
	log := logger.FromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	log.Info("Deploying executor",
		zap.String("registry", registryURL),
		zap.String("digest", digest))

	// Set up password provider
	var passwordProvider signer.PasswordProvider
	if c.Bool("password-stdin") {
		stdinProvider := signer.NewStdinPasswordProvider()
		if err := stdinProvider.ReadPasswordsFromStdin(); err != nil {
			return fmt.Errorf("failed to read passwords from stdin: %w", err)
		}
		passwordProvider = stdinProvider
	} else {
		// Combine explicit flags, environment, and interactive
		providers := []signer.PasswordProvider{}

		// Add explicit password flags
		flagPasswords := make(map[string]string)
		if blsPwd := c.String("bls-password"); blsPwd != "" {
			flagPasswords["BLS_PASSWORD"] = blsPwd
		}
		if ecdsaPwd := c.String("ecdsa-password"); ecdsaPwd != "" {
			flagPasswords["ECDSA_PASSWORD"] = ecdsaPwd
		}
		if len(flagPasswords) > 0 {
			providers = append(providers, signer.NewMapPasswordProvider(flagPasswords))
		}

		// Add environment provider
		providers = append(providers, signer.NewEnvironmentPasswordProvider())

		// Add interactive provider as last resort
		providers = append(providers, signer.NewInteractivePasswordProvider())

		passwordProvider = signer.NewCombinedPasswordProvider(providers...)
	}

	// Create signer resolver
	signerResolver := signer.NewDefaultSignerResolver(passwordProvider)

	// Resolve signer configurations
	blsSigner, err := signerResolver.ResolveSignerConfig(currentCtx, "BLS", "")
	if err != nil {
		log.Warn("No BLS signer configuration found", zap.Error(err))
	}

	ecdsaSigner, err := signerResolver.ResolveSignerConfig(currentCtx, "ECDSA", "")
	if err != nil {
		log.Warn("No ECDSA signer configuration found", zap.Error(err))
	}

	// Pull the runtime spec from OCI registry
	log.Info("Pulling runtime spec from OCI registry...")
	ociClient := client.NewOCIClient(log)

	// Clean up the digest (remove 0x prefix if present)
	digestClean := strings.TrimPrefix(digest, "0x")
	if !strings.HasPrefix(digestClean, "sha256:") {
		digestClean = "sha256:" + digestClean
	}

	spec, err := ociClient.PullRuntimeSpec(c.Context, registryURL, digestClean)
	if err != nil {
		return fmt.Errorf("failed to pull runtime spec: %w", err)
	}

	// Find executor component in spec
	executorComponent, found := spec.Spec["executor"]
	if !found {
		return fmt.Errorf("executor component not found in runtime spec")
	}

	// Prepare environment variables
	envVars := loadEnvironmentVariables(c, currentCtx)

	// Prepare signer configs map
	signerConfigs := make(map[string]*signer.SignerConfig)
	if blsSigner != nil {
		signerConfigs["BLS"] = blsSigner
	}
	if ecdsaSigner != nil {
		signerConfigs["ECDSA"] = ecdsaSigner
	}

	// Generate executor configuration
	configBuilder := templates.NewConfigBuilder()
	executorConfig, err := configBuilder.BuildExecutorConfig(signerConfigs, envVars)
	if err != nil {
		return fmt.Errorf("failed to build executor config: %w", err)
	}

	// Create temporary directory for config and keystores
	tempDir, err := os.MkdirTemp("", "hgctl-executor-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	// Note: Not using defer to remove tempDir - let user clean up after inspection

	// Create subdirectories
	configDir := filepath.Join(tempDir, "config")
	keystoreDir := filepath.Join(tempDir, "keystores")
	
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := os.MkdirAll(keystoreDir, 0755); err != nil {
		return fmt.Errorf("failed to create keystore directory: %w", err)
	}

	// Write executor config
	configPath := filepath.Join(configDir, "executor.yaml")
	if err := os.WriteFile(configPath, executorConfig, 0600); err != nil {
		return fmt.Errorf("failed to write executor config: %w", err)
	}
	
	// Copy keystore files
	contextName := currentCtx.Name
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	contextDir := filepath.Join(homeDir, ".hgctl", contextName)
	
	// Use the existing template package to copy keystore files
	configGen := templates.NewConfigGenerator(contextDir, envVars)
	if err := configGen.CopyKeystoreFiles(keystoreDir); err != nil {
		log.Warn("Failed to copy keystore files", zap.Error(err))
		// Don't fail the deployment if keystores aren't found
	}

	log.Info("Configuration written to", zap.String("path", configPath))

	// Container name
	containerName := fmt.Sprintf("hgctl-executor-%s", digest[:12])

	// Build docker run command
	dockerArgs := []string{"run", "-d", "--name", containerName}

	// Add volume mount for config only
	dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/config:ro", configDir))

	// Add environment variables
	for _, env := range executorComponent.Env {
		dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", env.Name, env.Value))
	}
	
	// Inject keystore and certificate contents as environment variables
	dockerArgs = injectFileContentsAsEnvVars(dockerArgs, contextDir, log)

	// Override config path
	dockerArgs = append(dockerArgs, "-e", "EXECUTOR_CONFIG_PATH=/config/executor.yaml")

	// Add ports
	for _, port := range executorComponent.Ports {
		dockerArgs = append(dockerArgs, "-p", fmt.Sprintf("%d:%d", port, port))
	}

	// Add image
	imageRef := fmt.Sprintf("%s@%s", executorComponent.Registry, executorComponent.Digest)
	dockerArgs = append(dockerArgs, imageRef)

	// Add command if specified
	if len(executorComponent.Command) > 0 {
		dockerArgs = append(dockerArgs, executorComponent.Command...)
	}

	// Execute docker command
	log.Info("Starting executor container...", zap.Strings("args", dockerArgs))

	cmd := exec.Command("docker", dockerArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start executor: %w", err)
	}

	log.Info("✅ Executor deployed successfully",
		zap.String("container", containerName),
		zap.String("config", configDir),
		zap.String("keystores", keystoreDir),
		zap.String("tempDir", tempDir))

	return nil
}

func loadEnvironmentVariables(c *cli.Context, ctx *config.Context) map[string]string {
	envVars := make(map[string]string)

	// Start with context environment variables
	for k, v := range ctx.EnvironmentVars {
		envVars[k] = v
	}

	// Load from env file if specified
	if envFile := c.String("env-file"); envFile != "" {
		if data, err := os.ReadFile(envFile); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					envVars[parts[0]] = parts[1]
				}
			}
		}
	}

	// Apply command-line env overrides
	for _, env := range c.StringSlice("env") {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	// Add system environment variables as fallback
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			if _, exists := envVars[parts[0]]; !exists {
				envVars[parts[0]] = parts[1]
			}
		}
	}

	return envVars
}

