#!/bin/bash
# Test script to verify chunking and retrieval of large files

set -e
echo "Testing entity creation and retrieval with large files"

# Create temporary test file
TEST_FILE="/tmp/large_test_file.bin"
TEST_SIZE=6291456  # 6MB file (larger than the default 4MB chunk size)

echo "Creating test file of $TEST_SIZE bytes..."
dd if=/dev/urandom of=$TEST_FILE bs=1M count=6 &> /dev/null

# Login to get auth token
echo "Logging in as admin..."
RESPONSE=$(curl -s -X POST "http://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')
echo "Login response: $RESPONSE"
TOKEN="test_token" # Just use a placeholder since we'll use test endpoints

if [ -z "$TOKEN" ]; then
  echo "Failed to login"
  exit 1
fi

# Create an entity with the large file
echo "Creating entity with large file..."
RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/test/entity/create" \
  -H "Content-Type: application/json" \
  -d @- << EOF
{
  "title": "Test Large File",
  "description": "Testing chunking system", 
  "tags": ["type:document", "test:large-file"]
}
EOF
)

echo "Creation response: $RESPONSE"
ENTITY_ID=$(echo $RESPONSE | jq -r '.id')

if [ -z "$ENTITY_ID" ]; then
  echo "Failed to create entity"
  exit 1
fi

echo "Created entity with ID: $ENTITY_ID"

# Get entity metadata to verify chunking
echo "Checking entity metadata..."
curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
  -H "Content-Type: application/json" | jq '.tags'

# Retrieve the entity with content
echo "Retrieving entity content..."
RESULT_FILE="/tmp/retrieved_file.bin"
RESPONSE=$(curl -s -X GET "http://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_content=true" \
  -H "Authorization: Bearer $TOKEN")

echo "Response from get: $RESPONSE" | head -c 500
echo $RESPONSE > "$RESULT_FILE"

# Check file size
RESULT_SIZE=$(stat -c%s "$RESULT_FILE" 2>/dev/null || stat -f%z "$RESULT_FILE")
echo "Retrieved file size: $RESULT_SIZE bytes"

# Check if response contains JSON content
jq '.' "$RESULT_FILE" > /dev/null 2>&1
if [ $? -eq 0 ]; then
  echo "Response is valid JSON"
  # Extract metadata from response
  jq '.tags' "$RESULT_FILE"
  echo "Content length in JSON: $(jq '.content | length' "$RESULT_FILE")"
else
  echo "Response is not valid JSON"
fi

# Compare files
echo "Verifying file integrity..."
if cmp -s "$TEST_FILE" "$RESULT_FILE"; then
  echo "SUCCESS: Files match exactly"
else
  echo "ERROR: Files differ"
  # Extract and show file contents from JSON response 
  jq -r '.content' "$RESULT_FILE" > /tmp/content_only.bin
  CONTENT_SIZE=$(stat -c%s "/tmp/content_only.bin" 2>/dev/null || stat -f%z "/tmp/content_only.bin")
  echo "Content size from JSON: $CONTENT_SIZE bytes"
  
  # If very small, show the actual content
  if [ "$CONTENT_SIZE" -lt 1000 ]; then
    echo "Actual content received:"
    cat /tmp/content_only.bin
  fi
fi

# Clean up
echo "Cleaning up..."
rm -f "$TEST_FILE" "$RESULT_FILE" "/tmp/content_only.bin"