package aggregator

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/avsAggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainListener"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"go.uber.org/zap"
)

type AvsAggregatorStore struct {
	//nolint:unused
	logger *zap.Logger
	//nolint:unused
	aggregators []*avsAggregator.AvsAggregator
	//nolint:unused
	chains map[config.ChainID]*avsAggregator.AvsAggregator
	//nolint:unused
	chainMessageInboxes map[config.ChainID]chan *chainListener.Event
}

type AggregatorConfig struct {
}

// Aggregator represents the main Aggregator server instance
type Aggregator struct {
	config *AggregatorConfig
	logger *zap.Logger
	//nolint:unused
	aggregatorStore *AvsAggregatorStore
	chainListeners  map[config.ChainID]chainListener.IChainListener
}

func NewAggregator(
	config *AggregatorConfig,
	chainListeners map[config.ChainID]chainListener.IChainListener,
	logger *zap.Logger,
) *Aggregator {
	return &Aggregator{
		config:         config,
		logger:         logger,
		chainListeners: chainListeners,
	}
}

func (a *Aggregator) Start(ctx context.Context) error {
	a.logger.Sugar().Infow("Starting Aggregator")

	// Startup process:
	// 1. Initialize the aggregator store using the config
	// 2. Initialize each aggregator in the store
	// 3. Start the chain listeners that will listen to events for that chain
	// 		e.g. inbox events, upgrade events, etc
	// 4. When new tasks come in from a chain, distribute to the appropriate aggregator

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

// distributeWorkForChainInbox consumes messages from the corresponding chainMessageInboxes for the chainId,
// determines which AvsAggregator to use based on the event payload, and sends it to that AvsAggregator
// by calling the `DistributeNewTask` function.
//
//nolint:unused
func (a *Aggregator) distributeWorkForChainInbox(chainId config.ChainID) {
	// 1. Distribute work to the appropriate AvsAggregator based on the chainId and avsAddress in the event
	// 2. Receive the result and post it to the corresponding outbox
}
