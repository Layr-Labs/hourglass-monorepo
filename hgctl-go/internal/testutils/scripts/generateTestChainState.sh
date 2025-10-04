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
L1_FORK_RPC_URL=https://chaotic-delicate-needle.ethereum-sepolia.quiknode.pro/de95f687aeb82f1e7dc579e7fa5c698931ff2c57/

anvilL1ChainId=31337
anvilL1StartBlock=23477799
anvilL1DumpStatePath=$HGCTL_ROOT/internal/testdata/anvil-l1-state.json
anvilL1ConfigPath=$HGCTL_ROOT/internal/testdata/anvil-l1-config.json
anvilL1RpcPort=8545
anvilL1RpcUrl="http://localhost:${anvilL1RpcPort}"

# base mainnet
L2_FORK_RPC_URL=https://dimensional-compatible-sky.base-sepolia.quiknode.pro/2bd3461589b9a684678254bbd88bccbd65c34b84/

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

# -----------------------------------------------------------------------------
# Load accounts from anvilConfig/accounts.json for reproducible testing
# -----------------------------------------------------------------------------
echo "Loading accounts from anvilConfig/accounts.json..."

ACCOUNTS_FILE="$PROJECT_ROOT/ponos/anvilConfig/accounts.json"

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

# Load unregistered operator (for testing registration flow)
unregisteredOperatorAccountPk=$(echo $seedAccounts | jq -r '.[] | select(.name == "unregistered_operator") | .private_key')
unregisteredOperatorAccountAddress=$(echo $seedAccounts | jq -r '.[] | select(.name == "unregistered_operator") | .address')

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
echo "  Unregistered Operator: $unregisteredOperatorAccountAddress"
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

echo "Creating ECDSA keystore for executor 1..."
create_keystore "executor-ecdsa" "ecdsa" "$EXECUTOR_ECDSA_PK" "$EXECUTOR_PASSWORD"

echo "Creating ECDSA keystore for executor 2..."
create_keystore "executor2-ecdsa" "ecdsa" "$execOperator2AccountPk" "$EXECUTOR_PASSWORD"

echo "Creating ECDSA keystore for executor 3..."
create_keystore "executor3-ecdsa" "ecdsa" "$execOperator3AccountPk" "$EXECUTOR_PASSWORD"

echo "Creating ECDSA keystore for executor 4..."
create_keystore "executor4-ecdsa" "ecdsa" "$execOperator4AccountPk" "$EXECUTOR_PASSWORD"

echo "Creating ECDSA keystore for unregistered operator..."
create_keystore "unregistered-operator-ecdsa" "ecdsa" "$unregisteredOperatorAccountPk" "$EXECUTOR_PASSWORD"

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
copy_keystore "executor2-ecdsa" "$KEYS_DIR/executor2-ecdsa-keystore.json"
copy_keystore "executor3-ecdsa" "$KEYS_DIR/executor3-ecdsa-keystore.json"
copy_keystore "executor4-ecdsa" "$KEYS_DIR/executor4-ecdsa-keystore.json"
copy_keystore "unregistered-operator-ecdsa" "$KEYS_DIR/unregistered-operator-ecdsa-keystore.json"

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

# Fund CrossChainRegistry owner for impersonation
fundAccount "0xBE1685C81aA44FF9FB319dD389addd9374383e90"

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
# Deploy and Setup L1 AVS with ManuallySetupAvsMainnet
# This single script does:
# 1. Deploys TaskAVSRegistrar contract (ProxyAdmin + Implementation + Proxy)
# 2. Creates operator sets (0 for aggregator, 1 for executors)
# 3. Configures operator sets in KeyRegistrar for ECDSA
# 4. Publishes operator set metadata URIs
# 5. Registers for multichain support via CrossChainRegistry
# -----------------------------------------------------------------------------
echo "Deploying and setting up L1 AVS with ManuallySetupAvsMainnet..."
export AVS_PRIVATE_KEY="0x$avsAccountPk"
forge script script/local/ManuallySetupAvsMainnet.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run(uint32,uint32)" 0 1

