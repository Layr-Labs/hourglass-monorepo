package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/tableTransporter"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	executorMemory "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"

	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	commonV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/common"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/avsExecutionManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	goEthereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	L1RPCUrl        = "http://127.0.0.1:8545"
	L2RPCUrl        = "http://127.0.0.1:9545"
	L2WSUrl         = "ws://127.0.0.1:9545"
	L1WsUrl         = "ws://127.0.0.1:8545"
	transportBlsKey = "0x5f8e6420b9cb0c940e3d3f8b99177980785906d16fb3571f70d7a05ecf5f2172"
)

// SignatureModeConfig defines the signature curve configuration for testing
type SignatureModeConfig struct {
	Name              string
	AggregatorCurve   config.CurveType
	ExecutorCurve     config.CurveType
	AggregatorOpsetId uint32
	ExecutorOpsetId   uint32
}

// Test_Aggregator is an integration test for the Aggregator component of the system.
//
// This test is designed to simulate an E2E on-chain flow with all components.
// - Both the aggregator and executor are registered as operators with the AllocationManager/AVSRegistrar with their peering data
// - The executor is started and boots up the performers
// - The aggregator is started with a poller calling a local anvil node
// - The test pushes a message to the mailbox and waits for the TaskVerified event to be emitted
func Test_Aggregator(t *testing.T) {
	// Define test configurations for both signature modes
	signatureConfigs := []SignatureModeConfig{
		{
			Name:              "BN254_Aggregator_ECDSA_Executor",
			AggregatorCurve:   config.CurveTypeBN254,
			ExecutorCurve:     config.CurveTypeECDSA,
			AggregatorOpsetId: 0,
			ExecutorOpsetId:   1,
		},
		{
			Name:              "ECDSA_Aggregator_BN254_Executor",
			AggregatorCurve:   config.CurveTypeECDSA,
			ExecutorCurve:     config.CurveTypeBN254,
			AggregatorOpsetId: 0,
			ExecutorOpsetId:   1,
		},
	}

	for _, sigConfig := range signatureConfigs {
		t.Run(sigConfig.Name, func(t *testing.T) {
			t.Run("Docker", func(t *testing.T) {
				runAggregatorTest(t, "docker", sigConfig)
			})

			t.Run("Kubernetes", func(t *testing.T) {
				runAggregatorTest(t, "kubernetes", sigConfig)
			})
		})
	}
}

