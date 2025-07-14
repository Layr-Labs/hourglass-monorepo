#!/bin/bash
set -ex

# Log output to file and console
exec > >(tee -a /var/log/user-data.log)
exec 2>&1

# Grow the root partition to use all available space
# This is important on Ubuntu as it might not automatically use the full EBS volume
echo "Expanding root filesystem to use full disk..."

# Detect the root device
ROOT_DEV=$(df / | tail -1 | awk '{print $1}' | sed 's/[0-9]*$//' | sed 's/p$//')
ROOT_PART=$(df / | tail -1 | awk '{print $1}')
PART_NUM=$(echo $ROOT_PART | grep -o '[0-9]*$')

echo "Root device: $ROOT_DEV"
echo "Root partition: $ROOT_PART"
echo "Partition number: $PART_NUM"

# First, grow the partition if needed
if [ -n "$PART_NUM" ]; then
    sudo growpart "$ROOT_DEV" "$PART_NUM" || true
fi

# Then resize the filesystem
sudo resize2fs "$ROOT_PART" || sudo xfs_growfs / || true

# Show disk usage after expansion
echo "Disk space after expansion:"
df -h /
lsblk

# Update system
sudo apt-get update -y
sudo apt-get upgrade -y

# Install cloud-guest-utils for growpart command
sudo apt-get install -y cloud-guest-utils

# Install Docker prerequisites
sudo apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

# Add Docker's official GPG key
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Set up the repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine and Docker Compose plugin
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Configure Docker daemon to use more space and proper storage driver
sudo mkdir -p /etc/docker
cat <<'EOL' | sudo tee /etc/docker/daemon.json
{
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.override_kernel_check=true"
  ],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOL

# Start and enable Docker
sudo systemctl daemon-reload
sudo systemctl start docker
sudo systemctl enable docker

# Add ubuntu user to docker group and fix permissions
sudo usermod -a -G docker ubuntu
sudo chmod 666 /var/run/docker.sock

# Verify Docker Compose v2 is installed
if ! docker compose version > /dev/null 2>&1; then
    echo "ERROR: Docker Compose v2 installation failed"
    exit 1
fi

# Install development tools and dependencies
echo "Installing development tools and dependencies..."
sudo apt-get install -y build-essential git make iproute2 coreutils netcat-openbsd

# Install AWS CLI (needed for SSM parameter retrieval)
echo "Installing AWS CLI..."
sudo apt-get install -y awscli

# Install CloudWatch Agent
wget -q https://amazoncloudwatch-agent.s3.amazonaws.com/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb
sudo dpkg -i -E ./amazon-cloudwatch-agent.deb > /dev/null 2>&1
rm ./amazon-cloudwatch-agent.deb

