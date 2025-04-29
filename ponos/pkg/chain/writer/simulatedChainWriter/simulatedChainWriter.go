package simulatedChainWriter

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
)

type SimulatedChainWriterConfig struct {
	Interval time.Duration
}

type SimulatedChainWriter struct {
	config          *SimulatedChainWriterConfig
	logger          *zap.Logger
	workOutputQueue chan *types.TaskResult
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewSimulatedChainWriter(
	config *SimulatedChainWriterConfig,
	workOutputQueue chan *types.TaskResult,
	logger *zap.Logger,
) *SimulatedChainWriter {
	return &SimulatedChainWriter{
		config:          config,
		workOutputQueue: workOutputQueue,
		logger:          logger,
	}
}

func (scw *SimulatedChainWriter) Start(ctx context.Context) error {
	scw.ctx, scw.cancel = context.WithCancel(ctx)

	go func() {
		scw.processQueue()
	}()

	return nil
}

func (scw *SimulatedChainWriter) processQueue() {
	ticker := time.NewTicker(scw.config.Interval)
	defer ticker.Stop()

	sugar := scw.logger.Sugar()
	sugar.Infow("SimulatedChainWriter started")

	for {
		select {
		case <-scw.ctx.Done():
			sugar.Infow("SimulatedChainWriter shutting down")
			return
		case <-ticker.C:
			scw.processBatch()
		case task, ok := <-scw.workOutputQueue:
			if !ok {
				sugar.Infow("Work output queue has been closed")
				return
			}
			_ = scw.submitResult(task)
		}
	}
}

func (scw *SimulatedChainWriter) processBatch() {
	scw.logger.Sugar().Debugw("Processing batch of tasks")

	for i := 0; i < 10; i++ {
		select {
		case <-scw.ctx.Done():
			return
		case task, ok := <-scw.workOutputQueue:
			if !ok {
				return
			}
			_ = scw.submitResult(task)
		default:
			return
		}
	}
}

func (scw *SimulatedChainWriter) submitResult(result *types.TaskResult) error {
	sugar := scw.logger.Sugar()
	sugar.Infow("Simulating submitResult", "task_id", result.TaskId)

	time.Sleep(5 * time.Millisecond)

	payload, err := json.Marshal(result)
	if err != nil {
		sugar.Errorw("Failed to marshal result", "err", err)
		return err
	}

	sugar.Infow("Simulating submitResult",
		"task_id", result.TaskId,
		"avs_address", result.AvsAddress,
		"callback_address", result.CallbackAddr,
		"chain_id", result.ChainId,
		"block_number", result.BlockNumber,
		"payload", string(payload),
	)
	return nil
}
