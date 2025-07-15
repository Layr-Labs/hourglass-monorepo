package clients

import (
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/test/bufconn"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	
	if config.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries to be 5, got %d", config.MaxRetries)
	}
	if config.InitialDelay != 1*time.Second {
		t.Errorf("Expected InitialDelay to be 1s, got %v", config.InitialDelay)
	}
	if config.MaxDelay != 30*time.Second {
		t.Errorf("Expected MaxDelay to be 30s, got %v", config.MaxDelay)
	}
	if config.BackoffMultiplier != 2.0 {
		t.Errorf("Expected BackoffMultiplier to be 2.0, got %f", config.BackoffMultiplier)
	}
	if config.ConnectionTimeout != 10*time.Second {
		t.Errorf("Expected ConnectionTimeout to be 10s, got %v", config.ConnectionTimeout)
	}
}

func TestNewGrpcClientWithRetry_Success(t *testing.T) {
	// Create a test server
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()
	defer server.Stop()
	
	// Test with bufconn (which should work immediately)
	config := &RetryConfig{
		MaxRetries:        2,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 5 * time.Second,
	}
	
	// This will fail because we can't actually connect to bufconn with regular dial
	// But we can test the retry logic with a mock server
	_, err := NewGrpcClientWithRetry("localhost:0", true, config)
	if err == nil {
		t.Error("Expected error for invalid address, got nil")
	}
	
	// Test that error contains retry information
	if err != nil && !contains(err.Error(), "failed to establish gRPC connection after") {
		t.Errorf("Expected retry error message, got: %v", err)
	}
}

func TestNewGrpcClientWithRetry_InvalidAddress(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:        2,
		InitialDelay:      10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 100 * time.Millisecond,
	}
	
	_, err := NewGrpcClientWithRetry("invalid-address:99999", true, config)
	if err == nil {
		t.Error("Expected error for invalid address, got nil")
	}
	
	// Should contain retry count
	if !contains(err.Error(), "failed to establish gRPC connection after 3 attempts") {
		t.Errorf("Expected retry count in error, got: %v", err)
	}
}

func TestNewGrpcClientWithRetry_NilConfig(t *testing.T) {
	_, err := NewGrpcClientWithRetry("localhost:0", true, nil)
	if err == nil {
		t.Error("Expected error for invalid address, got nil")
	}
	
	// Should use default config (5 retries + 1 initial = 6 attempts)
	if !contains(err.Error(), "failed to establish gRPC connection after 6 attempts") {
		t.Errorf("Expected default retry count in error, got: %v", err)
	}
}

func TestConnectionManager_NewConnectionManager(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:        3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 5 * time.Second,
	}
	
	cm := NewConnectionManager("localhost:8080", true, config)
	
	if cm.url != "localhost:8080" {
		t.Errorf("Expected URL to be 'localhost:8080', got %s", cm.url)
	}
	if cm.insecureConn != true {
		t.Errorf("Expected insecureConn to be true, got %v", cm.insecureConn)
	}
	if cm.retryConfig != config {
		t.Errorf("Expected retryConfig to be set, got %v", cm.retryConfig)
	}
	if cm.conn != nil {
		t.Errorf("Expected conn to be nil initially, got %v", cm.conn)
	}
}

func TestConnectionManager_NewConnectionManager_NilConfig(t *testing.T) {
	cm := NewConnectionManager("localhost:8080", true, nil)
	
	if cm.retryConfig == nil {
		t.Error("Expected retryConfig to be set to default, got nil")
	}
	if cm.retryConfig.MaxRetries != 5 {
		t.Errorf("Expected default MaxRetries to be 5, got %d", cm.retryConfig.MaxRetries)
	}
}

func TestConnectionManager_GetConnection_Failure(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:        1,
		InitialDelay:      10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 100 * time.Millisecond,
	}
	
	cm := NewConnectionManager("localhost:0", true, config)
	
	_, err := cm.GetConnection()
	if err == nil {
		t.Error("Expected error for invalid address, got nil")
	}
	
	if !contains(err.Error(), "failed to create connection") {
		t.Errorf("Expected connection creation error, got: %v", err)
	}
}

