package avsKubernetesPerformer

import (
	"context"
	"sync"
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
	corev1 "k8s.io/api/core/v1"
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
		AvsAddress:  "0x1234567890",
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
		config:             config,
		kubernetesConfig:   kubernetesConfig,
		logger:             logger,
		peeringFetcher:     mockPeeringFetcher,
		taskWaitGroups:     make(map[string]*sync.WaitGroup),
		drainingPerformers: make(map[string]struct{}),
	}

	return performer, mockCRDOps, mockClientWrapper, mockPeeringFetcher
}

func TestNewAvsKubernetesPerformer(t *testing.T) {
	config := &avsPerformer.AvsPerformerConfig{
		AvsAddress:  "0x1234567890",
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
	assert.Contains(t, id1, "performer-0x123456-")
	assert.Contains(t, id2, "performer-0x123456-")
}

func TestAvsKubernetesPerformer_ConvertPerformerResource(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	resource := &PerformerResource{
		performerID: "test-performer",
		avsAddress:  "0x1234567890",
		image: avsPerformer.PerformerImage{
			Repository: "test-repo",
			Tag:        "v1.0.0",
			Digest:     "sha256:abc123",
		},
		status: avsPerformer.PerformerResourceStatusInService,
	}

	metadata := performer.convertPerformerResource(resource)

	assert.Equal(t, "test-performer", metadata.PerformerID)
	assert.Equal(t, "0x1234567890", metadata.AvsAddress)
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

func TestAvsKubernetesPerformer_GetOrCreateTaskWaitGroup(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	// Test creating new wait group
	wg1 := performer.getOrCreateTaskWaitGroup("performer-1")
	assert.NotNil(t, wg1)

	// Test getting existing wait group
	wg2 := performer.getOrCreateTaskWaitGroup("performer-1")
	assert.Same(t, wg1, wg2) // Should be the same instance

	// Test creating different wait group
	wg3 := performer.getOrCreateTaskWaitGroup("performer-2")
	assert.NotNil(t, wg3)
	assert.NotSame(t, wg1, wg3)
}

func TestAvsKubernetesPerformer_CleanupTaskWaitGroup(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	// Create a wait group
	wg := performer.getOrCreateTaskWaitGroup("performer-1")
	assert.NotNil(t, wg)

	// Cleanup
	performer.cleanupTaskWaitGroup("performer-1")

	// Verify it's cleaned up by creating a new one
	wg2 := performer.getOrCreateTaskWaitGroup("performer-1")
	assert.NotSame(t, wg, wg2) // Should be a new instance
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
		avsAddress:  "0x1234567890",
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
		avsAddress:  "0x1234567890",
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

	env := performer.buildEnvironmentFromImage(image)

	// Should have default environment variables
	assert.Len(t, env, 2)
	assert.Equal(t, "AVS_ADDRESS", env[0].Name)
	assert.Equal(t, "0x1234567890", env[0].Value)
	assert.Equal(t, "GRPC_PORT", env[1].Name)
	assert.Equal(t, "8080", env[1].Value)
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

	env := performer.buildEnvironmentFromImage(image)

	// Should have default + custom environment variables
	assert.Len(t, env, 4)

	// Check defaults
	assert.Equal(t, "AVS_ADDRESS", env[0].Name)
	assert.Equal(t, "0x1234567890", env[0].Value)
	assert.Equal(t, "GRPC_PORT", env[1].Name)
	assert.Equal(t, "8080", env[1].Value)

	// Check custom values
	assert.Equal(t, "DATABASE_URL", env[2].Name)
	assert.Equal(t, "postgres://localhost/db", env[2].Value)
	assert.Equal(t, "API_ENDPOINT", env[3].Name)
	assert.Equal(t, "https://api.example.com", env[3].Value)
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
					ValueFrom: config.ValueFrom{
						SecretKeyRef: config.SecretKeyRef{
							Name: "api-secrets",
							Key:  "api-key",
						},
					},
				},
			},
		},
	}

	env := performer.buildEnvironmentFromImage(image)

	// Should have default + secret ref environment variables
	assert.Len(t, env, 3)

	// Check defaults
	assert.Equal(t, "AVS_ADDRESS", env[0].Name)
	assert.Equal(t, "0x1234567890", env[0].Value)
	assert.Equal(t, "GRPC_PORT", env[1].Name)
	assert.Equal(t, "8080", env[1].Value)

	// Check secret ref
	assert.Equal(t, "API_KEY", env[2].Name)
	assert.Empty(t, env[2].Value)
	assert.NotNil(t, env[2].ValueFrom)
	assert.NotNil(t, env[2].ValueFrom.SecretKeyRef)
	assert.Equal(t, "api-secrets", env[2].ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "api-key", env[2].ValueFrom.SecretKeyRef.Key)
	assert.Nil(t, env[2].ValueFrom.ConfigMapKeyRef)
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
					ValueFrom: config.ValueFrom{
						ConfigMapKeyRef: config.ConfigMapKeyRef{
							Name: "app-config",
							Key:  "config.json",
						},
					},
				},
			},
		},
	}

	env := performer.buildEnvironmentFromImage(image)

	// Should have default + configmap ref environment variables
	assert.Len(t, env, 3)

	// Check defaults
	assert.Equal(t, "AVS_ADDRESS", env[0].Name)
	assert.Equal(t, "0x1234567890", env[0].Value)
	assert.Equal(t, "GRPC_PORT", env[1].Name)
	assert.Equal(t, "8080", env[1].Value)

	// Check configmap ref
	assert.Equal(t, "APP_CONFIG", env[2].Name)
	assert.Empty(t, env[2].Value)
	assert.NotNil(t, env[2].ValueFrom)
	assert.NotNil(t, env[2].ValueFrom.ConfigMapKeyRef)
	assert.Equal(t, "app-config", env[2].ValueFrom.ConfigMapKeyRef.Name)
	assert.Equal(t, "config.json", env[2].ValueFrom.ConfigMapKeyRef.Key)
	assert.Nil(t, env[2].ValueFrom.SecretKeyRef)
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
					ValueFrom: config.ValueFrom{
						SecretKeyRef: config.SecretKeyRef{
							Name: "api-secrets",
							Key:  "api-key",
						},
					},
				},
			},
			{
				Name: "APP_CONFIG",
				KubernetesEnv: &config.KubernetesEnv{
					ValueFrom: config.ValueFrom{
						ConfigMapKeyRef: config.ConfigMapKeyRef{
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

	env := performer.buildEnvironmentFromImage(image)

	// Should have all environment variables
	assert.Len(t, env, 6) // 2 defaults + 2 direct + 2 refs

	// Check defaults
	assert.Equal(t, "AVS_ADDRESS", env[0].Name)
	assert.Equal(t, "0x1234567890", env[0].Value)
	assert.Equal(t, "GRPC_PORT", env[1].Name)
	assert.Equal(t, "8080", env[1].Value)

	// Check direct values
	assert.Equal(t, "DATABASE_URL", env[2].Name)
	assert.Equal(t, "postgres://localhost/db", env[2].Value)

	// Find and verify specific env vars
	var apiKeyEnv, appConfigEnv, logLevelEnv *corev1.EnvVar
	for i := range env {
		switch env[i].Name {
		case "API_KEY":
			apiKeyEnv = &env[i]
		case "APP_CONFIG":
			appConfigEnv = &env[i]
		case "LOG_LEVEL":
			logLevelEnv = &env[i]
		}
	}

	// Verify secret ref
	require.NotNil(t, apiKeyEnv)
	assert.Empty(t, apiKeyEnv.Value)
	assert.NotNil(t, apiKeyEnv.ValueFrom)
	assert.NotNil(t, apiKeyEnv.ValueFrom.SecretKeyRef)
	assert.Equal(t, "api-secrets", apiKeyEnv.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "api-key", apiKeyEnv.ValueFrom.SecretKeyRef.Key)

	// Verify configmap ref
	require.NotNil(t, appConfigEnv)
	assert.Empty(t, appConfigEnv.Value)
	assert.NotNil(t, appConfigEnv.ValueFrom)
	assert.NotNil(t, appConfigEnv.ValueFrom.ConfigMapKeyRef)
	assert.Equal(t, "app-config", appConfigEnv.ValueFrom.ConfigMapKeyRef.Name)
	assert.Equal(t, "config.json", appConfigEnv.ValueFrom.ConfigMapKeyRef.Key)

	// Verify direct value
	require.NotNil(t, logLevelEnv)
	assert.Equal(t, "debug", logLevelEnv.Value)
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

	env := performer.buildEnvironmentFromImage(image)

	// Should have default + valid environment variable only
	assert.Len(t, env, 3)

	// Check defaults
	assert.Equal(t, "AVS_ADDRESS", env[0].Name)
	assert.Equal(t, "0x1234567890", env[0].Value)
	assert.Equal(t, "GRPC_PORT", env[1].Name)
	assert.Equal(t, "8080", env[1].Value)

	// Check valid var
	assert.Equal(t, "VALID_VAR", env[2].Name)
	assert.Equal(t, "valid-value", env[2].Value)

	// Ensure EMPTY_VAR is not included
	for _, e := range env {
		assert.NotEqual(t, "EMPTY_VAR", e.Name)
	}
}
