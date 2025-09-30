package tableTransporter

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ICrossChainRegistry"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IKeyRegistrar"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IOperatorTableUpdater"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/multichain-go/pkg/blsSigner"
	"github.com/Layr-Labs/multichain-go/pkg/chainManager"
	"github.com/Layr-Labs/multichain-go/pkg/distribution"
	"github.com/Layr-Labs/multichain-go/pkg/operatorTableCalculator"
	"github.com/Layr-Labs/multichain-go/pkg/transport"
	"github.com/Layr-Labs/multichain-go/pkg/txSigner"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/wealdtech/go-merkletree/v2"
	"go.uber.org/zap"
)

// Map config.CurveType to KeyRegistrar curve type constants
const (
	ecdsaCurveType = 1
	bn254CurveType = 2
)

// OperatorKeyInfo contains generic key information for an operator
type OperatorKeyInfo struct {
	PrivateKeyHex   string
	Weights         []*big.Int
	OperatorAddress common.Address
}

// MultipleOperatorConfig contains configuration for multi-operator transport
type MultipleOperatorConfig struct {
	TransporterPrivateKey     string
	L1RpcUrl                  string
	L1ChainId                 uint64
	L2RpcUrl                  string
	L2ChainId                 uint64
	CrossChainRegistryAddress string
	ChainIdsToIgnore          []*big.Int
	Logger                    *zap.Logger

	// Multi-operator specific fields
	Operators              []OperatorKeyInfo
	AVSAddress             common.Address
	OperatorSetId          uint32
	CurveType              config.CurveType
	TransportBLSPrivateKey string
}

