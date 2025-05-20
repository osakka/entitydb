#!/bin/bash

# Very simple entity test

BASE_URL="https://localhost:8085"
echo "Testing simple entity creation..."

# Login as admin
echo -e "\n=== Login as admin ==="
LOGIN_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}')
  
ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r .token)
echo "Admin token: $ADMIN_TOKEN"

# Create a very simple entity
echo -e "\n=== Creating simple entity ==="
RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags": ["type:test", "name:simple"], "content": "Hello, world!"}')

echo "Response: $RESPONSE"

# Try without content field
echo -e "\n=== Creating entity without content ==="
RESPONSE2=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags": ["type:test", "name:nocontent"]}')

echo "Response: $RESPONSE2"

# Try with empty content
echo -e "\n=== Creating entity with empty content ==="
RESPONSE3=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags": ["type:test", "name:empty"], "content": ""}')

echo "Response: $RESPONSE3"