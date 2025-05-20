#!/bin/bash

# Show that all tags are temporal in EntityDB

BASE_URL="http://localhost:8085/api/v1"

# Login
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' | jq -r '.token')

echo "=== EntityDB Temporal Tags Demo ==="
echo

# Create a test entity
echo "1. Creating a test entity..."
ENTITY=$(curl -s -X POST "$BASE_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:demo",
      "status:active",
      "color:blue"
    ],
    "content": [
      {"type": "title", "value": "Temporal Demo"}
    ]
  }')

ENTITY_ID=$(echo "$ENTITY" | jq -r '.id')
echo "Created entity: $ENTITY_ID"
echo

# Get entity WITHOUT timestamps (default)
echo "2. Entity tags WITHOUT timestamps (default):"
curl -s -X GET "$BASE_URL/entities/get?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.tags'
echo

# Get entity WITH timestamps
echo "3. Entity tags WITH timestamps (include_timestamps=true):"
curl -s -X GET "$BASE_URL/entities/get?id=$ENTITY_ID&include_timestamps=true" \
  -H "Authorization: Bearer $TOKEN" | jq '.tags'
echo

echo "=== Explanation ==="
echo "• ALL tags in EntityDB are stored with nanosecond timestamps"
echo "• Format: timestamp|tag (e.g., 2025-05-18T21:17:06.504280777|type:demo)"
echo "• By default, APIs return tags WITHOUT timestamps for easier use"
echo "• Use include_timestamps=true to see the full temporal data"
echo "• This provides complete audit trail and time-travel capabilities"