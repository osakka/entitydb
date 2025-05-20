#!/bin/bash
# Integration test for entity relationship API endpoints

# Set up test environment
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test_utils.sh"

# Test variables
API_URL="http://localhost:8085"
AUTH_TOKEN=""
TEST_SOURCE_ID="ent_test_source_$(date +%s)"
TEST_TARGET_ID="ent_test_target_$(date +%s)"
TEST_REL_TYPE="depends_on"

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

# Function to check if entity feature is enabled
check_entity_features_enabled() {
    local output="$1"
    if echo "$output" | grep -q '"entity":\s*{\s*"enabled":\s*true'; then
        return 0
    else
        echo "Entity features are not enabled, skipping test"
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

# Login as admin to get auth token
echo "Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" -H "Content-Type: application/json" -d '{"username":"admin","password":"password"}')
AUTH_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -oP '"token":\s*"\K[^"]+')

if [[ -z "$AUTH_TOKEN" ]]; then
    echo "Failed to get auth token"
    exit 1
fi

# Check feature flags
echo "Checking if entity relationships feature is enabled..."
FEATURES_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/features" -H "Authorization: Bearer $AUTH_TOKEN")

ENTITY_ENABLED=$(echo "$FEATURES_RESPONSE" | grep -oP '"entity":\s*{\s*"enabled":\s*\K(true|false)')
RELATIONSHIPS_ENABLED=$(echo "$FEATURES_RESPONSE" | grep -oP '"relationships_enabled":\s*\K(true|false)')

if [[ "$ENTITY_ENABLED" != "true" || "$RELATIONSHIPS_ENABLED" != "true" ]]; then
    echo "Entity relationships feature is not enabled"
    echo "Would you like to enable it for testing? (y/n)"
    read -r ENABLE_RESPONSE
    
    if [[ "$ENABLE_RESPONSE" == "y" ]]; then
        echo "Enabling entity relationships feature..."
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
        echo "Skipping tests since entity relationships feature is not enabled"
        exit 0
    fi
fi

# Create test entities
echo "Creating test entities..."
curl -s -X POST "$API_URL/api/v1/test/entity/create" \
    -H "Authorization: Bearer $AUTH_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"id\": \"$TEST_SOURCE_ID\",
        \"tags\": [\"type:issue\", \"status:pending\"],
        \"content\": [
            {\"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%S.%NZ)\", \"type\": \"title\", \"value\": \"Test Source Entity\"},
            {\"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%S.%NZ)\", \"type\": \"description\", \"value\": \"Entity for testing relationships\"}
        ]
    }"

curl -s -X POST "$API_URL/api/v1/test/entity/create" \
    -H "Authorization: Bearer $AUTH_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"id\": \"$TEST_TARGET_ID\",
        \"tags\": [\"type:issue\", \"status:pending\"],
        \"content\": [
            {\"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%S.%NZ)\", \"type\": \"title\", \"value\": \"Test Target Entity\"},
            {\"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%S.%NZ)\", \"type\": \"description\", \"value\": \"Entity for testing relationships\"}
        ]
    }"

# Test 1: Create a relationship
run_test "Create Relationship" "
    curl -s -X POST \"$API_URL/api/entity/relationship\" \\
        -H \"Authorization: Bearer $AUTH_TOKEN\" \\
        -H \"Content-Type: application/json\" \\
        -d '{
            \"source_id\": \"$TEST_SOURCE_ID\",
            \"relationship_type\": \"$TEST_REL_TYPE\",
            \"target_id\": \"$TEST_TARGET_ID\",
            \"metadata\": {
                \"dependency_type\": \"blocker\",
                \"description\": \"Test relationship\"
            }
        }'
" 0 "check_response_contains '\"success\":\\s*true'"

# Test 2: Get Relationship
run_test "Get Relationship" "
    curl -s -X GET \"$API_URL/api/entity/relationship?source_id=$TEST_SOURCE_ID&relationship_type=$TEST_REL_TYPE&target_id=$TEST_TARGET_ID\" \\
        -H \"Authorization: Bearer $AUTH_TOKEN\"
" 0 "check_response_contains '\"relationship\"'"

# Test 3: List Relationships By Source
run_test "List Relationships By Source" "
    curl -s -X GET \"$API_URL/api/entity/relationship/source?source_id=$TEST_SOURCE_ID\" \\
        -H \"Authorization: Bearer $AUTH_TOKEN\"
" 0 "check_response_contains '\"relationships\"'"

# Test 4: List Relationships By Target
run_test "List Relationships By Target" "
    curl -s -X GET \"$API_URL/api/entity/relationship/target?target_id=$TEST_TARGET_ID\" \\
        -H \"Authorization: Bearer $AUTH_TOKEN\"
" 0 "check_response_contains '\"relationships\"'"

# Test 5: List Relationships By Type
run_test "List Relationships By Type" "
    curl -s -X GET \"$API_URL/api/entity/relationship/type?relationship_type=$TEST_REL_TYPE\" \\
        -H \"Authorization: Bearer $AUTH_TOKEN\"
" 0 "check_response_contains '\"relationships\"'"

# Test 6: Delete Relationship
run_test "Delete Relationship" "
    curl -s -X DELETE \"$API_URL/api/entity/relationship?source_id=$TEST_SOURCE_ID&relationship_type=$TEST_REL_TYPE&target_id=$TEST_TARGET_ID\" \\
        -H \"Authorization: Bearer $AUTH_TOKEN\"
" 0 "check_response_contains '\"success\":\\s*true'"

# Test 7: Verify Relationship Deleted
run_test "Verify Relationship Deleted" "
    curl -s -X GET \"$API_URL/api/entity/relationship?source_id=$TEST_SOURCE_ID&relationship_type=$TEST_REL_TYPE&target_id=$TEST_TARGET_ID\" \\
        -H \"Authorization: Bearer $AUTH_TOKEN\"
" 0 "check_response_contains 'not found\|\"error\"'"

# Clean up: Delete test entities
echo "Cleaning up test entities..."
curl -s -X DELETE "$API_URL/api/v1/test/entity/delete?id=$TEST_SOURCE_ID" \
    -H "Authorization: Bearer $AUTH_TOKEN"

curl -s -X DELETE "$API_URL/api/v1/test/entity/delete?id=$TEST_TARGET_ID" \
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