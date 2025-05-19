package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
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
	"testing"
	"time"
)

const (
	RPCUrl = "http://127.0.0.1:8545"
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
	_ = chainConfig

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
	// L1Chain & anvil setup
	// ------------------------------------------------------------------------
	ethereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   RPCUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)

	ethClient, err := ethereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get Ethereum contract caller: %v", err)
	}

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

	// ------------------------------------------------------------------------
	// register peering data
	// ------------------------------------------------------------------------
	aggCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.OperatorAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddress,
	}, ethClient, l)
	if err != nil {
		t.Fatalf("Failed to create contract caller: %v", err)
	}
	t.Logf("Registering avs peering data")
	_, err = operator.RegisterOperatorToOperatorSets(
		ctx,
		aggCc,
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

	execCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.ExecOperatorAccountPk,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddress,
	}, ethClient, l)
	if err != nil {
		t.Fatalf("Failed to create contract caller: %v", err)
	}
	t.Logf("Registering operator peering data")
	_, err = operator.RegisterOperatorToOperatorSets(
		ctx,
		execCc,
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
	execPdf := peeringDataFetcher.NewPeeringDataFetcher(execCc, l)
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

	tlp := transactionLogParser.NewTransactionLogParser(imContractStore, l)
	aggPdf := peeringDataFetcher.NewPeeringDataFetcher(aggCc, l)

	simulationDelay := time.Second
	if aggConfig.SimulationConfig != nil {
		simulationDelay = time.Duration(aggConfig.SimulationConfig.WriteDelaySeconds) * time.Second
	}

	agg, err := NewAggregatorWithRpcServer(
		aggConfig.ServerConfig.Port,
		&AggregatorConfig{
			AVSs:              aggConfig.Avss,
			Chains:            aggConfig.Chains,
			Address:           aggConfig.Operator.Address,
			PrivateKey:        aggConfig.Operator.OperatorPrivateKey,
			AggregatorUrl:     aggConfig.ServerConfig.AggregatorUrl,
			WriteDelaySeconds: simulationDelay,
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
	appCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.AppAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddress,
	}, ethClient, l)
	if err != nil {
		t.Fatalf("Failed to create contract caller: %v", err)
	}
	t.Logf("Pushing message to mailbox...")
	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))
	task, err := appCc.PublishMessageToInbox(ctx, chainConfig.AVSAccountAddress, 1, payloadJsonBytes)
	if err != nil {
		t.Fatalf("Failed to publish message to inbox: %v", err)
	}
	t.Logf("Task published: %+v", task)

	// ------------------------------------------------------------------------
	// Listen for TaskVerified event to know that the test is done
	// ------------------------------------------------------------------------
	wsEthClient, err := ethereumClient.GetWebsocketConnection("ws://localhost:8545")
	if err != nil {
		t.Fatalf("Failed to get websocket connection: %v", err)
	}

	taskVerified := false

	eventsChan := make(chan types.Log)
	sub, err := wsEthClient.SubscribeFilterLogs(ctx, goEthereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(chainConfig.MailboxContractAddress)},
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
	case <-time.After(60 * time.Second):
		t.Logf("Timeout after 60 seconds")
		cancel()
	}
	fmt.Printf("Test completed\n")

	_ = anvil.Process.Kill()
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
          "publicKey": "2d6b7590f1fea33186b11a795b5a6c5c77b3ebdd5563ad11404098c8e4d92a8209e5d2e5fd537eb2c253a9d13735935079bcb8902f09bbd7a117d07f3142d5f9039ca163db601221d77db55b0fe3876aab1ff8bdf90a205f60cb244633789f0020d166cd401deed5dcac545ae8d58ba6e024b7aa626c51ef74b23ef5fa170ba4",
          "crypto": {
            "cipher": "aes-128-ctr",
            "ciphertext": "de8e36c294f88c582d0f84ebadef0470b38dfd6209597e3f71013d780d033105",
            "cipherparams": {
              "iv": "780729b623bea9237293d11d949c6790"
            },
            "kdf": "scrypt",
            "kdfparams": {
              "dklen": 32,
              "n": 262144,
              "p": 1,
              "r": 8,
              "salt": "fc621449564675b56cfa22785b8fa362e63666a4f834e86f33683e5ccef700c2"
            },
            "mac": "a9e8175072147ef23ee6742aaeb96b4da0003a84925f1e74b78bedf4c6f8fd8a"
          },
          "uuid": "7c5feddd-b78f-404a-8548-7f84eac102e1",
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
operator:
  address: "0x90f79bf6eb2c4f870365e785982e1f101e93b906"
  operatorPrivateKey: "0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6"
  signingKeys:
    bls:
      password: ""
      keystore: | 
        {
          "publicKey": "1f9f528a1ab51aa8a8300d5abb3956d641d561942661020d93ec15217f72499513246c8fd468a8b1b982a252e7cf970e6bddf52c26c12341b5c6edc9787f94c312c44a2acc0f4a997ee5a06c8adb1451edd5c192bf05c53d142e895a163015c806ea90c5dfc90b58f428c633c0a571ae20f5febb4cb91e9f6ce09d248dcaabf8",
          "crypto": {
            "cipher": "aes-128-ctr",
            "ciphertext": "f011291fe6c96bcc74e4e5bd58d6dd169c27bf97ce3d69930cbc7836d9d968eb",
            "cipherparams": {
              "iv": "0b7426c25a24db1c90aec9c69c19a402"
            },
            "kdf": "scrypt",
            "kdfparams": {
              "dklen": 32,
              "n": 262144,
              "p": 1,
              "r": 8,
              "salt": "0d969931719e36f4946c8660bbb366737f07880ff1d2d9639e066acfec72eb53"
            },
            "mac": "095c9dfb4967d2bfe7d8a02cb9928c4e13f29d23254ab0b687b88022f2346551"
          },
          "uuid": "2f6cfbda-d9be-4a03-bf16-7750d1b67f22",
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
