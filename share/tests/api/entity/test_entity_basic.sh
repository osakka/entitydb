#!/bin/bash
# Basic Entity API Test

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
else
    fail "Entity list failed: $LIST_RESPONSE"
fi

# Test entity creation
echo -e "\n${YELLOW}Testing entity creation${NC}"
CREATE_RESPONSE=$(curl -s -X POST "${API_BASE}/entities" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "type": "issue",
        "title": "Test Entity",
        "description": "Created by test script",
        "tags": ["test", "api", "priority:low"],
        "status": "pending"
    }')

if echo "$CREATE_RESPONSE" | grep -q '"status":"ok"'; then
    pass "Entity created successfully"
    # Extract ID from the created entity in the "data" field
    ENTITY_ID=$(echo "$CREATE_RESPONSE" | grep -o '"data":{[^}]*}' | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    if [ -z "$ENTITY_ID" ]; then
        # Try alternative structure
        ENTITY_ID=$(echo "$CREATE_RESPONSE" | grep -o '"entity":{[^}]*}' | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    fi
    echo "Created entity ID: $ENTITY_ID"
else
    fail "Entity creation failed: $CREATE_RESPONSE"
    ENTITY_ID=""
fi

# Test entity retrieval
if [ -n "$ENTITY_ID" ]; then
    echo -e "\n${YELLOW}Testing entity retrieval${NC}"
    GET_RESPONSE=$(curl -s -X GET "${API_BASE}/entities/$ENTITY_ID" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$GET_RESPONSE" | grep -q '"status":"ok"'; then
        pass "Entity retrieved successfully"
    else
        fail "Entity retrieval failed: $GET_RESPONSE"
    fi
    
    # Test entity update
    echo -e "\n${YELLOW}Testing entity update${NC}"
    UPDATE_RESPONSE=$(curl -s -X PUT "${API_BASE}/entities/$ENTITY_ID" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "title": "Updated Test Entity",
            "status": "in_progress",
            "tags": ["test", "api", "priority:medium", "updated"]
        }')
    
    if echo "$UPDATE_RESPONSE" | grep -q '"status":"ok"'; then
        pass "Entity updated successfully"
    else
        fail "Entity update failed: $UPDATE_RESPONSE"
    fi
    
    # Test entity deletion
    echo -e "\n${YELLOW}Testing entity deletion${NC}"
    DELETE_RESPONSE=$(curl -s -X DELETE "${API_BASE}/entities/$ENTITY_ID" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$DELETE_RESPONSE" | grep -q '"status":"ok"'; then
        pass "Entity deleted successfully"
    else
        fail "Entity deletion failed: $DELETE_RESPONSE"
    fi
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