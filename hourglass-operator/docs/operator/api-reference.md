# API Reference

This document provides detailed API reference for the Hourglass Kubernetes Operator custom resources.

**Architecture**: The operator implements a **singleton pattern** managing only Performer resources. Executors are deployed independently by users as StatefulSets.

## API Groups

- **Group**: `hourglass.eigenlayer.io`
- **Version**: `v1alpha1`

## Performer

The Performer resource manages individual AVS workload containers with advanced scheduling and hardware requirements. This is the **only** custom resource managed by the singleton operator.

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
| `config` | PerformerConfig | No | Performer-specific configuration |
| `resources` | corev1.ResourceRequirements | No | Resource requirements |
| `scheduling` | *SchedulingConfig | No | Advanced scheduling constraints |
| `hardwareRequirements` | *HardwareRequirements | No | Specialized hardware needs |
| `imagePullSecrets` | []corev1.LocalObjectReference | No | Image pull secrets |

**Note**: The `executorRef` field has been **removed** in the singleton architecture. Performers are created directly by user-managed Executors.

### PerformerConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `grpcPort` | int32 | No | gRPC server port (default: 9090, range: 1-65535) |
| `env` | []corev1.EnvVar | No | Environment variables using standard k8s EnvVar type |
| `args` | []string | No | Additional command line arguments |
| `command` | []string | No | Override container entrypoint |
| `serviceAccountName` | string | No | Service account for the performer pod |

### SchedulingConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `nodeSelector` | map[string]string | No | Basic node selection labels |
| `tolerations` | []corev1.Toleration | No | Pod tolerations for tainted nodes |
| `affinity` | *corev1.Affinity | No | Full Kubernetes affinity specification |
| `runtimeClass` | *string | No | Container runtime class (e.g., "gvisor", "kata") |
| `priorityClassName` | *string | No | Priority class for pod scheduling |

**Important**: The field is now `affinity` (not `nodeAffinity`) to support the full Kubernetes affinity specification including node, pod, and anti-affinity rules.

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

## Service Discovery

The operator creates stable DNS names for performers following the pattern:

```
performer-{performer-name}.{namespace}.svc.cluster.local:{grpcPort}
```

**Examples:**
- `performer-example.avs-project-a.svc.cluster.local:9090`
- `performer-ml-worker.production.svc.cluster.local:8080`

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

### Affinity

Full Kubernetes affinity specification (node, pod, and anti-affinity):

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: "kubernetes.io/arch"
          operator: In
          values: ["amd64"]
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: "app"
            operator: In
            values: ["performer"]
        topologyKey: "kubernetes.io/hostname"
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

### Performer Validation

- `avsAddress` must be a valid Ethereum address format (0x followed by 40 hex characters)
- `image` must be a valid container image reference
- `config.grpcPort` must be in range 1-65535
- `hardwareRequirements.gpuCount` must be >= 0
- If `hardwareRequirements.teeRequired` is true, `teeType` should be specified

## Status Conditions

Performers support standard Kubernetes condition types:

### Common Condition Types

| Type | Status | Reason | Description |
|------|--------|--------|-------------|
| `Available` | True/False | Various | Resource is available and functioning |
| `Progressing` | True/False | Various | Resource is being updated |
| `Degraded` | True/False | Various | Resource is degraded but functional |

### Performer Condition Reasons

| Reason | Description |
|--------|-------------|
| `PodReady` | Performer pod is ready and running |
| `ServiceReady` | Service is created and endpoints are ready |
| `Scheduled` | Pod has been successfully scheduled to a node |
| `ImagePullFailed` | Failed to pull container image |
| `SchedulingFailed` | Failed to find suitable node for constraints |
| `HardwareUnavailable` | Required hardware (GPU/TEE) not available |
| `UpgradeInProgress` | Performer is being upgraded to new version |

## Examples

### Minimal Performer

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: minimal-performer
  namespace: my-avs-project
spec:
  avsAddress: "0x1234567890abcdef1234567890abcdef12345678"
  image: "myavs/performer:latest"
