package aggregator

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/coordinator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/executionManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/lifecycle"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/workQueue"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainListener/ethereumChainListener"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainListener/simulatedChainListener"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainWriter/simulatedChainWriter"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
	"time"
)

type Aggregator struct {
	config           *aggregatorConfig.AggregatorConfig
	logger           *zap.Logger
	listeners        []lifecycle.Lifecycle
	writers          []lifecycle.Lifecycle
	coordinator      lifecycle.Lifecycle
	executionManager lifecycle.Lifecycle
}

func NewAggregator(config *aggregatorConfig.AggregatorConfig, logger *zap.Logger) *Aggregator {
	taskQueue := workQueue.NewWorkQueue[types.Task]()
	resultQueue := workQueue.NewWorkQueue[types.TaskResult]()

	listeners := buildListeners(config, logger, taskQueue)

	writer := simulatedChainWriter.NewSimulatedChainWriter(
		&simulatedChainWriter.SimulatedChainWriterConfig{Interval: 50 * time.Millisecond},
		logger,
		resultQueue,
	)
	writers := []lifecycle.Lifecycle{writer}

	ponosExecutionManager := executionManager.NewInMemorySimulatedExecutionManager()
	ponosCoordinator := coordinator.NewPonosCoordinator(taskQueue, resultQueue, ponosExecutionManager, logger)

	return &Aggregator{
		config:           config,
		logger:           logger,
		listeners:        listeners,
		writers:          writers,
		coordinator:      ponosCoordinator,
		executionManager: ponosExecutionManager,
	}
}

func (a *Aggregator) Start(ctx context.Context) error {
	a.logger.Sugar().Infow("Starting aggregator...")

	startAll := func(components []lifecycle.Lifecycle, name string) error {
		for _, c := range components {
			a.logger.Sugar().Infow("Starting component", "group", name, "type", fmt.Sprintf("%T", c))
			if err := c.Start(ctx); err != nil {
				return fmt.Errorf("failed to start %s: %w", name, err)
			}
		}
		return nil
	}

	if err := startAll(a.listeners, "listener"); err != nil {
		return err
	}
	if err := startAll(a.writers, "writer"); err != nil {
		return err
	}
	if err := a.executionManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start executionManager: %w", err)
	}
	if err := a.coordinator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start coordinator: %w", err)
	}

	a.logger.Sugar().Infow("Aggregator fully started")
	return nil
}

func (a *Aggregator) Close() {
	stopAll := func(components []lifecycle.Lifecycle, name string) {
		for _, c := range components {
			if err := c.Close(); err != nil {
				a.logger.Sugar().Warnw(
					"Failed to stop component",
					"group",
					name, "type",
					fmt.Sprintf("%T", c),
					"error",
					err,
				)
			}
		}
	}

	stopAll(a.listeners, "listener")
	stopAll(a.writers, "writer")

	if err := a.coordinator.Close(); err != nil {
		a.logger.Sugar().Warnw("Failed to stop coordinator", "error", err)
	}
	if err := a.executionManager.Close(); err != nil {
		a.logger.Sugar().Warnw("Failed to stop execution manager", "error", err)
	}

	a.logger.Sugar().Infow("Aggregator stopped")
}

func buildListeners(
	cfg *aggregatorConfig.AggregatorConfig,
	logger *zap.Logger,
	taskQueue workQueue.IInputQueue[types.Task],
) []lifecycle.Lifecycle {
	var listeners []lifecycle.Lifecycle

	for i, chain := range cfg.Chains {
		chainId := chain.ChainID

		if cfg.Simulated {
			port := cfg.SimulatedPort + i
			listener := simulatedChainListener.NewSimulatedChainListener(
				taskQueue,
				&simulatedChainListener.SimulatedChainListenerConfig{
					Port: port,
				},
				logger,
				&chainId,
			)
			listeners = append(listeners, listener)
			logger.Sugar().Infow("Using simulated chain listener", "chainId", chainId, "port", port)
		} else {
			ethClient := ethereum.NewClient(&ethereum.EthereumClientConfig{
				BaseUrl:   chain.RpcURL,
				BlockType: ethereum.BlockType_Latest,
			}, logger)

			listener := ethereumChainListener.NewEthereumChainListener(
				ethClient,
				logger,
				taskQueue,
				"ethereum-mainnet-inbox",
				chain.ChainID,
			)
			listeners = append(listeners, listener)
			logger.Sugar().Infow("Using Ethereum chain listener", "chainId", chainId, "url", chain.RpcURL)
		}
	}

	return listeners
}
