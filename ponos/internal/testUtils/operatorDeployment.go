package testUtils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// OperatorDeployment represents a deployed Hourglass operator
type OperatorDeployment struct {
	Namespace   string
	ReleaseName string
	ChartPath   string
	KindCluster *KindCluster
	Logger      *zap.SugaredLogger
	Values      map[string]interface{}
}

// OperatorDeploymentConfig holds configuration for operator deployment
type OperatorDeploymentConfig struct {
	Namespace       string
	ReleaseName     string
	ChartPath       string
	Values          map[string]interface{}
	ImageRepository string
	ImageTag        string
	ImagePullPolicy string
	WaitTimeout     time.Duration
	Logger          *zap.SugaredLogger
}

// DefaultOperatorDeploymentConfig returns default configuration for operator deployment
func DefaultOperatorDeploymentConfig(projectRoot string, logger *zap.SugaredLogger) *OperatorDeploymentConfig {
	imagePullPolicy := "Always"
	if os.Getenv("USE_LOCAL_IMAGES") == "true" {
		imagePullPolicy = "Never" // Use locally loaded images
	}
	return &OperatorDeploymentConfig{
		Namespace:       "hourglass-system",
		ReleaseName:     "hourglass-operator",
		ChartPath:       filepath.Join(projectRoot, "..", "hourglass-operator", "charts", "hourglass-operator"),
		ImageRepository: "hourglass/operator",
		ImageTag:        "test",
		ImagePullPolicy: imagePullPolicy, // Use locally loaded images
		WaitTimeout:     3 * time.Minute,
		Logger:          logger,
		Values: map[string]interface{}{
			"replicaCount": 1,
			"image": map[string]interface{}{
				"repository": "hourglass/operator",
				"tag":        "test",
				"pullPolicy": imagePullPolicy,
			},
			"resources": map[string]interface{}{
				"limits": map[string]interface{}{
					"cpu":    "500m",
					"memory": "512Mi",
				},
				"requests": map[string]interface{}{
					"cpu":    "100m",
					"memory": "128Mi",
				},
			},
			"rbac": map[string]interface{}{
				"create": true,
			},
			"serviceAccount": map[string]interface{}{
				"create": true,
				"name":   "hourglass-operator",
			},
		},
	}
}

// DeployOperator deploys the Hourglass operator to the Kind cluster
func DeployOperator(ctx context.Context, cluster *KindCluster, config *OperatorDeploymentConfig) (*OperatorDeployment, func(), error) {
	if config.Logger == nil {
		return nil, nil, fmt.Errorf("logger is required")
	}

	deployment := &OperatorDeployment{
		Namespace:   config.Namespace,
		ReleaseName: config.ReleaseName,
		ChartPath:   config.ChartPath,
		KindCluster: cluster,
		Logger:      config.Logger,
		Values:      config.Values,
	}

	// Create namespace if it doesn't exist
	if err := cluster.CreateNamespace(ctx, config.Namespace); err != nil {
		return nil, nil, fmt.Errorf("failed to create namespace: %v", err)
	}

	// Install CRDs first
	if err := installCRDs(ctx, cluster, config); err != nil {
		return nil, nil, fmt.Errorf("failed to install CRDs: %v", err)
	}

	// Deploy operator using Helm
	if err := deployOperatorWithHelm(ctx, cluster, config); err != nil {
		return nil, nil, fmt.Errorf("failed to deploy operator: %v", err)
	}

	// Wait for operator to be ready
	if err := waitForOperatorReady(ctx, cluster, config); err != nil {
		return nil, nil, fmt.Errorf("operator not ready: %v", err)
	}

	config.Logger.Infof("Hourglass operator deployed successfully to namespace %s", config.Namespace)

	// Setup cleanup function
	cleanup := func() {
		config.Logger.Infof("Cleaning up operator deployment: %s", config.ReleaseName)

		// Set a shorter context for cleanup to avoid hanging
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Uninstall Helm release
		if err := uninstallHelmRelease(cleanupCtx, cluster, config); err != nil {
			config.Logger.Errorf("Failed to uninstall Helm release: %v", err)
		}

		// Delete CRDs
		if err := deleteCRDs(cleanupCtx, cluster, config); err != nil {
			config.Logger.Errorf("Failed to delete CRDs: %v", err)
		}

		// Delete namespace
		if err := cluster.DeleteNamespace(cleanupCtx, config.Namespace); err != nil {
			config.Logger.Errorf("Failed to delete namespace: %v", err)
		}
	}

	return deployment, cleanup, nil
}