# Extract the TaskAVSRegistrar proxy address from the broadcast
avsTaskRegistrarAddress=$(cat ./broadcast/ManuallySetupAvsMainnet.s.sol/$anvilL1ChainId/run-latest.json | jq -r '[.transactions[] | select(.transactionType == "CREATE")] | .[-1].contractAddress')
echo "L1 AVS TaskAVSRegistrar address: $avsTaskRegistrarAddress"

# -----------------------------------------------------------------------------
# Allowlist aggregator operator for operator set 0
# -----------------------------------------------------------------------------
echo "Allowlisting aggregator operator for operator set 0..."
export AGGREGATOR_PRIVATE_KEY="0x$operatorAccountPk"
forge script script/local/AllowlistOperators.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run(address)" "$avsTaskRegistrarAddress"

# -----------------------------------------------------------------------------
# Register operators with EigenLayer using hgctl CLI
# -----------------------------------------------------------------------------
echo "Registering aggregator operator with EigenLayer..."
"$HGCTL" eigenlayer register-operator \
    --keystore "$KEYS_DIR/aggregator-ecdsa-keystore.json" \
    --password "$AGGREGATOR_PASSWORD" \
    --metadata-uri "https://example.com/aggregator/metadata.json" \
    --allocation-delay "0"

echo "Registering executor operator 1 with EigenLayer..."
"$HGCTL" eigenlayer register-operator \
    --keystore "$KEYS_DIR/executor-ecdsa-keystore.json" \
    --password "$EXECUTOR_PASSWORD" \
    --metadata-uri "https://example.com/executor1/metadata.json" \
    --allocation-delay "0"

echo "Registering executor operator 2 with EigenLayer..."
"$HGCTL" eigenlayer register-operator \
    --keystore "$KEYS_DIR/executor2-ecdsa-keystore.json" \
    --password "$EXECUTOR_PASSWORD" \
    --metadata-uri "https://example.com/executor2/metadata.json" \
    --allocation-delay "0"

echo "Registering executor operator 3 with EigenLayer..."
"$HGCTL" eigenlayer register-operator \
    --keystore "$KEYS_DIR/executor3-ecdsa-keystore.json" \
    --password "$EXECUTOR_PASSWORD" \
    --metadata-uri "https://example.com/executor3/metadata.json" \
    --allocation-delay "0"

echo "Registering executor operator 4 with EigenLayer..."
"$HGCTL" eigenlayer register-operator \
    --keystore "$KEYS_DIR/executor4-ecdsa-keystore.json" \
    --password "$EXECUTOR_PASSWORD" \
    --metadata-uri "https://example.com/executor4/metadata.json" \
    --allocation-delay "0"

# -----------------------------------------------------------------------------
# Stake tokens for operators (WETH for aggregator, stETH for executors)
# -----------------------------------------------------------------------------
echo "Staking tokens for operators..."
echo "  Aggregator: $operatorAccountAddress (staker: $aggStakerAccountAddress)"
echo "  Exec Operator 1: $execOperatorAccountAddress (staker: $execStakerAccountAddress)"
echo "  Exec Operator 2: $execOperator2AccountAddress (staker: $execStaker2AccountAddress)"
echo "  Exec Operator 3: $execOperator3AccountAddress (staker: $execStaker3AccountAddress)"
echo "  Exec Operator 4: $execOperator4AccountAddress (staker: $execStaker4AccountAddress)"

export AGG_STAKER_PRIVATE_KEY="0x$aggStakerAccountPk"
export EXEC_STAKER_PRIVATE_KEY="0x$execStakerAccountPk"
export EXEC_STAKER2_PRIVATE_KEY="0x$execStaker2AccountPk"
export EXEC_STAKER3_PRIVATE_KEY="0x$execStaker3AccountPk"
export EXEC_STAKER4_PRIVATE_KEY="0x$execStaker4AccountPk"
forge script script/local/StakeStuff.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run()"

# Mine blocks to bypass ALLOCATION_CONFIGURATION_DELAY
echo "Mining blocks to finalize allocations..."
cast rpc --rpc-url $L1_RPC_URL anvil_mine 80
cast rpc --rpc-url $L2_RPC_URL anvil_mine 80

