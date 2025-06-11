#!/bin/bash
# Test script for widget dashboard functionality

echo "======================================"
echo "Widget Dashboard Test"
echo "======================================"
echo ""

# Check if server is running
echo "1. Checking if EntityDB server is running..."
if pgrep -f "entitydb" > /dev/null; then
    echo "✓ Server is running"
else
    echo "✗ Server is not running. Please start with: cd /opt/entitydb/src && ./bin/entitydbd.sh start"
    exit 1
fi

echo ""
echo "2. Server Status:"
echo "Navigate to: http://localhost:8085"
echo ""
echo "3. Widget System Tests:"
echo ""
echo "Dashboard Features to Test:"
echo "- Default tab should be 'Dashboard'"
echo "- Click 'Add Widget' button to open widget gallery"
echo "- Click on any widget to add it to dashboard"
echo "- Drag widgets to reorder them"
echo "- Double-click widget headers to cycle sizes (small/medium/large)"
echo "- Click 'Save Layout' to persist widget arrangement"
echo "- Click 'Reset' to return to default layout"
echo ""
echo "Expected Default Widgets:"
echo "- System Overview (medium)"
echo "- Health Score (small)"
echo "- Operations (small)"
echo "- Performance Trends (large)"
echo "- Storage (small)"
echo ""
echo "Console Debugging:"
echo "Open browser console (F12) to see:"
echo "- 'Initializing widget system...'"
echo "- 'Widget system setup complete'"
echo "- 'Adding widget to dashboard: [type]' when adding widgets"
echo ""
echo "Known Issues Fixed:"
echo "- Widget system now initializes on dashboard load"
echo "- Click handlers properly attached to widget previews"
echo "- Button sizing normalized in dashboard header"
echo "- Default layout loads 5 initial widgets"
echo ""
echo "======================================"