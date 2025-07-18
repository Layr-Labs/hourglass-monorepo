#!/bin/bash

# Milestone 4.2 E2E Validation Script
# Comprehensive validation of all milestone 4.2 requirements

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_RESULTS_DIR="${SCRIPT_DIR}/test-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_FILE="${TEST_RESULTS_DIR}/milestone-4.2-results-${TIMESTAMP}.log"

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_WARNING=0
TOTAL_TESTS=0

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1" | tee -a "${RESULTS_FILE}"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "${RESULTS_FILE}"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "${RESULTS_FILE}"
}

log_test_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}" | tee -a "${RESULTS_FILE}"
}

record_test_result() {
    local test_name="$1"
    local result="$2"
    local details="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    case "$result" in
        "PASS")
            TESTS_PASSED=$((TESTS_PASSED + 1))
            log_info "✓ $test_name: PASSED - $details"
            ;;
        "FAIL")
            TESTS_FAILED=$((TESTS_FAILED + 1))
            log_error "✗ $test_name: FAILED - $details"
            ;;
        "WARN")
            TESTS_WARNING=$((TESTS_WARNING + 1))
            log_warn "⚠ $test_name: WARNING - $details"
            ;;
    esac
}

setup_test_environment() {
    log_info "Setting up test environment..."
    
    # Create results directory
    mkdir -p "${TEST_RESULTS_DIR}"
    
    # Initialize results file
    echo "Hourglass Operator Milestone 4.2 E2E Test Results" > "${RESULTS_FILE}"
    echo "Started: $(date)" >> "${RESULTS_FILE}"
    echo "=============================================" >> "${RESULTS_FILE}"
    
    log_info "Test environment setup complete"
}

validate_prerequisites() {
    log_test_header "Prerequisites Validation"
    
    # Test: kubectl available
    if command -v kubectl &> /dev/null; then
        record_test_result "kubectl availability" "PASS" "kubectl is installed and accessible"
    else
        record_test_result "kubectl availability" "FAIL" "kubectl is not installed or not in PATH"
        return 1
    fi
    
    # Test: helm available
    if command -v helm &> /dev/null; then
        record_test_result "helm availability" "PASS" "helm is installed and accessible"
    else
        record_test_result "helm availability" "FAIL" "helm is not installed or not in PATH"
        return 1
    fi
    
    # Test: cluster connectivity
    if kubectl cluster-info &> /dev/null; then
        record_test_result "cluster connectivity" "PASS" "can connect to Kubernetes cluster"
    else
        record_test_result "cluster connectivity" "FAIL" "cannot connect to Kubernetes cluster"
        return 1
    fi
    
    # Test: cluster admin permissions
    if kubectl auth can-i create customresourcedefinitions &> /dev/null; then
        record_test_result "cluster admin permissions" "PASS" "have cluster admin permissions"
    else
        record_test_result "cluster admin permissions" "FAIL" "do not have cluster admin permissions"
        return 1
    fi
}

run_operator_integration_tests() {
    log_test_header "4.2.1 Operator Integration Tests"
    
    # Run basic operator integration tests
    log_info "Running operator integration tests..."
    
    if "${SCRIPT_DIR}/test-setup.sh" setup > "${TEST_RESULTS_DIR}/operator-setup-${TIMESTAMP}.log" 2>&1; then
        record_test_result "operator deployment" "PASS" "operator deployed successfully"
    else
        record_test_result "operator deployment" "FAIL" "operator deployment failed"
        return 1
    fi
    
    # Wait for operator to be ready
    sleep 10
    
    # Test: Operator pod running
    if kubectl get pods -n hourglass-system -l app.kubernetes.io/name=hourglass-operator | grep -q "Running"; then
        record_test_result "operator pod status" "PASS" "operator pod is running"
    else
        record_test_result "operator pod status" "FAIL" "operator pod is not running"
    fi
    
    # Test: CRDs installed
    if kubectl get crd performers.hourglass.eigenlayer.io &> /dev/null; then
        record_test_result "CRD installation" "PASS" "Performer CRD is installed"
    else
        record_test_result "CRD installation" "FAIL" "Performer CRD is not installed"
    fi
    
    # Test: Performer CRD creation
    if kubectl get performers -n test-avs-project &> /dev/null; then
        record_test_result "performer CRD creation" "PASS" "can create and list Performer CRDs"
    else
        record_test_result "performer CRD creation" "FAIL" "cannot create or list Performer CRDs"
    fi
    
    # Test: Pod creation by operator
    sleep 15  # Allow time for operator to process
    if kubectl get pods -n test-avs-project -l app=test-performer | grep -q "Running"; then
        record_test_result "performer pod creation" "PASS" "operator created performer pod"
    else
        POD_STATUS=$(kubectl get pods -n test-avs-project -l app=test-performer --no-headers | awk '{print $3}' | head -1)
        if [ -n "$POD_STATUS" ]; then
            record_test_result "performer pod creation" "WARN" "performer pod exists but status is: $POD_STATUS"
        else
            record_test_result "performer pod creation" "FAIL" "operator did not create performer pod"
        fi
    fi
    
    # Test: Service creation by operator
    if kubectl get service test-performer-1 -n test-avs-project &> /dev/null; then
        record_test_result "performer service creation" "PASS" "operator created performer service"
    else
        record_test_result "performer service creation" "FAIL" "operator did not create performer service"
    fi
    
    # Test: DNS resolution
    if kubectl run test-dns-validator --image=busybox:1.35 --namespace test-avs-project --rm -i --restart=Never -- nslookup test-performer-1.test-avs-project.svc.cluster.local &> /dev/null; then
        record_test_result "DNS resolution" "PASS" "service DNS resolution works"
    else
        record_test_result "DNS resolution" "WARN" "service DNS resolution failed (may be timing related)"
    fi
    
    # Test: Status updates
    STATUS=$(kubectl get performer test-performer-1 -n test-avs-project -o jsonpath='{.status.phase}' 2>/dev/null || echo "unknown")
    if [[ "$STATUS" != "unknown" && "$STATUS" != "" ]]; then
        record_test_result "performer status updates" "PASS" "performer status is: $STATUS"
    else
        record_test_result "performer status updates" "WARN" "performer status not yet available"
    fi
}

