#!/bin/bash

# EntityDB UUID System Demo

BASE_URL="http://localhost:8085/api/v1"

# Login
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' | jq -r '.token')

echo "=== EntityDB UUID-Only System Demo ==="
echo

# Try to create an entity with a custom ID (should be ignored)
echo "1. Attempting to create entity with custom ID..."
RESPONSE=$(curl -s -X POST "$BASE_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "my_custom_id",
    "tags": [
      "type:test",
      "name:UUID Test"
    ],
    "content": []
  }')

ENTITY_ID=$(echo "$RESPONSE" | jq -r '.id')
echo "Created entity ID: $ENTITY_ID"
echo "Notice: Custom ID was ignored, UUID generated instead"
echo

# Create another entity
echo "2. Creating another entity without specifying ID..."
RESPONSE2=$(curl -s -X POST "$BASE_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:test",
      "name:Second Test"
    ],
    "content": []
  }')

ENTITY_ID2=$(echo "$RESPONSE2" | jq -r '.id')
echo "Created entity ID: $ENTITY_ID2"
echo

echo "=== Benefits of UUID-Only System ==="
echo "✓ Guaranteed uniqueness - no ID collisions"
echo "✓ Immutable IDs - prevents ID manipulation"
echo "✓ Cleaner architecture - system IDs vs user IDs"
echo "✓ Better security - can't guess or manipulate IDs"
echo "✓ Simpler API - no need to handle custom ID cases"
echo

echo "=== User-Defined IDs Still Available as Tags ==="
echo "Creating a ticket with user-defined ID as a tag..."
TICKET=$(curl -s -X POST "$BASE_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:ticket",
      "id:ticket:CUSTOM-123",
      "project:DEMO",
      "status:open"
    ],
    "content": [
      {"type": "title", "value": "Demo ticket with custom ID tag"}
    ]
  }')

TICKET_UUID=$(echo "$TICKET" | jq -r '.id')
echo "System ID (UUID): $TICKET_UUID"
echo "User ID (tag): $(echo "$TICKET" | jq -r '.tags[] | select(startswith("id:ticket:"))')"
echo

echo "=== Summary ==="
echo "• System IDs: Always UUIDs, immutable, guaranteed unique"
echo "• User IDs: Stored as tags (id:*), flexible, can be anything"
echo "• Best practice: Use UUID for system operations, tags for user references"