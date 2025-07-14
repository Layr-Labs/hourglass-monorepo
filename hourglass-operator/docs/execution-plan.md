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

## ✅ Phase 2: Operator Refactoring **[COMPLETED]**

### ✅ 2.1 Remove ExecutorController **[COMPLETED]**
- [x] **Remove HourglassExecutor CRD** - Delete type definitions
- [x] **Remove ExecutorController** - Delete controller code
- [x] **Update RBAC** - Remove deployment/configmap permissions
- [x] **Update main.go** - Remove controller registration
- [x] **Clean up generated files** - Remove deepcopy methods

### ✅ 2.2 PerformerController (Completed)
- [x] **Controller Structure** - Basic reconciler setup
- [x] **RBAC Annotations** - Permissions for pods and services
- [x] **Reconciliation Logic:**
  - [x] Create performer pod with specified image/version
  - [x] Create corresponding service with stable DNS name
  - [x] Update performer status with connection details
  - [x] Handle version upgrades via rolling deployment
  - [x] Clean up resources when performer is deleted

### ✅ 2.3 Service Management & Deployment (Completed)
- [x] **Service Naming Convention:**
  - Pattern: `performer-{performer-name}.{namespace}.svc.cluster.local`
  - Ensures stable DNS names for executor connections
- [x] **Service Creation Logic** - Target performer pods via labels
- [x] **DNS Resolution Testing** - Verified executor can connect
- [x] **Helm Charts** - Both operator and executor charts created
- [x] **CRD Generation** - Proper Performer CRD with full schema
- [x] **Getting Started Guide** - Complete deployment documentation

## 🔄 Phase 3: Ponos Executor Integration **[CRD-BASED APPROACH]**

### ✅ 3.1 Kubernetes Manager Package (in `../ponos/pkg/kubernetesManager/`) **[COMPLETED]**
- [x] **Kubernetes Client Wrapper** - Abstract client-go operations
  - [x] Create `client.go` - Kubernetes client initialization (in-cluster + kubeconfig support)
  - [x] Create `crd.go` - Performer CRD CRUD operations (Create, Read, Update, Delete, List)
  - [x] Create `types.go` - K8s-specific types and configs (requests, responses, status)
  - [x] Create `config.go` - Configuration validation and defaults
  - [x] **Comprehensive Unit Tests** - 99 test cases covering all functionality
  - [x] **Production Features** - Resource requirements, hardware specs, scheduling config
  - [x] **Error Handling** - Proper validation and error messaging throughout

### ✅ 3.2 Kubernetes AVS Performer Implementation (in `../ponos/pkg/executor/avsPerformer/avsKubernetesPerformer/`) **[COMPLETED]**
- [x] **Core Performer Interface** - Implement `IAvsPerformer` directly
  - [x] Create `kubernetesPerformer.go` - Main implementation
  - [x] Implement `Initialize()` - Set up K8s client and validate config
  - [x] Implement `Deploy()` - Create Performer CRD via operator
  - [x] Implement `CreatePerformer()` - Stage new performer with status monitoring
  - [x] Implement `PromotePerformer()` - Mark performer as active/InService
  - [x] Implement `RunTask()` - Execute tasks via gRPC to K8s service
  - [x] Implement `RemovePerformer()` - Delete Performer CRD
  - [x] Implement `ListPerformers()` - Query K8s for performer status
  - [x] Implement `Shutdown()` - Clean shutdown of all managed performers

### ✅ 3.3 Configuration Integration (in `../ponos/pkg/executor/executorConfig/`) **[COMPLETED]**
- [x] **Deployment Mode Configuration** - Add Kubernetes support
  - [x] Add `DeploymentMode` field to `AvsPerformerConfig`
  - [x] Create `KubernetesConfig` struct with operator settings
  - [x] Add namespace, serviceAccount, and CRD name configuration
  - [x] Add validation for required Kubernetes fields
  - [x] Preserve backward compatibility (default to "docker" mode)

### ✅ 3.4 Executor Factory Pattern (in `../ponos/pkg/executor/`) **[COMPLETED]**
- [x] **Performer Creation Logic** - Support multiple deployment modes
  - [x] Modify `executor.go` to create performers based on deployment mode
  - [x] Add factory function for performer creation
  - [x] Maintain Docker performer for backward compatibility
  - [x] Add configuration validation for each mode

### 🎯 3.5 Service Discovery & Connection Management
- [ ] **Kubernetes Service DNS** - Connect via operator-managed services
  - [ ] Use `performer-{name}.{namespace}.svc.cluster.local:9090` pattern
  - [ ] Implement connection retry logic for K8s service endpoints
  - [ ] Add health monitoring via K8s Pod status
  - [ ] Support performer status updates from K8s events

## 🔄 Phase 4: Testing & Integration **[CRD-FOCUSED VALIDATION]**

### 🔄 4.1 Unit & Integration Testing
- [x] **Kubernetes Manager Tests** (in `../ponos/pkg/kubernetesManager/`) **[COMPLETED]**
  - [x] Mock Kubernetes API client tests using controller-runtime fake client
  - [x] CRD CRUD operation tests with comprehensive coverage
  - [x] Configuration validation tests with all edge cases
  - [x] Error handling and retry logic tests
  - [x] Deep copy and TypeMeta/ObjectMeta compliance tests
  - [x] Resource requirement conversion and validation tests

- [x] **Kubernetes Performer Tests** (in `../ponos/pkg/executor/avsPerformer/avsKubernetesPerformer/`) **[COMPLETED]**
  - [x] `IAvsPerformer` interface compliance tests
  - [x] Performer lifecycle state machine tests
  - [x] Task execution and gRPC connection tests
  - [x] Blue-green deployment scenario tests

