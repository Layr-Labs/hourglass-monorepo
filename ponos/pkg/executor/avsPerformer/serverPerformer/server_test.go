package serverPerformer

import (
	"context"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Mock implementations

type MockContainerManager struct {
	mock.Mock
}

func (m *MockContainerManager) Create(ctx context.Context, config *containerManager.ContainerConfig) (*containerManager.ContainerInfo, error) {
	args := m.Called(ctx, config)
	return args.Get(0).(*containerManager.ContainerInfo), args.Error(1)
}

func (m *MockContainerManager) Start(ctx context.Context, containerID string) error {
	args := m.Called(ctx, containerID)
	return args.Error(0)
}

func (m *MockContainerManager) Stop(ctx context.Context, containerID string, timeout time.Duration) error {
	args := m.Called(ctx, containerID, timeout)
	return args.Error(0)
}

func (m *MockContainerManager) Remove(ctx context.Context, containerID string, force bool) error {
	args := m.Called(ctx, containerID, force)
	return args.Error(0)
}

func (m *MockContainerManager) Inspect(ctx context.Context, containerID string) (*containerManager.ContainerInfo, error) {
	args := m.Called(ctx, containerID)
	return args.Get(0).(*containerManager.ContainerInfo), args.Error(1)
}

func (m *MockContainerManager) WaitForRunning(ctx context.Context, containerID string, timeout time.Duration) error {
	args := m.Called(ctx, containerID, timeout)
	return args.Error(0)
}

func (m *MockContainerManager) StartLivenessMonitoring(ctx context.Context, containerID string, config *containerManager.LivenessConfig) (<-chan containerManager.ContainerEvent, error) {
	args := m.Called(ctx, containerID, config)
	return args.Get(0).(<-chan containerManager.ContainerEvent), args.Error(1)
}

func (m *MockContainerManager) StopLivenessMonitoring(containerID string) {
	m.Called(containerID)
}

func (m *MockContainerManager) TriggerRestart(containerID string, reason string) error {
	args := m.Called(containerID, reason)
	return args.Error(0)
}

func (m *MockContainerManager) CreateNetworkIfNotExists(ctx context.Context, networkName string) error {
	args := m.Called(ctx, networkName)
	return args.Error(0)
}

func (m *MockContainerManager) GetContainerState(ctx context.Context, containerID string) (*containerManager.ContainerState, error) {
	args := m.Called(ctx, containerID)
	return args.Get(0).(*containerManager.ContainerState), args.Error(1)
}

func (m *MockContainerManager) GetResourceUsage(ctx context.Context, containerID string) (*containerManager.ResourceUsage, error) {
	args := m.Called(ctx, containerID)
	return args.Get(0).(*containerManager.ResourceUsage), args.Error(1)
}

func (m *MockContainerManager) IsRunning(ctx context.Context, containerID string) (bool, error) {
	args := m.Called(ctx, containerID)
	return args.Bool(0), args.Error(1)
}

func (m *MockContainerManager) RemoveNetwork(ctx context.Context, networkName string) error {
	args := m.Called(ctx, networkName)
	return args.Error(0)
}

func (m *MockContainerManager) StartHealthCheck(ctx context.Context, containerID string, config *containerManager.HealthCheckConfig) (<-chan bool, error) {
	args := m.Called(ctx, containerID, config)
	return args.Get(0).(<-chan bool), args.Error(1)
}

func (m *MockContainerManager) StopHealthCheck(containerID string) {
	m.Called(containerID)
}

func (m *MockContainerManager) RestartContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	args := m.Called(ctx, containerID, timeout)
	return args.Error(0)
}

func (m *MockContainerManager) SetRestartPolicy(containerID string, policy containerManager.RestartPolicy) error {
	args := m.Called(containerID, policy)
	return args.Error(0)
}

func (m *MockContainerManager) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockPeeringFetcher struct {
	mock.Mock
}

func (m *MockPeeringFetcher) ListAggregatorOperators(ctx context.Context, avsAddress string) ([]*peering.OperatorPeerInfo, error) {
	args := m.Called(ctx, avsAddress)
	return args.Get(0).([]*peering.OperatorPeerInfo), args.Error(1)
}

