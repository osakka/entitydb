#!/bin/bash

# Check database format

BASE_URL="http://localhost:8085"

echo "Checking database format..."

# Look at the binary files
echo -e "\n=== Binary database files ==="
ls -la /opt/entitydb/var/db/binary/ 2>/dev/null || echo "No binary db directory"

# Check environment variables
echo -e "\n=== Environment variables ==="
env | grep ENTITYDB || echo "No ENTITYDB env vars set"

# Create a test entity and see the format
echo -e "\n=== Creating test entity ==="
TEST_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:format-test", "test:db-format"]
  }')

TEST_ID=$(echo "$TEST_RESPONSE" | jq -r .id)
echo "Created entity: $TEST_ID"

# Get raw response
echo -e "\n=== Raw create response ==="
echo "$TEST_RESPONSE" | jq .

# Check internal tag format
echo -e "\n=== Examining first tag format ==="
FIRST_TAG=$(echo "$TEST_RESPONSE" | jq -r '.tags[0]')
echo "First tag: $FIRST_TAG"

# Count pipe characters
PIPE_COUNT=$(echo "$FIRST_TAG" | tr -cd '|' | wc -c)
echo "Number of pipes: $PIPE_COUNT"

# Parse the tag
if [ $PIPE_COUNT -eq 2 ]; then
    echo "Tag has double timestamp format (ISO|NANO|tag)"
    ISO_TS=$(echo "$FIRST_TAG" | cut -d'|' -f1)
    NANO_TS=$(echo "$FIRST_TAG" | cut -d'|' -f2)
    TAG_PART=$(echo "$FIRST_TAG" | cut -d'|' -f3)
    echo "  ISO timestamp: $ISO_TS"
    echo "  Nano timestamp: $NANO_TS"
    echo "  Tag part: $TAG_PART"
elif [ $PIPE_COUNT -eq 1 ]; then
    echo "Tag has single timestamp format"
    TS=$(echo "$FIRST_TAG" | cut -d'|' -f1)
    TAG_PART=$(echo "$FIRST_TAG" | cut -d'|' -f2)
    echo "  Timestamp: $TS"
    echo "  Tag part: $TAG_PART"
fi

# Check server logs
echo -e "\n=== Recent server logs ==="
tail -10 /opt/entitydb/var/log/entitydb.log 2>/dev/null | grep -i "temporal\|turbo" || echo "No temporal/turbo logs found"