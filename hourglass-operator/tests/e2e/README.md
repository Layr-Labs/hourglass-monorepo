# Hourglass Operator E2E Testing (Milestone 4.2)

This directory contains end-to-end tests for the Hourglass Kubernetes Operator to validate milestone 4.2 requirements.

## Test Overview

The E2E tests validate the complete operator workflow:

1. **Operator Deployment** - Deploy the singleton operator in a test cluster
2. **Executor Creation** - Create a test executor that uses Kubernetes deployment mode
3. **Performer CRD Creation** - Test executor creates Performer CRDs
4. **Pod & Service Creation** - Operator creates performer pods and services
5. **Service Discovery** - Validate DNS resolution and gRPC connectivity
6. **Health Monitoring** - Test performer health monitoring and status updates
7. **Multi-Performer Scenarios** - Test multiple performers per AVS
8. **Cross-Namespace Isolation** - Test performer isolation across namespaces

## Prerequisites

- Kubernetes cluster (v1.26+)
- kubectl configured with cluster admin permissions
- Helm v3.x installed
- Docker images available or buildable locally

## Quick Start

### 1. Setup Test Environment

```bash
# Run complete setup
./test-setup.sh setup

# Check status
./test-setup.sh status

# Run validation tests
./test-setup.sh test

# Cleanup
./test-setup.sh cleanup
```

### 2. Manual Testing

If you prefer to run tests manually:

```bash
# Deploy operator
helm install hourglass-operator ../../charts/hourglass-operator \
    --namespace hourglass-system \
    --create-namespace

# Create test namespace
kubectl create namespace test-avs-project

# Apply test configurations
kubectl apply -f test-configs/
```

## Test Scenarios

### 4.2.1 Operator Integration Tests

**Test: Deploy operator in test cluster**
- ✅ Operator pod starts successfully
- ✅ CRDs are installed
- ✅ RBAC permissions are configured
- ✅ Health checks pass

**Test: Create Performer CRDs via executor**
- ✅ Test executor can create Performer CRDs
- ✅ CRDs are accepted by API server
- ✅ Operator watches and processes CRDs

**Test: Verify Pod and Service creation by operator**
- ✅ Operator creates performer pods from CRDs
- ✅ Pods use specified image and resources
- ✅ Services are created with correct selectors
- ✅ Service DNS names follow expected pattern

**Test: Performer health monitoring and status updates**
- ✅ Operator monitors pod health
- ✅ Status updates are reflected in CRD status
- ✅ Health check failures are detected
- ✅ Status events are generated

**Test: Service DNS resolution and gRPC connectivity**
- ✅ Service DNS resolution works: `performer-{name}.{namespace}.svc.cluster.local`
- ✅ gRPC connectivity can be established
- ✅ Connection retry logic works
- ✅ Circuit breaker functions properly

### 4.2.2 Multi-Performer Scenarios

**Test: Multiple performers per AVS**
- ✅ Multiple performers can be created for same AVS
- ✅ Each performer gets unique pod and service
- ✅ Blue-green deployments work correctly
- ✅ Load balancing across performers

**Test: Cross-namespace performer isolation**
- ✅ Performers in different namespaces are isolated
- ✅ Operator manages performers across namespaces
- ✅ RBAC restrictions are enforced
- ✅ Resource quotas are respected

**Test: Concurrent deployment and removal operations**
- ✅ Concurrent performer creation works
- ✅ Concurrent performer deletion works
- ✅ No race conditions in operator
- ✅ Proper cleanup on termination

**Test: Performance testing with scale**
- ✅ Operator handles 10+ performers per namespace
- ✅ Response times remain acceptable
- ✅ Memory usage stays within limits
- ✅ CPU usage remains reasonable

## Test Configuration Files

### test-configs/operator-values.yaml
```yaml
# Operator configuration for testing
operator:
  image:
    repository: hourglass/operator
    tag: "latest"
    pullPolicy: IfNotPresent
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi
  env:
    logLevel: debug
```

### test-configs/executor-config.yaml
```yaml
# Test executor configuration
aggregator_endpoint: "test-aggregator.example.com:9090"
deployment_mode: "kubernetes"
performer_config:
  service_pattern: "performer-{name}.{namespace}.svc.cluster.local:9090"
  connection_timeout: "30s"
  retry_attempts: 3
kubernetes:
  namespace: "test-avs-project"
  operator_namespace: "hourglass-system"
```

### test-configs/test-performer.yaml
```yaml
# Test performer CRD
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: test-performer-1
  namespace: test-avs-project
spec:
  avsAddress: "0x1234567890abcdef1234567890abcdef12345678"
  image:
    repository: nginx
    tag: "1.21"
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "500m"
      memory: "256Mi"
```

## Validation Checks

### Automated Checks
- Pod creation and health
- Service creation and DNS resolution
- CRD status updates
- Event generation
- Resource utilization

### Manual Checks
- Operator logs for errors
- Performer pod functionality
- gRPC connectivity
- Performance metrics
- Resource cleanup

## Troubleshooting

### Common Issues

**Operator not starting**
```bash
# Check operator logs
kubectl logs -n hourglass-system deployment/hourglass-operator

# Check RBAC permissions
kubectl auth can-i create performers --as=system:serviceaccount:hourglass-system:hourglass-operator
```

**Performer pods not created**
```bash
# Check operator events
kubectl get events -n hourglass-system --sort-by=.lastTimestamp

# Check performer CRD status
kubectl get performers -n test-avs-project -o yaml
```

**DNS resolution issues**
```bash
# Test DNS from within cluster
kubectl run test-dns --image=busybox:1.35 --rm -i --restart=Never -- \
    nslookup performer-test-performer-1.test-avs-project.svc.cluster.local
```

**gRPC connectivity issues**
```bash
# Check service endpoints
kubectl get endpoints -n test-avs-project

# Test port connectivity
kubectl run test-grpc --image=busybox:1.35 --rm -i --restart=Never -- \
    nc -zv performer-test-performer-1.test-avs-project.svc.cluster.local 9090
```

## Test Results

Test results are automatically generated and saved to:
- `test-results/operator-integration.log`
- `test-results/multi-performer.log`
- `test-results/dns-connectivity.log`
- `test-results/performance.log`

## Contributing

To add new tests:

1. Create test scenario in `test-scenarios/`
2. Add validation logic to `test-setup.sh`
3. Update this README with new test description
4. Run tests and verify results

## Support

For issues with E2E tests:
- Check operator logs: `kubectl logs -n hourglass-system deployment/hourglass-operator`
- Review CRD status: `kubectl describe performers -n test-avs-project`
- Validate RBAC: `kubectl auth can-i <verb> <resource> --as=<serviceaccount>`