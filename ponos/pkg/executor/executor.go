package executor

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/server"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executorRpcServer"
	"go.uber.org/zap"
)

type Executor struct {
	logger        *zap.Logger
	config        *executorConfig.ExecutorConfig
	avsPerformers map[string]avsPerformer.IAvsPerformer
}

func NewExecutor(
	config *executorConfig.ExecutorConfig,
	rpcServer *executorRpcServer.ExecutorRpcServer,
	logger *zap.Logger,
) *Executor {
	return &Executor{
		logger:        logger,
		config:        config,
		avsPerformers: make(map[string]avsPerformer.IAvsPerformer),
	}
}

func (w *Executor) Initialize() error {
	w.logger.Sugar().Infow("Initializing AVS performers")

	for _, avs := range w.config.AvsPerformers {
		if _, ok := w.avsPerformers[avs.AvsAddress]; ok {
			w.logger.Sugar().Errorw("AVS performer already exists",
				zap.String("avsAddress", avs.AvsAddress),
				zap.String("processType", avs.ProcessType),
			)
		}

		switch avs.ProcessType {
		case string(avsPerformer.AvsProcessTypeServer):
			performer, err := server.NewAvsPerformerServer(&avsPerformer.AvsPerformerConfig{
				AvsAddress:  avs.AvsAddress,
				ProcessType: avsPerformer.AvsProcessType(avs.ProcessType),
				Image:       avsPerformer.PerformerImage{Repository: avs.Image.Repository, Tag: avs.Image.Tag},
			}, w.logger)
			if err != nil {
				w.logger.Sugar().Errorw("Failed to create AVS performer server",
					zap.String("avsAddress", avs.AvsAddress),
					zap.Error(err),
				)
				return fmt.Errorf("failed to create AVS performer server: %v", err)
			}
			w.avsPerformers[avs.AvsAddress] = performer

		default:
			w.logger.Sugar().Errorw("Unsupported AVS performer process type",
				zap.String("avsAddress", avs.AvsAddress),
				zap.String("processType", avs.ProcessType),
			)
			return fmt.Errorf("unsupported AVS performer process type: %s", avs.ProcessType)
		}
	}
	return nil
}

func (w *Executor) BootPerformers(ctx context.Context) error {
	w.logger.Sugar().Infow("Booting AVS performers")
	for avsAddress, performer := range w.avsPerformers {
		if err := performer.Initialize(ctx); err != nil {
			w.logger.Sugar().Errorw("Failed to initialize AVS performer",
				zap.String("avsAddress", avsAddress),
				zap.Error(err),
			)
			return fmt.Errorf("failed to initialize AVS performer: %v", err)
		}
	}
	go func() {
		select {
		case <-ctx.Done():
			w.logger.Sugar().Info("Shutting down AVS performers")
			for avsAddress, performer := range w.avsPerformers {
				if err := performer.Shutdown(); err != nil {
					w.logger.Sugar().Errorw("Failed to shutdown AVS performer",
						zap.String("avsAddress", avsAddress),
						zap.Error(err),
					)
				}
			}
		}
	}()
	return nil
}

func (w *Executor) Run() {
	w.logger.Info("Worker node is running", zap.String("version", "1.0.0"))
}
