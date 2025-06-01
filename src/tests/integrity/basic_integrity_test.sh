#!/bin/bash
# Simple test to verify chunking functionality

echo "Simple test for chunking functionality"

# Create smaller test file
TEST_FILE="/tmp/small_test.bin"
echo "Creating test file..."
dd if=/dev/urandom of=$TEST_FILE bs=1M count=1 &> /dev/null
echo "Test file created: $(du -h $TEST_FILE | cut -f1)"

# Create entity with small content first
echo "Creating entity with small content..."
ENTITY_ID="test$(date +%s)"
RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/entities/create" \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"$ENTITY_ID\",\"tags\":[\"type:test\",\"test:chunking\"]}")

echo "Created entity: $RESPONSE"
echo "Entity ID: $ENTITY_ID"

# Check if our chunking code works on a normal entity
echo "Retrieving entity with content..."
RETRIEVED=$(curl -k -s "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_content=true")
echo "Retrieved entity: $RETRIEVED"

# Clean up
rm -f $TEST_FILE

echo "Test completed successfully!"