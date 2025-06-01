#!/bin/bash
# Simple test for temporal tag functionality

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8085"
TOKEN=""

echo -e "${BLUE}========================================"
echo -e "Simple Temporal Tag Test"
echo -e "========================================${NC}"

# Login to get token
echo -e "${BLUE}Logging in...${NC}"
response=$(curl -s -X POST "$SERVER_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo -e "${RED}❌ Failed to login. Response: $response${NC}"
  exit 1
else
  echo -e "${GREEN}✅ Login successful${NC}"
fi

# Create a test entity
echo -e "${BLUE}Creating test entity...${NC}"
TEST_ID=$(date +%s)
create_response=$(curl -s -X POST "$SERVER_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"tags\": [\"type:test\", \"test:id:$TEST_ID\"],
    \"content\": \"Test entity for temporal tag test at $(date)\"
  }")

# Extract the ID from the response
entity_id=$(echo "$create_response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

if [ -z "$entity_id" ]; then
  echo -e "${RED}❌ Failed to create entity. Response: $create_response${NC}"
  exit 1
else
  echo -e "${GREEN}✅ Created entity with ID: $entity_id${NC}"
fi

# Wait for indexing
echo -e "${BLUE}Waiting for indexing...${NC}"
sleep 3

# Try to find entity by tag
echo -e "${BLUE}Searching for entity by tag type:test...${NC}"
search_response=$(curl -s -X GET "$SERVER_URL/api/v1/entities/list?tag=type:test" \
  -H "Authorization: Bearer $TOKEN")

if echo "$search_response" | grep -q "$entity_id"; then
  echo -e "${GREEN}✅ Successfully found entity by tag type:test${NC}"
else
  echo -e "${RED}❌ Could not find entity by tag type:test${NC}"
  echo -e "${YELLOW}Search response: $search_response${NC}"
fi

# Try to find entity by specific tag
echo -e "${BLUE}Searching for entity by specific tag test:id:$TEST_ID...${NC}"
specific_search=$(curl -s -X GET "$SERVER_URL/api/v1/entities/list?tag=test:id:$TEST_ID" \
  -H "Authorization: Bearer $TOKEN")

if echo "$specific_search" | grep -q "$entity_id"; then
  echo -e "${GREEN}✅ Successfully found entity by specific tag${NC}"
else
  echo -e "${RED}❌ Could not find entity by specific tag${NC}"
  echo -e "${YELLOW}Specific search response: $specific_search${NC}"
fi

echo -e "${BLUE}Test completed${NC}"