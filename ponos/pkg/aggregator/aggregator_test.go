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

			//t.Run("Kubernetes", func(t *testing.T) {
			//	runAggregatorTest(t, "kubernetes", sigConfig)
			//})
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

	// Setup Kind cluster for Kubernetes mode only
	var cluster *testUtils.KindCluster
	if mode == "kubernetes" {
		// Clean up any existing test clusters first to prevent port conflicts
		if err := testUtils.CleanupAllTestClusters(l.Sugar()); err != nil {
			t.Logf("Warning: Failed to cleanup existing test clusters: %v", err)
		}

		// Create Kind cluster
		kindConfig := testUtils.DefaultKindClusterConfig(l.Sugar())
		var clusterCleanup func()
		cluster, clusterCleanup, err = testUtils.CreateKindCluster(ctx, t, kindConfig)
		if err != nil {
			t.Fatalf("Failed to create Kind cluster: %v", err)
		}
		defer func() {
			// Nuclear option: just delete the cluster to avoid hanging cleanup
			t.Log("Using fast cluster deletion to avoid hanging cleanup")
			clusterCleanup()
		}()

		// Load performer image from executor config (assumes image is already built)
		if err := loadPerformerImage(ctx, cluster, l.Sugar()); err != nil {
			t.Fatalf("Failed to load performer image: %v", err)
		}

		// Install CRDs first (required for Performer objects)
		if err := installPerformerCRD(ctx, cluster, root, l.Sugar()); err != nil {
			t.Fatalf("Failed to install Performer CRD: %v", err)
		}

		// Load pre-built operator image
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
			// Run cleanup with timeout to avoid hanging
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

		// Create NodePort service to expose performer pods to the host
		if err := createPerformerNodePortService(ctx, cluster, l.Sugar()); err != nil {
			t.Fatalf("Failed to create NodePort service: %v", err)
		}
	}

	// ------------------------------------------------------------------------
	// Executor setup
	// ------------------------------------------------------------------------
	execConfig, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(getExecutorConfigYaml(mode, sigConfig.ExecutorCurve)))
	if err != nil {
		t.Fatalf("failed to create executor config: %v", err)
	}
	if err := execConfig.Validate(); err != nil {
		t.Fatalf("failed to validate executor config: %v", err)
	}
	// Set signing keys based on the curve type
	if sigConfig.ExecutorCurve == config.CurveTypeECDSA {
		execConfig.Operator.SigningKeys.ECDSA = &config.ECDSAKeyConfig{
			PrivateKey: chainConfig.ExecOperatorAccountPk,
		}
	}
	// BLS keys are already set in the YAML config when needed

	execConfig.Operator.OperatorPrivateKey = &config.ECDSAKeyConfig{
		PrivateKey: chainConfig.ExecOperatorAccountPk,
	}
	execConfig.Operator.Address = chainConfig.ExecOperatorAccountAddress
	execConfig.AvsPerformers[0].AvsAddress = chainConfig.AVSAccountAddress

	// Configure Kubernetes config based on mode
	if mode == "kubernetes" {
		// Set the actual kubeconfig path for the Kind cluster
		execConfig.Kubernetes.KubeConfigPath = cluster.KubeConfig
	}

	_, _, execGenericExecutorSigningKey, err := testUtils.ParseKeysFromConfig(execConfig.Operator, sigConfig.ExecutorCurve)
	if err != nil {
		t.Fatalf("Failed to parse keys from config: %v", err)
	}
	execSigner := inMemorySigner.NewInMemorySigner(execGenericExecutorSigningKey, sigConfig.ExecutorCurve)

	// ------------------------------------------------------------------------
	// Generate keys for executor operators (2 for Docker, 1 for Kubernetes)
	// ------------------------------------------------------------------------
	numExecutorOperators := 1
	if mode == "docker" {
		numExecutorOperators = 2
	}
	t.Logf("------------------------------------------- Setting up %d executor operator(s) for %s mode -------------------------------------------", numExecutorOperators, mode)

	// Get keys for aggregator and executors using the helper method
	// This returns the aggregator key and all executor keys at once
	_, execKeys, err := testUtils.GetKeysForCurveTypeFromChainConfig(t, sigConfig.AggregatorCurve, sigConfig.ExecutorCurve, chainConfig)
	if err != nil {
		t.Fatalf("Failed to get keys: %v", err)
	}

	// Verify we got the expected number of executor keys
	if len(execKeys) < numExecutorOperators {
		t.Fatalf("Expected at least %d executor keys, got %d", numExecutorOperators, len(execKeys))
	}

	// Use only the first 2 executor keys
	execKeys = execKeys[:numExecutorOperators]

	// ------------------------------------------------------------------------
	// 	Aggregator setup
	// ------------------------------------------------------------------------
	aggConfigYaml := getAggregatorConfigYaml(L1RPCUrl, L2RPCUrl, sigConfig.AggregatorCurve)

	aggConfig, err := aggregatorConfig.NewAggregatorConfigFromYamlBytes([]byte(aggConfigYaml))
	if err != nil {
		t.Fatalf("Failed to create aggregator config: %v", err)
	}

	aggConfig.Operator.Address = chainConfig.OperatorAccountAddress
	aggConfig.Operator.OperatorPrivateKey = &config.ECDSAKeyConfig{
		PrivateKey: chainConfig.OperatorAccountPrivateKey,
	}

	// Set signing keys based on the curve type for aggregator
	if sigConfig.AggregatorCurve == config.CurveTypeECDSA {
		aggConfig.Operator.SigningKeys.ECDSA = &config.ECDSAKeyConfig{
			PrivateKey: chainConfig.OperatorAccountPrivateKey,
		}
		aggConfig.Operator.SigningKeys.BLS = nil
	} else {
		// BLS keys are already set in the YAML config
		aggConfig.Operator.SigningKeys.ECDSA = nil
	}

	aggBn254PrivateSigningKey, aggEcdsaPrivateSigningKey, aggGenericExecutorSigningKey, err := testUtils.ParseKeysFromConfig(aggConfig.Operator, sigConfig.AggregatorCurve)
	if err != nil {
		t.Fatalf("Failed to parse keys from config: %v", err)
	}

	aggSigner := inMemorySigner.NewInMemorySigner(aggGenericExecutorSigningKey, sigConfig.AggregatorCurve)

	// ------------------------------------------------------------------------
	// L1Chain & l1Anvil setup
	// ------------------------------------------------------------------------
	anvilWg := &sync.WaitGroup{}

	// Both Docker and Kubernetes modes use localhost for executor/aggregator
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

	// ------------------------------------------------------------------------
	// L2Chain & l2Anvil setup
	// ------------------------------------------------------------------------
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

	// ------------------------------------------------------------------------
	// Configure operator sets dynamically based on test configuration
	// ------------------------------------------------------------------------
	t.Logf("------------------------------------------- Configuring operator sets dynamically -------------------------------------------")

	// Create AVS contract caller for configuring operator sets
	avsConfigPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create AVS config private key signer: %v", err)
	}

	avsConfigCaller, err := caller.NewContractCaller(l1EthClient, avsConfigPrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create AVS config caller: %v", err)
	}

	// Configure aggregator operator set with its curve type
	t.Logf("Configuring operator set %d with curve type %s for aggregator", sigConfig.AggregatorOpsetId, sigConfig.AggregatorCurve)
	_, err = avsConfigCaller.ConfigureAVSOperatorSet(ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		sigConfig.AggregatorOpsetId,
		sigConfig.AggregatorCurve)
	if err != nil {
		t.Fatalf("Failed to configure aggregator operator set %d: %v", sigConfig.AggregatorOpsetId, err)
	}

	// Configure executor operator set with its curve type
	t.Logf("Configuring operator set %d with curve type %s for executor", sigConfig.ExecutorOpsetId, sigConfig.ExecutorCurve)
	_, err = avsConfigCaller.ConfigureAVSOperatorSet(ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		sigConfig.ExecutorOpsetId,
		sigConfig.ExecutorCurve)
	if err != nil {
		t.Fatalf("Failed to configure executor operator set %d: %v", sigConfig.ExecutorOpsetId, err)
	}

	t.Logf("Successfully configured operator sets dynamically")

	// ------------------------------------------------------------------------
	// Validate operator signing keys match configured curve types
	// ------------------------------------------------------------------------
	// Validate aggregator keys
	if sigConfig.AggregatorCurve == config.CurveTypeBN254 && aggConfig.Operator.SigningKeys.BLS == nil {
		t.Fatalf("BLS signing key required for BN254 curve but not configured for aggregator")
	}
	if sigConfig.AggregatorCurve == config.CurveTypeECDSA && aggConfig.Operator.SigningKeys.ECDSA == nil {
		t.Fatalf("ECDSA signing key required but not configured for aggregator")
	}

	// Validate executor keys
	if sigConfig.ExecutorCurve == config.CurveTypeBN254 && execConfig.Operator.SigningKeys.BLS == nil {
		t.Fatalf("BLS signing key required for BN254 curve but not configured for executor")
	}
	if sigConfig.ExecutorCurve == config.CurveTypeECDSA && execConfig.Operator.SigningKeys.ECDSA == nil {
		t.Fatalf("ECDSA signing key required but not configured for executor")
	}

	t.Logf("Operator signing keys validated successfully")

	// ------------------------------------------------------------------------
	// Register multiple executor operators
	// ------------------------------------------------------------------------
	t.Logf("------------------------------------------- Registering aggregator and executor operators -------------------------------------------")
	// NOTE: we must register ALL opsets regardless of which curve type we are using, otherwise table transport fails
	allOperatorSetIds := []uint32{sigConfig.AggregatorOpsetId, sigConfig.ExecutorOpsetId}

	// Select the correct signing key based on curve type for aggregator
	var aggSigningKey interface{}
	if sigConfig.AggregatorCurve == config.CurveTypeBN254 {
		aggSigningKey = aggBn254PrivateSigningKey
	} else {
		aggSigningKey = aggEcdsaPrivateSigningKey
	}

	// Create executor operators based on mode
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

	// Register aggregator and all executor operators
	err = testUtils.SetupOperatorPeeringWithMultipleExecutors(
		ctx,
		chainConfig,
		config.ChainId(l1ChainId.Uint64()),
		l1EthClient,
		// aggregator with configured curve
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

	// ------------------------------------------------------------------------
	// Delegate stake to all operators
	// ------------------------------------------------------------------------
	t.Logf("------------------------------------------- Delegating stake to operators -------------------------------------------")

	// Build stake configs dynamically based on number of operators
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
		t.Fatalf("Failed to delegate stake to operators: %v", err)
	}

	t.Logf("All operator set IDs: %v", allOperatorSetIds)

	// Create generation reservations for each operator set
	t.Logf("------------------------ Creating generation reservations ------------------------")

	// Create a temporary AVS contract caller to set up the generation reservations
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

	// Wait for transactions to be mined
	time.Sleep(time.Second * 3)

	// update current block to account for transport
	currentBlock, err := l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}
	testUtils.DebugOpsetData(t, chainConfig, eigenlayerContractAddrs, l1EthClient, currentBlock, allOperatorSetIds)

	time.Sleep(time.Second * 6)

	l.Sugar().Infow("------------------------ Transporting L1 & L2 tables ------------------------")

	// For aggregator test, we need L2 transport - don't ignore L2 chain
	chainIdsToIgnore := []*big.Int{
		new(big.Int).SetUint64(1),    // Ethereum Mainnet
		new(big.Int).SetUint64(8453), // Base Mainnet
	}

	// Stake weights: 60% and 40% (both meet threshold)
	stakeWeights := []*big.Int{
		big.NewInt(1500000000000000000), // 1.5e18 = 60%
		big.NewInt(1000000000000000000), // 1e18   = 40%
	}

	operatorAddressList := []string{
		chainConfig.ExecOperatorAccountAddress,
		chainConfig.ExecOperator2AccountAddress,
	}

	// Build operator key infos for both operators
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

	// update current block to account for transport
	currentBlock, err = l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}
	l.Sugar().Infow("Current block number", zap.Uint64("blockNumber", currentBlock))
	testUtils.DebugOpsetData(t, chainConfig, eigenlayerContractAddrs, l1EthClient, currentBlock, allOperatorSetIds)

	// ------------------------------------------------------------------------
	// Setup executors (2 for Docker mode, 1 for Kubernetes mode)
	// ------------------------------------------------------------------------
	var executors []*executor.Executor

	if mode == "docker" {
		// Docker mode: Create 2 separate executor instances
		t.Logf("Creating 2 executor instances for Docker mode")

		for i := 0; i < numExecutorOperators; i++ {
			// Create separate config for each executor
			execConfigForOp, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(getExecutorConfigYaml(mode, sigConfig.ExecutorCurve)))
			if err != nil {
				t.Fatalf("Failed to create executor config for operator %d: %v", i, err)
			}

			// Set operator-specific fields
			operatorPkList := []string{chainConfig.ExecOperatorAccountPk, chainConfig.ExecOperator2AccountPk}
			operatorAddrList := []string{chainConfig.ExecOperatorAccountAddress, chainConfig.ExecOperator2AccountAddress}

			execConfigForOp.Operator.OperatorPrivateKey = &config.ECDSAKeyConfig{
				PrivateKey: operatorPkList[i],
			}
			execConfigForOp.Operator.Address = operatorAddrList[i]
			execConfigForOp.AvsPerformers[0].AvsAddress = chainConfig.AVSAccountAddress
			execConfigForOp.GrpcPort = 9000 + i

			// Set signing keys based on curve type
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

			// Set up executor signers based on curve type
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

			// Use in-memory storage for the executor
			execStoreForOp := executorMemory.NewInMemoryExecutorStore()
			exec, err := executor.NewExecutorWithRpcServers(execConfigForOp.GrpcPort, execConfigForOp.GrpcPort, execConfigForOp, l, execSignersForOp, execPdfForOp, execCc, execStoreForOp)
			if err != nil {
				t.Fatalf("Failed to create executor %d: %v", i, err)
			}
			executors = append(executors, exec)
			t.Logf("Created executor %d on port %d with address %s", i, execConfigForOp.GrpcPort, operatorAddrList[i])
		}
	} else {
		// Kubernetes mode: Create single executor (as before)
		t.Logf("Creating single executor instance for Kubernetes mode")
		execPdf := peeringDataFetcher.NewPeeringDataFetcher(l1ExecCc, l)

		// Set up executor signers based on curve type
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

		// Use in-memory storage for the executor
		execStore := executorMemory.NewInMemoryExecutorStore()
		realExec, err := executor.NewExecutorWithRpcServers(execConfig.GrpcPort, execConfig.GrpcPort, execConfig, l, execSigners, execPdf, l1ExecCc, execStore)
		if err != nil {
			t.Fatalf("Failed to create executor: %v", err)
		}
		executors = append(executors, realExec)
	}

	// ------------------------------------------------------------------------
	// Setup the aggregator
	// ------------------------------------------------------------------------
	eigenlayerContracts, err := eigenlayer.LoadContracts()
	if err != nil {
		t.Fatalf("failed to load contracts: %v", err)
	}

	imContractStore := inMemoryContractStore.NewInMemoryContractStore(eigenlayerContracts, l)

	tlp := transactionLogParser.NewTransactionLogParser(imContractStore, l)
	aggPdf := peeringDataFetcher.NewPeeringDataFetcher(l1AggCc, l)

	// Create in-memory storage for testing
	aggStore := memory.NewInMemoryAggregatorStore()

	// Set up aggregator signers based on curve type
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

	// ------------------------------------------------------------------------
	// Boot up everything
	// ------------------------------------------------------------------------
	// Initialize and run all executors
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

	// ------------------------------------------------------------------------
	// Listen for TaskVerified event to know that the test is done
	// ------------------------------------------------------------------------
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

	// ------------------------------------------------------------------------
	// Push a message to the mailbox
	// ------------------------------------------------------------------------
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