func runAggregatorTest(t *testing.T, mode string, sigConfig SignatureModeConfig) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	root := testUtils.GetProjectRootPath()
	t.Logf("Project root path: %s", root)

	chainConfig, err := testUtils.ReadChainConfig(root)
	if err != nil {
		t.Fatalf("Failed to read chain config: %v", err)
	}

	var cluster *testUtils.KindCluster
	if mode == "kubernetes" {

		if err := testUtils.CleanupAllTestClusters(l.Sugar()); err != nil {
			t.Logf("Warning: Failed to cleanup existing test clusters: %v", err)
		}

		kindConfig := testUtils.DefaultKindClusterConfig(l.Sugar())
		var clusterCleanup func()
		cluster, clusterCleanup, err = testUtils.CreateKindCluster(ctx, t, kindConfig)
		if err != nil {
			t.Fatalf("Failed to create Kind cluster: %v", err)
		}
		defer func() {
			t.Log("Using fast cluster deletion to avoid hanging cleanup")
			clusterCleanup()
		}()

		if err := loadPerformerImage(ctx, cluster, l.Sugar()); err != nil {
			t.Fatalf("Failed to load performer image: %v", err)
		}

		if err := installPerformerCRD(ctx, cluster, root, l.Sugar()); err != nil {
			t.Fatalf("Failed to install Performer CRD: %v", err)
		}

		if err := loadOperatorImage(ctx, cluster, l.Sugar()); err != nil {
			t.Fatalf("Failed to load operator image: %v", err)
		}

		// Deploy Hourglass operator
		operatorConfig := testUtils.DefaultOperatorDeploymentConfig(root, l.Sugar())
		hgOperator, operatorCleanup, err := testUtils.DeployOperator(ctx, cluster, operatorConfig)
		if err != nil {
			t.Fatalf("Failed to deploy operator: %v", err)
		}

		defer func() {
			t.Log("Running operator cleanup with timeout")
			done := make(chan struct{})
			go func() {
				operatorCleanup()
				close(done)
			}()

			select {
			case <-done:
				t.Log("Operator cleanup completed successfully")
			case <-time.After(45 * time.Second):
				t.Log("Operator cleanup timed out, proceeding with cluster deletion")
			}
		}()

		t.Logf("Operator deployed successfully: %s", hgOperator.ReleaseName)

		if err := createPerformerNodePortService(ctx, cluster, l.Sugar()); err != nil {
			t.Fatalf("Failed to create NodePort service: %v", err)
		}
	}

	execConfig, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(getExecutorConfigYaml(mode, sigConfig.ExecutorCurve)))
	if err != nil {
		t.Fatalf("failed to create executor config: %v", err)
	}
	if err := execConfig.Validate(); err != nil {
		t.Fatalf("failed to validate executor config: %v", err)
	}

	if sigConfig.ExecutorCurve == config.CurveTypeECDSA {
		execConfig.Operator.SigningKeys.ECDSA = &config.ECDSAKeyConfig{
			PrivateKey: chainConfig.ExecOperatorAccountPk,
		}
	}

	execConfig.Operator.OperatorPrivateKey = &config.ECDSAKeyConfig{
		PrivateKey: chainConfig.ExecOperatorAccountPk,
	}
	execConfig.Operator.Address = chainConfig.ExecOperatorAccountAddress
	execConfig.AvsPerformers[0].AvsAddress = chainConfig.AVSAccountAddress

	if mode == "kubernetes" {
		execConfig.Kubernetes.KubeConfigPath = cluster.KubeConfig
	}

	numExecutorOperators := 1
	if mode == "docker" {
		numExecutorOperators = 2
	}
	t.Logf("------------------------------------------- Setting up %d executor operator(s) for %s mode -------------------------------------------", numExecutorOperators, mode)

	aggKey, execKeys, err := testUtils.GetKeysForCurveTypeFromChainConfig(t, sigConfig.AggregatorCurve, sigConfig.ExecutorCurve, chainConfig)
	if err != nil {
		t.Fatalf("Failed to get keys: %v", err)
	}

	execSigner := inMemorySigner.NewInMemorySigner(execKeys[0].PrivateKey, sigConfig.ExecutorCurve)

	if len(execKeys) < numExecutorOperators {
		t.Fatalf("Expected at least %d executor keys, got %d", numExecutorOperators, len(execKeys))
	}

	execKeys = execKeys[:numExecutorOperators]

	aggConfigYaml := getAggregatorConfigYaml(L1RPCUrl, L2RPCUrl, sigConfig.AggregatorCurve)

	aggConfig, err := aggregatorConfig.NewAggregatorConfigFromYamlBytes([]byte(aggConfigYaml))
	if err != nil {
		t.Fatalf("Failed to create aggregator config: %v", err)
	}

	aggConfig.Operator.Address = chainConfig.OperatorAccountAddress
	aggConfig.Operator.OperatorPrivateKey = &config.ECDSAKeyConfig{
		PrivateKey: chainConfig.OperatorAccountPrivateKey,
	}

	if sigConfig.AggregatorCurve == config.CurveTypeECDSA {
		aggConfig.Operator.SigningKeys.ECDSA = &config.ECDSAKeyConfig{
			PrivateKey: chainConfig.OperatorAccountPrivateKey,
		}
		aggConfig.Operator.SigningKeys.BLS = nil
	} else {
		aggConfig.Operator.SigningKeys.ECDSA = nil
	}

	aggSigner := inMemorySigner.NewInMemorySigner(aggKey.PrivateKey, sigConfig.AggregatorCurve)

	anvilWg := &sync.WaitGroup{}

	l1RpcUrl := L1RPCUrl
	l2RpcUrl := L2RPCUrl
	l2WsUrl := L2WSUrl

	l1EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   l1RpcUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)
	l1EthClient, err := l1EthereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get Ethereum contract caller: %v", err)
	}

	l2EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   l2RpcUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)
	l2EthClient, err := l2EthereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get Ethereum contract caller: %v", err)
	}

	_ = testUtils.KillallAnvils()

	startErrorsChan := make(chan error, 2)
	anvilCtx, anvilCancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer anvilCancel()

	anvilWg.Add(1)
	l1Anvil, err := testUtils.StartL1Anvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start Anvil: %v", err)
	}
	go testUtils.WaitForAnvil(anvilWg, anvilCtx, t, l1EthereumClient, startErrorsChan)

	anvilWg.Add(1)
	l2Anvil, err := testUtils.StartL2Anvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start L2 Anvil: %v", err)
	}
	go testUtils.WaitForAnvil(anvilWg, anvilCtx, t, l2EthereumClient, startErrorsChan)

	anvilWg.Wait()
	close(startErrorsChan)
	for err := range startErrorsChan {
		if err != nil {
			anvilCancel()
			t.Fatalf("Failed to start Anvil: %v", err)
		}
	}
	anvilCancel()

	testUtils.SyncL2TimestampWithL1(t, l1EthClient, l2EthClient)

	l1ChainId, err := l1EthClient.ChainID(ctx)
	if err != nil {
		t.Fatalf("failed to get L1 chain ID: %v", err)
	}

	l2ChainId, err := l2EthClient.ChainID(ctx)
	if err != nil {
		t.Fatalf("Failed to get L2 chain ID: %v", err)
	}
	t.Logf("L2 Chain ID: %s", l2ChainId.String())

	eigenlayerContractAddrs, err := config.GetCoreContractsForChainId(config.ChainId(l1ChainId.Uint64()))
	if err != nil {
		t.Fatalf("Failed to get core contracts for chain ID: %v", err)
	}

	l2EigenlayerContractAddrs, err := config.GetCoreContractsForChainId(config.ChainId(l2ChainId.Uint64()))
	if err != nil {
		t.Fatalf("Failed to get core contracts for chain ID: %v", err)
	}

	l1AggPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.OperatorAccountPrivateKey, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create L1 aggregator private key bn254Signer: %v", err)
	}

	l1AggCc, err := caller.NewContractCaller(l1EthClient, l1AggPrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create contract caller: %v", err)
	}

	l1ExecPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.ExecOperatorAccountPk, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create L1 executor private key bn254Signer: %v", err)
	}

	l1ExecCc, err := caller.NewContractCaller(l1EthClient, l1ExecPrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create contract caller: %v", err)
	}

	reservations, err := l1AggCc.GetActiveGenerationReservations()
	if err != nil {
		t.Fatalf("Failed to get active generation reservations: %v", err)
	}
	for _, reservation := range reservations {
		fmt.Printf("Active generation reservation: %+v\n", reservation)
	}

	l.Sugar().Infow("Setting up operator peering",
		zap.String("AVSAccountAddress", chainConfig.AVSAccountAddress),
	)

	t.Logf("------------------------------------------- Configuring operator sets dynamically -------------------------------------------")

	avsConfigPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create AVS config private key signer: %v", err)
	}

	avsConfigCaller, err := caller.NewContractCaller(l1EthClient, avsConfigPrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create AVS config caller: %v", err)
	}

	t.Logf("Configuring operator set %d with curve type %s for aggregator", sigConfig.AggregatorOpsetId, sigConfig.AggregatorCurve)
	_, err = avsConfigCaller.ConfigureAVSOperatorSet(ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		sigConfig.AggregatorOpsetId,
		sigConfig.AggregatorCurve)
	if err != nil {
		t.Fatalf("Failed to configure aggregator operator set %d: %v", sigConfig.AggregatorOpsetId, err)
	}

	t.Logf("Configuring operator set %d with curve type %s for executor", sigConfig.ExecutorOpsetId, sigConfig.ExecutorCurve)
	_, err = avsConfigCaller.ConfigureAVSOperatorSet(ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		sigConfig.ExecutorOpsetId,
		sigConfig.ExecutorCurve)
	if err != nil {
		t.Fatalf("Failed to configure executor operator set %d: %v", sigConfig.ExecutorOpsetId, err)
	}

	t.Logf("Successfully configured operator sets dynamically")

	if sigConfig.AggregatorCurve == config.CurveTypeBN254 && aggConfig.Operator.SigningKeys.BLS == nil {
		t.Fatalf("BLS signing key required for BN254 curve but not configured for aggregator")
	}
	if sigConfig.AggregatorCurve == config.CurveTypeECDSA && aggConfig.Operator.SigningKeys.ECDSA == nil {
		t.Fatalf("ECDSA signing key required but not configured for aggregator")
	}

	if sigConfig.ExecutorCurve == config.CurveTypeBN254 && execConfig.Operator.SigningKeys.BLS == nil {
		t.Fatalf("BLS signing key required for BN254 curve but not configured for executor")
	}
	if sigConfig.ExecutorCurve == config.CurveTypeECDSA && execConfig.Operator.SigningKeys.ECDSA == nil {
		t.Fatalf("ECDSA signing key required but not configured for executor")
	}

	t.Logf("Operator signing keys validated successfully")

	t.Logf("------------------------------------------- Registering aggregator and executor operators -------------------------------------------")
	allOperatorSetIds := []uint32{sigConfig.AggregatorOpsetId, sigConfig.ExecutorOpsetId}

	aggSigningKey := aggKey.PrivateKey

	operatorPkList := []string{chainConfig.ExecOperatorAccountPk, chainConfig.ExecOperator2AccountPk}

	executorsWithSockets := make([]testUtils.ExecutorWithSocket, numExecutorOperators)
	for i := 0; i < numExecutorOperators; i++ {
		executorsWithSockets[i] = testUtils.ExecutorWithSocket{
			Executor: &operator.Operator{
				TransactionPrivateKey: operatorPkList[i],
				SigningPrivateKey:     execKeys[i].PrivateKey,
				Curve:                 sigConfig.ExecutorCurve,
				OperatorSetIds:        []uint32{sigConfig.ExecutorOpsetId},
			},
			Socket: fmt.Sprintf("localhost:%d", 9000+i),
		}
	}

	err = testUtils.SetupOperatorPeeringWithMultipleExecutors(
		ctx,
		chainConfig,
		config.ChainId(l1ChainId.Uint64()),
		l1EthClient,
		&operator.Operator{
			TransactionPrivateKey: chainConfig.OperatorAccountPrivateKey,
			SigningPrivateKey:     aggSigningKey,
			Curve:                 sigConfig.AggregatorCurve,
			OperatorSetIds:        []uint32{sigConfig.AggregatorOpsetId},
		},
		executorsWithSockets,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to set up operator peering: %v", err)
	}

	time.Sleep(time.Second * 6)

	t.Logf("------------------------------------------- Delegating stake to operators -------------------------------------------")

	t.Logf("Delegating stake to aggregator operator (operator set %d)", sigConfig.AggregatorOpsetId)
	aggStakeConfig := []*testUtils.StakerDelegationConfig{
		{
			StakerPrivateKey:   chainConfig.AggStakerAccountPrivateKey,
			StakerAddress:      chainConfig.AggStakerAccountAddress,
			OperatorPrivateKey: chainConfig.OperatorAccountPrivateKey,
			OperatorAddress:    chainConfig.OperatorAccountAddress,
			OperatorSetId:      sigConfig.AggregatorOpsetId,
			StrategyAddress:    testUtils.Strategy_WETH,
		},
	}

	err = testUtils.DelegateStakeToMultipleOperators(
		t,
		ctx,
		aggStakeConfig,
		chainConfig.AVSAccountAddress,
		l1EthClient,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to delegate stake to aggregator operator: %v", err)
	}

	t.Logf("Delegating stake to executor operators (operator set %d)", sigConfig.ExecutorOpsetId)
	stakerPkList := []string{chainConfig.ExecStakerAccountPrivateKey, chainConfig.ExecStaker2AccountPrivateKey}
	stakerAddrList := []string{chainConfig.ExecStakerAccountAddress, chainConfig.ExecStaker2AccountAddress}
	operatorAddrList := []string{chainConfig.ExecOperatorAccountAddress, chainConfig.ExecOperator2AccountAddress}

	stakeConfigs := make([]*testUtils.StakerDelegationConfig, numExecutorOperators)
	for i := 0; i < numExecutorOperators; i++ {
		stakeConfigs[i] = &testUtils.StakerDelegationConfig{
			StakerPrivateKey:   stakerPkList[i],
			StakerAddress:      stakerAddrList[i],
			OperatorPrivateKey: operatorPkList[i],
			OperatorAddress:    operatorAddrList[i],
			OperatorSetId:      sigConfig.ExecutorOpsetId,
			StrategyAddress:    testUtils.Strategy_STETH,
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
		t.Fatalf("Failed to delegate stake to executor operators: %v", err)
	}

	t.Logf("All operator set IDs: %v", allOperatorSetIds)

	t.Logf("------------------------ Creating generation reservations ------------------------")

	avsCcL1TempSigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create AVS L1 private key signer for generation reservation setup: %v", err)
	}

	avsCcL1Temp, err := caller.NewContractCaller(l1EthClient, avsCcL1TempSigner, l)
	if err != nil {
		t.Fatalf("Failed to create AVS contract caller for generation reservation setup: %v", err)
	}

	avsAddr := common.HexToAddress(chainConfig.AVSAccountAddress)
	maxStalenessPeriod := uint32(604800)

	aggCalculatorAddr, err := caller.GetTableCalculatorAddress(sigConfig.AggregatorCurve, config.ChainId_EthereumAnvil)
	if err != nil {
		t.Fatalf("Failed to delegate stake to operators: %v", err)
	}

	t.Logf("Creating generation reservation with %s table calculator %s for aggregator operator set %d",
		sigConfig.AggregatorCurve, aggCalculatorAddr.Hex(), sigConfig.AggregatorOpsetId)
	_, err = avsCcL1Temp.CreateGenerationReservation(
		ctx,
		avsAddr,
		sigConfig.AggregatorOpsetId,
		aggCalculatorAddr,
		avsAddr,
		maxStalenessPeriod,
	)
	if err != nil {
		t.Logf("Warning: Failed to create generation reservation for aggregator operator set %d: %v", sigConfig.AggregatorOpsetId, err)
	}

	execCalculatorAddr, err := caller.GetTableCalculatorAddress(sigConfig.ExecutorCurve, config.ChainId_EthereumAnvil)
	if err != nil {
		t.Fatalf("Failed to delegate stake to operators: %v", err)
	}

	t.Logf("Creating generation reservation with %s table calculator %s for executor operator set %d",
		sigConfig.ExecutorCurve, execCalculatorAddr.Hex(), sigConfig.ExecutorOpsetId)
	_, err = avsCcL1Temp.CreateGenerationReservation(
		ctx,
		avsAddr,
		sigConfig.ExecutorOpsetId,
		execCalculatorAddr,
		avsAddr,
		maxStalenessPeriod,
	)
	if err != nil {
		t.Logf("Warning: Failed to create generation reservation for executor operator set %d: %v", sigConfig.ExecutorOpsetId, err)
	}

	time.Sleep(time.Second * 3)

	currentBlock, err := l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}
	testUtils.DebugOpsetData(t, chainConfig, eigenlayerContractAddrs, l1EthClient, currentBlock, allOperatorSetIds)

	time.Sleep(time.Second * 6)

	l.Sugar().Infow("------------------------ Transporting L1 & L2 tables ------------------------")

	chainIdsToIgnore := []*big.Int{
		new(big.Int).SetUint64(1),
		new(big.Int).SetUint64(8453),
	}

	stakeWeights := []*big.Int{
		big.NewInt(600000000000000000), // 0.6e18 = 60%
		big.NewInt(400000000000000000), // 0.4e18 = 40%
	}

	operatorAddressList := []string{
		chainConfig.ExecOperatorAccountAddress,
		chainConfig.ExecOperator2AccountAddress,
	}

	operatorKeyInfos := make([]tableTransporter.OperatorKeyInfo, numExecutorOperators)
	for i := 0; i < numExecutorOperators; i++ {
		var privateKeyHex string
		if sigConfig.ExecutorCurve == config.CurveTypeBN254 {
			blsPrivKey := execKeys[i].PrivateKey.(*bn254.PrivateKey)
			privateKeyHex = fmt.Sprintf("0x%x", blsPrivKey.Bytes())
		} else {
			ecdsaPrivKey := execKeys[i].PrivateKey.(*cryptoLibsEcdsa.PrivateKey)
			privateKeyHex = fmt.Sprintf("0x%x", ecdsaPrivKey.Bytes())
		}

		operatorKeyInfos[i] = tableTransporter.OperatorKeyInfo{
			OperatorAddress: common.HexToAddress(operatorAddressList[i]),
			PrivateKeyHex:   privateKeyHex,
			Weights:         []*big.Int{stakeWeights[i]},
		}
	}

	contractAddresses := config.CoreContracts[config.ChainId_EthereumAnvil]

	err = tableTransporter.TransportTableWithSimpleMultiOperators(
		&tableTransporter.MultipleOperatorConfig{
			TransporterPrivateKey:     chainConfig.AVSAccountPrivateKey,
			L1RpcUrl:                  "http://localhost:8545",
			L1ChainId:                 31337,
			L2RpcUrl:                  l2RpcUrl,
			L2ChainId:                 31338,
			CrossChainRegistryAddress: contractAddresses.CrossChainRegistry,
			ChainIdsToIgnore:          chainIdsToIgnore,
			Logger:                    l,
			Operators:                 operatorKeyInfos,
			AVSAddress:                avsAddr,
			OperatorSetId:             sigConfig.ExecutorOpsetId,
			TransportBLSPrivateKey:    transportBlsKey,
			CurveType:                 sigConfig.ExecutorCurve,
		},
	)
	if err != nil {
		t.Fatalf("Failed to transport operator tables: %v", err)
	}
	l.Sugar().Infow("Sleeping for 6 seconds to allow table transport to complete")
	time.Sleep(time.Second * 6)

	l.Sugar().Infow("------------------------ Setting up mailbox ------------------------")

	avsCcL1PrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create AVS L1 private key bn254Signer: %v", err)
	}

	avsCcL1, err := caller.NewContractCaller(l1EthClient, avsCcL1PrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create AVS contract caller: %v", err)
	}
	err = testUtils.SetupTaskMailbox(
		ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		common.HexToAddress(chainConfig.AVSTaskHookAddressL1),
		[]uint32{sigConfig.ExecutorOpsetId},
		[]config.CurveType{sigConfig.ExecutorCurve},
		avsCcL1,
	)
	if err != nil {
		t.Fatalf("Failed to set up task mailbox: %v", err)
	}

	avsCcL2PrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, l2EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create AVS L2 private key bn254Signer: %v", err)
	}

	avsCcL2, err := caller.NewContractCaller(l2EthClient, avsCcL2PrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create AVS contract caller: %v", err)
	}
	err = testUtils.SetupTaskMailbox(
		ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		common.HexToAddress(chainConfig.AVSTaskHookAddressL2),
		[]uint32{sigConfig.ExecutorOpsetId},
		[]config.CurveType{sigConfig.ExecutorCurve},
		avsCcL2,
	)
	if err != nil {
		t.Fatalf("Failed to set up task mailbox: %v", err)
	}

	currentBlock, err = l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}
	l.Sugar().Infow("Current block number", zap.Uint64("blockNumber", currentBlock))
	testUtils.DebugOpsetData(t, chainConfig, eigenlayerContractAddrs, l1EthClient, currentBlock, allOperatorSetIds)

	var executors []*executor.Executor

	if mode == "docker" {
		t.Logf("Creating 2 executor instances for Docker mode")

		for i := 0; i < numExecutorOperators; i++ {
			execConfigForOp, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(getExecutorConfigYaml(mode, sigConfig.ExecutorCurve)))
			if err != nil {
				t.Fatalf("Failed to create executor config for operator %d: %v", i, err)
			}

			operatorPkList := []string{chainConfig.ExecOperatorAccountPk, chainConfig.ExecOperator2AccountPk}
			operatorAddrList := []string{chainConfig.ExecOperatorAccountAddress, chainConfig.ExecOperator2AccountAddress}

			execConfigForOp.Operator.OperatorPrivateKey = &config.ECDSAKeyConfig{
				PrivateKey: operatorPkList[i],
			}
			execConfigForOp.Operator.Address = operatorAddrList[i]
			execConfigForOp.AvsPerformers[0].AvsAddress = chainConfig.AVSAccountAddress
			execConfigForOp.GrpcPort = 9000 + i

			if sigConfig.ExecutorCurve == config.CurveTypeECDSA {
				execConfigForOp.Operator.SigningKeys.ECDSA = &config.ECDSAKeyConfig{
					PrivateKey: operatorPkList[i],
				}
			}

			execSignerForOp := inMemorySigner.NewInMemorySigner(execKeys[i].PrivateKey, sigConfig.ExecutorCurve)
			execPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(operatorPkList[i], l1EthClient, l)
			if err != nil {
				t.Fatalf("Failed to create private key signer for executor %d: %v", i, err)
			}
			execCc, err := caller.NewContractCaller(l1EthClient, execPrivateKeySigner, l)
			if err != nil {
				t.Fatalf("Failed to create contract caller for executor %d: %v", i, err)
			}

			execPdfForOp := peeringDataFetcher.NewPeeringDataFetcher(execCc, l)

			var execSignersForOp signer.Signers
			if sigConfig.ExecutorCurve == config.CurveTypeECDSA {
				execSignersForOp = signer.Signers{
					ECDSASigner: execSignerForOp,
				}
			} else {
				execSignersForOp = signer.Signers{
					BLSSigner: execSignerForOp,
				}
			}

			execStoreForOp := executorMemory.NewInMemoryExecutorStore()
			exec, err := executor.NewExecutorWithRpcServers(execConfigForOp.GrpcPort, execConfigForOp.GrpcPort, execConfigForOp, l, execSignersForOp, execPdfForOp, execCc, execStoreForOp)
			if err != nil {
				t.Fatalf("Failed to create executor %d: %v", i, err)
			}
			executors = append(executors, exec)
			t.Logf("Created executor %d on port %d with address %s", i, execConfigForOp.GrpcPort, operatorAddrList[i])
		}

	} else {

		t.Logf("Creating single executor instance for Kubernetes mode")
		execPdf := peeringDataFetcher.NewPeeringDataFetcher(l1ExecCc, l)

		var execSigners signer.Signers
		if sigConfig.ExecutorCurve == config.CurveTypeECDSA {
			execSigners = signer.Signers{
				ECDSASigner: execSigner,
			}
		} else {
			execSigners = signer.Signers{
				BLSSigner: execSigner,
			}
		}

		execStore := executorMemory.NewInMemoryExecutorStore()
		realExec, err := executor.NewExecutorWithRpcServers(execConfig.GrpcPort, execConfig.GrpcPort, execConfig, l, execSigners, execPdf, l1ExecCc, execStore)
		if err != nil {
			t.Fatalf("Failed to create executor: %v", err)
		}
		executors = append(executors, realExec)
	}

	eigenlayerContracts, err := eigenlayer.LoadContracts()
	if err != nil {
		t.Fatalf("failed to load contracts: %v", err)
	}

	imContractStore := inMemoryContractStore.NewInMemoryContractStore(eigenlayerContracts, l)

	tlp := transactionLogParser.NewTransactionLogParser(imContractStore, l)
	aggPdf := peeringDataFetcher.NewPeeringDataFetcher(l1AggCc, l)

	aggStore := memory.NewInMemoryAggregatorStore()

	var aggSigners signer.Signers
	if sigConfig.AggregatorCurve == config.CurveTypeECDSA {
		aggSigners = signer.Signers{
			ECDSASigner: aggSigner,
		}
	} else {
		aggSigners = signer.Signers{
			BLSSigner: aggSigner,
		}
	}

	agg, err := NewAggregatorWithManagementRpcServer(
		9002,
		&AggregatorConfig{
			AVSs:             aggConfig.Avss,
			Chains:           aggConfig.Chains,
			Address:          aggConfig.Operator.Address,
			PrivateKeyConfig: aggConfig.Operator.OperatorPrivateKey,
			L1ChainId:        aggConfig.L1ChainId,
		},
		imContractStore,
		tlp,
		aggPdf,
		aggSigners,
		aggStore,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	for i, exec := range executors {
		t.Logf("Initializing and running executor %d", i)
		if err := exec.Initialize(ctx); err != nil {
			t.Fatalf("Failed to initialize executor %d: %v", i, err)
		}
		if err := exec.Run(ctx); err != nil {
			t.Fatalf("Failed to run executor %d: %v", i, err)
		}
	}

	if err := agg.Initialize(); err != nil {
		t.Fatalf("Failed to initialize aggregator: %v", err)
	}

	go func() {
		if err := agg.Start(ctx); err != nil {
			cancel()
		}
	}()

	wsEthClient, err := l2EthereumClient.GetWebsocketConnection(l2WsUrl)
	if err != nil {
		t.Fatalf("Failed to get websocket connection: %v", err)
	}

	taskVerified := false

	eventsChan := make(chan types.Log)
	sub, err := wsEthClient.SubscribeFilterLogs(ctx, goEthereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(l2EigenlayerContractAddrs.TaskMailbox)},
	}, eventsChan)
	if err != nil {
		t.Fatalf("Failed to subscribe to events: %v", err)
	}
	defer close(eventsChan)
	go func() {
		for {
			select {
			case err := <-sub.Err():
				t.Logf("Error in subscription: %v", err)
				cancel()
				return
			case event := <-eventsChan:
				eventBytes, err := event.MarshalJSON()
				if err != nil {
					t.Logf("Failed to marshal event: %v", err)
					cancel()
					return
				}
				var eventLog *ethereum.EthereumEventLog
				if err := json.Unmarshal(eventBytes, &eventLog); err != nil {
					t.Logf("Failed to unmarshal event: %v", err)
					cancel()
					return
				}

				decodedLog, err := tlp.DecodeLog(nil, eventLog)
				if err != nil {
					t.Logf("Failed to decode log: %v", err)
					cancel()
					return
				}

				t.Logf("Received event: %+v", decodedLog)

				if decodedLog.EventName == "TaskVerified" {
					t.Logf("Task verified: %+v", decodedLog)
					taskVerified = true
					wsEthClient.Client().Close()
					cancel()
				}
			}
		}
	}()

	l2AppPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AppAccountPrivateKey, l2EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create L2 app private key bn254Signer: %v", err)
	}

	l2AppCc, err := caller.NewContractCaller(l2EthClient, l2AppPrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create contract caller: %v", err)
	}
	t.Logf("Pushing message to mailbox...")
	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))

	time.Sleep(10 * time.Second)

	task, err := l2AppCc.PublishMessageToInbox(ctx, chainConfig.AVSAccountAddress, 1, payloadJsonBytes)
	if err != nil {
		t.Fatalf("Failed to publish message to inbox: %v", err)
	}
	t.Logf("Task published: %+v", task)

	select {
	case <-ctx.Done():
		t.Logf("Context done: %v", ctx.Err())
	case <-time.After(250 * time.Second):
		t.Logf("Timeout after 250 seconds")
		cancel()
	}
	fmt.Printf("Test completed\n")

	time.Sleep(5 * time.Second)
	assert.True(t, taskVerified)

	_ = testUtils.KillAnvil(l1Anvil)
	_ = testUtils.KillAnvil(l2Anvil)
	cancel()
}

