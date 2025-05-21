#!/bin/bash
# Test script specifically for the streaming approach

echo "STREAMING TEST: Testing chunked content retrieval via streaming API"

# Step 1: Get an authentication token
echo "Step 1: Getting authentication token..."
LOGIN_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | sed 's/"token":"//;s/"//')

if [ -z "$TOKEN" ]; then
  echo "Failed to login, using default admin token"
  TOKEN="token_default_admin_token"
fi

echo "Got token: ${TOKEN:0:15}..."

# Step 2: Create a test file with markers
echo "Step 2: Creating test file with markers..."
TEST_FILE="/tmp/streaming_test.bin"
dd if=/dev/urandom of=$TEST_FILE bs=1M count=6 &> /dev/null
echo "STREAMING_START_MARKER" > /tmp/streaming_start.txt
echo "STREAMING_END_MARKER" > /tmp/streaming_end.txt
cat /tmp/streaming_start.txt $TEST_FILE /tmp/streaming_end.txt > /tmp/marked_file.bin
mv /tmp/marked_file.bin $TEST_FILE

TEST_SIZE=$(stat -c%s "$TEST_FILE" 2>/dev/null || stat -f%z "$TEST_FILE")
echo "Created test file with size: $(du -h $TEST_FILE | cut -f1) ($TEST_SIZE bytes)"

# Step 3: Create an entity with large file content
echo "Step 3: Creating entity with large content..."
BASE64_CONTENT=$(base64 $TEST_FILE | tr -d '\n')

JSON_DATA="{\"tags\":[\"type:test\",\"test:streaming\"],\"content\":\"$BASE64_CONTENT\"}"
echo "$JSON_DATA" > /tmp/streaming_create.json

CREATE_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  --data @/tmp/streaming_create.json)

ENTITY_ID=$(echo $CREATE_RESPONSE | grep -o '"id":"[^"]*"' | sed 's/"id":"//;s/"//')

if [ -z "$ENTITY_ID" ]; then
  echo "Failed to create entity, using default ID"
  ENTITY_ID="test_streaming_$(date +%s)"
fi

echo "Created entity with ID: $ENTITY_ID"

# Step 4: Verify chunking
echo "Step 4: Verifying entity is chunked..."
META_RESPONSE=$(curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")

CHUNK_TAGS=$(echo "$META_RESPONSE" | grep -o '"content:chunks:[^"]*"')
echo "Chunk tags: $CHUNK_TAGS"

if [ -z "$CHUNK_TAGS" ]; then
  echo "Entity doesn't have chunk tags, test may not work properly"
fi

# Step 5: Test streaming endpoint
echo "Step 5: Testing streaming endpoint..."
time curl -s -k -X GET "https://localhost:8085/api/v1/entities/stream?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -o /tmp/streamed_content.bin

# Check sizes
STREAMED_SIZE=$(stat -c%s "/tmp/streamed_content.bin" 2>/dev/null || stat -f%z "/tmp/streamed_content.bin")
echo "Original size: $TEST_SIZE bytes"
echo "Streamed size: $STREAMED_SIZE bytes"

# Check markers
START_MARKER=$(head -c 21 /tmp/streamed_content.bin)
END_MARKER=$(tail -c 19 /tmp/streamed_content.bin)

echo "Start marker: $START_MARKER"
echo "End marker: $END_MARKER"

if [ "$START_MARKER" = "STREAMING_START_MARKER" ] && [ "$END_MARKER" = "STREAMING_END_MARKER" ]; then
  echo "✅ MARKERS CHECK PASSED: Content integrity verified"
else
  echo "❌ MARKERS CHECK FAILED: Content corruption detected"
fi

if [ "$TEST_SIZE" -eq "$STREAMED_SIZE" ]; then
  echo "✅ SIZE CHECK PASSED: Retrieved content has correct size"
else 
  echo "❌ SIZE CHECK FAILED: Size mismatch - original: $TEST_SIZE, streamed: $STREAMED_SIZE"
fi

# Step 6: Test the dedicated chunks endpoint
echo "Step 6: Testing chunks endpoint..."
time curl -s -k -X GET "https://localhost:8085/api/v1/entities/chunks?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -o /tmp/chunked_content.bin

# Check chunked content
CHUNKED_SIZE=$(stat -c%s "/tmp/chunked_content.bin" 2>/dev/null || stat -f%z "/tmp/chunked_content.bin")
echo "Chunks size: $CHUNKED_SIZE bytes"

START_MARKER_CHUNKS=$(head -c 21 /tmp/chunked_content.bin)
END_MARKER_CHUNKS=$(tail -c 19 /tmp/chunked_content.bin)

echo "Start marker (chunks): $START_MARKER_CHUNKS"
echo "End marker (chunks): $END_MARKER_CHUNKS"

if [ "$START_MARKER_CHUNKS" = "STREAMING_START_MARKER" ] && [ "$END_MARKER_CHUNKS" = "STREAMING_END_MARKER" ]; then
  echo "✅ CHUNKS MARKERS CHECK PASSED: Content integrity verified"
else
  echo "❌ CHUNKS MARKERS CHECK FAILED: Content corruption detected"
fi

# Overall result
if [ "$START_MARKER" = "STREAMING_START_MARKER" ] && [ "$END_MARKER" = "STREAMING_END_MARKER" ] && [ "$TEST_SIZE" -eq "$STREAMED_SIZE" ]; then
  echo "✅ STREAMING TEST PASSED: Chunked content streaming works!"
else
  echo "❌ STREAMING TEST FAILED: Chunked content streaming has issues"
fi

# Clean up
echo "Cleaning up..."
rm -f "$TEST_FILE" "/tmp/streamed_content.bin" "/tmp/chunked_content.bin" "/tmp/streaming_start.txt" "/tmp/streaming_end.txt" "/tmp/streaming_create.json"