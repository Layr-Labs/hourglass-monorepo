package testUtils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TestNamespaceManager manages test namespaces in Kind clusters
type TestNamespaceManager struct {
	Logger *zap.SugaredLogger
}

// TestNamespaceConfig represents configuration for test namespaces
type TestNamespaceConfig struct {
	Name           string
	Labels         map[string]string
	Annotations    map[string]string
	ResourceQuota  *ResourceQuotaConfig
	NetworkPolicy  *NetworkPolicyConfig
	ServiceAccount *ServiceAccountConfig
	RoleBinding    *RoleBindingConfig
	CleanupOnExit  bool
	CleanupTimeout time.Duration
}

// ResourceQuotaConfig represents resource quota configuration
type ResourceQuotaConfig struct {
	Hard map[string]string
}

// NetworkPolicyConfig represents network policy configuration
type NetworkPolicyConfig struct {
	Enabled           bool
	AllowEgress       bool
	AllowedNamespaces []string
}

// ServiceAccountConfig represents service account configuration
type ServiceAccountConfig struct {
	Name   string
	Create bool
}

// RoleBindingConfig represents role binding configuration
type RoleBindingConfig struct {
	RoleName           string
	ServiceAccountName string
	ClusterRole        bool
}

// CreateTestNamespace creates a test namespace with the specified configuration
func (tnm *TestNamespaceManager) CreateTestNamespace(ctx context.Context, cluster *KindCluster, config *TestNamespaceConfig) (func(), error) {
	tnm.Logger.Infof("Creating test namespace: %s", config.Name)

	// Create namespace
	if err := tnm.createNamespace(ctx, cluster, config); err != nil {
		return nil, fmt.Errorf("failed to create namespace: %v", err)
	}

	// Create resource quota
	if config.ResourceQuota != nil {
		if err := tnm.createResourceQuota(ctx, cluster, config); err != nil {
			return nil, fmt.Errorf("failed to create resource quota: %v", err)
		}
	}

	// Create network policy
	if config.NetworkPolicy != nil && config.NetworkPolicy.Enabled {
		if err := tnm.createNetworkPolicy(ctx, cluster, config); err != nil {
			return nil, fmt.Errorf("failed to create network policy: %v", err)
		}
	}

	// Create service account
	if config.ServiceAccount != nil && config.ServiceAccount.Create {
		if err := tnm.createServiceAccount(ctx, cluster, config); err != nil {
			return nil, fmt.Errorf("failed to create service account: %v", err)
		}
	}

	// Create role and role binding
	if config.RoleBinding != nil {
		if err := tnm.createRoleBinding(ctx, cluster, config); err != nil {
			return nil, fmt.Errorf("failed to create role binding: %v", err)
		}
	}

	tnm.Logger.Infof("Test namespace %s created successfully", config.Name)

	// Setup cleanup function
	cleanup := func() {
		if config.CleanupOnExit {
			tnm.Logger.Infof("Cleaning up test namespace: %s", config.Name)
			if err := tnm.CleanupTestNamespace(ctx, cluster, config); err != nil {
				tnm.Logger.Errorf("Failed to cleanup test namespace: %v", err)
			}
		}
	}

	return cleanup, nil
}

// createNamespace creates the namespace
func (tnm *TestNamespaceManager) createNamespace(ctx context.Context, cluster *KindCluster, config *TestNamespaceConfig) error {
	tnm.Logger.Infof("Creating namespace: %s", config.Name)

	// Generate namespace YAML
	namespaceYAML := tnm.generateNamespaceYAML(config)

	// Apply namespace
	if err := cluster.RunKubectlWithInput(ctx, namespaceYAML, "apply", "-f", "-"); err != nil {
		return fmt.Errorf("failed to apply namespace: %v", err)
	}

	// Wait for namespace to be active
	if err := tnm.waitForNamespaceActive(ctx, cluster, config.Name, 30*time.Second); err != nil {
		return fmt.Errorf("namespace not active: %v", err)
	}

	return nil
}

