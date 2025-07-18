package testUtils

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

// NetworkConnectivity handles network configuration for Kind clusters
type NetworkConnectivity struct {
	Logger *zap.SugaredLogger
}

// NetworkConfig represents network configuration for Kind testing
type NetworkConfig struct {
	KindClusterName string
	AnvilL1Port     int
	AnvilL2Port     int
	AnvilL1WSPort   int
	AnvilL2WSPort   int
	HostIP          string
	UseHostNetwork  bool
}

// DefaultNetworkConfig returns default network configuration
func DefaultNetworkConfig(kindClusterName string) *NetworkConfig {
	return &NetworkConfig{
		KindClusterName: kindClusterName,
		AnvilL1Port:     8545,
		AnvilL2Port:     9545,
		AnvilL1WSPort:   8545,
		AnvilL2WSPort:   9545,
		HostIP:          "172.17.0.1", // Default Docker bridge IP
		UseHostNetwork:  true,
	}
}

// NewNetworkConnectivity creates a new NetworkConnectivity instance
func NewNetworkConnectivity(logger *zap.SugaredLogger) *NetworkConnectivity {
	return &NetworkConnectivity{
		Logger: logger,
	}
}

// SetupNetworkConnectivity configures network connectivity between Kind and Anvil
func (nc *NetworkConnectivity) SetupNetworkConnectivity(ctx context.Context, cluster *KindCluster, config *NetworkConfig) error {
	nc.Logger.Infof("Setting up network connectivity for Kind cluster %s", cluster.Name)

	// Detect host IP for Docker bridge
	hostIP, err := nc.GetDockerBridgeIP(ctx)
	if err != nil {
		nc.Logger.Warnf("Failed to detect Docker bridge IP, using default: %v", err)
		hostIP = config.HostIP
	}
	config.HostIP = hostIP

	nc.Logger.Infof("Using host IP: %s", hostIP)

	// Configure Kind cluster for host network access
	if err := nc.ConfigureKindForHostAccess(ctx, cluster, config); err != nil {
		return fmt.Errorf("failed to configure Kind for host access: %v", err)
	}

	// Test connectivity to Anvil ports
	if err := nc.TestConnectivity(ctx, cluster, config); err != nil {
		return fmt.Errorf("connectivity test failed: %v", err)
	}

	nc.Logger.Infof("Network connectivity configured successfully")
	return nil
}

// GetDockerBridgeIP detects the Docker bridge IP address
func (nc *NetworkConnectivity) GetDockerBridgeIP(ctx context.Context) (string, error) {
	nc.Logger.Infof("Detecting Docker bridge IP address")

	// Method 1: Check docker0 interface
	if ip, err := nc.getInterfaceIP("docker0"); err == nil {
		nc.Logger.Infof("Found Docker bridge IP via docker0: %s", ip)
		return ip, nil
	}

	// Method 2: Inspect Docker network
	cmd := exec.CommandContext(ctx, "docker", "network", "inspect", "bridge", "--format", "{{range .IPAM.Config}}{{.Gateway}}{{end}}")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to inspect Docker bridge network: %v", err)
	}

	ip := strings.TrimSpace(string(output))
	if ip != "" {
		nc.Logger.Infof("Found Docker bridge IP via network inspect: %s", ip)
		return ip, nil
	}

	// Method 3: Get default gateway from inside a container
	cmd = exec.CommandContext(ctx, "docker", "run", "--rm", "busybox", "route", "-n")
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get route info: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "0.0.0.0") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ip := parts[1]
				nc.Logger.Infof("Found Docker bridge IP via route: %s", ip)
				return ip, nil
			}
		}
	}

	return "", fmt.Errorf("could not detect Docker bridge IP")
}

// getInterfaceIP gets the IP address of a network interface
func (nc *NetworkConnectivity) getInterfaceIP(interfaceName string) (string, error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", fmt.Errorf("interface %s not found: %v", interfaceName, err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("failed to get addresses for interface %s: %v", interfaceName, err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no IPv4 address found for interface %s", interfaceName)
}

// ConfigureKindForHostAccess configures Kind cluster to access host services
func (nc *NetworkConnectivity) ConfigureKindForHostAccess(ctx context.Context, cluster *KindCluster, config *NetworkConfig) error {
	nc.Logger.Infof("Configuring Kind cluster for host access")

	// Create a ConfigMap with host network information
	configMapYAML := fmt.Sprintf(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: host-network-config
  namespace: kube-system
data:
  host-ip: "%s"
  anvil-l1-url: "http://%s:%d"
  anvil-l2-url: "http://%s:%d"
  anvil-l1-ws-url: "ws://%s:%d"
  anvil-l2-ws-url: "ws://%s:%d"
`, config.HostIP, config.HostIP, config.AnvilL1Port, config.HostIP, config.AnvilL2Port, config.HostIP, config.AnvilL1WSPort, config.HostIP, config.AnvilL2WSPort)

	// Apply ConfigMap
	if err := cluster.RunKubectlWithInput(ctx, configMapYAML, "apply", "-f", "-"); err != nil {
		return fmt.Errorf("failed to create host network ConfigMap: %v", err)
	}

	// Create a test pod to verify connectivity
	testPodYAML := fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: network-test
  namespace: default
spec:
  containers:
  - name: test
    image: busybox:latest
    imagePullPolicy: Never
    command: ["sleep", "300"]
    env:
    - name: HOST_IP
      value: "%s"
  restartPolicy: Never
`, config.HostIP)

	// Apply test pod
	if err := cluster.RunKubectlWithInput(ctx, testPodYAML, "apply", "-f", "-"); err != nil {
		return fmt.Errorf("failed to create network test pod: %v", err)
	}

	// Wait for test pod to be ready
	if err := cluster.WaitForPodReady(ctx, "default", "name=network-test", 60*time.Second); err != nil {
		return fmt.Errorf("network test pod not ready: %v", err)
	}

	nc.Logger.Infof("Kind cluster configured for host access")
	return nil
}

