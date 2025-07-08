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
anvilL1StartBlock=8712202
anvilL1DumpStatePath=./anvil-l1.json
anvilL1ConfigPath=./anvil-l1-config.json
anvilL1RpcPort=8545
anvilL1RpcUrl="http://localhost:${anvilL1RpcPort}"


# base mainnet
L2_FORK_RPC_URL=https://soft-alpha-grass.base-sepolia.quiknode.pro/fd5e4bf346247d9b6e586008a9f13df72ce6f5b2/

anvilL2ChainId=31338
anvilL2StartBlock=28063258
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

aggStakerAccountAddress=$(echo $seedAccounts | jq -r '.[5].address')
aggStakerAccountPk=$(echo $seedAccounts | jq -r '.[5].private_key')
echo "Agg staker account: $aggStakerAccountAddress"
echo "Agg staker account private key: $aggStakerAccountPk"

execStakerAccountAddress=$(echo $seedAccounts | jq -r '.[6].address')
execStakerAccountPk=$(echo $seedAccounts | jq -r '.[6].private_key')
echo "Exec staker account: $execStakerAccountAddress"
echo "Exec staker account private key: $execStakerAccountPk"

echo $deployAccount
echo $deployAccountPk

cd ../contracts

export L1_RPC_URL="http://localhost:${anvilL1RpcPort}"
export L2_RPC_URL="http://localhost:${anvilL2RpcPort}"

# -----------------------------------------------------------------------------
# Deploy mailbox contract
# -----------------------------------------------------------------------------
# Cert Verifier addresses are the same address across all chains
bn254CertVerifierAddress="0x998535833f3feE44ce720440E735554699f728a5"
ecdsaCertVerifierAddress="0xAD2F58A551bD0e77fa20b5531dA96eF440C392BF"
echo "Deploying mailbox contract to L1..."
forge script script/local/DeployTaskMailbox.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run(address, address)" $bn254CertVerifierAddress $ecdsaCertVerifierAddress
mailboxContractAddressL1=$(cat ./broadcast/DeployTaskMailbox.s.sol/$anvilL1ChainId/run-latest.json | jq -r '.transactions[2].contractAddress')
echo "Mailbox contract address: $mailboxContractAddressL1"

echo "Deploying mailbox contract to L2..."
forge script script/local/DeployTaskMailbox.s.sol --slow --rpc-url $L2_RPC_URL --broadcast --sig "run(address, address)" $bn254CertVerifierAddress $ecdsaCertVerifierAddress
mailboxContractAddressL2=$(cat ./broadcast/DeployTaskMailbox.s.sol/$anvilL2ChainId/run-latest.json | jq -r '.transactions[2].contractAddress')
echo "Mailbox contract address: $mailboxContractAddressL2"

# -----------------------------------------------------------------------------
# Deploy L1 avs contract
# -----------------------------------------------------------------------------
echo "Deploying L1 AVS contract..."
forge script script/local/DeployAVSL1Contracts.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run(address)" "${avsAccountAddress}"

# we need to get index 2 since thats where the actual proxy lives
avsTaskRegistrarAddress=$(cat ./broadcast/DeployAVSL1Contracts.s.sol/$anvilL1ChainId/run-latest.json | jq -r '.transactions[2].contractAddress')
echo "L1 AVS contract address: $l1ContractAddress"

# ------------------------------------------------`-----------------------------
# Setup L1 AVS
# -----------------------------------------------------------------------------
echo "Setting up L1 AVS..."
forge script script/local/SetupAVSL1.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run(address)" $avsTaskRegistrarAddress

# -----------------------------------------------------------------------------
# Setup L1 multichain
# -----------------------------------------------------------------------------
echo "Setting up L1 AVS..."
export L1_CHAIN_ID=$anvilL1ChainId
export L2_CHAIN_ID=$anvilL2ChainId
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

# -----------------------------------------------------------------------------
# Create operators
# -----------------------------------------------------------------------------
echo "Registering operators"
export AGGREGATOR_PRIVATE_KEY=$operatorAccountPk
export EXECUTOR_PRIVATE_KEY=$execOperatorAccountPk
forge script script/local/SetupOperators.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run()"

# -----------------------------------------------------------------------------
# Stake some stuff
# -----------------------------------------------------------------------------
echo "Aggregator addres: ${operatorAccountAddress}"
echo "Exec aggregator address: ${execOperatorAccountAddress}"
echo "Agg staker address: ${aggStakerAccountAddress}"
echo "Exec staker address: ${execStakerAccountAddress}"

echo "Staking all the things"
export AGG_STAKER_PRIVATE_KEY=$aggStakerAccountPk
export EXEC_STAKER_PRIVATE_KEY=$execStakerAccountPk
forge script script/local/StakeStuff.s.sol --slow --rpc-url $L1_RPC_URL --broadcast --sig "run()" -vvvv

# move past the global ALLOCATION_CONFIGURATION_DELAY which is 75 blocks for sepolia
cast rpc --rpc-url $L1_RPC_URL anvil_mine 80
cast rpc --rpc-url $L2_RPC_URL anvil_mine 80

echo "Ended at block number: "
cast block-number

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

function lowercaseAddress() {
    echo "$1" | tr '[:upper:]' '[:lower:]'
}

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
      "aggStakerAccountAddress": "$aggStakerAccountAddress",
      "aggStakerAccountPk": "$aggStakerAccountPk",
      "execStakerAccountAddress": "$execStakerAccountAddress",
      "execStakerAccountPk": "$execStakerAccountPk",
      "mailboxContractAddressL1": "$mailboxContractAddressL1",
      "mailboxContractAddressL2": "$mailboxContractAddressL2",
      "avsTaskRegistrarAddress": "$avsTaskRegistrarAddress",
      "avsTaskHookAddressL1": "$taskHookAddressL1",
      "avsTaskHookAddressL2": "$taskHookAddressL2",
      "destinationEnv": "anvil",
      "forkL1Block": $anvilL1StartBlock,
      "forkL2Block": $anvilL2StartBlock
}
