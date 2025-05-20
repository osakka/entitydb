#!/bin/bash

# Test complete entity workflow

BASE_URL="https://localhost:8085"
echo "Testing entity workflow..."

# Login as admin
echo -e "\n=== Login as admin ==="
LOGIN_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}')
  
ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r .token)
echo "Admin token exists: $([ ! -z "$ADMIN_TOKEN" ] && echo "Yes" || echo "No")"

# Create entity with simple string
echo -e "\n=== Creating entity with string ==="
RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags": ["test:workflow"], "content": "Hello, EntityDB!"}')

echo "Create response:"
echo "$RESPONSE" | jq .

ID=$(echo "$RESPONSE" | jq -r .id)
echo "Created entity ID: $ID"

# Check what was stored
echo -e "\n=== Checking stored content ==="
STORED_CONTENT=$(echo "$RESPONSE" | jq -r .content)
echo "Base64 content from create: $STORED_CONTENT"
echo "Decoded: $(echo "$STORED_CONTENT" | base64 -d)"

# Retrieve the entity
echo -e "\n=== Retrieving entity ==="
GET_RESPONSE=$(curl -sk -X GET "$BASE_URL/api/v1/entities/get?id=$ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "Get response:"
echo "$GET_RESPONSE" | jq .

# Compare content
echo -e "\n=== Comparing content ==="
RETRIEVED_CONTENT=$(echo "$GET_RESPONSE" | jq -r .content)
echo "Base64 content from get: $RETRIEVED_CONTENT"
echo "Decoded: $(echo "$RETRIEVED_CONTENT" | base64 -d)"

echo -e "\n=== Content comparison ==="
echo "Create content: $STORED_CONTENT"
echo "Get content:    $RETRIEVED_CONTENT"
echo "Same: $([ "$STORED_CONTENT" == "$RETRIEVED_CONTENT" ] && echo "No" || echo "Yes, different!")"

# Test large content
echo -e "\n=== Testing large content (10KB) ==="
LARGE_CONTENT=$(python3 -c "print('X' * 10240)")
LARGE_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
  "tags": ["test:large"],
  "content": "$LARGE_CONTENT"
}
EOF
)

LARGE_ID=$(echo "$LARGE_RESPONSE" | jq -r .id)
echo "Created large entity: $LARGE_ID"

# Check if chunked
LARGE_TAGS=$(echo "$LARGE_RESPONSE" | jq -r '.tags[]' | grep content:)
echo "Content tags: $LARGE_TAGS"

echo -e "\n✅ Workflow test complete"