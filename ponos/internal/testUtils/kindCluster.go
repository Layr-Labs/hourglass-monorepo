package testUtils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

// KindCluster represents a Kind Kubernetes cluster for testing
type KindCluster struct {
	Name       string
	ConfigPath string
	KubeConfig string
	Logger     *zap.SugaredLogger
}

// KindClusterConfig holds configuration for Kind cluster creation
type KindClusterConfig struct {
	Name               string
	NodeImage          string
	APIServerPort      int
	WorkerNodes        int
	ControlPlaneNodes  int
	KubeConfigPath     string
	ConfigTemplatePath string
	Logger             *zap.SugaredLogger
}

// DefaultKindClusterConfig returns default configuration for Kind cluster
func DefaultKindClusterConfig(logger *zap.SugaredLogger) *KindClusterConfig {
	return &KindClusterConfig{
		Name:              "hourglass-test",       // Use deterministic name since tests run in series
		NodeImage:         "kindest/node:v1.29.0", // Use stable Kubernetes version
		APIServerPort:     0,                      // Random port
		WorkerNodes:       1,                      // Single worker node for testing
		ControlPlaneNodes: 1,                      // Single control plane
		KubeConfigPath:    "",                     // Will be auto-generated
		Logger:            logger,
	}
}

// sanitizeTestName removes invalid characters from test names for Kind cluster names
func sanitizeTestName(testName string) string {
	// Kind cluster names must be RFC 1123 compliant
	sanitized := strings.ToLower(testName)
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ReplaceAll(sanitized, "/", "-")

	// Remove consecutive dashes
	for strings.Contains(sanitized, "--") {
		sanitized = strings.ReplaceAll(sanitized, "--", "-")
	}

	// Trim leading/trailing dashes
	sanitized = strings.Trim(sanitized, "-")

	// Ensure max length (Kind has limits)
	if len(sanitized) > 40 {
		sanitized = sanitized[:40]
	}

	return sanitized
}

// CreateKindCluster creates a new Kind cluster with the specified configuration
func CreateKindCluster(ctx context.Context, t *testing.T, config *KindClusterConfig) (*KindCluster, func(), error) {
	if config.Logger == nil {
		return nil, nil, fmt.Errorf("logger is required")
	}

	cluster := &KindCluster{
		Name:   config.Name,
		Logger: config.Logger,
	}

	// Create temporary directory for cluster configuration
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("kind-cluster-%s", config.Name))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp dir: %v", err)
	}

	// Generate Kind configuration
	kindConfigPath := filepath.Join(tempDir, "kind-config.yaml")
	if err := generateKindConfig(kindConfigPath, config); err != nil {
		os.RemoveAll(tempDir)
		return nil, nil, fmt.Errorf("failed to generate kind config: %v", err)
	}

	cluster.ConfigPath = kindConfigPath

	// Set kubeconfig path
	if config.KubeConfigPath == "" {
		cluster.KubeConfig = filepath.Join(tempDir, "kubeconfig")
	} else {
		cluster.KubeConfig = config.KubeConfigPath
	}

	// Create the cluster
	config.Logger.Infof("Creating Kind cluster: %s", config.Name)

	// Check if cluster already exists and delete it
	if clusterExists(config.Name) {
		config.Logger.Warnf("Cluster %s already exists, deleting it first", config.Name)
		if err := deleteKindCluster(config.Name, config.Logger); err != nil {
			config.Logger.Warnf("Failed to delete existing cluster: %v", err)
		}
	}

	// Create cluster with timeout
	createCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	createCmd := exec.CommandContext(createCtx, "kind", "create", "cluster",
		"--name", config.Name,
		"--config", kindConfigPath,
		"--kubeconfig", cluster.KubeConfig,
		"--wait", "60s",
	)

	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr

	if err := createCmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return nil, nil, fmt.Errorf("failed to create kind cluster: %v", err)
	}

	config.Logger.Infof("Kind cluster %s created successfully", config.Name)

	// Setup cleanup function
	cleanup := func() {
		config.Logger.Infof("Cleaning up Kind cluster: %s", config.Name)

		// Delete the cluster
		if err := deleteKindCluster(config.Name, config.Logger); err != nil {
			config.Logger.Errorf("Failed to delete Kind cluster: %v", err)
		}

		// Clean up temporary directory
		if err := os.RemoveAll(tempDir); err != nil {
			config.Logger.Errorf("Failed to clean up temp dir: %v", err)
		}
	}

	// Verify cluster is ready
	if err := waitForClusterReady(ctx, cluster, 5*time.Minute); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("cluster not ready: %v", err)
	}

	return cluster, cleanup, nil
}

