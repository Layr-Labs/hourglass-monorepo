package executor

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/badger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/stretchr/testify/assert"
)

const (
	L1RpcUrl = "http://127.0.0.1:8545"
)

func testWithKeyType(
	t *testing.T,
	curveType config.CurveType,
	executorConfigYaml string,
	aggregatorConfigYaml string,
) {
	t.Logf("Running test with curve type: %s", curveType)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(240*time.Second))

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

	// ------------------------------------------------------------------------
	// Executor setup
	// ------------------------------------------------------------------------
	execConfig, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(executorConfigYaml))
	if err != nil {
		t.Fatalf("failed to create executor config: %v", err)
	}
	execConfig.Operator.SigningKeys.ECDSA = &config.ECDSAKeyConfig{
		PrivateKey: chainConfig.ExecOperatorAccountPk,
	}
	execConfig.Operator.Address = chainConfig.ExecOperatorAccountAddress
	execConfig.AvsPerformers[0].AvsAddress = chainConfig.AVSAccountAddress

	// Get keys using the helper function
	aggKey, execKeys, err := testUtils.GetKeysForCurveTypeFromChainConfig(
		t,
		config.CurveTypeBN254, // Aggregator uses BN254
		config.CurveTypeECDSA, // Executor uses ECDSA
		chainConfig,
	)
	if err != nil {
		t.Fatalf("Failed to get keys from chain config: %v", err)
	}

	execSigner := inMemorySigner.NewInMemorySigner(execKeys[0].PrivateKey, config.CurveTypeECDSA)

	// ------------------------------------------------------------------------
	// Aggregator setup
	// ------------------------------------------------------------------------
	simAggConfig, err := aggregatorConfig.NewAggregatorConfigFromYamlBytes([]byte(aggregatorConfigYaml))
	if err != nil {
		t.Fatalf("Failed to create aggregator config: %v", err)
	}
	simAggConfig.Operator.SigningKeys.ECDSA = &config.ECDSAKeyConfig{
		PrivateKey: chainConfig.OperatorAccountPrivateKey,
	}
	simAggConfig.Operator.Address = chainConfig.OperatorAccountAddress
	simAggConfig.Avss[0].Address = chainConfig.AVSAccountAddress
	l1EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   L1RpcUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)

	l1EthClient, err := l1EthereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get L1 Ethereum contract caller: %v", err)
	}

	_ = testUtils.KillallAnvils()

	anvilWg := &sync.WaitGroup{}
	anvilWg.Add(1)
	startErrorsChan := make(chan error, 1)

	anvilCtx, anvilCancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer anvilCancel()

	l1Anvil, err := testUtils.StartL1Anvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start L1 Anvil: %v", err)
	}
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
		t.Fatalf("Failed to create L1 private key ecdsaSigner: %v", err)
	}

	l1CC, err := caller.NewContractCaller(l1EthClient, l1PrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create L2 contract caller: %v", err)
	}

	t.Logf("------------------------------------------- Setting up operator peering -------------------------------------------")
	// NOTE: we must register ALL opsets regardles of which curve type we are using, otherwise table transport fails
	aggOpsetId := uint32(0)
	execOpsetId := uint32(1)

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
	bn254CalculatorAddr, err := caller.GetTableCalculatorAddress(config.CurveTypeBN254, config.ChainId_EthereumAnvil)
	if err != nil {
		t.Fatalf("Failed to get table calculator address: %v", err)
	}
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
	ecdsaCalculatorAddr, err := caller.GetTableCalculatorAddress(config.CurveTypeECDSA, config.ChainId_EthereumAnvil)
	if err != nil {
		t.Fatalf("Failed to get table calculator address: %v", err)
	}
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

	err = testUtils.SetupOperatorPeering(
		ctx,
		chainConfig,
		config.ChainId(l1ChainId.Uint64()),
		l1EthClient,
		// aggregator is BN254
		&operator.Operator{
			TransactionPrivateKey: chainConfig.OperatorAccountPrivateKey,
			SigningPrivateKey:     aggKey.PrivateKey,
			Curve:                 config.CurveTypeBN254,
			OperatorSetIds:        []uint32{aggOpsetId},
		},
		// executor is ecdsa
		&operator.Operator{
			TransactionPrivateKey: chainConfig.ExecOperatorAccountPk,
			SigningPrivateKey:     execKeys[0].PrivateKey,
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

	pdf := peeringDataFetcher.NewPeeringDataFetcher(l1CC, l)

	signers := signer.Signers{
		ECDSASigner: execSigner,
	}
	// Create in-memory storage for the executor
	store := memory.NewInMemoryExecutorStore()
	exec, err := NewExecutorWithRpcServers(execConfig.GrpcPort, execConfig.GrpcPort, execConfig, l, signers, pdf, l1CC, store)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	if err := exec.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize executor: %v", err)
	}

	success := atomic.Bool{}
	success.Store(false)

	execClient, err := executorClient.NewExecutorClient(fmt.Sprintf("localhost:%d", execConfig.GrpcPort), true)
	if err != nil {
		t.Fatalf("Failed to create executor client: %v", err)
	}

	go func() {
		if err := exec.Run(ctx); err != nil {
			t.Errorf("Failed to run executor: %v", err)
			return
		}
	}()

	// give containers time to start.
	time.Sleep(5 * time.Second)
	taskId := "0x0000000000000000000000000000000000000000000000000000000000000001"
	strBlockNumber, err := l1EthereumClient.GetBlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get L1 block strBlockNumber: %v", err)
	}

	// Parse hex string (e.g., "0x86d4dd") to uint64
	blockNumber, err := strconv.ParseUint(strBlockNumber[2:], 16, 64)
	if err != nil {
		t.Fatalf("Failed to parse L1 block strBlockNumber %s: %v", strBlockNumber, err)
	}
	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))

	encodedMessage, err := util.EncodeTaskSubmissionMessageVersioned(
		taskId,
		simAggConfig.Avss[0].Address,
		chainConfig.ExecOperatorAccountAddress,
		42,
		blockNumber,
		1,
		payloadJsonBytes,
		1,
	)
	if err != nil {
		t.Fatalf("Failed to encode task submission message: %v", err)
	}

	// For BN254, we need to hash the message before signing to match the production behavior
	var payloadSig []byte
	if simAggConfig.Operator.SigningKeys.BLS != nil {
		// BN254 signing - sign the hashed message
		bn254Signer := inMemorySigner.NewInMemorySigner(aggKey.PrivateKey, config.CurveTypeBN254)
		signature, err := bn254Signer.SignMessage(encodedMessage[:])
		if err != nil {
			t.Fatalf("Failed to sign task payload with BN254: %v", err)
		}
		payloadSig = signature
	} else {
		t.Fatalf("signer should be BLS")
	}

	// send the task to the executor
	taskResult, err := execClient.SubmitTask(ctx, &executorV1.TaskSubmission{
		TaskId:             taskId,
		AggregatorAddress:  simAggConfig.Operator.Address,
		AvsAddress:         simAggConfig.Avss[0].Address,
		ExecutorAddress:    chainConfig.ExecOperatorAccountAddress,
		Payload:            payloadJsonBytes,
		Signature:          payloadSig,
		TaskBlockNumber:    blockNumber,
		ReferenceTimestamp: 42,
		OperatorSetId:      1,
		Version:            1,
	})

	if err != nil {
		cancel()
		time.Sleep(5 * time.Second)
		t.Fatalf("Failed to submit task: %v", err)
	}

	assert.NotNil(t, taskResult)

	if curveType == config.CurveTypeBN254 {

		sig, err := bn254.NewSignatureFromBytes(taskResult.ResultSignature)
		if err != nil {
			t.Fatalf("Failed to create BN254 signature from bytes: %v", err)
		}
		assert.NotNil(t, sig, "BN254 signature should be valid")
		assert.Len(t, taskResult.ResultSignature, 64, "BN254 signature should be 64 bytes")
		t.Logf("BN254 result signature present and valid format")

	} else if curveType == config.CurveTypeECDSA {

		sig, err := cryptoLibsEcdsa.NewSignatureFromBytes(taskResult.ResultSignature)
		if err != nil {
			t.Fatalf("Failed to create ECDSA signature from bytes: %v", err)
		}

		assert.NotNil(t, sig, "ECDSA signature should be valid")
		assert.Len(t, taskResult.ResultSignature, 65, "ECDSA signature should be 65 bytes")
		t.Logf("ECDSA result signature present and valid format")

		if len(taskResult.AuthSignature) > 0 {
			authSig, err := cryptoLibsEcdsa.NewSignatureFromBytes(taskResult.AuthSignature)
			if err != nil {
				t.Logf("Warning: Failed to create auth signature from bytes: %v", err)
			} else {
				assert.NotNil(t, authSig, "Auth signature should be valid")
				assert.Len(t, taskResult.AuthSignature, 65, "Auth signature should be 65 bytes")
				t.Logf("AuthSignature present and valid format, length: %d bytes", len(taskResult.AuthSignature))
			}
		}

	} else {
		t.Errorf("Unsupported curve type: %s", curveType)
	}

	cancel()
	t.Logf("Successfully verified signature for task %s", taskResult.TaskId)

	<-ctx.Done()
	t.Logf("Received shutdown signal, shutting down...")

	_ = testUtils.KillAnvil(l1Anvil)
}

