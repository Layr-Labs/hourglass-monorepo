package kubernetesManager

import (
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	corev1 "k8s.io/api/core/v1"
)

// Config represents the configuration for the Kubernetes manager
type Config struct {
	// Namespace is the Kubernetes namespace to operate in
	Namespace string `json:"namespace" yaml:"namespace"`

	// GenerateNamespace indicates if the namespace name should be generated to override Namespace
	GenerateNamespace bool `json:"generateNamespace,omitempty" yaml:"generateNamespace,omitempty"`

	// KubeconfigPath is the path to the kubeconfig file (optional, defaults to in-cluster)
	KubeconfigPath string `json:"kubeconfigPath,omitempty" yaml:"kubeconfigPath,omitempty"`

	// OperatorNamespace is the namespace where the hourglass-operator is deployed
	OperatorNamespace string `json:"operatorNamespace" yaml:"operatorNamespace"`

	// CRDGroup is the API group for Performer CRDs
	CRDGroup string `json:"crdGroup" yaml:"crdGroup"`

	// CRDVersion is the API version for Performer CRDs
	CRDVersion string `json:"crdVersion" yaml:"crdVersion"`

	// ServiceAccount is the service account to use for operations
	ServiceAccount string `json:"serviceAccount,omitempty" yaml:"serviceAccount,omitempty"`

	// ConnectionTimeout for Kubernetes API operations
	ConnectionTimeout time.Duration `json:"connectionTimeout" yaml:"connectionTimeout"`

	// RetryAttempts for failed operations
	RetryAttempts int `json:"retryAttempts" yaml:"retryAttempts"`

	// RetryBackoff for retry operations
	RetryBackoff time.Duration `json:"retryBackoff" yaml:"retryBackoff"`
}

// Custom environment types removed - using standard k8s corev1.EnvVar instead

// CreatePerformerRequest contains the parameters for creating a new performer
type CreatePerformerRequest struct {
	// Name is the unique name for this performer
	Name string

	// AVSAddress is the address of the AVS this performer belongs to
	AVSAddress string

	// Image is the container image to run
	Image string

	// ImagePullPolicy defines the pull policy for the container image
	ImagePullPolicy string

	// ImageTag is the specific tag/version of the image
	ImageTag string

	// ImageDigest is the digest of the image (optional, for immutable references)
	ImageDigest string

	// GRPCPort is the port the performer will serve gRPC on
	GRPCPort int32

	// Env is the list of environment variables using standard k8s EnvVar type
	Env []corev1.EnvVar

	// Resources specifies the compute resources for the performer
	Resources *ResourceRequirements

	// Scheduling specifies node selection and affinity rules
	Scheduling *SchedulingConfig

	// HardwareRequirements specifies specialized hardware needs
	HardwareRequirements *HardwareRequirementsConfig

	// ServiceAccountName is the name of the ServiceAccount to use for the performer pod
	ServiceAccountName string
}

// CreatePerformerResponse contains the result of creating a performer
type CreatePerformerResponse struct {
	// PerformerID is the unique identifier for the created performer
	PerformerID string

	// Endpoint is the gRPC endpoint for connecting to the performer
	Endpoint string

	// Status is the initial status of the performer
	Status PerformerStatus
}

// ResourceRequirements specifies compute resource requirements
type ResourceRequirements struct {
	// Requests specifies the minimum required resources
	Requests map[string]string `json:"requests,omitempty"`

	// Limits specifies the maximum allowed resources
	Limits map[string]string `json:"limits,omitempty"`
}

// SchedulingConfig specifies node selection and scheduling preferences
type SchedulingConfig struct {
	// NodeSelector is a map of node selector labels
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations allow the performer to tolerate node taints
	Tolerations []TolerationConfig `json:"tolerations,omitempty"`

	// RuntimeClass specifies the container runtime class
	RuntimeClass *string `json:"runtimeClass,omitempty"`

	// PriorityClassName indicates the priority class for scheduling
	PriorityClassName *string `json:"priorityClassName,omitempty"`
}

// TolerationConfig represents a toleration for node taints
type TolerationConfig struct {
	Key      string `json:"key,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
	Effect   string `json:"effect,omitempty"`
}

// HardwareRequirementsConfig specifies specialized hardware needs
type HardwareRequirementsConfig struct {
	// GPUType specifies the required GPU type
	GPUType string `json:"gpuType,omitempty"`

	// GPUCount is the number of GPUs required
	GPUCount int32 `json:"gpuCount,omitempty"`

	// TEERequired indicates if a Trusted Execution Environment is needed
	TEERequired bool `json:"teeRequired,omitempty"`

	// TEEType specifies the TEE technology
	TEEType string `json:"teeType,omitempty"`

	// CustomLabels for matching specialized hardware
	CustomLabels map[string]string `json:"customLabels,omitempty"`
}

// PerformerStatus represents the current status of a performer
type PerformerStatus struct {
	// Phase represents the current performer lifecycle phase
	Phase avsPerformer.PerformerResourceStatus

	// PodName is the name of the associated pod
	PodName string

	// ServiceName is the name of the associated service
	ServiceName string

	// GRPCEndpoint is the DNS name for gRPC connections
	GRPCEndpoint string

	// Ready indicates if the performer is ready to accept tasks
	Ready bool

	// Message contains any status message
	Message string

	// LastUpdated is when the status was last updated
	LastUpdated time.Time
}

// PerformerInfo contains metadata about a performer
type PerformerInfo struct {
	// PerformerID is the unique identifier
	PerformerID string

	// AVSAddress is the address of the AVS
	AVSAddress string

	// Image is the container image
	Image string

	// Version is the image version/tag
	Version string

	// Status is the current status
	Status PerformerStatus

	// CreatedAt is when the performer was created
	CreatedAt time.Time

	// UpdatedAt is when the performer was last updated
	UpdatedAt time.Time
}

// UpdatePerformerRequest contains parameters for updating a performer
type UpdatePerformerRequest struct {
	// PerformerID is the ID of the performer to update
	PerformerID string

	// Image is the new container image (optional)
	Image string

	// ImageTag is the new image tag (optional)
	ImageTag string

	// Status is the new status to set (optional)
	Status avsPerformer.PerformerResourceStatus
}

// PerformerEvent represents an event related to a performer
type PerformerEvent struct {
	// PerformerID is the ID of the performer
	PerformerID string

	// Type is the type of event
	Type string

	// Reason is the reason for the event
	Reason string

	// Message is a human-readable message
	Message string

	// Timestamp is when the event occurred
	Timestamp time.Time
}
