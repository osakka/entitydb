#!/bin/bash

# Test advanced query endpoint
echo "Testing EntityDB Advanced Query"

API_BASE="http://localhost:8085/api/v1"

# First login to get a token
echo "Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "${API_BASE}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "Failed to login. Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "Got token: ${TOKEN:0:20}..."
AUTH_HEADER="Authorization: Bearer $TOKEN"

# Test the auth
echo "Testing auth with token..."
TEST_RESPONSE=$(curl -s -X GET "${API_BASE}/auth/status" \
  -H "$AUTH_HEADER")
echo "Auth status: $TEST_RESPONSE"

# Create some test entities first
echo "Creating test entities..."

# Create entity 1
echo "Creating entity 1..."
curl -v -X POST "${API_BASE}/entities/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "tags": ["type:document", "status:active", "priority:high"],
    "content": [{"type": "title", "value": "Test Document 1"}]
  }' 2>&1 | grep -A 10 -B 10 "error"

# Create entity 2
curl -s -X POST "${API_BASE}/entities/create" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "tags": ["type:document", "status:inactive", "priority:low"],
    "content": [{"type": "title", "value": "Test Document 2"}]
  }' | jq '.'

# Create entity 3
curl -s -X POST "${API_BASE}/entities/create" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "tags": ["type:task", "status:active", "priority:high"],
    "content": [{"type": "title", "value": "Test Task 1"}]
  }' | jq '.'

echo -e "\n\nTesting advanced queries..."

# Test 1: Query by tag
echo -e "\n1. Query entities with tag type:document"
curl -s -X GET "${API_BASE}/entities/query?filter=tag:type&operator=eq&value=document" \
  -H "$AUTH_HEADER" | jq '.'

# Test 2: Query with sorting
echo -e "\n2. Query all entities sorted by created_at descending"
curl -s -X GET "${API_BASE}/entities/query?sort=created_at&order=desc&limit=5" \
  -H "$AUTH_HEADER" | jq '.'

# Test 3: Query with tag count filter
echo -e "\n3. Query entities with more than 2 tags"
curl -s -X GET "${API_BASE}/entities/query?filter=tag_count&operator=gt&value=2" \
  -H "$AUTH_HEADER" | jq '.'

# Test 4: Query with content filter
echo -e "\n4. Query entities with title content type"
curl -s -X GET "${API_BASE}/entities/query?filter=content_type&operator=eq&value=title" \
  -H "$AUTH_HEADER" | jq '.'

# Test 5: Query with pagination
echo -e "\n5. Query with pagination (limit 2, offset 1)"
curl -s -X GET "${API_BASE}/entities/query?limit=2&offset=1" \
  -H "$AUTH_HEADER" | jq '.'

echo -e "\nAdvanced query tests completed!"