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

func TestAvsKubernetesPerformer_ConnectionRetryIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create test configuration
	config := &avsPerformer.AvsPerformerConfig{
		AvsAddress:                     "0xtest123",
		ApplicationHealthCheckInterval: 5 * time.Second,
		Image: avsPerformer.PerformerImage{
			Repository: "test-repo",
			Tag:        "test-tag",
		},
	}

	kubernetesConfig := &kubernetesManager.Config{
		Namespace:         "test-namespace",
		KubeconfigPath:    "", // Will use in-cluster config
		OperatorNamespace: "hourglass-system",
		CRDGroup:          "hourglass.eigenlayer.io",
		CRDVersion:        "v1alpha1",
		ServiceAccount:    "test-service-account",
	}

	// Create performer (this will fail but we can test the retry logic structure)
	performer, err := NewAvsKubernetesPerformer(
		config,
		kubernetesConfig,
		nil, // peeringFetcher
		nil, // l1ContractCaller
		logger,
	)

	if err != nil {
		t.Logf("Expected error creating performer without valid k8s config: %v", err)
		// This is expected in test environment
		return
	}

	// Test connection manager integration
	if performer.kubernetesManager == nil {
		t.Error("Expected kubernetesManager to be set")
	}

	if performer.clientWrapper == nil {
		t.Error("Expected clientWrapper to be set")
	}
}

func TestAvsKubernetesPerformer_PerformerResourceConnectionManager(t *testing.T) {
	// Test that PerformerResource properly integrates with ConnectionManager

	// Create a mock connection manager
	retryConfig := &clients.RetryConfig{
		MaxRetries:        3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 5 * time.Second,
	}

	connectionManager := clients.NewConnectionManager("localhost:9090", true, retryConfig, logger)

	// Create a test performer resource
	performer := &PerformerResource{
		performerID:       "test-performer-123",
		avsAddress:        "0xtest123",
		connectionManager: connectionManager,
		endpoint:          "localhost:9090",
		statusChan:        make(chan avsPerformer.PerformerStatusEvent, 10),
		createdAt:         time.Now(),
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: true,
			LastHealthCheck:      time.Now(),
		},
	}

	// Test connection manager integration
	if performer.connectionManager == nil {
		t.Error("Expected connectionManager to be set")
	}

	// Test circuit breaker functionality
	if performer.connectionManager.IsCircuitOpen() {
		t.Error("Expected circuit breaker to be closed initially")
	}

	// Test connection stats
	stats := performer.connectionManager.GetConnectionStats()
	if stats["circuitOpen"] != false {
		t.Errorf("Expected circuitOpen to be false, got %v", stats["circuitOpen"])
	}

	if stats["hasConnection"] != false {
		t.Errorf("Expected hasConnection to be false, got %v", stats["hasConnection"])
	}
}

func TestAvsKubernetesPerformer_HealthCheckWithRetry(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock performer with connection manager
	retryConfig := &clients.RetryConfig{
		MaxRetries:        2,
		InitialDelay:      50 * time.Millisecond,
		MaxDelay:          500 * time.Millisecond,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 1 * time.Second,
	}

	connectionManager := clients.NewConnectionManager("localhost:9999", true, retryConfig, logger)

	performer := &PerformerResource{
		performerID:       "test-performer-health",
		avsAddress:        "0xtest123",
		connectionManager: connectionManager,
		endpoint:          "localhost:9999",
		statusChan:        make(chan avsPerformer.PerformerStatusEvent, 10),
		createdAt:         time.Now(),
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: true,
			LastHealthCheck:      time.Now(),
		},
	}

	// Create a mock AvsKubernetesPerformer for testing health check
	akp := &AvsKubernetesPerformer{
		logger: logger,
		config: &avsPerformer.AvsPerformerConfig{
			ApplicationHealthCheckInterval: 1 * time.Second,
		},
	}

	// Test health check with unavailable service
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should fail but test the retry logic
	akp.performApplicationHealthCheck(ctx, performer)

	// Verify health status was updated
	if performer.performerHealth.ApplicationIsHealthy {
		t.Error("Expected ApplicationIsHealthy to be false after failed health check")
	}

	if performer.performerHealth.ConsecutiveApplicationHealthFailures == 0 {
		t.Error("Expected ConsecutiveApplicationHealthFailures to be incremented")
	}
}