func TestConnectionManager_IsConnectionHealthy(t *testing.T) {
	cm := NewConnectionManager("localhost:8080", true, nil)
	
	// Test with nil connection
	if cm.isConnectionHealthy() {
		t.Error("Expected false for nil connection, got true")
	}
	
	// Test with mock connection (this is limited without actual connection)
	if cm.conn != nil {
		t.Error("Expected conn to be nil initially")
	}
}

func TestConnectionManager_IsCircuitOpen(t *testing.T) {
	cm := NewConnectionManager("localhost:8080", true, nil)
	
	// Initially circuit should be closed
	if cm.IsCircuitOpen() {
		t.Error("Expected circuit to be closed initially, got open")
	}
	
	// Simulate failures
	cm.unhealthyCount = 6
	cm.lastHealthy = time.Now().Add(-2 * time.Minute)
	
	if !cm.IsCircuitOpen() {
		t.Error("Expected circuit to be open after failures, got closed")
	}
	
	// Test with recent healthy time
	cm.lastHealthy = time.Now()
	if cm.IsCircuitOpen() {
		t.Error("Expected circuit to be closed with recent healthy time, got open")
	}
}

func TestConnectionManager_GetConnectionStats(t *testing.T) {
	cm := NewConnectionManager("localhost:8080", true, nil)
	
	stats := cm.GetConnectionStats()
	
	if stats["state"] != connectivity.Shutdown.String() {
		t.Errorf("Expected state to be %s, got %v", connectivity.Shutdown.String(), stats["state"])
	}
	if stats["unhealthyCount"] != 0 {
		t.Errorf("Expected unhealthyCount to be 0, got %v", stats["unhealthyCount"])
	}
	if stats["circuitOpen"] != false {
		t.Errorf("Expected circuitOpen to be false, got %v", stats["circuitOpen"])
	}
	if stats["hasConnection"] != false {
		t.Errorf("Expected hasConnection to be false, got %v", stats["hasConnection"])
	}
}

func TestConnectionManager_Close(t *testing.T) {
	cm := NewConnectionManager("localhost:8080", true, nil)
	
	// Test closing without connection
	err := cm.Close()
	if err != nil {
		t.Errorf("Expected no error closing without connection, got %v", err)
	}
	
	// Test that connection is nil after close
	if cm.conn != nil {
		t.Error("Expected conn to be nil after close")
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "connection reset",
			err:      errors.New("connection reset"),
			expected: true,
		},
		{
			name:     "connection timeout",
			err:      errors.New("connection timeout"),
			expected: true,
		},
		{
			name:     "no such host",
			err:      errors.New("no such host"),
			expected: true,
		},
		{
			name:     "network is unreachable",
			err:      errors.New("network is unreachable"),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      errors.New("permission denied"),
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v for error %v, got %v", tt.expected, tt.err, result)
			}
		})
	}
}

func TestTestConnection(t *testing.T) {
	// Test with nil connection
	if testConnection(nil) {
		t.Error("Expected false for nil connection, got true")
	}
	
	// We can't easily test with real connections without a lot of setup
	// This test mainly ensures the function doesn't panic with nil input
}

// Helper function to check if string contains substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || containsHelper(str, substr))
}

func containsHelper(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkNewConnectionManager(b *testing.B) {
	config := DefaultRetryConfig()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = NewConnectionManager("localhost:8080", true, config)
	}
}

func BenchmarkIsCircuitOpen(b *testing.B) {
	cm := NewConnectionManager("localhost:8080", true, nil)
	cm.unhealthyCount = 3
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = cm.IsCircuitOpen()
	}
}

func BenchmarkGetConnectionStats(b *testing.B) {
	cm := NewConnectionManager("localhost:8080", true, nil)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = cm.GetConnectionStats()
	}
}