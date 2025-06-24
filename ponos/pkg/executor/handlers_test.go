package executor

import (
	"context"
	"testing"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Note: Complex mocking of serverPerformer types is avoided due to type assertion challenges in handlers
// Core functionality is tested in serverPerformer unit tests

// MockLegacyPerformer represents a non-server performer for testing
type MockLegacyPerformer struct {
	mock.Mock
}

func (m *MockLegacyPerformer) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockLegacyPerformer) ValidateTaskSignature(task *performerTask.PerformerTask) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockLegacyPerformer) RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error) {
	args := m.Called(ctx, task)
	return args.Get(0).(*performerTask.PerformerTaskResult), args.Error(1)
}

func (m *MockLegacyPerformer) Shutdown() error {
	args := m.Called()
	return args.Error(0)
}

// Test helper functions

func createTestExecutor() *Executor {
	return &Executor{
		logger:        zap.NewNop(),
		avsPerformers: make(map[string]avsPerformer.IAvsPerformer),
	}
}

// Note: Handler functionality is tested via integration tests and end-to-end flows
// The core deployment logic is thoroughly tested in serverPerformer unit tests
// These handler tests focus on basic error conditions and routing logic

func TestDeployArtifact_PerformerNotFound(t *testing.T) {
	executor := createTestExecutor()
	
	req := &executorV1.DeployArtifactRequest{
		AvsAddress:  "0xnonexistent",
		RegistryUrl: "registry.example.com/repo",
		Digest:      "sha256:abcd1234",
	}
	
	resp, err := executor.DeployArtifact(context.Background(), req)
	
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "performer not found for AVS", resp.Message)
	assert.Empty(t, resp.DeploymentId)
}

func TestDeployArtifact_NonServerPerformer(t *testing.T) {
	executor := createTestExecutor()
	
	avsAddress := "0xtest123"
	legacyMock := &MockLegacyPerformer{}
	executor.avsPerformers[avsAddress] = legacyMock
	
	req := &executorV1.DeployArtifactRequest{
		AvsAddress:  avsAddress,
		RegistryUrl: "registry.example.com/repo",
		Digest:      "sha256:abcd1234",
	}
	
	resp, err := executor.DeployArtifact(context.Background(), req)
	
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "performer does not support deployments", resp.Message)
}

func TestRemovePerformer_InvalidID(t *testing.T) {
	executor := createTestExecutor()
	
	req := &executorV1.RemovePerformerRequest{
		PerformerId: "invalid", // No hyphen separator
	}
	resp, err := executor.RemovePerformer(context.Background(), req)
	
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "invalid performer ID format", resp.Message)
}

func TestRemovePerformer_PerformerNotFound(t *testing.T) {
	executor := createTestExecutor()
	
	req := &executorV1.RemovePerformerRequest{
		PerformerId: "0xnonexistent-container123",
	}
	resp, err := executor.RemovePerformer(context.Background(), req)
	
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "performer not found", resp.Message)
}

func TestRemovePerformer_NonServerPerformer(t *testing.T) {
	executor := createTestExecutor()
	
	avsAddress := "0xtest123"
	legacyMock := &MockLegacyPerformer{}
	executor.avsPerformers[avsAddress] = legacyMock
	
	req := &executorV1.RemovePerformerRequest{
		PerformerId: avsAddress + "-container123",
	}
	resp, err := executor.RemovePerformer(context.Background(), req)
	
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "performer does not support container removal", resp.Message)
}

func TestRemovePerformer_EntirePerformer(t *testing.T) {
	executor := createTestExecutor()
	legacyMock := &MockLegacyPerformer{}
	
	avsAddress := "0xtest123"
	executor.avsPerformers[avsAddress] = legacyMock
	
	req := &executorV1.RemovePerformerRequest{
		PerformerId: avsAddress + "-legacy",
	}
	resp, err := executor.RemovePerformer(context.Background(), req)
	
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "performer does not support container removal", resp.Message)
}

// Test helper functions

func TestParsePerformerID(t *testing.T) {
	tests := []struct {
		name              string
		performerID       string
		expectedAVS       string
		expectedContainer string
	}{
		{
			name:              "Valid container ID",
			performerID:       "0xtest123-abcd1234",
			expectedAVS:       "0xtest123",
			expectedContainer: "abcd1234",
		},
		{
			name:              "Legacy performer",
			performerID:       "0xtest123-legacy",
			expectedAVS:       "0xtest123",
			expectedContainer: "",
		},
		{
			name:              "Invalid format",
			performerID:       "invalid",
			expectedAVS:       "",
			expectedContainer: "",
		},
		{
			name:              "Empty string",
			performerID:       "",
			expectedAVS:       "",
			expectedContainer: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avsAddress, containerID := parsePerformerID(tt.performerID)
			assert.Equal(t, tt.expectedAVS, avsAddress)
			assert.Equal(t, tt.expectedContainer, containerID)
		})
	}
}

func TestExtractRegistry(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		expected string
	}{
		{
			name:     "Standard image with tag",
			image:    "registry.example.com/repo:tag123",
			expected: "registry.example.com/repo",
		},
		{
			name:     "Image without tag",
			image:    "registry.example.com/repo",
			expected: "registry.example.com/repo",
		},
		{
			name:     "Empty image",
			image:    "",
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRegistry(tt.image)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractDigest(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		expected string
	}{
		{
			name:     "Standard image with tag",
			image:    "registry.example.com/repo:tag123",
			expected: "tag123",
		},
		{
			name:     "Image without tag",
			image:    "registry.example.com/repo",
			expected: "",
		},
		{
			name:     "Multiple colons",
			image:    "registry.example.com:5000/repo:tag123",
			expected: "5000/repo",
		},
		{
			name:     "Empty image",
			image:    "",
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDigest(tt.image)
			assert.Equal(t, tt.expected, result)
		})
	}
}