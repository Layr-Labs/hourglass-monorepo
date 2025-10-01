#!/usr/bin/env bash

# Script to generate test chain state for hgctl integration tests
# Based on ponos generateTestChainState.sh but includes AVS setup

anvilL1Pid=""
anvilL2Pid=""

function cleanup() {
    echo "Cleaning up..."
    if [ ! -z "$anvilL1Pid" ]; then
        kill $anvilL1Pid 2>/dev/null || true
    fi
    if [ ! -z "$anvilL2Pid" ]; then
        kill $anvilL2Pid 2>/dev/null || true
    fi
    exit $?
}
trap cleanup EXIT ERR INT TERM

set -ex

# Navigate to the project root (where contracts directory exists)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Script is in hgctl-go/internal/testutils/scripts
HGCTL_ROOT="$SCRIPT_DIR/../../.."  # Go up to hgctl-go root
PROJECT_ROOT="$HGCTL_ROOT/.."       # Go up to hourglass-monorepo root

echo "Script dir: $SCRIPT_DIR"
echo "HGCTL root: $HGCTL_ROOT"
echo "Project root: $PROJECT_ROOT"

# Check for existing anvil processes
echo "Checking for existing anvil processes..."
if lsof -Pi :8545 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "ERROR: Port 8545 is already in use. Please kill the process using it."
    echo "You can run: lsof -i :8545"
    echo "Then kill the process with: kill -9 <PID>"
    exit 1
fi

if lsof -Pi :9545 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "ERROR: Port 9545 is already in use. Please kill the process using it."
    echo "You can run: lsof -i :9545"
    echo "Then kill the process with: kill -9 <PID>"
    exit 1
fi

# Build hgctl binary first
echo "Building hgctl binary..."
cd "$HGCTL_ROOT"
make build
cd "$PROJECT_ROOT"

# Path to hgctl binary
HGCTL="$HGCTL_ROOT/bin/hgctl"
KEYS_DIR="$HGCTL_ROOT/internal/testutils/keys"

# Create keys directory
mkdir -p "$KEYS_DIR"

# ethereum mainnet
L1_FORK_RPC_URL=https://late-crimson-dew.quiknode.pro/56c000eadf175378343de407c56e0ccd62801fe9

anvilL1ChainId=31337
anvilL1StartBlock=23477799
anvilL1DumpStatePath=$HGCTL_ROOT/internal/testdata/anvil-l1-state.json
anvilL1ConfigPath=$HGCTL_ROOT/internal/testdata/anvil-l1-config.json
anvilL1RpcPort=8545
anvilL1RpcUrl="http://localhost:${anvilL1RpcPort}"

# base mainnet
L2_FORK_RPC_URL=https://still-attentive-slug.base-mainnet.quiknode.pro/91bfa66d45c9f3ac7ef9e9ca35b2acc8ba41160a/

anvilL2ChainId=31338
anvilL2StartBlock=36235532
anvilL2DumpStatePath=$HGCTL_ROOT/internal/testdata/anvil-l2-state.json
anvilL2ConfigPath=$HGCTL_ROOT/internal/testdata/anvil-l2-config.json
anvilL2RpcPort=9545
anvilL2RpcUrl="http://localhost:${anvilL2RpcPort}"

# Create testdata directory if it doesn't exist
mkdir -p $HGCTL_ROOT/internal/testdata

# Make any existing anvil files writable before starting
echo "Making existing anvil files writable..."
chmod 644 "$anvilL1DumpStatePath" 2>/dev/null || true
chmod 644 "$anvilL1ConfigPath" 2>/dev/null || true
chmod 644 "$anvilL2DumpStatePath" 2>/dev/null || true
chmod 644 "$anvilL2ConfigPath" 2>/dev/null || true

# Load seed accounts from ponos config
seedAccounts=$(cat $PROJECT_ROOT/ponos/anvilConfig/accounts.json)

# -----------------------------------------------------------------------------
# Start Ethereum L1 (without dumping state yet)
# -----------------------------------------------------------------------------
echo "Starting L1 Anvil on port $anvilL1RpcPort..."
anvil \
    --fork-url $L1_FORK_RPC_URL \
    --dump-state $anvilL1DumpStatePath \
    --chain-id $anvilL1ChainId \
    --port $anvilL1RpcPort \
    --block-time 2 \
    --fork-block-number $anvilL1StartBlock &

