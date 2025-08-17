# Kubernetes Configuration Examples

This document provides comprehensive examples for configuring Hourglass executors in Kubernetes mode.

## Table of Contents

1. [Basic Kubernetes Configuration](#basic-kubernetes-configuration)
2. [Production Kubernetes Configuration](#production-kubernetes-configuration)
3. [Multi-AVS Kubernetes Configuration](#multi-avs-kubernetes-configuration)
4. [Development Kubernetes Configuration](#development-kubernetes-configuration)
5. [Advanced Features Examples](#advanced-features-examples)

## Basic Kubernetes Configuration

### Simple Single-AVS Setup

```yaml
# basic-kubernetes-config.yaml
aggregator_endpoint: "aggregator.example.com:9090"
aggregator_tls_enabled: false
deployment_mode: "kubernetes"
log_level: "info"
log_format: "json"

performer_config:
  service_pattern: "performer-{name}.{namespace}.svc.cluster.local:{port}"
  default_port: 9090
  connection_timeout: "30s"
  startup_timeout: "300s"
  retry_attempts: 3
  max_performers: 5

kubernetes:
  namespace: "my-avs-project"
  generateNamespace: true
  operator_namespace: "hourglass-system"
  performer_crd_group: "hourglass.eigenlayer.io"
  performer_crd_version: "v1alpha1"
  performer_crd_kind: "Performer"
  cleanup_on_shutdown: true
  cleanup_timeout: "60s"
  connection_timeout: "30s"
  in_cluster: true

chains:
  - name: "ethereum"
    chain_id: 1
    rpc_url: "https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY"
    task_mailbox_address: "0x1234567890abcdef1234567890abcdef12345678"
    block_confirmations: 12
    gas_limit: 300000
    event_filter:
      from_block: "latest"

avs_config:
  supported_avs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "my-avs"
      performer_image: "my-org/my-avs-performer"
      performer_version: "v1.0.0"
      deployment_mode: "kubernetes"
      default_resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "2"
          memory: "4Gi"
      env:
        - name: "AVS_CONFIG_PATH"
          value: "/etc/avs/config.json"
        - name: "LOG_LEVEL"
          value: "info"

metrics:
  enabled: true
  port: 8080
  path: "/metrics"

health:
  enabled: true
  port: 8090
  path: "/health"
```

### Helm Values for Basic Setup

```yaml
# basic-kubernetes-values.yaml
executor:
  name: "my-avs-executor"
  replicaCount: 1
  
  image:
    repository: "hourglass/executor"
    tag: "v1.2.0"
    pullPolicy: "IfNotPresent"

  env:
    deploymentMode: "kubernetes"
    logLevel: "info"
    logFormat: "json"

# Global configuration
global:
  operatorNamespace: "hourglass-system"

# Aggregator configuration
aggregator:
  endpoint: "aggregator.example.com:9090"
  tls:
    enabled: false

# Kubernetes-specific configuration
kubernetes:
  namespace: "my-avs-project"
  performerCRD:
    group: "hourglass.eigenlayer.io"
    version: "v1alpha1"
    kind: "Performer"
  connection:
    timeout: "30s"
    retryAttempts: 3
  cleanup:
    onShutdown: true
    timeout: "60s"

# AVS configuration
avs:
  supportedAvs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "my-avs"
      performer:
        image: "my-org/my-avs-performer"
        version: "v1.0.0"
        deploymentMode: "kubernetes"
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

# Blockchain configuration
chains:
  ethereum:
    enabled: true
    chainId: 1
    rpcUrl: "https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY"
    taskMailboxAddress: "0x1234567890abcdef1234567890abcdef12345678"
    blockConfirmations: 12

# Secrets (base64 encoded)
secrets:
  operatorKeys:
    ecdsaPrivateKey: "LS0tLS1CRUdJTi..."
    blsPrivateKey: "LS0tLS1CRUdJTi..."

# Monitoring
metrics:
  enabled: true
  port: 8080
  serviceMonitor:
    enabled: true
    namespace: "monitoring"

# Storage
persistence:
  enabled: true
  size: "10Gi"
  storageClass: "fast-ssd"
```

## Production Kubernetes Configuration

### High Availability Setup

```yaml
# production-kubernetes-config.yaml
aggregator_endpoint: "aggregator.production.com:9090"
aggregator_tls_enabled: true
deployment_mode: "kubernetes"
log_level: "warn"
log_format: "json"

performer_config:
  service_pattern: "performer-{name}.{namespace}.svc.cluster.local:{port}"
  default_port: 9090
  connection_timeout: "60s"
  startup_timeout: "600s"
  retry_attempts: 5
  max_performers: 10
  health_check_interval: "30s"
  unhealthy_threshold: 3

kubernetes:
  namespace: "production-avs"
  operator_namespace: "hourglass-system"
  performer_crd_group: "hourglass.eigenlayer.io"
  performer_crd_version: "v1alpha1"
  performer_crd_kind: "Performer"
  cleanup_on_shutdown: true
  cleanup_timeout: "300s"
  connection_timeout: "60s"
  in_cluster: true

chains:
  - name: "ethereum"
    chain_id: 1
    rpc_url: "https://eth-mainnet.alchemyapi.io/v2/PRODUCTION_API_KEY"
    task_mailbox_address: "0x1234567890abcdef1234567890abcdef12345678"
    block_confirmations: 12
    gas_limit: 500000
    event_filter:
      from_block: "latest"
      batch_size: 100

avs_config:
  supported_avs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "production-avs"
      performer_image: "my-org/my-avs-performer"
      performer_version: "v2.1.0"
      deployment_mode: "kubernetes"
      default_resources:
        requests:
          cpu: "1000m"
          memory: "2Gi"
        limits:
          cpu: "4000m"
          memory: "8Gi"
      env:
        - name: "AVS_CONFIG_PATH"
          value: "/etc/avs/config.json"
        - name: "LOG_LEVEL"
          value: "warn"
        - name: "METRICS_ENABLED"
          value: "true"
        - name: "TRACING_ENABLED"
          value: "true"
      node_selector:
        node-type: "high-performance"
      tolerations:
        - key: "avs-workload"
          operator: "Equal"
          value: "production"
          effect: "NoSchedule"
      affinity:
        pod_anti_affinity:
          preferred_during_scheduling_ignored_during_execution:
            - weight: 100
              pod_affinity_term:
                label_selector:
                  match_expressions:
                    - key: "app"
                      operator: "In"
                      values: ["avs-performer"]
                topology_key: "kubernetes.io/hostname"

metrics:
  enabled: true
  port: 8080
  path: "/metrics"

health:
  enabled: true
  port: 8090
  path: "/health"
  liveness_probe:
    initial_delay_seconds: 30
    period_seconds: 10
    timeout_seconds: 5
    failure_threshold: 3
  readiness_probe:
    initial_delay_seconds: 5
    period_seconds: 5
    timeout_seconds: 3
    failure_threshold: 3
```

### Production Helm Values

```yaml
# production-kubernetes-values.yaml
executor:
  name: "production-avs-executor"
  replicaCount: 3  # High availability
  
  image:
    repository: "my-org/hourglass-executor"
    tag: "v2.1.0"
    pullPolicy: "IfNotPresent"

  resources:
    requests:
      cpu: "1000m"
      memory: "2Gi"
    limits:
      cpu: "4000m"
      memory: "8Gi"

  # Pod security context
  securityContext:
    runAsNonRoot: true
    runAsUser: 65532
    runAsGroup: 65532
    fsGroup: 65532

  # Container security context
  containerSecurityContext:
    allowPrivilegeEscalation: false
    capabilities:
      drop:
        - ALL
    readOnlyRootFilesystem: true

  # Node selection
  nodeSelector:
    node-type: "high-performance"

  tolerations:
    - key: "avs-workload"
      operator: "Equal"
      value: "production"
      effect: "NoSchedule"

  affinity:
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          podAffinityTerm:
            labelSelector:
              matchExpressions:
                - key: "app"
                  operator: "In"
                  values: ["hourglass-executor"]
            topologyKey: "kubernetes.io/hostname"

# Production aggregator configuration
aggregator:
  endpoint: "aggregator.production.com:9090"
  tls:
    enabled: true
    certFile: "/etc/ssl/certs/aggregator.crt"
    keyFile: "/etc/ssl/private/aggregator.key"
    caFile: "/etc/ssl/certs/ca.crt"

# Production AVS configuration
avs:
  supportedAvs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "production-avs"
      performer:
        image: "my-org/my-avs-performer"
        version: "v2.1.0"
        deploymentMode: "kubernetes"
      resources:
        requests:
          cpu: "1000m"
          memory: "2Gi"
        limits:
          cpu: "4000m"
          memory: "8Gi"
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532
        readOnlyRootFilesystem: true
      nodeSelector:
        node-type: "high-performance"
      tolerations:
        - key: "avs-workload"
          operator: "Equal"
          value: "production"
          effect: "NoSchedule"

# Production blockchain configuration
chains:
  ethereum:
    enabled: true
    chainId: 1
    rpcUrl: "https://eth-mainnet.alchemyapi.io/v2/PRODUCTION_API_KEY"
    taskMailboxAddress: "0x1234567890abcdef1234567890abcdef12345678"
    blockConfirmations: 12
    gasLimit: 500000
    connectionPool:
      maxConnections: 10
      idleTimeout: "300s"

# Production monitoring
metrics:
  enabled: true
  port: 8080
  serviceMonitor:
    enabled: true
    namespace: "monitoring"
    interval: "30s"
    scrapeTimeout: "10s"
    labels:
      team: "avs-team"
      environment: "production"

# Production storage
persistence:
  enabled: true
  size: "100Gi"
  storageClass: "fast-ssd"
  accessMode: "ReadWriteOnce"

# Resource quotas
resourceQuota:
  enabled: true
  hard:
    requests.cpu: "10"
    requests.memory: "20Gi"
    limits.cpu: "40"
    limits.memory: "80Gi"
    persistentvolumeclaims: "10"

# Network policies
networkPolicy:
  enabled: true
  policyTypes:
    - "Ingress"
    - "Egress"
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: "hourglass-system"
        - namespaceSelector:
            matchLabels:
              name: "monitoring"
  egress:
    - to: []
      ports:
        - protocol: "TCP"
          port: 443  # HTTPS
        - protocol: "TCP"
          port: 9090  # Aggregator
```

## Multi-AVS Kubernetes Configuration

### Multiple AVS Support

```yaml
# multi-avs-kubernetes-config.yaml
aggregator_endpoint: "aggregator.example.com:9090"
aggregator_tls_enabled: false
deployment_mode: "kubernetes"
log_level: "info"
log_format: "json"

performer_config:
  service_pattern: "performer-{name}.{namespace}.svc.cluster.local:{port}"
  default_port: 9090
  connection_timeout: "30s"
  startup_timeout: "300s"
  retry_attempts: 3
  max_performers: 15

kubernetes:
  namespace: "multi-avs-project"
  operator_namespace: "hourglass-system"
  performer_crd_group: "hourglass.eigenlayer.io"
  performer_crd_version: "v1alpha1"
  performer_crd_kind: "Performer"
  cleanup_on_shutdown: true
  cleanup_timeout: "60s"
  connection_timeout: "30s"
  in_cluster: true

chains:
  - name: "ethereum"
    chain_id: 1
    rpc_url: "https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY"
    task_mailbox_address: "0x1234567890abcdef1234567890abcdef12345678"
    block_confirmations: 12
    gas_limit: 300000
    event_filter:
      from_block: "latest"

avs_config:
  supported_avs:
    # AVS 1: Data availability
    - address: "0x1111111111111111111111111111111111111111"
      name: "data-availability-avs"
      performer_image: "my-org/da-performer"
      performer_version: "v1.0.0"
      deployment_mode: "kubernetes"
      default_resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "2"
          memory: "4Gi"
      env:
        - name: "AVS_TYPE"
          value: "data-availability"
        - name: "STORAGE_PATH"
          value: "/data/da"
      node_selector:
        avs-type: "data-availability"
      
    # AVS 2: Oracle services
    - address: "0x2222222222222222222222222222222222222222"
      name: "oracle-avs"
      performer_image: "my-org/oracle-performer"
      performer_version: "v1.1.0"
      deployment_mode: "kubernetes"
      default_resources:
        requests:
          cpu: "1000m"
          memory: "2Gi"
        limits:
          cpu: "4"
          memory: "8Gi"
      env:
        - name: "AVS_TYPE"
          value: "oracle"
        - name: "PRICE_FEED_URL"
          value: "https://api.coinbase.com/v2/prices"
      node_selector:
        avs-type: "oracle"
      
    # AVS 3: Compute services
    - address: "0x3333333333333333333333333333333333333333"
      name: "compute-avs"
      performer_image: "my-org/compute-performer"
      performer_version: "v2.0.0"
      deployment_mode: "kubernetes"
      default_resources:
        requests:
          cpu: "2000m"
          memory: "4Gi"
        limits:
          cpu: "8"
          memory: "16Gi"
      env:
        - name: "AVS_TYPE"
          value: "compute"
        - name: "WORKER_THREADS"
          value: "8"
      node_selector:
        avs-type: "compute"
      tolerations:
        - key: "compute-intensive"
          operator: "Equal"
          value: "true"
          effect: "NoSchedule"

metrics:
  enabled: true
  port: 8080
  path: "/metrics"

health:
  enabled: true
  port: 8090
  path: "/health"
```

## Development Kubernetes Configuration

### Local Development Setup

```yaml
# development-kubernetes-config.yaml
aggregator_endpoint: "aggregator.dev.local:9090"
aggregator_tls_enabled: false
deployment_mode: "kubernetes"
log_level: "debug"
log_format: "text"

performer_config:
  service_pattern: "performer-{name}.{namespace}.svc.cluster.local:{port}"
  default_port: 9090
  connection_timeout: "15s"
  startup_timeout: "120s"
  retry_attempts: 2
  max_performers: 3

kubernetes:
  namespace: "dev-avs-project"
  operator_namespace: "hourglass-system"
  performer_crd_group: "hourglass.eigenlayer.io"
  performer_crd_version: "v1alpha1"
  performer_crd_kind: "Performer"
  cleanup_on_shutdown: true
  cleanup_timeout: "30s"
  connection_timeout: "15s"
  in_cluster: true

chains:
  - name: "ethereum"
    chain_id: 31337  # Local anvil
    rpc_url: "http://anvil.dev.local:8545"
    task_mailbox_address: "0x1234567890abcdef1234567890abcdef12345678"
    block_confirmations: 1
    gas_limit: 200000
    event_filter:
      from_block: "latest"

avs_config:
  supported_avs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "dev-avs"
      performer_image: "my-org/my-avs-performer"
      performer_version: "latest"
      deployment_mode: "kubernetes"
      default_resources:
        requests:
          cpu: "100m"
          memory: "256Mi"
        limits:
          cpu: "500m"
          memory: "1Gi"
      env:
        - name: "ENV"
          value: "development"
        - name: "DEBUG"
          value: "true"
        - name: "LOG_LEVEL"
          value: "debug"
      image_pull_policy: "Always"

metrics:
  enabled: true
  port: 8080
  path: "/metrics"

health:
  enabled: true
  port: 8090
  path: "/health"
```

### Development Helm Values

```yaml
# development-kubernetes-values.yaml
executor:
  name: "dev-avs-executor"
  replicaCount: 1
  
  image:
    repository: "my-org/hourglass-executor"
    tag: "latest"
    pullPolicy: "Always"

  resources:
    requests:
      cpu: "100m"
      memory: "256Mi"
    limits:
      cpu: "500m"
      memory: "1Gi"

  env:
    deploymentMode: "kubernetes"
    logLevel: "debug"
    logFormat: "text"

# Development aggregator configuration
aggregator:
  endpoint: "aggregator.dev.local:9090"
  tls:
    enabled: false

# Development AVS configuration
avs:
  supportedAvs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "dev-avs"
      performer:
        image: "my-org/my-avs-performer"
        version: "latest"
        deploymentMode: "kubernetes"
        pullPolicy: "Always"
      resources:
        requests:
          cpu: "100m"
          memory: "256Mi"
        limits:
          cpu: "500m"
          memory: "1Gi"
      env:
        - name: "ENV"
          value: "development"
        - name: "DEBUG"
          value: "true"

# Development blockchain configuration
chains:
  ethereum:
    enabled: true
    chainId: 31337
    rpcUrl: "http://anvil.dev.local:8545"
    taskMailboxAddress: "0x1234567890abcdef1234567890abcdef12345678"
    blockConfirmations: 1

# Development secrets (dummy values)
secrets:
  operatorKeys:
    ecdsaPrivateKey: "LS0tLS1CRUdJTi..."
    blsPrivateKey: "LS0tLS1CRUdJTi..."

# Development monitoring
metrics:
  enabled: true
  port: 8080
  serviceMonitor:
    enabled: false

# Development storage
persistence:
  enabled: false
```

## Advanced Features Examples

### Blue-Green Deployment Configuration

```yaml
# blue-green-kubernetes-config.yaml
aggregator_endpoint: "aggregator.example.com:9090"
deployment_mode: "kubernetes"
log_level: "info"

kubernetes:
  namespace: "production-avs"
  operator_namespace: "hourglass-system"
  performer_crd_group: "hourglass.eigenlayer.io"
  performer_crd_version: "v1alpha1"
  performer_crd_kind: "Performer"
  cleanup_on_shutdown: true
  cleanup_timeout: "60s"
  blue_green_deployment:
    enabled: true
    switch_strategy: "manual"  # or "automatic"
    health_check_duration: "5m"
    rollback_on_failure: true

avs_config:
  supported_avs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "production-avs"
      performer_image: "my-org/my-avs-performer"
      performer_version: "v2.0.0"
      deployment_mode: "kubernetes"
      blue_green_config:
        enabled: true
        active_color: "blue"  # Current active deployment
        inactive_color: "green"  # Standby deployment
        switch_threshold: 0.95  # Health percentage to switch
      default_resources:
        requests:
          cpu: "1000m"
          memory: "2Gi"
        limits:
          cpu: "4000m"
          memory: "8Gi"
```

### Auto-scaling Configuration

```yaml
# autoscaling-kubernetes-config.yaml
aggregator_endpoint: "aggregator.example.com:9090"
deployment_mode: "kubernetes"
log_level: "info"

kubernetes:
  namespace: "scalable-avs"
  operator_namespace: "hourglass-system"
  performer_crd_group: "hourglass.eigenlayer.io"
  performer_crd_version: "v1alpha1"
  performer_crd_kind: "Performer"
  cleanup_on_shutdown: true
  cleanup_timeout: "60s"
  auto_scaling:
    enabled: true
    min_replicas: 2
    max_replicas: 10
    target_cpu_utilization: 70
    target_memory_utilization: 80
    scale_up_cooldown: "5m"
    scale_down_cooldown: "10m"

avs_config:
  supported_avs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "scalable-avs"
      performer_image: "my-org/my-avs-performer"
      performer_version: "v1.0.0"
      deployment_mode: "kubernetes"
      auto_scaling:
        enabled: true
        min_replicas: 2
        max_replicas: 10
        metrics:
          - type: "Resource"
            resource:
              name: "cpu"
              target:
                type: "Utilization"
                average_utilization: 70
          - type: "Resource"
            resource:
              name: "memory"
              target:
                type: "Utilization"
                average_utilization: 80
      default_resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "2"
          memory: "4Gi"
```

### Monitoring and Alerting Configuration

```yaml
# monitoring-kubernetes-config.yaml
aggregator_endpoint: "aggregator.example.com:9090"
deployment_mode: "kubernetes"
log_level: "info"

kubernetes:
  namespace: "monitored-avs"
  operator_namespace: "hourglass-system"
  performer_crd_group: "hourglass.eigenlayer.io"
  performer_crd_version: "v1alpha1"
  performer_crd_kind: "Performer"
  cleanup_on_shutdown: true
  cleanup_timeout: "60s"

avs_config:
  supported_avs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "monitored-avs"
      performer_image: "my-org/my-avs-performer"
      performer_version: "v1.0.0"
      deployment_mode: "kubernetes"
      monitoring:
        enabled: true
        prometheus:
          scrape_interval: "30s"
          scrape_timeout: "10s"
          metrics_path: "/metrics"
          labels:
            team: "avs-team"
            environment: "production"
        alerts:
          - name: "PerformerDown"
            condition: "up == 0"
            for: "1m"
            severity: "critical"
            message: "Performer {{$labels.instance}} is down"
          - name: "HighCPUUsage"
            condition: "cpu_usage > 80"
            for: "5m"
            severity: "warning"
            message: "High CPU usage on {{$labels.instance}}"
          - name: "HighMemoryUsage"
            condition: "memory_usage > 85"
            for: "5m"
            severity: "warning"
            message: "High memory usage on {{$labels.instance}}"
        logging:
          enabled: true
          level: "info"
          format: "json"
          output: "stdout"
          fields:
            - "timestamp"
            - "level"
            - "message"
            - "performer_id"
            - "avs_address"
      default_resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "2"
          memory: "4Gi"

metrics:
  enabled: true
  port: 8080
  path: "/metrics"
  
health:
  enabled: true
  port: 8090
  path: "/health"
```

## Configuration Validation

All examples above include validation that ensures:

1. **Single Deployment Mode**: All AVS performers use `deployment_mode: "kubernetes"`
2. **Required Kubernetes Section**: Kubernetes configuration is present and valid
3. **Proper Resource Limits**: All performers have resource requests and limits
4. **Valid Image References**: All images are properly tagged and accessible
5. **Security Best Practices**: Non-root users, read-only filesystems, dropped capabilities

## Next Steps

- **Deploy Examples**: Use these configurations as starting points for your deployment
- **Customize**: Modify resource requirements, environment variables, and other settings
- **Test**: Validate configurations in a development environment first
- **Monitor**: Set up monitoring and alerting for production deployments
- **Scale**: Use auto-scaling features for production workloads
