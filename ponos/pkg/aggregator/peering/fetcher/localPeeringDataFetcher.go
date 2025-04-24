package fetcher

import (
	"go.uber.org/zap"
)

type LocalPeeringDataFetcherConfig[T any] struct {
	Peers []*T
}

type LocalPeeringDataFetcher[T any] struct {
	peers  []*T
	logger *zap.Logger
}

func NewLocalPeeringDataFetcher[T any](
	config *LocalPeeringDataFetcherConfig[T],
	logger *zap.Logger,
) *LocalPeeringDataFetcher[T] {
	return &LocalPeeringDataFetcher[T]{
		peers:  config.Peers,
		logger: logger,
	}
}

func (lpdf *LocalPeeringDataFetcher[T]) ListExecutorOperators() ([]*T, error) {
	return lpdf.peers, nil
}