// generateKindConfig creates a Kind configuration file
func generateKindConfig(configPath string, config *KindClusterConfig) error {
	kindConfig := fmt.Sprintf(`kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: %s
  extraPortMappings:
  # Allow access to Anvil nodes on host
  - containerPort: 30000
    hostPort: 30000
    protocol: TCP
  - containerPort: 30001
    hostPort: 30001
    protocol: TCP
  # NodePort range for services
  - containerPort: 30080
    hostPort: 30080
    protocol: TCP
  - containerPort: 30090
    hostPort: 30090
    protocol: TCP
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
`, config.NodeImage)

	// Add worker nodes if specified
	for i := 0; i < config.WorkerNodes; i++ {
		kindConfig += fmt.Sprintf(`- role: worker
  image: %s
`, config.NodeImage)
	}

	// Add networking configuration
	kindConfig += `networking:
  apiServerAddress: "127.0.0.1"
  disableDefaultCNI: false
  kubeProxyMode: "iptables"
`

	return os.WriteFile(configPath, []byte(kindConfig), 0644)
}

// clusterExists checks if a Kind cluster with the given name exists
func clusterExists(name string) bool {
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	clusters := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, cluster := range clusters {
		if strings.TrimSpace(cluster) == name {
			return true
		}
	}
	return false
}

// deleteKindCluster deletes a Kind cluster
func deleteKindCluster(name string, logger *zap.SugaredLogger) error {
	if !clusterExists(name) {
		logger.Infof("Cluster %s does not exist, nothing to delete", name)
		return nil
	}

	// First try fast deletion by killing containers directly
	logger.Infof("Attempting fast cluster deletion for %s", name)

	// Get the containers for this cluster
	listCmd := exec.Command("docker", "ps", "-a", "-q", "--filter", fmt.Sprintf("label=io.x-k8s.kind.cluster=%s", name))
	containers, err := listCmd.Output()
	if err != nil {
		logger.Warnf("Failed to list containers for cluster %s: %v", name, err)
	} else if len(containers) > 0 {
		// Kill containers directly
		containerIDs := strings.Fields(strings.TrimSpace(string(containers)))
		for _, containerID := range containerIDs {
			logger.Infof("Force killing container %s", containerID)
			killCmd := exec.Command("docker", "kill", containerID)
			_ = killCmd.Run() // Ignore errors

			removeCmd := exec.Command("docker", "rm", "-f", containerID)
			_ = removeCmd.Run() // Ignore errors
		}
	}

	// Then run the normal kind delete as cleanup
	deleteCmd := exec.Command("kind", "delete", "cluster", "--name", name)
	deleteCmd.Stdout = os.Stdout
	deleteCmd.Stderr = os.Stderr

	if err := deleteCmd.Run(); err != nil {
		logger.Warnf("Kind delete failed (containers may already be gone): %v", err)
	}

	logger.Infof("Successfully deleted Kind cluster: %s", name)
	return nil
}

// waitForClusterReady waits for the cluster to be ready
func waitForClusterReady(ctx context.Context, cluster *KindCluster, timeout time.Duration) error {
	cluster.Logger.Infof("Waiting for cluster %s to be ready...", cluster.Name)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for cluster to be ready")
		case <-ticker.C:
			if isClusterReady(cluster) {
				cluster.Logger.Infof("Cluster %s is ready", cluster.Name)
				return nil
			}
		}
	}
}

