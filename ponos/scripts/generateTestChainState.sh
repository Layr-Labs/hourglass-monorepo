#!/usr/bin/env bash

anvilL1Pid=""
anvilL2Pid=""

function cleanup() {
    kill $anvilL1Pid || true
    kill $anvilL2Pid || true

    exit $?
}
trap cleanup ERR
set -e

# ethereum holesky
L1_FORK_RPC_URL=https://practical-serene-mound.ethereum-sepolia.quiknode.pro/3aaa48bd95f3d6aed60e89a1a466ed1e2a440b61/

anvilL1ChainId=31337
anvilL1StartBlock=9085290
anvilL1DumpStatePath=./anvil-l1.json
anvilL1ConfigPath=./anvil-l1-config.json
anvilL1RpcPort=8545
anvilL1RpcUrl="http://localhost:${anvilL1RpcPort}"

# base mainnet
L2_FORK_RPC_URL=https://soft-alpha-grass.base-sepolia.quiknode.pro/fd5e4bf346247d9b6e586008a9f13df72ce6f5b2/

anvilL2ChainId=31338
anvilL2StartBlock=30327360
anvilL2DumpStatePath=./anvil-l2.json
anvilL2ConfigPath=./anvil-l2-config.json
anvilL2RpcPort=9545
anvilL2RpcUrl="http://localhost:${anvilL2RpcPort}"

# -----------------------------------------------------------------------------
# Generate new accounts instead of loading from file
# -----------------------------------------------------------------------------
echo "Generating new accounts for testing..."

# Function to generate a new account
generate_account() {
    local name=$1
    # Generate a new private key and address
    local wallet_output=$(cast wallet new)
    local address=$(echo "$wallet_output" | grep "Address:" | awk '{print $2}')
    local private_key=$(echo "$wallet_output" | grep "Private key:" | awk '{print $3}' | sed 's/0x//')
    echo "{\"name\": \"$name\", \"private_key\": \"$private_key\", \"address\": \"$address\"}"
}

# Generate all required accounts
DEPLOY_ACCOUNT=$(generate_account "deployer")
deployAccountPk=$(echo $DEPLOY_ACCOUNT | jq -r '.private_key')
deployAccountAddress=$(echo $DEPLOY_ACCOUNT | jq -r '.address')

AVS_ACCOUNT=$(generate_account "avs")
avsAccountPk=$(echo $AVS_ACCOUNT | jq -r '.private_key')
avsAccountAddress=$(echo $AVS_ACCOUNT | jq -r '.address')

APP_ACCOUNT=$(generate_account "app")
appAccountPk=$(echo $APP_ACCOUNT | jq -r '.private_key')
appAccountAddress=$(echo $APP_ACCOUNT | jq -r '.address')

OPERATOR_ACCOUNT=$(generate_account "operator")
operatorAccountPk=$(echo $OPERATOR_ACCOUNT | jq -r '.private_key')
operatorAccountAddress=$(echo $OPERATOR_ACCOUNT | jq -r '.address')

# Generate 4 executor operator accounts
EXEC_OPERATOR_ACCOUNT=$(generate_account "exec_operator")
execOperatorAccountPk=$(echo $EXEC_OPERATOR_ACCOUNT | jq -r '.private_key')
execOperatorAccountAddress=$(echo $EXEC_OPERATOR_ACCOUNT | jq -r '.address')

EXEC_OPERATOR_ACCOUNT2=$(generate_account "exec_operator2")
execOperator2AccountPk=$(echo $EXEC_OPERATOR_ACCOUNT2 | jq -r '.private_key')
execOperator2AccountAddress=$(echo $EXEC_OPERATOR_ACCOUNT2 | jq -r '.address')

EXEC_OPERATOR_ACCOUNT3=$(generate_account "exec_operator3")
execOperator3AccountPk=$(echo $EXEC_OPERATOR_ACCOUNT3 | jq -r '.private_key')
execOperator3AccountAddress=$(echo $EXEC_OPERATOR_ACCOUNT3 | jq -r '.address')

EXEC_OPERATOR_ACCOUNT4=$(generate_account "exec_operator4")
execOperator4AccountPk=$(echo $EXEC_OPERATOR_ACCOUNT4 | jq -r '.private_key')
execOperator4AccountAddress=$(echo $EXEC_OPERATOR_ACCOUNT4 | jq -r '.address')

AGG_STAKER_ACCOUNT=$(generate_account "agg_staker")
aggStakerAccountPk=$(echo $AGG_STAKER_ACCOUNT | jq -r '.private_key')
aggStakerAccountAddress=$(echo $AGG_STAKER_ACCOUNT | jq -r '.address')

