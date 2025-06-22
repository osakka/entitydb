#!/bin/bash
# Test 1: Authentication and Authorization Flows
# EntityDB E2E Production Readiness Testing

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="https://localhost:8085"
CURL_OPTS="-k -s"
TEST_RESULTS="/tmp/e2e_auth_results.log"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Utility functions
log_test() {
    echo -e "\n${YELLOW}TEST:${NC} $1"
    ((TESTS_RUN++))
}

log_pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    echo "  Details: $2"
    ((TESTS_FAILED++))
}

# Start test suite
echo "==================================="
echo "EntityDB Authentication Test Suite"
echo "==================================="
echo "Server: $BASE_URL"
echo "Time: $(date)"
echo ""

# Test 1.1: Admin login with correct credentials
log_test "Admin login with correct credentials"
LOGIN_RESPONSE=$(curl $CURL_OPTS -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}' 2>&1)

if echo "$LOGIN_RESPONSE" | grep -q '"token"'; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    log_pass "Admin login successful, token received"
    echo "  Token: ${TOKEN:0:20}..."
else
    log_fail "Admin login failed" "$LOGIN_RESPONSE"
    exit 1
fi

# Test 1.2: Failed login with incorrect credentials
log_test "Failed login with incorrect password"
FAIL_RESPONSE=$(curl $CURL_OPTS -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"wrongpassword"}' -w "\nHTTP_CODE:%{http_code}" 2>&1)

HTTP_CODE=$(echo "$FAIL_RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$HTTP_CODE" = "401" ]; then
    log_pass "Incorrect password properly rejected with 401"
else
    log_fail "Expected 401, got $HTTP_CODE" "$FAIL_RESPONSE"
fi

# Test 1.3: Failed login with non-existent user
log_test "Failed login with non-existent user"
FAIL_RESPONSE=$(curl $CURL_OPTS -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"nonexistent","password":"password"}' -w "\nHTTP_CODE:%{http_code}" 2>&1)

HTTP_CODE=$(echo "$FAIL_RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$HTTP_CODE" = "401" ]; then
    log_pass "Non-existent user properly rejected with 401"
else
    log_fail "Expected 401, got $HTTP_CODE" "$FAIL_RESPONSE"
fi

# Test 1.4: Token validation - valid token
log_test "Access protected endpoint with valid token"
STATS_RESPONSE=$(curl $CURL_OPTS -X GET "$BASE_URL/api/v1/dashboard/stats" \
    -H "Authorization: Bearer $TOKEN" -w "\nHTTP_CODE:%{http_code}" 2>&1)

HTTP_CODE=$(echo "$STATS_RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$HTTP_CODE" = "200" ]; then
    log_pass "Valid token accepted for protected endpoint"
else
    log_fail "Expected 200, got $HTTP_CODE" "$STATS_RESPONSE"
fi

# Test 1.5: Token validation - invalid token
log_test "Access protected endpoint with invalid token"
INVALID_RESPONSE=$(curl $CURL_OPTS -X GET "$BASE_URL/api/v1/dashboard/stats" \
    -H "Authorization: Bearer invalid-token-12345" -w "\nHTTP_CODE:%{http_code}" 2>&1)

HTTP_CODE=$(echo "$INVALID_RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$HTTP_CODE" = "401" ]; then
    log_pass "Invalid token properly rejected with 401"
else
    log_fail "Expected 401, got $HTTP_CODE" "$INVALID_RESPONSE"
fi

# Test 1.6: Token validation - missing token
log_test "Access protected endpoint without token"
NO_TOKEN_RESPONSE=$(curl $CURL_OPTS -X GET "$BASE_URL/api/v1/dashboard/stats" \
    -w "\nHTTP_CODE:%{http_code}" 2>&1)

HTTP_CODE=$(echo "$NO_TOKEN_RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$HTTP_CODE" = "401" ]; then
    log_pass "Missing token properly rejected with 401"
else
    log_fail "Expected 401, got $HTTP_CODE" "$NO_TOKEN_RESPONSE"
fi

# Test 1.7: Create a test user for permission testing
log_test "Create test user with limited permissions"
USER_RESPONSE=$(curl $CURL_OPTS -X POST "$BASE_URL/api/v1/users/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "testuser",
        "password": "testpass123",
        "email": "test@example.com",
        "tags": [
            "rbac:role:user",
            "rbac:perm:entity:view",
            "rbac:perm:entity:create"
        ]
    }' 2>&1)

