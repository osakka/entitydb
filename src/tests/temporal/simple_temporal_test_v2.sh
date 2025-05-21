#!/bin/bash

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB Temporal Test (Direct API)${NC}"
echo -e "${BLUE}========================================${NC}"

# First, stop the running server if any
echo -e "${BLUE}Stopping any running server...${NC}"
cd /opt/entitydb && ./bin/entitydbd.sh stop

# Delete the database to start fresh
echo -e "${BLUE}Deleting database files...${NC}"
rm -f /opt/entitydb/var/entities.ebf /opt/entitydb/var/entitydb.wal

# Start the server fresh
echo -e "${BLUE}Starting server...${NC}"
cd /opt/entitydb && ./bin/entitydbd.sh start

# Wait for server to start
echo -e "${BLUE}Waiting for server to start...${NC}"
sleep 5

# Login to get a token
echo -e "${BLUE}Logging in...${NC}"
login_response=$(curl -k -s -X POST "https://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo -e "${RED}❌ Failed to login. Response: $login_response${NC}"
  exit 1
else
  echo -e "${GREEN}✅ Login successful, got token${NC}"
fi

# Create a test entity
echo -e "${BLUE}Creating a test entity...${NC}"
entity_response=$(curl -k -s -X POST "https://localhost:8085/api/v1/entities/create" \
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
update_response=$(curl -k -s -X PUT "https://localhost:8085/api/v1/entities/update" \
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

# Test the history endpoint API directly
echo -e "${BLUE}Testing history endpoint...${NC}"
history_response=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/history?id=$entity_id" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${YELLOW}History response: $history_response${NC}"

# Test the as-of endpoint API directly
echo -e "${BLUE}Testing as-of endpoint...${NC}"
asof_response=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/as-of?id=$entity_id&as_of=$created_at" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${YELLOW}As-of response: $asof_response${NC}"

# Test the changes endpoint API directly
echo -e "${BLUE}Testing changes endpoint...${NC}"
changes_response=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/changes?id=$entity_id" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${YELLOW}Changes response: $changes_response${NC}"

# Test the diff endpoint API directly
echo -e "${BLUE}Testing diff endpoint...${NC}"
diff_response=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/diff?id=$entity_id&t1=$created_at&t2=$updated_at" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${YELLOW}Diff response: $diff_response${NC}"

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB Temporal Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Entity ID: $entity_id${NC}"
echo -e "${BLUE}Created At: $created_at${NC}"
echo -e "${BLUE}Updated At: $updated_at${NC}"
echo -e "${BLUE}========================================${NC}"