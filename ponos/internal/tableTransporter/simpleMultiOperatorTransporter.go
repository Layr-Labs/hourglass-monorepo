package tableTransporter

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ICrossChainRegistry"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IOperatorTableUpdater"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/multichain-go/pkg/blsSigner"
	"github.com/Layr-Labs/multichain-go/pkg/chainManager"
	"github.com/Layr-Labs/multichain-go/pkg/operatorTableCalculator"
	"github.com/Layr-Labs/multichain-go/pkg/transport"
	"github.com/Layr-Labs/multichain-go/pkg/txSigner"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// SimpleMultiOperatorConfig contains configuration for multi-operator transport
type SimpleMultiOperatorConfig struct {
	TransporterPrivateKey     string
	L1RpcUrl                  string
	L1ChainId                 uint64
	L2RpcUrl                  string
	L2ChainId                 uint64
	CrossChainRegistryAddress string
	ChainIdsToIgnore          []*big.Int
	Logger                    *zap.Logger

	// Multi-operator specific fields
	Operators              []OperatorBLSInfo
	AVSAddress             common.Address
	OperatorSetId          uint32
	TransportBLSPrivateKey string // BLS key for signing transport (can be one of the operators)
}

// TransportTableWithSimpleMultiOperators follows the same pattern as the original
// but supports multiple operators by:
// 1. Registering all operators in the KeyRegistrar
// 2. Building a merkle tree of operator info
// 3. Updating the generator with aggregate pubkey and merkle root
// 4. Calculating and transporting the stake table
func TransportTableWithSimpleMultiOperators(cfg *SimpleMultiOperatorConfig) error {
	ctx := context.Background()

	cm := chainManager.NewChainManager()

	// Add L1 chain
	l1AnvilConfig := &chainManager.ChainConfig{
		ChainID: cfg.L1ChainId,
		RPCUrl:  cfg.L1RpcUrl,
	}
	if err := cm.AddChain(l1AnvilConfig); err != nil {
		return fmt.Errorf("failed to add chain: %v", err)
	}

	holeskyClient, err := cm.GetChainForId(l1AnvilConfig.ChainID)
	if err != nil {
		return fmt.Errorf("failed to get chain for ID %d: %v", l1AnvilConfig.ChainID, err)
	}

	// Add L2 if configured
	if cfg.L2RpcUrl != "" && cfg.L2ChainId != 0 {
		l2ChainConfig := &chainManager.ChainConfig{
			ChainID: cfg.L2ChainId,
			RPCUrl:  cfg.L2RpcUrl,
		}
		if err := cm.AddChain(l2ChainConfig); err != nil {
			return fmt.Errorf("failed to add L2 chain: %v", err)
		}
		cfg.Logger.Sugar().Infow("Added L2 chain", zap.Any("chainConfig", l2ChainConfig))
	}

	txSign, err := txSigner.NewPrivateKeySigner(cfg.TransporterPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create private key signer: %v", err)
	}

	cfg.Logger.Sugar().Infow("Using CrossChainRegistryAddress",
		zap.String("crossChainRegistryAddress", cfg.CrossChainRegistryAddress),
	)

	// Get current block info
	blockNumber, err := holeskyClient.RPCClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get block number: %v", err)
	}

	block, err := holeskyClient.RPCClient.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return fmt.Errorf("failed to get block by number: %v", err)
	}

	referenceTimestamp := uint32(block.Time())

	// Transfer ownership of CrossChainRegistry
	transferOwnership(cfg.Logger, cfg.L1RpcUrl,
		common.HexToAddress(cfg.CrossChainRegistryAddress),
		cfg.TransporterPrivateKey)

	// Get supported chains
	ccRegistryCaller, err := ICrossChainRegistry.NewICrossChainRegistryCaller(
		common.HexToAddress(cfg.CrossChainRegistryAddress),
		holeskyClient.RPCClient)
	if err != nil {
		return fmt.Errorf("failed to create cross chain registry caller: %v", err)
	}

	chainIds, addresses, err := ccRegistryCaller.GetSupportedChains(&bind.CallOpts{})
	if err != nil {
		return fmt.Errorf("failed to get supported chains: %v", err)
	}

	// Create L1 client and contract caller for KeyRegistrar operations
	client, err := ethclient.Dial(cfg.L1RpcUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to L1 RPC: %v", err)
	}
	defer client.Close()

	signer, err := transactionSigner.NewPrivateKeySigner(cfg.TransporterPrivateKey, client, cfg.Logger)
	if err != nil {
		return fmt.Errorf("failed to create private key signer: %v", err)
	}

	contractCaller, err := caller.NewContractCaller(client, signer, cfg.Logger)
	if err != nil {
		return fmt.Errorf("failed to create contract caller: %v", err)
	}

	const CURVE_TYPE_KEY_REGISTRAR_BN254 = 2
	keyRegistrarAddress := common.HexToAddress("0xA4dB30D08d8bbcA00D40600bee9F029984dB162a")

	// Process each chain
	for i, chainId := range chainIds {
		addr := addresses[i]
		if chainId.Uint64() != cfg.L1ChainId && chainId.Uint64() != cfg.L2ChainId {
			continue
		}

		// Determine the correct RPC URL for this chain
		rpcURL := cfg.L1RpcUrl
		if chainId.Uint64() == cfg.L2ChainId {
			rpcURL = cfg.L2RpcUrl
		}

		// Transfer ownership of OperatorTableUpdater
		transferOwnership(cfg.Logger, rpcURL, addr, cfg.TransporterPrivateKey)

		// Use the provided operator set configuration
		gen := IOperatorTableUpdater.OperatorSet{
			Avs: cfg.AVSAddress,
			Id:  cfg.OperatorSetId,
		}

		// Configure curve type for the operator set
		if err := configureCurveTypeAsAVS(
			ctx,
			cfg.Logger,
			cfg.L1RpcUrl, // KeyRegistrar is on L1
			keyRegistrarAddress,
			gen.Avs,
			gen.Id,
			CURVE_TYPE_KEY_REGISTRAR_BN254,
		); err != nil {
			return fmt.Errorf("failed to configure curve type: %v", err)
		}

		// Register BLS keys in KeyRegistrar for all operators
		// This is necessary for the operators to be recognized
		if err := registerOperatorKeysInKeyRegistrar(
			ctx,
			cfg.Logger,
			contractCaller,
			cfg.Operators,
			gen,
		); err != nil {
			cfg.Logger.Sugar().Warnw("Failed to register some operator keys",
				zap.Error(err),
			)
			// Continue anyway - keys might already be registered
		}

		// Read current generator from the contract
		currentGen, err := getGenerator(ctx, cfg.Logger, cm, chainId, addr)
		if err != nil {
			// If we can't read the generator, use a default
			cfg.Logger.Sugar().Warnw("Failed to get current generator, using default",
				zap.Error(err),
			)
			// Use AVS address 0x0 with operator set ID 0 as default
			currentGen = IOperatorTableUpdater.OperatorSet{
				Avs: common.HexToAddress("0x0000000000000000000000000000000000000000"),
				Id:  0,
			}
		}

		// Switch to a different operator set ID for the generator
		// This ensures the generator is different from any operator set being updated
		generatorOpSet := IOperatorTableUpdater.OperatorSet{
			Avs: currentGen.Avs, // Keep the same AVS address
			Id:  currentGen.Id,   // Start with current ID
		}

		// Switch the operator set ID
		if generatorOpSet.Id == 1 {
			generatorOpSet.Id = 2
		} else {
			generatorOpSet.Id = 1
		}

		cfg.Logger.Sugar().Infow("Setting up generator operator set",
			zap.String("currentGeneratorAvs", currentGen.Avs.String()),
			zap.Uint32("currentGeneratorId", currentGen.Id),
			zap.String("newGeneratorAvs", generatorOpSet.Avs.String()),
			zap.Uint32("newGeneratorId", generatorOpSet.Id),
			zap.String("operatorTableUpdaterAddress", addr.String()),
			zap.String("chainId", chainId.String()),
		)

		// Update the generator with the transport BLS key
		if err := updateGeneratorFromContext(
			ctx,
			cfg.Logger,
			cm,
			chainId, // Use the actual chain ID for this OperatorTableUpdater
			addr,    // OperatorTableUpdater address
			txSign,
			cfg.TransportBLSPrivateKey,
			generatorOpSet, // Use the generator operator set, not our test operator set
		); err != nil {
			// Log the error but continue - generator might already be set
			cfg.Logger.Sugar().Warnw("Failed to update generator (may already be configured)",
				zap.Error(err),
			)
		} else {
			cfg.Logger.Sugar().Infow("Successfully updated generator",
				zap.String("generatorAvs", generatorOpSet.Avs.String()),
				zap.Uint32("generatorId", generatorOpSet.Id),
			)
		}
	}

	// Now calculate stake table root and transport
	tableCalc, err := operatorTableCalculator.NewStakeTableRootCalculator(&operatorTableCalculator.Config{
		CrossChainRegistryAddress: common.HexToAddress(cfg.CrossChainRegistryAddress),
	}, holeskyClient.RPCClient, cfg.Logger)
	if err != nil {
		return fmt.Errorf("failed to create StakeTableRootCalculator: %v", err)
	}

	root, tree, dist, err := tableCalc.CalculateStakeTableRoot(ctx, block.NumberU64())
	if err != nil {
		return fmt.Errorf("failed to calculate stake table root: %v", err)
	}

	// Log all operator sets found in the distribution
	cfg.Logger.Sugar().Infow("Calculated stake table root - checking operator sets",
		zap.String("root", fmt.Sprintf("0x%x", root)),
		zap.Uint64("blockNumber", block.NumberU64()),
	)

	allOpsets := dist.GetOperatorSets()
	cfg.Logger.Sugar().Infow("All operator sets in distribution",
		zap.Int("count", len(allOpsets)),
	)
	for i, opset := range allOpsets {
		cfg.Logger.Sugar().Infow("Operator set in distribution",
			zap.Int("index", i),
			zap.String("avs", opset.Avs.String()),
			zap.Uint32("id", opset.Id),
		)
	}

	// Create BLS signer for transport
	pk, err := bn254.NewPrivateKeyFromHexString(cfg.TransportBLSPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to convert transport BLS private key: %v", err)
	}

	inMemSigner, err := blsSigner.NewInMemoryBLSSigner(pk)
	if err != nil {
		return fmt.Errorf("failed to create in-memory BLS signer: %v", err)
	}

	stakeTransport, err := transport.NewTransport(
		&transport.TransportConfig{
			L1CrossChainRegistryAddress: common.HexToAddress(cfg.CrossChainRegistryAddress),
		},
		holeskyClient.RPCClient,
		inMemSigner,
		txSign,
		cm,
		cfg.Logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create transport: %v", err)
	}

	// Transport global table root
	err = stakeTransport.SignAndTransportGlobalTableRoot(
		ctx,
		root,
		referenceTimestamp,
		block.NumberU64(),
		cfg.ChainIdsToIgnore,
	)
	if err != nil {
		return fmt.Errorf("failed to sign and transport global table root: %v", err)
	}

	// Transport AVS stake tables - only for our specific operator set
	opsets := dist.GetOperatorSets()
	transportedOurOpset := false

	for _, opset := range opsets {
		// Only transport our specific operator set
		if opset.Avs != cfg.AVSAddress || opset.Id != cfg.OperatorSetId {
			cfg.Logger.Sugar().Debugw("Skipping operator set (not ours)",
				zap.String("avs", opset.Avs.String()),
				zap.Uint32("id", opset.Id),
				zap.String("ourAvs", cfg.AVSAddress.String()),
				zap.Uint32("ourId", cfg.OperatorSetId),
			)
			continue
		}

		cfg.Logger.Sugar().Infow("Transporting our operator set",
			zap.String("avs", opset.Avs.String()),
			zap.Uint32("id", opset.Id),
			zap.Uint32("referenceTimestamp", referenceTimestamp),
			zap.Uint64("blockNumber", block.NumberU64()),
		)

		// Get the operator table data for this operator set
		operatorTableBytes, exists := dist.GetTableData(opset)
		if !exists {
			cfg.Logger.Sugar().Warnw("No operator table bytes found for operator set",
				zap.String("avs", opset.Avs.String()),
				zap.Uint32("id", opset.Id),
			)
		} else {
			cfg.Logger.Sugar().Infow("Found operator table bytes for transport",
				zap.String("avs", opset.Avs.String()),
				zap.Uint32("id", opset.Id),
				zap.Int("bytesLength", len(operatorTableBytes)),
				zap.String("bytesHex", fmt.Sprintf("0x%x", operatorTableBytes[:32])), // First 32 bytes
			)
		}

		err = stakeTransport.SignAndTransportAvsStakeTable(
			ctx,
			referenceTimestamp,
			block.NumberU64(),
			opset,
			root,
			tree,
			dist,
			cfg.ChainIdsToIgnore,
		)
		if err != nil {
			return fmt.Errorf("failed to transport AVS stake table for our operator set: %w", err)
		}

		cfg.Logger.Sugar().Infow("Successfully transported our AVS stake table",
			zap.Any("opset", opset),
		)
		transportedOurOpset = true
	}

	if !transportedOurOpset {
		return fmt.Errorf("our operator set (AVS: %s, ID: %d) was not found in the calculated stake tables",
			cfg.AVSAddress.String(), cfg.OperatorSetId)
	}

	return nil
}

