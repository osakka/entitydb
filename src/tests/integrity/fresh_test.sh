#!/bin/bash
# Comprehensive test script for chunked content retrieval
# with a fresh database instance

set -e
echo "Testing chunked content retrieval on fresh database"

# Create test file
TEST_FILE="/tmp/test_chunk_file.bin"
TEST_SIZE=6291456  # 6MB (larger than default 4MB chunk size)

echo "Creating test file of $TEST_SIZE bytes..."
dd if=/dev/urandom of=$TEST_FILE bs=1M count=6 &> /dev/null

# Add recognizable markers at the beginning, middle, and end
echo "STARTMARKER_FOR_TESTING_123" > /tmp/start_marker.txt
echo "MIDDLE_MARKER_FOR_TESTING_456" > /tmp/middle_marker.txt
echo "END_MARKER_FOR_TESTING_789" > /tmp/end_marker.txt

# Replace beginning of file with start marker
cat /tmp/start_marker.txt > /tmp/test_with_markers.bin
dd if=$TEST_FILE bs=1 skip=25 count=3145728 >> /tmp/test_with_markers.bin 2> /dev/null

# Add middle marker
cat /tmp/middle_marker.txt >> /tmp/test_with_markers.bin
dd if=$TEST_FILE bs=1 skip=3145753 count=3145703 >> /tmp/test_with_markers.bin 2> /dev/null

# Add end marker
cat /tmp/end_marker.txt >> /tmp/test_with_markers.bin
mv /tmp/test_with_markers.bin $TEST_FILE

echo "Test file created with size: $(du -h $TEST_FILE | cut -f1)"
echo "Added markers: STARTMARKER_FOR_TESTING_123, MIDDLE_MARKER_FOR_TESTING_456, END_MARKER_FOR_TESTING_789"

# Login with admin/admin (the default admin account)
echo "Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

echo "Login response: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | sed 's/"token":"//;s/"//')

if [ -z "$TOKEN" ]; then
  echo "ERROR: Failed to get authentication token"
  exit 1
fi

echo "Got authentication token: ${TOKEN:0:10}...${TOKEN:(-10)}"

# Create entity with large file content
echo "Creating entity with large file (should trigger chunking)..."
BASE64_CONTENT=$(base64 $TEST_FILE | tr -d '\n')

# Save request to file to avoid command line length issues
cat > /tmp/create_request.json << EOF
{
  "tags": ["type:test", "test:chunking", "test:large-file"],
  "content": "$BASE64_CONTENT"
}
EOF

