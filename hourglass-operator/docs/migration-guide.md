# Migration Guide: Docker to Kubernetes Deployment Mode

This guide provides step-by-step instructions for migrating your Hourglass AVS executor from Docker mode to Kubernetes mode.

## Overview

Migrating from Docker to Kubernetes deployment mode involves:

1. **Infrastructure Setup**: Deploy Kubernetes cluster and Hourglass operator
2. **Configuration Migration**: Convert Docker configuration to Kubernetes format
3. **Deployment Strategy**: Choose between blue-green or rolling migration
4. **Testing & Validation**: Verify functionality in Kubernetes environment
5. **Cutover**: Switch traffic from Docker to Kubernetes deployment

## Prerequisites

### Before You Begin

- ✅ **Working Docker Deployment**: Current Docker-based executor running successfully
- ✅ **Kubernetes Cluster**: Running cluster with admin access (v1.26+)
- ✅ **Hourglass Operator**: Deployed and running in cluster
- ✅ **Network Connectivity**: Kubernetes cluster can reach your aggregator
- ✅ **Container Registry**: Access to push/pull performer images

### Required Tools

```bash
# Verify required tools are installed
kubectl version --client
helm version
docker version
```

## Migration Process

### Step 1: Analyze Current Docker Configuration

First, document your current Docker setup:

```bash
# Export current Docker configuration
docker inspect hourglass-executor > docker-config.json

# Document current resource usage
docker stats --no-stream hourglass-executor

# List current performer containers
docker ps --filter "label=hourglass.performer"

# Export current environment variables
docker inspect hourglass-executor | grep -A 20 "Env"
```

### Step 2: Install Kubernetes Infrastructure

#### Deploy Hourglass Operator

```bash
# Clone repository
git clone https://github.com/Layr-Labs/hourglass-monorepo.git
cd hourglass-monorepo/hourglass-operator

# Deploy operator using Helm
helm install hourglass-operator ./charts/hourglass-operator \
  --namespace hourglass-system \
  --create-namespace

# Verify operator deployment
kubectl get pods -n hourglass-system
kubectl get crd performers.hourglass.eigenlayer.io
```

#### Create Target Namespace

```bash
# Create namespace for your AVS
kubectl create namespace my-avs-project

# Add labels for monitoring/organization
kubectl label namespace my-avs-project \
  hourglass.eigenlayer.io/avs=my-avs \
  environment=production
```

### Step 3: Convert Configuration

#### Extract Docker Configuration

Create a configuration mapping from your Docker setup:

```bash
# Extract key configuration values
echo "Current Docker Configuration:"
echo "Aggregator Endpoint: $(docker inspect hourglass-executor | grep aggregator_endpoint)"
echo "AVS Addresses: $(docker inspect hourglass-executor | grep avs_address)"
echo "Performer Images: $(docker ps --filter label=hourglass.performer --format 'table {{.Image}}')"
```

#### Create Kubernetes Configuration

Convert your Docker configuration to Kubernetes format:

```yaml
# k8s-migration-values.yaml
executor:
  name: "my-avs-executor"
  replicaCount: 1  # Start with 1, scale later
  
  image:
    repository: "hourglass/executor"
    tag: "v1.2.0"  # Match your Docker version
    pullPolicy: "IfNotPresent"

  # Resource limits (based on Docker stats)
  resources:
    requests:
      cpu: "500m"      # Adjust based on Docker usage
      memory: "1Gi"    # Adjust based on Docker usage
    limits:
      cpu: "2"         # Adjust based on Docker usage
      memory: "4Gi"    # Adjust based on Docker usage

# Aggregator configuration (from Docker config)
aggregator:
  endpoint: "aggregator.example.com:9090"  # Replace with your endpoint
  tls:
    enabled: false  # Match your Docker TLS settings

# Blockchain configuration (from Docker config)
chains:
  ethereum:
    enabled: true
    chainId: 1  # Match your Docker chain ID
    rpcUrl: "https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY"
    taskMailboxAddress: "0x1234567890abcdef1234567890abcdef12345678"
    blockConfirmations: 12

# AVS configuration (from Docker config)
avs:
  supportedAvs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "my-avs"
      performer:
        image: "my-org/my-avs-performer"  # Same as Docker
        version: "v1.0.0"  # Same as Docker
        deploymentMode: "kubernetes"  # CHANGED FROM DOCKER
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "2"
          memory: "4Gi"
      env:
        - name: "AVS_CONFIG_PATH"
          value: "/etc/avs/config.json"
        # Add other environment variables from Docker

# Secrets (convert from Docker secrets)
secrets:
  operatorKeys:
    ecdsaPrivateKey: "LS0tLS1CRUdJTi..."  # Base64 encoded
    # blsPrivateKey: "LS0tLS1CRUdJTi..."  # If used

# Storage configuration
persistence:
  enabled: true
  size: "10Gi"  # Adjust based on needs
  storageClass: "gp2"  # Choose appropriate storage class
```

