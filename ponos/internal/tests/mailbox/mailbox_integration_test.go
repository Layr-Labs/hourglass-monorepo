package mailbox

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/EVMChainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
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
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"math/big"
	"os/exec"
	"slices"
	"sync"
	"testing"
	"time"
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

	// aggregator is bn254, executor is ecdsa
	aggKeysBN254, _, _, err := testUtils.GetKeysForCurveType(t, config.CurveTypeBN254)
	if err != nil {
		t.Fatalf("Failed to get keys for curve type %s: %v", config.CurveTypeBN254, err)
	}

	_, execKeysECDSA, _, err := testUtils.GetKeysForCurveType(t, config.CurveTypeECDSA)
	if err != nil {
		t.Fatalf("Failed to get keys for curve type %s: %v", config.CurveTypeECDSA, err)
	}

	chainConfig, err := testUtils.ReadChainConfig(root)
	if err != nil {
		t.Fatalf("Failed to read chain config: %v", err)
	}

	coreContracts, err := eigenlayer.LoadContracts()
	if err != nil {
		t.Fatalf("Failed to load core contracts: %v", err)
	}

	imContractStore := inMemoryContractStore.NewInMemoryContractStore(coreContracts, l)

	if err = testUtils.ReplaceMailboxAddressWithTestAddress(imContractStore, chainConfig); err != nil {
		t.Fatalf("Failed to replace mailbox address with test address: %v", err)
	}

	tlp := transactionLogParser.NewTransactionLogParser(imContractStore, l)

	l1EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   L1RpcUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)
	l2EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   L2RpcUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)

	logsChan := make(chan *chainPoller.LogWithBlock)

	var pollerConfig *EVMChainPoller.EVMChainPollerConfig
	var pollerEthClient *ethereum.Client
	if networkTarget == NetworkTarget_L1 {
		pollerConfig = &EVMChainPoller.EVMChainPollerConfig{
			ChainId:              config.ChainId_EthereumAnvil,
			PollingInterval:      time.Duration(10) * time.Second,
			InterestingContracts: imContractStore.ListContractAddressesForChain(config.ChainId_EthereumAnvil),
		}
		pollerEthClient = l1EthereumClient
	} else {
		pollerConfig = &EVMChainPoller.EVMChainPollerConfig{
			ChainId:              config.ChainId_BaseSepoliaAnvil,
			PollingInterval:      time.Duration(10) * time.Second,
			InterestingContracts: imContractStore.ListContractAddressesForChain(config.ChainId_BaseSepoliaAnvil),
		}
		pollerEthClient = l2EthereumClient
	}

	poller := EVMChainPoller.NewEVMChainPoller(pollerEthClient, logsChan, tlp, pollerConfig, l)

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

	anvilCtx, anvilCancel := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
	defer anvilCancel()

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

	l1CC, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.AppAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress, // technically not used...
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL2,
	}, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create L2 contract caller: %v", err)
	}

	var l2CC *caller.ContractCaller
	if networkTarget == NetworkTarget_L2 {
		l2CC, err = caller.NewContractCaller(&caller.ContractCallerConfig{
			PrivateKey:          chainConfig.AppAccountPrivateKey,
			AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress, // technically not used...
			TaskMailboxAddress:  chainConfig.MailboxContractAddressL2,
		}, l2EthClient, l)
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

	time.Sleep(time.Second * 6)

	l.Sugar().Infow("------------------------ Transporting L1 tables ------------------------")

	testUtils.TransportStakeTables(l, networkTarget == NetworkTarget_L2)
	l.Sugar().Infow("Sleeping for 6 seconds to allow table transport to complete")
	time.Sleep(time.Second * 6)

	l.Sugar().Infow("------------------------ Setting up mailbox ------------------------")

	mailboxEthClient := l1EthClient
	mailboxContractAddress := chainConfig.MailboxContractAddressL1
	avsTaskHookAddress := chainConfig.AVSTaskHookAddressL1
	if networkTarget == NetworkTarget_L2 {
		mailboxEthClient = l2EthClient
		mailboxContractAddress = chainConfig.MailboxContractAddressL2
		avsTaskHookAddress = chainConfig.AVSTaskHookAddressL2
	}

	avsCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.AVSAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  mailboxContractAddress,
	}, mailboxEthClient, l)
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
		for logWithBlock := range logsChan {
			fmt.Printf("Received logWithBlock: %+v\n", logWithBlock.Log)
			if logWithBlock.Log.EventName != "TaskCreated" {
				continue
			}
			t.Logf("Found created task log: %+v", logWithBlock.Log)
			assert.Equal(t, "TaskCreated", logWithBlock.Log.EventName)

			task, err := types.NewTaskFromLog(logWithBlock.Log, logWithBlock.Block, chainConfig.MailboxContractAddressL1)
			assert.Nil(t, err)

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
				task.BlockNumber,
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
				task.BlockNumber,
				task.OperatorSetId,
				100,
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
			taskOutputHash := util.GetKeccak256Digest(outputResult)

			// for ecdsa only
			ecdsaDigest, err := avsCc.CalculateECDSACertificateDigest(
				ctx,
				operatorPeersWeight.RootReferenceTimestamp,
				taskOutputHash,
			)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to calculate ECDSA certificate digest: %v", err)
				cancel()
				return
			}

			taskResult := &types.TaskResult{
				TaskId:          task.TaskId,
				AvsAddress:      chainConfig.AVSAccountAddress,
				OperatorSetId:   task.OperatorSetId,
				Output:          outputResult,
				OperatorAddress: chainConfig.ExecOperatorAccountAddress,
				Signature:       nil,
				OutputDigest:    nil,
			}
			signer := inMemorySigner.NewInMemorySigner(execKeysECDSA.PrivateKey, config.CurveTypeECDSA)
			sig, err := signer.SignMessageForSolidity(ecdsaDigest)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to sign message: %v", err)
				cancel()
				return
			}
			taskResult.Signature = sig
			taskResult.OutputDigest = ecdsaDigest[:]

			fmt.Printf("TaskResult %+v\n", taskResult)
			err = resultAgg.ProcessNewSignature(ctx, task.TaskId, taskResult)
			assert.Nil(t, err)

			assert.True(t, resultAgg.SigningThresholdMet())

			cert, err := resultAgg.GenerateFinalCertificate()
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to generate final certificate: %v", err)
				cancel()
				return
			}
			signedAt := time.Unix(int64(logWithBlock.Block.Timestamp.Value()), 0).Add(10 * time.Second)
			cert.SignedAt = &signedAt
			fmt.Printf("cert: %+v\n", cert)

			time.Sleep(4 * time.Second)

			fmt.Printf("Submitting task result to AVS\n\n\n")
			receipt, err := avsCc.SubmitECDSATaskResult(ctx, cert, operatorPeersWeight.RootReferenceTimestamp)
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

	// submit a task
	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))
	task, err := publishTaskCc.PublishMessageToInbox(ctx, chainConfig.AVSAccountAddress, 1, payloadJsonBytes)
	if err != nil {
		t.Errorf("Failed to publish message to inbox: %v", err)
	}
	t.Logf("Task published: %+v", task)

	select {
	case <-time.After(240 * time.Second):
		cancel()
		t.Errorf("Test timed out after 10 seconds")
	case <-ctx.Done():
		t.Logf("Test completed")
	}

	assert.False(t, hasErrors)

	_ = l1Anvil.Process.Kill()
	if l2Anvil != nil {
		_ = l2Anvil.Process.Kill()
	}
}

func Test_L1Mailbox(t *testing.T) {
	t.Run("BN254 & ECDSA - L1", func(t *testing.T) {
		testL1MailboxForCurve(t, config.CurveTypeECDSA, NetworkTarget_L1)
	})
	t.Run("BN254 & ECDSA - L2", func(t *testing.T) {
		testL1MailboxForCurve(t, config.CurveTypeECDSA, NetworkTarget_L2)
	})
}
