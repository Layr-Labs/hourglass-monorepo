package containerManager

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
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

// isTestContainerAvailable checks if the ponos test container image is available
func isTestContainerAvailable(t *testing.T) bool {
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}", "ponos-test-container:latest")
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Failed to check for test container image: %v", err)
		return false
	}

	if !strings.Contains(string(output), "ponos-test-container:latest") {
		t.Logf("Test container image not found. Run 'make build/test-container' to build it.")
		return false
	}

	return true
}

// These tests require Docker to be running and available
// Run with: go test -tags=integration

func TestDockerContainerManager_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if Docker is not available
	if !isDockerAvailable(t) {
		t.Skip("Docker is not available, skipping integration tests")
	}

	if !isTestContainerAvailable(t) {
		t.Skip("Test container image not available, skipping integration tests. Run 'make build/test-container' to build it.")
	}

	logger := zaptest.NewLogger(t)
	dcm, err := NewDockerContainerManager(nil, logger)
	require.NoError(t, err)
	defer func() { _ = dcm.Shutdown(context.Background()) }()

	ctx := context.Background()

	// Test with our test container
	config := &ContainerConfig{
		Hostname:    "test-ponos",
		Image:       "ponos-test-container:latest",
		NetworkName: "test-network",
		AutoRemove:  true,
		ExposedPorts: nat.PortSet{
			"8080/tcp": struct{}{},
		},
	}

	t.Run("create and manage container lifecycle", func(t *testing.T) {
		// Create container
		containerInfo, err := dcm.Create(ctx, config)
		require.NoError(t, err)
		assert.NotEmpty(t, containerInfo.ID)
		assert.Equal(t, "test-ponos", containerInfo.Hostname)

		// Start container
		err = dcm.Start(ctx, containerInfo.ID)
		require.NoError(t, err)

		// Check if running
		running, err := dcm.IsRunning(ctx, containerInfo.ID)
		require.NoError(t, err)
		assert.True(t, running)

		// Get container state
		state, err := dcm.GetContainerState(ctx, containerInfo.ID)
		require.NoError(t, err)
		assert.Equal(t, "running", state.Status)
		assert.False(t, state.OOMKilled)

		// Stop container
		err = dcm.Stop(ctx, containerInfo.ID, 5*time.Second)
		require.NoError(t, err)

		// Verify stopped
		running, err = dcm.IsRunning(ctx, containerInfo.ID)
		require.NoError(t, err)
		assert.False(t, running)

		// Remove container
		err = dcm.Remove(ctx, containerInfo.ID, true)
		require.NoError(t, err)
	})

	t.Run("network management", func(t *testing.T) {
		networkName := "test-network-" + time.Now().Format("20060102150405")

		// Create network
		err := dcm.CreateNetworkIfNotExists(ctx, networkName)
		require.NoError(t, err)

		// Create again (should not error)
		err = dcm.CreateNetworkIfNotExists(ctx, networkName)
		require.NoError(t, err)

		// Remove network
		err = dcm.RemoveNetwork(ctx, networkName)
		require.NoError(t, err)
	})
}

func TestDockerContainerManager_LivenessMonitoring_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !isDockerAvailable(t) {
		t.Skip("Docker is not available, skipping integration tests")
	}

	if !isTestContainerAvailable(t) {
		t.Skip("Test container image not available, skipping integration tests. Run 'make build/test-container' to build it.")
	}

	logger := zaptest.NewLogger(t)
	dcm, err := NewDockerContainerManager(nil, logger)
	require.NoError(t, err)
	defer func() { _ = dcm.Shutdown(context.Background()) }()

	ctx := context.Background()

	// Use test container with HTTP server
	config := &ContainerConfig{
		Hostname:   "test-monitoring",
		Image:      "ponos-test-container:latest",
		Env:        []string{},
		AutoRemove: true,
		ExposedPorts: nat.PortSet{
			"8080/tcp": struct{}{},
		},
	}

	t.Run("health monitoring with running container", func(t *testing.T) {
		// Create and start container
		containerInfo, err := dcm.Create(ctx, config)
		require.NoError(t, err)

		err = dcm.Start(ctx, containerInfo.ID)
		require.NoError(t, err)

		// Start liveness monitoring
		livenessConfig := &LivenessConfig{
			HealthCheckConfig: HealthCheckConfig{
				Enabled:          true,
				Interval:         500 * time.Millisecond,
				FailureThreshold: 2,
			},
			RestartPolicy: RestartPolicy{
				Enabled:     false, // Don't auto-restart for this test
				MaxRestarts: 0,
			},
			ResourceMonitoring:    true,
			ResourceCheckInterval: 1 * time.Second,
		}

		eventChan, err := dcm.StartLivenessMonitoring(ctx, containerInfo.ID, livenessConfig)
		require.NoError(t, err)
		require.NotNil(t, eventChan)

		// Wait a moment for monitoring to start
		time.Sleep(1 * time.Second)

		// Get resource usage
		usage, err := dcm.GetResourceUsage(ctx, containerInfo.ID)
		if err == nil { // Resource monitoring might not work in all environments
			assert.NotNil(t, usage)
			assert.True(t, usage.Timestamp.After(time.Now().Add(-5*time.Second)))
		}

		// Stop monitoring
		dcm.StopLivenessMonitoring(containerInfo.ID)

		// Clean up
		_ = dcm.Stop(ctx, containerInfo.ID, 5*time.Second)
		_ = dcm.Remove(ctx, containerInfo.ID, true)
	})

	t.Run("manual restart trigger", func(t *testing.T) {
		// Create and start container
		containerInfo, err := dcm.Create(ctx, config)
		require.NoError(t, err)

		err = dcm.Start(ctx, containerInfo.ID)
		require.NoError(t, err)

		// Start liveness monitoring with restart enabled
		livenessConfig := &LivenessConfig{
			HealthCheckConfig: HealthCheckConfig{
				Enabled:          true,
				Interval:         1 * time.Second,
				FailureThreshold: 3,
			},
			RestartPolicy: RestartPolicy{
				Enabled:        true,
				MaxRestarts:    1,
				RestartDelay:   1 * time.Second,
				RestartTimeout: 10 * time.Second,
			},
			ResourceMonitoring: false, // Disable to reduce noise
		}

		eventChan, err := dcm.StartLivenessMonitoring(ctx, containerInfo.ID, livenessConfig)
		require.NoError(t, err)

		// Trigger manual restart
		err = dcm.TriggerRestart(containerInfo.ID, "integration test")
		require.NoError(t, err)

		// Should receive unhealthy event
		select {
		case event := <-eventChan:
			assert.Equal(t, containerInfo.ID, event.ContainerID)
			assert.Equal(t, EventUnhealthy, event.Type)
			assert.Contains(t, event.Message, "integration test")
		case <-time.After(3 * time.Second):
			t.Fatal("Expected to receive restart event")
		}

		// Stop monitoring
		dcm.StopLivenessMonitoring(containerInfo.ID)

		// Clean up
		_ = dcm.Stop(ctx, containerInfo.ID, 5*time.Second)
		_ = dcm.Remove(ctx, containerInfo.ID, true)
	})
}

