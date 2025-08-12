# hgctl - Hourglass Control CLI 🚀

**A comprehensive CLI toolkit for deploying and managing Hourglass AVS (Actively Validated Services) and EigenLayer operator operations.**

`hgctl` streamlines AVS operations, enabling you to:
* Deploy and manage AVS components (aggregator, executor, performer) using EigenRuntime specifications
* Register and manage EigenLayer operators with full lifecycle support
* Handle keystores and signing operations (BLS/ECDSA, Web3Signer integration)
* Manage operator allocations, delegations, and deposits
* Configure multiple environments through contexts with hierarchical configuration
* Fetch and deploy AVS releases from OCI registries via ReleaseManager contracts

> **Note:** This is the Go implementation of hgctl, providing native performance, OCI artifact management via ORAS, and enhanced features for production AVS deployments.

## 📦 Installation

### Quick Install (Recommended)

Install the latest release binary directly:

```bash
# Install using the installation script
curl -fsSL https://raw.githubusercontent.com/Layr-Labs/hourglass-monorepo/master/hgctl-go/install-hgctl.sh | bash
```

The installer will:
- Detect your operating system and architecture
- Download the appropriate binary from GitHub releases
- Install to `$HOME/bin` (or location of your choice)
- Guide you through adding to PATH if needed

### Download Binary Manually

