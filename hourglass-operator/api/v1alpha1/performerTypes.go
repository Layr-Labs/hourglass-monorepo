package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HardwareRequirements defines specialized hardware needs
type HardwareRequirements struct {
	// GPUType specifies the required GPU type (e.g., "nvidia-tesla-v100", "nvidia-a100")
	GPUType string `json:"gpuType,omitempty"`

	// GPUCount is the number of GPUs required
	// +kubebuilder:validation:Minimum=0
	GPUCount int32 `json:"gpuCount,omitempty"`

	// TEERequired indicates if a Trusted Execution Environment is needed
	TEERequired bool `json:"teeRequired,omitempty"`

	// TEEType specifies the TEE technology (e.g., "sgx", "sev", "tdx")
	TEEType string `json:"teeType,omitempty"`

	// CustomLabels for matching specialized hardware
	CustomLabels map[string]string `json:"customLabels,omitempty"`
}

// PerformerSpec defines the desired state of Performer
type PerformerSpec struct {
	// AVSAddress is the unique identifier for this AVS
	// +kubebuilder:validation:Required
	AVSAddress string `json:"avsAddress"`

	// Image is the AVS performer container image
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Version is the container image version for upgrade tracking
	Version string `json:"version,omitempty"`

	// ExecutorRef is a reference to the parent HourglassExecutor
	ExecutorRef string `json:"executorRef,omitempty"`

	// Config contains performer-specific configuration
	Config PerformerConfig `json:"config,omitempty"`

	// Resources defines compute resource requirements
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Scheduling defines advanced scheduling requirements
	Scheduling *SchedulingConfig `json:"scheduling,omitempty"`

	// HardwareRequirements specifies specialized hardware needs
	HardwareRequirements *HardwareRequirements `json:"hardwareRequirements,omitempty"`

	// ImagePullSecrets for private container registries
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// PerformerConfig contains configuration for the performer
type PerformerConfig struct {
	// GRPCPort is the port on which the performer serves gRPC requests
	// +kubebuilder:default=9090
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	GRPCPort int32 `json:"grpcPort,omitempty"`

	// Environment variables for the performer container
	Environment map[string]string `json:"environment,omitempty"`

	// Args are additional command line arguments for the performer
	Args []string `json:"args,omitempty"`

	// Command overrides the default container entrypoint
	Command []string `json:"command,omitempty"`
}

// PerformerStatus defines the observed state of Performer
type PerformerStatus struct {
	// Phase represents the current performer lifecycle phase
	// +kubebuilder:validation:Enum=Pending;Running;Upgrading;Terminating;Failed
	Phase string `json:"phase,omitempty"`

	// PodName is the name of the associated pod
	PodName string `json:"podName,omitempty"`

	// ServiceName is the name of the associated service
	ServiceName string `json:"serviceName,omitempty"`

	// GRPCEndpoint is the DNS name for gRPC connections
	GRPCEndpoint string `json:"grpcEndpoint,omitempty"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastUpgrade tracks the most recent upgrade operation
	LastUpgrade *metav1.Time `json:"lastUpgrade,omitempty"`

	// ReadyTime indicates when the performer became ready
	ReadyTime *metav1.Time `json:"readyTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="AVS",type=string,JSONPath=`.spec.avsAddress`
//+kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.status.grpcEndpoint`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Performer is the Schema for the performers API
type Performer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PerformerSpec   `json:"spec,omitempty"`
	Status PerformerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PerformerList contains a list of Performer
type PerformerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Performer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Performer{}, &PerformerList{})
}
