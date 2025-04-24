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
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/simulators/simulatedAggregator"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

	execConfig, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(executorConfigYaml))
	if err != nil {
		t.Fatalf("failed to create executor config: %v", err)
	}

	storedKeys, err := keystore.ParseKeystoreJSON(execConfig.Operator.SigningKeys.BLS.Keystore)
	if err != nil {
		t.Fatalf("failed to parse keystore JSON: %v", err)
	}

	keyScheme, err := keystore.GetSigningScheme(storedKeys.CurveType)
	if err != nil {
		t.Fatalf("failed to get signing scheme: %v", err)
	}

	privateSigningKey, err := storedKeys.GetPrivateKey(execConfig.Operator.SigningKeys.BLS.Password, keyScheme)
	if err != nil {
		t.Fatalf("failed to get private key: %v", err)
	}

	sig := inMemorySigner.NewInMemorySigner(privateSigningKey)

	baseRpcServer, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: execConfig.GrpcPort,
	}, l)
	if err != nil {
		l.Sugar().Fatal("Failed to setup RPC server", zap.Error(err))
	}

	exec := NewExecutor(execConfig, baseRpcServer, l, sig)

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

	simAggConfig, err := aggregatorConfig.NewAggregatorConfigFromYamlBytes([]byte(aggregatorConfigYaml))
	if err != nil {
		t.Fatalf("Failed to create aggregator config: %v", err)
	}

	success := atomic.Bool{}
	success.Store(false)

	simAggregator, err := simulatedAggregator.NewSimulatedAggregator(simAggConfig, l, aggBaseRpcServer, func(result *aggregatorV1.TaskResult) {
		errors := false
		defer func() {
			success.Store(!errors)
			cancel()
		}()

		sig, err := keyScheme.NewSignatureFromBytes(result.Signature)
		if err != nil {
			errors = true
			t.Errorf("Failed to create signature from bytes: %v", err)
			return
		}

		verified, err := sig.Verify(privateSigningKey.Public(), result.Output)
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

	payload := []byte(`{"message": "hello world"}`)

	// TODO(seanmcgary): sign this with an aggregators key

	ack, err := execClient.SubmitTask(ctx, &executorV1.TaskSubmission{
		TaskId:            "0x1234taskId",
		AggregatorAddress: "0x1234aggregator",
		AvsAddress:        simAggConfig.Avss[0].Address,
		Payload:           payload,
		Signature:         []byte("totally a signature"),
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
	assert.True(t, success.Load(), "Task completed successfully")
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
          "publicKey": "27ca30a9935d5c8435d4a2854cc6c376c253d6a4cb6e0026ea7a98b12789fd0b297a85d16550fe94ee0433eeac9a6a854d64e61ed82d7484b4287cb289ea962212e3a593a27a8aa7e196adc51336857c6fb30791fb70ac5bd8a522d4d486d0e3043f3e74c00a9f10bed939b07a06ff1b9bbb47794e613aa597d3e364c540bdf7",
          "crypto": {
            "cipher": "aes-128-ctr",
            "ciphertext": "a0e75151edfb59c0a224a4ef74c6b572d98607a2fa48f85133a693b399d5c316",
            "cipherparams": {
              "iv": "6f02e76da70983d69a7cb9f072f3a384"
            },
            "kdf": "scrypt",
            "kdfparams": {
              "dklen": 32,
              "n": 262144,
              "p": 1,
              "r": 8,
              "salt": "c06be39f07a19428c69bec6c5d136a38ed5bb73ec0f68c81af1634dfda099ab8"
            },
            "mac": "28b382b5cb3fb663d7ac0e17dc7c7fdc9fe8065901e536fea12ce30c2d7a70a4"
          },
          "uuid": "2d34e7e1-0c94-4741-bfdc-a8aae120cf2f",
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
`

	aggregatorConfigYaml = `
---
chains:
  - name: ethereum
    network: mainnet
    chainId: 1
    rpcUrl: https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID
avss:
  - address: "0xavs1..."
    privateKey: "some private key"
    privateSigningKey: "some private signing key"
    privateSigningKeyType: "ecdsa"
    responseTimeout: 3000
    chainIds: [1]
`
)
