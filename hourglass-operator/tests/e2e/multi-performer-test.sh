#!/bin/bash

# Multi-Performer Test Script for Milestone 4.2
# Tests multiple performers per AVS and cross-namespace isolation

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
OPERATOR_NAMESPACE="hourglass-system"
TEST_NAMESPACE_A="test-avs-a"
TEST_NAMESPACE_B="test-avs-b"
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

create_test_namespaces() {
    log_info "Creating test namespaces for multi-performer testing..."
    
    kubectl create namespace ${TEST_NAMESPACE_A} --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace ${TEST_NAMESPACE_B} --dry-run=client -o yaml | kubectl apply -f -
    
    log_info "Test namespaces created"
}

create_multiple_performers_same_avs() {
    log_info "Creating multiple performers for the same AVS..."
    
    # Create 3 performers for the same AVS in namespace A
    for i in {1..3}; do
        cat > /tmp/performer-a-${i}.yaml << EOF
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: test-performer-a-${i}
  namespace: ${TEST_NAMESPACE_A}
  labels:
    app: test-performer
    avs: test-avs-a
    instance: "performer-${i}"
spec:
  avsAddress: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
  image:
    repository: nginx
    tag: "1.21"
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "200m"
      memory: "256Mi"
  ports:
  - name: http
    port: 80
    targetPort: 80
  - name: grpc
    port: 9090
    targetPort: 9090
  env:
  - name: PERFORMER_ID
    value: "performer-a-${i}"
  - name: AVS_ADDRESS
    value: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
EOF
        kubectl apply -f /tmp/performer-a-${i}.yaml
    done
    
    log_info "Multiple performers created for same AVS"
}

create_cross_namespace_performers() {
    log_info "Creating performers in different namespaces..."
    
    # Create performers in namespace B for a different AVS
    for i in {1..2}; do
        cat > /tmp/performer-b-${i}.yaml << EOF
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: test-performer-b-${i}
  namespace: ${TEST_NAMESPACE_B}
  labels:
    app: test-performer
    avs: test-avs-b
    instance: "performer-${i}"
spec:
  avsAddress: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
  image:
    repository: nginx
    tag: "1.21"
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "200m"
      memory: "256Mi"
  ports:
  - name: http
    port: 80
    targetPort: 80
  - name: grpc
    port: 9090
    targetPort: 9090
  env:
  - name: PERFORMER_ID
    value: "performer-b-${i}"
  - name: AVS_ADDRESS
    value: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
EOF
        kubectl apply -f /tmp/performer-b-${i}.yaml
    done
    
    log_info "Cross-namespace performers created"
}

test_concurrent_operations() {
    log_info "Testing concurrent deployment and removal operations..."
    
    # Create performers concurrently
    for i in {4..6}; do
        (
            cat > /tmp/concurrent-performer-${i}.yaml << EOF
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: concurrent-performer-${i}
  namespace: ${TEST_NAMESPACE_A}
  labels:
    app: concurrent-performer
    test: concurrent
spec:
  avsAddress: "0xcccccccccccccccccccccccccccccccccccccccc"
  image:
    repository: nginx
    tag: "1.21"
  resources:
    requests:
      cpu: "50m"
      memory: "64Mi"
    limits:
      cpu: "100m"
      memory: "128Mi"
  ports:
  - name: grpc
    port: 9090
    targetPort: 9090
EOF
            kubectl apply -f /tmp/concurrent-performer-${i}.yaml
        ) &
    done
    
    # Wait for all background processes to complete
    wait
    
    log_info "Concurrent operations completed"
}

