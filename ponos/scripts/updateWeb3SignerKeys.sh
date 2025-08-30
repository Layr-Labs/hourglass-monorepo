#!/usr/bin/env bash

set -e

echo "Updating Web3Signer keys from generated test data..."

# Read the chain config
CHAIN_CONFIG_FILE="./internal/testData/chain-config.json"

if [ ! -f "$CHAIN_CONFIG_FILE" ]; then
    echo "Error: chain-config.json not found. Please run generateTestChainState.sh first."
    exit 1
fi

# Extract keys using jq
OPERATOR_PK=$(jq -r '.operatorAccountPk' $CHAIN_CONFIG_FILE)
EXEC_OPERATOR_PK=$(jq -r '.execOperatorAccountPk' $CHAIN_CONFIG_FILE)
BN254_EXECUTOR_PK=$(jq -r '.bn254ExecutorAccountPk' $CHAIN_CONFIG_FILE)

# Web3Signer directories
WEB3SIGNER_DIR="${HOME}/.web3signer"
KEYS_DIR="${WEB3SIGNER_DIR}/key-files"

# Create directories if they don't exist
mkdir -p "$KEYS_DIR"

# Function to create BLS key file
create_bls_key_file() {
    local name=$1
    local private_key=$2
    local file_path="${KEYS_DIR}/${name}-bls.yaml"
    
    cat > "$file_path" << EOF
type: "file-raw"
keyType: "BLS"
privateKey: "0x${private_key}"
EOF
    echo "Created BLS key file: $file_path"
}

# Function to create ECDSA key file
create_ecdsa_key_file() {
    local name=$1
    local private_key=$2
    local file_path="${KEYS_DIR}/${name}-ecdsa.json"
    
    cat > "$file_path" << EOF
{
  "type": "file-raw",
  "keyType": "SECP256K1",
  "privateKey": "0x${private_key}"
}
EOF
    echo "Created ECDSA key file: $file_path"
}

# Create key files for each operator
echo "Creating Web3Signer key files..."

# Aggregator (BN254/BLS on operator set 0)
create_bls_key_file "aggregator" "$OPERATOR_PK"

# Executor (ECDSA on operator set 1)
create_ecdsa_key_file "executor" "$EXEC_OPERATOR_PK"

# BN254 Executor (BN254/BLS on operator set 2)
create_bls_key_file "bn254-executor" "$BN254_EXECUTOR_PK"

echo ""
echo "Web3Signer keys updated successfully!"
echo "Key files created in: $KEYS_DIR"
echo ""
echo "To start Web3Signer with these keys:"
echo "  web3signer --key-store-path=$KEYS_DIR eth2"
echo ""
echo "Or if using Docker:"
echo "  docker run -v $KEYS_DIR:/keys consensys/web3signer:latest --key-store-path=/keys eth2"