// testWithBadgerStorage tests the executor with persistent Badger storage and hydration
func testWithBadgerStorage(
	t *testing.T,
	curveType config.CurveType,
	executorConfigYaml string,
	aggregatorConfigYaml string,
) {
	t.Logf("Running Badger storage test with curve type: %s", curveType)

	// Create temporary directory for Badger storage
	tmpDir, err := os.MkdirTemp("", "executor-badger-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	t.Logf("Using temporary Badger storage directory: %s", tmpDir)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(240*time.Second))
	defer cancel()

	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	root := testUtils.GetProjectRootPath()
	chainConfig, err := testUtils.ReadChainConfig(root)
	require.NoError(t, err)

	// ------------------------------------------------------------------------
	// Common setup (same for both executor instances)
	// ------------------------------------------------------------------------
	execConfig, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(executorConfigYaml))
	require.NoError(t, err)
	execConfig.Operator.SigningKeys.ECDSA = &config.ECDSAKeyConfig{
		PrivateKey: chainConfig.ExecOperatorAccountPk,
	}
	execConfig.Operator.Address = chainConfig.ExecOperatorAccountAddress
	execConfig.AvsPerformers[0].AvsAddress = chainConfig.AVSAccountAddress

	// Add Badger storage configuration
	execConfig.Storage = &executorConfig.StorageConfig{
		Type: "badger",
		BadgerConfig: &executorConfig.BadgerConfig{
			Dir: tmpDir,
		},
	}

	_, execEcdsaPrivateSigningKey, execGenericExecutorSigningKey, err := testUtils.ParseKeysFromConfig(execConfig.Operator, config.CurveTypeECDSA)
	require.NoError(t, err)
	execSigner := inMemorySigner.NewInMemorySigner(execGenericExecutorSigningKey, config.CurveTypeECDSA)

	// Aggregator setup
	simAggConfig, err := aggregatorConfig.NewAggregatorConfigFromYamlBytes([]byte(aggregatorConfigYaml))
	require.NoError(t, err)
	simAggConfig.Operator.SigningKeys.ECDSA = &config.ECDSAKeyConfig{
		PrivateKey: chainConfig.OperatorAccountPrivateKey,
	}
	simAggConfig.Operator.Address = chainConfig.OperatorAccountAddress
	simAggConfig.Avss[0].Address = chainConfig.AVSAccountAddress

	aggBn254PrivateSigningKey, _, _, err := testUtils.ParseKeysFromConfig(simAggConfig.Operator, config.CurveTypeBN254)
	require.NoError(t, err)

	l1EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   L1RpcUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)

	l1EthClient, err := l1EthereumClient.GetEthereumContractCaller()
	require.NoError(t, err)

	// Start Anvil
	_ = testUtils.KillallAnvils()
	l1Anvil, err := testUtils.StartL1Anvil(root, ctx)
	require.NoError(t, err)
	defer func(cmd *exec.Cmd) {
		err := testUtils.KillAnvil(cmd)
		if err != nil {
			t.Logf("Failed to kill l1 anvil: %v", err)
		}
	}(l1Anvil)

	anvilWg := &sync.WaitGroup{}
	anvilWg.Add(1)
	startErrorsChan := make(chan error, 1)
	anvilCtx, anvilCancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	go testUtils.WaitForAnvil(anvilWg, anvilCtx, t, l1EthereumClient, startErrorsChan)
	anvilWg.Wait()
	close(startErrorsChan)
	for err := range startErrorsChan {
		require.NoError(t, err, "Failed to start Anvil")
	}
	anvilCancel()
	t.Log("Anvil is running")

	l1ChainId, err := l1EthClient.ChainID(ctx)
	require.NoError(t, err)

	l1PrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AppAccountPrivateKey, l1EthClient, l)
	require.NoError(t, err)

	l1CC, err := caller.NewContractCaller(l1EthClient, l1PrivateKeySigner, l)
	require.NoError(t, err)

	// Setup operator sets and peering (same as original test)
	aggOpsetId := uint32(0)
	execOpsetId := uint32(1)

	avsAddr := common.HexToAddress(chainConfig.AVSAccountAddress)
	avsPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, l1EthClient, l)
	require.NoError(t, err)
	avsConfigCaller, err := caller.NewContractCaller(l1EthClient, avsPrivateKeySigner, l)
	require.NoError(t, err)

	// Configure operator sets
	_, err = avsConfigCaller.ConfigureAVSOperatorSet(ctx, avsAddr, aggOpsetId, config.CurveTypeBN254)
	require.NoError(t, err)
	_, err = avsConfigCaller.ConfigureAVSOperatorSet(ctx, avsAddr, execOpsetId, config.CurveTypeECDSA)
	require.NoError(t, err)

	// Create generation reservations
	maxStalenessPeriod := uint32(0)
	bn254CalculatorAddr, err := caller.GetTableCalculatorAddress(config.CurveTypeBN254, config.ChainId_EthereumAnvil)
	if err != nil {
		t.Fatalf("Failed to create generation reservation for aggregator operator set: %v", err)
	}

	_, err = avsConfigCaller.CreateGenerationReservation(ctx, avsAddr, aggOpsetId, bn254CalculatorAddr, avsAddr, maxStalenessPeriod)
	require.NoError(t, err)

	ecdsaCalculatorAddr, err := caller.GetTableCalculatorAddress(config.CurveTypeECDSA, config.ChainId_EthereumAnvil)
	if err != nil {
		t.Fatalf("Failed to create generation reservation for aggregator operator set: %v", err)
	}
	_, err = avsConfigCaller.CreateGenerationReservation(ctx, avsAddr, execOpsetId, ecdsaCalculatorAddr, avsAddr, maxStalenessPeriod)
	require.NoError(t, err)

	// Setup operator peering
	err = testUtils.SetupOperatorPeering(
		ctx,
		chainConfig,
		config.ChainId(l1ChainId.Uint64()),
		l1EthClient,
		&operator.Operator{
			TransactionPrivateKey: chainConfig.OperatorAccountPrivateKey,
			SigningPrivateKey:     aggBn254PrivateSigningKey,
			Curve:                 config.CurveTypeBN254,
			OperatorSetIds:        []uint32{aggOpsetId},
		},
		&operator.Operator{
			TransactionPrivateKey: chainConfig.ExecOperatorAccountPk,
			SigningPrivateKey:     execEcdsaPrivateSigningKey,
			Curve:                 config.CurveTypeECDSA,
			OperatorSetIds:        []uint32{execOpsetId},
		},
		"localhost:9000",
		l,
	)
	require.NoError(t, err)

	// Delegate stake
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
	require.NoError(t, err)

	pdf := peeringDataFetcher.NewPeeringDataFetcher(l1CC, l)
	signers := signer.Signers{
		ECDSASigner: execSigner,
	}

	// Store performer ID for validation between phases
	var deployedPerformerID string
	var deployedContainerID string

	t.Run("Phase1_FirstExecutor", func(t *testing.T) {
		// Create Badger store
		store1, err := badger.NewBadgerExecutorStore(execConfig.Storage.BadgerConfig)
		require.NoError(t, err)

		exec1, err := NewExecutorWithRpcServers(execConfig.GrpcPort, execConfig.GrpcPort, execConfig, l, signers, pdf, l1CC, store1)
		require.NoError(t, err)

		err = exec1.Initialize(ctx)
		require.NoError(t, err)

		// Start executor in background
		exec1Ctx, exec1Cancel := context.WithCancel(ctx)
		go func() {
			if err := exec1.Run(exec1Ctx); err != nil {
				t.Logf("Executor 1 stopped: %v", err)
			}
		}()

		// Give containers time to start
		time.Sleep(5 * time.Second)

		// Submit task to first executor
		execClient, err := executorClient.NewExecutorClient(fmt.Sprintf("localhost:%d", execConfig.GrpcPort), true)
		require.NoError(t, err)

		taskId := "0x0000000000000000000000000000000000000000000000000000000000000001"
		strBlockNumber, err := l1EthereumClient.GetBlockNumber(ctx)
		require.NoError(t, err)
		blockNumber, err := strconv.ParseUint(strBlockNumber[2:], 16, 64)
		require.NoError(t, err)
		payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))

		encodedMessage, err := util.EncodeTaskSubmissionMessageVersioned(
			taskId,
			simAggConfig.Avss[0].Address,
			chainConfig.ExecOperatorAccountAddress,
			42,
			blockNumber,
			1,
			payloadJsonBytes,
			1,
		)
		require.NoError(t, err)

		// Sign message
		bn254Signer := inMemorySigner.NewInMemorySigner(aggBn254PrivateSigningKey, config.CurveTypeBN254)
		payloadSig, err := bn254Signer.SignMessage(encodedMessage[:])
		require.NoError(t, err)

		// Submit task
		taskResult, err := execClient.SubmitTask(ctx, &executorV1.TaskSubmission{
			TaskId:             taskId,
			AggregatorAddress:  simAggConfig.Operator.Address,
			AvsAddress:         simAggConfig.Avss[0].Address,
			ExecutorAddress:    chainConfig.ExecOperatorAccountAddress,
			Payload:            payloadJsonBytes,
			Signature:          payloadSig,
			TaskBlockNumber:    blockNumber,
			ReferenceTimestamp: 42,
			OperatorSetId:      1,
			Version:            1,
		})
		require.NoError(t, err)
		assert.NotNil(t, taskResult)
		t.Logf("Task %s executed successfully on first executor", taskResult.TaskId)

		// Capture performer state for validation
		states, err := store1.ListPerformerStates(ctx)
		require.NoError(t, err)
		require.Len(t, states, 1, "Should have one performer state persisted")

		deployedPerformerID = states[0].PerformerId
		deployedContainerID = states[0].ResourceId
		t.Logf("Captured performer ID: %s, container ID: %s", deployedPerformerID, deployedContainerID)

		// Gracefully shutdown first executor
		exec1Cancel()
		time.Sleep(2 * time.Second)

		// Close store to ensure all data is flushed
		err = store1.Close()
		require.NoError(t, err)
		t.Log("First executor shutdown complete, state persisted to disk")
	})

	t.Run("Phase2_SecondExecutor", func(t *testing.T) {
		// Create new Badger store instance with same directory
		store2, err := badger.NewBadgerExecutorStore(execConfig.Storage.BadgerConfig)
		require.NoError(t, err)
		defer store2.Close()

		// Verify persisted state is accessible
		states, err := store2.ListPerformerStates(ctx)
		require.NoError(t, err)
		require.Len(t, states, 1, "Should have one performer state from previous executor")
		assert.Equal(t, deployedPerformerID, states[0].PerformerId, "Should have same performer ID")
		assert.Equal(t, deployedContainerID, states[0].ResourceId, "Should have same container ID")

		// Use different port to avoid conflicts
		execConfig2 := *execConfig
		execConfig2.GrpcPort = 9091

		exec2, err := NewExecutorWithRpcServers(execConfig2.GrpcPort, execConfig2.GrpcPort, &execConfig2, l, signers, pdf, l1CC, store2)
		require.NoError(t, err)

		// Initialize - this should trigger rehydration
		err = exec2.Initialize(ctx)
		require.NoError(t, err)
		t.Log("Second executor initialized, rehydration should have occurred")

		// Start executor in background
		exec2Ctx, exec2Cancel := context.WithCancel(ctx)
		defer exec2Cancel()
		go func() {
			if err := exec2.Run(exec2Ctx); err != nil {
				t.Logf("Executor 2 stopped: %v", err)
			}
		}()

		// Wait for rehydration to complete
		time.Sleep(5 * time.Second)

		// Submit new task to verify rehydrated container works
		execClient2, err := executorClient.NewExecutorClient(fmt.Sprintf("localhost:%d", execConfig2.GrpcPort), true)
		require.NoError(t, err)

		taskId2 := "0x0000000000000000000000000000000000000000000000000000000000000002"
		strBlockNumber, err := l1EthereumClient.GetBlockNumber(ctx)
		require.NoError(t, err)
		blockNumber, err := strconv.ParseUint(strBlockNumber[2:], 16, 64)
		require.NoError(t, err)
		payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(5))

		encodedMessage, err := util.EncodeTaskSubmissionMessageVersioned(
			taskId2,
			simAggConfig.Avss[0].Address,
			chainConfig.ExecOperatorAccountAddress,
			43,
			blockNumber,
			1,
			payloadJsonBytes,
			1,
		)
		require.NoError(t, err)

		// Sign message
		bn254Signer := inMemorySigner.NewInMemorySigner(aggBn254PrivateSigningKey, config.CurveTypeBN254)
		payloadSig, err := bn254Signer.SignMessage(encodedMessage[:])
		require.NoError(t, err)

		// Submit task to rehydrated container
		taskResult, err := execClient2.SubmitTask(ctx, &executorV1.TaskSubmission{
			TaskId:             taskId2,
			AggregatorAddress:  simAggConfig.Operator.Address,
			AvsAddress:         simAggConfig.Avss[0].Address,
			ExecutorAddress:    chainConfig.ExecOperatorAccountAddress,
			Payload:            payloadJsonBytes,
			Signature:          payloadSig,
			TaskBlockNumber:    blockNumber,
			ReferenceTimestamp: 43,
			OperatorSetId:      1,
			Version:            1,
		})
		require.NoError(t, err)
		assert.NotNil(t, taskResult)
		t.Logf("Task %s executed successfully on rehydrated container", taskResult.TaskId)

		// Verify signature if ECDSA
		if curveType == config.CurveTypeECDSA {
			sig, err := cryptoLibsEcdsa.NewSignatureFromBytes(taskResult.ResultSignature)
			require.NoError(t, err)
			assert.NotNil(t, sig, "ECDSA signature should be valid")
			assert.Len(t, taskResult.ResultSignature, 65, "ECDSA signature should be 65 bytes")
			t.Log("Successfully verified signature from rehydrated container")
		}
	})

	t.Log("Badger storage test with hydration completed successfully")
}