// isClusterReady checks if the cluster is ready by running kubectl get nodes
func isClusterReady(cluster *KindCluster) bool {
	cmd := exec.Command("kubectl", "get", "nodes", "--kubeconfig", cluster.KubeConfig, "-o", "jsonpath={.items[*].status.conditions[?(@.type=='Ready')].status}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check if all nodes are ready
	statuses := strings.Fields(string(output))
	for _, status := range statuses {
		if strings.TrimSpace(status) != "True" {
			return false
		}
	}

	return len(statuses) > 0
}

// GetKubeConfigPath returns the path to the kubeconfig file
func (kc *KindCluster) GetKubeConfigPath() string {
	return kc.KubeConfig
}

// RunKubectl executes a kubectl command against the cluster
func (kc *KindCluster) RunKubectl(ctx context.Context, args ...string) ([]byte, error) {
	kubectlArgs := append([]string{"--kubeconfig", kc.KubeConfig}, args...)
	cmd := exec.CommandContext(ctx, "kubectl", kubectlArgs...)
	return cmd.Output()
}

// RunKubectlWithInput executes a kubectl command with stdin input
func (kc *KindCluster) RunKubectlWithInput(ctx context.Context, input string, args ...string) error {
	kubectlArgs := append([]string{"--kubeconfig", kc.KubeConfig}, args...)
	cmd := exec.CommandContext(ctx, "kubectl", kubectlArgs...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// LoadDockerImage loads a Docker image into the Kind cluster
func (kc *KindCluster) LoadDockerImage(ctx context.Context, imageName string) error {
	kc.Logger.Infof("Loading Docker image %s into Kind cluster %s", imageName, kc.Name)

	cmd := exec.CommandContext(ctx, "kind", "load", "docker-image", imageName, "--name", kc.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to load image %s into kind cluster: %v", imageName, err)
	}

	kc.Logger.Infof("Successfully loaded image %s into cluster", imageName)
	return nil
}

// WaitForPodReady waits for a pod to be ready
func (kc *KindCluster) WaitForPodReady(ctx context.Context, namespace, labelSelector string, timeout time.Duration) error {
	kc.Logger.Infof("Waiting for pod with selector %s in namespace %s to be ready", labelSelector, namespace)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for pod to be ready")
		case <-ticker.C:
			output, err := kc.RunKubectl(ctx, "get", "pods", "-n", namespace, "-l", labelSelector, "-o", "jsonpath={.items[*].status.phase}")
			if err != nil {
				continue
			}

			phases := strings.Fields(string(output))
			if len(phases) == 0 {
				continue
			}

			allRunning := true
			for _, phase := range phases {
				if strings.TrimSpace(phase) != "Running" {
					allRunning = false
					break
				}
			}

			if allRunning {
				kc.Logger.Infof("Pod with selector %s is ready", labelSelector)
				return nil
			}
		}
	}
}

// GetNodeIP gets the IP address of a Kind node (for accessing services)
func (kc *KindCluster) GetNodeIP(ctx context.Context) (string, error) {
	// For Kind, we typically use localhost since it's running in Docker
	// But we can get the container IP if needed
	cmd := exec.CommandContext(ctx, "docker", "inspect", fmt.Sprintf("%s-control-plane", kc.Name), "--format", "{{.NetworkSettings.IPAddress}}")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get node IP: %v", err)
	}

	ip := strings.TrimSpace(string(output))
	if ip == "" {
		// Fallback to localhost for Kind
		return "127.0.0.1", nil
	}

	return ip, nil
}

// CreateNamespace creates a namespace in the cluster
func (kc *KindCluster) CreateNamespace(ctx context.Context, namespace string) error {
	kc.Logger.Infof("Creating namespace %s in cluster %s", namespace, kc.Name)

	// Check if namespace already exists
	_, err := kc.RunKubectl(ctx, "get", "namespace", namespace)
	if err == nil {
		kc.Logger.Infof("Namespace %s already exists", namespace)
		return nil
	}

	// Create namespace
	_, err = kc.RunKubectl(ctx, "create", "namespace", namespace)
	if err != nil {
		return fmt.Errorf("failed to create namespace %s: %v", namespace, err)
	}

	kc.Logger.Infof("Successfully created namespace %s", namespace)
	return nil
}

// DeleteNamespace deletes a namespace from the cluster
func (kc *KindCluster) DeleteNamespace(ctx context.Context, namespace string) error {
	kc.Logger.Infof("Deleting namespace %s from cluster %s", namespace, kc.Name)

	// First delete all resources in the namespace with force
	_, err := kc.RunKubectl(ctx, "delete", "all", "--all", "-n", namespace, "--ignore-not-found", "--force", "--grace-period=0")
	if err != nil {
		kc.Logger.Warnf("Failed to delete all resources in namespace %s: %v", namespace, err)
	}

	// Then delete the namespace with force
	_, err = kc.RunKubectl(ctx, "delete", "namespace", namespace, "--ignore-not-found", "--force", "--grace-period=0")
	if err != nil {
		kc.Logger.Warnf("Failed to delete namespace %s: %v", namespace, err)
	}

	kc.Logger.Infof("Successfully deleted namespace %s", namespace)
	return nil
}

// CleanupAllTestClusters removes the test cluster to prevent port conflicts
func CleanupAllTestClusters(logger *zap.SugaredLogger) error {
	logger.Info("Cleaning up existing test cluster")

	// Delete the specific test cluster
	if err := deleteKindCluster("hourglass-test", logger); err != nil {
		logger.Warnf("Failed to delete test cluster: %v", err)
		return err
	}

	logger.Info("Finished cleaning up test cluster")
	return nil
}
