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

const (
	aggregatorOperatorAddr = "0x1234aggregator"
	aggregatorPublicKey    = "0b9adbefd52a9ef6d081d06dbdb8f5791321cd6676c19e1a594d845d4801e4551c61d9d2fcace2053b9928773cbaefb3a7b071be410ca21086941a4904d573261e2672e196e9e8528296807af313b0c4c27ac42a51db525842e4abbc66f4020426c07af913dd7703ebaef038004f892dc17d44f9f9c76c9c17dfb8de794e9213"
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

	aggPubKey, err := bn254.NewPublicKeyFromBytes([]byte(aggregatorPublicKey))
	if err != nil {
		t.Fatalf("Failed to create public key from bytes: %v", err)
	}

	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: []*peering.OperatorPeerInfo{
			{
				OperatorAddress: aggregatorOperatorAddr,
				PublicKey:       aggPubKey,
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
          "publicKey": "1e14b1b9a847b5bedea4e44e18541fe153a7d791d68f651bc86ac4be7dfe36000116d8d71bd4ae56331f58bec5967ae1adad91823b5d4b90746a5d85c3c8faaa13a4880b8c163984d2fec316803a146b1b97f63d95e6ac9536c061924b6367131799e6499a8ea979bc6fac9e01d0002a5547894e9212f09d6b2ed94843593145",
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
          "publicKey": "0b9adbefd52a9ef6d081d06dbdb8f5791321cd6676c19e1a594d845d4801e4551c61d9d2fcace2053b9928773cbaefb3a7b071be410ca21086941a4904d573261e2672e196e9e8528296807af313b0c4c27ac42a51db525842e4abbc66f4020426c07af913dd7703ebaef038004f892dc17d44f9f9c76c9c17dfb8de794e9213",
          "crypto": {
            "cipher": "aes-128-ctr",
            "ciphertext": "d364b7efca8f6df2a5d0a973976d2ae27893e8ba04bed1c8008c95557591ff73",
            "cipherparams": {
              "iv": "0b27b6a532f5519b011b2075372507cb"
            },
            "kdf": "scrypt",
            "kdfparams": {
              "dklen": 32,
              "n": 262144,
              "p": 1,
              "r": 8,
              "salt": "f9eb85d1059ac71f7971aea310029ed9cd8c0da6b03e2cc854ddffc55e21d2cd"
            },
            "mac": "295f897f1e50a71884d96debbec2fed3c23403776941017fa689567b9ac946d8"
          },
          "uuid": "4b9804d5-b594-4690-b4a3-dc1c76a0f110",
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
          "publicKey": "0b9adbefd52a9ef6d081d06dbdb8f5791321cd6676c19e1a594d845d4801e4551c61d9d2fcace2053b9928773cbaefb3a7b071be410ca21086941a4904d573261e2672e196e9e8528296807af313b0c4c27ac42a51db525842e4abbc66f4020426c07af913dd7703ebaef038004f892dc17d44f9f9c76c9c17dfb8de794e9213",
          "crypto": {
            "cipher": "aes-128-ctr",
            "ciphertext": "d364b7efca8f6df2a5d0a973976d2ae27893e8ba04bed1c8008c95557591ff73",
            "cipherparams": {
              "iv": "0b27b6a532f5519b011b2075372507cb"
            },
            "kdf": "scrypt",
            "kdfparams": {
              "dklen": 32,
              "n": 262144,
              "p": 1,
              "r": 8,
              "salt": "f9eb85d1059ac71f7971aea310029ed9cd8c0da6b03e2cc854ddffc55e21d2cd"
            },
            "mac": "295f897f1e50a71884d96debbec2fed3c23403776941017fa689567b9ac946d8"
          },
          "uuid": "4b9804d5-b594-4690-b4a3-dc1c76a0f110",
          "version": 4,
          "curveType": "bn254"
        }`
)
