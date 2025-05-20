#!/bin/bash
# Test sequence that runs multiple related tests in sequence

# Source the test framework
source "$(dirname "$0")/test_framework.sh"

# Print header
print_header

# Initialize with clean DB
initialize "clean" 

# Login first - this sets the SESSION_TOKEN for subsequent tests
login

# Store entity creation results so we can extract the ID for later tests
echo -e "${YELLOW}Creating test entity...${NC}"
ENTITY_RESPONSE=$(create_entity "[\"type:test\",\"status:active\",\"test:sequence\"]" "{\"description\":\"Test sequence entity\",\"created_by\":\"test_framework\"}")
ENTITY_ID=$(echo "$ENTITY_RESPONSE" | grep -o '"id":"[^"]*' | head -1 | sed 's/"id":"//')

if [[ -z "$ENTITY_ID" ]]; then
  echo -e "${RED}Failed to extract entity ID. Response: $ENTITY_RESPONSE${NC}"
  exit 1
fi

echo -e "${GREEN}Created entity with ID: $ENTITY_ID${NC}"

# Update the history test file to use our entity ID
HISTORY_TEST_FILE="$TEST_DIR/entity_history.test"
if [[ -f "$HISTORY_TEST_FILE" ]]; then
  # Create a temporary file
  TEMP_FILE="$(mktemp)"
  # Update the QUERY line with our entity ID
  cat "$HISTORY_TEST_FILE" | sed "s/QUERY=\"id=ENTITY_ID/QUERY=\"id=$ENTITY_ID/" > "$TEMP_FILE"
  # Replace the original file
  mv "$TEMP_FILE" "$HISTORY_TEST_FILE"
  
  echo -e "${YELLOW}Updated history test to use entity ID: $ENTITY_ID${NC}"
fi

# Now run the tests in sequence
run_test "list_entities" 
run_test "entity_history"

# Additional tests could be run here...

# Print final results
print_result