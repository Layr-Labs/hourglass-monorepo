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

func (m *MockPeeringFetcher) ListExecutorOperators(ctx context.Context, avsAddress string, number uint64) ([]*peering.OperatorPeerInfo, error) {
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
	performer, err := NewAvsKubernetesPerformer(config, kubernetesConfig, logger)

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

	// Verify it's marked as shutdown (but state still exists to prevent race conditions)
	performer.performerStatesMu.Lock()
	state, exists := performer.performerTaskStates["performer-1"]
	assert.True(t, exists)
	assert.Equal(t, PerformerStateShutdown, state.state)
	performer.performerStatesMu.Unlock()

	// Try to add task to shutdown performer (should fail)
	success = performer.tryAddTask("performer-1")
	assert.False(t, success)
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

// TestAvsKubernetesPerformer_RaceConditionFix tests that waitGroup references remain stable after cleanup
func TestAvsKubernetesPerformer_RaceConditionFix(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	// Add a task to create a state
	success := performer.tryAddTask("test-performer")
	assert.True(t, success)

	// Get the waitGroup reference before cleanup
	performer.performerStatesMu.Lock()
	state := performer.performerTaskStates["test-performer"]
	originalWaitGroup := state.waitGroup
	performer.performerStatesMu.Unlock()

	// Call cleanup (this should NOT delete the state anymore)
	performer.cleanupTaskWaitGroup("test-performer")

	// Verify state still exists and is marked as shutdown
	performer.performerStatesMu.Lock()
	state, exists := performer.performerTaskStates["test-performer"]
	assert.True(t, exists, "State should still exist after cleanup")
	assert.Equal(t, PerformerStateShutdown, state.state, "State should be marked as shutdown")

	// Verify waitGroup reference is the same (no new waitGroup created)
	assert.Same(t, originalWaitGroup, state.waitGroup, "WaitGroup reference should remain the same")
	performer.performerStatesMu.Unlock()

	// Try to add a task to shutdown performer - should fail
	success = performer.tryAddTask("test-performer")
	assert.False(t, success, "Should not allow tasks on shutdown performer")

	// Verify waitGroup is still the same after failed task addition
	performer.performerStatesMu.Lock()
	state = performer.performerTaskStates["test-performer"]
	assert.Same(t, originalWaitGroup, state.waitGroup, "WaitGroup should still be the same after failed task add")
	performer.performerStatesMu.Unlock()

	// Complete the original task to clean up the waitGroup
	performer.taskCompleted("test-performer")
}

// TestAvsKubernetesPerformer_StateTransitionUnidirectional tests that state transitions are unidirectional
func TestAvsKubernetesPerformer_StateTransitionUnidirectional(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	// Test: Active -> Draining transition
	success := performer.tryAddTask("performer-1")
	assert.True(t, success)

	performer.performerStatesMu.Lock()
	state := performer.performerTaskStates["performer-1"]
	assert.Equal(t, PerformerStateActive, state.state)

	// Manually set to draining
	state.state = PerformerStateDraining
	performer.performerStatesMu.Unlock()

	// Test: tryAddTask should fail for draining performer
	success = performer.tryAddTask("performer-1")
	assert.False(t, success, "Should not allow tasks on draining performer")

	// Test: Draining -> Shutdown transition
	performer.cleanupTaskWaitGroup("performer-1")
	performer.performerStatesMu.Lock()
	assert.Equal(t, PerformerStateShutdown, state.state)
	performer.performerStatesMu.Unlock()

	// Test: waitForTaskCompletion should NOT regress Shutdown back to Draining
	go func() {
		time.Sleep(1 * time.Second)
		performer.taskCompleted("performer-1")
	}()
	performer.waitForTaskCompletion("performer-1")
	performer.performerStatesMu.Lock()
	assert.Equal(t, PerformerStateShutdown, state.state, "Should remain shutdown, not regress to draining")
	performer.performerStatesMu.Unlock()

	// Test: tryAddTask should still fail for shutdown performer
	success = performer.tryAddTask("performer-1")
	assert.False(t, success, "Should not allow tasks on shutdown performer")
}

// TestAvsKubernetesPerformer_ConcurrentWaitForTaskCompletion tests that concurrent calls to waitForTaskCompletion are safe
func TestAvsKubernetesPerformer_ConcurrentWaitForTaskCompletion(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)

	// Add multiple tasks to the same performer
	performerID := "concurrent-test-performer"

	// Add 3 tasks
	for i := 0; i < 3; i++ {
		success := performer.tryAddTask(performerID)
		assert.True(t, success)
	}

	// Get the waitGroup reference
	performer.performerStatesMu.Lock()
	state := performer.performerTaskStates[performerID]
	waitGroupRef := state.waitGroup
	performer.performerStatesMu.Unlock()

	// Start multiple concurrent waitForTaskCompletion calls
	numWaiters := 3
	waitersStarted := make(chan struct{}, numWaiters)
	waitersFinished := make(chan struct{}, numWaiters)

	for i := 0; i < numWaiters; i++ {
		go func(waiterID int) {
			waitersStarted <- struct{}{}

			// All waiters should use the same waitGroup reference
			performer.waitForTaskCompletion(performerID)

			waitersFinished <- struct{}{}
		}(i)
	}

	// Wait for all waiters to start
	for i := 0; i < numWaiters; i++ {
		<-waitersStarted
	}

	// Give waiters a moment to get to the Wait() call
	time.Sleep(100 * time.Millisecond)

	// Verify all waiters are still using the same waitGroup
	performer.performerStatesMu.Lock()
	currentState := performer.performerTaskStates[performerID]
	assert.Same(t, waitGroupRef, currentState.waitGroup, "All waiters should use the same waitGroup")
	assert.Equal(t, PerformerStateDraining, currentState.state, "State should be draining")
	performer.performerStatesMu.Unlock()

	// Complete all tasks - this should release all waiters
	for i := 0; i < 3; i++ {
		performer.taskCompleted(performerID)
	}

	// Wait for all waiters to finish
	for i := 0; i < numWaiters; i++ {
		select {
		case <-waitersFinished:
			// Waiter finished successfully
		case <-time.After(1 * time.Second):
			t.Fatal("Waiter timed out - this indicates a deadlock or race condition")
		}
	}

	// Verify final state
	performer.performerStatesMu.Lock()
	finalState := performer.performerTaskStates[performerID]
	assert.Same(t, waitGroupRef, finalState.waitGroup, "WaitGroup reference should still be the same")
	performer.performerStatesMu.Unlock()
}

