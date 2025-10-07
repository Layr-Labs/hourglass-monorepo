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

# -----------------------------------------------------------------------------
# Validate required parameters
# -----------------------------------------------------------------------------
if [ -z "$ENVIRONMENT" ]; then
    echo "Error: ENVIRONMENT variable must be set (local or staging)"
    exit 1
fi

if [ "$ENVIRONMENT" != "local" ] && [ "$ENVIRONMENT" != "staging" ]; then
    echo "Error: ENVIRONMENT must be 'local' or 'staging', got: $ENVIRONMENT"
    exit 1
fi

if [ -z "$REGISTRY_URL" ]; then
    echo "Error: REGISTRY_URL variable must be set"
    exit 1
fi

# Navigate to the project root (where contracts directory exists)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Script is in hgctl-go/internal/testutils/scripts
HGCTL_ROOT="$SCRIPT_DIR/../../.."  # Go up to hgctl-go root
PROJECT_ROOT="$HGCTL_ROOT/.."       # Go up to hourglass-monorepo root
TEST_CONFIG_DIR="$HGCTL_ROOT/internal/testdata/.hgctl"  # Test-specific config directory

echo "Script dir: $SCRIPT_DIR"
echo "HGCTL root: $HGCTL_ROOT"
echo "Project root: $PROJECT_ROOT"
echo "Test config dir: $TEST_CONFIG_DIR"

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
L1_FORK_RPC_URL=https://shy-convincing-wave.ethereum-sepolia.quiknode.pro/3dd1c3a3090f08c2452c5bd135ecfbce22cde912

anvilL1ChainId=31337
anvilL1StartBlock=9349704
anvilL1DumpStatePath=$HGCTL_ROOT/internal/testdata/anvil-l1-state.json
anvilL1ConfigPath=$HGCTL_ROOT/internal/testdata/anvil-l1-config.json
anvilL1RpcPort=8545
anvilL1RpcUrl="http://localhost:${anvilL1RpcPort}"

# base mainnet
L2_FORK_RPC_URL=https://dimensional-compatible-sky.base-sepolia.quiknode.pro/2bd3461589b9a684678254bbd88bccbd65c34b84/

anvilL2ChainId=31338
anvilL2StartBlock=31958695
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

# -----------------------------------------------------------------------------
# Load accounts from anvilConfig/accounts.json for reproducible testing
# -----------------------------------------------------------------------------
echo "Loading accounts from anvilConfig/accounts.json..."

ACCOUNTS_FILE="$PROJECT_ROOT/hgctl-go/internal/testutils/anvilConfig/accounts.json"

if [ ! -f "$ACCOUNTS_FILE" ]; then
    echo "Error: accounts.json not found at $ACCOUNTS_FILE"
    exit 1
fi

# Load seed accounts
seedAccounts=$(cat $ACCOUNTS_FILE)

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
# Load operator and staker accounts from config
# -----------------------------------------------------------------------------
echo "Loading operator and staker accounts from config..."

# Load operator accounts from configuration
operatorAccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "operator") | .private_key')
operatorAccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "operator") | .address')

execOperatorAccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_operator") | .private_key')
execOperatorAccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_operator") | .address')

# Load additional executor operators
execOperator2AccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_operator2") | .private_key')
execOperator2AccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_operator2") | .address')

execOperator3AccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_operator3") | .private_key')
execOperator3AccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_operator3") | .address')

execOperator4AccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_operator4") | .private_key')
execOperator4AccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_operator4") | .address')

# Load or generate unregistered operators (for testing registration flow)
# Try to load from config first, generate if not present
unregisteredOperator1AccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "unregistered_operator1") | .private_key')
unregisteredOperator1AccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "unregistered_operator1") | .address')

unregisteredOperator2AccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "unregistered_operator2") | .private_key')
unregisteredOperator2AccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "unregistered_operator2") | .address')

# Generate unregistered operator 1 if not in config
if [ -z "$unregisteredOperator1AccountPk" ] || [ "$unregisteredOperator1AccountPk" = "null" ]; then
    echo "Generating unregistered_operator1 (not found in config)..."
    unregisteredOperator1AccountPk=$(openssl rand -hex 32)
    unregisteredOperator1AccountAddress=$(cast wallet address --private-key 0x$unregisteredOperator1AccountPk)
    echo "  Generated address: $unregisteredOperator1AccountAddress"
fi

# Generate unregistered operator 2 if not in config
if [ -z "$unregisteredOperator2AccountPk" ] || [ "$unregisteredOperator2AccountPk" = "null" ]; then
    echo "Generating unregistered_operator2 (not found in config)..."
    unregisteredOperator2AccountPk=$(openssl rand -hex 32)
    unregisteredOperator2AccountAddress=$(cast wallet address --private-key 0x$unregisteredOperator2AccountPk)
    echo "  Generated address: $unregisteredOperator2AccountAddress"
fi

# Load staker accounts
aggStakerAccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "agg_staker") | .private_key')
aggStakerAccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "agg_staker") | .address')

execStakerAccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_staker") | .private_key')
execStakerAccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_staker") | .address')

# Load additional executor stakers
execStaker2AccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_staker2") | .private_key')
execStaker2AccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_staker2") | .address')

execStaker3AccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_staker3") | .private_key')
execStaker3AccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_staker3") | .address')

execStaker4AccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_staker4") | .private_key')
execStaker4AccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "exec_staker4") | .address')

