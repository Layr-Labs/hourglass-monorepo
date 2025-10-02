package aggregation

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
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
	"go.uber.org/zap"
)

func Test_ECDSA_MultiOperator_Thresholds(t *testing.T) {

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

	operatorKeyMap := make(map[string]*testUtils.WrappedKeyPair)

	aggKey, execKeys, err := testUtils.GetKeysForCurveTypeFromChainConfig(
		t,
		config.CurveTypeECDSA,
		config.CurveTypeECDSA,
		chainConfig,
	)
	if err != nil {
		t.Fatalf("Failed to get keys: %v", err)
	}

	execKeys = execKeys[:numExecutorOperators]

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

	avsConfigPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create AVS config private key signer: %v", err)
	}

	avsConfigCaller, err := caller.NewContractCaller(l1EthClient, avsConfigPrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create AVS config caller: %v", err)
	}

	t.Logf("Configuring operator set %d with curve type ECDSA for 4 executors", executorOperatorSetId)
	_, err = avsConfigCaller.ConfigureAVSOperatorSet(ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		executorOperatorSetId,
		config.CurveTypeECDSA,
	)

	if err != nil {
		t.Fatalf("Failed to configure executor operator set %d: %v", executorOperatorSetId, err)
	}

	_, err = avsConfigCaller.ConfigureAVSOperatorSet(
		ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		aggregatorOperatorSetId,
		config.CurveTypeECDSA,
	)
	if err != nil {
		t.Fatalf("Failed to configure aggregator operator set %d: %v", aggregatorOperatorSetId, err)
	}

	operatorPkList := []string{
		chainConfig.ExecOperatorAccountPk,
		chainConfig.ExecOperator2AccountPk,
		chainConfig.ExecOperator3AccountPk,
		chainConfig.ExecOperator4AccountPk,
	}

	operatorKeyMap[strings.ToLower(chainConfig.ExecOperatorAccountAddress)] = execKeys[0]
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperator2AccountAddress)] = execKeys[1]
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperator3AccountAddress)] = execKeys[2]
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperator4AccountAddress)] = execKeys[3]

	executorsWithSockets := make([]testUtils.ExecutorWithSocket, numExecutorOperators)
	for i := 0; i < numExecutorOperators; i++ {
		executorsWithSockets[i] = testUtils.ExecutorWithSocket{
			Executor: &operator.Operator{
				TransactionPrivateKey: operatorPkList[i],
				SigningPrivateKey:     execKeys[i].PrivateKey,
				Curve:                 config.CurveTypeECDSA,
				OperatorSetIds:        []uint32{executorOperatorSetId},
			},
			Socket: fmt.Sprintf("localhost:%d", 9000+i),
		}
	}

	aggregatorOperator := &operator.Operator{
		TransactionPrivateKey: chainConfig.OperatorAccountPrivateKey,
		SigningPrivateKey:     aggKey.PrivateKey,
		Curve:                 config.CurveTypeECDSA,
		OperatorSetIds:        []uint32{aggregatorOperatorSetId},
	}

	t.Logf("Setting up operator peering with %d executors", numExecutorOperators)
	err = testUtils.SetupOperatorPeeringWithMultipleExecutors(
		ctx,
		chainConfig,
		config.ChainId(l1ChainId.Uint64()),
		l1EthClient,
		aggregatorOperator,
		executorsWithSockets,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to set up operator peering: %v", err)
	}

	t.Log("Verifying operator registration in AllocationManager")
	currentBlock, err := l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}

	registeredOperators, err := l1CC.GetOperatorSetMembersWithPeering(
		chainConfig.AVSAccountAddress,
		executorOperatorSetId,
		currentBlock,
	)
	if err != nil {
		t.Logf("Failed to get operator set members: %v", err)
	} else {
		t.Logf("Found %d operators registered to operator set %d at block %d:",
			len(registeredOperators), executorOperatorSetId, currentBlock)
		for i, op := range registeredOperators {
			t.Logf("  Operator %d: %s", i, op.OperatorAddress)
		}
	}

	// Also log what operators we expect to be registered
	t.Log("Expected operator addresses:")
	for i, exec := range executorsWithSockets {
		operatorAddr, _ := exec.Executor.DeriveAddress()
		t.Logf("  Executor %d: %s", i, operatorAddr.Hex())
	}

	t.Log("Delegating stake to executor operators")

	// Build stake configs for all 4 executors
	stakerPkList := []string{
		chainConfig.ExecStakerAccountPrivateKey,
		chainConfig.ExecStaker2AccountPrivateKey,
		chainConfig.ExecStaker3AccountPrivateKey,
		chainConfig.ExecStaker4AccountPrivateKey,
	}
	stakerAddrList := []string{
		chainConfig.ExecStakerAccountAddress,
		chainConfig.ExecStaker2AccountAddress,
		chainConfig.ExecStaker3AccountAddress,
		chainConfig.ExecStaker4AccountAddress,
	}
	operatorAddrList := []string{
		chainConfig.ExecOperatorAccountAddress,
		chainConfig.ExecOperator2AccountAddress,
		chainConfig.ExecOperator3AccountAddress,
		chainConfig.ExecOperator4AccountAddress,
	}

	// Define stake weights: 40%, 30%, 20%, 10%
	stakeWeights := []uint64{
		400000000000000000, // 40% of total
		300000000000000000, // 30%
		200000000000000000, // 20%
		100000000000000000, // 10%
	}

	stakeConfigs := make([]*testUtils.StakerDelegationConfig, numExecutorOperators)
	for i := 0; i < numExecutorOperators; i++ {
		stakeConfigs[i] = &testUtils.StakerDelegationConfig{
			StakerPrivateKey:   stakerPkList[i],
			StakerAddress:      stakerAddrList[i],
			OperatorPrivateKey: operatorPkList[i],
			OperatorAddress:    operatorAddrList[i],
			OperatorSetId:      executorOperatorSetId,
			StrategyAddress:    testUtils.Strategy_STETH,
			Magnitude:          stakeWeights[i],
		}
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

	// Create generation reservation to set up the operator table calculator
	avsAddr := common.HexToAddress(chainConfig.AVSAccountAddress)
	ecdsaCalculatorAddr, err := caller.GetTableCalculatorAddress(config.CurveTypeECDSA, config.ChainId_EthereumMainnet)
	if err != nil {
		t.Fatalf("Failed to create calculator address: %v", err)
	}

	t.Logf(
		"Creating generation reservation with ECDSA table calculator %s for executor operator set %d",
		ecdsaCalculatorAddr.Hex(),
		executorOperatorSetId,
	)

	_, err = avsConfigCaller.CreateGenerationReservation(
		ctx,
		avsAddr,
		executorOperatorSetId,
		ecdsaCalculatorAddr,
		avsAddr,
		maxStalenessPeriod,
	)

	if err != nil {
		t.Logf("Warning: Failed to create generation reservation: %v", err)
	}

	contractAddresses := config.CoreContracts[config.ChainId_EthereumMainnet]
	chainIdsToIgnore := []*big.Int{big.NewInt(8453)}

	ecdsaInfos := make([]tableTransporter.OperatorKeyInfo, len(execKeys))
	operatorAddressList := []string{
		chainConfig.ExecOperatorAccountAddress,
		chainConfig.ExecOperator2AccountAddress,
		chainConfig.ExecOperator3AccountAddress,
		chainConfig.ExecOperator4AccountAddress,
	}

	for i, keyPair := range execKeys {
		ecdsaPrivKey := keyPair.PrivateKey.(*ecdsa.PrivateKey)
		privateKeyHex := fmt.Sprintf("0x%x", ecdsaPrivKey.Bytes())
		ecdsaInfos[i] = tableTransporter.OperatorKeyInfo{
			PrivateKeyHex:   privateKeyHex,
			OperatorAddress: common.HexToAddress(operatorAddressList[i]),
		}
	}

	cfg := &tableTransporter.MultipleOperatorConfig{
		TransporterPrivateKey:     chainConfig.AVSAccountPrivateKey,
		L1RpcUrl:                  "http://localhost:8545",
		L1ChainId:                 1,
		L2RpcUrl:                  "",
		L2ChainId:                 0,
		CrossChainRegistryAddress: contractAddresses.CrossChainRegistry,
		ChainIdsToIgnore:          chainIdsToIgnore,
		Logger:                    l,
		Operators:                 ecdsaInfos,
		AVSAddress:                common.HexToAddress(chainConfig.AVSAccountAddress),
		OperatorSetId:             executorOperatorSetId,
		TransportBLSPrivateKey:    transportBlsKey,
		CurveType:                 config.CurveTypeECDSA,
	}

	err = tableTransporter.TransportTableWithSimpleMultiOperators(cfg)
	if err != nil {
		t.Fatalf("Failed to transport stake tables: %v", err)
	}

	currentBlock, err = l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}

	// After sorting by OperatorIndex, operators array is:
	// Operator 0 (OperatorIndex=0): 40% stake
	// Operator 1 (OperatorIndex=1): 30% stake
	// Operator 2 (OperatorIndex=2): 20% stake
	// Operator 3 (OperatorIndex=3): 10% stake
	testCases := []thresholdTestCase{
		{
			name:                       "Success_SingleLargerStakeWeightTaken",
			aggregationThreshold:       3500,           // 10%
			verificationThreshold:      3500,           // 10%
			respondingOperatorIdxs:     []int{3, 2, 0}, // Operator with 40% stake preferred
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true,
			operatorResponses: map[int][]byte{
				0: []byte("minority-chosen-response"), // 40% stake
				2: []byte("majority-response"),        // 20% stake
				3: []byte("majority-response"),        // 10% stake - total 30% for minority
			},
		},
		{
			name:                       "Success_LowThreshold_SingleHighStakeOperator",
			aggregationThreshold:       1000,     // 10%
			verificationThreshold:      1000,     // 10%
			respondingOperatorIdxs:     []int{0}, // Operator with 40% stake
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true, // 40% > 10%
		},
		{
			name:                       "Failure_HighVerificationThreshold_SingleHighStakeOperator",
			aggregationThreshold:       1000,     // 10% - aggregation succeeds
			verificationThreshold:      5000,     // 50% - verification should fail
			respondingOperatorIdxs:     []int{0}, // Operator with 40% stake
			shouldVerifySucceed:        false,
			shouldMeetSigningThreshold: true, // 40% > 10%
		},
		{
			name:                       "Success_ExactThreshold_SingleOperator",
			aggregationThreshold:       2000,     // 20%
			verificationThreshold:      2000,     // 20%
			respondingOperatorIdxs:     []int{2}, // Operator with exactly 20% stake
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true, // 20% >= 20%
		},
		{
			name:                       "Failure_InsufficientCombinedStake",
			aggregationThreshold:       2000,        // 20% - aggregation succeeds
			verificationThreshold:      4000,        // 40% - verification should fail
			respondingOperatorIdxs:     []int{2, 3}, // 20% + 10% = 30% combined
			shouldVerifySucceed:        false,
			shouldMeetSigningThreshold: true, // 30% > 20% (same response)
		},
		{
			name:                       "Success_AllOperators",
			aggregationThreshold:       9000,              // 90%
			verificationThreshold:      9000,              // 90%
			respondingOperatorIdxs:     []int{0, 1, 2, 3}, // 100% combined
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true, // 100% > 90% (same response)
		},
		{
			name:                       "Success_ExactThreshold_MultipleOperators",
			aggregationThreshold:       5000,        // 50%
			verificationThreshold:      5000,        // 50%
			respondingOperatorIdxs:     []int{0, 3}, // 40% + 10% = exactly 50%
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true, // 50% >= 50% (same response)
		},
		{
			name:                       "Success_AllOperators_",
			aggregationThreshold:       6000,              // 60%
			verificationThreshold:      6000,              // 60%
			respondingOperatorIdxs:     []int{3, 2, 1, 0}, // All operators respond
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true, // 70% >= 50% (same response)
			operatorResponses: map[int][]byte{
				0: []byte("higher-response"), // 40% stake
				1: []byte("higher-response"), // 30% stake - total 70% for majority
				2: []byte("lower-response"),  // 20% stake
				3: []byte("lower-response"),  // 10% stake - total 30% for minority
			},
		},
		{
			name:                       "ConflictingResponses_StakeWeightChoice",
			aggregationThreshold:       4000,           // 40% - aggregation threshold is low
			verificationThreshold:      4000,           // 40% - but verification requires high consensus
			respondingOperatorIdxs:     []int{2, 1, 0}, // 90% total stake responds
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true, // 40% >= 40%
			operatorResponses: map[int][]byte{
				0: []byte("response-a"), // 40% stake
				1: []byte("response-b"), // 30% stake
				2: []byte("response-c"), // 20% stake - all different responses
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testECDSAWithThresholds(
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

func testECDSAWithThresholds(
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
	t.Logf("Responding operator indices: %v", tc.respondingOperatorIdxs)
	t.Logf("Expected verification result: %v", tc.shouldVerifySucceed)

	taskId := fmt.Sprintf("0x%064x", time.Now().UnixNano()) // Unique task ID for each test
	taskInputData := []byte("test-ecdsa-task-input-data")
	deadline := time.Now().Add(1 * time.Minute)

	pdf := peeringDataFetcher.NewPeeringDataFetcher(l1CC, l)
	callerMap := map[config.ChainId]contractCaller.IContractCaller{
		config.ChainId_EthereumMainnet: l1CC,
	}

	opManager := operatorManager.NewOperatorManager(&operatorManager.OperatorManagerConfig{
		AvsAddress: chainConfig.AVSAccountAddress,
		ChainIds:   []config.ChainId{config.ChainId_EthereumMainnet},
		L1ChainId:  config.ChainId_EthereumMainnet,
	}, callerMap, pdf, l)

	operatorPeersWeight, err := opManager.GetExecutorPeersAndWeightsForBlock(
		ctx,
		config.ChainId_EthereumMainnet,
		currentBlock,
		executorOperatorSetId,
	)
	if err != nil {
		t.Fatalf("Failed to get operator peers and weights: %v", err)
	}

	// Create ECDSA aggregator with operator addresses (not public keys)
	var operators []*aggregation.Operator[common.Address]
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

		// For ECDSA, the "PublicKey" is the signing address derived from the ECDSA key
		weights := operatorPeersWeight.Weights[peer.OperatorAddress]

		operators = append(operators, &aggregation.Operator[common.Address]{
			Address:       peer.OperatorAddress,
			PublicKey:     opset.WrappedPublicKey.ECDSAAddress,
			OperatorIndex: opset.OperatorIndex,
			Weights:       weights,
		})
	}

	// Sort operators by OperatorIndex to ensure deterministic ordering
	// This makes the operators array match the on-chain operator table ordering
	sort.Slice(operators, func(i, j int) bool {
		return operators[i].OperatorIndex < operators[j].OperatorIndex
	})

	t.Logf("======= ECDSA Operators (sorted by OperatorIndex) =======")
	totalWeight := big.NewInt(0)
	for i, op := range operators {
		weight := op.Weights[0]
		totalWeight.Add(totalWeight, weight)
		t.Logf("Operator %d: Address=%s, Index=%d, Weight=%s",
			i, op.Address, op.OperatorIndex, weight.String())
	}
	t.Logf("Total weight: %s", totalWeight.String())

	agg, err := aggregation.NewECDSATaskResultAggregator(
		context.Background(),
		taskId,
		operatorPeersWeight.RootReferenceTimestamp,
		executorOperatorSetId,
		tc.aggregationThreshold,
		l1CC,
		taskInputData,
		&deadline,
		operators,
	)
	if err != nil {
		t.Fatalf("Failed to create ECDSA task result aggregator: %v", err)
	}

	// Calculate task message hash
	//var taskIdBytes [32]byte
	//copy(taskIdBytes[:], common.HexToHash(taskId).Bytes())
	//messageHash, err := l1CC.CalculateTaskMessageHash(ctx, taskIdBytes, taskInputData)
	//if err != nil {
	//	t.Fatalf("Failed to calculate task message hash: %v", err)
	//}

	totalSigningWeight := big.NewInt(0)
	for _, operatorIdx := range tc.respondingOperatorIdxs {
		operatorAddress := operatorAddressList[operatorIdx]

		var respondingOperator *aggregation.Operator[common.Address]
		for _, op := range operators {
			if strings.EqualFold(op.Address, operatorAddress) {
				respondingOperator = op
				break
			}
		}
		if respondingOperator == nil {
			t.Fatalf("Could not find operator at index %d", operatorIdx)
		}

		operatorKeys, ok := operatorKeyMap[strings.ToLower(respondingOperator.Address)]
		if !ok {
			t.Fatalf("Could not find ECDSA keys for operator %s", respondingOperator.Address)
		}

		// Determine what response this operator should provide
		operatorResponse := taskInputData // Default response
		if tc.operatorResponses != nil {
			if customResponse, ok := tc.operatorResponses[operatorIdx]; ok {
				operatorResponse = customResponse
				t.Logf("  Using custom response for operator %d: %s", operatorIdx, string(customResponse))
			}
		}

		// Calculate the message hash for this operator's response
		var taskIdBytes [32]byte
		copy(taskIdBytes[:], common.HexToHash(taskId).Bytes())
		operatorMessageHash, err := l1CC.CalculateTaskMessageHash(ctx, taskIdBytes, operatorResponse)
		if err != nil {
			t.Fatalf("Failed to calculate task message hash for operator %s: %v", respondingOperator.Address, err)
		}
		operatorECDSADigest, err := l1CC.CalculateECDSACertificateDigestBytes(
			ctx,
			operatorPeersWeight.RootReferenceTimestamp,
			operatorMessageHash,
		)
		if err != nil {
			t.Fatalf("Failed to calculate ECDSA certificate digest for operator %s: %v", respondingOperator.Address, err)
		}

		// Create signature from this operator
		responderSigner := inMemorySigner.NewInMemorySigner(operatorKeys.PrivateKey, config.CurveTypeECDSA)

		t.Logf("  Operator %s: signing with key that derives to address %s",
			respondingOperator.Address, operatorKeys.Address.Hex())
		t.Logf("  Operator %s: publicKey field in aggregator = %s",
			respondingOperator.Address, respondingOperator.PublicKey.Hex())

		resultSig, err := responderSigner.SignMessageForSolidity(operatorECDSADigest)
		if err != nil {
			t.Fatalf("Failed to sign ECDSA certificate for operator %s: %v", respondingOperator.Address, err)
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
			t.Fatalf("Failed to sign auth data for operator %s: %v", respondingOperator.Address, err)
		}

		taskResult := &types.TaskResult{
			TaskId:          taskId,
			AvsAddress:      chainConfig.AVSAccountAddress,
			OperatorSetId:   executorOperatorSetId,
			Output:          operatorResponse,
			OperatorAddress: respondingOperator.Address,
			ResultSignature: resultSig,
			AuthSignature:   authSig,
		}

		if err := agg.ProcessNewSignature(ctx, taskResult); err != nil {
			t.Fatalf("Failed to process signature from operator %s: %v", respondingOperator.Address, err)
		}

		// Use the actual weight from the operator, not the assumed position
		operatorWeight := respondingOperator.Weights[0]
		totalSigningWeight.Add(totalSigningWeight, operatorWeight)
		t.Logf("Processed signature from operator %d (%s) with weight %s",
			operatorIdx, respondingOperator.Address, operatorWeight.String())
	}

	// Calculate total signing percentage
	signingPercentage := new(big.Float).Quo(
		new(big.Float).SetInt(totalSigningWeight),
		new(big.Float).SetInt(totalWeight),
	)
	signingPercentage.Mul(signingPercentage, big.NewFloat(100))
	percentFloat, _ := signingPercentage.Float64()

	t.Logf("Total signing weight: %s (%.1f%% of total)", totalSigningWeight.String(), percentFloat)

	// Check if threshold is met for aggregation
	signingThresholdMet := agg.SigningThresholdMet()

	// Check if the signing threshold expectation matches reality
	if tc.shouldMeetSigningThreshold && !signingThresholdMet {
		t.Errorf("Expected signing threshold to be met but it was not. Most common response stake may be below %.1f%% threshold",
			float64(tc.aggregationThreshold)/100)
		return
	} else if !tc.shouldMeetSigningThreshold && signingThresholdMet {
		t.Errorf("Expected signing threshold to NOT be met but it was. Most common response stake exceeds %.1f%% threshold",
			float64(tc.aggregationThreshold)/100)
		return
	}

	if !signingThresholdMet {
		t.Logf("✓ Signing threshold correctly not met (most common response stake < %.1f%% threshold)",
			float64(tc.aggregationThreshold)/100)
		return
	}

	t.Logf("✓ Signing threshold correctly met (most common response stake >= %.1f%% threshold)",
		float64(tc.aggregationThreshold)/100)

	finalCert, err := agg.GenerateFinalCertificate()
	if err != nil {
		t.Fatalf("Failed to generate final certificate: %v", err)
	}

	// Get the actual number of signers from the certificate (those who signed the winning response)
	numSigners := len(finalCert.SignersSignatures)
	numNonSigners := len(operators) - numSigners
	t.Logf("Final certificate generated with %d signers (winning response) and %d non-signers", numSigners, numNonSigners)

	// Recalculate message hash based on the actual winning response
	// This is necessary when operators sign different responses - we need to verify against the winning one
	var winningTaskIdBytes [32]byte
	copy(winningTaskIdBytes[:], common.HexToHash(taskId).Bytes())
	winningMessageHash, err := l1CC.CalculateTaskMessageHash(ctx, winningTaskIdBytes, finalCert.TaskResponse)
	if err != nil {
		t.Fatalf("Failed to calculate message hash for winning response: %v", err)
	}

	// Get the combined signature from the certificate
	submitParams := finalCert.ToSubmitParams()

	combinedSig, err := caller.GetFinalECDSASignature(submitParams.SignersSignatures)
	if err != nil {
		t.Fatalf("Failed to combine signatures: %v", err)
	}

	t.Logf("Combined %d signatures into final signature", len(submitParams.SignersSignatures))

	// Verify the ECDSA certificate using the verification threshold
	valid, signers, err := l1CC.VerifyECDSACertificate(
		winningMessageHash,
		combinedSig,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		executorOperatorSetId,
		operatorPeersWeight.RootReferenceTimestamp,
		tc.verificationThreshold,
	)

	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "threshold") || strings.Contains(errStr, "insufficient") || strings.Contains(errStr, "weight") {
			if tc.shouldVerifySucceed {
				t.Errorf("Expected verification to succeed but got threshold error: %v", err)
			} else {
				t.Logf("✓ Test passed: Verification failed as expected with threshold error: %v", err)
			}
		} else {
			t.Errorf("Unexpected error during verification: %v", err)
		}
	} else {
		if !valid && tc.shouldVerifySucceed {
			t.Errorf("Expected verification to succeed but certificate was invalid")
		} else if valid && !tc.shouldVerifySucceed {
			t.Errorf("Expected verification to fail but certificate was valid")
		} else if valid && tc.shouldVerifySucceed {
			t.Logf("✓ Test passed: ECDSA certificate verification succeeded as expected")
			t.Logf("Certificate validated with %d signers", len(signers))
		} else {
			t.Logf("✓ Test passed: ECDSA certificate verification failed as expected")
		}
	}

	if submitParams.SignersSignatures != nil {
		t.Logf("Number of ECDSA signatures collected: %d", len(submitParams.SignersSignatures))
		for addr := range submitParams.SignersSignatures {
			t.Logf("  - Signature from: %s", addr.Hex())
		}
	}
}
