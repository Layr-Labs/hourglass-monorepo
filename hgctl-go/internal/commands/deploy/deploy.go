package deploy

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/dao"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
)

// Command returns the deploy command
func Command() *cli.Command {
	return &cli.Command{
		Name:  "deploy",
		Usage: "Deploy hourglass components",
		Subcommands: []*cli.Command{
			performerCommand(),
			executorCommand(),
			aggregatorCommand(),
		},
	}
}

// PlatformDeployer provides shared deployment functionality for all components
type PlatformDeployer struct {
	Context        *config.Context
	Log            logger.Logger
	AVSAddress     string
	OperatorSetID  uint32
	ReleaseID      string
	ContractClient *client.ContractClient
	EnvFile        string
	EnvFlags       []string
}

// DeploymentConfig holds deployment configuration
type DeploymentConfig struct {
	TempDir     string
	ConfigDir   string
	KeystoreDir string
	ConfigPath  string
	Env         map[string]string
}

// DeploymentArtifact represents a deployment artifact
type DeploymentArtifact struct {
	Registry string
	Digest   string
}

// NewPlatformDeployer creates a new platform deployer instance
func NewPlatformDeployer(
	ctx *config.Context,
	log logger.Logger,
	contractClient *client.ContractClient,
	avsAddress string,
	operatorSetID uint32,
	releaseID string,
	envFile string,
	envFlags []string,
) *PlatformDeployer {
	return &PlatformDeployer{
		Context:        ctx,
		Log:            log,
		AVSAddress:     avsAddress,
		OperatorSetID:  operatorSetID,
		ReleaseID:      releaseID,
		ContractClient: contractClient,
		EnvFile:        envFile,
		EnvFlags:       envFlags,
	}
}

// FetchRuntimeSpec retrieves the runtime specification via release manager
func (d *PlatformDeployer) FetchRuntimeSpec(ctx context.Context) (*runtime.Spec, error) {
	// Get AVS address from contract client
	d.Log.Info("Fetching release from ReleaseManager",
		zap.String("avs", d.AVSAddress),
		zap.Uint32("operatorSetID", d.OperatorSetID),
	)

	// Create OCI client and DAO
	ociClient := client.NewOCIClient(d.Log)
	specDAO := dao.NewEigenRuntimeSpecDAO(d.ContractClient, ociClient, d.OperatorSetID, d.Log)

	// Fetch runtime spec using DAO
	if d.ReleaseID == "" {
		return specDAO.GetLatestRuntimeSpec(ctx)
	}
	return specDAO.GetRuntimeSpec(ctx, d.ReleaseID)
}

// ExtractComponent extracts a component from the runtime spec
func (d *PlatformDeployer) ExtractComponent(spec *runtime.Spec, componentName string) (*runtime.ComponentSpec, error) {
	component, exists := spec.Spec[componentName]
	if !exists {
		return nil, fmt.Errorf("%s component not found in runtime spec", componentName)
	}

	d.Log.Info(fmt.Sprintf("Found %s component", componentName),
		zap.String("registry", component.Registry),
		zap.String("digest", component.Digest),
	)

	return &component, nil
}

