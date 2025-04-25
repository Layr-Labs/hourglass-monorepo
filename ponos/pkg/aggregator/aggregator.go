package aggregator

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/lifecycle"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/lifecycle/runnable"
	"go.uber.org/zap"
)

type Aggregator struct {
	chainPollers     []runnable.IRunnable
	chainWriters     []runnable.IRunnable
	executionManager runnable.IRunnable
	logger           *zap.Logger
	inboxAddress     string
}

type AggregatorConfig struct {
	ChainPollers     []runnable.IRunnable
	ChainWriters     []runnable.IRunnable
	ExecutionManager runnable.IRunnable
	Logger           *zap.Logger
	InboxAddress     string
}

func NewAggregator(params *AggregatorConfig) *Aggregator {
	return &Aggregator{
		chainPollers:     params.ChainPollers,
		chainWriters:     params.ChainWriters,
		executionManager: params.ExecutionManager,
		logger:           params.Logger,
		inboxAddress:     params.InboxAddress,
	}
}

func (a *Aggregator) Start(ctx context.Context) error {
	sugar := a.logger.Sugar()
	sugar.Infow("Starting aggregator with graceful shutdown support...")
	return lifecycle.RunContextWithShutdown(ctx, a.startAggregator, a.logger)
}

func (a *Aggregator) startAggregator(ctx context.Context) error {

	if err := a.executionManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start executionManager: %w", err)
	}

	for index, listener := range a.chainPollers {
		err := listener.Start(ctx)
		if err != nil && ctx.Err() == nil {
			a.logger.Sugar().Errorw("Listener failed", "index", index, "error", err)
			return err
		}
	}

	for index, writer := range a.chainWriters {
		err := writer.Start(ctx)
		if err != nil && ctx.Err() == nil {
			a.logger.Sugar().Errorw("Writer failed", "index", index, "error", err)
			return err
		}
	}

	a.logger.Sugar().Infow("Aggregator fully started")
	return nil
}
