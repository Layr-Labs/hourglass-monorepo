package client

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	releasemanager "github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ReleaseManager"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

// Release types
type OperatorSetRelease struct {
	Digest   string `json:"digest"`
	Registry string `json:"registry"`
}

type Release struct {
	ID                  string                        `json:"id"`
	OperatorSetReleases map[string]OperatorSetRelease `json:"operatorSetReleases"`
	UpgradeByTime       uint32                        `json:"upgradeByTime"`
}

// ReleaseArtifact represents an artifact in a release
type ReleaseArtifact struct {
	Digest       [32]byte
	RegistryName string
}

// ReleaseManagerRelease represents a release from the contract
type ReleaseManagerRelease struct {
	UpgradeByTime uint32
	Artifacts     []ReleaseArtifact
}

type ContractClient struct {
	ethClient *ethclient.Client
	logger    logger.Logger
}

func NewContractClient(rpcURL string, log logger.Logger) (*ContractClient, error) {
	if rpcURL == "" {
		return nil, fmt.Errorf("RPC URL is required")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	return &ContractClient{
		ethClient: client,
		logger:    log,
	}, nil
}

// GetRelease fetches a release from the ReleaseManager contract
func (c *ContractClient) GetRelease(ctx context.Context, releaseManagerAddr common.Address, avsAddress common.Address, operatorSetId uint32, releaseId *big.Int) (*ReleaseManagerRelease, error) {
	// Create release manager instance
	rm, err := releasemanager.NewReleaseManager(releaseManagerAddr, c.ethClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create release manager instance: %w", err)
	}

	// Create operator set
	operatorSet := releasemanager.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}

	// Get release from contract
	release, err := rm.GetRelease(&bind.CallOpts{Context: ctx}, operatorSet, releaseId)
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	// Convert to our internal type
	artifacts := make([]ReleaseArtifact, len(release.Artifacts))
	for i, artifact := range release.Artifacts {
		artifacts[i] = ReleaseArtifact{
			Digest:       artifact.Digest,
			RegistryName: artifact.Registry,
		}
	}

	return &ReleaseManagerRelease{
		UpgradeByTime: release.UpgradeByTime,
		Artifacts:     artifacts,
	}, nil
}

// GetNextReleaseId gets the next release ID for an operator set
func (c *ContractClient) GetNextReleaseId(ctx context.Context, releaseManagerAddr common.Address, avsAddress common.Address, operatorSetId uint32) (*big.Int, error) {
	// Create release manager instance
	rm, err := releasemanager.NewReleaseManager(releaseManagerAddr, c.ethClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create release manager instance: %w", err)
	}

	// Create operator set
	operatorSet := releasemanager.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}

	// Get total releases
	totalReleases, err := rm.GetTotalReleases(&bind.CallOpts{Context: ctx}, operatorSet)
	if err != nil {
		return nil, fmt.Errorf("failed to get total releases: %w", err)
	}

	return totalReleases, nil
}

// GetReleases fetches multiple releases
func (c *ContractClient) GetReleases(ctx context.Context, releaseManagerAddr common.Address, avsAddress common.Address, operatorSetIds []uint32, limit uint64) ([]*Release, error) {
	var releases []*Release

	for _, opSetId := range operatorSetIds {
		nextId, err := c.GetNextReleaseId(ctx, releaseManagerAddr, avsAddress, opSetId)
		if err != nil {
			c.logger.Warn("Failed to get next release ID",
				zap.Uint32("operatorSetId", opSetId),
				zap.Error(err))
			continue
		}

		// Iterate through releases
		for i := int64(0); i < nextId.Int64() && uint64(len(releases)) < limit; i++ {
			release, err := c.GetRelease(ctx, releaseManagerAddr, avsAddress, opSetId, big.NewInt(i))
			if err != nil {
				c.logger.Warn("Failed to get release",
					zap.Uint32("operatorSetId", opSetId),
					zap.Int64("releaseId", i),
					zap.Error(err))
				continue
			}

			if len(release.Artifacts) == 0 {
				continue
			}

			// Convert to internal release type
			internalRelease := &Release{
				ID: fmt.Sprintf("%d", i),
				OperatorSetReleases: map[string]OperatorSetRelease{
					fmt.Sprintf("%d", opSetId): {
						Digest:   fmt.Sprintf("0x%x", release.Artifacts[0].Digest),
						Registry: release.Artifacts[0].RegistryName,
					},
				},
				UpgradeByTime: release.UpgradeByTime,
			}

			releases = append(releases, internalRelease)
		}
	}

	return releases, nil
}

func (c *ContractClient) Close() {
	c.ethClient.Close()
}

// Helper to create a bound contract instance
func (c *ContractClient) bindContract(address common.Address, abi abi.ABI) *bind.BoundContract {
	return bind.NewBoundContract(address, abi, c.ethClient, c.ethClient, c.ethClient)
}
