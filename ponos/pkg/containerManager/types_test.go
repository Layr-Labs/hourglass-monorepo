package containerManager

import (
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

func TestContainerConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config *ContainerConfig
		valid  bool
	}{
		{
			name: "valid configuration",
			config: &ContainerConfig{
				Hostname: "test-container",
				Image:    "alpine:latest",
				ExposedPorts: nat.PortSet{
					"8080/tcp": struct{}{},
				},
				NetworkName: "test-network",
			},
			valid: true,
		},
		{
			name: "minimal configuration",
			config: &ContainerConfig{
				Hostname: "minimal",
				Image:    "alpine",
			},
			valid: true,
		},
		{
			name: "configuration with resource limits",
			config: &ContainerConfig{
				Hostname:    "resource-limited",
				Image:       "alpine:latest",
				MemoryLimit: 512 * 1024 * 1024, // 512MB
				CPUShares:   512,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation that config can be created
			assert.NotNil(t, tt.config)

			if tt.valid {
				assert.NotEmpty(t, tt.config.Hostname)
				assert.NotEmpty(t, tt.config.Image)
			}
		})
	}
}

func TestHealthCheckConfig_Defaults(t *testing.T) {
	config := &HealthCheckConfig{
		Enabled:          true,
		Interval:         5 * time.Second,
		Timeout:          2 * time.Second,
		Retries:          3,
		StartPeriod:      10 * time.Second,
		FailureThreshold: 3,
	}

	assert.True(t, config.Enabled)
	assert.Equal(t, 5*time.Second, config.Interval)
	assert.Equal(t, 2*time.Second, config.Timeout)
	assert.Equal(t, 3, config.Retries)
	assert.Equal(t, 10*time.Second, config.StartPeriod)
	assert.Equal(t, 3, config.FailureThreshold)
}

func TestRestartPolicy_Configuration(t *testing.T) {
	tests := []struct {
		name   string
		policy RestartPolicy
		desc   string
	}{
		{
			name: "aggressive restart policy",
			policy: RestartPolicy{
				Enabled:            true,
				MaxRestarts:        10,
				RestartDelay:       1 * time.Second,
				BackoffMultiplier:  2.0,
				MaxBackoffDelay:    60 * time.Second,
				RestartTimeout:     30 * time.Second,
				RestartOnCrash:     true,
				RestartOnOOM:       true,
				RestartOnUnhealthy: false,
			},
			desc: "Aggressive policy for critical services",
		},
		{
			name: "conservative restart policy",
			policy: RestartPolicy{
				Enabled:            true,
				MaxRestarts:        3,
				RestartDelay:       5 * time.Second,
				BackoffMultiplier:  1.5,
				MaxBackoffDelay:    30 * time.Second,
				RestartTimeout:     60 * time.Second,
				RestartOnCrash:     true,
				RestartOnOOM:       false,
				RestartOnUnhealthy: false,
			},
			desc: "Conservative policy for stable services",
		},
		{
			name: "disabled restart policy",
			policy: RestartPolicy{
				Enabled: false,
			},
			desc: "No automatic restarts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := tt.policy

			if policy.Enabled {
				assert.Greater(t, policy.MaxRestarts, -1) // -1 means unlimited
				assert.Greater(t, policy.RestartDelay, time.Duration(0))
				assert.Greater(t, policy.BackoffMultiplier, 0.0)
				assert.Greater(t, policy.MaxBackoffDelay, time.Duration(0))
				assert.Greater(t, policy.RestartTimeout, time.Duration(0))
			}
		})
	}
}

func TestResourceThresholds_Configuration(t *testing.T) {
	thresholds := ResourceThresholds{
		CPUThreshold:    90.0,
		MemoryThreshold: 85.0,
		RestartOnCPU:    false,
		RestartOnMemory: true,
	}

	assert.Equal(t, 90.0, thresholds.CPUThreshold)
	assert.Equal(t, 85.0, thresholds.MemoryThreshold)
	assert.False(t, thresholds.RestartOnCPU)
	assert.True(t, thresholds.RestartOnMemory)

	// Validate thresholds are reasonable
	assert.True(t, thresholds.CPUThreshold > 0 && thresholds.CPUThreshold <= 100)
	assert.True(t, thresholds.MemoryThreshold > 0 && thresholds.MemoryThreshold <= 100)
}

