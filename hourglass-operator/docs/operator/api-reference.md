# API Reference

This document provides detailed API reference for the Hourglass Kubernetes Operator custom resources.

## API Groups

- **Group**: `hourglass.eigenlayer.io`
- **Version**: `v1alpha1`

## HourglassExecutor

The HourglassExecutor resource manages the aggregation and execution layer of the Hourglass system.

### Resource Definition

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: HourglassExecutor
metadata:
  name: string
  namespace: string
spec:
  # HourglassExecutorSpec
status:
  # HourglassExecutorStatus
```

### HourglassExecutorSpec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `image` | string | Yes | Container image for the executor |
| `replicas` | *int32 | No | Number of executor instances (default: 1) |
| `config` | HourglassExecutorConfig | Yes | Executor configuration |
| `resources` | corev1.ResourceRequirements | No | Resource requirements |
| `nodeSelector` | map[string]string | No | Node selection constraints |
| `tolerations` | []corev1.Toleration | No | Pod tolerations |
| `imagePullSecrets` | []corev1.LocalObjectReference | No | Image pull secrets |

### HourglassExecutorConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `aggregatorEndpoint` | string | Yes | gRPC endpoint of the aggregator service |
| `operatorKeys` | map[string]string | Yes | BLS and ECDSA private keys |
| `chains` | []ChainConfig | Yes | Blockchain network configurations |
| `performerMode` | string | No | Deployment mode: "docker" or "kubernetes" (default: "kubernetes") |
| `kubernetes` | *KubernetesConfig | No | Kubernetes-specific configuration |
| `logLevel` | string | No | Logging level (default: "info") |

### ChainConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Chain identifier (e.g., "ethereum", "base") |
| `rpc` | string | Yes | RPC endpoint URL |
| `chainId` | int64 | Yes | Blockchain network ID |
| `taskMailboxAddress` | string | Yes | TaskMailbox contract address |

### KubernetesConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `namespace` | string | No | Target namespace for performers (default: "default") |
| `defaultScheduling` | *SchedulingConfig | No | Default scheduling constraints for performers |

### HourglassExecutorStatus

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | Current deployment phase: "Pending", "Running", "Stopped" |
| `replicas` | int32 | Total number of replicas |
| `readyReplicas` | int32 | Number of ready replicas |
| `conditions` | []metav1.Condition | Detailed status conditions |
| `lastConfigUpdate` | *metav1.Time | Timestamp of last configuration update |

## Performer

The Performer resource manages individual AVS workload containers with advanced scheduling capabilities.

### Resource Definition

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: string
  namespace: string
spec:
  # PerformerSpec
status:
  # PerformerStatus
```

### PerformerSpec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `avsAddress` | string | Yes | Unique identifier for the AVS |
| `image` | string | Yes | Container image for the performer |
| `version` | string | No | Image version for upgrade tracking |
| `executorRef` | string | No | Reference to parent HourglassExecutor |
| `config` | PerformerConfig | No | Performer-specific configuration |
| `resources` | corev1.ResourceRequirements | No | Resource requirements |
| `scheduling` | *SchedulingConfig | No | Advanced scheduling constraints |
| `hardwareRequirements` | *HardwareRequirements | No | Specialized hardware needs |
| `imagePullSecrets` | []corev1.LocalObjectReference | No | Image pull secrets |

### PerformerConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `grpcPort` | int32 | No | gRPC server port (default: 9090, range: 1-65535) |
| `environment` | map[string]string | No | Environment variables |
| `args` | []string | No | Additional command line arguments |
| `command` | []string | No | Override container entrypoint |

### SchedulingConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `nodeSelector` | map[string]string | No | Basic node selection labels |
| `nodeAffinity` | *corev1.NodeAffinity | No | Advanced node affinity rules |
| `tolerations` | []corev1.Toleration | No | Pod tolerations for tainted nodes |
| `runtimeClass` | *string | No | Container runtime class |

### HardwareRequirements

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `gpuType` | string | No | GPU type (e.g., "nvidia-tesla-v100", "nvidia-a100") |
| `gpuCount` | int32 | No | Number of GPUs required (minimum: 0) |
| `teeRequired` | bool | No | Whether TEE is required |
| `teeType` | string | No | TEE technology type (e.g., "sgx", "sev", "tdx") |
| `customLabels` | map[string]string | No | Custom hardware matching labels |

### PerformerStatus

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | Current phase: "Pending", "Running", "Upgrading", "Terminating", "Failed" |
| `podName` | string | Name of the associated pod |
| `serviceName` | string | Name of the associated service |
| `grpcEndpoint` | string | Full DNS name for gRPC connections |
| `conditions` | []metav1.Condition | Detailed status conditions |
| `lastUpgrade` | *metav1.Time | Timestamp of last upgrade |
| `readyTime` | *metav1.Time | When the performer became ready |

## Common Fields

### ResourceRequirements

Standard Kubernetes resource specification:

```yaml
resources:
  requests:
    cpu: "500m"
    memory: "1Gi"
    nvidia.com/gpu: "1"
  limits:
    cpu: "2"
    memory: "4Gi"
    nvidia.com/gpu: "1"
```

### NodeAffinity

Standard Kubernetes node affinity:

