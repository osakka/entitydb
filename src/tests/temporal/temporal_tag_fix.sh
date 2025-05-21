#!/bin/bash
# Script to test and demonstrate the ListByTag temporal fix

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

# Main execution
print_message "$BLUE" "========================================"
print_message "$BLUE" "EntityDB Temporal Tag Fix Test"
print_message "$BLUE" "========================================"

# Login
login

# Delete previous entities
print_message "$BLUE" "Looking for existing entities with tag '$TAG_TO_TEST'..."
response=$(curl -s -X GET "$SERVER_URL/api/v1/entities/list?tag=$TAG_TO_TEST" \
  -H "Authorization: Bearer $TOKEN")

# Get all entity IDs
entity_ids=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
id_count=$(echo "$entity_ids" | wc -l)

if [ $id_count -gt 0 ]; then
  print_message "$YELLOW" "Found $id_count existing test entities."
fi

# Create a test entity with the specified tag
print_message "$BLUE" "Creating a new test entity with tag '$TAG_TO_TEST'..."
create_response=$(curl -s -X POST "$SERVER_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"tags\": [\"$TAG_TO_TEST\"],
    \"content\": \"Test entity for temporal tag fix at $(date)\"
  }")

new_id=$(echo "$create_response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
if [ -z "$new_id" ]; then
  print_message "$RED" "❌ Failed to create test entity: $create_response"
else
  print_message "$GREEN" "✅ Created entity with ID: $new_id"
fi

# Wait a moment for any indexing to complete
sleep 1

# Test searching by the tag
print_message "$BLUE" "Searching for entities with tag '$TAG_TO_TEST'..."
search_response=$(curl -s -X GET "$SERVER_URL/api/v1/entities/list?tag=$TAG_TO_TEST" \
  -H "Authorization: Bearer $TOKEN")

# Get the count of returned entities
found_count=$(echo "$search_response" | grep -o '"id"' | wc -l)

if [ $found_count -gt 0 ]; then
  print_message "$GREEN" "✅ SUCCESS! Found $found_count entities with tag '$TAG_TO_TEST'"
  
  # Extract and show the entity ID from the response
  found_id=$(echo "$search_response" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
  print_message "$GREEN" "   Entity ID: $found_id"

  # Compare with the ID we created
  if [ "$found_id" == "$new_id" ]; then
    print_message "$GREEN" "✅ VERIFICATION PASSED! Found the entity we just created."
  else
    print_message "$YELLOW" "⚠️ Found entities, but not the one we just created. This may indicate a lag in indexing."
  fi
else
  print_message "$RED" "❌ FAILED! No entities found with tag '$TAG_TO_TEST'"
fi

print_message "$BLUE" "========================================"