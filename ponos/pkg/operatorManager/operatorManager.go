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
	L1ChainId      config.ChainId
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

func (om *OperatorManager) GetOperatorSetData(
	ctx context.Context,
	chainId config.ChainId,
	taskBlockNumber uint64,
	operatorSetId uint32,
) (
	*contractCaller.OperatorTableData,
	map[string][]*big.Int,
	uint32,
	error,
) {
	l1Cc, err := om.getContractCallerForChainId(om.config.L1ChainId)
	if err != nil {
		om.logger.Sugar().Errorw("Failed to get contract caller for chain ID",
			zap.Uint32("ChainId", uint32(chainId)),
			zap.Error(err),
		)
		return nil, nil, 0, err
	}

	var targetChainCc contractCaller.IContractCaller
	if chainId == om.config.L1ChainId {
		targetChainCc = l1Cc
	} else {
		targetChainCc, err = om.getContractCallerForChainId(chainId)
		if err != nil {
			om.logger.Sugar().Errorw("Failed to get contract caller for target chain ID",
				zap.Uint32("ChainId", uint32(chainId)),
				zap.Error(err),
			)
			return nil, nil, 0, err
		}
	}

	// no Weights found, go get the latest Weights
	om.logger.Sugar().Debugw("No Weights found for chain",
		zap.Uint32("chainId", uint32(chainId)),
		zap.String("avsAddress", om.config.AvsAddress),
		zap.Uint64("blockNumber", taskBlockNumber),
		zap.Uint32("operatorSetId", operatorSetId),
	)

	var supportedChainsBlockRef int64
	if chainId == om.config.L1ChainId {
		supportedChainsBlockRef = int64(taskBlockNumber)
	} else {
		// if this is not the L1, then we need to use the block number from the latest reference time
		// NOTE: there are potential edge cases where due to the L1 and L2 blocks not aligning 1 to 1.
		// the main risk is someone changing their tableUpdaterAddress to something different
		supportedChainsBlockRef = -1 // use latest block
	}
	destChainIds, tableUpdaterAddresses, err := l1Cc.GetSupportedChainsForMultichain(ctx, supportedChainsBlockRef)
	if err != nil {
		om.logger.Sugar().Errorw("Failed to get supported chains for multichain",
			zap.Uint64("blockNumber", taskBlockNumber),
			zap.String("avsAddress", om.config.AvsAddress),
			zap.Uint32("operatorSetId", operatorSetId),
			zap.Uint32("chainId", uint32(chainId)),
			zap.Error(err),
		)
		return nil, nil, 0, err
	}
	var destTableUpdaterAddress common.Address
	for i, destChainId := range destChainIds {
		if destChainId.Uint64() == uint64(chainId) {
			destTableUpdaterAddress = tableUpdaterAddresses[i]
			break
		}
	}

	// if there is no table updater, then this chain is likely misconfigured or not supported
	if destTableUpdaterAddress == (common.Address{}) {
		om.logger.Sugar().Errorw("No table updater address found for chain",
			zap.Uint32("ChainId", uint32(chainId)),
			zap.Uint64("BlockNumber", taskBlockNumber),
		)
		return nil, nil, 0, fmt.Errorf("no table updater address found for chain ID %d", chainId)
	}

	// this will tell us when the global root was last updated for this chain
	latestReferenceTimeAndBlock, err := targetChainCc.GetTableUpdaterReferenceTimeAndBlock(ctx, destTableUpdaterAddress, taskBlockNumber)
	if err != nil {
		om.logger.Sugar().Errorw("Failed to get latest reference time and block for table updater",
			zap.Uint32("ChainId", uint32(chainId)),
			zap.Uint64("BlockNumber", taskBlockNumber),
			zap.Error(err),
		)
		return nil, nil, 0, err
	}

	var blockForTableData uint64

	// We need to do some potential L2 to L1 block translation:
	// If our task came in on the L2, we need to get the operator table data
	// at the latest reference time and block for the L1.
	//
	// If this is the L1, we can use the task block number directly.
	if chainId == om.config.L1ChainId {
		blockForTableData = taskBlockNumber
	} else {
		// if this is not the L1, then we need to use the block number from the latest reference time
		blockForTableData = uint64(latestReferenceTimeAndBlock.LatestReferenceBlockNumber)
	}

	// weights and table data come from the L1
	tableData, err := l1Cc.GetOperatorTableDataForOperatorSet(ctx, common.HexToAddress(om.config.AvsAddress), operatorSetId, om.config.L1ChainId, blockForTableData)
	if err != nil {
		om.logger.Sugar().Errorw("Failed to get operator table data",
			zap.String("AvsAddress", om.config.AvsAddress),
			zap.Uint32("OperatorSetId", operatorSetId),
			zap.Uint64("BlockNumber", taskBlockNumber),
			zap.Error(err),
		)
		return nil, nil, 0, err
	}

	operatorWeights := make(map[string][]*big.Int, len(tableData.Operators))
	for i, operator := range tableData.Operators {
		weight := tableData.OperatorWeights[i]
		operatorWeights[operator.String()] = weight
	}

	var referenceTimestamp uint32
	if chainId == om.config.L1ChainId {
		referenceTimestamp = tableData.LatestReferenceTimestamp // use task block number as reference timestamp for L1
	} else {
		referenceTimestamp = latestReferenceTimeAndBlock.LatestReferenceTimestamp // use latest reference timestamp for L2
	}

	return tableData, operatorWeights, referenceTimestamp, nil
}

