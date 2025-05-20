#!/bin/bash
set -e

echo "Checking how timestamps are handled in turbo mode..."

# Login first
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username": "admin", "password": "admin"}' | \
    jq -r '.token')

echo "Token: $TOKEN"

# Create an entity with timestamps in tags
echo -e "\n1. Creating entity with temporal tags..."
ENTITY_ID=$(curl -s -X POST http://localhost:8085/api/v1/entities/create \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "tags": [
            "1732018871547277568|type:test",
            "1732018871547277568|turbo:check",
            "1732018871547277568|timestamp:visible"
        ],
        "content": [
            {
                "type": "test_data",
                "value": "Testing timestamp visibility",
                "timestamp": "2024-11-19T10:00:00Z"
            }
        ]
    }' | jq -r '.id')

echo "Created entity: $ENTITY_ID"

# Get entity normally (without include_timestamps)
echo -e "\n2. Getting entity normally (no include_timestamps)..."
curl -s -X GET "http://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
    -H "Authorization: Bearer $TOKEN" | jq

# Get entity with timestamps
echo -e "\n3. Getting entity with include_timestamps=true..."
curl -s -X GET "http://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_timestamps=true" \
    -H "Authorization: Bearer $TOKEN" | jq

# List by tag without timestamps
echo -e "\n4. List by tag (no include_timestamps)..."
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tag=turbo:check" \
    -H "Authorization: Bearer $TOKEN" | jq

# List by tag with timestamps
echo -e "\n5. List by tag with include_timestamps=true..."
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tag=turbo:check&include_timestamps=true" \
    -H "Authorization: Bearer $TOKEN" | jq

# Check temporal features - as-of query
echo -e "\n6. Testing temporal as-of query..."
curl -s -X GET "http://localhost:8085/api/v1/entities/as-of?id=$ENTITY_ID&timestamp=2024-11-19T10:00:00Z" \
    -H "Authorization: Bearer $TOKEN" | jq