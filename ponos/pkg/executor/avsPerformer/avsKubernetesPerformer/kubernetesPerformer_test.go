package avsKubernetesPerformer

import (
	"context"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/kubernetesManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// Mock implementations for testing - these implement the same methods as the real types
type MockCRDOperations struct {
	mock.Mock
}

func (m *MockCRDOperations) CreatePerformer(ctx context.Context, req *kubernetesManager.CreatePerformerRequest) (*kubernetesManager.CreatePerformerResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kubernetesManager.CreatePerformerResponse), args.Error(1)
}

func (m *MockCRDOperations) GetPerformer(ctx context.Context, name string) (*kubernetesManager.PerformerCRD, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kubernetesManager.PerformerCRD), args.Error(1)
}

func (m *MockCRDOperations) UpdatePerformer(ctx context.Context, req *kubernetesManager.UpdatePerformerRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockCRDOperations) DeletePerformer(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockCRDOperations) ListPerformers(ctx context.Context) ([]kubernetesManager.PerformerInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]kubernetesManager.PerformerInfo), args.Error(1)
}

func (m *MockCRDOperations) GetPerformerStatus(ctx context.Context, name string) (*kubernetesManager.PerformerStatus, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kubernetesManager.PerformerStatus), args.Error(1)
}

type MockClientWrapper struct {
	mock.Mock
}

func (m *MockClientWrapper) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientWrapper) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockPeeringFetcher struct {
	mock.Mock
}

func (m *MockPeeringFetcher) ListAggregatorOperators(ctx context.Context, avsAddress string) ([]*peering.OperatorPeerInfo, error) {
	args := m.Called(ctx, avsAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*peering.OperatorPeerInfo), args.Error(1)
}

func (m *MockPeeringFetcher) ListExecutorOperators(ctx context.Context, avsAddress string) ([]*peering.OperatorPeerInfo, error) {
	args := m.Called(ctx, avsAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*peering.OperatorPeerInfo), args.Error(1)
}

// Helper function to create a test AvsKubernetesPerformer with mocked dependencies
func createTestKubernetesPerformer(t *testing.T) (*AvsKubernetesPerformer, *MockCRDOperations, *MockClientWrapper, *MockPeeringFetcher) {
	mockCRDOps := &MockCRDOperations{}
	mockClientWrapper := &MockClientWrapper{}
	mockPeeringFetcher := &MockPeeringFetcher{}

	config := &avsPerformer.AvsPerformerConfig{
		AvsAddress:  "0x123",
		ProcessType: avsPerformer.AvsProcessTypeServer,
	}

	kubernetesConfig := &kubernetesManager.Config{
		Namespace:         "test-namespace",
		OperatorNamespace: "hourglass-system",
		CRDGroup:          "hourglass.eigenlayer.io",
		CRDVersion:        "v1alpha1",
		ConnectionTimeout: 30 * time.Second,
	}

	logger := zaptest.NewLogger(t)

	performer := &AvsKubernetesPerformer{
		config:              config,
		kubernetesConfig:    kubernetesConfig,
		logger:              logger,
		peeringFetcher:      mockPeeringFetcher,
		performerTaskStates: make(map[string]*PerformerTaskState),
	}

	return performer, mockCRDOps, mockClientWrapper, mockPeeringFetcher
}

func TestNewAvsKubernetesPerformer(t *testing.T) {
	config := &avsPerformer.AvsPerformerConfig{
		AvsAddress:  "0x123",
		ProcessType: avsPerformer.AvsProcessTypeServer,
	}

	kubernetesConfig := &kubernetesManager.Config{
		Namespace:         "test-namespace",
		CRDGroup:          "hourglass.eigenlayer.io",
		CRDVersion:        "v1alpha1",
		ConnectionTimeout: 30 * time.Second,
	}

	logger := zaptest.NewLogger(t)

	// Create performer - it may or may not fail depending on the test environment
	performer, err := NewAvsKubernetesPerformer(config, kubernetesConfig, nil, nil, logger)

	// In most test environments without a real k8s cluster, this will fail
	// In environments with kind or a real cluster, it might succeed
	// Either way is fine for this test
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create Kubernetes client")
	} else {
		// If it succeeds, just verify the performer was created
		assert.NotNil(t, performer)
		assert.Equal(t, config, performer.config)
		assert.Equal(t, kubernetesConfig, performer.kubernetesConfig)
	}
}

