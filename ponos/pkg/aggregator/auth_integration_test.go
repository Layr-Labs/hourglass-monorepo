package aggregator

import (
	"context"
	"fmt"
	"testing"
	"time"

	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	commonV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/common"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/auth"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/aggregatorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// TestAuthenticationWithRealAggregator tests authentication using a real aggregator setup
func TestAuthenticationWithRealAggregator(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(180*time.Second))
	defer cancel()

	// Setup logger
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	// Get project root and chain config
	root := testUtils.GetProjectRootPath()
	chainConfig, err := testUtils.ReadChainConfig(root)
	require.NoError(t, err)

	// Setup aggregator config
	aggConfig, err := aggregatorConfig.NewAggregatorConfigFromYamlBytes([]byte(authTestAggregatorConfigYaml))
	require.NoError(t, err)

	// Use a specific port for testing
	managementGrpcPort := 9092

	// Setup operator config
	aggConfig.Operator.Address = chainConfig.OperatorAccountAddress
	aggConfig.Operator.OperatorPrivateKey = &config.ECDSAKeyConfig{
		PrivateKey: chainConfig.OperatorAccountPrivateKey,
	}
	aggConfig.L1ChainId = 1

	// Parse keys for aggregator - using ECDSA for auth testing to match executor tests
	// This simplifies test setup as ECDSA just needs a hex private key
	_, _, aggGenericSigningKey, err := testUtils.ParseKeysFromConfig(aggConfig.Operator, config.CurveTypeECDSA)
	require.NoError(t, err)
	aggSigner := inMemorySigner.NewInMemorySigner(aggGenericSigningKey, config.CurveTypeECDSA)

	// Setup Ethereum client
	l1EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   "http://127.0.0.1:8545",
		BlockType: ethereum.BlockType_Latest,
	}, l)

	l1EthClient, err := l1EthereumClient.GetEthereumContractCaller()
	require.NoError(t, err)

	// Start Anvil
	_ = testUtils.KillallAnvils()
	l1Anvil, err := testUtils.StartL1Anvil(root, ctx)
	require.NoError(t, err)
	defer func() {
		_ = testUtils.KillAnvil(l1Anvil)
	}()

	// Wait for Anvil to be ready
	time.Sleep(2 * time.Second)

	// Setup contract caller
	l1PrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AppAccountPrivateKey, l1EthClient, l)
	require.NoError(t, err)

	l1CC, err := caller.NewContractCaller(l1EthClient, l1PrivateKeySigner, l)
	require.NoError(t, err)

	// Setup peering data fetcher
	pdf := peeringDataFetcher.NewPeeringDataFetcher(l1CC, l)

	// Setup contract store and transaction log parser
	// Load EigenLayer contracts that include mailbox contracts needed for AVS registration
	eigenlayerContracts, err := eigenlayer.LoadContracts()
	require.NoError(t, err)

	contractStore := inMemoryContractStore.NewInMemoryContractStore(eigenlayerContracts, l)
	tlp := transactionLogParser.NewTransactionLogParser(contractStore, l)

	// Enable authentication for testing
	authConfig := &auth.Config{
		IsEnabled: true,
	}

	// Create aggregator - using ECDSA signer for auth
	signers := signer.Signers{
		ECDSASigner: aggSigner,
	}
	store := memory.NewInMemoryAggregatorStore()

	agg, err := NewAggregatorWithManagementRpcServer(
		managementGrpcPort,
		&AggregatorConfig{
			Address:          aggConfig.Operator.Address,
			PrivateKeyConfig: aggConfig.Operator.OperatorPrivateKey,
			AVSs:             aggConfig.Avss,
			Chains:           aggConfig.Chains,
			L1ChainId:        aggConfig.L1ChainId,
			Authentication:   authConfig,
		},
		contractStore,
		tlp,
		pdf,
		signers,
		store,
		l,
	)
	require.NoError(t, err)

	err = agg.Initialize()
	require.NoError(t, err)

	// Start aggregator
	go func() {
		if err := agg.Start(ctx); err != nil {
			t.Logf("Aggregator stopped: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	serverAddr := fmt.Sprintf("localhost:%d", managementGrpcPort)

	// Test 1: GetChallengeToken with correct aggregator address
	t.Run("GetChallengeToken_Success", func(t *testing.T) {
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		client := aggregatorV1.NewAggregatorManagementServiceClient(conn)

		resp, err := client.GetChallengeToken(ctx, &aggregatorV1.AggregatorGetChallengeTokenRequest{
			AggregatorAddress: aggConfig.Operator.Address,
		})

		assert.NoError(t, err)
		assert.NotEmpty(t, resp.ChallengeToken)
		assert.Greater(t, resp.ExpiresAt, int64(0))
	})

	// Test 2: GetChallengeToken with wrong aggregator address
	t.Run("GetChallengeToken_WrongAggregator", func(t *testing.T) {
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		client := aggregatorV1.NewAggregatorManagementServiceClient(conn)

		_, err = client.GetChallengeToken(ctx, &aggregatorV1.AggregatorGetChallengeTokenRequest{
			AggregatorAddress: "0xWrongAggregator",
		})

		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, statusErr.Code())
	})

	// Test 3: Authenticated RegisterAvs request
	t.Run("Authenticated_RegisterAvs", func(t *testing.T) {
		authClient, err := aggregatorClient.NewAuthenticatedAggregatorClient(
			serverAddr,
			aggConfig.Operator.Address,
			aggSigner,
			true,
		)
		require.NoError(t, err)

		resp, err := authClient.RegisterAvs(ctx, &aggregatorV1.RegisterAvsRequest{
			AvsAddress: chainConfig.AVSAccountAddress,
			ChainIds:   []uint32{1},
		})

		// We expect this to succeed with authentication
		// The actual registration may fail due to missing contracts, but auth should pass
		if err != nil {
			statusErr, ok := status.FromError(err)
			if ok {
				// Should not be an authentication error
				assert.NotEqual(t, codes.Unauthenticated, statusErr.Code())
			}
		} else {
			assert.NotNil(t, resp)
		}
	})

	// Test 4: Unauthenticated RegisterAvs request should fail
	t.Run("Unauthenticated_RegisterAvs_Fails", func(t *testing.T) {
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		client := aggregatorV1.NewAggregatorManagementServiceClient(conn)

		_, err = client.RegisterAvs(ctx, &aggregatorV1.RegisterAvsRequest{
			AvsAddress: chainConfig.AVSAccountAddress,
			ChainIds:   []uint32{1},
		})

		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	})

	// Test 5: Token reuse should fail
	t.Run("Token_Reuse_Fails", func(t *testing.T) {
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		managementClient := aggregatorV1.NewAggregatorManagementServiceClient(conn)

		// Get a challenge token
		tokenResp, err := managementClient.GetChallengeToken(ctx, &aggregatorV1.AggregatorGetChallengeTokenRequest{
			AggregatorAddress: aggConfig.Operator.Address,
		})
		require.NoError(t, err)

		// Create auth signature
		signedMessage := auth.ConstructSignedMessage(tokenResp.ChallengeToken)
		signature, err := aggSigner.SignMessage(signedMessage)
		require.NoError(t, err)

		// Use a different AVS address to avoid "already exists" conflicts from previous tests
		tokenTestAvsAddress := "0xcafebabecafebabecafebabecafebabecafebabe"

		req := &aggregatorV1.RegisterAvsRequest{
			AvsAddress: tokenTestAvsAddress,
			ChainIds:   []uint32{1},
			Auth: &commonV1.AuthSignature{
				ChallengeToken: tokenResp.ChallengeToken,
				Signature:      signature,
			},
		}

		// First request should succeed (auth-wise)
		_, err = managementClient.RegisterAvs(ctx, req)
		// May fail due to other reasons, but not auth
		if err != nil {
			statusErr, ok := status.FromError(err)
			require.True(t, ok)
			// Should not be an authentication error
			assert.NotEqual(t, codes.Unauthenticated, statusErr.Code())
		}

		// Second request with same token should fail with auth error
		_, err = managementClient.RegisterAvs(ctx, req)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	})

	// Test 6: DeRegisterAvs with authentication
	t.Run("Authenticated_DeRegisterAvs", func(t *testing.T) {
		testCtx, testCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer testCancel()

		authClient, err := aggregatorClient.NewAuthenticatedAggregatorClient(
			serverAddr,
			aggConfig.Operator.Address,
			aggSigner,
			true,
		)
		require.NoError(t, err)

		// Use a different AVS address to ensure it's not registered
		unregisteredAvsAddress := "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"

		_, err = authClient.DeRegisterAvs(testCtx, &aggregatorV1.DeRegisterAvsRequest{
			AvsAddress: unregisteredAvsAddress,
		})

		// Should fail because AVS was never registered, but auth should pass
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, statusErr.Code()) // Not an auth error
		assert.Contains(t, statusErr.Message(), "not registered")
	})
}

