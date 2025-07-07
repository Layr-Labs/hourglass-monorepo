package types

import (
	"time"
)

// AVS represents an Actively Validated Service
type AVS struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Operator        string            `json:"operator"`
	ComputeSpec     ComputeSpec       `json:"compute_spec"`
	AttestationReq  AttestationRequirements `json:"attestation_requirements"`
	Status          AVSStatus         `json:"status"`
}

// ComputeSpec defines compute requirements for AVS
type ComputeSpec struct {
	Replicas        int               `json:"replicas"`
	Resources       ResourceRequirements `json:"resources"`
	NodeSelector    map[string]string `json:"node_selector"`
}

// ResourceRequirements for compute
type ResourceRequirements struct {
	CPU             string            `json:"cpu"`
	Memory          string            `json:"memory"`
	TEEType         string            `json:"tee_type"` // "SEV-SNP", "TDX", "SGX"
}

// AttestationRequirements for verification
type AttestationRequirements struct {
	MeasurementPolicy   string          `json:"measurement_policy"`
	MinTCBVersion      string          `json:"min_tcb_version"`
	AllowedMeasurements []string        `json:"allowed_measurements"`
	MaxAttestationAge   time.Duration   `json:"max_attestation_age"`
}

// AVSStatus represents current state
type AVSStatus struct {
	Phase           string            `json:"phase"`
	ReadyReplicas   int               `json:"ready_replicas"`
	Attestations    []AttestationInfo `json:"attestations"`
	LastUpdated     time.Time         `json:"last_updated"`
}

// AttestationInfo for tracking attestations
type AttestationInfo struct {
	InstanceID      string            `json:"instance_id"`
	Measurement     string            `json:"measurement"`
	Timestamp       time.Time         `json:"timestamp"`
	Valid           bool              `json:"valid"`
}