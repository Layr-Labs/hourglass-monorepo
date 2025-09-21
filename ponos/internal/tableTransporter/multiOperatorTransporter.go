package tableTransporter

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ICrossChainRegistry"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IOperatorTableUpdater"
	"github.com/Layr-Labs/multichain-go/pkg/chainManager"
	"github.com/Layr-Labs/multichain-go/pkg/txSigner"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/wealdtech/go-merkletree/v2"
	"github.com/wealdtech/go-merkletree/v2/keccak256"
	"go.uber.org/zap"
)

// OperatorBLSInfo contains the BLS key information for an operator
type OperatorBLSInfo struct {
	PrivateKeyHex    string
	PublicKey        *bn254.PublicKey
	Weights          []*big.Int
	OperatorAddress  common.Address // ECDSA address of the operator
}

// TransportTableWithMultipleOperators supports transporting tables with multiple operators
// for testing purposes. This function updates the GLOBAL generator in OperatorTableUpdater
// with custom operator info, which affects ALL operator sets in the system.
//
// IMPORTANT: To avoid conflicts with other operator sets, this function only transports
// the specific operator set provided in the parameters. Other operator sets will be skipped
// to prevent InvalidRoot errors caused by the modified global generator.
//
// This is primarily intended for testing scenarios where you need specific operator
// configurations that aren't registered through the normal operator registry flow.
func TransportTableWithMultipleOperators(
	transporterPrivateKey string,
	l1RpcUrl string,
	l1ChainId uint64,
	l2RpcUrl string,
	l2ChainId uint64,
	crossChainRegistryAddress string,
	operators []OperatorBLSInfo, // Multiple operators instead of single BLS key
	transportBLSPrivateKey string, // BLS key for signing the transport (can be one of the operators)
	chainIdsToIgnore []*big.Int,
	operatorSetId uint32, // Explicit operator set ID to use
	avsAddress common.Address, // AVS address for the operator set
	l *zap.Logger,
) error {
	ctx := context.Background()

	cm := chainManager.NewChainManager()

	l1AnvilConfig := &chainManager.ChainConfig{
		ChainID: l1ChainId,
		RPCUrl:  l1RpcUrl,
	}
	if err := cm.AddChain(l1AnvilConfig); err != nil {
		return fmt.Errorf("failed to add chain: %v", err)
	}

	holeskyClient, err := cm.GetChainForId(l1AnvilConfig.ChainID)
	if err != nil {
		return fmt.Errorf("failed to get chain for ID %d: %v", l1AnvilConfig.ChainID, err)
	}

	if l2RpcUrl != "" && l2ChainId != 0 {
		l2ChainConfig := &chainManager.ChainConfig{
			ChainID: l2ChainId,
			RPCUrl:  l2RpcUrl,
		}
		if err := cm.AddChain(l2ChainConfig); err != nil {
			return fmt.Errorf("failed to add L2 chain: %v", err)
		}
		l.Sugar().Infow("Added L2 chain", zap.Any("chainConfig", l2ChainConfig))
	}

	txSign, err := txSigner.NewPrivateKeySigner(transporterPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create private key signer: %v", err)
	}

	l.Sugar().Infow("Using CrossChainRegistryAddress",
		zap.String("crossChainRegistryAddress", crossChainRegistryAddress),
	)

	blockNumber, err := holeskyClient.RPCClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get block number: %v", err)
	}

	_, err = holeskyClient.RPCClient.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return fmt.Errorf("failed to get block by number: %v", err)
	}

	// Transfer ownership of CrossChainRegistry
	transferOwnership(l, l1RpcUrl,
		common.HexToAddress(crossChainRegistryAddress),
		transporterPrivateKey)

	// Get supported chains and configure each one
	ccRegistryCaller, err := ICrossChainRegistry.NewICrossChainRegistryCaller(
		common.HexToAddress(crossChainRegistryAddress),
		holeskyClient.RPCClient)
	if err != nil {
		return fmt.Errorf("failed to create cross chain registry caller: %v", err)
	}

	chainIds, addresses, err := ccRegistryCaller.GetSupportedChains(&bind.CallOpts{})
	if err != nil {
		return fmt.Errorf("failed to get supported chains: %v", err)
	}

	for i, chainId := range chainIds {
		addr := addresses[i]
		if chainId.Uint64() != l1ChainId && chainId.Uint64() != l2ChainId {
			continue
		}

		// Determine the correct RPC URL for this chain
		rpcURL := l1RpcUrl
		if chainId.Uint64() == l2ChainId {
			rpcURL = l2RpcUrl
		}

		// Transfer ownership of OperatorTableUpdater
		transferOwnership(l, rpcURL, addr, transporterPrivateKey)

		// Use the provided operator set configuration
		gen := IOperatorTableUpdater.OperatorSet{
			Avs: avsAddress,
			Id:  operatorSetId,
		}

		l.Sugar().Infow("Using operator set configuration",
			zap.String("avsAddress", avsAddress.String()),
			zap.Uint32("operatorSetId", operatorSetId),
			zap.String("chainId", chainId.String()),
		)

		// Connect to ethClient
		client, err := ethclient.Dial(l1RpcUrl)
		if err != nil {
			return fmt.Errorf("failed to connect to L1 RPC: %v", err)
		}
		defer client.Close()

		// Configure curve type for the generator operator set
		const CURVE_TYPE_KEY_REGISTRAR_BN254 = 2
		keyRegistrarAddress := common.HexToAddress("0xA4dB30D08d8bbcA00D40600bee9F029984dB162a")

		if err := configureCurveTypeAsAVS(
			ctx,
			l,
			l1RpcUrl,
			keyRegistrarAddress,
			gen.Avs,
			gen.Id,
			CURVE_TYPE_KEY_REGISTRAR_BN254,
		); err != nil {
			return fmt.Errorf("failed to configure curve type: %v", err)
		}

		// Update the generator with multiple operators
		if err := updateGeneratorWithMultipleOperators(
			ctx, l, cm, chainId, addr, txSign, operators, gen,
		); err != nil {
			return fmt.Errorf("failed to update generator: %v", err)
		}
	}

	// After updating the generator, we're done
	// The transport library will use the updated generator info when needed
	l.Sugar().Infow("Generator updated with custom operators for testing")

	return nil
}

