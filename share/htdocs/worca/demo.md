# Worca Dashboard Widget System Demo

## Overview

The Worca Dashboard now features a fully programmable widget system that integrates EntityDB metrics directly into the main application. Users can create multiple dashboards, add/remove widgets, and customize their workspace.

## Key Features

### 1. **Multiple Dashboards**
- Create unlimited custom dashboards
- Each dashboard has its own widget layout
- Dashboards are saved in EntityDB as entities
- Support for public/private dashboards

### 2. **Widget Library**

#### System Metrics Widgets
- **System Information**: Version, uptime, goroutines
- **Memory Usage**: Heap allocation, system memory with charts
- **Database Statistics**: Entity counts, tags, averages
- **Storage Usage**: Database, WAL, and index sizes
- **Temporal Statistics**: Temporal tag ratios and counts
- **Performance Metrics**: GC runs, pause times

#### Task Management Widgets
- **Task Overview**: Total, active, and completed task counts
- **Mini Kanban**: Compact kanban board view
- **Recent Activity**: Activity feed

#### Team Widgets
- **Team Members**: Member list with roles
- **Team Workload**: Task distribution charts

#### Analytics Widgets
- **Task Trends**: Historical task data
- **Velocity Chart**: Sprint velocity tracking

#### Custom Widgets
- **Custom HTML**: Add your own HTML content
- **Custom Chart**: Create custom data visualizations

### 3. **Drag & Drop Layout**
- Powered by Gridster.js
- Resize widgets by dragging corners
- Rearrange widgets by dragging headers
- Grid-based layout system

### 4. **Real-time Metrics**
- Metrics auto-refresh every 30 seconds
- Manual refresh button available
- Direct integration with EntityDB's `/api/v1/system/metrics` endpoint

### 5. **Widget Configuration**
- Some widgets support custom configuration
- Settings saved per widget instance
- Configuration stored in EntityDB

## How to Use

### Creating a Dashboard
1. Click the **"+ New Dashboard"** tab
2. Enter dashboard name and description
3. Choose an icon
4. Optionally make it public
5. Click "Save Dashboard"

### Adding Widgets
1. Click **"Edit Dashboard"** to enter edit mode
2. Click **"Add Widget"**
3. Browse widget categories or view all
4. Click on any widget to add it to your dashboard
5. The widget appears at the first available position

### Arranging Widgets
1. In edit mode, drag widget headers to move them
2. Drag widget edges/corners to resize
3. Click **"Save Layout"** when done

### Removing Widgets
1. In edit mode, click the **X** button on any widget
2. The widget is immediately removed
3. Save layout to persist changes

## Technical Implementation

### Widget Definition Structure
```javascript
const WIDGET_REGISTRY = {
    systemInfo: {
        name: 'System Information',
        icon: 'fa-server',
        category: 'metrics',
        defaultSize: { w: 4, h: 3 },
        component: 'SystemInfoWidget',
        configurable: false
    }
}
```

### Dashboard Storage
Dashboards are stored as EntityDB entities with:
- Type: `dashboard`
- Owner tag: `owner:username`
- Visibility: `visibility:public` or `visibility:private`
- Content: JSON with dashboard configuration and widget layout

### Metrics Integration
The widget system fetches metrics from:
- `/health` - Basic health check
- `/metrics` - Prometheus-format metrics
- `/api/v1/system/metrics` - Comprehensive EntityDB metrics

### Widget Data Flow
1. Metrics refresh timer calls `refreshMetrics()`
2. Fetches data from EntityDB API
3. Stores in `metricsData` reactive property
4. Widgets automatically update via Alpine.js reactivity

## Benefits

1. **Customizable Workspaces**: Each user can create their ideal dashboard layout
2. **Real-time Monitoring**: System metrics update automatically
3. **Integrated Experience**: No need to switch between apps for metrics
4. **Persistent Layouts**: All configurations saved in EntityDB
5. **Extensible**: Easy to add new widget types
6. **Responsive**: Works on desktop and tablet devices

## Future Enhancements

1. **Widget Marketplace**: Share custom widgets with the community
2. **Advanced Charts**: More visualization options
3. **Widget Templates**: Pre-built dashboard templates
4. **Export/Import**: Share dashboard configurations
5. **Mobile Optimization**: Better support for small screens
6. **Widget Permissions**: Control who can see specific widgets
7. **Data Sources**: Connect to external APIs for widgets

## Example Use Cases

### DevOps Dashboard
- System metrics widgets for monitoring
- Storage usage for capacity planning
- Performance metrics for optimization
- Recent activity for audit trail

### Project Management Dashboard
- Task overview for quick status
- Mini kanban for workflow visualization
- Team workload for resource planning
- Sprint progress for velocity tracking

### Executive Dashboard
- High-level KPI widgets
- Project status summaries
- Team performance metrics
- Custom charts for business metrics

## Summary

The Worca widget system transforms the dashboard from a static view into a dynamic, programmable workspace. By integrating EntityDB metrics directly into the application, users get a unified experience for both task management and system monitoring. The drag-and-drop interface makes customization intuitive, while the EntityDB backend ensures all configurations are safely persisted.

This implementation demonstrates the power of EntityDB as a flexible backend for modern applications, where everything from user data to UI configurations can be stored as entities with temporal tracking.