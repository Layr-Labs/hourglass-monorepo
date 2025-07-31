# Ponos - Aggregation and Execution

## Development

```bash
# install deps
make deps

# run the test suite
make test

# build the protos
make proto

# build all binaries
make all

# lint the project
make lint
```

## Key generation

Ponos comes with a `keygen` cli utility to make generating keys easy for testing

```bash
go run ./cmd/keygen/*.go generate --curve-type bn254 --output-dir ../testKeys --use-keystore
```

## Configuration

### Aggregator Configuration

The aggregator requires a configuration file in YAML or JSON format with the following structure:

```yaml
debug: false
l1ChainId: 31337

operator:
  address: "0x..."
  operatorPrivateKey:
    privateKey: "0x..."  # or use remoteSigner
  signingKeys:
    bls:
      keystore: "..."  # or keystoreFile
      password: "..."

chains:
  - name: "L1"
    version: "1.0"
    chainId: 31337
    rpcUrl: "http://localhost:8545"
    pollIntervalSeconds: 5
  - name: "L2"
    version: "1.0"
    chainId: 31338
    rpcUrl: "http://localhost:8546"
    pollIntervalSeconds: 5

avss:
  - address: "0x..."
    chainIds: [31337, 31338]
    avsRegistrarAddress: "0x..."

# Storage configuration (optional)
storage:
  type: "memory"  # Options: "memory" or "badger"
  badger:
    dir: "/var/lib/ponos/aggregator/badger"
    # Optional tuning parameters:
    # valueLogFileSize: 1073741824  # 1GB
    # numVersionsToKeep: 1
    # numLevelZeroTables: 5
    # numLevelZeroTablesStall: 10

# Optional contract overrides
overrideContracts:
  taskMailbox:
    contract: "0x..."
    chainIds: [31337]
```

### Executor Configuration

The executor requires a configuration file in YAML or JSON format with the following structure:

```yaml
debug: false
grpcPort: 9090
performerNetworkName: "hourglass-network"

operator:
  address: "0x..."
  operatorPrivateKey:
    privateKey: "0x..."  # or use remoteSigner
  signingKeys:
    bls:
      keystore: "..."  # or keystoreFile
      password: "..."

l1Chain:
  rpcUrl: "http://localhost:8545"
  chainId: 31337

avsPerformers:
  - image:
      repository: "my-avs-performer"
      tag: "latest"
    processType: "docker"
    avsAddress: "0x..."
    avsRegistrarAddress: "0x..."
    envs:
      - name: "AVS_CONFIG"
        value: "production"
      - name: "LOG_LEVEL"
        valueFromEnv: "EXECUTOR_LOG_LEVEL"

# Storage configuration (optional)
storage:
  type: "memory"  # Options: "memory" or "badger"
  badger:
    dir: "/var/lib/ponos/executor/badger"

# Optional contract overrides
overrideContracts:
  taskMailbox:
    contract: "0x..."
    chainIds: [31337]
```

### Key Configuration Elements

#### OperatorConfig
- `address`: Ethereum address of the operator
- `operatorPrivateKey`: Private key for transaction signing (supports remote signer)
- `signingKeys`: BLS and ECDSA keys for task signing

#### SigningKeys
- `bls`: BLS signing key configuration
  - `keystore`: Direct keystore JSON string
  - `keystoreFile`: Path to keystore file
  - `password`: Keystore password

#### Remote Signer Support
For production deployments, use remote signers:
```yaml
operatorPrivateKey:
  remoteSigner: true
  remoteSignerConfig:
    url: "https://web3signer.example.com"
    fromAddress: "0x..."
    publicKey: "0x..."
    caCert: "..."
    cert: "..."
    key: "..."
```

#### Environment Variables
Executor configurations support environment variable forwarding:
```yaml
envs:
  - name: "CONFIG_VALUE"
    value: "direct-value"
  - name: "FORWARDED_VALUE"
    valueFromEnv: "HOST_ENV_VAR"
```

## Persistence and Storage

Ponos supports optional data persistence for both aggregator and executor components to enable crash recovery and high availability. By default, services run with in-memory storage, but production deployments should use persistent storage.

### Storage Backends

1. **Memory** (default) - All data is kept in memory and lost on restart
   - Good for development and testing
   - Zero configuration required
   - No external dependencies

2. **BadgerDB** - Embedded key-value store with on-disk persistence
   - Production-ready persistent storage
   - Crash recovery support
   - Efficient performance with built-in compression

### What Gets Persisted

#### Aggregator
- **Chain State**: Last processed block number for each monitored chain
- **Tasks**: All tasks with their current status (pending, processing, completed, failed)
- **Configurations**: Operator set configurations and AVS settings
- **Recovery**: On restart, resumes from last processed block and requeues pending tasks

#### Executor
- **Performer State**: Active AVS containers/deployments with their status
- **Inflight Tasks**: Currently processing tasks
- **Deployment History**: Track deployment lifecycle and failures
- **Recovery**: On restart, verifies container health and resumes task processing

### Configuration

Add the storage section to your configuration files:

```yaml
# For development - no persistence
storage:
  type: "memory"

# For production - with persistence
storage:
  type: "badger"
  badger:
    dir: "/var/lib/ponos/aggregator/badger"  # or executor/badger
    # Optional tuning (defaults are usually fine)
    inMemory: false                  # Set true for testing
    valueLogFileSize: 1073741824     # 1GB
    numVersionsToKeep: 1             # Only keep latest version
    numLevelZeroTables: 5
    numLevelZeroTablesStall: 10
```

### Directory Permissions

Ensure the storage directory has appropriate permissions:

```bash
# Create directories with proper permissions
sudo mkdir -p /var/lib/ponos/{aggregator,executor}/badger
sudo chown -R ponos:ponos /var/lib/ponos
sudo chmod 750 /var/lib/ponos/{aggregator,executor}/badger
```

### Docker Volumes

When running in Docker, mount persistent volumes:

```yaml
# docker-compose.yml
services:
  aggregator:
    volumes:
      - aggregator-data:/var/lib/ponos/aggregator/badger
  
  executor:
    volumes:
      - executor-data:/var/lib/ponos/executor/badger

volumes:
  aggregator-data:
  executor-data:
```

### Monitoring Storage

BadgerDB automatically runs garbage collection every 5 minutes. Monitor disk usage:

```bash
# Check storage size
du -sh /var/lib/ponos/*/badger

# Monitor growth over time
watch -n 60 'du -sh /var/lib/ponos/*/badger'
```

### Backup and Recovery

For production deployments:

1. **Backup**: Stop the service and copy the BadgerDB directory
2. **Restore**: Stop the service, replace the directory, restart
3. **Migration**: Export data from one backend, import to another (future feature)

### Performance Considerations

- BadgerDB is optimized for SSDs
- Memory storage is fastest but provides no durability
- BadgerDB adds minimal overhead (< 5% in benchmarks)
- Tune `valueLogFileSize` based on your workload

### Troubleshooting

1. **Storage errors on startup**: Check directory permissions and disk space
2. **Performance degradation**: Monitor garbage collection and consider tuning
3. **Data corruption**: BadgerDB has built-in checksums; corrupted data will be detected
4. **Migration issues**: Ensure compatible storage versions when upgrading
