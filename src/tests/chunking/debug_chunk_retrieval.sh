#!/bin/bash
# Simpler test script to debug chunking retrieval

set -e
echo "Testing chunk retrieval with a simple approach"

# Create a smaller test file for quicker testing
TEST_FILE="/tmp/test_chunk.bin"
TEST_SIZE=5242880  # 5MB file (larger than the default 4MB chunk size)

echo "Creating test file of $TEST_SIZE bytes..."
dd if=/dev/urandom of=$TEST_FILE bs=1M count=5 &> /dev/null

# Login
echo "Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

echo "Login response: $LOGIN_RESPONSE"
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
  echo "Failed to login"
  exit 1
fi

echo "Got token: $TOKEN"

# Create entity with content (base64-encoded)
echo "Creating entity with large file..."
BASE64_CONTENT=$(base64 $TEST_FILE | tr -d '\n')
REQUEST="{\"tags\":[\"type:document\",\"test:large-file\"],\"content\":\"$BASE64_CONTENT\"}"

# Save request to a file to avoid command line length issues
echo "$REQUEST" > /tmp/create_request.json
echo "Request file size: $(stat -c%s /tmp/create_request.json 2>/dev/null || stat -f%z /tmp/create_request.json) bytes"

CREATE_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @/tmp/create_request.json)

echo "Creation response status: $?"
echo "Creation response (truncated): $(echo "$CREATE_RESPONSE" | head -c 200)"

ENTITY_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')

if [ -z "$ENTITY_ID" ] || [ "$ENTITY_ID" == "null" ]; then
  echo "Failed to create entity"
  exit 1
fi

echo "Created entity with ID: $ENTITY_ID"

# Check entity tags to verify chunking
echo "Checking entity metadata..."
META_RESPONSE=$(curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Metadata response: $META_RESPONSE"
echo "Tags: $(echo "$META_RESPONSE" | jq '.tags')"

# Now retrieve the entity with content
echo "Retrieving entity content..."
CONTENT_RESPONSE=$(curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_content=true" \
  -H "Authorization: Bearer $TOKEN")

echo "Content response size: $(echo "$CONTENT_RESPONSE" | wc -c) bytes"
echo "Content response tags: $(echo "$CONTENT_RESPONSE" | jq '.tags')"
echo "Content length in response: $(echo "$CONTENT_RESPONSE" | jq '.content | length')"

# Extract and decode content to check if we got all data
echo "Extracting and checking content..."
echo "$CONTENT_RESPONSE" | jq -r '.content' | base64 -d > /tmp/retrieved_content.bin 2>/dev/null

# Check sizes
ORIGINAL_SIZE=$(stat -c%s "$TEST_FILE" 2>/dev/null || stat -f%z "$TEST_FILE")
RETRIEVED_SIZE=$(stat -c%s "/tmp/retrieved_content.bin" 2>/dev/null || stat -f%z "/tmp/retrieved_content.bin")

echo "Original file size: $ORIGINAL_SIZE bytes"
echo "Retrieved file size: $RETRIEVED_SIZE bytes"

# Compare files if we got any content
if [ $RETRIEVED_SIZE -gt 0 ]; then
  echo "Comparing files..."
  if cmp -s "$TEST_FILE" "/tmp/retrieved_content.bin"; then
    echo "SUCCESS: Files match exactly"
  else
    echo "ERROR: Files differ"
    # Show file hashes
    echo "Original file hash: $(sha256sum $TEST_FILE | cut -d' ' -f1)"
    echo "Retrieved file hash: $(sha256sum /tmp/retrieved_content.bin | cut -d' ' -f1)"
  fi
else
  echo "ERROR: No content was retrieved"
fi

# Clean up
echo "Cleaning up..."
rm -f "$TEST_FILE" "/tmp/retrieved_content.bin" "/tmp/create_request.json"