func (m *MockPeeringFetcher) ListExecutorOperators(ctx context.Context, avsAddress string) ([]*peering.OperatorPeerInfo, error) {
	args := m.Called(ctx, avsAddress)
	return args.Get(0).([]*peering.OperatorPeerInfo), args.Error(1)
}

type MockPerformerClient struct {
	mock.Mock
}

func (m *MockPerformerClient) ExecuteTask(ctx context.Context, req *performerV1.TaskRequest, opts ...grpc.CallOption) (*performerV1.TaskResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*performerV1.TaskResponse), args.Error(1)
}

func (m *MockPerformerClient) HealthCheck(ctx context.Context, req *performerV1.HealthCheckRequest, opts ...grpc.CallOption) (*performerV1.HealthCheckResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*performerV1.HealthCheckResponse), args.Error(1)
}

func (m *MockPerformerClient) StartSync(ctx context.Context, req *performerV1.StartSyncRequest, opts ...grpc.CallOption) (*performerV1.StartSyncResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*performerV1.StartSyncResponse), args.Error(1)
}

// Test helper functions

func createTestServer(t *testing.T) (*AvsPerformerServer, *MockContainerManager, *MockPeeringFetcher) {
	config := &avsPerformer.AvsPerformerConfig{
		AvsAddress: "0xtest",
		Image: avsPerformer.PerformerImage{
			Repository: "test-repo",
			Tag:        "test-tag",
		},
		PerformerNetworkName: "test-network",
	}
	
	logger := zap.NewNop()
	mockContainerMgr := &MockContainerManager{}
	mockPeeringFetcher := &MockPeeringFetcher{}
	
	server := &AvsPerformerServer{
		config:           config,
		logger:           logger,
		containerManager: mockContainerMgr,
		peeringFetcher:   mockPeeringFetcher,
		containers:       make(map[string]*ContainerMetadata),
	}
	
	return server, mockContainerMgr, mockPeeringFetcher
}

func createMockContainerInfo(id string) *containerManager.ContainerInfo {
	return &containerManager.ContainerInfo{
		ID:       id,
		Hostname: "test-" + id,
		Status:   "running",
	}
}

// Tests

func TestDeployNewPerformerVersion(t *testing.T) {
	server, mockContainerMgr, _ := createTestServer(t)
	ctx := context.Background()
	
	// Test data
	registryURL := "test-registry"
	digest := "test-digest"
	activationTime := time.Now().Unix() + 60 // 1 minute in future
	
	containerInfo := createMockContainerInfo("container-123")
	
	// Mock expectations
	mockContainerMgr.On("Create", mock.Anything, mock.Anything).Return(containerInfo, nil)
	mockContainerMgr.On("Start", mock.Anything, "container-123").Return(nil)
	mockContainerMgr.On("WaitForRunning", mock.Anything, "container-123", mock.Anything).Return(nil)
	mockContainerMgr.On("Inspect", mock.Anything, "container-123").Return(containerInfo, nil)
	
	// Execute
	deploymentID, err := server.DeployNewPerformerVersion(ctx, registryURL, digest, activationTime)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, deploymentID)
	
	// Verify container was added to store
	server.containerMu.RLock()
	assert.Len(t, server.containers, 1)
	container := server.containers["container-123"]
	assert.NotNil(t, container)
	assert.Equal(t, ContainerStatusPending, container.Status)
	assert.Equal(t, registryURL, container.RegistryURL)
	assert.Equal(t, digest, container.ArtifactDigest)
	assert.Equal(t, activationTime, container.ActivationTime)
	server.containerMu.RUnlock()
	
	mockContainerMgr.AssertExpectations(t)
}

func TestDeployNewPerformerVersion_PastActivationTime(t *testing.T) {
	server, _, _ := createTestServer(t)
	ctx := context.Background()
	
	// Test with past activation time
	pastTime := time.Now().Unix() - 60 // 1 minute in past
	
	deploymentID, err := server.DeployNewPerformerVersion(ctx, "registry", "digest", pastTime)
	
	assert.Error(t, err)
	assert.Empty(t, deploymentID)
	assert.Contains(t, err.Error(), "activation time must be in the future")
}

