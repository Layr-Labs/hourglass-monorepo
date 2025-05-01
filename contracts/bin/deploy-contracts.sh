#!/bin/bash
set -e

VERSION=${1:-local}
CHAIN_ID=${2:-31337}
RPC_URL=${3:-http://localhost:8545}
BROADCAST_DIR="./broadcast/DeployTaskMailbox.s.sol/${CHAIN_ID}"

# Step 1: Deploy
forge script ./script/DeployTaskMailbox.s.sol \
  --rpc-url "$RPC_URL" \
  --broadcast \
  --chain-id "$CHAIN_ID"

# Step 2: Extract deployed contract addresses
DEPLOYED_JSON="${BROADCAST_DIR}/run-latest.json"

# Get contract map from broadcast file
jq -r '
  .transactions
  | map(select(.contractName != null))
  | map({ (.contractName): .contractAddress })
  | add
' "$DEPLOYED_JSON" > deployed-addresses.json

# Step 3: Create new format: [{"chainId": ..., "contracts": { "<version>": { ... }}}]
jq --arg version "$VERSION" '{($version): .}' deployed-addresses.json \
  | jq --argjson chainId "$CHAIN_ID" '{chainId: $chainId, contracts: .}' \
  | jq -s '.' > chain-contracts.json

ABI_OUT_DIR="./../ponos/contracts/abi/"
mkdir -p "${ABI_OUT_DIR}"

# Move full-chains.json into final location and rename
mv chain-contracts.json "${ABI_OUT_DIR}/chain-contracts.json"

# Remove the temporary deployed-addresses file
rm -f deployed-addresses.json

# Step 4: Compile bindings and copy the updated chains.json
./bin/compile-bindings.sh

# Get known .sol contract names (no extension)
contracts=$(find src -type f -name "*.sol" -exec basename {} .sol \;)
IFS=$'\n'

for contract_json in ./out/*.sol/*.json; do
  contract_file=$(basename "$contract_json")
  contract_name="${contract_file%.json}"

  mkdir -p "${ABI_OUT_DIR}/${VERSION}/"
  # Only copy if contract_name is in our src/ list
  if echo "$contracts" | grep -qx "$contract_name"; then
    jq '.abi' "$contract_json" > "${ABI_OUT_DIR}/${VERSION}/${contract_name}.abi.json"
  fi
done