# Configure CloudWatch Agent
# Get region using IMDSv2
TOKEN=$(curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600" -s)
REGION=$(curl -H "X-aws-ec2-metadata-token: $TOKEN" -s http://169.254.169.254/latest/meta-data/placement/region)

# Download CloudWatch Agent config from Parameter Store
if aws ssm get-parameter --name "/devstack/cloudwatch/agent-config.json" --region "$REGION" --query 'Parameter.Value' --output text > /tmp/cw-config.json 2>/dev/null; then
    sudo mv /tmp/cw-config.json /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json
    sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -a fetch-config -m ec2 -s -c file:/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json
else
    echo "WARNING: Could not download CloudWatch config from Parameter Store"
fi

# Install jq
JQ_VERSION="1.7.1"
sudo curl -sL "https://github.com/jqlang/jq/releases/download/jq-${JQ_VERSION}/jq-linux-amd64" -o /usr/local/bin/jq
sudo chmod +x /usr/local/bin/jq

# Install yq
YQ_VERSION="v4.35.1"
sudo curl -sL "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64" -o /usr/local/bin/yq
sudo chmod +x /usr/local/bin/yq

# Install Go
GO_VERSION="1.23.6"
curl -sL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" | sudo tar -C /usr/local -xz

GOMPLATE_VERSION="v4.1.0"
sudo curl -sL "https://github.com/hairyhenderson/gomplate/releases/download/${GOMPLATE_VERSION}/gomplate_linux-amd64" -o /usr/local/bin/gomplate
sudo chmod +x /usr/local/bin/gomplate

# Install Node.js and npm (required for DevKit)
echo "Installing Node.js..."
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs
# Verify npm is installed
npm --version || exit 1

# Add Go to PATH for all users
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
export PATH=$PATH:/usr/local/go/bin

# Install Foundry
sudo -u ubuntu bash -c 'curl -sL https://foundry.paradigm.xyz | bash'
sudo -u ubuntu bash -c 'source /home/ubuntu/.bashrc && /home/ubuntu/.foundry/bin/foundryup > /dev/null 2>&1'

# Make foundry available system-wide
sudo ln -sf /home/ubuntu/.foundry/bin/forge /usr/local/bin/forge
sudo ln -sf /home/ubuntu/.foundry/bin/cast /usr/local/bin/cast
sudo ln -sf /home/ubuntu/.foundry/bin/anvil /usr/local/bin/anvil
sudo ln -sf /home/ubuntu/.foundry/bin/chisel /usr/local/bin/chisel

# Install DevKit CLI from source
echo "Installing DevKit CLI from source..."
# Set up Go environment for root's build
export HOME=/root
export GOPATH=/root/go
export GOCACHE=/root/.cache/go-build
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

cd /tmp
git clone https://github.com/Layr-Labs/devkit-cli.git
cd devkit-cli
# Build and install DevKit
if make install; then
    # Move from ~/bin to system-wide location
    sudo mv ~/bin/devkit /usr/local/bin/
    sudo chmod +x /usr/local/bin/devkit
    echo "DevKit installed successfully"
    /usr/local/bin/devkit version
else
    echo "ERROR: Failed to build DevKit from source"
    exit 1
fi
cd /
rm -rf /tmp/devkit-cli

# Make sure devkit and other tools are in ubuntu user's PATH
echo "export PATH=$PATH:/usr/local/bin:/usr/local/go/bin" >> /home/ubuntu/.bashrc
echo "export PATH=$PATH:/home/ubuntu/.foundry/bin" >> /home/ubuntu/.bashrc

# Check disk space before proceeding
echo "Checking disk space..."
df -h /
echo "Docker root directory will be: /var/lib/docker"
df -h /var/lib/docker || df -h /var/lib || df -h /

# Ensure Docker has enough space to work with
echo "Docker disk usage before operations:"
sudo du -sh /var/lib/docker 2>/dev/null || echo "Docker directory not yet created"

# Wait for Docker to be fully ready
echo "Waiting for Docker to be ready..."
while ! sudo docker info > /dev/null 2>&1; do
    echo "Waiting for Docker daemon..."
    sleep 2
done

# Verify critical tools (removed gomplate - will be installed by DevKit init)
for tool in go forge jq yq docker devkit; do
    if ! command -v $tool &> /dev/null; then
        echo "ERROR: $tool is not installed"
        exit 1
    fi
done

# Create and build AVS project with proper permissions
sudo -u ubuntu bash << 'EOF'
set -ex
# Set up environment
export PATH=$PATH:/usr/local/bin:/usr/local/go/bin:/home/ubuntu/.foundry/bin
export HOME=/home/ubuntu
export GOPATH=/home/ubuntu/go
export GOCACHE=/home/ubuntu/.cache/go-build
export GOMODCACHE=/home/ubuntu/go/pkg/mod

# Create all directories as ubuntu user
mkdir -p ~/.cache/go-build
mkdir -p ~/go/pkg/mod
mkdir -p ~/go/bin

# Create project
cd ~
devkit avs create devstack
cd devstack

# Update devnet context with fork URLs and RPC URLs BEFORE building
L1_FORK_URL="https://practical-serene-mound.ethereum-sepolia.quiknode.pro/3aaa48bd95f3d6aed60e89a1a466ed1e2a440b61/"
L2_FORK_URL="https://tiniest-proud-surf.base-sepolia.quiknode.pro/7a175be1bac281923e1d87fb725ad4c965db10bd/"
CONTEXT_FILE="config/contexts/devnet.yaml"
if [ -f "$CONTEXT_FILE" ]; then
    yq eval '.context.chains.l1.fork.url = "'"${L1_FORK_URL}"'"' -i "$CONTEXT_FILE"
    yq eval '.context.chains.l2.fork.url = "'"${L2_FORK_URL}"'"' -i "$CONTEXT_FILE"
fi

# Check disk space before build
echo "Disk space before build:"
df -h /
echo "Docker disk usage:"
sudo docker system df || true

# Now build with the updated URLs
devkit avs build
EOF

if [ $? -ne 0 ]; then
    echo "ERROR: Failed to create/build AVS project"
    exit 1
fi

# Update Docker Compose port bindings
DOCKER_COMPOSE_FILE="/home/ubuntu/devstack/.hourglass/docker-compose.yml"
if [ -f "$DOCKER_COMPOSE_FILE" ]; then
    sudo -u ubuntu sed -i 's/"127\.0\.0\.1:/"0.0.0.0:/g' "$DOCKER_COMPOSE_FILE"
fi

# Download systemd service from Parameter Store
# Get region using IMDSv2 if not already set
if [ -z "$REGION" ]; then
    TOKEN=$(curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600" -s)
    REGION=$(curl -H "X-aws-ec2-metadata-token: $TOKEN" -s http://169.254.169.254/latest/meta-data/placement/region)
fi
PARAMETER_NAME="/devstack/systemd/devnet.service"

# Download the systemd service from Parameter Store
# Wait for SSM agent
# Get instance ID using IMDSv2
INSTANCE_ID=$(curl -H "X-aws-ec2-metadata-token: $TOKEN" -s http://169.254.169.254/latest/meta-data/instance-id)
for i in {1..30}; do
    if aws ssm describe-instance-information --region "$REGION" --filters "Key=InstanceIds,Values=$INSTANCE_ID" --query 'InstanceInformationList[0].PingStatus' --output text 2>/dev/null | grep -q "Online"; then
        break
    fi
    sleep 5
done

# Try to get the parameter
MAX_RETRIES=5
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if aws ssm get-parameter --name "$PARAMETER_NAME" --region "$REGION" --query 'Parameter.Value' --output text > /tmp/devnet.service 2>/dev/null; then
        echo "Successfully downloaded systemd service from Parameter Store"
        sudo mv /tmp/devnet.service /etc/systemd/system/devnet.service
        break
    else
        RETRY_COUNT=$((RETRY_COUNT + 1))
        if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
            echo "ERROR: Failed to download systemd service from Parameter Store after $MAX_RETRIES attempts"
            echo "Please ensure the SSM parameter $PARAMETER_NAME exists in region $REGION"
            exit 1
        fi
        echo "Failed to download from Parameter Store, retrying... ($RETRY_COUNT/$MAX_RETRIES)"
        sleep 10
    fi
done

# Set proper permissions
sudo chmod 644 /etc/systemd/system/devnet.service

# Reload systemd and enable the service
echo "Enabling devnet service..."
sudo systemctl daemon-reload
sudo systemctl enable devnet.service

# Create log file with proper permissions
sudo touch /var/log/devnet.log
sudo chown ubuntu:ubuntu /var/log/devnet.log

# Start the service
echo "Starting devnet service..."
sudo systemctl start devnet.service

# Wait for services to be ready
echo "Waiting for executor on port 9090 and aggregator on port 8081 to be available... This can take time to bootstrap anvil"
TIMEOUT=360
COUNTER=0
until nc -z localhost 9090 && nc -z localhost 8081; do
    if [ $COUNTER -ge $TIMEOUT ]; then
        echo "ERROR: Services did not start within $TIMEOUT seconds"
        echo "Checking service status..."
        sudo systemctl status devnet.service
        echo "Recent logs:"
        sudo journalctl -u devnet.service -n 50
        exit 1
    fi
    echo "Waiting for services... $COUNTER/$TIMEOUT"
    sleep 2
    COUNTER=$((COUNTER + 2))
done

# Final status check
sudo systemctl is-active devnet.service > /dev/null || exit 1