# Generate 4 executor staker accounts
EXEC_STAKER_ACCOUNT=$(generate_account "exec_staker")
execStakerAccountPk=$(echo $EXEC_STAKER_ACCOUNT | jq -r '.private_key')
execStakerAccountAddress=$(echo $EXEC_STAKER_ACCOUNT | jq -r '.address')

EXEC_STAKER_ACCOUNT2=$(generate_account "exec_staker2")
execStaker2AccountPk=$(echo $EXEC_STAKER_ACCOUNT2 | jq -r '.private_key')
execStaker2AccountAddress=$(echo $EXEC_STAKER_ACCOUNT2 | jq -r '.address')

EXEC_STAKER_ACCOUNT3=$(generate_account "exec_staker3")
execStaker3AccountPk=$(echo $EXEC_STAKER_ACCOUNT3 | jq -r '.private_key')
execStaker3AccountAddress=$(echo $EXEC_STAKER_ACCOUNT3 | jq -r '.address')

EXEC_STAKER_ACCOUNT4=$(generate_account "exec_staker4")
execStaker4AccountPk=$(echo $EXEC_STAKER_ACCOUNT4 | jq -r '.private_key')
execStaker4AccountAddress=$(echo $EXEC_STAKER_ACCOUNT4 | jq -r '.address')

# Export environment variables (with 0x prefix for Forge compatibility)
export PRIVATE_KEY_DEPLOYER="0x$deployAccountPk"
export PRIVATE_KEY_AVS="0x$avsAccountPk"
export PRIVATE_KEY_APP="0x$appAccountPk"
export PRIVATE_KEY_OPERATOR="0x$operatorAccountPk"
export PRIVATE_KEY_EXEC_OPERATOR="0x$execOperatorAccountPk"

# Print generated accounts
echo "Generated accounts:"
echo "Deploy account: $deployAccountAddress"
echo "AVS account: $avsAccountAddress"
echo "App account: $appAccountAddress"
echo "Operator account: $operatorAccountAddress"
echo "Exec Operator 1: $execOperatorAccountAddress"
echo "Exec Operator 2: $execOperator2AccountAddress"
echo "Exec Operator 3: $execOperator3AccountAddress"
echo "Exec Operator 4: $execOperator4AccountAddress"
echo "Agg staker account: $aggStakerAccountAddress"
echo "Exec staker 1: $execStakerAccountAddress"
echo "Exec staker 2: $execStaker2AccountAddress"
echo "Exec staker 3: $execStaker3AccountAddress"
echo "Exec staker 4: $execStaker4AccountAddress"

# Save accounts to a file for reference
cat <<EOF > ./generated-accounts.json
[
  {
    "name": "deployer",
    "address": "$deployAccountAddress",
    "private_key": "$deployAccountPk"
  },
  {
    "name": "avs",
    "address": "$avsAccountAddress",
    "private_key": "$avsAccountPk"
  },
  {
    "name": "app",
    "address": "$appAccountAddress",
    "private_key": "$appAccountPk"
  },
  {
    "name": "operator",
    "address": "$operatorAccountAddress",
    "private_key": "$operatorAccountPk"
  },
  {
    "name": "exec_operator",
    "address": "$execOperatorAccountAddress",
    "private_key": "$execOperatorAccountPk"
  },
  {
    "name": "exec_operator2",
    "address": "$execOperator2AccountAddress",
    "private_key": "$execOperator2AccountPk"
  },
  {
    "name": "exec_operator3",
    "address": "$execOperator3AccountAddress",
    "private_key": "$execOperator3AccountPk"
  },
  {
    "name": "exec_operator4",
    "address": "$execOperator4AccountAddress",
    "private_key": "$execOperator4AccountPk"
  },
  {
    "name": "agg_staker",
    "address": "$aggStakerAccountAddress",
    "private_key": "$aggStakerAccountPk"
  },
  {
    "name": "exec_staker",
    "address": "$execStakerAccountAddress",
    "private_key": "$execStakerAccountPk"
  },
  {
    "name": "exec_staker2",
    "address": "$execStaker2AccountAddress",
    "private_key": "$execStaker2AccountPk"
  },
  {
    "name": "exec_staker3",
    "address": "$execStaker3AccountAddress",
    "private_key": "$execStaker3AccountPk"
  },
  {
    "name": "exec_staker4",
    "address": "$execStaker4AccountAddress",
    "private_key": "$execStaker4AccountPk"
  }
]
EOF

