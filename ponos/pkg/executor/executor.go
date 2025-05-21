package executor

import (
	"context"
	"fmt"
	"sync"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/performerPoolManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Executor struct {
	logger        *zap.Logger
	config        *executorConfig.ExecutorConfig
	rpcServer     *rpcServer.RpcServer
	signer        signer.ISigner
	inflightTasks *sync.Map

	performerPoolManager *performerPoolManager.PerformerPoolManager
	peeringFetcher       peering.IPeeringDataFetcher
}

func NewExecutorWithRpcServer(
	port int,
	config *executorConfig.ExecutorConfig,
	logger *zap.Logger,
	signer signer.ISigner,
	peeringFetcher peering.IPeeringDataFetcher,
) (*Executor, error) {
	rpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: port,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC server: %v", err)
	}

	return NewExecutor(config, rpc, logger, signer, peeringFetcher), nil
}

func NewExecutor(
	config *executorConfig.ExecutorConfig,
	rpcServer *rpcServer.RpcServer,
	logger *zap.Logger,
	signer signer.ISigner,
	peeringFetcher peering.IPeeringDataFetcher,
) *Executor {
	return &Executor{
		logger:         logger,
		config:         config,
		rpcServer:      rpcServer,
		signer:         signer,
		inflightTasks:  &sync.Map{},
		peeringFetcher: peeringFetcher,
	}
}

func (e *Executor) Initialize() error {
	// Create the performer pool manager
	e.performerPoolManager = performerPoolManager.NewPerformerPoolManager(
		e.config,
		e.logger,
		e.peeringFetcher,
	)

	// Initialize the performer manager
	if err := e.performerPoolManager.Initialize(); err != nil {
		e.logger.Sugar().Errorw("Failed to initialize performer pool manager",
			zap.Error(err),
		)
		return fmt.Errorf("failed to initialize performer pool manager: %v", err)
	}

	if err := e.registerHandlers(e.rpcServer.GetGrpcServer()); err != nil {
		e.logger.Sugar().Errorw("Failed to register handlers",
			zap.Error(err),
		)
		return fmt.Errorf("failed to register handlers: %v", err)
	}

	return nil
}

func (e *Executor) BootPerformers(ctx context.Context) error {
	return e.performerPoolManager.BootPerformers(ctx)
}

func (e *Executor) Run(ctx context.Context) error {
	e.logger.Info("Executor is running",
		zap.String("version", "1.0.0"),
		zap.String("operatorAddress", e.config.Operator.Address),
	)
	if err := e.rpcServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start RPC server: %v", err)
	}
	return nil
}

func (e *Executor) registerHandlers(grpcServer *grpc.Server) error {
	executorV1.RegisterExecutorServiceServer(grpcServer, e)

	return nil
}
