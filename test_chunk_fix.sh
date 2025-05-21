#!/bin/bash
# Test script for the improved chunk handling

set -e  # Exit immediately if a command exits with a non-zero status

# ANSI color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing Improved Chunk Handling${NC}"
echo -e "${BLUE}========================================${NC}"

# Generate test data
echo -e "${BLUE}Generating test data (5MB)...${NC}"
dd if=/dev/urandom of=/tmp/test_5mb.bin bs=1M count=5 2>/dev/null

# Mark beginning and end for verification
echo -e "START_TEST_DATA" > /tmp/test_start.txt
cat /tmp/test_start.txt /tmp/test_5mb.bin > /tmp/test_5mb_marked.bin
echo -e "END_TEST_DATA" >> /tmp/test_5mb_marked.bin

echo -e "${BLUE}Creating test entity with chunked content...${NC}"

# Create entity using curl
TEST_DATA=$(cat /tmp/test_5mb_marked.bin | base64)
ENTITY_ID=$(curl -s -X POST "http://localhost:8085/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d "{
    \"tags\": [\"type:test_chunking\", \"test:improved_chunking\"],
    \"content\": \"$TEST_DATA\"
  }" | jq -r '.id')

if [ -z "$ENTITY_ID" ]; then
  echo -e "${RED}❌ Failed to create test entity${NC}"
  exit 1
fi

echo -e "${GREEN}✅ Created test entity with ID: $ENTITY_ID${NC}"

# Test retrieving the entity with content
echo -e "${BLUE}Retrieving entity with content...${NC}"
curl -s -o /tmp/retrieved_content.bin "http://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_content=true&raw=true"

# Check file size
ORIGINAL_SIZE=$(stat -c%s "/tmp/test_5mb_marked.bin")
RETRIEVED_SIZE=$(stat -c%s "/tmp/retrieved_content.bin")

echo -e "${BLUE}Original size: $ORIGINAL_SIZE bytes${NC}"
echo -e "${BLUE}Retrieved size: $RETRIEVED_SIZE bytes${NC}"

if [ "$ORIGINAL_SIZE" -eq "$RETRIEVED_SIZE" ]; then
  echo -e "${GREEN}✅ Size check passed: Original and retrieved files have the same size${NC}"
else
  echo -e "${RED}❌ Size check failed: Original and retrieved files have different sizes${NC}"
  exit 1
fi

# Check content
if grep -q "START_TEST_DATA" /tmp/retrieved_content.bin && grep -q "END_TEST_DATA" /tmp/retrieved_content.bin; then
  echo -e "${GREEN}✅ Content check passed: Start and end markers verified${NC}"
else
  echo -e "${RED}❌ Content check failed: Could not find start and end markers${NC}"
  exit 1
fi

echo -e "${GREEN}✅ All tests passed successfully!${NC}"
echo -e "${BLUE}========================================${NC}"

# Cleanup
rm -f /tmp/test_5mb.bin /tmp/test_start.txt /tmp/test_5mb_marked.bin /tmp/retrieved_content.bin

exit 0