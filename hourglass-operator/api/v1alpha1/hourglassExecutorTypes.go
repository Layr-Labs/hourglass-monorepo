package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ChainConfig defines configuration for a blockchain network
type ChainConfig struct {
	// Name is the identifier for this chain (e.g., "ethereum", "base")
	Name string `json:"name"`
	
	// RPC endpoint for this chain
	RPC string `json:"rpc"`
	
	// ChainID for this blockchain network
	ChainID int64 `json:"chainId"`
	
	// TaskMailbox contract address on this chain
	TaskMailboxAddress string `json:"taskMailboxAddress"`
}

// HourglassExecutorSpec defines the desired state of HourglassExecutor
type HourglassExecutorSpec struct {
	// Image is the Hourglass executor container image
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Replicas is the number of executor instances to run
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas,omitempty"`

	// Config contains executor-specific configuration
	Config HourglassExecutorConfig `json:"config"`

	// Resources defines compute resource requirements
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// NodeSelector constrains scheduling to nodes with matching labels
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations allows scheduling on nodes with matching taints
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// ImagePullSecrets for private container registries
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// HourglassExecutorConfig contains configuration for the executor
type HourglassExecutorConfig struct {
	// AggregatorEndpoint is the gRPC endpoint of the aggregator
	AggregatorEndpoint string `json:"aggregatorEndpoint"`

	// OperatorKeys contains BLS and ECDSA private keys for signing
	OperatorKeys map[string]string `json:"operatorKeys"`

	// Chains defines the blockchain networks this executor monitors
	Chains []ChainConfig `json:"chains"`

	// PerformerMode determines how performers are deployed ("docker" or "kubernetes")
	// +kubebuilder:default="kubernetes"
	PerformerMode string `json:"performerMode,omitempty"`

	// Kubernetes contains settings for Kubernetes performer deployment
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty"`

	// LogLevel sets the logging verbosity
	// +kubebuilder:default="info"
	LogLevel string `json:"logLevel,omitempty"`
}

// KubernetesConfig contains Kubernetes-specific settings
type KubernetesConfig struct {
	// Namespace where performers will be deployed
	// +kubebuilder:default="default"
	Namespace string `json:"namespace,omitempty"`

	// DefaultScheduling provides default scheduling constraints for performers
	DefaultScheduling *SchedulingConfig `json:"defaultScheduling,omitempty"`
}

// SchedulingConfig defines scheduling constraints
type SchedulingConfig struct {
	// NodeSelector constrains scheduling to nodes with matching labels
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// NodeAffinity defines advanced node selection rules
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`

	// Tolerations allows scheduling on nodes with matching taints
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// RuntimeClass specifies the container runtime
	RuntimeClass *string `json:"runtimeClass,omitempty"`
}

// HourglassExecutorStatus defines the observed state of HourglassExecutor
type HourglassExecutorStatus struct {
	// Phase represents the current deployment phase
	Phase string `json:"phase,omitempty"`

	// Replicas is the number of running executor replicas
	Replicas int32 `json:"replicas"`

	// ReadyReplicas is the number of ready executor replicas
	ReadyReplicas int32 `json:"readyReplicas"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastConfigUpdate tracks when configuration was last applied
	LastConfigUpdate *metav1.Time `json:"lastConfigUpdate,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
//+kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
//+kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
//+kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// HourglassExecutor is the Schema for the hourglassexecutors API
type HourglassExecutor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HourglassExecutorSpec   `json:"spec,omitempty"`
	Status HourglassExecutorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HourglassExecutorList contains a list of HourglassExecutor
type HourglassExecutorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HourglassExecutor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HourglassExecutor{}, &HourglassExecutorList{})
}