func loadPerformerImage(ctx context.Context, cluster *testUtils.KindCluster, logger *zap.SugaredLogger) error {
	execConfig, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(getExecutorConfigYaml("kubernetes", config.CurveTypeECDSA)))
	if err != nil {
		return fmt.Errorf("failed to parse executor config: %v", err)
	}

	if len(execConfig.AvsPerformers) == 0 {
		return fmt.Errorf("no AVS performers found in executor config")
	}

	performer := execConfig.AvsPerformers[0]
	imageName := fmt.Sprintf("%s:%s", performer.Image.Repository, performer.Image.Tag)

	logger.Infof("Loading performer image into Kind cluster: %s", imageName)
	if err := cluster.LoadDockerImage(ctx, imageName); err != nil {
		return fmt.Errorf("failed to load image into Kind cluster: %v", err)
	}

	logger.Infof("Successfully loaded performer image: %s", imageName)
	return nil
}

func installPerformerCRD(ctx context.Context, cluster *testUtils.KindCluster, projectRoot string, logger *zap.SugaredLogger) error {
	crdPath := filepath.Join(projectRoot, "..", "hourglass-operator", "config", "crd", "bases", "hourglass.eigenlayer.io_performers.yaml")

	logger.Infof("Installing Performer CRD from: %s", crdPath)

	output, err := cluster.RunKubectl(ctx, "apply", "-f", crdPath)
	if err != nil {
		return fmt.Errorf("failed to apply Performer CRD: %v\nOutput: %s", err, string(output))
	}

	logger.Infof("Performer CRD installed successfully")
	return nil
}

