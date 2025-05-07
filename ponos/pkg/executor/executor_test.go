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

const (
	aggregatorOperatorAddr = "0x1234aggregator"
	aggregatorPublicKey    = "10d2dcc53580b7c54f584ea9d0ce935c558243a898e9b221c3f7d172545455a726f701c37cda46ac284006cccba284c84bac254a1aa2a7bd10d46fa79cddb01d"
)

func signTaskPayload(payload []byte) ([]byte, error) {
	ks, err := keystore.ParseKeystoreJSON(aggregatorKeystore)
	if err != nil {
		return nil, err
	}
	keyScheme, err := keystore.GetSigningSchemeForCurveType(ks.CurveType)
	if err != nil {
		return nil, err
	}

	pk, err := ks.GetPrivateKey("", keyScheme)
	if err != nil {
		return nil, err
	}

	sig := inMemorySigner.NewInMemorySigner(pk)
	return sig.SignMessage(payload)
}

func bigIntToHex(i *big.Int) []byte {
	if i == nil {
		return nil
	}
	hexStr := i.Text(16)
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	return []byte("0x" + hexStr)
}

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

	keyScheme, err := keystore.GetSigningSchemeForCurveType(storedKeys.CurveType)
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

	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: []*peering.OperatorPeerInfo{
			{
				OperatorAddress: aggregatorOperatorAddr,
				PublicKey:       aggregatorPublicKey,
				OperatorSetIds:  []uint32{0},
				NetworkAddress:  "localhost",
			},
		},
	}, l)

	exec := NewExecutor(execConfig, baseRpcServer, l, sig, pdf)

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

	payloadJsonBytes := bigIntToHex(new(big.Int).SetUint64(4))

	payloadSig, err := signTaskPayload(payloadJsonBytes)
	if err != nil {
		t.Fatalf("Failed to sign task payload: %v", err)
	}

	ack, err := execClient.SubmitTask(ctx, &executorV1.TaskSubmission{
		TaskId:            "0x1234taskId",
		AggregatorAddress: aggregatorOperatorAddr,
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
          "publicKey": "1b62c6ebbb2e62704bccd850f3cb6e42c07263866d76e361f1bd436ea79eec20150fc9e5f63ce1a11ead51f062494d6f6ae6f1cd8e3d212525e5d25dc082c1b6",
          "crypto": {
            "cipher": "aes-128-ctr",
            "ciphertext": "751ea48ca668b7ae5b812690a8ded38a0e2675c0536fbbfeb4918a2c0c0ab732",
            "cipherparams": {
              "iv": "43c937bd6659eabfe166741f4d74dad7"
            },
            "kdf": "scrypt",
            "kdfparams": {
              "dklen": 32,
              "n": 262144,
              "p": 1,
              "r": 8,
              "salt": "fb4c8d27ddb45b7a7412ad3afa6b62bfdaea2c6d8dc1a1869f83adb47e72198e"
            },
            "mac": "8b8d33cd738dd37ef3c577a113e5f65d2563dc47d5142891610ec3edbba7bb5f"
          },
          "uuid": "741a2583-e42b-43a3-8f11-85fd4e2b2669",
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
  signingKeys:
    bls:
      password: ""
      keystore: | 
        {
          "publicKey": "10d2dcc53580b7c54f584ea9d0ce935c558243a898e9b221c3f7d172545455a726f701c37cda46ac284006cccba284c84bac254a1aa2a7bd10d46fa79cddb01d",
          "crypto": {
            "cipher": "aes-128-ctr",
            "ciphertext": "8a277efe25159b05fc5193aeb0c5346d12d565c0db8d28e8bb18904f63945c9c",
            "cipherparams": {
              "iv": "27b36cf4cbdeba506cd17ec00757c98a"
            },
            "kdf": "scrypt",
            "kdfparams": {
              "dklen": 32,
              "n": 262144,
              "p": 1,
              "r": 8,
              "salt": "a46e025642a031f83b0b78badbd6120a2e5ae3edf4dbe772e2877f723f88c9b9"
            },
            "mac": "647a73d802ae702f302c0012300cc2c7cc61142ec5f4f7edf0930f90144c6df2"
          },
          "uuid": "829bd1cf-2b64-4996-afa9-4664b2aafbf8",
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
	aggregatorKeystore = `{
          "publicKey": "10d2dcc53580b7c54f584ea9d0ce935c558243a898e9b221c3f7d172545455a726f701c37cda46ac284006cccba284c84bac254a1aa2a7bd10d46fa79cddb01d",
          "crypto": {
            "cipher": "aes-128-ctr",
            "ciphertext": "8a277efe25159b05fc5193aeb0c5346d12d565c0db8d28e8bb18904f63945c9c",
            "cipherparams": {
              "iv": "27b36cf4cbdeba506cd17ec00757c98a"
            },
            "kdf": "scrypt",
            "kdfparams": {
              "dklen": 32,
              "n": 262144,
              "p": 1,
              "r": 8,
              "salt": "a46e025642a031f83b0b78badbd6120a2e5ae3edf4dbe772e2877f723f88c9b9"
            },
            "mac": "647a73d802ae702f302c0012300cc2c7cc61142ec5f4f7edf0930f90144c6df2"
          },
          "uuid": "829bd1cf-2b64-4996-afa9-4664b2aafbf8",
          "version": 4,
          "curveType": "bn254"
        }`
)