// generateNamespaceYAML generates the namespace YAML
func (tnm *TestNamespaceManager) generateNamespaceYAML(config *TestNamespaceConfig) string {
	var labelsStr, annotationsStr string

	// Generate labels
	if len(config.Labels) > 0 {
		labelsStr = "  labels:\n"
		for key, value := range config.Labels {
			labelsStr += fmt.Sprintf("    %s: %s\n", key, value)
		}
	}

	// Generate annotations
	if len(config.Annotations) > 0 {
		annotationsStr = "  annotations:\n"
		for key, value := range config.Annotations {
			annotationsStr += fmt.Sprintf("    %s: %s\n", key, value)
		}
	}

	return fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
%s%s`, config.Name, labelsStr, annotationsStr)
}

// createResourceQuota creates a resource quota for the namespace
func (tnm *TestNamespaceManager) createResourceQuota(ctx context.Context, cluster *KindCluster, config *TestNamespaceConfig) error {
	tnm.Logger.Infof("Creating resource quota for namespace: %s", config.Name)

	// Generate resource quota YAML
	resourceQuotaYAML := tnm.generateResourceQuotaYAML(config)

	// Apply resource quota
	if err := cluster.RunKubectlWithInput(ctx, resourceQuotaYAML, "apply", "-f", "-"); err != nil {
		return fmt.Errorf("failed to apply resource quota: %v", err)
	}

	return nil
}

// generateResourceQuotaYAML generates the resource quota YAML
func (tnm *TestNamespaceManager) generateResourceQuotaYAML(config *TestNamespaceConfig) string {
	hardStr := ""
	for key, value := range config.ResourceQuota.Hard {
		hardStr += fmt.Sprintf("    %s: %s\n", key, value)
	}

	return fmt.Sprintf(`apiVersion: v1
kind: ResourceQuota
metadata:
  name: test-resource-quota
  namespace: %s
spec:
  hard:
%s`, config.Name, hardStr)
}

// createNetworkPolicy creates a network policy for the namespace
func (tnm *TestNamespaceManager) createNetworkPolicy(ctx context.Context, cluster *KindCluster, config *TestNamespaceConfig) error {
	tnm.Logger.Infof("Creating network policy for namespace: %s", config.Name)

	// Generate network policy YAML
	networkPolicyYAML := tnm.generateNetworkPolicyYAML(config)

	// Apply network policy
	if err := cluster.RunKubectlWithInput(ctx, networkPolicyYAML, "apply", "-f", "-"); err != nil {
		return fmt.Errorf("failed to apply network policy: %v", err)
	}

	return nil
}

// generateNetworkPolicyYAML generates the network policy YAML
func (tnm *TestNamespaceManager) generateNetworkPolicyYAML(config *TestNamespaceConfig) string {
	egressRules := ""
	if config.NetworkPolicy.AllowEgress {
		egressRules = `  - {}  # Allow all egress`

		// Add specific namespace rules
		if len(config.NetworkPolicy.AllowedNamespaces) > 0 {
			egressRules = "  egress:\n"
			for _, ns := range config.NetworkPolicy.AllowedNamespaces {
				egressRules += fmt.Sprintf(`  - to:
    - namespaceSelector:
        matchLabels:
          name: %s
`, ns)
			}
		}
	}

	return fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-network-policy
  namespace: %s
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: %s
  egress:
%s`, config.Name, config.Name, egressRules)
}

// createServiceAccount creates a service account for the namespace
func (tnm *TestNamespaceManager) createServiceAccount(ctx context.Context, cluster *KindCluster, config *TestNamespaceConfig) error {
	tnm.Logger.Infof("Creating service account: %s", config.ServiceAccount.Name)

	// Generate service account YAML
	serviceAccountYAML := tnm.generateServiceAccountYAML(config)

	// Apply service account
	if err := cluster.RunKubectlWithInput(ctx, serviceAccountYAML, "apply", "-f", "-"); err != nil {
		return fmt.Errorf("failed to apply service account: %v", err)
	}

	return nil
}

// generateServiceAccountYAML generates the service account YAML
func (tnm *TestNamespaceManager) generateServiceAccountYAML(config *TestNamespaceConfig) string {
	return fmt.Sprintf(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: %s
  namespace: %s
  labels:
    app: hourglass-test
    test: %s
`, config.ServiceAccount.Name, config.Name, sanitizeTestName(config.Name))
}