echo "Saved generated accounts to ./generated-accounts.json"

# Also save to anvilConfig directory for Web3Signer key generation
mkdir -p ../ponos/anvilConfig
mv ./generated-accounts.json ../ponos/anvilConfig/accounts.json
echo "Copied accounts to ../ponos/anvilConfig/accounts.json for Web3Signer key generation"

# -----------------------------------------------------------------------------
# Start Ethereum L1
# -----------------------------------------------------------------------------
anvil \
    --fork-url $L1_FORK_RPC_URL \
    --dump-state $anvilL1DumpStatePath \
    --config-out $anvilL1ConfigPath \
    --chain-id $anvilL1ChainId \
    --port $anvilL1RpcPort \
    --block-time 2 \
    --fork-block-number $anvilL1StartBlock &

anvilL1Pid=$!
sleep 3

# -----------------------------------------------------------------------------
# Start Base L2
# -----------------------------------------------------------------------------
anvil \
    --fork-url $L2_FORK_RPC_URL \
    --dump-state $anvilL2DumpStatePath \
    --config-out $anvilL2ConfigPath \
    --chain-id $anvilL2ChainId \
    --port $anvilL2RpcPort \
    --fork-block-number $anvilL2StartBlock &
anvilL2Pid=$!
sleep 3

function fundAccount() {
    address=$1
    echo "Funding address $address on L1"
    cast rpc --rpc-url $anvilL1RpcUrl anvil_setBalance $address '0x21E19E0C9BAB2400000' # 10,000 ETH

    echo "Funding address $address on L2"
    cast rpc --rpc-url $anvilL2RpcUrl anvil_setBalance $address '0x21E19E0C9BAB2400000' # 10,000 ETH
}

# Fund all generated accounts
echo "Funding generated accounts..."
fundAccount "$deployAccountAddress"
fundAccount "$avsAccountAddress"
fundAccount "$appAccountAddress"
fundAccount "$operatorAccountAddress"
fundAccount "$execOperatorAccountAddress"
fundAccount "$execOperator2AccountAddress"
fundAccount "$execOperator3AccountAddress"
fundAccount "$execOperator4AccountAddress"
fundAccount "$aggStakerAccountAddress"
fundAccount "$execStakerAccountAddress"
fundAccount "$execStaker2AccountAddress"
fundAccount "$execStaker3AccountAddress"
fundAccount "$execStaker4AccountAddress"

# Fund the special accounts used for table transport and whitelisting
fundAccount "0x8736311E6b706AfF3D8132Adf351387092802bA6"
fundAccount "0xb094Ba769b4976Dc37fC689A76675f31bc4923b0"

echo "All accounts funded with 10,000 ETH on both L1 and L2"

cd ../contracts

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
# Setup L1 AVS
# -----------------------------------------------------------------------------
echo "Setting up L1 AVS..."
forge script script/local/SetupAVSL1.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run(address)" $avsTaskRegistrarAddress


# -----------------------------------------------------------------------------
# Setup L1 multichain
# -----------------------------------------------------------------------------
echo "Setting up L1 multichain..."
export L1_CHAIN_ID=$anvilL1ChainId
export L2_CHAIN_ID=$anvilL2ChainId
cast rpc anvil_impersonateAccount "0xb094Ba769b4976Dc37fC689A76675f31bc4923b0" --rpc-url $L1_RPC_URL
forge script script/local/WhitelistDevnet.s.sol --slow --rpc-url $L1_RPC_URL --sender "0xb094Ba769b4976Dc37fC689A76675f31bc4923b0" --unlocked --broadcast --sig "run()"

# -----------------------------------------------------------------------------
# Deploy L2
# -----------------------------------------------------------------------------
echo "Deploying L2 contracts on L1..."
forge script script/local/DeployAVSL2Contracts.s.sol --slow --rpc-url $L1_RPC_URL --broadcast
taskHookAddressL1=$(cat ./broadcast/DeployAVSL2Contracts.s.sol/$anvilL1ChainId/run-latest.json | jq -r '.transactions[0].contractAddress')

echo "Deploying L2 contracts on L2..."
forge script script/local/DeployAVSL2Contracts.s.sol --slow --rpc-url $L2_RPC_URL --broadcast
taskHookAddressL2=$(cat ./broadcast/DeployAVSL2Contracts.s.sol/$anvilL2ChainId/run-latest.json | jq -r '.transactions[0].contractAddress')

