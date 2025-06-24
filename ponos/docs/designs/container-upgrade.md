# Container Upgrade Design

## Overview

This document describes the design for scheduled container deployment in the Hourglass AVS performer system. The system allows deploying new performer containers with scheduled activation times, maintaining both current active containers and a queue of pending containers for future activation.

## Core Architecture

### Container Metadata Store

The system maintains a map of containers with comprehensive metadata:

```go
type ContainerMetadata struct {
    Info             *containerManager.ContainerInfo
    Image            containerManager.ImageConfig
    ActivationTime   int64                              // Unix timestamp
    Status           ContainerStatus                    // Pending/Active/Expired
    Client           performerV1.PerformerServiceClient
    Endpoint         string
    DeploymentID     string                             // Track deployment requests
    CreatedAt        time.Time
    ActivatedAt      *time.Time                         // nil if not yet activated
    LastHealthCheck  time.Time
    RegistryURL      string                             // From gRPC request
    ArtifactDigest   string                             // From gRPC request
}

type ContainerStatus int
const (
    ContainerStatusPending  // Created, waiting for activation time
    ContainerStatusActive   // Currently processing tasks
    ContainerStatusExpired  // Superseded, marked for cleanup
)
```

Enhanced `AvsPerformerServer` struct:

```go
type AvsPerformerServer struct {
    // Existing fields...
    containers       map[string]*ContainerMetadata      // containerID -> metadata
    currentContainer string                              // ID of active container
    containerMu      sync.RWMutex                       // Protects container operations
}
```

## Lazy Activation Strategy

### No Background Scheduler

The system uses lazy evaluation instead of background schedulers for simplicity and efficiency:

- **Evaluation point**: During `RunTask()` processing
- **Simple logic**: Check if any pending container should now be active
- **Resource efficient**: Only evaluates when actually needed

```go
func (aps *AvsPerformerServer) RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error) {
    // 1. Lazy evaluation: activate pending containers if deadline reached
    aps.checkAndActivatePendingContainer(time.Now().Unix())
    
    // 2. Process task with current active container
    activeContainer := aps.getActiveContainer()
    res, err := activeContainer.Client.ExecuteTask(ctx, &performerV1.TaskRequest{...})
    return performerTask.NewTaskResultFromResultProto(res), err
}

func (aps *AvsPerformerServer) checkAndActivatePendingContainer(currentTime int64) {
    aps.containerMu.Lock()
    defer aps.containerMu.Unlock()
    
    // Find the container that should be active now
    var targetContainer *ContainerMetadata
    for _, container := range aps.containers {
        if container.ActivationTime <= currentTime && 
           container.Status == ContainerStatusPending {
            if targetContainer == nil || 
               container.ActivationTime > targetContainer.ActivationTime {
                targetContainer = container
            }
        }
    }
    
    if targetContainer != nil && targetContainer.Info.ID != aps.currentContainer {
        aps.activateContainer(targetContainer.Info.ID)
    }
}
```

## Container Reaping Strategy

### Health Check Integration

Container cleanup is handled by piggybacking on the existing application health check:

- **No additional goroutines**: Reuses existing health check goroutine
- **Frequency**: Every 60 health check cycles (60 seconds)
- **Grace period**: 1 hour for pending containers past their activation time

```go
func (aps *AvsPerformerServer) startApplicationHealthCheck(ctx context.Context) {
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()

        consecutiveFailures := 0
        reapCounter := 0
        const reapInterval = 60 // Reap every 60 seconds
        
        for {
            select {
            case <-healthCtx.Done():
                return
            case <-ticker.C:
                // 1. Periodic container reaping
                reapCounter++
                if reapCounter >= reapInterval {
                    aps.reapExpiredContainers()
                    reapCounter = 0
                }
                
                // 2. Continue with existing health check logic...
                if aps.performerClient != nil {
                    res, err := aps.performerClient.HealthCheck(healthCtx, &performerV1.HealthCheckRequest{})
                    // ... existing health check logic
                }
            }
        }
    }()
}

func (aps *AvsPerformerServer) reapExpiredContainers() {
    aps.containerMu.Lock()
    defer aps.containerMu.Unlock()
    
    currentTime := time.Now().Unix()
    expiredContainers := []string{}
    
    for containerID, metadata := range aps.containers {
        // Skip current active container
        if containerID == aps.currentContainer {
            continue
        }
        
        // Check if container has been superseded
        if aps.isContainerExpired(metadata, currentTime) {
            expiredContainers = append(expiredContainers, containerID)
        }
    }
    
    // Clean up expired containers
    for _, containerID := range expiredContainers {
        aps.removeExpiredContainer(containerID)
    }
}

func (aps *AvsPerformerServer) isContainerExpired(metadata *ContainerMetadata, currentTime int64) bool {
    // 1. Pending container that's way past its activation time
    if metadata.Status == ContainerStatusPending && 
       currentTime > metadata.ActivationTime + 3600 { // 1 hour grace period
        return true
    }
    
    // 2. Previously active containers that have been replaced
    if metadata.Status == ContainerStatusExpired {
        return true
    }
    
    return false
}
```

