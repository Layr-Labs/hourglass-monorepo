#!/bin/bash

set -e

# hgctl version
HGCTL_VERSION="v0.1.0.testnet-rc.2"

# Ensure version has 'v' prefix for consistency
if [[ ! "$HGCTL_VERSION" =~ ^v ]]; then
    HGCTL_VERSION="v${HGCTL_VERSION}"
fi

HGCTL_BASE_URL="https://github.com/Layr-Labs/hourglass-monorepo/releases/download"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $OS in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *) echo "Error: Unsupported OS: $OS"; exit 1 ;;
esac

case $ARCH in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Error: Unsupported architecture: $ARCH"; exit 1 ;;
esac

PLATFORM="${OS}-${ARCH}"

# Prompt for installation directory
if [[ -t 0 ]]; then
    # Interactive terminal available
    echo "Where would you like to install hgctl?"
    echo "1) $HOME/bin (recommended)"
    echo "2) /usr/local/bin (system-wide, requires sudo)"
    echo "3) Custom path"
    read -p "Enter choice (1-3) [1]: " choice
else
    # Non-interactive (piped), use default
    echo "Installing to $HOME/bin (default for non-interactive install)"
    choice=1
fi

case ${choice:-1} in
    1) INSTALL_DIR="$HOME/bin" ;;
    2) INSTALL_DIR="/usr/local/bin" ;;
    3) 
        read -p "Enter custom path: " INSTALL_DIR
        if [[ -z "$INSTALL_DIR" ]]; then
            echo "Error: No path provided"
            exit 1
        fi
        ;;
    *) echo "Invalid choice"; exit 1 ;;
esac

# Create directory if it doesn't exist
if [[ "$INSTALL_DIR" == "/usr/local/bin" ]]; then
    sudo mkdir -p "$INSTALL_DIR"
else
    mkdir -p "$INSTALL_DIR"
fi

# Download and install
# URL format: https://github.com/Layr-Labs/hourglass-monorepo/releases/download/hgctl-v0.1.0.preview-rc.1/hgctl-darwin-arm64-v0.1.0.preview-rc.1.tar.gz
HGCTL_URL="${HGCTL_BASE_URL}/hgctl-${HGCTL_VERSION}/hgctl-${PLATFORM}-${HGCTL_VERSION}.tar.gz"
echo "Downloading hgctl ${HGCTL_VERSION} for ${PLATFORM}..."

# Create temp directory for download
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# Download and extract
if ! curl -fsSL "$HGCTL_URL" -o "$TEMP_DIR/hgctl.tar.gz"; then
    echo "Error: Failed to download hgctl from $HGCTL_URL"
    echo "Please check if the release exists or try again later."
    exit 1
fi

# Extract the archive
tar xz -C "$TEMP_DIR" -f "$TEMP_DIR/hgctl.tar.gz"

# Find the hgctl binary (it might be in a subdirectory or have a platform-specific name)
HGCTL_BIN=""
if [[ -f "$TEMP_DIR/hgctl" ]]; then
    HGCTL_BIN="$TEMP_DIR/hgctl"
elif [[ -f "$TEMP_DIR/hgctl-${PLATFORM}" ]]; then
    HGCTL_BIN="$TEMP_DIR/hgctl-${PLATFORM}"
elif [[ -f "$TEMP_DIR/bin/hgctl" ]]; then
    HGCTL_BIN="$TEMP_DIR/bin/hgctl"
else
    # Try to find any hgctl* executable in the temp dir
    HGCTL_BIN=$(find "$TEMP_DIR" -name "hgctl*" -type f -perm +111 | head -1)
fi

if [[ -z "$HGCTL_BIN" ]] || [[ ! -f "$HGCTL_BIN" ]]; then
    echo "Error: Could not find hgctl binary in the extracted archive"
    echo "Archive contents:"
    ls -la "$TEMP_DIR"
    exit 1
fi

# Install the binary
if [[ "$INSTALL_DIR" == "/usr/local/bin" ]]; then
    sudo mv "$HGCTL_BIN" "$INSTALL_DIR/hgctl"
    sudo chmod +x "${INSTALL_DIR}/hgctl"
else
    mv "$HGCTL_BIN" "$INSTALL_DIR/hgctl"
    chmod +x "$INSTALL_DIR/hgctl"
fi

echo "✅ hgctl installed to $INSTALL_DIR/hgctl"

# Add to PATH if needed
if [[ "$INSTALL_DIR" == "$HOME/bin" ]] && [[ ":$PATH:" != *":$HOME/bin:"* ]]; then
    echo ""
    echo "💡 Add $HOME/bin to your PATH:"
    
    # Detect shell
    if [[ -n "$BASH_VERSION" ]]; then
        SHELL_RC="$HOME/.bashrc"
    elif [[ -n "$ZSH_VERSION" ]]; then
        SHELL_RC="$HOME/.zshrc"
    else
        SHELL_RC="$HOME/.$(basename "$SHELL")rc"
    fi
    
    echo "   echo 'export PATH=\"\$HOME/bin:\$PATH\"' >> $SHELL_RC"
    echo "   source $SHELL_RC"
fi

echo ""
echo "🚀 Verify installation: $INSTALL_DIR/hgctl --version"
echo "📚 Get started: $INSTALL_DIR/hgctl --help"