anvilL1Pid=$!
echo "L1 Anvil PID: $anvilL1Pid"

# Wait for L1 to be ready
echo "Waiting for L1 Anvil to be ready..."
for i in {1..30}; do
    if cast block-number --rpc-url "$anvilL1RpcUrl" >/dev/null 2>&1; then
        echo "L1 Anvil is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "ERROR: L1 Anvil failed to start after 30 seconds"
        exit 1
    fi
    sleep 1
done

# -----------------------------------------------------------------------------
# Start Base L2 (without dumping state yet)
# -----------------------------------------------------------------------------
echo "Starting L2 Anvil on port $anvilL2RpcPort..."
anvil \
    --fork-url $L2_FORK_RPC_URL \
    --dump-state $anvilL2DumpStatePath \
    --chain-id $anvilL2ChainId \
    --port $anvilL2RpcPort \
    --fork-block-number $anvilL2StartBlock &

anvilL2Pid=$!
echo "L2 Anvil PID: $anvilL2Pid"

# Wait for L2 to be ready
echo "Waiting for L2 Anvil to be ready..."
for i in {1..30}; do
    if cast block-number --rpc-url "$anvilL2RpcUrl" >/dev/null 2>&1; then
        echo "L2 Anvil is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "ERROR: L2 Anvil failed to start after 30 seconds"
        exit 1
    fi
    sleep 1
done

# -----------------------------------------------------------------------------
# Generate test keystores AFTER chains are running
# -----------------------------------------------------------------------------
echo "Generating test operator keystores..."

# For testing, we'll use deterministic BN254 keys
# BN254 private keys are 32 bytes (64 hex chars)
AGGREGATOR_BN254_PK="1234567890123456789012345678901234567890123456789012345678901234"
EXECUTOR_BN254_PK="abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd"

# Test passwords for integration tests
AGGREGATOR_PASSWORD="aggregator-test-password"
EXECUTOR_PASSWORD="executor-test-password"

# For funding, we need ECDSA addresses. Since BN254 keys don't directly map to Ethereum addresses,
# we'll use separate ECDSA keys for funding the operators
# Remove 0x prefix for hgctl
AGGREGATOR_ECDSA_PK="1234567890123456789012345678901234567890123456789012345678901234"
EXECUTOR_ECDSA_PK="abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd"

# Function to create keystore with overwrite
create_keystore() {
    local name=$1
    local type=$2
    local key=$3
    local password=$4

    # Create new keystore
    "$HGCTL" keystore create \
        --name "$name" \
        --type "$type" \
        --key "$key" \
        --password "$password"
}

# Get ECDSA addresses before creating context
AGGREGATOR_ADDRESS=$(cast wallet address --private-key 0x$AGGREGATOR_ECDSA_PK)
EXECUTOR_ADDRESS=$(cast wallet address --private-key 0x$EXECUTOR_ECDSA_PK)

# Delete existing test context if it exists (ignore errors)
echo "Removing any existing test context..."
rm -rf "$HOME/.hgctl/test" 2>/dev/null || true
rm -rf "$HOME/.config/hgctl/test" 2>/dev/null || true

# Create context with non-interactive flags
echo "Creating hgctl context 'test' with L1 RPC URL and operator address..."
"$HGCTL" context create \
    --l1-rpc-url "$anvilL1RpcUrl" \
    --l2-rpc-url "$anvilL2RpcUrl" \
    test
"$HGCTL" context set --operator-address "$AGGREGATOR_ADDRESS"

# Create BN254 keystores using hgctl
echo "Creating BN254 keystore for aggregator..."
create_keystore "aggregator" "bn254" "$AGGREGATOR_BN254_PK" "$AGGREGATOR_PASSWORD"

echo "Creating BN254 keystore for executor..."
create_keystore "executor" "bn254" "$EXECUTOR_BN254_PK" "$EXECUTOR_PASSWORD"

# Create ECDSA keystores for transaction signing
echo "Creating ECDSA keystore for aggregator..."
create_keystore "aggregator-ecdsa" "ecdsa" "$AGGREGATOR_ECDSA_PK" "$AGGREGATOR_PASSWORD"

echo "Creating ECDSA keystore for executor..."
create_keystore "executor-ecdsa" "ecdsa" "$EXECUTOR_ECDSA_PK" "$EXECUTOR_PASSWORD"