func (om *OperatorManager) filterAndJoinOperatorsWithData(
	operatorWeights map[string][]*big.Int,
	operatorSetId uint32,
	atBlockNumber uint64,
	referenceTimestamp uint32,
	chainId config.ChainId,
	operators []*peering.OperatorPeerInfo,
) (*PeerWeight, error) {
	// filter the list of Operators down to those that are in the operator set and have Weights
	operators = util.Filter(operators, func(op *peering.OperatorPeerInfo) bool {
		for opAddr := range operatorWeights {
			if strings.EqualFold(opAddr, op.OperatorAddress) && op.IncludesOperatorSetId(operatorSetId) {
				return true
			}
		}
		return false
	})

	return &PeerWeight{
		BlockNumber:            atBlockNumber,
		ChainId:                chainId,
		OperatorSetId:          operatorSetId,
		RootReferenceTimestamp: referenceTimestamp,
		Weights:                operatorWeights,
		Operators:              operators,
	}, nil
}

func (om *OperatorManager) GetExecutorPeersAndWeightsForBlock(
	ctx context.Context,
	chainId config.ChainId,
	atBlockNumber uint64,
	operatorSetId uint32,
) (*PeerWeight, error) {
	_, operatorWeights, referenceTimestamp, err := om.GetOperatorSetData(ctx, chainId, atBlockNumber, operatorSetId)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator set data: %w", err)
	}

	operators, err := om.peeringDataFetcher.ListExecutorOperators(ctx, om.config.AvsAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to list executor Operators: %w", err)
	}

	return om.filterAndJoinOperatorsWithData(
		operatorWeights,
		operatorSetId,
		atBlockNumber,
		referenceTimestamp,
		chainId,
		operators,
	)
}

func (om *OperatorManager) GetAggregatorPeersAndWeightsForBlock(
	ctx context.Context,
	chainId config.ChainId,
	atBlockNumber uint64,
	operatorSetId uint32,
) (*PeerWeight, error) {
	_, operatorWeights, referenceTimestamp, err := om.GetOperatorSetData(ctx, chainId, atBlockNumber, operatorSetId)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator set data: %w", err)
	}

	operators, err := om.peeringDataFetcher.ListAggregatorOperators(ctx, om.config.AvsAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to list executor Operators: %w", err)
	}

	return om.filterAndJoinOperatorsWithData(
		operatorWeights,
		operatorSetId,
		atBlockNumber,
		referenceTimestamp,
		chainId,
		operators,
	)
}

func (om *OperatorManager) getContractCallerForChainId(chainId config.ChainId) (contractCaller.IContractCaller, error) {
	cc, ok := om.contractCallers[chainId]
	if !ok {
		return nil, fmt.Errorf("no contract caller found for chain ID %d", chainId)
	}
	return cc, nil
}
