#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
STACK_NAME="QuartzStack"
REGION="${AWS_REGION:-us-east-1}"
RELEASE_MANAGER_ADDRESS="${RELEASE_MANAGER_ADDRESS:-}"
RPC_ENDPOINT="${RPC_ENDPOINT:-}"
EKS_NODE_ROLE_ARN="${EKS_NODE_ROLE_ARN:-}"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check AWS CLI
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI not found. Please install it first."
        exit 1
    fi
    
    # Check CDK CLI
    if ! command -v cdk &> /dev/null; then
        print_error "AWS CDK not found. Please install it: npm install -g aws-cdk"
        exit 1
    fi
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker not found. Please install Docker first."
        exit 1
    fi
    
    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        print_error "AWS credentials not configured. Please run 'aws configure'."
        exit 1
    fi
    
    print_status "Prerequisites check passed."
}

# Get AWS account ID
get_account_id() {
    aws sts get-caller-identity --query Account --output text
}

# Bootstrap CDK
bootstrap_cdk() {
    print_status "Checking CDK bootstrap..."
    
    ACCOUNT_ID=$(get_account_id)
    BOOTSTRAP_STACK="CDKToolkit"
    
    if ! aws cloudformation describe-stacks --stack-name $BOOTSTRAP_STACK --region $REGION &> /dev/null; then
        print_status "Bootstrapping CDK for account $ACCOUNT_ID in region $REGION..."
        cdk bootstrap aws://$ACCOUNT_ID/$REGION
    else
        print_status "CDK already bootstrapped."
    fi
}

# Install dependencies
install_dependencies() {
    print_status "Installing dependencies..."
    npm install
}

# Deploy CDK stack
deploy_stack() {
    print_status "Deploying Quartz Infrastructure..."
    
    CDK_ARGS=""
    
    if [ -n "$RELEASE_MANAGER_ADDRESS" ]; then
        CDK_ARGS="$CDK_ARGS -c releaseManagerAddress=$RELEASE_MANAGER_ADDRESS"
    else
        print_error "RELEASE_MANAGER_ADDRESS is required. Set it as an environment variable."
        exit 1
    fi
    
    if [ -n "$RPC_ENDPOINT" ]; then
        CDK_ARGS="$CDK_ARGS -c rpcEndpoint=$RPC_ENDPOINT"
    else
        print_error "RPC_ENDPOINT is required. Set it as an environment variable."
        exit 1
    fi
    
    if [ -n "$EKS_NODE_ROLE_ARN" ]; then
        CDK_ARGS="$CDK_ARGS -c eksNodeRoleArn=$EKS_NODE_ROLE_ARN"
    fi
    
    cdk deploy $STACK_NAME $CDK_ARGS --require-approval never
}

# Get stack outputs
get_outputs() {
    print_status "Getting stack outputs..."
    
    aws cloudformation describe-stacks \
        --stack-name $STACK_NAME \
        --region $REGION \
        --query 'Stacks[0].Outputs' \
        --output table
}

# Build and push Quartz adapter
deploy_quartz_adapter() {
    print_status "Building and pushing Quartz adapter..."
    
    ACCOUNT_ID=$(get_account_id)
    ECR_REGISTRY="$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com"
    
    # Get ECR repository URI from stack outputs
    ECR_URI=$(aws cloudformation describe-stacks \
        --stack-name $STACK_NAME \
        --region $REGION \
        --query 'Stacks[0].Outputs[?OutputKey==`ECRRepositoryURI`].OutputValue' \
        --output text)
    
    if [ -z "$ECR_URI" ]; then
        print_error "Could not find ECR repository URI. Make sure the stack is deployed."
        exit 1
    fi
    
    print_status "ECR Repository: $ECR_URI"
    
    # Change to lambda-adapter directory
    cd ../adapter
    
    # Login to ECR
    print_status "Logging in to ECR..."
    aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin $ECR_REGISTRY
    
    # Build and push
    print_status "Building Docker image..."
    docker build -t quartz-adapter:latest .
    
    print_status "Tagging image..."
    docker tag quartz-adapter:latest $ECR_URI:latest
    
    print_status "Pushing image to ECR..."
    docker push $ECR_URI:latest
    
    cd -
}

# Deploy example Lambda (optional)
deploy_example_lambda() {
    read -p "Do you want to deploy the example Lambda function? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_status "Deploying example Lambda function..."
        
        ECR_URI=$(aws cloudformation describe-stacks \
            --stack-name $STACK_NAME \
            --region $REGION \
            --query 'Stacks[0].Outputs[?OutputKey==`ECRRepositoryURI`].OutputValue' \
            --output text)
        
        cd ../example
        
        print_status "Building example Lambda..."
        docker build -t example-lambda:latest .
        
        print_status "Tagging example Lambda..."
        docker tag example-lambda:latest $ECR_URI:example
        
        print_status "Pushing example Lambda..."
        docker push $ECR_URI:example
        
        cd -
    fi
}

# Main deployment flow
main() {
    print_status "Starting Quartz Infrastructure deployment..."
    
    check_prerequisites
    bootstrap_cdk
    install_dependencies
    deploy_stack
    deploy_quartz_adapter
    deploy_example_lambda
    
    print_status "Deployment completed successfully!"
    echo
    get_outputs
    
    echo
    print_status "Next steps:"
    echo "1. Note the Registry API endpoint from the outputs above"
    echo "2. Configure your Executor to use the Quartz adapter container"
    echo "3. Monitor the Release Monitor function for new releases"
    echo "4. Check CloudWatch logs for any issues"
}

# Run main function
main