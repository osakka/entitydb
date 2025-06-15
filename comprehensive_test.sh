#!/bin/bash

echo "=== EntityDB Comprehensive Test Suite ==="
echo "Testing all features with sharded indexing enabled"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0

# Test function
test_api() {
    local test_name="$1"
    local expected_result="$2"
    local actual_result="$3"
    
    if [[ "$actual_result" == "$expected_result" ]]; then
        echo -e "${GREEN}✓${NC} $test_name"
        ((PASSED++))
    else
        echo -e "${RED}✗${NC} $test_name"
        echo "  Expected: $expected_result"
        echo "  Actual: $actual_result"
        ((FAILED++))
    fi
}

# Test 1: Authentication
echo "1. Testing Authentication..."
RESPONSE=$(curl -k -s -X POST https://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo "$RESPONSE" | jq -r '.token')
USER_ID=$(echo "$RESPONSE" | jq -r '.user_id')

if [[ ${#TOKEN} -eq 64 && "$USER_ID" != "null" ]]; then
    echo -e "${GREEN}✓${NC} Authentication successful"
    ((PASSED++))
else
    echo -e "${RED}✗${NC} Authentication failed"
    echo "Response: $RESPONSE"
    ((FAILED++))
fi

# Test 2: Basic Entity List
echo -e "\n2. Testing Entity List..."
ENTITY_LIST=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/list?limit=1" \
  -H "Authorization: Bearer $TOKEN")

if [[ $(echo "$ENTITY_LIST" | jq 'type') == '"array"' ]]; then
    echo -e "${GREEN}✓${NC} Entity list returned array"
    ((PASSED++))
else
    echo -e "${RED}✗${NC} Entity list failed"
    echo "Response: $ENTITY_LIST"
    ((FAILED++))
fi

# Test 3: Tag Query (Sessions)
echo -e "\n3. Testing Tag Queries..."
SESSIONS=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/query?tag=type:session&limit=3" \
  -H "Authorization: Bearer $TOKEN")

SESSION_COUNT=$(echo "$SESSIONS" | jq 'length')
if [[ "$SESSION_COUNT" -gt 0 ]]; then
    echo -e "${GREEN}✓${NC} Session query returned $SESSION_COUNT sessions"
    ((PASSED++))
else
    echo -e "${RED}✗${NC} Session query failed"
    echo "Response: $SESSIONS"
    ((FAILED++))
fi

# Test 4: User Query  
USER_QUERY=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/query?tag=type:user" \
  -H "Authorization: Bearer $TOKEN")

if [[ $(echo "$USER_QUERY" | jq 'type') == '"array"' ]]; then
    USER_COUNT=$(echo "$USER_QUERY" | jq 'length')
    echo -e "${GREEN}✓${NC} User query returned $USER_COUNT users"
    ((PASSED++))
else
    echo -e "${RED}✗${NC} User query failed"
    echo "Response: $USER_QUERY"
    ((FAILED++))
fi

# Test 5: System Metrics
echo -e "\n4. Testing System Metrics..."
METRICS=$(curl -k -s https://localhost:8085/api/v1/system/metrics)

TOTAL_ENTITIES=$(echo "$METRICS" | jq '.database.total_entities')
SESSION_COUNT_METRICS=$(echo "$METRICS" | jq '.database.entities_by_type.session')

if [[ "$TOTAL_ENTITIES" -gt 0 && "$SESSION_COUNT_METRICS" -gt 0 ]]; then
    echo -e "${GREEN}✓${NC} System metrics: $TOTAL_ENTITIES total entities, $SESSION_COUNT_METRICS sessions"
    ((PASSED++))
else
    echo -e "${RED}✗${NC} System metrics failed"
    echo "Total entities: $TOTAL_ENTITIES, Sessions: $SESSION_COUNT_METRICS"
    ((FAILED++))
fi

# Test 6: RBAC Metrics (Admin only)
echo -e "\n5. Testing RBAC Metrics..."
RBAC_METRICS=$(curl -k -s -X GET "https://localhost:8085/api/v1/rbac/metrics" \
  -H "Authorization: Bearer $TOKEN")

ACTIVE_SESSIONS=$(echo "$RBAC_METRICS" | jq '.sessions.active_count')
if [[ "$ACTIVE_SESSIONS" -gt 0 ]]; then
    echo -e "${GREEN}✓${NC} RBAC metrics: $ACTIVE_SESSIONS active sessions"
    ((PASSED++))
else
    echo -e "${RED}✗${NC} RBAC metrics failed"
    echo "Response: $RBAC_METRICS"
    ((FAILED++))
fi

# Test 7: Entity Creation
echo -e "\n6. Testing Entity Creation..."
CREATE_RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags":["type:test","name:comprehensive_test"],"content":"Test entity for comprehensive testing"}')

CREATED_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')
if [[ ${#CREATED_ID} -gt 10 && "$CREATED_ID" != "null" ]]; then
    echo -e "${GREEN}✓${NC} Entity created with ID: $CREATED_ID"
    ((PASSED++))
else
    echo -e "${RED}✗${NC} Entity creation failed"
    echo "Response: $CREATE_RESPONSE"
    ((FAILED++))
fi

# Test 8: Entity Retrieval
echo -e "\n7. Testing Entity Retrieval..."
if [[ ${#CREATED_ID} -gt 10 ]]; then
    GET_RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/get?id=$CREATED_ID" \
      -H "Authorization: Bearer $TOKEN")
    
    RETRIEVED_ID=$(echo "$GET_RESPONSE" | jq -r '.id')
    if [[ "$RETRIEVED_ID" == "$CREATED_ID" ]]; then
        echo -e "${GREEN}✓${NC} Entity retrieved successfully"
        ((PASSED++))
    else
        echo -e "${RED}✗${NC} Entity retrieval failed"
        echo "Response: $GET_RESPONSE"
        ((FAILED++))
    fi
else
    echo -e "${YELLOW}⚠${NC} Skipping retrieval test (no entity created)"
fi

# Test 9: Dashboard Status
echo -e "\n8. Testing Dashboard Access..."
DASHBOARD_RESPONSE=$(curl -k -s -o /dev/null -w "%{http_code}" https://localhost:8085/)

if [[ "$DASHBOARD_RESPONSE" == "200" ]]; then
    echo -e "${GREEN}✓${NC} Dashboard accessible (HTTP $DASHBOARD_RESPONSE)"
    ((PASSED++))
else
    echo -e "${RED}✗${NC} Dashboard access failed (HTTP $DASHBOARD_RESPONSE)"
    ((FAILED++))
fi

# Test 10: Health Check
echo -e "\n9. Testing Health Check..."
HEALTH_RESPONSE=$(curl -k -s https://localhost:8085/health)
HEALTH_STATUS=$(echo "$HEALTH_RESPONSE" | jq -r '.status')

if [[ "$HEALTH_STATUS" == "healthy" ]]; then
    echo -e "${GREEN}✓${NC} Health check passed"
    ((PASSED++))
else
    echo -e "${RED}✗${NC} Health check failed"
    echo "Response: $HEALTH_RESPONSE"
    ((FAILED++))
fi

# Summary
echo -e "\n=== Test Results ==="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo -e "Total:  $((PASSED + FAILED))"

if [[ $FAILED -eq 0 ]]; then
    echo -e "\n${GREEN}🎉 All tests passed! Sharded indexing working perfectly.${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed. Review the failures above.${NC}"
    exit 1
fi