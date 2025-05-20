#!/bin/bash
set -e

echo "Testing timestamp handling in turbo mode..."

# Login first
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username": "admin", "password": "admin"}' | \
    jq -r '.token')

echo "Token obtained"

# Create an entity with simple tags
echo -e "\n1. Creating entity with simple tags..."
RESPONSE=$(curl -s -X POST http://localhost:8085/api/v1/entities/create \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:test", "turbo:check", "feature:timestamps"],
        "content": []
    }')

echo "Create response: $RESPONSE"
ENTITY_ID=$(echo "$RESPONSE" | jq -r '.id')
echo "Created entity: $ENTITY_ID"

# Get entity without timestamps
echo -e "\n2. Getting entity (default - no timestamps)..."
curl -s -X GET "http://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
    -H "Authorization: Bearer $TOKEN" | jq '.tags'

# Get entity with timestamps
echo -e "\n3. Getting entity with include_timestamps=true..."
curl -s -X GET "http://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_timestamps=true" \
    -H "Authorization: Bearer $TOKEN" | jq '.tags'

# List by tag without timestamps
echo -e "\n4. List by tag (default - no timestamps)..."
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tag=type:test" \
    -H "Authorization: Bearer $TOKEN" | jq '.[0].tags' 2>/dev/null || echo "No results"

# List by tag with timestamps
echo -e "\n5. List by tag with include_timestamps=true..."
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tag=type:test&include_timestamps=true" \
    -H "Authorization: Bearer $TOKEN" | jq '.[0].tags' 2>/dev/null || echo "No results"

# Raw tag listing
echo -e "\n6. List raw tags..."
curl -s -X GET "http://localhost:8085/api/v1/entities/list-raw-tags" \
    -H "Authorization: Bearer $TOKEN" | jq | head -10