#!/bin/bash

set -e

echo "=== Testing Session Refresh Fix ==="
echo

# Login and get token
echo "1. Logging in as admin..."
LOGIN_RESPONSE=$(curl -k -s -X POST https://localhost:8085/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}')
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
EXPIRES_AT=$(echo "$LOGIN_RESPONSE" | jq -r '.expires_at')

echo "   Token: $TOKEN"
echo "   Expires: $EXPIRES_AT"
echo

# Wait a moment for indexing to complete
echo "2. Waiting 3 seconds for session indexing..."
sleep 3
echo

# Test session refresh
echo "3. Testing session refresh endpoint..."
REFRESH_RESPONSE=$(curl -k -s -X POST https://localhost:8085/api/v1/auth/refresh \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN")

echo "   Response: $REFRESH_RESPONSE"
echo

# Check if refresh was successful
if echo "$REFRESH_RESPONSE" | jq -e '.token' > /dev/null 2>&1; then
    NEW_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.token')
    NEW_EXPIRES=$(echo "$REFRESH_RESPONSE" | jq -r '.expires_at')
    echo "✅ Session refresh SUCCESS!"
    echo "   New token: $NEW_TOKEN"
    echo "   New expires: $NEW_EXPIRES"
    echo
    echo "4. Testing authenticated endpoint with new token..."
    WHO_RESPONSE=$(curl -k -s -X GET https://localhost:8085/api/v1/auth/whoami \
        -H "Authorization: Bearer $NEW_TOKEN")
    echo "   WhoAmI: $WHO_RESPONSE"
    
    if echo "$WHO_RESPONSE" | jq -e '.username' > /dev/null 2>&1; then
        echo "✅ New token works for authenticated endpoints!"
    else
        echo "❌ New token failed authentication"
    fi
else
    echo "❌ Session refresh FAILED!"
    echo "   Error response: $REFRESH_RESPONSE"
fi

echo
echo "=== Test Complete ==="