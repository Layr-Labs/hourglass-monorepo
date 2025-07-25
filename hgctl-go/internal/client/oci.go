package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
	"gopkg.in/yaml.v3"
)

type OCIClient struct {
	logger logger.Logger
}

func NewOCIClient(logger logger.Logger) *OCIClient {
	return &OCIClient{logger: logger}
}

// PullRuntimeSpec pulls a runtime spec from an OCI registry by digest
func (c *OCIClient) PullRuntimeSpec(ctx context.Context, registryName, digest string) (*runtime.Spec, error) {
	// Clean up the digest (remove 0x prefix if present)
	digest = strings.TrimPrefix(digest, "0x")
	if !strings.HasPrefix(digest, "sha256:") {
		digest = "sha256:" + digest
	}

	c.logger.Debug("Pulling runtime spec",
		zap.String("registry", registryName),
		zap.String("digest", digest))

	// Parse the registry URL
	ref, err := registry.ParseReference(registryName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry reference: %w", err)
	}

	// Create repository
	repo, err := remote.NewRepository(fmt.Sprintf("%s/%s", ref.Registry, ref.Repository))
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Create memory store for artifact
	memoryStore := memory.New()

	// Copy the artifact
	_, err = oras.Copy(ctx, repo, digest, memoryStore, digest, oras.DefaultCopyOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to pull artifact: %w", err)
	}

	// Extract runtime spec from the artifact
	spec, err := c.extractRuntimeSpec(ctx, memoryStore, digest)
	if err != nil {
		return nil, fmt.Errorf("failed to extract runtime spec: %w", err)
	}

	return spec, nil
}

// PullRuntimeSpecByTag pulls a runtime spec by tag reference
func (c *OCIClient) PullRuntimeSpecByTag(ctx context.Context, registryName, avsName string, operatorSetID uint32, version string) (*runtime.Spec, error) {
	tag := fmt.Sprintf("opset-%d-v%s", operatorSetID, version)
	fullRef := fmt.Sprintf("%s/%s:%s", registryName, avsName, tag)

	c.logger.Debug("Pulling runtime spec by tag",
		zap.String("reference", fullRef))

	// Parse the reference
	ref, err := registry.ParseReference(fullRef)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	// Create repository
	repo, err := remote.NewRepository(fmt.Sprintf("%s/%s", ref.Registry, ref.Repository))
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Create memory store
	memoryStore := memory.New()

	// Copy the artifact
	_, err = oras.Copy(ctx, repo, tag, memoryStore, tag, oras.DefaultCopyOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to pull artifact: %w", err)
	}

	// Get the manifest to find the digest
	desc, err := memoryStore.Resolve(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tag: %w", err)
	}

	// Extract runtime spec
	spec, err := c.extractRuntimeSpec(ctx, memoryStore, desc.Digest.String())
	if err != nil {
		return nil, fmt.Errorf("failed to extract runtime spec: %w", err)
	}

	return spec, nil
}

func (c *OCIClient) extractRuntimeSpec(ctx context.Context, store oras.ReadOnlyTarget, digest string) (*runtime.Spec, error) {
	// Fetch the manifest
	manifestDesc, err := store.Resolve(ctx, digest)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve manifest: %w", err)
	}

	manifestRC, err := store.Fetch(ctx, manifestDesc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer manifestRC.Close()

	manifestData, err := io.ReadAll(manifestRC)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse manifest
	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	c.logger.Debug("Found manifest",
		zap.Int("layers", len(manifest.Layers)))

	// Find the runtime spec layer
	var runtimeSpecLayer *ocispec.Descriptor
	for _, layer := range manifest.Layers {
		c.logger.Debug("Checking layer",
			zap.String("mediaType", layer.MediaType),
			zap.String("digest", layer.Digest.String()))

		// Check annotations
		if title, ok := layer.Annotations["org.opencontainers.image.title"]; ok && title == "runtime-spec.yaml" {
			runtimeSpecLayer = &layer
			break
		}

		// Check media type
		if layer.MediaType == "text/yaml" || layer.MediaType == "application/vnd.eigenlayer.runtime-spec.v1+yaml" {
			runtimeSpecLayer = &layer
			break
		}
	}

	// If no specific layer found, use the first layer
	if runtimeSpecLayer == nil && len(manifest.Layers) > 0 {
		runtimeSpecLayer = &manifest.Layers[0]
		c.logger.Debug("Using first layer as runtime spec")
	}

	if runtimeSpecLayer == nil {
		return nil, fmt.Errorf("no runtime spec layer found in artifact")
	}

	// Fetch the layer content
	layerRC, err := store.Fetch(ctx, *runtimeSpecLayer)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch layer: %w", err)
	}
	defer layerRC.Close()

	layerData, err := io.ReadAll(layerRC)
	if err != nil {
		return nil, fmt.Errorf("failed to read layer: %w", err)
	}

	// Parse the YAML content
	var spec runtime.Spec
	if err := yaml.Unmarshal(layerData, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse runtime spec: %w", err)
	}

	return &spec, nil
}
