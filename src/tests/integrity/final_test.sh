#!/bin/bash
# Final test script for chunked content retrieval

echo "FINAL TEST: Verifying chunked content retrieval"

# Step 1: Get a fresh authentication token
echo "Step 1: Getting authentication token..."
LOGIN_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

echo "Login response: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | sed 's/"token":"//;s/"//')

if [ -z "$TOKEN" ]; then
  echo "Failed to get authentication token"
  exit 1
fi

echo "Got token: ${TOKEN:0:20}..."

# Step 2: Create a test file with markers
echo "Step 2: Creating test file with markers..."
TEST_FILE="/tmp/test_file_final.bin"
dd if=/dev/urandom of=$TEST_FILE bs=1M count=6 &> /dev/null
echo "START_MARKER_TEXT" > /tmp/marker1.txt
echo "END_MARKER_TEXT" > /tmp/marker2.txt
cat /tmp/marker1.txt $TEST_FILE /tmp/marker2.txt > /tmp/marked_file.bin
mv /tmp/marked_file.bin $TEST_FILE

echo "Created test file with size: $(du -h $TEST_FILE | cut -f1)"

# Step 3: Create a new entity with chunked content
echo "Step 3: Creating entity with chunked content..."
BASE64_CONTENT=$(base64 $TEST_FILE | tr -d '\n')

echo "{\"tags\":[\"type:test\",\"test:chunking\",\"test:final\"],\"content\":\"$BASE64_CONTENT\"}" > /tmp/create_entity_final.json

CREATE_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  --data @/tmp/create_entity_final.json)

ENTITY_ID=$(echo $CREATE_RESPONSE | grep -o '"id":"[^"]*"' | sed 's/"id":"//;s/"//')

if [ -z "$ENTITY_ID" ]; then
  echo "Failed to create entity"
  exit 1
fi

echo "Created entity with ID: $ENTITY_ID"

# Step 4: Check metadata to confirm chunking
echo "Step 4: Checking entity metadata..."
METADATA=$(curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Entity tags: $(echo $METADATA | grep -o '"tags":\[[^]]*\]')"

CHUNK_TAG=$(echo $METADATA | grep -o '"[^"]*content:chunks:[^"]*"')
echo "Chunk tag: $CHUNK_TAG"

if [ -z "$CHUNK_TAG" ]; then
  echo "Entity was not properly chunked"
  exit 1
fi

# Step 5: Retrieve entity with content
echo "Step 5: Retrieving entity with content..."
time curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_content=true" \
  -H "Authorization: Bearer $TOKEN" \
  -o /tmp/final_response.json

RESPONSE_SIZE=$(stat -c%s "/tmp/final_response.json" 2>/dev/null || stat -f%z "/tmp/final_response.json")
echo "Response size: $RESPONSE_SIZE bytes"

# Extract content and verify
echo "Step 6: Verifying retrieved content..."
cat /tmp/final_response.json | grep -o '"content":"[^"]*"' | sed 's/"content":"//;s/"//' | base64 -d > /tmp/retrieved_content.bin

RETRIEVED_SIZE=$(stat -c%s "/tmp/retrieved_content.bin" 2>/dev/null || stat -f%z "/tmp/retrieved_content.bin")
ORIGINAL_SIZE=$(stat -c%s "$TEST_FILE" 2>/dev/null || stat -f%z "$TEST_FILE")

echo "Original file size: $ORIGINAL_SIZE bytes"
echo "Retrieved file size: $RETRIEVED_SIZE bytes"

# Check for markers
START_MARKER=$(head -c 17 /tmp/retrieved_content.bin)
END_MARKER=$(tail -c 15 /tmp/retrieved_content.bin)

echo "Start marker: $START_MARKER"
echo "End marker: $END_MARKER"

# Validate results
if [ "$START_MARKER" = "START_MARKER_TEXT" ] && [ "$END_MARKER" = "END_MARKER_TEXT" ]; then
  echo "✅ MARKERS CHECK PASSED: Content integrity verified"
else
  echo "❌ MARKERS CHECK FAILED: Content corruption detected"
fi

if [ "$RETRIEVED_SIZE" -eq "$ORIGINAL_SIZE" ]; then
  echo "✅ SIZE CHECK PASSED: Retrieved content has correct size"
else
  echo "❌ SIZE CHECK FAILED: Retrieved size ($RETRIEVED_SIZE) doesn't match original ($ORIGINAL_SIZE)"
fi

# Overall result
if [ "$START_MARKER" = "START_MARKER_TEXT" ] && [ "$END_MARKER" = "END_MARKER_TEXT" ] && [ "$RETRIEVED_SIZE" -eq "$ORIGINAL_SIZE" ]; then
  echo "✅ TEST PASSED: Chunked content retrieval is working correctly!"
else
  echo "❌ TEST FAILED: Issues with chunked content retrieval remain"
fi

echo "Test completed"