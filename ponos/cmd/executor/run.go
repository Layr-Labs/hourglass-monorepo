package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/fauxSigner"
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

		fmt.Printf("Executor config: %+v\n", Config)
		if err := Config.Validate(); err != nil {
			return err
		}

		l.Sugar().Infow("executor run")

		// TODO(seanmcgary): implement a real signer at some point
		fakeSigner := fauxSigner.NewFauxSigner()

		baseRpcServer, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{}, l)
		if err != nil {
			l.Sugar().Fatal("Failed to setup RPC server", zap.Error(err))
		}

		exec := executor.NewExecutor(Config, baseRpcServer, l, fakeSigner)

		if err := exec.Initialize(); err != nil {
			l.Sugar().Fatalw("Failed to initialize executor", zap.Error(err))
		}

		ctx, cancel := context.WithCancel(context.Background())

		if err := exec.BootPerformers(ctx); err != nil {
			l.Sugar().Fatalw("Failed to boot performers", zap.Error(err))
		}

		go func() {
			if err := baseRpcServer.Start(ctx); err != nil {
				l.Sugar().Fatal("Failed to start RPC server", zap.Error(err))
			}
		}()

		go func() {
			exec.Run()
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