func TestCheckAndActivatePendingContainer(t *testing.T) {
	server, _, _ := createTestServer(t)
	
	// Add a pending container with past activation time
	containerInfo := createMockContainerInfo("container-123")
	mockClient := &MockPerformerClient{}
	
	pastTime := time.Now().Unix() - 30 // 30 seconds ago
	pendingContainer := &ContainerMetadata{
		Info:           containerInfo,
		ActivationTime: pastTime,
		Status:         ContainerStatusPending,
		Client:         mockClient,
		DeploymentID:   "test-deployment",
	}
	
	server.containers["container-123"] = pendingContainer
	
	// Execute lazy activation
	server.checkAndActivatePendingContainer(time.Now().Unix())
	
	// Verify container was activated
	assert.Equal(t, "container-123", server.currentContainer)
	assert.Equal(t, ContainerStatusActive, pendingContainer.Status)
	assert.NotNil(t, pendingContainer.ActivatedAt)
}

func TestCheckAndActivatePendingContainer_FutureActivation(t *testing.T) {
	server, _, _ := createTestServer(t)
	
	// Add a pending container with future activation time
	containerInfo := createMockContainerInfo("container-123")
	
	futureTime := time.Now().Unix() + 60 // 1 minute in future
	pendingContainer := &ContainerMetadata{
		Info:           containerInfo,
		ActivationTime: futureTime,
		Status:         ContainerStatusPending,
		DeploymentID:   "test-deployment",
	}
	
	server.containers["container-123"] = pendingContainer
	
	// Execute lazy activation
	server.checkAndActivatePendingContainer(time.Now().Unix())
	
	// Verify container was NOT activated
	assert.Empty(t, server.currentContainer)
	assert.Equal(t, ContainerStatusPending, pendingContainer.Status)
	assert.Nil(t, pendingContainer.ActivatedAt)
}

func TestActivateContainer(t *testing.T) {
	server, _, _ := createTestServer(t)
	
	// Set up existing active container
	oldContainerInfo := createMockContainerInfo("old-container")
	oldContainer := &ContainerMetadata{
		Info:   oldContainerInfo,
		Status: ContainerStatusActive,
	}
	server.containers["old-container"] = oldContainer
	server.currentContainer = "old-container"
	
	// Set up new pending container
	newContainerInfo := createMockContainerInfo("new-container")
	mockClient := &MockPerformerClient{}
	newContainer := &ContainerMetadata{
		Info:         newContainerInfo,
		Status:       ContainerStatusPending,
		Client:       mockClient,
		DeploymentID: "test-deployment",
	}
	server.containers["new-container"] = newContainer
	
	// Execute activation
	err := server.activateContainer("new-container")
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "new-container", server.currentContainer)
	assert.Equal(t, ContainerStatusActive, newContainer.Status)
	assert.Equal(t, ContainerStatusExpired, oldContainer.Status)
	assert.NotNil(t, newContainer.ActivatedAt)
	
	// Verify legacy fields updated
	assert.Equal(t, newContainerInfo, server.containerInfo)
	assert.Equal(t, mockClient, server.performerClient)
}

func TestActivateContainer_NonExistentContainer(t *testing.T) {
	server, _, _ := createTestServer(t)
	
	err := server.activateContainer("non-existent")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container non-existent not found")
}

func TestActivateContainer_NotPending(t *testing.T) {
	server, _, _ := createTestServer(t)
	
	// Add active container (not pending)
	containerInfo := createMockContainerInfo("container-123")
	container := &ContainerMetadata{
		Info:   containerInfo,
		Status: ContainerStatusActive,
	}
	server.containers["container-123"] = container
	
	err := server.activateContainer("container-123")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not pending")
}

