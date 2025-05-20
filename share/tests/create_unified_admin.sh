#!/bin/bash

BASE_URL="https://localhost:8085/api/v1"

# Create admin user using the unified entity model
echo "Creating admin user using the unified entity model..."

# Use bcrypt hash for password 'admin'
# Generated using: `htpasswd -bnBC 10 "" admin | tr -d ':\n'`
ADMIN_HASH='$2y$10$kD0/KWUYhFcK.KmUTM/0juSB0wXBXLxA8xCfZ8YPZkJu3eYl62n/K'

ADMIN_RESPONSE=$(curl -sk -X POST $BASE_URL/test/entities/create \
  -H "Content-Type: application/json" \
  -d "{
    \"tags\": [
      \"type:user\", 
      \"id:username:admin\", 
      \"rbac:role:admin\", 
      \"rbac:perm:*\", 
      \"status:active\"
    ],
    \"content\": \"{\\\"username\\\":\\\"admin\\\",\\\"password_hash\\\":\\\"$ADMIN_HASH\\\",\\\"display_name\\\":\\\"Administrator\\\"}\"
  }")

echo "Admin creation response:"
echo "$ADMIN_RESPONSE" | jq '.' 2>/dev/null || echo "$ADMIN_RESPONSE"

# Test login  
echo -e "\nTesting admin login..."
LOGIN=$(curl -sk -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin"
  }')

echo "Login response:"
echo "$LOGIN" | jq '.' 2>/dev/null || echo "$LOGIN"

# List all user entities
echo -e "\nListing all user entities..."
USERS=$(curl -sk -X GET "$BASE_URL/test/entities/list?tag=type:user")
echo "$USERS" | jq '.' 2>/dev/null || echo "$USERS"