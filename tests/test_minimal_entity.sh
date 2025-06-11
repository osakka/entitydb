#!/bin/bash
# Test minimal entity creation

echo "Testing Minimal Entity Creation"
echo "==============================="
echo ""

HOST="${HOST:-https://localhost:8443}"
TOKEN="${1:-}"

if [ -z "$TOKEN" ]; then
    echo "Getting auth token..."
    TOKEN_RESPONSE=$(curl -k -s -X POST "$HOST/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin"}')
    
    TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.token')
    
    if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
        echo "Failed to get token. Response:"
        echo "$TOKEN_RESPONSE"
        exit 1
    fi
    echo "Got token: ${TOKEN:0:20}..."
fi

echo ""
echo "1. Creating a simple test entity..."
SIMPLE_ENTITY=$(cat <<EOF
{
  "id": "test_entity_$(date +%s)",
  "tags": ["type:test", "status:active"],
  "content": "Simple test content"
}
EOF
)

echo "Entity:"
echo "$SIMPLE_ENTITY" | jq .
echo ""

echo "Response:"
curl -k -X POST "$HOST/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$SIMPLE_ENTITY" \
  -w "\nHTTP Status: %{http_code}\n" \
  2>/dev/null | jq . || echo "Failed"

echo ""
echo "2. Creating dashboard layout entity..."
DASHBOARD_ENTITY=$(cat <<EOF
{
  "id": "dashboard_test_$(date +%s)",
  "tags": ["type:dashboard_layout", "user:admin", "version:1"],
  "content": "{\\"widgets\\":[],\\"theme\\":\\"light\\",\\"lastModified\\":\\"2025-01-01T00:00:00Z\\"}"
}
EOF
)

echo "Entity:"
echo "$DASHBOARD_ENTITY" | jq .
echo ""

echo "Response:"
curl -k -X POST "$HOST/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$DASHBOARD_ENTITY" \
  -w "\nHTTP Status: %{http_code}\n" \
  2>/dev/null | jq . || echo "Failed"

echo ""
echo "3. List all entities (first 5)..."
curl -k -s -X GET "$HOST/api/v1/entities/list" \
  -H "Authorization: Bearer $TOKEN" \
  2>/dev/null | jq -r '.[] | .id' | head -5

echo ""
echo "==============================="