```

### Performer with Environment Variables from Secrets/ConfigMaps

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: env-performer
  namespace: my-avs-project
spec:
  avsAddress: "0x1234567890abcdef1234567890abcdef12345678"
  image: "myavs/performer:latest"
  config:
    grpcPort: 9090
    env:
    # Direct value
    - name: LOG_LEVEL
      value: "info"
    # From secret
    - name: API_KEY
      valueFrom:
        secretKeyRef:
          name: api-secrets
          key: api-key
    # From configmap
    - name: CONFIG_DATA
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: config.json
    # From field reference
    - name: POD_NAME
      valueFrom:
        fieldRef:
          fieldPath: metadata.name
    # From resource field
    - name: MEM_LIMIT
      valueFrom:
        resourceFieldRef:
          containerName: performer
          resource: limits.memory
```

### GPU-Enabled Performer

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: gpu-performer
  namespace: ml-workloads
spec:
  avsAddress: "0xabcdef1234567890abcdef1234567890abcdef12"
  image: "myavs/ml-performer:v3.0.0"
  version: "v3.0.0"
  config:
    grpcPort: 9090
    env:
    - name: CUDA_VISIBLE_DEVICES
      value: "0,1"
    - name: LOG_LEVEL
      value: "info"
  resources:
    requests:
      nvidia.com/gpu: "2"
      cpu: "4"
      memory: "8Gi"
    limits:
      nvidia.com/gpu: "2"
      cpu: "8"
      memory: "16Gi"
  scheduling:
    nodeSelector:
      accelerator: "nvidia-tesla-a100"
    tolerations:
    - key: "nvidia.com/gpu"
      operator: "Exists"
      effect: "NoSchedule"
  hardwareRequirements:
    gpuType: "nvidia-a100"
    gpuCount: 2
```

### TEE-Enabled Performer

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: tee-performer
  namespace: secure-compute
spec:
  avsAddress: "0x9876543210987654321098765432109876543210"
  image: "myavs/secure-performer:v1.5.0"
  version: "v1.5.0"
  config:
    grpcPort: 8080
    env:
    - name: SGX_MODE
      value: "HW"
    - name: SECURITY_LEVEL
      value: "high"
  scheduling:
    nodeSelector:
      intel.feature.node.kubernetes.io/sgx: "true"
    tolerations:
    - key: "sgx"
      operator: "Equal"
      value: "enabled"
      effect: "NoSchedule"
    runtimeClass: "kata"
  hardwareRequirements:
    teeRequired: true
    teeType: "sgx"
    customLabels:
      sgx.intel.com/epc: "128Mi"
```

### Complex Performer with All Features

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: complex-performer
  namespace: production
spec:
  avsAddress: "0x1111222233334444555566667777888899990000"
  image: "myavs/performer:v2.1.0"
  version: "v2.1.0"
  config:
    grpcPort: 9090
    env:
    - name: LOG_LEVEL
      value: "debug"
    - name: WORKER_THREADS
      value: "8"
    - name: CACHE_SIZE
      value: "1GB"
    command: ["/usr/bin/performer"]
    args: 
      - "--config=/etc/performer/config.yaml"
      - "--enable-metrics"
      - "--metrics-port=8080"
  resources:
    requests:
      cpu: "4"
      memory: "8Gi"
    limits:
      cpu: "16"
      memory: "32Gi"
  scheduling:
    nodeSelector:
      node.kubernetes.io/instance-type: "m5.4xlarge"
      workload-type: "compute-intensive"
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
          - matchExpressions:
            - key: "kubernetes.io/arch"
              operator: In
              values: ["amd64"]
            - key: "node.kubernetes.io/instance-type"
              operator: In
              values: ["m5.4xlarge", "m5.8xlarge", "c5.4xlarge"]
        preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          preference:
            matchExpressions:
            - key: "zone"
              operator: In
              values: ["us-west-2a", "us-west-2b"]
      podAntiAffinity:
        preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          podAffinityTerm:
            labelSelector:
              matchExpressions:
              - key: "app"
                operator: In
                values: ["performer"]
            topologyKey: "kubernetes.io/hostname"
    tolerations:
    - key: "dedicated"
      operator: "Equal"
      value: "avs-workloads"
      effect: "NoSchedule"
    - key: "high-memory"
      operator: "Exists"
      effect: "NoSchedule"
    runtimeClass: "gvisor"
    priorityClassName: "high-priority"
  hardwareRequirements:
    customLabels:
      performance-tier: "premium"
      network-bandwidth: "10gbps"
  imagePullSecrets:
  - name: private-registry-secret
