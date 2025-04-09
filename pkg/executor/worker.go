package executor

import (
	"github.com/Layr-Labs/go-ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/go-ponos/pkg/executorRpcServer"
	"go.uber.org/zap"
)

type WorkerNode struct {
	logger *zap.Logger
	config *executorConfig.ExecutorConfig
}

func NewWorkerNode(
	logger *zap.Logger,
	config *executorConfig.ExecutorConfig,
	rpcServer *executorRpcServer.ExecutorRpcServer,
) *WorkerNode {
	return &WorkerNode{
		logger: logger,
		config: config,
	}
}

func (w *WorkerNode) Run() {
	w.logger.Info("Worker node is running", zap.String("version", "1.0.0"))
}
