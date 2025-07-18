package testUtils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ImageManager handles Docker image building and loading for Kind clusters
type ImageManager struct {
	ProjectRoot string
	Logger      *zap.SugaredLogger
}

// ImageConfig represents configuration for building and loading images
type ImageConfig struct {
	Name         string
	Tag          string
	BuildContext string
	Dockerfile   string
	BuildArgs    map[string]string
	LoadToKind   bool
}

// DefaultImageConfigs returns the default image configurations for testing
func DefaultImageConfigs(projectRoot string) []ImageConfig {
	return []ImageConfig{
		{
			Name:         "hourglass/operator",
			Tag:          "test",
			BuildContext: filepath.Join(projectRoot, "hourglass-operator"),
			Dockerfile:   "Dockerfile",
			BuildArgs:    map[string]string{},
			LoadToKind:   true,
		},
		{
			Name:         "hello-performer",
			Tag:          "latest",
			BuildContext: filepath.Join(projectRoot, "demo"),
			Dockerfile:   "Dockerfile",
			BuildArgs:    map[string]string{},
			LoadToKind:   true,
		},
	}
}

// NewImageManager creates a new ImageManager
func NewImageManager(projectRoot string, logger *zap.SugaredLogger) *ImageManager {
	return &ImageManager{
		ProjectRoot: projectRoot,
		Logger:      logger,
	}
}

// BuildAndLoadImages builds and loads all required images for testing
func (im *ImageManager) BuildAndLoadImages(ctx context.Context, cluster *KindCluster, configs []ImageConfig) error {
	im.Logger.Infof("Building and loading %d images for testing", len(configs))

	for _, config := range configs {
		if err := im.BuildImage(ctx, config); err != nil {
			return fmt.Errorf("failed to build image %s: %v", config.Name, err)
		}

		if config.LoadToKind {
			if err := im.LoadImageToKind(ctx, cluster, config); err != nil {
				return fmt.Errorf("failed to load image %s to Kind: %v", config.Name, err)
			}
		}
	}

	im.Logger.Infof("All images built and loaded successfully")
	return nil
}

// BuildImage builds a Docker image
func (im *ImageManager) BuildImage(ctx context.Context, config ImageConfig) error {
	fullImageName := fmt.Sprintf("%s:%s", config.Name, config.Tag)
	im.Logger.Infof("Building Docker image: %s", fullImageName)

	// Check if build context exists
	if _, err := os.Stat(config.BuildContext); os.IsNotExist(err) {
		return fmt.Errorf("build context does not exist: %s", config.BuildContext)
	}

	// Check if Dockerfile exists
	dockerfilePath := filepath.Join(config.BuildContext, config.Dockerfile)
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("Dockerfile does not exist: %s", dockerfilePath)
	}

	// Build Docker command
	args := []string{"build", "-t", fullImageName, "-f", config.Dockerfile}

	// Add build args
	for key, value := range config.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	// Add build context
	args = append(args, config.BuildContext)

	// Execute build command
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %v", err)
	}

	im.Logger.Infof("Successfully built image: %s", fullImageName)
	return nil
}

// LoadImageToKind loads a Docker image into the Kind cluster
func (im *ImageManager) LoadImageToKind(ctx context.Context, cluster *KindCluster, config ImageConfig) error {
	fullImageName := fmt.Sprintf("%s:%s", config.Name, config.Tag)
	im.Logger.Infof("Loading image %s to Kind cluster %s", fullImageName, cluster.Name)

	// Check if image exists locally
	if !im.ImageExists(ctx, fullImageName) {
		return fmt.Errorf("image %s does not exist locally", fullImageName)
	}

	// Load image into Kind cluster
	if err := cluster.LoadDockerImage(ctx, fullImageName); err != nil {
		return fmt.Errorf("failed to load image to Kind: %v", err)
	}

	im.Logger.Infof("Successfully loaded image %s to Kind cluster", fullImageName)
	return nil
}