func Test_Executor(t *testing.T) {
	// t.Run("BN254", func(t *testing.T) {
	// 	t.Skip("Executor is only setup as ECDSA for now")
	// 	testWithKeyType(t, config.CurveTypeBN254, executorConfigYaml, aggregatorConfigYaml)
	// })
	t.Run("ECDSA_InMemory", func(t *testing.T) {
		testWithKeyType(t, config.CurveTypeECDSA, executorConfigYaml, aggregatorConfigYaml)
	})
	t.Run("ECDSA_Badger", func(t *testing.T) {
		testWithBadgerStorage(t, config.CurveTypeECDSA, executorConfigYaml, aggregatorConfigYaml)
	})
}

const (
	executorConfigYaml = `
---
grpcPort: 9090
operator:
  address: "0x6ca766d180398847cEb1a58f03e029D65d88a878"
  operatorPrivateKey:
    privateKey: "0x71d173dbc3f00534ddd9bb0796c7df160c7d1af30e5aeeb601689c2026c244f6"
  signingKeys:
    bls:
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
                "salt": "ab09c61a936e1ad3fba62fa5798fe0cace021765b7da485662db23c7bb42df55"
              },
              "message": ""
            },
            "checksum": {
              "function": "sha256",
              "params": {},
              "message": "efc2d0baa038a0e5be02977b11e7cf55101abb348902b1b27e724bea89d96ace"
            },
            "cipher": {
              "function": "aes-128-ctr",
              "params": {
                "iv": "836eb161c0f87bb3ecccd81f9e8c7d9c"
              },
              "message": "3b288503f5ad1a7620fce9514b14f2dddb26427a5550f975a907c3de4190b985"
            }
          },
          "pubkey": "1a42c113530a95717e8d2b9d038d8ef87a028b9fd5ffc63e513742ec6e1c3aab0ccd2c8b00446a25fae666807b17e0c6fc18ef63bb754e44a8f81f8a6f6c6f171d17e059c183fc3f560fe537e3815911cbdefc80bc01aa89a536036264fec88d109c5a3f39e8b4b9a815172ad769f09922aba0879e2529d926a4b1476760c66b",
          "path": "m/1/0/0",
          "uuid": "34a93605-4399-409d-b2e6-bc51d8fadf21",
          "version": 4,
          "curveType": "bn254"
        }
      password: ""
l1Chain:
  rpcUrl: "http://localhost:8545"
  chainId: 31337
avsPerformers:
- image:
    repository: "hello-performer"
    tag: "latest"
  processType: "server"
  deploymentMode: "docker"
  avsAddress: "0xce2ac75be2e0951f1f7b288c7a6a9bfb6c331dc4"
  avsRegistrarAddress: "0xf4c5c29b14f0237131f7510a51684c8191f98e06"
`

	aggregatorConfigYaml = `
---
chains:
  - name: ethereum
    network: mainnet
    chainId: 31337
    rpcUrl: https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID
operator:
  address: "0xf701291C8276DFbc3A7ea29c13A16304Dbc9F845"
  operatorPrivateKey:
    privateKey: "0x97c00ea4a86647f8b5c885ec3a685999c4b59e8693dbe26172748084ba0deec7"
  signingKeys:
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
                "salt": "adf064e1f09d1001adef3fa51c40e0ed676103ca380ec30914d45d1e4a8ef1a3"
              },
              "message": ""
            },
            "checksum": {
              "function": "sha256",
              "params": {},
              "message": "dfc02d55ef7787fc6a01a76b247074dfed2750ec1e20fa19735d6d740ea78427"
            },
            "cipher": {
              "function": "aes-128-ctr",
              "params": {
                "iv": "14657c17d26112862db3b54ec5b214bd"
              },
              "message": "c65fff42226fade000cdde6f01894362e0f8908aa37a50e5434c2e2cd2d070b2"
            }
          },
          "pubkey": "110936c375f725bb6474b0facf2c243b8fa97f7066c51111049875d2674771402a1cd975e6b9d9bebf7e3e5f9d840e85d35afa22d3e6f6870f991f54fb137b132723a3d2d285a3d48ef9308a40e5a57d895816ed875e32152e4478d6d039e6c4114709c9fc9fdf80454b628307aea4f326d573a84ef5018d4551cf0879359650",
          "path": "m/1/0/0",
          "uuid": "894ceb28-b902-4062-be20-2e4d66606936",
          "version": 4,
          "curveType": "bn254"
        }

avss:
  - address: "0xce2ac75be2e0951f1f7b288c7a6a9bfb6c331dc4"
    privateKey: "some private key"
    privateSigningKey: "some private signing key"
    privateSigningKeyType: "ecdsa"
    responseTimeout: 3000
    chainIds: [31337]
    avsRegistrarAddress: "0xf4c5c29b14f0237131f7510a51684c8191f98e06"
`
)
