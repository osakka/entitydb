#!/bin/bash

# Login credentials
USERNAME="admin"
PASSWORD="admin"
SERVER_URL="https://localhost:8085"
TOKEN=""

# Define color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Clean database and restart server
echo -e "${BLUE}Stopping server and cleaning database...${NC}"
cd /opt/entitydb && ./bin/entitydbd.sh stop
rm -f /opt/entitydb/var/entities.ebf /opt/entitydb/var/entitydb.wal 
cd /opt/entitydb && ./bin/entitydbd.sh start

# Wait for server to start properly
sleep 5

# Test basic authentication
echo -e "${BLUE}Testing login with admin/admin credentials...${NC}"
login_response=$(curl -k -s -X POST "$SERVER_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

if [[ "$login_response" == *"token"* ]]; then
  echo -e "${GREEN}✅ Login successful${NC}"
else
  echo -e "${RED}❌ Login failed: $login_response${NC}"
  
  # Try to create admin directly
  echo -e "${YELLOW}Trying to create admin user via direct API call...${NC}"
  direct_response=$(curl -k -s -X POST "$SERVER_URL/api/v1/users/create" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\",\"is_admin\":true}")
    
  echo "Response: $direct_response"
  
  # Try login again
  echo -e "${BLUE}Trying login again after direct user creation...${NC}"
  login_response=$(curl -k -s -X POST "$SERVER_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")
    
  if [[ "$login_response" == *"token"* ]]; then
    echo -e "${GREEN}✅ Login successful after direct user creation${NC}"
  else
    echo -e "${RED}❌ Login still failed: $login_response${NC}"
    exit 1
  fi
fi

# Get token
TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo -e "${GREEN}Token: $TOKEN${NC}"

# Create test entity
echo -e "${BLUE}Creating test entity...${NC}"
entity_response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"tags\": [\"type:test\", \"test:simple\"],
    \"content\": \"Simple test entity content\"
  }")

if [[ "$entity_response" == *"id"* ]]; then
  echo -e "${GREEN}✅ Entity created successfully${NC}"
  ENTITY_ID=$(echo "$entity_response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
  echo "Entity ID: $ENTITY_ID"
else
  echo -e "${RED}❌ Entity creation failed: $entity_response${NC}"
fi