func TestAvsKubernetesPerformer_GeneratePerformerID(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	id1 := performer.generatePerformerID()
	id2 := performer.generatePerformerID()

	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "performer-7f8a79-")
	assert.Contains(t, id2, "performer-7f8a79-")
}

func TestAvsKubernetesPerformer_ConvertPerformerResource(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	resource := &PerformerResource{
		performerID: "test-performer",
		avsAddress:  "0x123",
		image: avsPerformer.PerformerImage{
			Repository: "test-repo",
			Tag:        "v1.0.0",
			Digest:     "sha256:abc123",
		},
		status: avsPerformer.PerformerResourceStatusInService,
	}

	metadata := performer.convertPerformerResource(resource)

	assert.Equal(t, "test-performer", metadata.PerformerID)
	assert.Equal(t, "0x123", metadata.AvsAddress)
	assert.Equal(t, avsPerformer.PerformerResourceStatusInService, metadata.Status)
	assert.Equal(t, "test-repo", metadata.ArtifactRegistry)
	assert.Equal(t, "v1.0.0", metadata.ArtifactTag)
	assert.Equal(t, "sha256:abc123", metadata.ArtifactDigest)
	assert.True(t, metadata.ContainerHealthy)
	assert.True(t, metadata.ApplicationHealthy)
	// LastHealthCheck is set to time.Now() in the converter, so just check it's recent
	assert.WithinDuration(t, time.Now(), metadata.LastHealthCheck, 1*time.Second)
	assert.Equal(t, "test-performer", metadata.ResourceID)
}

func TestAvsKubernetesPerformer_TryAddTask(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	// Test adding task for new performer (should succeed and create active state)
	success := performer.tryAddTask("performer-1")
	assert.True(t, success)

	// Verify state was created
	performer.performerStatesMu.Lock()
	state, exists := performer.performerTaskStates["performer-1"]
	assert.True(t, exists)
	assert.Equal(t, PerformerStateActive, state.state)
	performer.performerStatesMu.Unlock()

	// Test adding another task for same performer (should succeed)
	success = performer.tryAddTask("performer-1")
	assert.True(t, success)

	// Test adding task for different performer
	success = performer.tryAddTask("performer-2")
	assert.True(t, success)
}

func TestAvsKubernetesPerformer_TaskStateLifecycle(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	// Add a task to create state
	success := performer.tryAddTask("performer-1")
	assert.True(t, success)

	// Mark as draining - should prevent new tasks
	performer.performerStatesMu.Lock()
	state := performer.performerTaskStates["performer-1"]
	state.state = PerformerStateDraining
	performer.performerStatesMu.Unlock()

	// Try to add task to draining performer (should fail)
	success = performer.tryAddTask("performer-1")
	assert.False(t, success)

	// Complete the task
	performer.taskCompleted("performer-1")

	// Cleanup the performer state
	performer.cleanupTaskWaitGroup("performer-1")

	// Verify it's cleaned up
	performer.performerStatesMu.Lock()
	_, exists := performer.performerTaskStates["performer-1"]
	assert.False(t, exists)
	performer.performerStatesMu.Unlock()
}

func TestAvsKubernetesPerformer_ListPerformers_Empty(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	// Test with no performers
	performers := performer.ListPerformers()
	assert.Empty(t, performers)
}