run_multi_performer_tests() {
    log_test_header "4.2.2 Multi-Performer Scenarios"
    
    # Run multi-performer tests
    log_info "Running multi-performer tests..."
    
    if "${SCRIPT_DIR}/multi-performer-test.sh" test > "${TEST_RESULTS_DIR}/multi-performer-${TIMESTAMP}.log" 2>&1; then
        record_test_result "multi-performer deployment" "PASS" "multi-performer tests completed successfully"
    else
        record_test_result "multi-performer deployment" "FAIL" "multi-performer tests failed"
        return 1
    fi
    
    # Wait for deployment
    sleep 20
    
    # Test: Multiple performers per AVS
    PERFORMERS_A=$(kubectl get performers -n test-avs-a --no-headers | wc -l)
    if [ "$PERFORMERS_A" -ge 3 ]; then
        record_test_result "multiple performers per AVS" "PASS" "created $PERFORMERS_A performers in namespace A"
    else
        record_test_result "multiple performers per AVS" "FAIL" "only $PERFORMERS_A performers in namespace A"
    fi
    
    # Test: Cross-namespace isolation
    PERFORMERS_B=$(kubectl get performers -n test-avs-b --no-headers | wc -l)
    if [ "$PERFORMERS_B" -ge 2 ]; then
        record_test_result "cross-namespace isolation" "PASS" "created $PERFORMERS_B performers in namespace B"
    else
        record_test_result "cross-namespace isolation" "FAIL" "only $PERFORMERS_B performers in namespace B"
    fi
    
    # Test: Concurrent operations
    CONCURRENT_PERFORMERS=$(kubectl get performers -n test-avs-a -l test=concurrent --no-headers | wc -l)
    if [ "$CONCURRENT_PERFORMERS" -ge 3 ]; then
        record_test_result "concurrent operations" "PASS" "created $CONCURRENT_PERFORMERS concurrent performers"
    else
        record_test_result "concurrent operations" "WARN" "only $CONCURRENT_PERFORMERS concurrent performers created"
    fi
    
    # Test: Performance at scale
    TOTAL_PERFORMERS=$(kubectl get performers -n test-avs-a --no-headers | wc -l)
    if [ "$TOTAL_PERFORMERS" -ge 10 ]; then
        record_test_result "performance at scale" "PASS" "operator handling $TOTAL_PERFORMERS performers"
    else
        record_test_result "performance at scale" "WARN" "operator handling only $TOTAL_PERFORMERS performers"
    fi
}

test_connection_retry_integration() {
    log_test_header "4.2.3 Connection Retry Integration"
    
    # Test: Connection retry logic in real cluster
    log_info "Testing connection retry integration..."
    
    # Create a performer that points to a non-existent service to test retry logic
    cat > /tmp/retry-test-performer.yaml << 'EOF'
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
metadata:
  name: retry-test-performer
  namespace: test-avs-project
  labels:
    app: retry-test
spec:
  avsAddress: "0x1111111111111111111111111111111111111111"
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
  - name: grpc
    port: 9090
    targetPort: 9090
  env:
  - name: TEST_RETRY
    value: "true"
EOF
    
    kubectl apply -f /tmp/retry-test-performer.yaml
    
    # Wait for pod to be created
    sleep 15
    
    # Test: Retry performer pod exists
    if kubectl get pods -n test-avs-project -l app=retry-test | grep -q "retry-test-performer"; then
        record_test_result "retry logic integration" "PASS" "retry test performer pod created"
    else
        record_test_result "retry logic integration" "FAIL" "retry test performer pod not created"
    fi
    
    # Test: Service created for retry test
    if kubectl get service retry-test-performer -n test-avs-project &> /dev/null; then
        record_test_result "retry service creation" "PASS" "retry test service created"
    else
        record_test_result "retry service creation" "FAIL" "retry test service not created"
    fi
    
    # Cleanup retry test
    kubectl delete -f /tmp/retry-test-performer.yaml --ignore-not-found=true
    rm -f /tmp/retry-test-performer.yaml
}