// TransportTableWithSimpleMultiOperators follows the same pattern as the original
// but supports multiple operators by:
// 1. Registering all operators in the KeyRegistrar
// 2. Building a merkle tree of operator info
// 3. Updating the generator with aggregate pubkey and merkle root
// 4. Calculating and transporting the stake table
func TransportTableWithSimpleMultiOperators(cfg *MultipleOperatorConfig) error {
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

	l1ChainClient, err := cm.GetChainForId(l1AnvilConfig.ChainID)
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
	blockNumber, err := l1ChainClient.RPCClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get block number: %v", err)
	}

	block, err := l1ChainClient.RPCClient.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return fmt.Errorf("failed to get block by number: %v", err)
	}

	referenceTimestamp := uint32(block.Time())

	// Transfer ownership of CrossChainRegistry
	transferOwnership(cfg.Logger, cfg.L1RpcUrl,
		common.HexToAddress(cfg.CrossChainRegistryAddress),
		cfg.TransporterPrivateKey,
	)

	// Get supported chains
	ccRegistryCaller, err := ICrossChainRegistry.NewICrossChainRegistryCaller(
		common.HexToAddress(cfg.CrossChainRegistryAddress),
		l1ChainClient.RPCClient)
	if err != nil {
		return fmt.Errorf("failed to create cross chain registry caller: %v", err)
	}

	chainIds, updaterAddresses, err := ccRegistryCaller.GetSupportedChains(&bind.CallOpts{})
	if err != nil {
		return fmt.Errorf("failed to get supported chains: %v", err)
	}
	cfg.Logger.Sugar().Infow("Found supported chains", zap.Any("chainIds", chainIds))

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

	var curveTypeKeyRegistrar uint8
	if cfg.CurveType == config.CurveTypeBN254 {
		curveTypeKeyRegistrar = bn254CurveType
	} else if cfg.CurveType == config.CurveTypeECDSA {
		curveTypeKeyRegistrar = ecdsaCurveType
	} else {
		return fmt.Errorf("unsupported curve type: %s", cfg.CurveType)
	}

	keyRegistrarAddress := common.HexToAddress("0xA4dB30D08d8bbcA00D40600bee9F029984dB162a")

	gen := IOperatorTableUpdater.OperatorSet{
		Avs: cfg.AVSAddress,
		Id:  cfg.OperatorSetId,
	}

	if err := configureCurveTypeAsAVS(
		ctx,
		cfg.Logger,
		cfg.L1RpcUrl,
		keyRegistrarAddress,
		gen.Avs,
		gen.Id,
		curveTypeKeyRegistrar,
	); err != nil {
		return fmt.Errorf("failed to configure curve type: %v", err)
	}

	if err := registerOperatorKeysInKeyRegistrar(
		ctx,
		cfg.Logger,
		contractCaller,
		client,
		cfg.Operators,
		gen,
		cfg.CurveType,
	); err != nil {
		return fmt.Errorf("failed to register operators: %v", err)
	}

	for i, chainId := range chainIds {
		updaterAddress := updaterAddresses[i]
		if chainId.Uint64() != cfg.L1ChainId && chainId.Uint64() != cfg.L2ChainId {
			continue
		}

		rpcURL := cfg.L1RpcUrl
		if chainId.Uint64() == cfg.L2ChainId {
			rpcURL = cfg.L2RpcUrl
		}

		transferOwnership(cfg.Logger, rpcURL, updaterAddress, cfg.TransporterPrivateKey)

		currentGen, err := getGenerator(ctx, cfg.Logger, cm, chainId, updaterAddress)
		if err != nil {
			return fmt.Errorf("failed to get current generator: %v", err)
		}

		// Switch to a different operator set ID for the generator
		// This ensures the generator is different from any operator set being updated
		generatorOpSet := IOperatorTableUpdater.OperatorSet{
			Avs: currentGen.Avs, // Keep the same AVS address
			Id:  currentGen.Id,  // Start with current ID
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
			zap.String("operatorTableUpdaterAddress", updaterAddress.String()),
			zap.String("chainId", chainId.String()),
		)

		// The generator uses its own single-operator configuration with the transport BLS key
		// This is NOT the same as the target operator sets we're transporting
		if err := updateGeneratorFromContext(
			ctx,
			cfg.Logger,
			cm,
			chainId,
			updaterAddress,
			txSign,
			cfg.TransportBLSPrivateKey,
			generatorOpSet,
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
	cfg.Logger.Sugar().Infow("Creating StakeTableRootCalculator with CrossChainRegistry",
		zap.String("crossChainRegistryAddress", cfg.CrossChainRegistryAddress),
		zap.Uint64("blockNumber", block.NumberU64()),
	)

	tableCalc, err := operatorTableCalculator.NewStakeTableRootCalculator(&operatorTableCalculator.Config{
		CrossChainRegistryAddress: common.HexToAddress(cfg.CrossChainRegistryAddress),
	}, l1ChainClient.RPCClient, cfg.Logger)
	if err != nil {
		return fmt.Errorf("failed to create StakeTableRootCalculator: %v", err)
	}

	cfg.Logger.Sugar().Infow("Calculating stake table root from CrossChainRegistry",
		zap.String("info", "This fetches operator data from CrossChainRegistry, NOT from KeyRegistrar"),
	)

	root, tree, dist, err := tableCalc.CalculateStakeTableRoot(ctx, block.NumberU64())
	if err != nil {
		return fmt.Errorf("failed to calculate stake table root: %v", err)
	}

	// Log all operator sets found in the distribution
	cfg.Logger.Sugar().Infow("Calculated stake table root - checking operator sets",
		zap.String("root", fmt.Sprintf("0x%x", root)),
		zap.Uint64("blockNumber", block.NumberU64()),
	)

	// Debug: Log operator data from the distribution for our operator set
	for _, opset := range dist.GetOperatorSets() {
		if opset.Avs == cfg.AVSAddress && opset.Id == cfg.OperatorSetId {
			tableData, exists := dist.GetTableData(opset)
			if exists {
				cfg.Logger.Sugar().Infow("Found operator table data for our operator set",
					zap.String("avs", opset.Avs.String()),
					zap.Uint32("id", opset.Id),
					zap.Int("dataLength", len(tableData)),
				)

				// Log what operators the StakeTableRootCalculator found
				cfg.Logger.Sugar().Infow("StakeTableRootCalculator operator data preview",
					zap.String("first32bytes", fmt.Sprintf("0x%x", tableData[:min(32, len(tableData))])),
					zap.String("WARNING", "This data comes from CrossChainRegistry, NOT from KeyRegistrar where we registered our test operators"),
				)
			}
		}
	}

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
		&transport.TransportConfig{L1CrossChainRegistryAddress: common.HexToAddress(cfg.CrossChainRegistryAddress)},
		l1ChainClient.RPCClient,
		inMemSigner,
		txSign,
		cm,
		cfg.Logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create transport: %v", err)
	}

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

	// After transport, we need to explicitly update the operator table in the certificate verifier
	// The SignAndTransportAvsStakeTable updates the OperatorTableUpdater, but we need to ensure
	// the data is propagated to the BN254CertificateVerifier
	cfg.Logger.Sugar().Infow("Updating operator table in certificate verifier after transport")

	// For each chain, call updateOperatorTable to propagate the data
	for i, chainId := range chainIds {
		addr := updaterAddresses[i]
		if chainId.Uint64() != cfg.L1ChainId && chainId.Uint64() != cfg.L2ChainId {
			continue
		}

		// Update the operator table for our specific operator set
		if err := updateOperatorTableInVerifier(
			ctx,
			cfg.Logger,
			cm,
			chainId,
			addr,
			txSign,
			cfg.AVSAddress,
			cfg.OperatorSetId,
			referenceTimestamp,
			root,
			tree,
			dist,
		); err != nil {
			return fmt.Errorf("failed to update operator table: %w", err)
		}
	}

	return nil
}

