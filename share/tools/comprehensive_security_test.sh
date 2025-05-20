#!/bin/bash
# Comprehensive security testing script for EntityDB server
# Tests all security components with detailed validation

# Set up colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Test result counters
PASSED=0
FAILED=0

# Function to print a test result
test_result() {
  if [ "$1" = "PASS" ]; then
    echo -e "${GREEN}[PASS]${NC} $2"
    PASSED=$((PASSED+1))
  else
    echo -e "${RED}[FAIL]${NC} $2"
    FAILED=$((FAILED+1))
  fi
}

# Function to test API endpoint
test_api() {
  local method=$1
  local endpoint=$2
  local expected_status=$3
  local description=$4
  local data=$5
  local token=$6
  
  # Set up headers
  local headers=()
  if [ ! -z "$token" ]; then
    headers+=(-H "Authorization: Bearer $token")
  fi
  
  if [ ! -z "$data" ]; then
    headers+=(-H "Content-Type: application/json")
    # Send request with data
    response=$(curl -s -X "$method" "http://localhost:8087/$endpoint" "${headers[@]}" -d "$data" -w "\n%{http_code}")
  else
    # Send request without data
    response=$(curl -s -X "$method" "http://localhost:8087/$endpoint" "${headers[@]}" -w "\n%{http_code}")
  fi
  
  # Extract status code from response
  status_code=$(echo "$response" | tail -n1)
  # Extract JSON response (all but last line)
  json_response=$(echo "$response" | sed '$d')
  
  # Check if status code matches expected
  if [ "$status_code" -eq "$expected_status" ]; then
    test_result "PASS" "$description (Status: $status_code)"
  else
    test_result "FAIL" "$description (Expected: $expected_status, Got: $status_code)"
    echo -e "${YELLOW}Response: $json_response${NC}"
  fi
  
  # Return the JSON response for further testing
  echo "$json_response"
}

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

echo -e "\n${BLUE}=== Testing Security Components ===${NC}"

# 1. Test basic server status
echo -e "\n${BLUE}1. Testing Server Status${NC}"
test_api "GET" "api/v1/status" 200 "Server status endpoint"

# 2. Test Input Validation
echo -e "\n${BLUE}2. Testing Input Validation${NC}"

# 2.1 Test valid login
echo -e "\n${BLUE}2.1 Testing valid login${NC}"
login_response=$(test_api "POST" "api/v1/login" 200 "Valid login" '{"username":"admin","password":"password"}')

# Extract token from response
token=$(echo "$login_response" | jq -r '.data.token')
if [ "$token" != "null" ] && [ ! -z "$token" ]; then
  test_result "PASS" "Token extraction (Token: ${token:0:15}...)"
else
  test_result "FAIL" "Token extraction (No token received)"
  # Set a default token for testing - this may not work but prevents script from failing
  token="tk_admin_invalid"
fi

# 2.2 Test invalid login (bad username)
echo -e "\n${BLUE}2.2 Testing invalid login - bad username${NC}"
test_api "POST" "api/v1/login" 401 "Invalid login - bad username" '{"username":"nonexistent","password":"password"}'

# 2.3 Test invalid login (bad password)
echo -e "\n${BLUE}2.3 Testing invalid login - bad password${NC}"
test_api "POST" "api/v1/login" 401 "Invalid login - bad password" '{"username":"admin","password":"wrongpassword"}'

# 2.4 Test invalid login (missing fields)
echo -e "\n${BLUE}2.4 Testing invalid login - missing fields${NC}"
test_api "POST" "api/v1/login" 400 "Invalid login - missing username" '{"password":"password"}'
test_api "POST" "api/v1/login" 400 "Invalid login - missing password" '{"username":"admin"}'

# 3. Test Entity API Validation
echo -e "\n${BLUE}3. Testing Entity API Validation${NC}"

# 3.1 Test valid entity creation
echo -e "\n${BLUE}3.1 Testing valid entity creation${NC}"
entity_response=$(test_api "POST" "api/v1/entities/create" 200 "Valid entity creation" '{"type":"test","title":"Security Test Entity","tags":["security","test"]}' "$token")

# Extract entity ID from response
entity_id=$(echo "$entity_response" | jq -r '.data.id')
if [ "$entity_id" != "null" ] && [ ! -z "$entity_id" ]; then
  test_result "PASS" "Entity ID extraction (ID: $entity_id)"
else
  test_result "FAIL" "Entity ID extraction (No ID received)"
  # Set a default entity ID for testing
  entity_id="entity_unknown"
fi

# 3.2 Test invalid entity creation (invalid type)
echo -e "\n${BLUE}3.2 Testing invalid entity creation - invalid type${NC}"
test_api "POST" "api/v1/entities/create" 400 "Invalid entity creation - bad type" '{"type":"test with invalid characters!","title":"Invalid Test"}' "$token"

