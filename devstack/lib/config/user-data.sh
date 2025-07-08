#!/bin/bash
set -ex

# Log output to file and console
exec > >(tee -a /var/log/user-data.log)
exec 2>&1

echo "Starting DevStack setup at $(date)"

# Update system
sudo yum update -y

# Install Docker Engine with Compose v2 plugin
echo "Installing Docker Engine with Compose v2..."

# Detect OS version
if grep -q "Amazon Linux 2023" /etc/os-release; then
    echo "Detected Amazon Linux 2023"
    # For AL2023, use dnf and the amazonlinux repo
    sudo dnf install -y docker
    sudo systemctl start docker
    sudo systemctl enable docker
    
    # Install Docker Compose v2 plugin manually for AL2023
    echo "Installing Docker Compose v2 plugin..."
    DOCKER_CONFIG=${DOCKER_CONFIG:-/usr/local/lib/docker}
    sudo mkdir -p $DOCKER_CONFIG/cli-plugins
    sudo curl -SL "https://github.com/docker/compose/releases/latest/download/docker-compose-linux-$(uname -m)" -o $DOCKER_CONFIG/cli-plugins/docker-compose
    sudo chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose
else
    echo "Using standard Docker CE installation"
    # Remove any old Docker versions
    sudo yum remove -y docker docker-client docker-client-latest docker-common docker-latest docker-latest-logrotate docker-logrotate docker-engine
    
    # Install Docker's official repository
    sudo yum install -y yum-utils
    sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
    
    # Install Docker Engine, CLI, and Compose plugin
    sudo yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
fi

# Start and enable Docker
sudo systemctl start docker
sudo systemctl enable docker

# Add ec2-user to docker group and fix permissions
sudo usermod -a -G docker ec2-user
sudo chmod 666 /var/run/docker.sock

# Verify Docker Compose v2 is installed
echo "Verifying Docker Compose v2 installation..."
if docker compose version > /dev/null 2>&1; then
    echo "✅ Docker Compose v2 installed successfully"
    docker compose version
else
    echo "❌ Docker Compose v2 installation failed"
    exit 1
fi

# Install development tools and dependencies
echo "Installing development tools and dependencies..."
sudo yum groupinstall -y "Development Tools"
sudo yum install -y git make iproute coreutils nc

# Install jq (specific version)
echo "Installing jq v1.7.1..."
JQ_VERSION="1.7.1"
for i in {1..3}; do
    if sudo curl -L --retry 3 --retry-delay 5 "https://github.com/jqlang/jq/releases/download/jq-${JQ_VERSION}/jq-linux-amd64" -o /usr/local/bin/jq; then
        sudo chmod +x /usr/local/bin/jq
        break
    else
        echo "Attempt $i failed, retrying..."
        sleep 5
    fi
done

# Install yq (specific version)
echo "Installing yq v4.35.1..."
YQ_VERSION="v4.35.1"
for i in {1..3}; do
    if sudo curl -L --retry 3 --retry-delay 5 "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64" -o /usr/local/bin/yq; then
        sudo chmod +x /usr/local/bin/yq
        break
    else
        echo "Attempt $i failed, retrying..."
        sleep 5
    fi
done

# Install Go
echo "Installing Go..."
GO_VERSION="1.23.6"
curl -L "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
sudo tar -C /usr/local -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz

# Install gomplate
echo "Installing gomplate..."
GOMPLATE_VERSION="v4.1.0"
curl -L "https://github.com/hairyhenderson/gomplate/releases/download/${GOMPLATE_VERSION}/gomplate_linux-amd64" -o /usr/local/bin/gomplate
chmod +x /usr/local/bin/gomplate

# Add Go to PATH for all users
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
export PATH=$PATH:/usr/local/go/bin