# Debug: Show where keystores are created
echo "Checking for keystores in possible locations..."
ls -la "$HOME/.hgctl/test/keystores/" 2>/dev/null || echo "No keystores in ~/.hgctl/test/keystores/"

# Function to copy keystore with error checking
copy_keystore() {
    local name=$1
    local dest=$2

    if [ -f "$HOME/.hgctl/test/keystores/$name/key.json" ]; then
        cp "$HOME/.hgctl/test/keystores/$name/key.json" "$dest"
        echo "Copied $name keystore to $dest"
    elif [ -f "$HOME/.config/hgctl/test/keystores/$name/key.json" ]; then
        cp "$HOME/.config/hgctl/test/keystores/$name/key.json" "$dest"
        echo "Copied $name keystore to $dest"
    else
        echo "ERROR: Could not find $name keystore"
        exit 1
    fi
}

# Copy the generated keystores to our test keys directory
mkdir -p "$KEYS_DIR"
copy_keystore "aggregator" "$KEYS_DIR/aggregator-keystore.json"
copy_keystore "executor" "$KEYS_DIR/executor-keystore.json"
copy_keystore "aggregator-ecdsa" "$KEYS_DIR/aggregator-ecdsa-keystore.json"
copy_keystore "executor-ecdsa" "$KEYS_DIR/executor-ecdsa-keystore.json"

echo "Generated test keys:"
echo "  Aggregator ECDSA address (for funding): $AGGREGATOR_ADDRESS"
echo "  Executor ECDSA address (for funding): $EXECUTOR_ADDRESS"
echo "  Keys directory: $KEYS_DIR"
echo "  Aggregator keystore: $KEYS_DIR/aggregator-keystore.json"
echo "  Executor keystore: $KEYS_DIR/executor-keystore.json"

# -----------------------------------------------------------------------------
# Fund ALL accounts NOW before contract deployment
# -----------------------------------------------------------------------------
function fundAccount() {
    address=$1
    echo "Funding address $address on L1"
    cast rpc --rpc-url $anvilL1RpcUrl anvil_setBalance $address '0x21E19E0C9BAB2400000' # 10,000 ETH

    echo "Funding address $address on L2"
    cast rpc --rpc-url $anvilL2RpcUrl anvil_setBalance $address '0x21E19E0C9BAB2400000' # 10,000 ETH
}

# loop over the seed accounts (json array) and fund the accounts
numAccounts=$(echo $seedAccounts | jq '. | length - 1')
for i in $(seq 0 $numAccounts); do
    account=$(echo $seedAccounts | jq -r ".[$i]")
    address=$(echo $account | jq -r '.address')

    fundAccount $address
done

# fund the account used for table transport
fundAccount "0x8736311E6b706AfF3D8132Adf351387092802bA6"
fundAccount "0xb094Ba769b4976Dc37fC689A76675f31bc4923b0"

# Fund our generated operator accounts
echo "Funding generated operator accounts..."
fundAccount "$AGGREGATOR_ADDRESS"
fundAccount "$EXECUTOR_ADDRESS"

# deployer account
deployAccountAddress=$(echo $seedAccounts | jq -r '.[0].address')
deployAccountPk=$(echo $seedAccounts | jq -r '.[0].private_key')
export PRIVATE_KEY_DEPLOYER="0x$deployAccountPk"
echo "Deploy account: $deployAccountAddress"

# avs account
avsAccountAddress=$(echo $seedAccounts | jq -r '.[1].address')
avsAccountPk=$(echo $seedAccounts | jq -r '.[1].private_key')
export PRIVATE_KEY_AVS="0x$avsAccountPk"
echo "AVS account: $avsAccountAddress"

# app account
appAccountAddress=$(echo $seedAccounts | jq -r '.[2].address')
appAccountPk=$(echo $seedAccounts | jq -r '.[2].private_key')
export PRIVATE_KEY_APP="0x$appAccountPk"
echo "App account: $appAccountAddress"

# operator account (use ECDSA keys for transactions)
operatorAccountAddress=$AGGREGATOR_ADDRESS
operatorAccountPk=$AGGREGATOR_ECDSA_PK
export PRIVATE_KEY_OPERATOR="0x$operatorAccountPk"
echo "Operator account: $operatorAccountAddress"

