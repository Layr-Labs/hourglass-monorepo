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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	l1RpcUrl              = "http://127.0.0.1:8545"
	numExecutorOperators  = 4
	executorOperatorSetId = 1
	maxStalenessPeriod    = 604800
	transportBlsKey       = "0x5f8e6420b9cb0c940e3d3f8b99177980785906d16fb3571f70d7a05ecf5f2172"
)

type thresholdTestCase struct {
	name                  string
	aggregationThreshold  uint16
	verificationThreshold uint16
	respondingOperatorIdx int
	shouldVerifySucceed   bool
}

func Test_BN254_MultiOperator_NonSigners(t *testing.T) {
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
	operatorKeyMap := make(map[string]*testUtils.WrappedKeyPair)

	for index := 0; index < numExecutorOperators; index++ {
		_, execKeysBN254, _, err := testUtils.GetKeysForCurveType(t, config.CurveTypeBN254, chainConfig)
		if err != nil {
			t.Fatalf("Failed to get BN254 keys for executor %d: %v", index+1, err)
		}
		execKeys[index] = execKeysBN254
	}

	l1EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   l1RpcUrl,
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

	t.Logf("Configuring operator set %d with curve type BN254 for 4 executors", executorOperatorSetId)
	_, err = avsConfigCaller.ConfigureAVSOperatorSet(ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		executorOperatorSetId,
		config.CurveTypeBN254,
	)

	if err != nil {
		t.Fatalf("Failed to configure executor operator set %d: %v", executorOperatorSetId, err)
	}

	contractAddresses := config.CoreContracts[config.ChainId_EthereumAnvil]

	operators := []*operator.Operator{
		{
			TransactionPrivateKey: chainConfig.ExecOperatorAccountPk,
			SigningPrivateKey:     execKeys[0].PrivateKey,
			Curve:                 config.CurveTypeBN254,
			OperatorSetIds:        []uint32{executorOperatorSetId},
		},
		{
			TransactionPrivateKey: chainConfig.ExecOperator2AccountPk,
			SigningPrivateKey:     execKeys[1].PrivateKey,
			Curve:                 config.CurveTypeBN254,
			OperatorSetIds:        []uint32{executorOperatorSetId},
		},
		{
			TransactionPrivateKey: chainConfig.ExecOperator3AccountPk,
			SigningPrivateKey:     execKeys[2].PrivateKey,
			Curve:                 config.CurveTypeBN254,
			OperatorSetIds:        []uint32{executorOperatorSetId},
		},
		{
			TransactionPrivateKey: chainConfig.ExecOperator4AccountPk,
			SigningPrivateKey:     execKeys[3].PrivateKey,
			Curve:                 config.CurveTypeBN254,
			OperatorSetIds:        []uint32{executorOperatorSetId},
		},
	}

	operatorKeyMap[strings.ToLower(chainConfig.ExecOperatorAccountAddress)] = execKeys[0]
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperator2AccountAddress)] = execKeys[1]
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperator3AccountAddress)] = execKeys[2]
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperator4AccountAddress)] = execKeys[3]

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
			OperatorSetId:      executorOperatorSetId,
			StrategyAddress:    testUtils.Strategy_STETH,
		},
		{
			StakerPrivateKey:   chainConfig.ExecStaker2AccountPrivateKey,
			StakerAddress:      chainConfig.ExecStaker2AccountAddress,
			OperatorPrivateKey: chainConfig.ExecOperator2AccountPk,
			OperatorAddress:    chainConfig.ExecOperator2AccountAddress,
			OperatorSetId:      executorOperatorSetId,
			StrategyAddress:    testUtils.Strategy_STETH,
		},
		{
			StakerPrivateKey:   chainConfig.ExecStaker3AccountPrivateKey,
			StakerAddress:      chainConfig.ExecStaker3AccountAddress,
			OperatorPrivateKey: chainConfig.ExecOperator3AccountPk,
			OperatorAddress:    chainConfig.ExecOperator3AccountAddress,
			OperatorSetId:      executorOperatorSetId,
			StrategyAddress:    testUtils.Strategy_STETH,
		},
		{
			StakerPrivateKey:   chainConfig.ExecStaker4AccountPrivateKey,
			StakerAddress:      chainConfig.ExecStaker4AccountAddress,
			OperatorPrivateKey: chainConfig.ExecOperator4AccountPk,
			OperatorAddress:    chainConfig.ExecOperator4AccountAddress,
			OperatorSetId:      executorOperatorSetId,
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

	bn254CalculatorAddr := avsConfigCaller.GetTableCalculatorAddress(config.CurveTypeBN254)
	t.Logf(
		"Creating generation reservation with BN254 table calculator %s for executor operator set %d",
		bn254CalculatorAddr.Hex(),
		executorOperatorSetId,
	)

	_, err = avsConfigCaller.CreateGenerationReservation(
		ctx,
		avsAddr,
		executorOperatorSetId,
		bn254CalculatorAddr,
		avsAddr,
		maxStalenessPeriod,
	)

	if err != nil {
		t.Logf("Warning: Failed to create generation reservation: %v", err)
	}

	time.Sleep(time.Second * 3)

	chainIdsToIgnore := []*big.Int{
		big.NewInt(11155111), // Sepolia
		big.NewInt(84532),    // Base Sepolia
		big.NewInt(31338),    // L2 anvil
	}

	blsInfos := make([]tableTransporter.OperatorBLSInfo, len(execKeys))
	operatorAddressList := []string{
		chainConfig.ExecOperatorAccountAddress,
		chainConfig.ExecOperator2AccountAddress,
		chainConfig.ExecOperator3AccountAddress,
		chainConfig.ExecOperator4AccountAddress,
	}

	// Stake weights: 2, 1.5, 1, 0.5 = 5 total
	// Operator 0: 40%, Operator 1: 30%, Operator 2: 20%, Operator 3: 10%
	stakeWeights := []*big.Int{
		big.NewInt(2000000000000000000),
		big.NewInt(1500000000000000000),
		big.NewInt(1000000000000000000),
		big.NewInt(500000000000000000),
	}

	for i, keyPair := range execKeys {
		blsPrivKey := keyPair.PrivateKey.(*bn254.PrivateKey)
		blsInfos[i] = tableTransporter.OperatorBLSInfo{
			PrivateKeyHex:   fmt.Sprintf("0x%x", blsPrivKey.Bytes()),
			Weights:         []*big.Int{stakeWeights[i]},
			OperatorAddress: common.HexToAddress(operatorAddressList[i]),
		}
	}

	cfg := &tableTransporter.MultipleOperatorConfig{
		TransporterPrivateKey:     chainConfig.AVSAccountPrivateKey,
		L1RpcUrl:                  "http://localhost:8545",
		L1ChainId:                 31337,
		L2RpcUrl:                  "",
		L2ChainId:                 0,
		CrossChainRegistryAddress: contractAddresses.CrossChainRegistry,
		ChainIdsToIgnore:          chainIdsToIgnore,
		Logger:                    l,
		Operators:                 blsInfos,
		AVSAddress:                common.HexToAddress(chainConfig.AVSAccountAddress),
		OperatorSetId:             executorOperatorSetId,
		TransportBLSPrivateKey:    transportBlsKey,
	}

	err = tableTransporter.TransportTableWithSimpleMultiOperators(cfg)
	require.NoError(t, err, "Failed to transport stake tables")

	time.Sleep(time.Second * 6)

	currentBlock, err := l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}

	testCases := []thresholdTestCase{
		{
			name:                  "Success_LowThreshold_SingleHighStakeOperator",
			aggregationThreshold:  1000, // 10%
			verificationThreshold: 1000, // 10%
			respondingOperatorIdx: 0,    // Operator with 40% stake
			shouldVerifySucceed:   true,
		},
		{
			name:                  "Success_MediumThreshold_SingleHighStakeOperator",
			aggregationThreshold:  2500, // 25%
			verificationThreshold: 2500, // 25%
			respondingOperatorIdx: 0,    // Operator with 40% stake
			shouldVerifySucceed:   true,
		},
		{
			name:                  "Failure_HighVerificationThreshold_SingleHighStakeOperator",
			aggregationThreshold:  1000, // 10% - aggregation succeeds
			verificationThreshold: 5000, // 50% - verification should fail
			respondingOperatorIdx: 0,    // Operator with 40% stake
			shouldVerifySucceed:   false,
		},
		{
			name:                  "Failure_HighThreshold_SingleLowStakeOperator",
			aggregationThreshold:  1000, // 10% - aggregation succeeds
			verificationThreshold: 2000, // 20% - verification should fail
			respondingOperatorIdx: 3,    // Operator with 10% stake
			shouldVerifySucceed:   false,
		},
		{
			name:                  "Success_ExactThreshold_SingleOperator",
			aggregationThreshold:  2000, // 20%
			verificationThreshold: 2000, // 20%
			respondingOperatorIdx: 2,    // Operator with exactly 20% stake
			shouldVerifySucceed:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testBN254WithThresholds(
				t,
				ctx,
				l,
				chainConfig,
				l1CC,
				operatorKeyMap,
				executorOperatorSetId,
				currentBlock,
				tc,
			)
		})
	}
}

