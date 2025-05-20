#!/bin/bash

# EntityDB Temporal API Demo
# Shows how temporal system works with all API operations

BASE_URL="http://localhost:8085/api/v1"

# Start the server first
echo "Starting EntityDB server..."
/opt/entitydb/bin/entitydbd.sh start
sleep 3

# Login
echo "=== Temporal System API Demo ==="
echo "1. Logging in..."
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' | jq -r '.token')

echo "Token: ${TOKEN:0:20}..."
echo

# Create an entity
echo "2. Creating an entity..."
ENTITY=$(curl -s -X POST "$BASE_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:test",
      "status:active",
      "priority:high"
    ],
    "content": [
      {"type": "title", "value": "Temporal Test Entity"}
    ]
  }')

ENTITY_ID=$(echo "$ENTITY" | jq -r '.id')
echo "Created entity: $ENTITY_ID"
echo

# Get entity WITHOUT timestamps (default)
echo "3. Get entity WITHOUT timestamps (default API behavior):"
curl -s -X GET "$BASE_URL/entities/get?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.tags'
echo

# Get entity WITH timestamps
echo "4. Get entity WITH timestamps (include_timestamps=true):"
curl -s -X GET "$BASE_URL/entities/get?id=$ENTITY_ID&include_timestamps=true" \
  -H "Authorization: Bearer $TOKEN" | jq '.tags'
echo

# Update the entity
echo "5. Updating entity (changing status)..."
sleep 1  # Wait a bit to see timestamp difference
curl -s -X PUT "$BASE_URL/entities/update" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$ENTITY_ID\",
    \"tags\": [
      \"type:test\",
      \"status:completed\",
      \"priority:high\"
    ]
  }" > /dev/null
echo "Updated status from 'active' to 'completed'"
echo

# Get history
echo "6. Get entity history (temporal feature):"
curl -s -X GET "$BASE_URL/entities/history?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.[].tags' | head -20
echo

# List entities with query
echo "7. Query entities (tags returned without timestamps by default):"
curl -s -X GET "$BASE_URL/entities/list?tag=type:test" \
  -H "Authorization: Bearer $TOKEN" | jq '.[].tags'
echo

echo "=== Summary ==="
echo "• Storage: All tags have timestamps internally (nanosecond precision)"
echo "• Default API: Returns tags WITHOUT timestamps for easier use"
echo "• include_timestamps=true: Shows the full temporal data"
echo "• History/Temporal queries: Always show timestamps"
echo "• No special temporal handling needed for normal operations"
echo
echo "The temporal system is TRANSPARENT - you don't need to"
echo "handle timestamps unless you specifically want them!"