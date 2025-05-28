#!/usr/bin/env bash

set -e

# ethereum mainnet
L1_FORK_RPC_URL=https://tame-fabled-liquid.quiknode.pro/f27d4be93b4d7de3679f5c5ae881233f857407a0/

anvilL1ChinId=31337
anvilL1StartBlock=22396947
anvilL1DumpStatePath=./anvil-l1.json
anvilL1ConfigPath=./anvil-l1-config.json
anvilL1RpcPort=8545
anvilL1RpcUrl="http://localhost:${anvilL1RpcPort}"


# base mainnet
L2_FORK_RPC_URL=https://few-sly-dew.base-mainnet.quiknode.pro/eaecd36554bb2845570742c4e7aeda6f7dd0d5c1/

anvilL2ChinId=31338
anvilL2StartBlock=30611001
anvilL2DumpStatePath=./anvil-l2.json
anvilL2ConfigPath=./anvil-l2-config.json
anvilL2RpcPort=9545
anvilL2RpcUrl="http://localhost:${anvilL2RpcPort}"

seedAccounts=$(cat ./anvilConfig/accounts.json)

# -----------------------------------------------------------------------------
# Start Ethereum L1
# -----------------------------------------------------------------------------
anvil \
    --fork-url $L1_FORK_RPC_URL \
    --dump-state $anvilL1DumpStatePath \
    --config-out $anvilL1ConfigPath \
    --chain-id $anvilL1ChinId \
    --port $anvilL1RpcPort \
    --block-time 2 \
    --fork-block-number $anvilL1StartBlock &

anvilL1Pid=$!
sleep 3

# -----------------------------------------------------------------------------
# Start Base L2
# -----------------------------------------------------------------------------
# anvil \
#     --fork-url $L2_FORK_RPC_URL \
#     --dump-state $anvilL2DumpStatePath \
#     --config-out $anvilL2DumpStatePath \
#     --chain-id $anvilL2ChinId \
#     --fork-block-number $anvilL2StartBlock &
# anvilL2Pid=$!
# sleep 3

# loop over the seed accounts (json array) and fund the accounts
numAccounts=$(echo $seedAccounts | jq '. | length - 1')
for i in $(seq 0 $numAccounts); do
    account=$(echo $seedAccounts | jq -r ".[$i]")
    address=$(echo $account | jq -r '.address')
    echo "Funding address $address"
    cast rpc --rpc-url $anvilL1RpcUrl anvil_setBalance $address '0x21E19E0C9BAB2400000' # 10,000 ETH
    echo "Account $address funded with 10,000 ETH"
done


# deployer account
deployAccountAddress=$(echo $seedAccounts | jq -r '.[0].address')
deployAccountPk=$(echo $seedAccounts | jq -r '.[0].private_key')
export PRIVATE_KEY_DEPLOYER=$deployAccountPk
echo "Deploy account: $deployAccountAddress"
echo "Deploy account private key: $deployAccountPk"

# avs account
avsAccountAddress=$(echo $seedAccounts | jq -r '.[1].address')
avsAccountPk=$(echo $seedAccounts | jq -r '.[1].private_key')
export PRIVATE_KEY_AVS=$avsAccountPk
echo "AVS account: $avsAccountAddress"
echo "AVS account private key: $avsAccountPk"

# app account
appAccountAddress=$(echo $seedAccounts | jq -r '.[2].address')
appAccountPk=$(echo $seedAccounts | jq -r '.[2].private_key')
export PRIVATE_KEY_APP=$appAccountPk
echo "App account: $appAccountAddress"
echo "App account private key: $appAccountPk"

operatorAccountAddress=$(echo $seedAccounts | jq -r '.[3].address')
operatorAccountPk=$(echo $seedAccounts | jq -r '.[3].private_key')
export PRIVATE_KEY_OPERATOR=$operatorAccountPk
echo "Operator account: $operatorAccountAddress"
echo "Operator account private key: $operatorAccountPk"

execOperatorAccountAddress=$(echo $seedAccounts | jq -r '.[4].address')
execOperatorAccountPk=$(echo $seedAccounts | jq -r '.[4].private_key')
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

kill $anvilL1Pid
sleep 3

cd ../ponos

rm -rf ./internal/testData/anvil*.json

cp -R $anvilL1DumpStatePath internal/testData/anvil-l1-state.json
cp -R $anvilL1ConfigPath internal/testData/anvil-l1-config.json

# make the files read-only since anvil likes to overwrite things
chmod 444 internal/testData/anvil*

rm $anvilL1DumpStatePath
rm $anvilL1ConfigPath

# create a heredoc json file and dump it to internal/testData/chain-config.json
cat <<EOF > internal/testData/chain-config.json
{
  "l1": {
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
}
