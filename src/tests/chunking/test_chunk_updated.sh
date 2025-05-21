#!/bin/bash
# Test script to verify chunking and retrieval of large files

set -e
echo "Testing entity creation and retrieval with large files"

# Create temporary test file
TEST_FILE="/tmp/large_test_file.bin"
TEST_SIZE=6291456  # 6MB file (larger than the default 4MB chunk size)

echo "Creating test file of $TEST_SIZE bytes..."
dd if=/dev/urandom of=$TEST_FILE bs=1M count=6 &> /dev/null

# Step 1: Create an entity with test endpoint (no auth required)
echo "Creating entity with test endpoint..."
CREATE_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/test/entity/create" \
  -H "Content-Type: application/json" \
  -d @- << EOF
{
  "title": "Test Large File",
  "description": "Testing chunking system", 
  "tags": ["type:document", "test:large-file"]
}
EOF
)

echo "Creation response: $CREATE_RESPONSE"
ENTITY_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')

if [ -z "$ENTITY_ID" ] || [ "$ENTITY_ID" == "null" ]; then
  echo "Failed to create entity"
  exit 1
fi

echo "Created entity with ID: $ENTITY_ID"

# Step 2: Update entity with large content to trigger chunking
echo "Updating entity with large content..."
UPDATE_RESPONSE=$(curl -s -k -X PUT "https://localhost:8085/api/v1/entities/update?id=$ENTITY_ID" \
  -H "Content-Type: application/json" \
  -d @- << EOF
{
  "id": "$ENTITY_ID",
  "content": "$(base64 $TEST_FILE | tr -d '\n')"
}
EOF
)

echo "Update response status: $?"
echo "Update response (truncated): $(echo "$UPDATE_RESPONSE" | head -c 100)"

# Step 3: Get entity metadata to verify chunking
echo "Checking entity metadata for chunking tags..."
META_RESPONSE=$(curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
  -H "Content-Type: application/json")

echo "Entity tags:"
echo "$META_RESPONSE" | jq '.tags'

# Check for chunk-related tags
CHUNKS_TAG=$(echo "$META_RESPONSE" | jq -r '.tags[] | select(contains("content:chunks:"))' | tail -1)
CHUNK_SIZE_TAG=$(echo "$META_RESPONSE" | jq -r '.tags[] | select(contains("content:chunk-size:"))' | tail -1)

echo "Chunks tag: $CHUNKS_TAG"
echo "Chunk size tag: $CHUNK_SIZE_TAG"

if [ -z "$CHUNKS_TAG" ]; then
  echo "ERROR: Entity was not chunked as expected"
  exit 1
fi

# Step 4: Retrieve the entity with content
echo "Retrieving entity with content..."
CONTENT_RESPONSE=$(curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_content=true" \
  -H "Content-Type: application/json")

# Extract and save the content
echo "$CONTENT_RESPONSE" | jq -r '.content' | base64 -d > /tmp/retrieved_file.bin 2>/dev/null

# Check file size
ORIGINAL_SIZE=$(stat -c%s "$TEST_FILE" 2>/dev/null || stat -f%z "$TEST_FILE")
RESULT_SIZE=$(stat -c%s "/tmp/retrieved_file.bin" 2>/dev/null || stat -f%z "/tmp/retrieved_file.bin")

echo "Original file size: $ORIGINAL_SIZE bytes"
echo "Retrieved file size: $RESULT_SIZE bytes"

# Compare files
echo "Verifying file integrity..."
if [ "$ORIGINAL_SIZE" -eq "$RESULT_SIZE" ]; then
  echo "Size match: Original and retrieved files have the same size"
  
  if cmp -s "$TEST_FILE" "/tmp/retrieved_file.bin"; then
    echo "SUCCESS: Files match exactly - chunking and reassembly work correctly!"
  else
    echo "ERROR: Files have different content despite having the same size"
    # Show file hashes
    echo "Original file hash: $(sha256sum $TEST_FILE | cut -d' ' -f1)"
    echo "Retrieved file hash: $(sha256sum /tmp/retrieved_file.bin | cut -d' ' -f1)"
  fi
else
  echo "ERROR: File size mismatch - chunking or reassembly is not working correctly"
  echo "Original: $ORIGINAL_SIZE bytes, Retrieved: $RESULT_SIZE bytes"
  
  if [ "$RESULT_SIZE" -lt 1000 ]; then
    echo "Actual content received (first 100 bytes):"
    hexdump -C "/tmp/retrieved_file.bin" | head -5
  fi
fi

# Show entity content length from API response
CONTENT_LENGTH=$(echo "$CONTENT_RESPONSE" | jq '.content | length')
echo "Content length reported in API response: $CONTENT_LENGTH"

# Clean up
echo "Cleaning up..."
rm -f "$TEST_FILE" "/tmp/retrieved_file.bin"