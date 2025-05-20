#!/bin/bash

# Get the entity ID as a parameter
if [ -z "$1" ]; then
  echo "Usage: $0 <entity_id>"
  exit 1
fi

ENTITY_ID=$1
PORT=8085
PROTOCOL="https"

# Login to get a session token
echo "Logging in as admin..."
SESSION_TOKEN=$(curl -s -k -X POST "$PROTOCOL://localhost:$PORT/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

if [ -z "$SESSION_TOKEN" ] || [ "$SESSION_TOKEN" == "null" ]; then
  echo "Failed to login or get session token."
  exit 1
fi

# Get the entity
echo "Retrieving entity with ID: $ENTITY_ID"
curl -s -k "$PROTOCOL://localhost:$PORT/api/v1/entities/get?id=$ENTITY_ID" \
  -H "Authorization: Bearer $SESSION_TOKEN" | tee /tmp/entity.json

# Extract base64 content
CONTENT=$(jq -r '.content' /tmp/entity.json)

echo -e "\nBase64 content:"
echo "$CONTENT"

# Decode the content
echo -e "\nDecoded content:"
echo "$CONTENT" | base64 -d

# Try to decode it further if it's JSON encoded
echo -e "\nFurther decoding (if applicable):"
DECODED=$(echo "$CONTENT" | base64 -d)
if echo "$DECODED" | grep -q "application/octet-stream"; then
  # Extract the actual content value from the JSON structure
  echo "$DECODED" | jq -r '."application/octet-stream"'
else
  # If it's not in the expected format, just try to parse as JSON
  echo "$DECODED" | jq . 2>/dev/null || echo "Not JSON content"
fi

echo -e "\nTags:"
jq -r '.tags[]' /tmp/entity.json