// LoadEnvironmentVariables loads environment variables from all sources with proper precedence
func (d *PlatformDeployer) LoadEnvironmentVariables() map[string]string {
	envVars := make(map[string]string)

	// Add operator address from context as default if not already set
	if d.Context.OperatorAddress != "" {
		if _, exists := envVars["OPERATOR_ADDRESS"]; !exists {
			envVars["OPERATOR_ADDRESS"] = d.Context.OperatorAddress
			d.Log.Debug("Using operator address from context", zap.String("address", d.Context.OperatorAddress))
		}
	}

	if operatorKey, err := config.GetOperatorPrivateKey(d.Context); err == nil {
		envVars["OPERATOR_PRIVATE_KEY"] = operatorKey
		d.Log.Debug("Loaded operator private key")
	} else {
		d.Log.Error("Operator private key not available", zap.Error(err))
	}

	// Populate SystemSignerKeys configuration into environment variables
	if d.Context.SystemSignerKeys != nil {
		// Handle BN254 keystore configuration
		if d.Context.SystemSignerKeys.BN254 != nil {
			// Find the actual keystore path
			for _, ks := range d.Context.Keystores {
				if ks.Name == d.Context.SystemSignerKeys.BN254.Name {
					envVars["SYSTEM_BN254_KEYSTORE_PATH"] = ks.Path
					d.Log.Debug("Using system BN254 keystore from context",
						zap.String("keystore", ks.Name),
						zap.String("path", ks.Path))
					break
				}
			}
		}

		// Handle ECDSA configuration
		if d.Context.SystemSignerKeys.ECDSA != nil {
			if d.Context.SystemSignerKeys.ECDSA.Keystore != nil {
				// Find the actual keystore path
				for _, ks := range d.Context.Keystores {
					if ks.Name == d.Context.SystemSignerKeys.ECDSA.Keystore.Name {
						envVars["SYSTEM_ECDSA_KEYSTORE_PATH"] = ks.Path
						d.Log.Debug("Using system ECDSA keystore from context",
							zap.String("keystore", ks.Name),
							zap.String("path", ks.Path))
						break
					}
				}
			} else if d.Context.SystemSignerKeys.ECDSA.RemoteSignerConfig != nil {
				envVars["SYSTEM_ECDSA_WEB3SIGNER_URL"] = d.Context.SystemSignerKeys.ECDSA.RemoteSignerConfig.Url
				// Load cert contents if paths are provided
				if d.Context.SystemSignerKeys.ECDSA.RemoteSignerConfig.CACertPath != "" {
					if certData, err := os.ReadFile(d.Context.SystemSignerKeys.ECDSA.RemoteSignerConfig.CACertPath); err == nil {
						envVars["SYSTEM_WEB3SIGNER_CA_CERT"] = string(certData)
					}
				}
				if d.Context.SystemSignerKeys.ECDSA.RemoteSignerConfig.ClientCertPath != "" {
					if certData, err := os.ReadFile(d.Context.SystemSignerKeys.ECDSA.RemoteSignerConfig.ClientCertPath); err == nil {
						envVars["SYSTEM_WEB3SIGNER_CLIENT_CERT"] = string(certData)
					}
				}
				if d.Context.SystemSignerKeys.ECDSA.RemoteSignerConfig.ClientKeyPath != "" {
					if keyData, err := os.ReadFile(d.Context.SystemSignerKeys.ECDSA.RemoteSignerConfig.ClientKeyPath); err == nil {
						envVars["SYSTEM_WEB3SIGNER_CLIENT_KEY"] = string(keyData)
					}
				}
				d.Log.Debug("Using system ECDSA web3signer from context", zap.String("url", d.Context.SystemSignerKeys.ECDSA.RemoteSignerConfig.Url))
			} else if d.Context.SystemSignerKeys.ECDSA.PrivateKey {
				// PrivateKey flag indicates to use SYSTEM_PRIVATE_KEY env var
				d.Log.Debug("System ECDSA configured to use SYSTEM_PRIVATE_KEY environment variable")
			}
		}
	}

	// Add L1 chain ID from context if not already set
	if d.Context.L1ChainID != 0 {
		if _, exists := envVars["L1_CHAIN_ID"]; !exists {
			envVars["L1_CHAIN_ID"] = fmt.Sprintf("%d", d.Context.L1ChainID)
			d.Log.Debug("Using L1 chain ID from context", zap.Uint32("chainId", d.Context.L1ChainID))
		}
	}

	// Add L1 RPC URL from context if not already set
	if d.Context.L1RPCUrl != "" {
		if _, exists := envVars["L1_RPC_URL"]; !exists {
			// Translate localhost URLs for Docker on macOS
			envVars["L1_RPC_URL"] = translateLocalhostForDocker(d.Context.L1RPCUrl)
			if envVars["L1_RPC_URL"] != d.Context.L1RPCUrl {
				d.Log.Debug("Translated L1 RPC URL for Docker",
					zap.String("original", d.Context.L1RPCUrl),
					zap.String("translated", envVars["L1_RPC_URL"]))
			} else {
				d.Log.Debug("Using L1 RPC URL from context", zap.String("rpcUrl", d.Context.L1RPCUrl))
			}
		}
	}

	// Add L2 chain ID from context if not already set
	if d.Context.L2ChainID != 0 {
		if _, exists := envVars["L2_CHAIN_ID"]; !exists {
			envVars["L2_CHAIN_ID"] = fmt.Sprintf("%d", d.Context.L2ChainID)
			d.Log.Debug("Using L2 chain ID from context", zap.Uint32("chainId", d.Context.L2ChainID))
		}
	}

	// Add L2 RPC URL from context if not already set
	if d.Context.L2RPCUrl != "" {
		if _, exists := envVars["L2_RPC_URL"]; !exists {
			// Translate localhost URLs for Docker on macOS
			envVars["L2_RPC_URL"] = translateLocalhostForDocker(d.Context.L2RPCUrl)
			if envVars["L2_RPC_URL"] != d.Context.L2RPCUrl {
				d.Log.Debug("Translated L2 RPC URL for Docker",
					zap.String("original", d.Context.L2RPCUrl),
					zap.String("translated", envVars["L2_RPC_URL"]))
			} else {
				d.Log.Debug("Using L2 RPC URL from context", zap.String("rpcUrl", d.Context.L2RPCUrl))
			}
		}
	}

	if d.EnvFile != "" {
		d.loadEnvFile(d.EnvFile, envVars)
	}

	for _, env := range d.EnvFlags {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			if _, exists := envVars[parts[0]]; !exists {
				envVars[parts[0]] = parts[1]
			}
		}
	}

	if d.Context.EnvSecretsPath != "" {
		secretsPath := d.expandPath(d.Context.EnvSecretsPath)
		d.Log.Info("Loading environment from secrets file")
		d.loadEnvFile(secretsPath, envVars)
	}

	rpcURLKeys := []string{"L1_RPC_URL", "L2_RPC_URL", "RPC_URL", "ETH_RPC_URL"}
	for _, key := range rpcURLKeys {
		if url, exists := envVars[key]; exists && url != "" {
			translated := translateLocalhostForDocker(url)
			if translated != url {
				envVars[key] = translated
				d.Log.Debug("Translated RPC URL for Docker",
					zap.String("key", key),
					zap.String("original", url),
					zap.String("translated", translated))
			}
		}
	}

	return envVars
}

