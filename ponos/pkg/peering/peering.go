package peering

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
)

type OperatorSet struct {
	OperatorSetID  uint32           `json:"operatorSetId"`
	PublicKey      *bn254.PublicKey `json:"publicKey"`
	NetworkAddress string           `json:"networkAddress"`
}

func (os *OperatorSet) Clone() (*OperatorSet, error) {
	pk, err := bn254.NewPublicKeyFromBytes(os.PublicKey.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to copy public key: %w", err)
	}
	return &OperatorSet{
		OperatorSetID:  os.OperatorSetID,
		PublicKey:      pk,
		NetworkAddress: os.NetworkAddress,
	}, nil
}

type OperatorPeerInfo struct {
	OperatorAddress string         `json:"operatorAddress"`
	OperatorSets    []*OperatorSet `json:"operatorSets,omitempty"`
}

func (opi *OperatorPeerInfo) GetOperatorSet(operatorSetId uint32) (*OperatorSet, error) {
	for _, os := range opi.OperatorSets {
		if os.OperatorSetID == operatorSetId {
			return os, nil
		}
	}
	return nil, fmt.Errorf("operator set with ID %d not found in operator peer info", operatorSetId)
}

func (opi *OperatorPeerInfo) GetSocketForOperatorSet(operatorSetId uint32) (string, error) {
	os, err := opi.GetOperatorSet(operatorSetId)
	if err != nil {
		return "", fmt.Errorf("failed to get socket for operator set %d: %w", operatorSetId, err)
	}
	return os.NetworkAddress, nil
}

func (opi *OperatorPeerInfo) IncludesOperatorSetId(operatorSetId uint32) bool {
	for _, os := range opi.OperatorSets {
		if os.OperatorSetID == operatorSetId {
			return true
		}
	}
	return false
}

func (opi *OperatorPeerInfo) Clone() (*OperatorPeerInfo, error) {
	clonedOperatorSets := make([]*OperatorSet, len(opi.OperatorSets))
	for i, os := range opi.OperatorSets {
		clonedSet, err := os.Clone()
		if err != nil {
			return nil, fmt.Errorf("failed to clone operator set: %w", err)
		}
		clonedOperatorSets[i] = clonedSet
	}
	return &OperatorPeerInfo{
		OperatorAddress: opi.OperatorAddress,
		OperatorSets:    clonedOperatorSets,
	}, nil
}

type IPeeringDataFetcher interface {
	ListExecutorOperators(ctx context.Context, avsAddress string) ([]*OperatorPeerInfo, error)
	ListAggregatorOperators(ctx context.Context, avsAddress string) ([]*OperatorPeerInfo, error)
}

type IPeeringDataFetcherFactory interface {
	CreatePeeringDataFetcher() (IPeeringDataFetcher, error)
}
