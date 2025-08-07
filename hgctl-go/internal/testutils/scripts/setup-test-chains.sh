#!/usr/bin/env bash
# Wrapper script to set up test chains with proper cleanup

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
HGCTL_ROOT="$SCRIPT_DIR/../../.."
PROJECT_ROOT="$HGCTL_ROOT/.."

echo "Setting up test chains with cleanup..."

# 1. Clean up any existing anvil processes
echo "Stopping any existing anvil processes..."
pkill -f "anvil.*8545" || true
pkill -f "anvil.*9545" || true
sleep 2

# 2. Clean up existing test data files
echo "Cleaning up existing test data..."
chmod 644 "$HGCTL_ROOT/internal/testdata/anvil"* 2>/dev/null || true
rm -f "$HGCTL_ROOT/internal/testdata/anvil"* 2>/dev/null || true

# 3. Create necessary directories
echo "Creating directories..."
mkdir -p "$HGCTL_ROOT/internal/testdata"
mkdir -p "$HGCTL_ROOT/internal/testutils/chainData"
mkdir -p "$HGCTL_ROOT/internal/testutils/keys"

# 4. Run the main setup script
echo "Running chain setup..."
exec "$SCRIPT_DIR/generateTestChainState.sh" "$@"