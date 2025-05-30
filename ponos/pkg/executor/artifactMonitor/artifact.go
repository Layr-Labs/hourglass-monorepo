package artifactMonitor

import "context"

// IPerformerArtifactMonitor defines the interface for a performer artifact monitor
type IPerformerArtifactMonitor interface {
	// GetArtifacts returns a artifacts for the given AVS
	GetArtifacts(avsAddress string) ([]PerformerArtifact, error)

	// Start begins processing chain events and discovering operator sets
	Start(ctx context.Context)
}

// PerformerArtifact represents a published artifact version for an AVS
type PerformerArtifact struct {
	AvsAddress    string
	OperatorSetId string
	Digest        string
	Tag           string
	RegistryUrl   string
	PublishedAt   uint64 // Block number
}
