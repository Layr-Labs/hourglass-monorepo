package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/signerUtils"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/badger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the executor",
	RunE: func(cmd *cobra.Command, args []string) error {
		initRunCmd(cmd)

		l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: Config.Debug})

		if err := Config.Validate(); err != nil {
			return err
		}

		l.Sugar().Infow("executor run")

		execSigners, err := signerUtils.ParseSignersFromOperatorConfig(Config.Operator, l)
		if err != nil {
			return fmt.Errorf("failed to parse signers from operator config: %w", err)
		}

		baseRpcServer, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
			GrpcPort: Config.GrpcPort,
		}, l)
		if err != nil {
			l.Sugar().Fatal("Failed to setup RPC server", zap.Error(err))
		}

		var coreContracts []*contracts.Contract
		if len(Config.Contracts) > 0 {
			l.Sugar().Infow("Loading core contracts from runtime config")
			coreContracts, err = eigenlayer.LoadContractsFromRuntime(string(Config.Contracts))
			if err != nil {
				return fmt.Errorf("failed to load core contracts from runtime: %w", err)
			}
		} else {
			l.Sugar().Infow("Loading core contracts from embedded config")
			coreContracts, err = eigenlayer.LoadContracts()
			if err != nil {
				return fmt.Errorf("failed to load core contracts: %w", err)
			}
		}

		imContractStore := inMemoryContractStore.NewInMemoryContractStore(coreContracts, l)

		// Allow overriding contracts from the runtime config
		if Config.OverrideContracts != nil {
			if Config.OverrideContracts.TaskMailbox != nil && len(Config.OverrideContracts.TaskMailbox.Contract) > 0 {
				overrideContract, err := eigenlayer.LoadOverrideContract(Config.OverrideContracts.TaskMailbox.Contract)
				if err != nil {
					return fmt.Errorf("failed to load override contract: %w", err)
				}
				if err := imContractStore.OverrideContract(overrideContract.Name, Config.OverrideContracts.TaskMailbox.ChainIds, overrideContract); err != nil {
					return fmt.Errorf("failed to override contract: %w", err)
				}
			}
		}

		ethereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
			BaseUrl: Config.L1Chain.RpcUrl,
		}, l)

		ethClient, err := ethereumClient.GetEthereumContractCaller()
		if err != nil {
			return fmt.Errorf("failed to get ethereum contract caller: %w", err)
		}

		mailboxContract := util.Find(imContractStore.ListContracts(), func(c *contracts.Contract) bool {
			return c.ChainId == Config.L1Chain.ChainId && c.Name == config.ContractName_TaskMailbox
		})
		if mailboxContract == nil {
			return fmt.Errorf("task mailbox contract not found")
		}

		privateKeySigner, err := transactionSigner.NewTransactionSigner(Config.Operator.OperatorPrivateKey, ethClient, l)
		if err != nil {
			return fmt.Errorf("failed to create private key signer: %w", err)
		}

		cc, err := caller.NewContractCallerFromEthereumClient(ethereumClient, privateKeySigner, l)
		if err != nil {
			return fmt.Errorf("failed to initialize contract caller: %w", err)
		}

		pdf := peeringDataFetcher.NewPeeringDataFetcher(cc, l)

		// Create storage based on configuration
		var store storage.ExecutorStore
		if Config.Storage != nil {
			switch Config.Storage.Type {
			case "memory":
				l.Sugar().Infow("Using in-memory storage")
				store = memory.NewInMemoryExecutorStore()
			case "badger":
				l.Sugar().Infow("Using BadgerDB storage")
				badgerStore, err := badger.NewBadgerExecutorStore(Config.Storage.BadgerConfig)
				if err != nil {
					return fmt.Errorf("failed to create badger store: %w", err)
				}
				store = badgerStore
			default:
				return fmt.Errorf("unknown storage type: %s", Config.Storage.Type)
			}
		} else {
			l.Sugar().Infow("No storage configured, running without persistence")
		}

		exec := executor.NewExecutor(Config, baseRpcServer, l, execSigners, pdf, cc, store)

		ctx, cancel := context.WithCancel(context.Background())

		if err := exec.Initialize(ctx); err != nil {
			l.Sugar().Fatalw("Failed to initialize executor", zap.Error(err))
		}

		go func() {
			if err := exec.Run(ctx); err != nil {
				l.Sugar().Fatal("Failed to run executor", zap.Error(err))
			}
		}()

		gracefulShutdownNotifier := shutdown.CreateGracefulShutdownChannel()
		done := make(chan bool)
		shutdown.ListenForShutdown(gracefulShutdownNotifier, done, func() {
			l.Sugar().Info("Shutting down...")
			cancel()
			// Close storage if available
			if store != nil {
				if err := store.Close(); err != nil {
					l.Sugar().Errorw("Failed to close storage", "error", err)
				}
			}
		}, time.Second*5, l)
		return nil
	},
}

func initRunCmd(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err := viper.BindPFlag(config.KebabToSnakeCase(f.Name), f); err != nil {
			fmt.Printf("Failed to bind flag '%s' - %+v\n", f.Name, err)
		}
		if err := viper.BindEnv(f.Name); err != nil {
			fmt.Printf("Failed to bind env '%s' - %+v\n", f.Name, err)
		}

	})
}
