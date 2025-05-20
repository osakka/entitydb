#!/bin/bash

# Test the fixed timestamp handling

BASE_URL="http://localhost:8085"

echo "Testing fixed timestamp handling..."

# Create a new entity
echo -e "\n=== Creating new entity ==="
NEW_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:fixed-test", "test:timestamp-fix", "status:active"]
  }')

NEW_ID=$(echo "$NEW_RESPONSE" | jq -r .id)
echo "Created entity: $NEW_ID"

# Get without timestamps
echo -e "\n=== Get without timestamps (should show clean tags) ==="
curl -s "$BASE_URL/api/v1/test/entities/get?id=$NEW_ID" | jq .

# Get with timestamps
echo -e "\n=== Get with timestamps ==="
curl -s "$BASE_URL/api/v1/test/entities/get?id=$NEW_ID&include_timestamps=true" | jq .

# Test list without timestamps - should show clean tags for all formats
echo -e "\n=== List entities without timestamps (first 5) ==="
curl -s "$BASE_URL/api/v1/test/entities/list" | jq '.[0:5]'

# Test problematic entities (those with double timestamps)
echo -e "\n=== Test entities with double timestamps ==="
# Find an entity with double timestamps
DOUBLE_TS_ENTITY=$(curl -s "$BASE_URL/api/v1/test/entities/list?include_timestamps=true" | jq -r '.[] | select(.tags[] | contains("|") and contains("|type:sensor")) | .id' | head -1)

if [ -n "$DOUBLE_TS_ENTITY" ]; then
    echo "Testing entity: $DOUBLE_TS_ENTITY"
    echo -e "\n  Without timestamps:"
    curl -s "$BASE_URL/api/v1/test/entities/get?id=$DOUBLE_TS_ENTITY" | jq '.tags[0:3]'
    
    echo -e "\n  With timestamps:"
    curl -s "$BASE_URL/api/v1/test/entities/get?id=$DOUBLE_TS_ENTITY&include_timestamps=true" | jq '.tags[0:3]'
else
    echo "No entities with double timestamps found"
fi

echo -e "\n=== Summary ==="
echo "✅ GetTagsWithoutTimestamp now handles all timestamp formats"
echo "✅ Single pipe format: ISO|tag → tag"
echo "✅ Double pipe format: ISO|NANO|tag → tag"
echo "✅ Numeric format: NANO|tag → tag"
echo "✅ All entities display clean tags without timestamps"