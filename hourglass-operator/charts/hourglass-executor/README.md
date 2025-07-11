# Hourglass Executor Helm Chart

This Helm chart deploys a Hourglass Executor that works with the singleton Hourglass Operator to manage performer workloads in Kubernetes.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- The Hourglass Operator must be deployed in the cluster
- The `performers.hourglass.eigenlayer.io` CRD must be installed

## Installing the Chart

### Quick Start

1. **Create a values file** with your configuration:

```yaml
# my-executor-values.yaml
executor:
  name: "my-avs-executor"

aggregator:
  endpoint: "aggregator.example.com:9090"

chains:
  ethereum:
    rpcUrl: "https://eth-mainnet.alchemyapi.io/v2/YOUR_KEY"
    taskMailboxAddress: "0x1234567890abcdef1234567890abcdef12345678"

avs:
  supportedAvs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "my-avs"
      performer:
        image: "my-org/my-avs-performer"
        version: "v1.0.0"

secrets:
  operatorKeys:
    ecdsaPrivateKey: "your-base64-encoded-ecdsa-private-key"
```

2. **Install the chart**:

```bash
helm install my-executor ./charts/hourglass-executor \
  --namespace my-avs-namespace \
  --create-namespace \
  --values my-executor-values.yaml
```

### Alternative Installation Methods

#### Using Helm Repository (when published)

```bash
# Add the Hourglass Helm repository
helm repo add hourglass https://helm.hourglass.com
helm repo update

# Install the executor
helm install my-executor hourglass/hourglass-executor \
  --namespace my-avs-namespace \
  --create-namespace \
  --values my-executor-values.yaml
```

#### Using CLI Parameters

```bash
helm install my-executor ./charts/hourglass-executor \
  --namespace my-avs-namespace \
  --create-namespace \
  --set executor.name=my-executor \
  --set aggregator.endpoint=aggregator.example.com:9090 \
  --set chains.ethereum.rpcUrl=https://eth-mainnet.alchemyapi.io/v2/YOUR_KEY \
  --set chains.ethereum.taskMailboxAddress=0x1234567890abcdef1234567890abcdef12345678 \
  --set avs.supportedAvs[0].address=0x1234567890abcdef1234567890abcdef12345678 \
  --set avs.supportedAvs[0].performer.image=my-org/my-avs-performer \
  --set secrets.operatorKeys.ecdsaPrivateKey=your-base64-encoded-key
```

## Configuration

### Required Values

These values **must** be set for the executor to function:

| Parameter | Description | Example |
|-----------|-------------|---------|
| `aggregator.endpoint` | Aggregator service endpoint | `aggregator.example.com:9090` |
| `chains.ethereum.rpcUrl` | Ethereum RPC endpoint | `https://eth-mainnet.alchemyapi.io/v2/KEY` |
| `chains.ethereum.taskMailboxAddress` | TaskMailbox contract address | `0x1234...` |
| `avs.supportedAvs[].address` | AVS contract address | `0x1234...` |
| `avs.supportedAvs[].performer.image` | Performer container image | `my-org/performer:v1.0.0` |
| `secrets.operatorKeys.ecdsaPrivateKey` | Base64-encoded ECDSA private key | `LS0tLS1CRUdJTi...` |

### Common Configuration Examples

#### Basic Development Setup

```yaml
executor:
  name: "dev-executor"
  replicaCount: 1

aggregator:
  endpoint: "localhost:9090"
  tls:
    enabled: false

chains:
  ethereum:
    rpcUrl: "http://localhost:8545"  # Local Anvil
    taskMailboxAddress: "0x5FbDB2315678afecb367f032d93F642f64180aa3"

avs:
  supportedAvs:
    - address: "0x5FbDB2315678afecb367f032d93F642f64180aa3"
      name: "dev-avs"
      performer:
        image: "my-avs/performer"
        version: "latest"

development:
  enabled: true
  mockServices: true
```

#### Production High-Availability Setup

```yaml
executor:
  name: "prod-executor"
  replicaCount: 3
  resources:
    requests:
      cpu: "2"
      memory: "4Gi"
    limits:
      cpu: "8"
      memory: "16Gi"

aggregator:
  endpoint: "aggregator-ha.production.com:9090"
  tls:
    enabled: true

scheduling:
  podAntiAffinity:
    enabled: true
    type: required
  nodeSelector:
    node.kubernetes.io/instance-type: "m5.xlarge"
  tolerations:
    - key: "dedicated"
      operator: "Equal"
      value: "avs-workloads"
      effect: "NoSchedule"

persistence:
  enabled: true
  storageClass: "fast-ssd"
  size: "100Gi"

metrics:
  serviceMonitor:
    enabled: true
    namespace: "monitoring"

podDisruptionBudget:
  enabled: true
  minAvailable: 2

secrets:
  tls:
    enabled: true
    cert: "LS0tLS1CRUdJTi..."
    key: "LS0tLS1CRUdJTi..."
```

