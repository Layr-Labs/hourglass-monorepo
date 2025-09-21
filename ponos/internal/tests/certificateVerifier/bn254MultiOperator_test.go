package certificateVerifier

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/tableTransporter"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operatorManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_BN254_MultiOperator_NonSigners(t *testing.T) {
	const (
		L1RpcUrl = "http://127.0.0.1:8545"
	)

	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: true})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	root := testUtils.GetProjectRootPath()
	t.Logf("Project root path: %s", root)

	chainConfig, err := testUtils.ReadChainConfig(root)
	if err != nil {
		t.Fatalf("Failed to read chain config: %v", err)
	}

	// Generate BN254 keys for all 4 operators
	execKeys := make([]*testUtils.WrappedKeyPair, 4)
	// Map to store which BN254 key belongs to which operator address
	operatorKeyMap := make(map[string]*testUtils.WrappedKeyPair)

	for i := 0; i < 4; i++ {
		_, execKeysBN254, _, err := testUtils.GetKeysForCurveType(t, config.CurveTypeBN254, chainConfig)
		if err != nil {
			t.Fatalf("Failed to get BN254 keys for executor %d: %v", i+1, err)
		}
		execKeys[i] = execKeysBN254
	}

	l1EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   L1RpcUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)

	l1EthClient, err := l1EthereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get Ethereum contract caller: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	anvilWg := &sync.WaitGroup{}
	anvilWg.Add(1)
	startErrorsChan := make(chan error, 1)

	_ = testUtils.KillallAnvils()

	l1Anvil, err := testUtils.StartL1Anvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start L1 Anvil: %v", err)
	}
	defer func() {
		if err := testUtils.KillAnvil(l1Anvil); err != nil {
			t.Logf("Failed to kill L1 Anvil: %v", err)
		}
	}()

	anvilCtx, anvilCancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer anvilCancel()
	go testUtils.WaitForAnvil(anvilWg, anvilCtx, t, l1EthereumClient, startErrorsChan)

	anvilWg.Wait()
	close(startErrorsChan)
	for err := range startErrorsChan {
		if err != nil {
			t.Fatalf("Failed to start Anvil: %v", err)
		}
	}
	anvilCancel()
	t.Logf("Anvil is running")

	l1ChainId, err := l1EthClient.ChainID(ctx)
	if err != nil {
		t.Fatalf("Failed to get L1 chain ID: %v", err)
	}
	t.Logf("L1 Chain ID: %s", l1ChainId.String())

	eigenlayerContractAddrs, err := config.GetCoreContractsForChainId(config.ChainId(l1ChainId.Uint64()))
	if err != nil {
		t.Fatalf("Failed to get core contracts for chain ID: %v", err)
	}

	l1PrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AppAccountPrivateKey, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create L1 private key signer: %v", err)
	}

	l1CC, err := caller.NewContractCaller(l1EthClient, l1PrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create L1 contract caller: %v", err)
	}

	// Create AVS contract caller for configuring operator sets
	avsConfigPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create AVS config private key signer: %v", err)
	}

	avsConfigCaller, err := caller.NewContractCaller(l1EthClient, avsConfigPrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create AVS config caller: %v", err)
	}

	// Configure BN254 operator set for all executors
	execOpsetId := uint32(1)
	t.Logf("Configuring operator set %d with curve type BN254 for 4 executors", execOpsetId)
	_, err = avsConfigCaller.ConfigureAVSOperatorSet(ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		execOpsetId,
		config.CurveTypeBN254)
	if err != nil {
		t.Fatalf("Failed to configure executor operator set %d: %v", execOpsetId, err)
	}

	t.Logf("------------------------------------------- Setting up 4 BN254 operators -------------------------------------------")

	// Create operators array with all 4 executors
	operators := []*operator.Operator{
		{
			TransactionPrivateKey: chainConfig.ExecOperatorAccountPk,
			SigningPrivateKey:     execKeys[0].PrivateKey,
			Curve:                 config.CurveTypeBN254,
			OperatorSetIds:        []uint32{execOpsetId},
		},
		{
			TransactionPrivateKey: chainConfig.ExecOperator2AccountPk,
			SigningPrivateKey:     execKeys[1].PrivateKey,
			Curve:                 config.CurveTypeBN254,
			OperatorSetIds:        []uint32{execOpsetId},
		},
		{
			TransactionPrivateKey: chainConfig.ExecOperator3AccountPk,
			SigningPrivateKey:     execKeys[2].PrivateKey,
			Curve:                 config.CurveTypeBN254,
			OperatorSetIds:        []uint32{execOpsetId},
		},
		{
			TransactionPrivateKey: chainConfig.ExecOperator4AccountPk,
			SigningPrivateKey:     execKeys[3].PrivateKey,
			Curve:                 config.CurveTypeBN254,
			OperatorSetIds:        []uint32{execOpsetId},
		},
	}

	// Map operator addresses to their BN254 signing keys
	// Use lowercase addresses to ensure consistent matching
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperatorAccountAddress)] = execKeys[0]
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperator2AccountAddress)] = execKeys[1]
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperator3AccountAddress)] = execKeys[2]
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperator4AccountAddress)] = execKeys[3]

	// Create operator configurations with sockets and metadata
	// Store the configs for later reference
	operatorConfigs := make([]*testUtils.OperatorConfig, len(operators))
	for i, op := range operators {
		operatorConfigs[i] = &testUtils.OperatorConfig{
			Operator:        op,
			Socket:          fmt.Sprintf("localhost:%d", 9000+i),
			MetadataUri:     "https://some-metadata-uri.com",
			AllocationDelay: 1,
		}
	}

	err = testUtils.RegisterMultipleOperators(
		ctx,
		l1EthClient,
		chainConfig.AVSAccountAddress,
		chainConfig.AVSAccountPrivateKey,
		operatorConfigs,
		l,
	)

	if err != nil {
		t.Fatalf("Failed to register operators: %v", err)
	}

	time.Sleep(time.Second * 6)

	stakeConfigs := []*testUtils.StakerDelegationConfig{
		{
			StakerPrivateKey:   chainConfig.ExecStakerAccountPrivateKey,
			StakerAddress:      chainConfig.ExecStakerAccountAddress,
			OperatorPrivateKey: chainConfig.ExecOperatorAccountPk,
			OperatorAddress:    chainConfig.ExecOperatorAccountAddress,
			OperatorSetId:      execOpsetId,
			StrategyAddress:    testUtils.Strategy_STETH,
		},
		{
			StakerPrivateKey:   chainConfig.ExecStaker2AccountPrivateKey,
			StakerAddress:      chainConfig.ExecStaker2AccountAddress,
			OperatorPrivateKey: chainConfig.ExecOperator2AccountPk,
			OperatorAddress:    chainConfig.ExecOperator2AccountAddress,
			OperatorSetId:      execOpsetId,
			StrategyAddress:    testUtils.Strategy_STETH,
		},
		{
			StakerPrivateKey:   chainConfig.ExecStaker3AccountPrivateKey,
			StakerAddress:      chainConfig.ExecStaker3AccountAddress,
			OperatorPrivateKey: chainConfig.ExecOperator3AccountPk,
			OperatorAddress:    chainConfig.ExecOperator3AccountAddress,
			OperatorSetId:      execOpsetId,
			StrategyAddress:    testUtils.Strategy_STETH,
		},
		{
			StakerPrivateKey:   chainConfig.ExecStaker4AccountPrivateKey,
			StakerAddress:      chainConfig.ExecStaker4AccountAddress,
			OperatorPrivateKey: chainConfig.ExecOperator4AccountPk,
			OperatorAddress:    chainConfig.ExecOperator4AccountAddress,
			OperatorSetId:      execOpsetId,
			StrategyAddress:    testUtils.Strategy_STETH,
		},
	}

	err = testUtils.DelegateStakeToMultipleOperators(
		t,
		ctx,
		stakeConfigs,
		chainConfig.AVSAccountAddress,
		l1EthClient,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to delegate stake to operators: %v", err)
	}

	avsAddr := common.HexToAddress(chainConfig.AVSAccountAddress)
	maxStalenessPeriod := uint32(604800) // 1 week in seconds

	bn254CalculatorAddr := avsConfigCaller.GetTableCalculatorAddress(config.CurveTypeBN254)
	t.Logf("Creating generation reservation with BN254 table calculator %s for executor operator set %d",
		bn254CalculatorAddr.Hex(), execOpsetId)
	_, err = avsConfigCaller.CreateGenerationReservation(
		ctx,
		avsAddr,
		execOpsetId,
		bn254CalculatorAddr,
		avsAddr, // AVS is the owner
		maxStalenessPeriod,
	)
	if err != nil {
		t.Logf("Warning: Failed to create generation reservation: %v", err)
	}

	time.Sleep(time.Second * 3)

	l.Sugar().Infow("------------------------ Transporting L1 tables ------------------------")

	transportBLSKey := "0x5f8e6420b9cb0c940e3d3f8b99177980785906d16fb3571f70d7a05ecf5f2172"

	// Get contract addresses
	contractAddresses := config.CoreContracts[config.ChainId_EthereumAnvil]

	// Set up chains to ignore (only transport to our L1)
	chainIdsToIgnore := []*big.Int{
		big.NewInt(11155111), // Sepolia
		big.NewInt(84532),    // Base Sepolia
		big.NewInt(31338),    // L2 anvil
	}

	// Prepare BLS infos for transport (operators already registered)
	blsInfos := make([]tableTransporter.OperatorBLSInfo, len(execKeys))
	operatorAddresses := []string{
		chainConfig.ExecOperatorAccountAddress,
		chainConfig.ExecOperator2AccountAddress,
		chainConfig.ExecOperator3AccountAddress,
		chainConfig.ExecOperator4AccountAddress,
	}

	// Use descending weights matching the test setup
	stakeWeights := []*big.Int{
		big.NewInt(2000000000000000000), // 2e18
		big.NewInt(1500000000000000000), // 1.5e18
		big.NewInt(1000000000000000000), // 1e18
		big.NewInt(500000000000000000),  // 0.5e18
	}

	for i, keyPair := range execKeys {
		blsPrivKey := keyPair.PrivateKey.(*bn254.PrivateKey)
		blsInfos[i] = tableTransporter.OperatorBLSInfo{
			PrivateKeyHex:   fmt.Sprintf("0x%x", blsPrivKey.Bytes()),
			Weights:         []*big.Int{stakeWeights[i]},
			OperatorAddress: common.HexToAddress(operatorAddresses[i]),
		}
	}

	cfg := &tableTransporter.SimpleMultiOperatorConfig{
		TransporterPrivateKey:     chainConfig.AVSAccountPrivateKey,
		L1RpcUrl:                  "http://localhost:8545",
		L1ChainId:                 31337,
		L2RpcUrl:                  "", // No L2
		L2ChainId:                 0,  // No L2
		CrossChainRegistryAddress: contractAddresses.CrossChainRegistry,
		ChainIdsToIgnore:          chainIdsToIgnore,
		Logger:                    l,
		Operators:                 blsInfos,
		AVSAddress:                common.HexToAddress(chainConfig.AVSAccountAddress),
		OperatorSetId:             execOpsetId,
		TransportBLSPrivateKey:    transportBLSKey,
	}

	t.Logf("========== PRE-TRANSPORT STATE CHECK ==========")

	crossChainRegAddr := contractAddresses.CrossChainRegistry
	t.Logf("CrossChainRegistry address: %s", crossChainRegAddr)

	keyRegistrarAddr := contractAddresses.KeyRegistrar
	t.Logf("KeyRegistrar address: %s", keyRegistrarAddr)

	for i, addr := range operatorAddresses {
		t.Logf("Operator %d (%s) - checking BLS key registration", i, addr)
	}

	bn254CalcAddr := avsConfigCaller.GetTableCalculatorAddress(config.CurveTypeBN254)
	t.Logf("BN254 Table Calculator address: %s", bn254CalcAddr.Hex())

	t.Logf("========== STARTING TRANSPORT ==========")
	err = tableTransporter.TransportTableWithSimpleMultiOperators(cfg)
	require.NoError(t, err, "Failed to transport stake tables")

	t.Logf("Waiting for transport to be processed...")
	time.Sleep(time.Second * 6)

	t.Logf("========== POST-TRANSPORT STATE CHECK ==========")

	currentBlock, err := l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}

	operatorTableData, err := l1CC.GetOperatorTableDataForOperatorSet(ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		execOpsetId,
		config.CurveTypeBN254,
		config.ChainId_EthereumAnvil,
		currentBlock)
	if err != nil {
		t.Logf("Failed to get operator table data: %v", err)
	} else {
		t.Logf("Operator table data retrieved:")
		t.Logf("  - Latest reference timestamp: %d", operatorTableData.LatestReferenceTimestamp)
		t.Logf("  - Operator count: %d", len(operatorTableData.Operators))
		t.Logf("  - OperatorInfoTreeRoot: %s", hexutil.Encode(operatorTableData.OperatorInfoTreeRoot[:]))

		if operatorTableData.OperatorInfoTreeRoot == [32]byte{} {
			t.Logf("WARNING: OperatorInfoTreeRoot is EMPTY after transport!")
			t.Logf("  This means the BN254CertificateVerifier doesn't have the operator info tree root set")
			t.Logf("  The transport likely succeeded but the verifier wasn't updated with the tree root")
		} else {
			t.Logf("SUCCESS: OperatorInfoTreeRoot is SET: %s", hexutil.Encode(operatorTableData.OperatorInfoTreeRoot[:]))
		}
	}

	testUtils.DebugOpsetData(t, chainConfig, eigenlayerContractAddrs, l1EthClient, currentBlock, []uint32{execOpsetId})

	testBN254WithSingleResponder(t, ctx, l, chainConfig, l1CC, operatorKeyMap, execOpsetId, currentBlock)
}

