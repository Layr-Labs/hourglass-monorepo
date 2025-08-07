package executor

import (
	"context"
	"net"
	"testing"
	"time"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/auth"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// mockContractCaller implements contractCaller.IContractCaller for testing
type mockContractCaller struct{}

func (m *mockContractCaller) GetOperatorSetCurveType(avsAddress string, operatorSetId uint32) (config.CurveType, error) {
	return config.CurveTypeECDSA, nil
}

func (m *mockContractCaller) CalculateECDSACertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, digestBytes []byte) ([]byte, error) {
	return digestBytes, nil
}

// mockPeeringFetcher implements peering.IPeeringDataFetcher for testing
type mockPeeringFetcher struct{}

func (m *mockPeeringFetcher) GetPeeringData(avsAddress string) (*peering.PeeringData, error) {
	return &peering.PeeringData{}, nil
}

// TestAuthenticationIntegration tests the full authentication flow
func TestAuthenticationIntegration(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	// Create test config
	testConfig := &executorConfig.ExecutorConfig{
		Debug:                false,
		GrpcPort:             0, // Let system assign port
		PerformerNetworkName: "test-network",
		Operator: &config.OperatorConfig{
			Address:            "0xTestOperator123",
			OperatorPrivateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		AvsPerformers: []*executorConfig.AvsPerformerConfig{},
		L1Chain: &executorConfig.Chain{
			RpcUrl:  "http://localhost:8545",
			ChainId: 1,
		},
	}

	// Create signers
	mockSigners := signer.Signers{
		ECDSASigner: &mockSigner{},
		BLSSigner:   nil,
	}

	// Create RPC server
	rpcServerInstance, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: 0, // Let system assign port
	}, logger)
	require.NoError(t, err)

	// Create executor
	executor := NewExecutor(
		testConfig,
		rpcServerInstance,
		logger,
		mockSigners,
		&mockPeeringFetcher{},
		&mockContractCaller{},
	)

	// Initialize executor (registers handlers)
	err = executor.Initialize(ctx)
	require.NoError(t, err)

	// Start RPC server in background
	go func() {
		err := rpcServerInstance.Start(ctx)
		if err != nil && err != grpc.ErrServerStopped {
			t.Logf("RPC server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Get the actual port the server is listening on
	listener := rpcServerInstance.GetListener()
	require.NotNil(t, listener)
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)

	serverAddr := "localhost:" + port

	t.Run("Authenticated_DeployArtifact_Success", func(t *testing.T) {
		// Create authenticated client
		authClient, err := executorClient.NewAuthenticatedExecutorClient(
			serverAddr,
			testConfig.Operator.Address,
			mockSigners.ECDSASigner,
			true, // insecure connection for testing
		)
		require.NoError(t, err)

		// Make authenticated request
		resp, err := authClient.DeployArtifact(ctx, &executorV1.DeployArtifactRequest{
			AvsAddress:  "0xAVS123",
			Digest:      "sha256:abc123",
			RegistryUrl: "registry.example.com",
			Env: []*executorV1.PerformerEnv{
				{
					Name:  "TEST_ENV",
					Value: "test_value",
				},
			},
		})

		// Should fail with "AVS performer not found" since we haven't set up performers
		// But authentication should pass
		assert.Error(t, err)
		assert.Nil(t, resp)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "Failed to get or create AVS performer")
	})

	t.Run("Unauthenticated_DeployArtifact_Fails", func(t *testing.T) {
		// Create regular client (no authentication)
		conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		managementClient := executorV1.NewExecutorManagementServiceClient(conn)

		// Make request without auth
		resp, err := managementClient.DeployArtifact(ctx, &executorV1.DeployArtifactRequest{
			AvsAddress:  "0xAVS123",
			Digest:      "sha256:abc123",
			RegistryUrl: "registry.example.com",
			// No Auth field
		})

		assert.Error(t, err)
		assert.NotNil(t, resp) // Error response with success=false
		assert.False(t, resp.Success)
		assert.Equal(t, "Authentication failed", resp.Message)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	})

	t.Run("Invalid_Signature_Fails", func(t *testing.T) {
		// Create client with wrong signer
		wrongSigner := &mockSigner{
			signFunc: func(data []byte) ([]byte, error) {
				// Return wrong signature
				return []byte("wrong-signature"), nil
			},
		}

		authClient, err := executorClient.NewAuthenticatedExecutorClient(
			serverAddr,
			testConfig.Operator.Address,
			wrongSigner,
			true,
		)
		require.NoError(t, err)

		// Make request with wrong signature
		resp, err := authClient.DeployArtifact(ctx, &executorV1.DeployArtifactRequest{
			AvsAddress:  "0xAVS123",
			Digest:      "sha256:abc123",
			RegistryUrl: "registry.example.com",
		})

		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Success)
		assert.Equal(t, "Authentication failed", resp.Message)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	})

	t.Run("Reused_ChallengeToken_Fails", func(t *testing.T) {
		// Get a nonce directly
		conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		managementClient := executorV1.NewExecutorManagementServiceClient(conn)

		// Get challenge token
		tokenResp, err := managementClient.GetChallengeToken(ctx, &executorV1.GetChallengeTokenRequest{
			OperatorAddress: testConfig.Operator.Address,
		})
		require.NoError(t, err)
		require.NotEmpty(t, tokenResp.ChallengeToken)

		// Create auth signature manually
		req := &executorV1.ListPerformersRequest{
			AvsAddress: "0xAVS123",
		}
		requestBytes, err := auth.GetRequestWithoutAuth(req)
		require.NoError(t, err)

		signedMessage := auth.ConstructSignedMessage(tokenResp.ChallengeToken, "ListPerformers", requestBytes)
		signature, err := mockSigners.ECDSASigner.SignMessage(signedMessage)
		require.NoError(t, err)

		req.Auth = &executorV1.AuthSignature{
			ChallengeToken: tokenResp.ChallengeToken,
			Signature: signature,
		}

		// First request should succeed
		_, err = managementClient.ListPerformers(ctx, req)
		assert.NoError(t, err)

		// Second request with same nonce should fail
		_, err = managementClient.ListPerformers(ctx, req)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "invalid nonce")
	})

	t.Run("Wrong_Operator_Address_Fails", func(t *testing.T) {
		conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		managementClient := executorV1.NewExecutorManagementServiceClient(conn)

		// Try to get challenge token with wrong operator address
		_, err = managementClient.GetChallengeToken(ctx, &executorV1.GetChallengeTokenRequest{
			OperatorAddress: "0xWrongOperator",
		})

		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "operator address mismatch")
	})

	t.Run("Authenticated_ListPerformers_Success", func(t *testing.T) {
		authClient, err := executorClient.NewAuthenticatedExecutorClient(
			serverAddr,
			testConfig.Operator.Address,
			mockSigners.ECDSASigner,
			true,
		)
		require.NoError(t, err)

		// Should succeed with empty list since no performers are configured
		resp, err := authClient.ListPerformers(ctx, &executorV1.ListPerformersRequest{
			AvsAddress: "0xAVS123",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Empty(t, resp.Performers)
	})

	t.Run("Authenticated_RemovePerformer_Success", func(t *testing.T) {
		authClient, err := executorClient.NewAuthenticatedExecutorClient(
			serverAddr,
			testConfig.Operator.Address,
			mockSigners.ECDSASigner,
			true,
		)
		require.NoError(t, err)

		// Should fail with "performer not found" but authentication should pass
		resp, err := authClient.RemovePerformer(ctx, &executorV1.RemovePerformerRequest{
			PerformerId: "non-existent-performer",
		})

		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "performer with ID non-existent-performer not found")

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, statusErr.Code())
	})

	t.Run("SubmitTask_No_Auth_Required", func(t *testing.T) {
		// Create regular client (no authentication)
		conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		managementClient := executorV1.NewExecutorManagementServiceClient(conn)

		// SubmitTask should work without authentication
		_, err = client.SubmitTask(ctx, &executorV1.TaskSubmission{
			TaskId:             "task-123",
			AggregatorAddress:  "0xAggregator",
			AvsAddress:         "0xAVS123",
			Payload:            []byte("test-payload"),
			Signature:          []byte("task-signature"),
			OperatorSetId:      1,
			ReferenceTimestamp: 12345,
		})

		// Will fail with "AVS performer not found" but not authentication error
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "AVS performer not found")
	})

	t.Run("ChallengeToken_Expiration", func(t *testing.T) {
		// Create a new auth verifier with short token expiry
		shortExpiryTokenManager := auth.NewChallengeTokenManager(testConfig.Operator.Address, 100*time.Millisecond)
		shortExpiryVerifier := auth.NewVerifier(shortExpiryTokenManager, &mockSigner{})

		// Temporarily replace the executor's auth verifier
		originalVerifier := executor.authVerifier
		executor.authVerifier = shortExpiryVerifier
		defer func() {
			executor.authVerifier = originalVerifier
		}()

		conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		managementClient := executorV1.NewExecutorManagementServiceClient(conn)

		// Get challenge token
		tokenResp, err := managementClient.GetChallengeToken(ctx, &executorV1.GetChallengeTokenRequest{
			OperatorAddress: testConfig.Operator.Address,
		})
		require.NoError(t, err)

		// Wait for nonce to expire
		time.Sleep(150 * time.Millisecond)

		// Try to use expired nonce
		req := &executorV1.ListPerformersRequest{
			AvsAddress: "0xAVS123",
		}
		requestBytes, err := auth.GetRequestWithoutAuth(req)
		require.NoError(t, err)

		signedMessage := auth.ConstructSignedMessage(tokenResp.ChallengeToken, "ListPerformers", requestBytes)
		signature, err := mockSigners.ECDSASigner.SignMessage(signedMessage)
		require.NoError(t, err)

		req.Auth = &executorV1.AuthSignature{
			ChallengeToken: tokenResp.ChallengeToken,
			Signature: signature,
		}

		_, err = managementClient.ListPerformers(ctx, req)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "invalid nonce")
	})

	// Cleanup
	rpcServerInstance.Stop()
}