echo "Loaded operator accounts:"
echo "  Operator (aggregator): $operatorAccountAddress"
echo "  Exec Operator 1: $execOperatorAccountAddress"
echo "  Exec Operator 2: $execOperator2AccountAddress"
echo "  Exec Operator 3: $execOperator3AccountAddress"
echo "  Exec Operator 4: $execOperator4AccountAddress"
echo "  Unregistered Operator 1: $unregisteredOperator1AccountAddress"
echo "  Unregistered Operator 2: $unregisteredOperator2AccountAddress"
echo "  Agg staker: $aggStakerAccountAddress"
echo "  Exec staker 1: $execStakerAccountAddress"
echo "  Exec staker 2: $execStaker2AccountAddress"
echo "  Exec staker 3: $execStaker3AccountAddress"
echo "  Exec staker 4: $execStaker4AccountAddress"

# -----------------------------------------------------------------------------
# Generate test keystores AFTER chains are running
# -----------------------------------------------------------------------------
echo "Generating test operator keystores..."

# For testing, we'll use deterministic BN254 keys
# BN254 private keys are 32 bytes (64 hex chars)
AGGREGATOR_BN254_PK="1234567890123456789012345678901234567890123456789012345678901234"
EXECUTOR_BN254_PK="abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd"
TRANSPORT_BLS_KEY="0x5f8e6420b9cb0c940e3d3f8b99177980785906d16fb3571f70d7a05ecf5f2172"

# Test passwords for integration tests
AGGREGATOR_PASSWORD="aggregator-test-password"
EXECUTOR_PASSWORD="executor-test-password"

# Use the ECDSA keys from the accounts config for funding
AGGREGATOR_ECDSA_PK="$operatorAccountPk"
EXECUTOR_ECDSA_PK="$execOperatorAccountPk"

