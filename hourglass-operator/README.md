# Hourglass Operator

The Hourglass Operator manages Performer Custom Resources in Kubernetes, automatically creating and managing pods and services for AVS (Actively Validated Service) workloads.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Development Workflow](#development-workflow)
- [Common Development Tasks](#common-development-tasks)
- [Architecture](#architecture)
- [Testing](#testing)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- Go 1.23+ (managed via gvm in this project)
- Docker
- kubectl with access to a Kubernetes cluster
- make

## Development Workflow

### Initial Setup

```bash
# Install all development dependencies (controller-gen, kustomize, etc.)
make controller-gen kustomize envtest golangci-lint

# Generate all required files
make generate manifests
```

### When to Run What

#### After modifying API types (`api/v1alpha1/*.go`)

When you change any types in `api/v1alpha1/` (e.g., adding fields to PerformerSpec):

```bash
# 1. Generate DeepCopy methods
make generate

# 2. Generate CRD manifests and RBAC rules
make manifests

# 3. Update the CRD in your cluster
kubectl apply -f config/crd/bases/hourglass.eigenlayer.io_performers.yaml

# 4. If using Helm, also update the Helm chart's CRD
# Copy the generated CRD content to charts/hourglass-operator/templates/crd.yaml
```

#### After modifying controller logic (`internal/controller/*.go`)

```bash
# 1. Format and validate your code
make fmt vet

# 2. Run tests
make test

# 3. Run linting
make lint

# 4. Build and test locally
make run
```

#### After modifying RBAC annotations

If you change any `+kubebuilder:rbac` annotations in your controller:

```bash
# Regenerate RBAC manifests
make manifests
```

## Common Development Tasks

### Running the Operator Locally

```bash
# Run the operator against your current kubeconfig context
make run

# Or build the binary first
make build
./bin/manager
```

### Building and Testing

```bash
# Run all code generation, formatting, and tests
make all

# Run only tests
make test

# Run linting
make lint

# Fix linting issues automatically
make lint-fix

# Run all tests and linting (good before committing)
make test-all
```

### Working with CRDs

```bash
# Install CRDs to cluster
make install

# Uninstall CRDs from cluster
make uninstall

# Generate CRD YAML files (after changing API types)
make manifests
```

### Building Container Images

```bash
# Build the operator image
make docker-build IMG=myregistry/hourglass-operator:v1.0.0

# Push the image
make docker-push IMG=myregistry/hourglass-operator:v1.0.0

# Build for multiple platforms
make docker-buildx IMG=myregistry/hourglass-operator:v1.0.0 PLATFORMS=linux/amd64,linux/arm64
```

### Deployment

```bash
# Deploy to cluster (using kustomize)
make deploy IMG=myregistry/hourglass-operator:v1.0.0

# Undeploy from cluster
make undeploy

# Generate a single YAML file with all resources
make build-installer IMG=myregistry/hourglass-operator:v1.0.0
# Output will be in dist/install.yaml
```

## Architecture

### Key Components

1. **Performer CRD** (`api/v1alpha1/performerTypes.go`)
   - Defines the Performer custom resource schema
   - Includes specifications for AVS containers, resources, scheduling, etc.

2. **Performer Controller** (`internal/controller/performerController.go`)
   - Watches for Performer resources
   - Creates and manages pods and services
   - Handles lifecycle events (creation, updates, deletion)

3. **Generated Code**
   - `zz_generated.deepcopy.go`: DeepCopy methods for all types
   - `config/crd/bases/*.yaml`: CRD definitions
   - `config/rbac/*.yaml`: RBAC rules

### Resource Ownership

The operator follows Kubernetes best practices:
- Performers own their Pods and Services
- Uses finalizers for cleanup
- Sets owner references for garbage collection

## Testing

### Unit Tests

```bash
# Run unit tests with coverage
make test

# View coverage report
go tool cover -html=cover.out
```

### E2E Tests

```bash
# Run end-to-end tests
make test-e2e
```

### Manual Testing

1. Create a test Performer:
```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: test-performer
spec:
  avsAddress: "0x123..."
  image: "my-avs:latest"
  imagePullPolicy: Never
  config:
    grpcPort: 8080
```

2. Apply and verify:
```bash
kubectl apply -f test-performer.yaml
kubectl get performers
kubectl get pods -l hourglass.eigenlayer.io/performer=test-performer
```

## Troubleshooting

### Common Issues

1. **CRD changes not reflected in cluster**
   - Run `make manifests` to regenerate CRDs
   - Apply the updated CRD: `kubectl apply -f config/crd/bases/hourglass.eigenlayer.io_performers.yaml`

2. **"no matches for kind Performer" error**
   - CRD not installed: run `make install`
   - Wrong API version: check `apiVersion: hourglass.eigenlayer.io/v1alpha1`

3. **RBAC errors**
   - Regenerate RBAC: `make manifests`
   - Redeploy: `make deploy`

4. **Code generation issues**
   - Ensure `+kubebuilder` markers are correct
   - Run `make generate manifests`

### Debug Commands

```bash
# Check operator logs
kubectl logs -n hourglass-system deployment/hourglass-operator-controller-manager

# Describe a performer
kubectl describe performer <name>

# Check generated pod
kubectl describe pod performer-<name>

# View CRD schema
kubectl get crd performers.hourglass.eigenlayer.io -o yaml
```

## Helm Chart Development

The operator includes Helm charts in `charts/`. After making changes:

1. Update `charts/hourglass-operator/templates/crd.yaml` if CRD changed
2. Bump version in `charts/hourglass-operator/Chart.yaml`
3. Package: `helm package charts/hourglass-operator`
4. Update index: `helm repo index chart_releases/`

## Contributing

1. Make your changes
2. Run `make test-all` to ensure tests pass
3. Update documentation if needed
4. Submit PR with clear description

## Quick Reference

| Command | When to Use | What it Does |
|---------|-------------|--------------|
| `make generate` | After changing API types | Generates DeepCopy methods |
| `make manifests` | After changing API types or RBAC | Generates CRDs and RBAC |
| `make run` | Testing locally | Runs operator outside cluster |
| `make test` | Before committing | Runs unit tests |
| `make lint` | Before committing | Checks code style |
| `make docker-build` | Building for deployment | Creates container image |
| `make deploy` | Deploying to cluster | Applies all resources |

## Environment Variables

The operator supports these environment variables:

- `KUBECONFIG`: Path to kubeconfig file (default: `~/.kube/config`)
- `NAMESPACE`: Namespace to watch (default: all namespaces)
- `ENABLE_WEBHOOKS`: Enable admission webhooks (default: false)