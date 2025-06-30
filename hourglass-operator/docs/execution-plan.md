# Hourglass Kubernetes Operator - Implementation Milestones

## 🔄 **ARCHITECTURE CHANGE: Singleton Operator + User-Managed Executors**

**New Architecture**: Single Hourglass Operator instance manages `Performer` CRDs cluster-wide. Users deploy multiple Executors as StatefulSets independently.

**Benefits**: 
- ✅ Clean separation of concerns (no circular dependencies)
- ✅ Singleton operator scales to handle multiple AVS executors
- ✅ Each executor deployed independently with flexible configuration
- ✅ Proper StatefulSet semantics for Executor persistence and scaling
- ✅ Greater flexibility for production deployments
- ✅ Simpler operator focused on core responsibility

---

## 🎯 Phase 1: Foundation & CRDs

### ✅ 1.1 Project Setup
- [x] Initialize Go module with Kubebuilder/Operator SDK
- [x] Set up project structure following operator conventions
- [x] Create base Dockerfile for operator deployment
- [ ] Configure CI/CD pipeline for building and testing operator

### 🔄 1.2 Custom Resource Definitions **[REFACTOR NEEDED]**
- [x] ~~**HourglassExecutor CRD**~~ - **WILL BE REMOVED**
- [x] **Performer CRD** - Complete type definition with:
  - AVS address, image, version
  - Advanced scheduling (node affinity, tolerations)
  - Hardware requirements (GPU, TEE)
  - gRPC configuration
- [x] **Generated CRD manifests** in `config/crd/bases/`

### 🔄 1.3 RBAC Configuration **[REFACTOR NEEDED]**
- [x] ~~Define controller RBAC permissions~~ - **NEEDS UPDATE**
- [x] ~~Create ClusterRole YAML manifests~~ - **NEEDS UPDATE**
- [x] ~~Create ServiceAccount and ClusterRoleBinding YAMLs~~ - **NEEDS UPDATE**
- [ ] Test RBAC permissions in cluster

## 🔄 Phase 2: Operator Refactoring **[ARCHITECTURE CHANGE]**

### ❌ 2.1 Remove ExecutorController **[NEW TASK]**
- [ ] **Remove HourglassExecutor CRD** - Delete type definitions
- [ ] **Remove ExecutorController** - Delete controller code
- [ ] **Update RBAC** - Remove deployment/configmap permissions
- [ ] **Update main.go** - Remove controller registration
- [ ] **Clean up generated files** - Remove deepcopy methods

### ✅ 2.2 PerformerController (Keep As-Is)
- [x] **Controller Structure** - Basic reconciler setup
- [x] **RBAC Annotations** - Permissions for pods and services
- [x] **Reconciliation Logic:**
  - [x] Create performer pod with specified image/version
  - [x] Create corresponding service with stable DNS name
  - [x] Update performer status with connection details
  - [x] Handle version upgrades via rolling deployment
  - [x] Clean up resources when performer is deleted

### ✅ 2.3 Service Management (Already Working)
- [x] **Service Naming Convention:**
  - Pattern: `performer-{performer-name}.{namespace}.svc.cluster.local`
  - Ensures stable DNS names for executor connections
- [x] **Service Creation Logic** - Target performer pods via labels
- [ ] **DNS Resolution Testing** - Verify executor can connect

## ❌ Phase 3: Executor Integration **[UPDATED FOR NEW ARCHITECTURE]**

### ❌ 3.1 Executor Modifications (in `../ponos/`)
- [ ] **Kubernetes Client Integration** - Add K8s client to executor
- [ ] **KubernetesAVSPerformer Type** - Implement K8s-native performer interface
- [ ] **KubernetesPerformerManager** - Manage performer lifecycle via CRDs
- [ ] **Configuration Mode Selection** - Choose docker vs kubernetes deployment
- [ ] **Backward Compatibility** - Preserve existing Docker support
- [ ] **StatefulSet Configuration** - Update for StatefulSet deployment pattern

### ❌ 3.2 gRPC Connection Management
- [ ] **Service DNS Connection** - Connect via stable service names
- [ ] **Connection Pooling** - Pool gRPC connections to performers
- [ ] **Retry Logic** - Handle connection failures gracefully
- [ ] **Load Balancing** - Support multiple performer replicas

## ❌ Phase 4: User Experience & Deployment **[UPDATED FOR NEW ARCHITECTURE]**

### ❌ 4.1 StatefulSet Templates & Helm Charts
- [ ] **Executor StatefulSet Templates:**
  - [ ] Base StatefulSet YAML with persistent storage
  - [ ] ConfigMap templates for chain configurations
  - [ ] Secret management for operator keys
  - [ ] Service and networking configuration
- [ ] **Singleton Operator Deployment:**
  - [ ] Single operator pod with cluster-wide RBAC permissions
  - [ ] CRD installation and upgrades
  - [ ] Operator configuration for multi-executor support
- [ ] **Complete Helm Chart:**
  - [ ] Executor StatefulSet deployment (user-deployable)
  - [ ] Singleton operator deployment (cluster-wide)
  - [ ] RBAC and networking setup
  - [ ] Configurable values for multi-executor scenarios

