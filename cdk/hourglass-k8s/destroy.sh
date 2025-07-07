#!/bin/bash
set -e

echo "🗑️  Destroying Hourglass Kubernetes Infrastructure"
echo "================================================"
echo ""
echo "⚠️  WARNING: This will delete all resources including:"
echo "  - EKS cluster and all workloads"
echo "  - Anvil devnet nodes"
echo "  - All data in EFS volumes"
echo "  - Load balancers and other AWS resources"
echo ""
read -p "Are you sure you want to continue? (yes/no): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Destroy all stacks
echo "Destroying CDK stacks..."
cdk destroy --all --force

echo ""
echo "✅ Infrastructure destroyed successfully!"