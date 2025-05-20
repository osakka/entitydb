#!/bin/bash
# Master script to run all entity relationship tests

# Set up test environment
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="/opt/entitydb/var/log"
LOG_FILE="$LOG_DIR/entity_tests.log"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Make sure log directory exists
mkdir -p "$LOG_DIR"

# Initialize log file
echo "Entity Tests Run - $(date)" > "$LOG_FILE"
echo "=========================" >> "$LOG_FILE"

# Test counter
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test script
run_test_script() {
    local script_name="$1"
    local script_path="$SCRIPT_DIR/$script_name"
    
    echo -e "\n${YELLOW}Running test script: $script_name${NC}"
    echo -e "\nRunning test script: $script_name" >> "$LOG_FILE"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    # Check if the script exists and is executable
    if [[ ! -x "$script_path" ]]; then
        echo -e "${RED}✘ Script not found or not executable: $script_path${NC}"
        echo "✘ Script not found or not executable: $script_path" >> "$LOG_FILE"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    # Execute the script and capture output
    "$script_path" | tee -a "$LOG_FILE"
    local status=${PIPESTATUS[0]}
    
    if [[ "$status" -eq 0 ]]; then
        echo -e "${GREEN}✓ Test script passed${NC}"
        echo "✓ Test script passed" >> "$LOG_FILE"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✘ Test script failed with status $status${NC}"
        echo "✘ Test script failed with status $status" >> "$LOG_FILE"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    echo "-------------------------" >> "$LOG_FILE"
    return $status
}

# Run all entity test scripts
run_test_script "test_entity_relationship_api.sh"
run_test_script "test_entity_issue_conversion.sh"

# Print test summary
echo -e "\n${YELLOW}TEST SUMMARY${NC}"
echo -e "Total test scripts:  $TESTS_TOTAL"
echo -e "Passed:             ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed:             ${RED}$TESTS_FAILED${NC}"

# Write summary to log
echo -e "\nTEST SUMMARY" >> "$LOG_FILE"
echo "Total test scripts:  $TESTS_TOTAL" >> "$LOG_FILE"
echo "Passed:              $TESTS_PASSED" >> "$LOG_FILE"
echo "Failed:              $TESTS_FAILED" >> "$LOG_FILE"
echo "Test run completed at $(date)" >> "$LOG_FILE"

echo -e "\nTest log written to: $LOG_FILE"

# Return appropriate exit code
if [ "$TESTS_FAILED" -gt 0 ]; then
    exit 1
else
    exit 0
fi