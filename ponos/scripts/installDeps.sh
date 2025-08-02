#!/usr/bin/env bash

export BIN="/usr/local/bin/"
export VERSION="1.42.0"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
BINARY_NAME="buf-$(uname -s)-$(uname -m)"

url="https://github.com/bufbuild/buf/releases/download/v${VERSION}/${BINARY_NAME}.tar.gz"
echo $url

curl -L $url | tar xvz -C /tmp
sudo mv /tmp/buf/bin/* $BIN

rm -rf /tmp/buf

# install os-specific dependencies
arch=$(uname -m)
os=$(uname -s | tr '[:upper:]' '[:lower:]')

if [[ $os == "linux" ]]; then
    if [[ $arch == "x86_64" ]]; then
        sudo curl -Lo /usr/local/bin/kind https://kind.sigs.k8s.io/dl/v0.29.0/kind-linux-amd64
        sudo chmod +x /usr/local/bin/kind
    elif [[ $arch == "aarch64" ]] || [[ $arch == "arm64" ]]; then
        sudo curl -Lo /usr/local/bin/kind https://kind.sigs.k8s.io/dl/v0.29.0/kind-linux-arm64
        sudo chmod +x /usr/local/bin/kind
    else
        echo "Unsupported architecture: $arch"
        exit 1
    fi
elif [[ $os == "darwin" ]]; then
    brew install kind
else
    echo "Unsupported OS: $os"
    exit 1
fi
