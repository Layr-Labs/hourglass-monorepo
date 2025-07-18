package kubernetesManager

import (
	"fmt"
	"time"
)

const (
	// DefaultNamespace is the default Kubernetes namespace
	DefaultNamespace = "default"

	// DefaultOperatorNamespace is the default namespace for the hourglass-operator
	DefaultOperatorNamespace = "hourglass-system"

	// DefaultCRDGroup is the default API group for Performer CRDs
	DefaultCRDGroup = "hourglass.eigenlayer.io"

	// DefaultCRDVersion is the default API version for Performer CRDs
	DefaultCRDVersion = "v1alpha1"

	// DefaultConnectionTimeout is the default timeout for Kubernetes API operations
	DefaultConnectionTimeout = 30 * time.Second

	// DefaultRetryAttempts is the default number of retry attempts
	DefaultRetryAttempts = 3

	// DefaultRetryBackoff is the default backoff duration between retries
	DefaultRetryBackoff = 5 * time.Second

	// DefaultGRPCPort is the default gRPC port for performers
	DefaultGRPCPort = 9090
)

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *Config {
	return &Config{
		Namespace:         DefaultNamespace,
		OperatorNamespace: DefaultOperatorNamespace,
		CRDGroup:          DefaultCRDGroup,
		CRDVersion:        DefaultCRDVersion,
		ConnectionTimeout: DefaultConnectionTimeout,
		RetryAttempts:     DefaultRetryAttempts,
		RetryBackoff:      DefaultRetryBackoff,
	}
}

// Validate validates the configuration and returns an error if invalid
func (c *Config) Validate() error {
	if c.Namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}

	if c.OperatorNamespace == "" {
		return fmt.Errorf("operatorNamespace cannot be empty")
	}

	if c.CRDGroup == "" {
		return fmt.Errorf("crdGroup cannot be empty")
	}

	if c.CRDVersion == "" {
		return fmt.Errorf("crdVersion cannot be empty")
	}

	if c.ConnectionTimeout <= 0 {
		return fmt.Errorf("connectionTimeout must be positive")
	}

	if c.RetryAttempts < 0 {
		return fmt.Errorf("retryAttempts cannot be negative")
	}

	if c.RetryBackoff < 0 {
		return fmt.Errorf("retryBackoff cannot be negative")
	}

	return nil
}

// ApplyDefaults applies default values to unset fields
func (c *Config) ApplyDefaults() {
	if c.Namespace == "" {
		c.Namespace = DefaultNamespace
	}

	if c.OperatorNamespace == "" {
		c.OperatorNamespace = DefaultOperatorNamespace
	}

	if c.CRDGroup == "" {
		c.CRDGroup = DefaultCRDGroup
	}

	if c.CRDVersion == "" {
		c.CRDVersion = DefaultCRDVersion
	}

	if c.ConnectionTimeout == 0 {
		c.ConnectionTimeout = DefaultConnectionTimeout
	}

	if c.RetryAttempts == 0 {
		c.RetryAttempts = DefaultRetryAttempts
	}

	if c.RetryBackoff == 0 {
		c.RetryBackoff = DefaultRetryBackoff
	}
}

// ValidateCreatePerformerRequest validates a CreatePerformerRequest
func ValidateCreatePerformerRequest(req *CreatePerformerRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Name == "" {
		return fmt.Errorf("performer name cannot be empty")
	}

	if req.AVSAddress == "" {
		return fmt.Errorf("AVS address cannot be empty")
	}

	if req.Image == "" {
		return fmt.Errorf("image cannot be empty")
	}

	if req.GRPCPort <= 0 || req.GRPCPort > 65535 {
		return fmt.Errorf("gRPC port must be between 1 and 65535, got %d", req.GRPCPort)
	}

	// Validate resource requirements if provided
	if req.Resources != nil {
		if err := validateResourceRequirements(req.Resources); err != nil {
			return fmt.Errorf("invalid resource requirements: %w", err)
		}
	}

	// Validate hardware requirements if provided
	if req.HardwareRequirements != nil {
		if err := validateHardwareRequirements(req.HardwareRequirements); err != nil {
			return fmt.Errorf("invalid hardware requirements: %w", err)
		}
	}

	return nil
}

// validateResourceRequirements validates resource requirements
func validateResourceRequirements(resources *ResourceRequirements) error {
	// Validate CPU and memory resource formats
	for key, value := range resources.Requests {
		if err := validateResourceValue(key, value); err != nil {
			return fmt.Errorf("invalid request resource %s=%s: %w", key, value, err)
		}
	}

	for key, value := range resources.Limits {
		if err := validateResourceValue(key, value); err != nil {
			return fmt.Errorf("invalid limit resource %s=%s: %w", key, value, err)
		}
	}

	return nil
}

// validateResourceValue validates a resource value format
func validateResourceValue(key, value string) error {
	if value == "" {
		// Return specific error messages for CPU and memory
		switch key {
		case "cpu":
			return fmt.Errorf("CPU value cannot be empty")
		case "memory":
			return fmt.Errorf("memory value cannot be empty")
		default:
			return fmt.Errorf("resource value cannot be empty")
		}
	}

	// Basic validation - in a real implementation, you might want to parse
	// the value according to Kubernetes resource.Quantity format
	return nil
}

// validateHardwareRequirements validates hardware requirements
func validateHardwareRequirements(hw *HardwareRequirementsConfig) error {
	if hw.GPUCount < 0 {
		return fmt.Errorf("GPU count cannot be negative")
	}

	if hw.GPUCount > 0 && hw.GPUType == "" {
		return fmt.Errorf("GPU type must be specified when GPU count > 0")
	}

	if hw.TEERequired && hw.TEEType == "" {
		return fmt.Errorf("TEE type must be specified when TEE is required")
	}

	return nil
}

// ValidateUpdatePerformerRequest validates an UpdatePerformerRequest
func ValidateUpdatePerformerRequest(req *UpdatePerformerRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.PerformerID == "" {
		return fmt.Errorf("performer ID cannot be empty")
	}

	// At least one field must be provided for update
	if req.Image == "" && req.ImageTag == "" && req.Status == "" {
		return fmt.Errorf("at least one field must be provided for update")
	}

	return nil
}
