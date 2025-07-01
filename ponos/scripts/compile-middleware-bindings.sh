#!/bin/bash

pwd

BINDING_DIR=$(pwd)/pkg/middleware-bindings
JSON_DIR=./out

function create_binding {
    contract_name=$1
    echo "\tCompiling $contract_name"
    pwd
    mkdir -p $BINDING_DIR/${contract_name}

    contract_json_path="${JSON_DIR}/${contract_name}.sol/${contract_name}.json"

    binding_out_dir="${BINDING_DIR}/${contract_name}"
    mkdir -p $binding_out_dir || true

    cat $contract_json_path | jq -r '.abi' > $binding_out_dir/tmp.abi
    cat $contract_json_path | jq -r '.bytecode.object' > $binding_out_dir/tmp.bin

    abigen \
        --bin=$binding_out_dir/tmp.bin \
        --abi=$binding_out_dir/tmp.abi \
        --pkg="${contract_name}" \
        --out=$BINDING_DIR/$contract_name/binding.go \
        > /dev/null 2>&1

    if [[ $? == "1" ]];
    then
        echo "Failed to generate binding for $contract_json_path"
    fi
    rm $binding_out_dir/tmp.abi
    rm $binding_out_dir/tmp.bin
}

cd modules/eigenlayer-middleware

echo "Running forge build..."
forge build

contracts=$(find src -type f -name "*.sol" )
IFS=$'\n'


create_binding "IBN254TableCalculator"
create_binding "IECDSATableCalculator"

# echo "Generating bindings for contracts..."
# for contract_name in $contracts; do
# 	contract_name=$(basename $contract_name .sol)
# 	create_binding $contract_name
# done
