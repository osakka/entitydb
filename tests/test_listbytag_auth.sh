#!/bin/bash

# Test ListByTag endpoint to see if it can find admin user

echo "Testing ListByTag for admin user"
echo "================================"
echo ""

# Test direct ListByTag query
echo "1. Testing ListByTag for identity:username:admin"
curl -k -s "https://localhost:8085/api/v1/entities/listbytag?tag=identity:username:admin" | jq . || echo "Failed to query"

echo ""
echo "2. Testing ListByTag for type:user"
curl -k -s "https://localhost:8085/api/v1/entities/listbytag?tag=type:user" | jq . || echo "Failed to query"

echo ""
echo "3. Testing raw ListByTag endpoint with minimal request"
curl -k -s -X GET "https://localhost:8085/api/v1/entities/listbytag?tag=identity:username:admin" \
    -H "Accept: application/json" \
    -w "\nHTTP Status: %{http_code}\nTime: %{time_total}s\n"

echo ""
echo "4. Checking if temporal tags are the issue"
# The server stores tags as TIMESTAMP|tag internally
# But the API should handle this transparently
echo "Searching for admin users using list endpoint:"
curl -k -s "https://localhost:8085/api/v1/entities/list" | jq '.[] | select(.tags[]? | contains("username:admin"))' 2>/dev/null | jq -r '.id' | head -5