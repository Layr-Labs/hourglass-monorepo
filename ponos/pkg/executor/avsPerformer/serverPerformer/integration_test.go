package serverPerformer

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// isDockerAvailable checks if Docker is available and running
func isDockerAvailable(t *testing.T) bool {
	// Check if docker command exists
	if _, err := exec.LookPath("docker"); err != nil {
		t.Logf("Docker command not found: %v", err)
		return false
	}

	// Check if Docker daemon is running
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		t.Logf("Docker daemon not running: %v", err)
		return false
	}

	return true
}

// isHelloPerformerAvailable checks if the hello-performer container image is available
func isHelloPerformerAvailable(t *testing.T) bool {
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}", "hello-performer:latest")
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Failed to check for hello-performer image: %v", err)
		return false
	}

	if !strings.Contains(string(output), "hello-performer:latest") {
		t.Logf("hello-performer image not found. Run 'make build/test-performer-container' to build it.")
		return false
	}

	return true
}

// MockPeeringFetcher implements peering.IPeeringDataFetcher for testing
type MockPeeringFetcher struct{}

func (m *MockPeeringFetcher) ListExecutorOperators(ctx context.Context, avsAddress string) ([]*peering.OperatorPeerInfo, error) {
	// Create a mock BN254 public key
	_, mockPubKey, err := bn254.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	// Return mock peering data for tests
	return []*peering.OperatorPeerInfo{
		{
			OperatorAddress: "0x6B58f6762689DF33fe8fa3FC40Fb5a3089D3a8cc",
			OperatorSets: []*peering.OperatorSet{
				{
					OperatorSetID:  1,
					NetworkAddress: "localhost:9999",
					PublicKey:      mockPubKey,
				},
			},
		},
	}, nil
}

func (m *MockPeeringFetcher) ListAggregatorOperators(ctx context.Context, avsAddress string) ([]*peering.OperatorPeerInfo, error) {
	// Create a mock BN254 public key
	_, mockPubKey, err := bn254.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	// Return mock peering data for tests
	return []*peering.OperatorPeerInfo{
		{
			OperatorAddress: "0x7B58f6762689DF33fe8fa3FC40Fb5a3089D3a8dd",
			OperatorSets: []*peering.OperatorSet{
				{
					OperatorSetID:  0, // Aggregator operator set
					NetworkAddress: "localhost:9998",
					PublicKey:      mockPubKey,
				},
			},
		},
	}, nil
}