# -----------------------------------------------------------------------------
# Create operators
# -----------------------------------------------------------------------------
echo "Registering operators"
export AGGREGATOR_PRIVATE_KEY="0x$operatorAccountPk"
export EXECUTOR_PRIVATE_KEY="0x$execOperatorAccountPk"
export EXECUTOR2_PRIVATE_KEY="0x$execOperator2AccountPk"
export EXECUTOR3_PRIVATE_KEY="0x$execOperator3AccountPk"
export EXECUTOR4_PRIVATE_KEY="0x$execOperator4AccountPk"
forge script script/local/SetupOperators.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run()"

# -----------------------------------------------------------------------------
# Stake some stuff
# -----------------------------------------------------------------------------
echo "Aggregator address: ${operatorAccountAddress}"
echo "Exec aggregator address: ${execOperatorAccountAddress}"
echo "Agg staker address: ${aggStakerAccountAddress}"
echo "Exec staker address: ${execStakerAccountAddress}"

echo "Staking all the things"
export AGG_STAKER_PRIVATE_KEY="0x$aggStakerAccountPk"
export EXEC_STAKER_PRIVATE_KEY="0x$execStakerAccountPk"
export EXEC_STAKER2_PRIVATE_KEY="0x$execStaker2AccountPk"
export EXEC_STAKER3_PRIVATE_KEY="0x$execStaker3AccountPk"
export EXEC_STAKER4_PRIVATE_KEY="0x$execStaker4AccountPk"
forge script script/local/StakeStuff.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run()" -vvvv

# move past the global ALLOCATION_CONFIGURATION_DELAY which is 75 blocks for sepolia
cast rpc --rpc-url $L1_RPC_URL anvil_mine 80
cast rpc --rpc-url $L2_RPC_URL anvil_mine 80

echo "Ended at block number: "
cast block-number --rpc-url $L1_RPC_URL

kill $anvilL1Pid
kill $anvilL2Pid
sleep 3

cd ../ponos

rm -rf ./internal/testData/anvil*.json

cp -R $anvilL1DumpStatePath internal/testData/anvil-l1-state.json
cp -R $anvilL1ConfigPath internal/testData/anvil-l1-config.json
cp -R $anvilL2DumpStatePath internal/testData/anvil-l2-state.json
cp -R $anvilL2ConfigPath internal/testData/anvil-l2-config.json

# make the files read-only since anvil likes to overwrite things
chmod 444 internal/testData/anvil*

rm $anvilL1DumpStatePath
rm $anvilL1ConfigPath
rm $anvilL2DumpStatePath
rm $anvilL2ConfigPath

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
aggStakerAccountPublicKey=$(cast wallet public-key --private-key "0x$aggStakerAccountPk")
execStakerAccountPublicKey=$(cast wallet public-key --private-key "0x$execStakerAccountPk")
execStaker2AccountPublicKey=$(cast wallet public-key --private-key "0x$execStaker2AccountPk")
execStaker3AccountPublicKey=$(cast wallet public-key --private-key "0x$execStaker3AccountPk")
execStaker4AccountPublicKey=$(cast wallet public-key --private-key "0x$execStaker4AccountPk")
deployAccountAddress=$(lowercaseAddress $deployAccountAddress)

# create a heredoc json file and dump it to internal/testData/chain-config.json
cat <<EOF > internal/testData/chain-config.json
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
      "execOperatorAccountAddress": "$execOperatorAccountAddress",
      "execOperatorAccountPk": "$execOperatorAccountPk",
      "execOperatorAccountPublicKey": "$execOperatorAccountPublicKey",
      "execOperator2AccountAddress": "$execOperator2AccountAddress",
      "execOperator2AccountPk": "$execOperator2AccountPk",
      "execOperator2AccountPublicKey": "$execOperator2AccountPublicKey",
      "execOperator3AccountAddress": "$execOperator3AccountAddress",
      "execOperator3AccountPk": "$execOperator3AccountPk",
      "execOperator3AccountPublicKey": "$execOperator3AccountPublicKey",
      "execOperator4AccountAddress": "$execOperator4AccountAddress",
      "execOperator4AccountPk": "$execOperator4AccountPk",
      "execOperator4AccountPublicKey": "$execOperator4AccountPublicKey",
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
      "destinationEnv": "anvil",
      "forkL1Block": $anvilL1StartBlock,
      "forkL2Block": $anvilL2StartBlock
}
EOF

echo "Test chain state generated successfully!"
echo "Generated accounts saved to: ./generated-accounts.json"
echo "Chain config saved to: internal/testData/chain-config.json"