package avsAggregator

import (
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"go.uber.org/zap"
)

type ExecutorOperatorPeer struct {
}

type AvsAggregatorConfig struct {
	EnabledChains []config.ChainID
	AvsAddress    string
}

// AvsAggregator represents an aggregator instance for a specific AVS/chain combination.
//
// It knows how to handle the following responsibilities:
//   - Discover peering data on chain (for each configured chain) for AVS operators
//   - Connecting to Executor peers
//   - Distributing tasks to Executors
//   - Handle adding/removing peers from the peer list based on on-chain events when
//     operators are added/removed from the operator set
//   - Aggregates results from the Executors for a given task and returns the result to the
type AvsAggregator struct {
	config             *AvsAggregatorConfig
	logger             *zap.Logger
	peeringDataFetcher peering.IPeeringDataFetcher
}

func NewAvsAggregator(
	config *AvsAggregatorConfig,
	logger *zap.Logger,
	peeringDataFetcher peering.IPeeringDataFetcher,
) *AvsAggregator {
	return &AvsAggregator{
		config:             config,
		logger:             logger,
		peeringDataFetcher: peeringDataFetcher,
	}
}

func (aa *AvsAggregator) Initialize() error {
	peers, err := aa.peeringDataFetcher.ListExecutorOperators()
	if err != nil {
		aa.logger.Sugar().Errorf("failed to fetch executor operator peers: %v", err)
		return err
	}
	fmt.Printf("Executor operator peers: %v", peers)
	return nil
}

func (aa *AvsAggregator) DistributeNewTask(chainId config.ChainID, task interface{}) (interface{}, error) {
	return nil, nil
}

// FetchExecutorPeerData fetches the executor peer data from the AVSRegistrar contract
func (aa *AvsAggregator) FetchExecutorPeerData() {

}

// PeerJoined when an event is received that a new Executor operator has joined the operator set
func (aa *AvsAggregator) PeerJoined() {

}

// PeerLeft when an event is received that an Executor operator has left the operator set
func (aa *AvsAggregator) PeerLeft() {

}