// TestConnectivity tests network connectivity from Kind to Anvil
func (nc *NetworkConnectivity) TestConnectivity(ctx context.Context, cluster *KindCluster, config *NetworkConfig) error {
	nc.Logger.Infof("Testing network connectivity from Kind to Anvil")

	// Test L1 connectivity
	if err := nc.testAnvilConnectivity(ctx, cluster, config.HostIP, config.AnvilL1Port, "L1"); err != nil {
		return fmt.Errorf("L1 connectivity test failed: %v", err)
	}

	// Test L2 connectivity
	if err := nc.testAnvilConnectivity(ctx, cluster, config.HostIP, config.AnvilL2Port, "L2"); err != nil {
		return fmt.Errorf("L2 connectivity test failed: %v", err)
	}

	nc.Logger.Infof("Network connectivity tests passed")
	return nil
}

// testAnvilConnectivity tests connectivity to a specific Anvil instance
func (nc *NetworkConnectivity) testAnvilConnectivity(ctx context.Context, cluster *KindCluster, hostIP string, port int, name string) error {
	nc.Logger.Infof("Testing %s Anvil connectivity to %s:%d", name, hostIP, port)

	// Use network test pod to check connectivity
	testCommand := fmt.Sprintf("nc -z -w 5 %s %d", hostIP, port)

	output, err := cluster.RunKubectl(ctx, "exec", "-n", "default", "network-test", "--", "sh", "-c", testCommand)
	if err != nil {
		return fmt.Errorf("connectivity test failed: %v", err)
	}

	nc.Logger.Infof("%s Anvil connectivity test passed: %s", name, string(output))
	return nil
}

// CreateAnvilService creates a Kubernetes service that points to Anvil on the host
func (nc *NetworkConnectivity) CreateAnvilService(ctx context.Context, cluster *KindCluster, config *NetworkConfig, namespace string) error {
	nc.Logger.Infof("Creating Anvil services in namespace %s", namespace)

	// Create L1 Anvil service
	l1ServiceYAML := fmt.Sprintf(`
apiVersion: v1
kind: Service
metadata:
  name: anvil-l1
  namespace: %s
spec:
  type: ExternalName
  externalName: %s
  ports:
  - port: %d
    targetPort: %d
    protocol: TCP
---
apiVersion: v1
kind: Service
metadata:
  name: anvil-l2
  namespace: %s
spec:
  type: ExternalName
  externalName: %s
  ports:
  - port: %d
    targetPort: %d
    protocol: TCP
`, namespace, config.HostIP, config.AnvilL1Port, config.AnvilL1Port, namespace, config.HostIP, config.AnvilL2Port, config.AnvilL2Port)

	// Apply services
	if err := cluster.RunKubectlWithInput(ctx, l1ServiceYAML, "apply", "-f", "-"); err != nil {
		return fmt.Errorf("failed to create Anvil services: %v", err)
	}

	nc.Logger.Infof("Anvil services created in namespace %s", namespace)
	return nil
}

// GetAnvilURLsForExecutor returns the Anvil URLs that the executor should use
func (nc *NetworkConnectivity) GetAnvilURLsForExecutor(config *NetworkConfig) (string, string, string, string) {
	l1URL := fmt.Sprintf("http://%s:%d", config.HostIP, config.AnvilL1Port)
	l2URL := fmt.Sprintf("http://%s:%d", config.HostIP, config.AnvilL2Port)
	l1WSURL := fmt.Sprintf("ws://%s:%d", config.HostIP, config.AnvilL1WSPort)
	l2WSURL := fmt.Sprintf("ws://%s:%d", config.HostIP, config.AnvilL2WSPort)

	return l1URL, l2URL, l1WSURL, l2WSURL
}

