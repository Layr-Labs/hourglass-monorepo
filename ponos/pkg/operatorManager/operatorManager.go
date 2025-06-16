package operatorManager

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"math/big"
	"strings"
)

type OperatorManagerConfig struct {
	AvsAddress     string
	OperatorSetIds []uint32 // TODO(seanmcgary): this should get hydrated from the AVSConfig object
	ChainIds       []config.ChainId
}

type PeerWeight struct {
	BlockNumber            uint64
	ChainId                config.ChainId
	OperatorSetId          uint32
	RootReferenceTimestamp uint32
	Weights                map[string][]*big.Int
	Operators              []*peering.OperatorPeerInfo
}

type OperatorManager struct {
	config *OperatorManagerConfig

	contractCallers map[config.ChainId]contractCaller.IContractCaller

	peeringDataFetcher peering.IPeeringDataFetcher

	logger *zap.Logger
}

func NewOperatorManager(
	cfg *OperatorManagerConfig,
	ccs map[config.ChainId]contractCaller.IContractCaller,
	pdf peering.IPeeringDataFetcher,
	logger *zap.Logger,
) *OperatorManager {
	return &OperatorManager{
		config:             cfg,
		contractCallers:    ccs,
		peeringDataFetcher: pdf,
		logger:             logger,
	}
}

func (om *OperatorManager) GetOperatorPeersAndWeightsForBlock(
	ctx context.Context,
	chainId config.ChainId,
	blockNumber uint64,
	operatorSetId uint32,
) (*PeerWeight, error) {
	cc, err := om.getContractCallerForChainId(chainId)
	if err != nil {
		om.logger.Sugar().Errorw("Failed to get contract caller for chain ID",
			zap.Uint32("ChainId", uint32(chainId)),
			zap.Error(err),
		)
		return nil, err
	}

	// no Weights found, go get the latest Weights
	om.logger.Sugar().Debugw("No Weights found for chain",
		zap.Uint32("ChainId", uint32(chainId)),
		zap.String("AvsAddress", om.config.AvsAddress),
		zap.Uint64("BlockNumber", blockNumber),
		zap.Uint32("OperatorSetId", operatorSetId),
	)

	tableData, err := cc.GetOperatorTableDataForOperatorSet(ctx, common.HexToAddress(om.config.AvsAddress), operatorSetId, chainId, blockNumber)
	if err != nil {
		om.logger.Sugar().Errorw("Failed to get operator table data",
			zap.String("AvsAddress", om.config.AvsAddress),
			zap.Uint32("OperatorSetId", operatorSetId),
			zap.Uint64("BlockNumber", blockNumber),
			zap.Error(err),
		)
		return nil, err
	}

	operatorWeights := make(map[string][]*big.Int, len(tableData.Operators))
	for i, operator := range tableData.Operators {
		weight := tableData.OperatorWeights[i]
		operatorWeights[operator.String()] = weight
	}
	operators, err := om.peeringDataFetcher.ListExecutorOperators(ctx, om.config.AvsAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to list executor Operators: %w", err)
	}

	// filter the list of Operators down to those that are in the operator set and have Weights
	operators = util.Filter(operators, func(op *peering.OperatorPeerInfo) bool {
		for opAddr, _ := range operatorWeights {
			if strings.EqualFold(opAddr, op.OperatorAddress) && op.IncludesOperatorSetId(operatorSetId) {
				return true
			}
		}
		return false
	})

	return &PeerWeight{
		BlockNumber:            blockNumber,
		ChainId:                chainId,
		OperatorSetId:          operatorSetId,
		RootReferenceTimestamp: tableData.LatestReferenceTimestamp,
		Weights:                operatorWeights,
		Operators:              operators,
	}, nil

}

func (om *OperatorManager) getContractCallerForChainId(chainId config.ChainId) (contractCaller.IContractCaller, error) {
	cc, ok := om.contractCallers[chainId]
	if !ok {
		return nil, fmt.Errorf("no contract caller found for chain ID %d", chainId)
	}
	return cc, nil
}