# 3.3 Test invalid entity creation (missing fields)
echo -e "\n${BLUE}3.3 Testing invalid entity creation - missing fields${NC}"
test_api "POST" "api/v1/entities/create" 400 "Invalid entity creation - missing type" '{"title":"Missing Type Entity"}' "$token"
test_api "POST" "api/v1/entities/create" 400 "Invalid entity creation - missing title" '{"type":"test"}' "$token"

# 3.4 Test entity listing
echo -e "\n${BLUE}3.4 Testing entity listing${NC}"
test_api "GET" "api/v1/entities/list" 200 "Entity listing" "" "$token"

# 3.5 Test entity filtering
echo -e "\n${BLUE}3.5 Testing entity filtering${NC}"
test_api "GET" "api/v1/entities/list?type=test" 200 "Entity filtering by type" "" "$token"
test_api "GET" "api/v1/entities/list?tags=security" 200 "Entity filtering by tag" "" "$token"

# 4. Test Entity Relationship API Validation
echo -e "\n${BLUE}4. Testing Entity Relationship API Validation${NC}"

# 4.1 Test valid relationship creation
echo -e "\n${BLUE}4.1 Testing valid relationship creation${NC}"
test_api "POST" "api/v1/entity-relationships/create" 200 "Valid relationship creation" "{\"source_id\":\"$entity_id\",\"target_id\":\"entity_admin\",\"type\":\"test\"}" "$token"

# 4.2 Test invalid relationship creation (invalid source ID)
echo -e "\n${BLUE}4.2 Testing invalid relationship creation - invalid source ID${NC}"
test_api "POST" "api/v1/entity-relationships/create" 400 "Invalid relationship - bad source ID" '{"source_id":"invalid!id","target_id":"entity_admin","type":"test"}' "$token"

# 4.3 Test invalid relationship creation (missing fields)
echo -e "\n${BLUE}4.3 Testing invalid relationship creation - missing fields${NC}"
test_api "POST" "api/v1/entity-relationships/create" 400 "Invalid relationship - missing source" '{"target_id":"entity_admin","type":"test"}' "$token"
test_api "POST" "api/v1/entity-relationships/create" 400 "Invalid relationship - missing target" '{"source_id":"entity_test","type":"test"}' "$token"
test_api "POST" "api/v1/entity-relationships/create" 400 "Invalid relationship - missing type" '{"source_id":"entity_test","target_id":"entity_admin"}' "$token"

# 5. Test Authentication and Authorization
echo -e "\n${BLUE}5. Testing Authentication and Authorization${NC}"

# 5.1 Test access without token
echo -e "\n${BLUE}5.1 Testing access without token${NC}"
test_api "GET" "api/v1/entities/list" 401 "Entity listing without token"

# 5.2 Test with invalid token
echo -e "\n${BLUE}5.2 Testing with invalid token${NC}"
test_api "GET" "api/v1/entities/list" 401 "Entity listing with invalid token" "" "tk_invalid_token"

# 6. Test Audit Logging
echo -e "\n${BLUE}6. Testing Audit Logging${NC}"

# Check if audit logs exist
echo -e "\n${BLUE}6.1 Checking audit log files${NC}"
audit_files=$(find /opt/entitydb/var/log/audit/ -name "audit_*.log" | wc -l)
if [ "$audit_files" -gt 0 ]; then
  test_result "PASS" "Audit log files exist ($audit_files files found)"
else
  test_result "FAIL" "No audit log files found"
fi

# Check audit log entries
echo -e "\n${BLUE}6.2 Checking audit log entries${NC}"
latest_log=$(find /opt/entitydb/var/log/audit/ -name "audit_*.log" -type f -exec ls -t {} \; | head -1)
if [ ! -z "$latest_log" ]; then
  # Count entries in the latest log file
  entry_count=$(cat "$latest_log" | wc -l)
  if [ "$entry_count" -gt 0 ]; then
    test_result "PASS" "Audit log contains entries ($entry_count entries found)"
    # Show the last few entries
    echo -e "${YELLOW}Latest audit log entries:${NC}"
    tail -n 5 "$latest_log"
  else
    test_result "FAIL" "Audit log exists but contains no entries"
  fi
else
  test_result "FAIL" "Could not find latest audit log"
fi

# Stop the entity server
echo -e "\n${BLUE}Stopping EntityDB entity server...${NC}"
kill $SERVER_PID || true
wait $SERVER_PID 2>/dev/null || true

# Print summary
echo -e "\n${BLUE}=== Test Summary ===${NC}"
echo -e "Tests passed: ${GREEN}$PASSED${NC}"
echo -e "Tests failed: ${RED}$FAILED${NC}"
echo -e "Total tests: $((PASSED + FAILED))"

if [ "$FAILED" -eq 0 ]; then
  echo -e "\n${GREEN}All tests passed!${NC}"
  exit 0
else
  echo -e "\n${RED}Some tests failed!${NC}"
  exit 1
fi