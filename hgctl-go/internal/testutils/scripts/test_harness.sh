#!/usr/bin/env bash
# Test harness script for running integration tests

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$SCRIPT_DIR/../../.."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default test pattern
TEST_PATTERN="${1:-all}"

echo -e "${GREEN}🧪 Running hgctl integration tests${NC}"

# Ensure binary is built
echo -e "${YELLOW}Building hgctl binary...${NC}"
cd "$PROJECT_ROOT"
make build

# Check if chains are already running
if lsof -i :8545 >/dev/null 2>&1 || lsof -i :9545 >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Chains already running. Using existing chains.${NC}"
    SKIP_SETUP=true
else
    echo -e "${GREEN}✓ No existing chains detected${NC}"
    SKIP_SETUP=false
fi

# Run tests based on pattern
case "$TEST_PATTERN" in
    "all")
        echo -e "${GREEN}Running all integration tests...${NC}"
        # Don't use -skip-setup flag, let tests handle chain reuse
        # TODO: add |TestAggregatorDeployment
        go test -v ./internal/testutils/integration/... -run "TestOperatorRegistration|TestAdminLifecycle|TestAppointeeLifecycle"
        ;;
    "quick")
        echo -e "${GREEN}Running quick tests (no chain setup)...${NC}"
        go test -v ./internal/testutils/integration/... -skip-setup -run "TestBasic|TestCLI|TestHarness"
        ;;
    "keystore")
        echo -e "${GREEN}Running keystore tests...${NC}"
        go test -v ./internal/testutils/integration/... -run TestKeystore
        ;;
    "operator")
        echo -e "${GREEN}Running operator tests...${NC}"
        go test -v ./internal/testutils/integration/... -run TestOperator
        ;;
    "avs")
        echo -e "${GREEN}Running AVS tests...${NC}"
        go test -v ./internal/testutils/integration/... -run "TestAVS|TestKey"
        ;;
    "e2e")
        echo -e "${GREEN}Running end-to-end tests...${NC}"
        go test -v ./internal/testutils/integration/... -run TestComplete
        ;;
    *)
        echo -e "${GREEN}Running tests matching pattern: $TEST_PATTERN${NC}"
        go test -v ./internal/testutils/integration/... -run "$TEST_PATTERN"
        ;;
esac

# Capture exit code
EXIT_CODE=$?

# Show summary
if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ Tests passed!${NC}"
else
    echo -e "${RED}❌ Tests failed!${NC}"
fi

exit $EXIT_CODE