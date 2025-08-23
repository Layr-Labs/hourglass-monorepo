package avsContainerPerformer

import (
	"context"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/avsPerformerClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/localPeeringDataFetcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestOOMKillTriggersUnhealthyState verifies that an OOM kill event sets both
// container and application health to false
func TestOOMKillTriggersUnhealthyState(t *testing.T) {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockContainerMgr := NewMockContainerManager()
	mockHealthClient := NewMockHealthClient()
	mockPerformerClient := NewMockPerformerServiceClient()
	logger := zap.NewNop()

	config := &avsPerformer.AvsPerformerConfig{
		AvsAddress:                     "0xtest",
		ProcessType:                    "server",
		ApplicationHealthCheckInterval: 100 * time.Millisecond,
	}

	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: nil,
	}, logger)

	performer := NewAvsContainerPerformerWithContainerManager(
		config,
		pdf,
		logger,
		mockContainerMgr,
	)

	// Create initial container state
	testContainer := &PerformerContainer{
		performerID: "test-performer-1",
		info: &containerManager.ContainerInfo{
			ID:       "container-123",
			Hostname: "test-host",
			Status:   "running",
		},
		client: &avsPerformerClient.PerformerClient{
			HealthClient:    mockHealthClient,
			PerformerClient: mockPerformerClient,
		},
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: true,
		},
		status: avsPerformer.PerformerResourceStatusInService,
	}

	// Set as current container
	performer.currentContainer.Store(testContainer)

	// Start monitoring events
	eventChan, err := mockContainerMgr.StartLivenessMonitoring(ctx, testContainer.info.ID, nil)
	require.NoError(t, err)
	testContainer.eventChan = eventChan

	// Start the event monitoring in background
	go performer.monitorContainerEvents(ctx, testContainer)

	// Inject OOM kill event
	oomEvent := containerManager.ContainerEvent{
		ContainerID: testContainer.info.ID,
		Type:        containerManager.EventOOMKilled,
		State: containerManager.ContainerState{
			Status:       "oom-killed",
			ExitCode:     137,
			OOMKilled:    true,
			RestartCount: 1,
		},
		Timestamp: time.Now(),
		Message:   "Container killed due to OOM",
	}

	err = mockContainerMgr.InjectEvent(testContainer.info.ID, oomEvent)
	require.NoError(t, err)

	// Allow time for event processing
	time.Sleep(50 * time.Millisecond)

	// Verify state changes
	currentContainer := performer.currentContainer.Load().(*PerformerContainer)
	assert.False(t, currentContainer.performerHealth.ContainerIsHealthy,
		"Container health should be false after OOM kill")
	assert.False(t, currentContainer.performerHealth.ApplicationIsHealthy,
		"Application health should be false after OOM kill")
}

// TestRestartEventRecreatesClient verifies that a restart event resets health state
// and triggers client recreation
func TestRestartEventRecreatesClient(t *testing.T) {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockContainerMgr := NewMockContainerManager()
	mockHealthClient := NewMockHealthClient()
	mockPerformerClient := NewMockPerformerServiceClient()
	logger := zap.NewNop()

	config := &avsPerformer.AvsPerformerConfig{
		AvsAddress:                     "0xtest",
		ProcessType:                    "server",
		ApplicationHealthCheckInterval: 100 * time.Millisecond,
	}

	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: nil,
	}, logger)

	performer := NewAvsContainerPerformerWithContainerManager(
		config,
		pdf,
		logger,
		mockContainerMgr,
	)

	// Create initial container state with some failures
	originalClient := &avsPerformerClient.PerformerClient{
		HealthClient:    mockHealthClient,
		PerformerClient: mockPerformerClient,
	}
	
	testContainer := &PerformerContainer{
		performerID: "test-performer-1",
		info: &containerManager.ContainerInfo{
			ID:       "container-123",
			Hostname: "test-host",
			Status:   "running",
		},
		client: originalClient,
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:                   false,
			ApplicationIsHealthy:                 false,
			ConsecutiveApplicationHealthFailures: 2,
		},
		status: avsPerformer.PerformerResourceStatusInService,
	}

	// Set as current container
	performer.currentContainer.Store(testContainer)

	// Start monitoring events
	eventChan, err := mockContainerMgr.StartLivenessMonitoring(ctx, testContainer.info.ID, nil)
	require.NoError(t, err)
	testContainer.eventChan = eventChan

	// Start the event monitoring in background
	go performer.monitorContainerEvents(ctx, testContainer)

	// Inject restart event
	restartEvent := containerManager.ContainerEvent{
		ContainerID: testContainer.info.ID,
		Type:        containerManager.EventRestarted,
		State: containerManager.ContainerState{
			Status:       "running",
			ExitCode:     0,
			RestartCount: 1,
		},
		Timestamp: time.Now(),
		Message:   "Container restarted successfully",
	}

	err = mockContainerMgr.InjectEvent(testContainer.info.ID, restartEvent)
	require.NoError(t, err)

	// Allow time for event processing and client recreation (includes 2-second sleep in recreatePerformerClientForContainer)
	time.Sleep(2500 * time.Millisecond)

	// Verify state changes
	currentContainer := performer.currentContainer.Load().(*PerformerContainer)
	assert.False(t, currentContainer.performerHealth.ContainerIsHealthy,
		"Container health should be false after restart")
	assert.False(t, currentContainer.performerHealth.ApplicationIsHealthy,
		"Application health should be false after restart")
	assert.Equal(t, 0, currentContainer.performerHealth.ConsecutiveApplicationHealthFailures,
		"Consecutive health failures should be reset to 0")

	// Note: We can't directly verify client recreation in unit test since recreatePerformerClientForContainer
	// creates a new gRPC client which requires actual network connection. In integration tests, this would
	// be verified by checking that the client field has changed.
}

