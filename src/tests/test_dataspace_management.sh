#!/bin/bash
# Test dataset management API

BASE_URL="https://localhost:8085/api/v1"
CURL_OPTS="-k -s"

echo "=== EntityDB Dataset Management Test ==="
echo

# Login as admin
echo "1. Logging in as admin..."
LOGIN_RESPONSE=$(curl $CURL_OPTS -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Failed to login"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi
echo "✓ Logged in successfully"
echo

# List existing datasets
echo "2. Listing existing datasets..."
LIST_RESPONSE=$(curl $CURL_OPTS -X GET $BASE_URL/datasets \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $LIST_RESPONSE" | jq .
echo

# Create a new dataset
echo "3. Creating 'worca' dataset..."
CREATE_RESPONSE=$(curl $CURL_OPTS -X POST $BASE_URL/datasets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "worca",
    "description": "Workforce Orchestrator Application Dataset",
    "settings": {
      "theme": "oceanic",
      "features": "kanban,projects,teams"
    }
  }')
echo "Response: $CREATE_RESPONSE" | jq .
DATASPACE_ID=$(echo $CREATE_RESPONSE | jq -r '.id')
echo

# Get the created dataset
echo "4. Getting dataset details..."
GET_RESPONSE=$(curl $CURL_OPTS -X GET $BASE_URL/datasets/$DATASPACE_ID \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $GET_RESPONSE" | jq .
echo

# Update the dataset
echo "5. Updating dataset..."
UPDATE_RESPONSE=$(curl $CURL_OPTS -X PUT $BASE_URL/datasets/$DATASPACE_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "worca",
    "description": "Workforce Orchestrator - Ocean-Inspired Task Management",
    "settings": {
      "theme": "oceanic",
      "features": "kanban,projects,teams,analytics",
      "version": "1.0"
    }
  }')
echo "Response: $UPDATE_RESPONSE" | jq .
echo

# Create another dataset for testing
echo "6. Creating 'test' dataset..."
CREATE_TEST_RESPONSE=$(curl $CURL_OPTS -X POST $BASE_URL/datasets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test",
    "description": "Test Dataset",
    "settings": {}
  }')
TEST_DATASPACE_ID=$(echo $CREATE_TEST_RESPONSE | jq -r '.id')
echo "Response: $CREATE_TEST_RESPONSE" | jq .
echo

# List all datasets
echo "7. Listing all datasets..."
LIST_ALL_RESPONSE=$(curl $CURL_OPTS -X GET $BASE_URL/datasets \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $LIST_ALL_RESPONSE" | jq .
echo

# Delete the test dataset
echo "8. Deleting test dataset..."
DELETE_RESPONSE=$(curl $CURL_OPTS -X DELETE $BASE_URL/datasets/$TEST_DATASPACE_ID \
  -H "Authorization: Bearer $TOKEN")
echo "Delete status: $?"
echo

# List datasets again to confirm deletion
echo "9. Listing datasets after deletion..."
LIST_FINAL_RESPONSE=$(curl $CURL_OPTS -X GET $BASE_URL/datasets \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $LIST_FINAL_RESPONSE" | jq .
echo

echo "=== Test Complete ==="