#!/bin/bash

# Test to verify that metric entities don't trigger recovery

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# EntityDB API URL
API_URL="https://localhost:8085/api/v1"

echo -e "${YELLOW}Testing metric entity recovery fix...${NC}"

# Login
echo "Logging in..."
TOKEN=$(curl -sk -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | \
  jq -r '.token')

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Failed to login${NC}"
    exit 1
fi

echo "Token obtained: ${TOKEN:0:20}..."

# Create a test metric entity directly
echo -e "\nCreating test metric entity..."
METRIC_ID="metric_test_recovery_fix"

# First, check if the metric already exists
EXISTING=$(curl -sk -X GET "$API_URL/entities/get?id=$METRIC_ID" \
  -H "Authorization: Bearer $TOKEN" | \
  jq -r '.error // empty')

if [ -z "$EXISTING" ]; then
    echo "Metric already exists, deleting it first..."
    curl -sk -X DELETE "$API_URL/entities/delete?id=$METRIC_ID" \
        -H "Authorization: Bearer $TOKEN"
fi

# Create the metric
RESPONSE=$(curl -sk -X POST "$API_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$METRIC_ID\",
    \"tags\": [
      \"type:metric\",
      \"dataspace:system\",
      \"name:test_recovery\",
      \"unit:count\",
      \"description:Test metric for recovery fix\",
      \"value:1.00\"
    ],
    \"content\": \"\"
  }")

echo "Create response: $RESPONSE"

# Immediately try to get the metric
echo -e "\nFetching metric immediately after creation..."
METRIC=$(curl -sk -X GET "$API_URL/entities/get?id=$METRIC_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Get response: $METRIC"

# Check if the metric has recovery tags
HAS_RECOVERY=$(echo "$METRIC" | jq -r '.tags[]?' | grep -E "(status:recovered|recovery:partial|recovery:placeholder)" | wc -l)

if [ "$HAS_RECOVERY" -gt 0 ]; then
    echo -e "${RED}FAIL: Metric has recovery tags!${NC}"
    echo "Tags found:"
    echo "$METRIC" | jq -r '.tags[]?' | grep -E "(status:recovered|recovery:partial|recovery:placeholder)"
    exit 1
else
    echo -e "${GREEN}PASS: Metric does not have recovery tags${NC}"
fi

# Check the content
CONTENT=$(echo "$METRIC" | jq -r '.content // empty')
if [[ "$CONTENT" == *"could not be recovered"* ]]; then
    echo -e "${RED}FAIL: Metric has recovery placeholder content!${NC}"
    echo "Content: $CONTENT"
    exit 1
else
    echo -e "${GREEN}PASS: Metric content is correct${NC}"
fi

# Clean up
echo -e "\nCleaning up test metric..."
curl -sk -X DELETE "$API_URL/entities/delete?id=$METRIC_ID" \
    -H "Authorization: Bearer $TOKEN"

echo -e "\n${GREEN}All tests passed! Metric recovery fix is working correctly.${NC}"