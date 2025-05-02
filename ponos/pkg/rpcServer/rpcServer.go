package rpcServer

import (
	"context"
	"errors"
	"fmt"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type RpcServerConfig struct {
	GrpcPort int
}

type RpcServer struct {
	logger       *zap.Logger
	RpcConfig    *RpcServerConfig
	grpcListener *net.Listener
	grpcServer   *grpc.Server
}

func NewRpcServer(
	rpcConfig *RpcServerConfig,
	logger *zap.Logger,
) (*RpcServer, error) {
	grpc_zap.ReplaceGrpcLoggerV2(logger.WithOptions(zap.IncreaseLevel(zap.WarnLevel)))

	opts := []grpc_zap.Option{
		grpc_zap.WithDecider(func(fullMethodName string, err error) bool {
			return true
		}),
	}

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcConfig.GrpcPort))
	if err != nil {
		logger.Sugar().Errorw("Failed to listen",
			zap.Int("port", rpcConfig.GrpcPort),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.UnaryServerInterceptor(logger, opts...),
		),
	)
	reflection.Register(grpcServer)

	return &RpcServer{
		logger:       logger,
		RpcConfig:    rpcConfig,
		grpcListener: &grpcListener,
		grpcServer:   grpcServer,
	}, nil
}

func (rpc *RpcServer) Start(ctx context.Context) error {
	rpc.logger.Sugar().Infow("Starting gRPC server",
		zap.Int("port", rpc.RpcConfig.GrpcPort),
	)

	go func() {
		if err := rpc.grpcServer.Serve(*rpc.grpcListener); err != nil {
			rpc.logger.Sugar().Fatal("failed to serve reload server", zap.Error(err))
		}
	}()
	go func() {
		<-ctx.Done()
		err := ctx.Err()
		rpc.logger.Sugar().Info("received context.Done()")
		switch {
		case errors.Is(err, context.Canceled):
			rpc.logger.Sugar().Info("context canceled, shutting down")
		case errors.Is(err, context.DeadlineExceeded):
			rpc.logger.Sugar().Info("context deadline exceeded, shutting down")
		default:
			rpc.logger.Sugar().Info("Unknown error, shutting down")
		}

		rpc.grpcServer.GracefulStop()
	}()
	return nil
}

func (rpc *RpcServer) GetGrpcServer() *grpc.Server {
	return rpc.grpcServer
}
