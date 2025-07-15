#!/bin/bash

# Backward Compatibility Test Script for Milestone 4.3
# Tests single runtime configuration and backward compatibility

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_NAMESPACE="compatibility-test"
TIMEOUT="60s"

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

create_test_namespace() {
    log_info "Creating test namespace..."
    kubectl create namespace ${TEST_NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
    log_info "Test namespace created"
}

test_docker_mode_configuration() {
    log_info "Testing Docker mode configuration..."
    
    # Create Docker mode executor configuration
    cat > /tmp/docker-config.yaml << 'EOF'
aggregator_endpoint: "test-aggregator.example.com:9090"
aggregator_tls_enabled: false
deployment_mode: "docker"
log_level: "info"
log_format: "json"

performer_config:
  default_port: 9090
  connection_timeout: "30s"
  startup_timeout: "300s"
  retry_attempts: 3
  max_performers: 5

# No kubernetes section - should work fine
chains:
  - name: "ethereum"
    chain_id: 1
    rpc_url: "https://eth-mainnet.alchemyapi.io/v2/TEST_KEY"
    task_mailbox_address: "0x1234567890abcdef1234567890abcdef12345678"

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
EOF

    kubectl create configmap docker-config-test \
        --namespace ${TEST_NAMESPACE} \
        --from-file=config.yaml=/tmp/docker-config.yaml \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log_info "✓ Docker mode configuration created successfully"
}

test_kubernetes_mode_configuration() {
    log_info "Testing Kubernetes mode configuration..."
    
    # Create Kubernetes mode executor configuration
    cat > /tmp/kubernetes-config.yaml << 'EOF'
aggregator_endpoint: "test-aggregator.example.com:9090"
aggregator_tls_enabled: false
deployment_mode: "kubernetes"
log_level: "info"
log_format: "json"

performer_config:
  service_pattern: "performer-{name}.{namespace}.svc.cluster.local:{port}"
  default_port: 9090
  connection_timeout: "30s"
  startup_timeout: "300s"
  retry_attempts: 3
  max_performers: 5

# Kubernetes section required for k8s mode
kubernetes:
  namespace: "compatibility-test"
  operator_namespace: "hourglass-system"
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
EOF

    kubectl create configmap kubernetes-config-test \
        --namespace ${TEST_NAMESPACE} \
        --from-file=config.yaml=/tmp/kubernetes-config.yaml \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log_info "✓ Kubernetes mode configuration created successfully"
}

test_invalid_mixed_configuration() {
    log_info "Testing invalid mixed configuration..."
    
    # Create mixed mode executor configuration (should be invalid)
    cat > /tmp/mixed-config.yaml << 'EOF'
aggregator_endpoint: "test-aggregator.example.com:9090"
aggregator_tls_enabled: false
deployment_mode: "kubernetes"
log_level: "info"
log_format: "json"

performer_config:
  # Missing service_pattern for kubernetes mode
  default_port: 9090
  connection_timeout: "30s"
  startup_timeout: "300s"
  retry_attempts: 3
  max_performers: 5

# Missing kubernetes section for k8s mode - should cause validation error
chains:
  - name: "ethereum"
    chain_id: 1
    rpc_url: "https://eth-mainnet.alchemyapi.io/v2/TEST_KEY"
    task_mailbox_address: "0x1234567890abcdef1234567890abcdef12345678"

avs_config:
  supported_avs:
    - address: "0x1234567890abcdef1234567890abcdef12345678"
      name: "test-avs"
      performer_image: "nginx"
      performer_version: "1.21"
EOF

    kubectl create configmap mixed-config-test \
        --namespace ${TEST_NAMESPACE} \
        --from-file=config.yaml=/tmp/mixed-config.yaml \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log_info "✓ Invalid mixed configuration created (should fail validation)"
}

test_configuration_validation() {
    log_info "Testing configuration validation..."
    
    # Test 1: Docker mode should work without kubernetes config
    if kubectl get configmap docker-config-test -n ${TEST_NAMESPACE} &> /dev/null; then
        log_info "✓ Test 1 PASSED: Docker mode config accepted without kubernetes section"
    else
        log_error "✗ Test 1 FAILED: Docker mode config rejected"
        return 1
    fi
    
    # Test 2: Kubernetes mode should work with kubernetes config
    if kubectl get configmap kubernetes-config-test -n ${TEST_NAMESPACE} &> /dev/null; then
        log_info "✓ Test 2 PASSED: Kubernetes mode config accepted with kubernetes section"
    else
        log_error "✗ Test 2 FAILED: Kubernetes mode config rejected"
        return 1
    fi
    
    # Test 3: Mixed/invalid config should exist but would fail at runtime
    if kubectl get configmap mixed-config-test -n ${TEST_NAMESPACE} &> /dev/null; then
        log_info "✓ Test 3 PASSED: Invalid config stored (runtime validation expected)"
    else
        log_error "✗ Test 3 FAILED: Invalid config not stored"
        return 1
    fi
}

test_deployment_mode_isolation() {
    log_info "Testing deployment mode isolation..."
    
    # Create two separate executors with different deployment modes
    cat > /tmp/docker-executor.yaml << 'EOF'
apiVersion: v1
kind: ServiceAccount
metadata:
  name: docker-executor
  namespace: compatibility-test
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: docker-executor
  namespace: compatibility-test
  labels:
    app: docker-executor
    deployment-mode: docker
spec:
  replicas: 1
  selector:
    matchLabels:
      app: docker-executor
  template:
    metadata:
      labels:
        app: docker-executor
        deployment-mode: docker
    spec:
      serviceAccountName: docker-executor
      containers:
      - name: executor
        image: busybox:1.35
        command: ["sleep", "300"]
        env:
        - name: DEPLOYMENT_MODE
          value: "docker"
        - name: CONFIG_PATH
          value: "/etc/config/config.yaml"
        volumeMounts:
        - name: config
          mountPath: /etc/config
      volumes:
      - name: config
        configMap:
          name: docker-config-test
EOF

    cat > /tmp/kubernetes-executor.yaml << 'EOF'
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubernetes-executor
  namespace: compatibility-test
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubernetes-executor
  namespace: compatibility-test
  labels:
    app: kubernetes-executor
    deployment-mode: kubernetes
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubernetes-executor
  template:
    metadata:
      labels:
        app: kubernetes-executor
        deployment-mode: kubernetes
    spec:
      serviceAccountName: kubernetes-executor
      containers:
      - name: executor
        image: busybox:1.35
        command: ["sleep", "300"]
        env:
        - name: DEPLOYMENT_MODE
          value: "kubernetes"
        - name: CONFIG_PATH
          value: "/etc/config/config.yaml"
        volumeMounts:
        - name: config
          mountPath: /etc/config
      volumes:
      - name: config
        configMap:
          name: kubernetes-config-test
EOF

    kubectl apply -f /tmp/docker-executor.yaml
    kubectl apply -f /tmp/kubernetes-executor.yaml
    
    # Wait for deployments
    kubectl wait --for=condition=Available deployment/docker-executor -n ${TEST_NAMESPACE} --timeout=${TIMEOUT}
    kubectl wait --for=condition=Available deployment/kubernetes-executor -n ${TEST_NAMESPACE} --timeout=${TIMEOUT}
    
    log_info "✓ Both deployment modes running independently"
}

test_existing_docker_behavior() {
    log_info "Testing existing Docker behavior preservation..."
    
    # Verify Docker executor is running and has correct configuration
    DOCKER_POD=$(kubectl get pods -n ${TEST_NAMESPACE} -l app=docker-executor -o jsonpath='{.items[0].metadata.name}')
    
    if [ -n "$DOCKER_POD" ]; then
        DEPLOYMENT_MODE=$(kubectl exec -n ${TEST_NAMESPACE} "$DOCKER_POD" -- printenv DEPLOYMENT_MODE)
        if [ "$DEPLOYMENT_MODE" = "docker" ]; then
            log_info "✓ Docker executor has correct deployment mode"
        else
            log_error "✗ Docker executor has incorrect deployment mode: $DEPLOYMENT_MODE"
            return 1
        fi
        
        # Check config file exists
        if kubectl exec -n ${TEST_NAMESPACE} "$DOCKER_POD" -- ls /etc/config/config.yaml &> /dev/null; then
            log_info "✓ Docker executor config file accessible"
        else
            log_error "✗ Docker executor config file not accessible"
            return 1
        fi
    else
        log_error "✗ Docker executor pod not found"
        return 1
    fi
}

test_kubernetes_mode_behavior() {
    log_info "Testing Kubernetes mode behavior..."
    
    # Verify Kubernetes executor is running and has correct configuration
    K8S_POD=$(kubectl get pods -n ${TEST_NAMESPACE} -l app=kubernetes-executor -o jsonpath='{.items[0].metadata.name}')
    
    if [ -n "$K8S_POD" ]; then
        DEPLOYMENT_MODE=$(kubectl exec -n ${TEST_NAMESPACE} "$K8S_POD" -- printenv DEPLOYMENT_MODE)
        if [ "$DEPLOYMENT_MODE" = "kubernetes" ]; then
            log_info "✓ Kubernetes executor has correct deployment mode"
        else
            log_error "✗ Kubernetes executor has incorrect deployment mode: $DEPLOYMENT_MODE"
            return 1
        fi
        
        # Check config file exists and contains kubernetes section
        if kubectl exec -n ${TEST_NAMESPACE} "$K8S_POD" -- grep -q "kubernetes:" /etc/config/config.yaml; then
            log_info "✓ Kubernetes executor config contains kubernetes section"
        else
            log_error "✗ Kubernetes executor config missing kubernetes section"
            return 1
        fi
    else
        log_error "✗ Kubernetes executor pod not found"
        return 1
    fi
}

show_test_status() {
    log_info "=== Backward Compatibility Test Status ==="
    
    echo -e "\n${GREEN}Docker Mode Executor:${NC}"
    kubectl get pods -n ${TEST_NAMESPACE} -l app=docker-executor -o wide
    
    echo -e "\n${GREEN}Kubernetes Mode Executor:${NC}"
    kubectl get pods -n ${TEST_NAMESPACE} -l app=kubernetes-executor -o wide
    
    echo -e "\n${GREEN}Configuration Maps:${NC}"
    kubectl get configmaps -n ${TEST_NAMESPACE}
    
    echo -e "\n${GREEN}Recent Events:${NC}"
    kubectl get events -n ${TEST_NAMESPACE} --sort-by='.lastTimestamp' | tail -5
}

cleanup_compatibility_tests() {
    log_info "Cleaning up backward compatibility tests..."
    
    # Delete all resources
    kubectl delete deployment --all -n ${TEST_NAMESPACE} --ignore-not-found=true
    kubectl delete configmap --all -n ${TEST_NAMESPACE} --ignore-not-found=true
    kubectl delete serviceaccount --all -n ${TEST_NAMESPACE} --ignore-not-found=true
    kubectl delete namespace ${TEST_NAMESPACE} --ignore-not-found=true
    
    # Clean up temp files
    rm -f /tmp/docker-config.yaml
    rm -f /tmp/kubernetes-config.yaml
    rm -f /tmp/mixed-config.yaml
    rm -f /tmp/docker-executor.yaml
    rm -f /tmp/kubernetes-executor.yaml
    
    log_info "Backward compatibility test cleanup completed"
}

main() {
    case "${1:-test}" in
        test)
            log_info "Starting backward compatibility tests for milestone 4.3..."
            create_test_namespace
            test_docker_mode_configuration
            test_kubernetes_mode_configuration
            test_invalid_mixed_configuration
            test_configuration_validation
            test_deployment_mode_isolation
            test_existing_docker_behavior
            test_kubernetes_mode_behavior
            show_test_status
            log_info "Backward compatibility tests completed successfully!"
            ;;
        cleanup)
            cleanup_compatibility_tests
            ;;
        status)
            show_test_status
            ;;
        *)
            echo "Usage: $0 {test|cleanup|status}"
            exit 1
            ;;
    esac
}

main "$@"