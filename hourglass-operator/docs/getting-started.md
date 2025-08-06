# Getting Started with Hourglass Kubernetes Operator

This guide will walk you through deploying the Hourglass Kubernetes Operator and running your first AVS executor in a Kubernetes cluster.

## Architecture Overview

The Hourglass Kubernetes Operator uses a **singleton architecture**:

- **Single Operator**: One `hourglass-operator` deployment manages all `Performer` CRDs cluster-wide
- **Multiple Executors**: Each AVS team deploys their own `executor` as a StatefulSet
- **Automated Performers**: Executors create `Performer` CRDs, which the operator turns into running pods

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Kubernetes Cluster                                 │
│                                                                             │
│  ┌──────────────────┐              ┌──────────────────────────────────────┐ │
│  │ Hourglass        │              │        Multiple User Executors       │ │
│  │ Operator         │              │                                      │ │
│  │ (Singleton)      │              │ ┌──────────────┐  ┌─────────────────┐ │ │
│  │                  │              │ │ AVS-A        │  │ AVS-B           │ │ │
│  │ Manages All      │◄─────────────┤ │ Executor     │  │ Executor        │ │ │
│  │ Performer CRDs   │ Creates CRDs │ │ StatefulSet  │  │ StatefulSet     │ │ │
│  │                  │              │ │              │  │                 │ │ │
│  └──────────────────┘              │ └──────────────┘  └─────────────────┘ │ │
│           │                        └──────────────────────────────────────┘ │
│           │ Creates Performer Pods                                          │
│           ▼                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                    Performer Pods & Services                            │ │
│  │ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐ │ │
│  │ │AVS-A Perf-1 │ │AVS-A Perf-2 │ │AVS-B Perf-1 │ │AVS-B Perf-2         │ │ │
│  │ │             │ │             │ │             │ │                     │ │ │
│  │ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────────────┘ │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Prerequisites

### 1. Kubernetes Cluster
- Kubernetes v1.26+ (tested with v1.28+)
- kubectl configured to access your cluster
- Cluster admin permissions (for CRD installation)

### 2. Required Tools
```bash
# Install Helm (if not already installed)
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install kubectl (if not already installed)
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl && sudo mv kubectl /usr/local/bin/
```

### 3. Aggregator Endpoint
You need a running Hourglass aggregator that your executors can connect to.

## Step 1: Deploy the Hourglass Operator

### Option A: Using kubectl (Recommended)

```bash
# Clone the repository
git clone https://github.com/Layr-Labs/hourglass-monorepo.git
cd hourglass-monorepo/hourglass-operator

# Apply the CRDs
kubectl apply -f config/crd/bases/

# Create the operator namespace
kubectl create namespace hourglass-system

# Apply RBAC
kubectl apply -f config/rbac/ -n hourglass-system

# Deploy the operator
kubectl apply -f config/manager/manager.yaml -n hourglass-system
```

### Option B: Using Helm

```bash
# Install the operator using the local chart
cd hourglass-monorepo/hourglass-operator
helm install hourglass-operator ./charts/hourglass-operator \
  --namespace hourglass-system \
  --create-namespace
```

### Verify Operator Installation

```bash
# Check operator pod status
kubectl get pods -n hourglass-system

# Check CRD installation
kubectl get crd performers.hourglass.eigenlayer.io

# Check operator logs
kubectl logs -n hourglass-system -l app=hourglass-operator
```

## Step 2: Prepare Your AVS Configuration

### 1. Create Your Namespace
```bash
# Create a namespace for your AVS
kubectl create namespace my-avs-project
```

### 2. Generate Operator Keys

```bash
# Generate ECDSA private key
openssl ecparam -genkey -name secp256k1 -noout -out ecdsa-private-key.pem

# Convert to base64 for Kubernetes secret
ECDSA_KEY=$(cat ecdsa-private-key.pem | base64 -w 0)

# Optional: Generate BLS private key (if your AVS uses BLS)
# BLS_KEY=$(your-bls-key-generation-method | base64 -w 0)
```

### 3. Prepare Configuration Values

Create a values file for your executor:

```yaml
# my-avs-values.yaml
executor:
  name: "my-avs-executor"
  replicaCount: 1
  
  image:
    repository: hourglass/executor
    tag: "v1.2.0"

# Aggregator configuration (REQUIRED)
aggregator:
  endpoint: "your-aggregator.example.com:9090"
  tls:
    enabled: false

# Blockchain configuration (REQUIRED)
chains:
  ethereum:
    enabled: true
    rpcUrl: "https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY"
    taskMailboxAddress: "0x1234567890abcdef1234567890abcdef12345678"
    blockConfirmations: 12

# AVS configuration (REQUIRED)
avs:
  supportedAvs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "my-avs"
      performer:
        image: "my-org/my-avs-performer"
        version: "v1.0.0"
      resources:
        requests:
          cpu: 500m
          memory: 1Gi
        limits:
          cpu: 2
          memory: 4Gi

# Secrets (REQUIRED)
secrets:
  operatorKeys:
    ecdsaPrivateKey: "LS0tLS1CRUdJTi..."  # Replace with your base64 key
    # blsPrivateKey: "LS0tLS1CRUdJTi..."  # Optional

# Storage configuration
persistence:
  enabled: true
  size: 10Gi
```

## Step 3: Deploy Your Executor

### Using Helm

```bash
# Install your executor using the official Hourglass chart
cd hourglass-monorepo
helm install my-avs-executor ./ponos/charts/hourglass \
  --namespace my-avs-project \
  --values my-avs-values.yaml
```

### Using kubectl

If you prefer kubectl, use the example template:

