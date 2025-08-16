# Usage Examples

This document provides practical examples for deploying and configuring the Hourglass Kubernetes Operator in the **singleton architecture**.

**Architecture Note**: The operator now uses a singleton pattern where users deploy Executors independently as StatefulSets, and a single operator instance manages all Performer resources cluster-wide.

## Basic Examples

### User-Deployed Executor StatefulSet

**Note**: Executors are now deployed by users as StatefulSets, not as HourglassExecutor CRDs.

```yaml
# executor-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: basic-executor
  namespace: my-avs-project
spec:
  serviceName: basic-executor
  replicas: 1
  selector:
    matchLabels:
      app: basic-executor
  template:
    metadata:
      labels:
        app: basic-executor
    spec:
      serviceAccountName: executor-service-account
      containers:
      - name: executor
        image: "hourglass/executor:v1.2.0"
        env:
        - name: DEPLOYMENT_MODE
          value: "kubernetes"
        - name: KUBERNETES_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: AGGREGATOR_ENDPOINT
          value: "aggregator.example.com:9090"
        - name: LOG_LEVEL
          value: "info"
        ports:
        - containerPort: 8080
          name: metrics
        - containerPort: 8090
          name: health
        volumeMounts:
        - name: config
          mountPath: /etc/executor
        - name: data
          mountPath: /data
        resources:
          requests:
            cpu: "500m"
            memory: "1Gi"
          limits:
            cpu: "2"
            memory: "4Gi"
      volumes:
      - name: config
        configMap:
          name: executor-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
---
# Executor RBAC for managing Performers
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
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: executor-binding
  namespace: my-avs-project
subjects:
- kind: ServiceAccount
  name: executor-service-account
  namespace: my-avs-project
roleRef:
  kind: Role
  name: executor-role
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: executor-service-account
  namespace: my-avs-project
```

### Executor Configuration

```yaml
# executor-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: executor-config
  namespace: my-avs-project
data:
  config.yaml: |
    aggregator_endpoint: "aggregator.example.com:9090"
    performer_mode: "kubernetes"
    log_level: "info"
    chains:
      - name: "ethereum"
        rpc: "https://eth-mainnet.alchemyapi.io/v2/YOUR_KEY"
        chain_id: 1
        task_mailbox_address: "0x1234567890abcdef1234567890abcdef12345678"
    kubernetes:
      namespace: "my-avs-project"
      performer_service_pattern: "performer-{name}.{namespace}.svc.cluster.local:{port}"
    operator_keys:
      ecdsa: "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
      bls: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
```

### Simple Performer

```yaml
# Created by the user's Executor - managed by singleton operator
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: basic-performer
  namespace: my-avs-project
spec:
  avsAddress: "0x1234567890abcdef1234567890abcdef12345678"
  image: "myavs/performer:v1.0.0"
  version: "v1.0.0"
  config:
    grpcPort: 9090
    env:
    - name: LOG_LEVEL
      value: "info"
  resources:
    requests:
      cpu: "1"
      memory: "2Gi"
    limits:
      cpu: "4"
      memory: "8Gi"
```

### Performer with Secrets and ConfigMaps

```yaml
# Example showing environment variables from secrets and configmaps
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: secure-performer
  namespace: my-avs-project
spec:
  avsAddress: "0x1234567890abcdef1234567890abcdef12345678"
  image: "myavs/performer:v1.0.0"
  version: "v1.0.0"
  config:
    grpcPort: 9090
    env:
    # Direct value
    - name: LOG_LEVEL
      value: "info"
    # Value from secret
    - name: API_KEY
      valueFrom:
        secretKeyRef:
          name: api-secrets
          key: api-key
    # Value from configmap
    - name: CONFIG_DATA
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: config.json
    # Field reference
    - name: POD_NAME
      valueFrom:
        fieldRef:
          fieldPath: metadata.name
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          fieldPath: metadata.namespace
  resources:
    requests:
      cpu: "1"
      memory: "2Gi"
    limits:
      cpu: "4"
      memory: "8Gi"
```

## Advanced Examples

### Multi-Chain Executor with HA

