package containerManager

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// ContainerManagerTestSuite provides a comprehensive test suite
type ContainerManagerTestSuite struct {
	suite.Suite
	dcm    *DockerContainerManager
	logger *zap.Logger
	ctx    context.Context
}

// SetupSuite runs once before all tests
func (suite *ContainerManagerTestSuite) SetupSuite() {
	suite.logger = zaptest.NewLogger(suite.T())
	suite.ctx = context.Background()

	var err error
	suite.dcm, err = NewDockerContainerManager(nil, suite.logger)
	suite.Require().NoError(err)
}

// TearDownSuite runs once after all tests
func (suite *ContainerManagerTestSuite) TearDownSuite() {
	if suite.dcm != nil {
		_ = suite.dcm.Shutdown(suite.ctx)
	}
}

// SetupTest runs before each test
func (suite *ContainerManagerTestSuite) SetupTest() {
	// Reset any test-specific state if needed
}

// TearDownTest runs after each test
func (suite *ContainerManagerTestSuite) TearDownTest() {
	// Clean up any test-specific resources
}

// TestBasicFunctionality tests core container manager functionality
func (suite *ContainerManagerTestSuite) TestBasicFunctionality() {
	suite.T().Run("container manager creation", func(t *testing.T) {
		suite.NotNil(suite.dcm)
		suite.NotNil(suite.dcm.config)
		suite.NotNil(suite.dcm.logger)
		suite.NotNil(suite.dcm.healthChecks)
		suite.NotNil(suite.dcm.livenessMonitors)
	})

	suite.T().Run("default configuration", func(t *testing.T) {
		config := suite.dcm.config
		suite.Equal(30*time.Second, config.DefaultStartTimeout)
		suite.Equal(10*time.Second, config.DefaultStopTimeout)
		suite.NotNil(config.DefaultHealthCheckConfig)
		suite.NotNil(config.DefaultLivenessConfig)
	})
}

// TestMonitoringFunctionality tests monitoring capabilities
func (suite *ContainerManagerTestSuite) TestMonitoringFunctionality() {
	containerID := "test-container-monitoring"

	suite.T().Run("start liveness monitoring", func(t *testing.T) {
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

		eventChan, err := suite.dcm.StartLivenessMonitoring(suite.ctx, containerID, config)
		suite.NoError(err)
		suite.NotNil(eventChan)

		// Verify monitor was created
		suite.dcm.mu.RLock()
		monitor, exists := suite.dcm.livenessMonitors[containerID]
		suite.dcm.mu.RUnlock()

		suite.True(exists)
		suite.Equal(containerID, monitor.containerID)
		suite.Equal(0, monitor.restartCount)
	})

	suite.T().Run("trigger manual restart", func(t *testing.T) {
		err := suite.dcm.TriggerRestart(containerID, "test restart")
		suite.NoError(err)
	})

	suite.T().Run("stop liveness monitoring", func(t *testing.T) {
		suite.dcm.StopLivenessMonitoring(containerID)

		// Verify monitor was removed
		suite.dcm.mu.RLock()
		_, exists := suite.dcm.livenessMonitors[containerID]
		suite.dcm.mu.RUnlock()

		suite.False(exists)
	})
}

// TestHealthCheckFunctionality tests health check capabilities
func (suite *ContainerManagerTestSuite) TestHealthCheckFunctionality() {
	containerID := "test-container-health"

	suite.T().Run("start and stop health check", func(t *testing.T) {
		config := &HealthCheckConfig{
			Enabled:          true,
			Interval:         500 * time.Millisecond,
			FailureThreshold: 2,
		}

		healthChan, err := suite.dcm.StartHealthCheck(suite.ctx, containerID, config)
		suite.NoError(err)
		suite.NotNil(healthChan)

		// Verify health check was created
		suite.dcm.mu.RLock()
		_, exists := suite.dcm.healthChecks[containerID]
		suite.dcm.mu.RUnlock()
		suite.True(exists)

		// Stop health check
		suite.dcm.StopHealthCheck(containerID)

		// Verify health check was removed
		suite.dcm.mu.RLock()
		_, exists = suite.dcm.healthChecks[containerID]
		suite.dcm.mu.RUnlock()
		suite.False(exists)
	})
}

// TestUtilityFunctions tests utility functions
func (suite *ContainerManagerTestSuite) TestUtilityFunctions() {
	suite.T().Run("hash AVS address", func(t *testing.T) {
		address := "0x1234567890abcdef"
		hash := HashAvsAddress(address)
		suite.Len(hash, 6)
		suite.NotEmpty(hash)

		// Same input should produce same output
		hash2 := HashAvsAddress(address)
		suite.Equal(hash, hash2)
	})

	suite.T().Run("create default container config", func(t *testing.T) {
		avsAddress := "0x1234567890abcdef"
		imageRepo := "test/app"
		imageTag := "v1.0.0"
		containerPort := 8080
		networkName := "test-network"

		config := CreateDefaultContainerConfig(avsAddress, imageRepo, imageTag, containerPort, networkName)

		suite.Equal("avs-performer-"+HashAvsAddress(avsAddress), config.Hostname)
		suite.Equal("test/app:v1.0.0", config.Image)
		suite.Equal("test-network", config.NetworkName)
		suite.True(config.AutoRemove)
		suite.Equal("no", config.RestartPolicy)
	})
}

// TestErrorHandling tests error scenarios
func (suite *ContainerManagerTestSuite) TestErrorHandling() {
	suite.T().Run("trigger restart on non-existent container", func(t *testing.T) {
		err := suite.dcm.TriggerRestart("nonexistent", "test")
		suite.Error(err)
		suite.Contains(err.Error(), "no liveness monitor found")
	})

	suite.T().Run("stop monitoring on non-existent container", func(t *testing.T) {
		// Should not panic or error
		suite.dcm.StopLivenessMonitoring("nonexistent")
		suite.dcm.StopHealthCheck("nonexistent")
	})
}

// TestConcurrency tests concurrent operations
func (suite *ContainerManagerTestSuite) TestConcurrency() {
	suite.T().Run("concurrent monitoring operations", func(t *testing.T) {
		numGoroutines := 10
		done := make(chan bool, numGoroutines)

		config := &LivenessConfig{
			HealthCheckConfig: HealthCheckConfig{
				Enabled:          true,
				Interval:         100 * time.Millisecond,
				FailureThreshold: 3,
			},
			RestartPolicy: RestartPolicy{
				Enabled: false,
			},
			ResourceMonitoring: false,
		}

		// Start multiple monitoring operations concurrently
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				containerID := suite.T().Name() + string(rune('a'+id))

				eventChan, err := suite.dcm.StartLivenessMonitoring(suite.ctx, containerID, config)
				suite.NoError(err)
				suite.NotNil(eventChan)

				// Let it run briefly
				time.Sleep(50 * time.Millisecond)

				// Stop monitoring
				suite.dcm.StopLivenessMonitoring(containerID)

				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			select {
			case <-done:
				// Success
			case <-time.After(5 * time.Second):
				suite.Fail("Timeout waiting for concurrent operations")
			}
		}
	})
}

// TestShutdown tests cleanup functionality
func (suite *ContainerManagerTestSuite) TestShutdown() {
	suite.T().Run("shutdown cleans up resources", func(t *testing.T) {
		// Create a separate container manager for this test
		dcm, err := NewDockerContainerManager(nil, suite.logger)
		suite.NoError(err)

		// Start some monitoring
		containerID := "test-shutdown"
		config := &LivenessConfig{
			RestartPolicy: RestartPolicy{Enabled: true},
		}
		_, err = dcm.StartLivenessMonitoring(suite.ctx, containerID, config)
		suite.NoError(err)

		// Start health check
		healthConfig := &HealthCheckConfig{Enabled: true}
		_, err = dcm.StartHealthCheck(suite.ctx, containerID, healthConfig)
		suite.NoError(err)

		// Verify resources exist
		dcm.mu.RLock()
		suite.NotEmpty(dcm.livenessMonitors)
		suite.NotEmpty(dcm.healthChecks)
		dcm.mu.RUnlock()

		// Shutdown
		err = dcm.Shutdown(suite.ctx)
		suite.NoError(err)

		// Verify cleanup
		dcm.mu.RLock()
		suite.Empty(dcm.livenessMonitors)
		suite.Empty(dcm.healthChecks)
		dcm.mu.RUnlock()
	})
}

// TestSuite runs the entire test suite
func TestSuite(t *testing.T) {
	// Skip long-running tests in short mode
	if testing.Short() {
		t.Skip("Skipping test suite in short mode")
	}

	suite.Run(t, new(ContainerManagerTestSuite))
}

// TestMain provides setup and teardown for the entire test package
func TestMain(m *testing.M) {
	// Setup code here if needed

	// Run tests
	code := m.Run()

	// Teardown code here if needed

	os.Exit(code)
}