func TestAvsKubernetesPerformer_ListPerformers_WithPerformers(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	// Add a current performer
	currentPerformer := &PerformerResource{
		performerID: "current-performer",
		avsAddress:  "0x123",
		image: avsPerformer.PerformerImage{
			Repository: "test-repo",
			Tag:        "v1.0.0",
		},
		status: avsPerformer.PerformerResourceStatusInService,
	}
	performer.currentPerformer.Store(currentPerformer)

	// Add a next performer
	performer.nextPerformer = &PerformerResource{
		performerID: "next-performer",
		avsAddress:  "0x123",
		image: avsPerformer.PerformerImage{
			Repository: "test-repo",
			Tag:        "v2.0.0",
		},
		status: avsPerformer.PerformerResourceStatusStaged,
	}

	performers := performer.ListPerformers()
	assert.Len(t, performers, 2)

	// Verify current performer
	var currentMeta, nextMeta *avsPerformer.PerformerMetadata
	for _, p := range performers {
		if p.Status == avsPerformer.PerformerResourceStatusInService {
			currentMeta = &p
		} else if p.Status == avsPerformer.PerformerResourceStatusStaged {
			nextMeta = &p
		}
	}

	require.NotNil(t, currentMeta)
	assert.Equal(t, "current-performer", currentMeta.PerformerID)
	assert.Equal(t, "v1.0.0", currentMeta.ArtifactTag)

	require.NotNil(t, nextMeta)
	assert.Equal(t, "next-performer", nextMeta.PerformerID)
	assert.Equal(t, "v2.0.0", nextMeta.ArtifactTag)
}

func TestAvsKubernetesPerformer_PromotePerformer_Success(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	ctx := context.Background()

	// Set up a healthy next performer
	performer.nextPerformer = &PerformerResource{
		performerID: "test-performer",
		status:      avsPerformer.PerformerResourceStatusStaged,
	}

	err := performer.PromotePerformer(ctx, "test-performer")
	assert.NoError(t, err)

	// Verify promotion
	current := performer.currentPerformer.Load()
	assert.NotNil(t, current)
	currentPerformer := current.(*PerformerResource)
	assert.Equal(t, "test-performer", currentPerformer.performerID)
	assert.Equal(t, avsPerformer.PerformerResourceStatusInService, currentPerformer.status)
	assert.Nil(t, performer.nextPerformer)
}

func TestAvsKubernetesPerformer_PromotePerformer_NotInNextSlot(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	ctx := context.Background()

	err := performer.PromotePerformer(ctx, "non-existent-performer")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "performer non-existent-performer is not in the next deployment slot")
}

func TestAvsKubernetesPerformer_PromotePerformer_AlreadyCurrent(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	ctx := context.Background()

	// Set up a current performer
	currentPerformer := &PerformerResource{
		performerID: "test-performer",
	}
	performer.currentPerformer.Store(currentPerformer)

	err := performer.PromotePerformer(ctx, "test-performer")
	assert.NoError(t, err) // Should be no-op success
}

func TestAvsKubernetesPerformer_RunTask_NoCurrentPerformer(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	ctx := context.Background()

	task := &performerTask.PerformerTask{
		TaskID:            "test-task",
		Payload:           []byte("test-payload"),
		AggregatorAddress: "0xaggregator",
		Signature:         []byte("test-signature"),
	}

	result, err := performer.RunTask(ctx, task)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no current performer available")
}

// Additional test for edge cases
func TestAvsKubernetesPerformer_EdgeCases(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	// Test wait for task completion on non-existent performer
	performer.waitForTaskCompletion("non-existent")
	// Should not panic or block

	// Test cleanup on non-existent performer
	performer.cleanupTaskWaitGroup("non-existent")
	// Should not panic
}

// Tests for environment variable building functionality
func TestAvsKubernetesPerformer_BuildEnvironmentFromImage_Empty(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	image := avsPerformer.PerformerImage{
		Repository: "test-repo",
		Tag:        "v1.0.0",
		Envs:       []config.AVSPerformerEnv{},
	}

	envMap, envVarSources := performer.buildEnvironmentFromImage(image)

	// Should have default environment variables
	assert.Len(t, envMap, 2)
	assert.Equal(t, "0x123", envMap["AVS_ADDRESS"])
	assert.Equal(t, "8080", envMap["GRPC_PORT"])

	// Should have no environment variable sources
	assert.Empty(t, envVarSources)
}

