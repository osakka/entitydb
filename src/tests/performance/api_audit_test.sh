#!/bin/bash
# Comprehensive EntityDB API Endpoint Audit
# Tests every endpoint for functionality and proper responses

set -e

HOST="${1:-https://localhost:8085}"
VERBOSE="${2:-no}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Initialize counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
FAILURES=""

# Test function
test_endpoint() {
    local method="$1"
    local endpoint="$2"
    local auth="$3"
    local data="$4"
    local expected_status="$5"
    local description="$6"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Build curl command with proper quoting
    local curl_args=(-k -s -w '\n%{http_code}' -X "$method")
    
    if [ "$auth" = "yes" ] && [ ! -z "$TOKEN" ]; then
        curl_args+=(-H "Authorization: Bearer $TOKEN")
    fi
    
    if [ ! -z "$data" ]; then
        curl_args+=(-H "Content-Type: application/json" -d "$data")
    fi
    
    curl_args+=("$HOST$endpoint")
    
    if [ "$VERBOSE" = "yes" ]; then
        echo -e "${BLUE}Testing: $method $endpoint${NC}"
        echo "Command: curl ${curl_args[@]}"
    fi
    
    local response=$(curl "${curl_args[@]}" 2>/dev/null || echo "CURL_ERROR")
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [[ "$status_code" =~ ^$expected_status ]]; then
        echo -e "✅ ${GREEN}PASS${NC}: $description"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        if [ "$VERBOSE" = "yes" ]; then
            echo "  Response: $status_code"
            echo "  Body: $(echo $body | jq -c . 2>/dev/null || echo $body | head -c 100)"
        fi
    else
        echo -e "❌ ${RED}FAIL${NC}: $description"
        echo "  Expected: $expected_status, Got: $status_code"
        if [ "$VERBOSE" = "yes" ]; then
            echo "  Body: $(echo $body | jq . 2>/dev/null || echo $body)"
        fi
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILURES="$FAILURES\n  - $method $endpoint: Expected $expected_status, Got $status_code"
    fi
    echo ""
}

echo "======================================"
echo "🔍 EntityDB API Endpoint Audit"
echo "======================================"
echo "Host: $HOST"
echo "Date: $(date)"
echo ""

# 1. Test unauthenticated endpoints first
echo -e "${YELLOW}=== Unauthenticated Endpoints ===${NC}"
echo ""

test_endpoint "GET" "/health" "no" "" "200" "Health check endpoint"
test_endpoint "GET" "/metrics" "no" "" "200" "Prometheus metrics endpoint"
test_endpoint "GET" "/api/v1/system/metrics" "no" "" "200" "System metrics endpoint"
test_endpoint "GET" "/api/v1/rbac/metrics/public" "no" "" "200" "Public RBAC metrics"
test_endpoint "GET" "/api/v1/status" "no" "" "200|404" "API status check"
test_endpoint "GET" "/swagger/doc.json" "no" "" "200" "Swagger documentation"

# 2. Test authentication
echo -e "${YELLOW}=== Authentication Endpoints ===${NC}"
echo ""

