#!/bin/bash
# Test script for enhanced security features (validation, RBAC, password storage)

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
go build -o entitydb_server_entity server_db.go security_manager.go security_types.go simple_security.go security_bridge.go security_input_audit.go security_rbac.go security_password_upgrade.go

# Start the entity server in the background
echo -e "${BLUE}Starting EntityDB entity server on port 8087...${NC}"
./entitydb_server_entity -port 8087 &
SERVER_PID=$!

# Wait for server to start
sleep 2
echo -e "${BLUE}Server started with PID: $SERVER_PID${NC}"

# 1. Test Enhanced Input Validation
echo -e "\n${BLUE}1. Testing Enhanced Input Validation${NC}"

# 1.1 Login to get token
echo -e "\n${BLUE}1.1 Login to get authentication token${NC}"
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8087/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}')

echo "$LOGIN_RESPONSE" | jq .

# Extract token from response
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')
echo -e "${GREEN}Token: ${TOKEN}${NC}"

# 1.2 Test invalid entity type (now should fail validation)
echo -e "\n${BLUE}1.2 Testing entity creation with invalid type (should fail validation)${NC}"
curl -s -X POST http://localhost:8087/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"type":"invalid/type!","title":"Invalid Type Entity"}' | jq .

# 1.3 Test invalid tag format
echo -e "\n${BLUE}1.3 Testing entity creation with invalid tag format (should fail validation)${NC}"
curl -s -X POST http://localhost:8087/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"type":"test","title":"Invalid Tags Entity","tags":["UPPERCASE","invalid tag!"]}' | jq .

# 1.4 Test reserved type
echo -e "\n${BLUE}1.4 Testing entity creation with reserved type (should fail validation)${NC}"
curl -s -X POST http://localhost:8087/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"type":"system","title":"System Entity"}' | jq .

# 1.5 Test valid entity creation
echo -e "\n${BLUE}1.5 Testing valid entity creation (should succeed)${NC}"
curl -s -X POST http://localhost:8087/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"type":"valid","title":"Valid Entity","tags":["test","security"]}' | jq .

# 2. Test RBAC functionality
echo -e "\n${BLUE}2. Testing RBAC Functionality${NC}"

# 2.1 Login with read-only user
echo -e "\n${BLUE}2.1 Login with read-only user${NC}"
READONLY_LOGIN=$(curl -s -X POST http://localhost:8087/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"readonly_user","password":"password123"}')

echo "$READONLY_LOGIN" | jq .

# Extract readonly token
READONLY_TOKEN=$(echo "$READONLY_LOGIN" | jq -r '.data.token')
echo -e "${GREEN}Read-only Token: ${READONLY_TOKEN}${NC}"

# 2.2 Try to create entity with read-only user (should fail)
echo -e "\n${BLUE}2.2 Testing entity creation with read-only user (should fail)${NC}"
curl -s -X POST http://localhost:8087/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $READONLY_TOKEN" \
  -d '{"type":"readonly","title":"Readonly Test Entity"}' | jq .

# 2.3 Try to read entities with read-only user (should succeed)
echo -e "\n${BLUE}2.3 Testing entity listing with read-only user (should succeed)${NC}"
curl -s -X GET http://localhost:8087/api/v1/entities/list \
  -H "Authorization: Bearer $READONLY_TOKEN" | jq .

# 3. Test Secure Password Storage
echo -e "\n${BLUE}3. Testing Secure Password Storage${NC}"

# 3.1 Check server logs for password upgrade messages
echo -e "\n${BLUE}3.1 Checking server logs for password upgrade messages${NC}"
grep -i "password" /tmp/server_log.txt || echo "No password upgrade messages found"

# Stop the entity server
echo -e "\n${BLUE}Stopping EntityDB entity server...${NC}"
kill $SERVER_PID || true
wait $SERVER_PID 2>/dev/null || true

echo -e "\n${GREEN}Testing completed!${NC}"