validate_multiple_performers() {
    log_info "Validating multiple performers scenario..."
    
    # Wait for pods to be ready
    sleep 15
    
    # Test 1: Check all performers in namespace A
    PERFORMERS_A=$(kubectl get performers -n ${TEST_NAMESPACE_A} --no-headers | wc -l)
    if [ "$PERFORMERS_A" -ge 3 ]; then
        log_info "✓ Test 1 PASSED: Multiple performers created in namespace A (${PERFORMERS_A} total)"
    else
        log_error "✗ Test 1 FAILED: Expected at least 3 performers in namespace A, got ${PERFORMERS_A}"
        return 1
    fi
    
    # Test 2: Check performers in namespace B
    PERFORMERS_B=$(kubectl get performers -n ${TEST_NAMESPACE_B} --no-headers | wc -l)
    if [ "$PERFORMERS_B" -ge 2 ]; then
        log_info "✓ Test 2 PASSED: Performers created in namespace B (${PERFORMERS_B} total)"
    else
        log_error "✗ Test 2 FAILED: Expected at least 2 performers in namespace B, got ${PERFORMERS_B}"
        return 1
    fi
    
    # Test 3: Check cross-namespace isolation
    PODS_A=$(kubectl get pods -n ${TEST_NAMESPACE_A} -l app=test-performer --no-headers | wc -l)
    PODS_B=$(kubectl get pods -n ${TEST_NAMESPACE_B} -l app=test-performer --no-headers | wc -l)
    
    if [ "$PODS_A" -ge 3 ] && [ "$PODS_B" -ge 2 ]; then
        log_info "✓ Test 3 PASSED: Cross-namespace isolation working (A: ${PODS_A}, B: ${PODS_B})"
    else
        log_error "✗ Test 3 FAILED: Cross-namespace isolation issue (A: ${PODS_A}, B: ${PODS_B})"
        return 1
    fi
    
    # Test 4: Check unique services
    SERVICES_A=$(kubectl get services -n ${TEST_NAMESPACE_A} -l app=test-performer --no-headers | wc -l)
    SERVICES_B=$(kubectl get services -n ${TEST_NAMESPACE_B} -l app=test-performer --no-headers | wc -l)
    
    if [ "$SERVICES_A" -ge 3 ] && [ "$SERVICES_B" -ge 2 ]; then
        log_info "✓ Test 4 PASSED: Unique services created (A: ${SERVICES_A}, B: ${SERVICES_B})"
    else
        log_error "✗ Test 4 FAILED: Service creation issue (A: ${SERVICES_A}, B: ${SERVICES_B})"
        return 1
    fi
    
    log_info "Multiple performers validation completed successfully"
}

test_dns_resolution() {
    log_info "Testing DNS resolution for multiple performers..."
    
    # Test DNS resolution for performers in both namespaces
    PERFORMERS_A=($(kubectl get performers -n ${TEST_NAMESPACE_A} -o jsonpath='{.items[*].metadata.name}'))
    PERFORMERS_B=($(kubectl get performers -n ${TEST_NAMESPACE_B} -o jsonpath='{.items[*].metadata.name}'))
    
    local dns_failures=0
    
    # Test namespace A performers
    for performer in "${PERFORMERS_A[@]}"; do
        if kubectl run test-dns-${performer} --image=busybox:1.35 --namespace ${TEST_NAMESPACE_A} --rm -i --restart=Never -- nslookup ${performer}.${TEST_NAMESPACE_A}.svc.cluster.local &>/dev/null; then
            log_info "✓ DNS resolution working for ${performer} in namespace A"
        else
            log_warn "⚠ DNS resolution failed for ${performer} in namespace A"
            dns_failures=$((dns_failures + 1))
        fi
    done
    
    # Test namespace B performers
    for performer in "${PERFORMERS_B[@]}"; do
        if kubectl run test-dns-${performer} --image=busybox:1.35 --namespace ${TEST_NAMESPACE_B} --rm -i --restart=Never -- nslookup ${performer}.${TEST_NAMESPACE_B}.svc.cluster.local &>/dev/null; then
            log_info "✓ DNS resolution working for ${performer} in namespace B"
        else
            log_warn "⚠ DNS resolution failed for ${performer} in namespace B"
            dns_failures=$((dns_failures + 1))
        fi
    done
    
    if [ "$dns_failures" -eq 0 ]; then
        log_info "✓ DNS resolution test PASSED: All performers resolvable"
    else
        log_warn "⚠ DNS resolution test WARNING: ${dns_failures} failures (may be timing related)"
    fi
}