func TestAvsPerformerServer_AutomaticContainerPromotion_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !isDockerAvailable(t) {
		t.Fatal("Docker is not available, failing integration tests")
		return
	}

	if !isHelloPerformerAvailable(t) {
		t.Skip("hello-performer image not available, skipping integration tests. Run 'make build/test-performer-container' to build it.")
	}

	logger := zaptest.NewLogger(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create container manager
	containerMgr, err := containerManager.NewDefaultDockerContainerManager(logger)
	require.NoError(t, err)
	defer func() { _ = containerMgr.Shutdown(ctx) }()

	// Create mock peering fetcher
	peeringFetcher := &MockPeeringFetcher{}

	t.Run("automatic promotion when next container becomes healthy", func(t *testing.T) {
		// Use simple static AVS address for this test
		avsAddress := "0x1234567890abcdef"

		//// Ensure network cleanup no matter what
		//defer func() {
		//	if err := containerMgr.RemoveNetwork(ctx, networkName); err != nil {
		//		t.Logf("Network cleanup completed (expected if network was already cleaned): %v", err)
		//	}
		//}()

		// Create AVS performer config
		config := &avsPerformer.AvsPerformerConfig{
			AvsAddress:           avsAddress,
			ProcessType:          avsPerformer.AvsProcessTypeServer,
			Image:                avsPerformer.PerformerImage{Repository: "hello-performer", Tag: "latest"},
			WorkerCount:          1,
			PerformerNetworkName: "",
			SigningCurve:         "bn254",
		}

		// Create AVS performer server
		server := NewAvsPerformerServer(config, peeringFetcher, logger, containerMgr)
		defer func() { _ = server.Shutdown() }()

		// Initialize the server (creates initial currentContainer)
		err := server.Initialize(ctx)
		require.NoError(t, err)
		require.NotNil(t, server.currentContainer)
		require.NotNil(t, server.currentContainer.Info)

		// Store the original current container ID for comparison
		originalCurrentContainerID := server.currentContainer.Info.ID

		// Deploy a new container to nextContainer slot
		nextContainerMgr, err := containerManager.NewDefaultDockerContainerManager(logger)
		require.NoError(t, err)
		defer func() { _ = nextContainerMgr.Shutdown(ctx) }()

		deployConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:           avsAddress,
			ProcessType:          avsPerformer.AvsProcessTypeServer,
			Image:                avsPerformer.PerformerImage{Repository: "hello-performer", Tag: "latest"},
			WorkerCount:          1,
			PerformerNetworkName: "",
			SigningCurve:         "bn254",
		}

		err = server.DeployContainer(ctx, avsAddress, deployConfig, nextContainerMgr)
		require.NoError(t, err)
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

		// Wait for the promotion to complete
		time.Sleep(5 * time.Second)

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

		server := NewAvsPerformerServer(config, peeringFetcher, logger, containerMgr)
		defer func() { _ = server.Shutdown() }()

		// Reset to clean state
		err := server.Initialize(ctx)
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

		server := NewAvsPerformerServer(config, peeringFetcher, logger, containerMgr)
		defer func() { _ = server.Shutdown() }()

		// Reset to clean state
		err := server.Initialize(ctx)
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

	if !isHelloPerformerAvailable(t) {
		t.Skip("hello-performer image not available, skipping integration tests. Run 'make build/test-performer-container' to build it.")
	}

	logger := zaptest.NewLogger(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Create container manager
	containerMgr, err := containerManager.NewDefaultDockerContainerManager(logger)
	require.NoError(t, err)
	defer func() { _ = containerMgr.Shutdown(ctx) }()

	// Create mock peering fetcher
	peeringFetcher := &MockPeeringFetcher{}

	t.Run("complete blue-green deployment flow", func(t *testing.T) {
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

		// Create AVS performer server
		server := NewAvsPerformerServer(config, peeringFetcher, logger, containerMgr)
		defer func() { _ = server.Shutdown() }()

		// Step 1: Initialize server with initial container
		err := server.Initialize(ctx)
		require.NoError(t, err)
		require.NotNil(t, server.currentContainer)
		require.Nil(t, server.nextContainer)

		originalContainerID := server.currentContainer.Info.ID

		// Step 2: Deploy new version to next container
		nextContainerMgr, err := containerManager.NewDefaultDockerContainerManager(logger)
		require.NoError(t, err)
		defer func() { _ = nextContainerMgr.Shutdown(ctx) }()

		deployConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:           avsAddress,
			ProcessType:          avsPerformer.AvsProcessTypeServer,
			Image:                avsPerformer.PerformerImage{Repository: "hello-performer", Tag: "latest"},
			WorkerCount:          1,
			PerformerNetworkName: "",
			SigningCurve:         "bn254",
		}

		err = server.DeployContainer(ctx, avsAddress, deployConfig, nextContainerMgr)
		require.NoError(t, err)
		require.NotNil(t, server.nextContainer)

		nextContainerID := server.nextContainer.Info.ID

		// Step 3: Verify deployment state
		assert.Equal(t, originalContainerID, server.currentContainer.Info.ID, "Current container should be unchanged")
		assert.Equal(t, nextContainerID, server.nextContainer.Info.ID, "Next container should be deployed")
		assert.NotEqual(t, originalContainerID, nextContainerID, "Containers should be different")

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

		// Wait for promotion to complete
		time.Sleep(3 * time.Second)

		// Step 5: Verify promotion completed
		require.NotNil(t, server.currentContainer)
		assert.Equal(t, nextContainerID, server.currentContainer.Info.ID, "Next container should be promoted to current")
		assert.Nil(t, server.nextContainer, "Next container slot should be empty")

		// Step 6: Verify the new container is operational
		assert.NotNil(t, server.currentContainer.Client, "Performer client should be available")
		assert.NotNil(t, server.currentContainer.Manager, "Container manager should be available")
		assert.Equal(t, "running", server.currentContainer.Info.Status, "Container should be running")
	})
}