# Install Foundry (includes forge) as ec2-user
echo "Installing Foundry..."
sudo -u ec2-user bash -c 'curl -L https://foundry.paradigm.xyz | bash'
# Run foundryup as ec2-user
sudo -u ec2-user bash -c 'source /home/ec2-user/.bashrc && /home/ec2-user/.foundry/bin/foundryup'

# Make foundry available system-wide
sudo ln -sf /home/ec2-user/.foundry/bin/forge /usr/local/bin/forge
sudo ln -sf /home/ec2-user/.foundry/bin/cast /usr/local/bin/cast
sudo ln -sf /home/ec2-user/.foundry/bin/anvil /usr/local/bin/anvil
sudo ln -sf /home/ec2-user/.foundry/bin/chisel /usr/local/bin/chisel

# Install DevKit CLI
echo "Installing DevKit CLI..."
VERSION=v0.0.8
# Fix architecture naming - DevKit uses 'amd64' not 'x86_64'
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
fi
DISTRO=$(uname -s | tr '[:upper:]' '[:lower:]')

# Construct the download URL
DOWNLOAD_URL="https://s3.amazonaws.com/eigenlayer-devkit-releases/${VERSION}/devkit-${DISTRO}-${ARCH}-${VERSION}.tar.gz"
echo "System info: DISTRO=${DISTRO}, ARCH=${ARCH}, VERSION=${VERSION}"
echo "Downloading DevKit from: ${DOWNLOAD_URL}"

# Download to temp file first
TEMP_FILE=$(mktemp)
echo "Downloading to temporary file: ${TEMP_FILE}"

# Download with verbose output and save HTTP response code
HTTP_CODE=$(curl -L -w "%{http_code}" -o "$TEMP_FILE" "${DOWNLOAD_URL}" 2>&1 | tee -a /var/log/user-data.log | tail -n1)
echo "HTTP response code: ${HTTP_CODE}"

# Check if download was successful
if [ "${HTTP_CODE}" != "200" ]; then
    echo "ERROR: Failed to download DevKit. HTTP code: ${HTTP_CODE}"
    echo "Response content (first 500 chars):"
    head -c 500 "$TEMP_FILE"
    rm -f "$TEMP_FILE"
    exit 1
fi

# Check file size and type
FILE_SIZE=$(stat -c%s "$TEMP_FILE")
echo "Downloaded file size: ${FILE_SIZE} bytes"

# Check file type
echo "File type information:"
file "$TEMP_FILE"

# Show first few bytes in hex to debug
echo "First 20 bytes of file (hex):"
xxd -l 20 "$TEMP_FILE"

# First, let's see what's in the tar file
echo "Listing tar contents:"
tar -tvf "$TEMP_FILE" 2>&1 | head -20

# Try to extract with auto-detection of compression
echo "Attempting extraction with auto-detection (tar -xaf)..."
if ! sudo tar -xaf "$TEMP_FILE" -C /usr/local/bin 2>&1; then
    echo "Auto-detection failed, trying explicit methods..."
    
    # Try gzipped tar
    echo "Trying as gzipped tar (tar -xzvf)..."
    if ! sudo tar -xzvf "$TEMP_FILE" -C /usr/local/bin 2>&1; then
        echo "Gzipped tar failed, trying plain tar..."
        
        # Try plain tar
        if ! sudo tar -xvf "$TEMP_FILE" -C /usr/local/bin 2>&1; then
            echo "Plain tar failed. Checking file format..."
            
            # Get more info about the file
            file "$TEMP_FILE"
            echo "Tar version info:"
            tar --version
            
            # Try extracting to temp directory first
            echo "Trying extraction to temp directory..."
            TEMP_DIR=$(mktemp -d)
            if tar -xaf "$TEMP_FILE" -C "$TEMP_DIR" 2>&1; then
                echo "Extraction to temp successful, contents:"
                ls -la "$TEMP_DIR"
                # Move devkit binary if found
                if [ -f "$TEMP_DIR/devkit" ]; then
                    sudo mv "$TEMP_DIR/devkit" /usr/local/bin/
                    sudo chmod +x /usr/local/bin/devkit
                    echo "Moved devkit to /usr/local/bin"
                fi
            else
                echo "ERROR: All extraction attempts failed"
                echo "File header (first 512 bytes in hex):"
                xxd -l 512 "$TEMP_FILE"
            fi
            rm -rf "$TEMP_DIR"
        fi
    fi
