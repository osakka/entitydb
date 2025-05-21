#!/bin/bash
# Simple script to fix the temporal tag indexing issue

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8085"

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
    print_message "$GREEN" "✅ Login successful, got token: $TOKEN"
  fi
}

# Main execution
print_message "$BLUE" "========================================"
print_message "$BLUE" "Starting EntityDB Temporal Tag Fix Test"
print_message "$BLUE" "========================================"

# Login first
login

# Create a test entity with a tag
print_message "$BLUE" "Creating a test entity with tag 'type:test'..."
curl -s -X POST "$SERVER_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags": ["type:test"], "content": "Test entity for tag fix"}'

# Before we fixed the tag index, this search would return zero results
# Now, with our fix, it should find the entity
print_message "$BLUE" "Searching for entities with tag 'type:test'..."
response=$(curl -s -X GET "$SERVER_URL/api/v1/entities/list?tag=type:test" \
  -H "Authorization: Bearer $TOKEN")

count=$(echo "$response" | grep -o '"id"' | wc -l)

if [ "$count" -gt 0 ]; then
  print_message "$GREEN" "✅ FIX SUCCESSFUL! Found $count entities with tag 'type:test'"
else
  print_message "$RED" "❌ FIX FAILED! No entities found with tag 'type:test'"
fi

print_message "$BLUE" "========================================"