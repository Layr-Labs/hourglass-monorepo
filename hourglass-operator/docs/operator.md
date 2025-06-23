# Hourglass Kubernetes Operator - Detailed Execution Plan

## Phase 1: Foundation & CRDs

### 1.1 Project Setup
- Initialize Go module with Kubebuilder/Operator SDK
- Set up project structure following operator conventions
- Configure CI/CD pipeline for building and testing operator
- Create base Dockerfile for operator deployment

### 1.2 Custom Resource Definitions
**HourglassExecutor CRD**
```yaml
spec:
  image: string                    # Executor container image
  replicas: int                   # Number of executor instances
  config:                         # Executor configuration
    aggregatorEndpoint: string
    operatorKeys: map[string]string
    chains: []ChainConfig
  resources:                      # Resource requirements
    limits: ResourceList
    requests: ResourceList
  nodeSelector: map[string]string # Node placement constraints
```

**Performer CRD**
```yaml
spec:
  avsAddress: string              # AVS identifier
  image: string                   # Performer container image
  version: string                 # Container version for upgrades
  executorRef: string             # Reference to parent executor
  config:                         # Performer-specific config
    grpcPort: int32
    environment: map[string]string
  resources:                      # Resource requirements
    limits: ResourceList
    requests: ResourceList
  scheduling:                     # Advanced scheduling requirements
    nodeSelector: map[string]string # Basic node selection
    nodeAffinity:                 # Advanced node affinity rules
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms: []NodeSelectorTerm
      preferredDuringSchedulingIgnoredDuringExecution:
        - weight: int32
          preference: NodeSelectorTerm
    tolerations: []Toleration     # Tolerate taints (e.g., for specialized nodes)
    runtimeClass: string          # Container runtime (e.g., gvisor, kata-containers)
  hardwareRequirements:           # Specialized hardware needs
    gpuType: string               # nvidia-tesla-v100, nvidia-a100, etc.
    gpuCount: int32               # Number of GPUs required
    teeRequired: bool             # Requires Trusted Execution Environment
    teeType: string               # sgx, sev, tdx, etc.
    customLabels: map[string]string # Custom hardware labels
status:
  phase: string                   # Pending/Running/Upgrading/Terminating
  podName: string                 # Associated pod name
  serviceName: string             # Associated service name
  grpcEndpoint: string            # DNS name for gRPC connection
```

### 1.3 RBAC Configuration
- Define ClusterRole with permissions for:
  - Managing Pods, Services, Deployments
  - Reading/Writing custom resources
  - Accessing ConfigMaps and Secrets
- Create ServiceAccount and ClusterRoleBinding

## Phase 2: Core Controllers

### 2.1 ExecutorController
**Responsibilities:**
- Deploy executor pods based on HourglassExecutor specs
- Manage executor configuration via ConfigMaps
- Handle executor scaling and updates
- Monitor executor health and restart policies

**Reconciliation Logic:**
1. Ensure executor deployment exists with correct spec
2. Create/update ConfigMap with executor configuration
3. Handle rolling updates when spec changes
4. Update status with deployment state

### 2.2 PerformerController
**Responsibilities:**
- Create long-running performer pods
- Manage performer services for gRPC access
- Handle performer upgrades (rolling updates)
- Clean up performers when terminated by executor

**Reconciliation Logic:**
1. Create performer pod with specified image/version
2. Create corresponding service with stable DNS name
3. Update performer status with connection details
4. Handle version upgrades via rolling deployment
5. Clean up resources when performer is deleted

### 2.3 Service Management
**Service Naming Convention:**
- Pattern: `performer-{avs-address}-{hash}.{namespace}.svc.cluster.local`
- Ensures stable DNS names for executor connections
- Services target performer pods via labels

## Phase 3: Executor Integration

### 3.1 Executor Modifications
**Add Kubernetes Integration (Preserving Docker Support):**
- Keep existing Docker-based performer management for backward compatibility
- Add new Kubernetes client for CRD operations
- Implement new AVSPerformer type specifically for Kubernetes operator pattern
- Allow configuration to choose between Docker and Kubernetes deployment modes

**New Kubernetes AVSPerformer Type:**
```go
type KubernetesAVSPerformer struct {
    client       client.Client
    namespace    string
    avsAddress   string
    performerCR  *v1alpha1.Performer
    grpcConn     *grpc.ClientConn
    grpcEndpoint string
}

// Implements existing AVSPerformer interface
func (k *KubernetesAVSPerformer) SubmitTask(ctx context.Context, req *pb.SubmitTaskRequest) (*pb.SubmitTaskResponse, error)
func (k *KubernetesAVSPerformer) GetTaskResult(ctx context.Context, req *pb.GetTaskResultRequest) (*pb.GetTaskResultResponse, error)
func (k *KubernetesAVSPerformer) Close() error
```

