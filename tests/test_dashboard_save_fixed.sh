#!/bin/bash
# Test dashboard save functionality after fixes

echo "======================================"
echo "Dashboard Save Fix Test"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Changes Made to Fix Save Issues:${NC}"
echo ""
echo -e "${GREEN}1. Entity ID Format:${NC}"
echo "   - Changed from: dashboard_layout_user_9eb10d2e774c1aacbf13a73f1a669a66"
echo "   - Changed to: dashboard_layout_admin (using username)"
echo ""
echo -e "${GREEN}2. Tag Format:${NC}"
echo "   - Changed from: user:user_9eb10d2e774c1aacbf13a73f1a669a66"
echo "   - Changed to: user:admin (using username)"
echo ""
echo -e "${GREEN}3. List API Fix:${NC}"
echo "   - Changed from: /api/v1/entities/list?tags=type:dashboard_layout,user:..."
echo "   - Changed to: /api/v1/entities/list?tag=type:dashboard_layout&tag=user:admin"
echo ""
echo -e "${GREEN}4. Update API Fix:${NC}"
echo "   - Changed from: PUT /api/v1/entities/update (with ID in body)"
echo "   - Changed to: PUT /api/v1/entities/update?id=dashboard_layout_admin"
echo "   - Body now only contains tags and content"
echo ""

echo -e "${YELLOW}Testing Steps:${NC}"
echo ""
echo "1. Clear browser cache and refresh page"
echo "2. Login as admin/admin"
echo "3. Go to Dashboard tab"
echo "4. Add a widget"
echo "5. Check console for:"
echo "   - 'Fetching dashboard layout for user: admin'"
echo "   - 'Dashboard auto-saved' or 'Dashboard auto-saved (updated)'"
echo "6. Refresh page - layout should persist"
echo ""

echo -e "${YELLOW}API Calls to Monitor (Network Tab):${NC}"
echo ""
echo "GET  /api/v1/entities/list?tag=type:dashboard_layout&tag=user:admin"
echo "POST /api/v1/entities/create (first save)"
echo "PUT  /api/v1/entities/update?id=dashboard_layout_admin (subsequent saves)"
echo ""

echo -e "${YELLOW}Entity Structure:${NC}"
echo ""
echo "{"
echo "  \"id\": \"dashboard_layout_admin\","
echo "  \"tags\": ["
echo "    \"type:dashboard_layout\","
echo "    \"user:admin\","
echo "    \"version:1\""
echo "  ],"
echo "  \"content\": \"{\\\"widgets\\\":[...],\\\"theme\\\":\\\"light\\\",\\\"lastModified\\\":\\\"...\\\"}\""
echo "}"
echo ""

echo -e "${RED}If Still Having Issues:${NC}"
echo "1. Check server logs for 'Failed to list entities' errors"
echo "2. Verify user object has 'username' property"
echo "3. Check if entity creation permissions are working"
echo "4. Try manual entity creation with curl to test API"
echo ""

echo "======================================"