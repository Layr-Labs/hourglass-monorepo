package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	goEthereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"sync"
	"testing"
	"time"
)

const (
	L1RPCUrl = "http://127.0.0.1:8545"
	L2RPCUrl = "http://127.0.0.1:9545"
	L2WSUrl  = "ws://127.0.0.1:9545"
)

// Test_Aggregator is an integration test for the Aggregator component of the system.
//
// This test is designed to simulate an E2E on-chain flow with all components.
// - Both the aggreagator and executor are registered as operators with the AllocationManager/AVSRegistrar with their peering data
// - The executor is started and boots up the performers
// - The aggregator is started with a poller calling a local anvil node
// - The test pushes a message to the mailbox and waits for the TaskVerified event to be emitted
func Test_Aggregator(t *testing.T) {
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

	// ------------------------------------------------------------------------
	// Executor setup
	// ------------------------------------------------------------------------
	execConfig, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(executorConfigYaml))
	if err != nil {
		t.Fatalf("failed to create executor config: %v", err)
	}
	if err := execConfig.Validate(); err != nil {
		t.Fatalf("failed to validate executor config: %v", err)
	}

	execConfig.Operator.Address = chainConfig.ExecOperatorAccountAddress
	execConfig.Operator.OperatorPrivateKey = chainConfig.ExecOperatorAccountPk
	execConfig.AvsPerformers[0].AvsAddress = chainConfig.AVSAccountAddress

	storedKeys, err := keystore.ParseKeystoreJSON(execConfig.Operator.SigningKeys.BLS.Keystore)
	if err != nil {
		t.Fatalf("failed to parse keystore JSON: %v", err)
	}

	execPrivateSigningKey, err := storedKeys.GetBN254PrivateKey(execConfig.Operator.SigningKeys.BLS.Password)
	if err != nil {
		t.Fatalf("failed to get private key: %v", err)
	}
	execSigner := inMemorySigner.NewInMemorySigner(execPrivateSigningKey)

	// ------------------------------------------------------------------------
	// Aggregator setup
	// ------------------------------------------------------------------------
	aggConfig, err := aggregatorConfig.NewAggregatorConfigFromYamlBytes([]byte(aggregatorConfigYaml))
	if err != nil {
		t.Fatalf("Failed to create aggregator config: %v", err)
	}
	if err := aggConfig.Validate(); err != nil {
		t.Fatalf("Failed to validate aggregator config: %v", err)
	}

	aggConfig.Operator.Address = chainConfig.OperatorAccountAddress
	aggConfig.Operator.OperatorPrivateKey = chainConfig.OperatorAccountPrivateKey
	aggConfig.Avss[0].Address = chainConfig.AVSAccountAddress

	// run the AVS on the L2 (base)
	aggConfig.Avss[0].ChainIds = []uint{uint(config.ChainId_BaseAnvil)}

	aggStoredKeys, err := keystore.ParseKeystoreJSON(aggConfig.Operator.SigningKeys.BLS.Keystore)
	if err != nil {
		t.Fatalf("failed to parse keystore JSON: %v", err)
	}

	aggPrivateSigningKey, err := aggStoredKeys.GetBN254PrivateKey(aggConfig.Operator.SigningKeys.BLS.Password)
	if err != nil {
		t.Fatalf("failed to get private key: %v", err)
	}
	aggSigner := inMemorySigner.NewInMemorySigner(aggPrivateSigningKey)

	// ------------------------------------------------------------------------
	// L1Chain & l1Anvil setup
	// ------------------------------------------------------------------------
	anvilWg := &sync.WaitGroup{}

	l1EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   L1RPCUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)

	l1EthClient, err := l1EthereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get Ethereum contract caller: %v", err)
	}

	anvilWg.Add(1)
	l1Anvil, err := testUtils.StartL1Anvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start Anvil: %v", err)
	}

	go func() {
		defer anvilWg.Done()
		if os.Getenv("CI") == "" {
			fmt.Printf("Sleeping for 10 seconds\n\n")
			time.Sleep(10 * time.Second)
		} else {
			fmt.Printf("Sleeping for 30 seconds\n\n")
			time.Sleep(30 * time.Second)
		}
		fmt.Println("Checking if l1Anvil is up and running...")
	}()

	// ------------------------------------------------------------------------
	// L2Chain & l1Anvil setup
	// ------------------------------------------------------------------------
	anvilWg.Add(1)
	l2EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   L2RPCUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)

	l2EthClient, err := l2EthereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get Ethereum contract caller: %v", err)
	}

	l2Anvil, err := testUtils.StartL2Anvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start L2 Anvil: %v", err)
	}
	go func() {
		defer anvilWg.Done()
		if os.Getenv("CI") == "" {
			fmt.Printf("Sleeping for 10 seconds\n\n")
			time.Sleep(10 * time.Second)
		} else {
			fmt.Printf("Sleeping for 30 seconds\n\n")
			time.Sleep(30 * time.Second)
		}
		fmt.Println("Checking if l2Anvil is up and running...")
	}()
	anvilWg.Wait()

	// ------------------------------------------------------------------------
	// register peering data
	// ------------------------------------------------------------------------
	l1AggCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.OperatorAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
	}, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create contract caller: %v", err)
	}
	t.Logf("Registering avs peering data")
	_, err = operator.RegisterOperatorToOperatorSets(
		ctx,
		l1AggCc,
		common.HexToAddress(aggConfig.Operator.Address),
		common.HexToAddress(aggConfig.Avss[0].Address),
		[]uint32{0},
		aggPrivateSigningKey.Public(),
		aggPrivateSigningKey,
		"",
		7200,
		"http://something.com",
		l,
	)
	if err != nil {
		t.Fatalf("failed to register aggregator operator to operator sets: %v", err)
	}

	l1ExecCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.ExecOperatorAccountPk,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
	}, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create contract caller: %v", err)
	}
	t.Logf("Registering operator peering data")
	_, err = operator.RegisterOperatorToOperatorSets(
		ctx,
		l1ExecCc,
		common.HexToAddress(execConfig.Operator.Address),
		common.HexToAddress(execConfig.AvsPerformers[0].AvsAddress),
		[]uint32{1},
		execPrivateSigningKey.Public(),
		execPrivateSigningKey,
		fmt.Sprintf("localhost:%d", execConfig.GrpcPort),
		7200,
		"http://something.com",
		l,
	)
	if err != nil {
		t.Fatalf("failed to register executor operator to operator sets: %v", err)
	}

	// ------------------------------------------------------------------------
	// Setup the executor
	// ------------------------------------------------------------------------
	execPdf := peeringDataFetcher.NewPeeringDataFetcher(l1ExecCc, l)
	exec, err := executor.NewExecutorWithRpcServer(execConfig.GrpcPort, execConfig, l, execSigner, execPdf)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	_ = exec

	// ------------------------------------------------------------------------
	// Setup the aggregator
	// ------------------------------------------------------------------------
	coreContracts, err := eigenlayer.LoadContracts()
	if err != nil {
		t.Fatalf("failed to load contracts: %v", err)
	}

	imContractStore := inMemoryContractStore.NewInMemoryContractStore(coreContracts, l)

	if err = testUtils.ReplaceMailboxAddressWithTestAddress(imContractStore, chainConfig); err != nil {
		t.Fatalf("Failed to replace mailbox address with test address: %v", err)
	}

	tlp := transactionLogParser.NewTransactionLogParser(imContractStore, l)
	aggPdf := peeringDataFetcher.NewPeeringDataFetcher(l1AggCc, l)

	agg, err := NewAggregatorWithRpcServer(
		aggConfig.ServerConfig.Port,
		&AggregatorConfig{
			AVSs:          aggConfig.Avss,
			Chains:        aggConfig.Chains,
			Address:       aggConfig.Operator.Address,
			PrivateKey:    aggConfig.Operator.OperatorPrivateKey,
			AggregatorUrl: aggConfig.ServerConfig.AggregatorUrl,
		},
		imContractStore,
		tlp,
		aggPdf,
		aggSigner,
		l,
	)
	if err != nil {
		t.Logf("Failed to create aggregator: %v", err)
	}
	_ = agg

	// ------------------------------------------------------------------------
	// Boot up everything
	// ------------------------------------------------------------------------
	if err := exec.Initialize(); err != nil {
		t.Logf("Failed to initialize executor: %v", err)
		cancel()
	}

	if err := exec.BootPerformers(ctx); err != nil {
		t.Logf("Failed to boot performers: %v", err)
		cancel()
	}
	if err := exec.Run(ctx); err != nil {
		t.Logf("Failed to run executor: %v", err)
		cancel()
	}

	if err := agg.Initialize(); err != nil {
		cancel()
		t.Logf("Failed to initialize aggregator: %v", err)

	}

	go func() {
		if err := agg.Start(ctx); err != nil {
			cancel()
		}
	}()

	// ------------------------------------------------------------------------
	// Push a message to the mailbox
	// ------------------------------------------------------------------------
	l2AppCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.AppAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL2,
	}, l2EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create contract caller: %v", err)
	}
	t.Logf("Pushing message to mailbox...")
	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))
	task, err := l2AppCc.PublishMessageToInbox(ctx, chainConfig.AVSAccountAddress, 1, payloadJsonBytes)
	if err != nil {
		t.Fatalf("Failed to publish message to inbox: %v", err)
	}
	t.Logf("Task published: %+v", task)

	// ------------------------------------------------------------------------
	// Listen for TaskVerified event to know that the test is done
	// ------------------------------------------------------------------------
	wsEthClient, err := l2EthereumClient.GetWebsocketConnection(L2WSUrl)
	if err != nil {
		t.Fatalf("Failed to get websocket connection: %v", err)
	}

	taskVerified := false

	eventsChan := make(chan types.Log)
	sub, err := wsEthClient.SubscribeFilterLogs(ctx, goEthereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(chainConfig.MailboxContractAddressL2)},
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
					cancel()
				}

			}
		}
	}()

	select {
	case <-ctx.Done():
		t.Logf("Context done: %v", ctx.Err())
	case <-time.After(90 * time.Second):
		t.Logf("Timeout after 60 seconds")
		cancel()
	}
	fmt.Printf("Test completed\n")

	t.Cleanup(func() {
		_ = l1Anvil.Process.Kill()
		_ = l2Anvil.Process.Kill()
		cancel()
	})

	time.Sleep(5 * time.Second)
	assert.True(t, taskVerified)
}

