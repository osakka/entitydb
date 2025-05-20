#!/bin/bash

# Test different ID formats

BASE_URL="https://localhost:8085"
echo "Testing ID formats..."

# Login
LOGIN_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}')
  
ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r .token)
echo "Logged in"

# Create entity with custom UUID-style ID
echo -e "\n=== Creating with UUID ID ==="
UUID_ID=$(uuidgen)
echo "UUID: $UUID_ID"

UUID_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"id\": \"$UUID_ID\", \"tags\": [\"test:uuid\"], \"content\": \"UUID test\"}")

echo "Response: $UUID_RESPONSE"

# Try to get by UUID
echo -e "\n=== Getting by UUID ==="
UUID_GET=$(curl -sk -X GET "$BASE_URL/api/v1/entities/get?id=$UUID_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "Get response: $UUID_GET"

# Create without ID (auto-generated)
echo -e "\n=== Creating with auto ID ==="
AUTO_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags": ["test:auto"], "content": "Auto ID test"}')

echo "Response: $AUTO_RESPONSE"
AUTO_ID=$(echo "$AUTO_RESPONSE" | jq -r .id)
echo "Auto ID: $AUTO_ID"

# Try to get by auto ID
echo -e "\n=== Getting by auto ID ==="
AUTO_GET=$(curl -sk -X GET "$BASE_URL/api/v1/entities/get?id=$AUTO_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "Get response: $AUTO_GET"

# List both by tag
echo -e "\n=== Listing by tag ==="
echo "UUID tag results:"
curl -sk -X GET "$BASE_URL/api/v1/entities/list?tag=test:uuid" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -c '.[] | {id, tags}'

echo -e "\nAuto tag results:"
curl -sk -X GET "$BASE_URL/api/v1/entities/list?tag=test:auto" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -c '.[] | {id, tags}'

echo -e "\n✅ ID format test complete"