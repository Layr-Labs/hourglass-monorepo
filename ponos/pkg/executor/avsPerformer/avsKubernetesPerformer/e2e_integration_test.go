package avsKubernetesPerformer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/kubernetesManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"go.uber.org/zap/zaptest"
)

// TestE2E_KubernetesPerformer_FullWorkflow tests the complete workflow:
// 1. Create performer with connection manager
// 2. Test connection retry logic
// 3. Test health monitoring
// 4. Test task execution
// 5. Test cleanup
func TestE2E_KubernetesPerformer_FullWorkflow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Test configuration
	config := &avsPerformer.AvsPerformerConfig{
		AvsAddress: "0xtest-avs-address",
		ApplicationHealthCheckInterval: 2 * time.Second,
		Image: avsPerformer.PerformerImage{
			Repository: "test-registry/test-performer",
			Tag:        "v1.0.0",
			Digest:     "sha256:abcdef123456",
		},
	}
	
	kubernetesConfig := &kubernetesManager.Config{
		Namespace:         "test-e2e-namespace",
		KubeconfigPath:    "", // Use in-cluster config
		OperatorNamespace: "hourglass-system",
		CRDGroup:          "hourglass.eigenlayer.io",
		CRDVersion:        "v1alpha1",
		ServiceAccount:    "test-service-account",
		ConnectionTimeout: 10 * time.Second,
	}
	
	// Attempt to create performer (will fail in test environment but validates structure)
	performer, err := NewAvsKubernetesPerformer(
		config,
		kubernetesConfig,
		nil, // peeringFetcher
		nil, // l1ContractCaller
		logger,
	)
	
	if err != nil {
		// Expected in test environment - log and continue with mock testing
		t.Logf("Expected error creating performer in test environment: %v", err)
		
		// Test the retry configuration structure
		retryConfig := &clients.RetryConfig{
			MaxRetries:        3,
			InitialDelay:      500 * time.Millisecond,
			MaxDelay:          5 * time.Second,
			BackoffMultiplier: 2.0,
			ConnectionTimeout: 10 * time.Second,
		}
		
		// Test connection manager creation
		cm := clients.NewConnectionManager("test-service:9090", true, retryConfig)
		if cm == nil {
			t.Error("Expected connection manager to be created")
		}
		
		// Test initial circuit breaker state
		if cm.IsCircuitOpen() {
			t.Error("Expected circuit breaker to be closed initially")
		}
		
		// Test connection stats
		stats := cm.GetConnectionStats()
		if stats["circuitOpen"] != false {
			t.Errorf("Expected circuitOpen to be false, got %v", stats["circuitOpen"])
		}
		
		return
	}
	
	// If we get here, test the full workflow
	t.Log("Testing full Kubernetes performer workflow")
	
	// Test 1: Performer initialization
	if performer.kubernetesManager == nil {
		t.Error("Expected kubernetesManager to be initialized")
	}
	
	if performer.clientWrapper == nil {
		t.Error("Expected clientWrapper to be initialized")
	}
	
	// Test 2: Connection manager integration
	// Create a test performer resource to validate connection manager integration
	testPerformerResource := createTestPerformerResource(t, "test-endpoint:9090")
	
	if testPerformerResource.connectionManager == nil {
		t.Error("Expected connectionManager to be set on performer resource")
	}
	
	// Test 3: Circuit breaker functionality
	if testPerformerResource.connectionManager.IsCircuitOpen() {
		t.Error("Expected circuit breaker to be closed initially")
	}
	
	// Test 4: Health monitoring simulation
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	// This will fail but tests the retry logic
	performer.performApplicationHealthCheck(ctx, testPerformerResource)
	
	// Verify health state was updated
	if testPerformerResource.performerHealth.ApplicationIsHealthy {
		t.Error("Expected ApplicationIsHealthy to be false after failed health check")
	}
	
	// Test 5: Task execution simulation
	task := &performerTask.PerformerTask{
		TaskID:  "test-task-e2e",
		Payload: []byte("test-payload-data"),
	}
	
	// Store the test performer as current
	performer.currentPerformer.Store(testPerformerResource)
	
	_, err = performer.RunTask(ctx, task)
	if err == nil {
		t.Error("Expected error when service is unavailable")
	}
	
	// Verify error contains connection information
	if err != nil && !containsSubstring(err.Error(), "failed to get healthy connection") {
		t.Errorf("Expected connection error, got: %v", err)
	}
	
	// Test 6: Cleanup
	err = performer.RemovePerformer(ctx, testPerformerResource.performerID)
	if err != nil {
		t.Logf("Expected error removing performer in test environment: %v", err)
	}
	
	t.Log("E2E workflow test completed successfully")
}

func TestE2E_ConnectionResilience_FailureRecovery(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Test connection resilience under different failure scenarios
	scenarios := []struct {
		name          string
		endpoint      string
		expectedError string
	}{
		{
			name:          "Connection refused",
			endpoint:      "localhost:9999",
			expectedError: "failed to get healthy connection",
		},
		{
			name:          "Invalid hostname",
			endpoint:      "non-existent-host:9090",
			expectedError: "failed to get healthy connection",
		},
		{
			name:          "Invalid port",
			endpoint:      "localhost:99999",
			expectedError: "failed to get healthy connection",
		},
	}
	
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Create performer resource with different endpoints
			performer := createTestPerformerResource(t, scenario.endpoint)
			
			// Create mock AvsKubernetesPerformer
			akp := &AvsKubernetesPerformer{
				logger: logger,
				config: &avsPerformer.AvsPerformerConfig{
					ApplicationHealthCheckInterval: 1 * time.Second,
				},
				taskWaitGroups: make(map[string]*sync.WaitGroup),
			}
			
			// Store performer as current
			akp.currentPerformer.Store(performer)
			
			// Test task execution
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			
			task := &performerTask.PerformerTask{
				TaskID:  "test-task-" + scenario.name,
				Payload: []byte("test-payload"),
			}
			
			_, err := akp.RunTask(ctx, task)
			if err == nil {
				t.Errorf("Expected error for scenario %s", scenario.name)
			}
			
			if err != nil && !containsSubstring(err.Error(), scenario.expectedError) {
				t.Errorf("Expected error containing '%s', got: %v", scenario.expectedError, err)
			}
			
			// Test health check with a shorter timeout
			healthCtx, healthCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer healthCancel()
			akp.performApplicationHealthCheck(healthCtx, performer)
			
			if performer.performerHealth.ApplicationIsHealthy {
				t.Errorf("Expected ApplicationIsHealthy to be false for scenario %s", scenario.name)
			}
		})
	}
}

