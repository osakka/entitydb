#!/bin/bash
# EntityDB Server Status Test

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print header
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB Server Status Test${NC}"
echo -e "${BLUE}========================================${NC}"

# Check if the server is running
echo -e "${BLUE}Checking if EntityDB server is running...${NC}"
if pgrep -f "entitydb" > /dev/null; then
  echo -e "${GREEN}✅ EntityDB server is running${NC}"
else
  echo -e "${RED}❌ EntityDB server is NOT running${NC}"
  exit 1
fi

# Test basic API endpoints
echo -e "${BLUE}Testing API endpoints...${NC}"

# Test status endpoint
status_response=$(curl -k -s "https://localhost:8085/api/v1/status")
if [[ "$status_response" == *"EntityDB Consolidated Server"* ]] || [[ "$status_response" == *"status"* ]]; then
  echo -e "${GREEN}✅ Status endpoint is responding${NC}"
else
  echo -e "${RED}❌ Status endpoint is not responding correctly: $status_response${NC}"
fi

# Login to get token
echo -e "${BLUE}Testing authentication...${NC}"
login_response=$(curl -k -s -X POST "https://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

if [[ "$login_response" == *"token"* ]]; then
  echo -e "${GREEN}✅ Authentication is working${NC}"
  TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
else
  echo -e "${RED}❌ Authentication failed: $login_response${NC}"
  exit 1
fi

# Test entities endpoint
echo -e "${BLUE}Testing entities endpoint...${NC}"
entities_response=$(curl -k -s "https://localhost:8085/api/v1/entities/list" \
  -H "Authorization: Bearer $TOKEN")

if [[ "$entities_response" == *"id"* ]]; then
  echo -e "${GREEN}✅ Entities endpoint is working${NC}"
  entity_count=$(echo "$entities_response" | grep -o "\"id\":" | wc -l)
  echo -e "${BLUE}Found $entity_count entities in the database${NC}"
else
  echo -e "${RED}❌ Entities endpoint failed: $entities_response${NC}"
fi

# Check database files
echo -e "${BLUE}Checking database files...${NC}"
ls -lah /opt/entitydb/var/

if [ -f "/opt/entitydb/var/entities.ebf" ] && [ -f "/opt/entitydb/var/entitydb.wal" ]; then
  echo -e "${GREEN}✅ Database files exist${NC}"
  ebf_size=$(du -h /opt/entitydb/var/entities.ebf | cut -f1)
  wal_size=$(du -h /opt/entitydb/var/entitydb.wal | cut -f1)
  echo -e "${BLUE}EBF size: $ebf_size, WAL size: $wal_size${NC}"
else
  echo -e "${RED}❌ Database files are missing${NC}"
fi

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB Server Status Summary${NC}"
echo -e "${BLUE}========================================${NC}"
if pgrep -f "entitydb" > /dev/null && [[ "$status_response" == *"status"* ]] && [[ "$login_response" == *"token"* ]]; then
  echo -e "${GREEN}✅ EntityDB server is operational${NC}"
else
  echo -e "${RED}❌ EntityDB server has issues${NC}"
fi
echo -e "${BLUE}========================================${NC}"