#!/bin/bash
# Test entity listing API

# Include common test utilities
source "$(dirname "$0")/../test_utils.sh"

# Set up test environment
setup_test "Entity Listing API"

# Login as admin to get an auth token
TOKEN=$(login_admin)

# Test 1: List all entities
echo "TEST: Listing all entities"

RESPONSE=$(curl -s -X GET -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/entities")

# Verify response contains entities array
if [[ -n $(echo $RESPONSE | jq -r 'if type=="array" then "array" else empty end') ]]; then
  ENTITY_COUNT=$(echo $RESPONSE | jq 'length')
  pass "Successfully listed $ENTITY_COUNT entities"
else
  fail "Failed to list entities: $RESPONSE"
fi

# Test 2: Create a uniquely tagged entity for testing
echo "SETUP: Creating an entity with unique tag for filtering test"

UNIQUE_TAG="unique_test_tag_$(date +%s)"
ENTITY_DATA="{
  \"tags\": [\"type=test\", \"status=active\", \"test_tag=$UNIQUE_TAG\"],
  \"content\": [
    {\"type\": \"title\", \"value\": \"Test Entity With Unique Tag\"},
    {\"type\": \"description\", \"value\": \"This entity has a unique tag for testing filtering\"}
  ]
}"

RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d "$ENTITY_DATA" \
  "$BASE_URL/api/v1/entities")

# Extract entity ID from response
ENTITY_ID=$(echo $RESPONSE | jq -r '.id')

# Verify entity ID is present
if [[ -z "$ENTITY_ID" || "$ENTITY_ID" == "null" ]]; then
  fail "Failed to create tagged entity: $RESPONSE"
else
  pass "Tagged entity created with ID: $ENTITY_ID"
fi

# Test 3: List entities with tag filter
echo "TEST: Listing entities with tag filter"

RESPONSE=$(curl -s -X GET -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/entities?tag=test_tag%3D$UNIQUE_TAG")

# Verify we get exactly one entity with our unique tag
ENTITY_COUNT=$(echo $RESPONSE | jq 'length')
if [[ "$ENTITY_COUNT" -eq 1 ]]; then
  FOUND_ID=$(echo $RESPONSE | jq -r '.[0].id')
  if [[ "$FOUND_ID" == "$ENTITY_ID" ]]; then
    pass "Successfully filtered entity by tag"
  else
    fail "Filtered entity ID doesn't match created entity ID"
  fi
else
  fail "Expected 1 entity with unique tag, got $ENTITY_COUNT: $RESPONSE"
fi

# Test 4: Test listing entities with type filter
echo "TEST: Listing entities with type filter"

RESPONSE=$(curl -s -X GET -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/entities?tag=type%3Dtest")

# Verify we get at least one entity with type=test
ENTITY_COUNT=$(echo $RESPONSE | jq 'length')
if [[ "$ENTITY_COUNT" -gt 0 ]]; then
  pass "Successfully filtered entities by type, found $ENTITY_COUNT"
else
  fail "No entities found with type=test: $RESPONSE"
fi

# Summarize test results
summarize_tests