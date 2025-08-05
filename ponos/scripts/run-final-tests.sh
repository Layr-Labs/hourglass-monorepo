#!/bin/bash

# Run final testing suite for Ponos persistence layer
# This script runs all milestone 7.4 tests

set -e

echo "=== Ponos Final Testing Suite ==="
echo "Running all production validation tests..."
echo

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
SHORT_MODE=${1:-"short"}  # Pass "full" for extended tests
RESULTS_DIR="test-results-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$RESULTS_DIR"

# Function to run tests and capture results
run_test_suite() {
    local suite_name=$1
    local test_path=$2
    local timeout=$3
    
    echo -e "${YELLOW}Running $suite_name tests...${NC}"
    
    if [ "$SHORT_MODE" == "short" ]; then
        timeout_cmd="timeout $timeout"
    else
        timeout_cmd=""
    fi
    
    if $timeout_cmd go test -v -short "$test_path" > "$RESULTS_DIR/$suite_name.log" 2>&1; then
        echo -e "${GREEN}✓ $suite_name tests passed${NC}"
        return 0
    else
        echo -e "${RED}✗ $suite_name tests failed${NC}"
        echo "  See $RESULTS_DIR/$suite_name.log for details"
        return 1
    fi
}

# Function to run benchmarks
run_benchmarks() {
    local bench_name=$1
    local test_path=$2
    
    echo -e "${YELLOW}Running $bench_name benchmarks...${NC}"
    
    if go test -bench=. -benchmem -benchtime=10s "$test_path" > "$RESULTS_DIR/$bench_name-bench.log" 2>&1; then
        echo -e "${GREEN}✓ $bench_name benchmarks completed${NC}"
        # Extract key metrics
        echo "  Key results:"
        grep -E "Benchmark.*ns/op" "$RESULTS_DIR/$bench_name-bench.log" | head -5
        return 0
    else
        echo -e "${RED}✗ $bench_name benchmarks failed${NC}"
        return 1
    fi
}

# Track overall results
FAILED_TESTS=0

# 1. Extended Stability Tests
if ! run_test_suite "stability" "./internal/tests/stability" "300"; then
    ((FAILED_TESTS++))
fi

# 2. Upgrade Scenario Tests
if ! run_test_suite "upgrade" "./internal/tests/upgrade" "120"; then
    ((FAILED_TESTS++))
fi

# 3. Production Configuration Tests
if ! run_test_suite "production" "./internal/tests/production" "180"; then
    ((FAILED_TESTS++))
fi

# 4. Performance Benchmarks
if ! run_benchmarks "performance" "./internal/tests/performance"; then
    ((FAILED_TESTS++))
fi

# 5. Integration Tests (existing)
echo -e "${YELLOW}Running integration tests...${NC}"
if ! run_test_suite "integration" "./internal/tests/persistence" "120"; then
    ((FAILED_TESTS++))
fi

# 6. Storage Implementation Tests
echo -e "${YELLOW}Running storage tests...${NC}"
if go test -v -short ./pkg/aggregator/storage/... ./pkg/executor/storage/... > "$RESULTS_DIR/storage.log" 2>&1; then
    echo -e "${GREEN}✓ Storage tests passed${NC}"
else
    echo -e "${RED}✗ Storage tests failed${NC}"
    ((FAILED_TESTS++))
fi

echo
echo "=== Test Summary ==="
echo "Results saved to: $RESULTS_DIR/"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    
    # Generate summary report
    cat > "$RESULTS_DIR/summary.txt" <<EOF
Ponos Final Test Summary
========================
Date: $(date)
Mode: $SHORT_MODE

Test Results:
✓ Stability tests: PASSED
✓ Upgrade tests: PASSED
✓ Production tests: PASSED
✓ Performance benchmarks: COMPLETED
✓ Integration tests: PASSED
✓ Storage tests: PASSED

Key Performance Metrics:
$(grep -h "Benchmark.*ns/op" "$RESULTS_DIR"/*.log | grep -E "(SaveTask|GetTask|UpdateTaskStatus)" | head -10)

Production Readiness: YES
EOF
    
    echo
    echo "Production readiness confirmed!"
    echo "See $RESULTS_DIR/summary.txt for full report"
    
else
    echo -e "${RED}$FAILED_TESTS test suites failed ✗${NC}"
    echo "Please review the logs in $RESULTS_DIR/"
    exit 1
fi

# Optional: Run extended tests if requested
if [ "$SHORT_MODE" == "full" ]; then
    echo
    echo -e "${YELLOW}Running extended stability tests...${NC}"
    echo "Note: Extended tests must be explicitly enabled with RUN_EXTENDED_TESTS=true"
    
    # Run extended test with environment variable
    RUN_EXTENDED_TESTS=true go test -v -run TestExtendedStability ./internal/tests/stability -timeout 25h > "$RESULTS_DIR/extended-stability.log" 2>&1 &
    
    echo "Extended test running in background (PID: $!)"
    echo "Check $RESULTS_DIR/extended-stability.log for progress"
    echo
    echo "To run stability tests manually:"
    echo "  Short (5 min): go test -v -run TestShortStability ./internal/tests/stability"
    echo "  Extended (1-24h): RUN_EXTENDED_TESTS=true go test -v -run TestExtendedStability ./internal/tests/stability"
fi