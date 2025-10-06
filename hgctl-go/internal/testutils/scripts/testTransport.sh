#!/usr/bin/env bash

set -ex

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
HGCTL_ROOT="$SCRIPT_DIR/../../.."

TRANSPORT_CONFIG="/tmp/transport-config.json"

# Check if transport config exists
if [ ! -f "$TRANSPORT_CONFIG" ]; then
    echo "ERROR: Transport config not found at $TRANSPORT_CONFIG"
    echo "Please run 'make generate-test-state' first to generate the config."
    echo "The main script will pause before transport - you can then run this script."
    exit 1
fi

# Build the transport binary if not already built
if [ ! -f "$HGCTL_ROOT/bin/transport" ]; then
    echo "Building transport binary..."
    cd "$HGCTL_ROOT"
    make build-transport
fi

echo "Transport config:"
jq . "$TRANSPORT_CONFIG"

echo ""
echo "Running operator table transport..."
if "$HGCTL_ROOT/bin/transport" -config "$TRANSPORT_CONFIG" -v; then
    echo ""
    echo "✓ Operator table transport completed successfully"
else
    echo ""
    echo "✗ ERROR: Operator table transport failed"
    exit 1
fi