**New Kubernetes Performer Manager:**
```go
type KubernetesPerformerManager struct {
    client    client.Client
    namespace string
}

func (m *KubernetesPerformerManager) StartPerformer(ctx context.Context, req PerformerRequest) (AVSPerformer, error)
func (m *KubernetesPerformerManager) StopPerformer(ctx context.Context, avsAddress string) error
func (m *KubernetesPerformerManager) UpgradePerformer(ctx context.Context, avsAddress string, newVersion string) error
func (m *KubernetesPerformerManager) GetPerformerEndpoint(ctx context.Context, avsAddress string) (string, error)
```

**Configuration-Based Mode Selection:**
```go
type ExecutorConfig struct {
    // Existing fields...
    PerformerMode string `yaml:"performer_mode"` // "docker" or "kubernetes"
    Kubernetes    *KubernetesConfig `yaml:"kubernetes,omitempty"`
}

type KubernetesConfig struct {
    Namespace        string            `yaml:"namespace"`
    DefaultScheduling SchedulingConfig `yaml:"default_scheduling"`
}
```

### 3.2 gRPC Connection Management
- Executor connects to performers via service DNS names
- Implement connection pooling for performer gRPC clients
- Handle connection failures and retries
- Load balancing for multiple performer replicas (future)

## Phase 4: Deployment & Operations

### 4.1 Helm Charts
**Operator Deployment:**
- Operator pod with RBAC permissions
- CRD installation and upgrades
- ConfigMap for operator configuration

**Executor Deployment:**
- HourglassExecutor custom resource
- ConfigMap with chain configurations
- Secret management for operator keys

### 4.2 Monitoring & Observability
- Prometheus metrics for operator health
- Custom metrics for performer lifecycle events
- Logging integration with cluster logging stack
- Health checks and readiness probes

### 4.3 Security Considerations
- Pod Security Standards compliance
- Network policies for performer isolation
- Secret management for sensitive configuration
- Node affinity for secure node placement

## Phase 5: Advanced Features

### 5.1 High Availability
- Multiple executor replicas with leader election
- Performer affinity rules for distribution
- Graceful handling of node failures

### 5.2 Auto-scaling
- HPA integration for executor scaling
- Custom metrics for performer demand
- Resource-based scaling decisions

### 5.3 Upgrade Strategies
- Blue-green deployments for executors
- Rolling updates for performers
- Compatibility matrix management

## Scheduling Examples

**Bottlerocket Nodes:**
```yaml
scheduling:
  nodeSelector:
    node.kubernetes.io/os: bottlerocket
  tolerations:
  - key: "bottlerocket"
    operator: "Equal"
    value: "true"
    effect: "NoSchedule"
```

**GPU Workloads:**
```yaml
hardwareRequirements:
  gpuType: "nvidia-a100"
  gpuCount: 2
scheduling:
  nodeSelector:
    accelerator: nvidia-tesla-a100
  tolerations:
  - key: "nvidia.com/gpu"
    operator: "Exists"
    effect: "NoSchedule"
```

**TEE-Enabled Nodes:**
```yaml
hardwareRequirements:
  teeRequired: true
  teeType: "sgx"
scheduling:
  nodeSelector:
    intel.feature.node.kubernetes.io/sgx: "true"
  tolerations:
  - key: "sgx"
    operator: "Equal"
    value: "enabled"
    effect: "NoSchedule"
```

## Implementation Timeline

**Week 1-2:** Phase 1 (Foundation & CRDs)
**Week 3-4:** Phase 2 (Core Controllers)  
**Week 5-6:** Phase 3 (Executor Integration)
**Week 7-8:** Phase 4 (Deployment & Operations)
**Week 9-10:** Phase 5 (Advanced Features)

## Testing Strategy

### Unit Tests
- Controller reconciliation logic
- CRD validation and defaults
- Performer lifecycle state machines

### Integration Tests
- End-to-end operator deployment
- Executor � Performer communication
- Upgrade scenarios and rollbacks

### Performance Tests
- Performer creation/deletion latency
- Resource utilization under load
- Scale testing with multiple AVSs

## Risk Mitigation

**Kubernetes API Rate Limits:** Implement client-side rate limiting and caching
**Network Connectivity:** Design robust retry mechanisms for gRPC connections  
**Resource Contention:** Implement resource quotas and limits
**Upgrade Compatibility:** Maintain backward compatibility for existing deployments