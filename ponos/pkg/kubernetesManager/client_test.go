package kubernetesManager

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientWrapper(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name: "invalid config - negative timeout",
			config: &Config{
				Namespace:         "test",
				OperatorNamespace: "test",
				CRDGroup:          "test",
				CRDVersion:        "test",
				ConnectionTimeout: -1, // negative timeout should fail validation
			},
			expectError: true,
			errorMsg:    "invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClientWrapper(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestNewClientWrapper_WithValidConfig(t *testing.T) {
	// Skip this test if we don't have a valid Kubernetes environment
	// This test would require either in-cluster config or a valid kubeconfig file
	t.Skip("Skipping Kubernetes client test - requires valid Kubernetes environment")

	config := NewDefaultConfig()
	config.Namespace = "test-namespace"

	client, err := NewClientWrapper(config)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Kubernetes)
	assert.NotNil(t, client.CRDClient)
	assert.NotNil(t, client.RestConfig)
	assert.Equal(t, config, client.Config)
}

func TestGetDefaultKubeconfigPath(t *testing.T) {
	// Save original environment
	originalKubeconfig := os.Getenv("KUBECONFIG")
	originalHome := os.Getenv("HOME")
	
	defer func() {
		// Restore original environment
		if originalKubeconfig != "" {
			os.Setenv("KUBECONFIG", originalKubeconfig)
		} else {
			os.Unsetenv("KUBECONFIG")
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	t.Run("with KUBECONFIG environment variable", func(t *testing.T) {
		expectedPath := "/custom/kubeconfig/path"
		os.Setenv("KUBECONFIG", expectedPath)
		
		result := getDefaultKubeconfigPath()
		assert.Equal(t, expectedPath, result)
	})

	t.Run("with default home directory path", func(t *testing.T) {
		os.Unsetenv("KUBECONFIG")
		
		// Create a temporary directory to simulate home
		tempDir, err := os.MkdirTemp("", "test-home")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)
		
		// Create .kube/config file
		kubeDir := filepath.Join(tempDir, ".kube")
		err = os.MkdirAll(kubeDir, 0755)
		require.NoError(t, err)
		
		configPath := filepath.Join(kubeDir, "config")
		err = os.WriteFile(configPath, []byte("test-config"), 0644)
		require.NoError(t, err)
		
		os.Setenv("HOME", tempDir)
		
		result := getDefaultKubeconfigPath()
		assert.Equal(t, configPath, result)
	})

	t.Run("no kubeconfig found", func(t *testing.T) {
		os.Unsetenv("KUBECONFIG")
		
		// Set HOME to a directory without .kube/config
		tempDir, err := os.MkdirTemp("", "test-home-empty")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)
		
		os.Setenv("HOME", tempDir)
		
		result := getDefaultKubeconfigPath()
		assert.Empty(t, result)
	})

	t.Run("no home directory", func(t *testing.T) {
		os.Unsetenv("KUBECONFIG")
		os.Unsetenv("HOME")
		
		result := getDefaultKubeconfigPath()
		assert.Empty(t, result)
	})
}

func TestClientWrapper_TestConnection(t *testing.T) {
	// Skip this test if we don't have a valid Kubernetes environment
	t.Skip("Skipping Kubernetes connection test - requires valid Kubernetes environment")

	config := NewDefaultConfig()
	client, err := NewClientWrapper(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.TestConnection(ctx)

	// This test would only pass in a real Kubernetes environment
	// In a unit test environment, we expect it to fail
	assert.Error(t, err)
}

func TestClientWrapper_Getters(t *testing.T) {
	config := &Config{
		Namespace:         "test-namespace",
		OperatorNamespace: "test-operator-namespace",
		CRDGroup:          "test.group.io",
		CRDVersion:        "v1beta1",
		ConnectionTimeout: 60 * time.Second,
	}

	// Create a mock client wrapper for testing getters
	client := &ClientWrapper{
		Config: config,
	}

	assert.Equal(t, "test-namespace", client.GetNamespace())
	assert.Equal(t, "test-operator-namespace", client.GetOperatorNamespace())
	assert.Equal(t, "test.group.io", client.GetCRDGroup())
	assert.Equal(t, "v1beta1", client.GetCRDVersion())
}

func TestClientWrapper_Close(t *testing.T) {
	client := &ClientWrapper{}
	
	err := client.Close()
	assert.NoError(t, err)
}

func TestGetKubernetesConfig_Scenarios(t *testing.T) {
	tests := []struct {
		name           string
		kubeconfigPath string
		description    string
	}{
		{
			name:           "with explicit kubeconfig path",
			kubeconfigPath: "/custom/kubeconfig",
			description:    "should attempt to use provided kubeconfig path",
		},
		{
			name:           "with empty kubeconfig path",
			kubeconfigPath: "",
			description:    "should try in-cluster config first, then default kubeconfig",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test just verifies the function can be called without panicking
			// Actual Kubernetes configuration testing would require a real environment
			_, err := getKubernetesConfig(tt.kubeconfigPath)
			
			// We may or may not get an error depending on environment
			// Both outcomes are acceptable in unit tests
			t.Logf("getKubernetesConfig with path '%s': %v (%s)", tt.kubeconfigPath, err, tt.description)
		})
	}
}

func TestConfig_Integration(t *testing.T) {
	t.Run("config flow with defaults", func(t *testing.T) {
		config := NewDefaultConfig()
		config.Namespace = "custom-namespace"
		
		// Validate the config
		err := config.Validate()
		assert.NoError(t, err)
		
		// Verify values
		assert.Equal(t, "custom-namespace", config.Namespace)
		assert.Equal(t, DefaultOperatorNamespace, config.OperatorNamespace)
		assert.Equal(t, DefaultCRDGroup, config.CRDGroup)
		assert.Equal(t, DefaultCRDVersion, config.CRDVersion)
	})

	t.Run("config flow with apply defaults", func(t *testing.T) {
		config := &Config{
			Namespace: "test-namespace",
		}
		
		config.ApplyDefaults()
		err := config.Validate()
		
		assert.NoError(t, err)
		assert.Equal(t, "test-namespace", config.Namespace)
		assert.Equal(t, DefaultOperatorNamespace, config.OperatorNamespace)
	})
}