CREATE_RESPONSE=$(curl -s -k -X POST "https://localhost:8085/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  --data @/tmp/create_request.json)

ENTITY_ID=$(echo $CREATE_RESPONSE | grep -o '"id":"[^"]*"' | sed 's/"id":"//;s/"//')

if [ -z "$ENTITY_ID" ]; then
  echo "ERROR: Failed to extract entity ID"
  exit 1
fi

echo "Created entity with ID: $ENTITY_ID"

# Check entity metadata for chunking
echo "Checking entity metadata for chunking tags..."
META_RESPONSE=$(curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Entity tags:"
echo "$META_RESPONSE" | grep -o '"tags":\[[^]]*\]'

# Look for chunk tags
CHUNK_TAGS=$(echo "$META_RESPONSE" | grep -o '"[^"]*content:chunks:[^"]*"')
CHUNK_SIZE_TAGS=$(echo "$META_RESPONSE" | grep -o '"[^"]*content:chunk-size:[^"]*"')

echo "Chunk tags: $CHUNK_TAGS"
echo "Chunk size tags: $CHUNK_SIZE_TAGS"

if [ -z "$CHUNK_TAGS" ]; then
  echo "ERROR: Entity was not chunked as expected"
  exit 1
fi

# Retrieve the entity with content
echo "Retrieving entity with content (testing our fix)..."
time CONTENT_RESPONSE=$(curl -s -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_content=true" \
  -H "Authorization: Bearer $TOKEN")

# Save the base64 content for inspection
RECEIVED_CONTENT=$(echo "$CONTENT_RESPONSE" | grep -o '"content":"[^"]*"' | sed 's/"content":"//;s/"//')
echo "$RECEIVED_CONTENT" | base64 -d > /tmp/retrieved_content.bin

# Check file sizes
ORIGINAL_SIZE=$(stat -c%s "$TEST_FILE" 2>/dev/null || stat -f%z "$TEST_FILE")
RETRIEVED_SIZE=$(stat -c%s "/tmp/retrieved_content.bin" 2>/dev/null || stat -f%z "/tmp/retrieved_content.bin")

echo "Original file size: $ORIGINAL_SIZE bytes"
echo "Retrieved file size: $RETRIEVED_SIZE bytes"

# Check for markers
START_MARKER=$(head -c 28 /tmp/retrieved_content.bin)
END_MARKER=$(tail -c 27 /tmp/retrieved_content.bin)

# Check for middle marker
MIDDLE_POSITION=$((ORIGINAL_SIZE / 2))
dd if=/tmp/retrieved_content.bin bs=1 skip=$((MIDDLE_POSITION - 15)) count=30 of=/tmp/middle_section.bin 2>/dev/null
MIDDLE_MARKER=$(cat /tmp/middle_section.bin)

echo "Start marker: $START_MARKER"
echo "Middle marker contains 'MIDDLE': $(echo $MIDDLE_MARKER | grep -q 'MIDDLE' && echo 'YES' || echo 'NO')"
echo "End marker: $END_MARKER"

# Validate results
if [ "$ORIGINAL_SIZE" -eq "$RETRIEVED_SIZE" ]; then
  echo "✅ SIZE CHECK PASSED: Original and retrieved files are the same size"
else
  echo "❌ SIZE CHECK FAILED: Original ($ORIGINAL_SIZE bytes) and retrieved ($RETRIEVED_SIZE bytes) sizes differ"
fi

if [ "$START_MARKER" = "STARTMARKER_FOR_TESTING_123" ]; then
  echo "✅ START MARKER CHECK PASSED: Start marker found"
else
  echo "❌ START MARKER CHECK FAILED: Start marker not found"
  echo "Expected: STARTMARKER_FOR_TESTING_123"
  echo "Found: $START_MARKER"
fi

if echo "$MIDDLE_MARKER" | grep -q "MIDDLE_MARKER_FOR_TESTING_456"; then
  echo "✅ MIDDLE MARKER CHECK PASSED: Middle marker found"
else
  echo "❌ MIDDLE MARKER CHECK FAILED: Middle marker not found"
  echo "Expected to contain: MIDDLE_MARKER_FOR_TESTING_456"
  echo "Found: $MIDDLE_MARKER"
fi

if [ "$END_MARKER" = "END_MARKER_FOR_TESTING_789" ]; then
  echo "✅ END MARKER CHECK PASSED: End marker found"
else
  echo "❌ END MARKER CHECK FAILED: End marker not found"
  echo "Expected: END_MARKER_FOR_TESTING_789"
  echo "Found: $END_MARKER"
fi

# Overall test result
if [ "$ORIGINAL_SIZE" -eq "$RETRIEVED_SIZE" ] && \
   [ "$START_MARKER" = "STARTMARKER_FOR_TESTING_123" ] && \
   echo "$MIDDLE_MARKER" | grep -q "MIDDLE_MARKER_FOR_TESTING_456" && \
   [ "$END_MARKER" = "END_MARKER_FOR_TESTING_789" ]; then
  echo "✅ TEST PASSED: Chunked content retrieval works correctly!"
else
  echo "❌ TEST FAILED: Issues with chunked content retrieval"
fi

# Clean up
echo "Cleaning up..."
rm -f "$TEST_FILE" "/tmp/retrieved_content.bin" "/tmp/start_marker.txt" "/tmp/middle_marker.txt" "/tmp/end_marker.txt" "/tmp/create_request.json" "/tmp/middle_section.bin"

echo "Test completed!"