package peering

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/ethereum/go-ethereum/common"
)

type WrappedPublicKey struct {
	PublicKey    signing.PublicKey `json:"publicKey"`
	ECDSAAddress common.Address    `json:"ecdsaAddress"`
}

type OperatorSet struct {
	OperatorSetID    uint32           `json:"operatorSetId"`
	OperatorIndex    uint32           `json:"operatorIndex"`
	WrappedPublicKey WrappedPublicKey `json:"publicKey"`
	NetworkAddress   string           `json:"networkAddress"`
	CurveType        config.CurveType `json:"curveType"`
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

type IPeeringDataFetcher interface {
	ListExecutorOperators(ctx context.Context, avsAddress string, blockNumber uint64) ([]*OperatorPeerInfo, error)
	ListAggregatorOperators(ctx context.Context, avsAddress string, blockNumber uint64) ([]*OperatorPeerInfo, error)
}