### ❌ 4.2 Documentation & Examples **[NEW FOCUS]**
- [ ] **User Guide** - Step-by-step deployment instructions
- [ ] **StatefulSet Examples** - Common executor configurations
- [ ] **Migration Guide** - From managed to user-managed approach
- [ ] **Troubleshooting Guide** - Common issues and solutions
- [ ] **Best Practices** - Production deployment patterns

### ❌ 4.3 Validation & Testing
- [ ] **Admission Webhooks** - Validate executor-operator compatibility
- [ ] **End-to-End Tests** - Complete workflow validation
- [ ] **Performance Benchmarks** - Performer creation/deletion latency
- [ ] **Chaos Testing** - Network partition and node failure scenarios

## ❌ Phase 5: Production Readiness **[UPDATED FOCUS]**

### ❌ 5.1 Monitoring & Observability
- [ ] **Prometheus Metrics** - Singleton operator and performer health metrics
- [ ] **Custom Metrics** - Multi-executor performer lifecycle events
- [ ] **Logging Integration** - Cluster logging stack integration
- [ ] **Health Checks** - Readiness and liveness probes
- [ ] **Distributed Tracing** - Request flow through multiple executors and performers

### ❌ 5.2 Security & Compliance
- [ ] **Pod Security Standards** - Compliance implementation
- [ ] **Network Policies** - Performer isolation rules
- [ ] **Secret Management** - Secure configuration handling
- [ ] **RBAC Hardening** - Principle of least privilege
- [ ] **Audit Logging** - Security event tracking

### ❌ 5.3 Advanced Features
- [ ] **Multi-Executor Coordination** - Singleton operator managing multiple executors
- [ ] **Auto-scaling** - HPA integration for performers
- [ ] **Blue-Green Deployments** - Safe performer upgrades
- [ ] **Cross-Cluster Support** - Performers in different clusters
- [ ] **Backup & Recovery** - Executor state management
- [ ] **Resource Quotas** - Per-executor resource limits

---

## 📋 Implementation Status Summary

**✅ Completed (Phase 1 & Phase 2 Partial)**
- Project structure and Performer CRD
- PerformerController implementation
- Basic RBAC permissions

**🔄 Refactoring Needed (Phase 2)**
- Remove HourglassExecutor CRD and controller
- Update RBAC for performer-only operator
- Clean up generated code

**❌ Remaining Work**
- Executor integration in `../ponos/` (Phase 3)
- StatefulSet templates and Helm charts (Phase 4)
- Production readiness features (Phase 5)

---

## 🏗️ **New Architecture Overview**

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            Kubernetes Cluster                                   │
│                                                                                 │
│  ┌─────────────────────┐              ┌─────────────────────────────────────┐  │
│  │ Hourglass Operator  │              │        Multiple User Executors      │  │
│  │   (Singleton)       │              │                                     │  │
│  │                     │              │ ┌─────────────┐  ┌─────────────────┐ │  │
│  │ ┌─────────────────┐ │              │ │ AVS-A       │  │ AVS-B           │ │  │
│  │ │ Performer       │ │◄─────────────┤ │ Executor    │  │ Executor        │ │  │
│  │ │ Controller      │ │ Creates CRDs │ │ StatefulSet │  │ StatefulSet     │ │  │
│  │ │                 │ │              │ │ (ns: avs-a) │  │ (ns: avs-b)     │ │  │
│  │ └─────────────────┘ │              │ └─────────────┘  └─────────────────┘ │  │
│  └─────────────────────┘              └─────────────────────────────────────┘  │
│           │                                           │                        │
│           │ Manages All Performer CRDs                │ Creates Performer CRDs │
│           ▼                                           ▼                        │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │                      Performer Pods & Services                          │  │
│  │                                                                         │  │
│  │ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐ │  │
│  │ │AVS-A Perf-1 │ │AVS-A Perf-2 │ │AVS-B Perf-1 │ │AVS-B Perf-2         │ │  │
│  │ │(ns: avs-a)  │ │(ns: avs-a)  │ │(ns: avs-b)  │ │(ns: avs-b)          │ │  │
│  │ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────────────┘ │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

**Key Changes:**
- 🚫 No more HourglassExecutor CRD  
- ✅ **Singleton operator** manages all Performer CRDs cluster-wide
- ✅ **Multiple executors** deployed independently by users
- ✅ **Namespace isolation** for different AVS deployments
- ✅ Clean separation: Multiple Executors → K8s API → Single Operator → Performers

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

## 🗓️ Updated Implementation Timeline

- **Week 1-2:** ✅ Phase 1 (Foundation & CRDs) - **COMPLETED** *(needs refactoring)*
- **Week 3-4:** 🔄 Phase 2 (Operator Refactoring) - **IN PROGRESS** *(remove executor controller)*
- **Week 5-6:** ❌ Phase 3 (Executor Integration) - **PENDING** *(work in `../ponos/`)*
- **Week 7-8:** ❌ Phase 4 (User Experience & Deployment) - **PENDING** *(StatefulSet templates)*
- **Week 9-10:** ❌ Phase 5 (Production Readiness) - **PENDING** *(monitoring, security)*

## 🔄 **Next Immediate Steps**

1. **Phase 2.1**: Remove HourglassExecutor CRD and controller
2. **Phase 2.1**: Update RBAC for performer-only permissions  
3. **Phase 2.1**: Clean up generated code and documentation
4. **Phase 3.1**: Begin executor modifications in `../ponos/`
5. **Phase 4.1**: Create StatefulSet templates and Helm charts

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