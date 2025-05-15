package executor

import (
	"context"
	"fmt"
	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/localPeeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/simulations/simulatedAggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"math/big"
	"sync/atomic"
	"testing"
	"time"
)

func Test_Executor(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))

	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// executor setup
	execConfig, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(executorConfigYaml))
	if err != nil {
		t.Fatalf("failed to create executor config: %v", err)
	}

	storedKeys, err := keystore.ParseKeystoreJSON(execConfig.Operator.SigningKeys.BLS.Keystore)
	if err != nil {
		t.Fatalf("failed to parse keystore JSON: %v", err)
	}

	privateSigningKey, err := storedKeys.GetBN254PrivateKey(execConfig.Operator.SigningKeys.BLS.Password)
	if err != nil {
		t.Fatalf("failed to get private key: %v", err)
	}

	execSigner := inMemorySigner.NewInMemorySigner(privateSigningKey)

	// aggregator setup
	simAggConfig, err := aggregatorConfig.NewAggregatorConfigFromYamlBytes([]byte(aggregatorConfigYaml))
	if err != nil {
		t.Fatalf("Failed to create aggregator config: %v", err)
	}

	aggStoredKeys, err := keystore.ParseKeystoreJSON(simAggConfig.Operator.SigningKeys.BLS.Keystore)
	if err != nil {
		t.Fatalf("failed to parse keystore JSON: %v", err)
	}

	aggPrivateSigningKey, err := aggStoredKeys.GetBN254PrivateKey(simAggConfig.Operator.SigningKeys.BLS.Password)
	if err != nil {
		t.Fatalf("failed to get private key: %v", err)
	}

	baseRpcServer, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: execConfig.GrpcPort,
	}, l)
	if err != nil {
		l.Sugar().Fatal("Failed to setup RPC server", zap.Error(err))
	}

	pubKey := aggPrivateSigningKey.Public()
	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: []*peering.OperatorPeerInfo{
			{
				OperatorAddress: simAggConfig.Operator.Address,
				PublicKey:       pubKey,
				OperatorSetIds:  []uint32{0},
				NetworkAddress:  "localhost",
			},
		},
	}, l)

	exec := NewExecutor(execConfig, baseRpcServer, l, execSigner, pdf)

	if err := exec.Initialize(); err != nil {
		t.Fatalf("Failed to initialize executor: %v", err)
	}

	if err := exec.BootPerformers(ctx); err != nil {
		t.Fatalf("Failed to boot performers: %v", err)
	}

	// ------------------------------------------------------------------------
	// aggregator sim setup
	// ------------------------------------------------------------------------
	simAggPort := 5678
	aggBaseRpcServer, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: simAggPort,
	}, l)
	if err != nil {
		l.Sugar().Fatal("Failed to setup RPC server", zap.Error(err))
	}

	aggSigner := inMemorySigner.NewInMemorySigner(aggPrivateSigningKey)

	success := atomic.Bool{}
	success.Store(false)

	simAggregator, err := simulatedAggregator.NewSimulatedAggregator(simAggConfig, l, aggBaseRpcServer, func(result *aggregatorV1.TaskResult) {
		errors := false
		defer func() {
			success.Store(!errors)
			cancel()
		}()

		sig, err := bn254.NewSignatureFromBytes(result.Signature)
		if err != nil {
			errors = true
			t.Errorf("Failed to create signature from bytes: %v", err)
			return
		}

		digest := util.GetKeccak256Digest(result.Output)
		verified, err := sig.Verify(privateSigningKey.Public(), digest[:])
		if err != nil {
			errors = true
			t.Errorf("Failed to verify signature: %v", err)
			return
		}

		if !verified {
			errors = true
			t.Errorf("Signature verification failed")
			return
		}
		t.Logf("Successfully verified signature for task %s", result.TaskId)
	})
	if err != nil {
		t.Fatalf("Failed to create simulated aggregator: %v", err)
	}

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

	go func() {
		if err := simAggregator.Run(ctx); err != nil {
			t.Errorf("Failed to run simulated aggregator: %v", err)
			return
		}
	}()

	// give containers time to start.
	time.Sleep(5 * time.Second)

	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))
	payloadSig, err := aggSigner.SignMessage(payloadJsonBytes)

	if err != nil {
		t.Fatalf("Failed to sign task payload: %v", err)
	}

	ack, err := execClient.SubmitTask(ctx, &executorV1.TaskSubmission{
		TaskId:            "0x1234taskId",
		AggregatorAddress: simAggConfig.Operator.Address,
		AvsAddress:        simAggConfig.Avss[0].Address,
		Payload:           payloadJsonBytes,
		Signature:         payloadSig,
		AggregatorUrl:     fmt.Sprintf("localhost:%d", simAggPort),
	})
	if err != nil {
		cancel()
		time.Sleep(5 * time.Second)
		t.Fatalf("Failed to submit task: %v", err)
	}
	if ack == nil {
		cancel()
		time.Sleep(5 * time.Second)
		t.Fatalf("Ack is nil")
	}
	if ack.Success != true {
		cancel()
		time.Sleep(5 * time.Second)
		t.Fatalf("Ack success is false")
	}

	<-ctx.Done()
	t.Logf("Received shutdown signal, shutting down...")
	assert.True(t, success.Load(), "task completed successfully")
}

const (
	executorConfigYaml = `
---
grpcPort: 9090
operator:
  address: "0xoperator..."
  operatorPrivateKey: "..."
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
avsPerformers:
- image:
    repository: "hello-performer"
    tag: "latest"
  processType: "server"
  avsAddress: "0xavs1..."
  workerCount: 1
  signingCurve: "bn254"
`

	aggregatorConfigYaml = `
---
chains:
  - name: ethereum
    network: mainnet
    chainId: 31337
    rpcUrl: https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID
operator:
  address: "0x1234aggregator"
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
  - address: "0xavs1..."
    privateKey: "some private key"
    privateSigningKey: "some private signing key"
    privateSigningKeyType: "ecdsa"
    responseTimeout: 3000
    chainIds: [31337]
`
)
