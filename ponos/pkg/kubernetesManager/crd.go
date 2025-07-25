package kubernetesManager

import (
	"context"
	"fmt"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PerformerCRD represents the Performer Custom Resource Definition
// This mirrors the structure defined in hourglass-operator/api/v1alpha1/performerTypes.go
type PerformerCRD struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PerformerSpec      `json:"spec,omitempty"`
	Status PerformerStatusCRD `json:"status,omitempty"`
}

// GetObjectKind returns the object kind for the PerformerCRD
func (p *PerformerCRD) GetObjectKind() schema.ObjectKind {
	return &p.TypeMeta
}

// PerformerSpec defines the desired state of Performer
type PerformerSpec struct {
	// AVSAddress is the unique identifier for this AVS
	AVSAddress string `json:"avsAddress"`

	// Image is the AVS performer container image
	Image string `json:"image"`

	// ImagePullPolicy defines the pull policy for the container image
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
	HardwareRequirements *HardwareRequirementsConfig `json:"hardwareRequirements,omitempty"`

	// ImagePullSecrets for private container registries
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// PerformerConfig contains configuration for the performer
type PerformerConfig struct {
	// GRPCPort is the port on which the performer serves gRPC requests
	GRPCPort int32 `json:"grpcPort,omitempty"`

	// Environment variables for the performer container
	Environment map[string]string `json:"environment,omitempty"`

	// Args are additional command line arguments for the performer
	Args []string `json:"args,omitempty"`

	// Command overrides the default container entrypoint
	Command []string `json:"command,omitempty"`
}

// PerformerStatusCRD defines the observed state of Performer (from CRD)
type PerformerStatusCRD struct {
	// Phase represents the current performer lifecycle phase
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

// PerformerList contains a list of Performer
type PerformerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PerformerCRD `json:"items"`
}

// DeepCopyObject implements runtime.Object interface
func (p *PerformerCRD) DeepCopyObject() runtime.Object {
	return p.DeepCopy()
}

// DeepCopy creates a deep copy of PerformerCRD
func (p *PerformerCRD) DeepCopy() *PerformerCRD {
	if p == nil {
		return nil
	}
	out := new(PerformerCRD)
	p.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (p *PerformerCRD) DeepCopyInto(out *PerformerCRD) {
	*out = *p
	out.TypeMeta = p.TypeMeta
	p.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	p.Spec.DeepCopyInto(&out.Spec)
	p.Status.DeepCopyInto(&out.Status)
}

// DeepCopyInto for PerformerSpec
func (ps *PerformerSpec) DeepCopyInto(out *PerformerSpec) {
	*out = *ps
	ps.Resources.DeepCopyInto(&out.Resources)
	if ps.Scheduling != nil {
		in, out := ps.Scheduling, &out.Scheduling
		*out = new(SchedulingConfig)
		(*in).DeepCopyInto(*out)
	}
	if ps.HardwareRequirements != nil {
		in, out := ps.HardwareRequirements, &out.HardwareRequirements
		*out = new(HardwareRequirementsConfig)
		(*out).GPUType = in.GPUType
		(*out).GPUCount = in.GPUCount
		(*out).TEERequired = in.TEERequired
		(*out).TEEType = in.TEEType
		if in.CustomLabels != nil {
			(*out).CustomLabels = make(map[string]string, len(in.CustomLabels))
			for key, val := range in.CustomLabels {
				(*out).CustomLabels[key] = val
			}
		}
	}
	if ps.ImagePullSecrets != nil {
		in, out := &ps.ImagePullSecrets, &out.ImagePullSecrets
		*out = make([]corev1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
}

// DeepCopyInto for SchedulingConfig
func (sc *SchedulingConfig) DeepCopyInto(out *SchedulingConfig) {
	*out = *sc
	if sc.NodeSelector != nil {
		in, out := &sc.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if sc.Tolerations != nil {
		in, out := &sc.Tolerations, &out.Tolerations
		*out = make([]TolerationConfig, len(*in))
		copy(*out, *in)
	}
}

// DeepCopyInto for PerformerStatusCRD
func (ps *PerformerStatusCRD) DeepCopyInto(out *PerformerStatusCRD) {
	*out = *ps
	if ps.Conditions != nil {
		in, out := &ps.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if ps.LastUpgrade != nil {
		in, out := ps.LastUpgrade, &out.LastUpgrade
		*out = (*in).DeepCopy()
	}
	if ps.ReadyTime != nil {
		in, out := ps.ReadyTime, &out.ReadyTime
		*out = (*in).DeepCopy()
	}
}

// DeepCopyObject for PerformerList
func (pl *PerformerList) DeepCopyObject() runtime.Object {
	return pl.DeepCopy()
}

// DeepCopy for PerformerList
func (pl *PerformerList) DeepCopy() *PerformerList {
	if pl == nil {
		return nil
	}
	out := new(PerformerList)
	pl.DeepCopyInto(out)
	return out
}

// DeepCopyInto for PerformerList
func (pl *PerformerList) DeepCopyInto(out *PerformerList) {
	*out = *pl
	out.TypeMeta = pl.TypeMeta
	pl.ListMeta.DeepCopyInto(&out.ListMeta)
	if pl.Items != nil {
		in, out := &pl.Items, &out.Items
		*out = make([]PerformerCRD, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// CRDOperations provides operations for managing Performer CRDs
type CRDOperations struct {
	client    client.Client
	namespace string
	config    *Config
}

// NewCRDOperations creates a new CRD operations instance
func NewCRDOperations(client client.Client, config *Config) *CRDOperations {
	return &CRDOperations{
		client:    client,
		namespace: config.Namespace,
		config:    config,
	}
}

// CreatePerformer creates a new Performer CRD
func (c *CRDOperations) CreatePerformer(ctx context.Context, req *CreatePerformerRequest) (*CreatePerformerResponse, error) {
	if err := ValidateCreatePerformerRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	performer := &PerformerCRD{
		TypeMeta: metav1.TypeMeta{
			APIVersion: fmt.Sprintf("%s/%s", c.config.CRDGroup, c.config.CRDVersion),
			Kind:       "Performer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: c.namespace,
			Labels: map[string]string{
				"app":                               "hourglass-performer",
				"hourglass.eigenlayer.io/avs":       req.AVSAddress,
				"hourglass.eigenlayer.io/performer": req.Name,
			},
		},
		Spec: PerformerSpec{
			AVSAddress:      req.AVSAddress,
			Image:           req.Image,
			ImagePullPolicy: corev1.PullPolicy(req.ImagePullPolicy),
			Version:         req.ImageTag,
			Config: PerformerConfig{
				GRPCPort:    req.GRPCPort,
				Environment: req.Environment,
			},
		},
	}

	// Convert resource requirements
	if req.Resources != nil {
		performer.Spec.Resources = convertResourceRequirements(req.Resources)
	}

	// Convert scheduling config
	if req.Scheduling != nil {
		performer.Spec.Scheduling = req.Scheduling
	}

	// Convert hardware requirements
	if req.HardwareRequirements != nil {
		performer.Spec.HardwareRequirements = req.HardwareRequirements
	}

	err := c.client.Create(ctx, performer)
	if err != nil {
		return nil, fmt.Errorf("failed to create Performer CRD: %w", err)
	}

	// Generate the expected service endpoint
	endpoint := fmt.Sprintf("performer-%s.%s.svc.cluster.local:%d", req.Name, c.namespace, req.GRPCPort)

	return &CreatePerformerResponse{
		PerformerID: req.Name,
		Endpoint:    endpoint,
		Status: PerformerStatus{
			Phase:        avsPerformer.PerformerResourceStatusStaged,
			ServiceName:  fmt.Sprintf("performer-%s", req.Name),
			GRPCEndpoint: endpoint,
			Ready:        false,
			Message:      "Performer CRD created, waiting for operator to provision resources",
			LastUpdated:  time.Now(),
		},
	}, nil
}

// GetPerformer retrieves a Performer CRD by name
func (c *CRDOperations) GetPerformer(ctx context.Context, name string) (*PerformerCRD, error) {
	performer := &PerformerCRD{}
	key := types.NamespacedName{
		Name:      name,
		Namespace: c.namespace,
	}

	err := c.client.Get(ctx, key, performer)
	if err != nil {
		return nil, fmt.Errorf("failed to get Performer CRD %s: %w", name, err)
	}

	return performer, nil
}

// UpdatePerformer updates a Performer CRD
func (c *CRDOperations) UpdatePerformer(ctx context.Context, req *UpdatePerformerRequest) error {
	if err := ValidateUpdatePerformerRequest(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	performer, err := c.GetPerformer(ctx, req.PerformerID)
	if err != nil {
		return err
	}

	// Update fields if provided
	if req.Image != "" {
		performer.Spec.Image = req.Image
	}
	if req.ImageTag != "" {
		performer.Spec.Version = req.ImageTag
	}

	err = c.client.Update(ctx, performer)
	if err != nil {
		return fmt.Errorf("failed to update Performer CRD %s: %w", req.PerformerID, err)
	}

	return nil
}

// DeletePerformer deletes a Performer CRD
func (c *CRDOperations) DeletePerformer(ctx context.Context, name string) error {
	performer := &PerformerCRD{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.namespace,
		},
	}

	err := c.client.Delete(ctx, performer)
	if err != nil {
		return fmt.Errorf("failed to delete Performer CRD %s: %w", name, err)
	}

	return nil
}

// ListPerformers lists all Performer CRDs in the namespace
func (c *CRDOperations) ListPerformers(ctx context.Context) ([]PerformerInfo, error) {
	performerList := &PerformerList{}
	performerList.TypeMeta = metav1.TypeMeta{
		APIVersion: fmt.Sprintf("%s/%s", c.config.CRDGroup, c.config.CRDVersion),
		Kind:       "PerformerList",
	}

	err := c.client.List(ctx, performerList, client.InNamespace(c.namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list Performer CRDs: %w", err)
	}

	var performers []PerformerInfo
	for _, performer := range performerList.Items {
		info := PerformerInfo{
			PerformerID: performer.Name,
			AVSAddress:  performer.Spec.AVSAddress,
			Image:       performer.Spec.Image,
			Version:     performer.Spec.Version,
			Status: PerformerStatus{
				Phase:        avsPerformer.PerformerResourceStatus(performer.Status.Phase),
				PodName:      performer.Status.PodName,
				ServiceName:  performer.Status.ServiceName,
				GRPCEndpoint: performer.Status.GRPCEndpoint,
				Ready:        performer.Status.Ready,
				Message:      extractConditionMessage(performer.Status.Conditions),
				LastUpdated:  time.Now(),
			},
			CreatedAt: performer.CreationTimestamp.Time,
			UpdatedAt: performer.CreationTimestamp.Time,
		}
		performers = append(performers, info)
	}

	return performers, nil
}

// GetPerformerStatus gets the status of a specific performer
func (c *CRDOperations) GetPerformerStatus(ctx context.Context, name string) (*PerformerStatus, error) {
	performer, err := c.GetPerformer(ctx, name)
	if err != nil {
		return nil, err
	}

	status := &PerformerStatus{
		Phase:        avsPerformer.PerformerResourceStatus(performer.Status.Phase),
		PodName:      performer.Status.PodName,
		ServiceName:  performer.Status.ServiceName,
		GRPCEndpoint: performer.Status.GRPCEndpoint,
		Ready:        performer.Status.Ready,
		Message:      extractConditionMessage(performer.Status.Conditions),
		LastUpdated:  time.Now(),
	}

	return status, nil
}

// convertResourceRequirements converts our resource requirements to Kubernetes format
func convertResourceRequirements(req *ResourceRequirements) corev1.ResourceRequirements {
	k8sReq := corev1.ResourceRequirements{}

	if req.Requests != nil {
		k8sReq.Requests = make(corev1.ResourceList)
		for key, value := range req.Requests {
			k8sReq.Requests[corev1.ResourceName(key)] = parseQuantity(value)
		}
	}

	if req.Limits != nil {
		k8sReq.Limits = make(corev1.ResourceList)
		for key, value := range req.Limits {
			k8sReq.Limits[corev1.ResourceName(key)] = parseQuantity(value)
		}
	}

	return k8sReq
}

// parseQuantity parses a quantity string (simplified version)
func parseQuantity(value string) resource.Quantity {
	// This is a simplified version - in production, you'd use resource.ParseQuantity
	// For now, we'll create a simple quantity
	return resource.MustParse(value)
}

// extractConditionMessage extracts a message from conditions
func extractConditionMessage(conditions []metav1.Condition) string {
	if len(conditions) == 0 {
		return ""
	}

	// Return the most recent condition message
	latest := conditions[len(conditions)-1]
	return latest.Message
}
