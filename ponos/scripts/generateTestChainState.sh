#!/usr/bin/env bash

# ethereum mainnet
FORK_RPC_URL=https://tame-fabled-liquid.quiknode.pro/f27d4be93b4d7de3679f5c5ae881233f857407a0/

# launc h anvil to generate accounts and dump them to a file
anvil \
    --fork-url $FORK_RPC_URL \
    --dump-state ./anvil.json \
    --config-out ./anvil-config.json \
    --chain-id 31337 \
    --fork-block-number 22396947 &

anvilPid=$!
sleep 3
kill $anvilPid

echo "Parsing accounts"
# grab the first account from the anvil output
anvilConfig=$(cat anvil-config.json | jq '.')
anvilState=$(cat anvil.json | jq '.')

cat anvil-config.json | jq '.'


echo "re-loading anvil with state"
anvil \
    --fork-url $FORK_RPC_URL \
    --dump-state ./anvil-final.json \
    --config-out ./anvil-config-final.json \
    --chain-id 31337 \
    --fork-block-number 22396947 &
anvilPid=$!

sleep 3

# deployer account
deployAccountAddress=$(echo $anvilConfig | jq -r '.available_accounts[0]')
deployAccountPk=$(echo $anvilConfig | jq -r '.private_keys[0]')
export PRIVATE_KEY_DEPLOYER=$deployAccountPk
echo "Deploy account: $deployAccountAddress"
echo "Deploy account private key: $deployAccountPk"

# avs account
avsAccountAddress=$(echo $anvilConfig | jq -r '.available_accounts[1]')
avsAccountPk=$(echo $anvilConfig | jq -r '.private_keys[1]')
export PRIVATE_KEY_AVS=$avsAccountPk
echo "AVS account: $avsAccountAddress"
echo "AVS account private key: $avsAccountPk"

# app account
appAccountAddress=$(echo $anvilConfig | jq -r '.available_accounts[2]')
appAccountPk=$(echo $anvilConfig | jq -r '.private_keys[2]')
export PRIVATE_KEY_APP=$appAccountPk
echo "App account: $appAccountAddress"
echo "App account private key: $appAccountPk"

operatorAccountAddress=$(echo $anvilConfig | jq -r '.available_accounts[3]')
operatorAccountPk=$(echo $anvilConfig | jq -r '.private_keys[3]')
export PRIVATE_KEY_OPERATOR=$operatorAccountPk
echo "Operator account: $operatorAccountAddress"
echo "Operator account private key: $operatorAccountPk"

execOperatorAccountAddress=$(echo $anvilConfig | jq -r '.available_accounts[4]')
execOperatorAccountPk=$(echo $anvilConfig | jq -r '.private_keys[4]')
export PRIVATE_KEY_EXEC_OPERATOR=$appAccountPk
echo "Exec Operator account: $execOperatorAccountAddress"
echo "Exec Operator account private key: $execOperatorAccountPk"

echo $deployAccount
echo $deployAccountPk

# Get the ChainID from the anvil fork
chainId=$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' http://localhost:8545  | jq -r '.result' | xargs printf "%d\n")

echo "Chain ID: $chainId"

cd ../contracts

export RPC_URL="http://localhost:8545"

# -----------------------------------------------------------------------------
# Deploy mailbox contract
# -----------------------------------------------------------------------------
echo "Deploying mailbox contract..."
forge script script/local/DeployTaskMailbox.s.sol --slow --rpc-url $RPC_URL --broadcast

mailboxContractAddress=$(cat ./broadcast/DeployTaskMailbox.s.sol/$chainId/run-latest.json | jq -r '.transactions[0].contractAddress')
echo "Mailbox contract address: $mailboxContractAddress"

# -----------------------------------------------------------------------------
# Deploy L1 avs contract
# -----------------------------------------------------------------------------
echo "Deploying L1 AVS contract..."
forge script script/local/DeployAVSL1Contracts.s.sol --slow --rpc-url $RPC_URL --broadcast --sig "run(address)" "${avsAccountAddress}"

avsTaskRegistrarAddress=$(cat ./broadcast/DeployAVSL1Contracts.s.sol/$chainId/run-latest.json | jq -r '.transactions[0].contractAddress')
echo "L1 AVS contract address: $l1ContractAddress"

# -----------------------------------------------------------------------------
# Setup L1 AVS
# -----------------------------------------------------------------------------
echo "Setting up L1 AVS..."
forge script script/local/SetupAVSL1.s.sol --slow --rpc-url $RPC_URL --broadcast --sig "run(address)" $avsTaskRegistrarAddress

# -----------------------------------------------------------------------------
# Deploy L2
# -----------------------------------------------------------------------------
echo "Deploying L2 contracts..."
forge script script/local/DeployAVSL2Contracts.s.sol --slow --rpc-url $RPC_URL --broadcast
taskHookAddress=$(cat ./broadcast/DeployAVSL2Contracts.s.sol/$chainId/run-latest.json | jq -r '.transactions[0].contractAddress')
certificateVerifierAddress=$(cat ./broadcast/DeployAVSL2Contracts.s.sol/$chainId/run-latest.json | jq -r '.transactions[1].contractAddress')

# -----------------------------------------------------------------------------
# Setup L1 task mailbox config
# -----------------------------------------------------------------------------
echo "Setting up L1 AVS..."
forge script script/local/SetupAVSTaskMailboxConfig.s.sol --slow --rpc-url $RPC_URL --broadcast --sig "run(address, address, address)" $mailboxContractAddress $certificateVerifierAddress $taskHookAddress

# -----------------------------------------------------------------------------
# Create test task
# -----------------------------------------------------------------------------
# forge script script/CreateTask.s.sol --rpc-url $RPC_URL --broadcast --sig "run(address, address)" $mailboxContractAddress $avsAccountAddress

kill $anvilPid
sleep 3

cd ../ponos

rm -rf ./internal/testData/anvil*.json

cp -R ./anvil-final.json internal/testData/anvil-state.json
cp -R ./anvil-config-final.json internal/testData/anvil-config.json

# make the files read-only since anvil likes to overwrite things
chmod 444 internal/testData/anvil*

rm ./anvil-final.json
rm ./anvil-config-final.json
rm ./anvil.json
rm ./anvil-config.json

# create a heredoc json file and dump it to internal/testData/chain-config.json
cat <<EOF > internal/testData/chain-config.json
{
  "deployAccountAddress": "$deployAccountAddress",
  "deployAccountPk": "$deployAccountPk",
  "avsAccountAddress": "$avsAccountAddress",
  "avsAccountPk": "$avsAccountPk",
  "appAccountAddress": "$appAccountAddress",
  "appAccountPk": "$appAccountPk",
  "operatorAccountAddress": "$operatorAccountAddress",
  "operatorAccountPk": "$operatorAccountPk",
  "execOperatorAccountAddress": "$execOperatorAccountAddress",
  "execOperatorAccountPk": "$execOperatorAccountPk",
  "mailboxContractAddress": "$mailboxContractAddress",
  "avsTaskRegistrarAddress": "$avsTaskRegistrarAddress",
  "taskHookAddress": "$taskHookAddress",
  "certificateVerifierAddress": "$certificateVerifierAddress",
  "destinationEnv": "anvil"
}