func TestGetActiveContainer(t *testing.T) {
	server, _, _ := createTestServer(t)
	
	// Test with no active container
	activeContainer := server.getActiveContainer()
	assert.Nil(t, activeContainer)
	
	// Set up active container
	containerInfo := createMockContainerInfo("container-123")
	container := &ContainerMetadata{
		Info:   containerInfo,
		Status: ContainerStatusActive,
	}
	server.containers["container-123"] = container
	server.currentContainer = "container-123"
	
	// Test with active container
	activeContainer = server.getActiveContainer()
	assert.NotNil(t, activeContainer)
	assert.Equal(t, "container-123", activeContainer.Info.ID)
}

func TestGetAllContainerHealth(t *testing.T) {
	server, mockContainerMgr, _ := createTestServer(t)
	ctx := context.Background()
	
	// Set up test containers
	container1 := &ContainerMetadata{
		Info: createMockContainerInfo("container-1"),
		Image: avsPerformer.PerformerImage{
			Repository: "repo1",
			Tag:        "tag1",
		},
		Status:         ContainerStatusActive,
		ActivationTime: time.Now().Unix(),
	}
	
	container2 := &ContainerMetadata{
		Info: createMockContainerInfo("container-2"),
		Image: avsPerformer.PerformerImage{
			Repository: "repo2",
			Tag:        "tag2",
		},
		Status:         ContainerStatusPending,
		ActivationTime: time.Now().Unix() + 60,
	}
	
	server.containers["container-1"] = container1
	server.containers["container-2"] = container2
	
	// Mock container inspection
	mockContainerMgr.On("Inspect", mock.Anything, "container-1").Return(
		&containerManager.ContainerInfo{Status: "running"}, nil)
	mockContainerMgr.On("Inspect", mock.Anything, "container-2").Return(
		&containerManager.ContainerInfo{Status: "running"}, nil)
	
	// Execute
	statuses, err := server.GetAllContainerHealth(ctx)
	
	// Assertions
	assert.NoError(t, err)
	assert.Len(t, statuses, 2)
	
	// Verify status details
	for _, status := range statuses {
		assert.NotEmpty(t, status.ContainerID)
		assert.NotEmpty(t, status.Image)
		assert.Contains(t, []string{"Active", "Pending"}, status.Status.String())
		assert.Equal(t, ContainerHealthHealthy, status.ContainerHealth)
	}
	
	mockContainerMgr.AssertExpectations(t)
}

