package peering

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
)

type OperatorPeerInfo struct {
	NetworkAddress  string           `json:"networkAddress"`
	PublicKey       *bn254.PublicKey `json:"publicKey"`
	OperatorAddress string           `json:"operatorAddress"`
	OperatorSetIds  []uint32         `json:"operatorSetIds"`
}

func (opi *OperatorPeerInfo) Copy() (*OperatorPeerInfo, error) {
	operatorSetIds := make([]uint32, len(opi.OperatorSetIds))
	copy(operatorSetIds, opi.OperatorSetIds)
	clonedPubKey, err := bn254.NewPublicKeyFromBytes(opi.PublicKey.Bytes())
	if err != nil {
		return nil, err
	}
	return &OperatorPeerInfo{
		NetworkAddress:  opi.NetworkAddress,
		PublicKey:       clonedPubKey,
		OperatorAddress: opi.OperatorAddress,
		OperatorSetIds:  operatorSetIds,
	}, nil
}

type IPeeringDataFetcher interface {
	ListExecutorOperators(ctx context.Context, avsAddress string) ([]*OperatorPeerInfo, error)
	ListAggregatorOperators(ctx context.Context, avsAddress string) ([]*OperatorPeerInfo, error)
}

type IPeeringDataFetcherFactory interface {
	CreatePeeringDataFetcher() (IPeeringDataFetcher, error)
}
