package operatorManager

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

type OperatorManagerConfig struct {
	AvsAddress string
	ChainIds   []config.ChainId
	L1ChainId  config.ChainId
}

type PeerWeight struct {
	ChainId                config.ChainId
	OperatorSetId          uint32
	RootReferenceTimestamp uint32
	Weights                map[string][]*big.Int
	Operators              []*peering.OperatorPeerInfo
	CurveType              config.CurveType
	OperatorInfoTreeRoot   [32]byte
	OperatorInfos          []contractCaller.BN254OperatorInfo
}

type OperatorManager struct {
	config             *OperatorManagerConfig
	contractCallers    map[config.ChainId]contractCaller.IContractCaller
	peeringDataFetcher peering.IPeeringDataFetcher
	logger             *zap.Logger
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

func (om *OperatorManager) GetCurveTypeForOperatorSet(avsAddress string, operatorSetId uint32, blockNumber uint64) (config.CurveType, error) {
	l1Cc, err := om.getContractCallerForChainId(om.config.L1ChainId)
	if err != nil {
		om.logger.Sugar().Errorw("Failed to get contract caller for L1 chain ID",
			zap.Uint32("ChainId", uint32(om.config.L1ChainId)),
			zap.Error(err),
		)
		return config.CurveTypeUnknown, err
	}

	return l1Cc.GetOperatorSetCurveType(avsAddress, operatorSetId, blockNumber)
}

func (om *OperatorManager) GetExecutorPeersAndWeightsForTask(
	ctx context.Context,
	task *types.Task,
	curveType config.CurveType,
) (*PeerWeight, error) {
	l1BlockForTableData := task.L1ReferenceBlockNumber

	tableData, err := om.fetchOperatorTableData(
		ctx,
		common.HexToAddress(task.AVSAddress),
		task.OperatorSetId,
		task.ChainId,
		curveType,
		l1BlockForTableData,
		task.SourceBlockNumber,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch operator table data: %w", err)
	}

	operatorWeights := om.buildOperatorWeights(tableData)

	operators, err := om.peeringDataFetcher.ListExecutorOperators(ctx, om.config.AvsAddress, l1BlockForTableData)
	if err != nil {
		return nil, fmt.Errorf("failed to list executor operators: %w", err)
	}

	filteredOperators := om.filterOperatorsForSet(operators, operatorWeights, task.OperatorSetId)

	return &PeerWeight{
		ChainId:                task.ChainId,
		OperatorSetId:          task.OperatorSetId,
		RootReferenceTimestamp: task.ReferenceTimestamp,
		Weights:                operatorWeights,
		Operators:              filteredOperators,
		CurveType:              curveType,
		OperatorInfoTreeRoot:   tableData.OperatorInfoTreeRoot,
		OperatorInfos:          tableData.OperatorInfos,
	}, nil
}

// TODO(seanmcgary): extend/rename this later to support the aggregator as well when we add distributed aggregation
func (om *OperatorManager) GetExecutorPeersAndWeightsForBlock(
	ctx context.Context,
	chainId config.ChainId,
	taskBlockNumber uint64,
	operatorSetId uint32,
) (*PeerWeight, error) {

	isL1Task := chainId == om.config.L1ChainId

	convertedChainId := uint32(chainId)
	om.logger.Sugar().Infow("Getting executor peers and weights for block",
		zap.Uint32("chainId", convertedChainId),
		zap.Uint64("taskBlockNumber", taskBlockNumber),
		zap.Uint32("operatorSetId", operatorSetId),
	)

	l1Cc, err := om.getContractCallerForChainId(om.config.L1ChainId)
	if err != nil {
		om.logger.Sugar().Errorw("Failed to get contract caller for chain ID",
			zap.Uint32("ChainId", convertedChainId),
			zap.Error(err),
		)
		return nil, err
	}

	var targetChainCc contractCaller.IContractCaller
	if isL1Task {
		targetChainCc = l1Cc
	} else {
		targetChainCc, err = om.getContractCallerForChainId(chainId)
		if err != nil {
			om.logger.Sugar().Errorw("Failed to get contract caller for target chain ID",
				zap.Uint32("ChainId", convertedChainId),
				zap.Error(err),
			)
			return nil, err
		}
	}

	om.logger.Sugar().Debugw("Fetching supported chains for multichain",
		zap.Uint32("chainId", convertedChainId),
		zap.Uint64("taskBlockNumber", taskBlockNumber),
		zap.String("avsAddress", om.config.AvsAddress),
		zap.Uint32("operatorSetId", operatorSetId),
	)

	destChainIds, tableUpdaterAddresses, err := l1Cc.GetSupportedChainsForMultichain(ctx, 0)
	if err != nil {
		om.logger.Sugar().Errorw("Failed to get supported chains for multichain",
			zap.Uint64("SourceBlockNumber", taskBlockNumber),
			zap.Error(err),
		)
		return nil, err
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
			zap.Uint32("chainId", convertedChainId),
			zap.Uint64("taskBlockNumber", taskBlockNumber),
		)
		return nil, fmt.Errorf("no table updater address found for chain ID %d", chainId)
	}
	om.logger.Sugar().Infow("Found table updater address for chain",
		zap.Uint32("chainId", convertedChainId),
		zap.Uint64("blockNumber", taskBlockNumber),
		zap.String("tableUpdaterAddress", destTableUpdaterAddress.Hex()),
		zap.String("avsAddress", om.config.AvsAddress),
		zap.Uint32("operatorSetId", operatorSetId),
	)

	// this will tell us when the global root was last updated for this chain
	latestReferenceTimeAndBlock, err := targetChainCc.GetTableUpdaterReferenceTimeAndBlock(ctx, destTableUpdaterAddress, taskBlockNumber)
	if err != nil {
		om.logger.Sugar().Errorw("Failed to get latest reference time and block for table updater",
			zap.Uint32("chainId", convertedChainId),
			zap.Uint64("taskBlockNumber", taskBlockNumber),
			zap.Error(err),
		)
		return nil, err
	}
	om.logger.Sugar().Debugw("Latest reference time and block for table updater",
		zap.Uint32("chainId", convertedChainId),
		zap.Uint64("taskBlockNumber", taskBlockNumber),
		zap.Uint32("latestReferenceBlockNumber", latestReferenceTimeAndBlock.LatestReferenceBlockNumber),
		zap.Uint32("latestReferenceTimestamp", latestReferenceTimeAndBlock.LatestReferenceTimestamp),
		zap.String("avsAddress", om.config.AvsAddress),
		zap.Uint32("operatorSetId", operatorSetId),
	)

	var blockForTableData uint64
	// if this is not the L1, then we need to use the block number from the latest reference time
	blockForTableData = uint64(latestReferenceTimeAndBlock.LatestReferenceBlockNumber)
	if isL1Task {
		// If this is the L1, we can use the task block number directly.
		blockForTableData = taskBlockNumber
	}

	curveType, err := l1Cc.GetOperatorSetCurveType(om.config.AvsAddress, operatorSetId, blockForTableData)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator set curve type: %w", err)
	}
	om.logger.Sugar().Debugw("Got operator set curve type",
		zap.String("avsAddress", om.config.AvsAddress),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.Uint64("blockForTableData", blockForTableData),
		zap.String("curveType", curveType.String()),
	)

	tableData, err := l1Cc.GetOperatorTableDataForOperatorSet(
		ctx,
		common.HexToAddress(om.config.AvsAddress),
		operatorSetId,
		om.config.L1ChainId,
		blockForTableData,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get operator table: %w", err)
	}

	if curveType == config.CurveTypeBN254 {
		err = om.decorateOperatorTableRoot(
			ctx,
			tableData,
			chainId,
			taskBlockNumber,
			common.HexToAddress(om.config.AvsAddress),
			operatorSetId,
			latestReferenceTimeAndBlock,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to populate operator table root: %w", err)
		}
	}

	om.logger.Sugar().Debugw("Fetched operator table data",
		zap.String("avsAddress", om.config.AvsAddress),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.Uint64("taskBlockNumber", taskBlockNumber),
		zap.String("curveType", curveType.String()),
		zap.Int("operatorCount", len(tableData.Operators)),
		zap.Int("weightCount", len(tableData.OperatorWeights)),
		zap.String("operatorInfoTreeRoot", fmt.Sprintf("0x%x", tableData.OperatorInfoTreeRoot)),
	)

	operatorWeights := make(map[string][]*big.Int, len(tableData.Operators))
	for i, operator := range tableData.Operators {
		weight := tableData.OperatorWeights[i]
		operatorWeights[operator.String()] = weight
	}

	operators, err := om.peeringDataFetcher.ListExecutorOperators(ctx, om.config.AvsAddress, blockForTableData)
	if err != nil {
		return nil, fmt.Errorf("failed to list executor Operators: %w", err)
	}
	om.logger.Sugar().Debugw("Fetched executor operators",
		zap.String("avsAddress", om.config.AvsAddress),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.Any("executorOperators", operators),
	)

	// filter the list of Operators down to those that are in the operator set and have Weights
	operators = util.Filter(operators, func(op *peering.OperatorPeerInfo) bool {
		for opAddr := range operatorWeights {
			if strings.EqualFold(opAddr, op.OperatorAddress) && op.IncludesOperatorSetId(operatorSetId) {
				return true
			}
		}
		return false
	})

	var referenceTimestamp uint32
	if chainId == om.config.L1ChainId {
		referenceTimestamp = tableData.LatestReferenceTimestamp
	} else {
		referenceTimestamp = latestReferenceTimeAndBlock.LatestReferenceTimestamp
	}

	return &PeerWeight{
		ChainId:                chainId,
		OperatorSetId:          operatorSetId,
		RootReferenceTimestamp: referenceTimestamp,
		Weights:                operatorWeights,
		Operators:              operators,
		CurveType:              curveType,
		OperatorInfoTreeRoot:   tableData.OperatorInfoTreeRoot,
		OperatorInfos:          tableData.OperatorInfos,
	}, nil
}

