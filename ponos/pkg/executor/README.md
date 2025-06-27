# Executor Component

The executor is a core service in the Ponos system that manages and executes tasks for AVS (Actively Validated Services) performers. It provides a secure, containerized execution environment for AVS workloads with comprehensive lifecycle management.

## Overview

The executor acts as an orchestration layer that:
- Receives task submissions from aggregators via gRPC
- Manages containerized AVS performer deployments
- Executes tasks within isolated performer containers
- Handles health monitoring and lifecycle management
- Signs task results with operator keys before returning them

## Architecture

The executor follows a modular architecture with several key components:

- **gRPC Server**: Exposes the ExecutorService API for external communication
- **Container Manager**: Handles Docker container lifecycle operations
- **AVS Performers**: Interface-based design supporting different performer types
- **Signer**: Handles BLS/ECDSA signing operations for task results

## API Reference

The executor implements a gRPC-based `ExecutorService` with four main APIs:

### SubmitTask
Executes a task for a specific AVS performer.

```protobuf
rpc SubmitTask(TaskSubmission) returns (TaskResponse)
```

- Validates task signatures from aggregators
- Routes tasks to appropriate AVS performers
- Signs and returns task results
- Tracks in-flight tasks for monitoring

### DeployArtifact
Deploys new container images for AVS performers with zero-downtime.

```protobuf
rpc DeployArtifact(DeployArtifactRequest) returns (DeployArtifactResponse)
```

- Implements blue-green deployment pattern
- Performs health checking before promotion
- Supports rollback on deployment failures
- Prevents concurrent deployments per AVS

### ListPerformers
Lists all active performers and their current status.

```protobuf
rpc ListPerformers(ListPerformersRequest) returns (ListPerformersResponse)
```

- Returns performer health status
- Shows deployment information
- Supports filtering by AVS address

### RemovePerformer
Removes a specific performer and cleans up resources.

```protobuf
rpc RemovePerformer(RemovePerformerRequest) returns (RemovePerformerResponse)
```

## Configuration

The executor is configured via YAML with the following structure:

```yaml
grpcPort: 8080
performerNetworkName: "performer-network"
operator:
  address: "0x123..."
  signingKeys:
    bls:
      keystore:
        publicKey: "0x..."
        encryptedSecretKey: "..."
      password: "keystore-password"
avsPerformers:
  - avsAddress: "0xAVS123..."
    processType: "server"
    image:
      repository: "registry.io/avs/performer"
      tag: "v1.0.0"
    workerCount: 1
    signingCurve: "bn254"
    avsRegistrarAddress: "0xRegistrar123..."
    health:
      checkInterval: "30s"
      timeout: "5s"
```

### Configuration Fields

- **grpcPort**: Port for the gRPC server
- **performerNetworkName**: Docker network name for performer isolation
- **operator**: Operator identity and signing keys
- **avsPerformers**: List of AVS performers to manage
  - **processType**: Currently supports "server" type
  - **image**: Docker image configuration
  - **workerCount**: Number of parallel workers
  - **signingCurve**: "bn254" or "bls381"
  - **health**: Health check configuration

## Usage

### Starting the Executor

```go
// Create executor with dependencies
executor := NewExecutor(
    config,
    rpcServer,
    logger,
    signer,
    peeringFetcher,
)

// Initialize and start
ctx := context.Background()
if err := executor.Initialize(ctx); err != nil {
    log.Fatal(err)
}

if err := executor.Run(ctx); err != nil {
    log.Fatal(err)
}
```

### Client Integration

```go
// Connect to executor
conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
client := executorV1.NewExecutorServiceClient(conn)

// Submit a task
result, err := client.SubmitTask(ctx, &executorV1.TaskSubmission{
    TaskId:     "task-123",
    AvsAddress: "0xAVS123...",
    Payload:    []byte("task data"),
    Signature:  signature,
})

// Deploy new version
resp, err := client.DeployArtifact(ctx, &executorV1.DeployArtifactRequest{
    AvsAddress:  "0xAVS123...",
    RegistryUrl: "registry.io/avs/performer",
    Digest:      "sha256:abc123...",
})

// Monitor performers
performers, err := client.ListPerformers(ctx, &executorV1.ListPerformersRequest{})
```

## Key Features

### Container Management
- **Isolation**: Each AVS runs in isolated Docker containers
- **Lifecycle**: Full container lifecycle management (create, start, stop, remove)
- **Networking**: Custom Docker network for performer communication
- **Health Checks**: Both container and application-level health monitoring

### Deployment Management
- **Blue-Green**: Zero-downtime deployments with staged containers
- **Health Validation**: New deployments must pass health checks before promotion
- **Rollback**: Automatic rollback on deployment failures
- **Concurrency Control**: Prevents concurrent deployments for the same AVS

### Task Execution
- **Parallel Processing**: Configurable worker pools per AVS
- **Signature Validation**: Validates aggregator signatures on tasks
- **Result Signing**: Signs task results with operator keys
- **In-Flight Tracking**: Monitors active tasks for observability

### Security
- **BLS Signatures**: Support for BN254 and BLS381 curves
- **Keystore Management**: Encrypted key storage with password protection
- **Network Isolation**: Performers run in isolated Docker networks
- **Signed Results**: All task results are cryptographically signed

## Monitoring and Operations

### Logging
The executor uses structured logging with zap. Key log fields include:
- `avsAddress`: AVS identifier
- `taskId`: Task identifier
- `deploymentId`: Deployment identifier
- `performerId`: Container identifier

### Health Monitoring
- Container health status (healthy, unhealthy, starting)
- Application-level health checks via performer API
- Deployment health validation before promotion
- Configurable check intervals and timeouts

### Graceful Shutdown
The executor supports graceful shutdown:
1. Stops accepting new tasks
2. Waits for in-flight tasks to complete
3. Gracefully stops all performer containers
4. Cleans up resources

## Development

### Building
```bash
make build/executor
```

### Testing
```bash
make test
```

### Running Locally
```bash
make run/executor ARGS="--config config.yaml"
```

## Troubleshooting

### Common Issues

1. **Container Network Issues**
   - Ensure Docker daemon is running
   - Check network name conflicts
   - Verify performer network exists

2. **Deployment Failures**
   - Check image accessibility
   - Verify digest format
   - Review container logs
   - Check health endpoint configuration

3. **Task Execution Failures**
   - Verify performer is healthy
   - Check task payload format
   - Review performer logs
   - Validate signatures

### Debug Commands

```bash
# List performer containers
docker ps -f label=com.eigenlayer.hourglass.avs-address

# Check performer logs
docker logs <performer-id>

# Inspect performer network
docker network inspect performer-network

# Check executor logs
tail -f executor.log | jq .
```

## Integration with Ponos

The executor integrates with other Ponos components:

- **Aggregator**: Sends tasks and receives signed results
- **Peering**: Fetches operator peering data for validation
- **Signer**: Uses shared signing infrastructure
- **RPC Server**: Shares gRPC server implementation

See the [Ponos documentation](../../../README.md) for more details on the overall architecture.