#!/bin/bash
# Debug script for temporal tag functionality with more detailed output

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
echo -e "Debug Temporal Tag Test"
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

# Create test entities with different tag combinations
create_test_entity() {
  local tag_list=$1
  local content=$2
  local type=$3

  echo -e "${BLUE}Creating test entity with tags: $tag_list${NC}"
  
  # Construct tags JSON array
  tags_json="["
  for tag in $tag_list; do
    if [ "$tags_json" != "[" ]; then
      tags_json="$tags_json,"
    fi
    tags_json="$tags_json\"$tag\""
  done
  tags_json="$tags_json]"
  
  # Create entity
  create_response=$(curl -s -X POST "$SERVER_URL/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"tags\": $tags_json,
      \"content\": \"$content\"
    }")

  # Extract the ID from the response
  entity_id=$(echo "$create_response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

  if [ -z "$entity_id" ]; then
    echo -e "${RED}❌ Failed to create entity. Response: $create_response${NC}"
    return ""
  else
    echo -e "${GREEN}✅ Created entity with ID: $entity_id${NC}"
    return 0
  fi
  
  echo "$entity_id"
}

# Test search for a specific tag
test_tag_search() {
  local tag=$1
  local expected_count=$2
  local include_timestamps=$3
  
  timestamp_param=""
  if [ "$include_timestamps" = "true" ]; then
    timestamp_param="&include_timestamps=true"
  fi
  
  echo -e "${BLUE}Searching for tag '$tag' (include_timestamps=$include_timestamps)...${NC}"
  search_response=$(curl -s -X GET "$SERVER_URL/api/v1/entities/list?tag=$tag$timestamp_param" \
    -H "Authorization: Bearer $TOKEN")

  # Count entities in response
  entity_count=$(echo "$search_response" | grep -o '"id"' | wc -l)
  
  echo -e "${YELLOW}Search response: $search_response${NC}"
  
  if [ "$entity_count" -eq "$expected_count" ]; then
    echo -e "${GREEN}✅ Found expected $expected_count entities with tag '$tag'${NC}"
    return 0
  else
    echo -e "${RED}❌ Expected $expected_count entities, but found $entity_count with tag '$tag'${NC}"
    return 1
  fi
}

# First, create a test entity with specific tags
echo -e "${BLUE}Creating test entities...${NC}"

# Entity 1: Regular tags, no timestamps
ENTITY1_ID=$(create_test_entity "type:regular test:entity-1 status:active" "Regular entity with normal tags" "regular")

# Entity 2: With timestamp prefix (created by server internally)
ENTITY2_ID=$(create_test_entity "type:temporal test:entity-2 status:pending" "Entity with temporal tags" "temporal")

# Wait for indexing
echo -e "${BLUE}Waiting for indexing to complete...${NC}"
sleep 3

# Test all the variations
echo -e "${BLUE}Running search tests...${NC}"

# Test without timestamps
test_tag_search "type:regular" 1 "false"
test_tag_search "type:temporal" 1 "false"
test_tag_search "test:entity-1" 1 "false"
test_tag_search "test:entity-2" 1 "false"

# Test with timestamps
test_tag_search "type:regular" 1 "true"
test_tag_search "type:temporal" 1 "true"
test_tag_search "test:entity-1" 1 "true"
test_tag_search "test:entity-2" 1 "true"

echo -e "${BLUE}Tests completed${NC}"