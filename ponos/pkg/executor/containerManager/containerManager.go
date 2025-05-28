package containerManager

import (
	"context"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/performerCapacityPlanner"
)

// PullResult contains information about a pulled container
type PullResult struct {
	// ImageID is the ID of the pulled image (platform-specific)
	ImageID string

	// Tag is the tag of the pulled image
	Tag string

	// RequestedDigest is the original digest that was requested (may be an index digest)
	RequestedDigest string

	// RepoDigests are the repository digests from the pulled image (platform-specific)
	RepoDigests []string

	// Platform is the OS/architecture of the pulled image (e.g., "linux/amd64")
	Platform string
}

// IContainerManager defines the interface for container management operations
type IContainerManager interface {
	// PullContainer pulls a container image from the specified registry
	// artifact: The artifact version containing registry URL, digest, and/or tag
	// Returns a PullResult with information about the pulled image or an error
	PullContainer(ctx context.Context, artifact *performerCapacityPlanner.ArtifactVersion) (*PullResult, error)
}
