#!/bin/bash

# Comprehensive test suite for entity relationships

BASE_URL="http://localhost:8085/api/v1"

# Test utilities
timestamp() {
    date -u +"%Y-%m-%dT%H:%M:%S.%NZ"
}

check_response() {
    local response=$1
    local test_name=$2
    if [ -z "$response" ] || [ "$response" = "null" ] || [[ "$response" =~ "error" ]]; then
        echo "❌ $test_name: FAILED"
        echo "Response: $response"
        return 1
    else
        echo "✅ $test_name: PASSED"
        return 0
    fi
}

echo "=== EntityDB Relationship Tests ==="
echo "Starting at: $(timestamp)"
echo

# 1. Create test user and login
echo "=== Test 1: Create admin user and login ==="
USER_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:user", "rbac:role:admin"],
        "content": [
            {"type": "username", "value": "test_admin"},
            {"type": "password_hash", "value": "$2a$10$dummyhash"}
        ]
    }')

USER_ID=$(echo $USER_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
check_response "$USER_ID" "Create test user"

# Create a simple token for testing (no real auth)
TOKEN="test_token_admin_$(date +%s)"

# 2. Create various entity types
echo -e "\n=== Test 2: Create various entity types ==="

# Create project
PROJECT_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:project", "status:active"],
        "content": [
            {"type": "name", "value": "Main Project"},
            {"type": "description", "value": "Primary project for testing"}
        ]
    }')
PROJECT_ID=$(echo $PROJECT_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
check_response "$PROJECT_ID" "Create project"

# Create tasks
TASK1_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:task", "status:pending", "priority:high"],
        "content": [
            {"type": "title", "value": "Backend Development"},
            {"type": "description", "value": "Develop the API"}
        ]
    }')
TASK1_ID=$(echo $TASK1_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
check_response "$TASK1_ID" "Create task 1"

TASK2_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:task", "status:pending", "priority:medium"],
        "content": [
            {"type": "title", "value": "Frontend Development"},
            {"type": "description", "value": "Build the UI"}
        ]
    }')
TASK2_ID=$(echo $TASK2_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
check_response "$TASK2_ID" "Create task 2"

TASK3_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:task", "status:pending", "priority:low"],
        "content": [
            {"type": "title", "value": "Documentation"},
            {"type": "description", "value": "Write docs"}
        ]
    }')
