package serverPerformer

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/localPeeringDataFetcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// isDockerAvailable checks if Docker is available and running using containerManager
func isDockerAvailable(t *testing.T) bool {
	logger := zaptest.NewLogger(t)
	cm, err := containerManager.NewDockerContainerManager(nil, logger)
	if err != nil {
		t.Logf("Failed to create container manager: %v", err)
		return false
	}
	defer cm.Shutdown(context.Background())

	// Try to list containers as a way to check if Docker daemon is running
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// We can check by trying to inspect a non-existent container
	// If Docker is not running, this will fail with a connection error
	_, err = cm.Inspect(ctx, "non-existent-container-id")
	if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
		t.Logf("Docker daemon not running: %v", err)
		return false
	}

	return true
}

// createTestPeeringFetcher creates a LocalPeeringDataFetcher configured for testing
func createTestPeeringFetcher(t *testing.T, logger *zap.Logger) peering.IPeeringDataFetcher {
	// Create mock BN254 public keys for executor and aggregator
	_, executorPubKey, err := bn254.GenerateKeyPair()
	require.NoError(t, err)

	_, aggregatorPubKey, err := bn254.GenerateKeyPair()
	require.NoError(t, err)

	return localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		OperatorPeers: []*peering.OperatorPeerInfo{
			{
				OperatorAddress: "0x6B58f6762689DF33fe8fa3FC40Fb5a3089D3a8cc",
				OperatorSets: []*peering.OperatorSet{
					{
						OperatorSetID:  1,
						NetworkAddress: "localhost:9999",
						PublicKey:      executorPubKey,
					},
				},
			},
		},
		AggregatorPeers: []*peering.OperatorPeerInfo{
			{
				OperatorAddress: "0x7B58f6762689DF33fe8fa3FC40Fb5a3089D3a8dd",
				OperatorSets: []*peering.OperatorSet{
					{
						OperatorSetID:  0, // Aggregator operator set
						NetworkAddress: "localhost:9998",
						PublicKey:      aggregatorPubKey,
					},
				},
			},
		},
	}, logger)
}

