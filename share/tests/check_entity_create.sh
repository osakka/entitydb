#!/bin/bash
set -e

# Login
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username": "admin", "password": "admin"}' | \
    jq -r '.token')

echo "Token: $TOKEN"

# Create entity
ENTITY=$(curl -s -X POST http://localhost:8085/api/v1/entities/create \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:test", "simple:entity"],
        "content": []
    }' | jq)

echo "Created entity:"
echo "$ENTITY"

ENTITY_ID=$(echo "$ENTITY" | jq -r '.id')
echo "Entity ID: $ENTITY_ID"

# Try to get it
echo -e "\nGetting entity by ID:"
curl -s -X GET "http://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
    -H "Authorization: Bearer $TOKEN"

# List all entities with type:test
echo -e "\n\nListing entities by tag:"
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tag=type:test" \
    -H "Authorization: Bearer $TOKEN" | jq