```yaml
# ha-executor-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: ha-executor
  namespace: production-avs
spec:
  serviceName: ha-executor
  replicas: 3  # High availability
  selector:
    matchLabels:
      app: ha-executor
  template:
    metadata:
      labels:
        app: ha-executor
    spec:
      serviceAccountName: executor-service-account
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values: [ha-executor]
            topologyKey: kubernetes.io/hostname
      containers:
      - name: executor
        image: "hourglass/executor:v1.2.0"
        env:
        - name: DEPLOYMENT_MODE
          value: "kubernetes"
        - name: KUBERNETES_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: AGGREGATOR_ENDPOINT
          value: "aggregator-ha.example.com:9090"
        - name: LOG_LEVEL
          value: "info"
        - name: HA_MODE
          value: "true"
        - name: REPLICA_COUNT
          value: "3"
        ports:
        - containerPort: 8080
          name: metrics
        - containerPort: 8090
          name: health
        - containerPort: 9095
          name: grpc
        volumeMounts:
        - name: config
          mountPath: /etc/executor
        - name: data
          mountPath: /data
        resources:
          requests:
            cpu: "1"
            memory: "2Gi"
          limits:
            cpu: "4"
            memory: "8Gi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8090
          initialDelaySeconds: 5
          periodSeconds: 5
      nodeSelector:
        node.kubernetes.io/instance-type: "m5.xlarge"
      tolerations:
      - key: "dedicated"
        operator: "Equal"
        value: "avs-workloads"
        effect: "NoSchedule"
      volumes:
      - name: config
        configMap:
          name: ha-executor-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 50Gi
      storageClassName: fast-ssd
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
  image: "myavs/gpu-performer:v3.0.0"
  version: "v3.0.0"
  config:
    grpcPort: 9090
    env:
    - name: CUDA_VISIBLE_DEVICES
      value: "0,1"
    - name: LOG_LEVEL
      value: "debug"
    - name: GPU_MEMORY_FRACTION
      value: "0.8"
    - name: MODEL_PATH
      value: "/models/transformer.pt"
    args:
    - "--enable-gpu"
    - "--batch-size=32"
    - "--model-parallel"
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
      node.kubernetes.io/instance-type: "p3.8xlarge"
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
          - matchExpressions:
            - key: "nvidia.com/gpu.present"
              operator: In
              values: ["true"]
            - key: "nvidia.com/gpu.memory"
              operator: In
              values: ["40Gi", "80Gi"]
    tolerations:
    - key: "nvidia.com/gpu"
      operator: "Exists"
      effect: "NoSchedule"
    - key: "dedicated"
      operator: "Equal"
      value: "gpu-workloads"
      effect: "NoSchedule"
    runtimeClass: "nvidia"
    priorityClassName: "high-priority"
  hardwareRequirements:
    gpuType: "nvidia-a100"
    gpuCount: 2
    customLabels:
      gpu.memory: "40Gi"
      gpu.interconnect: "nvlink"
  imagePullSecrets:
  - name: private-registry-secret
```

### TEE-Enabled Performer

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: tee-performer
  namespace: secure-compute
spec:
  avsAddress: "0xfedcba0987654321fedcba0987654321fedcba09"
  image: "myavs/tee-performer:v2.0.0"
  version: "v2.0.0"
  config:
    grpcPort: 8080
    env:
    - name: TEE_MODE
      value: "sgx"
    - name: ENCLAVE_PATH
      value: "/opt/enclave.signed.so"
    - name: ATTESTATION_URL
      value: "https://attestation.intel.com"
    - name: LOG_LEVEL
      value: "info"
    - name: SECURITY_LEVEL
      value: "high"
    command:
    - "/usr/bin/tee-runner"
    args:
    - "--enclave=/opt/enclave.signed.so"
    - "--attestation-url=https://attestation.intel.com"
    - "--quote-provider=azure"
  resources:
    requests:
      cpu: "2"
      memory: "4Gi"
    limits:
      cpu: "4"
      memory: "8Gi"
  scheduling:
    nodeSelector:
      intel.feature.node.kubernetes.io/sgx: "true"
      node.kubernetes.io/instance-type: "m5.metal"
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
          - matchExpressions:
            - key: "intel.feature.node.kubernetes.io/sgx"
              operator: In
              values: ["true"]
            - key: "sgx.intel.com/epc"
              operator: Exists
    tolerations:
    - key: "sgx"
      operator: "Equal"
      value: "enabled"
      effect: "NoSchedule"
    - key: "dedicated"
      operator: "Equal"
      value: "tee-workloads"
      effect: "NoSchedule"
    runtimeClass: "kata-containers"
    priorityClassName: "security-critical"
  hardwareRequirements:
    teeRequired: true
    teeType: "sgx"
    customLabels:
      intel.sgx.version: "2.0"
      sgx.intel.com/epc: "128Mi"
      attestation.provider: "azure"
