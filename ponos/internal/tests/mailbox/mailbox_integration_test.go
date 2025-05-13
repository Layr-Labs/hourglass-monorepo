package mailbox

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
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

const (
	RPCUrl = "http://127.0.0.1:8545"
)

func Test_EVMChainPollerIntegration(t *testing.T) {
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

	tlp := transactionLogParser.NewTransactionLogParser(imContractStore, l)

	ethereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   RPCUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)

	logsChan := make(chan *chainPoller2.LogWithBlock)

	poller := EVMChainPoller.NewEVMChainPoller(ethereumClient, logsChan, tlp, &EVMChainPoller.EVMChainPollerConfig{
		ChainId:                 config.ChainId_EthereumAnvil,
		PollingInterval:         time.Duration(10) * time.Second,
		EigenLayerCoreContracts: imContractStore.ListContractAddresses(),
		InterestingContracts:    []string{},
	}, l)

	ethClient, err := ethereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get Ethereum contract caller: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	anvil, err := testUtils.StartAnvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start Anvil: %v", err)
	}

	if os.Getenv("CI") == "" {
		fmt.Printf("Sleeping for 10 seconds\n\n")
		time.Sleep(10 * time.Second)
	} else {
		fmt.Printf("Sleeping for 30 seconds\n\n")
		time.Sleep(30 * time.Second)
	}
	fmt.Println("Checking if anvil is up and running...")

	// goes after anvil since it has to get the chain ID
	cc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.AppAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddress,
	}, ethClient, l)
	if err != nil {
		t.Fatalf("Failed to create contract caller: %v", err)
	}

	chainId, err := ethClient.ChainID(ctx)
	if err != nil {
		t.Fatalf("Failed to get chain ID: %v", err)
	}

	if err := poller.Start(ctx); err != nil {
		cancel()
		t.Fatalf("Failed to start EVM Chain Poller: %v", err)
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

			prettyBytes, _ := json.MarshalIndent(logWithBlock.Log, "", "  ")
			fmt.Printf("Log: %s\n", string(prettyBytes))

			assert.Equal(t, "TaskCreated", logWithBlock.Log.EventName)

			task, err := types.NewTaskFromLog(logWithBlock.Log, logWithBlock.Block, chainConfig.MailboxContractAddress)
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
				ChainId:         config.ChainId(chainId.Uint64()),
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
				TaskMailboxAddress:  chainConfig.MailboxContractAddress,
			}, ethClient, l)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to create contract caller: %v", err)
				cancel()
				return
			}

			fmt.Printf("Submitting task result to AVS\n\n\n")
			receipt, err := avsCc.SubmitTaskResult(ctx, cert)
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
	task, err := cc.PublishMessageToInbox(ctx, chainConfig.AVSAccountAddress, 1, payloadJsonBytes)
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

	_ = anvil.Process.Kill()
	assert.False(t, hasErrors)
}
