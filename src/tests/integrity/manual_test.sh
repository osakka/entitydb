#!/bin/bash
# Manual test script to verify that our fix for chunked content retrieval works

set -e
echo "Manual test for chunked content retrieval"

# Create a test file
TEST_FILE="/tmp/test_data.bin"
TEST_SIZE=5242880  # 5MB

echo "Creating test file of $TEST_SIZE bytes..."
dd if=/dev/urandom of=$TEST_FILE bs=1M count=5 &> /dev/null

# Add a recognizable pattern at the beginning and end
echo "STARTMARKER" > /tmp/test_data_start.txt
echo "ENDMARKER" > /tmp/test_data_end.txt

cat /tmp/test_data_start.txt $TEST_FILE /tmp/test_data_end.txt > /tmp/test_data_marked.bin
mv /tmp/test_data_marked.bin $TEST_FILE

echo "Test file created with size: $(du -h $TEST_FILE | cut -f1)"

# Try direct API call to create entity
echo "Creating entity directly..."
ENTITY_ID=$(uuidgen | tr -d '-')
echo "Using ID: $ENTITY_ID"

# Encode content 
ENCODED_CONTENT=$(base64 $TEST_FILE | tr -d '\n')
CONTENT_START=${ENCODED_CONTENT:0:20}
CONTENT_END=${ENCODED_CONTENT: -20}
echo "Content encoded (${#ENCODED_CONTENT} bytes), start: $CONTENT_START..., end: ...$CONTENT_END"

# Create entity
echo "Creating entity with URL-based API..."
curl -k -s https://localhost:8085/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"$ENTITY_ID\", \"tags\":[\"type:test\", \"test:chunking\"], \"content\":\"$ENCODED_CONTENT\"}"

# Check if entity has chunked content
echo "Checking entity metadata..."
META=$(curl -k -s https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID)
echo "Entity metadata: $META"

# Try retrieving content
echo "Retrieving entity with content..."
CONTENT_RESP=$(curl -k -s https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_content=true)
echo "Content response (truncated): ${CONTENT_RESP:0:100}..."

# Extract content length
CONTENT_LEN=$(echo "$CONTENT_RESP" | jq '.content | length')
echo "Content length in response: $CONTENT_LEN bytes (original was $(du -b $TEST_FILE | cut -f1) bytes)"

# Clean up
rm -f $TEST_FILE /tmp/test_data_start.txt /tmp/test_data_end.txt