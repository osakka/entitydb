#!/bin/bash
# Integration test for entity-issue conversion with relationships

# Set up test environment
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test_utils.sh"

# Test variables
API_URL="http://localhost:8085"
AUTH_TOKEN=""
TEST_ISSUE_1="issue_test_$(date +%s)_1"
TEST_ISSUE_2="issue_test_$(date +%s)_2"
DEPENDENCY_ID=""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected_status="$3"
    local check_function="$4"
    
    echo -e "\n${YELLOW}Running test: $test_name${NC}"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    # Execute the command and capture both output and status code
    local output
    output=$(eval "$test_cmd" 2>&1)
    local status=$?
    
    # Check status code if expected_status is provided
    if [[ -n "$expected_status" && "$status" -ne "$expected_status" ]]; then
        echo -e "${RED}✘ Test failed: Expected status $expected_status, got $status${NC}"
        echo "Output: $output"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    # Run check function if provided
    if [[ -n "$check_function" ]]; then
        if ! $check_function "$output"; then
            echo -e "${RED}✘ Test failed: Check function returned false${NC}"
            echo "Output: $output"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    fi
    
    echo -e "${GREEN}✓ Test passed${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
}

# Function to check if dual write is enabled
check_dual_write_enabled() {
    local output="$1"
    if echo "$output" | grep -q '"dual_write_enabled":\s*true'; then
        return 0
    else
        echo "Dual write is not enabled, skipping test"
        TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
        TESTS_TOTAL=$((TESTS_TOTAL - 1))
        return 1
    fi
}

# Function to check if entity relationships feature is enabled
check_entity_relationships_enabled() {
    local output="$1"
    if echo "$output" | grep -q '"relationships_enabled":\s*true'; then
        return 0
    else
        echo "Entity relationships feature is not enabled, skipping test"
        TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
        TESTS_TOTAL=$((TESTS_TOTAL - 1))
        return 1
    fi
}

# Function to check if a response contains a specific string
check_response_contains() {
    local pattern="$1"
    local output="$2"
    if echo "$output" | grep -q "$pattern"; then
        return 0
    else
        echo "Response does not contain '$pattern'"
        return 1
    fi
}

# Function to extract ID from response
extract_id() {
    local output="$1"
    local id_pattern="$2"
    local id
    id=$(echo "$output" | grep -oP "$id_pattern")
    echo "$id"
}

# Login as admin to get auth token
echo "Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" -H "Content-Type: application/json" -d '{"username":"admin","password":"password"}')
AUTH_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -oP '"token":\s*"\K[^"]+')

if [[ -z "$AUTH_TOKEN" ]]; then
    echo "Failed to get auth token"
    exit 1
fi

# Check feature flags
echo "Checking if dual write and entity relationships features are enabled..."
FEATURES_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/features" -H "Authorization: Bearer $AUTH_TOKEN")

ENTITY_ENABLED=$(echo "$FEATURES_RESPONSE" | grep -oP '"entity":\s*{\s*"enabled":\s*\K(true|false)')
RELATIONSHIPS_ENABLED=$(echo "$FEATURES_RESPONSE" | grep -oP '"relationships_enabled":\s*\K(true|false)')
DUAL_WRITE_ENABLED=$(echo "$FEATURES_RESPONSE" | grep -oP '"dual_write_enabled":\s*\K(true|false)')

if [[ "$ENTITY_ENABLED" != "true" || "$RELATIONSHIPS_ENABLED" != "true" || "$DUAL_WRITE_ENABLED" != "true" ]]; then
    echo "Entity relationships and/or dual write features are not enabled"
    echo "Would you like to enable them for testing? (y/n)"
    read -r ENABLE_RESPONSE
    
    if [[ "$ENABLE_RESPONSE" == "y" ]]; then
        echo "Enabling entity relationships and dual write features..."
        curl -s -X PUT "$API_URL/api/v1/features" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -H "Content-Type: application/json" \
            -d '{
                "entity": {
                    "enabled": true,
                    "relationships_enabled": true,
                    "dual_write_enabled": true
                }
            }'
    else
        echo "Skipping tests since required features are not enabled"
        exit 0
    fi
fi

# Create test issues
echo "Creating test issues..."
ISSUE1_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/issues/create" \
    -H "Authorization: Bearer $AUTH_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"title\": \"Test Issue 1\",
        \"description\": \"Issue for testing entity relationships\",
        \"priority\": \"medium\",
        \"type\": \"issue\",
        \"workspace_id\": \"workspace_entitydb\"
    }")

ISSUE1_ID=$(echo "$ISSUE1_RESPONSE" | grep -oP '"id":\s*"\K[^"]+')

ISSUE2_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/issues/create" \
    -H "Authorization: Bearer $AUTH_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"title\": \"Test Issue 2\",
        \"description\": \"Issue for testing entity relationships\",
        \"priority\": \"medium\",
        \"type\": \"issue\",
        \"workspace_id\": \"workspace_entitydb\"
    }")

