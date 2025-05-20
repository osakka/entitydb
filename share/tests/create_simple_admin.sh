#!/bin/bash

BASE_URL="http://localhost:8085/api/v1"

# Create default admin/admin user  
echo "Creating simple admin user..."
RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "id": "user_default_admin",
        "tags": ["type:user", "id:username:admin", "rbac:role:admin", "rbac:perm:*"],
        "content": [
            {"type": "username", "value": "admin"},
            {"type": "password_hash", "value": "$2a$10$kxUiw4ax/xfuR0UMPlH4p.cY1fPoXa6iLx0WjNGwfCH8YfKoR0tnK"}
        ]
    }')

echo "Response: $RESPONSE"

# Test login
echo -e "\nTesting login with admin/admin..."
LOGIN=$(curl -s -v -X POST $BASE_URL/auth/login \
    -H "Content-Type: application/json" \
    -d '{
        "username": "admin",
        "password": "admin"
    }' 2>&1)

echo "Login response:"
echo "$LOGIN"