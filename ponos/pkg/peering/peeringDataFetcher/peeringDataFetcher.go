package peeringDataFetcher

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"go.uber.org/zap"
)

type PeeringDataFetcher struct {
	contractCaller contractCaller.IContractCaller
	logger         *zap.Logger
}

func NewPeeringDataFetcher(
	contractCaller contractCaller.IContractCaller,
	logger *zap.Logger,
) *PeeringDataFetcher {
	return &PeeringDataFetcher{
		contractCaller: contractCaller,
		logger:         logger,
	}
}

func (pdf *PeeringDataFetcher) ListExecutorOperators(ctx context.Context, avsAddress string) ([]*peering.OperatorPeerInfo, error) {
	return nil, nil
}

func (pdf *PeeringDataFetcher) ListAggregatorOperators(ctx context.Context, avsAddress string) ([]*peering.OperatorPeerInfo, error) {
	avsConfig, err := pdf.contractCaller.GetAVSConfig(ctx, avsAddress)
	if err != nil {
		pdf.logger.Sugar().Errorf("Failed to get AVS config", zap.Error(err))
		return nil, err
	}

	if avsConfig == nil {
		pdf.logger.Sugar().Errorf("AVS config is nil")
		return nil, nil
	}

	return pdf.contractCaller.GetOperatorSetMembersWithPeering(avsAddress, avsConfig.AggregatorOperatorSetId)
}
