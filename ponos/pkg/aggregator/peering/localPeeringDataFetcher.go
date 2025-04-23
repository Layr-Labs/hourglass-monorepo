package peering

import (
	"go.uber.org/zap"
)

type LocalPeeringDataFetcherConfig struct {
	Peers []*ExecutorOperatorPeerInfo
}

type LocalPeeringDataFetcher struct {
	peers  []*ExecutorOperatorPeerInfo
	logger *zap.Logger
}

func NewLocalPeeringDataFetcher(
	config *LocalPeeringDataFetcherConfig,
	logger *zap.Logger,
) *LocalPeeringDataFetcher {
	return &LocalPeeringDataFetcher{
		peers:  config.Peers,
		logger: logger,
	}
}

func (lpdf *LocalPeeringDataFetcher) ListExecutorOperators() ([]*ExecutorOperatorPeerInfo, error) {
	return lpdf.peers, nil
}
