package EVMChainPoller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	chainPoller2 "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

const (
	RPCUrl = "http://localhost:8545"
)

func Test_EVMChainPollerIntegration(t *testing.T) {
	t.Skip("Flaky, skipping for now")
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

	if err := os.Setenv(config.MAILBOX_CONTRACT_ADDRESS_OVERRIDE, chainConfig.MailboxContractAddress); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}

	mbJson, err := testUtils.ReadMailboxAbiJson(root)
	if err != nil {
		t.Fatalf("Failed to read mailbox ABI json: %v", err)
	}

	coreContracts, err := eigenlayer.LoadCoreContractsForL1Chain(config.ChainId_EthereumMainnet)
	if err != nil {
		t.Fatalf("Failed to load core contracts: %v", err)
	}

	// manually inject the mailbox contract
	mailboxContract := &contracts.Contract{
		Address:     chainConfig.MailboxContractAddress,
		AbiVersions: []string{string(mbJson)},
	}
	coreContracts[chainConfig.MailboxContractAddress] = mailboxContract

	imContractStore := inMemoryContractStore.NewInMemoryContractStore(coreContracts, l)

	tlp := transactionLogParser.NewTransactionLogParser(imContractStore, l)

	ethereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   RPCUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)

	logsChan := make(chan *chainPoller2.LogWithBlock)

	poller := NewEVMChainPoller(ethereumClient, logsChan, tlp, &EVMChainPollerConfig{
		ChainId:                 config.ChainId_EthereumMainnet,
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

	// goes after anvil since it has to get the chain ID
	cc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.AppAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddress,
	}, ethClient, l)
	if err != nil {
		t.Fatalf("Failed to create contract caller: %v", err)
	}

	if err := poller.Start(ctx); err != nil {
		cancel()
		t.Fatalf("Failed to start EVM Chain Poller: %v", err)
	}

	hasErrors := false
	go func() {
		for log := range logsChan {
			fmt.Printf("Received log: %+v\n", log.Log)
			if log.Log.EventName != "TaskCreated" {
				continue
			}
			assert.Equal(t, "TaskCreated", log.Log.EventName)
			assert.True(t, len(log.Log.Arguments[1].Value.(string)) > 0)
			assert.Equal(t, common.HexToAddress(chainConfig.AVSAccountAddress), log.Log.Arguments[2].Value.(common.Address))

			outputBytes, err := json.Marshal(log.Log.OutputData)
			if err != nil {
				t.Logf("Failed to marshal output data: %v", err)
				hasErrors = true
				return
			}
			type outputDataType struct {
				ExecutorOperatorSetId uint32
				TaskDeadline          uint64
				Payload               []byte
			}
			var od *outputDataType
			if err := json.Unmarshal(outputBytes, &od); err != nil {
				t.Logf("Failed to unmarshal output data: %v", err)
				hasErrors = true
				return
			}
			assert.True(t, len(od.Payload) > 0)

			cancel()
		}
	}()

	// submit a task
	payloadJsonBytes := []byte(`{ "numberToBeSquared": 4 }`)
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
