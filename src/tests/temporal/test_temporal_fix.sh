#!/bin/bash
# EntityDB Temporal Features Test

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="https://localhost:8085"
TOKEN=""

# Print header
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB Temporal Features Test (FIXED)${NC}"
echo -e "${BLUE}========================================${NC}"

# Test authentication
echo -e "${BLUE}Authenticating with admin user...${NC}"
login_response=$(curl -k -s -X POST "${SERVER_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo -e "${RED}❌ Failed to login. Response: $login_response${NC}"
  exit 1
else
  echo -e "${GREEN}✅ Login successful, got token${NC}"
fi

# Create a test entity with temporal data
echo -e "${BLUE}Creating a test entity with temporal data...${NC}"
entity_response=$(curl -k -s -X POST "${SERVER_URL}/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:temporal_test", "version:1"],
    "content": "Initial content for temporal testing"
  }')

entity_id=$(echo "$entity_response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
created_at=$(echo "$entity_response" | grep -o '"created_at":"[^"]*' | cut -d'"' -f4)

if [ -z "$entity_id" ]; then
  echo -e "${RED}❌ Failed to create entity. Response: $entity_response${NC}"
  exit 1
else
  echo -e "${GREEN}✅ Created entity with ID: $entity_id${NC}"
  echo -e "${BLUE}Initial timestamp: $created_at${NC}"
fi

# Sleep to ensure different timestamps
sleep 2

# Update the entity
echo -e "${BLUE}Updating entity with new content...${NC}"
update_response=$(curl -k -s -X PUT "${SERVER_URL}/api/v1/entities/update" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$entity_id\",
    \"tags\": [\"type:temporal_test\", \"version:2\", \"updated:true\"],
    \"content\": \"Updated content for temporal testing\"
  }")

updated_at=$(echo "$update_response" | grep -o '"updated_at":"[^"]*' | cut -d'"' -f4)

if [[ "$update_response" != *"$entity_id"* ]]; then
  echo -e "${RED}❌ Failed to update entity. Response: $update_response${NC}"
else
  echo -e "${GREEN}✅ Updated entity with new tags and content${NC}"
  echo -e "${BLUE}Update timestamp: $updated_at${NC}"
fi

# Sleep to ensure different timestamps
sleep 2

# Update the entity again
echo -e "${BLUE}Updating entity a second time...${NC}"
update_response2=$(curl -k -s -X PUT "${SERVER_URL}/api/v1/entities/update" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$entity_id\",
    \"tags\": [\"type:temporal_test\", \"version:3\", \"updated:true\", \"final:true\"],
    \"content\": \"Final content for temporal testing\"
  }")

updated_at2=$(echo "$update_response2" | grep -o '"updated_at":"[^"]*' | cut -d'"' -f4)

if [[ "$update_response2" != *"$entity_id"* ]]; then
  echo -e "${RED}❌ Failed to update entity. Response: $update_response2${NC}"
else
  echo -e "${GREEN}✅ Updated entity with final tags and content${NC}"
  echo -e "${BLUE}Update timestamp: $updated_at2${NC}"
fi

# Test the as-of endpoint with the fixed implementation
echo -e "${BLUE}Testing as-of endpoint (original)...${NC}"
asof_response=$(curl -k -s -X GET "${SERVER_URL}/api/v1/entities/as-of?id=$entity_id&as_of=$created_at" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${YELLOW}Response: $asof_response${NC}"

# Test the as-of-fixed endpoint
echo -e "${BLUE}Testing as-of-fixed endpoint...${NC}"
asof_fixed_response=$(curl -k -s -X GET "${SERVER_URL}/api/v1/entities/as-of-fixed?id=$entity_id&as_of=$created_at" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${YELLOW}Response: $asof_fixed_response${NC}"

# Check if the response has the expected version tag
if [[ "$asof_fixed_response" == *"version:1"* ]]; then
  echo -e "${GREEN}✅ As-of endpoint returned entity with correct version tag${NC}"
else
  echo -e "${RED}❌ As-of endpoint failed to return entity with correct version${NC}"
fi

# Test the history endpoint
echo -e "${BLUE}Testing history endpoint (original)...${NC}"
history_response=$(curl -k -s -X GET "${SERVER_URL}/api/v1/entities/history?id=$entity_id" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${YELLOW}Response: $history_response${NC}"

# Test the history-fixed endpoint
echo -e "${BLUE}Testing history-fixed endpoint...${NC}"
history_fixed_response=$(curl -k -s -X GET "${SERVER_URL}/api/v1/entities/history-fixed?id=$entity_id" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${YELLOW}Response: $history_fixed_response${NC}"

# Check if history has entries
history_count=$(echo "$history_fixed_response" | grep -o '"timestamp"' | wc -l)
if [ "$history_count" -gt 0 ]; then
  echo -e "${GREEN}✅ History endpoint returned $history_count entries${NC}"
else
  echo -e "${RED}❌ History endpoint failed to return any entries${NC}"
fi

# Test the changes endpoint
echo -e "${BLUE}Testing changes endpoint (original)...${NC}"
changes_response=$(curl -k -s -X GET "${SERVER_URL}/api/v1/entities/changes?id=$entity_id" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${YELLOW}Response: $changes_response${NC}"

# Test the changes-fixed endpoint
echo -e "${BLUE}Testing changes-fixed endpoint...${NC}"
changes_fixed_response=$(curl -k -s -X GET "${SERVER_URL}/api/v1/entities/changes-fixed?id=$entity_id" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${YELLOW}Response: $changes_fixed_response${NC}"

# Check if changes has entries
changes_count=$(echo "$changes_fixed_response" | grep -o '"timestamp"' | wc -l)
if [ "$changes_count" -gt 0 ]; then
  echo -e "${GREEN}✅ Changes endpoint returned $changes_count entries${NC}"
else
  echo -e "${RED}❌ Changes endpoint failed to return any entries${NC}"
fi

# Test the diff endpoint
echo -e "${BLUE}Testing diff endpoint (original)...${NC}"
diff_response=$(curl -k -s -X GET "${SERVER_URL}/api/v1/entities/diff?id=$entity_id&t1=$created_at&t2=$updated_at2" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${YELLOW}Response: $diff_response${NC}"

# Test the diff-fixed endpoint
echo -e "${BLUE}Testing diff-fixed endpoint...${NC}"
diff_fixed_response=$(curl -k -s -X GET "${SERVER_URL}/api/v1/entities/diff-fixed?id=$entity_id&t1=$created_at&t2=$updated_at2" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${YELLOW}Response: $diff_fixed_response${NC}"

# Check if diff has entries
if [[ "$diff_fixed_response" == *"added_tags"* ]] || [[ "$diff_fixed_response" == *"removed_tags"* ]]; then
  echo -e "${GREEN}✅ Diff endpoint returned tag differences${NC}"
else
  echo -e "${RED}❌ Diff endpoint failed to return tag differences${NC}"
fi

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Temporal Features Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"

success_count=0
total_tests=4

# Count successes
if [[ "$asof_fixed_response" == *"version:1"* ]]; then
  ((success_count++))
fi

if [ "$history_count" -gt 0 ]; then
  ((success_count++))
fi

if [ "$changes_count" -gt 0 ]; then
  ((success_count++))
fi

if [[ "$diff_fixed_response" == *"added_tags"* ]] || [[ "$diff_fixed_response" == *"removed_tags"* ]]; then
  ((success_count++))
fi

if [ "$success_count" -eq "$total_tests" ]; then
  echo -e "${GREEN}✅ All temporal features tests PASSED! ($success_count/$total_tests)${NC}"
else
  echo -e "${RED}❌ Some temporal features tests FAILED. ($success_count/$total_tests passed)${NC}"
fi

echo -e "${BLUE}========================================${NC}"