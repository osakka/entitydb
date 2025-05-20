#!/bin/bash

# Set the correct port and protocol
PORT=8085
PROTOCOL="https"

echo "Testing EntityDB Content Encoding with different types"
echo "===================================================="

# Login to get a session token
echo "Logging in as admin..."
SESSION_TOKEN=$(curl -s -k -X POST "$PROTOCOL://localhost:$PORT/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

if [ -z "$SESSION_TOKEN" ] || [ "$SESSION_TOKEN" == "null" ]; then
  echo "Failed to login or get session token. Check if the server is running."
  exit 1
fi

echo "Obtained session token: ${SESSION_TOKEN:0:10}..."

# Create entity with string content
echo -e "\n1. Creating entity with simple string content..."
curl -s -k -X POST "$PROTOCOL://localhost:$PORT/api/v1/entities/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -d '{
    "content": "This is a simple string content",
    "tags": ["type:test", "content:string"]
  }' | tee /tmp/entity1.json

ENTITY1_ID=$(jq -r '.id' /tmp/entity1.json)
echo "Entity created with ID: $ENTITY1_ID"

# Create entity with JSON content
echo -e "\n2. Creating entity with JSON content..."
curl -s -k -X POST "$PROTOCOL://localhost:$PORT/api/v1/entities/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -d '{
    "content": {
      "name": "Test Object",
      "values": [1, 2, 3],
      "nested": {
        "property": "value",
        "boolean": true
      }
    },
    "tags": ["type:test", "content:json"]
  }' | tee /tmp/entity2.json

ENTITY2_ID=$(jq -r '.id' /tmp/entity2.json)
echo "Entity created with ID: $ENTITY2_ID"

# Retrieving entities
echo -e "\n3. Retrieving entity with string content..."
curl -s -k "$PROTOCOL://localhost:$PORT/api/v1/entities/get?id=$ENTITY1_ID" \
  -H "Authorization: Bearer $SESSION_TOKEN" | jq

echo -e "\n4. Retrieving entity with JSON content..."
curl -s -k "$PROTOCOL://localhost:$PORT/api/v1/entities/get?id=$ENTITY2_ID" \
  -H "Authorization: Bearer $SESSION_TOKEN" | jq

# Testing querying by tags
echo -e "\n5. Testing query for entities with content:string tag..."
curl -s -k "$PROTOCOL://localhost:$PORT/api/v1/entities/query?tags=content:string" \
  -H "Authorization: Bearer $SESSION_TOKEN" | jq

echo -e "\n6. Testing query for entities with content:json tag..."
curl -s -k "$PROTOCOL://localhost:$PORT/api/v1/entities/query?tags=content:json" \
  -H "Authorization: Bearer $SESSION_TOKEN" | jq

echo -e "\n7. Testing query for all test entities..."
curl -s -k "$PROTOCOL://localhost:$PORT/api/v1/entities/query?tags=type:test" \
  -H "Authorization: Bearer $SESSION_TOKEN" | jq

echo -e "\nAll tests completed."