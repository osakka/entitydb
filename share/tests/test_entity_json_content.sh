#!/bin/bash

# Simple test for JSON content with the new entity model

BASE_URL="https://localhost:8085"
echo "Testing entity JSON content functionality..."

# Login as admin
echo -e "\n=== Login as admin ==="
LOGIN_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}')
  
echo "Login response: $LOGIN_RESPONSE"
ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r .token)

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo "Failed to get admin token"
    exit 1
fi

echo "Admin token: $ADMIN_TOKEN"

# Create a simple entity with JSON content
echo -e "\n=== Creating entity with JSON content ==="
# First, create JSON content and encode it as a string
JSON_CONTENT='{"message": "Hello, world!", "test": true}'
CONTENT_STRING=$(echo "$JSON_CONTENT" | jq -sR .)

RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
  "tags": ["type:test", "name:simple"],
  "content": $CONTENT_STRING
}
EOF
)

echo "Response: $RESPONSE"
ID=$(echo "$RESPONSE" | jq -r .id)

if [ -z "$ID" ] || [ "$ID" = "null" ]; then
    echo "Failed to create entity"
    echo "Error response:"
    echo "$RESPONSE" | jq .
    exit 1
fi

echo "Created entity: $ID"

# Retrieve the entity
echo -e "\n=== Retrieving entity ==="
GET_RESPONSE=$(curl -sk -X GET "$BASE_URL/api/v1/entities/get?id=$ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "Entity data:"
echo "$GET_RESPONSE" | jq .

# Check content field
echo -e "\n=== Checking content field ==="
CONTENT=$(echo "$GET_RESPONSE" | jq -r .content)
echo "Content: $CONTENT"

if [ "$CONTENT" = "null" ] || [ -z "$CONTENT" ]; then
    echo "✗ Content field is empty"
else
    echo "✓ Content field has data"
    # Try to parse as JSON
    echo "$CONTENT" | jq . 2>/dev/null && echo "✓ Content is valid JSON" || echo "✗ Content is not valid JSON"
fi

echo -e "\n✅ Test complete"