// installCRDs installs the Performer CRDs
func installCRDs(ctx context.Context, cluster *KindCluster, config *OperatorDeploymentConfig) error {
	config.Logger.Infof("Installing CRDs for Hourglass operator")

	// Look for CRD files in the hourglass-operator directory
	// The project root is ponos, so we need to go up one level to find hourglass-operator
	crdPath := filepath.Join(config.ChartPath, "..", "..", "config", "crd", "bases")
	if _, err := os.Stat(crdPath); os.IsNotExist(err) {
		// Try alternative path in chart directory
		crdPath = filepath.Join(config.ChartPath, "crds")
		if _, err := os.Stat(crdPath); os.IsNotExist(err) {
			return fmt.Errorf("CRD files not found in %s or %s", filepath.Join(config.ChartPath, "..", "..", "config", "crd", "bases"), crdPath)
		}
	}

	// Apply CRDs - specifically the Performer CRD file
	performerCRDPath := filepath.Join(crdPath, "hourglass.eigenlayer.io_performers.yaml")
	output, err := cluster.RunKubectl(ctx, "apply", "-f", performerCRDPath)
	if err != nil {
		return fmt.Errorf("failed to apply Performer CRD: %v\nOutput: %s", err, string(output))
	}

	// Wait for CRDs to be established
	if err := waitForCRDsReady(ctx, cluster, config); err != nil {
		return fmt.Errorf("CRDs not ready: %v", err)
	}

	config.Logger.Infof("CRDs installed successfully")
	return nil
}

// waitForCRDsReady waits for CRDs to be established
func waitForCRDsReady(ctx context.Context, cluster *KindCluster, config *OperatorDeploymentConfig) error {
	config.Logger.Infof("Waiting for CRDs to be established...")

	timeout := 60 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for CRDs to be ready")
		case <-ticker.C:
			// Check if Performer CRD is established
			output, err := cluster.RunKubectl(ctx, "get", "crd", "performers.hourglass.eigenlayer.io", "-o", "jsonpath={.status.conditions[?(@.type=='Established')].status}")
			if err != nil {
				continue
			}

			if strings.TrimSpace(string(output)) == "True" {
				config.Logger.Infof("CRDs are established")
				return nil
			}
		}
	}
}

// deployOperatorWithHelm deploys the operator using Helm
func deployOperatorWithHelm(ctx context.Context, cluster *KindCluster, config *OperatorDeploymentConfig) error {
	config.Logger.Infof("Deploying operator with Helm: %s", config.ReleaseName)

	// Generate values file
	valuesFile, err := generateValuesFile(config)
	if err != nil {
		return fmt.Errorf("failed to generate values file: %v", err)
	}
	defer os.Remove(valuesFile)

	// Check if Helm release already exists
	_, err = cluster.RunKubectl(ctx, "get", "secret", "-n", config.Namespace, "-l", fmt.Sprintf("name=%s", config.ReleaseName))
	if err == nil {
		config.Logger.Warnf("Helm release %s already exists, upgrading", config.ReleaseName)
		return upgradeHelmRelease(ctx, cluster, config, valuesFile)
	}

	// Install Helm release
	return installHelmRelease(ctx, cluster, config, valuesFile)
}

