#!/bin/bash

# Test script for verifying persistence functionality in the demo

echo "=== Ponos Persistence Test ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to check if services are ready
wait_for_services() {
    echo "Waiting for services to be ready..."
    local retries=30
    local count=0
    
    while [ $count -lt $retries ]; do
        if grpcurl -plaintext localhost:9090 list > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Services are ready${NC}"
            return 0
        fi
        echo -n "."
        sleep 2
        count=$((count+1))
    done
    
    echo -e "${RED}✗ Services failed to start${NC}"
    return 1
}

# Function to submit a task
submit_task() {
    local task_id=$1
    local payload=$2
    
    echo "Submitting task $task_id..."
    grpcurl -plaintext -d "{
        \"avsAddress\": \"0xCE2Ac75bE2E0951F1F7B288c7a6A9BfB6c331DC4\",
        \"taskId\": \"$task_id\",
        \"payload\": \"$payload\"
    }" localhost:9090 eigenlayer.hourglass.v1.ExecutorService/SubmitTask
}

# Main test flow
main() {
    echo "1. Starting services with docker compose..."
    docker compose up -d
    
    # Wait for services to be ready
    if ! wait_for_services; then
        echo "Failed to start services. Exiting."
        exit 1
    fi
    
    echo ""
    echo "2. Submitting test tasks..."
    
    # Submit a few tasks
    submit_task "0x0001" "eyAibnVtYmVyVG9CZVNxdWFyZWQiOiA0IH0="  # {"numberToBeSquared": 4}
    sleep 2
    submit_task "0x0002" "eyAibnVtYmVyVG9CZVNxdWFyZWQiOiA5IH0="  # {"numberToBeSquared": 9}
    sleep 2
    
    echo ""
    echo "3. Checking logs for persistence activity..."
    echo "Aggregator logs (last 20 lines):"
    docker compose logs aggregator | tail -20 | grep -E "(storage|persist|recover)" || echo "No persistence-related logs found"
    
    echo ""
    echo "Executor logs (last 20 lines):"
    docker compose logs executor | tail -20 | grep -E "(storage|persist|recover)" || echo "No persistence-related logs found"
    
    echo ""
    echo "4. Simulating crash and recovery..."
    echo "Stopping aggregator..."
    docker compose stop aggregator
    sleep 2
    
    echo "Restarting aggregator..."
    docker compose start aggregator
    sleep 5
    
    echo ""
    echo "5. Checking recovery logs..."
    docker compose logs aggregator | tail -30 | grep -E "(recover|restore|loaded)" || echo "No recovery logs found"
    
    echo ""
    echo "6. Submitting another task after recovery..."
    submit_task "0x0003" "eyAibnVtYmVyVG9CZVNxdWFyZWQiOiAxNiB9"  # {"numberToBeSquared": 16}
    
    echo ""
    echo -e "${GREEN}=== Test Complete ===${NC}"
    echo "Check the logs above to verify persistence is working correctly."
    echo ""
    echo "To clean up, run: docker compose down"
}

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    echo -e "${RED}grpcurl is not installed. Please install it with: brew install grpcurl${NC}"
    exit 1
fi

# Run the main test
main