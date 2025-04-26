package lifecycle

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"
	"go.uber.org/zap"
	"time"
)

func RunContextWithShutdown(ctx context.Context, startFunc func(ctx context.Context) error, logger *zap.Logger) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := startFunc(ctx); err != nil {
		return err
	}

	gracefulShutdownNotifier := shutdown.CreateGracefulShutdownChannel()
	done := make(chan bool)

	shutdown.ListenForShutdown(gracefulShutdownNotifier, done, func() {
		logger.Sugar().Info("Shutting down aggregator...")
		cancel()
	}, 5*time.Second, logger)

	return nil
}

func RunWithShutdown(startFunc func(ctx context.Context) error, logger *zap.Logger) error {
	return RunContextWithShutdown(context.Background(), startFunc, logger)
}
