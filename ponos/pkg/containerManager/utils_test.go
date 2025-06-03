package containerManager

import (
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
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
		containerPort int
		networkName   string
	}{
		{
			name:          "standard configuration",
			avsAddress:    "0x1234567890abcdef",
			imageRepo:     "myregistry/myapp",
			imageTag:      "v1.0.0",
			containerPort: 8080,
			networkName:   "avs-network",
		},
		{
			name:          "different port",
			avsAddress:    "0xabcdef1234567890",
			imageRepo:     "myapp",
			imageTag:      "latest",
			containerPort: 3000,
			networkName:   "custom-network",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CreateDefaultContainerConfig(
				tt.avsAddress,
				tt.imageRepo,
				tt.imageTag,
				tt.containerPort,
				tt.networkName,
			)

			// Verify hostname format
			expectedHostname := "avs-performer-" + HashAvsAddress(tt.avsAddress)
			assert.Equal(t, expectedHostname, config.Hostname)

			// Verify image format
			expectedImage := tt.imageRepo + ":" + tt.imageTag
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
