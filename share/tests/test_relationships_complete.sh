#!/bin/bash

# Complete test script for entity relationships in EntityDB

BASE_URL="http://localhost:8085"

echo "Testing Entity Relationships (Complete)..."

# Test 1: Create entities using test endpoint
echo -e "\n=== Test 1: Create test entities ==="
RESPONSE1=$(curl -s -X POST "$BASE_URL/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Task 1",
    "description": "First task",
    "tags": ["type:task", "priority:high"]
  }')

ENTITY1=$(echo "$RESPONSE1" | jq -r .id)
echo "Created entity 1: $ENTITY1"
echo "Response: $RESPONSE1"

RESPONSE2=$(curl -s -X POST "$BASE_URL/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Task 2", 
    "description": "Second task",
    "tags": ["type:task", "priority:medium"]
  }')

ENTITY2=$(echo "$RESPONSE2" | jq -r .id)
echo "Created entity 2: $ENTITY2"
echo "Response: $RESPONSE2"

# Test 2: Create relationships
echo -e "\n=== Test 2: Create relationship (Task 1 blocks Task 2) ==="
REL1=$(curl -s -X POST "$BASE_URL/api/v1/test/relationships/create" \
  -H "Content-Type: application/json" \
  -d "{
    \"source_id\": \"$ENTITY1\",
    \"relationship_type\": \"blocks\",
    \"target_id\": \"$ENTITY2\"
  }")
echo "Created relationship: $REL1"

# Test 3: Create another relationship
echo -e "\n=== Test 3: Create depends_on relationship ==="
REL2=$(curl -s -X POST "$BASE_URL/api/v1/test/relationships/create" \
  -H "Content-Type: application/json" \
  -d "{
    \"source_id\": \"$ENTITY2\",
    \"relationship_type\": \"depends_on\",
    \"target_id\": \"$ENTITY1\"
  }")
echo "Created relationship: $REL2"

# Test 4: Query relationships by source
echo -e "\n=== Test 4: Get relationships by source (Entity 1) ==="
curl -s "$BASE_URL/api/v1/test/relationships/list?source=$ENTITY1" | jq .

# Test 5: Query relationships by target
echo -e "\n=== Test 5: Get relationships by target (Entity 1) ==="
curl -s "$BASE_URL/api/v1/test/relationships/list?target=$ENTITY1" | jq .

# Test 6: Query relationships by source (Entity 2)
echo -e "\n=== Test 6: Get relationships by source (Entity 2) ==="
curl -s "$BASE_URL/api/v1/test/relationships/list?source=$ENTITY2" | jq .

# Test 7: Query relationships by target (Entity 2)
echo -e "\n=== Test 7: Get relationships by target (Entity 2) ==="
curl -s "$BASE_URL/api/v1/test/relationships/list?target=$ENTITY2" | jq .

# Test 8: Verify relationships are stored as entities
echo -e "\n=== Test 8: List all entities to see relationships ==="
# This will fail without auth, but let's try
curl -s "$BASE_URL/api/v1/entities/list" 2>/dev/null | jq . || echo "Need auth to list all entities"

echo -e "\n=== Relationship tests completed ===="