// generateValuesFile creates a temporary values file for Helm
func generateValuesFile(config *OperatorDeploymentConfig) (string, error) {
	tempFile, err := os.CreateTemp("", "operator-values-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	// Convert values to YAML
	valuesYAML := generateValuesYAML(config.Values)

	if _, err := tempFile.WriteString(valuesYAML); err != nil {
		return "", fmt.Errorf("failed to write values file: %v", err)
	}

	return tempFile.Name(), nil
}

// generateValuesYAML converts map to YAML string (simplified version)
func generateValuesYAML(values map[string]interface{}) string {
	var result strings.Builder

	for key, value := range values {
		writeValue(&result, key, value, 0)
	}

	return result.String()
}

// writeValue writes a value to the YAML builder
func writeValue(builder *strings.Builder, key string, value interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)

	switch v := value.(type) {
	case map[string]interface{}:
		builder.WriteString(fmt.Sprintf("%s%s:\n", prefix, key))
		for k, val := range v {
			writeValue(builder, k, val, indent+1)
		}
	case string:
		builder.WriteString(fmt.Sprintf("%s%s: %q\n", prefix, key, v))
	case int:
		builder.WriteString(fmt.Sprintf("%s%s: %d\n", prefix, key, v))
	case bool:
		builder.WriteString(fmt.Sprintf("%s%s: %t\n", prefix, key, v))
	default:
		builder.WriteString(fmt.Sprintf("%s%s: %v\n", prefix, key, v))
	}
}

// installHelmRelease installs a new Helm release
func installHelmRelease(ctx context.Context, cluster *KindCluster, config *OperatorDeploymentConfig, valuesFile string) error {
	config.Logger.Infof("Installing Helm release: %s", config.ReleaseName)

	// Use kubectl to simulate helm install
	// In a real scenario, you would use the Helm Go client library
	return applyHelmTemplate(ctx, cluster, config, valuesFile)
}

// upgradeHelmRelease upgrades an existing Helm release
func upgradeHelmRelease(ctx context.Context, cluster *KindCluster, config *OperatorDeploymentConfig, valuesFile string) error {
	config.Logger.Infof("Upgrading Helm release: %s", config.ReleaseName)

	// For testing purposes, we'll just reapply the templates
	return applyHelmTemplate(ctx, cluster, config, valuesFile)
}

// applyHelmTemplate applies Helm templates using kubectl
func applyHelmTemplate(ctx context.Context, cluster *KindCluster, config *OperatorDeploymentConfig, valuesFile string) error {
	// Generate Kubernetes manifests from the chart
	manifests, err := generateKubernetesManifests(config, valuesFile)
	if err != nil {
		return fmt.Errorf("failed to generate manifests: %v", err)
	}

	// Apply manifests
	return cluster.RunKubectlWithInput(ctx, manifests, "apply", "-f", "-")
}

// generateKubernetesManifests generates Kubernetes manifests for the operator
func generateKubernetesManifests(config *OperatorDeploymentConfig, valuesFile string) (string, error) {
	// This is a simplified version - in reality you'd use Helm templating
	// For testing, we'll generate basic manifests

	manifests := fmt.Sprintf(`
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: hourglass-operator
  namespace: %s
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: hourglass-operator
rules:
- apiGroups: [""]
  resources: ["pods", "services", "endpoints", "events"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: ["hourglass.eigenlayer.io"]
  resources: ["performers"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: ["hourglass.eigenlayer.io"]
  resources: ["performers/status"]
  verbs: ["get", "patch", "update"]
- apiGroups: ["hourglass.eigenlayer.io"]
  resources: ["performers/finalizers"]
  verbs: ["update"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: hourglass-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: hourglass-operator
subjects:
- kind: ServiceAccount
  name: hourglass-operator
  namespace: %s
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hourglass-operator
  namespace: %s
  labels:
    app: hourglass-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: hourglass-operator
  template:
    metadata:
      labels:
        app: hourglass-operator
    spec:
      serviceAccountName: hourglass-operator
      containers:
      - name: manager
        image: %s:%s
        imagePullPolicy: %s
        command:
        - /manager
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
          requests:
            cpu: 100m
            memory: 128Mi
        ports:
        - containerPort: 8080
          name: metrics
        - containerPort: 8081
          name: health
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 15
          periodSeconds: 20
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
`, config.Namespace, config.Namespace, config.Namespace, config.ImageRepository, config.ImageTag, config.ImagePullPolicy)

	return manifests, nil
}

// waitForOperatorReady waits for the operator to be ready
func waitForOperatorReady(ctx context.Context, cluster *KindCluster, config *OperatorDeploymentConfig) error {
	config.Logger.Infof("Waiting for operator to be ready...")

	return cluster.WaitForPodReady(ctx, config.Namespace, "app=hourglass-operator", config.WaitTimeout)
}

