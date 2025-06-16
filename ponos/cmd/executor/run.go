package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/localPeeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/simulations/peers"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"time"
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

		// Load up the keystore
		var err error
		var storedKeys *keystore.EIP2335Keystore
		if Config.Operator.SigningKeys.BLS.Keystore != "" {
			storedKeys, err = keystore.ParseKeystoreJSON(Config.Operator.SigningKeys.BLS.Keystore)
			if err != nil {
				return fmt.Errorf("failed to parse keystore JSON: %w", err)
			}
		} else {
			storedKeys, err = keystore.LoadKeystoreFile(Config.Operator.SigningKeys.BLS.KeystoreFile)
			if err != nil {
				return fmt.Errorf("failed to load keystore file: '%s' %w", Config.Operator.SigningKeys.BLS.KeystoreFile, err)
			}
		}

		privateSigningKey, err := storedKeys.GetBN254PrivateKey(Config.Operator.SigningKeys.BLS.Password)
		if err != nil {
			return fmt.Errorf("failed to get private key: %w", err)
		}

		sig := inMemorySigner.NewInMemorySigner(privateSigningKey)

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

		protocolContractAddresses, err := config.GetCoreContractsForChainId(Config.L1Chain.ChainId)
		if err != nil {
			l.Sugar().Fatalw("Failed to get protocol contract addresses", zap.Error(err))
		}

		var pdf peering.IPeeringDataFetcher
		if Config.Simulation != nil && Config.Simulation.SimulatePeering != nil && Config.Simulation.SimulatePeering.Enabled {
			simulatedPeers, err := peers.NewSimulatedPeersFromConfig(Config.Simulation.SimulatePeering.AggregatorPeers)
			if err != nil {
				l.Sugar().Fatalw("Failed to create simulated peers", zap.Error(err))
			}
			pdf = localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
				AggregatorPeers: simulatedPeers,
			}, l)
		} else {
			ethereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
				BaseUrl: Config.L1Chain.RpcUrl,
			}, l)

			mailboxContract := util.Find(imContractStore.ListContracts(), func(c *contracts.Contract) bool {
				return c.ChainId == Config.L1Chain.ChainId && c.Name == config.ContractName_TaskMailbox
			})
			if mailboxContract == nil {
				return fmt.Errorf("task mailbox contract not found")
			}

			cc, err := caller.NewContractCallerFromEthereumClient(&caller.ContractCallerConfig{
				PrivateKey:          "",
				AVSRegistrarAddress: Config.AvsPerformers[0].AVSRegistrarAddress,
				TaskMailboxAddress:  mailboxContract.Address,
				KeyRegistrarAddress: protocolContractAddresses.KeyRegistrar,
			}, ethereumClient, l)
			if err != nil {
				return fmt.Errorf("failed to initialize contract caller: %w", err)
			}

			pdf = peeringDataFetcher.NewPeeringDataFetcher(cc, l)
		}

		exec := executor.NewExecutor(Config, baseRpcServer, l, sig, pdf)

		if err := exec.Initialize(); err != nil {
			l.Sugar().Fatalw("Failed to initialize executor", zap.Error(err))
		}

		ctx, cancel := context.WithCancel(context.Background())

		if err := exec.BootPerformers(ctx); err != nil {
			l.Sugar().Fatalw("Failed to boot performers", zap.Error(err))
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
