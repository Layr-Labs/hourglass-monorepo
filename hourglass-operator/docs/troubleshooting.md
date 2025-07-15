# Hourglass Troubleshooting Guide

This guide provides solutions for common issues encountered when deploying and running Hourglass in both Docker and Kubernetes modes.

## Table of Contents

1. [General Troubleshooting](#general-troubleshooting)
2. [Docker Mode Issues](#docker-mode-issues)
3. [Kubernetes Mode Issues](#kubernetes-mode-issues)
4. [Configuration Issues](#configuration-issues)
5. [Performance Issues](#performance-issues)
6. [Monitoring and Debugging](#monitoring-and-debugging)

## General Troubleshooting

### Common Diagnostic Commands

```bash
# Check executor logs
kubectl logs -n my-avs-project deployment/my-avs-executor

# Check operator logs
kubectl logs -n hourglass-system deployment/hourglass-operator

# Check performer CRDs
kubectl get performers -n my-avs-project -o yaml

# Check recent events
kubectl get events -n my-avs-project --sort-by=.lastTimestamp
```

### Configuration Validation

```bash
# Validate executor configuration
hourglass-executor validate --config config.yaml

# Check deployment mode consistency
grep -n "deployment_mode" config.yaml
```

## Docker Mode Issues

### Issue: Executor Container Won't Start

**Symptoms:**
- Container exits immediately
- "Config file not found" errors
- Permission denied errors

**Diagnosis:**
```bash
# Check container logs
docker logs hourglass-executor

# Check file permissions
ls -la /path/to/config.yaml

# Verify container user
docker exec -it hourglass-executor whoami
```

**Solutions:**
```bash
# Fix file permissions
chmod 644 config.yaml
chown 1000:1000 config.yaml

# Mount config correctly
docker run -v $(pwd)/config.yaml:/etc/hourglass/config.yaml hourglass/executor

# Use correct user
docker run --user 1000:1000 hourglass/executor
```

### Issue: Performer Container Connection Failures

**Symptoms:**
- "Connection refused" errors
- Timeout errors when connecting to performers
- Performer containers start but aren't accessible

**Diagnosis:**
```bash
# Check network connectivity
docker exec -it hourglass-executor ping performer-container

# Check port bindings
docker ps -a | grep performer
netstat -tlnp | grep 9090

# Test gRPC connection
grpcurl -plaintext localhost:9090 list
```

**Solutions:**
```bash
# Fix network configuration
docker network create hourglass-network
docker run --network hourglass-network hourglass/executor

# Bind ports correctly
docker run -p 9090:9090 my-org/my-avs-performer

# Use correct service URLs
# Change: localhost:9090
# To: performer-container:9090
```

### Issue: Docker Resource Constraints

**Symptoms:**
- Out of memory errors
- CPU throttling
- Slow performance

**Diagnosis:**
```bash
# Check resource usage
docker stats
docker system df

# Check container limits
docker inspect hourglass-executor | grep -i memory
```

**Solutions:**
```bash
# Increase memory limits
docker run --memory=4g --memory-swap=8g hourglass/executor

# Increase CPU limits
docker run --cpus="2.0" hourglass/executor

# Use docker-compose for complex setups
version: '3.8'
services:
  executor:
    image: hourglass/executor
    deploy:
      resources:
        limits:
          memory: 4G
          cpus: '2.0'
```

## Kubernetes Mode Issues

### Issue: Operator Not Creating Performers

**Symptoms:**
- Performer CRDs created but no pods appear
- Operator logs show processing but no results
- "Failed to create performer" errors

**Diagnosis:**
```bash
# Check operator logs
kubectl logs -n hourglass-system deployment/hourglass-operator

# Check performer CRD status
kubectl get performers -n my-avs-project -o yaml

# Check operator RBAC permissions
kubectl auth can-i create pods --as=system:serviceaccount:hourglass-system:hourglass-operator -n my-avs-project
```

**Solutions:**
```bash
# Fix RBAC permissions
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: hourglass-operator
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: ["hourglass.eigenlayer.io"]
  resources: ["performers"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
EOF

# Restart operator
kubectl rollout restart deployment/hourglass-operator -n hourglass-system

# Check resource quotas
kubectl describe resourcequota -n my-avs-project
```

### Issue: Performer Pods in Pending State

**Symptoms:**
- Pods stuck in `Pending` status
- "Insufficient resources" events
- Node selector/affinity issues

**Diagnosis:**
```bash
# Check pod events
kubectl describe pod performer-pod-name -n my-avs-project

# Check node resources
kubectl top nodes
kubectl describe nodes

# Check resource requests
kubectl get pods -n my-avs-project -o yaml | grep -A 5 resources
```

**Solutions:**
```bash
# Scale up cluster or reduce resource requests
kubectl patch deployment my-avs-executor -n my-avs-project -p '{"spec":{"template":{"spec":{"containers":[{"name":"executor","resources":{"requests":{"cpu":"250m","memory":"512Mi"}}}]}}}}'

# Remove node selector if too restrictive
kubectl patch deployment my-avs-executor -n my-avs-project -p '{"spec":{"template":{"spec":{"nodeSelector":null}}}}'

# Check and fix taints/tolerations
kubectl taint nodes node1 special-workload=true:NoSchedule
```

### Issue: Service Discovery Problems

**Symptoms:**
- "Name resolution failed" errors
- Cannot connect to performer services
- DNS lookup failures

**Diagnosis:**
```bash
# Test DNS resolution
kubectl run test-dns --image=busybox:1.35 --rm -i --restart=Never -- nslookup performer-my-performer.my-avs-project.svc.cluster.local

# Check service endpoints
kubectl get endpoints -n my-avs-project
kubectl get services -n my-avs-project

# Check CoreDNS logs
kubectl logs -n kube-system deployment/coredns
```

**Solutions:**
```bash
# Fix service naming
kubectl patch service incorrect-service-name -n my-avs-project -p '{"metadata":{"name":"performer-my-performer"}}'

# Check service selector
kubectl patch service performer-my-performer -n my-avs-project -p '{"spec":{"selector":{"app":"my-performer"}}}'

# Restart CoreDNS if needed
kubectl rollout restart deployment/coredns -n kube-system
```

### Issue: Performer Health Check Failures

**Symptoms:**
- Pods failing readiness/liveness probes
- Frequent pod restarts
- "Unhealthy" status in performer CRDs

**Diagnosis:**
```bash
# Check pod health
kubectl get pods -n my-avs-project -o wide
kubectl describe pod performer-pod-name -n my-avs-project

# Test health endpoints manually
kubectl exec -it performer-pod-name -n my-avs-project -- wget -qO- http://localhost:8090/health
```

**Solutions:**
```bash
# Adjust health check timing
kubectl patch deployment my-avs-executor -n my-avs-project -p '{"spec":{"template":{"spec":{"containers":[{"name":"executor","readinessProbe":{"initialDelaySeconds":30,"periodSeconds":10}}]}}}}'

# Fix health endpoint
kubectl patch deployment my-avs-executor -n my-avs-project -p '{"spec":{"template":{"spec":{"containers":[{"name":"executor","readinessProbe":{"httpGet":{"path":"/health","port":8090}}}]}}}}'
```

### Issue: Cross-Namespace Permission Errors

**Symptoms:**
- Executor cannot create performers in its namespace
- "Forbidden" errors in operator logs
- RBAC permission denied

**Diagnosis:**
```bash
# Check executor service account permissions
kubectl auth can-i create performers --as=system:serviceaccount:my-avs-project:executor-service-account -n my-avs-project

# Check operator permissions
kubectl auth can-i create pods --as=system:serviceaccount:hourglass-system:hourglass-operator -n my-avs-project
```

**Solutions:**
```bash
# Create proper RBAC for executor
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: my-avs-project
  name: executor-role
rules:
- apiGroups: ["hourglass.eigenlayer.io"]
  resources: ["performers"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
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
EOF

# Grant operator cluster-wide permissions
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: hourglass-operator-binding
subjects:
- kind: ServiceAccount
  name: hourglass-operator
  namespace: hourglass-system
roleRef:
  kind: ClusterRole
  name: hourglass-operator
  apiGroup: rbac.authorization.k8s.io
EOF
```

## Configuration Issues

### Issue: Mixed Deployment Mode Error

**Symptoms:**
- "mixed deployment modes not supported" error
- Configuration validation failures
- Executor won't start

**Diagnosis:**
```bash
# Check configuration
grep -n "deployment_mode" config.yaml
```

**Solutions:**
```yaml
# Fix configuration - ensure all performers use same mode
avs_config:
  supported_avs:
    - address: "0x123..."
      deployment_mode: "kubernetes"  # All should be same
    - address: "0x456..."
      deployment_mode: "kubernetes"  # All should be same
```

### Issue: Missing Kubernetes Configuration

**Symptoms:**
- "kubernetes configuration is required" error
- Validation failures for Kubernetes mode
- Cannot connect to Kubernetes API

**Diagnosis:**
```bash
# Check if kubernetes section exists
grep -A 10 "kubernetes:" config.yaml

# Verify cluster connectivity
kubectl cluster-info
```

**Solutions:**
```yaml
# Add required kubernetes section
deployment_mode: "kubernetes"
kubernetes:
  namespace: "my-avs-project"
  operator_namespace: "hourglass-system"
  performer_crd_group: "hourglass.eigenlayer.io"
  performer_crd_version: "v1alpha1"
  performer_crd_kind: "Performer"
  cleanup_on_shutdown: true
  cleanup_timeout: "60s"
  connection_timeout: "30s"
  in_cluster: true
```

### Issue: Invalid Resource Specifications

**Symptoms:**
- Pods fail to schedule
- "Invalid resource format" errors
- Resource quota exceeded

**Diagnosis:**
```bash
# Check resource specifications
kubectl get pods -n my-avs-project -o yaml | grep -A 5 resources
```

**Solutions:**
```yaml
# Fix resource format
default_resources:
  requests:
    cpu: "500m"      # Correct format
    memory: "1Gi"    # Correct format
  limits:
    cpu: "2"         # Correct format (2 cores)
    memory: "4Gi"    # Correct format
```

## Performance Issues

### Issue: High Resource Usage

**Symptoms:**
- High CPU/memory usage
- Slow task processing
- Frequent OOM kills

**Diagnosis:**
```bash
# Check resource usage
kubectl top pods -n my-avs-project
kubectl top nodes

# Check metrics
kubectl get --raw /metrics | grep hourglass
```

**Solutions:**
```bash
# Increase resource limits
kubectl patch deployment my-avs-executor -n my-avs-project -p '{"spec":{"template":{"spec":{"containers":[{"name":"executor","resources":{"limits":{"cpu":"4","memory":"8Gi"}}}]}}}}'

# Enable auto-scaling
kubectl apply -f - <<EOF
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

### Issue: Slow Task Processing

**Symptoms:**
- Tasks taking too long to complete
- Queue backlog growing
- Timeout errors

**Diagnosis:**
```bash
# Check task metrics
kubectl exec -it my-avs-executor -- curl http://localhost:8080/metrics | grep task

# Check network latency
kubectl exec -it my-avs-executor -- ping aggregator.example.com
```

**Solutions:**
```bash
# Increase performer replicas
kubectl scale deployment my-avs-executor -n my-avs-project --replicas=5

# Optimize network configuration
kubectl patch deployment my-avs-executor -n my-avs-project -p '{"spec":{"template":{"spec":{"containers":[{"name":"executor","env":[{"name":"NETWORK_TIMEOUT","value":"60s"}]}]}}}}'
```

## Monitoring and Debugging

### Debug Mode Setup

```bash
# Enable debug logging
kubectl patch deployment my-avs-executor -n my-avs-project -p '{"spec":{"template":{"spec":{"containers":[{"name":"executor","env":[{"name":"LOG_LEVEL","value":"debug"}]}]}}}}'

# Port forward for debugging
kubectl port-forward deployment/my-avs-executor 8080:8080 -n my-avs-project

# Access metrics
curl http://localhost:8080/metrics
```

### Comprehensive Health Check

```bash
#!/bin/bash
# health-check.sh

echo "=== Hourglass Health Check ==="

# Check operator
echo "Checking operator..."
kubectl get pods -n hourglass-system -l app=hourglass-operator

# Check executors
echo "Checking executors..."
kubectl get pods -n my-avs-project -l app=hourglass-executor

# Check performers
echo "Checking performers..."
kubectl get performers -n my-avs-project

# Check services
echo "Checking services..."
kubectl get services -n my-avs-project

# Check recent events
echo "Recent events..."
kubectl get events -n my-avs-project --sort-by=.lastTimestamp | tail -10

# Check resource usage
echo "Resource usage..."
kubectl top pods -n my-avs-project
```

### Log Analysis

```bash
# Search for specific errors
kubectl logs -n my-avs-project deployment/my-avs-executor | grep -i error

# Follow logs in real time
kubectl logs -n my-avs-project deployment/my-avs-executor -f

# Get logs from all containers
kubectl logs -n my-avs-project deployment/my-avs-executor --all-containers=true

# Export logs for analysis
kubectl logs -n my-avs-project deployment/my-avs-executor --since=1h > executor-logs.txt
```

### Performance Profiling

```bash
# Enable profiling
kubectl patch deployment my-avs-executor -n my-avs-project -p '{"spec":{"template":{"spec":{"containers":[{"name":"executor","env":[{"name":"ENABLE_PROFILING","value":"true"}]}]}}}}'

# Access profiling endpoints
kubectl port-forward deployment/my-avs-executor 6060:6060 -n my-avs-project
go tool pprof http://localhost:6060/debug/pprof/profile
```

## Common Error Messages and Solutions

### "Connection refused"
- **Cause**: Service not available or incorrect URL
- **Solution**: Check service status and DNS resolution

### "Insufficient resources"
- **Cause**: Not enough CPU/memory in cluster
- **Solution**: Scale cluster or reduce resource requests

### "ImagePullBackOff"
- **Cause**: Cannot pull container image
- **Solution**: Check image name, registry access, and pull secrets

### "CrashLoopBackOff"
- **Cause**: Application crashes on startup
- **Solution**: Check application logs and configuration

### "Forbidden"
- **Cause**: RBAC permission denied
- **Solution**: Check and fix service account permissions

### "FailedMount"
- **Cause**: Cannot mount volume
- **Solution**: Check persistent volume claims and storage class

## Getting Help

### Information to Gather

When reporting issues, include:

1. **Environment Information**
   ```bash
   kubectl version
   kubectl cluster-info
   kubectl get nodes
   ```

2. **Hourglass Configuration**
   ```bash
   # Sanitized config (remove secrets)
   cat config.yaml | sed 's/password: .*/password: ***/'
   ```

3. **Logs**
   ```bash
   kubectl logs -n hourglass-system deployment/hourglass-operator
   kubectl logs -n my-avs-project deployment/my-avs-executor
   ```

4. **Resource Status**
   ```bash
   kubectl get all -n my-avs-project
   kubectl get performers -n my-avs-project -o yaml
   ```

### Support Channels

- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Check latest docs for updates
- **Community**: Join community discussions

### Advanced Debugging

For complex issues, enable verbose logging and collect comprehensive diagnostics:

```bash
# Enable verbose logging
kubectl patch deployment my-avs-executor -n my-avs-project -p '{"spec":{"template":{"spec":{"containers":[{"name":"executor","env":[{"name":"LOG_LEVEL","value":"trace"}]}]}}}}'

# Collect diagnostic bundle
kubectl cluster-info dump --output-directory=./hourglass-diagnostics
```