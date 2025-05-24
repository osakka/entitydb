#!/bin/bash
# Test dataspace management API

BASE_URL="https://localhost:8085/api/v1"
CURL_OPTS="-k -s"

echo "=== EntityDB Dataspace Management Test ==="
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

# List existing dataspaces
echo "2. Listing existing dataspaces..."
LIST_RESPONSE=$(curl $CURL_OPTS -X GET $BASE_URL/dataspaces \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $LIST_RESPONSE" | jq .
echo

# Create a new dataspace
echo "3. Creating 'worca' dataspace..."
CREATE_RESPONSE=$(curl $CURL_OPTS -X POST $BASE_URL/dataspaces \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "worca",
    "description": "Workforce Orchestrator Application Dataspace",
    "settings": {
      "theme": "oceanic",
      "features": "kanban,projects,teams"
    }
  }')
echo "Response: $CREATE_RESPONSE" | jq .
DATASPACE_ID=$(echo $CREATE_RESPONSE | jq -r '.id')
echo

# Get the created dataspace
echo "4. Getting dataspace details..."
GET_RESPONSE=$(curl $CURL_OPTS -X GET $BASE_URL/dataspaces/$DATASPACE_ID \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $GET_RESPONSE" | jq .
echo

# Update the dataspace
echo "5. Updating dataspace..."
UPDATE_RESPONSE=$(curl $CURL_OPTS -X PUT $BASE_URL/dataspaces/$DATASPACE_ID \
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

# Create another dataspace for testing
echo "6. Creating 'test' dataspace..."
CREATE_TEST_RESPONSE=$(curl $CURL_OPTS -X POST $BASE_URL/dataspaces \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test",
    "description": "Test Dataspace",
    "settings": {}
  }')
TEST_DATASPACE_ID=$(echo $CREATE_TEST_RESPONSE | jq -r '.id')
echo "Response: $CREATE_TEST_RESPONSE" | jq .
echo

# List all dataspaces
echo "7. Listing all dataspaces..."
LIST_ALL_RESPONSE=$(curl $CURL_OPTS -X GET $BASE_URL/dataspaces \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $LIST_ALL_RESPONSE" | jq .
echo

# Delete the test dataspace
echo "8. Deleting test dataspace..."
DELETE_RESPONSE=$(curl $CURL_OPTS -X DELETE $BASE_URL/dataspaces/$TEST_DATASPACE_ID \
  -H "Authorization: Bearer $TOKEN")
echo "Delete status: $?"
echo

# List dataspaces again to confirm deletion
echo "9. Listing dataspaces after deletion..."
LIST_FINAL_RESPONSE=$(curl $CURL_OPTS -X GET $BASE_URL/dataspaces \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $LIST_FINAL_RESPONSE" | jq .
echo

echo "=== Test Complete ==="