# -----------------------------------------------------------------------------
# Setup L1 multichain
# -----------------------------------------------------------------------------
echo "Setting up L1 multichain..."
export L1_CHAIN_ID=$anvilL1ChainId
export L2_CHAIN_ID=$anvilL2ChainId
export AVS_ADDRESS=$avsAccountAddress

# Whitelist chains
cast rpc anvil_impersonateAccount "0xBE1685C81aA44FF9FB319dD389addd9374383e90" --rpc-url $L1_RPC_URL
forge script script/local/WhitelistDevnet.s.sol --slow --rpc-url $L1_RPC_URL --sender "0xBE1685C81aA44FF9FB319dD389addd9374383e90" --unlocked --broadcast --sig "run()"

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
execOperator2AccountPublicKey=$(cast wallet public-key --private-key "0x$execOperator2AccountPk")
execOperator3AccountPublicKey=$(cast wallet public-key --private-key "0x$execOperator3AccountPk")
execOperator4AccountPublicKey=$(cast wallet public-key --private-key "0x$execOperator4AccountPk")
unregisteredOperatorAccountPublicKey=$(cast wallet public-key --private-key "0x$unregisteredOperatorAccountPk")
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
      "execOperatorAccountAddress": "$execOperatorAccountAddress",
      "execOperatorAccountPk": "$execOperatorAccountPk",
      "execOperatorAccountPublicKey": "$execOperatorAccountPublicKey",
      "execOperatorKeystorePath": "$KEYS_DIR/executor-keystore.json",
      "execOperatorKeystorePassword": "$EXECUTOR_PASSWORD",
      "execOperator2AccountAddress": "$execOperator2AccountAddress",
      "execOperator2AccountPk": "$execOperator2AccountPk",
      "execOperator2AccountPublicKey": "$execOperator2AccountPublicKey",
      "execOperator2KeystorePath": "$KEYS_DIR/executor2-ecdsa-keystore.json",
      "execOperator2KeystorePassword": "$EXECUTOR_PASSWORD",
      "execOperator3AccountAddress": "$execOperator3AccountAddress",
      "execOperator3AccountPk": "$execOperator3AccountPk",
      "execOperator3AccountPublicKey": "$execOperator3AccountPublicKey",
      "execOperator3KeystorePath": "$KEYS_DIR/executor3-ecdsa-keystore.json",
      "execOperator3KeystorePassword": "$EXECUTOR_PASSWORD",
      "execOperator4AccountAddress": "$execOperator4AccountAddress",
      "execOperator4AccountPk": "$execOperator4AccountPk",
      "execOperator4AccountPublicKey": "$execOperator4AccountPublicKey",
      "execOperator4KeystorePath": "$KEYS_DIR/executor4-ecdsa-keystore.json",
      "execOperator4KeystorePassword": "$EXECUTOR_PASSWORD",
      "unregisteredOperatorAccountAddress": "$unregisteredOperatorAccountAddress",
      "unregisteredOperatorAccountPk": "$unregisteredOperatorAccountPk",
      "unregisteredOperatorAccountPublicKey": "$unregisteredOperatorAccountPublicKey",
      "unregisteredOperatorKeystorePath": "$KEYS_DIR/unregistered-operator-ecdsa-keystore.json",
      "unregisteredOperatorKeystorePassword": "$EXECUTOR_PASSWORD",
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
      "avsTaskHookAddressL1": "$taskHookAddressL1",
      "avsTaskHookAddressL2": "$taskHookAddressL2",
      "keyRegistrarAddress": "0x54f4bC6bDEbe479173a2bbDc31dD7178408A57A4",
      "releaseManagerAddress": "0xeDA3CAd031c0cf367cF3f517Ee0DC98F9bA80C8F",
      "delegationManagerAddress": "0x39053D51B77DC0d36036Fc1fCc8Cb819df8Ef37A",
      "allocationManagerAddress": "0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39",
      "strategyManagerAddress": "0x858646372CC42E1A627fcE94aa7A7033e7CF075A",
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