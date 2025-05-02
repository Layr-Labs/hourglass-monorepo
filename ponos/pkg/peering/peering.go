package peering

import "context"

type OperatorPeerInfo struct {
	NetworkAddress  string
	PublicKey       string
	OperatorAddress string
	OperatorSetIds  []uint32
}

func (opi *OperatorPeerInfo) Copy() *OperatorPeerInfo {
	operatorSetIds := make([]uint32, len(opi.OperatorSetIds))
	copy(operatorSetIds, opi.OperatorSetIds)
	return &OperatorPeerInfo{
		NetworkAddress:  opi.NetworkAddress,
		PublicKey:       opi.PublicKey,
		OperatorAddress: opi.OperatorAddress,
		OperatorSetIds:  operatorSetIds,
	}
}

type IPeeringDataFetcher interface {
	ListExecutorOperators(ctx context.Context, avsAddress string) ([]*OperatorPeerInfo, error)
	ListAggregatorOperators(ctx context.Context, avsAddress string) ([]*OperatorPeerInfo, error)
}

type IPeeringDataFetcherFactory interface {
	CreatePeeringDataFetcher() (IPeeringDataFetcher, error)
}
