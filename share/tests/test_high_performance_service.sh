#!/bin/bash

# Test Turbo Service Functionality

API_URL="http://localhost:8085/api/v1"

echo "=== Testing EntityDB Turbo Service ==="

# 1. Login
echo -n "Testing login... "
TOKEN=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r .token)

if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
  echo "✓ Success"
else
  echo "✗ Failed"
  exit 1
fi

# 2. Get entity by ID
echo -n "Testing get by ID... "
RESULT=$(curl -s -X GET "$API_URL/entities/get?id=admin" \
  -H "Authorization: Bearer $TOKEN")

if [[ "$RESULT" == *"admin"* ]]; then
  echo "✓ Success"
else
  echo "✗ Failed"
  echo "$RESULT"
fi

# 3. List entities
echo -n "Testing list entities... "
RESULT=$(curl -s -X GET "$API_URL/entities/list" \
  -H "Authorization: Bearer $TOKEN")

if [[ "$RESULT" == *"entities"* ]]; then
  echo "✓ Success"
else
  echo "✗ Failed"
  echo "$RESULT"
fi

# 4. Create entity
echo -n "Testing create entity... "
RESULT=$(curl -s -X POST "$API_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:test", "status:active"],
    "content": [{"type":"name","value":"Test Entity"}]
  }')

ENTITY_ID=$(echo "$RESULT" | jq -r .id)
if [ "$ENTITY_ID" != "null" ] && [ -n "$ENTITY_ID" ]; then
  echo "✓ Success (ID: $ENTITY_ID)"
else
  echo "✗ Failed"
  echo "$RESULT"
fi

# 5. Query entities
echo -n "Testing query... "
RESULT=$(curl -s -X GET "$API_URL/entities/query?filter=type:test" \
  -H "Authorization: Bearer $TOKEN")

if [[ "$RESULT" == *"entities"* ]]; then
  echo "✓ Success"
else
  echo "✗ Failed"
  echo "$RESULT"
fi

# 6. Check turbo mode
echo -n "Checking turbo mode... "
LOG_ENTRY=$(grep "TurboEntityRepository" /opt/entitydb/var/entitydb.log | tail -1)
if [[ "$LOG_ENTRY" == *"TurboEntityRepository"* ]]; then
  echo "✓ Turbo mode active"
else
  echo "✗ Turbo mode not detected"
fi

echo -e "\n=== Service Test Complete ==="