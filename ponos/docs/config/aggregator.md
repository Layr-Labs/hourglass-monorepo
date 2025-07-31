# Aggregator Configuration Guide

The Ponos Aggregator is responsible for coordinating tasks across multiple chains, managing operator responses, and submitting aggregated results back to the blockchain.

## Configuration File

The aggregator accepts configuration in YAML or JSON format. The configuration file path can be specified via the `--config` flag when starting the aggregator.

### Full Configuration Example

```yaml
# Global chain configurations
chains:                                          # List of chains the aggregator monitors
  - name: ethereum                               # Human-readable chain name
    network: mainnet                             # Network identifier
    chainId: 1                                   # EIP-155 chain ID
    rpcUrl: "https://eth-mainnet.g.alchemy.com/v2/${ALCHEMY_KEY}"
    wsUrl: "wss://eth-mainnet.g.alchemy.com/v2/${ALCHEMY_KEY}"  # Optional WebSocket URL
    blockConfirmations: 12                       # Blocks to wait before processing
    pollInterval: 12000                          # Milliseconds between polls
    maxBlockRange: 1000                          # Max blocks to query at once
  - name: base
    network: mainnet
    chainId: 8453
    rpcUrl: "https://base-mainnet.g.alchemy.com/v2/${ALCHEMY_KEY}"
    blockConfirmations: 6
    pollInterval: 2000
    maxBlockRange: 5000

# AVS-specific configurations
avss:                                            # List of AVSs to aggregate for
  - address: "0xavs1..."                         # AVS contract address
    privateKey: "${AVS_PRIVATE_KEY}"             # Key for blockchain transactions
    privateSigningKey: "${AVS_SIGNING_KEY}"      # Key for message signing
    privateSigningKeyType: "ecdsa"               # Signing key type: ecdsa, bls
    responseTimeout: 3000                        # Response timeout in seconds
    chainIds: [1, 8453]                          # Chains to monitor for this AVS
    
    # Operator set configurations
    operatorSets:
      - id: 0                                    # Operator set ID
        minOperators: 3                          # Minimum operators required
        consensusThreshold: 6600                 # Threshold (66.00%)
        taskSLA: 3600                            # Task SLA in seconds
        
    # Task-specific configurations
    taskConfig:
      maxRetries: 3                              # Max retries for failed tasks
      retryDelay: 60                             # Seconds between retries
      batchSize: 100                             # Max tasks per batch
      
    # Advanced configurations
    advanced:
      skipBlockValidation: false                 # Skip block validation
      allowOperatorOverrides: true               # Allow operator overrides
      customGasLimits:                          # Custom gas limits
        submitResponse: 500000
        completeTask: 300000

# Storage configuration
storage:
  type: "badger"                                 # Storage backend: memory, badger
  badger:
    dir: "/var/lib/ponos/aggregator"             # Data directory
    valueLogFileSize: 2147483648                 # 2GB value log file size
    numVersionsToKeep: 1                         # Number of versions to keep
    numLevelZeroTables: 10                       # Level 0 tables
    numLevelZeroTablesStall: 20                  # Stall threshold
    compactL0OnClose: true                       # Compact on close
    readOnly: false                              # Read-only mode

# gRPC server configuration
grpc:
  enabled: true
  port: 50051
  maxRecvMsgSize: 4194304                        # 4MB max message size
  maxSendMsgSize: 4194304
  tls:
    enabled: false
    certFile: "/etc/ponos/tls/server.crt"
    keyFile: "/etc/ponos/tls/server.key"
    caFile: "/etc/ponos/tls/ca.crt"

# Metrics and monitoring
metrics:
  enabled: true
  port: 9091
  path: "/metrics"
  namespace: "ponos_aggregator"

# Logging configuration
logging:
  level: "info"                                  # debug, info, warn, error
  format: "json"                                 # json, text
  output: "stdout"                               # stdout, file
  file:
    path: "/var/log/ponos/aggregator.log"
    maxSize: 500                                 # MB
    maxBackups: 5
    maxAge: 30                                   # days
    compress: true

# Health checks
health:
  enabled: true
  port: 8080
  path: "/health"
  checkInterval: 30                              # seconds
```

### Configuration Parameters

#### Chains Section

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `chains[].name` | string | Yes | - | Human-readable chain name |
| `chains[].network` | string | Yes | - | Network identifier (mainnet, testnet, etc.) |
| `chains[].chainId` | integer | Yes | - | EIP-155 chain ID |
| `chains[].rpcUrl` | string | Yes | - | HTTP RPC endpoint URL |
| `chains[].wsUrl` | string | No | - | WebSocket RPC endpoint URL |
| `chains[].blockConfirmations` | integer | No | 12 | Blocks to wait before processing |
| `chains[].pollInterval` | integer | No | 12000 | Milliseconds between polls |
| `chains[].maxBlockRange` | integer | No | 1000 | Maximum blocks to query at once |

#### AVS Section

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `avss[].address` | string | Yes | - | AVS contract address |
| `avss[].privateKey` | string | Yes | - | Private key for transactions |
| `avss[].privateSigningKey` | string | Yes | - | Private key for signing |
| `avss[].privateSigningKeyType` | string | Yes | - | Key type: ecdsa, bls |
| `avss[].responseTimeout` | integer | Yes | - | Response timeout in seconds |
| `avss[].chainIds` | array | Yes | - | Chain IDs to monitor |
| `avss[].operatorSets` | array | No | - | Operator set configurations |
| `avss[].taskConfig` | object | No | - | Task-specific settings |

