package peeringDataFetcher

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	cryptoUtils "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/crypto"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/keccak256"
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
	t.Run("BN254", func(t *testing.T) {
		t.Skip()
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
		aggOperatorAddress := cryptoUtils.DeriveAddress(aggOperatorPrivateKey.PublicKey)
		assert.True(t, strings.EqualFold(aggOperatorAddress.String(), chainConfig.OperatorAccountAddress))

		// executor operator
		execOperatorPrivateKey, err := cryptoUtils.StringToECDSAPrivateKey(chainConfig.ExecOperatorAccountPk)
		if err != nil {
			l.Sugar().Fatalf("failed to convert private key: %v", err)
		}
		execOperatorAddress := cryptoUtils.DeriveAddress(execOperatorPrivateKey.PublicKey)
		assert.True(t, strings.EqualFold(execOperatorAddress.String(), chainConfig.ExecOperatorAccountAddress))

		ethClient, err := ethereumClient.GetEthereumContractCaller()
		if err != nil {
			l.Sugar().Fatalf("failed to get Ethereum contract caller: %v", err)
		}

		anvil, err := testUtils.StartL1Anvil(root, ctx)
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

		chainId, err := ethClient.ChainID(ctx)
		if err != nil {
			l.Sugar().Fatalf("failed to get chain ID: %v", err)
		}
		t.Logf("Chain ID: %d", chainId.Uint64())

		coreContracts, err := config.GetCoreContractsForChainId(config.ChainId(chainId.Uint64()))
		if err != nil {
			l.Sugar().Fatalf("failed to get core contracts for chain ID %d: %v", chainId.Uint64(), err)
		}
		t.Logf("Core contracts: %+v", coreContracts)

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
			avsCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
				PrivateKey:          chainConfig.AVSAccountPrivateKey,
				AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
				TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
				KeyRegistrarAddress: coreContracts.KeyRegistrar,
			}, ethClient, l)
			if err != nil {
				t.Fatalf("failed to create contract caller: %v", err)
			}

			operatorCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
				PrivateKey:          tc.privateKey,
				AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
				TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
				KeyRegistrarAddress: coreContracts.KeyRegistrar,
			}, ethClient, l)
			if err != nil {
				t.Fatalf("Failed to create contract caller: %v", err)
			}

			pk, _, err := bn254.GenerateKeyPair()
			if err != nil {
				t.Fatalf("Failed to generate key pair: %v", err)
			}

			socket := "localhost:8545"
			result, err := operator.RegisterOperatorToOperatorSets(
				ctx,
				avsCc,
				operatorCc,
				common.HexToAddress(chainConfig.AVSAccountAddress),
				tc.operatorSets,
				&operator.Operator{
					TransactionPrivateKey: tc.privateKey,
					SigningPrivateKey:     pk,
					Curve:                 config.CurveTypeBN254,
				},
				&operator.RegistrationConfig{
					Socket:          socket,
					MetadataUri:     "https://some-metadata-uri.com",
					AllocationDelay: 7200,
				},
				l,
			)
			assert.Nil(t, err)
			fmt.Printf("Result: %+v\n", result)

			// create a peeringDataFetcher and get the data
			pdf := NewPeeringDataFetcher(operatorCc, l)

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
				assert.Equal(t, peers[0].OperatorSets[0].NetworkAddress, socket)

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

			testSig, err := pk.Sign(testMessage)
			if err != nil {
				t.Fatalf("Failed to sign message: %v", err)
			}

			valid, err := testSig.Verify(peers[0].OperatorSets[0].PublicKey.(*bn254.PublicKey), testMessage)
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

	t.Run("ECDSA", func(t *testing.T) {
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
		aggOperatorAddress := cryptoUtils.DeriveAddress(aggOperatorPrivateKey.PublicKey)
		assert.True(t, strings.EqualFold(aggOperatorAddress.String(), chainConfig.OperatorAccountAddress))

		aggOperatorSigningKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(chainConfig.OperatorAccountPrivateKey)
		if err != nil {
			l.Sugar().Fatalf("failed to convert private key: %v", err)
		}

		// executor operator
		execOperatorPrivateKey, err := cryptoUtils.StringToECDSAPrivateKey(chainConfig.ExecOperatorAccountPk)
		if err != nil {
			l.Sugar().Fatalf("failed to convert private key: %v", err)
		}
		execOperatorAddress := cryptoUtils.DeriveAddress(execOperatorPrivateKey.PublicKey)
		assert.True(t, strings.EqualFold(execOperatorAddress.String(), chainConfig.ExecOperatorAccountAddress))

		ethClient, err := ethereumClient.GetEthereumContractCaller()
		if err != nil {
			l.Sugar().Fatalf("failed to get Ethereum contract caller: %v", err)
		}

		anvil, err := testUtils.StartL1Anvil(root, ctx)
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

		chainId, err := ethClient.ChainID(ctx)
		if err != nil {
			l.Sugar().Fatalf("failed to get chain ID: %v", err)
		}
		t.Logf("Chain ID: %d", chainId.Uint64())

		coreContracts, err := config.GetCoreContractsForChainId(config.ChainId(chainId.Uint64()))
		if err != nil {
			l.Sugar().Fatalf("failed to get core contracts for chain ID %d: %v", chainId.Uint64(), err)
		}
		t.Logf("Core contracts: %+v", coreContracts)

		testCases := []struct {
			txPrivateKey      string
			operatorAddress   string
			operatorSets      []uint32
			operatorType      string
			privateSigningKey *cryptoLibsEcdsa.PrivateKey
		}{
			{
				txPrivateKey:      chainConfig.OperatorAccountPrivateKey,
				operatorAddress:   chainConfig.OperatorAccountAddress,
				operatorSets:      []uint32{0},
				operatorType:      "aggregator",
				privateSigningKey: aggOperatorSigningKey,
			}, /*{
				txPrivateKey:   chainConfig.ExecOperatorAccountPk,
				operatorAddress:      chainConfig.ExecOperatorAccountAddress,
				operatorSets: []uint32{1},
				operatorType: "executor",
			},*/
		}

		hasErrors := false
		for _, tc := range testCases {
			avsCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
				PrivateKey:          chainConfig.AVSAccountPrivateKey,
				AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
				TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
				KeyRegistrarAddress: coreContracts.KeyRegistrar,
			}, ethClient, l)
			if err != nil {
				t.Fatalf("failed to create contract caller: %v", err)
			}

			operatorCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
				PrivateKey:          tc.txPrivateKey,
				AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
				TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
				KeyRegistrarAddress: coreContracts.KeyRegistrar,
			}, ethClient, l)
			if err != nil {
				t.Fatalf("Failed to create contract caller: %v", err)
			}

			derivedAddress, err := tc.privateSigningKey.DeriveAddress()

			messageHash, _ := avsCc.GetOperatorECDSAKeyRegistrationMessageHash(
				ctx,
				common.HexToAddress(tc.operatorAddress),
				common.HexToAddress(chainConfig.AVSAccountAddress),
				tc.operatorSets[0],
				derivedAddress,
			)
			fmt.Printf("Message hash: %s\n", hexutil.Encode(messageHash[:]))

			nativePrivateKey, err := cryptoUtils.StringToECDSAPrivateKey(tc.txPrivateKey)
			if err != nil {
				t.Fatalf("Failed to convert private key: %+v", err)
			}

			sig, err := crypto.Sign(messageHash[:], nativePrivateKey)
			if err != nil {
				t.Fatalf("Failed to sign message hash: %v", err)
			}
			fmt.Printf("Raw sig: %+v\n", hexutil.Encode(sig))

			socket := "localhost:8545"
			result, err := operator.RegisterOperatorToOperatorSets(
				ctx,
				avsCc,
				operatorCc,
				common.HexToAddress(chainConfig.AVSAccountAddress),
				tc.operatorSets,
				&operator.Operator{
					TransactionPrivateKey: tc.txPrivateKey,
					SigningPrivateKey:     tc.privateSigningKey,
					Curve:                 config.CurveTypeECDSA,
				},
				&operator.RegistrationConfig{
					Socket:          socket,
					MetadataUri:     "https://some-metadata-uri.com",
					AllocationDelay: 7200,
				},
				l,
			)
			fmt.Printf("Result: %+v\n", result)
			if err != nil {
				t.Fatalf("Failed to register operator: %v", err)
			}

			// create a peeringDataFetcher and get the data
			pdf := NewPeeringDataFetcher(operatorCc, l)

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
				assert.Equal(t, peers[0].OperatorSets[0].NetworkAddress, socket)

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

			var hash [32]byte
			copy(hash[:], keccak256.Hash(testMessage))
			testSig, err := tc.privateSigningKey.Sign(hash)
			if err != nil {
				t.Fatalf("Failed to sign message: %v", err)
			}

			pubKey := peers[0].OperatorSets[0].PublicKey.(*cryptoLibsEcdsa.PublicKey)
			valid, err := testSig.Verify(pubKey, hash)
			if err != nil {
				t.Fatalf("Failed to verify signature: %v", err)
			}
			assert.True(t, valid)
		}

		cancel()
		select {
		case <-time.After(240 * time.Second):
			cancel()
			t.Fatalf("Test timed out after 10 seconds")
		case <-ctx.Done():
			t.Logf("Test completed")
		}

		_ = anvil.Process.Kill()
		assert.False(t, hasErrors)
	})

}
