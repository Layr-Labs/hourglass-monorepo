package containerManager

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func BenchmarkHashAvsAddress(b *testing.B) {
	address := "0x1234567890abcdef1234567890abcdef12345678"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HashAvsAddress(address)
	}
}

func BenchmarkCreateDefaultContainerConfig(b *testing.B) {
	avsAddress := "0x1234567890abcdef1234567890abcdef12345678"
	imageRepo := "myregistry/myapp"
	imageTag := "v1.0.0"
	containerPort := 8080
	networkName := "avs-network"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CreateDefaultContainerConfig(avsAddress, imageRepo, imageTag, containerPort, networkName)
	}
}

func BenchmarkContainerManagerCreation(b *testing.B) {
	logger := zap.NewNop()
	config := &ContainerManagerConfig{
		DefaultStartTimeout: 30 * time.Second,
		DefaultStopTimeout:  10 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dcm, err := NewDockerContainerManager(config, logger)
		if err != nil {
			b.Fatal(err)
		}
		// Clean up
		_ = dcm.Shutdown(context.Background())
	}
}

func BenchmarkLivenessMonitoringSetup(b *testing.B) {
	logger := zap.NewNop()
	dcm, err := NewDockerContainerManager(nil, logger)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = dcm.Shutdown(context.Background()) }()

	ctx := context.Background()
	containerID := "benchmark-container"

	config := &LivenessConfig{
		HealthCheckConfig: HealthCheckConfig{
			Enabled:          true,
			Interval:         100 * time.Millisecond,
			FailureThreshold: 3,
		},
		RestartPolicy: RestartPolicy{
			Enabled:     true,
			MaxRestarts: 5,
		},
		ResourceMonitoring:    false, // Disable to avoid resource monitoring overhead
		ResourceCheckInterval: 1 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eventChan, err := dcm.StartLivenessMonitoring(ctx, containerID, config)
		if err != nil {
			b.Fatal(err)
		}

		// Clean up
		dcm.StopLivenessMonitoring(containerID)

		// Drain channel if needed
		select {
		case <-eventChan:
		default:
		}
	}
}

func BenchmarkTriggerRestart(b *testing.B) {
	logger := zap.NewNop()
	dcm, err := NewDockerContainerManager(nil, logger)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = dcm.Shutdown(context.Background()) }()

	ctx := context.Background()
	containerID := "benchmark-container"

	config := &LivenessConfig{
		HealthCheckConfig: HealthCheckConfig{
			Enabled:          true,
			Interval:         100 * time.Millisecond,
			FailureThreshold: 3,
		},
		RestartPolicy: RestartPolicy{
			Enabled: true,
		},
		ResourceMonitoring: false,
	}

	// Setup monitoring
	eventChan, err := dcm.StartLivenessMonitoring(ctx, containerID, config)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := dcm.TriggerRestart(containerID, "benchmark test")
		if err != nil {
			b.Fatal(err)
		}

		// Drain event channel
		select {
		case <-eventChan:
		case <-time.After(10 * time.Millisecond):
		}
	}

	// Clean up
	dcm.StopLivenessMonitoring(containerID)
}

func BenchmarkContainerStateCreation(b *testing.B) {
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &ContainerState{
			Status:       "running",
			ExitCode:     0,
			StartedAt:    now,
			RestartCount: 0,
			OOMKilled:    false,
			Error:        "",
			Restarting:   false,
		}
	}
}

func BenchmarkContainerEventCreation(b *testing.B) {
	now := time.Now()
	state := ContainerState{
		Status:       "running",
		ExitCode:     0,
		StartedAt:    now,
		RestartCount: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ContainerEvent{
			ContainerID: "container123",
			Type:        EventStarted,
			State:       state,
			Timestamp:   now,
			Message:     "Container started",
		}
	}
}

// BenchmarkConcurrentMonitoring tests performance with multiple concurrent monitors
func BenchmarkConcurrentMonitoring(b *testing.B) {
	logger := zap.NewNop()
	dcm, err := NewDockerContainerManager(nil, logger)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = dcm.Shutdown(context.Background()) }()

	ctx := context.Background()

	config := &LivenessConfig{
		HealthCheckConfig: HealthCheckConfig{
			Enabled:          true,
			Interval:         100 * time.Millisecond,
			FailureThreshold: 3,
		},
		RestartPolicy: RestartPolicy{
			Enabled: false, // Disable to avoid restart overhead
		},
		ResourceMonitoring: false, // Disable to focus on core monitoring
	}

	// Number of concurrent monitors
	numMonitors := 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Start multiple monitors concurrently
		eventChans := make([]<-chan ContainerEvent, numMonitors)
		for j := 0; j < numMonitors; j++ {
			containerID := string(rune('a' + j)) // Simple container ID generation
			eventChan, err := dcm.StartLivenessMonitoring(ctx, containerID, config)
			if err != nil {
				b.Fatal(err)
			}
			eventChans[j] = eventChan
		}

		// Let monitors run briefly
		time.Sleep(10 * time.Millisecond)

		// Stop all monitors
		for j := 0; j < numMonitors; j++ {
			containerID := string(rune('a' + j))
			dcm.StopLivenessMonitoring(containerID)

			// Drain event channels
			select {
			case <-eventChans[j]:
			default:
			}
		}
	}
}

// BenchmarkMemoryUsage tests memory efficiency
func BenchmarkMemoryUsage(b *testing.B) {
	logger := zap.NewNop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		dcm, err := NewDockerContainerManager(nil, logger)
		if err != nil {
			b.Fatal(err)
		}

		// Create some monitoring
		ctx := context.Background()
		config := &LivenessConfig{
			HealthCheckConfig: HealthCheckConfig{
				Enabled:          true,
				Interval:         100 * time.Millisecond,
				FailureThreshold: 3,
			},
			RestartPolicy:      RestartPolicy{Enabled: false},
			ResourceMonitoring: false,
		}

		_, err = dcm.StartLivenessMonitoring(ctx, "test", config)
		if err != nil {
			b.Fatal(err)
		}

		// Clean up
		_ = dcm.Shutdown(context.Background())
	}
}