func (om *OperatorManager) getContractCallerForChainId(chainId config.ChainId) (contractCaller.IContractCaller, error) {
	cc, ok := om.contractCallers[chainId]
	if !ok {
		return nil, fmt.Errorf("no contract caller found for chain ID %d", chainId)
	}
	return cc, nil
}

func (om *OperatorManager) fetchOperatorTableData(
	ctx context.Context,
	avsAddress common.Address,
	operatorSetId uint32,
	taskChainId config.ChainId,
	curveType config.CurveType,
	l1BlockNumber uint64,
	taskBlockNumber uint64,
) (*contractCaller.OperatorTableData, error) {

	l1Cc, err := om.getContractCallerForChainId(om.config.L1ChainId)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract caller for L1 chain: %w", err)
	}

	om.logger.Sugar().Debugw("Fetching operator table data with curve type",
		zap.String("avsAddress", om.config.AvsAddress),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.String("curveType", curveType.String()),
		zap.Uint64("l1BlockNumber", l1BlockNumber),
	)

	tableData, err := l1Cc.GetOperatorTableDataForOperatorSet(
		ctx, common.HexToAddress(om.config.AvsAddress),
		operatorSetId,
		om.config.L1ChainId,
		l1BlockNumber,
	)
	if err != nil {
		return nil, err
	}

	om.logger.Sugar().Debugw("Fetched operator table data in helper",
		zap.String("avsAddress", om.config.AvsAddress),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.String("curveType", curveType.String()),
		zap.Uint64("l1BlockNumber", l1BlockNumber),
		zap.Int("operatorCount", len(tableData.Operators)),
		zap.String("operatorInfoTreeRoot", fmt.Sprintf("0x%x", tableData.OperatorInfoTreeRoot)),
		zap.Int("operatorInfosCount", len(tableData.OperatorInfos)),
	)

	if curveType == config.CurveTypeBN254 {

		chainIds, tableUpdaterAddresses, err := l1Cc.GetSupportedChainsForMultichain(ctx, l1BlockNumber)

		if err != nil {
			return nil, fmt.Errorf("failed to get supported chains for multichain: %w", err)
		}

		var tableUpdaterAddr common.Address
		tableUpdaterAddressMap := make(map[uint64]common.Address)

		for i, id := range chainIds {
			tableUpdaterAddressMap[id.Uint64()] = tableUpdaterAddresses[i]

			if tableUpdaterAddr == (common.Address{}) && id.Uint64() == uint64(taskChainId) {
				tableUpdaterAddr = tableUpdaterAddresses[i]
			}
		}

		if tableUpdaterAddr == (common.Address{}) {
			return nil, fmt.Errorf("no table updater address found for chain ID %d", taskChainId)
		}

		taskCc, err := om.getContractCallerForChainId(taskChainId)
		if err != nil {
			return nil, fmt.Errorf("failed to get contract caller for chain ID %d: %w", taskChainId, err)
		}

		referenceTimeAndBlock, err := taskCc.GetTableUpdaterReferenceTimeAndBlock(ctx, tableUpdaterAddr, taskBlockNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to get time and block number for chain ID %d: %w", taskChainId, err)
		}

		err = om.decorateOperatorTableRoot(
			ctx,
			tableData,
			taskChainId,
			taskBlockNumber,
			avsAddress,
			operatorSetId,
			referenceTimeAndBlock,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to decorate operator table root: %w", err)
		}
	}

	return tableData, nil
}

