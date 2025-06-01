#!/bin/bash
# EntityDB Simple Stress Test

# Configuration
SERVER_URL="http://localhost:8085"
ENTITY_COUNT=50
TOKEN=""

# Color output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function for colored output
print_message() {
  echo -e "${1}${2}${NC}"
}

# Login to get token
print_message "$BLUE" "Logging in to EntityDB..."
response=$(curl -k -s -X POST "$SERVER_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')
  
TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  print_message "$RED" "❌ Failed to login."
  exit 1
else
  print_message "$GREEN" "✅ Login successful, got token"
fi

# Create entities in a loop
print_message "$BLUE" "Creating $ENTITY_COUNT entities..."
success_count=0

for i in $(seq 1 $ENTITY_COUNT); do
  # Generate simple content
  content="Test content for entity $i"
  
  # Create the entity
  response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"tags\": [\"type:stress_test\", \"number:$i\"],
      \"content\": \"$content\"
    }")
  
  # Check if successful
  entity_id=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
  if [ -n "$entity_id" ]; then
    ((success_count++))
    
    # Show progress
    if [ $((i % 10)) -eq 0 ]; then
      print_message "$GREEN" "Created $i entities so far..."
    fi
  fi
done

print_message "$GREEN" "✅ Successfully created $success_count/$ENTITY_COUNT entities"

# Run a query
print_message "$BLUE" "Running query for all stress test entities..."
start_time=$(date +%s.%N)
response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/list?tag=type:stress_test" \
  -H "Authorization: Bearer $TOKEN")
end_time=$(date +%s.%N)
query_time=$(echo "$end_time - $start_time" | bc)
entity_count=$(echo "$response" | grep -o "\"id\":" | wc -l)
print_message "$GREEN" "✅ Query returned $entity_count entities in $query_time seconds"

# Verify by retrieving a specific entity
test_number=50
print_message "$BLUE" "Retrieving test entity #$test_number..."
response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/list?tag=number:$test_number" \
  -H "Authorization: Bearer $TOKEN")
entity_found=$(echo "$response" | grep -o "\"id\":" | wc -l)

if [ "$entity_found" -gt 0 ]; then
  print_message "$GREEN" "✅ Entity #$test_number successfully retrieved"
else
  print_message "$RED" "❌ Failed to retrieve entity #$test_number"
fi

print_message "$BLUE" "Stress test completed."