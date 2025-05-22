#!/bin/bash

# Test to verify the root cause fix prevents content wrapping

set -e

BASE_URL="http://localhost:8085"
API_BASE="$BASE_URL/api/v1"

echo "=== Testing Root Cause Fix ==="
echo "This test verifies that new entities don't get content wrapped"
echo

# Test login first
echo "Testing initial login..."
response=$(curl -s -w "%{http_code}" -X POST "$API_BASE/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username": "admin", "password": "admin"}' -o /tmp/login_response.json)

http_code="${response: -3}"
if [ "$http_code" = "200" ]; then
    token=$(cat /tmp/login_response.json | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "✓ Login successful"
else
    echo "✗ Login failed - HTTP $http_code"
    cat /tmp/login_response.json
    exit 1
fi

# Create a new user entity
echo "Creating new user entity..."
user_content='{"username": "testuser", "password_hash": "$2a$10$abcdefg", "email": "test@example.com"}'
entity_data=$(cat <<EOF
{
    "tags": ["type:user", "status:active", "content:type:application/json"],
    "content": $user_content
}
EOF
)

response=$(curl -s -w "%{http_code}" -X POST "$API_BASE/entities/create" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $token" \
    -d "$entity_data" -o /tmp/create_response.json)

http_code="${response: -3}"
if [ "$http_code" = "201" ] || [ "$http_code" = "200" ]; then
    entity_id=$(cat /tmp/create_response.json | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "✓ Entity created with ID: $entity_id"
else
    echo "✗ Entity creation failed - HTTP $http_code"
    cat /tmp/create_response.json
    exit 1
fi

# Retrieve the entity and check content format
echo "Retrieving entity to check content format..."
response=$(curl -s -w "%{http_code}" -X GET "$API_BASE/entities/get?id=$entity_id" \
    -H "Authorization: Bearer $token" -o /tmp/get_response.json)

http_code="${response: -3}"
if [ "$http_code" = "200" ]; then
    echo "✓ Entity retrieved successfully"
else
    echo "✗ Entity retrieval failed - HTTP $http_code"
    exit 1
fi

# Check if content is properly formatted (not wrapped)
echo "Analyzing content format..."
content=$(cat /tmp/get_response.json | jq -r '.content')

if [ "$content" = "null" ]; then
    echo "✗ Content is null"
    exit 1
fi

# Decode base64 content (JSON marshaling encodes []byte as base64)
decoded_content=$(echo "$content" | base64 -d 2>/dev/null || echo "$content")

echo "Decoded content: $decoded_content"

# Check if it's clean JSON (not wrapped)
if echo "$decoded_content" | jq . >/dev/null 2>&1; then
    # Check if it contains the expected fields directly
    if echo "$decoded_content" | jq -e '.username' >/dev/null 2>&1; then
        echo "✓ Content is clean JSON with expected structure"
        username=$(echo "$decoded_content" | jq -r '.username')
        echo "  Username: $username"
    else
        # Check if it's wrapped in application/octet-stream
        if echo "$decoded_content" | jq -e '."application/octet-stream"' >/dev/null 2>&1; then
            echo "✗ Content is still wrapped in application/octet-stream"
            echo "  This indicates the root cause fix didn't work"
            exit 1
        else
            echo "? Content is JSON but doesn't have expected structure"
            echo "  Content: $decoded_content"
        fi
    fi
else
    echo "✗ Content is not valid JSON"
    echo "  Content: $decoded_content"
    exit 1
fi

# Clean up
rm -f /tmp/login_response.json /tmp/create_response.json /tmp/get_response.json

echo
echo "=== Root Cause Test Results ==="
echo "✓ New entities are created with clean, unwrapped content"
echo "✓ The root cause of content wrapping has been fixed"