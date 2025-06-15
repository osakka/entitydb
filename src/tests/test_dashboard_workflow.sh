#!/bin/bash
# Test complete dashboard workflow

echo "Testing Dashboard Workflow"
echo "========================="
echo ""

HOST="${HOST:-https://localhost:8085}"
USERNAME="admin"
PASSWORD="admin"

# 1. Login
echo "1. Login as admin..."
LOGIN_RESPONSE=$(curl -k -s -X POST "$HOST/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "Failed to login"
    echo "$LOGIN_RESPONSE"
    exit 1
fi

echo "✓ Login successful"
echo "Token: ${TOKEN:0:20}..."
echo ""

# 2. Create dashboard layout
echo "2. Creating dashboard layout..."
LAYOUT_JSON=$(cat <<EOF
{
  "id": "dashboard_layout_${USERNAME}",
  "tags": [
    "type:dashboard_layout",
    "user:${USERNAME}",
    "version:1"
  ],
  "content": "{\"widgets\":[{\"id\":\"widget-123\",\"type\":\"metrics\",\"size\":\"medium\",\"config\":{\"title\":\"System Metrics\"}}],\"theme\":\"light\",\"lastModified\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}"
}
EOF
)

CREATE_RESPONSE=$(curl -k -s -X POST "$HOST/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$LAYOUT_JSON")

CREATE_STATUS=$(echo "$CREATE_RESPONSE" | jq -r '.id // empty')

if [ -n "$CREATE_STATUS" ]; then
    echo "✓ Dashboard layout created"
else
    echo "Create response: $CREATE_RESPONSE"
    
    # Try update if already exists
    echo "Trying to update existing layout..."
    UPDATE_RESPONSE=$(curl -k -s -X PUT "$HOST/api/v1/entities/update?id=dashboard_layout_${USERNAME}" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"tags\":[\"type:dashboard_layout\",\"user:${USERNAME}\",\"version:1\"],\"content\":\"{\\\"widgets\\\":[{\\\"id\\\":\\\"widget-123\\\",\\\"type\\\":\\\"metrics\\\",\\\"size\\\":\\\"medium\\\",\\\"config\\\":{\\\"title\\\":\\\"System Metrics\\\"}}],\\\"theme\\\":\\\"light\\\",\\\"lastModified\\\":\\\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\\\"}\"}")
    
    echo "Update response: $UPDATE_RESPONSE"
fi
echo ""

# 3. List dashboard layouts
echo "3. Listing dashboard layouts..."
LIST_RESPONSE=$(curl -k -s -X GET "$HOST/api/v1/entities/list?tag=type:dashboard_layout" \
    -H "Authorization: Bearer $TOKEN")

LAYOUT_COUNT=$(echo "$LIST_RESPONSE" | jq '. | length')
echo "✓ Found $LAYOUT_COUNT dashboard layout(s)"

# Show user's layout
echo ""
echo "4. User's dashboard layout:"
echo "$LIST_RESPONSE" | jq -r ".[] | select(.tags[] | contains(\"user:$USERNAME\")) | {id: .id, content: .content}"

echo ""
echo "========================="
echo "Workflow complete!"
echo ""
echo "To test in browser:"
echo "1. Open $HOST in your browser"
echo "2. Login as admin/admin"
echo "3. Go to Dashboard tab"
echo "4. Click 'Add Widget' and add a widget"
echo "5. The layout should auto-save"
echo "6. Refresh the page - the widget should persist"