if echo "$USER_RESPONSE" | grep -q '"id"'; then
    TEST_USER_ID=$(echo "$USER_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    log_pass "Test user created successfully"
    echo "  User ID: $TEST_USER_ID"
else
    log_fail "Failed to create test user" "$USER_RESPONSE"
fi

# Test 1.8: Login as test user
log_test "Login as test user"
TEST_LOGIN=$(curl $CURL_OPTS -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"testpass123"}' 2>&1)

if echo "$TEST_LOGIN" | grep -q '"token"'; then
    TEST_TOKEN=$(echo "$TEST_LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    log_pass "Test user login successful"
else
    log_fail "Test user login failed" "$TEST_LOGIN"
fi

# Test 1.9: Test user can view entities (has permission)
log_test "Test user viewing entities (allowed)"
VIEW_RESPONSE=$(curl $CURL_OPTS -X GET "$BASE_URL/api/v1/entities/list" \
    -H "Authorization: Bearer $TEST_TOKEN" -w "\nHTTP_CODE:%{http_code}" 2>&1)

HTTP_CODE=$(echo "$VIEW_RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$HTTP_CODE" = "200" ]; then
    log_pass "Test user can view entities as expected"
else
    log_fail "Expected 200, got $HTTP_CODE" "$VIEW_RESPONSE"
fi

# Test 1.10: Test user cannot delete entities (no permission)
log_test "Test user deleting entities (forbidden)"
DELETE_RESPONSE=$(curl $CURL_OPTS -X DELETE "$BASE_URL/api/v1/entities/$TEST_USER_ID" \
    -H "Authorization: Bearer $TEST_TOKEN" -w "\nHTTP_CODE:%{http_code}" 2>&1)

HTTP_CODE=$(echo "$DELETE_RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$HTTP_CODE" = "403" ]; then
    log_pass "Test user properly forbidden from deleting (403)"
else
    log_fail "Expected 403, got $HTTP_CODE" "$DELETE_RESPONSE"
fi

# Test 1.11: Test user cannot access admin endpoints
log_test "Test user accessing admin endpoint (forbidden)"
ADMIN_RESPONSE=$(curl $CURL_OPTS -X GET "$BASE_URL/api/v1/rbac/metrics" \
    -H "Authorization: Bearer $TEST_TOKEN" -w "\nHTTP_CODE:%{http_code}" 2>&1)

HTTP_CODE=$(echo "$ADMIN_RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$HTTP_CODE" = "403" ]; then
    log_pass "Test user properly forbidden from admin endpoint (403)"
else
    log_fail "Expected 403, got $HTTP_CODE" "$ADMIN_RESPONSE"
fi

# Test 1.12: Concurrent session testing
log_test "Concurrent sessions for same user"
# Create 5 concurrent login sessions
for i in {1..5}; do
    (
        SESSION=$(curl $CURL_OPTS -X POST "$BASE_URL/api/v1/auth/login" \
            -H "Content-Type: application/json" \
            -d '{"username":"admin","password":"admin"}' 2>&1)
        
        if echo "$SESSION" | grep -q '"token"'; then
            echo "Session $i: SUCCESS"
        else
            echo "Session $i: FAILED - $SESSION"
        fi
    ) &
done
wait

log_pass "Concurrent session creation completed"

# Test 1.13: Session logout/invalidation
log_test "Logout invalidates session"
# First verify token works
PRE_LOGOUT=$(curl $CURL_OPTS -X GET "$BASE_URL/api/v1/dashboard/stats" \
    -H "Authorization: Bearer $TEST_TOKEN" -w "\nHTTP_CODE:%{http_code}" 2>&1)

PRE_CODE=$(echo "$PRE_LOGOUT" | grep "HTTP_CODE:" | cut -d: -f2)

# Logout
LOGOUT_RESPONSE=$(curl $CURL_OPTS -X POST "$BASE_URL/api/v1/auth/logout" \
    -H "Authorization: Bearer $TEST_TOKEN" -w "\nHTTP_CODE:%{http_code}" 2>&1)

# Try to use token after logout
POST_LOGOUT=$(curl $CURL_OPTS -X GET "$BASE_URL/api/v1/dashboard/stats" \
    -H "Authorization: Bearer $TEST_TOKEN" -w "\nHTTP_CODE:%{http_code}" 2>&1)

POST_CODE=$(echo "$POST_LOGOUT" | grep "HTTP_CODE:" | cut -d: -f2)

if [ "$PRE_CODE" = "200" ] && [ "$POST_CODE" = "401" ]; then
    log_pass "Token properly invalidated after logout"
else
    log_fail "Logout validation failed" "Pre: $PRE_CODE, Post: $POST_CODE"
fi

# Test 1.14: Cross-dataset access control
log_test "Cross-dataset access control"
# Create entity in private dataset as admin
PRIVATE_ENTITY=$(curl $CURL_OPTS -X POST "$BASE_URL/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "tags": [
            "type:secret",
            "dataset:admin-only",
            "content:classified"
        ],
        "content": "This is secret data"
    }' 2>&1)

if echo "$PRIVATE_ENTITY" | grep -q '"id"'; then
    PRIVATE_ID=$(echo "$PRIVATE_ENTITY" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    log_pass "Private entity created in admin dataset"
    
    # Try to access as test user (should fail)
    # Need new test user token first
    TEST_LOGIN2=$(curl $CURL_OPTS -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"testuser","password":"testpass123"}' 2>&1)
    
    TEST_TOKEN2=$(echo "$TEST_LOGIN2" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    
    ACCESS_RESPONSE=$(curl $CURL_OPTS -X GET "$BASE_URL/api/v1/entities/get?id=$PRIVATE_ID" \
        -H "Authorization: Bearer $TEST_TOKEN2" -w "\nHTTP_CODE:%{http_code}" 2>&1)
    
    HTTP_CODE=$(echo "$ACCESS_RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
    if [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "404" ]; then
        log_pass "Cross-dataset access properly restricted"
    else
        log_fail "Expected 403/404, got $HTTP_CODE" "$ACCESS_RESPONSE"
    fi
else
    log_fail "Failed to create private entity" "$PRIVATE_ENTITY"
fi

# Test 1.15: Permission inheritance
log_test "Permission inheritance (wildcard permissions)"
# Admin has rbac:perm:* which should grant all permissions
CONFIG_RESPONSE=$(curl $CURL_OPTS -X GET "$BASE_URL/api/v1/config" \
    -H "Authorization: Bearer $TOKEN" -w "\nHTTP_CODE:%{http_code}" 2>&1)

HTTP_CODE=$(echo "$CONFIG_RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$HTTP_CODE" = "200" ]; then
    log_pass "Wildcard permission grants access to config endpoint"
else
    log_fail "Expected 200, got $HTTP_CODE" "$CONFIG_RESPONSE"
fi

# Summary
echo ""
echo "==================================="
echo "Authentication Test Suite Complete"
echo "==================================="
echo "Tests Run:    $TESTS_RUN"
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All authentication tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed. Please review the output above.${NC}"
    exit 1
fi