#!/bin/bash

# Check raw tag format in EntityDB storage

BASE_URL="http://localhost:8085/api/v1"

# Login
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' | jq -r '.token')

echo "=== Raw Tag Storage in EntityDB ==="
echo

# Create a test entity using test endpoint
echo "Creating test entity..."
TEST_ENTITY=$(curl -s -X POST "$BASE_URL/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:raw_test", "status:checking", "version:1.0"],
    "content": [{"type": "note", "value": "Testing raw tag storage"}]
  }')

echo "Response from test endpoint:"
echo "$TEST_ENTITY" | jq '.tags' | head -10
echo

# Let's also check via the regular API
ENTITY_ID=$(echo "$TEST_ENTITY" | jq -r '.id')
echo "Getting entity via regular API:"
REGULAR=$(curl -s -X GET "$BASE_URL/entities/get?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Tags from regular API:"
echo "$REGULAR" | jq '.tags'
echo

echo "=== Key Points ==="
echo "• Internally, ALL tags have timestamps"
echo "• Storage format: timestamp|tag"
echo "• Regular API strips timestamps for convenience"
echo "• Test endpoints may show raw format"
echo "• This is the 'temporal-only' system"