#!/bin/bash
# Basic EntityDB Test

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="https://localhost:8085"

# Print header
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB Basic Test${NC}"
echo -e "${BLUE}========================================${NC}"

# Step 1: Create admin user manually
echo -e "${BLUE}Creating admin user directly in database...${NC}"
cd /opt/entitydb/src && go run tools/users/create_users.go

# Step 2: Restart server
echo -e "${BLUE}Restarting server to apply changes...${NC}"
cd /opt/entitydb && ./bin/entitydbd.sh stop && ./bin/entitydbd.sh start

# Step 3: Test basic API operations
echo -e "${BLUE}Testing basic API operations...${NC}"

# Test login
echo -e "${BLUE}Attempting login...${NC}"
login_response=$(curl -k -s -X POST "$SERVER_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

if [[ "$login_response" == *"token"* ]]; then
  echo -e "${GREEN}✅ Login successful${NC}"
  TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
else
  echo -e "${RED}❌ Login failed: $login_response${NC}"
  exit 1
fi

# Test entity creation
echo -e "${BLUE}Creating test entity...${NC}"
entity_response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:test", "basic:test"],
    "content": "Test content for basic testing"
  }')

if [[ "$entity_response" == *"id"* ]]; then
  echo -e "${GREEN}✅ Entity created successfully${NC}"
  ENTITY_ID=$(echo "$entity_response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
else
  echo -e "${RED}❌ Entity creation failed: $entity_response${NC}"
  exit 1
fi

# Test entity retrieval
echo -e "${BLUE}Retrieving test entity...${NC}"
get_response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/get?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")

if [[ "$get_response" == *"$ENTITY_ID"* ]]; then
  echo -e "${GREEN}✅ Entity retrieved successfully${NC}"
else
  echo -e "${RED}❌ Entity retrieval failed: $get_response${NC}"
  exit 1
fi

# Test entity update
echo -e "${BLUE}Updating test entity...${NC}"
update_response=$(curl -k -s -X PUT "$SERVER_URL/api/v1/entities/update" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$ENTITY_ID\",
    \"tags\": [\"type:test\", \"basic:test\", \"updated:true\"],
    \"content\": \"Updated test content for basic testing\"
  }")

if [[ "$update_response" == *"$ENTITY_ID"* ]]; then
  echo -e "${GREEN}✅ Entity updated successfully${NC}"
else
  echo -e "${RED}❌ Entity update failed: $update_response${NC}"
  exit 1
fi

# Test entity list
echo -e "${BLUE}Listing entities with tag 'type:test'...${NC}"
list_response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/list?tag=type:test" \
  -H "Authorization: Bearer $TOKEN")

if [[ "$list_response" == *"$ENTITY_ID"* ]]; then
  echo -e "${GREEN}✅ Entity list successful${NC}"
  entity_count=$(echo "$list_response" | grep -o "\"id\":" | wc -l)
  echo -e "${BLUE}Found $entity_count entities with tag 'type:test'${NC}"
else
  echo -e "${RED}❌ Entity list failed: $list_response${NC}"
  exit 1
fi

# Test query API
echo -e "${BLUE}Running query for entities with tag 'basic:test'...${NC}"
query_response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/query?tag=basic:test" \
  -H "Authorization: Bearer $TOKEN")

if [[ "$query_response" == *"$ENTITY_ID"* ]]; then
  echo -e "${GREEN}✅ Query API successful${NC}"
else
  echo -e "${RED}❌ Query API failed: $query_response${NC}"
  exit 1
fi

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB Basic Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✅ All basic tests passed successfully${NC}"
echo -e "${BLUE}Entity ID: $ENTITY_ID${NC}"
echo -e "${BLUE}========================================${NC}"

# Print database sizes
echo -e "${BLUE}Current database sizes:${NC}"
ls -lh /opt/entitydb/var/entities.ebf /opt/entitydb/var/entitydb.wal