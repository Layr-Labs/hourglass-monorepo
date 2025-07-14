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