Download pre-built binaries from [GitHub Releases](https://github.com/Layr-Labs/hourglass-monorepo/releases):

```bash
# Example for macOS ARM64
curl -LO https://github.com/Layr-Labs/hourglass-monorepo/releases/download/hgctl-v0.1.0.preview-rc.1/hgctl-darwin-arm64-v0.1.0.preview-rc.1.tar.gz
tar -xzf hgctl-darwin-arm64-v0.1.0.preview-rc.1.tar.gz
chmod +x hgctl
sudo mv hgctl /usr/local/bin/

# Example for Linux AMD64
curl -LO https://github.com/Layr-Labs/hourglass-monorepo/releases/download/hgctl-v0.1.0.preview-rc.1/hgctl-linux-amd64-v0.1.0.preview-rc.1.tar.gz
tar -xzf hgctl-linux-amd64-v0.1.0.preview-rc.1.tar.gz
chmod +x hgctl
sudo mv hgctl /usr/local/bin/
```

Available platforms:
- `darwin-amd64` - macOS Intel
- `darwin-arm64` - macOS Apple Silicon (M1/M2/M3)
- `linux-amd64` - Linux x86_64
- `linux-arm64` - Linux ARM64

### Build from Source

```bash
# Clone and build
git clone https://github.com/Layr-Labs/hourglass-monorepo
cd hourglass-monorepo/hgctl-go
make install  # Installs to ~/bin

# Or build only
make build    # Binary will be in ./bin/hgctl
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
| `hgctl context list` | List all available contexts |
| `hgctl context use` | Switch to a different context |
| `hgctl context set` | Configure context properties (RPC URLs, addresses, env vars) |
| `hgctl context show` | Display current context configuration |
| `hgctl context copy` | Copy an existing context |
| `hgctl context remove` | Remove a context |
| **AVS Deployment** |
| `hgctl deploy aggregator` | Deploy the aggregator component from OCI registry |
| `hgctl deploy executor` | Deploy the executor component from OCI registry |
| `hgctl deploy performer` | Deploy performer via executor gRPC interface |
| **Release Management** |
| `hgctl get release` | List available releases for an AVS |
| `hgctl describe release` | Get detailed information about a specific release |
| `hgctl get operator-set` | List operator sets for an AVS |
| `hgctl describe operator-set` | Get detailed operator set information |
| **Operator Management** |
| `hgctl el register-operator` | Register operator with EigenLayer |
| `hgctl el delegate` | Self-delegate stake as an operator |
| `hgctl el allocate` | Allocate stake to AVS operator sets |
| `hgctl el set-allocation-delay` | Configure allocation delay period |
| `hgctl el deposit` | Deposit tokens into EigenLayer strategies |
| `hgctl el register-avs` | Register operator with an AVS |
| `hgctl el register-key` | Register signing keys (BLS/ECDSA) with AVS |
| **Keystore Management** |
| `hgctl keystore create` | Create new BLS or ECDSA keystores |
| `hgctl keystore import` | Import existing keystore files |
| `hgctl keystore list` | List all registered keystores |
| `hgctl keystore show` | Display keystore details and export private key |
| **Signer Configuration** |
| `hgctl signer operator` | Configure operator signing keys |
| `hgctl signer system` | Configure system signing keys |
| **Performer Management** |
| `hgctl get performer` | List deployed performers |
| `hgctl remove performer` | Remove a deployed performer |

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
# 1. Create or import keystores
hgctl keystore create --name my-operator --type ecdsa
hgctl keystore create --name my-operator-bls --type bn254

# 2. Configure signing keys
hgctl signer operator keystore --keystore-name my-operator
hgctl signer system keystore --keystore-name my-operator-bls --type bn254

# 3. Register as EigenLayer operator
hgctl el register-operator --metadata-uri https://example.com/operator-metadata.json --allocation-delay 0

# 4. Self-delegate
hgctl el delegate

# 5. Deposit into strategies
hgctl el deposit --strategy 0x424246eF71b01ee33aA33aC590fd9a0855F5eFbc --token-address 0x7b79995e5f793A07Bc00c21412e50Ecae098E7f9 --amount '0.00001 ether'

# 6. Register keys with AVS
hgctl el register-key --operator-set-id 0 --key-type bn254 --keystore-path ~/.hgctl/sepolia/keystores/my-operator-bls/key.json --password test

# 7. Register with AVS
hgctl el register-avs --operator-set-ids 0 --socket https://operator.example.com:8080

# 8. Allocate stake (after configuration delay)
hgctl el allocate --operator-set-id 0 --strategy 0x424246eF71b01ee33aA33aC590fd9a0855F5eFbc --magnitude 1e13
```

---

## 🚧 Step-by-Step Guides

### 1️⃣ Context Setup

Contexts allow you to manage multiple environments (mainnet, testnet, local) with different configurations:

```bash
# Create a new context interactively
hgctl context create sepolia

# Or create with flags
hgctl context create mainnet \
  --l1-rpc-url https://mainnet.infura.io/v3/YOUR-KEY \
  --l2-rpc-url https://base.infura.io/v3/YOUR-KEY \
  --avs-address 0xYourAVSAddress \
  --release-manager 0xReleaseManagerAddress

# Switch to the context
hgctl context use mainnet

# Configure context properties
hgctl context set --l1-rpc-url https://mainnet.infura.io/v3/YOUR-KEY
hgctl context set --l2-rpc-url https://base.infura.io/v3/YOUR-KEY
hgctl context set --avs-address 0xYourAVSAddress
hgctl context set --operator-address 0xYourOperatorAddress
hgctl context set --operator-set-id 0

# Set environment variables (non-secret)
hgctl context set --env L1_CHAIN_ID=1 --env L2_CHAIN_ID=8453

# View current context
hgctl context show
```

### 2️⃣ Keystore Management

Before operating, you need to set up your signing keys:

```bash
# Create a new ECDSA keystore
hgctl keystore create \
  --name operator-ecdsa \
  --type ecdsa

# Create a new BN254 keystore for BLS signatures
hgctl keystore create \
  --name operator-bls \
  --type bn254

# Import an existing keystore
hgctl keystore import \
  --name existing-key \
  --path /path/to/keystore.json \
  --type ecdsa

# List all keystores
hgctl keystore list

# Show keystore details and export private key
hgctl keystore show --name operator-ecdsa
```

### 3️⃣ Operator Registration Flow

#### Step 1: Register with EigenLayer
```bash
hgctl el register-operator \
  --metadata-uri https://example.com/operator/metadata.json \
  --allocation-delay 0  # 0 for testnet, 86400 for mainnet (24 hours)
```

#### Step 2: Self-Delegate
```bash
# Required after registration
hgctl el delegate
```

#### Step 3: Deposit into Strategies
```bash
# Deposit WETH into a strategy
hgctl el deposit \
  --strategy 0x424246eF71b01ee33aA33aC590fd9a0855F5eFbc \
  --token-address 0x7b79995e5f793A07Bc00c21412e50Ecae098E7f9 \
  --amount '0.00001 ether'
```

#### Step 4: Register Keys with AVS
```bash
# Register BN254 key for operator set
hgctl el register-key \
  --operator-set-id 0 \
  --key-type bn254 \
  --keystore-path ~/.hgctl/sepolia/keystores/operator-bls/key.json \
  --password $KEYSTORE_PASSWORD

# Register ECDSA key (if required)
hgctl el register-key \
  --operator-set-id 0 \
  --key-type ecdsa \
  --key-address 0xYourECDSAAddress
```

#### Step 5: Register with AVS
```bash
hgctl el register-avs \
  --operator-set-ids 0,1 \
  --socket https://operator.example.com:8080
```

#### Step 6: Allocate Stake
```bash
# Wait for allocation delay period if configured
# Then allocate stake to operator set
hgctl el allocate \
  --operator-set-id 0 \
  --strategy 0x424246eF71b01ee33aA33aC590fd9a0855F5eFbc \
  --magnitude 1e13  # Allocation amount
```

### 4️⃣ AVS Component Deployment

Deploy AVS components using runtime specifications from OCI registries:

#### List Available Releases
```bash
# List all releases for your AVS
hgctl get release

# Get detailed release information
hgctl describe release --release-id 0 --operator-set-id 0
```

#### Deploy Aggregator
```bash
# Deploy latest aggregator release
hgctl deploy aggregator --operator-set-id 0

# Deploy specific release version
hgctl deploy aggregator --operator-set-id 0 --release-id 1

# With custom environment variables
hgctl deploy aggregator \
  --operator-set-id 0 \
  --env L1_RPC_URL=https://mainnet.infura.io/v3/KEY \
  --env L2_RPC_URL=https://base.infura.io/v3/KEY \
  --env-file aggregator-secrets.env
```

#### Deploy Executor
```bash
# Deploy executor (hosts gRPC API for performers)
hgctl deploy executor --operator-set-id 0

# With specific release and environment
hgctl deploy executor \
  --operator-set-id 0 \
  --release-id 1 \
  --env-file executor-config.env
```

#### Deploy Performer
```bash
# Deploy via executor gRPC (executor must be running)
hgctl deploy performer \
  --operator-set-id 0 \
  --env DATABASE_URL=postgres://localhost/avs_db \
  --env API_KEY=$API_KEY

# List deployed performers
hgctl get performer

# Remove a performer
hgctl remove performer --id performer-123
```

### 5️⃣ Environment Configuration

Environment variables are loaded in priority order (later sources override earlier):
1. Context configuration defaults
2. OS environment variables
3. Context environment variables (`hgctl context set --env`)
4. Environment file (`--env-file path/to/file`)
5. Command-line flags (`--env KEY=VALUE`)
6. Context secrets file (highest priority, set via context)

```bash
# Set context-wide environment variables
hgctl context set --env L1_CHAIN_ID=1 --env L2_CHAIN_ID=8453

# Create environment file for secrets
cat > secrets.env <<EOF
OPERATOR_PRIVATE_KEY=0x...
SYSTEM_PRIVATE_KEY=0x...
KEYSTORE_PASSWORD=...
DATABASE_URL=postgres://...
EOF

# Set secrets file in context (loaded with highest priority)
hgctl context set --env-secrets-path ~/secrets.env

# Deploy with additional environment overrides
hgctl deploy aggregator \
  --operator-set-id 0 \
  --env-file additional-config.env \
  --env DEBUG=true
```

#### Automatic Environment Variables

The following variables are automatically populated from context:
- `OPERATOR_ADDRESS` - From context operator address
- `OPERATOR_PRIVATE_KEY` - From configured operator signer
- `L1_CHAIN_ID` / `L2_CHAIN_ID` - From context chain IDs
- `L1_RPC_URL` / `L2_RPC_URL` - From context RPC URLs (auto-translated for Docker)
- `SYSTEM_*_KEYSTORE_PATH` - From system signer configuration
- `SYSTEM_WEB3SIGNER_*` - From Web3Signer configuration

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
hgctl context create <name>        # Create new context (interactive or with flags)
hgctl context list                 # List all contexts
hgctl context use <name>           # Switch context
hgctl context set [options]        # Configure context properties
hgctl context show                 # Display current context
hgctl context copy <src> <dst>     # Copy an existing context
hgctl context remove <name>        # Remove a context

# Context configuration options:
--l1-rpc-url <url>                 # L1 RPC endpoint
--l2-rpc-url <url>                 # L2 RPC endpoint
--avs-address <address>            # AVS contract address
--operator-address <address>       # Operator address
--operator-set-id <id>             # Default operator set ID
--release-manager <address>        # Release manager contract
--env KEY=VALUE                    # Set environment variables
--env-secrets-path <path>          # Path to secrets file
```

### Deployment Commands
```bash
# Component deployment
hgctl deploy aggregator [options]  # Deploy aggregator from OCI registry
hgctl deploy executor [options]    # Deploy executor from OCI registry
hgctl deploy performer [options]   # Deploy performer via executor gRPC

# Common deployment options:
--operator-set-id <id>            # Operator set ID (required)
--release-id <id>                 # Specific release ID (optional, uses latest if not specified)
--env KEY=VALUE                   # Set environment variable
--env-file <path>                 # Load environment from file

# Release management
hgctl get release                 # List available releases
hgctl describe release [options]  # Get detailed release info
hgctl get operator-set           # List operator sets
hgctl describe operator-set      # Get operator set details

# Performer management
hgctl get performer               # List deployed performers
hgctl remove performer --id <id>  # Remove a performer
```

### EigenLayer Commands
```bash
# All EigenLayer commands use the 'el' prefix or 'eigenlayer' full name

# Operator registration and management
hgctl el register-operator [options]  # Register with EigenLayer
  --metadata-uri <uri>                # Operator metadata URI
  --allocation-delay <seconds>        # Allocation delay period

hgctl el delegate [options]           # Self-delegate stake

hgctl el deposit [options]            # Deposit into strategies
  --strategy <address>                # Strategy contract address
  --token-address <address>           # Token to deposit
  --amount <amount>                    # Amount (e.g., '1 ether')

# AVS registration
hgctl el register-avs [options]       # Register with AVS
  --operator-set-ids <ids>            # Comma-separated operator set IDs
  --socket <url>                      # Operator socket URL

hgctl el register-key [options]       # Register signing keys
  --operator-set-id <id>              # Operator set ID
  --key-type <bn254|ecdsa>           # Key type
  --keystore-path <path>              # Path to keystore file
  --password <password>               # Keystore password

# Stake allocation
hgctl el allocate [options]          # Allocate to operator sets
  --operator-set-id <id>              # Operator set ID
  --strategy <address>                # Strategy address
  --magnitude <amount>                # Allocation magnitude

hgctl el set-allocation-delay        # Configure allocation delay
  --delay <seconds>                   # Delay in seconds
```

### Keystore Commands
```bash
# Keystore management
hgctl keystore create [options]      # Create new keystore
  --name <name>                      # Keystore name
  --type <ecdsa|bn254>               # Key type
  --key <private-key>                # Optional: use existing key

hgctl keystore import [options]      # Import existing keystore
  --name <name>                      # Keystore name
  --path <path>                      # Path to keystore file
  --type <ecdsa|bn254>               # Key type

hgctl keystore list                  # List all keystores
hgctl keystore show --name <name>    # Show keystore details
```

### Signer Commands
```bash
# Configure signing keys for operations
hgctl signer operator [options]      # Configure operator signing
  keystore --keystore-name <name>    # Use keystore for signing
  privatekey --key <key>              # Use private key directly

hgctl signer system [options]        # Configure system signing
  keystore --keystore-name <name> --type <bn254|ecdsa>
  privatekey --key <key>
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