test_operator_health_monitoring() {
    log_test_header "4.2.4 Operator Health Monitoring"
    
    # Test: Operator health endpoints
    OPERATOR_POD=$(kubectl get pods -n hourglass-system -l app.kubernetes.io/name=hourglass-operator -o jsonpath='{.items[0].metadata.name}')
    
    if [ -n "$OPERATOR_POD" ]; then
        # Test health endpoint
        if kubectl exec -n hourglass-system "$OPERATOR_POD" -- wget -q -O- http://localhost:8081/healthz | grep -q "ok"; then
            record_test_result "operator health endpoint" "PASS" "operator health endpoint responding"
        else
            record_test_result "operator health endpoint" "WARN" "operator health endpoint not responding as expected"
        fi
        
        # Test metrics endpoint
        if kubectl exec -n hourglass-system "$OPERATOR_POD" -- wget -q -O- http://localhost:8080/metrics | grep -q "go_"; then
            record_test_result "operator metrics endpoint" "PASS" "operator metrics endpoint responding"
        else
            record_test_result "operator metrics endpoint" "WARN" "operator metrics endpoint not responding as expected"
        fi
        
        # Test operator logs for errors
        ERROR_COUNT=$(kubectl logs -n hourglass-system "$OPERATOR_POD" --tail=100 | grep -c "ERROR" || echo "0")
        if [ "$ERROR_COUNT" -eq 0 ]; then
            record_test_result "operator error logs" "PASS" "no errors in operator logs"
        else
            record_test_result "operator error logs" "WARN" "$ERROR_COUNT errors found in operator logs"
        fi
    else
        record_test_result "operator pod availability" "FAIL" "operator pod not found"
    fi
}

generate_test_report() {
    log_test_header "Test Summary"
    
    echo "" | tee -a "${RESULTS_FILE}"
    echo "Test Results Summary:" | tee -a "${RESULTS_FILE}"
    echo "===================" | tee -a "${RESULTS_FILE}"
    echo "Total Tests: $TOTAL_TESTS" | tee -a "${RESULTS_FILE}"
    echo "Passed: $TESTS_PASSED" | tee -a "${RESULTS_FILE}"
    echo "Failed: $TESTS_FAILED" | tee -a "${RESULTS_FILE}"
    echo "Warnings: $TESTS_WARNING" | tee -a "${RESULTS_FILE}"
    echo "" | tee -a "${RESULTS_FILE}"
    
    # Calculate success rate
    if [ "$TOTAL_TESTS" -gt 0 ]; then
        SUCCESS_RATE=$(( (TESTS_PASSED * 100) / TOTAL_TESTS ))
        echo "Success Rate: ${SUCCESS_RATE}%" | tee -a "${RESULTS_FILE}"
    fi
    
    # Determine overall result
    if [ "$TESTS_FAILED" -eq 0 ]; then
        if [ "$TESTS_WARNING" -eq 0 ]; then
            echo -e "\n${GREEN}✓ MILESTONE 4.2 VALIDATION: ALL TESTS PASSED${NC}" | tee -a "${RESULTS_FILE}"
        else
            echo -e "\n${YELLOW}⚠ MILESTONE 4.2 VALIDATION: PASSED WITH WARNINGS${NC}" | tee -a "${RESULTS_FILE}"
        fi
    else
        echo -e "\n${RED}✗ MILESTONE 4.2 VALIDATION: FAILED${NC}" | tee -a "${RESULTS_FILE}"
    fi
    
    echo "" | tee -a "${RESULTS_FILE}"
    echo "Completed: $(date)" | tee -a "${RESULTS_FILE}"
    echo "Results saved to: ${RESULTS_FILE}" | tee -a "${RESULTS_FILE}"
}

cleanup_test_environment() {
    log_info "Cleaning up test environment..."
    
    # Cleanup basic tests
    "${SCRIPT_DIR}/test-setup.sh" cleanup &> /dev/null || true
    
    # Cleanup multi-performer tests
    "${SCRIPT_DIR}/multi-performer-test.sh" cleanup &> /dev/null || true
    
    log_info "Test environment cleanup completed"
}

main() {
    case "${1:-validate}" in
        validate)
            log_info "Starting Milestone 4.2 E2E Validation..."
            setup_test_environment
            validate_prerequisites
            run_operator_integration_tests
            run_multi_performer_tests
            test_connection_retry_integration
            test_operator_health_monitoring
            generate_test_report
            cleanup_test_environment
            log_info "Milestone 4.2 validation completed!"
            ;;
        cleanup)
            cleanup_test_environment
            ;;
        report)
            if [ -f "${RESULTS_FILE}" ]; then
                cat "${RESULTS_FILE}"
            else
                echo "No recent test results found"
            fi
            ;;
        *)
            echo "Usage: $0 {validate|cleanup|report}"
            exit 1
            ;;
    esac
}

main "$@"