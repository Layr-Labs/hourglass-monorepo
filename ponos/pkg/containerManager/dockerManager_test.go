package containerManager

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewDockerContainerManager(t *testing.T) {
	tests := []struct {
		name   string
		config *ContainerManagerConfig
	}{
		{
			name:   "nil config should use defaults",
			config: nil,
		},
		{
			name: "custom config should be used",
			config: &ContainerManagerConfig{
				DefaultStartTimeout: 60 * time.Second,
				DefaultStopTimeout:  20 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)

			dcm, err := NewDockerContainerManager(tt.config, logger)
			require.NoError(t, err)
			require.NotNil(t, dcm)

			assert.NotNil(t, dcm.config)
			assert.NotNil(t, dcm.logger)
			assert.NotNil(t, dcm.healthChecks)
			assert.NotNil(t, dcm.livenessMonitors)

			// Verify defaults are set
			if tt.config == nil {
				assert.Equal(t, 30*time.Second, dcm.config.DefaultStartTimeout)
				assert.Equal(t, 10*time.Second, dcm.config.DefaultStopTimeout)
				assert.NotNil(t, dcm.config.DefaultHealthCheckConfig)
				assert.NotNil(t, dcm.config.DefaultLivenessConfig)
			}
		})
	}
}

func TestDockerContainerManager_Configuration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	dcm, err := NewDockerContainerManager(nil, logger)
	require.NoError(t, err)
	defer func() { _ = dcm.Shutdown(context.Background()) }()

	// Test that the container manager has proper configuration
	assert.NotNil(t, dcm.config)
	assert.NotNil(t, dcm.logger)
	assert.NotNil(t, dcm.healthChecks)
	assert.NotNil(t, dcm.livenessMonitors)
	assert.NotNil(t, dcm.client)
}

func TestDockerContainerManager_StartLivenessMonitoring(t *testing.T) {
	logger := zaptest.NewLogger(t)
	dcm, err := NewDockerContainerManager(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()
	containerID := "container123"

	config := &LivenessConfig{
		HealthCheckConfig: HealthCheckConfig{
			Enabled:          true,
			Interval:         1 * time.Second,
			FailureThreshold: 2,
		},
		RestartPolicy: RestartPolicy{
			Enabled:     true,
			MaxRestarts: 3,
		},
		ResourceMonitoring:    true,
		ResourceCheckInterval: 2 * time.Second,
	}

	eventChan, err := dcm.StartLivenessMonitoring(ctx, containerID, config)

	require.NoError(t, err)
	assert.NotNil(t, eventChan)

	// Verify monitor was created
	dcm.mu.RLock()
	monitor, exists := dcm.livenessMonitors[containerID]
	dcm.mu.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, containerID, monitor.containerID)
	assert.Equal(t, config, monitor.config)
	assert.Equal(t, 0, monitor.restartCount)

	// Stop monitoring
	dcm.StopLivenessMonitoring(containerID)

	// Verify monitor was removed
	dcm.mu.RLock()
	_, exists = dcm.livenessMonitors[containerID]
	dcm.mu.RUnlock()

	assert.False(t, exists)
}

