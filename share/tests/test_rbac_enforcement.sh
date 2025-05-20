#!/bin/bash

# Test script for RBAC enforcement in EntityDB

BASE_URL="http://localhost:8085"

echo "Testing RBAC enforcement..."

# Test 1: Try to access entities without authentication
echo -e "\n=== Test 1: Access without authentication ==="
curl -s "$BASE_URL/api/v1/entities/list" | jq .

# Test 2: Login as admin
echo -e "\n=== Test 2: Login as admin ==="
ADMIN_TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' | jq -r .token)

echo "Admin token: $ADMIN_TOKEN"

# Test 3: Access entities with admin token
echo -e "\n=== Test 3: Access entities with admin token ==="
curl -s "$BASE_URL/api/v1/entities/list" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Test 4: Create a regular user with limited permissions
echo -e "\n=== Test 4: Create regular user ==="
curl -s -X POST "$BASE_URL/api/v1/users/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass",
    "role": "user"
  }' | jq .

# Test 5: Login as regular user
echo -e "\n=== Test 5: Login as regular user ==="
USER_TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass"}' | jq -r .token)

echo "User token: $USER_TOKEN"

# Test 6: Regular user can view entities
echo -e "\n=== Test 6: Regular user viewing entities ==="
curl -s "$BASE_URL/api/v1/entities/list" \
  -H "Authorization: Bearer user:testuser" | jq .

# Test 7: Regular user cannot access system stats (no system:view permission)
echo -e "\n=== Test 7: Regular user accessing system stats (should fail) ==="
curl -s "$BASE_URL/api/v1/dashboard/stats" \
  -H "Authorization: Bearer user:testuser" | jq .

# Test 8: Regular user cannot create other users (no user:create permission)
echo -e "\n=== Test 8: Regular user creating user (should fail) ==="
curl -s -X POST "$BASE_URL/api/v1/users/create" \
  -H "Authorization: Bearer user:testuser" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "anotheruser",
    "password": "pass"
  }' | jq .

# Test 9: Regular user can create entities
echo -e "\n=== Test 9: Regular user creating entity ==="
curl -s -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer user:testuser" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:test", "created_by:testuser"],
    "content": [{
      "type": "title",
      "value": "Test Entity"
    }]
  }' | jq .

echo -e "\n=== RBAC tests completed ==="