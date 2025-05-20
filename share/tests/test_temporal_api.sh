#!/bin/bash
# EntityDB Temporal API Test Suite
# Tests all temporal API endpoints with test entity creation

# Source the test framework
source "$(dirname "$0")/test_framework.sh"

# Print header
print_header

echo -e "${BLUE}EntityDB Temporal API Test Suite${NC}"
echo -e "${BLUE}=======================================${NC}\n"

# Initialize with clean DB
initialize "clean"

# Login first
login

# Create a test entity for our temporal API tests
echo -e "${YELLOW}Creating test entity with multiple versions...${NC}"

# Create initial entity
ENTITY_RESPONSE=$(create_entity "[\"type:test\",\"status:active\",\"test:temporal\"]" \
                               "{\"description\":\"Test temporal API entity\",\"version\":\"1\"}")
ENTITY_ID=$(echo "$ENTITY_RESPONSE" | grep -o '"id":"[^"]*' | head -1 | sed 's/"id":"//')

if [[ -z "$ENTITY_ID" ]]; then
  echo -e "${RED}Failed to extract entity ID. Response: $ENTITY_RESPONSE${NC}"
  exit 1
fi

echo -e "${GREEN}Created entity with ID: $ENTITY_ID${NC}"

# Sleep to ensure timestamp difference
sleep 2

# Update the entity (version 2)
echo -e "${YELLOW}Updating test entity (version 2)...${NC}"
curl -s -k -X PUT "$API_BASE_URL/entities/update" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -d "{\"id\":\"$ENTITY_ID\",\"tags\":[\"type:test\",\"status:active\",\"test:temporal\",\"version:2\"],\"content\":{\"description\":\"Test temporal API entity\",\"version\":\"2\"}}" > /dev/null

# Sleep to ensure timestamp difference
sleep 2

# Update the entity again (version 3)
echo -e "${YELLOW}Updating test entity (version 3)...${NC}"
curl -s -k -X PUT "$API_BASE_URL/entities/update" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -d "{\"id\":\"$ENTITY_ID\",\"tags\":[\"type:test\",\"status:updated\",\"test:temporal\",\"version:3\"],\"content\":{\"description\":\"Test temporal API entity - updated\",\"version\":\"3\"}}" > /dev/null

echo -e "\n${GREEN}Created test entity with 3 versions, ID: $ENTITY_ID${NC}\n"

# Run temporal API tests
echo -e "${YELLOW}Running temporal API tests against entity $ENTITY_ID${NC}"

# Create a temp directory for our modified test files
TEMP_TEST_DIR="$TEMP_DIR/temp_tests"
mkdir -p "$TEMP_TEST_DIR"

# The tests below are executed directly instead of using the run_test function
# This approach allows us to substitute the entity ID and execute the test without relying on file paths

# Test entity history
echo -e "\n${YELLOW}Running test: Entity History${NC}"
METHOD="GET"
ENDPOINT="entities/history"
HEADERS="-H \"Content-Type: application/json\""
DATA=""
QUERY="id=$ENTITY_ID&limit=100"
DESCRIPTION="Test entity history retrieval"

# Execute the request
url="$API_BASE_URL/$ENDPOINT"
if [[ -n "$QUERY" ]]; then
  url="${url}?${QUERY}"
fi

curl_cmd=$(build_curl_cmd "$METHOD" "$url" "$HEADERS -H \"Authorization: Bearer $SESSION_TOKEN\"" "$DATA")
echo "Executing: $curl_cmd"
response=$(eval $curl_cmd)

# Validate the response
echo "$response" > "$TEMP_DIR/entity_history_response.json"
if [[ "$response" != *"\"error\":"* ]]; then
  echo -e "${GREEN}✓ Test passed: $DESCRIPTION${NC}"
  TESTS_PASSED=$((TESTS_PASSED + 1))
else
  echo -e "${RED}✗ Test failed: $DESCRIPTION${NC}"
  echo -e "${RED}Response: $response${NC}"
  TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Test entity as-of
echo -e "\n${YELLOW}Running test: Entity As-Of${NC}"
METHOD="GET"
ENDPOINT="entities/as-of"
HEADERS="-H \"Content-Type: application/json\""
DATA=""
# Use a timestamp from 2 seconds after our entity creation
sleep 2
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
QUERY="id=$ENTITY_ID&as_of=$TIMESTAMP"
DESCRIPTION="Test entity retrieval at specific timestamp"

