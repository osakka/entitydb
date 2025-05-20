#!/bin/bash

BASE_URL="https://localhost:8085/api/v1"
ADMIN_ID="d4fa9b989e009bbd47612b93605da445"

# Use bcrypt hash for password 'admin'
# Generated with bcrypt
ADMIN_HASH='$2a$10$xqX8tfjz18M73dq5.Pc.PuxfyJ5vHoKCU7LXXgkQcRtHkqjJf9Iqe'

# Format the content value exactly like the existing admin entity
CONTENT="{\"application/octet-stream\":\"{\\\"application/octet-stream\\\":\\\"{\\\\\\\"display_name\\\\\\\":\\\\\\\"Administrator\\\\\\\",\\\\\\\"password_hash\\\\\\\":\\\\\\\"$ADMIN_HASH\\\\\\\",\\\\\\\"username\\\\\\\":\\\\\\\"admin\\\\\\\"}\\\"}\"}"

echo "Updating admin password hash..."
UPDATE_RESPONSE=$(curl -sk -X PUT $BASE_URL/entities/update \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$ADMIN_ID\",
    \"tags\": [
      \"type:user\", 
      \"id:username:admin\", 
      \"rbac:role:admin\", 
      \"rbac:perm:*\", 
      \"status:active\"
    ],
    \"content\": \"$CONTENT\"
  }")

echo "Admin update response:"
echo "$UPDATE_RESPONSE" | jq '.' 2>/dev/null || echo "$UPDATE_RESPONSE"

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