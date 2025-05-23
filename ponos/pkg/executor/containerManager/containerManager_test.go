package containerManager

import (
	"context"
	"fmt"
	"sync"
)

// MockContainerManager implements IContainerManager for testing
type MockContainerManager struct {
	mutex sync.Mutex

	// Track pulled containers
	PulledContainers map[string]*PullResult

	// For controlling return values in tests
	PullContainerFn func(ctx context.Context, registryUrl string, digest string) (*PullResult, error)
}

// NewMockContainerManager creates a new mock container manager
func NewMockContainerManager() *MockContainerManager {
	return &MockContainerManager{
		mutex:            sync.Mutex{},
		PulledContainers: make(map[string]*PullResult),
		PullContainerFn:  nil,
	}
}

// PullContainer implements the IContainerManager interface for testing
func (m *MockContainerManager) PullContainer(ctx context.Context, registryUrl string, digest string) (*PullResult, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Use custom function if provided
	if m.PullContainerFn != nil {
		return m.PullContainerFn(ctx, registryUrl, digest)
	}

	// Default implementation just records the call and returns a mock result
	key := createImageReference(registryUrl, digest)

	result := &PullResult{
		ImageID:         "mock-image-id",
		Tag:             key,
		RequestedDigest: digest,
		RepoDigests:     []string{"mock-repo-digest"},
		Platform:        "mock-platform",
	}

	m.PulledContainers[key] = result
	return result, nil
}

// createImageReference formats the registry URL and digest into a Docker image reference
// This is a helper function for the mock implementation
func createImageReference(registryUrl, digest string) string {
	return fmt.Sprintf("%s@%s", registryUrl, digest)
}
