package mailbox

import (
	"context"
	"fmt"
	"math/big"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/tableTransporter"
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
	NetworkTarget_L1   NetworkTarget = "l1"
	NetworkTarget_L2   NetworkTarget = "l2"
	L1RpcUrl                         = "http://127.0.0.1:8545"
	L2RpcUrl                         = "http://127.0.0.1:9545"
	maxStalenessPeriod               = 604800

	numExecutorOperators = 4
	transportBlsKey      = "0x5f8e6420b9cb0c940e3d3f8b99177980785906d16fb3571f70d7a05ecf5f2172"
)

func testL1MailboxForCurve(t *testing.T, curveType config.CurveType, networkTarget NetworkTarget) {
	if !slices.Contains([]config.CurveType{config.CurveTypeBN254, config.CurveTypeECDSA}, curveType) {
		t.Fatalf("Unsupported curve type: %s", curveType)
	}

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

	aggKey, execKeys, err := testUtils.GetKeysForCurveTypeFromChainConfig(
		t,
		config.CurveTypeBN254,
		config.CurveTypeECDSA,
		chainConfig,
	)
	if err != nil {
		t.Fatalf("Failed to get keys from chain config: %v", err)
	}
	execKeys = execKeys[:numExecutorOperators]

	// Create operator key map for lookup
	operatorKeyMap := make(map[string]*testUtils.WrappedKeyPair)
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperatorAccountAddress)] = execKeys[0]
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperator2AccountAddress)] = execKeys[1]
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperator3AccountAddress)] = execKeys[2]
	operatorKeyMap[strings.ToLower(chainConfig.ExecOperator4AccountAddress)] = execKeys[3]

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
			ChainId:              config.ChainId_BaseAnvil,
			PollingInterval:      time.Duration(10) * time.Second,
			InterestingContracts: imContractStore.ListContractAddressesForChain(config.ChainId_BaseAnvil),
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

	if networkTarget == NetworkTarget_L2 {
		testUtils.SyncL2TimestampWithL1(t, l1EthClient, l2EthClient)
	}

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

	t.Logf("------------------------------------------- Setting up operator peering -------------------------------------------")
	// NOTE: we must register ALL opsets regardless of which curve type we are using, otherwise table transport fails

	aggregator := &operator.Operator{
		TransactionPrivateKey: chainConfig.OperatorAccountPrivateKey,
		SigningPrivateKey:     aggKey.PrivateKey,
		Curve:                 config.CurveTypeBN254,
		OperatorSetIds:        []uint32{aggOpsetId},
	}

	executors := []testUtils.ExecutorWithSocket{
		{
			Executor: &operator.Operator{
				TransactionPrivateKey: chainConfig.ExecOperatorAccountPk,
				SigningPrivateKey:     execKeys[0].PrivateKey,
				Curve:                 config.CurveTypeECDSA,
				OperatorSetIds:        []uint32{execOpsetId},
			},
			Socket: fmt.Sprintf("localhost:%d", 9001),
		},
		{
			Executor: &operator.Operator{
				TransactionPrivateKey: chainConfig.ExecOperator2AccountPk,
				SigningPrivateKey:     execKeys[1].PrivateKey,
				Curve:                 config.CurveTypeECDSA,
				OperatorSetIds:        []uint32{execOpsetId},
			},
			Socket: fmt.Sprintf("localhost:%d", 9002),
		},
		{
			Executor: &operator.Operator{
				TransactionPrivateKey: chainConfig.ExecOperator3AccountPk,
				SigningPrivateKey:     execKeys[2].PrivateKey,
				Curve:                 config.CurveTypeECDSA,
				OperatorSetIds:        []uint32{execOpsetId},
			},
			Socket: fmt.Sprintf("localhost:%d", 9003),
		},
		{
			Executor: &operator.Operator{
				TransactionPrivateKey: chainConfig.ExecOperator4AccountPk,
				SigningPrivateKey:     execKeys[3].PrivateKey,
				Curve:                 config.CurveTypeECDSA,
				OperatorSetIds:        []uint32{execOpsetId},
			},
			Socket: fmt.Sprintf("localhost:%d", 9004),
		},
	}

	err = testUtils.SetupOperatorPeeringWithMultipleExecutors(
		ctx,
		chainConfig,
		config.ChainId_EthereumAnvil,
		l1EthClient,
		aggregator,
		executors,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to setup operator peering: %v", err)
	}

	time.Sleep(time.Second * 6)

	// Delegate stake to aggregator + 4 executors with different weights
	stakeConfigs := []*testUtils.StakerDelegationConfig{
		{
			StakerPrivateKey:   chainConfig.AggStakerAccountPrivateKey,
			StakerAddress:      chainConfig.AggStakerAccountAddress,
			OperatorPrivateKey: chainConfig.OperatorAccountPrivateKey,
			OperatorAddress:    chainConfig.OperatorAccountAddress,
			OperatorSetId:      aggOpsetId,
			StrategyAddress:    testUtils.Strategy_WETH,
		},
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

	t.Logf("All operator set IDs: %v", allOperatorSetIds)
	l.Sugar().Infow("Waiting for stake delegations to be processed on-chain...")
	time.Sleep(time.Second * 3)

	bn254CalculatorAddr, err := caller.GetTableCalculatorAddress(config.CurveTypeBN254, config.ChainId_EthereumAnvil)
	if err != nil {
		t.Fatalf("Failed to get BN254 calculator address: %v", err)
	}
	t.Logf(
		"Creating generation reservation with BN254 table calculator %s for aggregator operator set %d",
		bn254CalculatorAddr.Hex(),
		aggOpsetId,
	)

	_, err = avsConfigCaller.CreateGenerationReservation(
		ctx,
		avsAddr,
		aggOpsetId,
		bn254CalculatorAddr,
		avsAddr,
		maxStalenessPeriod,
	)
	if err != nil {
		t.Logf("Warning: Failed to create generation reservation for aggregator: %v", err)
	}

	ecdsaCalculatorAddr, err := caller.GetTableCalculatorAddress(config.CurveTypeECDSA, config.ChainId_EthereumAnvil)
	if err != nil {
		t.Fatalf("Failed to get ECDSA calculator address: %v", err)
	}
	t.Logf(
		"Creating generation reservation with ECDSA table calculator %s for executor operator set %d",
		ecdsaCalculatorAddr.Hex(),
		execOpsetId,
	)

	_, err = avsConfigCaller.CreateGenerationReservation(
		ctx,
		avsAddr,
		execOpsetId,
		ecdsaCalculatorAddr,
		avsAddr,
		maxStalenessPeriod,
	)
	if err != nil {
		t.Logf("Warning: Failed to create generation reservation: %v", err)
	}

	time.Sleep(time.Second * 3)

	l.Sugar().Infow("------------------------ Transporting L1 tables ------------------------")

	operatorAddressList := []string{
		chainConfig.ExecOperatorAccountAddress,
		chainConfig.ExecOperator2AccountAddress,
		chainConfig.ExecOperator3AccountAddress,
		chainConfig.ExecOperator4AccountAddress,
	}

	stakeWeights := []*big.Int{
		big.NewInt(400000000000000000), // 0.4e18 = 40%
		big.NewInt(300000000000000000), // 0.3e18 = 30%
		big.NewInt(200000000000000000), // 0.2e18 = 20%
		big.NewInt(100000000000000000), // 0.1e18 = 10%
	}

	operatorPrivateKeys := []string{
		chainConfig.ExecOperatorAccountPk,
		chainConfig.ExecOperator2AccountPk,
		chainConfig.ExecOperator3AccountPk,
		chainConfig.ExecOperator4AccountPk,
	}

	operatorKeyInfos := make([]tableTransporter.OperatorKeyInfo, numExecutorOperators)
	for i := 0; i < numExecutorOperators; i++ {
		operatorKeyInfos[i] = tableTransporter.OperatorKeyInfo{
			PrivateKeyHex:   operatorPrivateKeys[i],
			Weights:         []*big.Int{stakeWeights[i]},
			OperatorAddress: common.HexToAddress(operatorAddressList[i]),
		}
	}

	// Determine chains to ignore and L2 config based on network target
	var l2RpcUrl string
	var l2ChainId uint64
	chainIdsToIgnore := []*big.Int{
		big.NewInt(1),
		big.NewInt(8453),
	}

	if networkTarget == NetworkTarget_L1 {
		chainIdsToIgnore = append(chainIdsToIgnore, big.NewInt(31338))
	} else {
		l2RpcUrl = L2RpcUrl
		l2ChainId = 31338
	}

	transportConfig := &tableTransporter.MultipleOperatorConfig{
		TransporterPrivateKey:     chainConfig.AVSAccountPrivateKey,
		L1RpcUrl:                  L1RpcUrl,
		L1ChainId:                 31337,
		L2RpcUrl:                  l2RpcUrl,
		L2ChainId:                 l2ChainId,
		CrossChainRegistryAddress: eigenlayerContractAddrs.CrossChainRegistry,
		ChainIdsToIgnore:          chainIdsToIgnore,
		Logger:                    l,
		Operators:                 operatorKeyInfos,
		AVSAddress:                common.HexToAddress(chainConfig.AVSAccountAddress),
		OperatorSetId:             execOpsetId,
		CurveType:                 config.CurveTypeECDSA,
		TransportBLSPrivateKey:    transportBlsKey,
	}

	err = tableTransporter.TransportTableWithSimpleMultiOperators(transportConfig)
	if err != nil {
		t.Fatalf("Failed to transport stake tables: %v", err)
	}

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
		callerMap[config.ChainId_BaseAnvil] = l2CC
		opManagerChainIds = append(opManagerChainIds, config.ChainId_BaseAnvil)
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

			var operators []*aggregation.Operator[common.Address]
			for _, peer := range operatorPeersWeight.Operators {
				opset, err := peer.GetOperatorSet(task.OperatorSetId)
				if err != nil {
					hasErrors = true
					l.Sugar().Errorf("Failed to get operator set for peer %s: %v", peer.OperatorAddress, err)
					cancel()
					return
				}

				operators = append(operators, &aggregation.Operator[common.Address]{
					Address:       peer.OperatorAddress,
					PublicKey:     opset.WrappedPublicKey.ECDSAAddress,
					OperatorIndex: opset.OperatorIndex,
					Weights:       operatorPeersWeight.Weights[peer.OperatorAddress],
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
				10_000,
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

			// Define all executor operator addresses
			executorOperatorAddresses := []string{
				chainConfig.ExecOperatorAccountAddress,
				chainConfig.ExecOperator2AccountAddress,
				chainConfig.ExecOperator3AccountAddress,
				chainConfig.ExecOperator4AccountAddress,
			}

			// Calculate message hash once (same for all operators)
			var taskIdBytes [32]byte
			copy(taskIdBytes[:], common.HexToHash(task.TaskId).Bytes())
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

			// Process signatures from all 4 executor operators
			for i, operatorAddress := range executorOperatorAddresses {
				taskResult := &types.TaskResult{
					TaskId:          task.TaskId,
					AvsAddress:      chainConfig.AVSAccountAddress,
					OperatorSetId:   task.OperatorSetId,
					Output:          outputResult,
					OperatorAddress: operatorAddress,
					ResultSignature: nil,
					AuthSignature:   nil,
				}

				// Get the operator's key pair
				operatorKeys := operatorKeyMap[strings.ToLower(operatorAddress)]
				if operatorKeys == nil {
					hasErrors = true
					l.Sugar().Errorf("Could not find keys for operator %s", operatorAddress)
					cancel()
					return
				}

				signer := inMemorySigner.NewInMemorySigner(operatorKeys.PrivateKey, config.CurveTypeECDSA)

				// Log the actual signing address derived from the private key
				if ecdsaPK, ok := operatorKeys.PrivateKey.(*cryptoLibsEcdsa.PrivateKey); ok {
					signerAddress, _ := ecdsaPK.DeriveAddress()
					t.Logf("Operator %d - Signer address derived from private key: %s", i+1, signerAddress.Hex())
				}

				l.Sugar().Debugw("Signing result for operator",
					"operatorIndex", i+1,
					"taskId", taskResult.TaskId,
					"operatorAddress", taskResult.OperatorAddress,
					"outputLength", len(outputResult),
					"outputDigest", fmt.Sprintf("%x", messageHash),
					"certificateDigest", fmt.Sprintf("%x", certificateDigestBytes),
					"referenceTimestamp", operatorPeersWeight.RootReferenceTimestamp,
				)

				// Step 1: Sign the result using certificate digest
				resultSig, err := signer.SignMessageForSolidity(certificateDigestBytes)
				if err != nil {
					hasErrors = true
					l.Sugar().Errorf("Failed to sign result for operator %d: %v", i+1, err)
					cancel()
					return
				}
				taskResult.ResultSignature = resultSig

				l.Sugar().Debugw("Result signature created",
					"operatorIndex", i+1,
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
					l.Sugar().Errorf("Failed to sign auth data for operator %d: %v", i+1, err)
					cancel()
					return
				}
				taskResult.AuthSignature = authSig

				l.Sugar().Debugw("Auth signature created",
					"operatorIndex", i+1,
					"taskId", taskResult.TaskId,
					"authSigLength", len(authSig),
					"authSigHex", fmt.Sprintf("%x", authSig),
				)

				t.Logf("Processing signature from operator %d: %s", i+1, operatorAddress)
				fmt.Printf("TaskResult from operator %d: %+v\n", i+1, taskResult)

				err = resultAgg.ProcessNewSignature(ctx, taskResult)
				assert.Nil(t, err)

				// Check threshold after processing each signature
				thresholdMet := resultAgg.SigningThresholdMet()
				t.Logf("After operator %d: Signing threshold met = %v", i+1, thresholdMet)
			}

			// After all operators have signed, the threshold should be met
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