// CreateTempDirectories creates temporary directories for configuration
func (d *PlatformDeployer) CreateTempDirectories(componentType string) (*DeploymentConfig, error) {
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("hgctl-%s-*", componentType))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	cfg := &DeploymentConfig{
		TempDir:     tempDir,
		ConfigDir:   filepath.Join(tempDir, "config"),
		KeystoreDir: filepath.Join(tempDir, "keystores"),
	}

	if err := os.MkdirAll(cfg.ConfigDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := os.MkdirAll(cfg.KeystoreDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create keystore directory: %w", err)
	}

	d.Log.Info("Created temporary directories",
		zap.String("tempDir", cfg.TempDir),
		zap.String("configDir", cfg.ConfigDir),
		zap.String("keystoreDir", cfg.KeystoreDir))

	return cfg, nil
}

// ValidateKeystore validates that a keystore exists and is accessible
func (d *PlatformDeployer) ValidateKeystore(keystoreName string, keystorePassword string) (*signer.KeystoreReference, error) {
	var missing []string

	// Check for keystore configuration
	if keystoreName == "" {
		missing = append(missing, "KEYSTORE_NAME")
	}
	if keystorePassword == "" {
		missing = append(missing, "KEYSTORE_PASSWORD")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required signer configuration:\n  - %s", strings.Join(missing, "\n  - "))
	}

	var foundKeystore *signer.KeystoreReference
	for _, ks := range d.Context.Keystores {
		if ks.Name == keystoreName {
			foundKeystore = &ks
			break
		}
	}

	if foundKeystore == nil {
		return nil, fmt.Errorf("keystore '%s' not found in context '%s'", keystoreName, d.Context.Name)
	}

	// Verify keystore file exists
	if _, err := os.Stat(foundKeystore.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("keystore file not found at path: %s", foundKeystore.Path)
	}

	d.Log.Info("Validated keystore",
		zap.String("name", foundKeystore.Name),
		zap.String("type", foundKeystore.Type),
		zap.String("path", foundKeystore.Path))

	return foundKeystore, nil
}

