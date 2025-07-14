package kubernetesManager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	assert.Equal(t, DefaultNamespace, config.Namespace)
	assert.Equal(t, DefaultOperatorNamespace, config.OperatorNamespace)
	assert.Equal(t, DefaultCRDGroup, config.CRDGroup)
	assert.Equal(t, DefaultCRDVersion, config.CRDVersion)
	assert.Equal(t, DefaultConnectionTimeout, config.ConnectionTimeout)
	assert.Equal(t, DefaultRetryAttempts, config.RetryAttempts)
	assert.Equal(t, DefaultRetryBackoff, config.RetryBackoff)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config",
			config:      NewDefaultConfig(),
			expectError: false,
		},
		{
			name: "empty namespace",
			config: &Config{
				Namespace:         "",
				OperatorNamespace: "hourglass-system",
				CRDGroup:          "hourglass.eigenlayer.io",
				CRDVersion:        "v1alpha1",
				ConnectionTimeout: 30 * time.Second,
				RetryAttempts:     3,
				RetryBackoff:      5 * time.Second,
			},
			expectError: true,
			errorMsg:    "namespace cannot be empty",
		},
		{
			name: "empty operator namespace",
			config: &Config{
				Namespace:         "default",
				OperatorNamespace: "",
				CRDGroup:          "hourglass.eigenlayer.io",
				CRDVersion:        "v1alpha1",
				ConnectionTimeout: 30 * time.Second,
				RetryAttempts:     3,
				RetryBackoff:      5 * time.Second,
			},
			expectError: true,
			errorMsg:    "operatorNamespace cannot be empty",
		},
		{
			name: "empty CRD group",
			config: &Config{
				Namespace:         "default",
				OperatorNamespace: "hourglass-system",
				CRDGroup:          "",
				CRDVersion:        "v1alpha1",
				ConnectionTimeout: 30 * time.Second,
				RetryAttempts:     3,
				RetryBackoff:      5 * time.Second,
			},
			expectError: true,
			errorMsg:    "crdGroup cannot be empty",
		},
		{
			name: "empty CRD version",
			config: &Config{
				Namespace:         "default",
				OperatorNamespace: "hourglass-system",
				CRDGroup:          "hourglass.eigenlayer.io",
				CRDVersion:        "",
				ConnectionTimeout: 30 * time.Second,
				RetryAttempts:     3,
				RetryBackoff:      5 * time.Second,
			},
			expectError: true,
			errorMsg:    "crdVersion cannot be empty",
		},
		{
			name: "negative connection timeout",
			config: &Config{
				Namespace:         "default",
				OperatorNamespace: "hourglass-system",
				CRDGroup:          "hourglass.eigenlayer.io",
				CRDVersion:        "v1alpha1",
				ConnectionTimeout: -1 * time.Second,
				RetryAttempts:     3,
				RetryBackoff:      5 * time.Second,
			},
			expectError: true,
			errorMsg:    "connectionTimeout must be positive",
		},
		{
			name: "negative retry attempts",
			config: &Config{
				Namespace:         "default",
				OperatorNamespace: "hourglass-system",
				CRDGroup:          "hourglass.eigenlayer.io",
				CRDVersion:        "v1alpha1",
				ConnectionTimeout: 30 * time.Second,
				RetryAttempts:     -1,
				RetryBackoff:      5 * time.Second,
			},
			expectError: true,
			errorMsg:    "retryAttempts cannot be negative",
		},
		{
			name: "negative retry backoff",
			config: &Config{
				Namespace:         "default",
				OperatorNamespace: "hourglass-system",
				CRDGroup:          "hourglass.eigenlayer.io",
				CRDVersion:        "v1alpha1",
				ConnectionTimeout: 30 * time.Second,
				RetryAttempts:     3,
				RetryBackoff:      -1 * time.Second,
			},
			expectError: true,
			errorMsg:    "retryBackoff cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_ApplyDefaults(t *testing.T) {
	config := &Config{}
	config.ApplyDefaults()

	assert.Equal(t, DefaultNamespace, config.Namespace)
	assert.Equal(t, DefaultOperatorNamespace, config.OperatorNamespace)
	assert.Equal(t, DefaultCRDGroup, config.CRDGroup)
	assert.Equal(t, DefaultCRDVersion, config.CRDVersion)
	assert.Equal(t, DefaultConnectionTimeout, config.ConnectionTimeout)
	assert.Equal(t, DefaultRetryAttempts, config.RetryAttempts)
	assert.Equal(t, DefaultRetryBackoff, config.RetryBackoff)
}

