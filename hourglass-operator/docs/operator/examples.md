# Usage Examples

This document provides practical examples for deploying and configuring the Hourglass Kubernetes Operator.

## Basic Examples

### Simple Executor

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: HourglassExecutor
metadata:
  name: basic-executor
  namespace: hourglass
spec:
  image: "hourglass/executor:v1.0.0"
  replicas: 1
  config:
    aggregatorEndpoint: "aggregator.example.com:9090"
    performerMode: "kubernetes"
    logLevel: "info"
    chains:
    - name: "ethereum"
      rpc: "https://eth-mainnet.alchemyapi.io/v2/YOUR_KEY"
      chainId: 1
      taskMailboxAddress: "0x1234567890abcdef1234567890abcdef12345678"
    operatorKeys:
      ecdsa: "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
      bls: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
  resources:
    requests:
      cpu: "500m"
      memory: "1Gi"
    limits:
      cpu: "2"
      memory: "4Gi"
```

### Simple Performer

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: basic-performer
  namespace: hourglass
spec:
  avsAddress: "0x1234567890abcdef1234567890abcdef12345678"
  image: "myavs/performer:v1.0.0"
  version: "v1.0.0"
  config:
    grpcPort: 9090
    environment:
      LOG_LEVEL: "info"
  resources:
    requests:
      cpu: "1"
      memory: "2Gi"
    limits:
      cpu: "4"
      memory: "8Gi"
```

## Advanced Examples

### High-Availability Executor

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: HourglassExecutor
metadata:
  name: ha-executor
  namespace: hourglass
spec:
  image: "hourglass/executor:v1.2.0"
  replicas: 3
  config:
    aggregatorEndpoint: "aggregator-ha.example.com:9090"
    performerMode: "kubernetes"
    logLevel: "info"
    chains:
    - name: "ethereum"
      rpc: "https://eth-mainnet.alchemyapi.io/v2/YOUR_KEY"
      chainId: 1
      taskMailboxAddress: "0x1234567890abcdef1234567890abcdef12345678"
    - name: "base"
      rpc: "https://base-mainnet.infura.io/v3/YOUR_KEY"
      chainId: 8453
      taskMailboxAddress: "0xabcdef1234567890abcdef1234567890abcdef12"
    - name: "arbitrum"
      rpc: "https://arbitrum-mainnet.infura.io/v3/YOUR_KEY"
      chainId: 42161
      taskMailboxAddress: "0xfedcba0987654321fedcba0987654321fedcba09"
    operatorKeys:
      ecdsa: "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
      bls: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
    kubernetes:
      namespace: "hourglass"
      defaultScheduling:
        nodeSelector:
          node-role.kubernetes.io/worker: "true"
        tolerations:
        - key: "dedicated"
          operator: "Equal"
          value: "hourglass"
          effect: "NoSchedule"
  resources:
    requests:
      cpu: "1"
      memory: "2Gi"
    limits:
      cpu: "4"
      memory: "8Gi"
  nodeSelector:
    node.kubernetes.io/instance-type: "m5.xlarge"
  tolerations:
  - key: "dedicated"
    operator: "Equal"
    value: "hourglass"
    effect: "NoSchedule"
  imagePullSecrets:
  - name: private-registry-secret
```

### GPU-Enabled Performer

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: gpu-performer
  namespace: hourglass
spec:
  avsAddress: "0xabcdef1234567890abcdef1234567890abcdef12"
  image: "myavs/gpu-performer:v2.0.0"
  version: "v2.0.0"
  executorRef: "ha-executor"
  config:
    grpcPort: 9090
    environment:
      CUDA_VISIBLE_DEVICES: "0"
      LOG_LEVEL: "debug"
      GPU_MEMORY_FRACTION: "0.8"
    args:
    - "--enable-gpu"
    - "--batch-size=32"
  resources:
    requests:
      cpu: "2"
      memory: "4Gi"
    limits:
      cpu: "8"
      memory: "16Gi"
  scheduling:
    nodeSelector:
      accelerator: "nvidia-tesla-a100"
      node.kubernetes.io/instance-type: "p3.2xlarge"
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: "nvidia.com/gpu.present"
            operator: In
            values: ["true"]
    tolerations:
    - key: "nvidia.com/gpu"
      operator: "Exists"
      effect: "NoSchedule"
    - key: "dedicated"
      operator: "Equal"
      value: "gpu-workloads"
      effect: "NoSchedule"
  hardwareRequirements:
    gpuType: "nvidia-a100"
    gpuCount: 1
  imagePullSecrets:
  - name: private-registry-secret
```

### TEE-Enabled Performer

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: tee-performer
  namespace: hourglass
