package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/go-ponos/pkg/config"
	"github.com/Layr-Labs/go-ponos/pkg/executorRpcServer"
	"github.com/Layr-Labs/go-ponos/pkg/logger"
	"github.com/Layr-Labs/go-ponos/pkg/rpcServer"
	"github.com/Layr-Labs/go-ponos/pkg/shutdown"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"time"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the worker",
	Run: func(cmd *cobra.Command, args []string) {
		initRunCmd(cmd)

		l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})

		l.Sugar().Infow("worker run")

		baseRpcServer, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{}, l)
		if err != nil {
			l.Sugar().Fatal("Failed to setup RPC server", zap.Error(err))
		}

		_, err = executorRpcServer.NewExecutorRpcServer(baseRpcServer, l)
		if err != nil {
			l.Sugar().Fatal("Failed to setup executor RPC server", zap.Error(err))
		}

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			if err := baseRpcServer.Start(ctx); err != nil {
				l.Sugar().Fatal("Failed to start RPC server", zap.Error(err))
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
