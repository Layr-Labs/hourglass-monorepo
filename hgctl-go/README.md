# hgctl - Hourglass Control CLI 🚀

**A comprehensive CLI toolkit for deploying and managing Hourglass AVS (Actively Validated Services) and EigenLayer operator operations.**

`hgctl` streamlines AVS operations, enabling you to:
* Deploy and manage AVS components (aggregator, executor, performer)
* Register and manage EigenLayer operators
* Handle keystores and signing operations
* Manage operator allocations and delegations
* Configure multiple environments through contexts

> **Note:** This is the Go implementation of hgctl, providing native performance and enhanced features for production AVS deployments.

## 📦 Installation

### Quick Install (Recommended)
```bash
# Build from source
git clone https://github.com/Layr-Labs/hourglass-monorepo
cd hourglass-monorepo/hgctl-go
make install
```

### Manual Build
```bash
# Clone and build
git clone https://github.com/Layr-Labs/hourglass-monorepo
cd hourglass-monorepo/hgctl-go
make build

# Add to PATH
export PATH=$PATH:$(pwd)/bin
```

### Verify Installation
```bash
hgctl --version
hgctl --help
```

## 🌟 Key Commands Overview

| Command | Description |
|---------|-------------|
| **Context Management** |
| `hgctl context create` | Create a new context for environment configuration |
| `hgctl context use` | Switch to a different context |
| `hgctl context set` | Configure context properties |
| **AVS Deployment** |
| `hgctl deploy aggregator` | Deploy the aggregator component |
| `hgctl deploy executor` | Deploy the executor component |
| `hgctl deploy performer` | Deploy performer via executor gRPC |
| **Operator Management** |
| `hgctl register` | Register operator with EigenLayer |
| `hgctl delegate` | Delegate stake as an operator |
| `hgctl allocate` | Allocate stake to AVS operator sets |
| `hgctl register-avs` | Register operator with an AVS |
| `hgctl register-key` | Register signing keys with AVS |
| **Keystore Management** |
| `hgctl keystore create` | Create new BLS or ECDSA keystores |
| `hgctl keystore register` | Register existing keystore references |

---

## 🚦 Getting Started

### ✅ Prerequisites

