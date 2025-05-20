#!/bin/bash
# EntityDB Test Suite Runner
# Runs all test scripts and collects results

# Set color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Store script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Initialize test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run tests in a directory
run_tests_in_dir() {
    local dir=$1
    local dir_name=$(basename "$dir")
    
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Running $dir_name tests${NC}"
    echo -e "${BLUE}========================================${NC}"
    
    # Run each test script in the directory
    for test_script in "$dir"/*.sh; do
        if [ -f "$test_script" ] && [ -x "$test_script" ]; then
            test_name=$(basename "$test_script")
            echo -e "${YELLOW}Running test: $test_name${NC}"
            
            # Run the test script and capture its exit code
            "$test_script"
            exit_code=$?
            
            TOTAL_TESTS=$((TOTAL_TESTS+1))
            
            if [ $exit_code -eq 0 ]; then
                echo -e "${GREEN}Test passed: $test_name${NC}"
                PASSED_TESTS=$((PASSED_TESTS+1))
            else
                echo -e "${RED}Test failed: $test_name${NC}"
                FAILED_TESTS=$((FAILED_TESTS+1))
            fi
            
            echo -e "${YELLOW}----------------------------------------${NC}"
        fi
    done
}

# Print header
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB API Test Suite${NC}"
echo -e "${BLUE}========================================${NC}"

# Make sure server is running
if ! curl -s http://localhost:8085/api/v1/health > /dev/null; then
    echo -e "${RED}EntityDB server is not running! Please start the server before running tests.${NC}"
    echo -e "${YELLOW}Try running: ./bin/entitydbd.sh start${NC}"
    exit 1
fi

# Run tests in each subdirectory
run_tests_in_dir "$SCRIPT_DIR/auth"
run_tests_in_dir "$SCRIPT_DIR/agent"
run_tests_in_dir "$SCRIPT_DIR/session"
run_tests_in_dir "$SCRIPT_DIR/issue"
run_tests_in_dir "$SCRIPT_DIR/workspace"
run_tests_in_dir "$SCRIPT_DIR/rbac"

# Print summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Suite Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Total test scripts: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

# Return exit code based on test results
if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}All tests passed successfully!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi