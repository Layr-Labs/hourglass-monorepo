package containerManager

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashAvsAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{
			name:    "standard address",
			address: "0x1234567890abcdef",
		},
		{
			name:    "empty address",
			address: "",
		},
		{
			name:    "long address",
			address: "0x1234567890abcdef1234567890abcdef12345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashAvsAddress(tt.address)
			assert.Len(t, result, 6)
			assert.NotEmpty(t, result)

			// Same input should produce same output
			hash2 := HashAvsAddress(tt.address)
			assert.Equal(t, result, hash2)

			// Verify it's a valid hex string
			assert.Regexp(t, "^[a-f0-9]{6}$", result)
		})
	}
}

func TestCreateDefaultContainerConfig(t *testing.T) {
	tests := []struct {
		name          string
		avsAddress    string
		imageRepo     string
		imageTag      string
		imageDigest   string
		containerPort int
		networkName   string
	}{
		{
			name:          "standard configuration",
			avsAddress:    "0x1234567890abcdef",
			imageRepo:     "myregistry/myapp",
			imageTag:      "v1.0.0",
			imageDigest:   "",
			containerPort: 8080,
			networkName:   "avs-network",
		},
		{
			name:          "different port",
			avsAddress:    "0xabcdef1234567890",
			imageRepo:     "myapp",
			imageTag:      "latest",
			imageDigest:   "",
			containerPort: 3000,
			networkName:   "custom-network",
		},
		{
			name:          "with image digest",
			avsAddress:    "0xfedcba9876543210",
			imageRepo:     "myregistry/myapp",
			imageTag:      "",
			imageDigest:   "sha256:abcdef1234567890",
			containerPort: 8080,
			networkName:   "avs-network",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CreateDefaultContainerConfig(
				tt.avsAddress,
				tt.imageRepo,
				tt.imageTag,
				tt.imageDigest,
				tt.containerPort,
				tt.networkName,
			)

			// Verify hostname format (should include hash and timestamp)
			expectedPrefix := "avs-performer-" + HashAvsAddress(tt.avsAddress) + "-"
			assert.True(t, strings.HasPrefix(config.Hostname, expectedPrefix),
				"Hostname should start with %s, got %s", expectedPrefix, config.Hostname)

			// Verify hostname includes timestamp (should be numeric suffix after the hash)
			assert.Regexp(t, `^avs-performer-[a-f0-9]{6}-\d+$`, config.Hostname,
				"Hostname should follow pattern 'avs-performer-{6-digit-hash}-{timestamp}'")

			// Verify image format
			var expectedImage string
			if tt.imageDigest != "" {
				expectedImage = tt.imageRepo + "@" + tt.imageDigest
			} else {
				expectedImage = tt.imageRepo + ":" + tt.imageTag
			}
			assert.Equal(t, expectedImage, config.Image)

			// Verify port configuration
			expectedPort := nat.Port(fmt.Sprintf("%d/tcp", tt.containerPort))
			assert.Contains(t, config.ExposedPorts, expectedPort)

			portBindings, exists := config.PortBindings[expectedPort]
			assert.True(t, exists)
			assert.Len(t, portBindings, 1)
			assert.Equal(t, "0.0.0.0", portBindings[0].HostIP)
			assert.Equal(t, "", portBindings[0].HostPort) // Random port assignment

			// Verify network configuration
			assert.Equal(t, tt.networkName, config.NetworkName)

			// Verify default settings
			assert.True(t, config.AutoRemove)
			assert.Equal(t, "no", config.RestartPolicy)
			assert.False(t, config.Privileged)
			assert.False(t, config.ReadOnly)
			assert.Equal(t, int64(0), config.MemoryLimit)
			assert.Equal(t, int64(0), config.CPUShares)
		})
	}
}

func TestGetContainerEndpoint(t *testing.T) {
	tests := []struct {
		name             string
		containerInfo    *ContainerInfo
		containerPort    int
		networkName      string
		expectedEndpoint string
		expectError      bool
	}{
		{
			name: "custom network endpoint",
			containerInfo: &ContainerInfo{
				Hostname: "test-container",
				Ports:    nat.PortMap{},
			},
			containerPort:    8080,
			networkName:      "custom-network",
			expectedEndpoint: "test-container:8080",
			expectError:      false,
		},
		{
			name: "bridge network with port mapping",
			containerInfo: &ContainerInfo{
				Hostname: "test-container",
				Ports: nat.PortMap{
					"8080/tcp": []nat.PortBinding{
						{HostIP: "0.0.0.0", HostPort: "32000"},
					},
				},
			},
			containerPort:    8080,
			networkName:      "", // Empty means bridge network
			expectedEndpoint: "localhost:32000",
			expectError:      false,
		},
		{
			name: "bridge network without port mapping",
			containerInfo: &ContainerInfo{
				Hostname: "test-container",
				Ports:    nat.PortMap{},
			},
			containerPort:    8080,
			networkName:      "", // Empty means bridge network
			expectedEndpoint: "",
			expectError:      true,
		},
		{
			name: "bridge network with wrong port",
			containerInfo: &ContainerInfo{
				Hostname: "test-container",
				Ports: nat.PortMap{
					"9090/tcp": []nat.PortBinding{
						{HostIP: "0.0.0.0", HostPort: "32000"},
					},
				},
			},
			containerPort:    8080,
			networkName:      "", // Empty means bridge network
			expectedEndpoint: "",
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := GetContainerEndpoint(tt.containerInfo, tt.containerPort, tt.networkName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, endpoint)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedEndpoint, endpoint)
			}
		})
	}
}

func TestNewDefaultAvsPerformerLivenessConfig(t *testing.T) {
	config := NewDefaultAvsPerformerLivenessConfig()

	// Verify the factory method returns a properly configured LivenessConfig
	require.NotNil(t, config)

	// Verify HealthCheckConfig
	assert.True(t, config.HealthCheckConfig.Enabled)
	assert.Equal(t, 5*time.Second, config.HealthCheckConfig.Interval)
	assert.Equal(t, 2*time.Second, config.HealthCheckConfig.Timeout)
	assert.Equal(t, 3, config.HealthCheckConfig.Retries)
	assert.Equal(t, 10*time.Second, config.HealthCheckConfig.StartPeriod)
	assert.Equal(t, 3, config.HealthCheckConfig.FailureThreshold)

	// Verify RestartPolicy
	assert.True(t, config.RestartPolicy.Enabled)
	assert.Equal(t, 5, config.RestartPolicy.MaxRestarts)
	assert.Equal(t, 2*time.Second, config.RestartPolicy.RestartDelay)
	assert.Equal(t, 2.0, config.RestartPolicy.BackoffMultiplier)
	assert.Equal(t, 30*time.Second, config.RestartPolicy.MaxBackoffDelay)
	assert.Equal(t, 60*time.Second, config.RestartPolicy.RestartTimeout)
	assert.True(t, config.RestartPolicy.RestartOnCrash)
	assert.True(t, config.RestartPolicy.RestartOnOOM)
	assert.True(t, config.RestartPolicy.RestartOnUnhealthy)

	// Verify ResourceThresholds
	assert.Equal(t, 90.0, config.ResourceThresholds.CPUThreshold)
	assert.Equal(t, 90.0, config.ResourceThresholds.MemoryThreshold)
	assert.False(t, config.ResourceThresholds.RestartOnCPU)
	assert.False(t, config.ResourceThresholds.RestartOnMemory)

	// Verify other settings
	assert.True(t, config.ResourceMonitoring)
	assert.Equal(t, 30*time.Second, config.ResourceCheckInterval)
}