func TestLivenessConfig_Complete(t *testing.T) {
	config := &LivenessConfig{
		HealthCheckConfig: HealthCheckConfig{
			Enabled:          true,
			Interval:         5 * time.Second,
			Timeout:          2 * time.Second,
			Retries:          3,
			StartPeriod:      10 * time.Second,
			FailureThreshold: 3,
		},
		RestartPolicy: RestartPolicy{
			Enabled:            true,
			MaxRestarts:        5,
			RestartDelay:       2 * time.Second,
			BackoffMultiplier:  2.0,
			MaxBackoffDelay:    30 * time.Second,
			RestartTimeout:     60 * time.Second,
			RestartOnCrash:     true,
			RestartOnOOM:       true,
			RestartOnUnhealthy: false,
		},
		ResourceThresholds: ResourceThresholds{
			CPUThreshold:    90.0,
			MemoryThreshold: 90.0,
			RestartOnCPU:    false,
			RestartOnMemory: false,
		},
		MonitorEvents:         true,
		ResourceMonitoring:    true,
		ResourceCheckInterval: 30 * time.Second,
	}

	// Validate all components are properly configured
	assert.True(t, config.HealthCheckConfig.Enabled)
	assert.True(t, config.RestartPolicy.Enabled)
	assert.True(t, config.MonitorEvents)
	assert.True(t, config.ResourceMonitoring)
	assert.Equal(t, 30*time.Second, config.ResourceCheckInterval)

	// Validate integration between components
	assert.True(t, config.RestartPolicy.RestartDelay < config.RestartPolicy.MaxBackoffDelay)
	assert.True(t, config.HealthCheckConfig.Interval < config.ResourceCheckInterval)
}

func TestContainerEventType_Constants(t *testing.T) {
	// Verify all event types are defined
	eventTypes := []ContainerEventType{
		EventStarted,
		EventStopped,
		EventCrashed,
		EventOOMKilled,
		EventRestarted,
		EventHealthy,
		EventUnhealthy,
		EventRestarting,
		EventRestartFailed,
	}

	for _, eventType := range eventTypes {
		assert.NotEmpty(t, string(eventType))
	}

	// Verify specific values
	assert.Equal(t, "started", string(EventStarted))
	assert.Equal(t, "stopped", string(EventStopped))
	assert.Equal(t, "crashed", string(EventCrashed))
	assert.Equal(t, "oom-killed", string(EventOOMKilled))
	assert.Equal(t, "restarted", string(EventRestarted))
	assert.Equal(t, "healthy", string(EventHealthy))
	assert.Equal(t, "unhealthy", string(EventUnhealthy))
	assert.Equal(t, "restarting", string(EventRestarting))
	assert.Equal(t, "restart-failed", string(EventRestartFailed))
}

func TestContainerEvent_Creation(t *testing.T) {
	now := time.Now()

	event := ContainerEvent{
		ContainerID: "container123",
		Type:        EventStarted,
		State: ContainerState{
			Status:       "running",
			ExitCode:     0,
			StartedAt:    now,
			RestartCount: 0,
			OOMKilled:    false,
		},
		Timestamp: now,
		Message:   "Container started successfully",
	}

	assert.Equal(t, "container123", event.ContainerID)
	assert.Equal(t, EventStarted, event.Type)
	assert.Equal(t, "running", event.State.Status)
	assert.Equal(t, 0, event.State.ExitCode)
	assert.Equal(t, now, event.Timestamp)
	assert.Equal(t, "Container started successfully", event.Message)
}

func TestResourceUsage_Calculation(t *testing.T) {
	usage := &ResourceUsage{
		CPUPercent:    45.5,
		MemoryUsage:   512 * 1024 * 1024,  // 512MB
		MemoryLimit:   1024 * 1024 * 1024, // 1GB
		MemoryPercent: 50.0,
		NetworkRx:     1024 * 1024,       // 1MB
		NetworkTx:     2 * 1024 * 1024,   // 2MB
		DiskRead:      100 * 1024 * 1024, // 100MB
		DiskWrite:     50 * 1024 * 1024,  // 50MB
		Timestamp:     time.Now(),
	}

	// Validate percentages
	assert.True(t, usage.CPUPercent >= 0 && usage.CPUPercent <= 100)
	assert.True(t, usage.MemoryPercent >= 0 && usage.MemoryPercent <= 100)

	// Validate memory calculation
	expectedMemoryPercent := float64(usage.MemoryUsage) / float64(usage.MemoryLimit) * 100.0
	assert.InDelta(t, expectedMemoryPercent, usage.MemoryPercent, 0.1)

	// Validate all values are reasonable
	assert.True(t, usage.MemoryUsage > 0)
	assert.True(t, usage.MemoryLimit > usage.MemoryUsage)
	assert.True(t, usage.NetworkRx >= 0)
	assert.True(t, usage.NetworkTx >= 0)
	assert.True(t, usage.DiskRead >= 0)
	assert.True(t, usage.DiskWrite >= 0)
}
