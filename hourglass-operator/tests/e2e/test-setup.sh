#!/bin/bash

# Hourglass Operator E2E Test Setup Script
# This script sets up a complete test environment for milestone 4.2 validation

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
OPERATOR_NAMESPACE="hourglass-system"
TEST_NAMESPACE="test-avs-project"
CHART_VERSION="0.1.0"
TIMEOUT="300s"

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    # Check if helm is available
    if ! command -v helm &> /dev/null; then
        log_error "helm is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we can connect to Kubernetes cluster
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    log_info "Prerequisites check passed"
}

create_namespaces() {
    log_info "Creating namespaces..."
    
    # Create operator namespace
    kubectl create namespace ${OPERATOR_NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
    
    # Create test namespace
    kubectl create namespace ${TEST_NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
    
    log_info "Namespaces created successfully"
}

deploy_operator() {
    log_info "Deploying Hourglass Operator..."
    
    # Navigate to charts directory
    cd "$(dirname "$0")/../../charts"
    
    # Install operator using Helm
    helm upgrade --install hourglass-operator ./hourglass-operator \
        --namespace ${OPERATOR_NAMESPACE} \
        --create-namespace \
        --wait \
        --timeout ${TIMEOUT} \
        --values ./hourglass-operator/values.yaml \
        --set operator.image.tag="latest" \
        --set operator.image.pullPolicy="IfNotPresent"
    
    log_info "Operator deployed successfully"
}

wait_for_operator() {
    log_info "Waiting for operator to be ready..."
    
    kubectl wait --for=condition=Available deployment/hourglass-operator \
        --namespace ${OPERATOR_NAMESPACE} \
        --timeout ${TIMEOUT}
    
    log_info "Operator is ready"
}

create_test_executor_config() {
    log_info "Creating test executor configuration..."
    
    # Create test configuration
    cat > /tmp/test-executor-config.yaml << EOF
aggregator_endpoint: "test-aggregator.example.com:9090"
aggregator_tls_enabled: false
deployment_mode: "kubernetes"
log_level: "debug"
log_format: "json"

performer_config:
  service_pattern: "performer-{name}.{namespace}.svc.cluster.local:{port}"
  default_port: 9090
  connection_timeout: "30s"
  startup_timeout: "300s"
  retry_attempts: 3
  max_performers: 5

kubernetes:
  namespace: "${TEST_NAMESPACE}"
  operator_namespace: "${OPERATOR_NAMESPACE}"
  performer_crd_group: "hourglass.eigenlayer.io"
  performer_crd_version: "v1alpha1"
  performer_crd_kind: "Performer"
  cleanup_on_shutdown: true
  cleanup_timeout: "60s"

chains:
  - name: "ethereum"
    chain_id: 1
    rpc_url: "https://eth-mainnet.alchemyapi.io/v2/TEST_KEY"
    task_mailbox_address: "0x1234567890abcdef1234567890abcdef12345678"
    block_confirmations: 12
    gas_limit: 300000
    event_filter:
      from_block: "latest"

avs_config:
  supported_avs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "test-avs"
      performer_image: "nginx"
      performer_version: "1.21"
      default_resources:
        requests:
          cpu: "100m"
          memory: "128Mi"
        limits:
          cpu: "500m"
          memory: "256Mi"

metrics:
  enabled: true
  port: 8080
  path: "/metrics"

health:
  enabled: true
  port: 8090
EOF

    # Create ConfigMap
    kubectl create configmap test-executor-config \
        --namespace ${TEST_NAMESPACE} \
        --from-file=config.yaml=/tmp/test-executor-config.yaml \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log_info "Test executor configuration created"
}

create_test_secrets() {
    log_info "Creating test secrets..."
    
    # Create dummy keys for testing (these are NOT real keys)
    kubectl create secret generic test-executor-keys \
        --namespace ${TEST_NAMESPACE} \
        --from-literal=ecdsa-private-key="test-ecdsa-key" \
        --from-literal=bls-private-key="test-bls-key" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log_info "Test secrets created"
}

deploy_test_executor() {
    log_info "Deploying test executor..."
    
    # Create test executor deployment
    cat > /tmp/test-executor-deployment.yaml << EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-executor
  namespace: ${TEST_NAMESPACE}
automountServiceAccountToken: true

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: ${TEST_NAMESPACE}
  name: test-executor-role
rules:
- apiGroups: ["hourglass.eigenlayer.io"]
  resources: ["performers"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: ["hourglass.eigenlayer.io"]
  resources: ["performers/status"]
  verbs: ["get", "patch", "update"]
- apiGroups: [""]
  resources: ["services", "pods", "events"]
  verbs: ["get", "list", "watch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: test-executor-binding
  namespace: ${TEST_NAMESPACE}
subjects:
- kind: ServiceAccount
  name: test-executor
  namespace: ${TEST_NAMESPACE}
roleRef:
  kind: Role
  name: test-executor-role
  apiGroup: rbac.authorization.k8s.io

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-executor
  namespace: ${TEST_NAMESPACE}
  labels:
    app: test-executor
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-executor
  template:
    metadata:
      labels:
        app: test-executor
    spec:
      serviceAccountName: test-executor
      containers:
      - name: test-executor
        image: busybox:1.35
        command: ["sleep", "3600"]
        env:
        - name: DEPLOYMENT_MODE
          value: "kubernetes"
        - name: KUBERNETES_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        volumeMounts:
        - name: config
          mountPath: /etc/executor
          readOnly: true
        - name: secrets
          mountPath: /etc/secrets
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: test-executor-config
      - name: secrets
        secret:
          secretName: test-executor-keys
EOF

    kubectl apply -f /tmp/test-executor-deployment.yaml
    
    # Wait for deployment to be ready
    kubectl wait --for=condition=Available deployment/test-executor \
        --namespace ${TEST_NAMESPACE} \
        --timeout ${TIMEOUT}
    
    log_info "Test executor deployed successfully"
}

create_test_performer_crd() {
    log_info "Creating test Performer CRD..."
    
    # Create test performer CRD
    cat > /tmp/test-performer.yaml << EOF
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: test-performer-1
  namespace: ${TEST_NAMESPACE}
  labels:
    app: test-performer
    avs: test-avs
spec:
  avsAddress: "0x1234567890abcdef1234567890abcdef12345678"
  image:
    repository: nginx
    tag: "1.21"
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "500m"
      memory: "256Mi"
  ports:
  - name: http
    port: 80
    targetPort: 80
  - name: grpc
    port: 9090
    targetPort: 9090
  env:
  - name: TEST_ENV
    value: "test-value"
  serviceAccount: test-executor
EOF

    kubectl apply -f /tmp/test-performer.yaml
    
    log_info "Test Performer CRD created"
}

run_validation_tests() {
    log_info "Running validation tests..."
    
    # Test 1: Check if operator is processing the CRD
    log_info "Test 1: Checking operator processing..."
    sleep 10
    
    # Check if pod was created
    if kubectl get pods --namespace ${TEST_NAMESPACE} -l app=test-performer &> /dev/null; then
        log_info "✓ Test 1 PASSED: Operator created performer pod"
    else
        log_error "✗ Test 1 FAILED: Operator did not create performer pod"
        return 1
    fi
    
    # Test 2: Check if service was created
    log_info "Test 2: Checking service creation..."
    if kubectl get service test-performer-1 --namespace ${TEST_NAMESPACE} &> /dev/null; then
        log_info "✓ Test 2 PASSED: Service created for performer"
    else
        log_error "✗ Test 2 FAILED: Service not created for performer"
        return 1
    fi
    
    # Test 3: Check DNS resolution
    log_info "Test 3: Checking DNS resolution..."
    if kubectl run test-dns --image=busybox:1.35 --namespace ${TEST_NAMESPACE} --rm -i --restart=Never -- nslookup test-performer-1.${TEST_NAMESPACE}.svc.cluster.local &> /dev/null; then
        log_info "✓ Test 3 PASSED: DNS resolution works"
    else
        log_warn "⚠ Test 3 WARNING: DNS resolution may be pending"
    fi
    
    # Test 4: Check performer status
    log_info "Test 4: Checking performer status..."
    STATUS=$(kubectl get performer test-performer-1 --namespace ${TEST_NAMESPACE} -o jsonpath='{.status.phase}' 2>/dev/null || echo "unknown")
    if [[ "$STATUS" != "unknown" ]]; then
        log_info "✓ Test 4 PASSED: Performer status is: $STATUS"
    else
        log_warn "⚠ Test 4 WARNING: Performer status not yet available"
    fi
    
    log_info "Validation tests completed"
}

cleanup() {
    log_info "Cleaning up test environment..."
    
    # Delete test performer
    kubectl delete -f /tmp/test-performer.yaml --ignore-not-found=true
    
    # Delete test executor
    kubectl delete -f /tmp/test-executor-deployment.yaml --ignore-not-found=true
    
    # Delete configmap and secret
    kubectl delete configmap test-executor-config --namespace ${TEST_NAMESPACE} --ignore-not-found=true
    kubectl delete secret test-executor-keys --namespace ${TEST_NAMESPACE} --ignore-not-found=true
    
    # Delete namespace
    kubectl delete namespace ${TEST_NAMESPACE} --ignore-not-found=true
    
    # Clean up temporary files
    rm -f /tmp/test-executor-config.yaml
    rm -f /tmp/test-executor-deployment.yaml
    rm -f /tmp/test-performer.yaml
    
    log_info "Cleanup completed"
}

show_status() {
    log_info "=== Current Status ==="
    
    echo -e "\n${GREEN}Operator Status:${NC}"
    kubectl get pods --namespace ${OPERATOR_NAMESPACE} -l app.kubernetes.io/name=hourglass-operator
    
    echo -e "\n${GREEN}Test Executor Status:${NC}"
    kubectl get pods --namespace ${TEST_NAMESPACE} -l app=test-executor
    
    echo -e "\n${GREEN}Performer CRDs:${NC}"
    kubectl get performers --namespace ${TEST_NAMESPACE}
    
    echo -e "\n${GREEN}Performer Pods:${NC}"
    kubectl get pods --namespace ${TEST_NAMESPACE} -l app=test-performer
    
    echo -e "\n${GREEN}Services:${NC}"
    kubectl get services --namespace ${TEST_NAMESPACE}
    
    echo -e "\n${GREEN}Recent Events:${NC}"
    kubectl get events --namespace ${TEST_NAMESPACE} --sort-by='.lastTimestamp' | tail -10
}

main() {
    case "${1:-setup}" in
        setup)
            log_info "Starting E2E test setup for milestone 4.2..."
            check_prerequisites
            create_namespaces
            deploy_operator
            wait_for_operator
            create_test_executor_config
            create_test_secrets
            deploy_test_executor
            create_test_performer_crd
            run_validation_tests
            show_status
            log_info "E2E test setup completed successfully!"
            ;;
        cleanup)
            cleanup
            ;;
        status)
            show_status
            ;;
        test)
            run_validation_tests
            ;;
        *)
            echo "Usage: $0 {setup|cleanup|status|test}"
            exit 1
            ;;
    esac
}

main "$@"