func testBN254WithThresholds(
	t *testing.T,
	ctx context.Context,
	l *zap.Logger,
	chainConfig *testUtils.ChainConfig,
	l1CC contractCaller.IContractCaller,
	operatorKeyMap map[string]*testUtils.WrappedKeyPair,
	executorOperatorSetId uint32,
	currentBlock uint64,
	tc thresholdTestCase,
) {
	t.Logf("=== Testing: %s ===", tc.name)
	t.Logf("Aggregation threshold: %d/10000 (%.1f%%)", tc.aggregationThreshold, float64(tc.aggregationThreshold)/100)
	t.Logf("Verification threshold: %d/10000 (%.1f%%)", tc.verificationThreshold, float64(tc.verificationThreshold)/100)
	t.Logf("Responding operator index: %d", tc.respondingOperatorIdx)
	t.Logf("Expected verification result: %v", tc.shouldVerifySucceed)

	taskId := fmt.Sprintf("0x%064x", time.Now().UnixNano()) // Unique task ID for each test
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
		executorOperatorSetId,
	)
	if err != nil {
		t.Fatalf("Failed to get operator peers and weights: %v", err)
	}

	// Create BN254 aggregator
	var operators []*aggregation.Operator[signing.PublicKey]
	operatorAddressList := []string{
		chainConfig.ExecOperatorAccountAddress,
		chainConfig.ExecOperator2AccountAddress,
		chainConfig.ExecOperator3AccountAddress,
		chainConfig.ExecOperator4AccountAddress,
	}

	for _, peer := range operatorPeersWeight.Operators {
		opset, err := peer.GetOperatorSet(executorOperatorSetId)
		if err != nil {
			t.Fatalf("Failed to get operator set %d for peer %s: %v", executorOperatorSetId, peer.OperatorAddress, err)
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
	totalWeight := big.NewInt(0)
	for i, op := range operators {
		weight := op.Weights[0]
		totalWeight.Add(totalWeight, weight)
		t.Logf("Operator %d: Address=%s, Index=%d, Weight=%s",
			i, op.Address, op.OperatorIndex, weight.String())
	}
	t.Logf("Total weight: %s", totalWeight.String())

	agg, err := aggregation.NewBN254TaskResultAggregator(
		context.Background(),
		taskId,
		operatorPeersWeight.RootReferenceTimestamp,
		executorOperatorSetId,
		tc.aggregationThreshold, // Use test case aggregation threshold
		l1CC,
		taskInputData,
		&deadline,
		operators,
	)
	if err != nil {
		t.Fatalf("Failed to create BN254 task result aggregator: %v", err)
	}

	// Select the responding operator based on test case
	respondingOperatorAddress := operatorAddressList[tc.respondingOperatorIdx]
	var respondingOperator *aggregation.Operator[signing.PublicKey]
	for _, op := range operators {
		if strings.EqualFold(op.Address, respondingOperatorAddress) {
			respondingOperator = op
			break
		}
	}
	if respondingOperator == nil {
		t.Fatalf("Could not find operator at index %d", tc.respondingOperatorIdx)
	}

	operatorWeight := respondingOperator.Weights[0]
	operatorPercentage := new(big.Float).Quo(
		new(big.Float).SetInt(operatorWeight),
		new(big.Float).SetInt(totalWeight),
	)
	operatorPercentage.Mul(operatorPercentage, big.NewFloat(100))
	percentFloat, _ := operatorPercentage.Float64()

	t.Logf("Responding operator: %s (index %d) with weight %s (%.1f%% of total)",
		respondingOperator.Address, respondingOperator.OperatorIndex, operatorWeight.String(), percentFloat)

	operatorKeys, ok := operatorKeyMap[strings.ToLower(respondingOperator.Address)]
	if !ok {
		t.Fatalf("Could not find BN254 keys for operator %s", respondingOperator.Address)
	}
	responderPrivateKey := operatorKeys.PrivateKey

	// Create signature from the responding operator
	var taskIdBytes [32]byte
	copy(taskIdBytes[:], common.HexToHash(taskId).Bytes())
	messageHash, err := l1CC.CalculateTaskMessageHash(ctx, taskIdBytes, taskInputData)
	if err != nil {
		t.Fatalf("Failed to calculate task message hash: %v", err)
	}
	bn254DigestBytes, err := l1CC.CalculateBN254CertificateDigestBytes(
		ctx,
		operatorPeersWeight.RootReferenceTimestamp,
		messageHash,
	)
	if err != nil {
		t.Fatalf("Failed to calculate BN254 certificate digest: %v", err)
	}

	responderSigner := inMemorySigner.NewInMemorySigner(responderPrivateKey, config.CurveTypeBN254)

	resultSig, err := responderSigner.SignMessageForSolidity(bn254DigestBytes)
	if err != nil {
		t.Fatalf("Failed to sign BN254 certificate: %v", err)
	}

	resultSigDigest := util.GetKeccak256Digest(resultSig)
	authData := &types.AuthSignatureData{
		TaskId:          taskId,
		AvsAddress:      chainConfig.AVSAccountAddress,
		OperatorAddress: respondingOperator.Address,
		OperatorSetId:   executorOperatorSetId,
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
		OperatorSetId:   executorOperatorSetId,
		Output:          taskInputData,
		OperatorAddress: respondingOperator.Address,
		ResultSignature: resultSig,
		AuthSignature:   authSig,
	}

	// Process the signature
	if err := agg.ProcessNewSignature(ctx, taskResult); err != nil {
		t.Fatalf("Failed to process new signature: %v", err)
	}

	// Check if threshold is met for aggregation
	if !agg.SigningThresholdMet() {
		t.Logf("Aggregation threshold not met with operator weight %.1f%% < %.1f%% threshold",
			percentFloat, float64(tc.aggregationThreshold)/100)
		// If aggregation threshold is not met, we can't generate a certificate to verify
		return
	}

	t.Logf("Aggregation threshold met: operator weight %.1f%% >= %.1f%% threshold",
		percentFloat, float64(tc.aggregationThreshold)/100)

	finalCert, err := agg.GenerateFinalCertificate()
	if err != nil {
		t.Fatalf("Failed to generate final certificate: %v", err)
	}

	t.Logf("Final certificate generated with %d non-signers", len(operators)-1)

	// The certificate should include merkle proofs for the non-signing operators
	submitParams := finalCert.ToSubmitParams()
	t.Logf("Non-signer count: %d", len(submitParams.NonSignerOperators))

	// Now verify with the test case's verification threshold
	verified, err := l1CC.VerifyBN254Certificate(
		ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		executorOperatorSetId,
		submitParams,
		operatorPeersWeight.OperatorInfos,
		operatorPeersWeight.RootReferenceTimestamp,
		operatorPeersWeight.OperatorInfoTreeRoot,
		tc.verificationThreshold,
	)

	if err != nil {
		if tc.shouldVerifySucceed {
			t.Errorf("Expected verification to succeed but got error: %v", err)
		} else {
			t.Logf("Verification failed as expected with error: %v", err)
		}
	} else {
		t.Logf("BN254 certificate verification result: %v (threshold: %d/10000 = %.1f%%)",
			verified, tc.verificationThreshold, float64(tc.verificationThreshold)/100)

		if verified && !tc.shouldVerifySucceed {
			t.Errorf("Expected verification to fail but it succeeded")
		} else if !verified && tc.shouldVerifySucceed {
			t.Errorf("Expected verification to succeed but it failed")
		} else if verified && tc.shouldVerifySucceed {
			t.Logf("✓ Test passed: Verification succeeded as expected")
		} else {
			t.Logf("✓ Test passed: Verification failed as expected")
		}
	}
}
