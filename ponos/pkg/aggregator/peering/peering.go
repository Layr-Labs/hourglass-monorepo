package peering

type ExecutorOperatorPeerInfo struct {
	NetworkAddress string
	Port           int
	PublicKey      string
}

type IPeeringDataFetcher interface {
	ListExecutorOperators() ([]*ExecutorOperatorPeerInfo, error)
}