func TestConfig_ApplyDefaults_PreservesExistingValues(t *testing.T) {
	config := &Config{
		Namespace:         "custom-namespace",
		OperatorNamespace: "custom-operator-namespace",
		CRDGroup:          "custom.group.io",
		CRDVersion:        "v2alpha1",
		ConnectionTimeout: 60 * time.Second,
		RetryAttempts:     5,
		RetryBackoff:      10 * time.Second,
	}
	
	config.ApplyDefaults()

	// Verify existing values are preserved
	assert.Equal(t, "custom-namespace", config.Namespace)
	assert.Equal(t, "custom-operator-namespace", config.OperatorNamespace)
	assert.Equal(t, "custom.group.io", config.CRDGroup)
	assert.Equal(t, "v2alpha1", config.CRDVersion)
	assert.Equal(t, 60*time.Second, config.ConnectionTimeout)
	assert.Equal(t, 5, config.RetryAttempts)
	assert.Equal(t, 10*time.Second, config.RetryBackoff)
}

func TestValidateCreatePerformerRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     *CreatePerformerRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			request: &CreatePerformerRequest{
				Name:       "test-performer",
				AVSAddress: "0x123",
				Image:      "test-image:latest",
				GRPCPort:   9090,
			},
			expectError: false,
		},
		{
			name:        "nil request",
			request:     nil,
			expectError: true,
			errorMsg:    "request cannot be nil",
		},
		{
			name: "empty name",
			request: &CreatePerformerRequest{
				AVSAddress: "0x123",
				Image:      "test-image:latest",
				GRPCPort:   9090,
			},
			expectError: true,
			errorMsg:    "performer name cannot be empty",
		},
		{
			name: "empty AVS address",
			request: &CreatePerformerRequest{
				Name:     "test-performer",
				Image:    "test-image:latest",
				GRPCPort: 9090,
			},
			expectError: true,
			errorMsg:    "AVS address cannot be empty",
		},
		{
			name: "empty image",
			request: &CreatePerformerRequest{
				Name:       "test-performer",
				AVSAddress: "0x123",
				GRPCPort:   9090,
			},
			expectError: true,
			errorMsg:    "image cannot be empty",
		},
		{
			name: "invalid gRPC port - zero",
			request: &CreatePerformerRequest{
				Name:       "test-performer",
				AVSAddress: "0x123",
				Image:      "test-image:latest",
				GRPCPort:   0,
			},
			expectError: true,
			errorMsg:    "gRPC port must be between 1 and 65535",
		},
		{
			name: "invalid gRPC port - too high",
			request: &CreatePerformerRequest{
				Name:       "test-performer",
				AVSAddress: "0x123",
				Image:      "test-image:latest",
				GRPCPort:   70000,
			},
			expectError: true,
			errorMsg:    "gRPC port must be between 1 and 65535",
		},
		{
			name: "invalid resource requirements",
			request: &CreatePerformerRequest{
				Name:       "test-performer",
				AVSAddress: "0x123",
				Image:      "test-image:latest",
				GRPCPort:   9090,
				Resources: &ResourceRequirements{
					Requests: map[string]string{
						"cpu": "", // empty value
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid resource requirements",
		},
		{
			name: "invalid hardware requirements",
			request: &CreatePerformerRequest{
				Name:       "test-performer",
				AVSAddress: "0x123",
				Image:      "test-image:latest",
				GRPCPort:   9090,
				HardwareRequirements: &HardwareRequirementsConfig{
					GPUCount: 1,
					GPUType:  "", // missing GPU type when count > 0
				},
			},
			expectError: true,
			errorMsg:    "invalid hardware requirements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreatePerformerRequest(tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateResourceRequirements(t *testing.T) {
	tests := []struct {
		name        string
		resources   *ResourceRequirements
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid resources",
			resources: &ResourceRequirements{
				Requests: map[string]string{
					"cpu":    "100m",
					"memory": "128Mi",
				},
				Limits: map[string]string{
					"cpu":    "500m",
					"memory": "512Mi",
				},
			},
			expectError: false,
		},
		{
			name: "empty request value",
			resources: &ResourceRequirements{
				Requests: map[string]string{
					"cpu": "",
				},
			},
			expectError: true,
			errorMsg:    "resource value cannot be empty",
		},
		{
			name: "empty limit value",
			resources: &ResourceRequirements{
				Limits: map[string]string{
					"memory": "",
				},
			},
			expectError: true,
			errorMsg:    "resource value cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceRequirements(tt.resources)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateResourceValue(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid cpu value",
			key:         "cpu",
			value:       "100m",
			expectError: false,
		},
		{
			name:        "valid memory value",
			key:         "memory",
			value:       "128Mi",
			expectError: false,
		},
		{
			name:        "empty value",
			key:         "cpu",
			value:       "",
			expectError: true,
			errorMsg:    "resource value cannot be empty",
		},
		{
			name:        "empty cpu value",
			key:         "cpu",
			value:       "",
			expectError: true,
			errorMsg:    "CPU value cannot be empty",
		},
		{
			name:        "empty memory value",
			key:         "memory",
			value:       "",
			expectError: true,
			errorMsg:    "memory value cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceValue(tt.key, tt.value)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateHardwareRequirements(t *testing.T) {
	tests := []struct {
		name        string
		hardware    *HardwareRequirementsConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid hardware requirements",
			hardware: &HardwareRequirementsConfig{
				GPUType:  "nvidia-tesla-v100",
				GPUCount: 1,
			},
			expectError: false,
		},
		{
			name: "no GPU requirements",
			hardware: &HardwareRequirementsConfig{
				GPUCount: 0,
			},
			expectError: false,
		},
		{
			name: "negative GPU count",
			hardware: &HardwareRequirementsConfig{
				GPUCount: -1,
			},
			expectError: true,
			errorMsg:    "GPU count cannot be negative",
		},
		{
			name: "GPU count without type",
			hardware: &HardwareRequirementsConfig{
				GPUCount: 1,
				GPUType:  "",
			},
			expectError: true,
			errorMsg:    "GPU type must be specified when GPU count > 0",
		},
		{
			name: "TEE required without type",
			hardware: &HardwareRequirementsConfig{
				TEERequired: true,
				TEEType:     "",
			},
			expectError: true,
			errorMsg:    "TEE type must be specified when TEE is required",
		},
		{
			name: "valid TEE requirements",
			hardware: &HardwareRequirementsConfig{
				TEERequired: true,
				TEEType:     "intel-sgx",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHardwareRequirements(tt.hardware)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUpdatePerformerRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     *UpdatePerformerRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid update request",
			request: &UpdatePerformerRequest{
				PerformerID: "test-performer",
				Image:       "new-image:v2.0.0",
			},
			expectError: false,
		},
		{
			name:        "nil request",
			request:     nil,
			expectError: true,
			errorMsg:    "request cannot be nil",
		},
		{
			name: "empty performer ID",
			request: &UpdatePerformerRequest{
				Image: "new-image:v2.0.0",
			},
			expectError: true,
			errorMsg:    "performer ID cannot be empty",
		},
		{
			name: "no fields to update",
			request: &UpdatePerformerRequest{
				PerformerID: "test-performer",
			},
			expectError: true,
			errorMsg:    "at least one field must be provided for update",
		},
		{
			name: "update with image tag only",
			request: &UpdatePerformerRequest{
				PerformerID: "test-performer",
				ImageTag:    "v2.0.0",
			},
			expectError: false,
		},
		{
			name: "update with status only",
			request: &UpdatePerformerRequest{
				PerformerID: "test-performer",
				Status:      "Running",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdatePerformerRequest(tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}