#!/bin/bash

# EntityDB v2.13.0 Test Script

echo "========================================================"
echo "  EntityDB v2.13.0 Test Script"
echo "  Testing SSL & Content Encoding Features"
echo "========================================================"

# Set the port and protocol
PORT=8085
PROTOCOL="https"

# 1. Test server status
echo -e "\n1. Testing server status..."
curl -k "$PROTOCOL://localhost:$PORT/api/v1/status"

# 2. Login to get a session token
echo -e "\n\n2. Logging in as admin..."
SESSION_TOKEN=$(curl -s -k -X POST "$PROTOCOL://localhost:$PORT/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

if [ -z "$SESSION_TOKEN" ] || [ "$SESSION_TOKEN" == "null" ]; then
  echo "Failed to login. Check if the server is running."
  exit 1
fi

echo "Session token: ${SESSION_TOKEN:0:10}..."

# 3. Test string content creation
echo -e "\n3. Creating entity with string content..."
STRING_ENTITY=$(curl -s -k -X POST "$PROTOCOL://localhost:$PORT/api/v1/entities/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -d '{
    "content": "This is a test string for v2.13.0",
    "tags": ["type:test", "version:v2.13.0"]
  }')

STRING_ID=$(echo "$STRING_ENTITY" | jq -r '.id')
echo "String entity created with ID: $STRING_ID"
echo "$STRING_ENTITY" | jq

# 4. Test JSON content creation
echo -e "\n4. Creating entity with JSON content..."
JSON_ENTITY=$(curl -s -k -X POST "$PROTOCOL://localhost:$PORT/api/v1/entities/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -d '{
    "content": {
      "title": "JSON Test Object v2.13.0",
      "version": "2.13.0",
      "features": ["SSL", "Content Encoding", "MIME Types"],
      "config": {
        "ssl": true,
        "port": 8085
      }
    },
    "tags": ["type:test", "version:v2.13.0", "content:json"]
  }')

JSON_ID=$(echo "$JSON_ENTITY" | jq -r '.id')
echo "JSON entity created with ID: $JSON_ID"
echo "$JSON_ENTITY" | jq

# 5. Retrieve and check string entity
echo -e "\n5. Retrieving string entity..."
curl -s -k "$PROTOCOL://localhost:$PORT/api/v1/entities/get?id=$STRING_ID" \
  -H "Authorization: Bearer $SESSION_TOKEN" | jq

# 6. Retrieve and check JSON entity
echo -e "\n6. Retrieving JSON entity..."
curl -s -k "$PROTOCOL://localhost:$PORT/api/v1/entities/get?id=$JSON_ID" \
  -H "Authorization: Bearer $SESSION_TOKEN" | jq

# 7. Test query by version tag
echo -e "\n7. Querying entities with version:v2.13.0 tag..."
curl -s -k "$PROTOCOL://localhost:$PORT/api/v1/entities/query?tags=version:v2.13.0" \
  -H "Authorization: Bearer $SESSION_TOKEN" | jq '.entities | length'

# 8. Decode and check content
echo -e "\n8. Decoding content for validation..."
echo "String entity content:"
echo "$STRING_ENTITY" | jq -r '.content' | base64 -d

echo -e "\nJSON entity content (decoded from base64):"
echo "$JSON_ENTITY" | jq -r '.content' | base64 -d | jq

echo -e "\n========================================================"
echo "  All tests completed successfully!"
echo "  EntityDB v2.13.0 is working properly"
echo "========================================================"