// buildOperatorWeights converts operator table data to a weights map
func (om *OperatorManager) buildOperatorWeights(tableData *contractCaller.OperatorTableData) map[string][]*big.Int {
	operatorWeights := make(map[string][]*big.Int, len(tableData.Operators))
	for i, operator := range tableData.Operators {
		operatorWeights[operator.String()] = tableData.OperatorWeights[i]
	}
	return operatorWeights
}

// filterOperatorsForSet filters operators by weight and operator set membership
func (om *OperatorManager) filterOperatorsForSet(
	operators []*peering.OperatorPeerInfo,
	operatorWeights map[string][]*big.Int,
	operatorSetId uint32,
) []*peering.OperatorPeerInfo {
	return util.Filter(operators, func(op *peering.OperatorPeerInfo) bool {
		for opAddr := range operatorWeights {
			if strings.EqualFold(opAddr, op.OperatorAddress) && op.IncludesOperatorSetId(operatorSetId) {
				return true
			}
		}
		return false
	})
}

func (om *OperatorManager) decorateOperatorTableRoot(
	ctx context.Context,
	tableData *contractCaller.OperatorTableData,
	taskChainId config.ChainId,
	taskBlockNumber uint64,
	avsAddress common.Address,
	operatorSetId uint32,
	referenceBlock *contractCaller.LatestReferenceTimeAndBlock,
) error {

	taskCc, err := om.getContractCallerForChainId(taskChainId)
	if err != nil {
		return err
	}

	operatorTableCalculatorCc := taskCc
	if taskChainId != om.config.L1ChainId {
		operatorTableCalculatorCc, err = om.getContractCallerForChainId(om.config.L1ChainId)
		if err != nil {
			return fmt.Errorf("error getting l1 chain contract caller for chain ID %d: %w", taskChainId, err)
		}
	}

	latestTimestamp := referenceBlock.LatestReferenceTimestamp
	referenceBlockNumber := referenceBlock.LatestReferenceBlockNumber

	operatorInfos, err := operatorTableCalculatorCc.GetOperatorInfos(ctx, avsAddress, operatorSetId, uint64(referenceBlockNumber))
	if err != nil {
		return fmt.Errorf("error getting operator infos: %w", err)
	}

	operatorInfoTreeRoot, err := taskCc.GetOperatorInfoTreeRoot(
		ctx,
		avsAddress,
		operatorSetId,
		taskBlockNumber,
		latestTimestamp,
	)
	if err != nil {
		return fmt.Errorf("error getting operator info tree root: %w", err)
	}

	tableData.OperatorInfos = make([]contractCaller.BN254OperatorInfo, len(operatorInfos))
	for i, info := range operatorInfos {
		tableData.OperatorInfos[i] = contractCaller.BN254OperatorInfo{
			PubkeyX: info.Pubkey.X,
			PubkeyY: info.Pubkey.Y,
			Weights: info.Weights,
		}
	}

	tableData.OperatorInfoTreeRoot = operatorInfoTreeRoot
	tableData.LatestReferenceTimestamp = latestTimestamp
	tableData.LatestReferenceBlockNumber = referenceBlockNumber

	return nil
}