### Step 4: Deploy Kubernetes Environment

#### Deploy to Staging First

```bash
# Create staging namespace
kubectl create namespace my-avs-staging

# Deploy to staging with modified values
helm install my-avs-executor-staging ./charts/hourglass-executor \
  --namespace my-avs-staging \
  --values k8s-migration-values.yaml \
  --set executor.name=my-avs-executor-staging
```

#### Verify Staging Deployment

```bash
# Check staging deployment
kubectl get pods -n my-avs-staging
kubectl get performers -n my-avs-staging
kubectl logs -n my-avs-staging -l app=my-avs-executor-staging

# Test connectivity
kubectl exec -n my-avs-staging deploy/my-avs-executor-staging -- \
  ping aggregator.example.com

# Test performer creation
kubectl describe performer -n my-avs-staging
```

### Step 5: Performance and Functionality Testing

#### Compare Resource Usage

```bash
# Monitor Kubernetes resource usage
kubectl top pods -n my-avs-staging

# Compare with Docker usage
docker stats --no-stream hourglass-executor

# Check if resources need adjustment
kubectl describe pod -n my-avs-staging -l app=my-avs-executor-staging
```

#### Test Task Processing

```bash
# Submit test task to Kubernetes deployment
kubectl port-forward -n my-avs-staging svc/my-avs-executor-staging 9095:9095 &

# Submit test task
grpcurl -plaintext -d '{
  "avsAddress": "0x1234567890abcdef1234567890abcdef12345678",
  "taskId": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
  "payload": "test-payload"
}' localhost:9095 eigenlayer.hourglass.v1.ExecutorService/SubmitTask

# Monitor task processing
kubectl logs -n my-avs-staging -l app=my-avs-executor-staging -f
```

#### Performance Comparison

```bash
# Create performance comparison script
cat > performance-test.sh << 'EOF'
#!/bin/bash
echo "=== Performance Comparison ==="

echo "Docker Stats:"
docker stats --no-stream hourglass-executor | grep -v CONTAINER

echo "Kubernetes Stats:"
kubectl top pod -n my-avs-staging -l app=my-avs-executor-staging

echo "Docker Container Count:"
docker ps --filter "label=hourglass.performer" | wc -l

echo "Kubernetes Performer Count:"
kubectl get performers -n my-avs-staging --no-headers | wc -l
EOF

chmod +x performance-test.sh
./performance-test.sh
```

### Step 6: Production Deployment

#### Update Production Configuration

```yaml
# k8s-production-values.yaml
executor:
  name: "my-avs-executor"
  replicaCount: 2  # Increase for HA
  
  image:
    repository: "hourglass/executor"
    tag: "v1.2.0"
    pullPolicy: "IfNotPresent"

  resources:
    requests:
      cpu: "1000m"     # Increase for production
      memory: "2Gi"    # Increase for production
    limits:
      cpu: "4000m"     # Increase for production
      memory: "8Gi"    # Increase for production

  # Production security settings
  securityContext:
    runAsNonRoot: true
    runAsUser: 65532
    fsGroup: 65532

  # Production node selection
  nodeSelector:
    node-type: "high-performance"

  # Production tolerations
  tolerations:
    - key: "avs-workload"
      operator: "Equal"
      value: "production"
      effect: "NoSchedule"

# Production monitoring
metrics:
  enabled: true
  port: 8080
  serviceMonitor:
    enabled: true
    namespace: "monitoring"
    interval: "30s"
    labels:
      team: "avs-team"
      environment: "production"

# Production storage
persistence:
  enabled: true
  size: "100Gi"
  storageClass: "fast-ssd"
  accessMode: "ReadWriteOnce"
```

