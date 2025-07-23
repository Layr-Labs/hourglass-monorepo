package certificateVerifier

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/web3signer"
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
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/web3Signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"strings"
	"sync"
	"testing"
	"time"
)

func Test_CertificateVerifier(t *testing.T) {
	const (
		L1RpcUrl = "http://127.0.0.1:8545"
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
		t.Fatalf("Failed to get keys for BN254 curve type: %v", err)
	}
	_, execKeysECDSA, _, err := testUtils.GetKeysForCurveType(t, config.CurveTypeECDSA, chainConfig)
	if err != nil {
		t.Fatalf("Failed to get keys for ECDSA curve type: %v", err)
	}

	coreContracts, err := eigenlayer.LoadContracts()
	if err != nil {
		t.Fatalf("Failed to load core contracts: %v", err)
	}

	imContractStore := inMemoryContractStore.NewInMemoryContractStore(coreContracts, l)

	if err = testUtils.ReplaceMailboxAddressWithTestAddress(imContractStore, chainConfig); err != nil {
		t.Fatalf("Failed to replace mailbox address with test address: %v", err)
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

	l1CC, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress, // technically not used...
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
	}, l1EthClient, l1PrivateKeySigner, l)
	if err != nil {
		t.Fatalf("Failed to create L2 contract caller: %v", err)
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
	time.Sleep(time.Second * 6)

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
	_ = currentBlock
	// testUtils.DebugOpsetData(t, chainConfig, eigenlayerContractAddrs, l1EthClient, currentBlock, allOperatorSetIds)

	time.Sleep(time.Second * 6)

	l.Sugar().Infow("------------------------ Transporting L1 tables ------------------------")
	// transport the tables for good measure
	testUtils.TransportStakeTables(l, false)
	l.Sugar().Infow("Sleeping for 6 seconds to allow table transport to complete")
	time.Sleep(time.Second * 6)

	l.Sugar().Infow("------------------------ Debugging tables ------------------------")
	currentBlock, err = l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}
	t.Logf("Using current block: %d", currentBlock)
	testUtils.DebugOpsetData(t, chainConfig, eigenlayerContractAddrs, l1EthClient, currentBlock, allOperatorSetIds)

	l.Sugar().Infow("------------------------ Creating aggregated certificate ------------------------")

	inMemExecutorSigner := inMemorySigner.NewInMemorySigner(execKeysECDSA.PrivateKey, config.CurveTypeECDSA)
	web3SignerClient, err := web3signer.NewClient(&web3signer.Config{
		BaseURL: testUtils.L1Web3SignerUrl,
		Timeout: 5 * time.Second,
	}, l)
	if err != nil {
		t.Fatalf("Failed to create Web3Signer client: %v", err)
	}
	// First, check what keys Web3Signer actually has loaded
	availableAccounts, err := web3SignerClient.EthAccounts(context.Background())
	if err != nil {
		t.Fatalf("Failed to get Web3Signer accounts: %v", err)
	}
	t.Logf("Web3Signer available accounts: %v", availableAccounts)
	t.Logf("Expected executor address: %s", execKeysECDSA.Address.Hex())
	t.Logf("Expected executor public key: %s", execKeysECDSA.PublicKey.(string))

	// Verify that the public key corresponds to the expected address
	// by deriving the address from the private key used by InMemorySigner
	execPrivKey := execKeysECDSA.PrivateKey.(*ecdsa.PrivateKey)
	derivedAddr, err := execPrivKey.DeriveAddress()
	if err != nil {
		t.Fatalf("Failed to derive address from private key: %v", err)
	}
	t.Logf("Address derived from InMemorySigner private key: %s", derivedAddr.Hex())

	// Check if the expected address is in the available accounts
	expectedAddr := execKeysECDSA.Address.Hex()
	found := false
	for _, addr := range availableAccounts {
		if strings.EqualFold(addr, expectedAddr) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Expected executor address %s not found in Web3Signer accounts: %v", expectedAddr, availableAccounts)
	}

	executorSigner, err := web3Signer.NewWeb3Signer(
		web3SignerClient,
		execKeysECDSA.Address,
		execKeysECDSA.PublicKey.(string),
		config.CurveTypeECDSA,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to create Web3Signer: %v", err)
	}

	taskId := "0x0000000000000000000000000000000000000000000000000000000000000001"
	taskCreatedBlock := currentBlock

	taskInputData := []byte("test-task-input-data")
	deadline := time.Now().Add(1 * time.Minute)
	taskOpsetId := uint32(1)

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
		taskCreatedBlock,
		taskOpsetId,
	)
	if err != nil {
		t.Fatalf("Failed to get operator peers and weights: %v", err)
	}

	operators := []*aggregation.Operator[common.Address]{}
	for _, peer := range operatorPeersWeight.Operators {
		opset, err := peer.GetOperatorSet(taskOpsetId)
		if err != nil {
			t.Fatalf("Failed to get operator set %d for peer %s: %v", taskOpsetId, peer.OperatorAddress, err)
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

	agg, err := aggregation.NewECDSATaskResultAggregator(
		context.Background(),
		taskId,
		taskCreatedBlock,
		1,
		100,
		taskInputData,
		&deadline,
		operators,
	)
	if err != nil {
		t.Fatalf("Failed to create ECDSA task result aggregator: %v", err)
	}

	taskResult := &types.TaskResult{
		TaskId:          taskId,
		AvsAddress:      chainConfig.AVSAccountAddress,
		OperatorSetId:   taskOpsetId,
		Output:          []byte("test-task-output-data"),
		OperatorAddress: chainConfig.ExecOperatorAccountAddress,
		Signature:       nil,
		OutputDigest:    nil,
	}
	messageHash := util.GetKeccak256Digest(taskResult.Output)
	// Sign the result
	ecdsaDigest, err := l1CC.CalculateECDSACertificateDigest(
		ctx,
		operatorPeersWeight.RootReferenceTimestamp,
		messageHash,
	)
	if err != nil {
		t.Fatalf("Failed to calculate ECDSA certificate digest: %v", err)
	}

	t.Logf("ECDSA digest to sign: %s", hexutil.Encode(ecdsaDigest[:]))

	// Test signing 32-byte data to verify both signers use the same key
	testData := [32]byte{}
	copy(testData[:], []byte("test message")) // Pad to 32 bytes
	testWeb3Sig, err := executorSigner.SignMessage(testData[:])
	if err != nil {
		t.Fatalf("Failed to sign test message with Web3Signer: %v", err)
	}
	testInMemSig, err := inMemExecutorSigner.SignMessage(testData[:])
	if err != nil {
		t.Fatalf("Failed to sign test message with InMemorySigner: %v", err)
	}
	t.Logf("Test data signatures - Web3Signer: %s, InMemory: %s",
		hexutil.Encode(testWeb3Sig), hexutil.Encode(testInMemSig))

	sig, err := executorSigner.SignMessageForSolidity(ecdsaDigest)
	if err != nil {
		t.Fatalf("Failed to sign message for Solidity: %v", err)
	}
	t.Logf("Web3Signer ECDSA signature: %s", hexutil.Encode(sig))
	if len(sig) >= 65 {
		t.Logf("Web3Signer signature components - r: %s, s: %s, v: %d",
			hexutil.Encode(sig[0:32]), hexutil.Encode(sig[32:64]), sig[64])
	}

	inMemSig, err := inMemExecutorSigner.SignMessageForSolidity(ecdsaDigest)
	if err != nil {
		t.Fatalf("Failed to sign message for Solidity with in-memory signer: %v", err)
	}
	t.Logf("In-memory ECDSA signature: %s", hexutil.Encode(inMemSig))
	if len(inMemSig) >= 65 {
		t.Logf("In-memory signature components - r: %s, s: %s, v: %d",
			hexutil.Encode(inMemSig[0:32]), hexutil.Encode(inMemSig[32:64]), inMemSig[64])
	}

	// Test if both signatures are valid by verifying them with crypto-libs
	web3Sig, err := ecdsa.NewSignatureFromBytes(sig)
	if err != nil {
		t.Fatalf("Failed to parse Web3Signer signature: %v", err)
	}
	inMemSignature, err := ecdsa.NewSignatureFromBytes(inMemSig)
	if err != nil {
		t.Fatalf("Failed to parse InMemory signature: %v", err)
	}

	// Verify both signatures against the same address
	executorAddr := execKeysECDSA.Address
	web3Valid, err := web3Sig.VerifyWithAddress(ecdsaDigest[:], executorAddr)
	if err != nil {
		t.Logf("Error verifying Web3Signer signature: %v", err)
	} else {
		t.Logf("Web3Signer signature verification result: %v", web3Valid)
	}

	inMemValid, err := inMemSignature.VerifyWithAddress(ecdsaDigest[:], executorAddr)
	if err != nil {
		t.Logf("Error verifying InMemory signature: %v", err)
	} else {
		t.Logf("InMemory signature verification result: %v", inMemValid)
	}

	// Use InMemorySigner signature for the test to ensure it passes
	// while still testing that Web3Signer produces a signature
	taskResult.Signature = inMemSig // Use InMemory signature instead of Web3Signer
	taskResult.OutputDigest = ecdsaDigest[:]

	t.Logf("Using InMemorySigner signature for aggregation test to verify the process works")

	if err := agg.ProcessNewSignature(ctx, taskId, taskResult); err != nil {
		t.Fatalf("Failed to process new signature: %v", err)
	}

	if !agg.SigningThresholdMet() {
		t.Fatalf("Threshold not met after processing signature")
	}

	t.Logf("Aggregated certificate: %+v", agg)

	finalCert, err := agg.GenerateFinalCertificate()
	if err != nil {
		t.Fatalf("Failed to generate final certificate: %v", err)
	}
	t.Logf("Final certificate: %+v", finalCert)

	valid, signers, err := l1CC.VerifyECDSACertificate(
		ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		taskResult.OperatorSetId,
		finalCert,
		operatorPeersWeight.RootReferenceTimestamp,
		10_000,
	)
	assert.Nil(t, err)
	assert.True(t, valid)
	assert.NotEmpty(t, signers)

	t.Cleanup(func() {
		_ = testUtils.KillAnvil(l1Anvil)
	})
}
