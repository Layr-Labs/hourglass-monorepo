package peeringDataFetcher

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	cryptoUtils "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/crypto"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	RPCUrl = "http://127.0.0.1:8545"
)

func Test_PeeringDataFetcher(t *testing.T) {
	t.Run("setup operator peering data, then read it back and verify correctness", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

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

		ethereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
			BaseUrl:   RPCUrl,
			BlockType: ethereum.BlockType_Latest,
		}, l)

		// aggregator operator
		aggOperatorPrivateKey, err := cryptoUtils.StringToECDSAPrivateKey(chainConfig.OperatorAccountPrivateKey)
		if err != nil {
			l.Sugar().Fatalf("failed to convert private key: %v", err)
		}
		aggOperatorAddress := cryptoUtils.DeriveAddress(aggOperatorPrivateKey)
		assert.True(t, strings.EqualFold(aggOperatorAddress.String(), chainConfig.OperatorAccountAddress))

		// executor operator
		execOperatorPrivateKey, err := cryptoUtils.StringToECDSAPrivateKey(chainConfig.ExecOperatorAccountPk)
		if err != nil {
			l.Sugar().Fatalf("failed to convert private key: %v", err)
		}
		execOperatorAddress := cryptoUtils.DeriveAddress(execOperatorPrivateKey)
		assert.True(t, strings.EqualFold(execOperatorAddress.String(), chainConfig.ExecOperatorAccountAddress))

		ethClient, err := ethereumClient.GetEthereumContractCaller()
		if err != nil {
			l.Sugar().Fatalf("failed to get Ethereum contract caller: %v", err)
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

		testCases := []struct {
			privateKey   string
			address      string
			operatorSets []uint32
			operatorType string
		}{
			{
				privateKey:   chainConfig.OperatorAccountPrivateKey,
				address:      chainConfig.OperatorAccountAddress,
				operatorSets: []uint32{0},
				operatorType: "aggregator",
			}, {
				privateKey:   chainConfig.ExecOperatorAccountPk,
				address:      chainConfig.ExecOperatorAccountAddress,
				operatorSets: []uint32{1},
				operatorType: "executor",
			},
		}

		hasErrors := false
		for _, tc := range testCases {
			cc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
				PrivateKey:          tc.privateKey,
				AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
				TaskMailboxAddress:  chainConfig.MailboxContractAddress,
			}, ethClient, l)
			if err != nil {
				l.Sugar().Fatalf("failed to create contract caller: %v", err)
			}

			testOperatorAddress := common.HexToAddress(tc.address)

			privateKey, publicKey, err := bn254.GenerateKeyPair()
			if err != nil {
				l.Sugar().Fatalf("failed to generate key pair: %v", err)
			}

			g1Point, err := cc.GetOperatorRegistrationMessageHash(ctx, testOperatorAddress)
			if err != nil {
				l.Sugar().Fatalf("failed to get operator registration message hash: %v", err)
			}

			// Create G1 point from contract coordinates
			hashPoint := bn254.NewG1Point(g1Point.X, g1Point.Y)

			// Sign the hash point
			signature, err := privateKey.SignG1Point(hashPoint.G1Affine)
			if err != nil {
				l.Sugar().Fatalf("failed to sign hash point: %v", err)
			}

			result, err := cc.CreateOperatorAndRegisterWithAvs(
				ctx,
				common.HexToAddress(chainConfig.AVSAccountAddress),
				testOperatorAddress,
				tc.operatorSets,
				publicKey,
				signature,
				"localhost:8545",
				7200,
				"http://localhost:8545",
			)
			assert.Nil(t, err)
			fmt.Printf("Result: %+v\n", result)

			// create a peeringDataFetcher and get the data
			pdf := NewPeeringDataFetcher(cc, l)

			var peers []*peering.OperatorPeerInfo
			if tc.operatorType == "executor" {
				peers, err = pdf.ListExecutorOperators(ctx, chainConfig.AVSAccountAddress)
				if err != nil {
					t.Fatalf("Failed to list executor operators: %v", err)
				}
				assert.Equal(t, 1, len(peers))
				for _, peer := range peers {
					t.Logf("Executor Peer: %+v\n", peer)
				}

			} else if tc.operatorType == "aggregator" {
				peers, err = pdf.ListAggregatorOperators(ctx, chainConfig.AVSAccountAddress)
				if err != nil {
					t.Fatalf("Failed to list aggregator operators: %v", err)
				}
				assert.Equal(t, 1, len(peers))

				for _, peer := range peers {
					t.Logf("Aggregator Peer: %+v\n", peer)
				}
			}

			testMessage := []byte("test message")

			testSig, err := privateKey.Sign(testMessage)
			if err != nil {
				t.Fatalf("Failed to sign message: %v", err)
			}

			valid, err := testSig.Verify(peers[0].PublicKey, testMessage)
			if err != nil {
				t.Fatalf("Failed to verify signature: %v", err)
			}
			assert.True(t, valid)
		}

		cancel()
		select {
		case <-time.After(90 * time.Second):
			cancel()
			t.Fatalf("Test timed out after 10 seconds")
		case <-ctx.Done():
			t.Logf("Test completed")
		}

		_ = anvil.Process.Kill()
		assert.False(t, hasErrors)
	})

}