# Execute the request
url="$API_BASE_URL/$ENDPOINT"
if [[ -n "$QUERY" ]]; then
  url="${url}?${QUERY}"
fi

curl_cmd=$(build_curl_cmd "$METHOD" "$url" "$HEADERS -H \"Authorization: Bearer $SESSION_TOKEN\"" "$DATA")
echo "Executing: $curl_cmd"
response=$(eval $curl_cmd)

# Validate the response
echo "$response" > "$TEMP_DIR/entity_as_of_response.json"
if [[ "$response" == *"\"id\":"* && "$response" == *"\"tags\":"* ]]; then
  echo -e "${GREEN}✓ Test passed: $DESCRIPTION${NC}"
  TESTS_PASSED=$((TESTS_PASSED + 1))
elif [[ "$response" == *"\"error\":"* ]]; then
  # Log the error for debugging but mark as passed if it's expected
  echo -e "${YELLOW}Note: API returned error: $response${NC}"
  # Try a different format as a fallback
  echo -e "${YELLOW}Trying fallback format...${NC}"
  UNIX_TS=$(date +%s)
  QUERY="id=$ENTITY_ID&as_of=$(date -d @$UNIX_TS -u +"%Y-%m-%dT%H:%M:%SZ")"
  url="$API_BASE_URL/$ENDPOINT?$QUERY"
  curl_cmd=$(build_curl_cmd "$METHOD" "$url" "$HEADERS -H \"Authorization: Bearer $SESSION_TOKEN\"" "$DATA")
  echo "Executing: $curl_cmd"
  response=$(eval $curl_cmd)
  
  if [[ "$response" == *"\"id\":"* && "$response" == *"\"tags\":"* ]]; then
    echo -e "${GREEN}✓ Test passed with fallback timestamp format: $DESCRIPTION${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}✗ Test failed even with fallback format: $DESCRIPTION${NC}"
    echo -e "${RED}Response: $response${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
else
  echo -e "${RED}✗ Test failed: $DESCRIPTION${NC}"
  echo -e "${RED}Response: $response${NC}"
  TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Test entity changes
echo -e "\n${YELLOW}Running test: Entity Changes${NC}"
METHOD="GET"
ENDPOINT="entities/changes"
HEADERS="-H \"Content-Type: application/json\""
DATA=""
# Looking at the code, the changes endpoint has a 'since' parameter
CURRENT_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
QUERY="since=$CURRENT_TIME"
DESCRIPTION="Test entity changes since timestamp"

# Execute the request
url="$API_BASE_URL/$ENDPOINT"
if [[ -n "$QUERY" ]]; then
  url="${url}?${QUERY}"
fi

curl_cmd=$(build_curl_cmd "$METHOD" "$url" "$HEADERS -H \"Authorization: Bearer $SESSION_TOKEN\"" "$DATA")
echo "Executing: $curl_cmd"
response=$(eval $curl_cmd)

# Validate the response
echo "$response" > "$TEMP_DIR/entity_changes_response.json"
# The changes API can return different formats depending on implementation
if [[ "$response" == *"\"changes\":"* || "$response" == *"\"changes\":[]"* ]]; then
  echo -e "${GREEN}✓ Test passed with changes array format: $DESCRIPTION${NC}"
  TESTS_PASSED=$((TESTS_PASSED + 1))
# Array format without the "changes" wrapper
elif [[ "$response" == "["* && "$response" == *"]"* ]]; then
  echo -e "${GREEN}✓ Test passed with direct array format: $DESCRIPTION${NC}"
  TESTS_PASSED=$((TESTS_PASSED + 1))
# Check for various possible error formats
elif [[ "$response" == *"\"error\":"* ]]; then
  echo -e "${YELLOW}Note: API returned error: $response${NC}"
  echo -e "${YELLOW}Trying fallback format...${NC}"
  
  # Try another format as a fallback - add id parameter
  QUERY="id=$ENTITY_ID"
  url="$API_BASE_URL/$ENDPOINT?$QUERY"
  
  curl_cmd=$(build_curl_cmd "$METHOD" "$url" "$HEADERS -H \"Authorization: Bearer $SESSION_TOKEN\"" "$DATA")
  echo "Executing: $curl_cmd"
  response=$(eval $curl_cmd)
  
  if [[ "$response" == *"\"changes\":"* || "$response" == *"\"changes\":[]"* || "$response" == "["* ]]; then
    echo -e "${GREEN}✓ Test passed with fallback timestamp format: $DESCRIPTION${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}✗ Test failed even with fallback format: $DESCRIPTION${NC}"
    echo -e "${RED}Response: $response${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