#### Deploy Production Environment

```bash
# Deploy to production namespace
helm install my-avs-executor ./charts/hourglass-executor \
  --namespace my-avs-project \
  --values k8s-production-values.yaml

# Verify deployment
kubectl get pods -n my-avs-project
kubectl get performers -n my-avs-project
```

### Step 7: Migration Strategies

#### Option A: Blue-Green Migration (Recommended)

```bash
# 1. Keep Docker running (blue)
# 2. Deploy Kubernetes alongside (green)
# 3. Test Kubernetes thoroughly
# 4. Switch traffic to Kubernetes
# 5. Monitor for issues
# 6. Shutdown Docker after validation

# Monitor both systems
watch 'echo "=== Docker ===" && docker ps --filter label=hourglass && echo "=== Kubernetes ===" && kubectl get pods -n my-avs-project'
```

#### Option B: Rolling Migration

```bash
# 1. Reduce Docker replica count
docker update --cpus="0.5" hourglass-executor

# 2. Deploy Kubernetes with low resources
helm install my-avs-executor ./charts/hourglass-executor \
  --namespace my-avs-project \
  --values k8s-production-values.yaml \
  --set executor.resources.requests.cpu=500m \
  --set executor.resources.requests.memory=1Gi

# 3. Gradually increase Kubernetes resources
kubectl patch deployment my-avs-executor -n my-avs-project -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [
          {
            "name": "executor",
            "resources": {
              "requests": {
                "cpu": "1000m",
                "memory": "2Gi"
              }
            }
          }
        ]
      }
    }
  }
}'

# 4. Scale down Docker
docker stop hourglass-executor
```

### Step 8: Validation and Monitoring

#### Health Checks

```bash
# Create validation script
cat > migration-validation.sh << 'EOF'
#!/bin/bash
echo "=== Migration Validation ==="

# Check Kubernetes deployment
echo "Kubernetes Pods:"
kubectl get pods -n my-avs-project -l app=my-avs-executor

# Check performer CRDs
echo "Performer CRDs:"
kubectl get performers -n my-avs-project

# Check connectivity
echo "Aggregator Connectivity:"
kubectl exec -n my-avs-project deploy/my-avs-executor -- \
  nc -zv aggregator.example.com 9090

# Check performer pods
echo "Performer Pods:"
kubectl get pods -n my-avs-project -l app=hourglass-performer

# Check logs for errors
echo "Recent Errors:"
kubectl logs -n my-avs-project -l app=my-avs-executor --tail=50 | grep -i error
EOF

chmod +x migration-validation.sh
./migration-validation.sh
```

#### Performance Monitoring

```bash
# Set up monitoring dashboard
kubectl apply -f - << 'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: migration-monitoring
  namespace: my-avs-project
data:
  check-performance.sh: |
    #!/bin/bash
    echo "=== Performance Monitoring ==="
    
    # Resource usage
    echo "Resource Usage:"
    kubectl top pods -n my-avs-project
    
    # Task processing metrics
    echo "Task Metrics:"
    kubectl exec -n my-avs-project deploy/my-avs-executor -- \
      curl -s http://localhost:8080/metrics | grep -E "tasks_processed|tasks_failed"
    
    # Performer health
    echo "Performer Health:"
    kubectl get performers -n my-avs-project -o wide
EOF

# Run monitoring
kubectl exec -n my-avs-project deploy/my-avs-executor -- \
  /bin/bash -c "$(kubectl get configmap migration-monitoring -n my-avs-project -o jsonpath='{.data.check-performance\.sh}')"
```

### Step 9: Cleanup Docker Environment

#### After Successful Migration

```bash
# Stop Docker containers
docker stop hourglass-executor
docker stop $(docker ps -q --filter "label=hourglass.performer")

# Remove Docker containers
docker rm hourglass-executor
docker rm $(docker ps -aq --filter "label=hourglass.performer")

# Remove Docker images (optional)
docker rmi $(docker images --filter "label=hourglass" -q)

# Remove Docker networks (optional)
docker network rm $(docker network ls --filter "label=hourglass" -q)

# Clean up Docker volumes (optional)
docker volume rm $(docker volume ls --filter "label=hourglass" -q)
```