func TestDockerContainerManager_TriggerRestart(t *testing.T) {
	logger := zaptest.NewLogger(t)
	dcm, err := NewDockerContainerManager(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()
	containerID := "container123"

	config := &LivenessConfig{
		RestartPolicy: RestartPolicy{
			Enabled: true,
		},
	}

	// Start monitoring to create a monitor
	eventChan, err := dcm.StartLivenessMonitoring(ctx, containerID, config)
	require.NoError(t, err)
	require.NotNil(t, eventChan)

	// Test triggering restart
	err = dcm.TriggerRestart(containerID, "test restart")
	assert.NoError(t, err)

	// Should receive event on channel
	select {
	case event := <-eventChan:
		assert.Equal(t, containerID, event.ContainerID)
		assert.Equal(t, EventUnhealthy, event.Type)
		assert.Contains(t, event.Message, "test restart")
	case <-time.After(1 * time.Second):
		t.Fatal("Expected to receive restart event")
	}

	// Test with non-existent container
	err = dcm.TriggerRestart("nonexistent", "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no liveness monitor found")

	// Stop monitoring
	dcm.StopLivenessMonitoring(containerID)
}

func TestDockerContainerManager_RestartPolicyOperations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	dcm, err := NewDockerContainerManager(nil, logger)
	require.NoError(t, err)
	defer func() { _ = dcm.Shutdown(context.Background()) }()

	ctx := context.Background()
	containerID := "test-container"

	// Start monitoring first
	config := &LivenessConfig{
		HealthCheckConfig: HealthCheckConfig{
			Enabled:          true,
			Interval:         1 * time.Second,
			FailureThreshold: 3,
		},
		RestartPolicy: RestartPolicy{
			Enabled:     true,
			MaxRestarts: 3,
		},
		ResourceMonitoring: false,
	}

	_, err = dcm.StartLivenessMonitoring(ctx, containerID, config)
	require.NoError(t, err)

	// Test setting restart policy
	newPolicy := RestartPolicy{
		Enabled:     true,
		MaxRestarts: 5,
	}

	err = dcm.SetRestartPolicy(containerID, newPolicy)
	assert.NoError(t, err)

	// Verify policy was updated
	dcm.mu.RLock()
	monitor := dcm.livenessMonitors[containerID]
	dcm.mu.RUnlock()

	assert.Equal(t, 5, monitor.restartPolicy.MaxRestarts)

	// Clean up
	dcm.StopLivenessMonitoring(containerID)
}

func TestDockerContainerManager_HealthCheckOperations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	dcm, err := NewDockerContainerManager(nil, logger)
	require.NoError(t, err)
	defer func() { _ = dcm.Shutdown(context.Background()) }()

	ctx := context.Background()
	containerID := "test-container"

	config := &HealthCheckConfig{
		Enabled:          true,
		Interval:         100 * time.Millisecond,
		FailureThreshold: 2,
	}

	// Start health check (will fail since container doesn't exist, but we're testing the setup)
	healthChan, err := dcm.StartHealthCheck(ctx, containerID, config)
	require.NoError(t, err)
	assert.NotNil(t, healthChan)

	// Verify health check was created
	dcm.mu.RLock()
	_, exists := dcm.healthChecks[containerID]
	dcm.mu.RUnlock()
	assert.True(t, exists)

	// Stop health check
	dcm.StopHealthCheck(containerID)

	// Verify health check was removed
	dcm.mu.RLock()
	_, exists = dcm.healthChecks[containerID]
	dcm.mu.RUnlock()
	assert.False(t, exists)
}

func TestDockerContainerManager_ShutdownOperations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	dcm, err := NewDockerContainerManager(nil, logger)
	require.NoError(t, err)

	ctx := context.Background()
	containerID := "test-container"

	// Start some monitoring
	config := &LivenessConfig{
		HealthCheckConfig: HealthCheckConfig{
			Enabled:          true,
			Interval:         1 * time.Second,
			FailureThreshold: 3,
		},
		RestartPolicy:      RestartPolicy{Enabled: true},
		ResourceMonitoring: false,
	}
	_, err = dcm.StartLivenessMonitoring(ctx, containerID, config)
	require.NoError(t, err)

	// Start health check
	healthConfig := &HealthCheckConfig{Enabled: true}
	_, err = dcm.StartHealthCheck(ctx, containerID, healthConfig)
	require.NoError(t, err)

	// Verify resources exist
	dcm.mu.RLock()
	assert.NotEmpty(t, dcm.livenessMonitors)
	assert.NotEmpty(t, dcm.healthChecks)
	dcm.mu.RUnlock()

	// Shutdown
	err = dcm.Shutdown(ctx)
	assert.NoError(t, err)

	// Verify all monitors are cleaned up
	dcm.mu.RLock()
	assert.Empty(t, dcm.healthChecks)
	assert.Empty(t, dcm.livenessMonitors)
	dcm.mu.RUnlock()
}
