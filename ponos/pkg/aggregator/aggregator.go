package aggregator

import (
	"context"
	"fmt"
	aggregatorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	executorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/executionManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/lifecycle"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/workQueue"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainListener/ethereumChainListener"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainListener/simulatedChainListener"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainWriter/simulatedChainWriter"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
	"time"
)

type Aggregator struct {
	config           *aggregatorConfig.AggregatorConfig
	logger           *zap.Logger
	listeners        []lifecycle.Lifecycle
	writers          []lifecycle.Lifecycle
	executionManager lifecycle.Lifecycle
	executors        []lifecycle.Lifecycle
}

func NewAggregator(config *aggregatorConfig.AggregatorConfig, logger *zap.Logger) *Aggregator {
	taskQueue := workQueue.NewWorkQueue[types.Task]()
	resultQueue := workQueue.NewWorkQueue[types.TaskResult]()

	listeners := buildListeners(config, logger, taskQueue)

	// TODO: implement ponosChainWriter and use here.
	writer := simulatedChainWriter.NewSimulatedChainWriter(
		&simulatedChainWriter.SimulatedChainWriterConfig{Interval: 5 * time.Millisecond},
		logger,
		resultQueue,
	)
	writers := []lifecycle.Lifecycle{writer}
	peeringFetcher := peering.NewLocalPeeringDataFetcher(
		convertSimulationPeeringConfig(config.SimulationConfig.ExecutorPeerConfigs),
		logger,
	)
	var executors []lifecycle.Lifecycle
	if config.SimulationConfig.Enabled {
		aggregatorUrl := fmt.Sprintf("localhost:%d", config.SimulationConfig.Port)
		executors = append(executors, buildExecutors(
			config.SimulationConfig.ExecutorPeerConfigs,
			aggregatorUrl,
			config.SimulationConfig.SecureConnection,
			logger,
		)...)
	}
	aggregatorServer, err := loadAggregatorServer(config, logger)
	if err != nil {
		panic(err)
	}
	ponosExecutionManager := executionManager.NewPonosExecutionManager(
		aggregatorServer,
		taskQueue,
		resultQueue,
		peeringFetcher,
		logger,
	)
	return &Aggregator{
		config:           config,
		logger:           logger,
		listeners:        listeners,
		writers:          writers,
		executionManager: ponosExecutionManager,
		executors:        executors,
	}
}

func (a *Aggregator) Start(ctx context.Context) error {
	a.logger.Sugar().Infow("Starting aggregator...")

	if err := a.executionManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start executionManager: %w", err)
	}

	if err := lifecycle.StartAll(a.listeners, ctx, a.logger, "listener"); err != nil {
		return err
	}

	if err := lifecycle.StartAll(a.writers, ctx, a.logger, "writer"); err != nil {
		return err
	}

	if err := lifecycle.StartAll(a.executors, ctx, a.logger, "executor"); err != nil {
		return err
	}
	a.logger.Sugar().Infow("Aggregator fully started")
	return nil
}

func (a *Aggregator) Close() {
	lifecycle.StopAll(a.listeners, a.logger, "listener")

	if err := a.executionManager.Close(); err != nil {
		a.logger.Sugar().Warnw("Failed to stop execution manager", "error", err)
	}
	lifecycle.StopAll(a.writers, a.logger, "writer")
	lifecycle.StopAll(a.executors, a.logger, "executor")

	a.logger.Sugar().Infow("Aggregator stopped")
}

func loadAggregatorServer(config *aggregatorConfig.AggregatorConfig, logger *zap.Logger) (*rpcServer.RpcServer, error) {
	if config.SimulationConfig.Enabled {
		return rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{GrpcPort: config.SimulationConfig.Port}, logger)
	} else {
		return rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{GrpcPort: config.ServerConfig.Port}, logger)
	}
}

func buildListeners(
	cfg *aggregatorConfig.AggregatorConfig,
	logger *zap.Logger,
	taskQueue workQueue.IInputQueue[types.Task],
) []lifecycle.Lifecycle {
	var listeners []lifecycle.Lifecycle

	for i, chain := range cfg.Chains {
		chainId := chain.ChainID
		if cfg.SimulationConfig.Enabled {
			port := cfg.SimulationConfig.Port + i + 1
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

func buildExecutors(
	configs []aggregatorConfig.ExecutorPeerConfig,
	aggregatorUrl string,
	secConnection bool,
	logger *zap.Logger,
) []lifecycle.Lifecycle {
	var executors []lifecycle.Lifecycle

	for _, config := range configs {
		port := config.Port

		rpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{GrpcPort: port}, logger)
		if err != nil {
			panic(fmt.Errorf("failed to create rpcServer for executor: %w", err))
		}
		clientConn, err := clients.NewGrpcClient(aggregatorUrl, secConnection)
		if err != nil {
			panic(fmt.Errorf("failed to create aggregator client: %w", err))
		}
		aggregatorClient := aggregatorpb.NewAggregatorServiceClient(clientConn)

		exe := executor.NewSimulatedExecutorServer(rpc, aggregatorClient, config.PublicKey)
		executorpb.RegisterExecutorServiceServer(rpc.GetGrpcServer(), exe)

		executors = append(executors, exe)
	}

	return executors
}

func convertSimulationPeeringConfig(configs []aggregatorConfig.ExecutorPeerConfig) *peering.LocalPeeringDataFetcherConfig {
	var infos []*peering.ExecutorOperatorPeerInfo
	for _, config := range configs {
		info := &peering.ExecutorOperatorPeerInfo{
			NetworkAddress: config.NetworkAddress,
			Port:           config.Port,
			PublicKey:      config.PublicKey,
		}
		infos = append(infos, info)
	}
	return &peering.LocalPeeringDataFetcherConfig{
		Peers: infos,
	}
}
