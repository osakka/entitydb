#!/bin/bash

# Create test user for API testing
API_BASE="http://localhost:8085/api/v1"

# Login as admin to create new user
echo "Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -X POST "${API_BASE}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "entity_user_admin", "password": "admin123"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "Failed to login as admin. Response: $LOGIN_RESPONSE"
  # Try different credentials
  LOGIN_RESPONSE=$(curl -s -X POST "${API_BASE}/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username": "admin", "password": "changeme"}')
  
  TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
  if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "Failed with second attempt. Response: $LOGIN_RESPONSE"
    # Try test endpoint without auth
    echo "Testing without auth..."
    RESPONSE=$(curl -s -X GET "${API_BASE}/test/public")
    echo "Public endpoint response: $RESPONSE"
    exit 1
  fi
fi

echo "Got token: ${TOKEN:0:20}..."
AUTH_HEADER="Authorization: Bearer $TOKEN"

# Create a test user
echo "Creating test user..."
CREATE_RESPONSE=$(curl -s -X POST "${API_BASE}/users/create" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "roles": ["user"]
  }')

echo "User creation response: $CREATE_RESPONSE"