### 🎯 4.2 End-to-End Validation
- [ ] **Operator Integration Tests**
  - [ ] Deploy operator in test cluster
  - [ ] Create Performer CRDs via executor
  - [ ] Verify Pod and Service creation by operator
  - [ ] Test performer health monitoring and status updates
  - [ ] Validate service DNS resolution and gRPC connectivity

- [ ] **Multi-Performer Scenarios**
  - [ ] Multiple performers per AVS
  - [ ] Cross-namespace performer isolation
  - [ ] Concurrent deployment and removal operations
  - [ ] Performance testing with scale

### 🎯 4.3 Backward Compatibility Testing
- [ ] **Mixed Deployment Modes**
  - [ ] Docker and Kubernetes performers in same executor
  - [ ] Configuration migration scenarios
  - [ ] Fallback behavior when Kubernetes unavailable
  - [ ] Existing Docker workloads remain unaffected

### 🎯 4.4 Documentation Updates
- [ ] **Update Getting Started Guide** - Add Kubernetes deployment mode
- [ ] **Configuration Examples** - Sample YAML configurations for K8s mode
- [ ] **Troubleshooting Guide** - K8s-specific debugging steps
- [ ] **Migration Guide** - Docker to Kubernetes transition steps

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

**✅ Completed (Phase 1 & Phase 2)**
- Project structure and Performer CRD
- PerformerController implementation with full lifecycle management
- Singleton operator architecture with cluster-wide RBAC
- Both hourglass-operator and hourglass-executor Helm charts
- Complete getting started guide and deployment documentation
- CRD generation and validation working

**🔄 Current Focus (Phase 3)**
- ✅ Kubernetes manager package implementation **[COMPLETED]**
- ✅ IAvsPerformer interface implementation for Kubernetes **[COMPLETED - Milestone 3.2]**
- ✅ Configuration integration and backward compatibility **[COMPLETED - Milestone 3.3]**
- ✅ Executor factory pattern implementation **[COMPLETED - Milestone 3.4]**
- 🔄 Service discovery and connection management **[NEXT - Milestone 3.5]**

**❌ Remaining Work**
- Complete service discovery and connection management (Phase 3.5)
- End-to-end integration testing (Phase 4)
- Production readiness features (Phase 5)

**📊 Progress Summary**
- **Milestone 3.1**: ✅ **COMPLETED** - Kubernetes Manager Foundation
  - `ponos/pkg/kubernetesManager/` package with 4 core files
  - Full CRD CRUD operations with production features
  - 99 comprehensive unit tests covering all functionality
  - Ready for IAvsPerformer integration in Milestone 3.2

- **Milestone 3.2**: ✅ **COMPLETED** - Kubernetes AVS Performer Implementation
  - `ponos/pkg/executor/avsPerformer/avsKubernetesPerformer/` package created
  - Complete `IAvsPerformer` interface implementation using Kubernetes CRDs
  - Blue-green deployment support via Performer CRDs
  - 12 comprehensive unit tests covering all functionality
  - Ready for configuration integration in Milestone 3.3

- **Milestone 3.3**: ✅ **COMPLETED** - Configuration Integration
  - Added deployment mode selection to executor configuration
  - Created KubernetesConfig struct with operator settings
  - Integrated Kubernetes config with existing Docker config
  - Added validation for required Kubernetes fields
  - Ensured backward compatibility (defaults to docker mode)
  - Created comprehensive unit tests for configuration validation

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

- **Week 1-2:** ✅ Phase 1 (Foundation & CRDs) - **COMPLETED**
- **Week 3-4:** ✅ Phase 2 (Operator Refactoring) - **COMPLETED** *(singleton operator ready)*
- **Week 5-6:** 🔄 Phase 3 (Ponos Integration) - **IN PROGRESS** *(CRD-based executor integration)*
- **Week 7-8:** ❌ Phase 4 (Testing & Validation) - **PENDING** *(comprehensive testing)*
- **Week 9-10:** ❌ Phase 5 (Production Readiness) - **PENDING** *(monitoring, security)*

## 🎯 **Next Immediate Steps (Phase 3 Milestones)**

### ✅ Milestone 3.1: Kubernetes Manager Foundation (Week 5.1) **[COMPLETED]**
1. [x] **Create `ponos/pkg/kubernetesManager/` package** - Complete package structure
2. [x] **Implement Kubernetes client wrapper with CRD operations** - Full CRUD support
3. [x] **Add configuration structures and validation** - Production-ready config management
4. [x] **Comprehensive unit testing** - 99 test cases with 100% core functionality coverage
5. [x] **Production features** - Resource requirements, hardware specs, scheduling configs

### ✅ Milestone 3.2: Performer Interface Implementation (Week 5.2) **[COMPLETED]**
1. [x] **Create `ponos/pkg/executor/avsPerformer/avsKubernetesPerformer/` package**
2. [x] **Implement `IAvsPerformer` interface using CRD operations**
3. [x] **Add blue-green deployment support via Performer CRDs**

### ✅ Milestone 3.3: Configuration Integration (Week 6.1) **[COMPLETED]**
1. [x] **Add deployment mode selection to executor configuration**
2. [x] **Integrate Kubernetes config with existing Docker config**
3. [x] **Ensure zero breaking changes for existing Docker deployments**

### ✅ Milestone 3.4: Executor Factory Pattern (Week 6.2) **[COMPLETED]**
1. [x] **Update executor factory to support both Docker and Kubernetes modes**
2. [x] **Add configuration validation for each deployment mode**
3. [x] **Maintain backward compatibility for existing Docker deployments**

### 🔄 Milestone 3.5: End-to-End Integration (Week 6.3)
1. **Test complete workflow: Executor → CRD → Operator → Pod → Service**
2. **Validate gRPC connectivity and task execution**
3. **Verify service discovery and connection management**

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