func TestIsContainerExpired(t *testing.T) {
	server, _, _ := createTestServer(t)
	currentTime := time.Now().Unix()
	
	tests := []struct {
		name     string
		metadata *ContainerMetadata
		expected bool
	}{
		{
			name: "Active container",
			metadata: &ContainerMetadata{
				Status: ContainerStatusActive,
			},
			expected: false,
		},
		{
			name: "Pending container within grace period",
			metadata: &ContainerMetadata{
				Status:         ContainerStatusPending,
				ActivationTime: currentTime - 1800, // 30 minutes ago
			},
			expected: false,
		},
		{
			name: "Pending container past grace period",
			metadata: &ContainerMetadata{
				Status:         ContainerStatusPending,
				ActivationTime: currentTime - 7200, // 2 hours ago
			},
			expected: true,
		},
		{
			name: "Expired container",
			metadata: &ContainerMetadata{
				Status: ContainerStatusExpired,
			},
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.isContainerExpired(tt.metadata, currentTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReapExpiredContainers(t *testing.T) {
	server, mockContainerMgr, _ := createTestServer(t)
	currentTime := time.Now().Unix()
	
	// Set up containers
	activeContainer := &ContainerMetadata{
		Info:   createMockContainerInfo("active"),
		Status: ContainerStatusActive,
	}
	
	expiredContainer := &ContainerMetadata{
		Info:           createMockContainerInfo("expired"),
		Status:         ContainerStatusPending,
		ActivationTime: currentTime - 7200, // 2 hours ago
		DeploymentID:   "expired-deployment",
	}
	
	server.containers["active"] = activeContainer
	server.containers["expired"] = expiredContainer
	server.currentContainer = "active"
	
	// Mock expectations for removal
	mockContainerMgr.On("Stop", mock.Anything, "expired", mock.Anything).Return(nil)
	mockContainerMgr.On("Remove", mock.Anything, "expired", true).Return(nil)
	
	// Execute reaping
	server.reapExpiredContainers()
	
	// Verify expired container was removed
	server.containerMu.RLock()
	assert.Len(t, server.containers, 1)
	assert.Contains(t, server.containers, "active")
	assert.NotContains(t, server.containers, "expired")
	server.containerMu.RUnlock()
	
	mockContainerMgr.AssertExpectations(t)
}

func TestRemoveContainer(t *testing.T) {
	server, mockContainerMgr, _ := createTestServer(t)
	
	// Set up containers
	activeContainer := &ContainerMetadata{
		Info:   createMockContainerInfo("active"),
		Status: ContainerStatusActive,
	}
	
	pendingContainer := &ContainerMetadata{
		Info:         createMockContainerInfo("pending"),
		Status:       ContainerStatusPending,
		DeploymentID: "test-deployment",
	}
	
	server.containers["active"] = activeContainer
	server.containers["pending"] = pendingContainer
	server.currentContainer = "active"
	
	// Test removing active container (should fail)
	err := server.RemoveContainer("active")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove active container")
	
	// Test removing non-existent container
	err = server.RemoveContainer("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container non-existent not found")
	
	// Mock expectations for successful removal
	mockContainerMgr.On("Stop", mock.Anything, "pending", mock.Anything).Return(nil)
	mockContainerMgr.On("Remove", mock.Anything, "pending", true).Return(nil)
	
	// Test successful removal of pending container
	err = server.RemoveContainer("pending")
	assert.NoError(t, err)
	
	// Verify container was removed
	server.containerMu.RLock()
	assert.NotContains(t, server.containers, "pending")
	server.containerMu.RUnlock()
	
	mockContainerMgr.AssertExpectations(t)
}

func TestGetContainerHealth(t *testing.T) {
	server, mockContainerMgr, _ := createTestServer(t)
	
	tests := []struct {
		name           string
		containerStatus string
		expectedHealth ContainerHealthState
	}{
		{
			name:           "Running container",
			containerStatus: "running",
			expectedHealth: ContainerHealthHealthy,
		},
		{
			name:           "Exited container",
			containerStatus: "exited",
			expectedHealth: ContainerHealthCrashed,
		},
		{
			name:           "Dead container",
			containerStatus: "dead",
			expectedHealth: ContainerHealthCrashed,
		},
		{
			name:           "Paused container",
			containerStatus: "paused",
			expectedHealth: ContainerHealthUnhealthy,
		},
		{
			name:           "Unknown status",
			containerStatus: "unknown",
			expectedHealth: ContainerHealthUnknown,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containerInfo := &containerManager.ContainerInfo{
				Status: tt.containerStatus,
			}
			
			mockContainerMgr.On("Inspect", mock.Anything, "test-container").Return(containerInfo, nil).Once()
			
			health := server.getContainerHealth("test-container")
			assert.Equal(t, tt.expectedHealth, health)
		})
	}
	
	mockContainerMgr.AssertExpectations(t)
}

func TestGetApplicationHealth(t *testing.T) {
	server, _, _ := createTestServer(t)
	ctx := context.Background()
	
	// Test with nil client
	health := server.getApplicationHealth(ctx, nil)
	assert.Equal(t, AppHealthNotReady, health)
	
	// Test with healthy client
	mockClient := &MockPerformerClient{}
	mockClient.On("HealthCheck", mock.Anything, mock.Anything).Return(
		&performerV1.HealthCheckResponse{}, nil)
	
	health = server.getApplicationHealth(ctx, mockClient)
	assert.Equal(t, AppHealthHealthy, health)
	
	// Test with unhealthy client (error response)
	mockClientUnhealthy := &MockPerformerClient{}
	mockClientUnhealthy.On("HealthCheck", mock.Anything, mock.Anything).Return(
		(*performerV1.HealthCheckResponse)(nil), assert.AnError)
	
	health = server.getApplicationHealth(ctx, mockClientUnhealthy)
	assert.Equal(t, AppHealthUnhealthy, health)
	
	mockClient.AssertExpectations(t)
	mockClientUnhealthy.AssertExpectations(t)
}