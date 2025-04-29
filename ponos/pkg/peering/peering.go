package peering

type OperatorPeerInfo struct {
	NetworkAddress  string
	Port            int
	PublicKey       string
	OperatorAddress string
	OperatorSetId   uint64
}

type IPeeringDataFetcher interface {
	ListExecutorOperators() ([]*OperatorPeerInfo, error)
	ListAggregatorOperators() ([]*OperatorPeerInfo, error)
}

type IPeeringDataFetcherFactory interface {
	CreatePeeringDataFetcher() (IPeeringDataFetcher, error)
}
