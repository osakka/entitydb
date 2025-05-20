#!/bin/bash

# Simple test for entity creation

BASE_URL="http://localhost:8085"

echo "Simple entity test..."

# Login as admin
echo -e "\n=== Login as admin ==="
ADMIN_TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' | jq -r .token)

echo "Admin token: $ADMIN_TOKEN"

# Test token directly
echo -e "\n=== Test with token directly ==="
curl -s -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:test", "name:Test Entity"],
    "content": [{
      "type": "title",
      "value": "Test Entity"
    }]
  }' | jq .

# Test listing entities
echo -e "\n=== List entities ==="
curl -s "$BASE_URL/api/v1/entities/list" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

echo -e "\n=== Test complete ===="