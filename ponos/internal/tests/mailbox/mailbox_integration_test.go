package mailbox

import (
	"context"
	"fmt"
	"math/big"
	"os/exec"
	"slices"
	"sync"
	"testing"
	"time"

	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	aggregatorMemory "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/EVMChainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contextManager/taskBlockContextManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operatorManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type NetworkTarget string

const (
	NetworkTarget_L1 NetworkTarget = "l1"
	NetworkTarget_L2 NetworkTarget = "l2"
)

func testL1MailboxForCurve(t *testing.T, curveType config.CurveType, networkTarget NetworkTarget) {
	if !slices.Contains([]config.CurveType{config.CurveTypeBN254, config.CurveTypeECDSA}, curveType) {
		t.Fatalf("Unsupported curve type: %s", curveType)
	}
	const (
		L1RpcUrl = "http://127.0.0.1:8545"
		L2RpcUrl = "http://127.0.0.1:9545"
	)

	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	root := testUtils.GetProjectRootPath()
	t.Logf("Project root path: %s", root)

	chainConfig, err := testUtils.ReadChainConfig(root)
	if err != nil {
		t.Fatalf("Failed to read chain config: %v", err)
	}

	// aggregator is bn254, executor is ecdsa
	aggKeysBN254, _, _, err := testUtils.GetKeysForCurveType(t, config.CurveTypeBN254, chainConfig)
	if err != nil {
		t.Fatalf("Failed to get keys for curve type %s: %v", config.CurveTypeBN254, err)
	}

	_, execKeysECDSA, _, err := testUtils.GetKeysForCurveType(t, config.CurveTypeECDSA, chainConfig)
	if err != nil {
		t.Fatalf("Failed to get keys for curve type %s: %v", config.CurveTypeECDSA, err)
	}

	coreContracts, err := eigenlayer.LoadContracts()
	if err != nil {
		t.Fatalf("Failed to load core contracts: %v", err)
	}

	imContractStore := inMemoryContractStore.NewInMemoryContractStore(coreContracts, l)

	tlp := transactionLogParser.NewTransactionLogParser(imContractStore, l)

	l1EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   L1RpcUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)
	l2EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   L2RpcUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)

	taskQueue := make(chan *types.Task)

	var pollerConfig *EVMChainPoller.EVMChainPollerConfig
	var pollerEthClient *ethereum.EthereumClient
	if networkTarget == NetworkTarget_L1 {
		pollerConfig = &EVMChainPoller.EVMChainPollerConfig{
			ChainId:              config.ChainId_EthereumAnvil,
			PollingInterval:      time.Duration(10) * time.Second,
			InterestingContracts: imContractStore.ListContractAddressesForChain(config.ChainId_EthereumAnvil),
			AvsAddress:           chainConfig.AVSAccountAddress,
		}
		pollerEthClient = l1EthereumClient
	} else {
		pollerConfig = &EVMChainPoller.EVMChainPollerConfig{
			ChainId:              config.ChainId_BaseSepoliaAnvil,
			PollingInterval:      time.Duration(10) * time.Second,
			InterestingContracts: imContractStore.ListContractAddressesForChain(config.ChainId_BaseSepoliaAnvil),
			AvsAddress:           chainConfig.AVSAccountAddress,
		}
		pollerEthClient = l2EthereumClient
	}

	// Create an in-memory store for the poller
	aggStore := aggregatorMemory.NewInMemoryAggregatorStore()
	poller := EVMChainPoller.NewEVMChainPoller(
		pollerEthClient,
		taskQueue,
		tlp,
		pollerConfig,
		imContractStore,
		aggStore,
		taskBlockContextManager.NewTaskBlockContextManager(context.Background(), aggStore, l),
		l,
	)

	l1EthClient, err := l1EthereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get L1 Ethereum contract caller: %v", err)
	}

	l2EthClient, err := l2EthereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get L2 Ethereum contract caller: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	anvilWg := &sync.WaitGroup{}
	anvilWg.Add(1)
	startErrorsChan := make(chan error, 1)

	anvilCtx, anvilCancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer anvilCancel()

	_ = testUtils.KillallAnvils()

	l1Anvil, err := testUtils.StartL1Anvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start L1 Anvil: %v", err)
	}
	go testUtils.WaitForAnvil(anvilWg, anvilCtx, t, l1EthereumClient, startErrorsChan)

	var l2Anvil *exec.Cmd
	if networkTarget == NetworkTarget_L2 {
		anvilWg.Add(1)
		l2Anvil, err = testUtils.StartL2Anvil(root, ctx)
		if err != nil {
			t.Fatalf("Failed to start L2 Anvil: %v", err)
		}
		go testUtils.WaitForAnvil(anvilWg, anvilCtx, t, l2EthereumClient, startErrorsChan)
	}

	anvilWg.Wait()
	close(startErrorsChan)
	for err := range startErrorsChan {
		if err != nil {
			t.Errorf("Failed to start Anvil: %v", err)
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
		t.Fatalf("Failed to create L2 contract caller: %v", err)
	}

	var l2CC *caller.ContractCaller
	if networkTarget == NetworkTarget_L2 {
		l2PrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AppAccountPrivateKey, l2EthClient, l)
		if err != nil {
			t.Fatalf("Failed to create L2 private key signer: %v", err)
		}

		l2CC, err = caller.NewContractCaller(l2EthClient, l2PrivateKeySigner, l)
		if err != nil {
			t.Fatalf("Failed to create L2 contract caller: %v", err)
		}
	}

	reservations, err := l1CC.GetActiveGenerationReservations()
	if err != nil {
		t.Fatalf("Failed to get active generation reservations: %v", err)
	}
	for _, reservation := range reservations {
		fmt.Printf("Active generation reservation: %+v\n", reservation)
	}

	l.Sugar().Infow("Setting up operator peering",
		zap.String("AVSAccountAddress", chainConfig.AVSAccountAddress),
	)

	aggOpsetId := uint32(0)
	execOpsetId := uint32(1)

	allOperatorSetIds := []uint32{aggOpsetId, execOpsetId}

	t.Logf("------------------------------------------- Configuring operator sets -------------------------------------------")

	// Configure operator sets with their curve types
	avsAddr := common.HexToAddress(chainConfig.AVSAccountAddress)

	// Create AVS config caller for operator set configuration
	avsPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create AVS private key signer: %v", err)
	}

	avsConfigCaller, err := caller.NewContractCaller(l1EthClient, avsPrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create AVS config contract caller: %v", err)
	}

	// Configure BN254 operator set for aggregator
	_, err = avsConfigCaller.ConfigureAVSOperatorSet(ctx, avsAddr, aggOpsetId, config.CurveTypeBN254)
	if err != nil {
		t.Fatalf("Failed to configure BN254 operator set: %v", err)
	}
	t.Logf("Configured operator set %d with BN254 curve type", aggOpsetId)

	// Configure ECDSA operator set for executor
	_, err = avsConfigCaller.ConfigureAVSOperatorSet(ctx, avsAddr, execOpsetId, config.CurveTypeECDSA)
	if err != nil {
		t.Fatalf("Failed to configure ECDSA operator set: %v", err)
	}
	t.Logf("Configured operator set %d with ECDSA curve type", execOpsetId)

	// Create generation reservations for both operator sets
	maxStalenessPeriod := uint32(0) // 0 allows certificates to always be valid regardless of referenceTimestamp

	// BN254 table calculator for aggregator
	bn254CalculatorAddr := common.HexToAddress(caller.BN254TableCalculatorAddress)
	_, err = avsConfigCaller.CreateGenerationReservation(
		ctx,
		avsAddr,
		aggOpsetId,
		bn254CalculatorAddr,
		avsAddr, // AVS is the owner
		maxStalenessPeriod,
	)
	if err != nil {
		t.Fatalf("Failed to create generation reservation for BN254 operator set: %v", err)
	}
	t.Logf("Created generation reservation for operator set %d with BN254 table calculator", aggOpsetId)

	// ECDSA table calculator for executor
	ecdsaCalculatorAddr := common.HexToAddress(caller.ECDSATableCalculatorAddress)
	_, err = avsConfigCaller.CreateGenerationReservation(
		ctx,
		avsAddr,
		execOpsetId,
		ecdsaCalculatorAddr,
		avsAddr, // AVS is the owner
		maxStalenessPeriod,
	)
	if err != nil {
		t.Fatalf("Failed to create generation reservation for ECDSA operator set: %v", err)
	}
	t.Logf("Created generation reservation for operator set %d with ECDSA table calculator", execOpsetId)

	t.Logf("------------------------------------------- Setting up operator peering -------------------------------------------")
	// NOTE: we must register ALL opsets regardles of which curve type we are using, otherwise table transport fails

	err = testUtils.SetupOperatorPeering(
		ctx,
		chainConfig,
		config.ChainId(l1ChainId.Uint64()),
		l1EthClient,
		// aggregator is BN254
		&operator.Operator{
			TransactionPrivateKey: chainConfig.OperatorAccountPrivateKey,
			SigningPrivateKey:     aggKeysBN254.PrivateKey,
			Curve:                 config.CurveTypeBN254,
			OperatorSetIds:        []uint32{aggOpsetId},
		},
		// executor is ecdsa
		&operator.Operator{
			TransactionPrivateKey: chainConfig.ExecOperatorAccountPk,
			SigningPrivateKey:     execKeysECDSA.PrivateKey,
			Curve:                 config.CurveTypeECDSA,
			OperatorSetIds:        []uint32{execOpsetId},
		},
		"localhost:9000",
		l,
	)
	if err != nil {
		t.Fatalf("Failed to set up operator peering: %v", err)
	}

	err = testUtils.DelegateStakeToOperators(
		t,
		ctx,
		&testUtils.StakerDelegationConfig{
			StakerPrivateKey:   chainConfig.AggStakerAccountPrivateKey,
			StakerAddress:      chainConfig.AggStakerAccountAddress,
			OperatorPrivateKey: chainConfig.OperatorAccountPrivateKey,
			OperatorAddress:    chainConfig.OperatorAccountAddress,
			OperatorSetId:      0,
			StrategyAddress:    testUtils.Strategy_WETH,
		},
		&testUtils.StakerDelegationConfig{
			StakerPrivateKey:   chainConfig.ExecStakerAccountPrivateKey,
			StakerAddress:      chainConfig.ExecStakerAccountAddress,
			OperatorPrivateKey: chainConfig.ExecOperatorAccountPk,
			OperatorAddress:    chainConfig.ExecOperatorAccountAddress,
			OperatorSetId:      1,
			StrategyAddress:    testUtils.Strategy_STETH,
		},
		chainConfig.AVSAccountAddress,
		l1EthClient,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to delegate stake to operators: %v", err)
	}

	t.Logf("All operator set IDs: %v", allOperatorSetIds)
	// update current block to account for transport
	currentBlock, err := l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}
	testUtils.DebugOpsetData(t, chainConfig, eigenlayerContractAddrs, l1EthClient, currentBlock, allOperatorSetIds)

	time.Sleep(time.Second * 2)

	l.Sugar().Infow("------------------------ Transporting L1 tables ------------------------")

	testUtils.TransportStakeTables(l, networkTarget == NetworkTarget_L2)
	l.Sugar().Infow("Sleeping for 6 seconds to allow table transport to complete")
	time.Sleep(time.Second * 6)

	l.Sugar().Infow("------------------------ Setting up mailbox ------------------------")

	mailboxEthClient := l1EthClient
	avsTaskHookAddress := chainConfig.AVSTaskHookAddressL1
	if networkTarget == NetworkTarget_L2 {
		mailboxEthClient = l2EthClient
		avsTaskHookAddress = chainConfig.AVSTaskHookAddressL2
	}

	avsPrivateKeySigner, err = transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, mailboxEthClient, l)
	if err != nil {
		t.Fatalf("Failed to create AVS private key signer: %v", err)
	}

	avsCc, err := caller.NewContractCaller(mailboxEthClient, avsPrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create AVS contract caller: %v", err)
	}

	// setup mailbox with both exec types
	err = testUtils.SetupTaskMailbox(
		ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		common.HexToAddress(avsTaskHookAddress),
		[]uint32{execOpsetId},
		[]config.CurveType{config.CurveTypeECDSA},
		avsCc,
	)
	if err != nil {
		t.Fatalf("Failed to set up task mailbox: %v", err)
	}

	if err := poller.Start(ctx); err != nil {
		cancel()
		t.Fatalf("Failed to start EVM L1Chain Poller: %v", err)
	}

	pdf := peeringDataFetcher.NewPeeringDataFetcher(l1CC, l)

	callerMap := map[config.ChainId]contractCaller.IContractCaller{
		config.ChainId_EthereumAnvil: l1CC,
	}
	opManagerChainIds := []config.ChainId{config.ChainId_EthereumAnvil}

	if networkTarget == NetworkTarget_L2 {
		callerMap[config.ChainId_BaseSepoliaAnvil] = l2CC
		opManagerChainIds = append(opManagerChainIds, config.ChainId_BaseSepoliaAnvil)
	}

	opManager := operatorManager.NewOperatorManager(&operatorManager.OperatorManagerConfig{
		AvsAddress: chainConfig.AVSAccountAddress,
		ChainIds:   opManagerChainIds,
		L1ChainId:  config.ChainId_EthereumAnvil,
	}, callerMap, pdf, l)

	hasErrors := false
	go func() {
		for task := range taskQueue {
			fmt.Printf("Received task: %+v\n", task)
			t.Logf("Processing task: %+v", task)

			assert.Equal(t, common.HexToAddress(chainConfig.AVSAccountAddress), common.HexToAddress(task.AVSAddress))
			assert.True(t, len(task.TaskId) > 0)
			assert.True(t, len(task.Payload) > 0)

			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to create task session: %v", err)
				cancel()
				return
			}

			operatorPeersWeight, err := opManager.GetExecutorPeersAndWeightsForBlock(
				ctx,
				task.ChainId,
				task.SourceBlockNumber,
				task.OperatorSetId,
			)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to get operator peers and weights: %v", err)
				cancel()
				return
			}

			operators := []*aggregation.Operator[common.Address]{}
			for _, peer := range operatorPeersWeight.Operators {
				opset, err := peer.GetOperatorSet(task.OperatorSetId)
				if err != nil {
					hasErrors = true
					l.Sugar().Errorf("Failed to get operator set for peer %s: %v", peer.OperatorAddress, err)
					cancel()
					return
				}
				operators = append(operators, &aggregation.Operator[common.Address]{
					Address:   peer.OperatorAddress,
					PublicKey: opset.WrappedPublicKey.ECDSAAddress,
				})
			}
			t.Logf("======= Operators =======")
			for i, op := range operators {
				t.Logf("Operator %d: %+v", i, op)
			}

			resultAgg, err := aggregation.NewECDSATaskResultAggregator(
				ctx,
				task.TaskId,
				operatorPeersWeight.RootReferenceTimestamp,
				task.OperatorSetId,
				6667,
				l1CC,
				task.Payload,
				task.DeadlineUnixSeconds,
				operators,
			)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to create task result aggregator: %v", err)
				cancel()
				return
			}

			// ----------------------------------------------------------------
			// Compile the result
			// ----------------------------------------------------------------
			outputResult := util.BigIntToHex(new(big.Int).SetUint64(16))

			taskResult := &types.TaskResult{
				TaskId:          task.TaskId,
				AvsAddress:      chainConfig.AVSAccountAddress,
				OperatorSetId:   task.OperatorSetId,
				Output:          outputResult,
				OperatorAddress: chainConfig.ExecOperatorAccountAddress,
				ResultSignature: nil,
				AuthSignature:   nil,
			}

			signer := inMemorySigner.NewInMemorySigner(execKeysECDSA.PrivateKey, config.CurveTypeECDSA)
			// Log the actual signing address derived from the private key
			if ecdsaPK, ok := execKeysECDSA.PrivateKey.(*cryptoLibsEcdsa.PrivateKey); ok {
				signerAddress, _ := ecdsaPK.DeriveAddress()
				t.Logf("Signer address derived from private key: %s", signerAddress.Hex())
			}

			// Step 1: Sign the result using certificate digest (matching the executor)
			var taskIdBytes [32]byte
			copy(taskIdBytes[:], common.HexToHash(taskResult.TaskId).Bytes())
			messageHash, err := l1CC.CalculateTaskMessageHash(ctx, taskIdBytes, outputResult)
			if err != nil {
				t.Errorf("Failed to calculate task message hash: %v", err)
				return
			}
			certificateDigestBytes, err := l1CC.CalculateECDSACertificateDigestBytes(
				ctx,
				operatorPeersWeight.RootReferenceTimestamp,
				messageHash,
			)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to calculate certificate digest: %v", err)
				cancel()
				return
			}

			l.Sugar().Debugw("Signing result",
				"taskId", taskResult.TaskId,
				"operatorAddress", taskResult.OperatorAddress,
				"outputLength", len(outputResult),
				"outputDigest", fmt.Sprintf("%x", messageHash),
				"certificateDigest", fmt.Sprintf("%x", certificateDigestBytes),
				"referenceTimestamp", operatorPeersWeight.RootReferenceTimestamp,
			)

			resultSig, err := signer.SignMessageForSolidity(certificateDigestBytes)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to sign result: %v", err)
				cancel()
				return
			}
			taskResult.ResultSignature = resultSig

			l.Sugar().Debugw("Result signature created",
				"taskId", taskResult.TaskId,
				"resultSigLength", len(resultSig),
				"resultSigHex", fmt.Sprintf("%x", resultSig),
			)

			// Step 2: Sign the auth data (unique per operator)
			resultSigDigest := util.GetKeccak256Digest(taskResult.ResultSignature)
			authData := &types.AuthSignatureData{
				TaskId:          taskResult.TaskId,
				AvsAddress:      taskResult.AvsAddress,
				OperatorAddress: taskResult.OperatorAddress,
				OperatorSetId:   taskResult.OperatorSetId,
				ResultSigDigest: resultSigDigest,
			}
			authBytes := authData.ToSigningBytes()
			authSig, err := signer.SignMessage(authBytes)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to sign auth data: %v", err)
				cancel()
				return
			}
			taskResult.AuthSignature = authSig

			l.Sugar().Debugw("Auth signature created",
				"taskId", taskResult.TaskId,
				"authSigLength", len(authSig),
				"authSigHex", fmt.Sprintf("%x", authSig),
			)

			fmt.Printf("TaskResult %+v\n", taskResult)
			err = resultAgg.ProcessNewSignature(ctx, taskResult)
			assert.Nil(t, err)

			assert.True(t, resultAgg.SigningThresholdMet())

			cert, err := resultAgg.GenerateFinalCertificate()
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to generate final certificate: %v", err)
				cancel()
				return
			}
			// Use task's deadline or current time plus offset for signing time
			signedAt := time.Now().Add(10 * time.Second)
			if task.DeadlineUnixSeconds != nil {
				// Use a time before the deadline
				signedAt = task.DeadlineUnixSeconds.Add(-10 * time.Second)
			}
			cert.SignedAt = &signedAt
			fmt.Printf("cert: %+v\n", cert)

			time.Sleep(4 * time.Second)

			fmt.Printf("Submitting task result to AVS\n\n\n")
			receipt, err := avsCc.SubmitECDSATaskResult(ctx, cert.ToSubmitParams(), operatorPeersWeight.RootReferenceTimestamp)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to submit task result: %v", err)
				time.Sleep(time.Second * 300)
				cancel()
				return
			}
			assert.Nil(t, err)
			fmt.Printf("Receipt: %+v\n", receipt)

			cancel()
		}
	}()

	publishTaskCc := l1CC
	if networkTarget == NetworkTarget_L2 {
		publishTaskCc = l2CC
	}

	time.Sleep(10 * time.Second)

	// submit a task
	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))
	task, err := publishTaskCc.PublishMessageToInbox(ctx, chainConfig.AVSAccountAddress, 1, payloadJsonBytes)
	if err != nil {
		t.Errorf("Failed to publish message to inbox: %v", err)
	}
	t.Logf("Task published: %+v", task)

	select {
	case <-time.After(260 * time.Second):
		cancel()
		t.Errorf("Test timed out after 240 seconds")
	case <-ctx.Done():
		t.Logf("Test completed")
	}

	assert.False(t, hasErrors)

	_ = testUtils.KillAnvil(l1Anvil)
	if l2Anvil != nil {
		_ = testUtils.KillAnvil(l2Anvil)
	}
}

func Test_Mailbox(t *testing.T) {
	t.Run("BN254 & ECDSA - L1", func(t *testing.T) {
		testL1MailboxForCurve(t, config.CurveTypeECDSA, NetworkTarget_L1)
	})
	t.Run("BN254 & ECDSA - L2", func(t *testing.T) {
		testL1MailboxForCurve(t, config.CurveTypeECDSA, NetworkTarget_L2)
	})
}