// uninstallHelmRelease uninstalls the Helm release
func uninstallHelmRelease(ctx context.Context, cluster *KindCluster, config *OperatorDeploymentConfig) error {
	config.Logger.Infof("Uninstalling Helm release: %s", config.ReleaseName)

	// Delete the deployment with force and grace period 0
	_, err := cluster.RunKubectl(ctx, "delete", "deployment", "hourglass-operator", "-n", config.Namespace, "--ignore-not-found", "--force", "--grace-period=0")
	if err != nil {
		config.Logger.Warnf("Failed to delete deployment: %v", err)
	}

	// Delete RBAC resources with force
	_, err = cluster.RunKubectl(ctx, "delete", "clusterrolebinding", "hourglass-operator", "--ignore-not-found", "--force", "--grace-period=0")
	if err != nil {
		config.Logger.Warnf("Failed to delete clusterrolebinding: %v", err)
	}

	_, err = cluster.RunKubectl(ctx, "delete", "clusterrole", "hourglass-operator", "--ignore-not-found", "--force", "--grace-period=0")
	if err != nil {
		config.Logger.Warnf("Failed to delete clusterrole: %v", err)
	}

	_, err = cluster.RunKubectl(ctx, "delete", "serviceaccount", "hourglass-operator", "-n", config.Namespace, "--ignore-not-found", "--force", "--grace-period=0")
	if err != nil {
		config.Logger.Warnf("Failed to delete serviceaccount: %v", err)
	}

	config.Logger.Infof("Helm release uninstalled successfully")
	return nil
}

// deleteCRDs deletes the CRDs
func deleteCRDs(ctx context.Context, cluster *KindCluster, config *OperatorDeploymentConfig) error {
	config.Logger.Infof("Deleting CRDs")

	// Nuclear option: patch away finalizers first
	config.Logger.Infof("Removing finalizers from all performers")
	_, err := cluster.RunKubectl(ctx, "patch", "performers", "--all", "--all-namespaces", "--type=merge", "-p", `{"metadata":{"finalizers":null}}`, "--ignore-not-found")
	if err != nil {
		config.Logger.Warnf("Failed to remove finalizers: %v", err)
	}

	// Then try to delete any performer instances with force
	_, err = cluster.RunKubectl(ctx, "delete", "performers", "--all", "--all-namespaces", "--ignore-not-found", "--force", "--grace-period=0")
	if err != nil {
		config.Logger.Warnf("Failed to delete performer instances: %v", err)
	}

	// Then delete the CRD itself with force
	_, err = cluster.RunKubectl(ctx, "delete", "crd", "performers.hourglass.eigenlayer.io", "--ignore-not-found", "--force", "--grace-period=0")
	if err != nil {
		config.Logger.Warnf("Failed to delete CRDs: %v", err)
	}

	config.Logger.Infof("CRDs deleted successfully")
	return nil
}

// GetOperatorPods returns the operator pods
func (od *OperatorDeployment) GetOperatorPods(ctx context.Context) ([]byte, error) {
	return od.KindCluster.RunKubectl(ctx, "get", "pods", "-n", od.Namespace, "-l", "app=hourglass-operator", "-o", "json")
}

// GetOperatorLogs returns the operator logs
func (od *OperatorDeployment) GetOperatorLogs(ctx context.Context) ([]byte, error) {
	return od.KindCluster.RunKubectl(ctx, "logs", "-n", od.Namespace, "-l", "app=hourglass-operator", "--tail=100")
}

// IsOperatorReady checks if the operator is ready
func (od *OperatorDeployment) IsOperatorReady(ctx context.Context) (bool, error) {
	output, err := od.KindCluster.RunKubectl(ctx, "get", "pods", "-n", od.Namespace, "-l", "app=hourglass-operator", "-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return false, err
	}

	phases := strings.Fields(string(output))
	if len(phases) == 0 {
		return false, nil
	}

	for _, phase := range phases {
		if strings.TrimSpace(phase) != "Running" {
			return false, nil
		}
	}

	return true, nil
}
