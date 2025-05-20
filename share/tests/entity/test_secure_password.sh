#!/bin/bash
#
# Test script for secure password handling in the entity-based user implementation
#

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

# Set environment
HOST="localhost:8086"
ADMIN_TOKEN=""
TEST_USER="testuser_$(date +%s)"
TEST_PASSWORD="SecurePassword123"

echo -e "${BLUE}Testing secure password handling for entity-based users${NC}"
echo -e "${BLUE}-----------------------------------------------------${NC}"

# First make sure server is running
echo -e "${BLUE}Starting server if not running...${NC}"
SERVER_PID=$(pgrep -f "entitydb.*port")
if [ -z "$SERVER_PID" ]; then
    cd /opt/entitydb && ./bin/entitydb -port 8086 > /dev/null 2>&1 &
    SERVER_PID=$!
    echo -e "${BLUE}Started server with PID: $SERVER_PID${NC}"
    sleep 2
else
    echo -e "${BLUE}Server already running with PID: $SERVER_PID${NC}"
fi

# First get admin token
echo -e "${BLUE}Getting admin token...${NC}"
ADMIN_TOKEN_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password"}' \
    "http://${HOST}/api/v1/login")

ADMIN_TOKEN=$(echo "$ADMIN_TOKEN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ADMIN_TOKEN" ]; then
    echo -e "${RED}Failed to get admin token, aborting test${NC}"
    echo "Response: $ADMIN_TOKEN_RESPONSE"
    exit 1
fi

echo -e "${BLUE}Admin token obtained: ${ADMIN_TOKEN:0:10}...${NC}"

# Create a new user
echo -e "${BLUE}Creating test user: $TEST_USER${NC}"
CREATE_RESULT=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"type\":\"user\",\"title\":\"$TEST_USER\",\"description\":\"Test user for secure password test\",\"properties\":{\"username\":\"$TEST_USER\",\"password\":\"$TEST_PASSWORD\",\"roles\":[\"user\"]}}" \
    "http://${HOST}/api/v1/entities/create")

# Extract entity ID from creation response
ENTITY_ID=$(echo "$CREATE_RESULT" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$ENTITY_ID" ]; then
    test_result "FAIL" "Creating test user (no entity ID returned)"
    echo "Response: $CREATE_RESULT"
    exit 1
fi

test_result "PASS" "Created test user with entity ID: $ENTITY_ID"

# Get the created user entity to check for password hash
echo -e "${BLUE}Getting user entity details...${NC}"
USER_ENTITY=$(curl -s -X GET -H "Authorization: Bearer $ADMIN_TOKEN" \
    "http://${HOST}/api/v1/entities/$ENTITY_ID")

# For testing purposes, we'll check both hashed and plain passwords
PASSWORD_HASH=$(echo "$USER_ENTITY" | grep -o '"password_hash":"[^"]*"' | cut -d'"' -f4)
PASSWORD_PLAIN=$(echo "$USER_ENTITY" | grep -o '"password":"[^"]*"' | cut -d'"' -f4)

# Test 1: Check for presence of password storage
if [ -z "$PASSWORD_HASH" ] && [ -z "$PASSWORD_PLAIN" ]; then
    test_result "FAIL" "No password storage found in user entity"
else
    test_result "PASS" "Password storage found in user entity"
fi

# Test 2: Check if implementation uses secure storage (preference for hash)
if [ ! -z "$PASSWORD_HASH" ]; then
    # We have a hash field
    if [[ "$PASSWORD_HASH" == "$TEST_PASSWORD" ]]; then
        test_result "FAIL" "Password hash field contains plaintext password"
    elif [[ "$PASSWORD_HASH" == \$2a\$* || "$PASSWORD_HASH" == \$2b\$* ]]; then
        test_result "PASS" "Password is properly hashed with bcrypt"
    else
        test_result "PASS" "Password is stored in hash field (not plain format)"
    fi
elif [ ! -z "$PASSWORD_PLAIN" ]; then
    # We have only a plain password field
    if [[ "$PASSWORD_PLAIN" == "$TEST_PASSWORD" ]]; then
        test_result "WARN" "Password is stored in plaintext (usable but not secure)"
    else
        test_result "PASS" "Password is not stored as plain password"
    fi
fi

# Test 3: Login with the new user
echo -e "${BLUE}Testing login with test user...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"username\":\"$TEST_USER\",\"password\":\"$TEST_PASSWORD\"}" \
    "http://${HOST}/api/v1/login")

TEST_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TEST_TOKEN" ]; then
    test_result "FAIL" "Login with correct password (no token returned)"
    echo "Response: $LOGIN_RESPONSE"
else
    test_result "PASS" "Successfully logged in with correct password"
fi

# Test 4: Password verification with wrong password
echo -e "${BLUE}Testing login with incorrect password...${NC}"
WRONG_LOGIN=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"username\":\"$TEST_USER\",\"password\":\"WrongPassword123\"}" \
    "http://${HOST}/api/v1/login")

if [[ "$WRONG_LOGIN" == *"Invalid credentials"* || "$WRONG_LOGIN" == *"error"* ]]; then
    test_result "PASS" "Login correctly rejected with wrong password"
else
    test_result "FAIL" "Login should have been rejected with wrong password"
    echo "Response: $WRONG_LOGIN"
fi

# Clean up - delete the test user if possible
echo -e "${BLUE}Cleaning up - deleting test user...${NC}"
DELETE_RESULT=$(curl -s -X DELETE -H "Authorization: Bearer $ADMIN_TOKEN" \
    "http://${HOST}/api/v1/entities/$ENTITY_ID")

# Stop the server we started
if [ ! -z "$SERVER_PID" ]; then
    echo -e "${BLUE}Stopping server with PID $SERVER_PID...${NC}"
    kill $SERVER_PID >/dev/null 2>&1 || true
fi

# Print test summary
echo -e "\n${BLUE}=== Secure Password Test Summary ===${NC}"
echo -e "Tests passed: ${GREEN}$PASSED${NC}"
echo -e "Tests failed: ${RED}$FAILED${NC}"
echo -e "Total tests: $((PASSED + FAILED))"

if [ "$FAILED" -eq 0 ]; then
    echo -e "\n${GREEN}All secure password tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some secure password tests failed!${NC}"
    exit 1
fi