// updateGeneratorWithMultipleOperators updates the generator with multiple operators
func updateGeneratorWithMultipleOperators(
	ctx context.Context,
	logger *zap.Logger,
	cm chainManager.IChainManager,
	chainId *big.Int,
	updaterAddr common.Address,
	txSign txSigner.ITransactionSigner,
	operators []OperatorBLSInfo,
	gen IOperatorTableUpdater.OperatorSet,
) error {
	if len(operators) == 0 {
		return fmt.Errorf("no operators provided")
	}

	chain, err := cm.GetChainForId(chainId.Uint64())
	if err != nil {
		return fmt.Errorf("get chain %d: %w", chainId.Uint64(), err)
	}

	updaterTx, err := IOperatorTableUpdater.NewIOperatorTableUpdater(updaterAddr, chain.RPCClient)
	if err != nil {
		return fmt.Errorf("bind updater tx: %w", err)
	}

	// Build operator info array and calculate aggregate pubkey
	aggregatePubkey := bn254.NewZeroG1Point()
	var totalWeights []*big.Int

	// Calculate operator info leaves for merkle tree
	var leaves [][]byte
	calculatorAddr := common.HexToAddress("0xff58A373c18268F483C1F5cA03Cf885c0C43373a")

	for _, op := range operators {
		// Parse private key if provided, otherwise use public key
		var pubKey *bn254.PublicKey
		if op.PrivateKeyHex != "" {
			sk, err := bn254.NewPrivateKeyFromHexString(strings.TrimPrefix(op.PrivateKeyHex, "0x"))
			if err != nil {
				return fmt.Errorf("parse BLS private key: %w", err)
			}
			pubKey = sk.Public()
		} else if op.PublicKey != nil {
			pubKey = op.PublicKey
		} else {
			return fmt.Errorf("operator must have either private key or public key")
		}

		// Convert to G1 point
		g1 := bn254.NewZeroG1Point().AddPublicKey(pubKey)
		g1b, err := g1.ToPrecompileFormat()
		if err != nil {
			return fmt.Errorf("g1 bytes: %w", err)
		}

		pkG1 := G1Point{
			X: new(big.Int).SetBytes(g1b[0:32]),
			Y: new(big.Int).SetBytes(g1b[32:64]),
		}

		// Use provided weights or default to 1
		weights := op.Weights
		if len(weights) == 0 {
			weights = []*big.Int{big.NewInt(1)}
		}

		info := BN254OperatorInfo{Pubkey: pkG1, Weights: weights}

		// Add to aggregate pubkey
		aggregatePubkey = aggregatePubkey.Add(g1)

		// Add weights to total
		if len(totalWeights) == 0 {
			totalWeights = make([]*big.Int, len(weights))
			for i := range weights {
				totalWeights[i] = new(big.Int)
			}
		}
		for i, w := range weights {
			if i < len(totalWeights) {
				totalWeights[i].Add(totalWeights[i], w)
			}
		}

		// Calculate leaf for merkle tree
		leaf, err := calcOperatorInfoLeaf(ctx, logger, chain.RPCClient, calculatorAddr, info)
		if err != nil {
			return fmt.Errorf("calc operator info leaf: %w", err)
		}
		leaves = append(leaves, leaf[:])
	}

	// Create merkle tree from operator leaves
	var merkleRoot [32]byte
	if len(leaves) == 1 {
		// Single operator case - the root is the leaf itself
		copy(merkleRoot[:], leaves[0])
	} else {
		// Multiple operators - build merkle tree
		tree, err := merkletree.NewTree(
			merkletree.WithData(leaves),
			merkletree.WithHashType(keccak256.New()),
		)
		if err != nil {
			return fmt.Errorf("failed to create merkle tree: %w", err)
		}
		copy(merkleRoot[:], tree.Root())
	}

	// Convert aggregate pubkey to contract format
	aggG1b, err := aggregatePubkey.ToPrecompileFormat()
	if err != nil {
		return fmt.Errorf("aggregate g1 bytes: %w", err)
	}

	aggPkG1 := IOperatorTableUpdater.BN254G1Point{
		X: new(big.Int).SetBytes(aggG1b[0:32]),
		Y: new(big.Int).SetBytes(aggG1b[32:64]),
	}

	// Create generator info
	genInfo := IOperatorTableUpdater.IOperatorTableCalculatorTypesBN254OperatorSetInfo{
		OperatorInfoTreeRoot: merkleRoot,
		NumOperators:         big.NewInt(int64(len(operators))),
		AggregatePubkey:      aggPkG1,
		TotalWeights:         totalWeights,
	}

	logger.Sugar().Infow("Updating generator with multiple operators",
		zap.Int("numOperators", len(operators)),
		zap.String("merkleRoot", fmt.Sprintf("%x", merkleRoot)),
	)

	// Update the generator
	auth, err := txSign.GetTransactOpts(ctx, chainId)
	if err != nil {
		return fmt.Errorf("get tx opts: %w", err)
	}

	tx, err := updaterTx.UpdateGenerator(auth, gen, genInfo)
	if err != nil {
		return fmt.Errorf("updateGenerator tx: %w", err)
	}

	rcpt, err := bind.WaitMined(ctx, chain.RPCClient, tx)
	if err != nil {
		return fmt.Errorf("wait mined: %w", err)
	}
	if rcpt.Status != 1 {
		return fmt.Errorf("updateGenerator reverted: %s", tx.Hash().Hex())
	}

	logger.Sugar().Infow("Successfully updated generator with multiple operators",
		zap.String("txHash", tx.Hash().Hex()),
	)

	return nil
}