// TestAuthenticationClientHelpers tests the authenticated client helper methods
func TestAuthenticationClientHelpers(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	// Setup test server
	testConfig := &executorConfig.ExecutorConfig{
		Debug:                false,
		GrpcPort:             0,
		PerformerNetworkName: "test-network",
		Operator: &config.OperatorConfig{
			Address:            "0xTestOperator456",
			OperatorPrivateKey: "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		},
		AvsPerformers: []*executorConfig.AvsPerformerConfig{},
		L1Chain: &executorConfig.Chain{
			RpcUrl:  "http://localhost:8545",
			ChainId: 1,
		},
	}

	mockSigners := signer.Signers{
		ECDSASigner: &mockSigner{},
		BLSSigner:   nil,
	}

	rpcServerInstance, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: 0,
	}, logger)
	require.NoError(t, err)

	executor := NewExecutor(
		testConfig,
		rpcServerInstance,
		logger,
		mockSigners,
		&mockPeeringFetcher{},
		&mockContractCaller{},
	)

	err = executor.Initialize(ctx)
	require.NoError(t, err)

	go func() {
		_ = rpcServerInstance.Start(ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	listener := rpcServerInstance.GetListener()
	require.NotNil(t, listener)
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	serverAddr := "localhost:" + port

	t.Run("GetClient_Returns_Underlying_Client", func(t *testing.T) {
		authClient, err := executorClient.NewAuthenticatedExecutorClient(
			serverAddr,
			testConfig.Operator.Address,
			mockSigners.ECDSASigner,
			true,
		)
		require.NoError(t, err)

		// Get underlying client
		underlyingClient := authClient.GetClient()
		assert.NotNil(t, underlyingClient)

		// Verify it's a valid executor service client
		_, ok := underlyingClient.(executorV1.ExecutorServiceClient)
		assert.True(t, ok)
	})

	t.Run("SubmitTask_Passthrough", func(t *testing.T) {
		authClient, err := executorClient.NewAuthenticatedExecutorClient(
			serverAddr,
			testConfig.Operator.Address,
			mockSigners.ECDSASigner,
			true,
		)
		require.NoError(t, err)

		// SubmitTask should pass through without authentication
		_, err = authClient.SubmitTask(ctx, &executorV1.TaskSubmission{
			TaskId:             "task-456",
			AggregatorAddress:  "0xAggregator",
			AvsAddress:         "0xAVS456",
			Payload:            []byte("test-payload-2"),
			Signature:          []byte("task-signature-2"),
			OperatorSetId:      1,
			ReferenceTimestamp: 12345,
		})

		// Will fail with AVS performer error, not auth error
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "AVS performer not found")
	})

	// Cleanup
	rpcServerInstance.Stop()
}