// ImageExists checks if a Docker image exists locally
func (im *ImageManager) ImageExists(ctx context.Context, imageName string) bool {
	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", imageName)
	err := cmd.Run()
	return err == nil
}

// BuildOperatorImage builds the Hourglass operator image
func (im *ImageManager) BuildOperatorImage(ctx context.Context) error {
	im.Logger.Infof("Building Hourglass operator image")

	operatorDir := filepath.Join(im.ProjectRoot, "hourglass-operator")

	// Check if operator directory exists
	if _, err := os.Stat(operatorDir); os.IsNotExist(err) {
		return fmt.Errorf("operator directory does not exist: %s", operatorDir)
	}

	// Check if Dockerfile exists
	dockerfilePath := filepath.Join(operatorDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("operator Dockerfile does not exist: %s", dockerfilePath)
	}

	config := ImageConfig{
		Name:         "hourglass/operator",
		Tag:          "test",
		BuildContext: operatorDir,
		Dockerfile:   "Dockerfile",
		BuildArgs:    map[string]string{},
		LoadToKind:   false,
	}

	return im.BuildImage(ctx, config)
}

// BuildPerformerImage builds the test performer image
func (im *ImageManager) BuildPerformerImage(ctx context.Context) error {
	im.Logger.Infof("Building test performer image")

	demoDir := filepath.Join(im.ProjectRoot, "demo")

	// Check if demo directory exists
	if _, err := os.Stat(demoDir); os.IsNotExist(err) {
		return fmt.Errorf("demo directory does not exist: %s", demoDir)
	}

	// Check if Dockerfile exists
	dockerfilePath := filepath.Join(demoDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("performer Dockerfile does not exist: %s", dockerfilePath)
	}

	config := ImageConfig{
		Name:         "hello-performer",
		Tag:          "latest",
		BuildContext: demoDir,
		Dockerfile:   "Dockerfile",
		BuildArgs:    map[string]string{},
		LoadToKind:   false,
	}

	return im.BuildImage(ctx, config)
}

// PullRequiredImages pulls any required base images
func (im *ImageManager) PullRequiredImages(ctx context.Context) error {
	im.Logger.Infof("Pulling required base images")

	// List of base images commonly needed
	baseImages := []string{
		"golang:1.21-alpine",
		"alpine:3.18",
		"busybox:1.35",
	}

	for _, image := range baseImages {
		if err := im.PullImage(ctx, image); err != nil {
			im.Logger.Warnf("Failed to pull base image %s: %v", image, err)
			// Continue with other images
		}
	}

	return nil
}

// PullImage pulls a Docker image
func (im *ImageManager) PullImage(ctx context.Context, imageName string) error {
	im.Logger.Infof("Pulling Docker image: %s", imageName)

	cmd := exec.CommandContext(ctx, "docker", "pull", imageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull image %s: %v", imageName, err)
	}

	im.Logger.Infof("Successfully pulled image: %s", imageName)
	return nil
}

// CleanupImages removes test images from the local Docker daemon
func (im *ImageManager) CleanupImages(ctx context.Context, configs []ImageConfig) error {
	im.Logger.Infof("Cleaning up test images")

	for _, config := range configs {
		fullImageName := fmt.Sprintf("%s:%s", config.Name, config.Tag)
		if err := im.RemoveImage(ctx, fullImageName); err != nil {
			im.Logger.Warnf("Failed to remove image %s: %v", fullImageName, err)
		}
	}

	return nil
}

// RemoveImage removes a Docker image
func (im *ImageManager) RemoveImage(ctx context.Context, imageName string) error {
	im.Logger.Infof("Removing Docker image: %s", imageName)

	cmd := exec.CommandContext(ctx, "docker", "rmi", imageName, "--force")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove image %s: %v", imageName, err)
	}

	im.Logger.Infof("Successfully removed image: %s", imageName)
	return nil
}

