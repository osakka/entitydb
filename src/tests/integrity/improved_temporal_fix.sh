#!/bin/bash
# Improved script to test and demonstrate the ListByTag temporal fix

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8085"
TAG_TO_TEST="type:test"  # The tag we'll search for
TOKEN=""

# Function for colored output
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

# Login to get token
login() {
  print_message "$BLUE" "Logging in to EntityDB..."
  
  local response=$(curl -s -X POST "$SERVER_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}')
  
  TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
  
  if [ -z "$TOKEN" ]; then
    print_message "$RED" "❌ Failed to login. Response: $response"
    exit 1
  else
    print_message "$GREEN" "✅ Login successful, got token"
  fi
}

# Test the tag search functionality
test_tag_search() {
  print_message "$BLUE" "Creating a test entity with tag '$TAG_TO_TEST'..."
  
  # Generate a unique test ID
  TEST_ID=$(date +%s)
  
  # Create a test entity
  create_response=$(curl -s -X POST "$SERVER_URL/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"tags\": [\"$TAG_TO_TEST\", \"test:id:$TEST_ID\"],
      \"content\": \"Test entity for temporal tag fix at $(date)\"
    }")
  
  # Extract entity ID
  new_id=$(echo "$create_response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
  if [ -z "$new_id" ]; then
    print_message "$RED" "❌ Failed to create test entity: $create_response"
    return 1
  else
    print_message "$GREEN" "✅ Created entity with ID: $new_id"
  fi
  
  # Wait for indexing to complete
  print_message "$BLUE" "Waiting 2 seconds for indexing..."
  sleep 2
  
  # Test searching by the generic tag
  print_message "$BLUE" "Searching for entities with tag '$TAG_TO_TEST'..."
  search_response=$(curl -s -X GET "$SERVER_URL/api/v1/entities/list?tag=$TAG_TO_TEST" \
    -H "Authorization: Bearer $TOKEN")
  
  # Get the count of returned entities
  found_count=$(echo "$search_response" | grep -o '"id"' | wc -l)
  
  if [ $found_count -gt 0 ]; then
    print_message "$GREEN" "✅ Found $found_count entities with tag '$TAG_TO_TEST'"
    
    # Check if our newly created entity is in the results
    if echo "$search_response" | grep -q "$new_id"; then
      print_message "$GREEN" "✅ Successfully found the entity we just created!"
    else
      print_message "$YELLOW" "⚠️ Entity we created was not found in the results."
    fi
  else
    print_message "$RED" "❌ No entities found with tag '$TAG_TO_TEST'"
  fi
  
  # Now test searching by the unique test ID tag
  print_message "$BLUE" "Searching for entity with unique tag 'test:id:$TEST_ID'..."
  specific_search=$(curl -s -X GET "$SERVER_URL/api/v1/entities/list?tag=test:id:$TEST_ID" \
    -H "Authorization: Bearer $TOKEN")
  
  # Check if our entity is found by the specific tag
  if echo "$specific_search" | grep -q "$new_id"; then
    print_message "$GREEN" "✅ Successfully found the entity by its unique test tag!"
  else
    print_message "$RED" "❌ Entity not found by its unique test tag. This indicates a temporal tag issue."
  fi
  
  # Return the entity ID for later use
  echo "$new_id"
}

# Test temporal features to see if they properly handle temporal tags
test_temporal_features() {
  local entity_id=$1
  if [ -z "$entity_id" ]; then
    print_message "$RED" "❌ No entity ID provided to test temporal features."
    return 1
  fi
  
  print_message "$BLUE" "Testing temporal features on entity $entity_id..."
  
  # Get current time in RFC3339 format
  now=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  
  # Test as-of endpoint with current time
  print_message "$BLUE" "Testing as-of endpoint with current time..."
  as_of_response=$(curl -s -X GET "$SERVER_URL/api/v1/entities/as-of?id=$entity_id&as_of=$now" \
    -H "Authorization: Bearer $TOKEN")
  
  if echo "$as_of_response" | grep -q "$entity_id"; then
    print_message "$GREEN" "✅ As-of endpoint working correctly!"
  else
    print_message "$RED" "❌ As-of endpoint failed to return the entity."
  fi
  
  # Test history endpoint 
  print_message "$BLUE" "Testing history endpoint..."
  history_response=$(curl -s -X GET "$SERVER_URL/api/v1/entities/history?id=$entity_id" \
    -H "Authorization: Bearer $TOKEN")
  
  if [[ "$history_response" == *"timestamp"* ]]; then
    print_message "$GREEN" "✅ History endpoint working correctly!"
  else
    print_message "$RED" "❌ History endpoint failed to return the entity history."
  fi
}

# Main execution
print_message "$BLUE" "========================================"
print_message "$BLUE" "Improved EntityDB Temporal Tag Fix Test"
print_message "$BLUE" "========================================"

# Login
login

# Run the tag search test and get the entity ID
entity_id=$(test_tag_search)

# Test temporal features if an entity was created
if [ -n "$entity_id" ]; then
  test_temporal_features "$entity_id"
fi

print_message "$BLUE" "========================================"
print_message "$BLUE" "Test complete!"
print_message "$BLUE" "========================================"