// PullDockerImage pulls the Docker image
func (d *PlatformDeployer) PullDockerImage(imageRef string, componentType string) error {
	d.Log.Info(fmt.Sprintf("Pulling %s image...", componentType), zap.String("image", imageRef))

	pullCmd := exec.Command("docker", "pull", imageRef)
	var pullStdout, pullStderr bytes.Buffer
	pullCmd.Stdout = &pullStdout
	pullCmd.Stderr = &pullStderr

	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("failed to pull %s image: %w\nstderr: %s", componentType, err, pullStderr.String())
	}

	d.Log.Info(fmt.Sprintf("Successfully pulled %s image", componentType))
	return nil
}

// CheckContainerRunning checks if a container exists and is running
func (d *PlatformDeployer) CheckContainerRunning(containerName string) (bool, string, error) {
	// Check if container exists
	checkCmd := exec.Command("docker", "ps", "-q", "-f", fmt.Sprintf("name=%s", containerName))
	var checkStdout bytes.Buffer
	checkCmd.Stdout = &checkStdout

	if err := checkCmd.Run(); err != nil {
		return false, "", nil // Container doesn't exist
	}

	containerID := strings.TrimSpace(checkStdout.String())
	if containerID == "" {
		return false, "", nil // Container doesn't exist or is not running
	}

	// Get container details to verify it's running
	inspectCmd := exec.Command("docker", "inspect", "--format", "{{.State.Running}}", containerID)
	var inspectStdout bytes.Buffer
	inspectCmd.Stdout = &inspectStdout

	if err := inspectCmd.Run(); err != nil {
		return false, "", fmt.Errorf("failed to inspect container: %w", err)
	}

	isRunning := strings.TrimSpace(inspectStdout.String()) == "true"
	return isRunning, containerID, nil
}

// GetContainerPort retrieves the exposed port for a running container
func (d *PlatformDeployer) GetContainerPort(containerName string, internalPort string) (string, error) {
	cmd := exec.Command("docker", "port", containerName, internalPort)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}

	// Output format is "0.0.0.0:9010" or "[::]:9010"
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return "", fmt.Errorf("no port mapping found for %s", internalPort)
	}

	// Extract just the port number
	parts := strings.Split(output, ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected port format: %s", output)
	}

	return parts[len(parts)-1], nil
}

// CleanupExistingContainer stops and removes existing container if it exists
func (d *PlatformDeployer) CleanupExistingContainer(containerName string) {
	// Check if container exists
	checkCmd := exec.Command("docker", "ps", "-a", "-q", "-f", fmt.Sprintf("name=%s", containerName))
	var checkStdout bytes.Buffer
	checkCmd.Stdout = &checkStdout

	if err := checkCmd.Run(); err == nil && strings.TrimSpace(checkStdout.String()) != "" {
		d.Log.Info("Found existing container, removing...", zap.String("container", containerName))

		// Stop container
		stopCmd := exec.Command("docker", "stop", containerName)
		_ = stopCmd.Run() // Ignore error if container is not running

		// Remove container
		rmCmd := exec.Command("docker", "rm", "-f", containerName)
		if err := rmCmd.Run(); err != nil {
			d.Log.Warn("Failed to remove existing container", zap.Error(err))
		}
	}
}

// MountKeystores adds keystore volume mounts to docker arguments
func (d *PlatformDeployer) MountKeystores(dockerArgs *[]string, keystore *signer.KeystoreReference) error {
	// Ensure the keystore path is absolute
	absPath, err := filepath.Abs(keystore.Path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for keystore: %w", err)
	}

	// Verify the keystore file exists and is readable
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("keystore file not accessible at %s: %w", absPath, err)
	}

	// Mount the specific keystore file
	*dockerArgs = append(*dockerArgs, "-v", fmt.Sprintf("%s:/keystores/operator.keystore.json:ro", absPath))

	d.Log.Debug("Mounted keystore",
		zap.String("source", absPath),
		zap.String("target", "/keystores/operator.keystore.json"),
		zap.String("type", keystore.Type))

	return nil
}