// loadPerformerImage loads the performer image referenced in executorConfigYaml into the Kind cluster
func loadPerformerImage(ctx context.Context, cluster *testUtils.KindCluster, logger *zap.SugaredLogger) error {
	// Parse executor config to extract image info (use ECDSA as default for loading image)
	execConfig, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(getExecutorConfigYaml("kubernetes", config.CurveTypeECDSA)))
	if err != nil {
		return fmt.Errorf("failed to parse executor config: %v", err)
	}

	if len(execConfig.AvsPerformers) == 0 {
		return fmt.Errorf("no AVS performers found in executor config")
	}

	// Get the first performer's image (assuming single performer for now)
	performer := execConfig.AvsPerformers[0]
	imageName := fmt.Sprintf("%s:%s", performer.Image.Repository, performer.Image.Tag)

	logger.Infof("Loading performer image into Kind cluster: %s", imageName)

	// Load the image into Kind cluster (assumes image is already built locally)
	if err := cluster.LoadDockerImage(ctx, imageName); err != nil {
		return fmt.Errorf("failed to load image into Kind cluster: %v", err)
	}

	logger.Infof("Successfully loaded performer image: %s", imageName)
	return nil
}

// installPerformerCRD installs the Performer CRD required for the test
func installPerformerCRD(ctx context.Context, cluster *testUtils.KindCluster, projectRoot string, logger *zap.SugaredLogger) error {
	// Path to the Performer CRD file
	crdPath := filepath.Join(projectRoot, "..", "hourglass-operator", "config", "crd", "bases", "hourglass.eigenlayer.io_performers.yaml")

	logger.Infof("Installing Performer CRD from: %s", crdPath)

	// Apply the CRD
	output, err := cluster.RunKubectl(ctx, "apply", "-f", crdPath)
	if err != nil {
		return fmt.Errorf("failed to apply Performer CRD: %v\nOutput: %s", err, string(output))
	}

	logger.Infof("Performer CRD installed successfully")
	return nil
}

