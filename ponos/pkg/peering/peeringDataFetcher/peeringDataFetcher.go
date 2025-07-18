package peeringDataFetcher

import (
	"context"
	"fmt"
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
	avsConfig, err := pdf.contractCaller.GetAVSConfig(avsAddress)
	if err != nil {
		pdf.logger.Sugar().Errorw("Failed to get AVS config",
			zap.String("avsAddress", avsAddress),
			zap.Error(err),
		)
		return nil, err
	}
	// map[operatorAddress]*peering.OperatorPeerInfo
	operatorPeeringInfos := map[string]*peering.OperatorPeerInfo{}

	// iterate over all operator sets and get their members with peering info
	for _, operatorSetId := range avsConfig.ExecutorOperatorSetIds {
		peeringInfos, err := pdf.contractCaller.GetOperatorSetMembersWithPeering(avsAddress, operatorSetId)
		if err != nil {
			return nil, fmt.Errorf("failed to get operator set members with peering %w", err)
		}
		for _, peeringInfo := range peeringInfos {
			infos, ok := operatorPeeringInfos[peeringInfo.OperatorAddress]
			if !ok {
				operatorPeeringInfos[peeringInfo.OperatorAddress] = peeringInfo
				continue
			}
			infos.OperatorSets = append(infos.OperatorSets, peeringInfo.OperatorSets...)
		}
	}
	result := make([]*peering.OperatorPeerInfo, 0, len(operatorPeeringInfos))
	for _, info := range operatorPeeringInfos {
		result = append(result, info)
	}

	return result, nil
}

func (pdf *PeeringDataFetcher) ListAggregatorOperators(ctx context.Context, avsAddress string) ([]*peering.OperatorPeerInfo, error) {
	avsConfig, err := pdf.contractCaller.GetAVSConfig(avsAddress)
	if err != nil {
		pdf.logger.Sugar().Errorw("Failed to get AVS config", zap.Error(err))
		return nil, err
	}

	if avsConfig == nil {
		pdf.logger.Sugar().Errorw("AVS config is nil")
		return nil, nil
	}

	return pdf.contractCaller.GetOperatorSetMembersWithPeering(avsAddress, avsConfig.AggregatorOperatorSetId)
}
