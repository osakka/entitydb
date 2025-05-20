#!/bin/bash
# Improved wrapper to run API tests from the /opt/entitydb/share/tests/api directory

# Set color codes
BLUE='\033[0;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Testing flags
SKIP_SERVER_CHECK=0
SPECIFIC_TEST=""
VERBOSE=0

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-server-check)
            SKIP_SERVER_CHECK=1
            shift
            ;;
        --test)
            SPECIFIC_TEST="$2"
            shift 2
            ;;
        --verbose|-v)
            VERBOSE=1
            shift
            ;;
        --help|-h)
            echo -e "${BLUE}EntityDB API Test Runner${NC}"
            echo -e "${BLUE}===================${NC}"
            echo "Usage: $(basename $0) [options]"
            echo ""
            echo "Options:"
            echo "  --no-server-check     Skip server availability check"
            echo "  --test <test_name>    Run a specific test or test directory"
            echo "  --verbose, -v         Show more detailed output"
            echo "  --help, -h            Show this help message"
            echo ""
            echo "Examples:"
            echo "  $(basename $0) --test rbac/test_rbac_permissions.sh"
            echo "  $(basename $0) --test rbac"
            echo "  $(basename $0) --verbose"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

if [ $VERBOSE -eq 1 ]; then
    echo -e "${YELLOW}Running in verbose mode${NC}"
fi

# Server health check
if [ $SKIP_SERVER_CHECK -eq 0 ]; then
    echo -e "${YELLOW}Checking if EntityDB server is running...${NC}"
    if ! curl -s http://localhost:8085/api/v1/status > /dev/null; then
        echo -e "${RED}EntityDB server is not running! Please start the server before running tests.${NC}"
        echo -e "${YELLOW}Try running: /opt/entitydb/bin/entitydbd.sh start${NC}"
        exit 1
    else
        echo -e "${GREEN}Server is running and responsive.${NC}"
    fi
else
    echo -e "${YELLOW}Skipping server check as requested.${NC}"
fi

# Set the base test directory
TEST_BASE_DIR="/opt/entitydb/share/tests/api"

if [ ! -d "$TEST_BASE_DIR" ]; then
    echo -e "${RED}Test directory not found: $TEST_BASE_DIR${NC}"
    exit 1
fi

# Function to run all tests
run_all_tests() {
    echo -e "${BLUE}Running all API tests from $TEST_BASE_DIR${NC}"
    bash "$TEST_BASE_DIR/run_all_tests.sh"
    return $?
}

# Function to run a specific test
run_specific_test() {
    local test_path="$TEST_BASE_DIR/$SPECIFIC_TEST"
    
    # Check if it's a directory or file
    if [ -d "$test_path" ]; then
        # It's a directory, look for a run_all_tests.sh script
        if [ -f "$test_path/run_all_tests.sh" ]; then
            echo -e "${BLUE}Running all tests in directory: $test_path${NC}"
            cd "$test_path" && bash "./run_all_tests.sh"
            return $?
        else
            # Run each test script in the directory
            echo -e "${BLUE}Running all test scripts in directory: $test_path${NC}"
            local test_count=0
            local pass_count=0
            local fail_count=0
            
            for test_script in "$test_path"/*.sh; do
                if [ -x "$test_script" ]; then
                    echo -e "${YELLOW}Running test: $(basename $test_script)${NC}"
                    bash "$test_script"
                    if [ $? -eq 0 ]; then
                        echo -e "${GREEN}Test passed: $(basename $test_script)${NC}"
                        pass_count=$((pass_count+1))
                    else
                        echo -e "${RED}Test failed: $(basename $test_script)${NC}"
                        fail_count=$((fail_count+1))
                    fi
                    test_count=$((test_count+1))
                fi
            done
            
            echo -e "${BLUE}===================${NC}"
            echo -e "${BLUE}Test Summary${NC}"
            echo -e "${BLUE}===================${NC}"
            echo -e "Total tests: $test_count"
            echo -e "${GREEN}Passed: $pass_count${NC}"
            echo -e "${RED}Failed: $fail_count${NC}"
            
            if [ $fail_count -gt 0 ]; then
                return 1
            else
                return 0
            fi
        fi
    elif [ -f "$test_path" ]; then
        # It's a single test file
        echo -e "${BLUE}Running specific test: $test_path${NC}"
        bash "$test_path"
        return $?
    else
        echo -e "${RED}Test not found: $test_path${NC}"
        return 1
    fi
}

# Run tests based on command line arguments
if [ -n "$SPECIFIC_TEST" ]; then
    if [ $VERBOSE -eq 1 ]; then
        echo -e "${YELLOW}Running specific test: $SPECIFIC_TEST${NC}"
    fi
    run_specific_test
    TEST_RESULT=$?
else
    if [ $VERBOSE -eq 1 ]; then
        echo -e "${YELLOW}Running all tests${NC}"
    fi
    run_all_tests
    TEST_RESULT=$?
fi

# Show result
if [ $TEST_RESULT -eq 0 ]; then
    echo -e "${GREEN}All API tests completed successfully!${NC}"
else
    echo -e "${RED}Some API tests failed!${NC}"
fi

exit $TEST_RESULT