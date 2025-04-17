package coordinator

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/executionManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/workQueue"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
	"time"
)

type PonosCoordinator struct {
	taskQueue        workQueue.IOutputQueue[types.Task]
	resultQueue      workQueue.IInputQueue[types.TaskResult]
	executionManager executionManager.IExecutionManager
	logger           *zap.Logger
	cancelFunc       context.CancelFunc
}

func NewPonosCoordinator(
	taskQueue workQueue.IOutputQueue[types.Task],
	resultQueue workQueue.IInputQueue[types.TaskResult],
	executionManager executionManager.IExecutionManager,
	logger *zap.Logger,
) *PonosCoordinator {
	return &PonosCoordinator{
		taskQueue:        taskQueue,
		resultQueue:      resultQueue,
		executionManager: executionManager,
		logger:           logger,
	}
}

func (pc *PonosCoordinator) Start(ctx context.Context) error {
	pc.logger.Info("PonosCoordinator started")

	ctx, pc.cancelFunc = context.WithCancel(ctx)

	go pc.runTaskProcessor(ctx)
	go pc.runResultProcessor(ctx)

	return nil
}

func (pc *PonosCoordinator) Close() error {
	pc.logger.Info("PonosCoordinator shutting down...")
	if pc.cancelFunc != nil {
		pc.cancelFunc()
	}
	return nil
}

func (pc *PonosCoordinator) runTaskProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			pc.logger.Info("Task loop shutting down")
			return
		default:
			task := pc.taskQueue.Dequeue()
			if err := pc.executionManager.ExecuteTask(task); err != nil {
				pc.logger.Error("Failed to execute task", zap.String("task_id", task.TaskId), zap.Error(err))
			} else {
				pc.logger.Info("Task dispatched to execution manager", zap.String("task_id", task.TaskId))
			}
		}
	}
}

func (pc *PonosCoordinator) runResultProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			pc.logger.Info("Result collector shutting down")
			return
		default:
			results := pc.executionManager.LoadResults()
			for _, result := range results {
				if err := pc.resultQueue.Enqueue(result); err != nil {
					pc.logger.Error("Failed to enqueue result", zap.String("task_id", result.TaskId), zap.Error(err))
				} else {
					pc.logger.Info("Result submitted", zap.String("task_id", result.TaskId))
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}
