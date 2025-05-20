#!/bin/bash

BASE_URL="http://localhost:8085/api/v1"

# Create test user with bcrypt hash for "password123"
echo "Creating test user..."
USER_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:user", "rbac:role:admin"],
        "content": [
            {"type": "username", "value": "test_user"},
            {"type": "password_hash", "value": "$2a$10$xqX8tfjz18M73dq5.Pc.PuxfyJ5vHoKCU7LXXgkQcRtHkqjJf9Iqe"}
        ]
    }')

echo "User creation response:"
echo "$USER_RESPONSE" | jq '.' 2>/dev/null || echo "$USER_RESPONSE"

# Test login
echo -e "\nTesting login..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/auth/login \
    -H "Content-Type: application/json" \
    -d '{
        "username": "test_user", 
        "password": "password123"
    }')

echo "Login response:"
echo "$LOGIN_RESPONSE" | jq '.' 2>/dev/null || echo "$LOGIN_RESPONSE"