func TestDockerContainerManager_HealthCheck_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !isDockerAvailable(t) {
		t.Skip("Docker is not available, skipping integration tests")
	}

	if !isTestContainerAvailable(t) {
		t.Skip("Test container image not available, skipping integration tests. Run 'make build/test-container' to build it.")
	}

	logger := zaptest.NewLogger(t)
	dcm, err := NewDockerContainerManager(nil, logger)
	require.NoError(t, err)
	defer func() { _ = dcm.Shutdown(context.Background()) }()

	ctx := context.Background()

	config := &ContainerConfig{
		Hostname:   "test-health",
		Image:      "ponos-test-container:latest",
		AutoRemove: true,
		ExposedPorts: nat.PortSet{
			"8080/tcp": struct{}{},
		},
	}

	t.Run("health check with running container", func(t *testing.T) {
		// Create and start container
		containerInfo, err := dcm.Create(ctx, config)
		require.NoError(t, err)

		err = dcm.Start(ctx, containerInfo.ID)
		require.NoError(t, err)

		// Start health check
		healthConfig := &HealthCheckConfig{
			Enabled:          true,
			Interval:         200 * time.Millisecond,
			FailureThreshold: 2,
		}

		healthChan, err := dcm.StartHealthCheck(ctx, containerInfo.ID, healthConfig)
		require.NoError(t, err)
		require.NotNil(t, healthChan)

		// Should receive healthy status
		select {
		case healthy := <-healthChan:
			assert.True(t, healthy)
		case <-time.After(2 * time.Second):
			t.Fatal("Expected to receive health status")
		}

		// Stop health check
		dcm.StopHealthCheck(containerInfo.ID)

		// Clean up
		_ = dcm.Stop(ctx, containerInfo.ID, 5*time.Second)
		_ = dcm.Remove(ctx, containerInfo.ID, true)
	})
}

func TestDockerContainerManager_DefaultConfig_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !isDockerAvailable(t) {
		t.Skip("Docker is not available, skipping integration tests")
	}

	if !isTestContainerAvailable(t) {
		t.Skip("Test container image not available, skipping integration tests. Run 'make build/test-container' to build it.")
	}

	logger := zaptest.NewLogger(t)
	dcm, err := NewDockerContainerManager(nil, logger)
	require.NoError(t, err)
	defer func() { _ = dcm.Shutdown(context.Background()) }()

	ctx := context.Background()

	// Test CreateDefaultContainerConfig with real container
	avsAddress := "0x1234567890abcdef"
	imageRepo := "ponos-test-container"
	imageTag := "latest"
	containerPort := 8080
	networkName := "test-default-network"

	config := CreateDefaultContainerConfig(avsAddress, imageRepo, imageTag, containerPort, networkName)

	t.Run("default configuration creates valid container", func(t *testing.T) {
		// Create container with default config
		containerInfo, err := dcm.Create(ctx, config)
		require.NoError(t, err)
		assert.NotEmpty(t, containerInfo.ID)

		// Verify hostname was set correctly
		expectedHostname := "avs-performer-" + HashAvsAddress(avsAddress)
		assert.Equal(t, expectedHostname, containerInfo.Hostname)

		// Start container
		err = dcm.Start(ctx, containerInfo.ID)
		require.NoError(t, err)

		// Inspect container to verify configuration
		inspectedInfo, err := dcm.Inspect(ctx, containerInfo.ID)
		require.NoError(t, err)
		assert.Equal(t, expectedHostname, inspectedInfo.Hostname)

		// Clean up
		_ = dcm.Stop(ctx, containerInfo.ID, 5*time.Second)
		_ = dcm.Remove(ctx, containerInfo.ID, true)

		// Clean up network
		if !strings.Contains(networkName, "bridge") {
			_ = dcm.RemoveNetwork(ctx, networkName)
		}
	})
}
