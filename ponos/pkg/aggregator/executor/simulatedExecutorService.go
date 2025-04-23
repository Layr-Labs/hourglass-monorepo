package executor

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/common/v1"
	aggregatorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	executorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"log"
)

type SimulatedExecutorServer struct {
	rpcServer        *rpcServer.RpcServer
	aggregatorClient aggregatorpb.AggregatorServiceClient
	operatorAddress  string
}

func NewSimulatedExecutorServer(
	rpcServer *rpcServer.RpcServer,
	client aggregatorpb.AggregatorServiceClient,
	operatorAddress string,
) *SimulatedExecutorServer {
	return &SimulatedExecutorServer{
		rpcServer:        rpcServer,
		aggregatorClient: client,
		operatorAddress:  operatorAddress,
	}
}

func (s *SimulatedExecutorServer) Start(ctx context.Context) error {
	return s.rpcServer.Start(ctx)
}

func (s *SimulatedExecutorServer) Close() error {
	return nil
}

func (s *SimulatedExecutorServer) SubmitTask(ctx context.Context, req *executorpb.TaskSubmission) (*v1.SubmitAck, error) {
	log.Printf("Received task %s from aggregator %s", req.TaskId, req.AggregatorAddress)

	result := &aggregatorpb.TaskResult{
		TaskId:          req.TaskId,
		OperatorAddress: s.operatorAddress,
		Output:          []byte("simulatedOutput"),
		Signature:       []byte("simulatedSig"),
	}

	ack, err := s.aggregatorClient.SubmitTaskResult(ctx, result)
	if err != nil {
		log.Printf("Failed to send result: %v", err)
		return &v1.SubmitAck{Success: false, Message: "error sending result"}, nil
	}

	return ack, nil
}