const (
	executorConfigYaml = `
---
grpcPort: 9090
operator:
  address: "0x15d34aaf54267db7d7c367839aaf71a00a2c6a65"
  operatorPrivateKey: "0x47e179ec197488593b187f80a00eb0da91f1b9d0b13f8733639f19c30a34926a"
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
    avsAddress: "0x70997970c51812dc3a010c7d01b50e0d17dc79c8"
    workerCount: 1
    signingCurve: "bn254"
    avsRegistrarAddress: "0xf4c5c29b14f0237131f7510a51684c8191f98e06"
`

	aggregatorConfigYaml = `
---
l1ChainId: 31337
chains:
  - name: ethereum
    network: mainnet
    chainId: 31337
    rpcUrl: http://localhost:8545
    pollIntervalSeconds: 10
  - name: base
    network: mainnet
    chainId: 31338
    rpcUrl: http://localhost:9545
    pollIntervalSeconds: 2
operator:
  address: "0x90f79bf6eb2c4f870365e785982e1f101e93b906"
  operatorPrivateKey: "0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6"
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
        }

avss:
  - address: "0x70997970c51812dc3a010c7d01b50e0d17dc79c8"
    responseTimeout: 3000
    chainIds: [31337]
    signingCurve: "bn254"
    avsRegistrarAddress: "0xf4c5c29b14f0237131f7510a51684c8191f98e06"
`
)
