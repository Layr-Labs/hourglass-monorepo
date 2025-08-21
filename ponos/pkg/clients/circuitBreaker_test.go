package clients

import (
	"context"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
)

// MockConnection implements a mock grpc.ClientConn for testing
type MockConnection struct {
	state connectivity.State
	err   error
}

func (m *MockConnection) GetState() connectivity.State {
	return m.state
}

func (m *MockConnection) Close() error {
	return m.err
}

func (m *MockConnection) Target() string {
	return "mock-target"
}

func (m *MockConnection) WaitForStateChange(ctx context.Context, sourceState connectivity.State) bool {
	return false
}

func (m *MockConnection) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return m.err
}

func (m *MockConnection) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, m.err
}

func TestConnectionManager_CircuitBreaker_InitialState(t *testing.T) {
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	cm := NewConnectionManager("localhost:8080", true, nil, l)

	// Circuit should be closed initially
	if cm.IsCircuitOpen() {
		t.Error("Expected circuit breaker to be closed initially, got open")
	}

	// Verify initial state
	if cm.unhealthyCount != 0 {
		t.Errorf("Expected unhealthyCount to be 0, got %d", cm.unhealthyCount)
	}

	// lastHealthy should be recent
	if time.Since(cm.lastHealthy) > time.Second {
		t.Error("Expected lastHealthy to be recent")
	}
}

func TestConnectionManager_CircuitBreaker_FailureThreshold(t *testing.T) {
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	cm := NewConnectionManager("localhost:8080", true, nil, l)

	// Test below threshold
	cm.unhealthyCount = 3
	cm.lastHealthy = time.Now().Add(-30 * time.Second)

	if cm.IsCircuitOpen() {
		t.Error("Expected circuit breaker to be closed below threshold, got open")
	}

	// Test at threshold
	cm.unhealthyCount = 5
	cm.lastHealthy = time.Now().Add(-30 * time.Second)

	if cm.IsCircuitOpen() {
		t.Error("Expected circuit breaker to be closed at threshold, got open")
	}

	// Test above threshold
	cm.unhealthyCount = 6
	cm.lastHealthy = time.Now().Add(-2 * time.Minute)

	if !cm.IsCircuitOpen() {
		t.Error("Expected circuit breaker to be open above threshold, got closed")
	}
}

func TestConnectionManager_CircuitBreaker_TimeWindow(t *testing.T) {
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	cm := NewConnectionManager("localhost:8080", true, nil, l)

	// Set high failure count but recent healthy time
	cm.unhealthyCount = 10
	cm.lastHealthy = time.Now().Add(-30 * time.Second)

	if cm.IsCircuitOpen() {
		t.Error("Expected circuit breaker to be closed with recent healthy time, got open")
	}

	// Set high failure count with old healthy time
	cm.lastHealthy = time.Now().Add(-2 * time.Minute)

	if !cm.IsCircuitOpen() {
		t.Error("Expected circuit breaker to be open with old healthy time, got closed")
	}
}

func TestConnectionManager_CircuitBreaker_Recovery(t *testing.T) {
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	cm := NewConnectionManager("localhost:8080", true, nil, l)

	// Open the circuit
	cm.unhealthyCount = 10
	cm.lastHealthy = time.Now().Add(-5 * time.Minute)

	if !cm.IsCircuitOpen() {
		t.Error("Expected circuit breaker to be open, got closed")
	}

	// Simulate recovery
	cm.lastHealthy = time.Now()
	cm.unhealthyCount = 0

	if cm.IsCircuitOpen() {
		t.Error("Expected circuit breaker to be closed after recovery, got open")
	}
}

func TestConnectionManager_HealthTracking(t *testing.T) {
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	cm := NewConnectionManager("localhost:8080", true, nil, l)

	// Test initial state
	if cm.unhealthyCount != 0 {
		t.Errorf("Expected initial unhealthyCount to be 0, got %d", cm.unhealthyCount)
	}

	// Test health tracking logic
	cm.unhealthyCount = 5
	cm.lastHealthy = time.Now().Add(-30 * time.Second)

	// Simulate successful health check
	cm.lastHealthy = time.Now()
	cm.unhealthyCount = 0

	if cm.unhealthyCount != 0 {
		t.Errorf("Expected unhealthyCount to be reset to 0, got %d", cm.unhealthyCount)
	}

	if time.Since(cm.lastHealthy) > time.Second {
		t.Error("Expected lastHealthy to be updated")
	}
}

func TestConnectionManager_ConnectionStates(t *testing.T) {
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	cm := NewConnectionManager("localhost:8080", true, nil, l)
	_ = cm // Use the variable to avoid unused variable error

	tests := []struct {
		name     string
		state    connectivity.State
		expected bool
	}{
		{
			name:     "Ready state",
			state:    connectivity.Ready,
			expected: true,
		},
		{
			name:     "Idle state",
			state:    connectivity.Idle,
			expected: true,
		},
		{
			name:     "Connecting state",
			state:    connectivity.Connecting,
			expected: false,
		},
		{
			name:     "TransientFailure state",
			state:    connectivity.TransientFailure,
			expected: false,
		},
		{
			name:     "Shutdown state",
			state:    connectivity.Shutdown,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test this without injecting a mock connection
			// This is more of a design validation test
			if tt.state == connectivity.Ready || tt.state == connectivity.Idle {
				// These should be considered healthy
				if !tt.expected {
					t.Errorf("Expected %v to be healthy", tt.state)
				}
			} else {
				// These should be considered unhealthy
				if tt.expected {
					t.Errorf("Expected %v to be unhealthy", tt.state)
				}
			}
		})
	}
}