// TestAvsKubernetesPerformer_WaitGroupStabilityUnderConcurrentOperations tests waitGroup stability under various concurrent operations
func TestAvsKubernetesPerformer_WaitGroupStabilityUnderConcurrentOperations(t *testing.T) {
	performer, _, _, _ := createTestKubernetesPerformer(t)
	performerID := "stability-test-performer"

	// Add initial task
	success := performer.tryAddTask(performerID)
	assert.True(t, success)

	// Get initial waitGroup reference
	performer.performerStatesMu.Lock()
	initialState := performer.performerTaskStates[performerID]
	initialWaitGroup := initialState.waitGroup
	performer.performerStatesMu.Unlock()

	// Start concurrent operations
	var wg sync.WaitGroup
	tasksAdded := make(chan int, 1) // Channel to communicate how many tasks were actually added

	// Goroutine 1: Try to add more tasks (should succeed until cleanup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		successfulAdds := 0
		for i := 0; i < 5; i++ {
			if performer.tryAddTask(performerID) {
				successfulAdds++
			}
			time.Sleep(10 * time.Millisecond)
		}
		tasksAdded <- successfulAdds // Send count to completion goroutine
	}()

	// Goroutine 2: Call waitForTaskCompletion
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond) // Let some tasks get added first
		performer.waitForTaskCompletion(performerID)
	}()

	// Goroutine 3: Complete tasks gradually
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond) // Let draining start

		// Get the count of successfully added tasks
		var additionalTasksAdded int
		select {
		case additionalTasksAdded = <-tasksAdded:
		case <-time.After(500 * time.Millisecond):
			// Fallback if channel read times out
			additionalTasksAdded = 0
		}

		// Complete tasks: 1 initial + however many were actually added in goroutine 1
		totalTasksToComplete := 1 + additionalTasksAdded
		for i := 0; i < totalTasksToComplete; i++ {
			performer.taskCompleted(performerID)
			time.Sleep(20 * time.Millisecond)
		}
	}()

	// Goroutine 4: Call cleanup after some time
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond)
		performer.cleanupTaskWaitGroup(performerID)
	}()

	// Wait for all operations to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All operations completed
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out - indicates a deadlock or race condition")
	}

	// Verify final state - waitGroup reference should still be the same
	performer.performerStatesMu.Lock()
	finalState := performer.performerTaskStates[performerID]
	assert.True(t, finalState != nil, "State should still exist")
	assert.Same(t, initialWaitGroup, finalState.waitGroup, "WaitGroup reference should never change")
	assert.Equal(t, PerformerStateShutdown, finalState.state, "Final state should be shutdown")
	performer.performerStatesMu.Unlock()
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
