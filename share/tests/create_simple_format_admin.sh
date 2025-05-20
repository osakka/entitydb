#!/bin/bash

BASE_URL="https://localhost:8085/api/v1"

# Create test admin user with simpler content format
echo "Creating admin user with simple content format..."

# Use bcrypt hash for password 'admin'
ADMIN_HASH='$2a$10$xqX8tfjz18M73dq5.Pc.PuxfyJ5vHoKCU7LXXgkQcRtHkqjJf9Iqe'

# Try simpler content format
ADMIN_RESPONSE=$(curl -sk -X POST $BASE_URL/test/entities/create \
  -H "Content-Type: application/json" \
  -d "{
    \"tags\": [
      \"type:user\", 
      \"id:username:simpleadmin\", 
      \"rbac:role:admin\", 
      \"rbac:perm:*\", 
      \"status:active\",
      \"content:type:text/plain\"
    ],
    \"content\": \"{\\\"username\\\":\\\"simpleadmin\\\",\\\"password_hash\\\":\\\"$ADMIN_HASH\\\",\\\"display_name\\\":\\\"Simple Admin\\\"}\"
  }")

echo "Admin creation response:"
echo "$ADMIN_RESPONSE" | jq '.' 2>/dev/null || echo "$ADMIN_RESPONSE"

# Test login with the new user
echo -e "\nTesting simpleadmin login..."
LOGIN=$(curl -sk -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "simpleadmin",
    "password": "admin"
  }')

echo "Login response:"
echo "$LOGIN" | jq '.' 2>/dev/null || echo "$LOGIN"

# Also try a variation with different content encoding
echo -e "\nCreating admin2 user with different content encoding..."
ADMIN2_RESPONSE=$(curl -sk -X POST $BASE_URL/test/entities/create \
  -H "Content-Type: application/json" \
  -d "{
    \"tags\": [
      \"type:user\", 
      \"id:username:admin2\", 
      \"rbac:role:admin\", 
      \"rbac:perm:*\", 
      \"status:active\"
    ],
    \"content\": [{
      \"type\": \"username\",
      \"value\": \"admin2\"
    }, {
      \"type\": \"password_hash\",
      \"value\": \"$ADMIN_HASH\"
    }]
  }")

echo "Admin2 creation response:"
echo "$ADMIN2_RESPONSE" | jq '.' 2>/dev/null || echo "$ADMIN2_RESPONSE"

# Test login with admin2
echo -e "\nTesting admin2 login..."
LOGIN2=$(curl -sk -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin2",
    "password": "admin"
  }')

echo "Login2 response:"
echo "$LOGIN2" | jq '.' 2>/dev/null || echo "$LOGIN2"