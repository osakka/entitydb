#!/bin/bash
#
# Test script for the EntityDB API
# Performs basic tests to ensure the API is working correctly

# Output formatting
green='\033[0;32m'
red='\033[0;31m'
yellow='\033[0;33m'
clear='\033[0m'

SERVER="http://localhost:8085"
TOKEN_FILE="/tmp/entitydb_test_token.txt"

# Default test credentials
USERNAME="admin"
PASSWORD="password"

echo -e "${yellow}EntityDB API Test Script${clear}"
echo "================================="
echo

# Store test results
PASS_COUNT=0
FAIL_COUNT=0

# Function to mark a test as passed
pass_test() {
  echo -e "${green}✓ PASS:${clear} $1"
  PASS_COUNT=$((PASS_COUNT + 1))
}

# Function to mark a test as failed
fail_test() {
  echo -e "${red}✗ FAIL:${clear} $1"
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

# Test 1: Server availability
echo "Test 1: Server availability"
if curl -s -o /dev/null -w "%{http_code}" "${SERVER}" | grep -q "200\|301\|302"; then
  pass_test "Server is accessible at ${SERVER}"
else
  fail_test "Cannot access server at ${SERVER}"
  echo "Make sure the EntityDB server is running."
  exit 1
fi
echo

# Test 2: Authentication
echo "Test 2: Authentication"
AUTH_RESPONSE=$(curl -s -X POST "${SERVER}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"${USERNAME}\",\"password\":\"${PASSWORD}\"}")

if echo "${AUTH_RESPONSE}" | grep -q "token"; then
  pass_test "Authentication successful"
  # Extract and save token
  TOKEN=$(echo "${AUTH_RESPONSE}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
  if [ -z "${TOKEN}" ]; then
    # Try alternative format
    TOKEN=$(echo "${AUTH_RESPONSE}" | grep -o '"data":{[^}]*"token":"[^"]*"' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
  fi
  echo "${TOKEN}" > "${TOKEN_FILE}"
  echo "Token saved to ${TOKEN_FILE}"
else
  fail_test "Authentication failed: ${AUTH_RESPONSE}"
  echo "Please check your credentials."
  exit 1
fi
echo

# Load the token for subsequent tests
TOKEN=$(cat "${TOKEN_FILE}")

# Test 3: Fetch entities
echo "Test 3: Fetch entities"
ENTITIES_RESPONSE=$(curl -s -X GET "${SERVER}/api/v1/entities" \
  -H "Authorization: Bearer ${TOKEN}")

if echo "${ENTITIES_RESPONSE}" | grep -q '"status":"ok"'; then
  pass_test "Successfully fetched entities"
  # Count entities
  if echo "${ENTITIES_RESPONSE}" | grep -q '"data":\['; then
    ENTITY_COUNT=$(echo "${ENTITIES_RESPONSE}" | grep -o '"data":\[[^]]*\]' | grep -o '{' | wc -l)
    echo "Found ${ENTITY_COUNT} entities"
  else
    echo "No entities found or unexpected response format"
  fi
else
  fail_test "Failed to fetch entities: ${ENTITIES_RESPONSE}"
fi
echo

# Test 4: Create an entity
echo "Test 4: Create an entity"
CREATE_RESPONSE=$(curl -s -X POST "${SERVER}/api/v1/entities" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{
    "type": "issue",
    "title": "API Test Issue",
    "description": "This issue was created by the API test script",
    "tags": ["test", "api", "status:pending", "priority:low"],
    "properties": {
      "test_run": true,
      "created_by": "test_script"
    }
  }')

if echo "${CREATE_RESPONSE}" | grep -q '"status":"ok"'; then
  pass_test "Successfully created test entity"
  # Extract entity ID
  ENTITY_ID=$(echo "${CREATE_RESPONSE}" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  if [ -n "${ENTITY_ID}" ]; then
    echo "Created entity with ID: ${ENTITY_ID}"
  else
    echo "Entity created but could not extract ID"
  fi
else
  fail_test "Failed to create entity: ${CREATE_RESPONSE}"
  ENTITY_ID=""
fi
echo

# Test 5: Fetch entity by ID (if we have one)
if [ -n "${ENTITY_ID}" ]; then
  echo "Test 5: Fetch entity by ID"
  GET_RESPONSE=$(curl -s -X GET "${SERVER}/api/v1/entities/${ENTITY_ID}" \
    -H "Authorization: Bearer ${TOKEN}")
  
  if echo "${GET_RESPONSE}" | grep -q '"status":"ok"'; then
    pass_test "Successfully fetched entity by ID"
  else
    fail_test "Failed to fetch entity by ID: ${GET_RESPONSE}"
  fi
  echo
  
  # Test 6: Update the entity
  echo "Test 6: Update entity"
  UPDATE_RESPONSE=$(curl -s -X PUT "${SERVER}/api/v1/entities/${ENTITY_ID}" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}" \
    -d '{
      "title": "Updated API Test Issue",
      "tags": ["test", "api", "status:in_progress", "priority:medium"],
      "properties": {
        "test_run": true,
        "created_by": "test_script",
        "updated": true
      }
    }')
  
  if echo "${UPDATE_RESPONSE}" | grep -q '"status":"ok"'; then
    pass_test "Successfully updated entity"
  else
    fail_test "Failed to update entity: ${UPDATE_RESPONSE}"
  fi
  echo
fi

# Test 7: Entity relationships
if [ -n "${ENTITY_ID}" ]; then
  echo "Test 7: Entity relationships"
  
  # Find another entity to create a relationship with
  OTHER_ENTITY_ID=$(curl -s -X GET "${SERVER}/api/v1/entities?limit=1" \
    -H "Authorization: Bearer ${TOKEN}" | 
    grep -o '"id":"[^"]*"' | grep -v "${ENTITY_ID}" | head -1 | cut -d'"' -f4)
  
  if [ -n "${OTHER_ENTITY_ID}" ]; then
    echo "Found other entity with ID: ${OTHER_ENTITY_ID}"
    
    # Create relationship
    REL_RESPONSE=$(curl -s -X POST "${SERVER}/api/v1/entity-relationships" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer ${TOKEN}" \
      -d "{
        \"source_id\": \"${ENTITY_ID}\",
        \"target_id\": \"${OTHER_ENTITY_ID}\",
        \"type\": \"depends_on\",
        \"properties\": {
          \"test_run\": true
        }
      }")
    
    if echo "${REL_RESPONSE}" | grep -q '"status":"ok"'; then
      pass_test "Successfully created relationship"
      
      # List relationships
      LIST_REL_RESPONSE=$(curl -s -X GET "${SERVER}/api/v1/entity-relationships?source_id=${ENTITY_ID}" \
        -H "Authorization: Bearer ${TOKEN}")
      
      if echo "${LIST_REL_RESPONSE}" | grep -q '"status":"ok"'; then
        pass_test "Successfully listed relationships"
      else
        fail_test "Failed to list relationships: ${LIST_REL_RESPONSE}"
      fi
    else
      fail_test "Failed to create relationship: ${REL_RESPONSE}"
    fi
  else
    echo "Could not find another entity for relationship test"
  fi
  echo
fi

# Test 8: Clean up (optional)
if [ -n "${ENTITY_ID}" ]; then
  echo "Test 8: Clean up test entity"
  DELETE_RESPONSE=$(curl -s -X DELETE "${SERVER}/api/v1/entities/${ENTITY_ID}" \
    -H "Authorization: Bearer ${TOKEN}")
  
  if echo "${DELETE_RESPONSE}" | grep -q '"status":"ok"'; then
    pass_test "Successfully deleted test entity"
  else
    fail_test "Failed to delete test entity: ${DELETE_RESPONSE}"
  fi
  echo
fi

# Summary
echo "================================="
echo -e "Test Summary: ${green}${PASS_COUNT} passed${clear}, ${red}${FAIL_COUNT} failed${clear}"

# Clean up token file
rm -f "${TOKEN_FILE}"
echo "Cleaned up temporary token file"

if [ "${FAIL_COUNT}" -eq 0 ]; then
  echo -e "${green}All tests passed!${clear}"
  exit 0
else
  echo -e "${red}Some tests failed.${clear} Please check the output for details."
  exit 1
fi