```yaml
nodeAffinity:
  requiredDuringSchedulingIgnoredDuringExecution:
    nodeSelectorTerms:
    - matchExpressions:
      - key: "kubernetes.io/arch"
        operator: In
        values: ["amd64"]
  preferredDuringSchedulingIgnoredDuringExecution:
  - weight: 100
    preference:
      matchExpressions:
      - key: "node.kubernetes.io/instance-type"
        operator: In
        values: ["m5.large", "m5.xlarge"]
```

### Tolerations

Standard Kubernetes tolerations:

```yaml
tolerations:
- key: "dedicated"
  operator: "Equal"
  value: "hourglass"
  effect: "NoSchedule"
- key: "nvidia.com/gpu"
  operator: "Exists"
  effect: "NoSchedule"
```

## Validation Rules

### HourglassExecutor Validation

- `image` must be a valid container image reference
- `replicas` must be >= 1 if specified
- `config.aggregatorEndpoint` must be a valid endpoint
- `config.chains` must contain at least one chain
- Each chain must have a unique `name`
- `config.performerMode` must be "docker" or "kubernetes"

### Performer Validation

- `avsAddress` must be a valid Ethereum address format
- `image` must be a valid container image reference
- `config.grpcPort` must be in range 1-65535
- `hardwareRequirements.gpuCount` must be >= 0
- If `hardwareRequirements.teeRequired` is true, `teeType` should be specified

## Status Conditions

Both resources support standard Kubernetes condition types:

### Common Condition Types

| Type | Status | Reason | Description |
|------|--------|--------|-------------|
| `Available` | True/False | Various | Resource is available and functioning |
| `Progressing` | True/False | Various | Resource is being updated |
| `Degraded` | True/False | Various | Resource is degraded but functional |

### HourglassExecutor Conditions

| Reason | Description |
|--------|-------------|
| `DeploymentAvailable` | Underlying deployment is available |
| `ConfigMapReady` | Configuration has been applied |
| `RollingUpdate` | Rolling update in progress |

### Performer Conditions

| Reason | Description |
|--------|-------------|
| `PodReady` | Performer pod is ready |
| `ServiceReady` | Service is created and ready |
| `Scheduled` | Pod has been scheduled to a node |
| `ImagePullFailed` | Failed to pull container image |
| `SchedulingFailed` | Failed to find suitable node |

## Examples

### Minimal HourglassExecutor

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: HourglassExecutor
metadata:
  name: minimal-executor
spec:
  image: "hourglass/executor:latest"
  config:
    aggregatorEndpoint: "aggregator.example.com:9090"
    operatorKeys:
      ecdsa: "0x..."
      bls: "0x..."
    chains:
    - name: "ethereum"
      rpc: "https://mainnet.infura.io/v3/YOUR_KEY"
      chainId: 1
      taskMailboxAddress: "0x..."
```

### Minimal Performer

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: minimal-performer
spec:
  avsAddress: "0x1234567890abcdef1234567890abcdef12345678"
  image: "myavs/performer:latest"
```

### Complex Performer with All Features

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: complex-performer
spec:
  avsAddress: "0x1234567890abcdef1234567890abcdef12345678"
  image: "myavs/performer:v2.0.0"
  version: "v2.0.0"
  executorRef: "main-executor"
  config:
    grpcPort: 9090
    environment:
      LOG_LEVEL: "debug"
      WORKER_THREADS: "4"
    command: ["/usr/bin/performer"]
    args: ["--config", "/etc/config.yaml"]
  resources:
    requests:
      cpu: "2"
      memory: "4Gi"
    limits:
      cpu: "8"
      memory: "16Gi"
  scheduling:
    nodeSelector:
      node.kubernetes.io/instance-type: "m5.2xlarge"
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: "kubernetes.io/arch"
            operator: In
            values: ["amd64"]
    tolerations:
    - key: "dedicated"
      operator: "Equal"
      value: "avs-workloads"
      effect: "NoSchedule"
    runtimeClass: "gvisor"
  hardwareRequirements:
    gpuType: "nvidia-a100"
    gpuCount: 2
    teeRequired: true
    teeType: "sgx"
    customLabels:
      intel.sgx.version: "2.0"
  imagePullSecrets:
  - name: private-registry
```

## kubectl Commands

### List Resources

```bash
# List all executors
kubectl get hourglassexecutors

# List all performers
kubectl get performers

# List both with custom columns
kubectl get hourglassexecutors,performers -o custom-columns=\
NAME:.metadata.name,\
PHASE:.status.phase,\
READY:.status.readyReplicas,\
AGE:.metadata.creationTimestamp
```

### Describe Resources

```bash
# Describe executor
kubectl describe hourglassexecutor my-executor

# Describe performer
kubectl describe performer my-performer
```

### Edit Resources

```bash
# Edit executor
kubectl edit hourglassexecutor my-executor

# Patch performer image
kubectl patch performer my-performer --type='merge' \
  -p='{"spec":{"image":"myavs/performer:v2.1.0","version":"v2.1.0"}}'
```

### Status and Logs

```bash
# Check executor status
kubectl get hourglassexecutor my-executor -o yaml

# Check performer logs
kubectl logs performer-my-performer

# Follow logs
kubectl logs -f deployment/my-executor
```