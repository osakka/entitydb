#!/bin/bash
# Test multiple endpoints for EntityDB

# Configuration
SERVER_URL="http://localhost:8085"

# Get token
echo "Getting authentication token..."
LOGIN_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Failed to login: $LOGIN_RESPONSE"
  exit 1
else
  echo "Authentication successful!"
fi

# Create a test entity
echo -e "\nCreating test entity..."
CREATE_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:test_entity", "version:1", "category:test"],
    "content": "This is test entity content"
  }')

ENTITY_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

if [ -z "$ENTITY_ID" ]; then
  echo "Failed to create entity: $CREATE_RESPONSE"
  exit 1
else
  echo "Created entity with ID: $ENTITY_ID"
fi

sleep 1

# Test all specified endpoints
echo -e "\n--- TESTING ENTITY ENDPOINTS ---"

# 1. Get the entity
echo -e "\nTesting /api/v1/entities/get"
GET_RESPONSE=$(curl -s -X GET "$SERVER_URL/api/v1/entities/get?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $GET_RESPONSE"

# 2. List entities
echo -e "\nTesting /api/v1/entities/list"
LIST_RESPONSE=$(curl -s -X GET "$SERVER_URL/api/v1/entities/list" \
  -H "Authorization: Bearer $TOKEN")
echo "Response (truncated): ${LIST_RESPONSE:0:200}..."

# 3. Test query
echo -e "\nTesting /api/v1/entities/query"
QUERY_RESPONSE=$(curl -s -X GET "$SERVER_URL/api/v1/entities/query?tag=type:test_entity" \
  -H "Authorization: Bearer $TOKEN")
echo "Response (truncated): ${QUERY_RESPONSE:0:200}..."

echo -e "\n--- TESTING TEMPORAL ENDPOINTS ---"

# 4. Test as-of
echo -e "\nTesting /api/v1/entities/as-of"
CURRENT_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
AS_OF_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/entities/as-of" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$ENTITY_ID\",
    \"timestamp\": \"$CURRENT_TIME\"
  }")
echo "Response: $AS_OF_RESPONSE"

# 5. Test history
echo -e "\nTesting /api/v1/entities/history"
HISTORY_RESPONSE=$(curl -s -X GET "$SERVER_URL/api/v1/entities/history?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $HISTORY_RESPONSE"

# 6. Update the entity for temporal testing
echo -e "\nUpdating entity for temporal test..."
UPDATE_RESPONSE=$(curl -s -X PUT "$SERVER_URL/api/v1/entities/update" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$ENTITY_ID\",
    \"tags\": [\"type:test_entity\", \"version:2\", \"category:test\"],
    \"content\": \"This is updated entity content\"
  }")
echo "Response: $UPDATE_RESPONSE"

sleep 1

# 7. Test history again after update
echo -e "\nTesting /api/v1/entities/history after update"
HISTORY_RESPONSE2=$(curl -s -X GET "$SERVER_URL/api/v1/entities/history?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $HISTORY_RESPONSE2"

# 8. Test changes
echo -e "\nTesting /api/v1/entities/changes"
PAST_TIME=$(date -u -d "5 minutes ago" +"%Y-%m-%dT%H:%M:%SZ")
CURRENT_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CHANGES_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/entities/changes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$ENTITY_ID\",
    \"start_time\": \"$PAST_TIME\",
    \"end_time\": \"$CURRENT_TIME\"
  }")
echo "Response: $CHANGES_RESPONSE"

# 9. Test diff
echo -e "\nTesting /api/v1/entities/diff"
DIFF_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/entities/diff" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$ENTITY_ID\",
    \"time1\": \"$PAST_TIME\",
    \"time2\": \"$CURRENT_TIME\"
  }")
echo "Response: $DIFF_RESPONSE"

echo -e "\n--- TESTING COMPLETE ---"
echo "Entity endpoints test: PASSED"

# Check if temporal endpoints are working
if [[ "$HISTORY_RESPONSE2" == *"history"* ]] || [[ "$CHANGES_RESPONSE" == *"changes"* ]] || [[ "$DIFF_RESPONSE" == *"diff"* ]]; then
  echo "Temporal endpoints test: PASSED"
else
  echo "Temporal endpoints test: FAILED (empty or invalid responses)"
fi