func getAggregatorConfigYaml(l1RpcUrl, l2RpcUrl string, curveType config.CurveType) string {
	var signingKeysSection string
	if curveType == config.CurveTypeECDSA {
		signingKeysSection = `  signingKeys:
    ecdsa:
      privateKey: "0xebf9626f161c0f7becb1aad06b6eab6290c312c03ce12eeeb9ff9b2e204c8a35"`
	} else {
		signingKeysSection = `  signingKeys:
    bls:
      password: ""
      keystore: |
        {
          "crypto": {
            "kdf": {
              "function": "scrypt",
              "params": {
                "dklen": 32,
                "n": 262144,
                "p": 1,
                "r": 8,
                "salt": "dfca382309f4848f5b19e68b210a4352483ac2932ed85fd33dcf18a65cf6df00"
              },
              "message": ""
            },
            "checksum": {
              "function": "sha256",
              "params": {},
              "message": "2a199250fa26519cf2126a1412146401841dcf01bf3b7247400e0a7a76c4250b"
            },
            "cipher": {
              "function": "aes-128-ctr",
              "params": {
                "iv": "677edd29eff1f8635a51f66f71bc5c83"
              },
              "message": "162d9d639a04c1ba85eca100875408dcc19fcd4c3d046137a73c777dde1f8347"
            }
          },
          "pubkey": "2d9070dd755001e31106e8fd58e12f391d09748e5e729512847a944f59966c3311647e4f059bc95ca7f82ecf104758658faa6c3fd18e520c84ba494659b0c6aa015b70ece5cf79963f6295b2db088213732f8bd5c2c456039cd76991e8f24fc225de170c25e59665e9ed95313f43f0bfc93122445e048c9a91fbdea84c71d169",
          "path": "m/1/0/0",
          "uuid": "3b7d7ab3-4472-417f-8f2f-8b2a7011a463",
          "version": 4,
          "curveType": "bn254"
        }`
	}

	return fmt.Sprintf(`
---
chains:
  - name: ethereum
    network: sepolia
    chainId: 31337
    rpcUrl: %s
    pollIntervalSeconds: 5
  - name: base
    network: sepolia
    chainId: 31338
    rpcUrl: %s
    pollIntervalSeconds: 5
l1ChainId: 31337
operator:
  address: "0x204AFca2CBE13C81140f16Fe05063D1b62B56d7D"
  operatorPrivateKey:
    privateKey: "0xebf9626f161c0f7becb1aad06b6eab6290c312c03ce12eeeb9ff9b2e204c8a35"
%s
avss:
  - address: "0x8e14dB002737F89745bc98F987caeB18D0d47635"
    responseTimeout: 3000
    chainIds: [31338]
    avsRegistrarAddress: "0x005ba0ba463b0709380afdf8ff5045c461878c90"
`, l1RpcUrl, l2RpcUrl, signingKeysSection)
}

