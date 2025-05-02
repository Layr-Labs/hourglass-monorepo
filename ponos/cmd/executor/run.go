package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/localPeeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/keystore"
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

		storedKeys, err := keystore.ParseKeystoreJSON(Config.Operator.SigningKeys.BLS.Keystore)
		if err != nil {
			return fmt.Errorf("failed to parse keystore JSON: %w", err)
		}

		keyScheme, err := keystore.GetSigningSchemeForCurveType(storedKeys.CurveType)
		if err != nil {
			return fmt.Errorf("failed to get signing scheme: %w", err)
		}

		privateSigningKey, err := storedKeys.GetPrivateKey(Config.Operator.SigningKeys.BLS.Password, keyScheme)
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

		var pdf *localPeeringDataFetcher.LocalPeeringDataFetcher
		if Config.Simulation.SimulatePeering.Enabled {
			pdf = localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
				AggregatorPeers: util.Map(Config.Simulation.SimulatePeering.AggregatorPeers, func(p config.SimulatedPeer, i uint64) *peering.OperatorPeerInfo {
					return &peering.OperatorPeerInfo{
						OperatorAddress: p.OperatorAddress,
						PublicKey:       p.PublicKey,
						OperatorSetIds:  []uint32{p.OperatorSetId},
					}
				}),
			}, l)
		} else {
			return fmt.Errorf("peering data fetcher not implemented")
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
