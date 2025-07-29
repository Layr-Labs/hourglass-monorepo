# hgctl Configuration Templates

This document describes how hgctl generates configuration files for Hourglass aggregator and executor components.

## Overview

hgctl automatically generates YAML configuration files from templates when deploying aggregator or executor components. The templates use environment variable substitution to customize the configuration based on your deployment context.

## Directory Structure

Configuration files and secrets are organized per context in the `.hgctl` directory:

```
~/.hgctl/
├── {context}/                    # e.g., mainnet, testnet, devnet
│   ├── config.env               # Environment variables for this context
│   ├── operator.bls.keystore.json       # BLS keystore file
│   ├── operator.ecdsa.keystore.json     # ECDSA keystore file (optional)
│   ├── web3signer-bls-ca.crt           # Web3 signer CA cert (optional)
│   ├── web3signer-bls-client.crt       # Web3 signer client cert (optional)
│   ├── web3signer-bls-client.key       # Web3 signer client key (optional)
│   ├── web3signer-ecdsa-ca.crt         # Web3 signer CA cert for ECDSA (optional)
│   ├── web3signer-ecdsa-client.crt     # Web3 signer client cert for ECDSA (optional)
│   └── web3signer-ecdsa-client.key     # Web3 signer client key for ECDSA (optional)
```

## Environment Variables

### Common Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `OPERATOR_ADDRESS` | Operator Ethereum address | Yes | - |
| `OPERATOR_PRIVATE_KEY` | Operator private key for transactions | No | - |
| `AVS_ADDRESS` | AVS contract address | Yes | - |
| `L1_CHAIN_ID` | L1 chain ID | Yes | - |
| `L1_RPC_URL` | L1 RPC endpoint URL | Yes | - |

### BLS Signing Configuration

**Option 1: Local Keystore**
| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `BLS_KEYSTORE_PASSWORD` | Password for BLS keystore | Yes | - |

**Option 2: Web3 Signer**
| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `USE_WEB3_SIGNER_BLS` | Enable Web3 signer for BLS | No | false |
| `WEB3_SIGNER_BLS_URL` | Web3 signer service URL | Yes* | - |
| `WEB3_SIGNER_BLS_FROM_ADDRESS` | Address to sign from | Yes* | - |
| `WEB3_SIGNER_BLS_PUBLIC_KEY` | Public key | Yes* | - |

*Required when `USE_WEB3_SIGNER_BLS=true`

### ECDSA Signing Configuration

**Option 1: Local Private Key**
| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `ECDSA_PRIVATE_KEY` | ECDSA private key | No | - |

**Option 2: Web3 Signer**
| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `USE_WEB3_SIGNER_ECDSA` | Enable Web3 signer for ECDSA | No | false |
| `WEB3_SIGNER_ECDSA_URL` | Web3 signer service URL | Yes* | - |
| `WEB3_SIGNER_ECDSA_FROM_ADDRESS` | Address to sign from | Yes* | - |
| `WEB3_SIGNER_ECDSA_PUBLIC_KEY` | Public key | Yes* | - |

*Required when `USE_WEB3_SIGNER_ECDSA=true`

### Aggregator-Specific Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `AGGREGATOR_PORT` | gRPC server port | No | 9000 |
| `DEBUG` | Enable debug mode | No | false |
| `L2_CHAIN_ID` | L2 chain ID | No | - |
| `L2_RPC_URL` | L2 RPC endpoint URL | No* | - |
| `RESPONSE_TIMEOUT` | Task response timeout (ms) | No | 3000 |
| `AVS_REGISTRAR_ADDRESS` | AVS registrar contract | No | - |

*Required if `L2_CHAIN_ID` is set

### Executor-Specific Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `EXECUTOR_PORT` | gRPC server port | No | 9090 |
| `PERFORMER_NETWORK_NAME` | Docker network name | No | hgctl-performer-network |
| `PERFORMER_REGISTRY` | Docker registry for performer | Yes | - |
| `PERFORMER_TAG` | Docker image tag | No | latest |
| `WORKER_COUNT` | Number of parallel workers | No | 1 |
| `SIGNING_CURVE` | Signing curve (bn254/bls381) | No | bn254 |

## Setting Up Configuration

### 1. Create Context Directory

```bash
mkdir -p ~/.hgctl/mainnet
```

### 2. Add Keystore Files

Copy your BLS keystore to the context directory:
```bash
cp operator.bls.keystore.json ~/.hgctl/mainnet/
```

### 3. Create Environment File

Create `~/.hgctl/mainnet/config.env`:
```bash
# Operator configuration
OPERATOR_ADDRESS=0x1234...
BLS_KEYSTORE_PASSWORD=your-keystore-password

# Network configuration
L1_CHAIN_ID=1
L1_RPC_URL=https://eth-mainnet.example.com
L2_CHAIN_ID=8453
L2_RPC_URL=https://base-mainnet.example.com

# AVS configuration
AVS_ADDRESS=0xABC...
AVS_REGISTRAR_ADDRESS=0xDEF...
PERFORMER_REGISTRY=registry.example.com/my-avs/performer
```

### 4. Deploy Components

The deploy commands will automatically use the configuration from your context:

```bash
# Deploy aggregator
hgctl deploy aggregator 0xAVS... --operator-set-id 1

# Deploy executor  
hgctl deploy executor 0xAVS... --operator-set-id 1
```

## Web3 Signer Setup

To use Web3 signer instead of local keystores:

### 1. Add TLS Certificates (if required)

```bash
# For BLS signing
cp web3signer-ca.crt ~/.hgctl/mainnet/web3signer-bls-ca.crt
cp client.crt ~/.hgctl/mainnet/web3signer-bls-client.crt
cp client.key ~/.hgctl/mainnet/web3signer-bls-client.key
```

### 2. Configure Environment

Add to `config.env`:
```bash
# Enable Web3 signer for BLS
USE_WEB3_SIGNER_BLS=true
WEB3_SIGNER_BLS_URL=https://web3signer.example.com:9000
WEB3_SIGNER_BLS_FROM_ADDRESS=0x1234...
WEB3_SIGNER_BLS_PUBLIC_KEY=0xabcd...

# Optional: Enable Web3 signer for ECDSA
USE_WEB3_SIGNER_ECDSA=true
WEB3_SIGNER_ECDSA_URL=https://web3signer.example.com:9000
WEB3_SIGNER_ECDSA_FROM_ADDRESS=0x5678...
WEB3_SIGNER_ECDSA_PUBLIC_KEY=0xefgh...
```

## Security Best Practices

1. **Never commit secrets**: Keep your `.hgctl` directory in `.gitignore`
2. **File permissions**: Keystore and key files are automatically set to 0600 (owner read/write only)
3. **Use Web3 signer**: For production, consider using Web3 signer instead of local private keys
4. **Separate contexts**: Use different contexts for mainnet, testnet, and development
5. **Environment isolation**: Each context has its own isolated configuration

## Troubleshooting

### Missing Configuration
If you see "failed to generate config", ensure:
- Required environment variables are set
- Context directory exists (`~/.hgctl/{context}/`)
- Keystore files are present (if not using Web3 signer)

### Invalid Templates
The templates expect specific environment variables. Check:
- Variable names match exactly (case-sensitive)
- No typos in variable names
- Values are properly formatted (addresses with 0x prefix, etc.)

### Container Fails to Start
Check the generated configuration:
```bash
# Find the container
docker ps -a | grep hgctl-

# Check its configuration mount
docker inspect <container-id> | grep -A5 Mounts

# View the generated config
docker exec <container-id> cat /config/aggregator.yaml
```