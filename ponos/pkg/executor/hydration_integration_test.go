package executor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/badger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestContainerHydrationIntegration tests that containers can be rehydrated from persistent storage
func TestContainerHydrationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(180*time.Second))
	defer cancel()

	// Setup logger
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	// Get project root and chain config
	root := testUtils.GetProjectRootPath()
	chainConfig, err := testUtils.ReadChainConfig(root)
	require.NoError(t, err)

	// Create temporary directory for Badger storage that will be shared between executors
	tmpDir, err := os.MkdirTemp("", "hydration-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Logf("Using temporary storage directory: %s", tmpDir)

	// Base executor config with persistent storage
	baseExecutorConfig := &executorConfig.ExecutorConfig{
		Debug:                false,
		GrpcPort:             9092,
		PerformerNetworkName: "",
		Operator: &config.OperatorConfig{
			Address: chainConfig.ExecOperatorAccountAddress,
			SigningKeys: config.SigningKeys{
				ECDSA: &config.ECDSAKeyConfig{
					PrivateKey: chainConfig.ExecOperatorAccountPk,
				},
			},
		},
		AvsPerformers: []*executorConfig.AvsPerformerConfig{
			{
				AvsAddress:     chainConfig.AVSAccountAddress,
				ProcessType:    "server",
				DeploymentMode: executorConfig.DeploymentModeDocker,
				Image: &executorConfig.PerformerImage{
					Repository: "sleepy-hello-performer",
					Tag:        "latest",
				},
				Envs: []config.AVSPerformerEnv{
					{
						Name:  "TEST_ENV",
						Value: "hydration_test",
					},
					{
						Name:         "DYNAMIC_ENV",
						ValueFromEnv: "PATH",
					},
				},
			},
		},
		Storage: &executorConfig.StorageConfig{
			Type: "badger",
			BadgerConfig: &executorConfig.BadgerConfig{
				Dir: tmpDir, // This directory will persist between executor instances
			},
		},
	}

	// Parse keys for executor
	_, _, execGenericSigningKey, err := testUtils.ParseKeysFromConfig(baseExecutorConfig.Operator, config.CurveTypeECDSA)
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

	signers := signer.Signers{
		ECDSASigner: execSigner,
	}

	// Variable to store the performer ID for verification
	var deployedPerformerID string
	var deployedContainerID string

	// Test Phase 1: Deploy container with first executor
	t.Run("Phase1_DeployWithFirstExecutor", func(t *testing.T) {
		// Create Badger store for first executor
		store1, err := badger.NewBadgerExecutorStore(baseExecutorConfig.Storage.BadgerConfig)
		require.NoError(t, err)

		// Create first executor
		exec1, err := NewExecutorWithRpcServers(
			baseExecutorConfig.GrpcPort,
			baseExecutorConfig.GrpcPort,
			baseExecutorConfig,
			l,
			signers,
			pdf,
			l1CC,
			store1,
		)
		require.NoError(t, err)

		// Initialize executor (this will deploy containers for configured AVS performers)
		err = exec1.Initialize(ctx)
		require.NoError(t, err)

		// Start executor in background
		exec1Ctx, exec1Cancel := context.WithCancel(ctx)
		go func() {
			if err := exec1.Run(exec1Ctx); err != nil {
				t.Logf("Executor 1 stopped: %v", err)
			}
		}()

		// Wait for server to start
		time.Sleep(3 * time.Second)

		// Connect to executor and verify deployment
		serverAddr := fmt.Sprintf("localhost:%d", baseExecutorConfig.GrpcPort)
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		managementClient := executorV1.NewExecutorManagementServiceClient(conn)

		// List performers to verify deployment
		listResp, err := managementClient.ListPerformers(ctx, &executorV1.ListPerformersRequest{})
		require.NoError(t, err)
		require.Len(t, listResp.Performers, 1, "Should have one performer deployed")

		performer := listResp.Performers[0]
		assert.Equal(t, strings.ToLower(chainConfig.AVSAccountAddress), strings.ToLower(performer.AvsAddress))
		assert.Equal(t, "sleepy-hello-performer", performer.ArtifactRegistry)
		assert.Equal(t, "latest", performer.ArtifactTag)
		assert.Equal(t, "InService", performer.Status)
		assert.True(t, performer.ResourceHealthy)
		assert.True(t, performer.ApplicationHealthy)

		// Store performer ID and container ID for later verification
		deployedPerformerID = performer.PerformerId
		deployedContainerID = performer.ContainerId

		t.Logf("Deployed performer ID: %s", deployedPerformerID)
		t.Logf("Deployed container ID: %s", deployedContainerID)

		// Verify data was persisted to storage
		states, err := store1.ListPerformerStates(ctx)
		require.NoError(t, err)
		require.Len(t, states, 1, "Should have one performer state persisted")

		state := states[0]
		assert.Equal(t, deployedPerformerID, state.PerformerId)
		assert.Equal(t, deployedContainerID, state.ContainerId)
		assert.Equal(t, "running", state.Status)
		assert.NotEmpty(t, state.ContainerEndpoint)
		assert.NotEmpty(t, state.ContainerHostname)
		assert.Len(t, state.EnvironmentVars, 2, "Should have 2 environment variables")

		// Gracefully shutdown first executor
		exec1Cancel()
		time.Sleep(2 * time.Second)

		// Close the store to ensure all data is flushed to disk
		err = store1.Close()
		require.NoError(t, err)

		t.Log("First executor shutdown complete, state persisted to disk")
	})

	// Test Phase 2: Rehydrate container with second executor
	t.Run("Phase2_RehydrateWithSecondExecutor", func(t *testing.T) {
		// Create new Badger store instance pointing to SAME directory
		store2, err := badger.NewBadgerExecutorStore(baseExecutorConfig.Storage.BadgerConfig)
		require.NoError(t, err)
		defer store2.Close()

		// Verify persisted state is accessible
		states, err := store2.ListPerformerStates(ctx)
		require.NoError(t, err)
		require.Len(t, states, 1, "Should have one performer state from previous executor")

		t.Logf("Found persisted performer state: %s", states[0].PerformerId)

		// Create second executor with same storage
		// Use a different port to avoid conflicts
		secondExecutorConfig := *baseExecutorConfig
		secondExecutorConfig.GrpcPort = 9093

		exec2, err := NewExecutorWithRpcServers(
			secondExecutorConfig.GrpcPort,
			secondExecutorConfig.GrpcPort,
			&secondExecutorConfig,
			l,
			signers,
			pdf,
			l1CC,
			store2,
		)
		require.NoError(t, err)

		// Initialize executor - this should trigger rehydration
		err = exec2.Initialize(ctx)
		require.NoError(t, err)

		// Start executor in background
		exec2Ctx, exec2Cancel := context.WithCancel(ctx)
		defer exec2Cancel()
		go func() {
			if err := exec2.Run(exec2Ctx); err != nil {
				t.Logf("Executor 2 stopped: %v", err)
			}
		}()

		// Wait for server to start and rehydration to complete
		time.Sleep(3 * time.Second)

		// Connect to second executor
		serverAddr := fmt.Sprintf("localhost:%d", secondExecutorConfig.GrpcPort)
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		managementClient := executorV1.NewExecutorManagementServiceClient(conn)

		// Wait for the rehydrated container to become healthy (up to 30 seconds)
		var performer *executorV1.Performer
		for i := 0; i < 10; i++ {
			listResp, err := managementClient.ListPerformers(ctx, &executorV1.ListPerformersRequest{})
			require.NoError(t, err)
			require.Len(t, listResp.Performers, 1, "Should have one performer rehydrated")

			performer = listResp.Performers[0]
			if performer.ApplicationHealthy {
				break
			}
			t.Logf("Waiting for rehydrated container to become healthy (attempt %d/10, status: %s, app healthy: %v)",
				i+1, performer.Status, performer.ApplicationHealthy)
			time.Sleep(3 * time.Second)
		}

		// Verify the rehydrated performer
		assert.Equal(t, deployedPerformerID, performer.PerformerId, "Should have same performer ID")
		assert.Equal(t, deployedContainerID, performer.ContainerId, "Should have same container ID")
		assert.Equal(t, strings.ToLower(chainConfig.AVSAccountAddress), strings.ToLower(performer.AvsAddress))
		assert.Equal(t, "sleepy-hello-performer", performer.ArtifactRegistry)
		assert.Equal(t, "running", performer.Status)
		assert.True(t, performer.ResourceHealthy, "Rehydrated container should be healthy")
		assert.True(t, performer.ApplicationHealthy, "Rehydrated application should be healthy")

		t.Log("Successfully rehydrated performer from persistent storage")
	})

	// Test Phase 4: Verify cleanup of non-running containers
	t.Run("Phase3_CleanupNonRunningContainers", func(t *testing.T) {
		// This test verifies that if a container is stopped between executor restarts, the state is properly cleaned up

		// First, we need to stop the container manually
		store3, err := badger.NewBadgerExecutorStore(baseExecutorConfig.Storage.BadgerConfig)
		require.NoError(t, err)
		defer store3.Close()

		// Add a fake performer state with non-existent container
		fakeState := &storage.PerformerState{
			PerformerId:        "fake-performer-123",
			AvsAddress:         "0xfake",
			ContainerId:        "non-existent-container",
			Status:             "running",
			ArtifactRegistry:   "fake-registry",
			ArtifactTag:        "v1.0.0",
			DeploymentMode:     "docker",
			CreatedAt:          time.Now(),
			LastHealthCheck:    time.Now(),
			ContainerHealthy:   true,
			ApplicationHealthy: true,
			NetworkName:        "test-network",
			ContainerEndpoint:  "localhost:8080",
			ContainerHostname:  "fake-host",
			InternalPort:       8080,
		}

		err = store3.SavePerformerState(ctx, fakeState.PerformerId, fakeState)
		require.NoError(t, err)

		// Close store to flush
		err = store3.Close()
		require.NoError(t, err)

		// Create new executor that should clean up the fake state
		store4, err := badger.NewBadgerExecutorStore(baseExecutorConfig.Storage.BadgerConfig)
		require.NoError(t, err)
		defer store4.Close()

		// Verify we have 2 states before cleanup (1 real, 1 fake)
		statesBefore, err := store4.ListPerformerStates(ctx)
		require.NoError(t, err)
		assert.Len(t, statesBefore, 2, "Should have 2 performer states before cleanup")

		// Create third executor - should clean up fake state during rehydration
		thirdExecutorConfig := *baseExecutorConfig
		thirdExecutorConfig.GrpcPort = 9094

		exec3, err := NewExecutorWithRpcServers(
			thirdExecutorConfig.GrpcPort,
			thirdExecutorConfig.GrpcPort,
			&thirdExecutorConfig,
			l,
			signers,
			pdf,
			l1CC,
			store4,
		)
		require.NoError(t, err)

		// Initialize - this should trigger cleanup of non-existent container
		err = exec3.Initialize(ctx)
		require.NoError(t, err)

		// Wait for cleanup to complete
		time.Sleep(2 * time.Second)

		// Verify fake state was cleaned up
		statesAfter, err := store4.ListPerformerStates(ctx)
		require.NoError(t, err)
		assert.Len(t, statesAfter, 1, "Should have only 1 performer state after cleanup")
		assert.Equal(t, deployedPerformerID, statesAfter[0].PerformerId, "Should keep real performer")

		t.Log("Successfully cleaned up state for non-running container")
	})
}