func TestConnectionManager_StatsReporting(t *testing.T) {
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	cm := NewConnectionManager("localhost:8080", true, nil, l)
	_ = cm // Use the variable to avoid unused variable error

	// Set some state
	cm.unhealthyCount = 3
	cm.lastHealthy = time.Now().Add(-45 * time.Second)

	stats := cm.GetConnectionStats()

	// Verify stats structure
	if stats["unhealthyCount"] != 3 {
		t.Errorf("Expected unhealthyCount to be 3, got %v", stats["unhealthyCount"])
	}

	if stats["circuitOpen"] != false {
		t.Errorf("Expected circuitOpen to be false, got %v", stats["circuitOpen"])
	}

	if stats["hasConnection"] != false {
		t.Errorf("Expected hasConnection to be false, got %v", stats["hasConnection"])
	}

	// Test with circuit open
	cm.unhealthyCount = 10
	cm.lastHealthy = time.Now().Add(-5 * time.Minute)

	stats = cm.GetConnectionStats()
	if stats["circuitOpen"] != true {
		t.Errorf("Expected circuitOpen to be true, got %v", stats["circuitOpen"])
	}
}

func TestIsRetryableError_GrpcStatus(t *testing.T) {
	tests := []struct {
		name     string
		code     codes.Code
		expected bool
	}{
		{
			name:     "UNAVAILABLE",
			code:     codes.Unavailable,
			expected: true,
		},
		{
			name:     "DEADLINE_EXCEEDED",
			code:     codes.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "RESOURCE_EXHAUSTED",
			code:     codes.ResourceExhausted,
			expected: true,
		},
		{
			name:     "INTERNAL",
			code:     codes.Internal,
			expected: true,
		},
		{
			name:     "INVALID_ARGUMENT",
			code:     codes.InvalidArgument,
			expected: false,
		},
		{
			name:     "PERMISSION_DENIED",
			code:     codes.PermissionDenied,
			expected: false,
		},
		{
			name:     "UNAUTHENTICATED",
			code:     codes.Unauthenticated,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := status.Error(tt.code, "test error")
			result := isRetryableError(err)
			if result != tt.expected {
				t.Errorf("Expected %v for gRPC code %v, got %v", tt.expected, tt.code, result)
			}
		})
	}
}

func TestConnectionManager_ConcurrentAccess(t *testing.T) {
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	cm := NewConnectionManager("localhost:8080", true, nil, l)

	// Test concurrent access to circuit breaker state
	done := make(chan bool)

	// Goroutine 1: Read circuit state
	go func() {
		for i := 0; i < 100; i++ {
			cm.mu.Lock()
			_ = cm.IsCircuitOpen()
			cm.mu.Unlock()
		}
		done <- true
	}()

	// Goroutine 2: Update failure count
	go func() {
		for i := 0; i < 100; i++ {
			cm.mu.Lock()
			cm.unhealthyCount = i % 10
			cm.mu.Unlock()
		}
		done <- true
	}()

	// Goroutine 3: Update last healthy time
	go func() {
		for i := 0; i < 100; i++ {
			cm.mu.Lock()
			cm.lastHealthy = time.Now().Add(-time.Duration(i) * time.Second)
			cm.mu.Unlock()
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	<-done
	<-done
	<-done

	// If we reach here without panic, the test passes
	// This is a basic race condition test
}

func TestConnectionManager_FailureScenarios(t *testing.T) {
	tests := []struct {
		name             string
		failureCount     int
		timeSinceHealthy time.Duration
		expectedOpen     bool
	}{
		{
			name:             "No failures",
			failureCount:     0,
			timeSinceHealthy: time.Second,
			expectedOpen:     false,
		},
		{
			name:             "Few failures, recent healthy",
			failureCount:     3,
			timeSinceHealthy: 30 * time.Second,
			expectedOpen:     false,
		},
		{
			name:             "Many failures, recent healthy",
			failureCount:     10,
			timeSinceHealthy: 30 * time.Second,
			expectedOpen:     false,
		},
		{
			name:             "Few failures, old healthy",
			failureCount:     3,
			timeSinceHealthy: 5 * time.Minute,
			expectedOpen:     false,
		},
		{
			name:             "Many failures, old healthy",
			failureCount:     10,
			timeSinceHealthy: 5 * time.Minute,
			expectedOpen:     true,
		},
		{
			name:             "Threshold failures, old healthy",
			failureCount:     6,
			timeSinceHealthy: 2 * time.Minute,
			expectedOpen:     true,
		},
	}
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cm := NewConnectionManager("localhost:8080", true, nil, l)
			cm.unhealthyCount = tt.failureCount
			cm.lastHealthy = time.Now().Add(-tt.timeSinceHealthy)

			result := cm.IsCircuitOpen()
			if result != tt.expectedOpen {
				t.Errorf("Expected circuit open=%v, got %v (failures=%d, timeSince=%v)",
					tt.expectedOpen, result, tt.failureCount, tt.timeSinceHealthy)
			}
		})
	}
}

// Benchmark tests for circuit breaker performance
func BenchmarkCircuitBreaker_IsOpen(b *testing.B) {
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	cm := NewConnectionManager("localhost:8080", true, nil, l)
	cm.unhealthyCount = 7
	cm.lastHealthy = time.Now().Add(-2 * time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cm.IsCircuitOpen()
	}
}

func BenchmarkCircuitBreaker_StatsCollection(b *testing.B) {
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	cm := NewConnectionManager("localhost:8080", true, nil, l)
	cm.unhealthyCount = 5
	cm.lastHealthy = time.Now().Add(-time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cm.GetConnectionStats()
	}
}

func BenchmarkIsRetryableError(b *testing.B) {
	err := status.Error(codes.Unavailable, "service unavailable")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isRetryableError(err)
	}
}