// getAggregatorConfigYaml returns the aggregator configuration with configurable RPC URLs and curve type
func getAggregatorConfigYaml(l1RpcUrl, l2RpcUrl string, curveType config.CurveType) string {
	// Build signing keys section based on curve type
	var signingKeysSection string
	if curveType == config.CurveTypeECDSA {
		signingKeysSection = `  signingKeys:
    ecdsa:
      privateKey: "0xebf9626f161c0f7becb1aad06b6eab6290c312c03ce12eeeb9ff9b2e204c8a35"`
	} else {
		// BN254 keystore - using the actual test keystore from testKeys/aggregator
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
	// BLS keystore for when executor uses BN254 - using the actual test keystore
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

	// Build signing keys section based on curve type
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

// loadOperatorImage loads the pre-built Hourglass operator image into the Kind cluster
func loadOperatorImage(ctx context.Context, cluster *testUtils.KindCluster, logger *zap.SugaredLogger) error {
	logger.Info("Loading pre-built Hourglass operator image into Kind cluster")

	// Load the pre-built operator image
	if err := cluster.LoadDockerImage(ctx, "hourglass/operator:test"); err != nil {
		return fmt.Errorf("failed to load operator image to Kind: %v", err)
	}

	logger.Info("Successfully loaded Hourglass operator image")
	return nil
}

// createPerformerNodePortService creates a NodePort service to expose performer pods
func createPerformerNodePortService(ctx context.Context, cluster *testUtils.KindCluster, logger *zap.SugaredLogger) error {
	logger.Info("Creating NodePort service to expose performer pods")

	// Create the NodePort service YAML
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

	// Apply the service
	if err := cluster.RunKubectlWithInput(ctx, serviceYAML, "apply", "-f", "-"); err != nil {
		return fmt.Errorf("failed to create NodePort service: %v", err)
	}

	logger.Info("NodePort service created successfully on port 30080")
	return nil
}

// Test_DeRegisterAvs tests the DeRegisterAvs handler function
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
				// Pre-register an AVS to deregister later
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
				// Pre-register an AVS
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
				// Verify the AVS was actually removed from the map
				_, exists := agg.avsManagers[tt.request.AvsAddress]
				assert.False(t, exists, "AVS should be removed from avsManagers")
				assert.Equal(t, initialCount-1, len(agg.avsManagers), "AVS count should decrease by 1")
			}
		})
	}
}

