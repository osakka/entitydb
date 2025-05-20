#!/bin/bash

# Test script for entity relationships in EntityDB

BASE_URL="http://localhost:8085"

# Login as admin to get token
echo "Logging in as admin..."
ADMIN_TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' | jq -r .token)

echo "Admin token: $ADMIN_TOKEN"

echo "Testing Entity Relationships..."

# Test 1: Create entities to relate
echo -e "\n=== Test 1: Create test entities ==="
ENTITY1=$(curl -s -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:task", "title:Task 1"],
    "content": [{
      "type": "title",
      "value": "Task 1"
    }]
  }' | jq -r .id)

echo "Created entity 1: $ENTITY1"

ENTITY2=$(curl -s -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:task", "title:Task 2"],
    "content": [{
      "type": "title",
      "value": "Task 2"
    }]
  }' | jq -r .id)

echo "Created entity 2: $ENTITY2"

# Test 2: Create relationship
echo -e "\n=== Test 2: Create relationship (Task 1 blocks Task 2) ==="
curl -s -X POST "$BASE_URL/api/v1/entity-relationships" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"source_id\": \"$ENTITY1\",
    \"relationship_type\": \"blocks\",
    \"target_id\": \"$ENTITY2\"
  }" | jq .

# Test 3: Get relationships by source
echo -e "\n=== Test 3: Get relationships by source ==="
curl -s "$BASE_URL/api/v1/entity-relationships?source=$ENTITY1" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Test 4: Get relationships by target
echo -e "\n=== Test 4: Get relationships by target ==="
curl -s "$BASE_URL/api/v1/entity-relationships?target=$ENTITY2" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Test 5: Create another relationship
echo -e "\n=== Test 5: Create depends_on relationship ==="
curl -s -X POST "$BASE_URL/api/v1/entity-relationships" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"source_id\": \"$ENTITY2\",
    \"relationship_type\": \"depends_on\",
    \"target_id\": \"$ENTITY1\"
  }" | jq .

# Test 6: Get all relationships for entity 1
echo -e "\n=== Test 6: Get all relationships for entity 1 ==="
echo "As source:"
curl -s "$BASE_URL/api/v1/entity-relationships?source=$ENTITY1" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

echo "As target:"
curl -s "$BASE_URL/api/v1/entity-relationships?target=$ENTITY1" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Test 7: Test persistence - verify relationships are stored as entities
echo -e "\n=== Test 7: Verify relationships are stored as entities ==="
curl -s "$BASE_URL/api/v1/entities/list?tag=type:relationship" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

echo -e "\n=== Relationship tests completed ===="