# exec operator account (use ECDSA keys for transactions)
execOperatorAccountAddress=$EXECUTOR_ADDRESS
execOperatorAccountPk=$EXECUTOR_ECDSA_PK
export PRIVATE_KEY_EXEC_OPERATOR="0x$execOperatorAccountPk"
echo "Exec Operator account: $execOperatorAccountAddress"

# -----------------------------------------------------------------------------
# Deploy contracts and setup everything
# -----------------------------------------------------------------------------
cd $PROJECT_ROOT/contracts

# Ensure Foundry dependencies are installed
echo "Installing Foundry dependencies..."
forge install || true  # Continue even if already installed
forge build  # Build contracts to ensure everything is ready

export L1_RPC_URL="http://localhost:${anvilL1RpcPort}"
export L2_RPC_URL="http://localhost:${anvilL2RpcPort}"

# -----------------------------------------------------------------------------
# Deploy L1 avs contract
# -----------------------------------------------------------------------------
echo "Deploying L1 AVS contract..."
forge script script/local/DeployAVSL1Contracts.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run(address)" "${avsAccountAddress}"

# we need to get index 2 since thats where the actual proxy lives
avsTaskRegistrarAddress=$(cat ./broadcast/DeployAVSL1Contracts.s.sol/$anvilL1ChainId/run-latest.json | jq -r '.transactions[2].contractAddress')
echo "L1 AVS contract address: $avsTaskRegistrarAddress"

# -----------------------------------------------------------------------------
# Setup L1 AVS - THIS IS THE CRITICAL STEP FOR OPERATOR SETS!
# -----------------------------------------------------------------------------
echo "Setting up L1 AVS (creating operator sets)..."
forge script script/local/SetupAVSL1.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run(address)" $avsTaskRegistrarAddress

# -----------------------------------------------------------------------------
# Configure Operator Sets for AVS
# -----------------------------------------------------------------------------
echo "Configuring operator sets for AVS..."
forge script script/local/ConfigureOperatorSets.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run(address)" $avsAccountAddress

# -----------------------------------------------------------------------------
# Allowlist aggregator operator for operator set 0
# -----------------------------------------------------------------------------
echo "Allowlisting aggregator operator for operator set 0..."
export AGGREGATOR_PRIVATE_KEY="0x$operatorAccountPk"
forge script script/local/AllowlistOperators.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run(address)" "$avsTaskRegistrarAddress"

# -----------------------------------------------------------------------------
# Setup L1 multichain
# -----------------------------------------------------------------------------
echo "Setting up L1 multichain..."
export L1_CHAIN_ID=$anvilL1ChainId
export L2_CHAIN_ID=$anvilL2ChainId
export AVS_ADDRESS=$avsAccountAddress

# Whitelist chains
cast rpc anvil_impersonateAccount "0xb094Ba769b4976Dc37fC689A76675f31bc4923b0" --rpc-url $L1_RPC_URL
forge script script/local/WhitelistDevnet.s.sol --slow --rpc-url $L1_RPC_URL --sender "0xb094Ba769b4976Dc37fC689A76675f31bc4923b0" --unlocked --broadcast --sig "run()"

forge script script/local/SetupAVSMultichain.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run()"

# -----------------------------------------------------------------------------
# Deploy L2
# -----------------------------------------------------------------------------
echo "Deploying L2 contracts on L1..."
forge script script/local/DeployAVSL2Contracts.s.sol --slow --rpc-url $L1_RPC_URL --broadcast
taskHookAddressL1=$(cat ./broadcast/DeployAVSL2Contracts.s.sol/$anvilL1ChainId/run-latest.json | jq -r '.transactions[0].contractAddress')

echo "Deploying L2 contracts on L2..."
forge script script/local/DeployAVSL2Contracts.s.sol --slow --rpc-url $L2_RPC_URL --broadcast
taskHookAddressL2=$(cat ./broadcast/DeployAVSL2Contracts.s.sol/$anvilL2ChainId/run-latest.json | jq -r '.transactions[0].contractAddress')

# Still need to mine some blocks for any configuration delays
echo "Mining initial blocks..."
cast rpc --rpc-url $L1_RPC_URL anvil_mine 10
cast rpc --rpc-url $L2_RPC_URL anvil_mine 10

