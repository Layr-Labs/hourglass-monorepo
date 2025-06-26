package mailbox

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IBN254CertificateVerifier"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	chainPoller2 "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/EVMChainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"math/big"
	"sync"
	"testing"
	"time"
)

func Test_L2Mailbox(t *testing.T) {
	// t.Skip()
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

	aggPrivateKey, _, err := bn254.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	execPrivateKey, execPublicKey, err := bn254.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
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

	logsChan := make(chan *chainPoller2.LogWithBlock)

	l1Poller := EVMChainPoller.NewEVMChainPoller(l1EthereumClient, logsChan, tlp, &EVMChainPoller.EVMChainPollerConfig{
		ChainId:              config.ChainId_EthereumAnvil,
		PollingInterval:      time.Duration(5) * time.Second,
		InterestingContracts: imContractStore.ListContractAddressesForChain(config.ChainId_EthereumAnvil),
	}, l)
	_ = l1Poller

	l2Poller := EVMChainPoller.NewEVMChainPoller(l2EthereumClient, logsChan, tlp, &EVMChainPoller.EVMChainPollerConfig{
		ChainId:              config.ChainId_BaseSepoliaAnvil,
		PollingInterval:      time.Duration(5) * time.Second,
		InterestingContracts: imContractStore.ListContractAddressesForChain(config.ChainId_BaseSepoliaAnvil),
	}, l)

	l1EthClient, err := l1EthereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get Ethereum contract caller: %v", err)
	}
	l2EthClient, err := l2EthereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get Ethereum contract caller: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	anvilWg := &sync.WaitGroup{}

	anvilWg.Add(1)
	l1Anvil, err := testUtils.StartL1Anvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start L1 Anvil: %v", err)
	}

	anvilWg.Add(1)
	l2Anvil, err := testUtils.StartL2Anvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start L2 Anvil: %v", err)
	}

	startErrorsChan := make(chan error, 2)
	anvilCtx, anvilCancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer anvilCancel()

	go testUtils.WaitForAnvil(anvilWg, anvilCtx, t, l1EthereumClient, startErrorsChan)
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
		t.Fatalf("Failed to get L1 chain ID: %v", err)
	}
	t.Logf("L1 Chain ID: %s", l1ChainId.String())

	l2ChainId, err := l2EthClient.ChainID(ctx)
	if err != nil {
		t.Fatalf("Failed to get L2 chain ID: %v", err)
	}
	t.Logf("L2 Chain ID: %s", l2ChainId.String())

	l1EigenlayerContractAddrs, err := config.GetCoreContractsForChainId(config.ChainId(l1ChainId.Uint64()))
	if err != nil {
		t.Fatalf("Failed to get core contracts for chain ID: %v", err)
	}

	l1CC, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:                chainConfig.AppAccountPrivateKey,
		AVSRegistrarAddress:       chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:        chainConfig.MailboxContractAddressL1,
		CrossChainRegistryAddress: l1EigenlayerContractAddrs.CrossChainRegistry,
		KeyRegistrarAddress:       l1EigenlayerContractAddrs.KeyRegistrar,
	}, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create L1 contract caller: %v", err)
	}

	l2CC, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:         chainConfig.AppAccountPrivateKey,
		TaskMailboxAddress: chainConfig.MailboxContractAddressL2,
	}, l2EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create L2 contract caller: %v", err)
	}

	reservations, err := l1CC.GetActiveGenerationReservations()
	if err != nil {
		t.Fatalf("Failed to get active L1 generation reservations: %v", err)
	}
	for _, reservation := range reservations {
		fmt.Printf("L1 active generation reservation: %+v\n", reservation)
	}

	l.Sugar().Infow("Setting up operator peering",
		zap.String("AVSAccountAddress", chainConfig.AVSAccountAddress),
	)
	err = testUtils.SetupOperatorPeering(
		ctx,
		chainConfig,
		config.ChainId(l1ChainId.Uint64()),
		l1EthClient,
		aggPrivateKey,
		execPrivateKey,
		"localhost:9000",
		l,
	)
	if err != nil {
		t.Fatalf("Failed to set up operator peering: %v", err)
	}

	currentL1Block, err := l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}

	debugOpsetData(t, chainConfig, l1EigenlayerContractAddrs, l1EthClient, currentL1Block)

	l.Sugar().Infow("------------------------ Transporting L1 & L2 tables ------------------------")
	// transport the tables for good measure
	testUtils.TransportStakeTables(l, true)
	l.Sugar().Infow("Sleeping for 6 seconds to allow table transport to complete")
	time.Sleep(time.Second * 6)

	l.Sugar().Infow("------------------------ Setting up mailbox ------------------------")
	avsCcL1, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.AVSAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
	}, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create AVS contract caller: %v", err)
	}
	err = testUtils.SetupTaskMailbox(
		ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		common.HexToAddress(chainConfig.AVSTaskHookAddressL1),
		[]uint32{1},
		[]string{"bn254"},
		avsCcL1,
	)
	if err != nil {
		t.Fatalf("Failed to set up task mailbox: %v", err)
	}

	avsCcL2, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.AVSAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL2,
	}, l2EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create AVS contract caller: %v", err)
	}
	err = testUtils.SetupTaskMailbox(
		ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		common.HexToAddress(chainConfig.AVSTaskHookAddressL2),
		[]uint32{1},
		[]string{"bn254"},
		avsCcL2,
	)
	if err != nil {
		t.Fatalf("Failed to set up task mailbox: %v", err)
	}

	// update current block to account for transport
	currentL1Block, err = l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}

	l.Sugar().Infow("------------------------ Starting pollers ------------------------")
	// if err := l1Poller.Start(ctx); err != nil {
	// 	cancel()
	// 	t.Fatalf("Failed to start EVM L1Chain Poller: %v", err)
	// }
	if err := l2Poller.Start(ctx); err != nil {
		cancel()
		t.Fatalf("Failed to start EVM L2Chain Poller: %v", err)
	}

	tableData, err := l1CC.GetOperatorTableDataForOperatorSet(
		ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		1,
		config.ChainId(l1ChainId.Uint64()),
		currentL1Block,
	)
	if err != nil {
		t.Fatalf("Failed to get operator table data: %v", err)
	}
	fmt.Printf("Operator table data: %+v\n", tableData)

	currentL2Block, err := l2EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}
	fmt.Printf("TableUpdaterAddresses: %+v\n", tableData.TableUpdaterAddresses)
	l2TableUpdaterAddress, ok := tableData.TableUpdaterAddresses[l2ChainId.Uint64()]
	if !ok {
		t.Fatalf("l2 table updater address not found")
	}

	// need to get the reference block number/timestamp from the table updater to know when it was last updated.
	// its entirely possible that the global table root was updated but the table was not.
	latestReferenceTimeAndBlock, err := l2CC.GetTableUpdaterReferenceTimeAndBlock(ctx, l2TableUpdaterAddress, currentL2Block)
	if err != nil {
		t.Fatalf("Failed to get latest reference time and block: %v", err)
	}
	fmt.Printf("Latest reference time and block: %+v\n", latestReferenceTimeAndBlock)

	cv, err := IBN254CertificateVerifier.NewIBN254CertificateVerifier(common.HexToAddress("0x824604a31b580Aec16D8Dd7ae9A27661Dc65cBA3"), l2EthClient)
	if err != nil {
		t.Fatalf("Failed to create certificate verifier: %v", err)
	}

	opsetInfo, err := cv.GetOperatorSetInfo(&bind.CallOpts{
		Context: ctx,
	}, IBN254CertificateVerifier.OperatorSet{
		Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		Id:  1,
	}, latestReferenceTimeAndBlock.LatestReferenceTimestamp)
	if err != nil {
		t.Fatalf("Failed to get operator set info: %v", err)
	}
	fmt.Printf("\n\nOperatorset info: %+v\n\n", opsetInfo)

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

			operators := []*aggregation.Operator{
				{
					Address:   chainConfig.ExecOperatorAccountAddress,
					PublicKey: execPublicKey,
				},
			}

			resultAgg, err := aggregation.NewTaskResultAggregator(
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

			outputResult := util.BigIntToHex(new(big.Int).SetUint64(16))
			signer := inMemorySigner.NewInMemorySigner(execPrivateKey)
			digest := util.GetKeccak256Digest(outputResult)

			sig, err := signer.SignMessageForSolidity(digest)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to sign message: %v", err)
				cancel()
				return
			}

			taskResult := &types.TaskResult{
				TaskId:          task.TaskId,
				AvsAddress:      chainConfig.AVSAccountAddress,
				CallbackAddr:    chainConfig.AVSAccountAddress,
				OperatorSetId:   1,
				Output:          outputResult,
				ChainId:         task.ChainId,
				BlockNumber:     task.BlockNumber,
				BlockHash:       task.BlockHash,
				OperatorAddress: chainConfig.ExecOperatorAccountAddress,
				Signature:       sig,
			}
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
			signedAt := time.Now()
			cert.SignedAt = &signedAt
			fmt.Printf("cert: %+v\n", cert)

			time.Sleep(10 * time.Second)

			avsCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
				PrivateKey:          chainConfig.AVSAccountPrivateKey,
				AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
				TaskMailboxAddress:  chainConfig.MailboxContractAddressL2,
			}, l2EthClient, l)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to create contract caller: %v", err)
				cancel()
				return
			}

			fmt.Printf("Submitting task result to AVS\n\n\n")
			fmt.Printf("Using timestamp: %d\n", tableData.LatestReferenceTimestamp)
			receipt, err := avsCc.SubmitTaskResult(ctx, cert, tableData.LatestReferenceTimestamp)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to submit task result: %v", err)
				cancel()
				return
			}
			assert.Nil(t, err)
			fmt.Printf("Receipt: %+v\n", receipt)

			cancel()
		}
	}()

	// submit a task
	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))
	task, err := l2CC.PublishMessageToInbox(ctx, chainConfig.AVSAccountAddress, 1, payloadJsonBytes)
	if err != nil {
		t.Fatalf("Failed to publish message to inbox: %v", err)
	}
	t.Logf("Task published: %+v", task)

	select {
	case <-time.After(150 * time.Second):
		cancel()
		t.Fatalf("Test timed out after 10 seconds")
	case <-ctx.Done():
		t.Logf("Test completed")
	}

	_ = l1Anvil.Process.Kill()
	_ = l2Anvil.Process.Kill()
	assert.False(t, hasErrors)
}