// BuildDockerArgs builds common docker run arguments
func (d *PlatformDeployer) BuildDockerArgs(containerName string, component *runtime.ComponentSpec, cfg *DeploymentConfig) []string {
	dockerArgs := []string{"run", "-d", "--name", containerName}
	dockerArgs = append(dockerArgs, "--restart", "unless-stopped")

	// Add volume mount for config
	dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/config:ro", cfg.ConfigDir))

	// Add environment variables from component
	for _, env := range component.Env {
		dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", env.Name, env.Value))
	}

	// Add ports
	for _, port := range component.Ports {
		dockerArgs = append(dockerArgs, "-p", fmt.Sprintf("%d:%d", port, port))
	}

	return dockerArgs
}

// RunDockerContainer executes the docker run command
func (d *PlatformDeployer) RunDockerContainer(dockerArgs []string, componentType string) (string, error) {
	d.Log.Info(fmt.Sprintf("Starting %s container...", componentType))
	d.Log.Debug("Docker arguments", zap.Strings("args", dockerArgs))

	cmd := exec.Command("docker", dockerArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to start %s container: %w\nstderr: %s", componentType, err, stderr.String())
	}

	containerID := strings.TrimSpace(stdout.String())
	return containerID, nil
}

// InjectFileContentsAsEnvVars injects file contents as environment variables for legacy support
func (d *PlatformDeployer) InjectFileContentsAsEnvVars(dockerArgs []string) []string {
	homeDir, _ := os.UserHomeDir()
	contextDir := filepath.Join(homeDir, ".hgctl", d.Context.Name)

	// This assumes the injectFileContentsAsEnvVars function exists in the package
	return injectFileContentsAsEnvVars(dockerArgs, contextDir, d.Log)
}

// Helper methods

func (d *PlatformDeployer) loadEnvFile(path string, envVars map[string]string) {
	if data, err := os.ReadFile(path); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				// Remove surrounding quotes if present
				value = strings.Trim(value, `"'`)
				envVars[key] = value
			}
		}
	} else if path != "" {
		d.Log.Warn("Failed to read env file", zap.String("path", path), zap.Error(err))
	}
}

func (d *PlatformDeployer) expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

// injectFileContentsAsEnvVars reads keystore and certificate files and injects them as environment variables
func injectFileContentsAsEnvVars(dockerArgs []string, contextDir string, log logger.Logger) []string {
	fileEnvMappings := map[string]string{
		// BN254 keystore
		"operator.bls.keystore.json": "BLS_KEYSTORE_CONTENT",
		// ECDSA keystore
		"operator.ecdsa.keystore.json": "ECDSA_KEYSTORE_CONTENT",
		// Web3 signer certificates for BN254
		"web3signer-bls-ca.crt":     "WEB3_SIGNER_BLS_CA_CERT_CONTENT",
		"web3signer-bls-client.crt": "WEB3_SIGNER_BLS_CLIENT_CERT_CONTENT",
		"web3signer-bls-client.key": "WEB3_SIGNER_BLS_CLIENT_KEY_CONTENT",
		// Web3 signer certificates for ECDSA
		"web3signer-ecdsa-ca.crt":     "WEB3_SIGNER_ECDSA_CA_CERT_CONTENT",
		"web3signer-ecdsa-client.crt": "WEB3_SIGNER_ECDSA_CLIENT_CERT_CONTENT",
		"web3signer-ecdsa-client.key": "WEB3_SIGNER_ECDSA_CLIENT_KEY_CONTENT",
	}

	for fileName, envVar := range fileEnvMappings {
		filePath := filepath.Join(contextDir, fileName)
		if content, err := os.ReadFile(filePath); err == nil {
			// File exists, inject its content
			dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", envVar, string(content)))
			log.Info("Injected file content as environment variable",
				zap.String("file", fileName),
				zap.String("envVar", envVar))
		}
	}

	return dockerArgs
}
