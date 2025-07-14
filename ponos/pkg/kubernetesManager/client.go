package kubernetesManager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// ClientWrapper wraps both the standard Kubernetes client and the controller-runtime client
type ClientWrapper struct {
	// Kubernetes is the standard Kubernetes clientset
	Kubernetes kubernetes.Interface

	// CRDClient is the controller-runtime client for CRD operations
	CRDClient client.Client

	// RestConfig is the Kubernetes REST config
	RestConfig *rest.Config

	// Config is the kubernetesManager configuration
	Config *Config
}

// NewClientWrapper creates a new Kubernetes client wrapper
func NewClientWrapper(cfg *Config) (*ClientWrapper, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Apply defaults and validate config
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Get Kubernetes REST config
	restConfig, err := getKubernetesConfig(cfg.KubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	// Set timeout
	restConfig.Timeout = cfg.ConnectionTimeout

	// Create Kubernetes clientset
	kubernetesClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create controller-runtime client for CRD operations
	crdClient, err := client.New(restConfig, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create CRD client: %w", err)
	}

	return &ClientWrapper{
		Kubernetes: kubernetesClient,
		CRDClient:  crdClient,
		RestConfig: restConfig,
		Config:     cfg,
	}, nil
}

// getKubernetesConfig gets the Kubernetes configuration
func getKubernetesConfig(kubeconfigPath string) (*rest.Config, error) {
	// If kubeconfigPath is provided, use it
	if kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	// Try in-cluster config first
	if restConfig, err := rest.InClusterConfig(); err == nil {
		return restConfig, nil
	}

	// Fall back to kubeconfig file
	kubeconfigPath = getDefaultKubeconfigPath()
	if kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	// Try controller-runtime config as last resort
	return config.GetConfig()
}

// getDefaultKubeconfigPath gets the default kubeconfig path
func getDefaultKubeconfigPath() string {
	// Check KUBECONFIG environment variable
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	// Check default location in home directory
	if home := homedir.HomeDir(); home != "" {
		kubeconfigPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeconfigPath); err == nil {
			return kubeconfigPath
		}
	}

	return ""
}

// TestConnection tests the connection to the Kubernetes cluster
func (c *ClientWrapper) TestConnection(ctx context.Context) error {
	// Test standard Kubernetes client
	_, err := c.Kubernetes.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to Kubernetes API: %w", err)
	}

	// Test namespace access
	_, err = c.Kubernetes.CoreV1().Namespaces().Get(ctx, c.Config.Namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to access namespace %s: %w", c.Config.Namespace, err)
	}

	return nil
}

// GetNamespace returns the configured namespace
func (c *ClientWrapper) GetNamespace() string {
	return c.Config.Namespace
}

// GetOperatorNamespace returns the configured operator namespace
func (c *ClientWrapper) GetOperatorNamespace() string {
	return c.Config.OperatorNamespace
}

// GetCRDGroup returns the configured CRD group
func (c *ClientWrapper) GetCRDGroup() string {
	return c.Config.CRDGroup
}

// GetCRDVersion returns the configured CRD version
func (c *ClientWrapper) GetCRDVersion() string {
	return c.Config.CRDVersion
}

// Close closes the client connections (if needed in the future)
func (c *ClientWrapper) Close() error {
	// Currently, there's nothing to close for these clients
	// This method is here for future extensibility
	return nil
}