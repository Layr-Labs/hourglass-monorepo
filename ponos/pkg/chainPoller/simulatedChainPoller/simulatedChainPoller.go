package simulatedChainPoller

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/lifecycle/runnable"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/manualPushChainPoller"
	"math/big"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
)

type SimulatedChainPollerConfig struct {
	ChainId      *config.ChainId
	TaskInterval time.Duration
	Port         int
}

type SimulatedChainPoller struct {
	taskQueue chan *types.Task
	config    *SimulatedChainPollerConfig
	logger    *zap.Logger

	manualPoller runnable.IRunnable
}

func NewSimulatedChainPoller(
	taskQueue chan *types.Task,
	config *SimulatedChainPollerConfig,
	logger *zap.Logger,
) *SimulatedChainPoller {
	manualPoller := manualPushChainPoller.NewManualPushChainPoller(taskQueue, &manualPushChainPoller.ManualPushChainPollerConfig{
		ChainId: config.ChainId,
		Port:    config.Port,
	}, logger)

	return &SimulatedChainPoller{
		taskQueue:    taskQueue,
		config:       config,
		logger:       logger,
		manualPoller: manualPoller,
	}
}

func (scl *SimulatedChainPoller) Start(ctx context.Context) error {
	sugar := scl.logger.Sugar()
	sugar.Infow("SimulatedChainPoller starting", "port", scl.config.Port)

	go func() {
		if err := scl.manualPoller.Start(ctx); err != nil {
			sugar.Errorw("Manual chain poller error", "error", err)
		}
	}()

	if scl.config.TaskInterval > 0 {
		go func() {
			scl.generatePeriodicTasks(ctx)
		}()
	} else {
		return fmt.Errorf("polling interval must be greater than 0")
	}

	return nil
}

func (scl *SimulatedChainPoller) generatePeriodicTasks(ctx context.Context) {
	ticker := time.NewTicker(scl.config.TaskInterval)
	defer ticker.Stop()

	sugar := scl.logger.Sugar()
	sugar.Infow("Starting periodic task generation")

	for {
		select {
		case <-ctx.Done():
			sugar.Infow("Stopping periodic task generation")
			return
		case <-ticker.C:
			task := &types.Task{
				TaskId:        fmt.Sprintf("periodic-task-%d", time.Now().UnixNano()),
				AVSAddress:    "0xPeriodicTaskAVS",
				OperatorSetId: 123456,
				CallbackAddr:  "0xPeriodicTaskCallback",
				Payload:       []byte(`{"type":"periodic","timestamp":` + fmt.Sprintf("%d", time.Now().Unix()) + `}`),
				Deadline:      big.NewInt(time.Now().Add(1 * time.Hour).UnixMilli()),
				StakeRequired: 0.75,
				ChainId:       *scl.config.ChainId,
			}

			select {
			case scl.taskQueue <- task:
				sugar.Infow("Generated periodic task", "taskID", task.TaskId)
			case <-time.After(1 * time.Second):
				sugar.Warnw("Failed to enqueue periodic task (channel full or closed)", "taskID", task.TaskId)
			case <-ctx.Done():
				return
			}
		}
	}
}
