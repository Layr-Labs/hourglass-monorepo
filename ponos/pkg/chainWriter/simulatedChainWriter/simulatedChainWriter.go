package simulatedChainWriter

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/workQueue"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
)

type SimulatedChainWriterConfig struct {
	Interval time.Duration
}

type SimulatedChainWriter struct {
	config          *SimulatedChainWriterConfig
	logger          *zap.SugaredLogger
	workOutputQueue workQueue.IOutputQueue[types.TaskResult]
	client          *http.Client
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

func NewSimulatedChainWriter(
	config *SimulatedChainWriterConfig,
	logger *zap.Logger,
	workOutputQueue workQueue.IOutputQueue[types.TaskResult],
) *SimulatedChainWriter {
	return &SimulatedChainWriter{
		config:          config,
		logger:          logger.Sugar(),
		workOutputQueue: workOutputQueue,
		client:          &http.Client{Timeout: 5 * time.Second},
	}
}

func (scw *SimulatedChainWriter) Start(parent context.Context) error {
	scw.ctx, scw.cancel = context.WithCancel(parent)
	scw.wg.Add(1)

	go func() {
		defer scw.wg.Done()
		ticker := time.NewTicker(scw.config.Interval)
		defer ticker.Stop()

		scw.logger.Infow("SimulatedChainWriter started")

		for {
			select {
			case <-scw.ctx.Done():
				scw.logger.Infow("SimulatedChainWriter shutting down")
				return
			case <-ticker.C:
				scw.drainQueue()
			}
		}
	}()

	return nil
}

func (scw *SimulatedChainWriter) drainQueue() {
	for {
		select {
		case <-scw.ctx.Done():
			return
		default:
			task := scw.workOutputQueue.Dequeue()
			_ = scw.submitResult(task)
		}
	}
}

func (scw *SimulatedChainWriter) submitResult(result *types.TaskResult) error {
	scw.logger.Infow("Simulating submitResult", "task_id", result.TaskId)

	time.Sleep(50 * time.Millisecond)

	payload, err := json.Marshal(result)
	if err != nil {
		scw.logger.Errorw("Failed to marshal result", "err", err)
		return err
	}
	scw.logger.Infow("Simulating submitResult",
		"task_id", result.TaskId,
		"avs_address", result.AvsAddress,
		"callback_address", result.CallbackAddr,
		"chain_id", result.ChainId,
		"block_number", result.BlockNumber,
		"payload", payload,
	)
	return nil
}

func (scw *SimulatedChainWriter) Close() error {
	if scw.cancel != nil {
		scw.cancel()
	}
	scw.wg.Wait()
	scw.logger.Infow("SimulatedChainWriter closed")
	return nil
}