spec:
  avsAddress: "0xfedcba0987654321fedcba0987654321fedcba09"
  image: "myavs/tee-performer:v1.5.0"
  version: "v1.5.0"
  executorRef: "ha-executor"
  config:
    grpcPort: 9090
    environment:
      TEE_MODE: "sgx"
      ENCLAVE_PATH: "/opt/enclave.so"
      LOG_LEVEL: "info"
    command:
    - "/usr/bin/tee-runner"
    args:
    - "--enclave=/opt/enclave.so"
    - "--attestation-url=https://attestation.intel.com"
  resources:
    requests:
      cpu: "1"
      memory: "2Gi"
    limits:
      cpu: "4"
      memory: "8Gi"
  scheduling:
    nodeSelector:
      intel.feature.node.kubernetes.io/sgx: "true"
      node.kubernetes.io/instance-type: "m5.metal"
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: "intel.feature.node.kubernetes.io/sgx"
            operator: In
            values: ["true"]
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
  hardwareRequirements:
    teeRequired: true
    teeType: "sgx"
    customLabels:
      intel.sgx.version: "2.0"
      intel.sgx.epc-size: "128MB"
```

### Bottlerocket Node Performer

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: bottlerocket-performer
  namespace: hourglass
spec:
  avsAddress: "0x1111222233334444555566667777888899990000"
  image: "myavs/secure-performer:v1.8.0"
  version: "v1.8.0"
  executorRef: "ha-executor"
  config:
    grpcPort: 9090
    environment:
      SECURITY_LEVEL: "high"
      LOG_LEVEL: "warn"
      AUDIT_ENABLED: "true"
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
    nodeAffinity:
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
```

## Deployment Workflows

### Complete Stack Deployment

```bash
# 1. Create namespace
kubectl create namespace hourglass

# 2. Create secrets for operator keys
kubectl create secret generic operator-keys \
  --from-literal=ecdsa="0xabcdef..." \
  --from-literal=bls="0x123456..." \
  -n hourglass

# 3. Create private registry secret
kubectl create secret docker-registry private-registry-secret \
  --docker-server=registry.example.com \
  --docker-username=username \
  --docker-password=password \
  -n hourglass

# 4. Deploy executor
kubectl apply -f ha-executor.yaml

# 5. Wait for executor to be ready
kubectl wait --for=condition=Available deployment/ha-executor -n hourglass --timeout=300s

# 6. Deploy performers
kubectl apply -f gpu-performer.yaml
kubectl apply -f tee-performer.yaml
kubectl apply -f bottlerocket-performer.yaml

# 7. Verify deployment
kubectl get hourglassexecutors,performers -n hourglass
```

### Scaling Operations

```bash
# Scale executor replicas
kubectl patch hourglassexecutor ha-executor -n hourglass \
  --type='merge' -p='{"spec":{"replicas":5}}'

# Update performer image
kubectl patch performer gpu-performer -n hourglass \
  --type='merge' -p='{"spec":{"image":"myavs/gpu-performer:v2.1.0","version":"v2.1.0"}}'
```

### Monitoring and Debugging

```bash
# Check executor status
kubectl describe hourglassexecutor ha-executor -n hourglass

# Check performer status
kubectl describe performer gpu-performer -n hourglass

# View executor logs
kubectl logs deployment/ha-executor -n hourglass

# View performer logs
kubectl logs performer-gpu-performer -n hourglass

# Check service endpoints
kubectl get endpoints -n hourglass

# Test service connectivity
kubectl run -it --rm debug --image=busybox --restart=Never -- \
  nslookup performer-gpu-performer.hourglass.svc.cluster.local
```

## Resource Quotas and Limits

### Namespace Resource Quota

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: hourglass-quota
  namespace: hourglass
spec:
  hard:
    requests.cpu: "20"
    requests.memory: "40Gi"
    limits.cpu: "50"
    limits.memory: "100Gi"
    nvidia.com/gpu: "10"
    persistentvolumeclaims: "10"
    count/hourglassexecutors.hourglass.eigenlayer.io: "5"
    count/performers.hourglass.eigenlayer.io: "20"
```

### Limit Range

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: hourglass-limits
  namespace: hourglass
spec:
  limits:
  - default:
      cpu: "2"
      memory: "4Gi"
    defaultRequest:
      cpu: "500m"
      memory: "1Gi"
    type: Container
  - max:
      cpu: "8"
      memory: "32Gi"
    min:
      cpu: "100m"
      memory: "128Mi"
    type: Container
```

## Network Policies

### Isolation Policy

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: hourglass-isolation
  namespace: hourglass
spec:
  podSelector:
    matchLabels:
      app: hourglass-performer
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: hourglass-executor
    ports:
    - protocol: TCP
      port: 9090
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
  - to: []
    ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 80
```

## Service Mesh Integration

### Istio Configuration

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: performer-routing
  namespace: hourglass
spec:
  hosts:
  - performer-gpu-performer.hourglass.svc.cluster.local
  http:
  - match:
    - headers:
        x-executor-id:
          exact: "ha-executor"
    route:
    - destination:
        host: performer-gpu-performer.hourglass.svc.cluster.local
        port:
          number: 9090
      weight: 100
    timeout: 30s
    retries:
      attempts: 3
      perTryTimeout: 10s
```

### Service Monitor for Prometheus

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: hourglass-metrics
  namespace: hourglass
spec:
  selector:
    matchLabels:
      app: hourglass-executor
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

This comprehensive set of examples demonstrates the flexibility and power of the Hourglass Kubernetes Operator across various deployment scenarios and requirements.