* [Docker](https://docs.docker.com/engine/install/) (latest)
* [Go](https://go.dev/doc/install) (v1.21+)
* Access to Ethereum RPC endpoints (mainnet/testnet)
* Operator private keys or keystores

### 🚀 Quick Start - AVS Deployment

```bash
# 1. Create and configure a context
hgctl context create production
hgctl context use production
hgctl context set --rpc-url https://mainnet.infura.io/v3/YOUR-KEY
hgctl context set --avs-address 0xYourAVSAddress

# 2. Deploy AVS components
hgctl deploy aggregator --operator-set-id 0
hgctl deploy executor --operator-set-id 0
hgctl deploy performer --operator-set-id 0 --env-file performer.env
```

### 🔑 Quick Start - Operator Registration

```bash
# 1. Create or register keystores
hgctl keystore create --name my-operator --key-type ecdsa
hgctl keystore create --name my-operator-bls --key-type bn254

# 2. Register as EigenLayer operator
hgctl register --metadata-uri https://example.com/operator-metadata.json

# 3. Self-delegate
hgctl delegate

# 4. Register with AVS
hgctl register-avs --operator-set-ids 0 --socket https://operator.example.com:8080

# 5. Allocate stake
hgctl allocate --operator-set-id 0 --strategy 0xBeaC0eeEeeeeEEeEeEEEEeeEEeEeeeEeeEEBEaC0 --magnitude 1e18
```

---

## 🚧 Step-by-Step Guides

### 1️⃣ Context Setup

Contexts allow you to manage multiple environments (mainnet, testnet, local) with different configurations:

```bash
# Create a new context
hgctl context create mainnet

# Switch to the context
hgctl context use mainnet

# Configure essential addresses
hgctl context set --rpc-url https://mainnet.infura.io/v3/YOUR-KEY
hgctl context set --avs-address 0xYourAVSAddress
hgctl context set --operator-address 0xYourOperatorAddress
hgctl context set --delegation-manager 0xDelegationManagerAddress
hgctl context set --allocation-manager 0xAllocationManagerAddress

# View current context
hgctl context show
```

### 2️⃣ Keystore Management

Before operating, you need to set up your signing keys:

```bash
# Create a new ECDSA keystore
hgctl keystore create \
  --name operator-ecdsa \
  --key-type ecdsa

# Create a new BN254 keystore for BLS signatures
hgctl keystore create \
  --name operator-bls \
  --key-type bn254

# Register an existing keystore
hgctl keystore register \
  --name existing-key \
  --path /path/to/keystore.json \
  --key-type ecdsa

# List all keystores
hgctl keystore list
```

### 3️⃣ Operator Registration Flow

#### Step 1: Register with EigenLayer
```bash
hgctl register \
  --metadata-uri https://example.com/operator/metadata.json \
  --allocation-delay 86400  # 24 hours
```

#### Step 2: Self-Delegate
```bash
# Required after registration
hgctl delegate
```

#### Step 3: Register Keys with AVS
```bash
# Register ECDSA key
hgctl register-key \
  --operator-set-id 0 \
  --key-type ecdsa \
  --key-address 0xYourECDSAAddress

# Register BN254 key
hgctl register-key \
  --operator-set-id 0 \
  --key-type bn254 \
  --keystore-path /path/to/bn254.keystore \
  --password $KEYSTORE_PASSWORD
```

#### Step 4: Register with AVS
```bash
hgctl register-avs \
  --operator-set-ids 0,1 \
  --socket https://operator.example.com:8080
```

#### Step 5: Allocate Stake
```bash
hgctl allocate \
  --operator-set-id 0 \
  --strategy 0xBeaC0eeEeeeeEEeEeEEEEeeEEeEeeeEeeEEBEaC0 \
  --magnitude 1e18  # Full allocation
```

### 4️⃣ AVS Component Deployment

Deploy AVS components using runtime specifications from OCI registries:

#### Deploy Aggregator
```bash
# Basic deployment
hgctl deploy aggregator --operator-set-id 0

# With custom environment
hgctl deploy aggregator \
  --operator-set-id 0 \
  --env L1_RPC_URL=https://mainnet.infura.io/v3/KEY \
  --env-file aggregator-secrets.env
```

#### Deploy Executor
```bash
# Deploy executor (hosts gRPC API for performers)
hgctl deploy executor --operator-set-id 0
```

#### Deploy Performer
```bash
# Deploy via executor gRPC (executor must be running)
hgctl deploy performer \
  --operator-set-id 0 \
  --env DATABASE_URL=postgres://localhost/avs_db \
  --env API_KEY=$API_KEY
```

### 5️⃣ Environment Configuration

Environment variables are loaded in priority order:
1. Command-line flags (`--env KEY=VALUE`)
2. Environment file (`--env-file path/to/file`)
3. Context environment variables
4. OS environment variables
5. Context secrets file (highest priority)

```bash
# Create environment file for secrets
cat > secrets.env <<EOF
OPERATOR_PRIVATE_KEY=0x...
KEYSTORE_PASSWORD=...
DATABASE_URL=postgres://...
EOF

# Deploy with environment configuration
hgctl deploy aggregator \
  --operator-set-id 0 \
  --env-file secrets.env \
  --env L1_CHAIN_ID=1
```

---

## 📋 Command Reference

### Global Options
```bash
--verbose, -v              Enable verbose logging
--output, -o <format>      Output format (table|json|yaml)
--help, -h                 Show help
```

### Context Commands
```bash
# Context management
hgctl context create <name>        # Create new context
hgctl context list                 # List all contexts
hgctl context use <name>           # Switch context
hgctl context set [options]        # Configure context
hgctl context show                 # Display current context
```

### Deployment Commands
```bash
# Component deployment
hgctl deploy aggregator [options]  # Deploy aggregator
hgctl deploy executor [options]    # Deploy executor
hgctl deploy performer [options]   # Deploy performer via gRPC

# Common options:
--operator-set-id <id>            # Operator set ID
--release-id <id>                 # Specific release ID
--env KEY=VALUE                   # Set environment variable
--env-file <path>                 # Load environment from file
--dry-run                         # Validate without deploying
```

### Operator Commands
```bash
# Registration and management
hgctl register [options]           # Register with EigenLayer
hgctl delegate [options]           # Delegate stake
hgctl register-avs [options]       # Register with AVS
hgctl register-key [options]       # Register signing keys
hgctl allocate [options]          # Allocate to operator sets
hgctl set-allocation-delay        # Configure allocation delay
hgctl deposit [options]           # Deposit into strategies
```

### Resource Commands
```bash
# Query resources
hgctl get performer               # List performers
hgctl get release                 # List releases
hgctl get operator-set           # List operator sets

# Describe resources
hgctl describe release [options]  # Release details
hgctl describe operator-set      # Operator set details
```

---

## 🔧 Configuration

### Directory Structure
```
~/.hgctl/
├── config.yaml                   # Global configuration
└── contexts/
    ├── mainnet/                  # Context-specific files
    │   ├── aggregator.yaml       # Component configs
    │   └── keystores/            # Keystore references
    └── testnet/
```

### Context Configuration
```yaml
# ~/.hgctl/config.yaml
currentContext: mainnet
contexts:
  mainnet:
    rpcUrl: https://mainnet.infura.io/v3/KEY
    avsAddress: "0x..."
    operatorAddress: "0x..."
    delegationManager: "0x..."
    allocationManager: "0x..."
    environmentVars:
      L1_CHAIN_ID: "1"
      AGGREGATOR_PORT: "9000"
    keystores:
      - name: operator-ecdsa
        type: ecdsa
        path: /secure/path/keystore.json
```

### Security Best Practices

1. **Never store private keys in configuration files**
2. **Use environment files for secrets** (`--env-file`)
3. **Store keystores in secure locations** with restricted permissions
4. **Use context secrets files** for sensitive environment variables
5. **Enable Web3 Signer** for production deployments

---

## 🛠 Development

### Building from Source
```bash
# Clone repository
git clone https://github.com/Layr-Labs/hourglass-monorepo
cd hourglass-monorepo/hgctl-go

# Build
make build

# Run tests
make test

# Run integration tests
make test-integration

# Install locally
make install
```

### Project Structure
```
hgctl-go/
├── cmd/hgctl/              # CLI entry point
├── internal/
│   ├── commands/           # Command implementations
│   │   ├── deploy/         # Deployment commands
│   │   ├── operator/       # Operator management
│   │   └── middleware/     # Shared middleware
│   ├── client/             # API clients (contract, OCI, executor)
│   ├── config/             # Configuration management
│   ├── keystore/           # Keystore operations
│   ├── signer/             # Signing implementations
│   └── templates/          # Config templates
├── Makefile
└── go.mod
```

---

## 🐛 Troubleshooting

### Common Issues

**"Required addresses not configured"**
```bash
# Ensure all required addresses are set
hgctl context show
hgctl context set --delegation-manager 0x...
```

**"Executor not available"**
```bash
# Check executor is running
docker ps | grep executor
docker logs hgctl-executor-<avs-address>
```

**"Transaction failed"**
```bash
# Enable verbose mode for transaction details
hgctl register --verbose

# Check operator balance
cast balance $OPERATOR_ADDRESS
```

### Debug Mode
```bash
# Enable verbose logging
export HGCTL_LOG_LEVEL=debug
hgctl --verbose <command>

# View transaction details
hgctl --verbose allocate --operator-set-id 0 --strategy 0x...
```

---

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

## 🙋 Support

For questions and support:
- Open an issue in the [GitHub repository](https://github.com/Layr-Labs/hourglass-monorepo/issues)
- Join our [Discord community](https://discord.gg/eigenlayer)
- Check the [documentation](https://docs.eigenlayer.xyz)

---

Made with ❤️ by the EigenLayer team