## Deployment Workflow

### End-to-End Process

1. **gRPC Request**: `DeployArtifact` with registry URL and digest
2. **Container Creation**: Create and start new container immediately
3. **Queue Addition**: Add to pending containers with future activation time
4. **Lazy Activation**: Container becomes active during next task processing after deadline
5. **Old Container Cleanup**: Previous active container marked as expired and reaped

```go
func (aps *AvsPerformerServer) DeployNewPerformerVersion(
    ctx context.Context, 
    registryURL string,
    digest string,
    activationTime int64,
) (string, error) {
    if activationTime <= time.Now().Unix() {
        return "", fmt.Errorf("activation time must be in the future")
    }
    
    // Create image config from RPC parameters
    image := containerManager.ImageConfig{
        Repository: registryURL,
        Tag:        digest,
    }
    
    // Create and start container
    containerInfo, endpoint, err := aps.createAndStartContainer(ctx, image)
    if err != nil {
        return "", err
    }
    
    // Create client
    client, err := aps.createPerformerClient(endpoint)
    if err != nil {
        aps.cleanupFailedContainer(ctx, containerInfo.ID)
        return "", err
    }
    
    deploymentID := generateDeploymentID()
    
    // Add to pending containers
    aps.addPendingContainer(&ContainerMetadata{
        Info:           containerInfo,
        Image:          image,
        ActivationTime: activationTime,
        Status:         ContainerStatusPending,
        Client:         client,
        Endpoint:       endpoint,
        DeploymentID:   deploymentID,
        CreatedAt:      time.Now(),
        RegistryURL:    registryURL,
        ArtifactDigest: digest,
    })
    
    return deploymentID, nil
}
```

## Safe Container Switching

### Atomic Activation Process

```go
func (aps *AvsPerformerServer) activateContainer(containerID string) error {
    aps.containerMu.Lock()
    defer aps.containerMu.Unlock()
    
    // 1. Verify new container is healthy before switching
    if !aps.isContainerHealthy(containerID) {
        return fmt.Errorf("cannot activate unhealthy container %s", containerID)
    }
    
    // 2. Perform application health check
    metadata := aps.containers[containerID]
    if !aps.isApplicationHealthy(metadata.Client) {
        return fmt.Errorf("container %s application is not responding", containerID)
    }
    
    // 3. Atomic switch
    oldContainer := aps.currentContainer
    aps.currentContainer = containerID
    aps.containers[containerID].Status = ContainerStatusActive
    now := time.Now()
    aps.containers[containerID].ActivatedAt = &now
    
    // 4. Mark old container for cleanup
    if oldContainer != "" {
        if oldMeta, exists := aps.containers[oldContainer]; exists {
            oldMeta.Status = ContainerStatusExpired
        }
    }
    
    aps.logger.Info("Activated new container",
        append(aps.logFields(),
            zap.String("newContainerID", containerID),
            zap.String("previousContainerID", oldContainer),
        )...,
    )
    
    return nil
}
```

## Health Monitoring

### Multi-Container Health Strategy

- **Active Container**: Full monitoring (container + application health)
- **Pending Containers**: Container-level health only (resource efficient)
- **Near-Activation Containers**: Start application health checks ~1 minute before activation

```go
type ContainerHealthStatus struct {
    ContainerID       string
    Status            ContainerStatus
    ActivationTime    int64
    ContainerHealth   ContainerHealthState
    ApplicationHealth ApplicationHealthState
    LastHealthCheck   time.Time
    Endpoint          string
    Image            string
}

func (aps *AvsPerformerServer) GetAllContainerHealth(ctx context.Context) ([]ContainerHealthStatus, error) {
    aps.containerMu.RLock()
    defer aps.containerMu.RUnlock()
    
    var statuses []ContainerHealthStatus
    for containerID, metadata := range aps.containers {
        status := ContainerHealthStatus{
            ContainerID:    containerID,
            Status:         metadata.Status,
            ActivationTime: metadata.ActivationTime,
            Endpoint:       metadata.Endpoint,
            Image:          fmt.Sprintf("%s:%s", metadata.Image.Repository, metadata.Image.Tag),
        }
        
        // Get container-level health
        status.ContainerHealth = aps.getContainerHealth(containerID)
        
        // Get application-level health (only if container is running)
        if status.ContainerHealth == ContainerHealthHealthy {
            status.ApplicationHealth = aps.getApplicationHealth(ctx, metadata.Client)
        }
        
        statuses = append(statuses, status)
    }
    return statuses, nil
}
```

