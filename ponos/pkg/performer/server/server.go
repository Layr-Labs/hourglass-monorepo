package server

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performer/worker"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"go.uber.org/zap"
	"time"
)

type PonosPerformerConfig struct {
	Port    int
	Timeout time.Duration
}

type PonosPerformer struct {
	config     *PonosPerformerConfig
	rpcServer  *rpcServer.RpcServer
	taskWorker worker.IWorker
	logger     *zap.Logger
}

func NewPonosPerformer(
	cfg *PonosPerformerConfig,
	rpcServer *rpcServer.RpcServer,
	worker worker.IWorker,
	logger *zap.Logger,
) *PonosPerformer {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	pp := &PonosPerformer{
		config:     cfg,
		rpcServer:  rpcServer,
		taskWorker: worker,
		logger:     logger,
	}
	pp.registerHandlers()

	return pp
}

func NewPonosPerformerWithRpcServer(
	cfg *PonosPerformerConfig,
	worker worker.IWorker,
	logger *zap.Logger,
) (*PonosPerformer, error) {
	rpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: cfg.Port,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC server: %w", err)
	}
	return NewPonosPerformer(cfg, rpc, worker, logger), nil
}

func (pp *PonosPerformer) registerHandlers() {
	performerV1.RegisterPerformerServiceServer(pp.rpcServer.GetGrpcServer(), pp)
}

func (pp *PonosPerformer) Start(ctx context.Context) error {
	go func() {
		if err := pp.rpcServer.Start(ctx); err != nil {
			pp.logger.Sugar().Errorw("Failed to start RPC server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	pp.logger.Sugar().Infow("Shutting down grpc server")
	return nil
}
