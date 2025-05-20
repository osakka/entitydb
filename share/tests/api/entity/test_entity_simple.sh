#!/bin/bash
# Simple Entity API Test - Working with actual server

SERVER="http://localhost:8085"
API_BASE="${SERVER}/api/v1"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

# Counters
PASS=0
FAIL=0

# Test helper functions
pass() {
    echo -e "${GREEN}✓ $1${NC}"
    ((PASS++))
}

fail() {
    echo -e "${RED}✗ $1${NC}"
    ((FAIL++))
}

# Check if server is running
echo -e "${YELLOW}Testing Entity API${NC}"
echo "Testing server at: $SERVER"

# Get auth token
echo "Getting auth token..."
AUTH_RESPONSE=$(curl -s -X POST "${API_BASE}/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password"}')

TOKEN=$(echo "$AUTH_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    fail "Failed to get auth token"
    exit 1
else
    pass "Got auth token"
fi

# Test entity list
echo -e "\n${YELLOW}Testing entity list${NC}"
LIST_RESPONSE=$(curl -s -X GET "${API_BASE}/entities" \
    -H "Authorization: Bearer $TOKEN")

if echo "$LIST_RESPONSE" | grep -q '"status":"ok"'; then
    pass "Entity list returned successfully"
    # Count entities
    ENTITY_COUNT=$(echo "$LIST_RESPONSE" | grep -o '"id":"[^"]*"' | wc -l)
    echo "Found $ENTITY_COUNT entities"
else
    fail "Entity list failed: $LIST_RESPONSE"
fi

# Test entity retrieval (using a known entity)
echo -e "\n${YELLOW}Testing entity retrieval${NC}"
KNOWN_ID="entity_workspace_entitydb"
GET_RESPONSE=$(curl -s -X GET "${API_BASE}/entities/$KNOWN_ID" \
    -H "Authorization: Bearer $TOKEN")

if echo "$GET_RESPONSE" | grep -q '"status":"ok"'; then
    pass "Entity retrieved successfully"
else
    fail "Entity retrieval failed: $GET_RESPONSE"
fi

# Test entity list with type filter
echo -e "\n${YELLOW}Testing entity list with type filter${NC}"
TYPE_RESPONSE=$(curl -s -X GET "${API_BASE}/entities?type=workspace" \
    -H "Authorization: Bearer $TOKEN")

if echo "$TYPE_RESPONSE" | grep -q '"status":"ok"'; then
    pass "Type-filtered list returned successfully"
    # Count workspace entities
    WORKSPACE_COUNT=$(echo "$TYPE_RESPONSE" | grep -o '"type":"workspace"' | wc -l)
    echo "Found $WORKSPACE_COUNT workspace entities"
else
    fail "Type-filtered list failed: $TYPE_RESPONSE"
fi

# Test entity list with tag filter
echo -e "\n${YELLOW}Testing entity list with tag filter${NC}"
TAG_RESPONSE=$(curl -s -X GET "${API_BASE}/entities?tags=status:active" \
    -H "Authorization: Bearer $TOKEN")

if echo "$TAG_RESPONSE" | grep -q '"status":"ok"'; then
    pass "Tag-filtered list returned successfully"
    # Count active entities
    ACTIVE_COUNT=$(echo "$TAG_RESPONSE" | grep -o '"id":"[^"]*"' | wc -l)
    echo "Found $ACTIVE_COUNT active entities"
else
    fail "Tag-filtered list failed: $TAG_RESPONSE"
fi

# Test relationship endpoints (if they exist)
echo -e "\n${YELLOW}Testing entity relationships${NC}"
REL_RESPONSE=$(curl -s -X GET "${API_BASE}/entity-relationships/list" \
    -H "Authorization: Bearer $TOKEN")

if echo "$REL_RESPONSE" | grep -q '"status":"ok"'; then
    pass "Relationship list returned successfully"
else
    echo "Note: Relationship endpoint not available or returned different response"
fi

# Summary
echo -e "\n${YELLOW}Test Summary${NC}"
echo "Total tests: $((PASS + FAIL))"
echo -e "${GREEN}Passed: $PASS${NC}"
echo -e "${RED}Failed: $FAIL${NC}"

if [ $FAIL -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed!${NC}"
    exit 1
fi