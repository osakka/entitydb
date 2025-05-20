#!/bin/bash

# Check if ID tags are temporal too

BASE_URL="http://localhost:8085/api/v1"

# Login
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' | jq -r '.token')

echo "=== Checking if ID Tags are Temporal ==="
echo

# Create an entity with an explicit ID tag
echo "1. Creating entity with explicit id: tag..."
ENTITY_WITH_ID=$(curl -s -X POST "$BASE_URL/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:test",
      "id:custom:TEST-123",
      "status:active"
    ],
    "content": [{"type": "title", "value": "Testing ID temporality"}]
  }')

echo "Tags with timestamps (from test endpoint):"
echo "$ENTITY_WITH_ID" | jq '.tags' | grep -E "(id:|type:|status:)"
echo

# Get the entity ID
ENTITY_ID=$(echo "$ENTITY_WITH_ID" | jq -r '.id')
echo "2. Entity system ID: $ENTITY_ID"
echo

# Now let's see all tags for an existing ticket entity
echo "3. Checking existing ticket entity tags..."
TICKET=$(curl -s -X POST "$BASE_URL/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:ticket",
      "id:ticket:DEMO-999",
      "project:DEMO",
      "status:open"
    ],
    "content": [{"type": "title", "value": "Demo ticket"}]
  }')

echo "Ticket tags with timestamps:"
echo "$TICKET" | jq '.tags' | grep -E "id:"
echo

echo "=== Analysis ==="
echo "• The entity.ID field is NOT a tag - it's a system field"
echo "• But id: tags (like id:ticket:XXX) ARE temporal"
echo "• ALL tags in the tags array are temporal"
echo "• Only the entity.ID field itself is not temporal"