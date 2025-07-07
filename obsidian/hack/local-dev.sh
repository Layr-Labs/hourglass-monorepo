#!/bin/bash
set -e

echo "Starting Obsidian local development environment..."

# Check prerequisites
command -v kind >/dev/null 2>&1 || { echo "kind is required but not installed. Aborting." >&2; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo "kubectl is required but not installed. Aborting." >&2; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "docker is required but not installed. Aborting." >&2; exit 1; }

# Create kind cluster
echo "Creating kind cluster..."
cat <<EOF | kind create cluster --name obsidian-dev --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30080
    hostPort: 8080
    protocol: TCP
- role: worker
  extraMounts:
  - hostPath: /dev
    containerPath: /dev
- role: worker
  extraMounts:
  - hostPath: /dev
    containerPath: /dev
EOF

# Build images
echo "Building Docker images..."
make docker-build

# Load images into kind
echo "Loading images into kind cluster..."
kind load docker-image localhost:5000/obsidian-service:v1.0.0 --name obsidian-dev
kind load docker-image localhost:5000/obsidian-operator:v1.0.0 --name obsidian-dev

# Install CRDs
echo "Installing CRDs..."
make manifests
kubectl apply -f deployments/kubernetes/crds/

# Deploy operator
echo "Deploying operator..."
kubectl apply -f deployments/kubernetes/namespace.yaml
kubectl apply -f deployments/kubernetes/rbac.yaml
kubectl apply -f deployments/kubernetes/operator-deployment.yaml

# Wait for operator to be ready
echo "Waiting for operator to be ready..."
kubectl wait --for=condition=available --timeout=60s deployment/avs-operator -n obsidian-system

# Deploy example AVS
echo "Deploying example AVS..."
kubectl apply -f deployments/kubernetes/example-avs.yaml

echo "Local development environment is ready!"
echo "Access the service at: http://localhost:8080"
echo ""
echo "Useful commands:"
echo "  kubectl get avs"
echo "  kubectl logs -n obsidian-system deployment/avs-operator"
echo "  kubectl port-forward svc/example-avs-service 8080:8080"