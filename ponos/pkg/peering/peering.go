package peering

import "context"

type OperatorPeerInfo struct {
	NetworkAddress  string
	PublicKey       string
	OperatorAddress string
	OperatorSetIds  []uint32
}

type IPeeringDataFetcher interface {
	ListExecutorOperators(ctx context.Context, avsAddress string) ([]*OperatorPeerInfo, error)
	ListAggregatorOperators(ctx context.Context, avsAddress string) ([]*OperatorPeerInfo, error)
}

type IPeeringDataFetcherFactory interface {
	CreatePeeringDataFetcher() (IPeeringDataFetcher, error)
}