fi

# Check if devkit was extracted successfully
if [ -f /usr/local/bin/devkit ]; then
    echo "DevKit extracted successfully"
    sudo chmod +x /usr/local/bin/devkit
else
    echo "WARNING: devkit binary not found in /usr/local/bin after extraction"
fi

# Clean up temp file
rm -f "$TEMP_FILE"

# If extraction failed, try alternative download method
if [ ! -f /usr/local/bin/devkit ]; then
    echo "ERROR: devkit binary not found after extraction. Trying alternative download..."
    
    # Alternative: Try downloading the binary directly (not tar)
    BINARY_URL="https://s3.amazonaws.com/eigenlayer-devkit-releases/${VERSION}/devkit-${DISTRO}-${ARCH}"
    echo "Attempting direct binary download from: ${BINARY_URL}"
    
    if curl -L -o /usr/local/bin/devkit "${BINARY_URL}"; then
        chmod +x /usr/local/bin/devkit
        echo "Direct binary download successful"
    else
        echo "ERROR: Direct binary download also failed"
        
        # Last resort: Try GitHub releases
        GITHUB_URL="https://github.com/Layr-Labs/devkit/releases/download/${VERSION}/devkit-${DISTRO}-${ARCH}-${VERSION}.tar.gz"
        echo "Trying GitHub releases as fallback: ${GITHUB_URL}"
        
        TEMP_FILE=$(mktemp)
        if curl -L -o "$TEMP_FILE" "${GITHUB_URL}"; then
            echo "GitHub download successful, attempting extraction..."
            tar -xzvf "$TEMP_FILE" -C /usr/local/bin || tar -xvf "$TEMP_FILE" -C /usr/local/bin
        fi
        rm -f "$TEMP_FILE"
    fi
fi

# Verify DevKit installation
echo "Checking if devkit binary exists..."
if [ -f /usr/local/bin/devkit ]; then
    echo "DevKit binary found at /usr/local/bin/devkit"
    chmod +x /usr/local/bin/devkit
    echo "DevKit version:"
    /usr/local/bin/devkit version
else
    echo "ERROR: DevKit installation failed - binary not found"
    exit 1
fi

# Make sure devkit and other tools are in ec2-user's PATH
echo "export PATH=$PATH:/usr/local/bin:/usr/local/go/bin" >> /home/ec2-user/.bashrc
echo "export PATH=$PATH:/home/ec2-user/.foundry/bin" >> /home/ec2-user/.bashrc

# Wait for Docker to be fully ready
echo "Waiting for Docker to be ready..."
while ! sudo docker info > /dev/null 2>&1; do
    echo "Waiting for Docker daemon..."
    sleep 2
done

# Verify all dependencies are installed before proceeding
echo "Verifying dependencies..."
echo "Go version: $(go version)"
echo "Forge version: $(forge --version)"
echo "jq version: $(jq --version)"
echo "yq version: $(yq --version)"
echo "Docker version: $(docker --version)"

# Double-check critical tools exist
for tool in go forge jq yq docker devkit gomplate; do
    if ! command -v $tool &> /dev/null; then
        echo "ERROR: $tool is not installed or not in PATH"
        exit 1
    fi
done
echo "All required tools are installed"

# Create AVS project as ec2-user with full PATH set
echo "Creating AVS project..."
if ! sudo -u ec2-user bash -c 'export PATH=$PATH:/usr/local/bin:/usr/local/go/bin:/home/ec2-user/.foundry/bin && cd /home/ec2-user && devkit avs create devstack'; then
    echo "ERROR: Failed to create AVS project"
    exit 1
