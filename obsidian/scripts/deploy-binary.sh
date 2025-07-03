#!/bin/bash
set -e

# Configuration
VERSION=${VERSION:-$(git describe --tags --always --dirty)}
BUCKET_NAME=${BUCKET_NAME:-"obsidian-artifacts-${AWS_ACCOUNT_ID}-${AWS_REGION}"}
BINARY_NAME="obsidian"

echo "Building Obsidian binary version: ${VERSION}"

# Build for Linux AMD64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
  -ldflags "-X main.Version=${VERSION} -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.GitCommit=$(git rev-parse HEAD)" \
  -o ${BINARY_NAME}-linux-amd64 \
  ./cmd/obsidian

# Create tarball with binary and config
tar -czf obsidian-${VERSION}.tar.gz \
  ${BINARY_NAME}-linux-amd64 \
  config/production.yaml \
  README.md

# Upload to S3
echo "Uploading to S3 bucket: ${BUCKET_NAME}"

# Upload versioned artifact
aws s3 cp obsidian-${VERSION}.tar.gz s3://${BUCKET_NAME}/releases/obsidian-${VERSION}.tar.gz

# Update latest pointer
aws s3 cp obsidian-${VERSION}.tar.gz s3://${BUCKET_NAME}/releases/obsidian-latest.tar.gz

# Upload just the binary for direct download
aws s3 cp ${BINARY_NAME}-linux-amd64 s3://${BUCKET_NAME}/binaries/obsidian-linux-amd64-${VERSION}
aws s3 cp ${BINARY_NAME}-linux-amd64 s3://${BUCKET_NAME}/binaries/obsidian-linux-amd64-latest

echo "Upload complete!"
echo "Latest binary URL: s3://${BUCKET_NAME}/binaries/obsidian-linux-amd64-latest"
echo "Latest release URL: s3://${BUCKET_NAME}/releases/obsidian-latest.tar.gz"

# Cleanup
rm -f ${BINARY_NAME}-linux-amd64 obsidian-${VERSION}.tar.gz