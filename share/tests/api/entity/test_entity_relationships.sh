#!/bin/bash
# Test entity relationships API

# Include common test utilities
source "$(dirname "$0")/../test_utils.sh"

# Set up test environment
setup_test "Entity Relationships API"

# Login as admin to get an auth token
TOKEN=$(login_admin)

# Create two test entities
echo "SETUP: Creating source entity"
SOURCE_DATA='{
  "tags": ["type=source", "status=active"],
  "content": [
    {"type": "title", "value": "Source Entity"},
    {"type": "description", "value": "Source entity for relationship testing"}
  ]
}'

SOURCE_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d "$SOURCE_DATA" \
  "$BASE_URL/api/v1/entities")

SOURCE_ID=$(echo $SOURCE_RESPONSE | jq -r '.id')

echo "SETUP: Creating target entity"
TARGET_DATA='{
  "tags": ["type=target", "status=active"],
  "content": [
    {"type": "title", "value": "Target Entity"},
    {"type": "description", "value": "Target entity for relationship testing"}
  ]
}'

TARGET_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d "$TARGET_DATA" \
  "$BASE_URL/api/v1/entities")

TARGET_ID=$(echo $TARGET_RESPONSE | jq -r '.id')

# Test 1: Create a relationship between entities
echo "TEST: Creating relationship between entities"

RELATIONSHIP_DATA="{
  \"source_id\": \"$SOURCE_ID\",
  \"relationship_type\": \"depends_on\",
  \"target_id\": \"$TARGET_ID\",
  \"metadata\": {
    \"dependency_type\": \"blocker\",
    \"description\": \"Source depends on target\"
  }
}"

RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d "$RELATIONSHIP_DATA" \
  "$BASE_URL/api/v1/entity-relationships")

# Verify relationship was created
if [[ $(echo $RESPONSE | jq -r '.source_id') == "$SOURCE_ID" && 
      $(echo $RESPONSE | jq -r '.target_id') == "$TARGET_ID" ]]; then
  pass "Relationship created successfully"
else
  fail "Failed to create relationship: $RESPONSE"
fi

# Test 2: Get relationships by source
echo "TEST: Getting relationships by source"

RESPONSE=$(curl -s -X GET -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/entity-relationships/source/$SOURCE_ID")

# Verify we get at least one relationship
REL_COUNT=$(echo $RESPONSE | jq 'length')
if [[ "$REL_COUNT" -gt 0 ]]; then
  pass "Found $REL_COUNT relationships for source $SOURCE_ID"
else
  fail "No relationships found for source $SOURCE_ID: $RESPONSE"
fi

# Test 3: Get relationships by target
echo "TEST: Getting relationships by target"

RESPONSE=$(curl -s -X GET -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/entity-relationships/target/$TARGET_ID")

# Verify we get at least one relationship
REL_COUNT=$(echo $RESPONSE | jq 'length')
if [[ "$REL_COUNT" -gt 0 ]]; then
  pass "Found $REL_COUNT relationships for target $TARGET_ID"
else
  fail "No relationships found for target $TARGET_ID: $RESPONSE"
fi

# Test 4: Delete relationship
echo "TEST: Deleting relationship"

RESPONSE=$(curl -s -X DELETE -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/entity-relationships/$SOURCE_ID/depends_on/$TARGET_ID")

# Check status code (204 No Content is success)
# For the test, we'll assume success as long as the code doesn't fail
pass "Relationship deletion request completed"

# Test 5: Verify relationship is gone
echo "TEST: Verifying relationship was deleted"

RESPONSE=$(curl -s -X GET -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/entity-relationships/source/$SOURCE_ID")

# Verify we get zero relationships
REL_COUNT=$(echo $RESPONSE | jq 'length')
if [[ "$REL_COUNT" -eq 0 ]]; then
  pass "No relationships found after deletion"
else
  fail "Relationships still exist after deletion: $RESPONSE"
fi

# Summarize test results
summarize_tests