fi

echo "Project created successfully"

# Build the AVS
echo "Building AVS..."
if ! sudo -u ec2-user bash -c 'export PATH=$PATH:/usr/local/bin:/usr/local/go/bin:/home/ec2-user/.foundry/bin && cd /home/ec2-user/devstack && devkit avs build'; then
    echo "ERROR: Failed to build AVS"
    exit 1
fi

echo "AVS built successfully"

# Update the devnet context with fork URL
echo "Updating devnet context with fork URL..."
FORK_URL="FORK_URL_PLACEHOLDER"
echo "Fork URL to be used: ${FORK_URL}"

CONTEXT_FILE="/home/ec2-user/devstack/config/contexts/devnet.yaml"
if [ -f "$CONTEXT_FILE" ]; then
    # Use yq to update the fork URLs in the context file
    sudo -u ec2-user yq eval '.context.chains.l1.fork.url = "'"${FORK_URL}"'"' -i "$CONTEXT_FILE"
    sudo -u ec2-user yq eval '.context.chains.l2.fork.url = "'"${FORK_URL}"'"' -i "$CONTEXT_FILE"
    echo "Updated fork URLs in devnet context"

    # Verify the update - simplified version
    echo "L1 fork URL:"
    sudo -u ec2-user yq eval '.context.chains.l1.fork.url' "$CONTEXT_FILE"
    echo "L2 fork URL:"
    sudo -u ec2-user yq eval '.context.chains.l2.fork.url' "$CONTEXT_FILE"
else
    echo "WARNING: Context file not found at $CONTEXT_FILE"
fi

echo "Updating Docker Compose to allow external access..."
DOCKER_COMPOSE_FILE="/home/ec2-user/devstack/.hourglass/docker-compose.yml"

if [ -f "$DOCKER_COMPOSE_FILE" ]; then
    echo "Found docker-compose.yml, updating port bindings..."

    # Backup the original file
    sudo -u ec2-user cp "$DOCKER_COMPOSE_FILE" "${DOCKER_COMPOSE_FILE}.backup"

    # Update port bindings from 127.0.0.1 to 0.0.0.0
    sudo -u ec2-user sed -i 's/"127\.0\.0\.1:/"0.0.0.0:/g' "$DOCKER_COMPOSE_FILE"

    echo "Updated port bindings in docker-compose.yml"

    # Verify the changes
    echo "New port configuration:"
    grep -E "ports:|-.*(9090|8081|9000)" "$DOCKER_COMPOSE_FILE"
else
    echo "WARNING: docker-compose.yml not found at $DOCKER_COMPOSE_FILE"
fi

# Start devnet directly instead of using systemd for simplicity
echo "Starting devnet in background..."
cd /home/ec2-user/devstack

# Create log file with proper permissions
sudo touch /var/log/devnet.log
sudo chown ec2-user:ec2-user /var/log/devnet.log

# Start devnet as ec2-user in background with full PATH
# No need to pass fork URL as env vars since it's now in the context file
sudo -u ec2-user bash -c 'export PATH=$PATH:/usr/local/bin:/usr/local/go/bin:/home/ec2-user/.foundry/bin && cd /home/ec2-user/devstack && nohup devkit avs devnet start > /var/log/devnet.log 2>&1 &'

# Wait for services to be ready
echo "Waiting for executor on port 9090 and aggregator on port 8081 to be available... This can take time to bootstrap anvil"
TIMEOUT=120
COUNTER=0
until nc -z localhost 9090 && nc -z localhost 8081; do
    if [ $COUNTER -ge $TIMEOUT ]; then
        echo "ERROR: Services did not start within $TIMEOUT seconds"
        cat /var/log/devnet.log
        exit 1
    fi
    echo "Waiting for services... $COUNTER/$TIMEOUT"
    sleep 2
    COUNTER=$((COUNTER + 2))
done

echo "DevStack setup completed at $(date)"