TASK3_ID=$(echo $TASK3_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
check_response "$TASK3_ID" "Create task 3"

# 3. Create relationships
echo -e "\n=== Test 3: Create various relationships ==="

# Project contains tasks
REL1_RESPONSE=$(curl -s -X POST $BASE_URL/test/relationships/create \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"source_id\": \"$PROJECT_ID\",
        \"relationship_type\": \"contains\",
        \"target_id\": \"$TASK1_ID\"
    }")
check_response "$REL1_RESPONSE" "Project contains Task 1"

REL2_RESPONSE=$(curl -s -X POST $BASE_URL/test/relationships/create \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"source_id\": \"$PROJECT_ID\",
        \"relationship_type\": \"contains\",
        \"target_id\": \"$TASK2_ID\"
    }")
check_response "$REL2_RESPONSE" "Project contains Task 2"

REL3_RESPONSE=$(curl -s -X POST $BASE_URL/test/relationships/create \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"source_id\": \"$PROJECT_ID\",
        \"relationship_type\": \"contains\",
        \"target_id\": \"$TASK3_ID\"
    }")
check_response "$REL3_RESPONSE" "Project contains Task 3"

# Task dependencies
REL4_RESPONSE=$(curl -s -X POST $BASE_URL/test/relationships/create \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"source_id\": \"$TASK1_ID\",
        \"relationship_type\": \"blocks\",
        \"target_id\": \"$TASK2_ID\"
    }")
check_response "$REL4_RESPONSE" "Task 1 blocks Task 2"

REL5_RESPONSE=$(curl -s -X POST $BASE_URL/test/relationships/create \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"source_id\": \"$TASK2_ID\",
        \"relationship_type\": \"blocks\",
        \"target_id\": \"$TASK3_ID\"
    }")
check_response "$REL5_RESPONSE" "Task 2 blocks Task 3"

# User assignments
REL6_RESPONSE=$(curl -s -X POST $BASE_URL/test/relationships/create \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"source_id\": \"$USER_ID\",
        \"relationship_type\": \"assigned_to\",
        \"target_id\": \"$TASK1_ID\"
    }")
check_response "$REL6_RESPONSE" "User assigned to Task 1"

# 4. Query relationships
echo -e "\n=== Test 4: Query relationships by source ==="

# Get all relationships from project
PROJECT_RELS=$(curl -s -X GET "$BASE_URL/test/relationships/list?source=$PROJECT_ID")
echo "Project relationships:"
echo "$PROJECT_RELS" | jq '.' 2>/dev/null || echo "$PROJECT_RELS"
CONTAINS_COUNT=$(echo "$PROJECT_RELS" | grep -o "contains" | wc -l)
if [ "$CONTAINS_COUNT" -eq 3 ]; then
    echo "✅ Project has 3 'contains' relationships"
else
    echo "❌ Expected 3 'contains' relationships, got $CONTAINS_COUNT"
fi

echo -e "\n=== Test 5: Query relationships by target ==="

# Get all relationships targeting Task 2
TASK2_RELS=$(curl -s -X GET "$BASE_URL/test/relationships/list?target=$TASK2_ID")
echo "Task 2 incoming relationships:"
echo "$TASK2_RELS" | jq '.' 2>/dev/null || echo "$TASK2_RELS"
INCOMING_COUNT=$(echo "$TASK2_RELS" | grep -o "\"target_id\":\"$TASK2_ID\"" | wc -l)
if [ "$INCOMING_COUNT" -eq 2 ]; then
    echo "✅ Task 2 has 2 incoming relationships"
else
    echo "❌ Expected 2 incoming relationships, got $INCOMING_COUNT"
fi

echo -e "\n=== Test 6: Query relationships by type ==="

# Get all blocking relationships
BLOCKING_RELS=$(curl -s -X GET "$BASE_URL/test/relationships/list?type=blocks")
echo "Blocking relationships:"
echo "$BLOCKING_RELS" | jq '.' 2>/dev/null || echo "$BLOCKING_RELS"

echo -e "\n=== Test 7: Complex queries ==="

# Get all tasks in project (via contains relationship)
echo "Tasks in project (should be 3):"
PROJECT_TASKS=$(curl -s -X GET "$BASE_URL/test/relationships/list?source=$PROJECT_ID" | \
    grep -o '"target_id":"[^"]*' | cut -d'"' -f4)
echo "$PROJECT_TASKS"

# Get blocking chain
echo -e "\nBlocking chain (Task1 -> Task2 -> Task3):"
T1_BLOCKS=$(curl -s -X GET "$BASE_URL/test/relationships/list?source=$TASK1_ID&type=blocks")
echo "Task 1 blocks: $(echo "$T1_BLOCKS" | grep -o '"target_id":"[^"]*' | cut -d'"' -f4)"

T2_BLOCKS=$(curl -s -X GET "$BASE_URL/test/relationships/list?source=$TASK2_ID&type=blocks")
echo "Task 2 blocks: $(echo "$T2_BLOCKS" | grep -o '"target_id":"[^"]*' | cut -d'"' -f4)"

# 8. Test bidirectional relationships
echo -e "\n=== Test 8: Bidirectional relationship ==="

# Create peer relationship
PEER_REL=$(curl -s -X POST $BASE_URL/test/relationships/create \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"source_id\": \"$TASK1_ID\",
        \"relationship_type\": \"related_to\",
        \"target_id\": \"$TASK2_ID\"
    }")
check_response "$PEER_REL" "Task 1 related to Task 2"

# Query both directions
echo "Relationships from Task 1:"
curl -s -X GET "$BASE_URL/test/relationships/list?source=$TASK1_ID" | jq '.' 2>/dev/null

echo -e "\nRelationships to Task 2:"
curl -s -X GET "$BASE_URL/test/relationships/list?target=$TASK2_ID" | jq '.' 2>/dev/null

# 9. Test relationship modification
echo -e "\n=== Test 9: Relationship temporal history ==="

# Since relationships are entities, check history
REL_ID=$(echo "$REL1_RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
if [ ! -z "$REL_ID" ]; then
    echo "Checking history of relationship: $REL_ID"
    HISTORY=$(curl -s -X GET "$BASE_URL/entities/history?id=$REL_ID")
    echo "$HISTORY" | jq '.' 2>/dev/null || echo "$HISTORY"
fi

# 10. Test edge cases
echo -e "\n=== Test 10: Edge cases ==="

# Try to create duplicate relationship
DUP_REL=$(curl -s -X POST $BASE_URL/test/relationships/create \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"source_id\": \"$PROJECT_ID\",
        \"relationship_type\": \"contains\",
        \"target_id\": \"$TASK1_ID\"
    }")
echo "Duplicate relationship attempt:"
echo "$DUP_REL" | jq '.' 2>/dev/null || echo "$DUP_REL"

# Try invalid source
INVALID_REL=$(curl -s -X POST $BASE_URL/test/relationships/create \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "source_id": "invalid_id",
        "relationship_type": "contains",
        "target_id": "'"$TASK1_ID"'"
    }')
echo -e "\nInvalid source attempt:"
echo "$INVALID_REL" | jq '.' 2>/dev/null || echo "$INVALID_REL"

# 11. Performance test
echo -e "\n=== Test 11: Performance with many relationships ==="

# Create 50 relationships
echo "Creating 50 relationships..."
START_TIME=$(date +%s%N)
for i in {1..50}; do
    TEST_ENTITY=$(curl -s -X POST $BASE_URL/test/entities/create \
        -H "Content-Type: application/json" \
        -d "{
            \"tags\": [\"type:test\", \"index:$i\"],
            \"content\": [{\"type\": \"name\", \"value\": \"Test $i\"}]
        }")
    ENTITY_ID=$(echo "$TEST_ENTITY" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
    
    curl -s -X POST $BASE_URL/test/relationships/create \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"source_id\": \"$PROJECT_ID\",
            \"relationship_type\": \"test_rel\",
            \"target_id\": \"$ENTITY_ID\"
        }" > /dev/null
done
END_TIME=$(date +%s%N)
DURATION=$((($END_TIME - $START_TIME) / 1000000))
echo "Created 50 relationships in ${DURATION}ms"

# Query performance
START_TIME=$(date +%s%N)
MANY_RELS=$(curl -s -X GET "$BASE_URL/test/relationships/list?source=$PROJECT_ID")
END_TIME=$(date +%s%N)
QUERY_DURATION=$((($END_TIME - $START_TIME) / 1000000))
REL_COUNT=$(echo "$MANY_RELS" | grep -o "\"id\":" | wc -l)
echo "Queried $REL_COUNT relationships in ${QUERY_DURATION}ms"

# 12. Summary
echo -e "\n=== Test Summary ==="
echo "Completed at: $(timestamp)"
echo "Total relationships created: ~60"
echo "Test coverage: Create, Query (by source/target/type), History, Edge cases, Performance"

echo -e "\n=== All tests completed ==="