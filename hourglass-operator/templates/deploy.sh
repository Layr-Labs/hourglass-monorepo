#!/bin/bash

# Hourglass Executor Deployment Script
# This script helps deploy user-managed executors with the singleton operator

set -euo pipefail

# Default values
EXECUTOR_NAME=""
NAMESPACE=""
EXECUTOR_IMAGE="hourglass/executor:v1.2.0"
EXECUTOR_VERSION="v1.2.0"
AGGREGATOR_ENDPOINT=""
AVS_ADDRESS=""
ETH_RPC_URL=""
ECDSA_PRIVATE_KEY=""
BLS_PRIVATE_KEY=""
DEPLOYMENT_TYPE="basic"
DRY_RUN=false
APPLY=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Usage function
usage() {
    cat << EOF
Hourglass Executor Deployment Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -n, --name EXECUTOR_NAME         Name of the executor (required)
    -s, --namespace NAMESPACE        Kubernetes namespace (required)
    -i, --image IMAGE                Executor container image (default: hourglass/executor:v1.2.0)
    -v, --version VERSION            Executor version (default: v1.2.0)
    -a, --aggregator ENDPOINT        Aggregator endpoint (required)
    -A, --avs-address ADDRESS        AVS contract address (required)
    -r, --rpc-url URL               Ethereum RPC URL (required)
    -e, --ecdsa-key KEY             ECDSA private key (required)
    -b, --bls-key KEY               BLS private key (optional)
    -t, --type TYPE                 Deployment type: basic|production|multi-gpu|tee (default: basic)
    -d, --dry-run                   Generate manifests without applying
    -y, --apply                     Apply manifests to cluster
    -h, --help                      Show this help message

EXAMPLES:
    # Basic deployment (dry run)
    $0 -n my-executor -s my-namespace -a aggregator.example.com:9090 \\
       -A 0x1234... -r https://eth-mainnet.alchemyapi.io/v2/KEY \\
       -e 0xabcdef... --dry-run

    # Production deployment with application
    $0 -n prod-executor -s production-avs -t production \\
       -a aggregator-ha.example.com:9090 -A 0x1234... \\
       -r https://eth-mainnet.infura.io/v3/KEY -e 0xabcdef... \\
       -b 0x567890... --apply

    # Multi-GPU deployment
    $0 -n gpu-executor -s ml-workloads -t multi-gpu \\
       -a aggregator.ml.com:9090 -A 0x1234... \\
       -r https://eth-mainnet.alchemyapi.io/v2/KEY -e 0xabcdef... --apply

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--name)
            EXECUTOR_NAME="$2"
            shift 2
            ;;
        -s|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -i|--image)
            EXECUTOR_IMAGE="$2"
            shift 2
            ;;
        -v|--version)
            EXECUTOR_VERSION="$2"
            shift 2
            ;;
        -a|--aggregator)
            AGGREGATOR_ENDPOINT="$2"
            shift 2
            ;;
        -A|--avs-address)
            AVS_ADDRESS="$2"
            shift 2
            ;;
        -r|--rpc-url)
            ETH_RPC_URL="$2"
            shift 2
            ;;
        -e|--ecdsa-key)
            ECDSA_PRIVATE_KEY="$2"
            shift 2
            ;;
        -b|--bls-key)
            BLS_PRIVATE_KEY="$2"
            shift 2
            ;;
        -t|--type)
            DEPLOYMENT_TYPE="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -y|--apply)
            APPLY=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Validate required parameters
if [[ -z "$EXECUTOR_NAME" ]]; then
    log_error "Executor name is required (-n/--name)"
    exit 1
fi

if [[ -z "$NAMESPACE" ]]; then
    log_error "Namespace is required (-s/--namespace)"
    exit 1
fi

if [[ -z "$AGGREGATOR_ENDPOINT" ]]; then
    log_error "Aggregator endpoint is required (-a/--aggregator)"
    exit 1
fi

if [[ -z "$AVS_ADDRESS" ]]; then
    log_error "AVS address is required (-A/--avs-address)"
    exit 1
fi

if [[ -z "$ETH_RPC_URL" ]]; then
    log_error "Ethereum RPC URL is required (-r/--rpc-url)"
    exit 1
fi

if [[ -z "$ECDSA_PRIVATE_KEY" ]]; then
    log_error "ECDSA private key is required (-e/--ecdsa-key)"
    exit 1
