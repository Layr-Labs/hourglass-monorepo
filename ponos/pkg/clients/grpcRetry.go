package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// RetryConfig contains configuration for connection retry behavior
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// InitialDelay is the initial delay between retries
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// BackoffMultiplier is the multiplier for exponential backoff
	BackoffMultiplier float64
	// ConnectionTimeout is the timeout for establishing a connection
	ConnectionTimeout time.Duration
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:        5,
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 10 * time.Second,
	}
}

// NewGrpcClientWithRetry creates a gRPC client without forcing connection
// The actual connection will be established lazily on first use
func NewGrpcClientWithRetry(url string, insecureConn bool, retryConfig *RetryConfig) (*grpc.ClientConn, error) {
	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	}

	var creds grpc.DialOption
	if strings.Contains(url, "localhost:") || strings.Contains(url, "127.0.0.1:") || insecureConn {
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		creds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: false}))
	}

	// Add retry interceptor for unary calls
	opts := []grpc.DialOption{
		creds,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(math.MaxInt32)),
		grpc.WithUnaryInterceptor(retryUnaryInterceptor(retryConfig)),
	}

	// Simply create the lazy connection - don't test it
	conn, err := grpc.NewClient(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return conn, nil
}

// retryUnaryInterceptor creates a unary interceptor that retries failed requests
func retryUnaryInterceptor(config *RetryConfig) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var err error
		delay := config.InitialDelay

		for attempt := 0; attempt <= config.MaxRetries; attempt++ {
			// Create a context with timeout for this attempt
			attemptCtx, cancel := context.WithTimeout(ctx, config.ConnectionTimeout)
			err = invoker(attemptCtx, method, req, reply, cc, opts...)
			cancel()

			// If successful or non-retryable error, return immediately
			if err == nil || !isRetryableError(err) {
				return err
			}

			// Don't retry if we've hit max attempts
			if attempt == config.MaxRetries {
				break
			}

			// Don't retry if the parent context is cancelled
			if ctx.Err() != nil {
				return ctx.Err()
			}

			// Exponential backoff
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * config.BackoffMultiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}

		return fmt.Errorf("request failed after %d attempts: %w", config.MaxRetries+1, err)
	}
}

// testConnection tests if the gRPC connection is healthy
func testConnection(conn *grpc.ClientConn) bool {
	if conn == nil {
		return false
	}

	state := conn.GetState()
	// Only consider Ready state as healthy for active connections
	// Idle means the connection hasn't been established yet
	return state == connectivity.Ready
}

// ConnectionManager manages gRPC connections with retry and health monitoring
type ConnectionManager struct {
	url            string
	insecureConn   bool
	retryConfig    *RetryConfig
	conn           *grpc.ClientConn
	lastHealthy    time.Time
	unhealthyCount int
	logger         *zap.Logger

	// mu protects all mutable fields (conn, lastHealthy, unhealthyCount)
	mu sync.Mutex
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(url string, insecureConn bool, retryConfig *RetryConfig, l *zap.Logger) *ConnectionManager {
	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	}

	return &ConnectionManager{
		url:          url,
		insecureConn: insecureConn,
		retryConfig:  retryConfig,
		lastHealthy:  time.Now(),
		logger:       l,
	}
}

// GetConnection returns a healthy connection, creating or reconnecting if necessary
// LOCKING: This method ACQUIRES the mutex lock for the entire operation
func (cm *ConnectionManager) GetConnection() (*grpc.ClientConn, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.conn != nil && cm.isConnectionHealthy() {
		cm.logger.Sugar().Infow("Connection is healthy",
			zap.String("url", cm.url),
			zap.String("state", cm.conn.GetState().String()),
		)
		return cm.conn, nil
	}

	// Connection is unhealthy or doesn't exist, create a new one
	if cm.conn != nil {
		cm.conn.Close()
	}

	cm.logger.Sugar().Infow("Creating new gRPC connection",
		zap.String("url", cm.url),
		zap.Bool("insecureConn", cm.insecureConn),
	)
	conn, err := NewGrpcClientWithRetry(cm.url, cm.insecureConn, cm.retryConfig)
	if err != nil {
		cm.unhealthyCount++
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	cm.conn = conn
	cm.lastHealthy = time.Now()
	cm.unhealthyCount = 0
	return cm.conn, nil
}

// isConnectionHealthy checks if the current connection is healthy
// LOCKING: This method ASSUMES the caller already holds cm.mu lock (does NOT acquire lock)
func (cm *ConnectionManager) isConnectionHealthy() bool {

	if cm.conn == nil {
		return false
	}

	state := cm.conn.GetState()
	// Consider Ready or Idle as acceptable states
	// Idle is fine for lazy connections that will connect on first use
	isHealthy := state == connectivity.Ready || state == connectivity.Idle

	if isHealthy {
		cm.lastHealthy = time.Now()
		cm.unhealthyCount = 0
	} else {
		cm.unhealthyCount++
	}

	return isHealthy
}

// IsCircuitOpen returns true if the circuit breaker is open (too many failures)
// LOCKING: This method ASSUMES the caller already holds cm.mu lock (does NOT acquire lock)
func (cm *ConnectionManager) IsCircuitOpen() bool {

	// Simple circuit breaker: open if we've had more than 5 consecutive failures
	// and the last healthy time was more than 1 minute ago
	return cm.unhealthyCount > 5 && time.Since(cm.lastHealthy) > time.Minute
}

// GetConnectionStats returns connection statistics
// LOCKING: This method ACQUIRES the mutex lock for the entire operation
func (cm *ConnectionManager) GetConnectionStats() map[string]interface{} {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	state := connectivity.Shutdown
	if cm.conn != nil {
		state = cm.conn.GetState()
	}

	return map[string]interface{}{
		"state":          state.String(),
		"lastHealthy":    cm.lastHealthy,
		"unhealthyCount": cm.unhealthyCount,
		"circuitOpen":    cm.IsCircuitOpen(),
		"hasConnection":  cm.conn != nil,
	}
}

// Close closes the connection manager
// LOCKING: This method ACQUIRES the mutex lock for the entire operation
func (cm *ConnectionManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.conn != nil {
		err := cm.conn.Close()
		cm.conn = nil
		return err
	}
	return nil
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check gRPC status codes
	if grpcStatus, ok := status.FromError(err); ok {
		switch grpcStatus.Code() {
		case 14: // UNAVAILABLE
			return true
		case 4: // DEADLINE_EXCEEDED
			return true
		case 8: // RESOURCE_EXHAUSTED
			return true
		case 13: // INTERNAL
			return true
		}
	}

	// Check for connection errors
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "connection timeout") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "network is unreachable")
}
