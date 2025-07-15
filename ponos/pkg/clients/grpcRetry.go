package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"strings"
	"time"

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

// GrpcClientWithRetry creates a gRPC client with retry logic
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

	opts := []grpc.DialOption{
		creds,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(math.MaxInt32)),
	}

	var conn *grpc.ClientConn
	var err error

	delay := retryConfig.InitialDelay
	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), retryConfig.ConnectionTimeout)
		
		conn, err = grpc.NewClient(url, opts...)
		if err == nil {
			// Wait for connection to be ready with timeout
			waitCtx, waitCancel := context.WithTimeout(ctx, retryConfig.ConnectionTimeout)
			if conn.WaitForStateChange(waitCtx, connectivity.Idle) {
				if testConnection(conn) {
					waitCancel()
					cancel()
					return conn, nil
				}
			}
			waitCancel()
			
			// Connection failed test, close and retry
			conn.Close()
			err = fmt.Errorf("connection test failed")
		}
		cancel()

		if attempt == retryConfig.MaxRetries {
			break
		}

		// Exponential backoff with jitter
		time.Sleep(delay)
		delay = time.Duration(float64(delay) * retryConfig.BackoffMultiplier)
		if delay > retryConfig.MaxDelay {
			delay = retryConfig.MaxDelay
		}
	}

	return nil, fmt.Errorf("failed to establish gRPC connection after %d attempts: %w", retryConfig.MaxRetries+1, err)
}

// testConnection tests if the gRPC connection is healthy
func testConnection(conn *grpc.ClientConn) bool {
	if conn == nil {
		return false
	}

	state := conn.GetState()
	return state == connectivity.Ready || state == connectivity.Idle
}

// ConnectionManager manages gRPC connections with retry and health monitoring
type ConnectionManager struct {
	url          string
	insecureConn bool
	retryConfig  *RetryConfig
	conn         *grpc.ClientConn
	lastHealthy  time.Time
	unhealthyCount int
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(url string, insecureConn bool, retryConfig *RetryConfig) *ConnectionManager {
	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	}

	return &ConnectionManager{
		url:          url,
		insecureConn: insecureConn,
		retryConfig:  retryConfig,
		lastHealthy:  time.Now(),
	}
}

// GetConnection returns a healthy connection, creating or reconnecting if necessary
func (cm *ConnectionManager) GetConnection() (*grpc.ClientConn, error) {
	if cm.conn != nil && cm.isConnectionHealthy() {
		return cm.conn, nil
	}

	// Connection is unhealthy or doesn't exist, create a new one
	if cm.conn != nil {
		cm.conn.Close()
	}

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
func (cm *ConnectionManager) isConnectionHealthy() bool {
	if cm.conn == nil {
		return false
	}

	state := cm.conn.GetState()
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
func (cm *ConnectionManager) IsCircuitOpen() bool {
	// Simple circuit breaker: open if we've had more than 5 consecutive failures
	// and the last healthy time was more than 1 minute ago
	return cm.unhealthyCount > 5 && time.Since(cm.lastHealthy) > time.Minute
}

// GetConnectionStats returns connection statistics
func (cm *ConnectionManager) GetConnectionStats() map[string]interface{} {
	state := connectivity.Shutdown
	if cm.conn != nil {
		state = cm.conn.GetState()
	}

	return map[string]interface{}{
		"state":           state.String(),
		"lastHealthy":     cm.lastHealthy,
		"unhealthyCount":  cm.unhealthyCount,
		"circuitOpen":     cm.IsCircuitOpen(),
		"hasConnection":   cm.conn != nil,
	}
}

// Close closes the connection manager
func (cm *ConnectionManager) Close() error {
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