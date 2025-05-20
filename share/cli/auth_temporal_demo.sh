#!/bin/bash

# EntityDB Authentication with Temporal System Demo

BASE_URL="http://localhost:8085/api/v1"

# Start server
echo "=== Authentication & Temporal System Demo ==="
/opt/entitydb/bin/entitydbd.sh start
sleep 3

# Show how authentication works with temporal system
echo "1. Authentication is NOT affected by temporal system"
echo "   - You login with username/password as normal"
echo "   - Temporal tracking happens automatically in background"
echo

# Login normally
echo "2. Login request (no temporal handling needed):"
echo "POST /auth/login"
echo '{"username": "admin", "password": "admin"}'
echo
echo "Response:"
RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}')
echo "$RESPONSE" | jq
TOKEN=$(echo "$RESPONSE" | jq -r '.token')
echo

# Show what's stored internally
echo "3. Behind the scenes - User entity with temporal tags:"
echo "   The 'admin' user entity has temporal tags internally:"
echo "   - 2025-05-18T22:40:39.903749423+01:00|type:user"
echo "   - 2025-05-18T22:40:39.903749423+01:00|id:username:admin"
echo "   - 2025-05-18T22:40:39.903749423+01:00|rbac:role:admin"
echo

# Create a new user
echo "4. Creating a new user (no temporal handling required):"
curl -s -X POST "$BASE_URL/users/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "roles": ["user"]
  }' | jq
echo

# Login as new user
echo "5. Login as new user (temporal system transparent):"
NEW_LOGIN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass123"}')
echo "$NEW_LOGIN" | jq
NEW_TOKEN=$(echo "$NEW_LOGIN" | jq -r '.token')
echo

# Use the token
echo "6. Using auth token (temporal system handles session tracking):"
curl -s -X GET "$BASE_URL/entities/list" \
  -H "Authorization: Bearer $NEW_TOKEN" | jq -r '.[0].id' | head -5
echo

echo "=== Key Points ==="
echo "• Authentication works EXACTLY as before"
echo "• No temporal handling needed in auth requests"
echo "• Login: Send username/password, get token"
echo "• Use token: Bearer token in Authorization header"
echo "• Temporal tracking happens automatically:"
echo "  - User entities have timestamped tags"
echo "  - Sessions tracked temporally"
echo "  - Audit trail maintained automatically"
echo
echo "The temporal system is INVISIBLE to authentication!"