# Hourglass Kubernetes Operator - Implementation Milestones

## 🎯 Phase 1: Foundation & CRDs

### ✅ 1.1 Project Setup
- [x] Initialize Go module with Kubebuilder/Operator SDK
- [x] Set up project structure following operator conventions
- [x] Create base Dockerfile for operator deployment
- [ ] Configure CI/CD pipeline for building and testing operator

### ✅ 1.2 Custom Resource Definitions
- [x] **HourglassExecutor CRD** - Complete type definition with:
  - Image, replicas, configuration
  - Chain configs, operator keys
  - Resource requirements, scheduling
  - Kubernetes-specific settings
- [x] **Performer CRD** - Complete type definition with:
  - AVS address, image, version
  - Advanced scheduling (node affinity, tolerations)
  - Hardware requirements (GPU, TEE)
  - gRPC configuration
- [x] **Generated CRD manifests** in `config/crd/bases/`

### ✅ 1.3 RBAC Configuration
- [x] Define controller RBAC permissions in code annotations
- [x] Create ClusterRole YAML manifests
- [x] Create ServiceAccount and ClusterRoleBinding YAMLs
- [ ] Test RBAC permissions in cluster

## ✅ Phase 2: Core Controllers

### ✅ 2.1 ExecutorController
- [x] **Controller Structure** - Basic reconciler setup
- [x] **RBAC Annotations** - Permissions for deployments, configmaps, services
- [x] **Reconciliation Logic:**
  - [x] Ensure executor deployment exists with correct spec
  - [x] Create/update ConfigMap with executor configuration
  - [x] Handle rolling updates when spec changes
  - [x] Update status with deployment state
- [x] **Health Monitoring** - Restart policies and health checks

### ✅ 2.2 PerformerController  
- [x] **Controller Structure** - Basic reconciler setup
- [x] **RBAC Annotations** - Permissions for pods and services
- [x] **Reconciliation Logic:**
  - [x] Create performer pod with specified image/version
  - [x] Create corresponding service with stable DNS name
  - [x] Update performer status with connection details
  - [x] Handle version upgrades via rolling deployment
  - [x] Clean up resources when performer is deleted

### ❌ 2.3 Service Management
- [ ] **Service Naming Convention:**
  - Pattern: `performer-{avs-address}-{hash}.{namespace}.svc.cluster.local`
  - Ensures stable DNS names for executor connections
- [ ] **Service Creation Logic** - Target performer pods via labels
- [ ] **DNS Resolution Testing** - Verify executor can connect

## ❌ Phase 3: Executor Integration

### ❌ 3.1 Executor Modifications  
- [ ] **Kubernetes Client Integration** - Add K8s client to executor
- [ ] **KubernetesAVSPerformer Type** - Implement K8s-native performer interface
- [ ] **KubernetesPerformerManager** - Manage performer lifecycle via CRDs
- [ ] **Configuration Mode Selection** - Choose docker vs kubernetes deployment
- [ ] **Backward Compatibility** - Preserve existing Docker support

### ❌ 3.2 gRPC Connection Management
- [ ] **Service DNS Connection** - Connect via stable service names
- [ ] **Connection Pooling** - Pool gRPC connections to performers
- [ ] **Retry Logic** - Handle connection failures gracefully
- [ ] **Load Balancing** - Support multiple performer replicas

## ❌ Phase 4: Deployment & Operations

### ❌ 4.1 Helm Charts
- [ ] **Operator Deployment:**
  - [ ] Operator pod with RBAC permissions
  - [ ] CRD installation and upgrades  
  - [ ] ConfigMap for operator configuration
- [ ] **Executor Deployment:**
  - [ ] HourglassExecutor custom resource templates
  - [ ] ConfigMap with chain configurations
  - [ ] Secret management for operator keys

### ❌ 4.2 Monitoring & Observability
- [ ] **Prometheus Metrics** - Operator and performer health metrics
- [ ] **Custom Metrics** - Performer lifecycle events
- [ ] **Logging Integration** - Cluster logging stack integration
- [ ] **Health Checks** - Readiness and liveness probes

### ❌ 4.3 Security Considerations
- [ ] **Pod Security Standards** - Compliance implementation
- [ ] **Network Policies** - Performer isolation rules
- [ ] **Secret Management** - Secure configuration handling
- [ ] **Node Affinity** - Secure node placement rules

## ❌ Phase 5: Advanced Features

### ❌ 5.1 High Availability
- [ ] **Leader Election** - Multiple executor replicas with leader election
- [ ] **Affinity Rules** - Performer distribution across nodes
- [ ] **Failure Handling** - Graceful node failure recovery

### ❌ 5.2 Auto-scaling
- [ ] **HPA Integration** - Horizontal Pod Autoscaler for executors
- [ ] **Custom Metrics** - Performer demand-based scaling
- [ ] **Resource Scaling** - Resource-based scaling decisions

### ❌ 5.3 Upgrade Strategies
- [ ] **Blue-Green Deployments** - Safe executor upgrades
- [ ] **Rolling Updates** - Performer version upgrades
- [ ] **Compatibility Matrix** - Version compatibility management

---

## 📋 Implementation Status Summary

**✅ Completed (Phase 1 & 2)**
- Project structure and CRD types
- Basic controller framework
- RBAC permission definitions
- Full ExecutorController implementation
- Full PerformerController implementation

**❌ Remaining Work**
- Service management (2.3)
- Executor integration (Phase 3)
- Deployment tooling (Phase 4)
- Advanced features (Phase 5)

---

## 🎯 Scheduling Examples

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

## 🗓️ Implementation Timeline

- **Week 1-2:** ✅ Phase 1 (Foundation & CRDs) - **COMPLETED**
- **Week 3-4:** ✅ Phase 2 (Core Controllers) - **COMPLETED**  
- **Week 5-6:** ❌ Phase 3 (Executor Integration) - **PENDING**
- **Week 7-8:** ❌ Phase 4 (Deployment & Operations) - **PENDING** 
- **Week 9-10:** ❌ Phase 5 (Advanced Features) - **PENDING**

## 🧪 Testing Strategy

### ❌ Unit Tests
- [ ] Controller reconciliation logic
- [ ] CRD validation and defaults
- [ ] Performer lifecycle state machines

### ❌ Integration Tests
- [ ] End-to-end operator deployment
- [ ] Executor → Performer communication
- [ ] Upgrade scenarios and rollbacks

### ❌ Performance Tests
- [ ] Performer creation/deletion latency
- [ ] Resource utilization under load
- [ ] Scale testing with multiple AVSs

## ⚠️ Risk Mitigation

- [ ] **Kubernetes API Rate Limits:** Implement client-side rate limiting and caching
- [ ] **Network Connectivity:** Design robust retry mechanisms for gRPC connections  
- [ ] **Resource Contention:** Implement resource quotas and limits
- [ ] **Upgrade Compatibility:** Maintain backward compatibility for existing deployments