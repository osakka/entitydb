#!/bin/bash
# Common utilities for EntityDB entity API tests
# To be sourced by test scripts

# Server configuration
SERVER_HOST="localhost"
SERVER_PORT="8085"
BASE_URL="http://${SERVER_HOST}:${SERVER_PORT}/api/v1"

# Output colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
PASSED=0
FAILED=0
TOTAL=0

# Current admin credentials
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="password"

# Set up test environment
setup_test() {
    local test_name="$1"
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Test suite: ${test_name}${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo "Base URL: $BASE_URL"
    echo "Date: $(date)"
    echo ""
    
    # Reset counters
    PASSED=0
    FAILED=0
    TOTAL=0
}

# Helper function to make API calls
call_api() {
    local method="$1"
    local endpoint="$2"
    shift 2
    
    # Build the full URL
    url="${BASE_URL}${endpoint}"
    
    # Make the request, passing through any additional arguments
    response=$(curl -s -X "$method" "$url" -w "\nSTATUS:%{http_code}" "$@")
    
    echo "$response"
}

# Login as admin and get token
login_admin() {
    response=$(curl -s -X POST "$BASE_URL/auth/login" \
      -H "Content-Type: application/json" \
      -d "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}")
    
    token=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$token" ]; then
        token=$(echo "$response" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    fi
    
    if [ -z "$token" ]; then
        echo "Failed to login as admin" >&2
        return 1
    else
        echo "$token"
        return 0
    fi
}

# Mark a test as passed
pass() {
    message="$1"
    PASSED=$((PASSED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "${GREEN}✓ PASS:${NC} $message"
}

# Mark a test as failed
fail() {
    message="$1"
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "${RED}✗ FAIL:${NC} $message"
}

# Log info message
log_info() {
    message="$1"
    echo -e "${BLUE}INFO:${NC} $message"
}

# Log step
log_step() {
    message="$1"
    echo -e "${YELLOW}STEP:${NC} $message"
}

# Assert HTTP status code
assert_status() {
    response="$1"
    expected_status="$2"
    message="${3:-Expected status $expected_status}"
    
    status=$(echo "$response" | grep -o 'STATUS:[0-9]*' | cut -d':' -f2)
    content=$(echo "$response" | sed '/STATUS:[0-9]*/d')
    
    if [ "$status" = "$expected_status" ]; then
        pass "Response status code is $status as expected"
    else
        fail "$message - Got status $status instead of $expected_status: $content"
    fi
}

# Assert JSON value
assert_json() {
    response="$1"
    jq_path="$2"
    expected_value="$3"
    message="${4:-Expected $jq_path to be $expected_value}"
    
    content=$(echo "$response" | sed '/STATUS:[0-9]*/d')
    actual_value=$(echo "$content" | jq -r "$jq_path")
    
    if [ "$actual_value" = "$expected_value" ]; then
        pass "$message"
    else
        fail "$message - Got '$actual_value' instead of '$expected_value'"
    fi
}

# Finish test and report summary
finish_test() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Test summary:${NC}"
    echo -e "${BLUE}  Total:  ${TOTAL}${NC}"
    echo -e "${GREEN}  Passed: ${PASSED}${NC}"
    
    if [ $FAILED -gt 0 ]; then
        echo -e "${RED}  Failed: ${FAILED}${NC}"
        echo -e "${RED}Test suite failed${NC}"
        exit 1
    else
        echo -e "${GREEN}All tests passed${NC}"
        exit 0
    fi
}

# Summarize test results
summarize_tests() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Test summary:${NC}"
    echo -e "${BLUE}  Total:  ${TOTAL}${NC}"
    echo -e "${GREEN}  Passed: ${PASSED}${NC}"
    
    if [ $FAILED -gt 0 ]; then
        echo -e "${RED}  Failed: ${FAILED}${NC}"
        echo -e "${RED}Test suite failed${NC}"
        exit 1
    else
        echo -e "${GREEN}All tests passed${NC}"
        exit 0
    fi
}

# Helper function to create a test entity
create_test_entity() {
    local type="$1"
    local title="$2"
    local token="$3"
    
    response=$(call_api POST "/entities" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $token" \
      -d "{
        \"type\": \"$type\",
        \"title\": \"$title\",
        \"tags\": {
          \"status\": \"active\"
        }
      }")
    
    echo "$response"
}

# Helper function to create a relationship between entities
create_test_relationship() {
    local source_id="$1"
    local rel_type="$2"
    local target_id="$3"
    local token="$4"
    
    response=$(call_api POST "/entity-relationships" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $token" \
      -d "{
        \"source_id\": \"$source_id\",
        \"relationship_type\": \"$rel_type\",
        \"target_id\": \"$target_id\"
      }")
    
    echo "$response"
}