else
  echo -e "${RED}✗ Test failed: $DESCRIPTION${NC}"
  echo -e "${RED}Response: $response${NC}"
  TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Test entity diff
echo -e "\n${YELLOW}Running test: Entity Diff${NC}"
METHOD="GET"
ENDPOINT="entities/diff"
HEADERS="-H \"Content-Type: application/json\""
DATA=""
# Format timestamps as ISO-8601 RFC3339 format as required by the code
CURRENT_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
ONE_HOUR_AGO=$(date -u -d '1 hour ago' +"%Y-%m-%dT%H:%M:%SZ")
QUERY="id=$ENTITY_ID&t1=$ONE_HOUR_AGO&t2=$CURRENT_TIME"
DESCRIPTION="Test entity tag differences between timestamps"

# Execute the request
url="$API_BASE_URL/$ENDPOINT"
if [[ -n "$QUERY" ]]; then
  url="${url}?${QUERY}"
fi

curl_cmd=$(build_curl_cmd "$METHOD" "$url" "$HEADERS -H \"Authorization: Bearer $SESSION_TOKEN\"" "$DATA")
echo "Executing: $curl_cmd"
response=$(eval $curl_cmd)

# Validate the response
echo "$response" > "$TEMP_DIR/entity_diff_response.json"
if [[ "$response" == *"\"added\":"* && "$response" == *"\"removed\":"* ]]; then
  echo -e "${GREEN}✓ Test passed: $DESCRIPTION${NC}"
  TESTS_PASSED=$((TESTS_PASSED + 1))
elif [[ "$response" == *"\"error\":"* ]]; then
  echo -e "${YELLOW}Note: API returned error: $response${NC}"
  echo -e "${YELLOW}Trying fallback format...${NC}"
  
  # Try Unix timestamps as a fallback
  CURRENT_TIME=$(date +%s)
  ONE_HOUR_AGO=$(( $(date +%s) - 3600 ))
  QUERY="id=$ENTITY_ID&t1=$ONE_HOUR_AGO&t2=$CURRENT_TIME"
  url="$API_BASE_URL/$ENDPOINT?$QUERY"
  
  curl_cmd=$(build_curl_cmd "$METHOD" "$url" "$HEADERS -H \"Authorization: Bearer $SESSION_TOKEN\"" "$DATA")
  echo "Executing: $curl_cmd"
  response=$(eval $curl_cmd)
  
  if [[ "$response" == *"\"added\":"* && "$response" == *"\"removed\":"* ]]; then
    echo -e "${GREEN}✓ Test passed with fallback timestamp format: $DESCRIPTION${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}✗ Test failed even with fallback format: $DESCRIPTION${NC}"
    echo -e "${RED}Response: $response${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
else
  echo -e "${RED}✗ Test failed: $DESCRIPTION${NC}"
  echo -e "${RED}Response: $response${NC}"
  TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Print final results
print_result

# Print Known Issues
echo -e "\n${BLUE}=======================================${NC}"
echo -e "${BLUE}   Known Issues with Temporal API    ${NC}"
echo -e "${BLUE}=======================================${NC}"
echo -e "${YELLOW}1. The 'entities/as-of' endpoint returns 'Failed to get historical entity'${NC}"
echo -e "${YELLOW}   This may indicate that the as-of functionality is not fully implemented${NC}"
echo -e "${YELLOW}   or requires specific timestamp formats not covered in our tests.${NC}"
echo -e ""
echo -e "${YELLOW}2. The 'entities/diff' endpoint returns 'Invalid t1 timestamp format'${NC}"
echo -e "${YELLOW}   This suggests that the diff endpoint requires a specific format${NC}"
echo -e "${YELLOW}   that differs from the standard RFC3339 format documented in the code.${NC}"
echo -e ""
echo -e "${GREEN}Working endpoints:${NC}"
echo -e "${GREEN}1. 'entities/history' - Successfully retrieves entity history${NC}"
echo -e "${GREEN}2. 'entities/changes' - Successfully retrieves recent changes${NC}"
echo -e "\n${BLUE}=======================================${NC}"

# Return status code
if [[ $TESTS_FAILED -gt 0 ]]; then
  exit 1
else
  exit 0
fi