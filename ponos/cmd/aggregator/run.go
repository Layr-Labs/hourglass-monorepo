package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/lifecycle/runnable"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/ethereumChainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/simulatedChainPoller"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/executionManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/peering/fetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainWriter/simulatedChainWriter"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/simulations/executor/service"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"

	aggregatorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	executorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	inboxAddress = "ethereum-mainnet-inbox"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the aggregator",
	RunE: func(cmd *cobra.Command, args []string) error {
		initRunCmd(cmd)
		log, _ := logger.NewLogger(&logger.LoggerConfig{Debug: Config.Debug})
		sugar := log.Sugar()

		if err := Config.Validate(); err != nil {
			sugar.Errorw("Invalid configuration", "error", err)
			return err
		}

		sugar.Infof("Aggregator config: %+v\n", Config)
		sugar.Infow("Building aggregator components...")

		taskQueue := make(chan *types.Task, 100)
		resultQueue := make(chan *types.TaskResult, 100)

		if Config.SimulationConfig.Enabled {
			sugar.Infow("Starting simulation executors...")
			executors := buildExecutors(cmd.Context(), &Config, log)

			for i, executor := range executors {
				if err := executor.Start(cmd.Context()); err != nil {
					sugar.Errorw("Failed to start executor", "index", i, "error", err)
				}
			}
		}

		listeners := buildListeners(&Config, taskQueue, log)
		writers := buildWriters(resultQueue, log)
		execManager := buildExecutionManager(&Config, taskQueue, resultQueue, log)

		agg := aggregator.NewAggregator(&aggregator.AggregatorConfig{
			Logger:           log,
			ChainPollers:     listeners,
			ChainWriters:     writers,
			ExecutionManager: execManager,
			InboxAddress:     inboxAddress,
		})

		if err := agg.Start(cmd.Context()); err != nil {
			return err
		}

		sugar.Infow("Context cancelled, shutting down aggregator")
		return nil
	},
}

func initRunCmd(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err := viper.BindPFlag(f.Name, f); err != nil {
			fmt.Printf("Failed to bind flag '%s': %+v\n", f.Name, err)
		}
		if err := viper.BindEnv(f.Name); err != nil {
			fmt.Printf("Failed to bind env '%s': %+v\n", f.Name, err)
		}
	})
}

func buildListeners(cfg *aggregatorConfig.AggregatorConfig, taskQueue chan *types.Task, logger *zap.Logger) []runnable.IRunnable {
	var listeners []runnable.IRunnable

	for i, chain := range cfg.Chains {
		if cfg.SimulationConfig.Enabled {
			port := cfg.SimulationConfig.Port + i + 1

			listenerConfig := &simulatedChainPoller.SimulatedChainPollerConfig{
				ChainId:         &chain.ChainID,
				Port:            port,
				PollingInterval: 500 * time.Millisecond,
				TaskInterval:    250 * time.Millisecond,
			}

			listener := simulatedChainPoller.NewSimulatedChainPoller(
				taskQueue,
				listenerConfig,
				logger,
			)
			listeners = append(listeners, listener)
			logger.Sugar().Infow("Created simulated chain listener", "chainId", chain.ChainID, "port", port)
		} else {
			ethClient := ethereum.NewClient(&ethereum.EthereumClientConfig{
				BaseUrl:   chain.RpcURL,
				BlockType: ethereum.BlockType_Latest,
			}, logger)

			listenerConfig := ethereumChainPoller.NewEthereumChainPollerDefaultConfig(
				chain.ChainID,
				"ethereum-mainnet-inbox",
			)

			listener := ethereumChainPoller.NewEthereumChainPoller(
				ethClient,
				taskQueue,
				listenerConfig,
				logger,
			)
			listeners = append(listeners, listener)
			logger.Sugar().Infow("Created Ethereum chain listener", "chainId", chain.ChainID, "url", chain.RpcURL)
		}
	}

	return listeners
}

