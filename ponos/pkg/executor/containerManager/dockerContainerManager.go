package containerManager

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/performerCapacityPlanner"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

// DockerContainerManager implements IContainerManager using Docker
type DockerContainerManager struct {
	dockerClient *client.Client
	logger       *zap.Logger
}

// NewDockerContainerManager creates a new Docker container manager
func NewDockerContainerManager(logger *zap.Logger) (*DockerContainerManager, error) {
	// Create Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &DockerContainerManager{
		dockerClient: dockerClient,
		logger:       logger,
	}, nil
}

// PullContainer pulls a container image from the specified registry
func (m *DockerContainerManager) PullContainer(
	ctx context.Context,
	artifact *performerCapacityPlanner.ArtifactVersion,
) (*PullResult, error) {
	if artifact == nil {
		return nil, fmt.Errorf("artifact cannot be nil")
	}

	// Determine what to use for pulling - prefer digest over tag
	var identifier string
	if artifact.Digest != "" {
		identifier = artifact.Digest
	} else if artifact.Tag != "" {
		identifier = artifact.Tag
	} else {
		return nil, fmt.Errorf("artifact must have either digest or tag")
	}

	m.logger.Sugar().Infow("Pulling container",
		"registryUrl", artifact.RegistryUrl,
		"digest", artifact.Digest,
		"tag", artifact.Tag,
		"identifier", identifier,
	)

	// Format image reference
	imageRef := formatImageReference(artifact.RegistryUrl, identifier)

	// Get current OS and architecture for platform-specific pulling
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Prepare platform string (e.g., "linux/amd64")
	platformStr := fmt.Sprintf("%s/%s", os, arch)

	m.logger.Sugar().Infow("Pulling image with platform specification",
		"imageRef", imageRef,
		"platform", platformStr,
	)

	// Pull image with platform specification
	// Docker API handles OCI indexes automatically when platform is specified
	pullOpts := image.PullOptions{
		Platform: platformStr,
	}

	pullReader, err := m.dockerClient.ImagePull(ctx, imageRef, pullOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image %s for platform %s: %w", imageRef, platformStr, err)
	}
	defer pullReader.Close()

	// Read and process the pull output
	_, err = io.ReadAll(pullReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read pull output: %w", err)
	}

	// Inspect the pulled image using the newer API method
	inspectResult, err := m.dockerClient.ImageInspect(ctx, imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect pulled image: %w", err)
	}

	// Create result with both the requested digest and platform-specific information
	result := &PullResult{
		ImageID:         inspectResult.ID,
		Tag:             imageRef,
		RequestedDigest: identifier,
		RepoDigests:     inspectResult.RepoDigests,
		Platform:        platformStr,
	}

	m.logger.Sugar().Infow("Successfully pulled container",
		"imageId", result.ImageID,
		"tag", result.Tag,
		"requestedIdentifier", result.RequestedDigest,
		"repoDigests", result.RepoDigests,
		"platform", result.Platform,
	)

	return result, nil
}

// formatImageReference formats the registry URL and digest into a Docker image reference
func formatImageReference(registryUrl, digest string) string {
	// If digest contains '@', it's already in digest format
	if strings.Contains(digest, "@") {
		return fmt.Sprintf("%s%s", registryUrl, digest)
	}
	// Otherwise, treat it as a tag
	return fmt.Sprintf("%s:%s", registryUrl, digest)
}