func TestAvsPerformerServer_AutomaticContainerPromotion_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !isDockerAvailable(t) {
		t.Fatal("Docker is not available, failing integration tests")
		return
	}

	logger := zaptest.NewLogger(t)
	// Create mock peering fetcher
	peeringFetcher := createTestPeeringFetcher(t, logger)

	// Create a short health check interval for tests
	testHealthCheckInterval := 1 * time.Second

	t.Run("automatic promotion when next container becomes healthy", func(t *testing.T) {
		// Use simple static AVS address for this test
		avsAddress := "0x1234567890abcdef"

		// Create AVS performer config
		config := &avsPerformer.AvsPerformerConfig{
			AvsAddress:           avsAddress,
			ProcessType:          avsPerformer.AvsProcessTypeServer,
			Image:                avsPerformer.PerformerImage{Repository: "hello-performer", Tag: "latest"},
			WorkerCount:          1,
			PerformerNetworkName: "",
			SigningCurve:         "bn254",
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Create container manager
		containerMgr, err := containerManager.NewDockerContainerManager(containerManager.DefaultContainerManagerConfig(), logger)
		require.NoError(t, err)

		// Create AVS performer server with test health check config
		server := NewAvsPerformerServerWithHealthCheckInterval(config, peeringFetcher, logger, containerMgr, testHealthCheckInterval)

		// Initialize the server (creates initial currentContainer)
		err = server.Initialize(ctx)
		require.NoError(t, err)
		require.NotNil(t, server.currentContainer)
		require.NotNil(t, server.currentContainer.Info)

		// Store the original current container ID for comparison
		originalCurrentContainerID := server.currentContainer.Info.ID

		// Deploy a new container to nextContainer slot
		deployImage := avsPerformer.PerformerImage{
			Repository: "hello-performer",
			Tag:        "latest",
		}

		statusChan, err := server.DeployContainer(ctx, avsAddress, deployImage)
		require.NoError(t, err)
		require.NotNil(t, statusChan)
		require.NotNil(t, server.nextContainer)
		require.NotNil(t, server.nextContainer.Info)

		// Store the next container ID
		nextContainerID := server.nextContainer.Info.ID
		assert.NotEqual(t, originalCurrentContainerID, nextContainerID)

		// Wait for the containers to fully initialize their gRPC services
		// hello-performer takes time to start up and be ready for health checks
		t.Logf("Waiting for containers to fully initialize...")
		time.Sleep(5 * time.Second)

		// Now simulate a healthy event - this should trigger promotion since app health should be ready
		healthyEvent := containerManager.ContainerEvent{
			ContainerID: nextContainerID,
			Type:        containerManager.EventHealthy,
			Timestamp:   time.Now(),
			Message:     "Container health check passed",
			State: containerManager.ContainerState{
				Status:       "running",
				ExitCode:     0,
				StartedAt:    time.Now(),
				RestartCount: 0,
				OOMKilled:    false,
				Error:        "",
				Restarting:   false,
			},
		}

		// Call the event handler directly to simulate receiving a healthy event
		server.handleContainerEvent(ctx, healthyEvent)

		// Wait for the periodic health check to run and send status to channel
		// With our test config, the periodic health check runs every 1 second
		t.Logf("Waiting for container to become application-healthy...")

		// Monitor the status channel for health events
		select {
		case statusEvent := <-statusChan:
			assert.Equal(t, avsPerformer.PerformerHealthy, statusEvent.Status, "Should receive healthy status")
			assert.Equal(t, nextContainerID, statusEvent.ContainerID, "Status should be for next container")
			t.Logf("Received healthy status event for container %s", statusEvent.ContainerID)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for healthy status event")
		}

		// Now manually promote the container (no longer automatic)
		err = server.PromoteContainer(ctx)
		require.NoError(t, err, "Promotion should succeed for healthy container")

		// Verify that the next container was promoted to current container
		require.NotNil(t, server.currentContainer)
		require.NotNil(t, server.currentContainer.Info)
		assert.Equal(t, nextContainerID, server.currentContainer.Info.ID, "Next container should have been promoted to current")

		// Verify that nextContainer slot is now empty
		assert.Nil(t, server.nextContainer, "Next container slot should be empty after promotion")

		// Verify that the original current container is no longer the current one
		assert.NotEqual(t, originalCurrentContainerID, server.currentContainer.Info.ID, "Original current container should have been replaced")
	})

	t.Run("no promotion when healthy event is for current container", func(t *testing.T) {
		// Use simple static AVS address for this test
		avsAddress := "0x1234567890abcdef2222222222222222"

		config := &avsPerformer.AvsPerformerConfig{
			AvsAddress:           avsAddress,
			ProcessType:          avsPerformer.AvsProcessTypeServer,
			Image:                avsPerformer.PerformerImage{Repository: "hello-performer", Tag: "latest"},
			WorkerCount:          1,
			PerformerNetworkName: "",
			SigningCurve:         "bn254",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Create container manager
		containerMgr, err := containerManager.NewDockerContainerManager(containerManager.DefaultContainerManagerConfig(), logger)
		require.NoError(t, err)

		// Create AVS performer server with test health check config
		server := NewAvsPerformerServerWithHealthCheckInterval(config, peeringFetcher, logger, containerMgr, testHealthCheckInterval)

		err = server.Initialize(ctx)
		require.NoError(t, err)
		require.NotNil(t, server.currentContainer)

		currentContainerID := server.currentContainer.Info.ID

		// Simulate a healthy event for the current container (not next container)
		healthyEvent := containerManager.ContainerEvent{
			ContainerID: currentContainerID,
			Type:        containerManager.EventHealthy,
			Timestamp:   time.Now(),
			Message:     "Container health check passed",
			State: containerManager.ContainerState{
				Status:    "running",
				ExitCode:  0,
				StartedAt: time.Now(),
			},
		}

		// Call the event handler
		server.handleContainerEvent(ctx, healthyEvent)

		// Wait a moment
		time.Sleep(1 * time.Second)

		// Verify that nothing changed - current container should remain the same
		require.NotNil(t, server.currentContainer)
		assert.Equal(t, currentContainerID, server.currentContainer.Info.ID, "Current container should remain unchanged")
		assert.Nil(t, server.nextContainer, "Next container should still be nil")
	})

	t.Run("no promotion when no next container exists", func(t *testing.T) {
		// Create container manager for this test scenario
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Create container manager
		containerMgr, err := containerManager.NewDockerContainerManager(containerManager.DefaultContainerManagerConfig(), logger)
		require.NoError(t, err)

		// Use simple static AVS address for this test
		avsAddress := "0x1234567890abcdef3333333333333333"

		config := &avsPerformer.AvsPerformerConfig{
			AvsAddress:           avsAddress,
			ProcessType:          avsPerformer.AvsProcessTypeServer,
			Image:                avsPerformer.PerformerImage{Repository: "hello-performer", Tag: "latest"},
			WorkerCount:          1,
			PerformerNetworkName: "",
			SigningCurve:         "bn254",
		}

		// Create AVS performer server with test health check config
		server := NewAvsPerformerServerWithHealthCheckInterval(config, peeringFetcher, logger, containerMgr, testHealthCheckInterval)

		// Reset to clean state
		err = server.Initialize(ctx)
		require.NoError(t, err)
		require.NotNil(t, server.currentContainer)

		currentContainerID := server.currentContainer.Info.ID

		// Ensure no next container exists
		server.nextContainer = nil

		// Simulate a healthy event for a random container ID
		healthyEvent := containerManager.ContainerEvent{
			ContainerID: "random-container-id",
			Type:        containerManager.EventHealthy,
			Timestamp:   time.Now(),
			Message:     "Container health check passed",
			State: containerManager.ContainerState{
				Status:    "running",
				ExitCode:  0,
				StartedAt: time.Now(),
			},
		}

		// Call the event handler
		server.handleContainerEvent(ctx, healthyEvent)

		// Wait a moment
		time.Sleep(1 * time.Second)

		// Verify that nothing changed
		require.NotNil(t, server.currentContainer)
		assert.Equal(t, currentContainerID, server.currentContainer.Info.ID, "Current container should remain unchanged")
		assert.Nil(t, server.nextContainer, "Next container should still be nil")
	})
}

func TestAvsPerformerServer_ContainerDeploymentFlow_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !isDockerAvailable(t) {
		t.Skip("Docker is not available, skipping integration tests")
	}

	logger := zaptest.NewLogger(t)

	// Create mock peering fetcher
	peeringFetcher := createTestPeeringFetcher(t, logger)

	// Create a short health check interval for tests
	testHealthCheckInterval := 1 * time.Second

	t.Run("complete blue-green deployment flow", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Create container manager for this test scenario
		containerMgr, err := containerManager.NewDockerContainerManager(containerManager.DefaultContainerManagerConfig(), logger)
		require.NoError(t, err)

		// Use simple static AVS address for this test
		avsAddress := "0x1234567890abcdef4444444444444444"

		// Create AVS performer config
		config := &avsPerformer.AvsPerformerConfig{
			AvsAddress:           avsAddress,
			ProcessType:          avsPerformer.AvsProcessTypeServer,
			Image:                avsPerformer.PerformerImage{Repository: "hello-performer", Tag: "latest"},
			WorkerCount:          1,
			PerformerNetworkName: "",
			SigningCurve:         "bn254",
		}

		// Create AVS performer server with test health check config
		server := NewAvsPerformerServerWithHealthCheckInterval(config, peeringFetcher, logger, containerMgr, testHealthCheckInterval)

		// Step 1: Initialize server with initial container
		err = server.Initialize(ctx)
		require.NoError(t, err)
		require.NotNil(t, server.currentContainer)
		require.Nil(t, server.nextContainer)

		originalContainerID := server.currentContainer.Info.ID

		// Step 2: Deploy new version to next container
		deployImage := avsPerformer.PerformerImage{
			Repository: "hello-performer",
			Tag:        "latest",
		}

		statusChan, err := server.DeployContainer(ctx, avsAddress, deployImage)
		require.NoError(t, err)
		require.NotNil(t, statusChan)
		require.NotNil(t, server.nextContainer)

		nextContainerID := server.nextContainer.Info.ID

		// Step 3: Verify deployment state
		assert.Equal(t, originalContainerID, server.currentContainer.Info.ID, "Current container should be unchanged")
		assert.Equal(t, nextContainerID, server.nextContainer.Info.ID, "Next container should be deployed")
		assert.NotEqual(t, originalContainerID, nextContainerID, "Containers should be different")

		// Wait for containers to be ready
		time.Sleep(5 * time.Second)

		// Step 4: Simulate automatic promotion via healthy event
		healthyEvent := containerManager.ContainerEvent{
			ContainerID: nextContainerID,
			Type:        containerManager.EventHealthy,
			Timestamp:   time.Now(),
			Message:     "Container health check passed",
			State: containerManager.ContainerState{
				Status:       "running",
				ExitCode:     0,
				StartedAt:    time.Now(),
				RestartCount: 0,
				OOMKilled:    false,
				Error:        "",
				Restarting:   false,
			},
		}

		server.handleContainerEvent(ctx, healthyEvent)

		// Wait for the container to become application-healthy
		t.Logf("Waiting for container to become application-healthy...")

		// Monitor the status channel for health events
		select {
		case statusEvent := <-statusChan:
			assert.Equal(t, avsPerformer.PerformerHealthy, statusEvent.Status, "Should receive healthy status")
			assert.Equal(t, nextContainerID, statusEvent.ContainerID, "Status should be for next container")
			t.Logf("Received healthy status event for container %s", statusEvent.ContainerID)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for healthy status event")
		}

		// Manually promote the container
		err = server.PromoteContainer(ctx)
		require.NoError(t, err, "Promotion should succeed for healthy container")

		// Step 5: Verify promotion completed
		require.NotNil(t, server.currentContainer)
		assert.Equal(t, nextContainerID, server.currentContainer.Info.ID, "Next container should be promoted to current")
		assert.Nil(t, server.nextContainer, "Next container slot should be empty")

		// Step 6: Verify the new container is operational
		assert.NotNil(t, server.currentContainer.Client, "Performer client should be available")
		assert.Equal(t, "running", server.currentContainer.Info.Status, "Container should be running")
	})
}