# Test login
echo "Testing authentication..."
LOGIN_RESPONSE=$(curl -k -s -X POST "$HOST/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // empty')
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.user.id // empty')

if [ ! -z "$TOKEN" ]; then
    echo -e "✅ ${GREEN}Authentication successful${NC}"
    echo "  Token: ${TOKEN:0:20}..."
    echo "  User ID: $USER_ID"
else
    echo -e "❌ ${RED}Authentication failed${NC}"
    echo "  Response: $LOGIN_RESPONSE"
    exit 1
fi
echo ""

test_endpoint "POST" "/api/v1/auth/login" "no" '{"username":"admin","password":"admin"}' "200" "User login"
test_endpoint "GET" "/api/v1/auth/whoami" "yes" "" "200" "Get current user"
test_endpoint "POST" "/api/v1/auth/refresh" "yes" "" "200" "Refresh token"
test_endpoint "POST" "/api/v1/auth/logout" "yes" "" "200" "User logout"

# 3. Test entity operations
echo -e "${YELLOW}=== Entity Operations ===${NC}"
echo ""

# Create test entity
TEST_ENTITY_ID="test_audit_entity_$(date +%s)"
test_endpoint "POST" "/api/v1/entities/create" "yes" \
    '{"id":"'$TEST_ENTITY_ID'","tags":["type:test","audit:test"],"content":"test content"}' \
    "200|201" "Create entity"

test_endpoint "GET" "/api/v1/entities/list" "yes" "" "200" "List entities"
test_endpoint "GET" "/api/v1/entities/get?id=$TEST_ENTITY_ID" "yes" "" "200" "Get entity by ID"
test_endpoint "GET" "/api/v1/entities/query?tags=type:test" "yes" "" "200" "Query entities by tags"
test_endpoint "GET" "/api/v1/entities/listbytag?tag=type:test" "yes" "" "200" "List entities by tag"
test_endpoint "PUT" "/api/v1/entities/update" "yes" \
    '{"id":"'$TEST_ENTITY_ID'","tags":["type:test","audit:updated"]}' \
    "200" "Update entity"

# 4. Test temporal operations
echo -e "${YELLOW}=== Temporal Operations ===${NC}"
echo ""

TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
test_endpoint "GET" "/api/v1/entities/history?id=$TEST_ENTITY_ID" "yes" "" "200" "Get entity history"
test_endpoint "GET" "/api/v1/entities/as-of?timestamp=$TIMESTAMP&tags=type:test" "yes" "" "200" "Query as-of timestamp"
test_endpoint "GET" "/api/v1/entities/changes?from=$TIMESTAMP&to=$TIMESTAMP" "yes" "" "200" "Get changes in time range"
test_endpoint "GET" "/api/v1/entities/diff?id=$TEST_ENTITY_ID&from=$TIMESTAMP&to=$TIMESTAMP" "yes" "" "200" "Get entity diff"

# 5. Test chunking operations
echo -e "${YELLOW}=== Content Chunking ===${NC}"
echo ""

# Create large entity for chunking test
LARGE_CONTENT=$(python3 -c "print('x' * 5000)")
CHUNK_ENTITY_ID="test_chunk_entity_$(date +%s)"
test_endpoint "POST" "/api/v1/entities/create" "yes" \
    '{"id":"'$CHUNK_ENTITY_ID'","tags":["type:chunk_test"],"content":"'$LARGE_CONTENT'"}' \
    "200|201" "Create large entity for chunking"

test_endpoint "GET" "/api/v1/entities/get-chunk?id=$CHUNK_ENTITY_ID&chunk=0" "yes" "" "200|404" "Get entity chunk"
test_endpoint "GET" "/api/v1/entities/stream-content?id=$CHUNK_ENTITY_ID" "yes" "" "200" "Stream entity content"

# 6. Test relationship operations
echo -e "${YELLOW}=== Entity Relationships ===${NC}"
echo ""

# Create second entity for relationship
REL_ENTITY_ID="test_rel_entity_$(date +%s)"
test_endpoint "POST" "/api/v1/entities/create" "yes" \
    '{"id":"'$REL_ENTITY_ID'","tags":["type:test"]}' \
    "200|201" "Create entity for relationship"

test_endpoint "POST" "/api/v1/entity-relationships" "yes" \
    '{"from_entity_id":"'$TEST_ENTITY_ID'","to_entity_id":"'$REL_ENTITY_ID'","relationship_type":"relates_to"}' \
    "200|201" "Create entity relationship"

test_endpoint "GET" "/api/v1/entity-relationships?entity_id=$TEST_ENTITY_ID" "yes" "" "200" "Get entity relationships"

# 7. Test dataset operations
echo -e "${YELLOW}=== Dataset Management ===${NC}"
echo ""

TEST_DATASET="test_dataset_$(date +%s)"
test_endpoint "GET" "/api/v1/datasets" "yes" "" "200" "List datasets"
test_endpoint "POST" "/api/v1/datasets" "yes" \
    '{"name":"'$TEST_DATASET'","description":"Test dataset"}' \
    "200|201" "Create dataset"
test_endpoint "GET" "/api/v1/datasets/$TEST_DATASET" "yes" "" "200|404" "Get dataset details"
test_endpoint "PUT" "/api/v1/datasets/$TEST_DATASET" "yes" \
    '{"description":"Updated test dataset"}' \
    "200|404" "Update dataset"

# Test dataset-scoped operations
test_endpoint "POST" "/api/v1/datasets/$TEST_DATASET/entities/create" "yes" \
    '{"id":"dataset_entity_test","tags":["type:test"]}' \
    "200|201|404" "Create entity in dataset"
test_endpoint "GET" "/api/v1/datasets/$TEST_DATASET/entities/query?tags=type:test" "yes" "" "200|404" "Query entities in dataset"

# 8. Test user management
echo -e "${YELLOW}=== User Management ===${NC}"
echo ""

TEST_USER="test_user_$(date +%s)"
test_endpoint "POST" "/api/v1/users/create" "yes" \
    '{"username":"'$TEST_USER'","password":"TestPass123!","tags":["rbac:role:user"]}' \
    "200|201|403" "Create user (requires user:create permission)"

test_endpoint "POST" "/api/v1/users/change-password" "yes" \
    '{"old_password":"admin","new_password":"admin"}' \
    "200" "Change password"

# 9. Test admin operations
echo -e "${YELLOW}=== Admin Operations ===${NC}"
echo ""

test_endpoint "GET" "/api/v1/dashboard/stats" "yes" "" "200" "Dashboard statistics"
test_endpoint "GET" "/api/v1/config" "yes" "" "200|403" "Get configuration"
test_endpoint "GET" "/api/v1/feature-flags" "yes" "" "200|403" "Get feature flags"
test_endpoint "GET" "/api/v1/admin/health" "yes" "" "200|403" "Admin health check"
test_endpoint "GET" "/api/v1/admin/log-level" "yes" "" "200|403" "Get log level"
test_endpoint "GET" "/api/v1/admin/trace-subsystems" "yes" "" "200|403" "Get trace subsystems"

# 10. Test metrics operations
echo -e "${YELLOW}=== Metrics Operations ===${NC}"
echo ""

test_endpoint "GET" "/api/v1/metrics/comprehensive" "no" "" "200" "Comprehensive metrics"
test_endpoint "GET" "/api/v1/metrics/history?metric=entity_count&period=1h" "yes" "" "200" "Metric history"
test_endpoint "GET" "/api/v1/metrics/available" "yes" "" "200" "Available metrics list"
test_endpoint "GET" "/api/v1/application/metrics?app=entitydb" "yes" "" "200" "Application metrics"
test_endpoint "GET" "/api/v1/rbac/metrics" "yes" "" "200|403" "RBAC metrics (admin only)"

# 11. Test configuration endpoints
echo -e "${YELLOW}=== Configuration Management ===${NC}"
echo ""

test_endpoint "POST" "/api/v1/config/set" "yes" \
    '{"key":"test_config","value":"test_value"}' \
    "200|403" "Set configuration"

test_endpoint "POST" "/api/v1/feature-flags/set" "yes" \
    '{"flag":"test_feature","enabled":true}' \
    "200|403" "Set feature flag"

# Print summary
echo ""
echo "======================================"
echo "📊 Test Summary"
echo "======================================"
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Failed Endpoints:${NC}"
    echo -e "$FAILURES"
    echo ""
fi

# Calculate success rate
SUCCESS_RATE=$(echo "scale=2; ($PASSED_TESTS / $TOTAL_TESTS) * 100" | bc)
echo "Success Rate: $SUCCESS_RATE%"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}✅ All API endpoints are functioning correctly!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some endpoints failed testing${NC}"
    exit 1
fi