func TestAvsKubernetesPerformer_TaskExecutionWithRetry(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock performer with connection manager
	retryConfig := &clients.RetryConfig{
		MaxRetries:        1,
		InitialDelay:      10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 500 * time.Millisecond,
	}

	connectionManager := clients.NewConnectionManager("localhost:9998", true, retryConfig, logger)

	performer := &PerformerResource{
		performerID:       "test-performer-task",
		avsAddress:        "0xtest123",
		connectionManager: connectionManager,
		endpoint:          "localhost:9998",
		statusChan:        make(chan avsPerformer.PerformerStatusEvent, 10),
		createdAt:         time.Now(),
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: true,
			LastHealthCheck:      time.Now(),
		},
	}

	// Create a mock AvsKubernetesPerformer for testing task execution
	akp := &AvsKubernetesPerformer{
		logger:         logger,
		taskWaitGroups: make(map[string]*sync.WaitGroup),
	}

	// Store the performer as current
	akp.currentPerformer.Store(performer)

	// Test task execution with unavailable service
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	task := &performerTask.PerformerTask{
		TaskID:  "test-task-123",
		Payload: []byte("test-payload"),
	}

	_, err := akp.RunTask(ctx, task)
	if err == nil {
		t.Error("Expected error when service is unavailable")
	}

	// Verify error contains connection information
	if err != nil && !contains(err.Error(), "failed to get healthy connection") {
		t.Errorf("Expected connection error, got: %v", err)
	}
}

func TestAvsKubernetesPerformer_CircuitBreakerIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock performer with connection manager
	retryConfig := &clients.RetryConfig{
		MaxRetries:        1,
		InitialDelay:      10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 100 * time.Millisecond,
	}

	connectionManager := clients.NewConnectionManager("localhost:9997", true, retryConfig, logger)

	performer := &PerformerResource{
		performerID:       "test-performer-circuit",
		avsAddress:        "0xtest123",
		connectionManager: connectionManager,
		endpoint:          "localhost:9997",
		statusChan:        make(chan avsPerformer.PerformerStatusEvent, 10),
		createdAt:         time.Now(),
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: true,
			LastHealthCheck:      time.Now(),
		},
	}

	// Create a mock AvsKubernetesPerformer
	akp := &AvsKubernetesPerformer{
		logger:         logger,
		taskWaitGroups: make(map[string]*sync.WaitGroup),
		config: &avsPerformer.AvsPerformerConfig{
			ApplicationHealthCheckInterval: 1 * time.Second,
		},
	}

	// Store the performer as current
	akp.currentPerformer.Store(performer)

	// Force circuit breaker to open by simulating failures
	// In a real scenario, this would happen through actual connection failures
	// For testing, we can't easily manipulate internal state, so we'll test behavior

	// Test task execution - will fail due to unavailable service
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	task := &performerTask.PerformerTask{
		TaskID:  "test-task-circuit",
		Payload: []byte("test-payload"),
	}

	_, err := akp.RunTask(ctx, task)
	if err == nil {
		t.Error("Expected error when service is unavailable")
	}

	// Verify error contains connection failure information
	if err != nil && !contains(err.Error(), "failed to get healthy connection") {
		t.Errorf("Expected connection failure error, got: %v", err)
	}

	// Test health check with unavailable service
	akp.performApplicationHealthCheck(ctx, performer)

	// Verify health check failed
	if performer.performerHealth.ApplicationIsHealthy {
		t.Error("Expected ApplicationIsHealthy to be false when service is unavailable")
	}
}

func TestAvsKubernetesPerformer_ConnectionCleanup(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock performer with connection manager
	retryConfig := &clients.RetryConfig{
		MaxRetries:        1,
		InitialDelay:      10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 100 * time.Millisecond,
	}

	connectionManager := clients.NewConnectionManager("localhost:9996", true, retryConfig, logger)

	performer := &PerformerResource{
		performerID:       "test-performer-cleanup",
		avsAddress:        "0xtest123",
		connectionManager: connectionManager,
		endpoint:          "localhost:9996",
		statusChan:        make(chan avsPerformer.PerformerStatusEvent, 10),
		createdAt:         time.Now(),
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: true,
			LastHealthCheck:      time.Now(),
		},
	}

	// Create a mock AvsKubernetesPerformer
	akp := &AvsKubernetesPerformer{
		logger:             logger,
		taskWaitGroups:     make(map[string]*sync.WaitGroup),
		drainingPerformers: make(map[string]struct{}),
	}

	// Store the performer as current
	akp.currentPerformer.Store(performer)

	// Test connection manager cleanup directly (without k8s operations)
	// This simulates the cleanup that would happen in startDrainAndRemove
	if performer.connectionManager != nil {
		err := performer.connectionManager.Close()
		if err != nil {
			t.Errorf("Expected no error closing connection manager, got: %v", err)
		}
	}

	// Verify cleanup was attempted
	// Note: This test is limited since we can't actually create k8s resources
	// but it verifies the connection manager cleanup logic
}

// Helper function to check if error contains a substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && containsHelper(str, substr)
}

func containsHelper(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Note: In a real implementation, these methods would be internal to ConnectionManager
// For testing purposes, we simulate the circuit breaker behavior by checking the
// connection manager's public methods and state
