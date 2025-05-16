package service

import (
	"context"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"go.uber.org/zap"
	"log"
)

type SimulatedExecutorServer struct {
	rpcServer       *rpcServer.RpcServer
	operatorAddress string
}

func NewSimulatedExecutorWithRpcServer(
	port int,
	logger *zap.Logger,
	operatorAddress string,
) (*SimulatedExecutorServer, error) {
	server, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: port,
	}, logger)
	if err != nil {
		return nil, err
	}

	return NewSimulatedExecutorServer(server, operatorAddress), nil
}

func NewSimulatedExecutorServer(
	rpcServer *rpcServer.RpcServer,
	operatorAddress string,
) *SimulatedExecutorServer {
	es := &SimulatedExecutorServer{
		rpcServer:       rpcServer,
		operatorAddress: operatorAddress,
	}

	executorV1.RegisterExecutorServiceServer(rpcServer.GetGrpcServer(), es)
	return es
}

func (s *SimulatedExecutorServer) Start(ctx context.Context) error {
	return s.rpcServer.Start(ctx)
}

func (s *SimulatedExecutorServer) Close() error {
	return nil
}

func (s *SimulatedExecutorServer) SubmitTask(ctx context.Context, req *executorV1.TaskSubmission) (*executorV1.TaskResult, error) {
	log.Printf("Received task %s from aggregator %s", req.TaskId, req.AggregatorAddress)

	return &executorV1.TaskResult{
		TaskId:          req.TaskId,
		OperatorAddress: s.operatorAddress,
		AvsAddress:      req.AvsAddress,
		Output:          []byte("simulatedOutput"),
		Signature:       []byte("simulatedSig"),
	}, nil
}