echo "Current block number: "
cast block-number --rpc-url $L1_RPC_URL

# Kill anvil processes
kill $anvilL1Pid
kill $anvilL2Pid
sleep 3

cd $HGCTL_ROOT

# Make the files read-only to prevent accidental overwrites
chmod 444 internal/testdata/anvil*

function lowercaseAddress() {
    echo "$1" | tr '[:upper:]' '[:lower:]'
}

deployAccountPublicKey=$(cast wallet public-key --private-key "0x$deployAccountPk")
avsAccountPublicKey=$(cast wallet public-key --private-key "0x$avsAccountPk")
appAccountPublicKey=$(cast wallet public-key --private-key "0x$appAccountPk")
operatorAccountPublicKey=$(cast wallet public-key --private-key "0x$operatorAccountPk")
execOperatorAccountPublicKey=$(cast wallet public-key --private-key "0x$execOperatorAccountPk")
deployAccountAddress=$(lowercaseAddress $deployAccountAddress)

# Create chainData directory if it doesn't exist
mkdir -p $HGCTL_ROOT/internal/testutils/chainData

# create a heredoc json file and dump it to internal/testutils/chainData/chain-config.json
cat <<EOF > $HGCTL_ROOT/internal/testutils/chainData/chain-config.json
{
      "deployAccountAddress": "$deployAccountAddress",
      "deployAccountPk": "$deployAccountPk",
      "deployAccountPublicKey": "$deployAccountPublicKey",
      "avsAccountAddress": "$avsAccountAddress",
      "avsAccountPk": "$avsAccountPk",
      "avsAccountPublicKey": "$avsAccountPublicKey",
      "appAccountAddress": "$appAccountAddress",
      "appAccountPk": "$appAccountPk",
      "appAccountPublicKey": "$appAccountPublicKey",
      "operatorAccountAddress": "$operatorAccountAddress",
      "operatorAccountPk": "$operatorAccountPk",
      "operatorAccountPublicKey": "$operatorAccountPublicKey",
      "operatorKeystorePath": "$KEYS_DIR/aggregator-keystore.json",
      "operatorKeystorePassword": "$AGGREGATOR_PASSWORD",
      "execOperatorAccountAddress": "$execOperatorAccountAddress",
      "execOperatorAccountPk": "$execOperatorAccountPk",
      "execOperatorAccountPublicKey": "$execOperatorAccountPublicKey",
      "execOperatorKeystorePath": "$KEYS_DIR/executor-keystore.json",
      "execOperatorKeystorePassword": "$EXECUTOR_PASSWORD",
      "avsTaskRegistrarAddress": "$avsTaskRegistrarAddress",
      "avsTaskHookAddressL1": "$taskHookAddressL1",
      "avsTaskHookAddressL2": "$taskHookAddressL2",
      "keyRegistrarAddress": "0xa4db30d08d8bbca00d40600bee9f029984db162a",
      "releaseManagerAddress": "0x59c8d715dca616e032b744a753c017c9f3e16bf4",
      "delegationManagerAddress": "0xd4a7e1bd8015057293f0d0a557088c286942e84b",
      "allocationManagerAddress": "0x42583067658071247ec8CE0A516A58f682002d07",
      "strategyManagerAddress": "0xdfB5f6CE42aAA7830E94ECFCcAd411beF4d4D5b6",
      "destinationEnv": "anvil",
      "forkL1Block": $anvilL1StartBlock,
      "forkL2Block": $anvilL2StartBlock,
      "l1ChainId": $anvilL1ChainId,
      "l2ChainId": $anvilL2ChainId,
      "l1RPC": "$L1_RPC_URL",
      "l2RPC": "$L2_RPC_URL"
}
EOF

echo "Test chain state generated successfully!"
echo "Anvil state files saved to: $HGCTL_ROOT/internal/testdata/"
echo "Chain config saved to: $HGCTL_ROOT/internal/testutils/chainData/chain-config.json"
echo ""
echo "Generated operator keystores:"
echo "  Aggregator: $KEYS_DIR/aggregator-keystore.json"
echo "  Executor: $KEYS_DIR/executor-keystore.json"
echo ""
echo "The saved state includes:"
echo "  - All accounts funded with 10,000 ETH"
echo "  - Deployed AVS contracts"
echo "  - Configured operator sets"
echo "  - All multichain setup complete"