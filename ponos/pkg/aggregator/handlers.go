package aggregator

import (
	"context"
	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *Aggregator) RegisterAvs(ctx context.Context, request *aggregatorV1.RegisterAvsRequest) (*aggregatorV1.RegisterAvsResponse, error) {
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
	return nil, status.Errorf(codes.Unimplemented, "DeRegisterAvs is not implemented yet")
}