// TestRestartFailureTriggersContainerRecreation verifies that a restart failure
// with "recreation needed" message sets the correct health states
func TestRestartFailureTriggersContainerRecreation(t *testing.T) {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockContainerMgr := NewMockContainerManager()
	mockHealthClient := NewMockHealthClient()
	mockPerformerClient := NewMockPerformerServiceClient()
	logger := zap.NewNop()

	config := &avsPerformer.AvsPerformerConfig{
		AvsAddress:                     "0xtest",
		ProcessType:                    "server",
		ApplicationHealthCheckInterval: 100 * time.Millisecond,
	}

	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: nil,
	}, logger)

	performer := NewAvsContainerPerformerWithContainerManager(
		config,
		pdf,
		logger,
		mockContainerMgr,
	)

	// Create initial container state
	testContainer := &PerformerContainer{
		performerID: "test-performer-1",
		info: &containerManager.ContainerInfo{
			ID:       "container-123",
			Hostname: "test-host",
			Status:   "running",
		},
		client: &avsPerformerClient.PerformerClient{
			HealthClient:    mockHealthClient,
			PerformerClient: mockPerformerClient,
		},
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: true,
		},
		status: avsPerformer.PerformerResourceStatusInService,
		image: avsPerformer.PerformerImage{
			Repository: "test-repo",
			Tag:        "latest",
		},
	}

	// Set as current container
	performer.currentContainer.Store(testContainer)

	// Start monitoring events
	eventChan, err := mockContainerMgr.StartLivenessMonitoring(ctx, testContainer.info.ID, nil)
	require.NoError(t, err)
	testContainer.eventChan = eventChan

	// Start the event monitoring in background
	go performer.monitorContainerEvents(ctx, testContainer)

	// Inject restart failed event WITHOUT "recreation needed" message first
	restartFailedEvent := containerManager.ContainerEvent{
		ContainerID: testContainer.info.ID,
		Type:        containerManager.EventRestartFailed,
		State: containerManager.ContainerState{
			Status:       "stopped",
			ExitCode:     1,
			RestartCount: 3,
			Error:        "restart failed",
		},
		Timestamp: time.Now(),
		Message:   "Container restart failed",
	}

	err = mockContainerMgr.InjectEvent(testContainer.info.ID, restartFailedEvent)
	require.NoError(t, err)

	// Allow time for event processing
	time.Sleep(50 * time.Millisecond)

	// Verify health states are set to false
	currentContainer := performer.currentContainer.Load().(*PerformerContainer)
	assert.False(t, currentContainer.performerHealth.ContainerIsHealthy,
		"Container health should be false after restart failure")
	assert.False(t, currentContainer.performerHealth.ApplicationIsHealthy,
		"Application health should be false after restart failure")

	// Now inject restart failed event WITH "recreation needed" message
	restartFailedWithRecreationEvent := containerManager.ContainerEvent{
		ContainerID: testContainer.info.ID,
		Type:        containerManager.EventRestartFailed,
		State: containerManager.ContainerState{
			Status:       "stopped",
			ExitCode:     1,
			RestartCount: 4,
			Error:        "restart failed",
		},
		Timestamp: time.Now(),
		Message:   "Container restart failed: recreation needed",
	}

	err = mockContainerMgr.InjectEvent(testContainer.info.ID, restartFailedWithRecreationEvent)
	require.NoError(t, err)

	// Allow time for event processing
	time.Sleep(50 * time.Millisecond)

	// Verify health states remain false
	currentContainer = performer.currentContainer.Load().(*PerformerContainer)
	assert.False(t, currentContainer.performerHealth.ContainerIsHealthy,
		"Container health should remain false after restart failure with recreation")
	assert.False(t, currentContainer.performerHealth.ApplicationIsHealthy,
		"Application health should remain false after restart failure with recreation")

	// Note: Actual container recreation (calling createAndStartContainer) would be tested 
	// in integration tests since it requires Docker interactions
}