// BenchmarkAuthentication benchmarks the authentication overhead
func BenchmarkAuthentication(b *testing.B) {
	ctx := context.Background()
	logger := zaptest.NewLogger(b)

	testConfig := &executorConfig.ExecutorConfig{
		Debug:                false,
		GrpcPort:             0,
		PerformerNetworkName: "test-network",
		Operator: &config.OperatorConfig{
			Address:            "0xBenchmarkOperator",
			OperatorPrivateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		AvsPerformers: []*executorConfig.AvsPerformerConfig{},
		L1Chain: &executorConfig.Chain{
			RpcUrl:  "http://localhost:8545",
			ChainId: 1,
		},
	}

	mockSigners := signer.Signers{
		ECDSASigner: &mockSigner{},
		BLSSigner:   nil,
	}

	rpcServerInstance, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: 0,
	}, logger)
	require.NoError(b, err)

	executor := NewExecutor(
		testConfig,
		rpcServerInstance,
		logger,
		mockSigners,
		&mockPeeringFetcher{},
		&mockContractCaller{},
	)

	err = executor.Initialize(ctx)
	require.NoError(b, err)

	go func() {
		_ = rpcServerInstance.Start(ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	listener := rpcServerInstance.GetListener()
	require.NotNil(b, listener)
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(b, err)
	serverAddr := "localhost:" + port

	authClient, err := executorClient.NewAuthenticatedExecutorClient(
		serverAddr,
		testConfig.Operator.Address,
		mockSigners.ECDSASigner,
		true,
	)
	require.NoError(b, err)

	b.ResetTimer()

	// Benchmark authenticated requests
	for i := 0; i < b.N; i++ {
		_, _ = authClient.ListPerformers(ctx, &executorV1.ListPerformersRequest{
			AvsAddress: "0xBenchmark",
		})
	}

	b.StopTimer()
	rpcServerInstance.Stop()
}