fi

# Validate deployment type
case $DEPLOYMENT_TYPE in
    basic|production|multi-gpu|tee)
        ;;
    *)
        log_error "Invalid deployment type: $DEPLOYMENT_TYPE"
        log_error "Valid types: basic, production, multi-gpu, tee"
        exit 1
        ;;
esac

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl is not installed or not in PATH"
    exit 1
fi

# Check if we can connect to the cluster
if ! kubectl cluster-info &> /dev/null; then
    log_error "Cannot connect to Kubernetes cluster"
    exit 1
fi

# Function to encode to base64
encode_base64() {
    echo -n "$1" | base64 -w 0
}

# Generate secrets
ECDSA_PRIVATE_KEY_B64=$(encode_base64 "$ECDSA_PRIVATE_KEY")
BLS_PRIVATE_KEY_B64=""
if [[ -n "$BLS_PRIVATE_KEY" ]]; then
    BLS_PRIVATE_KEY_B64=$(encode_base64 "$BLS_PRIVATE_KEY")
fi

# Set deployment-specific variables
case $DEPLOYMENT_TYPE in
    basic)
        REPLICA_COUNT=1
        CPU_REQUESTS="500m"
        MEMORY_REQUESTS="1Gi"
        CPU_LIMITS="2"
        MEMORY_LIMITS="4Gi"
        STORAGE_SIZE="10Gi"
        STORAGE_CLASS="standard"
        ;;
    production)
        REPLICA_COUNT=3
        CPU_REQUESTS="1"
        MEMORY_REQUESTS="2Gi"
        CPU_LIMITS="4"
        MEMORY_LIMITS="8Gi"
        STORAGE_SIZE="50Gi"
        STORAGE_CLASS="fast-ssd"
        ;;
    multi-gpu)
        REPLICA_COUNT=1
        CPU_REQUESTS="2"
        MEMORY_REQUESTS="4Gi"
        CPU_LIMITS="8"
        MEMORY_LIMITS="16Gi"
        STORAGE_SIZE="100Gi"
        STORAGE_CLASS="fast-ssd"
        ;;
    tee)
        REPLICA_COUNT=2
        CPU_REQUESTS="1"
        MEMORY_REQUESTS="2Gi"
        CPU_LIMITS="4"
        MEMORY_LIMITS="8Gi"
        STORAGE_SIZE="20Gi"
        STORAGE_CLASS="fast-ssd"
        ;;
esac

# Output directory
OUTPUT_DIR="./generated-manifests/${EXECUTOR_NAME}"
mkdir -p "$OUTPUT_DIR"

log_info "Generating manifests for executor: $EXECUTOR_NAME"
log_info "Namespace: $NAMESPACE"
log_info "Deployment type: $DEPLOYMENT_TYPE"

# Function to substitute variables in template
substitute_template() {
    local template_file="$1"
    local output_file="$2"
    
    cat "$template_file" | \
    sed "s/\${EXECUTOR_NAME}/$EXECUTOR_NAME/g" | \
    sed "s/\${NAMESPACE}/$NAMESPACE/g" | \
    sed "s/\${EXECUTOR_IMAGE}/$EXECUTOR_IMAGE/g" | \
    sed "s/\${EXECUTOR_VERSION}/$EXECUTOR_VERSION/g" | \
    sed "s/\${AGGREGATOR_ENDPOINT}/$AGGREGATOR_ENDPOINT/g" | \
    sed "s/\${AVS_ADDRESS_1}/$AVS_ADDRESS/g" | \
    sed "s/\${ETH_RPC_URL}/$ETH_RPC_URL/g" | \
    sed "s/\${ECDSA_PRIVATE_KEY_B64}/$ECDSA_PRIVATE_KEY_B64/g" | \
    sed "s/\${BLS_PRIVATE_KEY_B64}/$BLS_PRIVATE_KEY_B64/g" | \
    sed "s/\${REPLICA_COUNT:-1}/$REPLICA_COUNT/g" | \
    sed "s/\${CPU_REQUESTS:-500m}/$CPU_REQUESTS/g" | \
    sed "s/\${MEMORY_REQUESTS:-1Gi}/$MEMORY_REQUESTS/g" | \
    sed "s/\${CPU_LIMITS:-2}/$CPU_LIMITS/g" | \
    sed "s/\${MEMORY_LIMITS:-4Gi}/$MEMORY_LIMITS/g" | \
    sed "s/\${DATA_STORAGE_SIZE:-10Gi}/$STORAGE_SIZE/g" | \
    sed "s/\${STORAGE_CLASS:-standard}/$STORAGE_CLASS/g" | \
    sed "s/\${SERVICE_ACCOUNT_NAME:-executor-service-account}/executor-service-account/g" \
    > "$output_file"
}