// updateOperatorTableInVerifier explicitly calls updateOperatorTable on the OperatorTableUpdater
// to ensure the operator table data is propagated to the BN254CertificateVerifier
func updateOperatorTableInVerifier(
	ctx context.Context,
	logger *zap.Logger,
	cm chainManager.IChainManager,
	chainId *big.Int,
	updaterAddr common.Address,
	txSign txSigner.ITransactionSigner,
	avsAddress common.Address,
	operatorSetId uint32,
	referenceTimestamp uint32,
	globalTableRoot [32]byte,
	tree *merkletree.MerkleTree,
	dist *distribution.Distribution,
) error {
	chain, err := cm.GetChainForId(chainId.Uint64())
	if err != nil {
		return fmt.Errorf("get chain %d: %w", chainId.Uint64(), err)
	}

	// Get the operator set - use the distribution's OperatorSet type
	opset := distribution.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}

	// Get the operator table data for this operator set
	operatorTableBytes, exists := dist.GetTableData(opset)
	if !exists {
		return fmt.Errorf("no operator table bytes found for operator set (AVS: %s, ID: %d)",
			avsAddress.String(), operatorSetId)
	}

	// Get the merkle proof for this operator set
	operatorSetIndex, exists := dist.GetTableIndex(opset)
	if !exists {
		return fmt.Errorf("operator set not found in distribution")
	}

	// Generate the merkle proof for this operator set
	proof, err := tree.GenerateProofWithIndex(operatorSetIndex, 0)
	if err != nil {
		return fmt.Errorf("generate merkle proof: %w", err)
	}

	// Convert proof to bytes - flatten the hashes into a single byte array
	proofBytes := make([]byte, 0)
	for _, hash := range proof.Hashes {
		proofBytes = append(proofBytes, hash...)
	}

	// Create the OperatorTableUpdater binding
	updaterContract, err := IOperatorTableUpdater.NewIOperatorTableUpdater(updaterAddr, chain.RPCClient)
	if err != nil {
		return fmt.Errorf("bind operator table updater: %w", err)
	}

	// Call updateOperatorTable to propagate the data to BN254CertificateVerifier
	auth, err := txSign.GetTransactOpts(ctx, chainId)
	if err != nil {
		return fmt.Errorf("get transact opts: %w", err)
	}

	logger.Sugar().Infow("Calling updateOperatorTable to propagate data to certificate verifier",
		"avs", avsAddress.String(),
		"operatorSetId", operatorSetId,
		"referenceTimestamp", referenceTimestamp,
		"operatorSetIndex", operatorSetIndex,
		"operatorTableBytesLen", len(operatorTableBytes),
		"chainId", chainId.String(),
		"updaterAddress", updaterAddr,
	)

	tx, err := updaterContract.UpdateOperatorTable(
		auth,
		referenceTimestamp,
		globalTableRoot,
		uint32(operatorSetIndex),
		proofBytes,
		operatorTableBytes,
	)
	if err != nil {
		// Check if it's already updated
		if strings.Contains(err.Error(), "TableUpdateForPastTimestamp") {
			logger.Sugar().Infow("Operator table already updated for this timestamp",
				"avs", avsAddress.String(),
				"operatorSetId", operatorSetId,
				"referenceTimestamp", referenceTimestamp,
			)
			return nil
		}
		return fmt.Errorf("updateOperatorTable tx: %w", err)
	}

	rcpt, err := bind.WaitMined(ctx, chain.RPCClient, tx)
	if err != nil {
		return fmt.Errorf("wait mined: %w", err)
	}
	if rcpt.Status != 1 {
		return fmt.Errorf("updateOperatorTable reverted: %s", tx.Hash().Hex())
	}

	logger.Sugar().Infow("Successfully updated operator table in certificate verifier",
		"txHash", tx.Hash().Hex(),
		"avs", avsAddress.String(),
		"operatorSetId", operatorSetId,
		"referenceTimestamp", referenceTimestamp,
	)

	return nil
}

