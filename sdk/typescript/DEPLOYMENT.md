# Hourglass Performer Deployment Guide

This guide covers deploying Hourglass TypeScript performers as Docker containers.

## Quick Start

### 1. Build Your Performer

```typescript
// performer.ts
import { SolidityWorker } from '@hourglass/performer';

class MyPerformer extends SolidityWorker {
  async handleSolidityTask(params: any) {
    // Your AVS logic here
    return { result: params.value * 2 };
  }
}

new MyPerformer().start();
```

### 2. Create Dockerfile

```dockerfile
FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy built code
COPY dist/ ./dist/

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD node -e "const http = require('http'); \
        const options = { hostname: 'localhost', port: 8080, path: '/health', timeout: 5000 }; \
        const req = http.request(options, (res) => { \
            process.exit(res.statusCode === 200 ? 0 : 1); \
        }); \
        req.on('error', () => process.exit(1)); \
        req.end();"

# Run performer
CMD ["node", "dist/performer.js"]
```

### 3. Build and Run

```bash
# Build your performer
npm run build

# Build Docker image
docker build -t my-performer .

# Run container
docker run -p 8080:8080 my-performer
```

## Environment Configuration

### Environment Variables

Configure your performer using environment variables:

```bash
# Server configuration
PORT=8080                    # Port to listen on
HOST=0.0.0.0                # Host to bind to
TIMEOUT=10000               # Request timeout in ms

# Application configuration
NODE_ENV=production         # Environment (development, production, test)
DEBUG=false                 # Enable debug mode
LOG_LEVEL=info             # Log level (error, warn, info, debug)

# Performance configuration
MAX_CONCURRENT_TASKS=10     # Maximum concurrent tasks
TASK_TIMEOUT=30000         # Task timeout in ms

# Health and monitoring
HEALTH_CHECK_INTERVAL=30000 # Health check interval in ms
METRICS_ENABLED=true        # Enable metrics collection
METRICS_PORT=9090          # Metrics port

# TypeChain configuration
TYPECHAIN_ENABLED=true      # Enable TypeChain
CONTRACTS_PATH=./contracts  # Path to contract ABIs

# Security configuration
CORS_ENABLED=true          # Enable CORS
CORS_ORIGINS=*             # Allowed CORS origins

# Storage configuration
DATA_DIR=./data            # Data directory
LOGS_DIR=./logs            # Logs directory
```

### Environment Files

Create `.env` files for different environments:

```bash
# .env.production
NODE_ENV=production
DEBUG=false
LOG_LEVEL=info
MAX_CONCURRENT_TASKS=50
TASK_TIMEOUT=45000
CORS_ENABLED=true
# Set CORS_ORIGINS to your domain in production
```

## Container Deployment

### Docker Run

```bash
# Basic run
docker run -p 8080:8080 my-performer

# With environment variables
docker run \
  -e NODE_ENV=production \
  -e LOG_LEVEL=info \
  -e MAX_CONCURRENT_TASKS=20 \
  -p 8080:8080 \
  my-performer

# With volumes for persistence
docker run \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/contracts:/app/contracts \
  -p 8080:8080 \
  my-performer
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  performer:
    build: .
    ports:
      - "8080:8080"
    environment:
      - NODE_ENV=production
      - LOG_LEVEL=info
      - MAX_CONCURRENT_TASKS=20
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
      - ./contracts:/app/contracts
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "8080"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

Run with:
```bash
docker-compose up -d
```

### Kubernetes

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-performer
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-performer
  template:
    metadata:
      labels:
        app: my-performer
    spec:
      containers:
      - name: performer
        image: my-performer:latest
        ports:
        - containerPort: 8080
        env:
        - name: NODE_ENV
          value: "production"
        - name: LOG_LEVEL
          value: "info"
        - name: MAX_CONCURRENT_TASKS
          value: "20"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: my-performer-service
spec:
  selector:
    app: my-performer
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

## Production Configuration

### Security Best Practices

1. **Run as non-root user**:
```dockerfile
RUN addgroup -g 1001 -S nodejs && \
    adduser -S performer -u 1001
USER performer
```

2. **Use read-only filesystem**:
```yaml
securityContext:
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1001
```

3. **Limit CORS origins**:
```bash
CORS_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
```

### Resource Limits

Set appropriate resource limits:

```yaml
resources:
  requests:
    memory: "256Mi"    # Minimum memory
    cpu: "250m"        # Minimum CPU
  limits:
    memory: "512Mi"    # Maximum memory
    cpu: "500m"        # Maximum CPU
```

### Health Checks

The performer includes built-in health checks:

```bash
# Health check endpoint
curl http://localhost:8080/health

# Docker health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1
```

### Monitoring

Enable metrics collection:

```bash
METRICS_ENABLED=true
METRICS_PORT=9090

# Access metrics
curl http://localhost:9090/metrics
```

### Logging

Configure structured logging:

```bash
LOG_LEVEL=info          # Production log level
LOGS_DIR=/app/logs      # Log directory
```

Logs are output to stdout/stderr and can be collected by your container orchestration platform.

## Troubleshooting

### Common Issues

1. **Port already in use**:
   - Change the PORT environment variable
   - Use different host port: `-p 8081:8080`

2. **Permission denied**:
   - Ensure proper user permissions
   - Check volume mount permissions

3. **Health check failing**:
   - Verify performer is listening on correct port
   - Check for startup errors in logs

4. **TypeChain errors**:
   - Ensure contract ABIs are mounted correctly
   - Verify CONTRACTS_PATH is correct

### Debug Mode

Enable debug mode for troubleshooting:

```bash
DEBUG=true
LOG_LEVEL=debug
NODE_ENV=development
```

### Container Logs

View container logs:

```bash
# Docker
docker logs <container-id>

# Docker Compose
docker-compose logs performer

# Kubernetes
kubectl logs deployment/my-performer
```

## Integration with Hourglass

Your containerized performer integrates with the Hourglass ecosystem:

1. **Executor Discovery**: The Hourglass executor discovers and connects to your performer
2. **Task Distribution**: Tasks are distributed via gRPC to your performer
3. **Result Aggregation**: Results are collected and aggregated by the Hourglass system

Ensure your performer is accessible on the network where the Hourglass executor runs.

## Next Steps

1. **Scale**: Use multiple replicas for high availability
2. **Monitor**: Set up monitoring and alerting
3. **Optimize**: Tune performance based on your workload
4. **Secure**: Implement proper security measures for production

For more information, see the [Hourglass documentation](https://docs.hourglass.io).