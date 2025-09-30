package certificateVerifier

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
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

type ecdsaThresholdTestCase struct {
	name                       string
	aggregationThreshold       uint16
	verificationThreshold      uint16
	respondingOperatorIdxs     []int
	shouldVerifySucceed        bool
	shouldMeetSigningThreshold bool
	operatorResponses          map[int][]byte
}

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

	// Generate ECDSA keys for operators
	execKeys := make([]*testUtils.WrappedKeyPair, 4)
	operatorKeyMap := make(map[string]*testUtils.WrappedKeyPair)

	// Generate unique ECDSA keys for each operator
	operatorPrivateKeyStrings := []string{
		chainConfig.ExecOperatorAccountPk,
		chainConfig.ExecOperator2AccountPk,
		chainConfig.ExecOperator3AccountPk,
		chainConfig.ExecOperator4AccountPk,
	}

	for index := 0; index < 4; index++ {
		// Create ECDSA private key from the operator's private key string
		ecdsaPrivKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(operatorPrivateKeyStrings[index])
		if err != nil {
			t.Fatalf("Failed to create ECDSA private key for operator %d: %v", index+1, err)
		}

		derivedAddress, err := ecdsaPrivKey.DeriveAddress()
		if err != nil {
			t.Fatalf("Failed to derive address for operator %d: %v", index+1, err)
		}

		execKeys[index] = &testUtils.WrappedKeyPair{
			PrivateKey: ecdsaPrivKey,
			Address:    derivedAddress,
		}

		t.Logf("Operator %d signing address: %s", index+1, derivedAddress.Hex())
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

	t.Logf("Configuring operator set %d with curve type ECDSA for 4 executors", executorOperatorSetId)
	_, err = avsConfigCaller.ConfigureAVSOperatorSet(ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		executorOperatorSetId,
		config.CurveTypeECDSA,
	)

	if err != nil {
		t.Fatalf("Failed to configure executor operator set %d: %v", executorOperatorSetId, err)
	}

	operators := []*operator.Operator{
		{
			TransactionPrivateKey: chainConfig.ExecOperatorAccountPk,
			SigningPrivateKey:     execKeys[0].PrivateKey,
			Curve:                 config.CurveTypeECDSA,
			OperatorSetIds:        []uint32{executorOperatorSetId},
		},
		{
			TransactionPrivateKey: chainConfig.ExecOperator2AccountPk,
			SigningPrivateKey:     execKeys[1].PrivateKey,
			Curve:                 config.CurveTypeECDSA,
			OperatorSetIds:        []uint32{executorOperatorSetId},
		},
		{
			TransactionPrivateKey: chainConfig.ExecOperator3AccountPk,
			SigningPrivateKey:     execKeys[2].PrivateKey,
			Curve:                 config.CurveTypeECDSA,
			OperatorSetIds:        []uint32{executorOperatorSetId},
		},
		{
			TransactionPrivateKey: chainConfig.ExecOperator4AccountPk,
			SigningPrivateKey:     execKeys[3].PrivateKey,
			Curve:                 config.CurveTypeECDSA,
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

	// Setup stake delegation
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

	// Create generation reservation to set up the operator table calculator
	avsAddr := common.HexToAddress(chainConfig.AVSAccountAddress)
	ecdsaCalculatorAddr, err := caller.GetTableCalculatorAddress(config.CurveTypeECDSA, config.ChainId_EthereumAnvil)
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

	time.Sleep(time.Second * 3)

	// Transport operator tables using the standard transport method
	// ECDSA doesn't require special BLS operator info
	contractAddresses := config.CoreContracts[config.ChainId_EthereumAnvil]
	chainIdsToIgnore := []*big.Int{
		big.NewInt(1),     // Ethereum Mainnet
		big.NewInt(8453),  // Base Mainnet
		big.NewInt(31338), // L2 anvil
	}

	// Use the standard transport method for ECDSA
	tableTransporter.TransportTable(
		chainConfig.AVSAccountPrivateKey, // Transport using AVS account
		"http://localhost:8545",
		31337,
		"", // No L2
		0,  // No L2 chain ID
		contractAddresses.CrossChainRegistry,
		transportBlsKey, // Still use this for signing transport
		chainIdsToIgnore,
		l,
	)

	time.Sleep(time.Second * 6)

	currentBlock, err := l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}

	// After sorting by OperatorIndex, operators array is:
	// Operator 0 (OperatorIndex=0): 40% stake
	// Operator 1 (OperatorIndex=1): 30% stake
	// Operator 2 (OperatorIndex=2): 20% stake
	// Operator 3 (OperatorIndex=3): 10% stake
	testCases := []ecdsaThresholdTestCase{
		{
			name:                       "Success_SingleOperator_HighStake",
			aggregationThreshold:       1000,     // 10%
			verificationThreshold:      1000,     // 10%
			respondingOperatorIdxs:     []int{0}, // Operator 0 with 40% stake
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true,
		},
		{
			name:                       "Success_TwoOperators_CombinedStake",
			aggregationThreshold:       4000,        // 40%
			verificationThreshold:      4000,        // 40%
			respondingOperatorIdxs:     []int{1, 2}, // 30% + 20% = 50% combined
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true, // 50% > 40%
		},
		{
			name:                       "Failure_InsufficientCombinedStake",
			aggregationThreshold:       2000,        // 20% - aggregation succeeds
			verificationThreshold:      4000,        // 40% - verification should fail
			respondingOperatorIdxs:     []int{2, 3}, // 20% + 10% = 30% combined
			shouldVerifySucceed:        false,
			shouldMeetSigningThreshold: true, // 30% > 20%
		},
		{
			name:                       "Success_AllOperators",
			aggregationThreshold:       9000,              // 90%
			verificationThreshold:      9000,              // 90%
			respondingOperatorIdxs:     []int{0, 1, 2, 3}, // 100% combined
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true, // 100% > 90%
		},
		{
			name:                       "Failure_SingleLowStakeOperator",
			aggregationThreshold:       500,      // 5% - aggregation succeeds
			verificationThreshold:      2000,     // 20% - verification should fail
			respondingOperatorIdxs:     []int{3}, // Operator 3 with 10% stake
			shouldVerifySucceed:        false,
			shouldMeetSigningThreshold: true, // 10% > 5%
		},
		{
			name:                       "Success_ExactThreshold_MultipleOperators",
			aggregationThreshold:       5000,        // 50%
			verificationThreshold:      5000,        // 50%
			respondingOperatorIdxs:     []int{0, 3}, // 40% + 10% = exactly 50%
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true, // 50% >= 50%
		},
		{
			name:                       "ConflictingResponses_MajorityWins",
			aggregationThreshold:       6000,              // 60%
			verificationThreshold:      6000,              // 60%
			respondingOperatorIdxs:     []int{0, 1, 2, 3}, // All operators respond
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true, // 70% (majority) > 60%
			operatorResponses: map[int][]byte{
				0: []byte("majority-response"), // 40% stake
				1: []byte("majority-response"), // 30% stake - total 70% for majority
				2: []byte("minority-response"), // 20% stake
				3: []byte("minority-response"), // 10% stake - total 30% for minority
			},
		},
		{
			name:                       "ConflictingResponses_InsufficientConsensus",
			aggregationThreshold:       3000,           // 30% - aggregation threshold is low
			verificationThreshold:      7000,           // 70% - but verification requires high consensus
			respondingOperatorIdxs:     []int{0, 1, 2}, // 90% total stake responds
			shouldVerifySucceed:        false,
			shouldMeetSigningThreshold: true, // 40% (largest) > 30%
			operatorResponses: map[int][]byte{
				0: []byte("response-a"), // 40% stake
				1: []byte("response-b"), // 30% stake
				2: []byte("response-c"), // 20% stake - all different responses
			},
		},
		{
			name:                       "StakeWeighted_MinorityOperatorsMajorityStake",
			aggregationThreshold:       7000,        // 70% threshold
			verificationThreshold:      7000,        // 70% threshold
			respondingOperatorIdxs:     []int{0, 1}, // 2 out of 4 operators respond
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true, // 70% stake (40% + 30%) >= 70%
			operatorResponses: map[int][]byte{
				0: []byte("consensus-response"), // 40% stake
				1: []byte("consensus-response"), // 30% stake - total 70% stake
			},
		},
		{
			name:                       "StakeWeighted_MajorityOperatorsMinorityStake",
			aggregationThreshold:       4000,           // 40% threshold
			verificationThreshold:      4000,           // 40% threshold
			respondingOperatorIdxs:     []int{1, 2, 3}, // 3 out of 4 operators respond
			shouldVerifySucceed:        false,
			shouldMeetSigningThreshold: false, // No single response has 40% stake
			operatorResponses: map[int][]byte{
				1: []byte("response-a"), // 30% stake - most common response below 40%
				2: []byte("response-b"), // 20% stake
				3: []byte("response-c"), // 10% stake
			},
		},
		{
			name:                       "TieBreaking_HigherStakeWins",
			aggregationThreshold:       3500,              // 35% threshold
			verificationThreshold:      3500,              // 35% threshold
			respondingOperatorIdxs:     []int{3, 2, 1, 0}, // All operators
			shouldVerifySucceed:        true,
			shouldMeetSigningThreshold: true, // 40% (single op) > 35%
			operatorResponses: map[int][]byte{
				0: []byte("high-stake-response"), // 40% stake - 1 operator wins
				1: []byte("low-stake-response"),  // 30% stake
				2: []byte("low-stake-response"),  // 20% stake - 2 operators total 50%
				3: []byte("third-response"),      // 10% stake
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
	tc ecdsaThresholdTestCase,
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
