#!/bin/bash
# Simple Temporal API Test Script for EntityDB

# Configuration
SERVER_URL="https://localhost:8085"  # Make sure this matches the actual server URL

# Login and get token
echo "Logging in to get token..."
LOGIN_RESPONSE=$(curl -s -k -X POST "$SERVER_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Failed to login: $LOGIN_RESPONSE"
  exit 1
else
  echo "Login successful, got token"
fi

# Create a test entity
echo "Creating test entity..."
CREATE_RESPONSE=$(curl -s -k -X POST "$SERVER_URL/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:temporal_test", "test:v1"],
    "content": "Initial version content"
  }')

ENTITY_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

if [ -z "$ENTITY_ID" ]; then
  echo "Failed to create entity: $CREATE_RESPONSE"
  exit 1
else
  echo "Created entity with ID: $ENTITY_ID"
fi

# Wait a moment to ensure timestamp difference
sleep 1

# Update the entity
echo "Updating entity..."
sleep 2  # Ensure the entity is fully created before updating
UPDATE_RESPONSE=$(curl -s -k -X PUT "$SERVER_URL/api/v1/entities/update" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$ENTITY_ID\",
    \"tags\": [\"type:temporal_test\", \"test:v2\"],
    \"content\": \"Updated version content\"
  }")
echo "Update response: $UPDATE_RESPONSE"

if [[ "$UPDATE_RESPONSE" == *"error"* ]]; then
  echo "Failed to update entity: $UPDATE_RESPONSE"
  exit 1
else
  echo "Updated entity"
fi

# Get entity history
echo "Getting entity history..."
HISTORY_RESPONSE=$(curl -s -k -X GET "$SERVER_URL/api/v1/entities/history?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")

if [[ "$HISTORY_RESPONSE" == *"timestamp"* ]]; then
  echo "Got entity history: $HISTORY_RESPONSE"
  
  # Count number of versions based on timestamp entries
  VERSION_COUNT=$(echo "$HISTORY_RESPONSE" | grep -o '"timestamp"' | wc -l)
  echo "Entity has $VERSION_COUNT timestamp entries"
  
  if [ "$VERSION_COUNT" -ge 1 ]; then
    echo "TEMPORAL HISTORY TEST PASSED!"
  else
    echo "TEMPORAL HISTORY TEST FAILED! Expected at least 1 timestamp entry, got $VERSION_COUNT"
  fi
else
  echo "Failed to get entity history: $HISTORY_RESPONSE"
  exit 1
fi

# Get current timestamp - add 5 minutes to ensure it's future from when entity was created
CURRENT_TIME=$(date -u -d "+5 minutes" +"%Y-%m-%dT%H:%M:%SZ")

# Get entity as of a past time
echo "Getting entity as of current time..."
AS_OF_RESPONSE=$(curl -s -k -X GET "$SERVER_URL/api/v1/entities/as-of?id=$ENTITY_ID&as_of=$CURRENT_TIME" \
  -H "Authorization: Bearer $TOKEN")

if [[ "$AS_OF_RESPONSE" != *"error"* ]]; then
  echo "Got entity as of current time: $AS_OF_RESPONSE"
  echo "TEMPORAL AS-OF TEST PASSED!"
else
  echo "Failed to get entity as-of: $AS_OF_RESPONSE"
  echo "TEMPORAL AS-OF TEST FAILED!"
fi

echo "All tests completed"