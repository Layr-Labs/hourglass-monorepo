package lifecycle

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"
	"go.uber.org/zap"
	"time"
)

func RunWithShutdown(startFunc func(ctx context.Context) error, logger *zap.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
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

func StopAll(components []Lifecycle, logger *zap.Logger, name string) {
	for _, c := range components {
		if err := c.Close(); err != nil {
			// TODO: emit metric
			logger.Sugar().Warnw(
				"Failed to stop component",
				"group",
				name, "type",
				fmt.Sprintf("%T", c),
				"error",
				err,
			)
		}
	}
}

func StartAll(components []Lifecycle, ctx context.Context, logger *zap.Logger, name string) error {
	for _, c := range components {
		logger.Sugar().Infow("Starting component", "group", name, "type", fmt.Sprintf("%T", c))
		if err := c.Start(ctx); err != nil {
			// TODO: emit metric
			return fmt.Errorf("failed to start %s: %w", name, err)
		}
	}
	return nil
}
