package executor

import (
	"context"
	"fmt"
	"testing"
	"time"

	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	commonV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/common"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/auth"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// TestAuthenticationWithRealExecutor tests authentication using a real executor setup
func TestAuthenticationWithRealExecutor(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(180*time.Second))
	defer cancel()

	// Setup logger
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	// Get project root and chain config
	root := testUtils.GetProjectRootPath()
	chainConfig, err := testUtils.ReadChainConfig(root)
	require.NoError(t, err)

	// Setup executor config
	execConfig, err := executorConfig.NewExecutorConfigFromYamlBytes([]byte(authTestExecutorConfigYaml))
	require.NoError(t, err)

	// Use a specific port for testing
	execConfig.GrpcPort = 9091
	execConfig.Operator.SigningKeys.ECDSA = &config.ECDSAKeyConfig{
		PrivateKey: chainConfig.ExecOperatorAccountPk,
	}
	execConfig.Operator.Address = chainConfig.ExecOperatorAccountAddress
	if len(execConfig.AvsPerformers) > 0 {
		execConfig.AvsPerformers[0].AvsAddress = chainConfig.AVSAccountAddress
	}

	// Enable authentication for testing
	execConfig.AuthConfig = &auth.Config{
		IsEnabled: true,
	}

	// Parse keys for executor
	_, _, execGenericSigningKey, err := testUtils.ParseKeysFromConfig(execConfig.Operator, config.CurveTypeECDSA)
	require.NoError(t, err)
	execSigner := inMemorySigner.NewInMemorySigner(execGenericSigningKey, config.CurveTypeECDSA)

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

	// Create executor
	signers := signer.Signers{
		ECDSASigner: execSigner,
	}
	store := memory.NewInMemoryExecutorStore()

	exec, err := NewExecutorWithRpcServers(execConfig.GrpcPort, execConfig.GrpcPort, execConfig, l, signers, pdf, l1CC, store)
	require.NoError(t, err)

	err = exec.Initialize(ctx)
	require.NoError(t, err)

	// Start executor
	go func() {
		if err := exec.Run(ctx); err != nil {
			t.Logf("Executor stopped: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	serverAddr := fmt.Sprintf("localhost:%d", execConfig.GrpcPort)

	// Test 1: GetChallengeToken with correct operator address
	t.Run("GetChallengeToken_Success", func(t *testing.T) {
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		client := executorV1.NewExecutorManagementServiceClient(conn)

		resp, err := client.GetChallengeToken(ctx, &executorV1.GetChallengeTokenRequest{
			OperatorAddress: execConfig.Operator.Address,
		})

		assert.NoError(t, err)
		assert.NotEmpty(t, resp.ChallengeToken)
		assert.Greater(t, resp.ExpiresAt, int64(0))
	})

	// Test 2: GetChallengeToken with wrong operator address
	t.Run("GetChallengeToken_WrongOperator", func(t *testing.T) {
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		client := executorV1.NewExecutorManagementServiceClient(conn)

		_, err = client.GetChallengeToken(ctx, &executorV1.GetChallengeTokenRequest{
			OperatorAddress: "0xWrongOperator",
		})

		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, statusErr.Code())
	})

	// Test 3: Authenticated ListPerformers request
	t.Run("Authenticated_ListPerformers", func(t *testing.T) {
		authClient, err := executorClient.NewAuthenticatedExecutorClient(
			serverAddr,
			execConfig.Operator.Address,
			execSigner,
			true,
		)
		require.NoError(t, err)

		resp, err := authClient.ListPerformers(ctx, &executorV1.ListPerformersRequest{
			AvsAddress: "",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		// Empty list is fine, we're testing auth not performers
	})

	// Test 4: Unauthenticated ListPerformers request should fail
	t.Run("Unauthenticated_ListPerformers_Fails", func(t *testing.T) {
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		client := executorV1.NewExecutorManagementServiceClient(conn)

		_, err = client.ListPerformers(ctx, &executorV1.ListPerformersRequest{
			AvsAddress: "",
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

		managementClient := executorV1.NewExecutorManagementServiceClient(conn)

		// Get a challenge token
		tokenResp, err := managementClient.GetChallengeToken(ctx, &executorV1.GetChallengeTokenRequest{
			OperatorAddress: execConfig.Operator.Address,
		})
		require.NoError(t, err)

		// Create auth signature (simplified - only sign the token)
		signedMessage := auth.ConstructSignedMessage(tokenResp.ChallengeToken)
		signature, err := execSigner.SignMessage(signedMessage)
		require.NoError(t, err)

		req := &executorV1.ListPerformersRequest{
			AvsAddress: "",
			Auth: &commonV1.AuthSignature{
				ChallengeToken: tokenResp.ChallengeToken,
				Signature:      signature,
			},
		}

		// First request should succeed
		_, err = managementClient.ListPerformers(ctx, req)
		assert.NoError(t, err)

		// Second request with same token should fail
		_, err = managementClient.ListPerformers(ctx, req)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	})

	// Test 6: SubmitTask should work without authentication
	t.Run("SubmitTask_No_Auth_Required", func(t *testing.T) {
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		client := executorV1.NewExecutorServiceClient(conn)

		// This will fail with "AVS performer not found" but that's OK
		// We're testing that auth is not required
		_, err = client.SubmitTask(ctx, &executorV1.TaskSubmission{
			TaskId:            "task-123",
			AggregatorAddress: "0xAggregator",
			AvsAddress:        chainConfig.AVSAccountAddress,
			Payload:           []byte("test"),
			Signature:         []byte("sig"),
			TaskBlockNumber:   0,
			OperatorSetId:     1,
		})

		// Should get Internal error (performer not found), not Unauthenticated
		if err != nil {
			statusErr, ok := status.FromError(err)
			if ok {
				assert.NotEqual(t, codes.Unauthenticated, statusErr.Code())
			}
		}
	})

	// Test 7: RemovePerformer with authentication
	t.Run("Authenticated_RemovePerformer", func(t *testing.T) {
		authClient, err := executorClient.NewAuthenticatedExecutorClient(
			serverAddr,
			execConfig.Operator.Address,
			execSigner,
			true,
		)
		require.NoError(t, err)

		_, err = authClient.RemovePerformer(ctx, &executorV1.RemovePerformerRequest{
			PerformerId: "non-existent",
		})

		// Will fail with NotFound, but auth should pass
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, statusErr.Code())
	})
}

// TestChallengeTokenManager tests the challenge token manager directly
func TestChallengeTokenManager(t *testing.T) {
	operatorAddress := "0xTestOperator123"
	tokenManager := auth.NewChallengeTokenManager(operatorAddress, 5*time.Minute)

	t.Run("GenerateToken_Success", func(t *testing.T) {
		entry, err := tokenManager.GenerateChallengeToken(operatorAddress)
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.NotEmpty(t, entry.Token)
		assert.False(t, entry.Used)
	})

	t.Run("GenerateToken_WrongOperator", func(t *testing.T) {
		_, err := tokenManager.GenerateChallengeToken("0xWrongOperator")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "entity mismatch")
	})

	t.Run("UseToken_Success", func(t *testing.T) {
		entry, err := tokenManager.GenerateChallengeToken(operatorAddress)
		require.NoError(t, err)

		err = tokenManager.UseChallengeToken(entry.Token)
		assert.NoError(t, err)
	})

	t.Run("UseToken_AlreadyUsed", func(t *testing.T) {
		entry, err := tokenManager.GenerateChallengeToken(operatorAddress)
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
		shortExpiryManager := auth.NewChallengeTokenManager(operatorAddress, 1*time.Millisecond)
		entry, err := shortExpiryManager.GenerateChallengeToken(operatorAddress)
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(2 * time.Millisecond)

		err = shortExpiryManager.UseChallengeToken(entry.Token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})
}

// TestAuthVerifier tests the auth verifier directly
func TestAuthVerifier(t *testing.T) {
	operatorAddress := "0xTestOperator123"
	tokenManager := auth.NewChallengeTokenManager(operatorAddress, 5*time.Minute)

	// Create a test signer with a proper ECDSA private key
	// This is a test private key for testing purposes only
	testPrivateKeyHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	testPrivateKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(testPrivateKeyHex)
	require.NoError(t, err)

	testSigner := inMemorySigner.NewInMemorySigner(testPrivateKey, config.CurveTypeECDSA)
	verifier := auth.NewVerifier(tokenManager, testSigner)

	t.Run("VerifyAuthentication_Success", func(t *testing.T) {
		// Generate a token
		entry, err := verifier.GenerateChallengeToken(operatorAddress)
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
		entry, err := verifier.GenerateChallengeToken(operatorAddress)
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

// Minimal executor config for authentication testing
const authTestExecutorConfigYaml = `
---
grpcPort: 9091
performerNetworkName: "test-network"
operator:
  address: "0xoperator..."
  operatorPrivateKey:
    privateKey: "..."
  signingKeys:
    ecdsa:
      privateKey: "..."
l1Chain:
  rpcUrl: "http://localhost:8545"
  chainId: 1
avsPerformers: []
`
