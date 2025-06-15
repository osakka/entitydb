#!/bin/bash
# Debug dashboard save issues

echo "Testing EntityDB Dashboard Save"
echo "==============================="
echo ""

# Test variables
TOKEN="${1:-}"
HOST="https://localhost:8443"
USERNAME="admin"

if [ -z "$TOKEN" ]; then
    echo "Usage: $0 <auth_token>"
    echo ""
    echo "To get token:"
    echo "1. Open browser developer tools (F12)"
    echo "2. Go to Application/Storage -> Local Storage"
    echo "3. Find 'entitydb-admin-token' and copy the value"
    echo ""
    echo "Or login via curl:"
    echo "curl -k -X POST $HOST/api/v1/auth/login -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"admin\"}'"
    exit 1
fi

echo "1. Testing entity list endpoint..."
echo "Command: curl -k -X GET \"$HOST/api/v1/entities/list?tag=type:dashboard_layout&tag=user:$USERNAME\" -H \"Authorization: Bearer $TOKEN\""
echo ""
curl -k -X GET "$HOST/api/v1/entities/list?tag=type:dashboard_layout&tag=user:$USERNAME" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  2>/dev/null | jq . || echo "Failed to parse JSON"

echo ""
echo "2. Testing entity creation..."
ENTITY_JSON=$(cat <<EOF
{
  "id": "dashboard_layout_${USERNAME}",
  "tags": [
    "type:dashboard_layout",
    "user:${USERNAME}",
    "version:1"
  ],
  "content": "{\"widgets\":[],\"theme\":\"light\",\"lastModified\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}"
}
EOF
)

echo "Entity to create:"
echo "$ENTITY_JSON" | jq .
echo ""

echo "Command: curl -k -X POST \"$HOST/api/v1/entities/create\" -H \"Authorization: Bearer $TOKEN\" -d '<entity>'"
echo ""
curl -k -X POST "$HOST/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$ENTITY_JSON" \
  -w "\nHTTP Status: %{http_code}\n" \
  2>/dev/null | jq . || echo "Failed to parse JSON"

echo ""
echo "3. Testing with different query formats..."
echo ""

# Try with wildcard
echo "Testing wildcard query: tag=type:dashboard_layout*"
curl -k -X GET "$HOST/api/v1/entities/list?tag=type:dashboard_layout*" \
  -H "Authorization: Bearer $TOKEN" \
  -w "\nHTTP Status: %{http_code}\n" \
  2>/dev/null | jq . || echo "Failed to parse JSON"

echo ""
echo "4. Testing simple list all entities..."
curl -k -X GET "$HOST/api/v1/entities/list" \
  -H "Authorization: Bearer $TOKEN" \
  -w "\nHTTP Status: %{http_code}\n" \
  2>/dev/null | head -20

echo ""
echo "==============================="
echo "Debug complete"