## gRPC API Integration

### External Interface

The system integrates with the executor's gRPC API:

#### DeployArtifact RPC

```go
func (e *ExecutorService) DeployArtifact(ctx context.Context, req *executorpb.DeployArtifactRequest) (*executorpb.DeployArtifactResponse, error) {
    performer := e.getPerformerByAVS(req.AvsAddress)
    if performer == nil {
        return &executorpb.DeployArtifactResponse{
            Success: false,
            Message: "performer not found for AVS",
        }, nil
    }

    activationTime := time.Now().Unix() + 30 // 30 second delay for validation
    
    deploymentID, err := performer.DeployNewPerformerVersion(ctx, req.RegistryUrl, req.Digest, activationTime)
    if err != nil {
        return &executorpb.DeployArtifactResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    return &executorpb.DeployArtifactResponse{
        Success:      true,
        Message:      "deployment scheduled successfully",
        DeploymentId: deploymentID,
    }, nil
}
```

#### ListPerformers RPC

```go
func (e *ExecutorService) ListPerformers(ctx context.Context, req *executorpb.ListPerformersRequest) (*executorpb.ListPerformersResponse, error) {
    var performers []*executorpb.Performer
    
    for avsAddress, performer := range e.performers {
        if req.AvsAddress != "" && req.AvsAddress != avsAddress {
            continue
        }

        healthStatuses, err := performer.GetAllContainerHealth(ctx)
        if err != nil {
            continue
        }

        for _, status := range healthStatuses {
            performers = append(performers, &executorpb.Performer{
                PerformerId:        fmt.Sprintf("%s-%s", avsAddress, status.ContainerID[:8]),
                AvsAddress:         avsAddress,
                Status:            status.Status.String(),
                ArtifactRegistry:  status.Image,
                ArtifactDigest:    status.ContainerID,
                ResourceHealthy:   status.ContainerHealth == ContainerHealthHealthy,
                ApplicationHealthy: status.ApplicationHealth == AppHealthHealthy,
                LastHealthCheck:   status.LastHealthCheck.Format(time.RFC3339),
                ContainerId:       status.ContainerID,
            })
        }
    }

    return &executorpb.ListPerformersResponse{
        Performers: performers,
    }, nil
}
```

## DRY Improvements

### Helper Methods Created

The implementation leverages several helper methods to reduce code duplication:

- **`logFields()` / `logFieldsWithContainer()`**: Centralized logging fields
- **`cleanupFailedContainer()`**: Unified container cleanup logic
- **`createAndStartContainer()`**: Container lifecycle management
- **`createPerformerClient()`**: Standardized client creation
- **`createLivenessConfig()`**: Reusable configuration creation
- **`retryWithBackoff()`**: Generalized retry pattern

## Edge Cases and Error Handling

### Failure Scenarios

1. **Deployment Failures**: Clean up failed containers, don't affect active container
2. **Activation Failures**: Keep using current active container, log errors
3. **Container Crashes**: Only recreate active container using existing logic
4. **No Tasks for Extended Periods**: Containers eventually reaped via health check
5. **Multiple Pending Containers**: Activate the latest one past its deadline
6. **Clock Skew**: 1-hour grace period handles timing discrepancies

### Thread Safety

- **`sync.RWMutex`** protects container store operations
- **Atomic operations** for current container switching
- **Context cancellation** for health check coordination

## Benefits

### Simplicity & Efficiency

- **No background schedulers**: Lazy evaluation only when needed
- **Minimal goroutines**: Reuses existing health check infrastructure
- **Resource efficient**: Only monitors active containers fully
- **Deterministic**: Activation happens during task processing

### Robustness

- **Health validation** before container switching
- **Graceful degradation** if activation fails
- **Automatic cleanup** of expired containers
- **Thread-safe** operations with proper locking

### Observability

- **Comprehensive health status** for all containers
- **Deployment tracking** with unique IDs
- **Detailed logging** with consistent fields
- **External visibility** via gRPC APIs

## Implementation Phases

1. **Phase 1**: Container metadata store and basic queue management
2. **Phase 2**: Lazy activation logic and safe container switching
3. **Phase 3**: Health check integration and container reaping
4. **Phase 4**: gRPC API handlers and external interface
5. **Phase 5**: Testing and monitoring enhancements

This design provides scheduled container deployment with minimal complexity while maintaining robust health monitoring and safe container switching capabilities.