#### GPU-Enabled ML Workloads

```yaml
executor:
  name: "ml-executor"

avs:
  supportedAvs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "ml-avs"
      performer:
        image: "ml-project/gpu-performer"
        version: "v3.0.0"
      resources:
        requests:
          nvidia.com/gpu: "1"
          cpu: "4"
          memory: "8Gi"
        limits:
          nvidia.com/gpu: "1"
          cpu: "16"
          memory: "32Gi"
      hardware:
        gpu:
          required: true
          type: "nvidia-a100"
          count: 1
      scheduling:
        nodeSelector:
          accelerator: "nvidia-tesla-a100"
        tolerations:
          - key: "nvidia.com/gpu"
            operator: "Exists"
            effect: "NoSchedule"
```

#### TEE (Trusted Execution Environment) Setup

```yaml
executor:
  name: "secure-executor"

avs:
  supportedAvs:
    - address: "0xabcdef1234567890abcdef1234567890abcdef12"
      name: "secure-avs"
      performer:
        image: "secure-project/tee-performer"
        version: "v1.5.0"
      hardware:
        tee:
          required: true
          type: "sgx"
      scheduling:
        nodeSelector:
          intel.feature.node.kubernetes.io/sgx: "true"
        tolerations:
          - key: "sgx"
            operator: "Equal"
            value: "enabled"
            effect: "NoSchedule"
        runtimeClass: "kata-containers"

# Enhanced security for TEE workloads
networkPolicy:
  enabled: true
```

### Multi-Chain Configuration

```yaml
chains:
  ethereum:
    enabled: true
    rpcUrl: "https://eth-mainnet.infura.io/v3/YOUR_PROJECT_ID"
    wsUrl: "wss://eth-mainnet.infura.io/ws/v3/YOUR_PROJECT_ID"
    taskMailboxAddress: "0x1234567890abcdef1234567890abcdef12345678"
    
  base:
    enabled: true
    rpcUrl: "https://mainnet.base.org"
    wsUrl: "wss://mainnet.base.org"
    taskMailboxAddress: "0xabcdef1234567890abcdef1234567890abcdef12"
```

## Upgrading

### Helm Upgrade

```bash
helm upgrade my-executor ./charts/hourglass-executor \
  --namespace my-avs-namespace \
  --values my-executor-values.yaml
```

### Rolling Back

```bash
# List releases
helm history my-executor -n my-avs-namespace

# Roll back to previous version
helm rollback my-executor 1 -n my-avs-namespace
```

## Monitoring

### Metrics

The executor exposes Prometheus metrics on port 8080 at `/metrics`. Common metrics include:

- `performer_count` - Number of active performers
- `task_processing_duration` - Time taken to process tasks
- `performer_connection_errors` - Number of performer connection errors

### Health Checks

Health check endpoints:

- `/health` - Liveness probe
- `/ready` - Readiness probe  
- `/startup` - Startup probe

### Logs

View executor logs:

```bash
kubectl logs -f statefulset/my-executor -n my-avs-namespace
```

View performer status:

```bash
kubectl get performers -n my-avs-namespace
kubectl describe performer my-performer -n my-avs-namespace
```

## Troubleshooting

### Common Issues

#### 1. Executor Pod Not Starting

Check the pod events and logs:

```bash
kubectl describe pod my-executor-0 -n my-avs-namespace
kubectl logs my-executor-0 -n my-avs-namespace
```

Common causes:
- Missing required configuration values
- Invalid private keys
- Insufficient RBAC permissions
- Node resource constraints

#### 2. Performer Creation Failing

Check the operator logs:

```bash
kubectl logs -f deployment/hourglass-operator -n hourglass-system
```

Check performer status:

```bash
kubectl get performers -n my-avs-namespace -o yaml
```

#### 3. Connection Issues

Verify network connectivity:

```bash
# Test aggregator connection from executor pod
kubectl exec -it my-executor-0 -n my-avs-namespace -- \
  curl -v telnet://aggregator.example.com:9090

# Test performer service discovery
kubectl exec -it my-executor-0 -n my-avs-namespace -- \
  nslookup performer-my-performer.my-avs-namespace.svc.cluster.local
```

### Debug Mode

Enable debug logging:

```yaml
executor:
  env:
    logLevel: debug
```

## Uninstalling

Remove the executor:

```bash
helm uninstall my-executor -n my-avs-namespace
```

Clean up namespace (if desired):

```bash
kubectl delete namespace my-avs-namespace
```

## Values Reference

See [values.yaml](values.yaml) for the complete list of configurable values.

## Contributing

Please read the [Contributing Guide](../../CONTRIBUTING.md) for details on how to contribute to this project.

## License

This chart is licensed under the [MIT License](../../LICENSE).