// Test_DeRegisterAvs_ContextCancellation tests that the context cancellation mechanism works
func Test_DeRegisterAvs_ContextCancellation(t *testing.T) {
	ctx := context.Background()
	agg := createTestAggregator(t)

	// Simulate an AVS being started (normally done in Start() method)
	avsAddress := "0x123avs"
	mockAEM := &avsExecutionManager.AvsExecutionManager{}

	// Create a mock cancel function to verify it gets called
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

	// Verify initial state
	assert.Equal(t, 1, len(agg.avsManagers))

	// Deregister the AVS
	request := &aggregatorV1.DeRegisterAvsRequest{
		AvsAddress: avsAddress,
		Auth:       nil,
	}

	response, err := agg.DeRegisterAvs(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)

	// Verify the cancel function was called
	assert.True(t, cancelCalled, "Cancel function should have been called")

	// Verify cleanup of the map
	_, avsExists := agg.avsManagers[avsAddress]
	assert.False(t, avsExists, "AVS should be removed from avsManagers map")

	// Verify final state
	assert.Equal(t, 0, len(agg.avsManagers))
}

// Test_DeRegisterAvs_ThreadSafety tests concurrent register/deregister operations
func Test_DeRegisterAvs_ThreadSafety(t *testing.T) {
	ctx := context.Background()
	agg := createTestAggregator(t)

	const numGoroutines = 10
	const numOperations = 20

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Run concurrent register/deregister operations
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				avsAddress := fmt.Sprintf("0x%d%davs", goroutineID, j)

				// Register AVS by directly adding to maps (simulating successful registration)
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

				// Small delay to increase chance of race conditions
				time.Sleep(time.Microsecond * 100)

				// Attempt to deregister
				request := &aggregatorV1.DeRegisterAvsRequest{
					AvsAddress: avsAddress,
					Auth:       nil,
				}

				_, err := agg.DeRegisterAvs(ctx, request)
				// Either succeeds or fails with "not registered" - both are valid outcomes
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
