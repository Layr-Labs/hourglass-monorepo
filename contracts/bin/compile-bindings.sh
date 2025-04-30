#!/bin/bash

set -e

RELEASE_ID=${1:-local}
BINDING_DIR=./pkg/bindings
JSON_DIR=./out
ABI_DIR=../ponos/contracts/abi/${RELEASE_ID}
CHAIN_MAP_FILE=../ponos/contracts/abi/${RELEASE_ID}/chains.json

function create_binding {
    contract_name=$1

    mkdir -p $BINDING_DIR/${contract_name}
    contract_json_path="${JSON_DIR}/${contract_name}.sol/${contract_name}.json"
    binding_out_dir="${BINDING_DIR}/${contract_name}"
    abi_out_path="${ABI_DIR}/${contract_name}.abi.json"

    mkdir -p $binding_out_dir || true
    mkdir -p $ABI_DIR

    cat $contract_json_path | jq -r '.abi' > $binding_out_dir/tmp.abi
    cat $contract_json_path | jq -r '.bytecode.object' > $binding_out_dir/tmp.bin

    cp $binding_out_dir/tmp.abi $abi_out_path

    abigen \
        --bin=$binding_out_dir/tmp.bin \
        --abi=$binding_out_dir/tmp.abi \
        --pkg="${contract_name}" \
        --out=$BINDING_DIR/$contract_name/binding.go \
        > /dev/null 2>&1

    if [[ $? == "1" ]]; then
        echo "Failed to generate binding for $contract_json_path"
    fi

    rm $binding_out_dir/tmp.abi
    rm $binding_out_dir/tmp.bin
}

contracts=$(find src -type f -name "*.sol" )
IFS=$'\n'
mkdir -p $JSON_DIR || true

for contract_name in $contracts; do
    contract_name=$(basename $contract_name .sol)
    create_binding $contract_name
done

CHAIN_IDS=${2:-31337}

chains_json="{}"

contract_names=()
for contract_path in $contracts; do
    contract_name=$(basename "$contract_path" .sol)
    contract_names+=("$contract_name")
done

for chain_id in $(echo "$CHAIN_IDS" | tr ',' '\n'); do
    contracts_json=$(printf '%s\n' "${contract_names[@]}" | jq -Rn '
        [inputs] | map({(.): "0x0000000000000000000000000000000000000000"}) | add
    ')
    chains_json=$(echo "$chains_json" | jq --argjson contracts "$contracts_json" --arg chain "$chain_id" '. + {($chain): $contracts}')
done

echo "$chains_json" | jq '.' > "$CHAIN_MAP_FILE"


