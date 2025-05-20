#!/bin/bash

# Test JSON content with proper format

BASE_URL="https://localhost:8085"
echo "Testing JSON entity creation..."

# Login as admin
echo -e "\n=== Login as admin ==="
LOGIN_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}')
  
ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r .token)
echo "Admin token: $ADMIN_TOKEN"

# Create entity with JSON object content
echo -e "\n=== Creating entity with JSON object ==="
RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags": ["type:document", "format:json"], "content": {"message": "Hello from JSON", "count": 42, "active": true}}')

echo "Response: $RESPONSE"
JSON_ID=$(echo "$RESPONSE" | jq -r .id)

# Create entity with string content
echo -e "\n=== Creating entity with string content ==="
RESPONSE2=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags": ["type:document", "format:text"], "content": "Plain text content"}')

echo "Response: $RESPONSE2"
TEXT_ID=$(echo "$RESPONSE2" | jq -r .id)

# Retrieve and decode content
echo -e "\n=== Retrieving JSON entity ==="
JSON_GET=$(curl -sk -X GET "$BASE_URL/api/v1/entities/get?id=$JSON_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "Raw response: $JSON_GET"
ENCODED_CONTENT=$(echo "$JSON_GET" | jq -r .content)
echo "Encoded content: $ENCODED_CONTENT"

# Decode the base64 content
echo -e "\n=== Decoding content ==="
if [ ! -z "$ENCODED_CONTENT" ] && [ "$ENCODED_CONTENT" != "null" ]; then
    DECODED=$(echo "$ENCODED_CONTENT" | base64 -d)
    echo "Decoded content: $DECODED"
    
    # Try to parse as JSON
    echo "$DECODED" | jq . 2>/dev/null && echo "✓ Valid JSON" || echo "✗ Not valid JSON"
fi

echo -e "\n✅ Test complete"