# Function to create keystore with overwrite
create_keystore() {
    local name=$1
    local type=$2
    local key=$3
    local password=$4

    # Create new keystore
    "$HGCTL" --config-dir "$TEST_CONFIG_DIR" keystore create \
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
rm -rf "$TEST_CONFIG_DIR/" 2>/dev/null || true

# Create context with non-interactive flags
echo "Creating hgctl context 'test' with L1 RPC URL and operator address..."
"$HGCTL" --config-dir "$TEST_CONFIG_DIR" context create \
    --l1-rpc-url "$anvilL1RpcUrl" \
    --l2-rpc-url "$anvilL2RpcUrl" \
    test
"$HGCTL" --config-dir "$TEST_CONFIG_DIR" context set --operator-address "$AGGREGATOR_ADDRESS"

# Create BN254 keystores using hgctl
echo "Creating BN254 keystore for aggregator..."
create_keystore "aggregator" "bn254" "$AGGREGATOR_BN254_PK" "$AGGREGATOR_PASSWORD"

echo "Creating BN254 keystore for executor..."
create_keystore "executor" "bn254" "$EXECUTOR_BN254_PK" "$EXECUTOR_PASSWORD"

# Create ECDSA keystores for transaction signing
echo "Creating ECDSA keystore for aggregator..."
create_keystore "aggregator-ecdsa" "ecdsa" "$AGGREGATOR_ECDSA_PK" "$AGGREGATOR_PASSWORD"

echo "Creating ECDSA keystore for executor 1..."
create_keystore "executor-ecdsa" "ecdsa" "$EXECUTOR_ECDSA_PK" "$EXECUTOR_PASSWORD"

echo "Creating ECDSA keystore for executor 2..."
create_keystore "executor2-ecdsa" "ecdsa" "$execOperator2AccountPk" "$EXECUTOR_PASSWORD"

echo "Creating ECDSA keystore for executor 3..."
create_keystore "executor3-ecdsa" "ecdsa" "$execOperator3AccountPk" "$EXECUTOR_PASSWORD"

echo "Creating ECDSA keystore for executor 4..."
create_keystore "executor4-ecdsa" "ecdsa" "$execOperator4AccountPk" "$EXECUTOR_PASSWORD"

echo "Creating ECDSA keystore for unregistered operator 1..."
create_keystore "unregistered-operator1-ecdsa" "ecdsa" "$unregisteredOperator1AccountPk" "$EXECUTOR_PASSWORD"

echo "Creating ECDSA keystore for unregistered operator 2..."
create_keystore "unregistered-operator2-ecdsa" "ecdsa" "$unregisteredOperator2AccountPk" "$EXECUTOR_PASSWORD"

# Create system ECDSA keystores for registered operators (these generate random keys)
# These are separate from operator signing keys and used for system-level operations
echo "Creating system ECDSA keystores for registered operators..."
create_keystore "aggregator-system" "ecdsa" "" "$AGGREGATOR_PASSWORD"
create_keystore "executor-system" "ecdsa" "" "$EXECUTOR_PASSWORD"
create_keystore "executor2-system" "ecdsa" "" "$EXECUTOR_PASSWORD"
create_keystore "executor3-system" "ecdsa" "" "$EXECUTOR_PASSWORD"
create_keystore "executor4-system" "ecdsa" "" "$EXECUTOR_PASSWORD"

# Create system keystores for unregistered operators (for testing)
# These will NOT be registered with EigenLayer
echo "Creating system keystores for unregistered operators..."
create_keystore "unregistered1-system-bn254" "bn254" "" "$EXECUTOR_PASSWORD"
create_keystore "unregistered1-system-ecdsa" "ecdsa" "" "$EXECUTOR_PASSWORD"
create_keystore "unregistered2-system-bn254" "bn254" "" "$EXECUTOR_PASSWORD"
create_keystore "unregistered2-system-ecdsa" "ecdsa" "" "$EXECUTOR_PASSWORD"

# Debug: Show where keystores are created
echo "Checking for keystores..."
ls -la "$TEST_CONFIG_DIR/test/keystores/" 2>/dev/null || echo "No keystores found in $TEST_CONFIG_DIR/test/keystores/"

# Function to copy keystore with error checking
copy_keystore() {
    local name=$1
    local dest=$2

    if [ -f "$TEST_CONFIG_DIR/test/keystores/$name/key.json" ]; then
        cp "$TEST_CONFIG_DIR/test/keystores/$name/key.json" "$dest"
        echo "Copied $name keystore to $dest"
    else
        echo "ERROR: Could not find $name keystore at $TEST_CONFIG_DIR/test/keystores/$name/key.json"
        exit 1
    fi
}

# Copy the generated keystores to our test keys directory
mkdir -p "$KEYS_DIR"
copy_keystore "aggregator" "$KEYS_DIR/aggregator-keystore.json"
copy_keystore "executor" "$KEYS_DIR/executor-keystore.json"
copy_keystore "aggregator-ecdsa" "$KEYS_DIR/aggregator-ecdsa-keystore.json"
copy_keystore "executor-ecdsa" "$KEYS_DIR/executor-ecdsa-keystore.json"
copy_keystore "executor2-ecdsa" "$KEYS_DIR/executor2-ecdsa-keystore.json"
copy_keystore "executor3-ecdsa" "$KEYS_DIR/executor3-ecdsa-keystore.json"
copy_keystore "executor4-ecdsa" "$KEYS_DIR/executor4-ecdsa-keystore.json"
copy_keystore "unregistered-operator1-ecdsa" "$KEYS_DIR/unregistered-operator1-ecdsa-keystore.json"
copy_keystore "unregistered-operator2-ecdsa" "$KEYS_DIR/unregistered-operator2-ecdsa-keystore.json"

# Copy system keystores for registered operators
copy_keystore "aggregator-system" "$KEYS_DIR/aggregator-system-keystore.json"
copy_keystore "executor-system" "$KEYS_DIR/executor-system-keystore.json"
copy_keystore "executor2-system" "$KEYS_DIR/executor2-system-keystore.json"
copy_keystore "executor3-system" "$KEYS_DIR/executor3-system-keystore.json"
copy_keystore "executor4-system" "$KEYS_DIR/executor4-system-keystore.json"

# Copy system keystores for unregistered operators
copy_keystore "unregistered1-system-bn254" "$KEYS_DIR/unregistered1-system-bn254-keystore.json"
copy_keystore "unregistered1-system-ecdsa" "$KEYS_DIR/unregistered1-system-ecdsa-keystore.json"
copy_keystore "unregistered2-system-bn254" "$KEYS_DIR/unregistered2-system-bn254-keystore.json"
copy_keystore "unregistered2-system-ecdsa" "$KEYS_DIR/unregistered2-system-ecdsa-keystore.json"

# Extract system key information using hgctl keystore show
echo "Extracting system key information for registered operators..."
AGGREGATOR_SYSTEM_PK=$("$HGCTL" --config-dir "$TEST_CONFIG_DIR" keystore show --name aggregator-system --password "$AGGREGATOR_PASSWORD" | grep "Private key:" | awk '{print $3}')
EXECUTOR_SYSTEM_PK=$("$HGCTL" --config-dir "$TEST_CONFIG_DIR" keystore show --name executor-system --password "$EXECUTOR_PASSWORD" | grep "Private key:" | awk '{print $3}')
EXECUTOR2_SYSTEM_PK=$("$HGCTL" --config-dir "$TEST_CONFIG_DIR" keystore show --name executor2-system --password "$EXECUTOR_PASSWORD" | grep "Private key:" | awk '{print $3}')
EXECUTOR3_SYSTEM_PK=$("$HGCTL" --config-dir "$TEST_CONFIG_DIR" keystore show --name executor3-system --password "$EXECUTOR_PASSWORD" | grep "Private key:" | awk '{print $3}')
EXECUTOR4_SYSTEM_PK=$("$HGCTL" --config-dir "$TEST_CONFIG_DIR" keystore show --name executor4-system --password "$EXECUTOR_PASSWORD" | grep "Private key:" | awk '{print $3}')

echo "Extracting system key information for unregistered operators..."
UNREGISTERED1_SYSTEM_BN254_PK=$("$HGCTL" --config-dir "$TEST_CONFIG_DIR" keystore show --name unregistered1-system-bn254 --password "$EXECUTOR_PASSWORD" | grep "Private key:" | awk '{print $3}')
UNREGISTERED1_SYSTEM_ECDSA_PK=$("$HGCTL" --config-dir "$TEST_CONFIG_DIR" keystore show --name unregistered1-system-ecdsa --password "$EXECUTOR_PASSWORD" | grep "Private key:" | awk '{print $3}')
UNREGISTERED2_SYSTEM_BN254_PK=$("$HGCTL" --config-dir "$TEST_CONFIG_DIR" keystore show --name unregistered2-system-bn254 --password "$EXECUTOR_PASSWORD" | grep "Private key:" | awk '{print $3}')
UNREGISTERED2_SYSTEM_ECDSA_PK=$("$HGCTL" --config-dir "$TEST_CONFIG_DIR" keystore show --name unregistered2-system-ecdsa --password "$EXECUTOR_PASSWORD" | grep "Private key:" | awk '{print $3}')

# Derive addresses from system keys
AGGREGATOR_SYSTEM_ADDRESS=$(cast wallet address --private-key 0x$AGGREGATOR_SYSTEM_PK)
EXECUTOR_SYSTEM_ADDRESS=$(cast wallet address --private-key 0x$EXECUTOR_SYSTEM_PK)
EXECUTOR2_SYSTEM_ADDRESS=$(cast wallet address --private-key 0x$EXECUTOR2_SYSTEM_PK)
EXECUTOR3_SYSTEM_ADDRESS=$(cast wallet address --private-key 0x$EXECUTOR3_SYSTEM_PK)
EXECUTOR4_SYSTEM_ADDRESS=$(cast wallet address --private-key 0x$EXECUTOR4_SYSTEM_PK)
UNREGISTERED1_SYSTEM_ECDSA_ADDRESS=$(cast wallet address --private-key 0x$UNREGISTERED1_SYSTEM_ECDSA_PK)
UNREGISTERED2_SYSTEM_ECDSA_ADDRESS=$(cast wallet address --private-key 0x$UNREGISTERED2_SYSTEM_ECDSA_PK)

echo "Generated test keys:"
echo "  Aggregator ECDSA address (for funding): $AGGREGATOR_ADDRESS"
echo "  Executor ECDSA address (for funding): $EXECUTOR_ADDRESS"
echo "  Aggregator system address: $AGGREGATOR_SYSTEM_ADDRESS"
echo "  Executor system address: $EXECUTOR_SYSTEM_ADDRESS"
echo "  Executor 2 system address: $EXECUTOR2_SYSTEM_ADDRESS"
echo "  Executor 3 system address: $EXECUTOR3_SYSTEM_ADDRESS"
echo "  Executor 4 system address: $EXECUTOR4_SYSTEM_ADDRESS"
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

# Fund CrossChainRegistry owner for impersonation
fundAccount "0xBE1685C81aA44FF9FB319dD389addd9374383e90"

# Fund our generated operator accounts
echo "Funding generated operator accounts..."
fundAccount "$AGGREGATOR_ADDRESS"
fundAccount "$EXECUTOR_ADDRESS"

# Fund system key addresses
echo "Funding system key addresses..."
fundAccount "$AGGREGATOR_SYSTEM_ADDRESS"
fundAccount "$EXECUTOR_SYSTEM_ADDRESS"
fundAccount "$EXECUTOR2_SYSTEM_ADDRESS"
fundAccount "$EXECUTOR3_SYSTEM_ADDRESS"
fundAccount "$EXECUTOR4_SYSTEM_ADDRESS"

# Fund unregistered operators for integration testing
echo "Funding unregistered operator accounts..."
fundAccount "$unregisteredOperator1AccountAddress"
fundAccount "$unregisteredOperator2AccountAddress"

# Fund unregistered operator system key addresses (ECDSA only, BN254 keys don't have addresses)
echo "Funding unregistered operator system key addresses..."
fundAccount "$UNREGISTERED1_SYSTEM_ECDSA_ADDRESS"
fundAccount "$UNREGISTERED2_SYSTEM_ECDSA_ADDRESS"

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
# Install DevKit if not already installed
# -----------------------------------------------------------------------------
if ! command -v devkit &> /dev/null; then
    echo "Installing DevKit..."
    curl -fsSL https://raw.githubusercontent.com/Layr-Labs/devkit-cli/main/install-devkit.sh | bash

    # Source the shell config to get devkit in PATH
    export PATH="$HOME/.devkit/bin:$PATH"
fi

echo "DevKit version: $(devkit --version)"

# -----------------------------------------------------------------------------
# Create AVS project using DevKit
# -----------------------------------------------------------------------------
cd $PROJECT_ROOT
AVS_PROJECT_DIR="$PROJECT_ROOT/integration-test-avs"

# Clean up any existing AVS project
if [ -d "$AVS_PROJECT_DIR" ]; then
    echo "Removing existing AVS project..."
    rm -rf "$AVS_PROJECT_DIR"
fi

echo "Creating AVS project with DevKit..."
devkit avs create integration-test-avs

cd "$AVS_PROJECT_DIR"

devkit telemetry --disable

# -----------------------------------------------------------------------------
# Configure DevKit context for local anvil chains
# -----------------------------------------------------------------------------
echo "Configuring DevKit context for local testnet..."
export L1_RPC_URL="http://localhost:${anvilL1RpcPort}"
export L2_RPC_URL="http://localhost:${anvilL2RpcPort}"

# Delete existing testnet context if it exists
rm -rf "$HOME/.devkit/contexts/testnet" 2>/dev/null || true
rm -rf "$HOME/.config/devkit/contexts/testnet" 2>/dev/null || true

# Create testnet context for local anvil deployment
echo "Creating DevKit context for testnet..."
devkit avs context create \
  --context testnet \
  --l1-rpc-url "$L1_RPC_URL" \
  --l2-rpc-url "$L2_RPC_URL" \
  --deployer-private-key "0x$deployAccountPk" \
  --app-private-key "0x$appAccountPk" \
  --avs-private-key "0x$avsAccountPk" \
  --avs-metadata-url "https://example.com/integration-test-avs/metadata.json"

echo "Building AVS with DevKit..."
devkit avs build

# -----------------------------------------------------------------------------
# Deploy AVS contracts using DevKit
# -----------------------------------------------------------------------------
echo "Deploying L1 AVS contracts with DevKit..."
devkit avs deploy contracts l1

# Extract taskAVSRegistrar proxy address from DevKit context YAML file
DEVKIT_CONTEXT_FILE="$AVS_PROJECT_DIR/config/contexts/testnet.yaml"
avsTaskRegistrarAddress=$(grep -B 1 'name: taskAVSRegistrar$' "$DEVKIT_CONTEXT_FILE" | grep "address:" | awk '{print $2}')
echo "L1 AVS TaskAVSRegistrar (Proxy) address: $avsTaskRegistrarAddress"

# Validate we got a valid address
if [ -z "$avsTaskRegistrarAddress" ] || [ "$avsTaskRegistrarAddress" = "\"\"" ]; then
    echo "ERROR: Failed to extract taskAVSRegistrar address from DevKit context"
    exit 1
fi

# -----------------------------------------------------------------------------
# Set release metadata URIs for operator sets
# -----------------------------------------------------------------------------
echo "Setting release metadata URIs..."
devkit avs release uri --operator-set-id 0 --avs-address "$AVS_ADDRESS" --metadata-uri "http://integration-test-uri/operator-set-0"
devkit avs release uri --operator-set-id 1 --avs-address "$AVS_ADDRESS" --metadata-uri "http://integration-test-uri/operator-set-1"

# -----------------------------------------------------------------------------
# Publish AVS release to ReleaseManager
# -----------------------------------------------------------------------------
echo "Publishing AVS release to $REGISTRY_URL..."
devkit avs release publish --registry "$REGISTRY_URL" --upgrade-by-time 2759793679

# -----------------------------------------------------------------------------
# Allowlist aggregator operator for operator set 0 (permissioned set)
# -----------------------------------------------------------------------------
echo "Allowlisting aggregator operator for operator set 0..."
cd "$PROJECT_ROOT/contracts"

export AGGREGATOR_PRIVATE_KEY="0x$operatorAccountPk"

export PRIVATE_KEY_DEPLOYER="0x$avsAccountPk"
forge script script/local/AllowlistOperators.s.sol \
    --rpc-url "$L1_RPC_URL" \
    --broadcast \
    --sig "run(address)" \
    "$avsTaskRegistrarAddress"

# -----------------------------------------------------------------------------
# Register operators with EigenLayer using Foundry script
# -----------------------------------------------------------------------------
echo "Registering operators with EigenLayer..."

# Register aggregator operator for operator set 0
echo "Registering aggregator operator (operator set 0)..."
forge script script/local/RegisterOperator.s.sol \
    --rpc-url "$L1_RPC_URL" \
    --broadcast \
    --via-ir \
    --sig "run(bytes32,bytes32,uint32,string,address,uint32,string)" \
    "0x$operatorAccountPk" \
    "0x$AGGREGATOR_SYSTEM_PK" \
    0 \
    "https://example.com/aggregator/metadata.json" \
    "$avsAccountAddress" \
    0 \
    "localhost:9010"

# Register executor operator 1 for operator set 1
echo "Registering executor operator 1 (operator set 1)..."
forge script script/local/RegisterOperator.s.sol \
    --rpc-url "$L1_RPC_URL" \
    --broadcast \
    --via-ir \
    --sig "run(bytes32,bytes32,uint32,string,address,uint32,string)" \
    "0x$execOperatorAccountPk" \
    "0x$EXECUTOR_SYSTEM_PK" \
    0 \
    "https://example.com/executor1/metadata.json" \
    "$avsAccountAddress" \
    1 \
    "localhost:9090"

# Register executor operator 2 for operator set 1
echo "Registering executor operator 2 (operator set 1)..."
forge script script/local/RegisterOperator.s.sol \
    --rpc-url "$L1_RPC_URL" \
    --broadcast \
    --via-ir \
    --sig "run(bytes32,bytes32,uint32,string,address,uint32,string)" \
    "0x$execOperator2AccountPk" \
    "0x$EXECUTOR2_SYSTEM_PK" \
    0 \
    "https://example.com/executor2/metadata.json" \
    "$avsAccountAddress" \
    1 \
    "localhost:9080"

# Register executor operator 3 for operator set 1
echo "Registering executor operator 3 (operator set 1)..."
forge script script/local/RegisterOperator.s.sol \
    --rpc-url "$L1_RPC_URL" \
    --broadcast \
    --via-ir \
    --sig "run(bytes32,bytes32,uint32,string,address,uint32,string)" \
    "0x$execOperator3AccountPk" \
    "0x$EXECUTOR3_SYSTEM_PK" \
    0 \
    "https://example.com/executor3/metadata.json" \
    "$avsAccountAddress" \
    1 \
    "localhost:9070"

# Register executor operator 4 for operator set 1
echo "Registering executor operator 4 (operator set 1)..."
forge script script/local/RegisterOperator.s.sol \
    --rpc-url "$L1_RPC_URL" \
    --broadcast \
    --via-ir \
    --sig "run(bytes32,bytes32,uint32,string,address,uint32,string)" \
    "0x$execOperator4AccountPk" \
    "0x$EXECUTOR4_SYSTEM_PK" \
    0 \
    "https://example.com/executor4/metadata.json" \
    "$avsAccountAddress" \
    1 \
    "localhost:9060"

# -----------------------------------------------------------------------------
# Stake tokens for operators (WETH for aggregator, stETH for executors)
# -----------------------------------------------------------------------------
echo "Staking tokens for operators..."
echo "  Aggregator: $operatorAccountAddress (staker: $aggStakerAccountAddress)"
echo "  Exec Operator 1: $execOperatorAccountAddress (staker: $execStakerAccountAddress)"
echo "  Exec Operator 2: $execOperator2AccountAddress (staker: $execStaker2AccountAddress)"
echo "  Exec Operator 3: $execOperator3AccountAddress (staker: $execStaker3AccountAddress)"
echo "  Exec Operator 4: $execOperator4AccountAddress (staker: $execStaker4AccountAddress)"

cd "$PROJECT_ROOT/contracts"

SEPOLIA_ALLOCATION_MANAGER="0x42583067658071247ec8CE0A516A58f682002d07"
SEPOLIA_DELEGATION_MANAGER="0xD4A7E1Bd8015057293f0D0A557088c286942e84b"
SEPOLIA_STRATEGY_MANAGER="0x2E3D6c0744b10eb0A4e6F679F71554a39Ec47a5D"
SEPOLIA_STRATEGY_WETH="0x424246eF71b01ee33aA33aC590fd9a0855F5eFbc"
SEPOLIA_STRATEGY_STETH="0x8b29d91e67b013e855EaFe0ad704aC4Ab086a574"

export AGG_STAKER_PRIVATE_KEY="0x$aggStakerAccountPk"
export EXEC_STAKER_PRIVATE_KEY="0x$execStakerAccountPk"
export EXEC_STAKER2_PRIVATE_KEY="0x$execStaker2AccountPk"
export EXEC_STAKER3_PRIVATE_KEY="0x$execStaker3AccountPk"
export EXEC_STAKER4_PRIVATE_KEY="0x$execStaker4AccountPk"

forge script script/local/StakeWithStrategies.sol --slow --rpc-url $L1_RPC_URL --broadcast --via-ir \
    --sig "run(address,address,address,address,address)" \
    "$SEPOLIA_ALLOCATION_MANAGER" \
    "$SEPOLIA_DELEGATION_MANAGER" \
    "$SEPOLIA_STRATEGY_MANAGER" \
    "$SEPOLIA_STRATEGY_WETH" \
    "$SEPOLIA_STRATEGY_STETH"

# Mine blocks to bypass ALLOCATION_CONFIGURATION_DELAY
echo "Mining blocks to finalize allocations..."
cast rpc --rpc-url $L1_RPC_URL anvil_mine 80
cast rpc --rpc-url $L2_RPC_URL anvil_mine 80

# -----------------------------------------------------------------------------
# Setup L1 multichain
# -----------------------------------------------------------------------------

CROSS_CHAIN_REGISTRY="0x287381B1570d9048c4B4C7EC94d21dDb8Aa1352a"

echo "Setting up L1 multichain..."
export L1_CHAIN_ID=$anvilL1ChainId
export L2_CHAIN_ID=$anvilL2ChainId
export AVS_ADDRESS=$avsAccountAddress
export CROSS_CHAIN_REGISTRY="$CROSS_CHAIN_REGISTRY"
export TABLE_UPDATER_ADDRESS="0xB02A15c6Bd0882b35e9936A9579f35FB26E11476"
export BN254_TABLE_CALCULATOR="0xa19E3B00cf4aC46B5e6dc0Bbb0Fb0c86D0D65603"
export ECDSA_TABLE_CALCULATOR="0xaCB5DE6aa94a1908E6FA577C2ade65065333B450"

# Whitelist chains in CrossChainRegistry
echo "Whitelisting anvil chains in CrossChainRegistry..."
echo "  CrossChainRegistry: $CROSS_CHAIN_REGISTRY"
echo "  Table Updater: $TABLE_UPDATER_ADDRESS"
echo "  L1 Chain ID: $anvilL1ChainId"
echo "  L2 Chain ID: $anvilL2ChainId"

# Get CrossChainRegistry owner
CROSS_CHAIN_REGISTRY_OWNER="0xb094Ba769b4976Dc37fC689A76675f31bc4923b0"
echo "Using CrossChainRegistry owner: $CROSS_CHAIN_REGISTRY_OWNER"
cast rpc anvil_impersonateAccount "$CROSS_CHAIN_REGISTRY_OWNER" --rpc-url $L1_RPC_URL
forge script script/local/WhitelistDevnet.s.sol --slow --rpc-url $L1_RPC_URL --sender "$CROSS_CHAIN_REGISTRY_OWNER" --unlocked --broadcast --sig "run()"

# -----------------------------------------------------------------------------
# Transport operator tables for multichain support
# -----------------------------------------------------------------------------
echo "Transporting operator tables for executor operator set..."

# Build the transport binary if not already built
if [ ! -f "$HGCTL_ROOT/bin/transport" ]; then
    echo "Building transport binary..."
    cd "$HGCTL_ROOT"
    go build -o bin/transport ./cmd/transport
    cd "$AVS_PROJECT_DIR"
fi

CROSS_CHAIN_REGISTRY="0x287381B1570d9048c4B4C7EC94d21dDb8Aa1352a"

# Create transport config for executor operator set (ID 1)
cat > /tmp/transport-config.json <<EOF
{
  "transporterKey": "$avsAccountPk",
  "l1RpcUrl": "$L1_RPC_URL",
  "l1ChainId": $anvilL1ChainId,
  "l2RpcUrl": "$L2_RPC_URL",
  "l2ChainId": $anvilL2ChainId,
  "crossChainRegistry": "$CROSS_CHAIN_REGISTRY",
  "keyRegistrarAddress": "0xA4dB30D08d8bbcA00D40600bee9F029984dB162a",
  "avsAddress": "$AVS_ADDRESS",
  "operatorSetId": 1,
  "curveType": "ECDSA",
  "transportBlsKey": "$TRANSPORT_BLS_KEY",
  "operators": [
    {
      "address": "$execOperatorAccountAddress",
      "privateKey": "$execOperatorAccountPk"
    },
    {
      "address": "$execOperator2AccountAddress",
      "privateKey": "$execOperator2AccountPk"
    },
    {
      "address": "$execOperator3AccountAddress",
      "privateKey": "$execOperator3AccountPk"
    },
    {
      "address": "$execOperator4AccountAddress",
      "privateKey": "$execOperator4AccountPk"
    }
  ],
  "chainsToIgnore": [11155111, 84532]
}
EOF

echo "Transport config for operator set 1:"
cat /tmp/transport-config.json | jq .

echo "Running operator table transport..."
"$HGCTL_ROOT/bin/transport" -config /tmp/transport-config.json -v

if [ $? -eq 0 ]; then
    echo "Operator table transport completed successfully for operator set 1"
else
    echo "ERROR: Operator table transport failed for operator set 1"
    cleanup
    exit 1
fi

# Create transport config for executor operator set (ID 0)
cat > /tmp/transport-config.json <<EOF
{
  "transporterKey": "$avsAccountPk",
  "l1RpcUrl": "$L1_RPC_URL",
  "l1ChainId": $anvilL1ChainId,
  "l2RpcUrl": "$L2_RPC_URL",
  "l2ChainId": $anvilL2ChainId,
  "crossChainRegistry": "$CROSS_CHAIN_REGISTRY",
  "keyRegistrarAddress": "0xA4dB30D08d8bbcA00D40600bee9F029984dB162a",
  "avsAddress": "$AVS_ADDRESS",
  "operatorSetId": 0,
  "curveType": "ECDSA",
  "transportBlsKey": "$TRANSPORT_BLS_KEY",
  "operators": [
    {
      "address": "$operatorAccountAddress",
      "privateKey": "$operatorAccountPk"
    }
  ],
  "chainsToIgnore": [11155111, 84532]
}
EOF

echo "Transport config for operator set 0:"
cat /tmp/transport-config.json | jq .

echo "Running operator table transport..."
"$HGCTL_ROOT/bin/transport" -config /tmp/transport-config.json -v

if [ $? -eq 0 ]; then
    echo "Operator table transport completed successfully for operator set 0"
else
    echo "ERROR: Operator table transport failed for operator set 0"
    cleanup
    exit 1
fi

# Clean up temp config
rm -f /tmp/transport-config.json

cd "$AVS_PROJECT_DIR"

echo "Deploying L2 AVS contracts with DevKit..."
devkit avs deploy contracts l2

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
execOperator2AccountPublicKey=$(cast wallet public-key --private-key "0x$execOperator2AccountPk")
execOperator3AccountPublicKey=$(cast wallet public-key --private-key "0x$execOperator3AccountPk")
execOperator4AccountPublicKey=$(cast wallet public-key --private-key "0x$execOperator4AccountPk")
unregisteredOperator1AccountPublicKey=$(cast wallet public-key --private-key "0x$unregisteredOperator1AccountPk")
unregisteredOperator2AccountPublicKey=$(cast wallet public-key --private-key "0x$unregisteredOperator2AccountPk")
aggStakerAccountPublicKey=$(cast wallet public-key --private-key "0x$aggStakerAccountPk")
execStakerAccountPublicKey=$(cast wallet public-key --private-key "0x$execStakerAccountPk")
execStaker2AccountPublicKey=$(cast wallet public-key --private-key "0x$execStaker2AccountPk")
execStaker3AccountPublicKey=$(cast wallet public-key --private-key "0x$execStaker3AccountPk")
execStaker4AccountPublicKey=$(cast wallet public-key --private-key "0x$execStaker4AccountPk")
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
      "operatorSystemAddress": "$AGGREGATOR_SYSTEM_ADDRESS",
      "operatorSystemPk": "$AGGREGATOR_SYSTEM_PK",
      "operatorSystemKeystorePath": "$KEYS_DIR/aggregator-system-keystore.json",
      "operatorSystemKeystorePassword": "$AGGREGATOR_PASSWORD",
      "execOperatorAccountAddress": "$execOperatorAccountAddress",
      "execOperatorAccountPk": "$execOperatorAccountPk",
      "execOperatorAccountPublicKey": "$execOperatorAccountPublicKey",
      "execOperatorKeystorePath": "$KEYS_DIR/executor-keystore.json",
      "execOperatorKeystorePassword": "$EXECUTOR_PASSWORD",
      "execOperatorSystemAddress": "$EXECUTOR_SYSTEM_ADDRESS",
      "execOperatorSystemPk": "$EXECUTOR_SYSTEM_PK",
      "execOperatorSystemKeystorePath": "$KEYS_DIR/executor-system-keystore.json",
      "execOperatorSystemKeystorePassword": "$EXECUTOR_PASSWORD",
      "execOperator2AccountAddress": "$execOperator2AccountAddress",
      "execOperator2AccountPk": "$execOperator2AccountPk",
      "execOperator2AccountPublicKey": "$execOperator2AccountPublicKey",
      "execOperator2KeystorePath": "$KEYS_DIR/executor2-ecdsa-keystore.json",
      "execOperator2KeystorePassword": "$EXECUTOR_PASSWORD",
      "execOperator2SystemAddress": "$EXECUTOR2_SYSTEM_ADDRESS",
      "execOperator2SystemPk": "$EXECUTOR2_SYSTEM_PK",
      "execOperator2SystemKeystorePath": "$KEYS_DIR/executor2-system-keystore.json",
      "execOperator2SystemKeystorePassword": "$EXECUTOR_PASSWORD",
      "execOperator3AccountAddress": "$execOperator3AccountAddress",
      "execOperator3AccountPk": "$execOperator3AccountPk",
      "execOperator3AccountPublicKey": "$execOperator3AccountPublicKey",
      "execOperator3KeystorePath": "$KEYS_DIR/executor3-ecdsa-keystore.json",
      "execOperator3KeystorePassword": "$EXECUTOR_PASSWORD",
      "execOperator3SystemAddress": "$EXECUTOR3_SYSTEM_ADDRESS",
      "execOperator3SystemPk": "$EXECUTOR3_SYSTEM_PK",
      "execOperator3SystemKeystorePath": "$KEYS_DIR/executor3-system-keystore.json",
      "execOperator3SystemKeystorePassword": "$EXECUTOR_PASSWORD",
      "execOperator4AccountAddress": "$execOperator4AccountAddress",
      "execOperator4AccountPk": "$execOperator4AccountPk",
      "execOperator4AccountPublicKey": "$execOperator4AccountPublicKey",
      "execOperator4KeystorePath": "$KEYS_DIR/executor4-ecdsa-keystore.json",
      "execOperator4KeystorePassword": "$EXECUTOR_PASSWORD",
      "execOperator4SystemAddress": "$EXECUTOR4_SYSTEM_ADDRESS",
      "execOperator4SystemPk": "$EXECUTOR4_SYSTEM_PK",
      "execOperator4SystemKeystorePath": "$KEYS_DIR/executor4-system-keystore.json",
      "execOperator4SystemKeystorePassword": "$EXECUTOR_PASSWORD",
      "unregisteredOperator1AccountAddress": "$unregisteredOperator1AccountAddress",
      "unregisteredOperator1AccountPk": "$unregisteredOperator1AccountPk",
      "unregisteredOperator1AccountPublicKey": "$unregisteredOperator1AccountPublicKey",
      "unregisteredOperator1KeystorePath": "$KEYS_DIR/unregistered-operator1-ecdsa-keystore.json",
      "unregisteredOperator1KeystorePassword": "$EXECUTOR_PASSWORD",
      "unregisteredOperator1SystemBN254Pk": "$UNREGISTERED1_SYSTEM_BN254_PK",
      "unregisteredOperator1SystemBN254KeystorePath": "$KEYS_DIR/unregistered1-system-bn254-keystore.json",
      "unregisteredOperator1SystemBN254KeystorePassword": "$EXECUTOR_PASSWORD",
      "unregisteredOperator1SystemECDSAPk": "$UNREGISTERED1_SYSTEM_ECDSA_PK",
      "unregisteredOperator1SystemECDSAAddress": "$UNREGISTERED1_SYSTEM_ECDSA_ADDRESS",
      "unregisteredOperator1SystemECDSAKeystorePath": "$KEYS_DIR/unregistered1-system-ecdsa-keystore.json",
      "unregisteredOperator1SystemECDSAKeystorePassword": "$EXECUTOR_PASSWORD",
      "unregisteredOperator2AccountAddress": "$unregisteredOperator2AccountAddress",
      "unregisteredOperator2AccountPk": "$unregisteredOperator2AccountPk",
      "unregisteredOperator2AccountPublicKey": "$unregisteredOperator2AccountPublicKey",
      "unregisteredOperator2KeystorePath": "$KEYS_DIR/unregistered-operator2-ecdsa-keystore.json",
      "unregisteredOperator2KeystorePassword": "$EXECUTOR_PASSWORD",
      "unregisteredOperator2SystemBN254Pk": "$UNREGISTERED2_SYSTEM_BN254_PK",
      "unregisteredOperator2SystemBN254KeystorePath": "$KEYS_DIR/unregistered2-system-bn254-keystore.json",
      "unregisteredOperator2SystemBN254KeystorePassword": "$EXECUTOR_PASSWORD",
      "unregisteredOperator2SystemECDSAPk": "$UNREGISTERED2_SYSTEM_ECDSA_PK",
      "unregisteredOperator2SystemECDSAAddress": "$UNREGISTERED2_SYSTEM_ECDSA_ADDRESS",
      "unregisteredOperator2SystemECDSAKeystorePath": "$KEYS_DIR/unregistered2-system-ecdsa-keystore.json",
      "unregisteredOperator2SystemECDSAKeystorePassword": "$EXECUTOR_PASSWORD",
      "aggStakerAccountAddress": "$aggStakerAccountAddress",
      "aggStakerAccountPk": "$aggStakerAccountPk",
      "aggStakerAccountPublicKey": "$aggStakerAccountPublicKey",
      "execStakerAccountAddress": "$execStakerAccountAddress",
      "execStakerAccountPk": "$execStakerAccountPk",
      "execStakerAccountPublicKey": "$execStakerAccountPublicKey",
      "execStaker2AccountAddress": "$execStaker2AccountAddress",
      "execStaker2AccountPk": "$execStaker2AccountPk",
      "execStaker2AccountPublicKey": "$execStaker2AccountPublicKey",
      "execStaker3AccountAddress": "$execStaker3AccountAddress",
      "execStaker3AccountPk": "$execStaker3AccountPk",
      "execStaker3AccountPublicKey": "$execStaker3AccountPublicKey",
      "execStaker4AccountAddress": "$execStaker4AccountAddress",
      "execStaker4AccountPk": "$execStaker4AccountPk",
      "execStaker4AccountPublicKey": "$execStaker4AccountPublicKey",
      "avsTaskRegistrarAddress": "$avsTaskRegistrarAddress",
      "keyRegistrarAddress": "0xA4dB30D08d8bbcA00D40600bee9F029984dB162a",
      "releaseManagerAddress": "0xd9Cb89F1993292dEC2F973934bC63B0f2A702776",
      "delegationManagerAddress": "0xD4A7E1Bd8015057293f0D0A557088c286942e84b",
      "allocationManagerAddress": "0x42583067658071247ec8CE0A516A58f682002d07",
      "strategyManagerAddress": "0x2E3D6c0744b10eb0A4e6F679F71554a39Ec47a5D",
      "taskMailboxAddress": "0x132b466d9d5723531f68797519dfed701ac2c749",
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

echo "Cleaning up test avs content"
cd "$PROJECT_ROOT"
rm -rf integration-test-avs