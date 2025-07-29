package dao

import (
	"context"
	"fmt"
	"math/big"

	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
)

// EigenRuntimeSpecDAO implements SpecDAO for EigenLayer runtime specs
type EigenRuntimeSpecDAO struct {
	contractClient *client.ContractClient
	ociClient      *client.OCIClient
	operatorSetID  uint32
	log            logger.Logger
}

// NewEigenRuntimeSpecDAO creates a new EigenRuntimeSpecDAO instance
func NewEigenRuntimeSpecDAO(
	contractClient *client.ContractClient,
	ociClient *client.OCIClient,
	operatorSetID uint32,
	log logger.Logger,
) *EigenRuntimeSpecDAO {
	return &EigenRuntimeSpecDAO{
		contractClient: contractClient,
		ociClient:      ociClient,
		operatorSetID:  operatorSetID,
		log:            log,
	}
}

// GetRuntimeSpec retrieves a specific runtime spec by release ID
func (dao *EigenRuntimeSpecDAO) GetRuntimeSpec(ctx context.Context, releaseID string) (*runtime.Spec, error) {
	// Parse release ID
	releaseIdBig, ok := new(big.Int).SetString(releaseID, 10)
	if !ok {
		return nil, fmt.Errorf("invalid release ID: %s", releaseID)
	}

	// Get release from contract
	release, err := dao.contractClient.GetRelease(ctx, dao.operatorSetID, releaseIdBig)
	if err != nil {
		return nil, fmt.Errorf("failed to get release %s: %w", releaseID, err)
	}

	dao.log.Info("Using specific release", zap.String("releaseId", releaseID))

	// Extract and pull runtime spec
	return dao.pullRuntimeSpec(ctx, release)
}

// GetLatestRuntimeSpec retrieves the latest runtime spec
func (dao *EigenRuntimeSpecDAO) GetLatestRuntimeSpec(ctx context.Context) (*runtime.Spec, error) {
	// Get latest release ID
	nextReleaseId, err := dao.contractClient.GetReleaseCount(ctx, dao.operatorSetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next release ID: %w", err)
	}

	if nextReleaseId.Uint64() == 0 {
		return nil, fmt.Errorf("no releases found for operator set %d", dao.operatorSetID)
	}

	// Latest release ID is nextReleaseId - 1
	releaseIndex := new(big.Int).Sub(nextReleaseId, big.NewInt(1))

	// Get release from contract
	release, err := dao.contractClient.GetRelease(ctx, dao.operatorSetID, releaseIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get release %d: %w", releaseIndex.Uint64(), err)
	}

	dao.log.Info("Using latest release", zap.Uint64("releaseId", releaseIndex.Uint64()))

	// Extract and pull runtime spec
	return dao.pullRuntimeSpec(ctx, release)
}

// pullRuntimeSpec extracts the runtime spec from a release
func (dao *EigenRuntimeSpecDAO) pullRuntimeSpec(ctx context.Context, release *client.ReleaseManagerRelease) (*runtime.Spec, error) {
	if len(release.Artifacts) == 0 {
		return nil, fmt.Errorf("no artifacts found in release")
	}

	// Extract first artifact (assuming EigenRuntime spec)
	artifact := release.Artifacts[0]

	// Format digest
	digest := fmt.Sprintf("sha256:%x", artifact.Digest)

	dao.log.Info("Found release artifact",
		zap.String("registry", artifact.RegistryName),
		zap.String("digest", digest))

	// Pull runtime spec from OCI registry
	spec, _, err := dao.ociClient.PullRuntimeSpec(ctx, artifact.RegistryName, digest)
	if err != nil {
		return nil, fmt.Errorf("failed to pull runtime spec: %w", err)
	}

	return spec, nil
}
