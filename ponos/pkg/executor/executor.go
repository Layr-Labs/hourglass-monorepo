package executor

import (
	"context"
	"fmt"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/artifactMonitor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/runtime"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/serverPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"strings"
	"sync"
)

type Executor struct {
	logger            *zap.Logger
	config            *executorConfig.ExecutorConfig
	avsPerformers     map[string]avsPerformer.IAvsPerformer
	runtimeController runtime.IContainerRuntimeController
	rpcServer         *rpcServer.RpcServer
	signer            signer.ISigner

	inflightTasks *sync.Map

	peeringFetcher  peering.IPeeringDataFetcher
	artifactMonitor *artifactMonitor.PerformerArtifactMonitor
	poller          chainPoller.IChainPoller
}

func NewExecutorWithRpcServer(
	port int,
	config *executorConfig.ExecutorConfig,
	logger *zap.Logger,
	signer signer.ISigner,
	peeringFetcher peering.IPeeringDataFetcher,
	runtimeController runtime.IContainerRuntimeController,
	artifactMonitor *artifactMonitor.PerformerArtifactMonitor,
) (*Executor, error) {
	rpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: port,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC server: %v", err)
	}

	return NewExecutor(config, rpc, logger, signer, peeringFetcher, runtimeController, artifactMonitor, nil), nil
}

func NewExecutor(
	config *executorConfig.ExecutorConfig,
	rpcServer *rpcServer.RpcServer,
	logger *zap.Logger,
	signer signer.ISigner,
	peeringFetcher peering.IPeeringDataFetcher,
	runtimeController runtime.IContainerRuntimeController,
	artifactMonitor *artifactMonitor.PerformerArtifactMonitor,
	poller chainPoller.IChainPoller,
) *Executor {
	return &Executor{
		logger:            logger,
		config:            config,
		avsPerformers:     make(map[string]avsPerformer.IAvsPerformer),
		rpcServer:         rpcServer,
		signer:            signer,
		inflightTasks:     &sync.Map{},
		peeringFetcher:    peeringFetcher,
		runtimeController: runtimeController,
		artifactMonitor:   artifactMonitor,
		poller:            poller,
	}
}

func (e *Executor) Initialize() error {
	e.logger.Sugar().Infow("Initializing AVS performers")

	for _, avs := range e.config.AvsPerformers {
		avsAddress := strings.ToLower(avs.AvsAddress)
		if _, ok := e.avsPerformers[avsAddress]; ok {
			e.logger.Sugar().Errorw("AVS performer already exists",
				zap.String("avsAddress", avsAddress),
				zap.String("processType", avs.ProcessType),
			)
		}

		switch avs.ProcessType {
		case string(avsPerformer.AvsProcessTypeServer):
			performer, err := serverPerformer.NewAvsPerformerServer(
				&avsPerformer.AvsPerformerConfig{
					AvsAddress:           avsAddress,
					ProcessType:          avsPerformer.AvsProcessType(avs.ProcessType),
					Image:                avsPerformer.PerformerImage{Repository: avs.Image.Repository, Tag: avs.Image.Tag},
					WorkerCount:          avs.WorkerCount,
					PerformerNetworkName: e.config.PerformerNetworkName,
					SigningCurve:         avs.SigningCurve,
				},
				e.peeringFetcher,
				e.artifactMonitor,
				e.runtimeController,
				e.logger,
			)
			if err != nil {
				e.logger.Sugar().Errorw("Failed to create AVS performer server",
					zap.String("avsAddress", avsAddress),
					zap.Error(err),
				)
				return fmt.Errorf("failed to create AVS performer server: %v", err)
			}
			e.avsPerformers[avsAddress] = performer

		default:
			e.logger.Sugar().Errorw("Unsupported AVS performer process type",
				zap.String("avsAddress", avsAddress),
				zap.String("processType", avs.ProcessType),
			)
			return fmt.Errorf("unsupported AVS performer process type: %s", avs.ProcessType)
		}
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
	e.logger.Sugar().Infow("Booting AVS performers")
	for avsAddress, performer := range e.avsPerformers {
		if err := performer.Initialize(ctx); err != nil {
			e.logger.Sugar().Errorw("Failed to initialize AVS performer",
				zap.String("avsAddress", avsAddress),
				zap.Error(err),
			)
			return fmt.Errorf("failed to initialize AVS performer: %v", err)
		}
		performer.Start(ctx)
	}

	go func() {
		<-ctx.Done()
		e.logger.Sugar().Info("Shutting down AVS performers")
		for avsAddress, performer := range e.avsPerformers {
			if err := performer.Shutdown(); err != nil {
				e.logger.Sugar().Errorw("Failed to shutdown AVS performer",
					zap.String("avsAddress", avsAddress),
					zap.Error(err),
				)
			}
		}
	}()
	return nil
}

func (e *Executor) Start(ctx context.Context) error {
	e.logger.Info("Executor is running",
		zap.String("version", "1.0.0"),
		zap.String("operatorAddress", e.config.Operator.Address),
	)

	e.artifactMonitor.Start(ctx)
	err := e.poller.Start(ctx)
	if err != nil {
		return err
	}

	if err = e.rpcServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start RPC server: %v", err)
	}
	return nil
}

func (e *Executor) registerHandlers(grpcServer *grpc.Server) error {
	executorV1.RegisterExecutorServiceServer(grpcServer, e)

	return nil
}