```

## User-Managed Executor Integration

Since Executors are now deployed by users as StatefulSets, they create Performer CRDs using the Kubernetes API. Here's how Executors typically interact with Performers:

### Executor Environment Variables

```yaml
# In Executor StatefulSet
env:
- name: DEPLOYMENT_MODE
  value: "kubernetes"
- name: KUBERNETES_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: PERFORMER_SERVICE_PATTERN
  value: "performer-{name}.{namespace}.svc.cluster.local:{port}"
```

### Executor RBAC

Executors need permissions to manage Performers in their namespace:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: my-avs-project
  name: executor-role
rules:
- apiGroups: ["hourglass.eigenlayer.io"]
  resources: ["performers"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get", "list", "watch"]
```

## kubectl Commands

### List Resources

```bash
# List all performers across cluster (operator manages cluster-wide)
kubectl get performers --all-namespaces

# List performers in specific namespace
kubectl get performers -n my-avs-project

# List with custom columns
kubectl get performers -o custom-columns=\
NAME:.metadata.name,\
NAMESPACE:.metadata.namespace,\
AVS:.spec.avsAddress,\
PHASE:.status.phase,\
ENDPOINT:.status.grpcEndpoint,\
AGE:.metadata.creationTimestamp
```

### Describe Resources

```bash
# Describe performer
kubectl describe performer my-performer -n my-avs-project

# Check performer status across all namespaces
kubectl get performers --all-namespaces -o wide
```

### Edit Resources

```bash
# Edit performer
kubectl edit performer my-performer -n my-avs-project

# Patch performer image
kubectl patch performer my-performer -n my-avs-project --type='merge' \
  -p='{"spec":{"image":"myavs/performer:v2.1.0","version":"v2.1.0"}}'

# Scale resources
kubectl patch performer my-performer -n my-avs-project --type='merge' \
  -p='{"spec":{"resources":{"requests":{"cpu":"2","memory":"4Gi"}}}}'
```

### Status and Logs

```bash
# Check performer status
kubectl get performer my-performer -n my-avs-project -o yaml

# Check performer logs (via pod)
kubectl logs -l "app=performer,performer-name=my-performer" -n my-avs-project

# Follow logs
kubectl logs -f $(kubectl get pod -l "app=performer,performer-name=my-performer" -n my-avs-project -o jsonpath='{.items[0].metadata.name}') -n my-avs-project

# Check service endpoints
kubectl get endpoints -l "app=performer" -n my-avs-project

# Test DNS resolution from executor
kubectl exec -it <executor-pod> -n my-avs-project -- \
  nslookup performer-my-performer.my-avs-project.svc.cluster.local
```

### Multi-Namespace Operations

```bash
# Monitor all performers across multiple AVS projects
kubectl get performers --all-namespaces --watch

# Count performers per namespace
kubectl get performers --all-namespaces --no-headers | awk '{print $1}' | sort | uniq -c

# Find performers by AVS address
kubectl get performers --all-namespaces -o json | \
  jq -r '.items[] | select(.spec.avsAddress=="0x1234...") | "\(.metadata.namespace)/\(.metadata.name)"'
```

## Migration Notes

### From Previous Architecture

If migrating from a version with HourglassExecutor CRDs:

1. **Remove executorRef**: The `executorRef` field is no longer supported
2. **Update RBAC**: Executors need permissions to manage Performers in their namespace
3. **Service Discovery**: DNS names remain the same (`performer-{name}.{namespace}.svc.cluster.local`)
4. **Status Monitoring**: Use namespace-scoped queries for executor-specific monitoring

### API Changes

| Field | Old Architecture | New Architecture | Notes |
|-------|------------------|------------------|-------|
| `executorRef` | Required reference | **Removed** | No longer needed |
| `scheduling.nodeAffinity` | Separate field | Part of `affinity` | Use full Kubernetes affinity spec |
| RBAC scope | Operator manages all | User executors manage per-namespace | Better isolation |