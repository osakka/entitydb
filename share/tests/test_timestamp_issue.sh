#!/bin/bash

# Test timestamp stripping issue

BASE_URL="http://localhost:8085"

echo "Testing timestamp stripping issue..."

# Create a simple entity
echo -e "\n=== Creating entity ==="
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:debug", "test:simple"]
  }')

echo "Create response:"
echo "$CREATE_RESPONSE" | jq .

ENTITY_ID=$(echo "$CREATE_RESPONSE" | jq -r .id)
echo "Entity ID: $ENTITY_ID"

# Get entity normally (should fail if bug present)
echo -e "\n=== Getting entity without timestamps ==="
GET_NORMAL=$(curl -s -X GET "$BASE_URL/api/v1/test/entities/get?id=$ENTITY_ID")
echo "Normal get response:"
echo "$GET_NORMAL" | jq .

# Get entity with timestamps
echo -e "\n=== Getting entity with timestamps ==="
GET_WITH_TS=$(curl -s -X GET "$BASE_URL/api/v1/test/entities/get?id=$ENTITY_ID&include_timestamps=true")
echo "With timestamps response:"
echo "$GET_WITH_TS" | jq .

# Raw database check
echo -e "\n=== Let's check the raw database entry ==="
cat /opt/entitydb/var/db/binary/entities.log | grep "$ENTITY_ID" || echo "No entry found in log"

# Also list entities
echo -e "\n=== List all entities (no timestamps) ==="
LIST_NORMAL=$(curl -s -X GET "$BASE_URL/api/v1/test/entities/list")
echo "Normal list response:"
echo "$LIST_NORMAL" | jq .

echo -e "\n=== List all entities (with timestamps) ==="
LIST_WITH_TS=$(curl -s -X GET "$BASE_URL/api/v1/test/entities/list?include_timestamps=true")  
echo "With timestamps list response:"
echo "$LIST_WITH_TS" | jq .