#!/bin/bash

echo "🧪 EntityDB Unified Sharded Indexing Test Suite"
echo "=============================================="
echo ""

# Server configuration
SERVER="${SERVER:-https://localhost:8085}"
echo "🔗 Testing server: $SERVER"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Helper function to run test
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_pattern="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -n "🔍 Testing: $test_name... "
    
    result=$(eval "$test_command" 2>&1)
    if echo "$result" | grep -q "$expected_pattern"; then
        echo -e "${GREEN}✅ PASS${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ FAIL${NC}"
        echo "   Expected: $expected_pattern"
        echo "   Got: $result"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Authentication test
echo "🔐 Testing Authentication System"
echo "--------------------------------"

LOGIN_RESULT=$(curl -s -k -X POST "$SERVER/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo "$LOGIN_RESULT" | jq -r '.token // empty')

if [[ -n "$TOKEN" && "$TOKEN" != "null" ]]; then
    echo -e "✅ Authentication successful - Token received"
    echo "   Token: ${TOKEN:0:20}..."
else
    echo -e "❌ Authentication failed"
    echo "   Response: $LOGIN_RESULT"
    exit 1
fi

echo ""

# Test 1: Basic Entity Operations
echo "📦 Testing Entity CRUD Operations"
echo "----------------------------------"

# Create test entity
run_test "Create Entity" \
    "curl -s -k -X POST '$SERVER/api/v1/entities/create' \
        -H 'Authorization: Bearer $TOKEN' \
        -H 'Content-Type: application/json' \
        -d '{\"id\":\"test-unified-001\",\"tags\":[\"type:test\",\"environment:unified\",\"status:active\"]}'" \
    '"success":true'

# Wait for indexing
sleep 1

# Get entity by ID
run_test "Get Entity by ID" \
    "curl -s -k -X GET '$SERVER/api/v1/entities/get?id=test-unified-001' \
        -H 'Authorization: Bearer $TOKEN'" \
    '"id":"test-unified-001"'

# Test 2: Tag-Based Queries (Sharded Index)
echo ""
echo "🏷️  Testing Tag-Based Queries (Sharded Index)"
echo "----------------------------------------------"

# List entities by tag
run_test "List by Tag (type:test)" \
    "curl -s -k -X GET '$SERVER/api/v1/entities/list?tag=type:test' \
        -H 'Authorization: Bearer $TOKEN'" \
    'test-unified-001'

run_test "List by Tag (environment:unified)" \
    "curl -s -k -X GET '$SERVER/api/v1/entities/list?tag=environment:unified' \
        -H 'Authorization: Bearer $TOKEN'" \
    'test-unified-001'

run_test "List by Tag (status:active)" \
    "curl -s -k -X GET '$SERVER/api/v1/entities/list?tag=status:active' \
        -H 'Authorization: Bearer $TOKEN'" \
    'test-unified-001'

# Test 3: Advanced Queries
echo ""
echo "🔎 Testing Advanced Query Operations"
echo "------------------------------------"

# Query with multiple filters
run_test "Advanced Query" \
    "curl -s -k -X GET '$SERVER/api/v1/entities/query?tags=type:test,environment:unified&match_all=true' \
        -H 'Authorization: Bearer $TOKEN'" \
    'test-unified-001'

# Test 4: Entity Updates and Tag Changes
echo ""
echo "🔄 Testing Entity Updates"
echo "-------------------------"

# Add new tag to entity
run_test "Add Tag to Entity" \
    "curl -s -k -X PUT '$SERVER/api/v1/entities/update' \
        -H 'Authorization: Bearer $TOKEN' \
        -H 'Content-Type: application/json' \
        -d '{\"id\":\"test-unified-001\",\"tags\":[\"type:test\",\"environment:unified\",\"status:active\",\"updated:true\"]}'" \
    '"success":true'

# Wait for indexing
sleep 1

# Verify new tag is searchable
run_test "Search by New Tag" \
    "curl -s -k -X GET '$SERVER/api/v1/entities/list?tag=updated:true' \
        -H 'Authorization: Bearer $TOKEN'" \
    'test-unified-001'

# Test 5: Temporal Operations
echo ""
echo "⏰ Testing Temporal Operations"
echo "------------------------------"

# Create entity with temporal tags
TIMESTAMP=$(date +%s%N)
run_test "Create Entity with Temporal Data" \
    "curl -s -k -X POST '$SERVER/api/v1/entities/create' \
        -H 'Authorization: Bearer $TOKEN' \
        -H 'Content-Type: application/json' \
        -d '{\"id\":\"temporal-test-001\",\"tags\":[\"type:temporal\",\"${TIMESTAMP}|value:100\"]}'" \
    '"success":true'

# Wait for indexing
sleep 1

# Query temporal entity
run_test "Query Temporal Entity" \
    "curl -s -k -X GET '$SERVER/api/v1/entities/list?tag=type:temporal' \
        -H 'Authorization: Bearer $TOKEN'" \
    'temporal-test-001'

# Test 6: System Health with Sharded Index
echo ""
echo "🏥 Testing System Health"
echo "------------------------"

run_test "Health Check" \
    "curl -s -k -X GET '$SERVER/health'" \
    '"status":"healthy"'

run_test "System Metrics" \
    "curl -s -k -X GET '$SERVER/api/v1/system/metrics'" \
    '"entities":'

# Test 7: Dashboard Functionality
echo ""
echo "📊 Testing Dashboard Integration"
echo "--------------------------------"

run_test "Dashboard Stats" \
    "curl -s -k -X GET '$SERVER/api/v1/dashboard/stats' \
        -H 'Authorization: Bearer $TOKEN'" \
    '"entity_count":'

# Test 8: Performance Verification
echo ""
echo "⚡ Testing Performance"
echo "---------------------"

# Measure response time for tag query
start_time=$(date +%s%N)
curl -s -k -X GET "$SERVER/api/v1/entities/list?tag=type:test" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
end_time=$(date +%s%N)
response_time=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds

if [[ $response_time -lt 1000 ]]; then
    echo -e "✅ Tag query performance: ${response_time}ms (Good)"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "⚠️  Tag query performance: ${response_time}ms (Slow)"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# Test 9: Concurrent Operations
echo ""
echo "🔄 Testing Concurrent Operations"
echo "--------------------------------"

# Test concurrent tag queries
echo "Running 5 concurrent tag queries..."
for i in {1..5}; do
    curl -s -k -X GET "$SERVER/api/v1/entities/list?tag=type:test" \
        -H "Authorization: Bearer $TOKEN" > /tmp/concurrent_test_$i.json &
done

wait
success_count=0
for i in {1..5}; do
    if grep -q "test-unified-001" /tmp/concurrent_test_$i.json; then
        success_count=$((success_count + 1))
    fi
    rm -f /tmp/concurrent_test_$i.json
done

if [[ $success_count -eq 5 ]]; then
    echo -e "✅ Concurrent operations: All 5 queries successful"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "❌ Concurrent operations: Only $success_count/5 queries successful"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# Cleanup
echo ""
echo "🧹 Cleanup"
echo "----------"

curl -s -k -X DELETE "$SERVER/api/v1/entities/delete?id=test-unified-001" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
curl -s -k -X DELETE "$SERVER/api/v1/entities/delete?id=temporal-test-001" \
    -H "Authorization: Bearer $TOKEN" > /dev/null

echo "Test entities cleaned up"

# Final Results
echo ""
echo "📊 Test Results Summary"
echo "======================="
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

if [[ $FAILED_TESTS -eq 0 ]]; then
    echo ""
    echo -e "${GREEN}🎉 ALL TESTS PASSED! Unified sharded indexing system is working perfectly.${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}❌ Some tests failed. Please review the failures above.${NC}"
    exit 1
fi