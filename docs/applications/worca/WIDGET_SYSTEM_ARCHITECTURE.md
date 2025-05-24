# Worca Widget System Architecture

> **Current Implementation State**: Post-Layout Enhancement (Pre-tag)
> **Last Updated**: 2025-05-24

## Overview

Worca features a modular widget system that provides a customizable dashboard experience with full-screen layout optimization. The system enables users to create multiple dashboards, add widgets from a comprehensive registry, and manage layouts through an intuitive interface.

## Architecture Components

### 1. Widget Registry (`worca-widgets.js`)

The widget registry defines all available widget types with their metadata:

```javascript
const WIDGET_REGISTRY = {
    systemInfo: {
        name: 'System Information',
        icon: 'fa-server',
        category: 'metrics',
        defaultSize: { w: 4, h: 3 },
        component: 'SystemInfoWidget'
    },
    // ... additional widgets
};
```

**Widget Categories:**
- **Metrics**: System info, memory usage, database stats, storage usage
- **Tasks**: Task overview, mini kanban, recent activity
- **Team**: Team workload, team members
- **Projects**: Project status, sprint progress
- **Analytics**: Task trends, velocity charts
- **Custom**: Custom HTML, custom charts

### 2. Layout System

**Full-Screen Responsive Layout:**
- Sidebar: Fixed 280px width (collapsible)
- Main Content: Remaining screen width with proper scrolling
- Content Area: Consistent 20px margins (top/bottom/sides)
- Height Calculation: `calc(100vh - 60px - 40px)` for perfect spacing

**CSS Grid Implementation:**
```css
.widget-grid {
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    grid-auto-rows: 100px;
    gap: 16px;
    width: 100%;
    min-height: 100%;
}
```

### 3. State Management

**Dashboard Structure:**
```javascript
{
    id: 'unique-id',
    name: 'Dashboard Name',
    icon: 'fa-icon-name',
    widgets: [
        {
            id: 'widget-id',
            type: 'systemInfo',
            x: 0, y: 0,    // Grid position
            w: 3, h: 2     // Grid size (columns x rows)
        }
    ]
}
```

**Persistence:**
- Primary: EntityDB storage (planned)
- Fallback: localStorage for development
- Auto-save on layout changes

### 4. Integration with Main Application

The widget system is integrated into the main `worca()` Alpine.js function:

```javascript
// Widget System Properties
dashboards: [],
currentDashboard: null,
widgets: [],
editMode: false,
showAddWidget: false,
metricsData: null,

// Widget Registry Access
get WIDGET_REGISTRY() {
    return window.WIDGET_REGISTRY || {};
},

get WIDGET_CATEGORIES() {
    return window.WIDGET_CATEGORIES || {};
}
```

## Key Features

### 1. Dashboard Management
- **Multiple Dashboards**: Users can create unlimited dashboards
- **Dashboard Switching**: Tabbed interface for quick switching
- **Persistent Storage**: Dashboards saved automatically
- **Default Dashboard**: Auto-created with essential widgets

### 2. Edit Mode
- **Toggle Edit Mode**: Switch between view/edit modes
- **Visual Feedback**: Edit mode styling and controls
- **Auto-Save**: Changes saved when exiting edit mode
- **Widget Actions**: Add/remove widgets in edit mode

### 3. Widget Gallery
- **Categorized Widgets**: Organized by function (metrics, tasks, etc.)
- **One-Click Addition**: Simple widget addition process
- **Default Sizing**: Intelligent default sizes per widget type
- **Auto-Positioning**: Smart positioning for new widgets

### 4. Data Integration
- **Real-time Metrics**: Live system metrics from `/api/v1/system/metrics`
- **Auto-refresh**: 30-second refresh interval
- **Error Handling**: Graceful degradation for missing data
- **Dynamic Content**: Widget content updates automatically

## Layout Improvements

### Full-Screen Optimization
1. **Container Hierarchy:**
   ```
   app-container (flex, full height)
   ├── sidebar (280px fixed width)
   └── main-content (remaining width)
       └── content-area (scrollable, consistent margins)
   ```

2. **Responsive Behavior:**
   - Consistent 20px margins on all sides
   - Perfect spacing from taskbar to bottom
   - Internal scrolling for content overflow
   - Better space utilization across all views

3. **Cross-View Consistency:**
   - Dashboard, Kanban, Projects, Team, Analytics all use same layout
   - Consistent spacing and visual hierarchy
   - Proper content boundaries

## Implementation Status

### ✅ Completed
- Full-screen layout with consistent margins
- Widget registry and categorization
- Dashboard creation and management
- Edit mode with add/remove functionality
- Integration with main Worca application
- Real-time metrics integration
- localStorage persistence
- Simplified CSS Grid layout (removed complex Gridster)
- Comprehensive console logging for debugging

### 🔄 In Progress
- Widget drag-and-drop positioning
- Advanced widget configuration
- EntityDB persistence integration

### 📋 Planned
- Widget resizing in edit mode
- Custom widget creation
- Dashboard sharing and permissions
- Advanced analytics widgets
- Widget import/export functionality

## Development Guidelines

### Adding New Widgets

1. **Register Widget Type:**
   ```javascript
   // In worca-widgets.js
   myWidget: {
       name: 'My Widget',
       icon: 'fa-custom-icon',
       category: 'custom',
       defaultSize: { w: 4, h: 3 },
       component: 'MyWidgetComponent'
   }
   ```

2. **Implement Rendering:**
   ```javascript
   // In worca.js renderWidgetContent()
   case 'myWidget':
       return `<div>My custom widget content</div>`;
   ```

3. **Add Data Source:**
   ```javascript
   // In getWidgetData() method
   case 'myWidget':
       return this.myWidgetData || {};
   ```

### Layout Debugging

The system includes comprehensive logging for debugging layout and widget issues:

```javascript
// Use logWidgetState() for debugging
this.logWidgetState('OPERATION_NAME');
```

Console output includes:
- Current dashboard state
- Widget count and details
- Edit mode status
- Widget registry access
- Positioning information

## Security Considerations

- **Input Validation**: All widget configurations validated
- **XSS Prevention**: Content properly escaped in widget rendering
- **RBAC Integration**: Widget access controlled by user permissions
- **Data Isolation**: Widget data scoped to user context

## Performance Optimization

- **Lazy Loading**: Widgets render content on demand
- **Efficient Updates**: Minimal DOM manipulation
- **Memory Management**: Proper cleanup of event listeners
- **Caching**: Widget data cached for improved performance

## Future Enhancements

1. **Advanced Positioning**: Drag-and-drop with collision detection
2. **Widget Marketplace**: Community-contributed widgets
3. **Theming System**: Custom widget themes and styling
4. **Real-time Collaboration**: Multi-user dashboard editing
5. **Widget Templates**: Pre-configured dashboard templates
6. **Advanced Analytics**: Custom chart and data visualization widgets

---

*This document reflects the current state of the Worca widget system architecture and serves as a guide for future development and maintenance.*