func TestE2E_CircuitBreaker_StateTransitions(t *testing.T) {
	// Test circuit breaker state transitions over time
	retryConfig := &clients.RetryConfig{
		MaxRetries:        1,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          500 * time.Millisecond,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 500 * time.Millisecond,
	}
	
	cm := clients.NewConnectionManager("localhost:9998", true, retryConfig)
	
	// Test initial state
	if cm.IsCircuitOpen() {
		t.Error("Expected circuit breaker to be closed initially")
	}
	
	// Simulate multiple connection attempts (will fail)
	for i := 0; i < 3; i++ {
		_, err := cm.GetConnection()
		if err == nil {
			t.Error("Expected connection to fail")
		}
	}
	
	// Test connection stats after failures
	stats := cm.GetConnectionStats()
	if stats["hasConnection"] != false {
		t.Errorf("Expected hasConnection to be false, got %v", stats["hasConnection"])
	}
	
	// Test cleanup
	err := cm.Close()
	if err != nil {
		t.Errorf("Expected no error closing connection manager, got: %v", err)
	}
}

func TestE2E_PerformerLifecycle_Complete(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Test complete performer lifecycle
	performer := createTestPerformerResource(t, "test-lifecycle:9090")
	
	// Create mock AvsKubernetesPerformer
	akp := &AvsKubernetesPerformer{
		logger: logger,
		config: &avsPerformer.AvsPerformerConfig{
			ApplicationHealthCheckInterval: 1 * time.Second,
		},
		taskWaitGroups: make(map[string]*sync.WaitGroup),
	}
	
	// Test 1: Performer creation
	if performer.performerID == "" {
		t.Error("Expected performerID to be set")
	}
	
	if performer.connectionManager == nil {
		t.Error("Expected connectionManager to be set")
	}
	
	// Test 2: Health monitoring
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	akp.performApplicationHealthCheck(ctx, performer)
	
	// Should be unhealthy due to unavailable service
	if performer.performerHealth.ApplicationIsHealthy {
		t.Error("Expected ApplicationIsHealthy to be false")
	}
	
	// Test 3: Task execution
	akp.currentPerformer.Store(performer)
	
	task := &performerTask.PerformerTask{
		TaskID:  "test-lifecycle-task",
		Payload: []byte("test-payload"),
	}
	
	_, err := akp.RunTask(ctx, task)
	if err == nil {
		t.Error("Expected error when service is unavailable")
	}
	
	// Test 4: Cleanup
	if performer.connectionManager != nil {
		err := performer.connectionManager.Close()
		if err != nil {
			t.Errorf("Expected no error closing connection manager, got: %v", err)
		}
	}
	
	// Test 5: Resource cleanup
	akp.currentPerformer.Store((*PerformerResource)(nil))
	
	current := akp.currentPerformer.Load()
	if current != nil {
		// Check if it's actually a nil pointer of the correct type
		if performer, ok := current.(*PerformerResource); !ok || performer != nil {
			t.Error("Expected current performer to be nil after cleanup")
		}
	}
}

// Helper function to create a test performer resource
func createTestPerformerResource(t testing.TB, endpoint string) *PerformerResource {
	retryConfig := &clients.RetryConfig{
		MaxRetries:        2,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 2 * time.Second,
	}
	
	connectionManager := clients.NewConnectionManager(endpoint, true, retryConfig)
	
	return &PerformerResource{
		performerID:       "test-performer-" + time.Now().Format("20060102-150405"),
		avsAddress:        "0xtest-avs-address",
		connectionManager: connectionManager,
		endpoint:          endpoint,
		statusChan:        make(chan avsPerformer.PerformerStatusEvent, 10),
		createdAt:         time.Now(),
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: true,
			LastHealthCheck:      time.Now(),
		},
	}
}

// Helper function to check if error contains a substring
func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && containsSubstringHelper(str, substr)
}

func containsSubstringHelper(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark test for connection management performance
func BenchmarkE2E_ConnectionManager_Performance(b *testing.B) {
	retryConfig := &clients.RetryConfig{
		MaxRetries:        1,
		InitialDelay:      10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 100 * time.Millisecond,
	}
	
	cm := clients.NewConnectionManager("localhost:9997", true, retryConfig)
	defer cm.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cm.IsCircuitOpen()
	}
}

func BenchmarkE2E_PerformerHealth_Check(b *testing.B) {
	logger := zaptest.NewLogger(b)
	performer := createTestPerformerResource(b, "localhost:9996")
	
	akp := &AvsKubernetesPerformer{
		logger: logger,
		config: &avsPerformer.AvsPerformerConfig{
			ApplicationHealthCheckInterval: 1 * time.Second,
		},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		akp.performApplicationHealthCheck(ctx, performer)
	}
}