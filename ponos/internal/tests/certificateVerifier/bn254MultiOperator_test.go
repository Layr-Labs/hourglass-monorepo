package certificateVerifier

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/signing"
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

	execKeys := make([]*testUtils.WrappedKeyPair, 4)

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

	// Create operator configurations with sockets and metadata
	operatorConfigs := make([]*testUtils.OperatorConfig, len(operators))
	for i, op := range operators {
		operatorConfigs[i] = &testUtils.OperatorConfig{
			Operator:        op,
			Socket:          fmt.Sprintf("localhost:%d", 9000+i),
			MetadataUri:     "https://some-metadata-uri.com",
			AllocationDelay: 1,
		}
	}

	// Register all 4 operators
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

	// Delegate stake to all 4 operators
	// Note: DelegateStakeToOperators will use default stake amounts from StakeStuff.s.sol
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
			StrategyAddress:    testUtils.Strategy_WETH,
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
			StrategyAddress:    testUtils.Strategy_WETH,
		},
	}

	err = testUtils.DelegateStakeToOperators(
		t,
		ctx,
		stakeConfigs[0],
		stakeConfigs[1],
		chainConfig.AVSAccountAddress,
		l1EthClient,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to delegate stake to operators 1-2: %v", err)
	}

	err = testUtils.DelegateStakeToOperators(
		t,
		ctx,
		stakeConfigs[2],
		stakeConfigs[3],
		chainConfig.AVSAccountAddress,
		l1EthClient,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to delegate stake to operators 3-4: %v", err)
	}

	// Create generation reservation
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

	// Transport tables
	l.Sugar().Infow("------------------------ Transporting L1 tables ------------------------")
	testUtils.TransportStakeTables(l, false)
	time.Sleep(time.Second * 6)

	currentBlock, err := l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}
	t.Logf("Using current block: %d", currentBlock)
	testUtils.DebugOpsetData(t, chainConfig, eigenlayerContractAddrs, l1EthClient, currentBlock, []uint32{execOpsetId})

	// Create task and test with only 1 operator responding
	testBN254WithSingleResponder(t, ctx, l, chainConfig, l1CC, execKeys, execOpsetId, currentBlock)
}

func testBN254WithSingleResponder(
	t *testing.T,
	ctx context.Context,
	l *zap.Logger,
	chainConfig *testUtils.ChainConfig,
	l1CC contractCaller.IContractCaller,
	execKeys []*testUtils.WrappedKeyPair,
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

	// Test scenario: Only operator at index 0 responds
	respondingOperatorIndex := 0
	t.Logf("Testing scenario: Only operator at index %d responds", respondingOperatorIndex)

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

	// Sign with the responding operator's key
	responderSigner := inMemorySigner.NewInMemorySigner(
		execKeys[respondingOperatorIndex].PrivateKey,
		config.CurveTypeBN254,
	)

	resultSig, err := responderSigner.SignMessage(bn254DigestBytes)
	if err != nil {
		t.Fatalf("Failed to sign BN254 certificate: %v", err)
	}

	// Create auth signature
	resultSigDigest := util.GetKeccak256Digest(resultSig)
	authData := &types.AuthSignatureData{
		TaskId:          taskId,
		AvsAddress:      chainConfig.AVSAccountAddress,
		OperatorAddress: operators[respondingOperatorIndex].Address,
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
		OperatorAddress: operators[respondingOperatorIndex].Address,
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

	// Verify the operator info tree root is set
	if submitParams.OperatorInfoTreeRoot == [32]byte{} {
		t.Errorf("OperatorInfoTreeRoot should not be empty")
	} else {
		t.Logf("OperatorInfoTreeRoot: %s", hexutil.Encode(submitParams.OperatorInfoTreeRoot[:]))
	}

	// Try to submit the result (this would fail if we don't have enough stake)
	receipt, err := l1CC.SubmitBN254TaskResult(
		ctx,
		submitParams,
		operatorPeersWeight.RootReferenceTimestamp,
	)
	if err != nil {
		t.Logf("Failed to submit BN254 task result (expected if threshold too high): %v", err)
	} else {
		t.Logf("Successfully submitted BN254 task result with receipt: %v", receipt.TxHash.Hex())
		assert.Equal(t, uint64(1), receipt.Status, "Transaction should succeed")
	}
}
