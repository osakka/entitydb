# Dashboard Troubleshooting Guide

## Quick Debug Steps

### 1. Check in Browser Console
Open browser developer tools (F12) and check for these messages when you switch to Dashboard tab:

```
[Dashboard] Setting up widget system...
[Dashboard] Container found: <div>
[Dashboard] Widget system created
[Dashboard] Widgets registered
[Dashboard] Loading dashboard layout...
```

### 2. Test Pages Available

1. **Simple Debug Page**: https://localhost:8085/dashboard-debug.html
   - Tests authentication, entity API, and widget system separately
   - Step-by-step buttons to isolate issues

2. **Widget System Test**: https://localhost:8085/widget-test.html
   - Tests just the widget system without EntityDB complexity
   - Verifies widget-system.js is working

3. **Main Dashboard**: https://localhost:8085/ (Dashboard tab)
   - Should now show at least a "Test Widget" if working

### 3. Common Issues & Fixes

#### Issue: Blank Dashboard
**Check:** Browser console for errors
**Fix:** 
- Clear browser cache and reload
- Check if widget-system.js loaded (Network tab)
- Look for [Dashboard] error messages

#### Issue: "Widget container not found"
**Check:** Dashboard tab is fully loaded
**Fix:** The system will retry automatically after 500ms

#### Issue: No widgets showing
**Check:** Look for "Test Widget Working!" message
**Fix:** The test widget should always show if system is working

### 4. Manual Test in Console

While on the Dashboard tab, paste this in browser console:

```javascript
// Check if widget system exists
console.log('Widget System:', entityDBAdmin.widgetSystem);

// Manually add a test widget
if (entityDBAdmin.widgetSystem) {
    entityDBAdmin.widgetSystem.addWidget({
        id: 'manual-test',
        type: 'test',
        size: 'medium',
        config: { title: 'Manual Test' }
    });
} else {
    console.error('Widget system not initialized');
}
```

### 5. Entity API Test

Check if dashboard layouts are saving:

```javascript
// In browser console
fetch('/api/v1/entities/list?tag=type:dashboard_layout', {
    headers: {
        'Authorization': `Bearer ${localStorage.getItem('entitydb-admin-token')}`
    }
})
.then(r => r.json())
.then(data => console.log('Dashboard layouts:', data));
```

### 6. What's Working Now

- ✅ Widget system JavaScript included
- ✅ Test widget registered
- ✅ Debug logging added throughout
- ✅ Entity API working (after server restart)
- ✅ Dashboard save/load functionality implemented
- ✅ Available widgets populated including test widget

### 7. If Nothing Works

1. Open https://localhost:8085/widget-test.html
2. Click buttons in order:
   - Initialize Widget System
   - Register Test Widget  
   - Add Widget
3. If this works, the issue is in the main dashboard integration
4. If this fails, check browser console for JavaScript errors

### 8. Server Issues

If entity API hangs:
```bash
cd /opt/entitydb
./bin/entitydbd.sh restart
```

Then wait 30 seconds for server to fully start before testing again.