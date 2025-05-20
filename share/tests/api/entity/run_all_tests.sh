#!/bin/bash
# Run all entity API tests

# Set color codes
BLUE='\033[0;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Test statistics
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Running Entity API Tests${NC}"
echo -e "${BLUE}========================================${NC}"

# Run each test file
for test in "$SCRIPT_DIR"/test_*.sh; do
    if [ -x "$test" ]; then
        echo -e "${YELLOW}Running $(basename "$test")...${NC}"
        "$test"
        result=$?
        
        TOTAL_TESTS=$((TOTAL_TESTS+1))
        
        if [ $result -eq 0 ]; then
            echo -e "${GREEN}Test passed: $(basename "$test")${NC}"
            PASSED_TESTS=$((PASSED_TESTS+1))
        else
            echo -e "${RED}Test failed: $(basename "$test")${NC}"
            FAILED_TESTS=$((FAILED_TESTS+1))
        fi
        
        echo -e "${YELLOW}----------------------------------------${NC}"
    fi
done

# Print summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Entity API Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Total tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

# Exit with error if any test failed
if [ $FAILED_TESTS -gt 0 ]; then
    exit 1
else
    exit 0
fi