// createRoleBinding creates a role and role binding for the namespace
func (tnm *TestNamespaceManager) createRoleBinding(ctx context.Context, cluster *KindCluster, config *TestNamespaceConfig) error {
	tnm.Logger.Infof("Creating role binding for namespace: %s", config.Name)

	// Generate role and role binding YAML
	roleYAML := tnm.generateRoleYAML(config)

	// Apply role and role binding
	if err := cluster.RunKubectlWithInput(ctx, roleYAML, "apply", "-f", "-"); err != nil {
		return fmt.Errorf("failed to apply role binding: %v", err)
	}

	return nil
}

// generateRoleYAML generates the role and role binding YAML
func (tnm *TestNamespaceManager) generateRoleYAML(config *TestNamespaceConfig) string {
	if config.RoleBinding.ClusterRole {
		return fmt.Sprintf(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: %s
rules:
- apiGroups: [""]
  resources: ["pods", "services", "endpoints"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["hourglass.eigenlayer.io"]
  resources: ["performers"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: %s
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: %s
subjects:
- kind: ServiceAccount
  name: %s
  namespace: %s
`, config.RoleBinding.RoleName, config.RoleBinding.RoleName, config.RoleBinding.RoleName, config.RoleBinding.ServiceAccountName, config.Name)
	} else {
		return fmt.Sprintf(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: %s
  namespace: %s
rules:
- apiGroups: [""]
  resources: ["pods", "services", "endpoints"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["hourglass.eigenlayer.io"]
  resources: ["performers"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: %s
  namespace: %s
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: %s
subjects:
- kind: ServiceAccount
  name: %s
  namespace: %s
`, config.RoleBinding.RoleName, config.Name, config.RoleBinding.RoleName, config.Name, config.RoleBinding.RoleName, config.RoleBinding.ServiceAccountName, config.Name)
	}
}

// waitForNamespaceActive waits for the namespace to be active
func (tnm *TestNamespaceManager) waitForNamespaceActive(ctx context.Context, cluster *KindCluster, namespace string, timeout time.Duration) error {
	tnm.Logger.Infof("Waiting for namespace %s to be active", namespace)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for namespace to be active")
		case <-ticker.C:
			output, err := cluster.RunKubectl(ctx, "get", "namespace", namespace, "-o", "jsonpath={.status.phase}")
			if err != nil {
				continue
			}

			if strings.TrimSpace(string(output)) == "Active" {
				tnm.Logger.Infof("Namespace %s is active", namespace)
				return nil
			}
		}
	}
}

// CleanupTestNamespace cleans up the test namespace
func (tnm *TestNamespaceManager) CleanupTestNamespace(ctx context.Context, cluster *KindCluster, config *TestNamespaceConfig) error {
	tnm.Logger.Infof("Cleaning up test namespace: %s", config.Name)

	// Delete namespace (this will cascade delete all resources)
	if err := cluster.DeleteNamespace(ctx, config.Name); err != nil {
		return fmt.Errorf("failed to delete namespace: %v", err)
	}

	// Clean up cluster-level resources if created
	if config.RoleBinding != nil && config.RoleBinding.ClusterRole {
		_, err := cluster.RunKubectl(ctx, "delete", "clusterrole", config.RoleBinding.RoleName, "--ignore-not-found")
		if err != nil {
			tnm.Logger.Warnf("Failed to delete cluster role: %v", err)
		}

		_, err = cluster.RunKubectl(ctx, "delete", "clusterrolebinding", config.RoleBinding.RoleName, "--ignore-not-found")
		if err != nil {
			tnm.Logger.Warnf("Failed to delete cluster role binding: %v", err)
		}
	}

	tnm.Logger.Infof("Test namespace %s cleaned up successfully", config.Name)
	return nil
}

// ListTestNamespaces lists all test namespaces
func (tnm *TestNamespaceManager) ListTestNamespaces(ctx context.Context, cluster *KindCluster) ([]string, error) {
	tnm.Logger.Infof("Listing test namespaces")

	output, err := cluster.RunKubectl(ctx, "get", "namespaces", "-l", "app=hourglass-test", "-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return nil, fmt.Errorf("failed to list test namespaces: %v", err)
	}

	namespaces := strings.Fields(string(output))
	tnm.Logger.Infof("Found %d test namespaces", len(namespaces))
	return namespaces, nil
}

// CleanupAllTestNamespaces cleans up all test namespaces
func (tnm *TestNamespaceManager) CleanupAllTestNamespaces(ctx context.Context, cluster *KindCluster) error {
	tnm.Logger.Infof("Cleaning up all test namespaces")

	namespaces, err := tnm.ListTestNamespaces(ctx, cluster)
	if err != nil {
		return fmt.Errorf("failed to list test namespaces: %v", err)
	}

	for _, namespace := range namespaces {
		if err := cluster.DeleteNamespace(ctx, namespace); err != nil {
			tnm.Logger.Warnf("Failed to delete namespace %s: %v", namespace, err)
		}
	}

	tnm.Logger.Infof("All test namespaces cleaned up")
	return nil
}

// GetNamespaceInfo returns information about a test namespace
func (tnm *TestNamespaceManager) GetNamespaceInfo(ctx context.Context, cluster *KindCluster, namespace string) (map[string]interface{}, error) {
	tnm.Logger.Infof("Getting info for namespace: %s", namespace)

	// Get namespace details
	output, err := cluster.RunKubectl(ctx, "get", "namespace", namespace, "-o", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace info: %v", err)
	}

	// Get resource usage
	resourceOutput, err := cluster.RunKubectl(ctx, "top", "pods", "-n", namespace, "--no-headers")
	if err != nil {
		tnm.Logger.Warnf("Failed to get resource usage: %v", err)
		resourceOutput = []byte("Resource usage not available")
	}

	return map[string]interface{}{
		"namespace_details": string(output),
		"resource_usage":    string(resourceOutput),
	}, nil
}

// ValidateNamespaceSetup validates that the namespace is properly set up
func (tnm *TestNamespaceManager) ValidateNamespaceSetup(ctx context.Context, cluster *KindCluster, config *TestNamespaceConfig) error {
	tnm.Logger.Infof("Validating namespace setup: %s", config.Name)

	// Check namespace exists
	_, err := cluster.RunKubectl(ctx, "get", "namespace", config.Name)
	if err != nil {
		return fmt.Errorf("namespace does not exist: %v", err)
	}

	// Check resource quota if configured
	if config.ResourceQuota != nil {
		_, err := cluster.RunKubectl(ctx, "get", "resourcequota", "test-resource-quota", "-n", config.Name)
		if err != nil {
			return fmt.Errorf("resource quota does not exist: %v", err)
		}
	}

	// Check service account if configured
	if config.ServiceAccount != nil && config.ServiceAccount.Create {
		_, err := cluster.RunKubectl(ctx, "get", "serviceaccount", config.ServiceAccount.Name, "-n", config.Name)
		if err != nil {
			return fmt.Errorf("service account does not exist: %v", err)
		}
	}

	// Check role binding if configured
	if config.RoleBinding != nil {
		if config.RoleBinding.ClusterRole {
			_, err := cluster.RunKubectl(ctx, "get", "clusterrole", config.RoleBinding.RoleName)
			if err != nil {
				return fmt.Errorf("cluster role does not exist: %v", err)
			}
		} else {
			_, err := cluster.RunKubectl(ctx, "get", "role", config.RoleBinding.RoleName, "-n", config.Name)
			if err != nil {
				return fmt.Errorf("role does not exist: %v", err)
			}
		}
	}

	tnm.Logger.Infof("Namespace setup validated successfully: %s", config.Name)
	return nil
}