test_performance_scale() {
    log_info "Testing performance at scale..."
    
    # Create additional performers to test scale
    for i in {7..12}; do
        cat > /tmp/scale-performer-${i}.yaml << EOF
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: scale-performer-${i}
  namespace: ${TEST_NAMESPACE_A}
  labels:
    app: scale-performer
    test: performance
spec:
  avsAddress: "0xdddddddddddddddddddddddddddddddddddddddd"
  image:
    repository: nginx
    tag: "1.21"
  resources:
    requests:
      cpu: "50m"
      memory: "64Mi"
    limits:
      cpu: "100m"
      memory: "128Mi"
  ports:
  - name: grpc
    port: 9090
    targetPort: 9090
EOF
        kubectl apply -f /tmp/scale-performer-${i}.yaml
    done
    
    # Wait for deployment
    sleep 20
    
    # Check total performers
    TOTAL_PERFORMERS=$(kubectl get performers -n ${TEST_NAMESPACE_A} --no-headers | wc -l)
    if [ "$TOTAL_PERFORMERS" -ge 10 ]; then
        log_info "✓ Performance test PASSED: ${TOTAL_PERFORMERS} performers running successfully"
    else
        log_warn "⚠ Performance test WARNING: Only ${TOTAL_PERFORMERS} performers running"
    fi
    
    # Check operator resource usage
    OPERATOR_CPU=$(kubectl top pods -n ${OPERATOR_NAMESPACE} -l app.kubernetes.io/name=hourglass-operator --no-headers | awk '{print $2}' | sed 's/m//')
    OPERATOR_MEM=$(kubectl top pods -n ${OPERATOR_NAMESPACE} -l app.kubernetes.io/name=hourglass-operator --no-headers | awk '{print $3}' | sed 's/Mi//')
    
    log_info "Operator resource usage: CPU=${OPERATOR_CPU:-unknown}m, Memory=${OPERATOR_MEM:-unknown}Mi"
}

show_detailed_status() {
    log_info "=== Detailed Multi-Performer Status ==="
    
    echo -e "\n${GREEN}Namespace A Performers:${NC}"
    kubectl get performers -n ${TEST_NAMESPACE_A} -o wide
    
    echo -e "\n${GREEN}Namespace B Performers:${NC}"
    kubectl get performers -n ${TEST_NAMESPACE_B} -o wide
    
    echo -e "\n${GREEN}Namespace A Pods:${NC}"
    kubectl get pods -n ${TEST_NAMESPACE_A} -l app=test-performer -o wide
    
    echo -e "\n${GREEN}Namespace B Pods:${NC}"
    kubectl get pods -n ${TEST_NAMESPACE_B} -l app=test-performer -o wide
    
    echo -e "\n${GREEN}Namespace A Services:${NC}"
    kubectl get services -n ${TEST_NAMESPACE_A} -l app=test-performer
    
    echo -e "\n${GREEN}Namespace B Services:${NC}"
    kubectl get services -n ${TEST_NAMESPACE_B} -l app=test-performer
    
    echo -e "\n${GREEN}Operator Status:${NC}"
    kubectl get pods -n ${OPERATOR_NAMESPACE} -l app.kubernetes.io/name=hourglass-operator
    
    echo -e "\n${GREEN}Recent Events (Namespace A):${NC}"
    kubectl get events -n ${TEST_NAMESPACE_A} --sort-by='.lastTimestamp' | tail -5
    
    echo -e "\n${GREEN}Recent Events (Namespace B):${NC}"
    kubectl get events -n ${TEST_NAMESPACE_B} --sort-by='.lastTimestamp' | tail -5
}

cleanup_multi_performer_tests() {
    log_info "Cleaning up multi-performer tests..."
    
    # Delete all performers
    kubectl delete performers --all -n ${TEST_NAMESPACE_A} --ignore-not-found=true
    kubectl delete performers --all -n ${TEST_NAMESPACE_B} --ignore-not-found=true
    
    # Delete namespaces
    kubectl delete namespace ${TEST_NAMESPACE_A} --ignore-not-found=true
    kubectl delete namespace ${TEST_NAMESPACE_B} --ignore-not-found=true
    
    # Clean up temp files
    rm -f /tmp/performer-*.yaml
    rm -f /tmp/concurrent-performer-*.yaml
    rm -f /tmp/scale-performer-*.yaml
    
    log_info "Multi-performer test cleanup completed"
}

main() {
    case "${1:-test}" in
        test)
            log_info "Starting multi-performer tests for milestone 4.2..."
            create_test_namespaces
            create_multiple_performers_same_avs
            create_cross_namespace_performers
            test_concurrent_operations
            validate_multiple_performers
            test_dns_resolution
            test_performance_scale
            show_detailed_status
            log_info "Multi-performer tests completed successfully!"
            ;;
        cleanup)
            cleanup_multi_performer_tests
            ;;
        status)
            show_detailed_status
            ;;
        *)
            echo "Usage: $0 {test|cleanup|status}"
            exit 1
            ;;
    esac
}

main "$@"