func buildWriters(resultQueue chan *types.TaskResult, logger *zap.Logger) []runnable.IRunnable {
	// TODO: implement ponosChainWriter and use when appropriate
	writerConfig := &simulatedChainWriter.SimulatedChainWriterConfig{
		Interval: 1 * time.Millisecond,
	}

	writer := simulatedChainWriter.NewSimulatedChainWriter(
		writerConfig,
		resultQueue,
		logger,
	)

	logger.Sugar().Infow("Created simulated chain writer")
	return []runnable.IRunnable{writer}
}

func buildExecutionManager(
	cfg *aggregatorConfig.AggregatorConfig,
	taskQueue chan *types.Task,
	resultQueue chan *types.TaskResult,
	logger *zap.Logger,
) runnable.IRunnable {
	peeringFetcher := fetcher.NewLocalPeeringDataFetcher[peering.ExecutorOperatorPeerInfo](
		convertSimulationPeeringConfig(cfg.SimulationConfig.ExecutorPeerConfigs),
		logger,
	)

	aggregatorServer, err := loadAggregatorServer(cfg, logger)
	if err != nil {
		logger.Sugar().Fatalw("Failed to create aggregator server", "error", err)
	}

	emConfig := executionManager.NewPonosExecutionManagerDefaultConfig()
	emConfig.SecureConnection = cfg.ServerConfig.SecureConnection

	manager := executionManager.NewPonosExecutionManager(
		aggregatorServer,
		taskQueue,
		resultQueue,
		peeringFetcher,
		emConfig,
		logger,
	)

	logger.Sugar().Infow("Created execution manager")
	return manager
}

func buildExecutors(ctx context.Context, cfg *aggregatorConfig.AggregatorConfig, logger *zap.Logger) []runnable.IRunnable {
	var executors []runnable.IRunnable
	aggregatorUrl := fmt.Sprintf("localhost:%d", cfg.SimulationConfig.Port)

	for _, config := range cfg.SimulationConfig.ExecutorPeerConfigs {
		port := config.Port

		rpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{GrpcPort: port}, logger)
		if err != nil {
			logger.Sugar().Fatalw("Failed to create rpcServer for executor", "error", err)
		}

		clientConn, err := clients.NewGrpcClient(aggregatorUrl, cfg.SimulationConfig.SecureConnection)
		if err != nil {
			logger.Sugar().Fatalw("Failed to create aggregator client", "error", err)
		}

		aggregatorClient := aggregatorpb.NewAggregatorServiceClient(clientConn)
		exe := service.NewSimulatedExecutorServer(rpc, aggregatorClient, config.PublicKey)
		executorpb.RegisterExecutorServiceServer(rpc.GetGrpcServer(), exe)
		err = rpc.Start(ctx)
		if err != nil {
			logger.Sugar().Fatalw("Failed to start executor", "error", err)
			panic(err)
		}

		executors = append(executors, exe)
		logger.Sugar().Infow("Created simulated executor", "publicKey", config.PublicKey, "port", port)
	}

	return executors
}

func loadAggregatorServer(cfg *aggregatorConfig.AggregatorConfig, logger *zap.Logger) (*rpcServer.RpcServer, error) {
	if cfg.SimulationConfig.Enabled {
		return rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{GrpcPort: cfg.SimulationConfig.Port}, logger)
	} else {
		return rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{GrpcPort: cfg.ServerConfig.Port}, logger)
	}
}

func convertSimulationPeeringConfig(
	configs []aggregatorConfig.ExecutorPeerConfig,
) *fetcher.LocalPeeringDataFetcherConfig[peering.ExecutorOperatorPeerInfo] {
	var infos []*peering.ExecutorOperatorPeerInfo
	for _, config := range configs {
		info := &peering.ExecutorOperatorPeerInfo{
			NetworkAddress: config.NetworkAddress,
			Port:           config.Port,
			PublicKey:      config.PublicKey,
		}
		infos = append(infos, info)
	}
	return &fetcher.LocalPeeringDataFetcherConfig[peering.ExecutorOperatorPeerInfo]{
		Peers: infos,
	}
}