## Post-Migration Checklist

### ✅ Functional Verification

- [ ] All performer pods are running and healthy
- [ ] Executor can connect to aggregator
- [ ] Task processing is working correctly
- [ ] Performance metrics are within expected ranges
- [ ] No error messages in logs
- [ ] Service discovery is working
- [ ] Resource usage is optimal

### ✅ Operational Verification

- [ ] Monitoring and alerting configured
- [ ] Backup and disaster recovery tested
- [ ] Scaling policies configured
- [ ] Security policies applied
- [ ] Documentation updated
- [ ] Team trained on Kubernetes operations

### ✅ Performance Verification

- [ ] Task processing latency acceptable
- [ ] Resource utilization efficient
- [ ] Auto-scaling working (if enabled)
- [ ] Load balancing effective
- [ ] Network performance adequate

## Troubleshooting Common Migration Issues

### Issue: Performer Pods Not Starting

**Symptoms:**
- Performer CRDs created but no pods appear
- Operator logs show processing but no results

**Solutions:**
```bash
# Check operator logs
kubectl logs -n hourglass-system -l app=hourglass-operator

# Check RBAC permissions
kubectl auth can-i create pods --as=system:serviceaccount:hourglass-system:hourglass-operator -n my-avs-project

# Check resource quotas
kubectl describe resourcequota -n my-avs-project
```

### Issue: Configuration Validation Errors

**Symptoms:**
- "mixed deployment modes not supported" error
- Executor won't start with new configuration

**Solutions:**
```bash
# Validate configuration
kubectl apply --dry-run=client -f k8s-config.yaml

# Check for mixed deployment modes
grep -n "deployment_mode" k8s-config.yaml
```

### Issue: Performance Degradation

**Symptoms:**
- Tasks taking longer to process
- Higher resource usage than Docker

**Solutions:**
```bash
# Increase resources
kubectl patch deployment my-avs-executor -n my-avs-project -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [
          {
            "name": "executor",
            "resources": {
              "requests": {
                "cpu": "2000m",
                "memory": "4Gi"
              },
              "limits": {
                "cpu": "8000m",
                "memory": "16Gi"
              }
            }
          }
        ]
      }
    }
  }
}'

# Enable horizontal pod autoscaling
kubectl apply -f - << 'EOF'
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-avs-executor-hpa
  namespace: my-avs-project
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-avs-executor
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
EOF
```

## Rollback Procedure

If issues occur, you can rollback to Docker:

```bash
# 1. Stop Kubernetes deployment
kubectl scale deployment my-avs-executor -n my-avs-project --replicas=0

# 2. Start Docker containers
docker start hourglass-executor

# 3. Verify Docker functionality
docker logs hourglass-executor
docker ps --filter "label=hourglass.performer"

# 4. Update DNS/load balancer to point to Docker
# (depends on your infrastructure)
```

## Best Practices for Migration

### Planning
- **Test in staging first** - Never migrate production directly
- **Plan for rollback** - Have a rollback plan ready
- **Monitor closely** - Watch metrics during and after migration
- **Migrate gradually** - Use blue-green or rolling deployment

### Configuration
- **Keep secrets secure** - Use Kubernetes secrets for sensitive data
- **Set resource limits** - Prevent resource exhaustion
- **Use health checks** - Configure liveness and readiness probes
- **Enable monitoring** - Set up metrics and alerting

### Operations
- **Document changes** - Update operational procedures
- **Train team** - Ensure team knows Kubernetes operations
- **Test disaster recovery** - Verify backup and restore procedures
- **Monitor performance** - Watch for performance regressions

## Next Steps

After successful migration:

1. **Optimize Performance**: Fine-tune resource allocation and scaling policies
2. **Implement GitOps**: Set up automated deployments with ArgoCD or Flux
3. **Enhance Security**: Implement NetworkPolicies and PodSecurityPolicies
4. **Set Up Monitoring**: Configure comprehensive monitoring with Prometheus/Grafana
5. **Plan Multi-Region**: Consider multi-region deployment for disaster recovery

For additional help, see:
- [Kubernetes Configuration Examples](./examples/kubernetes-config-examples.md)
- [Troubleshooting Guide](./troubleshooting.md)
- [Production Best Practices](./production-best-practices.md)