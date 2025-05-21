#!/bin/bash
# Test script to show what's going on with our fix

echo "Testing chunking fix directly..."

# Check if the server is running
curl -s -k https://localhost:8085/api/v1/status || {
  echo "Server not responding, please start it first"
  exit 1
}

# Get the current code that handles chunked entity retrieval
echo "Current code in GetEntity function:"
cd /opt/entitydb/src/api
grep -A10 "if includeContent && entity.IsChunked()" entity_handler.go

echo "====================" 
echo "Creating a small test file (100KB)"
echo "Test data content" > /tmp/test_data.txt
head -c 100K /dev/urandom >> /tmp/test_data.txt

echo "File created with size: $(du -h /tmp/test_data.txt | cut -f1)"
echo "===================="

# Manually trace through how the GetEntity code would handle chunking
echo "Checking if IsChunked() works:"
cd /opt/entitydb/src
grep -A10 "IsChunked" models/entity.go

echo "===================="
echo "Checking GetContentMetadata():"
grep -A10 "GetContentMetadata" models/entity.go

echo "===================="
echo "Looking at how chunking works for entity creation:"
cd /opt/entitydb/src/api
grep -A20 "AutoChunkThreshold" entity_handler.go

echo "===================="
echo "Looking at how metadata tags are stored:"
grep -A5 "entity.AddTag(\"content:chunks:" entity_handler.go