```

### Bottlerocket Node Performer

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: bottlerocket-performer
  namespace: secure-os
spec:
  avsAddress: "0x1111222233334444555566667777888899990000"
  image: "myavs/secure-performer:v1.8.0"
  version: "v1.8.0"
  config:
    grpcPort: 9090
    env:
    - name: SECURITY_LEVEL
      value: "high"
    - name: LOG_LEVEL
      value: "warn"
    - name: AUDIT_ENABLED
      value: "true"
    - name: READONLY_ROOTFS
      value: "true"
  resources:
    requests:
      cpu: "500m"
      memory: "1Gi"
    limits:
      cpu: "2"
      memory: "4Gi"
  scheduling:
    nodeSelector:
      node.kubernetes.io/os: "bottlerocket"
      kubernetes.io/arch: "amd64"
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
          - matchExpressions:
            - key: "node.kubernetes.io/os"
              operator: In
              values: ["bottlerocket"]
        preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          preference:
            matchExpressions:
            - key: "node.kubernetes.io/instance-type"
              operator: In
              values: ["m5.large", "m5.xlarge"]
    tolerations:
    - key: "bottlerocket"
      operator: "Equal"
      value: "true"
      effect: "NoSchedule"
    - key: "dedicated"
      operator: "Equal"
      value: "secure-workloads"
      effect: "NoSchedule"
    runtimeClass: "gvisor"
    priorityClassName: "system-critical"
```

## Multi-Namespace Deployment Scenarios

### Scenario 1: Multiple AVS Projects

```bash
# Deploy AVS Project A
kubectl create namespace avs-project-a
kubectl apply -f executor-a-statefulset.yaml -n avs-project-a
kubectl apply -f performer-a1.yaml -n avs-project-a
kubectl apply -f performer-a2.yaml -n avs-project-a

# Deploy AVS Project B
kubectl create namespace avs-project-b  
kubectl apply -f executor-b-statefulset.yaml -n avs-project-b
kubectl apply -f performer-b1.yaml -n avs-project-b

# Single operator manages all performers
kubectl get performers --all-namespaces

# Output shows performers across namespaces:
# NAMESPACE       NAME           AVS                                        PHASE     ENDPOINT
# avs-project-a   performer-a1   0x1234567890abcdef1234567890abcdef12345678   Running   performer-a1.avs-project-a.svc.cluster.local:9090
# avs-project-a   performer-a2   0x1234567890abcdef1234567890abcdef12345678   Running   performer-a2.avs-project-a.svc.cluster.local:9090
# avs-project-b   performer-b1   0xabcdef1234567890abcdef1234567890abcdef12   Running   performer-b1.avs-project-b.svc.cluster.local:9090
```

### Scenario 2: Development vs Production

```bash
# Development environment
kubectl create namespace dev-environment
kubectl apply -f executor-dev.yaml -n dev-environment
kubectl apply -f performer-dev.yaml -n dev-environment

# Production environment
kubectl create namespace production
kubectl apply -f executor-prod.yaml -n production
kubectl apply -f performer-prod.yaml -n production

# Same operator handles both environments
kubectl get performers -n dev-environment
kubectl get performers -n production
```

## Deployment Workflows

### Complete Multi-Namespace Stack Deployment

```bash
# 1. Deploy singleton operator (cluster-wide)
kubectl apply -f operator-deployment.yaml

# 2. Create namespaces for different AVS projects
kubectl create namespace ml-workloads
kubectl create namespace secure-compute
kubectl create namespace production-avs

# 3. Create secrets in each namespace
for ns in ml-workloads secure-compute production-avs; do
  kubectl create secret generic operator-keys \
    --from-literal=ecdsa="0xabcdef..." \
    --from-literal=bls="0x123456..." \
    -n $ns
    
  kubectl create secret docker-registry private-registry-secret \
    --docker-server=registry.example.com \
    --docker-username=username \
    --docker-password=password \
    -n $ns
done

# 4. Deploy executors in each namespace
kubectl apply -f ml-executor-statefulset.yaml -n ml-workloads
kubectl apply -f secure-executor-statefulset.yaml -n secure-compute
kubectl apply -f prod-executor-statefulset.yaml -n production-avs

# 5. Wait for executors to be ready
kubectl wait --for=condition=Ready pod -l app=executor -n ml-workloads --timeout=300s
kubectl wait --for=condition=Ready pod -l app=executor -n secure-compute --timeout=300s
kubectl wait --for=condition=Ready pod -l app=executor -n production-avs --timeout=300s

# 6. Verify executors create performers (they will create them automatically)
# Monitor performer creation across all namespaces
kubectl get performers --all-namespaces --watch

# 7. Verify full deployment
kubectl get statefulsets,performers,services --all-namespaces -l app=executor
```

### Scaling Operations

```bash
# Scale executor replicas in specific namespace
kubectl patch statefulset ha-executor -n production-avs \
  --type='merge' -p='{"spec":{"replicas":5}}'

# Update performer image via executor (executor will update the CRD)
kubectl patch performer gpu-performer -n ml-workloads \
  --type='merge' -p='{"spec":{"image":"myavs/gpu-performer:v3.1.0","version":"v3.1.0"}}'

# Scale resources for performer
kubectl patch performer tee-performer -n secure-compute \
  --type='merge' -p='{"spec":{"resources":{"requests":{"cpu":"4","memory":"8Gi"}}}}'
```