// registerOperatorKeysInKeyRegistrar registers the keys in KeyRegistrar
// by impersonating each operator to avoid permission issues
func registerOperatorKeysInKeyRegistrar(
	ctx context.Context,
	logger *zap.Logger,
	contractCaller *caller.ContractCaller,
	client *ethclient.Client,
	operators []OperatorKeyInfo,
	gen IOperatorTableUpdater.OperatorSet,
	curveType config.CurveType,
) error {
	for _, op := range operators {
		if err := registerSingleOperatorKey(ctx, logger, contractCaller, client, op, gen, curveType); err != nil {
			logger.Sugar().Warnw("Failed to register operator key",
				zap.String("operator", op.OperatorAddress.String()),
				zap.Error(err),
			)
			return err
		}
	}

	return nil
}

// checkOperatorRegistered checks if an operator is already registered in the KeyRegistrar
func checkOperatorRegistered(
	ctx context.Context,
	client *ethclient.Client,
	operatorAddress common.Address,
	avsAddress common.Address,
	operatorSetId uint32,
) (bool, error) {
	keyRegistrarAddress := common.HexToAddress("0xA4dB30D08d8bbcA00D40600bee9F029984dB162a")
	keyRegistrar, err := IKeyRegistrar.NewIKeyRegistrar(keyRegistrarAddress, client)
	if err != nil {
		return false, fmt.Errorf("failed to create KeyRegistrar: %w", err)
	}

	// Create operator set struct
	operatorSet := IKeyRegistrar.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}

	// Check if the operator is registered using the IsRegistered method
	isRegistered, err := keyRegistrar.IsRegistered(&bind.CallOpts{Context: ctx}, operatorSet, operatorAddress)
	if err != nil {
		return false, nil
	}

	return isRegistered, nil
}

// registerSingleOperatorKey registers a single operator's key by impersonating them
func registerSingleOperatorKey(
	ctx context.Context,
	logger *zap.Logger,
	contractCaller *caller.ContractCaller,
	client *ethclient.Client,
	op OperatorKeyInfo,
	gen IOperatorTableUpdater.OperatorSet,
	curveType config.CurveType,
) error {
	// First check if the operator is already registered
	if isRegistered, err := checkOperatorRegistered(ctx, client, op.OperatorAddress, gen.Avs, gen.Id); err != nil {
		logger.Sugar().Warnw("Failed to check if operator is registered",
			zap.String("operator", op.OperatorAddress.String()),
			zap.Error(err),
		)
		// Continue with registration attempt anyway
	} else if isRegistered {
		logger.Sugar().Infow("Operator already registered in KeyRegistrar, skipping",
			zap.String("operator", op.OperatorAddress.String()),
			zap.String("avs", gen.Avs.String()),
			zap.Uint32("operatorSetId", gen.Id),
		)
		return nil
	}

	if op.PrivateKeyHex == "" {
		return fmt.Errorf("operator must have private key for registration")
	}

	var keyData []byte
	var signature []byte
	var err error

	// Handle different curve types
	if curveType == config.CurveTypeBN254 {
		// Parse BN254 private key
		sk, err := bn254.NewPrivateKeyFromHexString(strings.TrimPrefix(op.PrivateKeyHex, "0x"))
		if err != nil {
			return fmt.Errorf("parse BN254 private key: %w", err)
		}

		pubKey := sk.Public()

		// Encode key data for registration
		keyData, err = contractCaller.EncodeBN254KeyData(pubKey)
		if err != nil {
			return fmt.Errorf("encode BN254 key data: %w", err)
		}

		// Get message hash and sign with operator's BN254 key
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
			return fmt.Errorf("BN254 sign: %w", err)
		}
		signature = sig.Bytes()

	} else if curveType == config.CurveTypeECDSA {
		// Parse ECDSA private key
		ecdsaPrivKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(strings.TrimPrefix(op.PrivateKeyHex, "0x"))
		if err != nil {
			return fmt.Errorf("parse ECDSA private key: %w", err)
		}

		// Derive the address from the private key
		derivedAddress, err := ecdsaPrivKey.DeriveAddress()
		if err != nil {
			return fmt.Errorf("derive ECDSA address: %w", err)
		}

		// For ECDSA, the key data is just the signing address (20 bytes)
		keyData = derivedAddress.Bytes()

		// Get message hash using the ECDSA-specific method
		msgHash, err := contractCaller.GetOperatorECDSAKeyRegistrationMessageHash(
			ctx,
			op.OperatorAddress,
			gen.Avs,
			gen.Id,
			derivedAddress,
		)
		if err != nil {
			return fmt.Errorf("failed to get ECDSA operator registration message: %w", err)
		}

		// Sign the message hash with ECDSA
		sig, err := ecdsaPrivKey.Sign(msgHash[:])
		if err != nil {
			return fmt.Errorf("ECDSA sign: %w", err)
		}
		// Convert ECDSA signature to bytes
		signature = sig.Bytes()

	} else {
		return fmt.Errorf("unsupported curve type: %s", curveType)
	}

	// Now use impersonation to register as the operator
	if err = registerKeyAsOperator(
		ctx,
		logger,
		contractCaller,
		client,
		op.OperatorAddress,
		gen.Avs,
		gen.Id,
		keyData,
		signature,
	); err != nil {
		return fmt.Errorf("failed to register key as operator: %w", err)
	}

	logger.Sugar().Infow("Successfully registered operator key",
		zap.String("operator", op.OperatorAddress.String()),
		zap.String("curveType", string(curveType)),
	)

	return nil
}

