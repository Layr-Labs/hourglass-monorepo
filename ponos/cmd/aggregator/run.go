package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainListener"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainListener/ethereumChainListener"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainListener/simulatedChainListener"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"time"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the aggregator",
	Run: func(cmd *cobra.Command, args []string) {
		initRunCmd(cmd)

		cfg := aggregatorConfig.NewAggregatorConfig()
		fmt.Printf("Config: %+v\n", cfg)

		l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: cfg.Debug})

		l.Sugar().Infow("aggregator run")

		ctx, cancel := context.WithCancel(context.Background())

		ethereumClient := ethereum.NewClient(&ethereum.EthereumClientConfig{
			BaseUrl:   "https://special-yolo-river.ethereum-holesky.quiknode.pro/2d21099a19e7c896a22b9fcc23dc8ce80f2214a5/",
			BlockType: ethereum.BlockType_Latest,
		}, l)

		listeners := map[config.ChainID]chainListener.IChainListener{}

		if cfg.Simulated {
			listeners[config.ChainID(1)] = simulatedChainListener.NewSimulatedChainListener(&simulatedChainListener.SimulatedChainListenerConfig{
				Port: cfg.SimulatedPort,
			}, l)
			l.Sugar().Infow("Using simulated chain listener")
		} else {
			listeners[config.ChainID(1)] = ethereumChainListener.NewEthereumChainListener(ethereumClient, l)
		}

		agg := aggregator.NewAggregator(&aggregator.AggregatorConfig{}, l, listeners)

		go func() {
			if err := agg.Start(ctx); err != nil {
				l.Sugar().Fatalw("Failed to start aggregator", zap.Error(err))
			}
		}()

		gracefulShutdownNotifier := shutdown.CreateGracefulShutdownChannel()
		done := make(chan bool)
		shutdown.ListenForShutdown(gracefulShutdownNotifier, done, func() {
			l.Sugar().Info("Shutting down...")
			cancel()
		}, time.Second*5, l)
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