### Monitoring and Debugging

```bash
# Check singleton operator health
kubectl logs -n hourglass-system deployment/hourglass-operator-controller-manager

# Check executor status in specific namespace
kubectl describe statefulset ha-executor -n production-avs

# Check performer status across all namespaces
kubectl describe performers --all-namespaces

# View executor logs in specific namespace
kubectl logs statefulset/ha-executor -n production-avs

# View performer logs (via pod)
kubectl logs -l "app=performer,performer-name=gpu-performer" -n ml-workloads

# Check service endpoints across namespaces
kubectl get endpoints -l "app=performer" --all-namespaces

# Test service connectivity from executor
kubectl exec -it ha-executor-0 -n production-avs -- \
  nslookup performer-gpu-performer.ml-workloads.svc.cluster.local

# Monitor resource usage by namespace
kubectl top pods --all-namespaces -l app=performer
```

## Resource Management

### Namespace Resource Quotas

```yaml
# Apply to each AVS namespace
apiVersion: v1
kind: ResourceQuota
metadata:
  name: avs-quota
  namespace: ml-workloads
spec:
  hard:
    requests.cpu: "50"
    requests.memory: "100Gi"
    limits.cpu: "100"
    limits.memory: "200Gi"
    nvidia.com/gpu: "20"
    persistentvolumeclaims: "10"
    count/performers.hourglass.eigenlayer.io: "50"
    count/statefulsets.apps: "5"
```

### Priority Classes

```yaml
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: security-critical
value: 1000
globalDefault: false
description: "Critical security workloads (TEE performers)"
---
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high-priority
value: 500
globalDefault: false
description: "High priority GPU workloads"
```

## Network Policies

### Cross-Namespace Communication

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: performer-isolation
  namespace: ml-workloads
spec:
  podSelector:
    matchLabels:
      app: performer
  policyTypes:
  - Ingress
  - Egress
  ingress:
  # Allow from executors in same namespace
  - from:
    - podSelector:
        matchLabels:
          app: executor
    ports:
    - protocol: TCP
      port: 9090
  # Allow from operator (cluster-wide)
  - from:
    - namespaceSelector:
        matchLabels:
          name: hourglass-system
    ports:
    - protocol: TCP
      port: 9090
  egress:
  # DNS resolution
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
  # External APIs
  - to: []
    ports:
    - protocol: TCP
      port: 443
```

## Service Mesh Integration

### Istio Configuration for Multi-Namespace

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: performer-routing
  namespace: ml-workloads
spec:
  hosts:
  - performer-gpu-performer.ml-workloads.svc.cluster.local
  http:
  - match:
    - headers:
        x-executor-id:
          regex: ".*-executor"
    route:
    - destination:
        host: performer-gpu-performer.ml-workloads.svc.cluster.local
        port:
          number: 9090
      weight: 100
    timeout: 30s
    retries:
      attempts: 3
      perTryTimeout: 10s
---
# Cross-namespace service discovery
apiVersion: networking.istio.io/v1beta1
kind: ServiceEntry
metadata:
  name: cross-namespace-performers
  namespace: production-avs
spec:
  hosts:
  - performer-gpu-performer.ml-workloads.svc.cluster.local
  ports:
  - number: 9090
    name: grpc
    protocol: GRPC
  location: MESH_EXTERNAL
  resolution: DNS
```

### Prometheus Monitoring

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: executor-metrics
  namespace: ml-workloads
spec:
  selector:
    matchLabels:
      app: executor
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
---
# Cluster-wide operator monitoring
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: operator-metrics
  namespace: hourglass-system
spec:
  selector:
    matchLabels:
      app: hourglass-operator
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

## Migration from Previous Architecture

### Migration Steps

```bash
# 1. Backup existing resources
kubectl get hourglassexecutors -o yaml > executor-backup.yaml
kubectl get performers -o yaml > performer-backup.yaml

# 2. Deploy new singleton operator
kubectl apply -f singleton-operator-manifests/

# 3. Create executor StatefulSets (manual conversion required)
# Convert HourglassExecutor CRDs to StatefulSet YAMLs

# 4. Remove executorRef from Performer CRDs
kubectl patch performers --all --type=json \
  -p='[{"op": "remove", "path": "/spec/executorRef"}]'

# 5. Remove old HourglassExecutor CRDs
kubectl delete crd hourglassexecutors.hourglass.eigenlayer.io

# 6. Verify migration
kubectl get performers --all-namespaces
kubectl get statefulsets --all-namespaces -l app=executor
```

This comprehensive set of examples demonstrates the flexibility and power of the singleton Hourglass Kubernetes Operator across various deployment scenarios, supporting multiple independent executor deployments while maintaining centralized performer management.