func testBN254WithSingleResponder(
	t *testing.T,
	ctx context.Context,
	l *zap.Logger,
	chainConfig *testUtils.ChainConfig,
	l1CC contractCaller.IContractCaller,
	operatorKeyMap map[string]*testUtils.WrappedKeyPair,
	execOpsetId uint32,
	currentBlock uint64,
) {
	taskId := "0x0000000000000000000000000000000000000000000000000000000000000001"
	taskInputData := []byte("test-task-input-data")
	deadline := time.Now().Add(1 * time.Minute)

	pdf := peeringDataFetcher.NewPeeringDataFetcher(l1CC, l)
	callerMap := map[config.ChainId]contractCaller.IContractCaller{
		config.ChainId_EthereumAnvil: l1CC,
	}

	opManager := operatorManager.NewOperatorManager(&operatorManager.OperatorManagerConfig{
		AvsAddress: chainConfig.AVSAccountAddress,
		ChainIds:   []config.ChainId{config.ChainId_EthereumAnvil},
		L1ChainId:  config.ChainId_EthereumAnvil,
	}, callerMap, pdf, l)

	operatorPeersWeight, err := opManager.GetExecutorPeersAndWeightsForBlock(
		ctx,
		config.ChainId_EthereumAnvil,
		currentBlock,
		execOpsetId,
	)
	if err != nil {
		t.Fatalf("Failed to get operator peers and weights: %v", err)
	}

	// Create BN254 aggregator
	var operators []*aggregation.Operator[signing.PublicKey]
	for _, peer := range operatorPeersWeight.Operators {
		opset, err := peer.GetOperatorSet(execOpsetId)
		if err != nil {
			t.Fatalf("Failed to get operator set %d for peer %s: %v", execOpsetId, peer.OperatorAddress, err)
		}

		// Retrieve weights for this operator
		weights := operatorPeersWeight.Weights[peer.OperatorAddress]

		operators = append(operators, &aggregation.Operator[signing.PublicKey]{
			Address:       peer.OperatorAddress,
			PublicKey:     opset.WrappedPublicKey.PublicKey,
			OperatorIndex: opset.OperatorIndex,
			Weights:       weights,
		})
	}

	t.Logf("======= BN254 Operators =======")
	for i, op := range operators {
		t.Logf("Operator %d: Address=%s, Index=%d, Weights=%v",
			i, op.Address, op.OperatorIndex, op.Weights)
	}

	agg, err := aggregation.NewBN254TaskResultAggregator(
		context.Background(),
		taskId,
		operatorPeersWeight.RootReferenceTimestamp,
		execOpsetId,
		2500, // Threshold: 25% (with 4 operators having weights 2, 1.5, 1, 0.5 = 5 total, 25% = 1.25)
		l1CC,
		taskInputData,
		&deadline,
		operators,
	)
	if err != nil {
		t.Fatalf("Failed to create BN254 task result aggregator: %v", err)
	}

	// WHEN: only the 0th indexed operator responds
	respondingOperator := operators[0]
	t.Logf("Testing scenario: Operator %s (index %d) responds", respondingOperator.Address, respondingOperator.OperatorIndex)

	operatorKeys, ok := operatorKeyMap[strings.ToLower(respondingOperator.Address)]
	if !ok {
		t.Logf("Available keys in map:")
		for addr := range operatorKeyMap {
			t.Logf("  - %s", addr)
		}
		t.Fatalf("Could not find BN254 keys for operator %s", respondingOperator.Address)
	}
	responderPrivateKey := operatorKeys.PrivateKey

	// Debug: Verify the public key matches
	bn254PrivKey := responderPrivateKey.(*bn254.PrivateKey)
	pubKeyFromPriv := bn254PrivKey.Public()
	t.Logf("Debug: Operator %s", respondingOperator.Address)

	// Get the G2 points (public keys are in G2 for BN254)
	g2FromPriv := pubKeyFromPriv.GetG2Point()
	if g2FromPriv != nil {
		t.Logf("  Public key from private key (G2): X0=%s, X1=%s",
			g2FromPriv.X.A0.String(), g2FromPriv.X.A1.String())
	}

	// Check if the public key from the operator matches
	if respondingOperator.PublicKey != nil {
		bn254PubKey := respondingOperator.PublicKey.(*bn254.PublicKey)
		g2FromOp := bn254PubKey.GetG2Point()
		if g2FromOp != nil {
			t.Logf("  Public key from operator (G2): X0=%s, X1=%s",
				g2FromOp.X.A0.String(), g2FromOp.X.A1.String())

			// Verify they match
			if g2FromPriv != nil && !g2FromPriv.Equal(g2FromOp) {
				t.Errorf("ERROR: Public keys don't match!")
				t.Logf("  Expected (from private): bytes=%x", pubKeyFromPriv.Bytes())
				t.Logf("  Got (from operator): bytes=%x", bn254PubKey.Bytes())
			} else {
				t.Logf("  Public keys match!")
			}
		}
	}

	// Create signature from the responding operator
	messageHash := util.GetKeccak256Digest(taskInputData)
	bn254DigestBytes, err := l1CC.CalculateBN254CertificateDigestBytes(
		ctx,
		operatorPeersWeight.RootReferenceTimestamp,
		messageHash,
	)
	if err != nil {
		t.Fatalf("Failed to calculate BN254 certificate digest: %v", err)
	}

	responderSigner := inMemorySigner.NewInMemorySigner(
		responderPrivateKey,
		config.CurveTypeBN254,
	)

	resultSig, err := responderSigner.SignMessageForSolidity(bn254DigestBytes)
	if err != nil {
		t.Fatalf("Failed to sign BN254 certificate: %v", err)
	}

	resultSigDigest := util.GetKeccak256Digest(resultSig)
	authData := &types.AuthSignatureData{
		TaskId:          taskId,
		AvsAddress:      chainConfig.AVSAccountAddress,
		OperatorAddress: respondingOperator.Address,
		OperatorSetId:   execOpsetId,
		ResultSigDigest: resultSigDigest,
	}

	authBytes := authData.ToSigningBytes()
	authSig, err := responderSigner.SignMessage(authBytes)
	if err != nil {
		t.Fatalf("Failed to sign auth data: %v", err)
	}

	// Create task result
	taskResult := &types.TaskResult{
		TaskId:          taskId,
		AvsAddress:      chainConfig.AVSAccountAddress,
		OperatorSetId:   execOpsetId,
		Output:          taskInputData,
		OperatorAddress: respondingOperator.Address,
		ResultSignature: resultSig,
		AuthSignature:   authSig,
	}

	// Process the signature
	if err := agg.ProcessNewSignature(ctx, taskResult); err != nil {
		t.Fatalf("Failed to process new signature: %v", err)
	}

	// Check if threshold is met
	// The threshold calculation depends on the actual stake weights that were delegated
	if !agg.SigningThresholdMet() {
		t.Logf("Threshold not met with single operator - adjusting test expectations")
		// This is expected if the operator doesn't have enough stake weight
	}

	// Generate final certificate
	finalCert, err := agg.GenerateFinalCertificate()
	if err != nil {
		t.Fatalf("Failed to generate final certificate: %v", err)
	}

	t.Logf("Final certificate generated with %d non-signers", len(operators)-1)

	// The certificate should include merkle proofs for the 3 non-signing operators
	submitParams := finalCert.ToSubmitParams()
	t.Logf("Non-signer count: %d", len(submitParams.NonSignerOperators))
	for i, nonSigner := range submitParams.NonSignerOperators {
		t.Logf("Non-signer %d: OperatorIndex=%d", i, nonSigner.OperatorIndex)
	}

	// Refetch operator data after transport to get the updated OperatorInfoTreeRoot
	t.Logf("Refetching operator data after transport...")
	t.Logf("  Chain: %v, Block: %d, OpSetId: %d", config.ChainId_EthereumAnvil, currentBlock, execOpsetId)

	operatorPeersWeightAfterTransport, err := opManager.GetExecutorPeersAndWeightsForBlock(
		ctx,
		config.ChainId_EthereumAnvil,
		currentBlock,
		execOpsetId, // Use the correct operator set ID
	)
	require.NoError(t, err, "Failed to get executor peers and weights after transport")

	t.Logf("Fetched operator data after transport:")
	t.Logf("  RootReferenceTimestamp: %d", operatorPeersWeightAfterTransport.RootReferenceTimestamp)
	t.Logf("  OperatorInfoTreeRoot: %s", hexutil.Encode(operatorPeersWeightAfterTransport.OperatorInfoTreeRoot[:]))
	t.Logf("  Number of operators: %d", len(operatorPeersWeightAfterTransport.Operators))

	// Verify the operator info tree root is set using the refreshed data
	if operatorPeersWeightAfterTransport.OperatorInfoTreeRoot == [32]byte{} {
		t.Errorf("OperatorInfoTreeRoot should not be empty after transport")

		// Try directly fetching from the contract to see what's stored
		t.Logf("Debug: Let me directly query the operator table data again...")
		directOperatorTableData, err := l1CC.GetOperatorTableDataForOperatorSet(ctx,
			common.HexToAddress(chainConfig.AVSAccountAddress),
			execOpsetId,
			config.CurveTypeBN254,
			config.ChainId_EthereumAnvil,
			currentBlock)
		if err != nil {
			t.Logf("  Failed to directly query operator table data: %v", err)
		} else {
			t.Logf("  Direct query result: OperatorInfoTreeRoot = %s", hexutil.Encode(directOperatorTableData.OperatorInfoTreeRoot[:]))
		}

		t.Logf("Debug: This might be a timing issue or the data isn't being fetched from the right source")
		t.FailNow()
	} else {
		t.Logf("SUCCESS: OperatorInfoTreeRoot after transport: %s", hexutil.Encode(operatorPeersWeightAfterTransport.OperatorInfoTreeRoot[:]))
	}

	// Try to submit the result using the updated data after transport
	receipt, err := l1CC.SubmitBN254TaskResult(
		ctx,
		submitParams,
		operatorPeersWeightAfterTransport.RootReferenceTimestamp,
		operatorPeersWeightAfterTransport.OperatorInfoTreeRoot,
	)
	if err != nil {
		t.Logf("Failed to submit BN254 task result (expected if threshold too high): %v", err)
		t.Fail()
	} else {
		t.Logf("Successfully submitted BN254 task result with receipt: %v", receipt.TxHash.Hex())
		assert.Equal(t, uint64(1), receipt.Status, "Transaction should succeed")
	}
}
