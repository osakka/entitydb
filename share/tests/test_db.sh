#!/bin/bash

# Test the EntityDB after initialization

# Login as admin
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

echo "Token: $TOKEN"

# Create a test entity
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "issue",
    "title": "Test Issue",
    "tags": ["type:issue", "status:active"],
    "description": "This is a test issue created after db init"
  }'

echo ""

# List entities
echo "Listing entities:"
curl http://localhost:8085/api/v1/entities/list \
  -H "Authorization: Bearer $TOKEN" | jq