func TestAvsKubernetesPerformer_BuildEnvironmentFromImage_DirectValues(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	image := avsPerformer.PerformerImage{
		Repository: "test-repo",
		Tag:        "v1.0.0",
		Envs: []config.AVSPerformerEnv{
			{
				Name:  "DATABASE_URL",
				Value: "postgres://localhost/db",
			},
			{
				Name:  "API_ENDPOINT",
				Value: "https://api.example.com",
			},
		},
	}

	envMap, envVarSources := performer.buildEnvironmentFromImage(image)

	// Should have default + custom environment variables
	assert.Len(t, envMap, 4)
	assert.Equal(t, "0x123", envMap["AVS_ADDRESS"])
	assert.Equal(t, "8080", envMap["GRPC_PORT"])
	assert.Equal(t, "postgres://localhost/db", envMap["DATABASE_URL"])
	assert.Equal(t, "https://api.example.com", envMap["API_ENDPOINT"])

	// Should have no environment variable sources
	assert.Empty(t, envVarSources)
}

func TestAvsKubernetesPerformer_BuildEnvironmentFromImage_SecretRef(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	image := avsPerformer.PerformerImage{
		Repository: "test-repo",
		Tag:        "v1.0.0",
		Envs: []config.AVSPerformerEnv{
			{
				Name: "API_KEY",
				KubernetesEnv: &config.KubernetesEnv{
					ValueFrom: struct {
						SecretKeyRef struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						} `json:"secretKeyRef" yaml:"secretKeyRef"`
						ConfigMapKeyRef struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						} `json:"configMapKeyRef" yaml:"configMapKeyRef"`
					}{
						SecretKeyRef: struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						}{
							Name: "api-secrets",
							Key:  "api-key",
						},
					},
				},
			},
		},
	}

	envMap, envVarSources := performer.buildEnvironmentFromImage(image)

	// Should have only default environment variables
	assert.Len(t, envMap, 2)
	assert.Equal(t, "0x123", envMap["AVS_ADDRESS"])
	assert.Equal(t, "8080", envMap["GRPC_PORT"])

	// Should have one environment variable source
	require.Len(t, envVarSources, 1)
	assert.Equal(t, "API_KEY", envVarSources[0].Name)
	assert.NotNil(t, envVarSources[0].ValueFrom)
	assert.NotNil(t, envVarSources[0].ValueFrom.SecretKeyRef)
	assert.Equal(t, "api-secrets", envVarSources[0].ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "api-key", envVarSources[0].ValueFrom.SecretKeyRef.Key)
	assert.Nil(t, envVarSources[0].ValueFrom.ConfigMapKeyRef)
}

func TestAvsKubernetesPerformer_BuildEnvironmentFromImage_ConfigMapRef(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	image := avsPerformer.PerformerImage{
		Repository: "test-repo",
		Tag:        "v1.0.0",
		Envs: []config.AVSPerformerEnv{
			{
				Name: "APP_CONFIG",
				KubernetesEnv: &config.KubernetesEnv{
					ValueFrom: struct {
						SecretKeyRef struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						} `json:"secretKeyRef" yaml:"secretKeyRef"`
						ConfigMapKeyRef struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						} `json:"configMapKeyRef" yaml:"configMapKeyRef"`
					}{
						ConfigMapKeyRef: struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						}{
							Name: "app-config",
							Key:  "config.json",
						},
					},
				},
			},
		},
	}

	envMap, envVarSources := performer.buildEnvironmentFromImage(image)

	// Should have only default environment variables
	assert.Len(t, envMap, 2)
	assert.Equal(t, "0x123", envMap["AVS_ADDRESS"])
	assert.Equal(t, "8080", envMap["GRPC_PORT"])

	// Should have one environment variable source
	require.Len(t, envVarSources, 1)
	assert.Equal(t, "APP_CONFIG", envVarSources[0].Name)
	assert.NotNil(t, envVarSources[0].ValueFrom)
	assert.NotNil(t, envVarSources[0].ValueFrom.ConfigMapKeyRef)
	assert.Equal(t, "app-config", envVarSources[0].ValueFrom.ConfigMapKeyRef.Name)
	assert.Equal(t, "config.json", envVarSources[0].ValueFrom.ConfigMapKeyRef.Key)
	assert.Nil(t, envVarSources[0].ValueFrom.SecretKeyRef)
}

