package peering

type OperatorPeerInfo struct {
	NetworkAddress  string
	PublicKey       string
	OperatorAddress string
	OperatorSetIds  []uint32
}

type IPeeringDataFetcher interface {
	ListExecutorOperators() ([]*OperatorPeerInfo, error)
	ListAggregatorOperators() ([]*OperatorPeerInfo, error)
}

type IPeeringDataFetcherFactory interface {
	CreatePeeringDataFetcher() (IPeeringDataFetcher, error)
}
