// test_helpers.go contains test utilities and hand-written mocks for testing containerPerformer.
// These mocks provide custom behavior like event injection and call tracking that wouldn't
// be captured by generated mocks.

package avsContainerPerformer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	healthV1 "github.com/Layr-Labs/protocol-apis/gen/protos/grpc/health/v1"
	"google.golang.org/grpc"
)

// MockContainerManager provides a testable mock implementation of ContainerManager
type MockContainerManager struct {
	mu sync.Mutex

	// Event injection
	eventChannels map[string]chan containerManager.ContainerEvent

	// Tracking
	createCalls []CreateCall
	removeCalls []string

	// Behavior control
	createError  error
	removeError  error
	restartError error
}

type CreateCall struct {
	Config    *containerManager.ContainerConfig
	Timestamp time.Time
}

func NewMockContainerManager() *MockContainerManager {
	return &MockContainerManager{
		eventChannels: make(map[string]chan containerManager.ContainerEvent),
		createCalls:   []CreateCall{},
		removeCalls:   []string{},
	}
}

func (m *MockContainerManager) Create(ctx context.Context, config *containerManager.ContainerConfig) (*containerManager.ContainerInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.createCalls = append(m.createCalls, CreateCall{
		Config:    config,
		Timestamp: time.Now(),
	})

	if m.createError != nil {
		return nil, m.createError
	}

	return &containerManager.ContainerInfo{
		ID:       fmt.Sprintf("container-%d", len(m.createCalls)),
		Hostname: config.Hostname,
		Status:   "running",
	}, nil
}

func (m *MockContainerManager) StartLivenessMonitoring(ctx context.Context, containerID string, config *containerManager.LivenessConfig) (<-chan containerManager.ContainerEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	eventChan := make(chan containerManager.ContainerEvent, 100)
	m.eventChannels[containerID] = eventChan
	return eventChan, nil
}

func (m *MockContainerManager) InjectEvent(containerID string, event containerManager.ContainerEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ch, ok := m.eventChannels[containerID]; ok {
		select {
		case ch <- event:
			return nil
		default:
			return fmt.Errorf("event channel full")
		}
	}
	return fmt.Errorf("no event channel for container %s", containerID)
}

func (m *MockContainerManager) Remove(ctx context.Context, containerID string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.removeCalls = append(m.removeCalls, containerID)

	if m.removeError != nil {
		return m.removeError
	}
	return nil
}

func (m *MockContainerManager) GetCreateCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.createCalls)
}

func (m *MockContainerManager) GetRemoveCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.removeCalls)
}

func (m *MockContainerManager) WasContainerRemoved(containerID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range m.removeCalls {
		if id == containerID {
			return true
		}
	}
	return false
}

// Implement remaining ContainerManager interface methods with minimal implementations
func (m *MockContainerManager) Start(ctx context.Context, containerID string) error {
	return nil
}

func (m *MockContainerManager) Stop(ctx context.Context, containerID string, timeout time.Duration) error {
	return nil
}

func (m *MockContainerManager) Inspect(ctx context.Context, containerID string) (*containerManager.ContainerInfo, error) {
	return &containerManager.ContainerInfo{
		ID:       containerID,
		Hostname: "test-host",
		Status:   "running",
	}, nil
}

func (m *MockContainerManager) IsRunning(ctx context.Context, containerID string) (bool, error) {
	return true, nil
}

func (m *MockContainerManager) WaitForRunning(ctx context.Context, containerID string, timeout time.Duration) error {
	return nil
}

func (m *MockContainerManager) CreateNetworkIfNotExists(ctx context.Context, networkName string) error {
	return nil
}

func (m *MockContainerManager) RemoveNetwork(ctx context.Context, networkName string) error {
	return nil
}

func (m *MockContainerManager) StartHealthCheck(ctx context.Context, containerID string, config *containerManager.HealthCheckConfig) (<-chan bool, error) {
	return make(chan bool), nil
}

func (m *MockContainerManager) StopHealthCheck(containerID string) {}

func (m *MockContainerManager) StopLivenessMonitoring(containerID string) {}

func (m *MockContainerManager) GetContainerState(ctx context.Context, containerID string) (*containerManager.ContainerState, error) {
	return &containerManager.ContainerState{
		Status:    "running",
		ExitCode:  0,
		StartedAt: time.Now(),
	}, nil
}

func (m *MockContainerManager) RestartContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	return m.restartError
}

func (m *MockContainerManager) SetRestartPolicy(containerID string, policy containerManager.RestartPolicy) error {
	return nil
}

func (m *MockContainerManager) GetResourceUsage(ctx context.Context, containerID string) (*containerManager.ResourceUsage, error) {
	return &containerManager.ResourceUsage{
		CPUPercent:    10.0,
		MemoryUsage:   100000,
		MemoryLimit:   1000000,
		MemoryPercent: 10.0,
		Timestamp:     time.Now(),
	}, nil
}

func (m *MockContainerManager) TriggerRestart(containerID string, reason string) error {
	return m.restartError
}

func (m *MockContainerManager) Shutdown(ctx context.Context) error {
	return nil
}

// MockHealthClient provides a testable mock implementation for the health client
type MockHealthClient struct {
	mu sync.Mutex

	// Tracking
	healthChecks int

	// Behavior control
	healthCheckError error
}

func NewMockHealthClient() *MockHealthClient {
	return &MockHealthClient{}
}

func (m *MockHealthClient) Check(ctx context.Context, req *healthV1.HealthCheckRequest, opts ...grpc.CallOption) (*healthV1.HealthCheckResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.healthChecks++

	if m.healthCheckError != nil {
		return nil, m.healthCheckError
	}

	return &healthV1.HealthCheckResponse{
		Status: healthV1.HealthCheckResponse_SERVING,
	}, nil
}

func (m *MockHealthClient) Watch(ctx context.Context, req *healthV1.HealthCheckRequest, opts ...grpc.CallOption) (healthV1.Health_WatchClient, error) {
	return nil, fmt.Errorf("Watch not implemented in mock")
}

// MockPerformerServiceClient provides a testable mock implementation for the performer service client
type MockPerformerServiceClient struct {
	mu sync.Mutex

	// Tracking
	taskExecutions int

	// Behavior control
	taskError error
}

func NewMockPerformerServiceClient() *MockPerformerServiceClient {
	return &MockPerformerServiceClient{}
}

func (m *MockPerformerServiceClient) ExecuteTask(ctx context.Context, req *performerV1.TaskRequest, opts ...grpc.CallOption) (*performerV1.TaskResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.taskExecutions++

	if m.taskError != nil {
		return nil, m.taskError
	}

	return &performerV1.TaskResponse{
		TaskId: req.TaskId,
		Result: []byte("mock result"),
	}, nil
}

// StartSync is not used in our tests, so we stub it out
func (m *MockPerformerServiceClient) StartSync(ctx context.Context, req *performerV1.StartSyncRequest, opts ...grpc.CallOption) (*performerV1.StartSyncResponse, error) {
	return nil, fmt.Errorf("StartSync not implemented in mock")
}
