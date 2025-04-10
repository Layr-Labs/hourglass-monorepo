package executorRpcServer

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ExecutorRpcServer struct {
	executor.UnimplementedExecutorServiceServer
	rpcServer *rpcServer.RpcServer
	logger    *zap.Logger
}

func NewExecutorRpcServer(
	rpcServer *rpcServer.RpcServer,
	logger *zap.Logger,
) (*ExecutorRpcServer, error) {

	exec := &ExecutorRpcServer{
		logger:    logger,
		rpcServer: rpcServer,
	}

	if err := exec.registerHandlers(rpcServer.GetGrpcServer()); err != nil {
		logger.Sugar().Errorw("Failed to register handlers",
			zap.Error(err),
		)
		return nil, err
	}
	return exec, nil
}

func (rpc *ExecutorRpcServer) registerHandlers(grpcServer *grpc.Server) error {
	executor.RegisterExecutorServiceServer(grpcServer, rpc)

	return nil
}
