package peering

type ExecutorOperatorPeerInfo struct {
	NetworkAddress string
	Port           int
	PublicKey      string
}

type IPeeringDataFetcher[T any] interface {
	ListExecutorOperators() ([]*T, error)
}
