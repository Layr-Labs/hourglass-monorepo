package avsExecutionManager

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
)

type IAggregationStrategy interface {
	IsLeaderForBlock(ctx context.Context, block *ethereum.EthereumBlock) (bool, error)
}

// SingleAvsAggregationStrategy is a simple aggregation strategy that assumes there is only one aggregator for the AVS
// and is always the leader for any block.
type SingleAvsAggregationStrategy struct{}

func NewSingleAvsAggregationStrategy() *SingleAvsAggregationStrategy {
	return &SingleAvsAggregationStrategy{}
}
func (s *SingleAvsAggregationStrategy) IsLeaderForBlock(ctx context.Context, block *ethereum.EthereumBlock) (bool, error) {
	return true, nil
}
