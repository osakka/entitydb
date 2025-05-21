#!/bin/bash
# Test script for the improved chunk handling with smaller files

# ANSI color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing Improved Chunk Handling (Simple)${NC}"
echo -e "${BLUE}========================================${NC}"

# Generate a much smaller test file (1MB)
echo -e "${BLUE}Generating test data (1MB)...${NC}"
dd if=/dev/urandom of=/tmp/test_1mb.bin bs=1M count=1 2>/dev/null

# Add markers for verification
echo "START_TEST_DATA" > /tmp/test_1mb_marked.bin
cat /tmp/test_1mb.bin >> /tmp/test_1mb_marked.bin
echo "END_TEST_DATA" >> /tmp/test_1mb_marked.bin

# Calculate checksum for verification later
CHECKSUM=$(sha256sum /tmp/test_1mb_marked.bin | awk '{print $1}')
echo -e "${BLUE}Original file checksum: $CHECKSUM${NC}"

echo -e "${BLUE}Creating test entity with content...${NC}"

# Create entity with the test content - direct approach with smaller file
ENTITY_ID=$(curl -s -X POST "http://localhost:8085/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d "{\"tags\": [\"type:test_chunking\", \"test:simple_test\"], \"content\": \"$(base64 -w 0 < /tmp/test_1mb_marked.bin)\"}" | grep -o '"id":"[^"]*"' | cut -d':' -f2 | tr -d '"')

if [ -z "$ENTITY_ID" ]; then
  echo -e "${RED}❌ Failed to create test entity${NC}"
  exit 1
fi

echo -e "${GREEN}✅ Created test entity with ID: $ENTITY_ID${NC}"

# Test retrieving the entity with content
echo -e "${BLUE}Retrieving entity with content...${NC}"
curl -s -o /tmp/retrieved_content.bin "http://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_content=true&raw=true"

# Check file size
ORIGINAL_SIZE=$(stat -c%s "/tmp/test_1mb_marked.bin")
RETRIEVED_SIZE=$(stat -c%s "/tmp/retrieved_content.bin")

echo -e "${BLUE}Original size: $ORIGINAL_SIZE bytes${NC}"
echo -e "${BLUE}Retrieved size: $RETRIEVED_SIZE bytes${NC}"

if [ "$ORIGINAL_SIZE" -eq "$RETRIEVED_SIZE" ]; then
  echo -e "${GREEN}✅ Size check passed: Files have the same size${NC}"
else
  echo -e "${RED}❌ Size check failed: Files have different sizes${NC}"
  exit 1
fi

# Check content markers
if grep -q "START_TEST_DATA" /tmp/retrieved_content.bin && grep -q "END_TEST_DATA" /tmp/retrieved_content.bin; then
  echo -e "${GREEN}✅ Content check passed: Start and end markers verified${NC}"
else
  echo -e "${RED}❌ Content check failed: Missing markers${NC}"
  exit 1
fi

# Verify checksum of retrieved file
RETRIEVED_CHECKSUM=$(sha256sum /tmp/retrieved_content.bin | awk '{print $1}')
echo -e "${BLUE}Retrieved file checksum: $RETRIEVED_CHECKSUM${NC}"

if [ "$CHECKSUM" = "$RETRIEVED_CHECKSUM" ]; then
  echo -e "${GREEN}✅ Checksum verification passed${NC}"
else
  echo -e "${RED}❌ Checksum verification failed${NC}"
  exit 1
fi

echo -e "${GREEN}✅ All tests passed successfully!${NC}"
echo -e "${BLUE}========================================${NC}"

# Cleanup
rm -f /tmp/test_1mb.bin /tmp/test_1mb_marked.bin /tmp/retrieved_content.bin

exit 0