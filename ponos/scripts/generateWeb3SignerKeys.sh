#!/usr/bin/env bash

seedAccounts=$(cat ./anvilConfig/accounts.json)

keysDir="./internal/testData/web3signer/keys"
keystoresDir="./internal/testData/web3signer/keystores"
passwordsDir="./internal/testData/web3signer/passwords"

for account in $(echo $seedAccounts | jq -rc '.[]'); do
    pk=$(echo $account | jq -r '.private_key')
    address=$(echo $account | jq -r '.address')
    echo "PrivateKey: $pk"
    echo "Address: $address"

    filePath="${keysDir}/${address}.yaml"
    echo "Creating key file at: $filePath"

    password="test"

    cast wallet import --keystore-dir $keystoresDir --unsafe-password $password --private-key $pk "${address}.json"

    echo -n "$password" > "${passwordsDir}/${address}.txt"

    # echo a heredoc to the $filePath
    cat << EOF > "$filePath"
type: "file-raw"
keyType: "SECP256K1"  # For ETH1
privateKey: "${pk}"
EOF
done
