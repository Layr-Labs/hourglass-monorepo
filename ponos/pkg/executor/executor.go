package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/serverPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/deployment"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Executor struct {
	logger        *zap.Logger
	config        *executorConfig.ExecutorConfig
	avsPerformers map[string]avsPerformer.IAvsPerformer
	rpcServer     *rpcServer.RpcServer
	signer        signer.ISigner
	containerMgr  *containerManager.DockerContainerManager
	inflightTasks *sync.Map
	deploymentMgr *deployment.Manager

	peeringFetcher peering.IPeeringDataFetcher
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
		avsPerformers:  make(map[string]avsPerformer.IAvsPerformer),
		rpcServer:      rpcServer,
		signer:         signer,
		inflightTasks:  &sync.Map{},
		deploymentMgr:  deployment.NewManager(logger),
		peeringFetcher: peeringFetcher,
	}
}

func (e *Executor) Initialize(ctx context.Context) error {
	e.logger.Sugar().Infow("Initializing AVS performers")
	containerMgr, err := containerManager.NewDockerContainerManager(
		containerManager.DefaultContainerManagerConfig(),
		e.logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create container manager for executor: %v", err)
	}
	e.containerMgr = containerMgr

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
			performer := serverPerformer.NewAvsPerformerServer(
				&avsPerformer.AvsPerformerConfig{
					AvsAddress:           avsAddress,
					ProcessType:          avsPerformer.AvsProcessType(avs.ProcessType),
					WorkerCount:          avs.WorkerCount,
					PerformerNetworkName: e.config.PerformerNetworkName,
					SigningCurve:         avs.SigningCurve,
				},
				e.peeringFetcher,
				e.logger,
				containerMgr,
			)
			err = performer.Initialize(ctx)
			if err != nil {
				return err
			}

			// Deploy container using deployment manager
			deploymentConfig := deployment.DeploymentConfig{
				AvsAddress: avsAddress,
				Image: avsPerformer.PerformerImage{
					Repository: avs.Image.Repository,
					Tag:        avs.Image.Tag,
				},
				Timeout: 1 * time.Minute,
			}

			result, err := e.deploymentMgr.Deploy(ctx, deploymentConfig, performer)
			if err != nil {
				e.logger.Sugar().Errorw("Failed to deploy container for AVS performer during startup",
					zap.String("avsAddress", avsAddress),
					zap.Error(err),
				)
				return fmt.Errorf("failed to deploy container for AVS %s: %w", avsAddress, err)
			}

			e.logger.Sugar().Infow("AVS performer container deployed successfully",
				zap.String("avsAddress", avsAddress),
				zap.String("deploymentId", result.DeploymentID),
				zap.String("performerId", result.PerformerID),
			)

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
