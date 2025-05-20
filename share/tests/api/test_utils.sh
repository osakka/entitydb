#!/bin/bash
# Common utilities for EntityDB test scripts
# To be sourced by test scripts

# Server configuration
SERVER_HOST="localhost"
SERVER_PORT="8085"
BASE_URL="http://${SERVER_HOST}:${SERVER_PORT}/api/v1"

# Output colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
PASSED=0
FAILED=0
TOTAL=0

# Current admin credentials (updated after the recent schema fixes)
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="password"

# Helper function to get token
get_auth_token() {
    local username="$1"
    local password="$2"
    
    echo "Attempting to log in with username: $username" >&2
    
    if [[ "$username" == test-user-* ]]; then
        # For test users, use a different login endpoint
        response=$(curl -s -X POST -H "Content-Type: application/json" \
            -d "{\"username\":\"$username\",\"password\":\"$password\"}" \
            "http://${SERVER_HOST}:${SERVER_PORT}/auth/login")
    else
        response=$(curl -s -X POST -H "Content-Type: application/json" \
            -d "{\"username\":\"$username\",\"password\":\"$password\"}" \
            "$BASE_URL/auth/login")
    fi
    
    echo "Auth response: $response" >&2
    
    # Try different formats for token extraction
    token=""
    
    # Look for "access_token" field first (standard format)
    token=$(echo "$response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    
    # If not found, try "token" field
    if [ -z "$token" ]; then
        token=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    fi
    
    # If still not found, try for "jwt" field
    if [ -z "$token" ]; then
        token=$(echo "$response" | grep -o '"jwt":"[^"]*"' | cut -d'"' -f4)
    fi
    
    # Finally try for first JWT pattern in the response
    if [ -z "$token" ]; then
        token=$(echo "$response" | grep -o 'eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*' | head -1)
    fi
    
    # For debugging
    if [ -z "$token" ]; then
        echo "Failed to extract token from response" >&2
    else
        echo "Token extracted successfully" >&2
    fi
    
    echo "$token"
}

# Helper function to extract ID from response
extract_id() {
    local response="$1"
    local id=""
    
    # Try different ID formats
    if echo "$response" | grep -q '"id":[0-9]*'; then
        id=$(echo "$response" | grep -o '"id":[0-9]*' | cut -d':' -f2)
    elif echo "$response" | grep -q '"id":"[0-9]*"'; then
        id=$(echo "$response" | grep -o '"id":"[0-9]*"' | cut -d'"' -f4)
    fi
    
    echo "$id"
}

# Helper function to register a test agent
register_test_agent() {
    local handle="$1"
    local name="$2"
    local specialization="$3"
    local auth_token="$4"
    
    response=$(test_endpoint "/agents/create" "POST" "{\"handle\":\"$handle\",\"name\":\"$name\",\"specialization\":\"$specialization\"}" 201 "Registering test agent" "$auth_token")
    
    echo "$response"
}

# Helper function to create a test issue
create_test_issue() {
    local title="$1"
    local description="$2"
    local priority="$3"
    local auth_token="$4"
    local issue_type="${5:-task}" # Default to task type if not specified
    local workspace_id="${6:-system}" # Default to system workspace if not specified
    
    response=$(test_endpoint "/issues/create" "POST" "{\"title\":\"$title\",\"description\":\"$description\",\"priority\":\"$priority\",\"type\":\"$issue_type\",\"workspace_id\":\"$workspace_id\"}" 201 "Creating test issue" "$auth_token")
    
    echo "$response"
}

# Helper function to create a test session
create_test_session() {
    local agent="$1"
    local workspace="$2"
    local name="$3"
    local description="$4"
    local auth_token="$5"
    
    response=$(test_endpoint "/sessions/create" "POST" "{\"agent\":\"$agent\",\"workspace\":\"$workspace\",\"name\":\"$name\",\"description\":\"$description\"}" 201 "Creating test session" "$auth_token")
    
    echo "$response"
}

# Function to test API endpoints
test_endpoint() {
    local endpoint=$1
    local method=${2:-GET}
    local data=${3:-""}
    local expected_status=${4:-200}
    local description=${5:-"Testing $endpoint"}
    local auth_token=${6:-""}
    
    echo -e "${YELLOW}$description${NC}"
    
    # Build command based on method and data
    cmd="curl -s -X $method -w '\n%{http_code}' -H 'Content-Type: application/json'"
    
    if [ -n "$auth_token" ]; then
        cmd="$cmd -H 'Authorization: Bearer $auth_token'"
    fi
    
    if [ -n "$data" ]; then
        cmd="$cmd -d '$data'"
    fi
    
    # Special handling for RBAC endpoints for test compatibility
    if [[ "$endpoint" == "/rbac/"* || "$endpoint" == "/user/"* ]]; then
        # Use direct endpoint without adding api/v1 prefix
        cmd="$cmd http://${SERVER_HOST}:${SERVER_PORT}$endpoint"
    elif [[ "$endpoint" == "/api/v1/rbac/"* ]]; then
        # Strip the duplicate /api/v1 prefix if it's already in the endpoint
        fixed_endpoint="${endpoint#/api/v1}"
        cmd="$cmd http://${SERVER_HOST}:${SERVER_PORT}$fixed_endpoint"
    else
        cmd="$cmd $BASE_URL$endpoint"
    fi
    
    # Execute the command
    result=$(eval $cmd)
    
    # Extract HTTP status code
    status_code=$(echo "$result" | tail -n 1)
    response_body=$(echo "$result" | head -n -1)
    
    # Print request details
    if [[ "$endpoint" == "/rbac/"* || "$endpoint" == "/user/"* ]]; then
        echo -e "  Request: $method http://${SERVER_HOST}:${SERVER_PORT}$endpoint"
    else
        echo -e "  Request: $method $BASE_URL$endpoint"
    fi
    if [ -n "$data" ]; then
        echo -e "  Data: $data"
    fi
    if [ -n "$auth_token" ]; then
        token_preview="${auth_token:0:10}...${auth_token: -10}"
        echo -e "  Auth: Bearer $token_preview"
    fi
    
    # Check if status code is as expected
    if [ "$status_code" -eq "$expected_status" ]; then
        echo -e "  ${GREEN}✓ Status: $status_code (as expected)${NC}"
        PASSED=$((PASSED+1))
    else
        echo -e "  ${RED}✗ Status: $status_code (expected $expected_status)${NC}"
        FAILED=$((FAILED+1))
    fi
    
    # Print response summary
    echo -e "  Response: $(echo "$response_body" | tr -d '\n' | head -c 100)..."
    echo
    
    TOTAL=$((TOTAL+1))
    
    # Return response body for parsing
    echo "$response_body"
}

# Function to print test summary
print_summary() {
    echo -e "${YELLOW}===============================${NC}"
    echo -e "Tests completed: $TOTAL"
    echo -e "${GREEN}Passed: $PASSED${NC}"
    echo -e "${RED}Failed: $FAILED${NC}"
    
    # Return status code based on test results
    if [ $FAILED -eq 0 ]; then
        echo -e "${GREEN}All tests passed successfully!${NC}"
        exit 0
    else
        echo -e "${RED}Some tests failed!${NC}"
        exit 1
    fi
}

# Function to clean up test resources
cleanup() {
    # Override in test scripts if needed
    echo "No cleanup required"
}

# Set up trap to ensure cleanup runs on exit
trap cleanup EXIT