# Generate namespace manifest
cat > "$OUTPUT_DIR/namespace.yaml" << EOF
apiVersion: v1
kind: Namespace
metadata:
  name: $NAMESPACE
  labels:
    app.kubernetes.io/name: hourglass-executor
    app.kubernetes.io/part-of: hourglass
    deployment-type: $DEPLOYMENT_TYPE
EOF

# Generate RBAC manifests
log_info "Generating RBAC manifests..."
substitute_template "rbac/executor-rbac.yaml" "$OUTPUT_DIR/rbac.yaml"

# Generate ConfigMap manifests
log_info "Generating ConfigMap manifests..."
substitute_template "executor/configmap.yaml" "$OUTPUT_DIR/configmap.yaml"

# Generate Secret manifests
log_info "Generating Secret manifests..."
substitute_template "executor/secrets.yaml" "$OUTPUT_DIR/secrets.yaml"

# Generate StatefulSet manifests
log_info "Generating StatefulSet manifests..."
substitute_template "executor/statefulset.yaml" "$OUTPUT_DIR/statefulset.yaml"

# Generate combined manifest
log_info "Generating combined manifest..."
cat "$OUTPUT_DIR/namespace.yaml" \
    "$OUTPUT_DIR/rbac.yaml" \
    "$OUTPUT_DIR/secrets.yaml" \
    "$OUTPUT_DIR/configmap.yaml" \
    "$OUTPUT_DIR/statefulset.yaml" \
    > "$OUTPUT_DIR/all-in-one.yaml"

log_success "Manifests generated in: $OUTPUT_DIR"

# Apply manifests if requested
if [[ "$APPLY" == "true" ]]; then
    log_info "Applying manifests to cluster..."
    
    # Check if namespace exists
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_warning "Namespace $NAMESPACE already exists"
    else
        log_info "Creating namespace: $NAMESPACE"
        kubectl apply -f "$OUTPUT_DIR/namespace.yaml"
    fi
    
    # Apply other manifests
    kubectl apply -f "$OUTPUT_DIR/rbac.yaml"
    kubectl apply -f "$OUTPUT_DIR/secrets.yaml"
    kubectl apply -f "$OUTPUT_DIR/configmap.yaml"
    kubectl apply -f "$OUTPUT_DIR/statefulset.yaml"
    
    log_success "Manifests applied successfully!"
    
    # Wait for deployment
    log_info "Waiting for executor to be ready..."
    kubectl wait --for=condition=Ready pod -l app=$EXECUTOR_NAME -n $NAMESPACE --timeout=300s
    
    log_success "Executor is ready!"
    
    # Show status
    echo
    log_info "Executor status:"
    kubectl get statefulsets,pods,services -n $NAMESPACE -l app=$EXECUTOR_NAME
    
    echo
    log_info "To view logs:"
    echo "kubectl logs -f statefulset/$EXECUTOR_NAME -n $NAMESPACE"
    
    echo
    log_info "To check performers:"
    echo "kubectl get performers -n $NAMESPACE"
    
elif [[ "$DRY_RUN" == "true" ]]; then
    log_info "Dry run - manifests generated but not applied"
    echo
    log_info "To apply manually:"
    echo "kubectl apply -f $OUTPUT_DIR/all-in-one.yaml"
    
    echo
    log_info "To apply step by step:"
    echo "kubectl apply -f $OUTPUT_DIR/namespace.yaml"
    echo "kubectl apply -f $OUTPUT_DIR/rbac.yaml"
    echo "kubectl apply -f $OUTPUT_DIR/secrets.yaml"
    echo "kubectl apply -f $OUTPUT_DIR/configmap.yaml"
    echo "kubectl apply -f $OUTPUT_DIR/statefulset.yaml"
else
    log_info "Manifests generated - use --apply to deploy or --dry-run to preview"
fi

echo
log_success "Deployment script completed!"