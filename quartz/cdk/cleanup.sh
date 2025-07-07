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

# Get AWS account ID
get_account_id() {
    aws sts get-caller-identity --query Account --output text
}

# Delete Lambda functions created by the factory
delete_lambda_functions() {
    print_status "Looking for Lambda functions created by the factory..."
    
    # List all Lambda functions with the naming pattern
    FUNCTIONS=$(aws lambda list-functions \
        --region $REGION \
        --query 'Functions[?starts_with(FunctionName, `avs-`) && ends_with(FunctionName, `-performer`)].FunctionName' \
        --output text)
    
    if [ -n "$FUNCTIONS" ]; then
        print_warning "Found the following Lambda functions:"
        echo "$FUNCTIONS" | tr '\t' '\n'
        
        read -p "Do you want to delete these Lambda functions? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            for func in $FUNCTIONS; do
                print_status "Deleting Lambda function: $func"
                aws lambda delete-function --function-name $func --region $REGION || true
            done
        fi
    else
        print_status "No Lambda functions found."
    fi
}

# Delete API Gateways created by the factory
delete_api_gateways() {
    print_status "Looking for API Gateways created by the factory..."
    
    # List all REST APIs with the naming pattern
    APIS=$(aws apigateway get-rest-apis \
        --region $REGION \
        --query 'items[?ends_with(name, `-performer-api`)].{id:id,name:name}' \
        --output json)
    
    if [ "$APIS" != "[]" ]; then
        print_warning "Found the following API Gateways:"
        echo "$APIS" | jq -r '.[] | "\(.name) (ID: \(.id))"'
        
        read -p "Do you want to delete these API Gateways? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "$APIS" | jq -r '.[].id' | while read -r api_id; do
                print_status "Deleting API Gateway: $api_id"
                aws apigateway delete-rest-api --rest-api-id $api_id --region $REGION || true
            done
        fi
    else
        print_status "No API Gateways found."
    fi
}

# Delete IAM roles created by the factory
delete_iam_roles() {
    print_status "Looking for IAM roles created by the factory..."
    
    # List all roles with the naming pattern
    ROLES=$(aws iam list-roles \
        --query 'Roles[?starts_with(RoleName, `avs-`) && ends_with(RoleName, `-performer-role`)].RoleName' \
        --output text)
    
    if [ -n "$ROLES" ]; then
        print_warning "Found the following IAM roles:"
        echo "$ROLES" | tr '\t' '\n'
        
        read -p "Do you want to delete these IAM roles? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            for role in $ROLES; do
                print_status "Deleting IAM role: $role"
                # Detach policies first
                aws iam list-attached-role-policies --role-name $role --query 'AttachedPolicies[].PolicyArn' --output text | \
                    xargs -n1 -I{} aws iam detach-role-policy --role-name $role --policy-arn {} || true
                # Delete role
                aws iam delete-role --role-name $role || true
            done
        fi
    else
        print_status "No IAM roles found."
    fi
}

# Empty and delete ECR repository
delete_ecr_repository() {
    print_status "Checking ECR repository..."
    
    REPO_NAME="quartz-performer"
    
    if aws ecr describe-repositories --repository-names $REPO_NAME --region $REGION &> /dev/null; then
        print_warning "ECR repository '$REPO_NAME' exists."
        
        # List images
        IMAGES=$(aws ecr list-images --repository-name $REPO_NAME --region $REGION --query 'imageIds[].imageDigest' --output text)
        
        if [ -n "$IMAGES" ]; then
            print_warning "Repository contains images."
            read -p "Do you want to delete all images? (y/N) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                # Delete all images
                aws ecr batch-delete-image \
                    --repository-name $REPO_NAME \
                    --region $REGION \
                    --image-ids "$(aws ecr list-images --repository-name $REPO_NAME --region $REGION --query 'imageIds')"
            fi
        fi
    fi
}

# Delete CDK stack
delete_stack() {
    print_status "Deleting CDK stack..."
    
    if aws cloudformation describe-stacks --stack-name $STACK_NAME --region $REGION &> /dev/null; then
        cdk destroy $STACK_NAME --force
    else
        print_status "Stack $STACK_NAME not found."
    fi
}

# Main cleanup flow
main() {
    print_warning "This will clean up the Quartz Infrastructure."
    print_warning "This action cannot be undone!"
    
    read -p "Are you sure you want to continue? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Cleanup cancelled."
        exit 0
    fi
    
    print_status "Starting cleanup..."
    
    # Clean up resources created by the factory
    delete_lambda_functions
    delete_api_gateways
    delete_iam_roles
    
    # Delete the main stack
    delete_stack
    
    print_status "Cleanup completed!"
    print_warning "Note: Some resources like CloudWatch logs may still exist."
}

# Run main function
main