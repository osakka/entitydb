#!/bin/bash
# Test dashboard save functionality

echo "======================================"
echo "Dashboard Save Functionality Test"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Testing Dashboard Save Features:${NC}"
echo ""

echo "1. Open browser console (F12) to see debug messages"
echo ""

echo "2. Test Save Operations:"
echo "   - Add a widget → Check console for 'Layout changed' message"
echo "   - Remove a widget → Layout auto-saves"
echo "   - Resize widget (double-click header) → Layout auto-saves"
echo "   - Drag to reorder → Layout auto-saves on drop"
echo "   - Click 'Save Layout' → Manual save with notification"
echo ""

echo "3. Console Messages to Look For:"
echo -e "   ${GREEN}✓${NC} 'Fetching dashboard layout for user: admin'"
echo -e "   ${GREEN}✓${NC} 'Layout changed: [array of widgets]'"
echo -e "   ${GREEN}✓${NC} 'Saving dashboard layout...'"
echo -e "   ${GREEN}✓${NC} 'Current layout to save: [array]'"
echo -e "   ${GREEN}✓${NC} 'Dashboard layout saved successfully' or 'Layout exists, updating...'"
echo ""

echo "4. Testing Save Persistence:"
echo "   a. Add/remove/reorder widgets"
echo "   b. Click 'Save Layout' button"
echo "   c. Refresh the page (F5)"
echo "   d. Your layout should be restored"
echo ""

echo "5. API Endpoints Used:"
echo "   - GET  /api/v1/entities/list?tags=type:dashboard_layout,user:admin"
echo "   - POST /api/v1/entities/create (first save)"
echo "   - PUT  /api/v1/entities/update (subsequent saves)"
echo ""

echo "6. Entity Storage Format:"
echo "   ID: dashboard_layout_admin"
echo "   Tags: type:dashboard_layout, user:admin, version:1"
echo "   Content: JSON with widgets array and metadata"
echo ""

echo -e "${YELLOW}Note:${NC} The save function now:"
echo "- Uses current widget system layout (not stale data)"
echo "- Handles user ID properly (falls back to username)"
echo "- Shows notifications for success/failure"
echo "- Auto-saves on widget changes (add/remove/resize/reorder)"
echo "- Updates existing layouts instead of creating duplicates"
echo ""

echo "======================================"