#!/bin/bash
# Simple EntityDB Temporal Test

# Configuration
SERVER_URL="https://localhost:8085"
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

# Create an entity with initial content
print_message "$BLUE" "Creating an entity with initial content..."
create_response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:temporal_test", "version:1"],
    "content": "Initial content for temporal testing"
  }')

entity_id=$(echo "$create_response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

if [ -z "$entity_id" ]; then
  print_message "$RED" "❌ Failed to create entity."
  exit 1
else
  print_message "$GREEN" "✅ Created entity with ID: $entity_id"
  print_message "$BLUE" "Initial timestamp: $(echo "$create_response" | grep -o '"created_at":"[^"]*' | cut -d'"' -f4)"
fi

# Sleep to ensure different timestamps
sleep 2

# Update the entity with new content
print_message "$BLUE" "Updating entity with new content..."
update_response=$(curl -k -s -X PUT "$SERVER_URL/api/v1/entities/update" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$entity_id\",
    \"tags\": [\"type:temporal_test\", \"version:2\", \"updated:true\"],
    \"content\": \"Updated content for temporal testing\"
  }")

update_timestamp=$(echo "$update_response" | grep -o '"updated_at":"[^"]*' | cut -d'"' -f4)

if [ -z "$update_timestamp" ]; then
  print_message "$RED" "❌ Failed to update entity."
  exit 1
else
  print_message "$GREEN" "✅ Updated entity with ID: $entity_id"
  print_message "$BLUE" "Update timestamp: $update_timestamp"
fi

# Sleep to ensure different timestamps
sleep 2

# Get entity history
print_message "$BLUE" "Getting entity history..."
history_response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/history?id=$entity_id" \
  -H "Authorization: Bearer $TOKEN")

history_count=$(echo "$history_response" | grep -o '"timestamp"' | wc -l)

if [ "$history_count" -lt 2 ]; then
  print_message "$RED" "❌ Failed to get proper history. Expected at least 2 entries, got $history_count."
else
  print_message "$GREEN" "✅ Got entity history with $history_count entries"
fi

# Get entity as of first timestamp
first_timestamp=$(echo "$create_response" | grep -o '"created_at":"[^"]*' | cut -d'"' -f4)
print_message "$BLUE" "Getting entity as of first timestamp: $first_timestamp..."
asof_response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/as-of?id=$entity_id&timestamp=$first_timestamp" \
  -H "Authorization: Bearer $TOKEN")

asof_tags=$(echo "$asof_response" | grep -o '"tags":\[[^]]*\]' | grep "version:1")

if [ -z "$asof_tags" ]; then
  print_message "$RED" "❌ Failed to get entity as of first timestamp. Expected version:1 tag."
else
  print_message "$GREEN" "✅ Successfully retrieved entity as of first timestamp with correct tags"
fi

# Get entity changes
print_message "$BLUE" "Getting entity changes..."
changes_response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/changes?id=$entity_id" \
  -H "Authorization: Bearer $TOKEN")

changes_count=$(echo "$changes_response" | grep -o '"added"\|"removed"' | wc -l)

if [ "$changes_count" -lt 1 ]; then
  print_message "$RED" "❌ Failed to get proper changes. Expected at least 1 change."
else
  print_message "$GREEN" "✅ Got entity changes with changes detected"
fi

# Get entity diff
print_message "$BLUE" "Getting entity diff between timestamps..."
diff_response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/diff?id=$entity_id&from_timestamp=$first_timestamp&to_timestamp=$update_timestamp" \
  -H "Authorization: Bearer $TOKEN")

diff_content=$(echo "$diff_response" | grep -o '"added_tags"\|"removed_tags"' | wc -l)

if [ "$diff_content" -lt 1 ]; then
  print_message "$RED" "❌ Failed to get proper diff. Expected tag differences."
else
  print_message "$GREEN" "✅ Got entity diff showing tag changes"
fi

# Summary
print_message "$BLUE" "=========== Temporal Test Summary ==========="
print_message "$GREEN" "Entity created with ID: $entity_id"
print_message "$GREEN" "History entries: $history_count"
print_message "$GREEN" "As-of retrieval: $([ -n "$asof_tags" ] && echo "Success" || echo "Failed")"
print_message "$GREEN" "Changes detection: $([ "$changes_count" -ge 1 ] && echo "Success" || echo "Failed")"
print_message "$GREEN" "Diff functionality: $([ "$diff_content" -ge 1 ] && echo "Success" || echo "Failed")"

success_count=0
[ -n "$entity_id" ] && ((success_count++))
[ "$history_count" -ge 2 ] && ((success_count++))
[ -n "$asof_tags" ] && ((success_count++))
[ "$changes_count" -ge 1 ] && ((success_count++))
[ "$diff_content" -ge 1 ] && ((success_count++))

if [ "$success_count" -eq 5 ]; then
  print_message "$GREEN" "✅ All temporal tests PASSED!"
else
  print_message "$RED" "❌ Some temporal tests FAILED! ($success_count/5 passed)"
fi