# Obsidian - Verifiable Confidential Computing Platform

Obsidian is a platform for running verifiable confidential computing workloads. It provides attestation of software authenticity while serving non-deterministic information, designed to work with the Hourglass framework and Constellation for confidential Kubernetes deployments.

## Overview

Obsidian enables running non-deterministic workloads (e.g., AI inference, random number generation, timestamp services) while providing cryptographic proof of the software's authenticity through hardware-based attestation (SEV-SNP, TDX, SGX).

## Architecture Components

1. **Obsidian Service**: Core attestation service that processes compute requests
2. **AVS Operator**: Kubernetes operator managing AVS (Actively Validated Services) CRDs
3. **Attestation Agent**: Node-level agent for platform attestation
4. **On-chain Registry**: Blockchain contract for attestation verification

## Quick Start

### Prerequisites

- Go 1.21+
- Docker
- Kubernetes (kind for local development)
- kubectl
- controller-gen (for CRD generation)

### Local Development

```bash
# Start local Kubernetes cluster with Obsidian
./hack/local-dev.sh

# Build and run locally
make build
make run-local

# Run tests
make test
```

### Building

```bash
# Build binaries
make build

# Build Docker images
make docker-build

# Push images
make docker-push
```

### Deployment

```bash
# Generate CRDs
make manifests

# Deploy to Kubernetes
make deploy

# Deploy example AVS
kubectl apply -f deployments/kubernetes/example-avs.yaml
```

## API Usage

### Health Check
```bash
curl http://localhost:8080/health
```

### Get Attestation
```bash
curl http://localhost:8080/attestation
```

### Compute Request
```bash
curl -X POST http://localhost:8080/api/compute \
  -H "Content-Type: application/json" \
  -H "X-Nonce: unique-nonce-123" \
  -d '{
    "type": "random",
    "input": {}
  }'
```

### Verify Output
```bash
curl http://localhost:8080/api/verify/{output-id}
```

## AVS Custom Resource

Example AVS deployment:

```yaml
apiVersion: hourglass.io/v1alpha1
kind: AVS
metadata:
  name: example-avs
spec:
  operator: "operator-1"
  serviceImage: "obsidian-service:v1.0.0"
  replicas: 3
  computeRequirements:
    cpu: "1"
    memory: "2Gi"
    teeType: "SEV-SNP"
    nodeSelector:
      feature.node.kubernetes.io/cpu-sev: "true"
  attestationPolicy:
    allowedMeasurements:
      - "7b068c0c3ac29afe264134536b9be26f1e4ccd575b88d3e3be77e768414ce98d"
    maxAttestationAge: "1h"
    requireSEV: true
```

## Security Features

- Hardware-based attestation (SEV-SNP, TDX, SGX)
- Cryptographic binding of outputs to attestations
- Nonce-based freshness guarantees
- Configurable measurement policies
- On-chain attestation verification

## Integration with Hourglass

Obsidian is designed to work seamlessly with the Hourglass framework:

1. AVS operators register their services
2. Obsidian provides attestation for compute workloads
3. Results are aggregated using BLS signatures
4. On-chain verification ensures authenticity

## Development

### Project Structure

```
obsidian/
├── cmd/                    # Entry points
├── pkg/                    # Core packages
├── api/v1alpha1/          # CRD definitions
├── deployments/           # Deployment configs
├── build/                 # Dockerfiles
└── hack/                  # Development scripts
```

### Testing

```bash
# Unit tests
make test

# Integration tests (requires kind)
./hack/local-dev.sh
kubectl apply -f deployments/kubernetes/example-avs.yaml
kubectl port-forward svc/example-avs-service 8080:8080
```

## License

Apache 2.0