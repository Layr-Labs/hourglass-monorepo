package simulatedAggregator

import (
	"context"
	"fmt"
	v1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/common/v1"
	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"go.uber.org/zap"
)

type TaskResultInspector func(result *aggregatorV1.TaskResult)

type SimulatedAggregator struct {
	rpcServer           *rpcServer.RpcServer
	logger              *zap.Logger
	config              *aggregatorConfig.AggregatorConfig
	taskResultInspector TaskResultInspector
}

func NewSimulatedAggregator(
	config *aggregatorConfig.AggregatorConfig,
	logger *zap.Logger,
	rpcServer *rpcServer.RpcServer,
	inspector TaskResultInspector,
) (*SimulatedAggregator, error) {
	sa := &SimulatedAggregator{
		rpcServer:           rpcServer,
		logger:              logger,
		config:              config,
		taskResultInspector: inspector,
	}
	sa.initializeHandlers()
	return sa, nil
}

func (sa *SimulatedAggregator) initializeHandlers() {
	aggregatorV1.RegisterAggregatorServiceServer(sa.rpcServer.GetGrpcServer(), sa)
}

func (sa *SimulatedAggregator) SubmitTaskResult(ctx context.Context, result *aggregatorV1.TaskResult) (*v1.SubmitAck, error) {
	sa.logger.Sugar().Infow("Received task result from executor",
		zap.String("taskId", result.TaskId),
		zap.String("operatorAddress", result.OperatorAddress),
		zap.Any("output", result.Output),
		zap.Binary("signature", result.Signature),
	)

	if sa.taskResultInspector != nil {
		sa.taskResultInspector(result)
	}

	return &v1.SubmitAck{
		Message: "task result received",
		Success: true,
	}, nil
}

func (sa *SimulatedAggregator) Run(ctx context.Context) error {
	if err := sa.rpcServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start RPC server: %v", err)
	}
	return nil
}
