#!/bin/bash

# Test entity get specifically

BASE_URL="https://localhost:8085"
echo "Testing entity retrieval..."

# Login as admin
echo -e "\n=== Login as admin ==="
LOGIN_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}')
  
ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r .token)
echo "Admin token: $ADMIN_TOKEN"

# Create a simple entity
echo -e "\n=== Creating test entity ==="
CREATE_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags": ["test:retrieval"], "content": "Test content"}')

echo "Create response: $CREATE_RESPONSE"
ID=$(echo "$CREATE_RESPONSE" | jq -r .id)
echo "Entity ID: $ID"

# Try getting with full ID
echo -e "\n=== Getting entity by ID ==="
GET_URL="$BASE_URL/api/v1/entities/get?id=$ID"
echo "GET URL: $GET_URL"

GET_RESPONSE=$(curl -sk -X GET "$GET_URL" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "Get response: $GET_RESPONSE"

# Try listing all entities
echo -e "\n=== Listing all entities ==="
LIST_RESPONSE=$(curl -sk -X GET "$BASE_URL/api/v1/entities/list" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "List response:"
echo "$LIST_RESPONSE" | jq -c '.[] | {id, tags}'

# Try getting by tag
echo -e "\n=== Getting by tag ==="
TAG_RESPONSE=$(curl -sk -X GET "$BASE_URL/api/v1/entities/list?tag=test:retrieval" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "Tag response:"
echo "$TAG_RESPONSE" | jq .

echo -e "\n✅ Test complete"