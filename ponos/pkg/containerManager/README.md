# Container Manager Tests

This package provides comprehensive tests for the containerManager module.

## Running Tests

### Unit Tests Only (Recommended)
Run the core unit tests that don't require Docker:

```bash
go test -short -v -run="TestHashAvsAddress|TestCreateDefaultContainerConfig|TestGetContainerEndpoint|TestContainerConfig|TestHealthCheckConfig|TestRestartPolicy|TestResourceThresholds|TestLivenessConfig|TestContainerEventType|TestContainerEvent|TestResourceUsage"
```

### All Unit Tests (Basic Container Manager functionality)
```bash
go test -short -v
```

### Integration Tests (Requires Docker)
Run tests that require a running Docker daemon:

```bash
go test -v -run="Integration"
```

**Note:** Integration tests will be skipped if Docker is not available.

### Benchmarks
Run performance benchmarks:

```bash
go test -bench=BenchmarkHashAvsAddress -benchmem
go test -bench=BenchmarkCreateDefaultContainerConfig -benchmem  
go test -bench=BenchmarkContainerStateCreation -benchmem
```

### Test Coverage
Generate coverage report:

```bash
go test -short -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Test Structure

- **`dockerManager_test.go`** - Tests for core Docker container manager functionality
- **`utils_test.go`** - Tests for utility functions (hashing, config creation, endpoints)
- **`types_test.go`** - Tests for type definitions and validation
- **`integration_test.go`** - Integration tests requiring Docker
- **`benchmark_test.go`** - Performance benchmarks
- **`suite_test.go`** - Comprehensive test suite using testify/suite

## Features Tested

### Core Functionality
- ✅ Container manager creation and configuration  
- ✅ Liveness monitoring setup and teardown
- ✅ Health check operations
- ✅ Restart policy management
- ✅ Manual restart triggers
- ✅ Resource cleanup on shutdown

### Utility Functions  
- ✅ AVS address hashing
- ✅ Default container configuration creation
- ✅ Container endpoint resolution
- ✅ Network endpoint handling (custom vs bridge networks)

### Configuration Types
- ✅ Container configuration validation
- ✅ Health check configuration defaults
- ✅ Restart policy configuration
- ✅ Resource threshold validation
- ✅ Liveness configuration completeness

### Event System
- ✅ Container event type constants
- ✅ Event creation and structure
- ✅ Resource usage calculation

### Integration (with Docker)
- ✅ Real container lifecycle management
- ✅ Network creation and management  
- ✅ Container monitoring with actual containers
- ✅ Health check integration
- ✅ Default configuration validation

## Performance Benchmarks

The benchmark suite measures:

- **Hash Operations** - AVS address hashing performance
- **Configuration Creation** - Container config generation speed
- **Manager Creation** - Container manager instantiation overhead
- **Monitoring Setup** - Liveness monitoring setup performance
- **Memory Usage** - Memory allocation patterns
- **Concurrent Operations** - Performance with multiple monitors

## Requirements

- **Go 1.23+**
- **Docker** (for integration tests only)
- **Dependencies:**
  - `github.com/stretchr/testify` - Test assertions and mocking
  - `go.uber.org/zap` - Logging in tests
  - `github.com/docker/docker` - Docker client
  - `github.com/docker/go-connections` - Docker networking utilities

## Test Philosophy

The tests are designed to:

1. **Test core logic without external dependencies** - Most tests focus on the container manager's internal logic rather than Docker integration
2. **Graceful degradation** - Integration tests are skipped if Docker is unavailable  
3. **Comprehensive coverage** - Tests cover happy paths, error cases, edge cases, and concurrent operations
4. **Performance validation** - Benchmarks ensure the container manager performs acceptably under load
5. **Race condition detection** - Tests run with `-race` flag to detect concurrency issues

## Known Issues

- Some race conditions may occur during test teardown due to goroutine cleanup timing
- Integration tests require Docker daemon to be running and accessible
- Benchmark tests may show variability based on system performance

## Example Usage

```go
// Create a container manager
dcm, err := NewDockerContainerManager(nil, logger)
require.NoError(t, err)
defer dcm.Shutdown(context.Background())

// Start liveness monitoring
config := &LivenessConfig{
    HealthCheckConfig: HealthCheckConfig{
        Enabled:          true,
        Interval:         5 * time.Second,
        FailureThreshold: 3,
    },
    RestartPolicy: RestartPolicy{
        Enabled:     true,
        MaxRestarts: 5,
    },
    ResourceMonitoring: true,
}

eventChan, err := dcm.StartLivenessMonitoring(ctx, containerID, config)
require.NoError(t, err)

// Monitor for events
select {
case event := <-eventChan:
    // Handle container event
default:
    // No events
}
```