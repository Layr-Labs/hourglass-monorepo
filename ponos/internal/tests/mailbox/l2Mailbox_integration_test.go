package mailbox

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
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
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"testing"
	"time"
)

func Test_L2Mailbox(t *testing.T) {
	const (
		L1RpcUrl = "http://127.0.0.1:8545"
		L2RpcUrl = "http://127.0.0.1:9545"
	)

	t.Skip()
	// t.Skip("Flaky, skipping for now")
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
		PollingInterval:      time.Duration(10) * time.Second,
		InterestingContracts: imContractStore.ListContractAddressesForChain(config.ChainId_EthereumAnvil),
	}, l)

	l2Poller := EVMChainPoller.NewEVMChainPoller(l2EthereumClient, logsChan, tlp, &EVMChainPoller.EVMChainPollerConfig{
		ChainId:              config.ChainId_EthereumAnvil,
		PollingInterval:      time.Duration(10) * time.Second,
		InterestingContracts: imContractStore.ListContractAddressesForChain(config.ChainId_EthereumAnvil),
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

	l1Anvil, err := testUtils.StartL1Anvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start L1 Anvil: %v", err)
	}

	l2Anvil, err := testUtils.StartL2Anvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start L2 Anvil: %v", err)
	}

	if os.Getenv("CI") == "" {
		fmt.Printf("Sleeping for 10 seconds\n\n")
		time.Sleep(10 * time.Second)
	} else {
		fmt.Printf("Sleeping for 30 seconds\n\n")
		time.Sleep(30 * time.Second)
	}
	fmt.Println("Checking if l1Anvil is up and running...")

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

	l2CC, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.AppAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress, // technically not used...
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL2,
	}, l2EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create L2 contract caller: %v", err)
	}

	if err := l1Poller.Start(ctx); err != nil {
		cancel()
		t.Fatalf("Failed to start EVM L1Chain Poller: %v", err)
	}
	if err := l2Poller.Start(ctx); err != nil {
		cancel()
		t.Fatalf("Failed to start EVM L2Chain Poller: %v", err)
	}

	execPrivateKey, execPublicKey, err := bn254.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

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

			sig, err := signer.SignMessage(digest[:])
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
				ChainId:         config.ChainId(l2ChainId.Uint64()),
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
			signedAt := time.Unix(int64(logWithBlock.Block.Timestamp.Value()), 0).Add(10 * time.Second)
			cert.SignedAt = &signedAt
			fmt.Printf("cert: %+v\n", cert)

			time.Sleep(10 * time.Second)

			avsCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
				PrivateKey:          chainConfig.AVSAccountPrivateKey,
				AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
				TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
			}, l2EthClient, l)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to create contract caller: %v", err)
				cancel()
				return
			}

			fmt.Printf("Submitting task result to AVS\n\n\n")
			// TODO(seanmcgary): use the global root timestamp here
			receipt, err := avsCc.SubmitTaskResult(ctx, cert, uint32(0))
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
	case <-time.After(90 * time.Second):
		cancel()
		t.Fatalf("Test timed out after 10 seconds")
	case <-ctx.Done():
		t.Logf("Test completed")
	}

	_ = l1Anvil.Process.Kill()
	_ = l2Anvil.Process.Kill()
	assert.False(t, hasErrors)
}
