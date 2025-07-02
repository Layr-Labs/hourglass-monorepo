package executor

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/keystore"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/localPeeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/stretchr/testify/assert"
)

func parseKeysFromConfig(
	operatorConfig *config.OperatorConfig,
	curveType config.CurveType,
) (*bn254.PrivateKey, *cryptoLibsEcdsa.PrivateKey, interface{}, error) {
	var genericExecutorSigningKey interface{}
	var bn254PrivateSigningKey *bn254.PrivateKey
	var ecdsaPrivateSigningKey *cryptoLibsEcdsa.PrivateKey
	var err error

	if curveType == config.CurveTypeBN254 {
		storedKeys, err := keystore.ParseKeystoreJSON(operatorConfig.SigningKeys.BLS.Keystore)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse keystore JSON: %w", err)
		}

		bn254PrivateSigningKey, err = storedKeys.GetBN254PrivateKey(operatorConfig.SigningKeys.BLS.Password)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to get private key: %w", err)
		}
		genericExecutorSigningKey = bn254PrivateSigningKey
	} else if curveType == config.CurveTypeECDSA {
		ecdsaPrivateSigningKey, err = cryptoLibsEcdsa.NewPrivateKeyFromHexString(operatorConfig.SigningKeys.ECDSA)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to get ECDSA private key: %w", err)
		}
		genericExecutorSigningKey = ecdsaPrivateSigningKey
	} else {
		return nil, nil, nil, fmt.Errorf("unsupported curve type: %s", curveType)
	}
	return bn254PrivateSigningKey, ecdsaPrivateSigningKey, genericExecutorSigningKey, nil
}

func testWithKeyType(
	t *testing.T,
	curveType config.CurveType,
	executorConfigYaml string,
	aggregatorConfigYaml string,
) {
	t.Logf("Running test with curve type: %s", curveType)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))

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

	// executor setup
	execConfig, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(executorConfigYaml))
	if err != nil {
		t.Fatalf("failed to create executor config: %v", err)
	}
	execConfig.Operator.SigningKeys.ECDSA = chainConfig.ExecOperatorAccountPk

	execBn254PrivateSigningKey, execEcdsaPrivateSigningKey, execGenericExecutorSigningKey, err := parseKeysFromConfig(execConfig.Operator, curveType)
	if err != nil {
		t.Fatalf("Failed to parse keys from config: %v", err)
	}
	execSigner := inMemorySigner.NewInMemorySigner(execGenericExecutorSigningKey, curveType)

	// aggregator setup
	simAggConfig, err := aggregatorConfig.NewAggregatorConfigFromYamlBytes([]byte(aggregatorConfigYaml))
	if err != nil {
		t.Fatalf("Failed to create aggregator config: %v", err)
	}
	simAggConfig.Operator.SigningKeys.ECDSA = chainConfig.OperatorAccountPrivateKey

	aggBn254PrivateSigningKey, aggEcdsaPrivateSigningKey, aggGenericExecutorSigningKey, err := parseKeysFromConfig(simAggConfig.Operator, curveType)
	if err != nil {
		t.Fatalf("Failed to parse keys from config: %v", err)
	}
	aggSigner := inMemorySigner.NewInMemorySigner(aggGenericExecutorSigningKey, curveType)

	peer := &peering.OperatorSet{
		OperatorSetID:    0,
		WrappedPublicKey: peering.WrappedPublicKey{},
		NetworkAddress:   fmt.Sprintf("localhost:%d", execConfig.GrpcPort),
		CurveType:        curveType,
	}
	if curveType == config.CurveTypeBN254 {
		peer.WrappedPublicKey.PublicKey = aggBn254PrivateSigningKey.Public()
	} else if curveType == config.CurveTypeECDSA {
		derivedAddress, err := aggEcdsaPrivateSigningKey.DeriveAddress()
		if err != nil {
			t.Fatalf("Failed to derive ECDSA address: %v", err)
		}
		peer.WrappedPublicKey.ECDSAAddress = derivedAddress
	}
	fmt.Printf("Peer: %+v\n", peer)

	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: []*peering.OperatorPeerInfo{
			{
				OperatorAddress: simAggConfig.Operator.Address,
				OperatorSets:    []*peering.OperatorSet{peer},
			},
		},
	}, l)

	exec, err := NewExecutorWithRpcServer(execConfig.GrpcPort, execConfig, l, execSigner, pdf)
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

	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))
	payloadHash := util.GetKeccak256Digest(payloadJsonBytes)

	payloadSig, err := aggSigner.SignMessage(payloadHash[:])
	if err != nil {
		t.Fatalf("Failed to sign task payload: %v", err)
	}

	// send the task to the executor
	taskResult, err := execClient.SubmitTask(ctx, &executorV1.TaskSubmission{
		TaskId:            "0x1234taskId",
		AggregatorAddress: simAggConfig.Operator.Address,
		AvsAddress:        simAggConfig.Avss[0].Address,
		Payload:           payloadJsonBytes,
		Signature:         payloadSig,
	})
	if err != nil {
		cancel()
		time.Sleep(5 * time.Second)
		t.Fatalf("Failed to submit task: %v", err)
	}
	assert.NotNil(t, taskResult)

	digest := util.GetKeccak256Digest(taskResult.Output)

	var verified bool
	if curveType == config.CurveTypeBN254 {
		sig, err := bn254.NewSignatureFromBytes(taskResult.Signature)
		assert.Nil(t, err)
		verified, err = sig.VerifySolidityCompatible(execBn254PrivateSigningKey.Public(), digest)
	} else if curveType == config.CurveTypeECDSA {
		sig, err := cryptoLibsEcdsa.NewSignatureFromBytes(taskResult.Signature)
		assert.Nil(t, err)
		derivedAddress, err := execEcdsaPrivateSigningKey.DeriveAddress()
		if err != nil {
			t.Fatalf("Failed to derive ECDSA address: %v", err)
		}
		verified, err = sig.VerifyWithAddress(digest[:], derivedAddress)
	} else {
		t.Fatalf("Unsupported curve type: %s", curveType)
	}
	assert.Nil(t, err)
	assert.True(t, verified)
	cancel()
	t.Logf("Successfully verified signature for task %s", taskResult.TaskId)

	<-ctx.Done()
	t.Logf("Received shutdown signal, shutting down...")
	time.Sleep(3 * time.Second)
}

func Test_Executor(t *testing.T) {

	t.Run("BN254", func(t *testing.T) {
		testWithKeyType(t, config.CurveTypeBN254, executorConfigYaml, aggregatorConfigYaml)
	})
	t.Run("ECDSA", func(t *testing.T) {
		testWithKeyType(t, config.CurveTypeECDSA, executorConfigYaml, aggregatorConfigYaml)
	})

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
  avsAddress: "0xavs1..."
  workerCount: 1
  signingCurve: "bn254"
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
  address: "0x1234aggregator"
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
  - address: "0xavs1..."
    privateKey: "some private key"
    privateSigningKey: "some private signing key"
    privateSigningKeyType: "ecdsa"
    responseTimeout: 3000
    chainIds: [31337]
    avsRegistrarAddress: "0xf4c5c29b14f0237131f7510a51684c8191f98e06"
`
)
