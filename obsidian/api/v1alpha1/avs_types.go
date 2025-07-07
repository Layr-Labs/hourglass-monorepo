package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AVSSpec defines the desired state of AVS
type AVSSpec struct {
	// Operator managing this AVS
	Operator string `json:"operator"`

	// Service configuration
	ServiceImage string `json:"serviceImage"`
	Replicas     int32  `json:"replicas"`

	// Compute requirements
	ComputeRequirements ComputeRequirements `json:"computeRequirements"`

	// Attestation policy
	AttestationPolicy AttestationPolicy `json:"attestationPolicy"`

	// Networking
	ServicePort int32 `json:"servicePort,omitempty"`
}

// ComputeRequirements for the AVS
type ComputeRequirements struct {
	CPU         string            `json:"cpu"`
	Memory      string            `json:"memory"`
	TEEType     string            `json:"teeType"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// AttestationPolicy defines verification requirements
type AttestationPolicy struct {
	AllowedMeasurements []string `json:"allowedMeasurements"`
	MaxAttestationAge   string   `json:"maxAttestationAge"`
	RequireSEV          bool     `json:"requireSEV"`
}

// AVSStatus defines the observed state of AVS
type AVSStatus struct {
	Phase          string              `json:"phase"`
	ReadyReplicas  int32               `json:"readyReplicas"`
	TotalReplicas  int32               `json:"totalReplicas"`
	Attestations   []AttestationStatus `json:"attestations,omitempty"`
	LastUpdated    metav1.Time         `json:"lastUpdated"`
}

// AttestationStatus for each instance
type AttestationStatus struct {
	PodName     string      `json:"podName"`
	InstanceID  string      `json:"instanceId"`
	Measurement string      `json:"measurement"`
	Valid       bool        `json:"valid"`
	LastChecked metav1.Time `json:"lastChecked"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=avs
// +kubebuilder:printcolumn:name="Operator",type=string,JSONPath=`.spec.operator`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// AVS is the Schema for the avs API
type AVS struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AVSSpec   `json:"spec,omitempty"`
	Status AVSStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AVSList contains a list of AVS
type AVSList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AVS `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AVS{}, &AVSList{})
}