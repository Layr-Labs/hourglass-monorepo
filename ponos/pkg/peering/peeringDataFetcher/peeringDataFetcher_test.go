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
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/ethereum/go-ethereum/common"
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

		_ = testUtils.KillallAnvils()

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

		// First, create the AVS private key signer and contract caller for operator set configuration
		avsPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, ethClient, l)
		if err != nil {
			t.Fatalf("Failed to create AVS private key signer: %v", err)
		}

		avsCc, err := caller.NewContractCaller(ethClient, avsPrivateKeySigner, l)
		if err != nil {
			t.Fatalf("failed to create contract caller: %v", err)
		}

		// Configure operator set with BN254 curve type
		_, err = avsCc.ConfigureAVSOperatorSet(
			ctx,
			common.HexToAddress(chainConfig.AVSAccountAddress),
			0, // aggregator operator set
			config.CurveTypeBN254,
		)
		if err != nil {
			t.Fatalf("Failed to configure AVS operator set: %v", err)
		}

		// Create generation reservation for BN254 operator set
		_, err = avsCc.CreateGenerationReservation(
			ctx,
			common.HexToAddress(chainConfig.AVSAccountAddress),
			0, // operator set id
			common.HexToAddress(caller.BN254TableCalculatorAddress),
			common.HexToAddress(chainConfig.AVSAccountAddress), // AVS is the owner
			0, // maxStalenessPeriod - 0 means always valid
		)
		if err != nil {
			t.Fatalf("Failed to create generation reservation: %v", err)
		}

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
			},
		}

		hasErrors := false
		for _, tc := range testCases {

			operatorPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(tc.privateKey, ethClient, l)
			if err != nil {
				t.Fatalf("Failed to create operator private key signer: %v", err)
			}

			operatorCc, err := caller.NewContractCaller(ethClient, operatorPrivateKeySigner, l)
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
					AllocationDelay: 1,
				},
				l,
			)
			fmt.Printf("Result: %+v\n", result)
			if err != nil {
				t.Fatalf("Failed to register operator: %v", err)
			}

			// create a peeringDataFetcher and get the data
			pdf := NewPeeringDataFetcher(operatorCc, l)

			// Check AVS config first
			avsConfig, err := operatorCc.GetAVSConfig(chainConfig.AVSAccountAddress, 0)
			if err != nil {
				t.Fatalf("Failed to get AVS config: %v", err)
			}
			t.Logf("AVS Config - AggregatorOperatorSetId: %d, ExecutorOperatorSetIds: %v",
				avsConfig.AggregatorOperatorSetId, avsConfig.ExecutorOperatorSetIds)

			// BN254 test only has aggregator operators
			var peers []*peering.OperatorPeerInfo
			peers, err = pdf.ListAggregatorOperators(ctx, chainConfig.AVSAccountAddress, 0)
			if err != nil {
				t.Fatalf("Failed to list aggregator operators: %v", err)
			}
			assert.Equal(t, 1, len(peers))

			for _, peer := range peers {
				t.Logf("Aggregator Peer: %+v\n", peer)
			}

			testMessage := []byte("test message")

			testSig, err := pk.Sign(testMessage)
			if err != nil {
				t.Fatalf("Failed to sign message: %v", err)
			}

			wrappedPubKey := peers[0].OperatorSets[0].WrappedPublicKey

			valid, err := testSig.Verify(wrappedPubKey.PublicKey.(*bn254.PublicKey), testMessage)
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

		_ = testUtils.KillAnvil(anvil)
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

		// executor operator
		execOperatorPrivateKey, err := cryptoUtils.StringToECDSAPrivateKey(chainConfig.ExecOperatorAccountPk)
		if err != nil {
			l.Sugar().Fatalf("failed to convert private key: %v", err)
		}
		execOperatorAddress := cryptoUtils.DeriveAddress(execOperatorPrivateKey.PublicKey)
		assert.True(t, strings.EqualFold(execOperatorAddress.String(), chainConfig.ExecOperatorAccountAddress))

		execOperatorSigningKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(chainConfig.ExecOperatorAccountPk)
		if err != nil {
			l.Sugar().Fatalf("failed to convert private key: %v", err)
		}

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

		// First, create the AVS private key signer and contract caller for operator set configuration
		avsPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, ethClient, l)
		if err != nil {
			t.Fatalf("Failed to create AVS private key signer: %v", err)
		}

		avsCc, err := caller.NewContractCaller(ethClient, avsPrivateKeySigner, l)
		if err != nil {
			t.Fatalf("failed to create contract caller: %v", err)
		}

		// Configure operator set with ECDSA curve type for executor
		_, err = avsCc.ConfigureAVSOperatorSet(
			ctx,
			common.HexToAddress(chainConfig.AVSAccountAddress),
			1, // executor operator set
			config.CurveTypeECDSA,
		)
		if err != nil {
			t.Fatalf("Failed to configure AVS operator set: %v", err)
		}

		// Create generation reservation for ECDSA operator set
		_, err = avsCc.CreateGenerationReservation(
			ctx,
			common.HexToAddress(chainConfig.AVSAccountAddress),
			1, // operator set id
			common.HexToAddress(caller.ECDSATableCalculatorAddress),
			common.HexToAddress(chainConfig.AVSAccountAddress), // AVS is the owner
			0, // maxStalenessPeriod - 0 means always valid
		)
		if err != nil {
			t.Fatalf("Failed to create generation reservation: %v", err)
		}

		testCases := []struct {
			txPrivateKey      string
			operatorAddress   string
			operatorSets      []uint32
			operatorType      string
			privateSigningKey *cryptoLibsEcdsa.PrivateKey
		}{
			{
				txPrivateKey:      chainConfig.ExecOperatorAccountPk,
				operatorAddress:   chainConfig.ExecOperatorAccountAddress,
				operatorSets:      []uint32{1},
				operatorType:      "executor",
				privateSigningKey: execOperatorSigningKey,
			},
		}

		hasErrors := false
		for _, tc := range testCases {

			operatorPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(tc.txPrivateKey, ethClient, l)
			if err != nil {
				t.Fatalf("Failed to create operator private key signer: %v", err)
			}

			operatorCc, err := caller.NewContractCaller(ethClient, operatorPrivateKeySigner, l)
			if err != nil {
				t.Fatalf("Failed to create contract caller: %v", err)
			}

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
					AllocationDelay: 1,
				},
				l,
			)
			fmt.Printf("Result: %+v\n", result)
			if err != nil {
				t.Errorf("Failed to register operator: %v", err)
				hasErrors = true
				cancel()
				break
			}

			// create a peeringDataFetcher and get the data
			pdf := NewPeeringDataFetcher(operatorCc, l)

			// Check AVS config first
			avsConfig, err := operatorCc.GetAVSConfig(chainConfig.AVSAccountAddress, 0)
			if err != nil {
				t.Fatalf("Failed to get AVS config: %v", err)
			}
			t.Logf("AVS Config - AggregatorOperatorSetId: %d, ExecutorOperatorSetIds: %v",
				avsConfig.AggregatorOperatorSetId, avsConfig.ExecutorOperatorSetIds)

			// ECDSA test only has executor operators
			var peers []*peering.OperatorPeerInfo
			peers, err = pdf.ListExecutorOperators(ctx, chainConfig.AVSAccountAddress, 0)
			if err != nil {
				t.Fatalf("Failed to list executor operators: %v", err)
			}
			t.Logf("Listed %d executor operators", len(peers))
			assert.Equal(t, 1, len(peers))
			for _, peer := range peers {
				t.Logf("Executor Peer: %+v\n", peer)
			}
			assert.Equal(t, peers[0].OperatorSets[0].NetworkAddress, socket)

			testMessage := []byte("test message")

			hash := keccak256.Hash(testMessage)

			testSig, err := tc.privateSigningKey.Sign(hash)
			if err != nil {
				t.Errorf("Failed to sign message: %v", err)
				hasErrors = true
				cancel()
				break
			}

			addr := peers[0].OperatorSets[0].WrappedPublicKey.ECDSAAddress
			valid, err := testSig.VerifyWithAddress(hash[:], addr)
			if err != nil {
				t.Errorf("Failed to verify signature: %v", err)
				hasErrors = true
				cancel()
				break
			}
			assert.True(t, valid)
		}

		cancel()
		select {
		case <-time.After(240 * time.Second):
			cancel()
			t.Logf("Test timed out after 240 seconds")
		case <-ctx.Done():
			t.Logf("Test completed")
		}

		assert.False(t, hasErrors)
		_ = testUtils.KillAnvil(anvil)
	})

}