// TestHydrationWithEnvironmentVariables tests that environment variables are properly persisted and restored
func TestHydrationWithEnvironmentVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create temporary directory for storage
	tmpDir, err := os.MkdirTemp("", "hydration-env-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test environment variables with both direct values and host env resolution
	testEnvVars := []config.AVSPerformerEnv{
		{
			Name:  "STATIC_VAR",
			Value: "static_value_123",
		},
		{
			Name:  "ANOTHER_STATIC",
			Value: "another_value_456",
		},
		{
			Name:         "FROM_HOST",
			ValueFromEnv: "HOME", // Will resolve from host environment
		},
	}

	// Create performer state with environment variables
	store1, err := badger.NewBadgerExecutorStore(&executorConfig.BadgerConfig{
		Dir: tmpDir,
	})
	require.NoError(t, err)

	// Convert to storage format
	var envRecords []storage.EnvironmentVarRecord
	for _, env := range testEnvVars {
		envRecords = append(envRecords, storage.EnvironmentVarRecord{
			Name:         env.Name,
			Value:        env.Value,
			ValueFromEnv: env.ValueFromEnv,
		})
	}

	testState := &storage.PerformerState{
		PerformerId:        "env-test-performer",
		AvsAddress:         "0xenvtest",
		ContainerId:        "env-test-container",
		Status:             "running",
		ArtifactRegistry:   "test-registry",
		ArtifactTag:        "v1.0.0",
		DeploymentMode:     "docker",
		CreatedAt:          time.Now(),
		LastHealthCheck:    time.Now(),
		ContainerHealthy:   true,
		ApplicationHealthy: true,
		NetworkName:        "test-network",
		ContainerEndpoint:  "localhost:8080",
		ContainerHostname:  "test-host",
		InternalPort:       8080,
		EnvironmentVars:    envRecords,
	}

	// Save state
	err = store1.SavePerformerState(ctx, testState.PerformerId, testState)
	require.NoError(t, err)

	// Close first store
	err = store1.Close()
	require.NoError(t, err)

	// Open second store and verify environment variables
	store2, err := badger.NewBadgerExecutorStore(&executorConfig.BadgerConfig{
		Dir: tmpDir,
	})
	require.NoError(t, err)
	defer store2.Close()

	// Retrieve state
	retrievedState, err := store2.GetPerformerState(ctx, testState.PerformerId)
	require.NoError(t, err)

	// Verify environment variables were persisted correctly
	require.Len(t, retrievedState.EnvironmentVars, 3, "Should have 3 environment variables")

	// Check each environment variable
	envMap := make(map[string]storage.EnvironmentVarRecord)
	for _, env := range retrievedState.EnvironmentVars {
		envMap[env.Name] = env
	}

	// Verify static variables
	staticVar, ok := envMap["STATIC_VAR"]
	assert.True(t, ok, "Should have STATIC_VAR")
	assert.Equal(t, "static_value_123", staticVar.Value)
	assert.Empty(t, staticVar.ValueFromEnv)

	anotherStatic, ok := envMap["ANOTHER_STATIC"]
	assert.True(t, ok, "Should have ANOTHER_STATIC")
	assert.Equal(t, "another_value_456", anotherStatic.Value)
	assert.Empty(t, anotherStatic.ValueFromEnv)

	// Verify host environment variable reference
	fromHost, ok := envMap["FROM_HOST"]
	assert.True(t, ok, "Should have FROM_HOST")
	assert.Empty(t, fromHost.Value)
	assert.Equal(t, "HOME", fromHost.ValueFromEnv)

	t.Log("Environment variables successfully persisted and restored")
}
