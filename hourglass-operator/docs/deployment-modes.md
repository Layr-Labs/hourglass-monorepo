# Hourglass Deployment Modes

This guide explains the two deployment modes available for Hourglass AVS performers and how to choose between them.

## Overview

Hourglass supports two deployment modes for AVS performers:

1. **Docker Mode** - Traditional container-based deployment (default)
2. **Kubernetes Mode** - Cloud-native deployment using Kubernetes CRDs

> **Important**: Each executor instance must use a single deployment mode. Mixed modes within the same executor are not supported.

## Docker Mode (Default)

### When to Use Docker Mode

- **Local Development**: Testing and development environments
- **Simple Deployments**: Single-node or basic multi-node setups
- **Existing Docker Infrastructure**: Teams already using Docker Compose or similar
- **Quick Prototyping**: Fast iteration and testing

### How Docker Mode Works

```
┌─────────────────────────────────────────────────────────────────┐
│                          Docker Host                            │
│                                                                 │
│  ┌─────────────────┐              ┌─────────────────────────┐    │
│  │   Executor      │              │    AVS Performers       │    │
│  │   Container     │              │                         │    │
│  │                 │──────────────┤  ┌─────────────────────┐ │    │
│  │ - Config        │ Docker API   │  │ Performer Container │ │    │
│  │ - Lifecycle     │              │  │ (your AVS logic)    │ │    │
│  │ - Monitoring    │              │  └─────────────────────┘ │    │
│  └─────────────────┘              └─────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

### Docker Mode Configuration

```yaml
# executor-config.yaml
aggregator_endpoint: "aggregator.example.com:9090"
deployment_mode: "docker"  # or omit (defaults to docker)
log_level: "info"

# No kubernetes section needed
avs_config:
  supported_avs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "my-avs"
      performer_image: "my-org/my-avs-performer"
      performer_version: "v1.0.0"
      default_resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "2"
          memory: "4Gi"
```

### Docker Mode Benefits

- ✅ **Simple Setup**: No Kubernetes cluster required
- ✅ **Fast Development**: Quick container builds and deployments
- ✅ **Resource Efficient**: Lower overhead for small deployments
- ✅ **Easy Debugging**: Direct container access and logs

### Docker Mode Limitations

- ❌ **Manual Scaling**: No automatic scaling capabilities
- ❌ **Basic Health Checks**: Limited failure recovery
- ❌ **Single Point of Failure**: Host failures affect all performers
- ❌ **Manual Updates**: No rolling updates or blue-green deployments

## Kubernetes Mode

### When to Use Kubernetes Mode

- **Production Deployments**: High availability and scaling requirements
- **Cloud Infrastructure**: Running on managed Kubernetes services
- **Advanced Features**: Need for rolling updates, auto-scaling, health monitoring
- **Multi-Tenant Environments**: Namespace isolation and resource quotas

### How Kubernetes Mode Works

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

### Kubernetes Mode Configuration

```yaml
# executor-config.yaml
aggregator_endpoint: "aggregator.example.com:9090"
deployment_mode: "kubernetes"  # REQUIRED for k8s mode
log_level: "info"

performer_config:
  service_pattern: "performer-{name}.{namespace}.svc.cluster.local:{port}"
  default_port: 9090
  connection_timeout: "30s"
  startup_timeout: "300s"
  retry_attempts: 3
  max_performers: 5

# REQUIRED for kubernetes mode
kubernetes:
  namespace: "my-avs-project"
  operator_namespace: "hourglass-system"
  performer_crd_group: "hourglass.eigenlayer.io"
  performer_crd_version: "v1alpha1"
  performer_crd_kind: "Performer"
  cleanup_on_shutdown: true
  cleanup_timeout: "60s"

avs_config:
  supported_avs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "my-avs"
      performer_image: "my-org/my-avs-performer"
      performer_version: "v1.0.0"
      default_resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "2"
          memory: "4Gi"
```

### Kubernetes Mode Benefits

- ✅ **High Availability**: Automatic failover and recovery
- ✅ **Auto-Scaling**: Horizontal Pod Autoscaler support
- ✅ **Rolling Updates**: Zero-downtime deployments
- ✅ **Blue-Green Deployments**: Safe production updates
- ✅ **Resource Management**: Requests, limits, and quotas
- ✅ **Service Discovery**: Built-in DNS and service mesh integration
- ✅ **Monitoring**: Prometheus metrics and health checks
- ✅ **Namespace Isolation**: Multi-tenant security

### Kubernetes Mode Requirements

- ❌ **Kubernetes Cluster**: Requires running cluster
- ❌ **Hourglass Operator**: Must deploy singleton operator
- ❌ **Additional Complexity**: More moving parts to manage
- ❌ **Learning Curve**: Kubernetes knowledge required

## Choosing Between Deployment Modes

### Use Docker Mode When:

- **Development Environment**: Local testing and development
- **Simple Production**: Single-node or basic multi-node deployments
- **Resource Constraints**: Limited infrastructure or budget
- **Fast Prototyping**: Quick iterations and testing
- **Existing Docker Infrastructure**: Already using Docker ecosystem

### Use Kubernetes Mode When:

- **Production at Scale**: High availability requirements
- **Cloud Infrastructure**: Running on managed Kubernetes services
- **Advanced Features**: Need auto-scaling, rolling updates, monitoring
- **Multi-Tenant**: Multiple teams or environments
- **CI/CD Integration**: Automated deployment pipelines

## Configuration Validation

The executor configuration includes validation to prevent common mistakes:

### ✅ Valid Configurations

```yaml
# All Docker mode performers
avs_config:
  supported_avs:
    - address: "0x123..."
      deployment_mode: "docker"
    - address: "0x456..."
      deployment_mode: "docker"  # Same mode ✅
