package aggregator

import (
	"context"
	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/auth"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *Aggregator) RegisterAvs(ctx context.Context, request *aggregatorV1.RegisterAvsRequest) (*aggregatorV1.RegisterAvsResponse, error) {
	a.logger.Sugar().Infow("RegisterAvs called",
		zap.String("avsAddress", request.AvsAddress),
		zap.Any("chainIds", request.ChainIds),
		zap.Bool("hasAuth", request.Auth != nil),
		zap.Bool("authEnabled", a.authVerifier != nil),
	)

	// Verify authentication
	if err := auth.HandleAuthError(a.verifyAuth(request.Auth)); err != nil {
		return nil, err
	}

	err := a.registerAvs(&aggregatorConfig.AggregatorAvs{
		Address: request.AvsAddress,
		ChainIds: util.Map(request.ChainIds, func(id uint32, i uint64) uint {
			return uint(id)
		}),
	})
	return &aggregatorV1.RegisterAvsResponse{
		Success: err == nil,
	}, nil
}

func (a *Aggregator) DeRegisterAvs(ctx context.Context, request *aggregatorV1.DeRegisterAvsRequest) (*aggregatorV1.DeRegisterAvsResponse, error) {
	// Verify authentication
	if err := auth.HandleAuthError(a.verifyAuth(request.Auth)); err != nil {
		return nil, err
	}

	return nil, status.Errorf(codes.Unimplemented, "DeRegisterAvs is not implemented yet")
}

// GetChallengeToken generates a challenge token for authentication
func (a *Aggregator) GetChallengeToken(ctx context.Context, request *aggregatorV1.AggregatorGetChallengeTokenRequest) (*aggregatorV1.AggregatorGetChallengeTokenResponse, error) {
	if a.authVerifier == nil {
		a.logger.Sugar().Warnw("GetChallengeToken called but authentication is not configured")
		return nil, status.Error(codes.Unimplemented, "authentication is not configured")
	}

	entry, err := a.authVerifier.GenerateChallengeToken(request.AggregatorAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate challenge token: %v", err)
	}

	return &aggregatorV1.AggregatorGetChallengeTokenResponse{
		ChallengeToken: entry.Token,
		ExpiresAt:      entry.ExpiresAt.Unix(),
	}, nil
}