func getExecutorConfigYaml(mode string, curveType config.CurveType) string {
	blsKeystore := `
        {
          "crypto": {
            "kdf": {
              "function": "scrypt",
              "params": {
                "dklen": 32,
                "n": 262144,
                "p": 1,
                "r": 8,
                "salt": "be920dab5644b5036299788e5a4082fd03c978cc35903b528af754fe7aeccb41"
              },
              "message": ""
            },
            "checksum": {
              "function": "sha256",
              "params": {},
              "message": "28566410c36025d243d0ea9e061ccb46651f09d63ebba598752db2f781d040da"
            },
            "cipher": {
              "function": "aes-128-ctr",
              "params": {
                "iv": "cbaff55d36de018603dc9a336ac3bdc7"
              },
              "message": "3d261076c91fdc6b1de390d0136b22c2a79b83b2838d55dd646218b7cec58396"
            }
          },
          "pubkey": "11d5ec232840a49a1b48d4a6dc0b2e2cb6d5d4d7fc0ef45233f91b98a384d7090f19ac8105e5eaab41aea1ce0021511627a0063ef06f5815cc38bcf0ef4a671e292df403d6a7d6d331b6992dc5b2a06af62bb9c61d7a037a0cd33b88a87950412746cea67ee4b7d3cf0d9f97fdd5bca4690895df14930d78f28db3ff287acea9",
          "path": "m/1/0/0",
          "uuid": "8df75d34-4383-4ff4-a3c0-c47717c72e86",
          "version": 4,
          "curveType": "bn254"
        }`

	var signingKeysSection string
	if curveType == config.CurveTypeECDSA {
		signingKeysSection = `  signingKeys:
    ecdsa:
      privateKey: "0xafc712ca53b54dd37c0fdd382213912e469e617a6e81a8defbc5affcbde4284b"`
	} else {
		signingKeysSection = fmt.Sprintf(`  signingKeys:
    bls:
      password: ""
      keystore: |%s`, blsKeystore)
	}

	if mode == "kubernetes" {
		return fmt.Sprintf(`
---
grpcPort: 9000
operator:
  address: "0x8A07AFaD6AbF7cB9B9da595d0301Fef5145a0BE8"
  operatorPrivateKey:
    privateKey: "0xafc712ca53b54dd37c0fdd382213912e469e617a6e81a8defbc5affcbde4284b"
%s
l1Chain:
  rpcUrl: "http://localhost:8545"
  chainId: 31337
kubernetes:
  namespace: "default"
  operatorNamespace: "hourglass-system"
  crdGroup: "hourglass.eigenlayer.io"
  crdVersion: "v1alpha1"
  connectionTimeout: 30s
  inCluster: false
  kubeConfigPath: "/tmp/kind-kubeconfig"
avsPerformers:
- image:
    repository: "hello-performer"
    tag: "latest"
  processType: "server"
  avsAddress: "0x8e14dB002737F89745bc98F987caeB18D0d47635"
  avsRegistrarAddress: "0x7675776c164b786084474f5cc0c9c3d27118e4d1"
  deploymentMode: "kubernetes"
  kubernetes:
    endpointOverride: "localhost:30080"
`, signingKeysSection)
	} else {
		return fmt.Sprintf(`
---
debug: true
grpcPort: 9000
operator:
  address: "0x8A07AFaD6AbF7cB9B9da595d0301Fef5145a0BE8"
  operatorPrivateKey:
    privateKey: "0xafc712ca53b54dd37c0fdd382213912e469e617a6e81a8defbc5affcbde4284b"
%s
l1Chain:
  rpcUrl: "http://localhost:8545"
  chainId: 31337
avsPerformers:
- image:
    repository: "hello-performer"
    tag: "latest"
  processType: "server"
  avsAddress: "0x8e14dB002737F89745bc98F987caeB18D0d47635"
  avsRegistrarAddress: "0x7675776c164b786084474f5cc0c9c3d27118e4d1"
  deploymentMode: "docker"
`, signingKeysSection)
	}
}