```bash
# Copy and customize the basic executor template
cp templates/examples/basic-executor.yaml my-avs-executor.yaml

# Edit the configuration values
vim my-avs-executor.yaml

# Apply the configuration
kubectl apply -f my-avs-executor.yaml
```

## Step 4: Verify Your Deployment

### Check Executor Status

```bash
# Check executor pod
kubectl get pods -n my-avs-project

# Check executor logs
kubectl logs -n my-avs-project -l app=my-avs-executor

# Check executor service
kubectl get svc -n my-avs-project
```

### Check Performer CRDs

```bash
# List all performers
kubectl get performers -n my-avs-project

# Get detailed performer status
kubectl describe performer -n my-avs-project

# Check performer pods (managed by operator)
kubectl get pods -n my-avs-project -l app=hourglass-performer
```

### Check Connectivity

```bash
# Test gRPC endpoint connectivity
kubectl exec -n my-avs-project deploy/my-avs-executor -- \
  grpcurl -plaintext performer-my-avs.my-avs-project.svc.cluster.local:9090 list

# Check operator logs for performer management
kubectl logs -n hourglass-system -l app=hourglass-operator
```

## Step 5: Test Task Execution

### Submit a Test Task

```bash
# Port-forward to access executor locally
kubectl port-forward -n my-avs-project svc/my-avs-executor 9095:9095

# Submit a test task (in another terminal)
grpcurl -plaintext -d '{
  "avsAddress": "0x1234567890abcdef1234567890abcdef12345678",
  "taskId": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
  "payload": "test-payload"
}' localhost:9095 eigenlayer.hourglass.v1.ExecutorService/SubmitTask
```

### Monitor Task Processing

```bash
# Check executor logs for task processing
kubectl logs -n my-avs-project -l app=my-avs-executor -f

# Check performer logs for task execution
kubectl logs -n my-avs-project -l app=hourglass-performer -f

# Check operator logs for performer lifecycle events
kubectl logs -n hourglass-system -l app=hourglass-operator -f
```

## Step 6: Scaling and Management

### Scale Your Executor

```bash
# Scale executor replicas
kubectl scale statefulset my-avs-executor -n my-avs-project --replicas=3

# Check scaled pods
kubectl get pods -n my-avs-project -l app=my-avs-executor
```

### Upgrade Performer Images

```bash
# Update performer CRD with new image
kubectl patch performer my-avs -n my-avs-project --type='merge' -p='{
  "spec": {
    "image": "my-org/my-avs-performer:v1.1.0",
    "version": "v1.1.0"
  }
}'

# Watch rolling update
kubectl get pods -n my-avs-project -l hourglass.eigenlayer.io/performer=my-avs -w
```

## Advanced Configuration

### GPU Workloads

```yaml
# In your values file
avs:
  supportedAvs:
    - address: "0x..."
      name: "gpu-avs"
      performer:
        image: "my-org/gpu-performer"
        version: "v1.0.0"
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

### Secure Node Scheduling

```yaml
# Schedule performers on Bottlerocket nodes
avs:
  supportedAvs:
    - address: "0x..."
      name: "secure-avs"
      performer:
        image: "my-org/secure-performer"
        version: "v1.0.0"
      scheduling:
        nodeSelector:
          node.kubernetes.io/os: bottlerocket
        tolerations:
        - key: "bottlerocket"
          operator: "Equal"
          value: "true"
          effect: "NoSchedule"
```

### TEE-Enabled Performers

```yaml
# For Intel SGX workloads
avs:
  supportedAvs:
    - address: "0x..."
      name: "tee-avs"
      performer:
        image: "my-org/tee-performer"
        version: "v1.0.0"
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

## Troubleshooting

### Common Issues

1. **Operator Not Starting**
   ```bash
   # Check operator logs
   kubectl logs -n hourglass-system -l app=hourglass-operator
   
   # Check RBAC permissions
   kubectl auth can-i create performers --as=system:serviceaccount:hourglass-system:operator
   ```

2. **Executor Can't Connect to Aggregator**
   ```bash
   # Check network connectivity
   kubectl exec -n my-avs-project deploy/my-avs-executor -- nc -zv aggregator.example.com 9090
   
   # Check TLS configuration
   kubectl logs -n my-avs-project -l app=my-avs-executor | grep -i tls
   ```

3. **Performer Pods Not Starting**
   ```bash
   # Check performer CRD status
   kubectl describe performer my-avs -n my-avs-project
   
   # Check operator logs for errors
   kubectl logs -n hourglass-system -l app=hourglass-operator | grep -i error
   
   # Check resource constraints
   kubectl describe nodes
   ```

4. **DNS Resolution Issues**
   ```bash
   # Test DNS resolution from executor
   kubectl exec -n my-avs-project deploy/my-avs-executor -- nslookup performer-my-avs.my-avs-project.svc.cluster.local
   
   # Check CoreDNS logs
   kubectl logs -n kube-system -l k8s-app=kube-dns
   ```

### Getting Help

- **Documentation**: Check the [operator documentation](./operator/)
- **Examples**: See [example configurations](../templates/examples/)
- **Issues**: Report problems on [GitHub Issues](https://github.com/Layr-Labs/hourglass-monorepo/issues)

## Next Steps

1. **Production Deployment**: Review the [production deployment guide](./operator/production-deployment.md)
2. **Monitoring**: Set up [monitoring and alerting](./operator/monitoring.md)
3. **Security**: Implement [security best practices](./operator/security.md)
4. **Multi-Chain**: Configure [multi-chain support](./operator/multi-chain.md)

## Configuration Reference

For detailed configuration options, see:
- [API Reference](./operator/api-reference.md)
- [Helm Chart Values](../../ponos/charts/hourglass/values.yaml)
- [Example Configurations](../templates/examples/)