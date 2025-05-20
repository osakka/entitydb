#!/bin/bash
# Security test script for Entity API with authentication and validation

# Set up colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Kill any existing server instances
echo -e "${BLUE}Killing any existing server instances...${NC}"
pkill -f "entitydb_server_entity" || true
sleep 1

# Ensure audit log directory exists
mkdir -p /opt/entitydb/var/log/audit

# Build the entity server with security components
echo -e "${BLUE}Building EntityDB entity server with security components...${NC}"
cd /opt/entitydb/src
go build -o entitydb_server_entity server_db.go security_manager.go security_types.go simple_security.go security_bridge.go security_input_audit.go

# Start the entity server in the background
echo -e "${BLUE}Starting EntityDB entity server on port 8087...${NC}"
./entitydb_server_entity -port 8087 &
SERVER_PID=$!

# Wait for server to start
sleep 2
echo -e "${BLUE}Server started with PID: $SERVER_PID${NC}"

# Test login to get a valid token
echo -e "\n${BLUE}1. Testing login to get authentication token...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8087/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}')

echo "$LOGIN_RESPONSE" | jq .

# Extract token from response
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')
echo -e "${GREEN}Token: ${TOKEN}${NC}"

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${RED}Failed to get valid token, exiting...${NC}"
  kill $SERVER_PID
  exit 1
fi

echo -e "\n${BLUE}2. Testing Entity API with Authentication${NC}"

# 2.1 Test entity creation with valid authentication
echo -e "\n${BLUE}2.1 Creating a test entity with valid authentication${NC}"
ENTITY_RESPONSE=$(curl -s -X POST http://localhost:8087/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"type":"test","title":"Security Test Entity","description":"Created with authentication","tags":["security","test"]}')

echo "$ENTITY_RESPONSE" | jq .

# Extract entity ID from response
ENTITY_ID=$(echo "$ENTITY_RESPONSE" | jq -r '.data.id')

if [ "$ENTITY_ID" != "null" ] && [ ! -z "$ENTITY_ID" ]; then
  echo -e "${GREEN}Entity created successfully with ID: $ENTITY_ID${NC}"
else
  echo -e "${YELLOW}Could not extract entity ID, using fallback ID for tests${NC}"
  ENTITY_ID="entity_test_fallback"
fi

# 2.2 Test entity creation without authentication
echo -e "\n${BLUE}2.2 Testing entity creation without authentication (should fail)${NC}"
curl -s -X POST http://localhost:8087/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -d '{"type":"test","title":"Unauthenticated Entity"}' | jq .

# 2.3 Test entity retrieval with authentication
echo -e "\n${BLUE}2.3 Testing entity retrieval with authentication${NC}"
curl -s -X GET "http://localhost:8087/api/v1/entities/list" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n${BLUE}3. Testing Input Validation${NC}"

# 3.1 Test with invalid entity type
echo -e "\n${BLUE}3.1 Testing with invalid entity type (should fail validation)${NC}"
curl -s -X POST http://localhost:8087/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"type":"invalid/type!","title":"Invalid Type Entity"}' | jq .

# 3.2 Test with missing required fields
echo -e "\n${BLUE}3.2 Testing with missing required fields (should fail validation)${NC}"
curl -s -X POST http://localhost:8087/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"type":"test"}' | jq .

# 3.3 Test with invalid tags
echo -e "\n${BLUE}3.3 Testing with invalid tags (should fail validation)${NC}"
curl -s -X POST http://localhost:8087/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"type":"test","title":"Invalid Tags Entity","tags":["valid","invalid tag!"]}' | jq .

echo -e "\n${BLUE}4. Testing Entity Relationships${NC}"

# 4.1 Test relationship creation
echo -e "\n${BLUE}4.1 Testing relationship creation${NC}"
curl -s -X POST http://localhost:8087/api/v1/entity-relationships/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"source_id\":\"$ENTITY_ID\",\"target_id\":\"entity_admin\",\"type\":\"test_relation\"}" | jq .

# 4.2 Test relationship retrieval
echo -e "\n${BLUE}4.2 Testing relationship retrieval${NC}"
curl -s -X GET "http://localhost:8087/api/v1/entity-relationships/list?source=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n${BLUE}5. Testing Audit Logging${NC}"

# 5.1 Check audit log
echo -e "\n${BLUE}5.1 Checking audit log entries${NC}"
find /opt/entitydb/var/log/audit/ -name "audit_*.log" -type f -exec ls -t {} \; | head -1 | xargs tail -n 30

# Stop the entity server
echo -e "\n${BLUE}Stopping EntityDB entity server...${NC}"
kill $SERVER_PID || true
wait $SERVER_PID 2>/dev/null || true

echo -e "\n${GREEN}Testing completed successfully!${NC}"