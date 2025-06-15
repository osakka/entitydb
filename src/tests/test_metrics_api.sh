#!/bin/bash

# Test EntityDB Metrics API

echo "🔍 Testing EntityDB Metrics API..."

# Use HTTPS with -k to ignore certificate warnings
BASE_URL="https://localhost:8085"
CURL_OPTS="-k -s"

# Test 1: Health endpoint
echo -e "\n📊 Test 1: Health endpoint"
curl $CURL_OPTS $BASE_URL/health | jq '.'

# Test 2: Prometheus metrics endpoint
echo -e "\n📊 Test 2: Prometheus metrics endpoint"
curl $CURL_OPTS $BASE_URL/metrics | head -20

# Test 3: System metrics endpoint (requires auth)
echo -e "\n📊 Test 3: System metrics endpoint"

# First login to get token
echo "Logging in..."
LOGIN_RESPONSE=$(curl $CURL_OPTS -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')

if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
    echo "✅ Login successful, fetching system metrics..."
    
    # Get system metrics
    curl $CURL_OPTS $BASE_URL/api/v1/system/metrics \
      -H "Authorization: Bearer $TOKEN" | jq '.'
else
    echo "❌ Login failed, response:"
    echo "$LOGIN_RESPONSE" | jq '.'
fi

echo -e "\n✅ Metrics API test complete!"