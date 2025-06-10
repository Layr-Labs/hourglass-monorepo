package containerManager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestDockerContainerManager_ImagePulling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping image pulling test in short mode")
	}

	if !isDockerAvailable(t) {
		t.Skip("Docker is not available, skipping image pulling tests")
	}

	logger := zaptest.NewLogger(t)
	dcm, err := NewDockerContainerManager(nil, logger)
	require.NoError(t, err)
	defer func() { _ = dcm.Shutdown(context.Background()) }()

	ctx := context.Background()

	t.Run("ensureImageExists with existing image", func(t *testing.T) {
		// Test with alpine which should already exist locally or be pullable
		err := dcm.ensureImageExists(ctx, "alpine:latest")
		assert.NoError(t, err)
	})

	t.Run("ensureImageExists with small test image", func(t *testing.T) {
		// Test with a very small image that might not exist locally
		err := dcm.ensureImageExists(ctx, "hello-world:latest")
		assert.NoError(t, err)
	})

	t.Run("ensureImageExists with non-existent image", func(t *testing.T) {
		// Test with an image that definitely doesn't exist
		err := dcm.ensureImageExists(ctx, "this-image-definitely-does-not-exist:never")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to pull image")
	})
}

func TestDockerContainerManager_CreateWithImagePull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping image pulling integration test in short mode")
	}

	if !isDockerAvailable(t) {
		t.Skip("Docker is not available, skipping image pulling tests")
	}

	logger := zaptest.NewLogger(t)
	dcm, err := NewDockerContainerManager(nil, logger)
	require.NoError(t, err)
	defer func() { _ = dcm.Shutdown(context.Background()) }()

	ctx := context.Background()

	t.Run("create container with automatic image pull", func(t *testing.T) {
		// Use hello-world which is very small and likely to be pulled quickly
		config := &ContainerConfig{
			Hostname:   "test-pull",
			Image:      "hello-world:latest",
			AutoRemove: true,
		}

		// This should automatically pull the image if not present
		containerInfo, err := dcm.Create(ctx, config)
		require.NoError(t, err)
		assert.NotEmpty(t, containerInfo.ID)
		assert.Equal(t, "test-pull", containerInfo.Hostname)

		// Clean up - remove the container
		err = dcm.Remove(ctx, containerInfo.ID, true)
		// Don't require no error since AutoRemove might handle it
		if err != nil && !containsAny(err.Error(), []string{"already in progress", "No such container"}) {
			t.Logf("Container removal warning: %v", err)
		}
	})
}

// Helper function to check if error message contains any of the given strings
func containsAny(str string, substrings []string) bool {
	for _, substr := range substrings {
		for i := 0; i <= len(str)-len(substr); i++ {
			if str[i:i+len(substr)] == substr {
				return true
			}
		}
	}
	return false
}
