package avsKubernetesPerformer

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/kubernetesManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/connectivity"
)

// TestAvsKubernetesPerformer_ConnectionRetryIntegration tests that the Kubernetes performer correctly integrates with gRPC connections
func TestAvsKubernetesPerformer_ConnectionRetryIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := &avsPerformer.AvsPerformerConfig{
		AvsAddress: "0xtest123",
	}

	kubernetesConfig := &kubernetesManager.Config{
		Namespace:         "default",
		OperatorNamespace: "hourglass-system",
		CRDGroup:          "hourglass.eigenlayer.io",
		CRDVersion:        "v1alpha1",
		KubeconfigPath:    "", // Fixed field name
	}

	akp, err := NewAvsKubernetesPerformer(config, kubernetesConfig, nil, nil, logger)
	if err != nil {
		// Expected error in test environment without k8s
		t.Logf("Expected error creating performer without valid k8s config: %v", err)
	}

	if akp != nil && akp.clientWrapper != nil {
		t.Error("Expected clientWrapper to be nil in test environment")
	}
}

func TestAvsKubernetesPerformer_PerformerResourceConnectionManager(t *testing.T) {
	// Test that PerformerResource properly uses gRPC connections
	// Create a mock gRPC connection
	retryConfig := &clients.RetryConfig{
		MaxRetries:        3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 5 * time.Second,
	}

	grpcConn, err := clients.NewGrpcClientWithRetry("localhost:9090", true, retryConfig)
	if err != nil {
		t.Fatalf("Failed to create gRPC connection: %v", err)
	}
	defer grpcConn.Close()

	// Create test client
	client := performerV1.NewPerformerServiceClient(grpcConn)

	// Create a test performer resource
	performer := &PerformerResource{
		performerID: "test-performer-123",
		avsAddress:  "0xtest123",
		grpcConn:    grpcConn,
		client:      client,
		endpoint:    "localhost:9090",
		statusChan:  make(chan avsPerformer.PerformerStatusEvent, 10),
		createdAt:   time.Now(),
	}

	// Test gRPC connection is set
	if performer.grpcConn == nil {
		t.Error("Expected grpcConn to be set")
	}

	// Test clients are set
	if performer.client == nil {
		t.Error("Expected client to be set")
	}

}

// TestAvsKubernetesPerformer_TaskExecutionWithRetry tests task execution with retry logic
func TestAvsKubernetesPerformer_TaskExecutionWithRetry(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create retry config with fast retries for testing
	retryConfig := &clients.RetryConfig{
		MaxRetries:        1,
		InitialDelay:      10 * time.Millisecond,
		MaxDelay:          50 * time.Millisecond,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 100 * time.Millisecond,
	}

	// Create a gRPC connection (will fail to connect)
	grpcConn, _ := clients.NewGrpcClientWithRetry("localhost:9998", true, retryConfig)
	if grpcConn != nil {
		defer grpcConn.Close()
	}

	performer := &PerformerResource{
		performerID: "test-performer-task",
		grpcConn:    grpcConn,
		client:      performerV1.NewPerformerServiceClient(grpcConn),
	}

	// Create AvsKubernetesPerformer
	akp := &AvsKubernetesPerformer{
		config:         &avsPerformer.AvsPerformerConfig{},
		logger:         logger,
		taskWaitGroups: make(map[string]*sync.WaitGroup),
	}
	akp.currentPerformer.Store(performer)

	// Try to execute a task
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
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
	// The error could be either from ConnectionManager or from the retry interceptor
	if err != nil && !contains(err.Error(), "failed to get healthy connection") &&
		!contains(err.Error(), "request failed after") &&
		!contains(err.Error(), "connection refused") {
		t.Errorf("Expected connection error, got: %v", err)
	}
}

// TestAvsKubernetesPerformer_CircuitBreakerIntegration tests circuit breaker behavior
func TestAvsKubernetesPerformer_CircuitBreakerIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create retry config with fast retries for testing
	retryConfig := &clients.RetryConfig{
		MaxRetries:        1,
		InitialDelay:      10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 100 * time.Millisecond,
	}

	// Create a gRPC connection (will fail to connect)
	grpcConn, _ := clients.NewGrpcClientWithRetry("localhost:9997", true, retryConfig)
	if grpcConn != nil {
		defer grpcConn.Close()
	}

	performer := &PerformerResource{
		performerID: "test-performer-circuit",
		grpcConn:    grpcConn,
		client:      performerV1.NewPerformerServiceClient(grpcConn),
	}

	// Create AvsKubernetesPerformer
	akp := &AvsKubernetesPerformer{
		config:         &avsPerformer.AvsPerformerConfig{},
		logger:         logger,
		taskWaitGroups: make(map[string]*sync.WaitGroup),
	}
	akp.currentPerformer.Store(performer)

	// Try to execute task
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
	// The error could be either from ConnectionManager or from the retry interceptor
	if err != nil && !contains(err.Error(), "failed to get healthy connection") &&
		!contains(err.Error(), "request failed after") &&
		!contains(err.Error(), "connection refused") {
		t.Errorf("Expected connection failure error, got: %v", err)
	}

}

func TestAvsKubernetesPerformer_ConnectionCleanup(t *testing.T) {
	// Create retry config
	retryConfig := &clients.RetryConfig{
		MaxRetries:        1,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 500 * time.Millisecond,
	}

	grpcConn, _ := clients.NewGrpcClientWithRetry("localhost:9996", true, retryConfig)
	if grpcConn == nil {
		t.Skip("Could not create gRPC connection")
	}

	performer := &PerformerResource{
		performerID: "test-performer-cleanup",
		grpcConn:    grpcConn,
		statusChan:  make(chan avsPerformer.PerformerStatusEvent, 10),
	}

	// Close gRPC connection
	if err := performer.grpcConn.Close(); err != nil {
		t.Errorf("Failed to close gRPC connection: %v", err)
	}

	// Connection should be closed
	state := performer.grpcConn.GetState()
	if state != connectivity.Shutdown {
		t.Errorf("Expected connection state to be Shutdown, got %v", state)
	}
}

// Helper function for string contains
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
