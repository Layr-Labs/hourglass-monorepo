package executor

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/server"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/connectedAggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ConnectedAggregatorStore struct {
	connectedAggregators map[string]*connectedAggregator.ConnectedAggregator
	logger               *zap.Logger
}

func NewConnectedAggregatorStore(logger *zap.Logger) *ConnectedAggregatorStore {
	return &ConnectedAggregatorStore{
		connectedAggregators: make(map[string]*connectedAggregator.ConnectedAggregator),
		logger:               logger,
	}
}

type Executor struct {
	logger        *zap.Logger
	config        *executorConfig.ExecutorConfig
	avsPerformers map[string]avsPerformer.IAvsPerformer
	rpcServer     *rpcServer.RpcServer
	aggregators   map[string]*connectedAggregator.ConnectedAggregator
	signer        signer.Signer
}

func NewExecutor(
	config *executorConfig.ExecutorConfig,
	rpcServer *rpcServer.RpcServer,
	logger *zap.Logger,
	signer signer.Signer,
) *Executor {
	return &Executor{
		logger:        logger,
		config:        config,
		avsPerformers: make(map[string]avsPerformer.IAvsPerformer),
		rpcServer:     rpcServer,
		signer:        signer,
	}
}

func (e *Executor) Initialize() error {
	e.logger.Sugar().Infow("Initializing AVS performers")

	for _, avs := range e.config.AvsPerformers {
		if _, ok := e.avsPerformers[avs.AvsAddress]; ok {
			e.logger.Sugar().Errorw("AVS performer already exists",
				zap.String("avsAddress", avs.AvsAddress),
				zap.String("processType", avs.ProcessType),
			)
		}

		switch avs.ProcessType {
		case string(avsPerformer.AvsProcessTypeServer):
			performer, err := server.NewAvsPerformerServer(&avsPerformer.AvsPerformerConfig{
				AvsAddress:  avs.AvsAddress,
				ProcessType: avsPerformer.AvsProcessType(avs.ProcessType),
				Image:       avsPerformer.PerformerImage{Repository: avs.Image.Repository, Tag: avs.Image.Tag},
			}, e.logger)
			if err != nil {
				e.logger.Sugar().Errorw("Failed to create AVS performer server",
					zap.String("avsAddress", avs.AvsAddress),
					zap.Error(err),
				)
				return fmt.Errorf("failed to create AVS performer server: %v", err)
			}
			e.avsPerformers[avs.AvsAddress] = performer

		default:
			e.logger.Sugar().Errorw("Unsupported AVS performer process type",
				zap.String("avsAddress", avs.AvsAddress),
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

func (e *Executor) Run() {
	e.logger.Info("Worker node is running", zap.String("version", "1.0.0"))
}

func (e *Executor) registerHandlers(grpcServer *grpc.Server) error {
	executor.RegisterExecutorServiceServer(grpcServer, e)

	return nil
}
