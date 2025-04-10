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