// registerKeyAsOperator registers a key by impersonating the operator (for testing)
// This uses anvil_impersonateAccount for local development
func registerKeyAsOperator(
	ctx context.Context,
	logger *zap.Logger,
	contractCaller *caller.ContractCaller,
	client *ethclient.Client,
	operatorAddress common.Address,
	avsAddress common.Address,
	operatorSetId uint32,
	keyData []byte,
	signature []byte,
) error {
	// Impersonate the operator account
	var ok bool
	if err := client.Client().CallContext(ctx, &ok, "anvil_impersonateAccount", operatorAddress.Hex()); err != nil {
		return fmt.Errorf("failed to impersonate operator %s: %w", operatorAddress.String(), err)
	}

	// Fund the account so it can pay gas
	_ = client.Client().CallContext(ctx, &ok, "anvil_setBalance", operatorAddress.Hex(), "0x56BC75E2D63100000") // 100 ETH

	logger.Sugar().Infow("Impersonating operator for key registration",
		"operator", operatorAddress.String(),
	)

	// Build transaction options with the impersonated operator as sender
	auth := &bind.TransactOpts{
		From:    operatorAddress,
		Context: ctx,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			// When impersonating with anvil, we don't need to sign
			// Anvil will handle it for the impersonated account
			return tx, nil
		},
		NoSend: false,
	}

	// Get the KeyRegistrar contract
	keyRegistrarAddress := common.HexToAddress("0xA4dB30D08d8bbcA00D40600bee9F029984dB162a")
	keyRegistrar, err := IKeyRegistrar.NewIKeyRegistrar(keyRegistrarAddress, client)
	if err != nil {
		_ = client.Client().CallContext(ctx, &ok, "anvil_stopImpersonatingAccount", operatorAddress.Hex())
		return fmt.Errorf("failed to create KeyRegistrar: %w", err)
	}

	// Create operator set struct
	operatorSet := IKeyRegistrar.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}

	logger.Sugar().Debugw("Registering key with KeyRegistrar as impersonated operator",
		"operatorAddress", operatorAddress.String(),
		"avsAddress", avsAddress.String(),
		"operatorSetId", operatorSetId,
		"keyData", hexutil.Encode(keyData),
		"signature", hexutil.Encode(signature),
	)

	// Register the key
	tx, err := keyRegistrar.RegisterKey(
		auth,
		operatorAddress,
		operatorSet,
		keyData,
		signature,
	)
	if err != nil {
		// Stop impersonation before returning error
		_ = client.Client().CallContext(ctx, &ok, "anvil_stopImpersonatingAccount", operatorAddress.Hex())
		return fmt.Errorf("failed to register key: %w", err)
	}

	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		_ = client.Client().CallContext(ctx, &ok, "anvil_stopImpersonatingAccount", operatorAddress.Hex())
		return fmt.Errorf("failed to wait for tx mining: %w", err)
	}

	if receipt.Status != 1 {
		_ = client.Client().CallContext(ctx, &ok, "anvil_stopImpersonatingAccount", operatorAddress.Hex())
		return fmt.Errorf("transaction reverted: %s", tx.Hash().Hex())
	}

	logger.Sugar().Infow("Successfully registered key for impersonated operator",
		"operator", operatorAddress.String(),
		"txHash", tx.Hash().Hex(),
	)

	// Stop impersonating the operator
	_ = client.Client().CallContext(ctx, &ok, "anvil_stopImpersonatingAccount", operatorAddress.Hex())

	return nil
}