ISSUE2_ID=$(echo "$ISSUE2_RESPONSE" | grep -oP '"id":\s*"\K[^"]+')

if [[ -z "$ISSUE1_ID" || -z "$ISSUE2_ID" ]]; then
    echo "Failed to create test issues"
    exit 1
fi

echo "Created issues: $ISSUE1_ID and $ISSUE2_ID"

# Test 1: Create a dependency between issues
run_test "Create Issue Dependency" "
    curl -s -X POST \"$API_URL/api/v1/issues/dependency/add\" \\
        -H \"Authorization: Bearer $AUTH_TOKEN\" \\
        -H \"Content-Type: application/json\" \\
        -d '{
            \"issue_id\": \"$ISSUE1_ID\",
            \"depends_on_id\": \"$ISSUE2_ID\",
            \"dependency_type\": \"blocker\",
            \"description\": \"Test dependency\"
        }'
" 0 "check_response_contains '\"success\":\\s*true\|\"id\"'"

# Extract dependency ID for cleanup
DEPENDENCY_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/issues/dependencies?issue_id=$ISSUE1_ID" \
    -H "Authorization: Bearer $AUTH_TOKEN")
DEPENDENCY_ID=$(echo "$DEPENDENCY_RESPONSE" | grep -oP '"id":\s*"\K[^"]+')

echo "Created dependency: $DEPENDENCY_ID"

# Test 2: Check if entity relationship was created (due to dual-write)
run_test "Check Entity Relationship Created" "
    curl -s -X GET \"$API_URL/api/entity/relationship/source?source_id=$ISSUE1_ID\" \\
        -H \"Authorization: Bearer $AUTH_TOKEN\"
" 0 "check_response_contains '$ISSUE2_ID'"

# Test 3: Get entity relationship details
run_test "Get Entity Relationship Details" "
    curl -s -X GET \"$API_URL/api/entity/relationship?source_id=$ISSUE1_ID&relationship_type=depends_on&target_id=$ISSUE2_ID\" \\
        -H \"Authorization: Bearer $AUTH_TOKEN\"
" 0 "check_response_contains 'dependency_type\|blocker'"

# Test 4: Create entity relationship and check if issue dependency is created
TEST_REL_TYPE="related_to"
run_test "Create Entity Relationship" "
    curl -s -X POST \"$API_URL/api/entity/relationship\" \\
        -H \"Authorization: Bearer $AUTH_TOKEN\" \\
        -H \"Content-Type: application/json\" \\
        -d '{
            \"source_id\": \"$ISSUE1_ID\",
            \"relationship_type\": \"$TEST_REL_TYPE\",
            \"target_id\": \"$ISSUE2_ID\",
            \"metadata\": {
                \"relationship_name\": \"related_issue\",
                \"description\": \"Test relationship\"
            }
        }'
" 0 "check_response_contains '\"success\":\\s*true'"

# Test 5: Delete issue dependency and check if entity relationship is deleted
if [[ -n "$DEPENDENCY_ID" ]]; then
    run_test "Delete Issue Dependency" "
        curl -s -X DELETE \"$API_URL/api/v1/issues/dependency/remove?id=$DEPENDENCY_ID\" \\
            -H \"Authorization: Bearer $AUTH_TOKEN\"
    " 0 "check_response_contains '\"success\":\\s*true'"

    # Check if entity relationship was deleted
    run_test "Check Entity Relationship Deleted" "
        curl -s -X GET \"$API_URL/api/entity/relationship?source_id=$ISSUE1_ID&relationship_type=depends_on&target_id=$ISSUE2_ID\" \\
            -H \"Authorization: Bearer $AUTH_TOKEN\"
    " 0 "check_response_contains 'not found\|\"error\"'"
fi

# Test 6: Delete entity relationship
run_test "Delete Entity Relationship" "
    curl -s -X DELETE \"$API_URL/api/entity/relationship?source_id=$ISSUE1_ID&relationship_type=$TEST_REL_TYPE&target_id=$ISSUE2_ID\" \\
        -H \"Authorization: Bearer $AUTH_TOKEN\"
" 0 "check_response_contains '\"success\":\\s*true'"

# Clean up: Delete test issues
echo "Cleaning up test issues..."
curl -s -X DELETE "$API_URL/api/v1/issues/delete?id=$ISSUE1_ID" \
    -H "Authorization: Bearer $AUTH_TOKEN"

curl -s -X DELETE "$API_URL/api/v1/issues/delete?id=$ISSUE2_ID" \
    -H "Authorization: Bearer $AUTH_TOKEN"

# Print test summary
echo -e "\n${YELLOW}TEST SUMMARY${NC}"
echo -e "Total tests:  $TESTS_TOTAL"
echo -e "Passed:      ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed:      ${RED}$TESTS_FAILED${NC}"
echo -e "Skipped:     ${YELLOW}$TESTS_SKIPPED${NC}"

# Return appropriate exit code
if [ "$TESTS_FAILED" -gt 0 ]; then
    exit 1
else
    exit 0
fi