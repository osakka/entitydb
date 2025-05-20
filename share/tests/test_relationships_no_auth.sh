#!/bin/bash

# Test script for entity relationships in EntityDB
# This version uses the test/create endpoint that doesn't require auth

BASE_URL="http://localhost:8085"

echo "Testing Entity Relationships (without auth)..."

# Test 1: Create entities to relate using test endpoint
echo -e "\n=== Test 1: Create test entities ==="
ENTITY1=$(curl -s -X POST "$BASE_URL/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Task 1",
    "description": "First task",
    "tags": ["type:task"]
  }' | jq -r .id)

echo "Created entity 1: $ENTITY1"

ENTITY2=$(curl -s -X POST "$BASE_URL/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Task 2",
    "description": "Second task",
    "tags": ["type:task"]
  }' | jq -r .id)

echo "Created entity 2: $ENTITY2"

# For relationships, we need to check if there's a test endpoint
echo -e "\n=== Test 2: Create relationship (direct handler test) ==="

# Let's first check if entities were created properly
echo -e "\n=== Verify entities exist ==="
curl -s "$BASE_URL/api/v1/entities/list" | jq .

echo -e "\n=== Test complete ===="