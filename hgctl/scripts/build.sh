#!/bin/bash

# Build TypeScript
npx tsc

# Copy proto files
mkdir -p dist/protos
cp src/protos/* dist/protos/

# Copy ABI files
mkdir -p dist/abis
cp src/abis/* dist/abis/

echo "Build complete!"