// registerOperatorKeysInKeyRegistrar just registers the BLS keys in KeyRegistrar
// without building generator info (the transport process will handle that)
func registerOperatorKeysInKeyRegistrar(
	ctx context.Context,
	logger *zap.Logger,
	contractCaller *caller.ContractCaller,
	operators []OperatorBLSInfo,
	gen IOperatorTableUpdater.OperatorSet,
) error {
	if len(operators) == 0 {
		return fmt.Errorf("no operators provided")
	}

	for _, op := range operators {
		// Parse private key
		var sk *bn254.PrivateKey
		if op.PrivateKeyHex != "" {
			var err error
			sk, err = bn254.NewPrivateKeyFromHexString(strings.TrimPrefix(op.PrivateKeyHex, "0x"))
			if err != nil {
				return fmt.Errorf("parse BLS private key: %w", err)
			}
		} else {
			return fmt.Errorf("operator must have private key for registration")
		}

		pubKey := sk.Public()

		// Encode key data for registration
		keyData, err := contractCaller.EncodeBN254KeyData(pubKey)
		if err != nil {
			return fmt.Errorf("encode key data: %w", err)
		}

		// Get message hash and sign with operator's address
		msgHash, err := contractCaller.GetOperatorRegistrationMessageHash(
			ctx,
			op.OperatorAddress,
			gen.Avs,
			gen.Id,
			keyData,
		)
		if err != nil {
			return fmt.Errorf("failed to get operator registration message: %w", err)
		}

		sig, err := sk.SignSolidityCompatible(msgHash)
		if err != nil {
			return fmt.Errorf("BLS sign: %w", err)
		}

		// Register in KeyRegistrar with operator's address
		if _, err := contractCaller.RegisterKeyWithKeyRegistrar(
			ctx,
			op.OperatorAddress,
			gen.Avs,
			gen.Id,
			keyData,
			sig.Bytes(),
		); err != nil {
			logger.Sugar().Warnw("Failed to register key (may already be registered)",
				zap.String("operator", op.OperatorAddress.String()),
				zap.Error(err),
			)
			// Continue with other operators
		} else {
			logger.Sugar().Infow("Successfully registered operator key",
				zap.String("operator", op.OperatorAddress.String()),
			)
		}
	}

	return nil
}