// WaitForAnvilAvailability waits for Anvil to be available from Kind
func (nc *NetworkConnectivity) WaitForAnvilAvailability(ctx context.Context, cluster *KindCluster, config *NetworkConfig, timeout time.Duration) error {
	nc.Logger.Infof("Waiting for Anvil to be available from Kind cluster")

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for Anvil to be available")
		case <-ticker.C:
			// Test both L1 and L2 connectivity
			if err := nc.testAnvilConnectivity(ctx, cluster, config.HostIP, config.AnvilL1Port, "L1"); err != nil {
				nc.Logger.Warnf("L1 Anvil not yet available: %v", err)
				continue
			}

			if err := nc.testAnvilConnectivity(ctx, cluster, config.HostIP, config.AnvilL2Port, "L2"); err != nil {
				nc.Logger.Warnf("L2 Anvil not yet available: %v", err)
				continue
			}

			nc.Logger.Infof("Anvil is available from Kind cluster")
			return nil
		}
	}
}

// CleanupNetworkResources cleans up network-related resources
func (nc *NetworkConnectivity) CleanupNetworkResources(ctx context.Context, cluster *KindCluster, namespace string) error {
	nc.Logger.Infof("Cleaning up network resources")

	// Delete test pod
	_, err := cluster.RunKubectl(ctx, "delete", "pod", "network-test", "-n", "default", "--ignore-not-found")
	if err != nil {
		nc.Logger.Warnf("Failed to delete network test pod: %v", err)
	}

	// Delete Anvil services
	_, err = cluster.RunKubectl(ctx, "delete", "service", "anvil-l1", "anvil-l2", "-n", namespace, "--ignore-not-found")
	if err != nil {
		nc.Logger.Warnf("Failed to delete Anvil services: %v", err)
	}

	// Delete ConfigMap
	_, err = cluster.RunKubectl(ctx, "delete", "configmap", "host-network-config", "-n", "kube-system", "--ignore-not-found")
	if err != nil {
		nc.Logger.Warnf("Failed to delete host network ConfigMap: %v", err)
	}

	nc.Logger.Infof("Network resources cleaned up")
	return nil
}

// ValidateNetworkRequirements validates network requirements for testing
func (nc *NetworkConnectivity) ValidateNetworkRequirements(ctx context.Context, config *NetworkConfig) error {
	nc.Logger.Infof("Validating network requirements")

	// Check if required ports are available
	if err := nc.checkPortAvailability(config.AnvilL1Port); err != nil {
		return fmt.Errorf("Anvil L1 port %d is not available: %v", config.AnvilL1Port, err)
	}

	if err := nc.checkPortAvailability(config.AnvilL2Port); err != nil {
		return fmt.Errorf("Anvil L2 port %d is not available: %v", config.AnvilL2Port, err)
	}

	// Check if Docker is running
	if err := nc.checkDockerRunning(ctx); err != nil {
		return fmt.Errorf("Docker is not running: %v", err)
	}

	// Check if Kind is available
	if err := nc.checkKindAvailable(ctx); err != nil {
		return fmt.Errorf("Kind is not available: %v", err)
	}

	nc.Logger.Infof("Network requirements validated")
	return nil
}

// checkPortAvailability checks if a port is available for binding
func (nc *NetworkConnectivity) checkPortAvailability(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("port %d is not available: %v", port, err)
	}
	defer listener.Close()
	return nil
}

// checkDockerRunning checks if Docker daemon is running
func (nc *NetworkConnectivity) checkDockerRunning(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker is not running: %v", err)
	}
	return nil
}

// checkKindAvailable checks if Kind is available
func (nc *NetworkConnectivity) checkKindAvailable(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "kind", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Kind is not available: %v", err)
	}
	return nil
}

// CreateHostAliasManifest creates a manifest to add host aliases to pods
func (nc *NetworkConnectivity) CreateHostAliasManifest(config *NetworkConfig, namespace string) string {
	return fmt.Sprintf(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: host-aliases
  namespace: %s
data:
  host-ip: "%s"
  anvil-l1-host: "%s"
  anvil-l2-host: "%s"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: network-config
  namespace: %s
data:
  L1_RPC_URL: "http://%s:%d"
  L2_RPC_URL: "http://%s:%d"
  L1_WS_URL: "ws://%s:%d"
  L2_WS_URL: "ws://%s:%d"
`, namespace, config.HostIP, config.HostIP, config.HostIP, namespace, config.HostIP, config.AnvilL1Port, config.HostIP, config.AnvilL2Port, config.HostIP, config.AnvilL1WSPort, config.HostIP, config.AnvilL2WSPort)
}

// GetNetworkConfigForExecutor returns network configuration for the executor
func (nc *NetworkConnectivity) GetNetworkConfigForExecutor(config *NetworkConfig) map[string]string {
	return map[string]string{
		"L1_RPC_URL": fmt.Sprintf("http://%s:%d", config.HostIP, config.AnvilL1Port),
		"L2_RPC_URL": fmt.Sprintf("http://%s:%d", config.HostIP, config.AnvilL2Port),
		"L1_WS_URL":  fmt.Sprintf("ws://%s:%d", config.HostIP, config.AnvilL1WSPort),
		"L2_WS_URL":  fmt.Sprintf("ws://%s:%d", config.HostIP, config.AnvilL2WSPort),
	}
}
