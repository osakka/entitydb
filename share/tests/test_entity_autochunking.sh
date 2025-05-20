#!/bin/bash

# Test autochunking functionality with the new entity model

BASE_URL="https://localhost:8085"
echo "Testing entity autochunking functionality..."

# Login as admin
echo -e "\n=== Login as admin ==="
LOGIN_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}')

ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r .token)

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo "Failed to get admin token"
    exit 1
fi

echo "Admin token: $ADMIN_TOKEN"

# Create a small file entity (should not be chunked)
echo -e "\n=== Creating entity with small content ==="
SMALL_CONTENT=$(python3 -c "print('x' * 1000)")
SMALL_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
  "tags": ["type:document", "name:small.txt"],
  "content": "$SMALL_CONTENT"
}
EOF
)

SMALL_ID=$(echo "$SMALL_RESPONSE" | jq -r .id)
echo "Created small entity: $SMALL_ID"

# Create a large file entity (should be auto-chunked)
echo -e "\n=== Creating entity with large content (5MB) ==="
# Generate 5MB of content
LARGE_CONTENT=$(python3 -c "print('x' * 5242880)")
LARGE_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
  "tags": ["type:document", "name:large.txt"],
  "content": "$LARGE_CONTENT"
}
EOF
)

LARGE_ID=$(echo "$LARGE_RESPONSE" | jq -r .id)
echo "Created large entity: $LARGE_ID"

# Retrieve the small entity
echo -e "\n=== Retrieving small entity ==="
SMALL_GET=$(curl -sk -X GET "$BASE_URL/api/v1/entities/get?id=$SMALL_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "Small entity tags:"
echo "$SMALL_GET" | jq '.tags'
echo "Small entity content length: $(echo "$SMALL_GET" | jq '.content | length')"

# Retrieve the large entity  
echo -e "\n=== Retrieving large entity ==="
LARGE_GET=$(curl -sk -X GET "$BASE_URL/api/v1/entities/get?id=$LARGE_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "Large entity tags:"
echo "$LARGE_GET" | jq '.tags'
echo "Large entity content length: $(echo "$LARGE_GET" | jq '.content | length')"

# Check for chunk metadata
echo -e "\n=== Checking for chunk metadata ==="
if echo "$LARGE_GET" | jq -r '.tags[]' | grep -q "content:chunks:"; then
    echo "✓ Large entity has chunk metadata"
    echo "$LARGE_GET" | jq -r '.tags[]' | grep -E "content:(chunks|checksum|size):"
else
    echo "✗ Large entity missing chunk metadata"
fi

# Test entity with binary content
echo -e "\n=== Creating entity with binary content ==="
# Create a small binary file
BINARY_CONTENT=$(echo -n "Binary data test" | base64)
BINARY_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
  "tags": ["type:binary", "name:test.bin", "content:encoding:base64"],
  "content": "$BINARY_CONTENT"
}
EOF
)

BINARY_ID=$(echo "$BINARY_RESPONSE" | jq -r .id)
echo "Created binary entity: $BINARY_ID"

echo -e "\n=== Test summary ==="
echo "Small entity: $SMALL_ID (1KB)"
echo "Large entity: $LARGE_ID (5MB, should be chunked)"
echo "Binary entity: $BINARY_ID"

echo -e "\n✅ Autochunking test complete"