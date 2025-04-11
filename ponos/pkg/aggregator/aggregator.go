package aggregator

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/avsAggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainListener"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"go.uber.org/zap"
)

type AvsAggregatorStore struct {
	logger      *zap.Logger
	aggregators []*avsAggregator.AvsAggregator
	chains      map[config.ChainID]*avsAggregator.AvsAggregator
}

type AggregatorConfig struct {
}

// Aggregator represents the main Aggregator server instance
type Aggregator struct {
	config          *AggregatorConfig
	logger          *zap.Logger
	aggregatorStore *AvsAggregatorStore
	chainListeners  map[config.ChainID]chainListener.IChainListener
}

func NewAggregator(
	config *AggregatorConfig,
	logger *zap.Logger,
	chainListeners map[config.ChainID]chainListener.IChainListener,
) *Aggregator {
	return &Aggregator{
		config:         config,
		logger:         logger,
		chainListeners: chainListeners,
	}
}

func (a *Aggregator) Start(ctx context.Context) error {
	a.logger.Sugar().Infow("Starting Aggregator")

	if err := a.listenToChains(ctx); err != nil {
		a.logger.Sugar().Errorf("failed to start chain listeners: %v", err)
		return err
	}
	return nil
}

func (a *Aggregator) listenToChains(ctx context.Context) error {
	// Create a channel to receive events from the chain listeners
	queue := make(chan *chainListener.Event, 1000)

	for chainId, listener := range a.chainListeners {
		go func(chainId config.ChainID, chainListener chainListener.IChainListener) {
			a.logger.Sugar().Infow("Starting chain listener",
				zap.Uint("chainId", uint(chainId)),
			)
			err := chainListener.ListenForInboxEvents(ctx, queue, "")
			if err != nil {
				a.logger.Sugar().Errorf("failed to listen for inbox events: %v", err)
			}
		}(chainId, listener)
	}
	return nil
}
