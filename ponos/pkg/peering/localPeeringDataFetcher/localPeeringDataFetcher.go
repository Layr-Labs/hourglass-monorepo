package localPeeringDataFetcher

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"go.uber.org/zap"
)

type LocalPeeringDataFetcherConfig struct {
	OperatorPeers   []*peering.OperatorPeerInfo
	AggregatorPeers []*peering.OperatorPeerInfo
}

type LocalPeeringDataFetcher struct {
	operatorPeers   []*peering.OperatorPeerInfo
	aggregatorPeers []*peering.OperatorPeerInfo
	logger          *zap.Logger
}

func NewLocalPeeringDataFetcher(
	config *LocalPeeringDataFetcherConfig,
	logger *zap.Logger,
) *LocalPeeringDataFetcher {
	return &LocalPeeringDataFetcher{
		operatorPeers:   config.OperatorPeers,
		aggregatorPeers: config.AggregatorPeers,
		logger:          logger,
	}
}

func (lpdf *LocalPeeringDataFetcher) ListExecutorOperators() ([]*peering.OperatorPeerInfo, error) {
	return lpdf.operatorPeers, nil
}

func (lpdf *LocalPeeringDataFetcher) ListAggregatorOperators() ([]*peering.OperatorPeerInfo, error) {
	return lpdf.aggregatorPeers, nil
}
