#!/bin/bash
set -e

RELEASE_ID=${1:-local}
CHAIN_ID=${2:-31337}
RPC_URL=${3:-http://localhost:8545}
BROADCAST_DIR="./broadcast/DeployTaskMailbox.s.sol/${CHAIN_ID}"

# Step 1: Deploy
forge script ./script/DeployTaskMailbox.s.sol \
  --rpc-url $RPC_URL \
  --broadcast \
  --chain-id $CHAIN_ID

# Step 2: Extract deployed contract addresses
DEPLOYED_JSON="${BROADCAST_DIR}/run-latest.json"

# Get contract map from broadcast file
jq -r '
  .transactions
  | map(select(.contractName != null))
  | map({ (.contractName): .contractAddress })
  | add
' $DEPLOYED_JSON > deployed-addresses.json

# Step 3: Convert to chainId => contract map
jq --arg chain "$CHAIN_ID" '{ ($chain): . }' deployed-addresses.json > full-chains.json

# Step 4: Call compile-bindings.sh and overwrite chains.json
./bin/compile-bindings.sh $RELEASE_ID $CHAIN_ID
cp full-chains.json ./../ponos/contracts/abi/${RELEASE_ID}/chains.json
