package executor

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"math/big"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

	_, execEcdsaPrivateSigningKey, execGenericExecutorSigningKey, err := testUtils.ParseKeysFromConfig(execConfig.Operator, config.CurveTypeECDSA)
	if err != nil {
		t.Fatalf("Failed to parse keys from config: %v", err)
	}
	execSigner := inMemorySigner.NewInMemorySigner(execGenericExecutorSigningKey, config.CurveTypeECDSA)

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

	aggBn254PrivateSigningKey, _, _, err := testUtils.ParseKeysFromConfig(simAggConfig.Operator, config.CurveTypeBN254)
	if err != nil {
		t.Fatalf("Failed to parse keys from config: %v", err)
	}
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

	err = testUtils.SetupOperatorPeering(
		ctx,
		chainConfig,
		config.ChainId(l1ChainId.Uint64()),
		l1EthClient,
		// aggregator is BN254
		&operator.Operator{
			TransactionPrivateKey: chainConfig.OperatorAccountPrivateKey,
			SigningPrivateKey:     aggBn254PrivateSigningKey,
			Curve:                 config.CurveTypeBN254,
			OperatorSetIds:        []uint32{aggOpsetId},
		},
		// executor is ecdsa
		&operator.Operator{
			TransactionPrivateKey: chainConfig.ExecOperatorAccountPk,
			SigningPrivateKey:     execEcdsaPrivateSigningKey,
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
	blockNumber, err := strconv.ParseUint(strBlockNumber, 10, 64)
	if err != nil {
		t.Fatalf("Failed to parse L1 block strBlockNumber: %v", err)
	}
	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))
	encodedMessage := util.EncodeTaskSubmissionMessage(
		taskId,
		simAggConfig.Avss[0].Address,
		chainConfig.ExecOperatorAccountAddress,
		1,
		blockNumber,
		payloadJsonBytes,
	)

	// For BN254, we need to sign directly without hashing
	var payloadSig []byte
	if simAggConfig.Operator.SigningKeys.BLS != nil {
		// BN254 signing - sign the raw message
		bn254Key := aggBn254PrivateSigningKey
		sig, err := bn254Key.Sign(encodedMessage)
		if err != nil {
			t.Fatalf("Failed to sign task payload with BN254: %v", err)
		}
		payloadSig = sig.Bytes()
	} else {
		t.Fatalf("signer should be BLS")
	}

	// send the task to the executor
	taskResult, err := execClient.SubmitTask(ctx, &executorV1.TaskSubmission{
		TaskId:            taskId,
		AggregatorAddress: simAggConfig.Operator.Address,
		AvsAddress:        simAggConfig.Avss[0].Address,
		ExecutorAddress:   chainConfig.ExecOperatorAccountAddress,
		Payload:           payloadJsonBytes,
		Signature:         payloadSig,
		TaskBlockNumber:   blockNumber,
		OperatorSetId:     1,
	})
	if err != nil {
		cancel()
		time.Sleep(5 * time.Second)
		t.Fatalf("Failed to submit task: %v", err)
	}
	assert.NotNil(t, taskResult)

	if curveType == config.CurveTypeBN254 {
		// BN254 signature verification - check signature format
		sig, err := bn254.NewSignatureFromBytes(taskResult.ResultSignature)
		if err != nil {
			t.Fatalf("Failed to create BN254 signature from bytes: %v", err)
		}
		// Just verify the signature is valid BN254 format
		assert.NotNil(t, sig, "BN254 signature should be valid")
		assert.Len(t, taskResult.ResultSignature, 64, "BN254 signature should be 64 bytes")
		t.Logf("BN254 result signature present and valid format")
	} else if curveType == config.CurveTypeECDSA {
		// ECDSA signature verification - check signature format
		sig, err := cryptoLibsEcdsa.NewSignatureFromBytes(taskResult.ResultSignature)
		if err != nil {
			t.Fatalf("Failed to create ECDSA signature from bytes: %v", err)
		}
		// Just verify the signature is valid ECDSA format
		assert.NotNil(t, sig, "ECDSA signature should be valid")
		assert.Len(t, taskResult.ResultSignature, 65, "ECDSA signature should be 65 bytes")
		t.Logf("ECDSA result signature present and valid format")

		// Also verify the AuthSignature (used for identity verification)
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

func Test_Executor(t *testing.T) {
	// t.Run("BN254", func(t *testing.T) {
	// 	t.Skip("Executor is only setup as ECDSA for now")
	// 	testWithKeyType(t, config.CurveTypeBN254, executorConfigYaml, aggregatorConfigYaml)
	// })
	t.Run("ECDSA", func(t *testing.T) {
		testWithKeyType(t, config.CurveTypeECDSA, executorConfigYaml, aggregatorConfigYaml)
	})

}

const (
	executorConfigYaml = `
---
grpcPort: 9090
operator:
  address: "0x9b18a6d836e9b2b6541fa9c7247f46b4a4a2f2fc"
  operatorPrivateKey:
    privateKey: "0x40a4c2aa3c75c735a5e3deaeb77cf5b6ea73bf12771f634e07a82d501f420849"
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
  address: "0x9b18a6d836e9b2b6541fa9c7247f46b4a4a2f2fc"
  operatorPrivateKey:
    privateKey: "0x40a4c2aa3c75c735a5e3deaeb77cf5b6ea73bf12771f634e07a82d501f420849"
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
  - address: "0xce2ac75be2e0951f1f7b288c7a6a9bfb6c331dc4"
    privateKey: "some private key"
    privateSigningKey: "some private signing key"
    privateSigningKeyType: "ecdsa"
    responseTimeout: 3000
    chainIds: [31337]
    avsRegistrarAddress: "0xf4c5c29b14f0237131f7510a51684c8191f98e06"
`
)
