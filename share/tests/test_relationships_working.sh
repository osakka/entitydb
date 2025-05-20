#!/bin/bash

# Test script for entity relationships in EntityDB

BASE_URL="http://localhost:8085"

echo "Testing Entity Relationships..."

# First check if test endpoints are working
echo -e "\n=== Test 1: Check test status endpoint ==="
curl -s "$BASE_URL/api/v1/test/status" | jq .

# Create entities using test endpoint
echo -e "\n=== Test 2: Create test entities ==="
ENTITY1=$(curl -s -X POST "$BASE_URL/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:task", "title:Task 1"],
    "content": [{
      "type": "title",
      "value": "Task 1"
    }, {
      "type": "description",
      "value": "First task"
    }]
  }' | jq -r .id)

echo "Created entity 1: $ENTITY1"

ENTITY2=$(curl -s -X POST "$BASE_URL/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:task", "title:Task 2"],
    "content": [{
      "type": "title",
      "value": "Task 2"
    }, {
      "type": "description",
      "value": "Second task"
    }]
  }' | jq -r .id)

echo "Created entity 2: $ENTITY2"

# Create relationships
echo -e "\n=== Test 3: Create relationship (Task 1 blocks Task 2) ==="
curl -s -X POST "$BASE_URL/api/v1/test/relationships/create" \
  -H "Content-Type: application/json" \
  -d "{
    \"source_id\": \"$ENTITY1\",
    \"relationship_type\": \"blocks\",
    \"target_id\": \"$ENTITY2\"
  }" | jq .

# Query relationships by source
echo -e "\n=== Test 4: Get relationships by source ==="
curl -s "$BASE_URL/api/v1/test/relationships/list?source=$ENTITY1" | jq .

# Query relationships by target
echo -e "\n=== Test 5: Get relationships by target ==="
curl -s "$BASE_URL/api/v1/test/relationships/list?target=$ENTITY2" | jq .

# List all entities to see relationships stored as entities
echo -e "\n=== Test 6: List all entities (includes relationships) ==="
curl -s "$BASE_URL/api/v1/test/entities/list" | jq '.[] | select(.tags[] | contains("type:relationship"))'

echo -e "\n=== Relationship tests completed ===="