// GetImageInfo returns information about a Docker image
func (im *ImageManager) GetImageInfo(ctx context.Context, imageName string) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", imageName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image %s: %v", imageName, err)
	}

	// For simplicity, return raw output as string
	// In a real implementation, you might parse the JSON
	return map[string]interface{}{
		"raw_output": string(output),
		"exists":     true,
	}, nil
}

// ValidateImageRequirements validates that all required images can be built
func (im *ImageManager) ValidateImageRequirements(ctx context.Context, configs []ImageConfig) error {
	im.Logger.Infof("Validating image requirements")

	for _, config := range configs {
		// Check if build context exists
		if _, err := os.Stat(config.BuildContext); os.IsNotExist(err) {
			return fmt.Errorf("build context for %s does not exist: %s", config.Name, config.BuildContext)
		}

		// Check if Dockerfile exists
		dockerfilePath := filepath.Join(config.BuildContext, config.Dockerfile)
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			return fmt.Errorf("Dockerfile for %s does not exist: %s", config.Name, dockerfilePath)
		}

		im.Logger.Infof("Image %s requirements validated", config.Name)
	}

	return nil
}

// WaitForImageInKind waits for an image to be available in the Kind cluster
func (im *ImageManager) WaitForImageInKind(ctx context.Context, cluster *KindCluster, imageName string, timeout time.Duration) error {
	im.Logger.Infof("Waiting for image %s to be available in Kind cluster", imageName)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for image %s to be available in Kind", imageName)
		case <-ticker.C:
			// Check if image is available in Kind cluster
			if available, err := im.IsImageAvailableInKind(ctx, cluster, imageName); err != nil {
				im.Logger.Warnf("Error checking image availability: %v", err)
				continue
			} else if available {
				im.Logger.Infof("Image %s is available in Kind cluster", imageName)
				return nil
			}
		}
	}
}

// IsImageAvailableInKind checks if an image is available in the Kind cluster
func (im *ImageManager) IsImageAvailableInKind(ctx context.Context, cluster *KindCluster, imageName string) (bool, error) {
	// Run docker exec on the Kind node to check if image exists
	nodeContainerName := fmt.Sprintf("%s-control-plane", cluster.Name)

	cmd := exec.CommandContext(ctx, "docker", "exec", nodeContainerName, "crictl", "images", "--filter", fmt.Sprintf("reference=%s", imageName))
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check image in Kind node: %v", err)
	}

	// Check if the image appears in the output
	return strings.Contains(string(output), imageName), nil
}

// ListImagesInKind lists all images available in the Kind cluster
func (im *ImageManager) ListImagesInKind(ctx context.Context, cluster *KindCluster) ([]string, error) {
	im.Logger.Infof("Listing images in Kind cluster %s", cluster.Name)

	nodeContainerName := fmt.Sprintf("%s-control-plane", cluster.Name)

	cmd := exec.CommandContext(ctx, "docker", "exec", nodeContainerName, "crictl", "images")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list images in Kind node: %v", err)
	}

	// Parse the output to extract image names
	lines := strings.Split(string(output), "\n")
	var images []string

	for _, line := range lines {
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "IMAGE") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				images = append(images, parts[0])
			}
		}
	}

	return images, nil
}

// PreloadCommonImages preloads commonly used images to speed up testing
func (im *ImageManager) PreloadCommonImages(ctx context.Context, cluster *KindCluster) error {
	im.Logger.Infof("Preloading common images to Kind cluster")

	commonImages := []string{
		"alpine:3.18",
		"busybox:1.35",
		"nginx:1.21-alpine",
	}

	for _, image := range commonImages {
		// Pull if not exists
		if !im.ImageExists(ctx, image) {
			if err := im.PullImage(ctx, image); err != nil {
				im.Logger.Warnf("Failed to pull common image %s: %v", image, err)
				continue
			}
		}

		// Load to Kind
		if err := cluster.LoadDockerImage(ctx, image); err != nil {
			im.Logger.Warnf("Failed to load common image %s to Kind: %v", image, err)
			continue
		}
	}

	im.Logger.Infof("Common images preloaded successfully")
	return nil
}
