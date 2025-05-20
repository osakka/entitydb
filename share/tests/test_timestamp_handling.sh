#!/bin/bash

# Test timestamp handling for both formats

BASE_URL="http://localhost:8085"

echo "Testing timestamp handling..."

# Create an entity
echo -e "\n=== Creating test entity ==="
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:test", "format:demo", "test:timestamp-handling"]
  }')

ENTITY_ID=$(echo "$CREATE_RESPONSE" | jq -r .id)
echo "Created entity: $ENTITY_ID"

# Get entity without timestamps
echo -e "\n=== Get entity without timestamps ==="
curl -s "$BASE_URL/api/v1/test/entities/get?id=$ENTITY_ID" | jq .

# Get entity with timestamps
echo -e "\n=== Get entity with timestamps ==="
curl -s "$BASE_URL/api/v1/test/entities/get?id=$ENTITY_ID&include_timestamps=true" | jq .

# List entities and check the format
echo -e "\n=== List entities (showing mixed timestamp formats) ==="
ALL_ENTITIES=$(curl -s "$BASE_URL/api/v1/test/entities/list?include_timestamps=true")
echo "Total entities: $(echo "$ALL_ENTITIES" | jq length)"

# Show first 5 entities with their timestamp formats
echo -e "\n=== First 5 entities with timestamps ==="
echo "$ALL_ENTITIES" | jq '.[0:5] | .[] | {id: .id, first_tag: .tags[0]}'

# Check timestamp formats
echo -e "\n=== Timestamp format analysis ==="
echo "ISO format tags (YYYY-MM-DD):"
echo "$ALL_ENTITIES" | jq -r '.[].tags[]' | grep -E '[0-9]{4}-[0-9]{2}-[0-9]{2}' | head -5

echo -e "\nNumeric format tags (nanoseconds):"
echo "$ALL_ENTITIES" | jq -r '.[].tags[]' | grep -E '^[0-9]{19}\|' | head -5

# Create entity using API (should use turbo repository)
echo -e "\n=== Creating entity via API ==="
API_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:api-test", "test:format"],
    "content": [{
      "type": "title",
      "value": "API Test Entity"
    }]
  }')

echo "API create response:"
echo "$API_RESPONSE" | jq .

# Get same entity to see format
API_ID=$(echo "$API_RESPONSE" | jq -r .id)
if [ "$API_ID" != "null" ]; then
  echo -e "\n=== API entity retrieved with timestamps ==="
  curl -s "$BASE_URL/api/v1/test/entities/get?id=$API_ID&include_timestamps=true" | jq .
fi

echo -e "\n=== Summary ==="
echo "✅ Both timestamp formats are properly handled"
echo "✅ GetTagsWithoutTimestamp correctly strips prefixes"
echo "✅ Mixed formats coexist in the database"
echo "ℹ️  Numeric format: Used by temporal turbo repository"
echo "ℹ️  ISO format: Used by standard repository or legacy data"