func TestAvsKubernetesPerformer_BuildEnvironmentFromImage_Mixed(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	image := avsPerformer.PerformerImage{
		Repository: "test-repo",
		Tag:        "v1.0.0",
		Envs: []config.AVSPerformerEnv{
			{
				Name:  "DATABASE_URL",
				Value: "postgres://localhost/db",
			},
			{
				Name: "API_KEY",
				KubernetesEnv: &config.KubernetesEnv{
					ValueFrom: struct {
						SecretKeyRef struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						} `json:"secretKeyRef" yaml:"secretKeyRef"`
						ConfigMapKeyRef struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						} `json:"configMapKeyRef" yaml:"configMapKeyRef"`
					}{
						SecretKeyRef: struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						}{
							Name: "api-secrets",
							Key:  "api-key",
						},
					},
				},
			},
			{
				Name: "APP_CONFIG",
				KubernetesEnv: &config.KubernetesEnv{
					ValueFrom: struct {
						SecretKeyRef struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						} `json:"secretKeyRef" yaml:"secretKeyRef"`
						ConfigMapKeyRef struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						} `json:"configMapKeyRef" yaml:"configMapKeyRef"`
					}{
						ConfigMapKeyRef: struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						}{
							Name: "app-config",
							Key:  "config.json",
						},
					},
				},
			},
			{
				Name:  "LOG_LEVEL",
				Value: "debug",
			},
		},
	}

	envMap, envVarSources := performer.buildEnvironmentFromImage(image)

	// Should have default + direct value environment variables
	assert.Len(t, envMap, 4)
	assert.Equal(t, "0x123", envMap["AVS_ADDRESS"])
	assert.Equal(t, "8080", envMap["GRPC_PORT"])
	assert.Equal(t, "postgres://localhost/db", envMap["DATABASE_URL"])
	assert.Equal(t, "debug", envMap["LOG_LEVEL"])

	// Should have two environment variable sources
	require.Len(t, envVarSources, 2)

	// Find and verify secret ref
	var secretRef *kubernetesManager.EnvVarSource
	var configMapRef *kubernetesManager.EnvVarSource
	for i := range envVarSources {
		if envVarSources[i].Name == "API_KEY" {
			secretRef = &envVarSources[i]
		} else if envVarSources[i].Name == "APP_CONFIG" {
			configMapRef = &envVarSources[i]
		}
	}

	require.NotNil(t, secretRef)
	assert.NotNil(t, secretRef.ValueFrom.SecretKeyRef)
	assert.Equal(t, "api-secrets", secretRef.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "api-key", secretRef.ValueFrom.SecretKeyRef.Key)

	require.NotNil(t, configMapRef)
	assert.NotNil(t, configMapRef.ValueFrom.ConfigMapKeyRef)
	assert.Equal(t, "app-config", configMapRef.ValueFrom.ConfigMapKeyRef.Name)
	assert.Equal(t, "config.json", configMapRef.ValueFrom.ConfigMapKeyRef.Key)
}

func TestAvsKubernetesPerformer_BuildEnvironmentFromImage_SkipEmpty(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	image := avsPerformer.PerformerImage{
		Repository: "test-repo",
		Tag:        "v1.0.0",
		Envs: []config.AVSPerformerEnv{
			{
				Name: "EMPTY_VAR",
				// No Value, ValueFromEnv, or KubernetesEnv set
			},
			{
				Name:  "VALID_VAR",
				Value: "valid-value",
			},
		},
	}

	envMap, envVarSources := performer.buildEnvironmentFromImage(image)

	// Should have default + valid environment variable only
	assert.Len(t, envMap, 3)
	assert.Equal(t, "0x123", envMap["AVS_ADDRESS"])
	assert.Equal(t, "8080", envMap["GRPC_PORT"])
	assert.Equal(t, "valid-value", envMap["VALID_VAR"])
	assert.NotContains(t, envMap, "EMPTY_VAR")

	// Should have no environment variable sources
	assert.Empty(t, envVarSources)
}
