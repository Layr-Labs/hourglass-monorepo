# Obsidian: AWS-Based Container Orchestration for Hourglass

Obsidian is a secure, scalable container orchestration platform designed to enable AWS-based AVS (Actively Validated Services) execution within the Hourglass framework.

## Architecture

Obsidian consists of three core services:

1. **Orchestrator Service**: Manages container lifecycle, task execution, and resource allocation
2. **Registry Service**: Handles container image management, caching, and security scanning
3. **Proxy Service**: Provides controlled access to external services with rate limiting and filtering

## Quick Start

### Local Development

1. Install dependencies:
```bash
make deps
make install-tools
```

2. Generate protobuf code:
```bash
make proto
```

3. Build the binary:
```bash
make build
```

4. Run with development configuration:
```bash
make run
```

### Docker Deployment

Build and run using Docker:
```bash
make docker
docker run -p 8080:8080 -p 9090:9090 obsidian:latest
```

### AWS Deployment

Deploy using CDK:

1. Install CDK dependencies:
```bash
make cdk-install
```

2. Deploy to AWS:
```bash
# Development (single EC2 instance)
make cdk-deploy-dev

# Production (hybrid deployment)
make cdk-deploy-prod
```

## Configuration

Obsidian uses YAML configuration files. See `config/development.yaml` for an example.

Key configuration sections:
- `orchestrator`: Container management settings
- `registry`: Image registry configuration
- `proxy`: External service proxy settings
- `server`: HTTP/gRPC server configuration

## Integration with Hourglass

Obsidian implements the `IAvsPerformer` interface for seamless integration with Hourglass:

```go
import "github.com/hourglass/obsidian/pkg/performer"

performer, err := performer.NewObsidianPerformer(&performer.Config{
    ObsidianEndpoint: "localhost:9090",
    AvsAddress:       "0x...",
    Resources: &performer.ResourceLimits{
        CPU:    "2",
        Memory: "4Gi",
        Disk:   "20Gi",
    },
})
```

## API Endpoints

### HTTP Endpoints
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

### gRPC Services
- `OrchestratorService` - Container and task management
- `RegistryService` - Image management
- `ProxyService` - External service proxy

## Development

### Running Tests
```bash
make test
```

### Linting
```bash
make lint
```

### Monitoring

Obsidian exposes Prometheus metrics on port 9091:
- Container metrics
- Task execution metrics
- Registry cache metrics
- Proxy request metrics

## Security

- All containers run with minimal privileges
- Image vulnerability scanning enabled by default
- Rate limiting and request filtering for external calls
- AWS IAM integration for authentication

## License

See LICENSE file in the repository root.