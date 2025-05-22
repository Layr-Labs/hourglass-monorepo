package performerCapacityPlanner

import "context"

// IPerformerCapacityPlanner defines the interface for a performer capacity planner
type IPerformerCapacityPlanner interface {
	// GetCapacityPlan returns a capacity plan for the given AVS
	GetCapacityPlan(avsAddress string) (*PerformerCapacityPlan, error)

	// Start begins processing chain events and discovering operator sets
	Start(ctx context.Context)
}

// ArtifactVersion represents a published artifact version for an AVS
type ArtifactVersion struct {
	AvsAddress    string
	OperatorSetId string
	Digest        string
	RegistryUrl   string
	PublishedAt   uint64 // Block number
}

// PerformerCapacityPlan represents the desired state for performers of an AVS
type PerformerCapacityPlan struct {
	// The desired number of performers
	TargetCount int

	// The digest/version this capacity plan applies to (optional)
	Digest string

	// The latest artifact version for this AVS
	LatestArtifact *ArtifactVersion
}
