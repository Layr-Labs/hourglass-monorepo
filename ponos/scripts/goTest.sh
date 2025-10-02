#!/usr/bin/env bash

set -e  # Exit on any error
set -u  # Exit on undefined variables
set -o pipefail  # Exit if any command in a pipe fails

# spin up a new web3signer docker container for L1
web3signerL1Name="web3signer-l1"
web3signerL1Port=9100
web3signerL1HttpPort=9101
web3signerL1ChainId=1

web3signerL2Name="web3signer-l2"
web3signerL2Port=9200
web3signerL2HttpPort=9201
web3signerL2ChainId=8453

cleanup_containers() {
    echo "Cleaning up containers..."
    docker rm -f $web3signerL1Name || true
    docker rm -f $web3signerL2Name || true
}

trap cleanup_containers ERR EXIT SIGINT SIGTERM

function runWeb3SignerContainer() {
    local name=$1
    local port=$2
    local chainId=$3

    docker run \
        --rm \
        --name $name \
        -v ./internal/testData/web3signer:/web3signer \
        -i \
        -p "${port}:${port}" \
        --detach \
         consensys/web3signer:develop \
            --key-store-path=/web3signer/keys \
            --http-listen-port=$port \
            eth1 \
            --chain-id $chainId \
            --keystores-path=/web3signer/keystores \
            --keystores-passwords-path=/web3signer/passwords
}

runWeb3SignerContainer $web3signerL1Name $web3signerL1Port $web3signerL1ChainId
runWeb3SignerContainer $web3signerL2Name $web3signerL2Port $web3signerL2ChainId

echo "Sleeping to let web3signer containers start..."
sleep 3


# run the tests
# GOFLAGS="-count=1" $(GO) test -v -p 1 -parallel 1 ./...
# Set a longer timeout for integration tests (default is 10m)
go test -timeout 35m $@
