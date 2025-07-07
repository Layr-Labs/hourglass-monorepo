#!/bin/bash
set -e

echo "🚀 Deploying Hourglass Kubernetes Infrastructure"
echo "=============================================="

# Check prerequisites
echo "Checking prerequisites..."

if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI is not installed. Please install it first."
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed. Please install Node.js first."
    exit 1
fi

if ! command -v cdk &> /dev/null; then
    echo "⚠️  CDK CLI is not installed. Installing..."
    npm install -g aws-cdk
fi

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    echo "❌ AWS credentials not configured. Please run 'aws configure'."
    exit 1
fi

# Install dependencies
echo "Installing dependencies..."
npm install

# Bootstrap CDK if needed
ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
REGION=${AWS_DEFAULT_REGION:-us-east-1}

echo "Bootstrapping CDK for account $ACCOUNT in region $REGION..."
cdk bootstrap aws://$ACCOUNT/$REGION

# Build TypeScript
echo "Building TypeScript..."
npm run build

# Deploy stacks
echo "Deploying stacks..."
echo ""
echo "This will deploy:"
echo "  1. EKS Cluster with managed node groups"
echo "  2. Anvil L1/L2 devnet on ECS Fargate"
echo "  3. Hourglass Ponos services (aggregator/executor)"
echo "  4. Obsidian AVS operator (without TEE)"
echo ""
echo "⏰ This process will take approximately 20-30 minutes"
echo ""

# Deploy all stacks
cdk deploy --all --require-approval never

# Get cluster credentials
CLUSTER_NAME=$(aws eks list-clusters --query 'clusters[?contains(@, `hourglass-eks`)]' --output text)
if [ -n "$CLUSTER_NAME" ]; then
    echo ""
    echo "✅ Deployment complete!"
    echo ""
    echo "Configure kubectl:"
    echo "  aws eks update-kubeconfig --name $CLUSTER_NAME --region $REGION"
    echo ""
    echo "View AVS resources:"
    echo "  kubectl get avs -A"
    echo ""
    echo "Check Ponos services:"
    echo "  kubectl get pods -n hourglass"
    echo ""
    echo "Get aggregator endpoint:"
    echo "  kubectl get svc -n hourglass aggregator-lb"
fi