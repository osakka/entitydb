#!/bin/bash

BASE_URL="https://localhost:8085/api/v1"

# Use bcrypt hash that I know works
# This is the hash for "password123" generated properly
HASH='$2a$10$xqX8tfjz18M73dq5.Pc.PuxfyJ5vHoKCU7LXXgkQcRtHkqjJf9Iqe'

echo "Creating admin user..."
ADMIN_RESPONSE=$(curl -sk -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d "{
        \"id\": \"user_admin\",
        \"tags\": [\"type:user\", \"rbac:role:admin\", \"rbac:perm:*\"],
        \"content\": [
            {\"type\": \"username\", \"value\": \"admin\"},
            {\"type\": \"password_hash\", \"value\": \"$HASH\"}
        ]
    }")

echo "Admin creation response:"
echo "$ADMIN_RESPONSE" | jq '.' 2>/dev/null || echo "$ADMIN_RESPONSE"

# Test login  
echo -e "\nTesting admin login..."
LOGIN=$(curl -sk -X POST $BASE_URL/auth/login \
    -H "Content-Type: application/json" \
    -d '{
        "username": "admin",
        "password": "password123"
    }')

echo "Login response:"
echo "$LOGIN" | jq '.' 2>/dev/null || echo "$LOGIN"

# List all users
echo -e "\nListing all user entities..."
USERS=$(curl -sk -X GET "$BASE_URL/test/entities/list?tag=type:user")
echo "$USERS" | jq '.' 2>/dev/null || echo "$USERS"