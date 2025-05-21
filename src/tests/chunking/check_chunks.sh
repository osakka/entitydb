#!/bin/bash
# Simple script to check if our chunk retrieval fix is working

echo "Checking chunk retrieval functionality"

# Get authentication token
echo "Logging in..."
LOGIN_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | sed 's/"token":"//;s/"//')

if [ -z "$TOKEN" ]; then
  echo "Failed to get token"
  exit 1
fi

echo "Got token: ${TOKEN:0:10}..."

# List all entities and check for chunked ones
echo "Listing all entities..."
ENTITIES=$(curl -s -k -X GET "https://localhost:8085/api/v1/entities/list" \
  -H "Authorization: Bearer $TOKEN")

# Save entities list for inspection
echo $ENTITIES > /tmp/entities_list.json

# Look for chunked entities
echo "Looking for chunked entities..."
CHUNK_PATTERN='"content:chunks:[0-9]'
CHUNKED_ENTITIES=$(echo $ENTITIES | grep -o "$CHUNK_PATTERN")

if [ -z "$CHUNKED_ENTITIES" ]; then
  echo "No chunked entities found. Creating a new one..."
  
  # Create a test file
  TEST_FILE="/tmp/chunk_test.bin"
  dd if=/dev/urandom of=$TEST_FILE bs=1M count=5 &> /dev/null
  echo "START_TEST_MARKER" > /tmp/test_marker.txt
  cat /tmp/test_marker.txt $TEST_FILE > /tmp/test_with_marker.bin
  mv /tmp/test_with_marker.bin $TEST_FILE
  
  # Create entity
  BASE64_CONTENT=$(base64 $TEST_FILE | tr -d '\n')
  
  echo "{\"tags\":[\"type:test\",\"test:chunk-test\"],\"content\":\"$BASE64_CONTENT\"}" > /tmp/create_entity.json
  
  CREATE_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    --data @/tmp/create_entity.json)
  
  ENTITY_ID=$(echo $CREATE_RESPONSE | grep -o '"id":"[^"]*"' | sed 's/"id":"//;s/"//')
  
  if [ -z "$ENTITY_ID" ]; then
    echo "Failed to create chunked entity"
    exit 1
  fi
  
  echo "Created entity with ID: $ENTITY_ID"
else
  echo "Found chunked entities: $CHUNKED_ENTITIES"
  # Extract an entity ID that has chunks
  ENTITY_ID=$(echo $ENTITIES | grep -o '"id":"[^"]*","tags":\[.*content:chunks:[0-9]' | head -1 | sed 's/"id":"//;s/","tags.*//')
  echo "Using chunked entity with ID: $ENTITY_ID"
fi

# Retrieve entity with content
echo "Retrieving entity $ENTITY_ID with content..."
CONTENT_RESPONSE=$(curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_content=true" \
  -H "Authorization: Bearer $TOKEN")

# Save response for inspection
echo $CONTENT_RESPONSE > /tmp/content_response.json

# Extract content length
CONTENT_LENGTH=$(echo $CONTENT_RESPONSE | grep -o '"content":"[^"]*"' | wc -c)
echo "Content response length: $CONTENT_LENGTH bytes"

# Check if content field is actually present
if echo $CONTENT_RESPONSE | grep -q '"content":"'; then
  echo "Content field is present in response"
  # Extract a sample of the content
  CONTENT_SAMPLE=$(echo $CONTENT_RESPONSE | grep -o '"content":"[^"]*"' | sed 's/"content":"//;s/"//' | head -c 20)
  echo "Content sample (first 20 chars): $CONTENT_SAMPLE"
else
  echo "Content field is MISSING from response"
fi

# Check the code to ensure it's actually working
echo "Checking if our chunking fix is in the code..."
if grep -q "HandleChunkedContent" /opt/entitydb/src/api/entity_handler.go; then
  echo "HandleChunkedContent function is being called"
else
  echo "HandleChunkedContent function is NOT being called"
fi

# Check if the entity_handler_fix.go file is being loaded
if ps -ef | grep entitydb | grep -v grep | grep -q "debug"; then
  echo "Server is running in debug mode, should be logging chunk retrieval"
else
  echo "Server is not running in debug mode"
fi

echo "Check completed!"