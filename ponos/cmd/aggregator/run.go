package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	aggregatorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	executorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/executionManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/lifecycle/runnable"
	ethereumPoller "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chain/poller/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chain/writer/simulatedChainWriter"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/loadGenerator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/fetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/simulators/simulatedExecutor/service"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
		})

		if Config.LoadGenConfig.Enabled {
			go func() {
				loadGenerator.NewLoadGenerator(&Config.LoadGenConfig, log).Run(cmd.Context())
			}()
		}

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

	for _, chain := range cfg.Chains {
		if cfg.SimulationConfig.Enabled {
			//	port := cfg.SimulationConfig.Port + i + 1
			//
			//	listenerConfig := &simulated.SimulatedChainPollerConfig{
			//		ChainId:      &chain.ChainID,
			//		Port:         port,
			//		TaskInterval: 250 * time.Millisecond,
			//	}
			//
			//	listener := simulated.NewSimulatedChainPoller(
			//		taskQueue,
			//		listenerConfig,
			//		logger,
			//	)
			//	listeners = append(listeners, listener)
			//	logger.Sugar().Infow("Created simulated chain listener", "chainId", chain.ChainID, "port", port)
			//} else {
			ethClient := ethereum.NewClient(&ethereum.EthereumClientConfig{
				BaseUrl:   chain.RpcURL,
				BlockType: ethereum.BlockType_Latest,
			}, logger)

			parsedABI, mailboxAddr, err := loadABIAndAddress(
				cfg.SimulationConfig.ReleaseId,
				chain.Environment,
				chain.ChainID,
				"TaskMailbox")
			if err != nil {
				logger.Sugar().Fatalw("Failed to load contract ABI & mailbox address", "error", err)
			}
			listenerConfig := ethereumPoller.NewEthereumChainPollerDefaultConfig(
				chain.ChainID,
				mailboxAddr.String(),
			)
			if err != nil {
				logger.Sugar().Errorw("Error parsing ABI.", err)
				panic(err)
			}

			logParser := transactionLogParser.NewTransactionLogParser(&parsedABI, logger)
			block, err := ethClient.GetBlockByNumber(context.Background(), chain.BlockNumber)
			if err != nil {
				logger.Sugar().Errorw("Error getting block by number", "error", err)
				panic(err)
			}

			listener := ethereumPoller.NewEthereumChainPoller(
				ethClient,
				taskQueue,
				logParser,
				block,
				&parsedABI,
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
	peeringFetcher := fetcher.NewLocalPeeringDataFetcher(
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

func loadABIAndAddress(
	releaseID string,
	environment string,
	chainID config.ChainId,
	contractName string,
) (abi.ABI, common.Address, error) {
	basePath, err := os.Getwd()
	if err != nil {
		return abi.ABI{}, common.Address{}, err
	}
	abiPath := filepath.Join(
		basePath,
		"contracts",
		"abi",
		environment,
		releaseID,
	)

	abiBytes, err := os.ReadFile(filepath.Join(abiPath, fmt.Sprintf("%s.abi.json", contractName)))
	if err != nil {
		return abi.ABI{}, common.Address{}, fmt.Errorf("failed to read ABI: %w", err)
	}
	parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return abi.ABI{}, common.Address{}, fmt.Errorf("failed to parse ABI: %w", err)
	}

	chainMapPath := filepath.Join(abiPath, "chains.json")
	mapBytes, err := os.ReadFile(chainMapPath)
	if err != nil {
		return parsedABI, common.Address{}, fmt.Errorf("failed to read chains.json: %w", err)
	}

	var chainMap map[config.ChainId]map[string]string
	if err := json.Unmarshal(mapBytes, &chainMap); err != nil {
		return parsedABI, common.Address{}, fmt.Errorf("failed to parse chains.json: %w", err)
	}

	addrStr := chainMap[chainID][contractName]
	if addrStr == "" {
		return parsedABI, common.Address{}, fmt.Errorf("no address found for contract %s on chain %d", contractName, chainID)
	}

	return parsedABI, common.HexToAddress(addrStr), nil
}

func convertSimulationPeeringConfig(configs []aggregatorConfig.ExecutorPeerConfig) *fetcher.LocalPeeringDataFetcherConfig {
	return &fetcher.LocalPeeringDataFetcherConfig{
		OperatorPeers: util.Map(configs, func(config aggregatorConfig.ExecutorPeerConfig, i uint64) *peering.OperatorPeerInfo {
			return &peering.OperatorPeerInfo{
				NetworkAddress: config.NetworkAddress,
				Port:           config.Port,
				PublicKey:      config.PublicKey,
			}
		}),
	}
}