func loadOperatorImage(ctx context.Context, cluster *testUtils.KindCluster, logger *zap.SugaredLogger) error {
	logger.Info("Loading pre-built Hourglass operator image into Kind cluster")

	if err := cluster.LoadDockerImage(ctx, "hourglass/operator:test"); err != nil {
		return fmt.Errorf("failed to load operator image to Kind: %v", err)
	}

	logger.Info("Successfully loaded Hourglass operator image")
	return nil
}

func createPerformerNodePortService(ctx context.Context, cluster *testUtils.KindCluster, logger *zap.SugaredLogger) error {
	logger.Info("Creating NodePort service to expose performer pods")

	serviceYAML := `
apiVersion: v1
kind: Service
metadata:
  name: performer-nodeport
  namespace: default
spec:
  type: NodePort
  selector:
    app: hourglass-performer
  ports:
  - port: 8080
    targetPort: 8080
    nodePort: 30080
    protocol: TCP
`

	if err := cluster.RunKubectlWithInput(ctx, serviceYAML, "apply", "-f", "-"); err != nil {
		return fmt.Errorf("failed to create NodePort service: %v", err)
	}

	logger.Info("NodePort service created successfully on port 30080")
	return nil
}

func Test_DeRegisterAvs(t *testing.T) {
	tests := []struct {
		name            string
		setupAggregator func(t *testing.T) *Aggregator
		request         *aggregatorV1.DeRegisterAvsRequest
		expectSuccess   bool
		expectError     bool
		errorCode       codes.Code
		errorContains   string
	}{
		{
			name: "successful_deregistration",
			setupAggregator: func(t *testing.T) *Aggregator {
				agg := createTestAggregator(t)

				mockAEM := &avsExecutionManager.AvsExecutionManager{}
				avsInfo := &AvsExecutionManagerInfo{
					Address:          "0x123avs",
					ExecutionManager: mockAEM,
					CancelFunc:       nil,
				}
				agg.avsManagers["0x123avs"] = avsInfo
				return agg
			},
			request: &aggregatorV1.DeRegisterAvsRequest{
				AvsAddress: "0x123avs",
				Auth:       nil, // No auth for this test
			},
			expectSuccess: true,
			expectError:   false,
		},
		{
			name: "deregister_nonexistent_avs",
			setupAggregator: func(t *testing.T) *Aggregator {
				return createTestAggregator(t)
			},
			request: &aggregatorV1.DeRegisterAvsRequest{
				AvsAddress: "0x999nonexistent",
				Auth:       nil,
			},
			expectSuccess: false,
			expectError:   true,
			errorCode:     codes.Internal,
			errorContains: "AVS 0x999nonexistent is not registered",
		},
		{
			name: "deregister_with_auth_disabled_but_provided",
			setupAggregator: func(t *testing.T) *Aggregator {
				agg := createTestAggregator(t)

				mockAEM := &avsExecutionManager.AvsExecutionManager{}
				avsInfo := &AvsExecutionManagerInfo{
					Address:          "0x456avs",
					ExecutionManager: mockAEM,
					CancelFunc:       nil,
				}
				agg.avsManagers["0x456avs"] = avsInfo
				return agg
			},
			request: &aggregatorV1.DeRegisterAvsRequest{
				AvsAddress: "0x456avs",
				Auth: &commonV1.AuthSignature{
					ChallengeToken: "test-challenge-token",
					Signature:      []byte("test-signature"),
				},
			},
			expectSuccess: false,
			expectError:   true,
			errorCode:     codes.Unimplemented,
			errorContains: "authentication is not enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			agg := tt.setupAggregator(t)

			// Store initial count of AVS execution managers
			initialCount := len(agg.avsManagers)

			response, err := agg.DeRegisterAvs(ctx, tt.request)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCode != codes.OK {
					st, ok := status.FromError(err)
					require.True(t, ok, "Error should be a gRPC status error")
					assert.Equal(t, tt.errorCode, st.Code())
				}
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, tt.expectSuccess, response.Success)

			if tt.expectSuccess {
				_, exists := agg.avsManagers[tt.request.AvsAddress]
				assert.False(t, exists, "AVS should be removed from avsManagers")
				assert.Equal(t, initialCount-1, len(agg.avsManagers), "AVS count should decrease by 1")
			}
		})
	}
}