#### Storage Section

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `storage.type` | string | No | memory | Storage backend type |
| `storage.badger.dir` | string | Yes* | - | Data directory (*if type is badger) |
| `storage.badger.valueLogFileSize` | integer | No | 1GB | Value log file size |
| `storage.badger.numVersionsToKeep` | integer | No | 1 | Number of versions to keep |
| `storage.badger.compactL0OnClose` | boolean | No | true | Compact level 0 on close |

### Environment Variables

The aggregator supports configuration via environment variables:

- `AGGREGATOR_CONFIG_PATH`: Path to configuration file
- `AGGREGATOR_LOG_LEVEL`: Override log level
- `AGGREGATOR_GRPC_PORT`: Override gRPC port
- `AGGREGATOR_METRICS_PORT`: Override metrics port
- `AGGREGATOR_STORAGE_DIR`: Override storage directory

### Security Considerations

1. **Private Key Management**:
   - Never commit private keys to version control
   - Use environment variables or key management systems
   - Rotate keys regularly

2. **TLS Configuration**:
   - Enable TLS for production gRPC servers
   - Use proper certificate rotation
   - Verify client certificates

3. **File Permissions**:
   ```bash
   chmod 600 aggregator-config.yaml
   chown ponos:ponos aggregator-config.yaml
   ```

4. **Network Security**:
   - Use private RPC endpoints
   - Enable rate limiting
   - Monitor for anomalous activity

### Validation

The aggregator validates configuration on startup:

- Chain RPC connectivity
- Contract existence at AVS addresses
- Private key validity
- Storage accessibility
- Port availability

Common validation errors:
- Invalid chain IDs
- Unreachable RPC endpoints
- Malformed Ethereum addresses
- Insufficient storage permissions
- Port conflicts

### Example Configurations

#### Minimal Configuration

```yaml
chains:
  - name: ethereum
    network: mainnet
    chainId: 1
    rpcUrl: "${ETH_RPC_URL}"

avss:
  - address: "0x1234567890123456789012345678901234567890"
    privateKey: "${AVS_PRIVATE_KEY}"
    privateSigningKey: "${AVS_SIGNING_KEY}"
    privateSigningKeyType: "ecdsa"
    responseTimeout: 3000
    chainIds: [1]
```

#### Multi-Chain Production Configuration

```yaml
chains:
  - name: ethereum
    network: mainnet
    chainId: 1
    rpcUrl: "${ETH_RPC_URL}"
    wsUrl: "${ETH_WS_URL}"
    blockConfirmations: 12
    pollInterval: 12000
  - name: base
    network: mainnet
    chainId: 8453
    rpcUrl: "${BASE_RPC_URL}"
    blockConfirmations: 6
    pollInterval: 2000
  - name: arbitrum
    network: mainnet
    chainId: 42161
    rpcUrl: "${ARB_RPC_URL}"
    blockConfirmations: 6
    pollInterval: 500

avss:
  - address: "${AVS_ADDRESS_1}"
    privateKey: "${AVS_PRIVATE_KEY_1}"
    privateSigningKey: "${AVS_SIGNING_KEY_1}"
    privateSigningKeyType: "bls"
    responseTimeout: 3600
    chainIds: [1, 8453, 42161]
    operatorSets:
      - id: 0
        minOperators: 5
        consensusThreshold: 7500  # 75%
        taskSLA: 3600
    taskConfig:
      maxRetries: 5
      retryDelay: 120
      batchSize: 50

storage:
  type: "badger"
  badger:
    dir: "/data/ponos/aggregator"
    valueLogFileSize: 4294967296  # 4GB
    numVersionsToKeep: 1
    numLevelZeroTables: 20
    numLevelZeroTablesStall: 40

grpc:
  enabled: true
  port: 50051
  tls:
    enabled: true
    certFile: "/etc/ponos/tls/server.crt"
    keyFile: "/etc/ponos/tls/server.key"

metrics:
  enabled: true
  port: 9091

logging:
  level: "info"
  format: "json"
  output: "file"
  file:
    path: "/var/log/ponos/aggregator.log"
    maxSize: 1000
    maxBackups: 10
    maxAge: 90
    compress: true
```

### Performance Tuning

1. **Chain Polling**:
   - Adjust `pollInterval` based on chain block time
   - Increase `maxBlockRange` for catching up
   - Use WebSocket connections for real-time updates

2. **Storage Optimization**:
   - Tune BadgerDB parameters for workload
   - Enable compression for large payloads
   - Regular garbage collection

3. **Resource Allocation**:
   - Allocate sufficient memory for in-flight tasks
   - Monitor goroutine count
   - Profile CPU usage during peak loads

### Monitoring and Alerts

Key metrics to monitor:
- Task processing latency
- Chain synchronization lag
- Operator response rates
- Storage usage growth
- gRPC connection count

Recommended alerts:
- Chain RPC failures
- Task timeout rates > 5%
- Storage usage > 80%
- Memory usage > 90%
- Goroutine leaks