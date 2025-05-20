#!/bin/bash

# Create user directly with proper hash

BASE_URL="https://localhost:8085"
echo "Creating user with proper password hash..."

# Generate bcrypt hash
HASH='$2a$10$W6qJBVqK9KjWvP1O7Cov8uXXD8lP6HH2.nDPdQu6OJD1Xy0G9.Nxu' # "test123"

# Create user entity
RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d "{
    \"tags\": [
      \"type:user\",
      \"id:username:testuser\",
      \"rbac:role:admin\",
      \"rbac:perm:*\",
      \"status:active\"
    ],
    \"content\": {
      \"username\": \"testuser\",
      \"password_hash\": \"$HASH\",
      \"display_name\": \"Test User\"
    }
  }")

echo "Create response:"
echo "$RESPONSE" | jq .

# Test login
echo -e "\n=== Testing login ==="
LOGIN=$(curl -sk -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "test123"}')

echo "Login response:"
echo "$LOGIN" | jq .

TOKEN=$(echo "$LOGIN" | jq -r .token)
if [ "$TOKEN" != "null" ] && [ ! -z "$TOKEN" ]; then
    echo -e "\n✅ Login successful! Token: ${TOKEN:0:20}..."
    
    # Test entity creation with auth
    echo -e "\n=== Creating test entity ==="
    ENTITY=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"tags": ["test:success"], "content": "Hello from authenticated request!"}')
    
    echo "Entity response:"
    echo "$ENTITY" | jq .
else
    echo -e "\n❌ Login failed"
fi