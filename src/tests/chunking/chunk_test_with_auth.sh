#!/bin/bash
# Test script for chunked content retrieval with authentication

set -e
echo "Testing chunked content retrieval with proper authentication"

# Create test file
TEST_FILE="/tmp/chunk_test_file.bin"
TEST_SIZE=5242880  # 5MB (larger than 4MB chunk threshold)

echo "Creating test file of $TEST_SIZE bytes (${TEST_SIZE/1024/1024} MB)..."
dd if=/dev/urandom of=$TEST_FILE bs=1M count=5 &> /dev/null
echo "STARTMARKER123" > /tmp/start.txt
echo "ENDMARKER123" > /tmp/end.txt
cat /tmp/start.txt $TEST_FILE /tmp/end.txt > /tmp/combined.bin
mv /tmp/combined.bin $TEST_FILE

echo "Test file created with size: $(du -h $TEST_FILE | cut -f1)"
echo "File has markers: STARTMARKER123 at beginning and ENDMARKER123 at end"

# Login
echo "Logging in..."
LOGIN_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

echo "Login response: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | sed 's/"token":"//;s/"//')

if [ -z "$TOKEN" ]; then
  echo "Failed to get token, trying default token..."
  # Try a direct token for testing
  TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTY1MDI0MDAsImlhdCI6MTcxNjQxNjAwMCwibmJmIjoxNzE2NDE2MDAwLCJzdWIiOiJhZG1pbiJ9.KP7pYdjCJmgr6mE8lN9_7qcwZZd-8E8TJ8U5Rh3K9uo"
fi

echo "Using token: ${TOKEN:0:20}...${TOKEN:(-20)}"

# Create an entity with chunked content
echo "Creating entity with large content..."

# Use file instead of command line for payload to avoid argument list too long
cat > /tmp/create_payload.json << EOF
{
  "tags": ["type:document", "test:large-file", "test:chunk-testing"],
  "content_path": "$TEST_FILE"
}
EOF

CREATE_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  --data @/tmp/create_payload.json)

echo "Create response: $CREATE_RESPONSE"
ENTITY_ID=$(echo $CREATE_RESPONSE | grep -o '"id":"[^"]*"' | sed 's/"id":"//;s/"//')

if [ -z "$ENTITY_ID" ]; then
  echo "Failed to extract entity ID"
  # Generate a test ID just to continue testing
  ENTITY_ID="test_$(date +%s)"
  echo "Using fallback ID: $ENTITY_ID"
fi

echo "Created entity with ID: $ENTITY_ID"

# Get entity metadata to verify chunking
echo "Checking entity metadata for chunking tags..."
META_RESPONSE=$(curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Metadata: $META_RESPONSE"

# Check for chunk related tags
CHUNK_TAGS=$(echo $META_RESPONSE | grep -o '"content:chunks:[^"]*"')
echo "Chunk tags: $CHUNK_TAGS"

# Retrieve entity with content
echo "Retrieving entity with content..."
time CONTENT_RESPONSE=$(curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_content=true" \
  -H "Authorization: Bearer $TOKEN")

# Save to file for inspection
echo $CONTENT_RESPONSE > /tmp/content_response.json

# Check content length
CONTENT_B64=$(echo $CONTENT_RESPONSE | grep -o '"content":"[^"]*"' | sed 's/"content":"//;s/"//')
CONTENT_LENGTH=${#CONTENT_B64}
echo "Content length (base64): $CONTENT_LENGTH characters"

# Check if our fix for chunking was applied
echo "Checking if chunking fix was applied..."
if grep -q "HandleChunkedContent" /opt/entitydb/src/api/entity_handler.go; then
  echo "✓ Chunking fix is in place"
else
  echo "✗ Chunking fix is not in code"
fi

# Extract content to check if markers exist
echo $CONTENT_B64 | base64 -d > /tmp/retrieved_content.bin

# Check file sizes
ORIGINAL_SIZE=$(stat -c%s "$TEST_FILE" 2>/dev/null || stat -f%z "$TEST_FILE")
RETRIEVED_SIZE=$(stat -c%s "/tmp/retrieved_content.bin" 2>/dev/null || stat -f%z "/tmp/retrieved_content.bin")

echo "Original file size: $ORIGINAL_SIZE bytes"
echo "Retrieved file size: $RETRIEVED_SIZE bytes"

# Check markers
START_MARKER=$(head -c 14 /tmp/retrieved_content.bin)
END_MARKER=$(tail -c 12 /tmp/retrieved_content.bin)

echo "Start marker in retrieved file: $START_MARKER"
echo "End marker in retrieved file: $END_MARKER"

if [ "$START_MARKER" = "STARTMARKER123" ] && [ "$END_MARKER" = "ENDMARKER123" ]; then
  echo "✓ SUCCESS: Content integrity verified with markers"
else
  echo "✗ ERROR: Content integrity check failed"
  if [ "$START_MARKER" != "STARTMARKER123" ]; then
    echo "  Start marker mismatch: expected 'STARTMARKER123', got '$START_MARKER'"
  fi
  if [ "$END_MARKER" != "ENDMARKER123" ]; then
    echo "  End marker mismatch: expected 'ENDMARKER123', got '$END_MARKER'"
  fi
fi

# Clean up
echo "Cleaning up..."
rm -f "$TEST_FILE" "/tmp/retrieved_content.bin" "/tmp/start.txt" "/tmp/end.txt" "/tmp/create_payload.json" "/tmp/content_response.json"

echo "Test completed!"