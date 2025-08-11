package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SchedulingConfig defines advanced scheduling requirements
// +kubebuilder:object:generate=true
type SchedulingConfig struct {
	// NodeSelector is a map of node selector labels for scheduling
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations allow the performer to tolerate node taints
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity defines node affinity scheduling rules
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// RuntimeClass specifies the container runtime class
	RuntimeClass *string `json:"runtimeClass,omitempty"`

	// PriorityClassName indicates the priority class for scheduling
	PriorityClassName *string `json:"priorityClassName,omitempty"`
}

// HardwareRequirements defines specialized hardware needs
// +kubebuilder:object:generate=true
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
// +kubebuilder:object:generate=true
type PerformerSpec struct {
	// AVSAddress is the unique identifier for this AVS
	// +kubebuilder:validation:Required
	AVSAddress string `json:"avsAddress"`

	// Image is the AVS performer container image
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// ImagePullPolicy defines the pull policy for the container image
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	// +kubebuilder:default="IfNotPresent"
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Version is the container image version for upgrade tracking
	Version string `json:"version,omitempty"`

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

// EnvVarSource represents a source for the value of an environment variable
// +kubebuilder:object:generate=true
type EnvVarSource struct {
	// Name of the environment variable
	Name string `json:"name"`

	// ValueFrom specifies the source of the environment variable value
	ValueFrom *EnvValueFrom `json:"valueFrom,omitempty"`
}

// EnvValueFrom describes a source for the value of an environment variable
// +kubebuilder:object:generate=true
type EnvValueFrom struct {
	// SecretKeyRef selects a key of a secret in the pod's namespace
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`

	// ConfigMapKeyRef selects a key of a config map in the pod's namespace
	ConfigMapKeyRef *corev1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
}

// PerformerConfig contains configuration for the performer
// +kubebuilder:object:generate=true
type PerformerConfig struct {
	// GRPCPort is the port on which the performer serves gRPC requests
	// +kubebuilder:default=9090
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	GRPCPort int32 `json:"grpcPort,omitempty"`

	// Environment variables for the performer container
	Environment map[string]string `json:"environment,omitempty"`

	// EnvironmentFrom variables for the performer container (references to secrets/configmaps)
	EnvironmentFrom []EnvVarSource `json:"environmentFrom,omitempty"`

	// Args are additional command line arguments for the performer
	Args []string `json:"args,omitempty"`

	// Command overrides the default container entrypoint
	Command []string `json:"command,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use for the performer pod
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// PerformerStatus defines the observed state of Performer
// +kubebuilder:object:generate=true
type PerformerStatus struct {
	// Phase represents the current performer lifecycle phase
	// +kubebuilder:validation:Enum=Pending;Running;Upgrading;Terminating;Failed
	Phase string `json:"phase,omitempty"`

	// Ready indicates if the performer is ready to accept requests
	Ready bool `json:"ready,omitempty"`

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
