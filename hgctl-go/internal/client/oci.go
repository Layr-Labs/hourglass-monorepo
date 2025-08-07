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
func (c *OCIClient) PullRuntimeSpec(ctx context.Context, registryName, digest string) (*runtime.Spec, []byte, error) {
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
		return nil, nil, fmt.Errorf("failed to parse registry reference: %w", err)
	}

	// Create repository
	repo, err := remote.NewRepository(fmt.Sprintf("%s/%s", ref.Registry, ref.Repository))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Create memory store for artifact
	memoryStore := memory.New()

	// Use extended copy options to ensure all blobs are copied
	copyOpts := oras.CopyOptions{
		CopyGraphOptions: oras.CopyGraphOptions{
			Concurrency: 3,
		},
	}

	// Copy the artifact - this should copy the manifest and all referenced blobs
	_, err = oras.Copy(ctx, repo, digest, memoryStore, "", copyOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to pull artifact: %w", err)
	}

	// First, let's fetch the manifest to get the actual layer digest dynamically
	manifestDesc, err := memoryStore.Resolve(ctx, digest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve manifest: %w", err)
	}

	manifestRC, err := memoryStore.Fetch(ctx, manifestDesc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer func(manifestRC io.ReadCloser) {
		err := manifestRC.Close()
		if err != nil {
			c.logger.Error("failed to close manifest reader", zap.Error(err))
		}
	}(manifestRC)

	manifestData, err := io.ReadAll(manifestRC)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	// Find the YAML layer
	var specLayer *ocispec.Descriptor
	for _, layer := range manifest.Layers {
		if layer.MediaType == "text/yaml" {
			specLayer = &layer
			break
		}
	}

	if specLayer == nil {
		return nil, nil, fmt.Errorf("no yaml layer found in manifest")
	}

	// Fetch the spec layer
	specRC, err := memoryStore.Fetch(ctx, *specLayer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch runtime spec layer: %w", err)
	}
	defer func(specRC io.ReadCloser) {
		err := specRC.Close()
		if err != nil {
			c.logger.Error("failed to close runtime spec reader", zap.Error(err))
		}
	}(specRC)

	specBytes, err := io.ReadAll(specRC)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read runtime spec: %w", err)
	}

	// Parse the spec
	var spec runtime.Spec
	if err := yaml.Unmarshal(specBytes, &spec); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal runtime spec: %w", err)
	}

	return &spec, specBytes, nil
}
