# Hourglass Kubernetes Infrastructure

This CDK project deploys a complete Hourglass AVS infrastructure on AWS using EKS, ECS, and other AWS services.

## Overview

The infrastructure includes:

1. **Amazon EKS Cluster**: Kubernetes cluster for running AVS workloads
2. **Anvil Devnet**: L1/L2 blockchain nodes running on ECS Fargate
3. **Hourglass Services**: Ponos aggregator and executor deployed on EKS
4. **Obsidian Operator**: Kubernetes operator for managing AVS resources (without TEE requirements)

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         AWS Account                              │
│                                                                  │
│  ┌────────────────────┐        ┌─────────────────────────────┐ │
│  │    ECS Fargate     │        │         EKS Cluster         │ │
│  │  ┌──────────────┐  │        │  ┌───────────────────────┐  │ │
│  │  │  Anvil L1    │  │        │  │  Hourglass Namespace  │  │ │
│  │  │  Port: 8545  │  │        │  │  - Aggregator         │  │ │
│  │  └──────────────┘  │        │  │  - Executor           │  │ │
│  │  ┌──────────────┐  │        │  └───────────────────────┘  │ │
│  │  │  Anvil L2    │  │        │  ┌───────────────────────┐  │ │
│  │  │  Port: 9545  │  │        │  │  Obsidian Namespace   │  │ │
│  │  └──────────────┘  │        │  │  - AVS Operator       │  │ │
│  └────────────────────┘        │  └───────────────────────┘  │ │
│                                │  ┌───────────────────────┐  │ │
│                                │  │   Default Namespace   │  │ │
│                                │  │  - Example AVS        │  │ │
│                                │  │  - Monitoring AVS     │  │ │
│                                │  └───────────────────────┘  │ │
│                                └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

- AWS CLI configured with appropriate credentials
- Node.js 18+ and npm
- Docker (for building container images)
- kubectl (for interacting with the EKS cluster)

## Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd cdk/hourglass-k8s

# Deploy everything
./deploy.sh

# Configure kubectl
aws eks update-kubeconfig --name hourglass-eks --region us-east-1

# Check deployment status
kubectl get pods -A
kubectl get avs -A
```

## Deployment

### Manual Deployment

```bash
# Install dependencies
npm install

# Bootstrap CDK (first time only)
cdk bootstrap

# Deploy all stacks
cdk deploy --all

# Or deploy individual stacks
cdk deploy HourglassK8sCluster
cdk deploy HourglassAnvilDevnet
cdk deploy HourglassServices
cdk deploy ObsidianOperator
```

### Configuration

The deployment uses configuration files in the `config/` directory:

- `accounts.json`: Test accounts for the devnet
- `devnet.json`: Devnet configuration including chain IDs, RPC URLs, and contract addresses

## Stack Details

### 1. K8s Cluster Stack (`k8s-cluster-stack.ts`)

Creates an EKS cluster with:
- 2 node groups: general workloads and executor workloads (with Docker-in-Docker support)
- AWS Load Balancer Controller
- EBS CSI Driver
- GP3 storage class

### 2. Anvil Devnet Stack (`anvil-devnet-stack.ts`)

Deploys Anvil nodes on ECS Fargate:
- L1 node (Ethereum Holesky fork)
- L2 node (Base Sepolia fork)
- Persistent state using EFS
- Internal Network Load Balancers

### 3. Hourglass Services Stack (`hourglass-services-stack.ts`)

Deploys Ponos components:
- Aggregator service with gRPC endpoint
- Executor service with Docker-in-Docker capability
- ConfigMaps with service configuration
- LoadBalancer for external access

### 4. Obsidian Operator Stack (`obsidian-operator-stack.ts`)

Deploys the AVS operator:
- Custom Resource Definition (CRD) for AVS
- Operator deployment without TEE requirements
- Example AVS deployments for testing

## Usage

### Creating an AVS

```yaml
apiVersion: hourglass.io/v1alpha1
kind: AVS
metadata:
  name: my-avs
  namespace: default
spec:
  operator: "operator-address"
  serviceImage: "my-avs-image:latest"
  replicas: 3
  computeRequirements:
    cpu: "500m"
    memory: "1Gi"
    teeType: "NONE"  # No TEE required
  servicePort: 8080
```

Apply the manifest:
```bash
kubectl apply -f my-avs.yaml
```

### Accessing Services

```bash
# Get aggregator endpoint
kubectl get svc -n hourglass aggregator-lb

# Port forward for local testing
kubectl port-forward -n hourglass svc/aggregator-lb 9000:9000

# Submit a task via gRPC
grpcurl -plaintext \
  -d '{"avsAddress": "0x...", "taskId": "0x...", "payload": "..."}' \
  localhost:9000 eigenlayer.hourglass.v1.AggregatorService/SubmitTask
```

### Monitoring

```bash
# View operator logs
kubectl logs -n obsidian-system deployment/avs-operator -f

# View AVS status
kubectl describe avs my-avs

# Check Ponos services
kubectl get pods -n hourglass
kubectl logs -n hourglass deployment/aggregator -f
kubectl logs -n hourglass deployment/executor -f
```

## Cost Optimization

The infrastructure uses several cost-saving measures:
- Spot instances for general workloads
- Single NAT gateway
- Fargate for Anvil nodes (pay per use)
- Auto-scaling for node groups

## Cleanup

```bash
# Destroy all stacks
cdk destroy --all

# Or use the script
./destroy.sh
```

## Troubleshooting

### EKS Access Issues
```bash
# Update kubeconfig
aws eks update-kubeconfig --name hourglass-eks --region us-east-1

# Check cluster status
aws eks describe-cluster --name hourglass-eks
```

### Pod Issues
```bash
# Check pod events
kubectl describe pod <pod-name> -n <namespace>

# View logs
kubectl logs <pod-name> -n <namespace>
```

### Anvil Connection Issues
```bash
# Check ECS task status
aws ecs list-tasks --cluster hourglass-anvil-devnet
aws ecs describe-tasks --cluster hourglass-anvil-devnet --tasks <task-arn>
```

## Integration with devkit

This infrastructure is compatible with the existing devkit commands:

```bash
# Configure devkit to use the deployed infrastructure
export AGGREGATOR_URL=$(kubectl get svc -n hourglass aggregator-lb -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'):9000
export L1_RPC_URL=http://<anvil-l1-endpoint>:8545
export L2_RPC_URL=http://<anvil-l2-endpoint>:8545

# Use devkit commands as usual
devkit avs submit-task --avs-address 0x...
```

## Security Considerations

- All services run in private subnets
- IAM roles follow least privilege principle
- Secrets should be stored in AWS Secrets Manager (not implemented in this example)
- Network policies can be added for additional isolation

## Future Enhancements

1. **Production Readiness**
   - Use AWS Secrets Manager for sensitive data
   - Implement proper contract deployment (not mocked)
   - Add monitoring with CloudWatch and X-Ray
   - Implement backup strategies

2. **TEE Support**
   - Add SEV-SNP capable instances
   - Implement real attestation service
   - Add attestation verification

3. **Scaling**
   - Implement Horizontal Pod Autoscaler
   - Add Cluster Autoscaler
   - Optimize resource requests/limits

## License

See LICENSE file in the repository root.