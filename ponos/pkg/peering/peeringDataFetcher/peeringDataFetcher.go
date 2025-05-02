package peeringDataFetcher

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
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
	avsConfig, err := pdf.contractCaller.GetAVSConfig(ctx, avsAddress)
	if err != nil {
		pdf.logger.Sugar().Errorf("Failed to get AVS config", zap.Error(err))
		return nil, err
	}
	operatorPeeringInfos := pdf.groupOperatorPeeringAcrossOpSets(avsAddress, avsConfig)
	return reduceOperatorPeers(operatorPeeringInfos), nil
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

func (pdf *PeeringDataFetcher) groupOperatorPeeringAcrossOpSets(
	avsAddress string,
	avsConfig *contractCaller.AVSConfig,
) map[string][]*peering.OperatorPeerInfo {
	operatorPeeringInfos := map[string][]*peering.OperatorPeerInfo{}
	for _, operatorSetId := range avsConfig.ExecutorOperatorSetIds {
		peeringInfos, err := pdf.contractCaller.GetOperatorSetMembersWithPeering(avsAddress, operatorSetId)
		if err != nil {
			return nil
		}
		for _, peeringInfo := range peeringInfos {
			infos, ok := operatorPeeringInfos[peeringInfo.OperatorAddress]
			if !ok {
				infos = []*peering.OperatorPeerInfo{}
				infos = append(infos, peeringInfo)
				operatorPeeringInfos[peeringInfo.OperatorAddress] = infos
			} else {
				operatorPeeringInfos[peeringInfo.OperatorAddress] = append(infos, peeringInfo)
			}
		}

	}
	return operatorPeeringInfos
}

func reduceOperatorPeers(operatorPeeringInfos map[string][]*peering.OperatorPeerInfo) []*peering.OperatorPeerInfo {
	result := make([]*peering.OperatorPeerInfo, 0, len(operatorPeeringInfos))
	for _, peeringInfos := range operatorPeeringInfos {
		operatorSetIds := util.Reduce(peeringInfos, func(operatorSetIds []uint32, next *peering.OperatorPeerInfo) []uint32 {
			return append(operatorSetIds, next.OperatorSetIds...)
		}, []uint32{})

		merged := *peeringInfos[0]
		merged.OperatorSetIds = operatorSetIds
		result = append(result, &merged)
	}
	return result
}
