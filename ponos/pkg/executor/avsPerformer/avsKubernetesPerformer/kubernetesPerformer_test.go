package avsKubernetesPerformer

import (
	"context"
	"sync"
	"testing"
	"time"

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
		AvsAddress:                     "0x123",
		ProcessType:                    avsPerformer.AvsProcessTypeServer,
		WorkerCount:                    1,
		ApplicationHealthCheckInterval: 5 * time.Second,
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

	// This will fail because we can't create real Kubernetes clients in tests
	_, err := NewAvsKubernetesPerformer(config, kubernetesConfig, nil, nil, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create Kubernetes client")
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

	now := time.Now()
	resource := &PerformerResource{
		performerID: "test-performer",
		avsAddress:  "0x123",
		image: avsPerformer.PerformerImage{
			Repository: "test-repo",
			Tag:        "v1.0.0",
			Digest:     "sha256:abc123",
		},
		status: avsPerformer.PerformerResourceStatusInService,
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: true,
			LastHealthCheck:      now,
		},
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
	assert.Equal(t, now, metadata.LastHealthCheck)
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
		avsAddress:  "0x123",
		image: avsPerformer.PerformerImage{
			Repository: "test-repo",
			Tag:        "v1.0.0",
		},
		status: avsPerformer.PerformerResourceStatusInService,
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: true,
			LastHealthCheck:      time.Now(),
		},
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
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: true,
			LastHealthCheck:      time.Now(),
		},
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
		performerHealth: &avsPerformer.PerformerHealth{
			ApplicationIsHealthy: true,
		},
		status: avsPerformer.PerformerResourceStatusStaged,
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

func TestAvsKubernetesPerformer_PromotePerformer_Unhealthy(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	ctx := context.Background()

	// Set up an unhealthy next performer
	performer.nextPerformer = &PerformerResource{
		performerID: "test-performer",
		performerHealth: &avsPerformer.PerformerHealth{
			ApplicationIsHealthy: false,
		},
	}

	err := performer.PromotePerformer(ctx, "test-performer")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot promote unhealthy performer")
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
