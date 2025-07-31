# Executor Configuration Guide

The Ponos Executor is responsible for running AVS workloads in isolated Docker containers and managing task execution on behalf of operators.

## Configuration File

The executor accepts configuration in YAML or JSON format. The configuration file path can be specified via the `--config` flag when starting the executor.

### Full Configuration Example

```yaml
# Operator identity and credentials
operator:
  address: "0xoperator..."              # The operator's Ethereum address
  operatorPrivateKey:
    privateKey: "..."                    # Private key for blockchain transactions
  signingKeys:                           # Keys for task response signing
    ecdsa:
      privateKey: ""                     # ECDSA key for standard signing
    bls:
      privateKey: ""                     # BLS key for aggregatable signatures

# AVS configurations
avss:                                    # List of AVSs this executor will run
  - image:                               # Docker image configuration
      repository: "eigenlabs/avs"        # Docker repository and image name
      tag: "v1.0.0"                      # Image version tag
    processType: "server"                # Process type: server, one-off, etc.
    avsAddress: "0xavs1..."              # AVS contract address for task filtering
    env:                                 # Optional environment variables
      - name: "API_KEY"
        value: "..."
      - name: "LOG_LEVEL"
        value: "debug"
    resources:                           # Optional resource limits
      cpuLimit: "2"                      # CPU cores limit
      memoryLimit: "4Gi"                 # Memory limit
      gpuRequired: false                 # GPU requirement

# Storage configuration (optional)
storage:
  type: "badger"                         # Storage backend: memory, badger
  badger:
    dir: "/var/lib/ponos/executor"       # Data directory for BadgerDB
    valueLogFileSize: 1073741824         # 1GB value log file size
    numVersionsToKeep: 1                 # Number of versions to keep
    numLevelZeroTables: 5                # Level 0 tables
    numLevelZeroTablesStall: 10          # Stall threshold

# Metrics and monitoring (optional)
metrics:
  enabled: true
  port: 9090
  path: "/metrics"

# Logging configuration (optional)
logging:
  level: "info"                          # Log level: debug, info, warn, error
  format: "json"                         # Log format: json, text
  output: "stdout"                       # Output: stdout, file
  file:                                  # If output is "file"
    path: "/var/log/ponos/executor.log"
    maxSize: 100                         # MB
    maxBackups: 3
    maxAge: 30                           # days
```

### Configuration Parameters

#### Operator Section

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operator.address` | string | Yes | The operator's Ethereum address |
| `operator.operatorPrivateKey.privateKey` | string | Yes | Private key for blockchain transactions |
| `operator.signingKeys.ecdsa.privateKey` | string | Conditional | ECDSA signing key (required if AVS uses ECDSA) |
| `operator.signingKeys.bls.privateKey` | string | Conditional | BLS signing key (required if AVS uses BLS) |

#### AVS Section

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `avss[].image.repository` | string | Yes | Docker image repository |
| `avss[].image.tag` | string | Yes | Docker image tag |
| `avss[].processType` | string | Yes | Process type (server, one-off) |
| `avss[].avsAddress` | string | Yes | AVS contract address |
| `avss[].env` | array | No | Environment variables for the container |
| `avss[].resources` | object | No | Resource limits for the container |

#### Storage Section

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `storage.type` | string | No | memory | Storage backend type |
| `storage.badger.dir` | string | Yes* | - | Data directory (*if type is badger) |
| `storage.badger.valueLogFileSize` | int | No | 1GB | Value log file size |
| `storage.badger.numVersionsToKeep` | int | No | 1 | Number of versions to keep |

### Environment Variables

The executor also supports configuration via environment variables:

- `EXECUTOR_CONFIG_PATH`: Path to the configuration file
- `EXECUTOR_OPERATOR_ADDRESS`: Override operator address
- `EXECUTOR_LOG_LEVEL`: Override log level
- `EXECUTOR_METRICS_PORT`: Override metrics port

### Security Considerations

1. **Private Key Storage**: Never commit private keys to version control. Use environment variables or secure key management systems.

2. **File Permissions**: Ensure configuration files containing private keys have restricted permissions:
   ```bash
   chmod 600 executor-config.yaml
   ```

3. **Docker Socket Access**: The executor requires access to the Docker daemon. Ensure proper permissions:
   ```bash
   sudo usermod -aG docker $USER
   ```

### Validation

The executor validates configuration on startup. Common validation errors:

- Missing required fields
- Invalid Ethereum addresses
- Inaccessible Docker images
- Invalid resource specifications
- Permission issues with storage directories

### Example Configurations

#### Minimal Configuration

```yaml
operator:
  address: "0x1234567890123456789012345678901234567890"
  operatorPrivateKey:
    privateKey: "${OPERATOR_PRIVATE_KEY}"
  signingKeys:
    ecdsa:
      privateKey: "${SIGNING_PRIVATE_KEY}"
avss:
  - image:
      repository: "myavs/service"
      tag: "latest"
    processType: "server"
    avsAddress: "0xabcdef1234567890123456789012345678901234"
```

#### Production Configuration

```yaml
operator:
  address: "${OPERATOR_ADDRESS}"
  operatorPrivateKey:
    privateKey: "${OPERATOR_PRIVATE_KEY}"
  signingKeys:
    ecdsa:
      privateKey: "${ECDSA_SIGNING_KEY}"
    bls:
      privateKey: "${BLS_SIGNING_KEY}"

avss:
  - image:
      repository: "registry.example.com/avs/service"
      tag: "v2.1.0"
    processType: "server"
    avsAddress: "${AVS_CONTRACT_ADDRESS}"
    env:
      - name: "NODE_ENV"
        value: "production"
      - name: "API_ENDPOINT"
        value: "https://api.example.com"
    resources:
      cpuLimit: "4"
      memoryLimit: "8Gi"

storage:
  type: "badger"
  badger:
    dir: "/data/ponos/executor"
    valueLogFileSize: 2147483648  # 2GB
    numVersionsToKeep: 1
    numLevelZeroTables: 10
    numLevelZeroTablesStall: 20

metrics:
  enabled: true
  port: 9090

logging:
  level: "info"
  format: "json"
  output: "file"
  file:
    path: "/var/log/ponos/executor.log"
    maxSize: 500
    maxBackups: 5
    maxAge: 30
```

### Configuration Reload

The executor supports configuration hot-reload for certain parameters without restart:

- AVS environment variables
- Logging levels
- Metrics configuration

To reload configuration:
```bash
kill -HUP <executor-pid>
```

Parameters requiring restart:
- Operator address and keys
- AVS list changes
- Storage backend configuration