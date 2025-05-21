#!/bin/bash
# Minimal test for chunk handling

set -e  # Exit immediately if a command exits with a non-zero status

# Generate a tiny test file
echo "START_TEST_DATA" > /tmp/test_file.txt
cat /dev/urandom | head -c 10000 >> /tmp/test_file.txt
echo "END_TEST_DATA" >> /tmp/test_file.txt

# Base64 encode the file for the JSON request
DATA=$(base64 -w 0 < /tmp/test_file.txt)

# Create our entity with the minimal data
echo "Creating test entity..."
RESPONSE=$(curl -s -X POST "http://localhost:8085/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d "{\"tags\":[\"type:test\"], \"id\":\"test1234\", \"content\":\"$DATA\"}")

echo "Response: $RESPONSE"

# Extract ID
ENTITY_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | cut -d':' -f2 | tr -d '"')
if [ -z "$ENTITY_ID" ]; then
  ENTITY_ID="test1234"
fi

echo "Entity ID: $ENTITY_ID"

# Get the entity with content
echo "Retrieving entity..."
curl -v "http://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_content=true" > /tmp/result.json

# Cleanup
rm -f /tmp/test_file.txt /tmp/result.json

echo "Test completed"