func Test_DeRegisterAvs_ContextCancellation(t *testing.T) {
	ctx := context.Background()
	agg := createTestAggregator(t)

	avsAddress := "0x123avs"
	mockAEM := &avsExecutionManager.AvsExecutionManager{}

	cancelCalled := false
	mockCancel := func() {
		cancelCalled = true
	}

	avsInfo := &AvsExecutionManagerInfo{
		Address:          avsAddress,
		ExecutionManager: mockAEM,
		CancelFunc:       mockCancel,
	}
	agg.avsManagers[avsAddress] = avsInfo

	assert.Equal(t, 1, len(agg.avsManagers))

	request := &aggregatorV1.DeRegisterAvsRequest{
		AvsAddress: avsAddress,
		Auth:       nil,
	}

	response, err := agg.DeRegisterAvs(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)

	assert.True(t, cancelCalled, "Cancel function should have been called")

	_, avsExists := agg.avsManagers[avsAddress]
	assert.False(t, avsExists, "AVS should be removed from avsManagers map")

	assert.Equal(t, 0, len(agg.avsManagers))
}

func Test_DeRegisterAvs_ThreadSafety(t *testing.T) {
	ctx := context.Background()
	agg := createTestAggregator(t)

	const numGoroutines = 10
	const numOperations = 20

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				avsAddress := fmt.Sprintf("0x%d%davs", goroutineID, j)

				agg.avsMutex.Lock()
				if _, exists := agg.avsManagers[avsAddress]; !exists {
					mockAEM := &avsExecutionManager.AvsExecutionManager{}
					mockCancel := func() {}
					avsInfo := &AvsExecutionManagerInfo{
						Address:          avsAddress,
						ExecutionManager: mockAEM,
						CancelFunc:       mockCancel,
					}
					agg.avsManagers[avsAddress] = avsInfo
				}
				agg.avsMutex.Unlock()

				time.Sleep(time.Microsecond * 100)

				request := &aggregatorV1.DeRegisterAvsRequest{
					AvsAddress: avsAddress,
					Auth:       nil,
				}

				_, err := agg.DeRegisterAvs(ctx, request)
				if err != nil {
					assert.Contains(t, err.Error(), "not registered")
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify no maps are corrupted and all entries are properly cleaned up
	agg.avsMutex.RLock()
	avsCount := len(agg.avsManagers)

	// Verify data integrity - each AVS info should have consistent fields
	for avsAddr, avsInfo := range agg.avsManagers {
		assert.Equal(t, avsAddr, avsInfo.Address, "AVS info address should match map key")
		assert.NotNil(t, avsInfo.ExecutionManager, "Execution manager should not be nil")
		// CancelFunc can be nil if AVS was registered but not yet started
	}
	agg.avsMutex.RUnlock()

	t.Logf("Thread safety test completed. Final state: %d AVSs registered", avsCount)
}

// createTestAggregator creates a minimal aggregator for testing
func createTestAggregator(t *testing.T) *Aggregator {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	// Create minimal aggregator with required fields
	agg := &Aggregator{
		logger: l,
		config: &AggregatorConfig{
			Address:   "0xaggregator",
			L1ChainId: config.ChainId_EthereumMainnet,
		},
		rootCtx:      context.Background(), // Set root context for tests
		avsManagers:  make(map[string]*AvsExecutionManagerInfo),
		authVerifier: nil, // No auth verification for these tests
	}

	return agg
}
