package peering

import "context"

type OperatorPeerInfo struct {
	NetworkAddress  string   `json:"networkAddress"`
	PublicKey       string   `json:"publicKey"`
	OperatorAddress string   `json:"operatorAddress"`
	OperatorSetIds  []uint32 `json:"operatorSetIds"`
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