// TestAggregatorChallengeTokenManager tests the challenge token manager for aggregator
func TestAggregatorChallengeTokenManager(t *testing.T) {
	aggregatorAddress := "0xTestAggregator123"
	tokenManager := auth.NewChallengeTokenManager(aggregatorAddress, 5*time.Minute)

	t.Run("GenerateToken_Success", func(t *testing.T) {
		entry, err := tokenManager.GenerateChallengeToken(aggregatorAddress)
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.NotEmpty(t, entry.Token)
		assert.False(t, entry.Used)
	})

	t.Run("GenerateToken_WrongAggregator", func(t *testing.T) {
		_, err := tokenManager.GenerateChallengeToken("0xWrongAggregator")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "entity mismatch")
	})

	t.Run("UseToken_Success", func(t *testing.T) {
		entry, err := tokenManager.GenerateChallengeToken(aggregatorAddress)
		require.NoError(t, err)

		err = tokenManager.UseChallengeToken(entry.Token)
		assert.NoError(t, err)
	})

	t.Run("UseToken_AlreadyUsed", func(t *testing.T) {
		entry, err := tokenManager.GenerateChallengeToken(aggregatorAddress)
		require.NoError(t, err)

		err = tokenManager.UseChallengeToken(entry.Token)
		require.NoError(t, err)

		// Try to use again
		err = tokenManager.UseChallengeToken(entry.Token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already used")
	})

	t.Run("UseToken_NotFound", func(t *testing.T) {
		err := tokenManager.UseChallengeToken("invalid-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("UseToken_Expired", func(t *testing.T) {
		// Create manager with very short expiration
		shortExpiryManager := auth.NewChallengeTokenManager(aggregatorAddress, 1*time.Millisecond)
		entry, err := shortExpiryManager.GenerateChallengeToken(aggregatorAddress)
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(2 * time.Millisecond)

		err = shortExpiryManager.UseChallengeToken(entry.Token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})
}

// TestAggregatorAuthVerifier tests the auth verifier for aggregator
func TestAggregatorAuthVerifier(t *testing.T) {
	aggregatorAddress := "0xTestAggregator123"
	tokenManager := auth.NewChallengeTokenManager(aggregatorAddress, 5*time.Minute)

	// Create a test signer with ECDSA private key (matching executor pattern)
	testPrivateKeyHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	testPrivateKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(testPrivateKeyHex)
	require.NoError(t, err)

	testSigner := inMemorySigner.NewInMemorySigner(testPrivateKey, config.CurveTypeECDSA)
	verifier := auth.NewVerifier(tokenManager, testSigner)

	t.Run("VerifyAuthentication_Success", func(t *testing.T) {
		// Generate a token
		entry, err := verifier.GenerateChallengeToken(aggregatorAddress)
		require.NoError(t, err)

		// Create auth signature
		signedMessage := auth.ConstructSignedMessage(entry.Token)
		signature, err := testSigner.SignMessage(signedMessage)
		require.NoError(t, err)

		authSig := &commonV1.AuthSignature{
			ChallengeToken: entry.Token,
			Signature:      signature,
		}

		// Verify
		err = verifier.VerifyAuthentication(authSig)
		assert.NoError(t, err)
	})

	t.Run("VerifyAuthentication_MissingAuth", func(t *testing.T) {
		err := verifier.VerifyAuthentication(nil)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	})

	t.Run("VerifyAuthentication_InvalidToken", func(t *testing.T) {
		authSig := &commonV1.AuthSignature{
			ChallengeToken: "invalid-token",
			Signature:      []byte("signature"),
		}

		err := verifier.VerifyAuthentication(authSig)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	})

	t.Run("VerifyAuthentication_WrongSignature", func(t *testing.T) {
		// Generate a token
		entry, err := verifier.GenerateChallengeToken(aggregatorAddress)
		require.NoError(t, err)

		authSig := &commonV1.AuthSignature{
			ChallengeToken: entry.Token,
			Signature:      []byte("wrong-signature"),
		}

		err = verifier.VerifyAuthentication(authSig)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	})
}

// Minimal aggregator config for authentication testing
const authTestAggregatorConfigYaml = `
---
managementServerGrpcPort: 9092
l1ChainId: 1
operator:
  address: "0xaggregator..."
  operatorPrivateKey:
    privateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
  signingKeys:
    ecdsa:
      privateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
chains:
  - chainId: 1
    name: "L1 Anvil"
    rpcURL: "http://localhost:8545"
avss: []
`
