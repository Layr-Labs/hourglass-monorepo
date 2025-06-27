package avsExecutionManager

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operatorManager"
	"go.uber.org/zap"
	"golang.org/x/crypto/sha3"
	"math/big"
	"strings"
	"sync"
)

const (
	// TODO(seanmcgary): this will come from the AVSConfig eventually...
	aggregatorOperatorSetId = 0
)

type StakeWeightedAggregationStrategy struct {
	sync.Mutex
	logger                  *zap.Logger
	operatorManager         *operatorManager.OperatorManager
	previousBlock           *ethereum.EthereumBlock
	currentBlock            *ethereum.EthereumBlock
	isLeaderForCurrentBlock bool
	avsAddress              string
}

func NewStakeWeightedAggregationStrategy(
	avsAddress string,
	om *operatorManager.OperatorManager,
	logger *zap.Logger,
) *StakeWeightedAggregationStrategy {
	return &StakeWeightedAggregationStrategy{
		logger:          logger,
		operatorManager: om,
		avsAddress:      avsAddress,
	}
}

func (swas *StakeWeightedAggregationStrategy) IsLeaderForBlock(ctx context.Context, block *ethereum.EthereumBlock) (bool, error) {
	swas.Lock()
	defer swas.Unlock()

	if swas.previousBlock != nil && swas.previousBlock.Number == block.Number {
		swas.logger.Debug("Already processed block", zap.Uint64("blockNumber", block.Number.Value()))
		return swas.isLeaderForCurrentBlock, nil
	}

	if swas.previousBlock != swas.currentBlock {
		swas.previousBlock = swas.currentBlock
		swas.currentBlock = block
		swas.isLeaderForCurrentBlock = false
	}

	aggregators, err := swas.operatorManager.GetAggregatorPeersAndWeightsForBlock(
		ctx,
		block.ChainId,
		block.Number.Value(),
		aggregatorOperatorSetId,
	)
	if err != nil {
		return false, fmt.Errorf("failed to get aggregator peers and weights: %w", err)
	}
	if aggregators == nil || len(aggregators.Operators) == 0 {
		swas.logger.Error("no aggregators found for block",
			zap.Uint64("blockNumber", block.Number.Value()),
			zap.Uint("chainId", uint(block.ChainId)),
			zap.String("avsAddress", swas.avsAddress),
		)

		return false, fmt.Errorf("no aggregators found for block %d", block.Number.Value())
	}
	if len(aggregators.Operators) == 1 && strings.EqualFold(aggregators.Operators[0].OperatorAddress, swas.avsAddress) {
		swas.isLeaderForCurrentBlock = true
		return true, nil
	}

	// use mixHash + avsAddress as a seed to determine if this AVS is the leader based on stake weight
	swas.logger.Debug("Calculating leader for block",
		zap.Uint64("blockNumber", block.Number.Value()),
		zap.String("avsAddress", swas.avsAddress),
		zap.String("mixHash", block.MixHash.Value()),
	)

	selectedLeader, err := selectStakeWeightedLeader(aggregators, block.MixHash.Value(), swas.avsAddress)
	if err != nil {
		return false, err
	}

	isLeader := strings.EqualFold(selectedLeader, swas.avsAddress)
	swas.isLeaderForCurrentBlock = isLeader

	swas.logger.Debug("Leader selection result",
		zap.Bool("isLeader", isLeader),
		zap.String("selectedLeader", selectedLeader),
	)

	return isLeader, nil
}

// selectStakeWeightedLeader selects a leader from aggregators based on their stake weights using mixHash and avsAddress as seed
func selectStakeWeightedLeader(aggregators *operatorManager.PeerWeight, mixHash string, avsAddress string) (string, error) {
	if aggregators == nil || len(aggregators.Operators) == 0 {
		return "", fmt.Errorf("no operators provided")
	}

	// Calculate total stake weight and build operator weights map
	totalWeight := big.NewInt(0)
	operatorWeights := make(map[string]*big.Int)

	for _, operator := range aggregators.Operators {
		weights, exists := aggregators.Weights[operator.OperatorAddress]
		if !exists || len(weights) == 0 {
			continue
		}
		// The 0th element will be the sum of shares across all strategies in the opset
		weight := weights[0]
		if weight == nil {
			weight = new(big.Int).SetInt64(0)
		}
		if weight.Cmp(big.NewInt(0)) > 0 {
			operatorWeights[operator.OperatorAddress] = weight
			totalWeight.Add(totalWeight, weight)
		}
	}

	if totalWeight.Cmp(big.NewInt(0)) == 0 {
		return "", fmt.Errorf("no operators with stake weight found")
	}

	// Use mixHash + avsAddress as random seed
	seed := calculateSeedFromMixHashAndAVS(mixHash, avsAddress)

	// Calculate weighted random selection
	target := new(big.Int).Mod(seed, totalWeight)
	currentSum := big.NewInt(0)

	for _, operator := range aggregators.Operators {
		weight, exists := operatorWeights[operator.OperatorAddress]
		if !exists {
			continue
		}

		currentSum.Add(currentSum, weight)
		if target.Cmp(currentSum) < 0 {
			return operator.OperatorAddress, nil
		}
	}

	// Should never reach here if logic is correct
	return "", fmt.Errorf("failed to select leader")
}

// calculateSeedFromMixHashAndAVS creates a seed using keccak256(mixHash + avsAddress) to match Solidity behavior
func calculateSeedFromMixHashAndAVS(mixHash string, avsAddress string) *big.Int {
	// Remove 0x prefix if present
	mixHash = strings.TrimPrefix(mixHash, "0x")
	avsAddress = strings.TrimPrefix(avsAddress, "0x")

	// Create keccak256 hasher (equivalent to Solidity's keccak256)
	hasher := sha3.NewLegacyKeccak256()

	// Write mixHash bytes (padded to 32 bytes like Solidity abi.encode)
	mixHashBytes := make([]byte, 32)
	if len(mixHash) > 0 {
		if mixHashData, err := hex.DecodeString(mixHash); err == nil {
			// Right-pad shorter values, left-pad longer values to fit 32 bytes
			if len(mixHashData) <= 32 {
				copy(mixHashBytes[32-len(mixHashData):], mixHashData)
			} else {
				copy(mixHashBytes, mixHashData[:32])
			}
		}
	}
	hasher.Write(mixHashBytes)

	// Write avsAddress bytes (padded to 32 bytes like Solidity abi.encode)
	avsAddressBytes := make([]byte, 32)
	if len(avsAddress) > 0 {
		if avsData, err := hex.DecodeString(avsAddress); err == nil {
			// Right-pad shorter values, left-pad longer values to fit 32 bytes
			if len(avsData) <= 32 {
				copy(avsAddressBytes[32-len(avsData):], avsData)
			} else {
				copy(avsAddressBytes, avsData[:32])
			}
		}
	}
	hasher.Write(avsAddressBytes)

	hash := hasher.Sum(nil)

	// Convert to big.Int
	seed := new(big.Int).SetBytes(hash)
	return seed
}