```

```yaml
# All Kubernetes mode performers
deployment_mode: "kubernetes"
kubernetes:
  namespace: "my-avs"
  # ... other kubernetes config
avs_config:
  supported_avs:
    - address: "0x123..."
      deployment_mode: "kubernetes"
    - address: "0x456..."
      deployment_mode: "kubernetes"  # Same mode ✅
```

### ❌ Invalid Configurations

```yaml
# Mixed deployment modes (NOT SUPPORTED)
avs_config:
  supported_avs:
    - address: "0x123..."
      deployment_mode: "docker"
    - address: "0x456..."
      deployment_mode: "kubernetes"  # Different mode ❌
```

**Error message**: `"mixed deployment modes not supported: all performers must use the same deployment mode (either 'docker' or 'kubernetes')"`

## Migration Path

### Docker → Kubernetes Migration

1. **Deploy Kubernetes Infrastructure**
   ```bash
   # Install Hourglass Operator
   helm install hourglass-operator ./charts/hourglass-operator \
     --namespace hourglass-system \
     --create-namespace
   ```

2. **Update Configuration**
   ```yaml
   # Change deployment mode
   deployment_mode: "kubernetes"
   
   # Add kubernetes section
   kubernetes:
     namespace: "my-avs-project"
     operator_namespace: "hourglass-system"
     # ... other settings
   ```

3. **Test in Staging**
   ```bash
   # Deploy to staging namespace first
   kubectl apply -f kubernetes-config.yaml -n staging
   ```

4. **Rolling Migration**
   ```bash
   # Blue-green deployment
   kubectl apply -f kubernetes-config.yaml -n production
   # Validate then switch traffic
   ```

### Kubernetes → Docker Migration

1. **Prepare Docker Environment**
   ```bash
   # Set up Docker host
   docker-compose up -d
   ```

2. **Update Configuration**
   ```yaml
   # Change deployment mode
   deployment_mode: "docker"
   
   # Remove kubernetes section
   # kubernetes: ...  # Remove this
   ```

3. **Deploy to Docker**
   ```bash
   # Start executor in docker mode
   ./executor run --config docker-config.yaml
   ```

## Best Practices

### Docker Mode Best Practices

- **Use docker-compose** for multi-container setups
- **Volume mounts** for persistent data
- **Health checks** in Dockerfile
- **Resource limits** in docker-compose.yml
- **Log aggregation** for monitoring

### Kubernetes Mode Best Practices

- **Resource requests/limits** on all containers
- **Liveness/readiness probes** for health checks
- **Namespace isolation** for security
- **NetworkPolicies** for network security
- **PodSecurityPolicies** for runtime security
- **Monitoring and alerting** with Prometheus

## Troubleshooting

### Docker Mode Issues

**Problem**: Performer container won't start
```bash
# Check container logs
docker logs <container-id>

# Check resource usage
docker stats

# Check network connectivity
docker exec -it <container-id> ping aggregator.example.com
```

**Problem**: Task execution failures
```bash
# Check executor logs
docker logs hourglass-executor

# Check performer health
curl http://localhost:9090/health
```

### Kubernetes Mode Issues

**Problem**: Performer pods not created
```bash
# Check operator logs
kubectl logs -n hourglass-system deployment/hourglass-operator

# Check performer CRDs
kubectl get performers -n my-avs-project -o yaml

# Check events
kubectl get events -n my-avs-project --sort-by=.lastTimestamp
```

**Problem**: Service discovery failures
```bash
# Test DNS resolution
kubectl run test-dns --image=busybox:1.35 --rm -i --restart=Never -- \
  nslookup performer-my-performer.my-avs-project.svc.cluster.local

# Check service endpoints
kubectl get endpoints -n my-avs-project
```

## Next Steps

- **Docker Mode**: See [Docker Deployment Guide](./docker-deployment.md)
- **Kubernetes Mode**: See [Kubernetes Deployment Guide](./kubernetes-deployment.md)
- **Configuration Reference**: See [Configuration API Reference](./api-reference.md)
- **Troubleshooting**: See [Troubleshooting Guide](./troubleshooting.md)