# EntityDB Platform Documentation Update

## Overview

The EntityDB (EntityDB) Platform has been significantly improved with the following enhancements:

1. Simple API Server implementation
2. Web Dashboard with real-time data visualization
3. Improved API endpoints
4. Documentation updates

## Simple API Server

A new simplified API server has been implemented to address issues with the original server implementation. The simplified server:

- Provides all necessary API endpoints for the dashboard
- Uses simulated data that matches the expected data models
- Handles CORS properly for frontend integration
- Follows RESTful API conventions
- Serves the web dashboard files

### Key Endpoints

The simple API server provides these primary endpoints:

- `/api/v1/status` - Returns system status information
- `/api/v1/agents/list` - Returns a list of all agents
- `/api/v1/tasks/list` - Returns a list of all tasks
- `/api/v1/projects/list` - Returns a list of all projects
- `/api/v1/sessions/list` - Returns a list of all sessions
- `/api/v1/dashboard/stats` - Returns key statistics for the dashboard
- `/api/v1/activities` - Returns recent system activities
- `/api/v1/system/resources` - Returns system resource usage

### Starting the Simple API Server

The simple API server can be started using the provided script:

```bash
cd /opt/entitydb && ./bin/run_simple_api.sh start
```

Other commands include:

```bash
./bin/run_simple_api.sh stop      # Stop the server
./bin/run_simple_api.sh restart   # Restart the server
./bin/run_simple_api.sh status    # Check server status
```

## Web Dashboard

The web dashboard has been completely redesigned to provide a clean, modern interface for the EntityDB platform. Key features include:

### Layout Options

The dashboard supports two layout modes:

1. **Grid Layout** - Widgets are organized in a responsive grid
2. **Floating Layout** - Widgets can be positioned freely

Toggle between these layouts using the "Switch to Floating Mode" / "Switch to Grid Mode" button.

### Theme Options

The dashboard supports light and dark themes. Toggle between them using the "Toggle Dark Mode" / "Toggle Light Mode" button.

### Widgets

The dashboard includes these key widgets:

1. **System Statistics** - Shows active agents, tasks, projects, and sessions
2. **Recent Activity** - Displays recent system activities with timestamps
3. **System Resources** - Shows CPU, memory, disk, and network usage
4. **System Status** - Shows system status and API information

### Data Refresh

Dashboard data refreshes automatically every 30 seconds. You can also manually refresh the data using the "Refresh Data" button.

## Next Steps

The following enhancements are planned for future updates:

1. **Role-Based Access Control (RBAC)**
   - Design role hierarchy (Admin, Manager, Agent)
   - Implement authentication with role validation
   - Add permission checks to API endpoints

2. **Issue Management**
   - Update models from Task to Issue
   - Implement issue hierarchy (Epic-Story-Task)
   - Implement issue state transitions and history

3. **Team Management**
   - Implement Teams (formerly pools)
   - Update agent model with avatar and job designation support

4. **Client Integration**
   - Update client commands for issues
   - Implement agent login/logout flow
   - Connect client to API endpoints

5. **Enhanced Dashboard**
   - Implement team visualization
   - Add issue hierarchy visualization
   - Add metrics visualizations and charts

6. **Database Integration**
   - Update database schema for new models
   - Add metrics tables for all entities
   - Switch from mock data to database persistence

## Running the System

### Starting the Server

```bash
# Use the simplified API server
cd /opt/entitydb && ./bin/run_simple_api.sh start

# Or use the original server (not recommended currently)
cd /opt/entitydb && ./bin/entitydbd.sh start
```

### Accessing the Dashboard

Open your browser and navigate to:

```
http://localhost:8085/
```

### Using the API

You can interact with the API directly:

```bash
# Check server status
curl http://localhost:8085/api/v1/status

# Get dashboard statistics
curl http://localhost:8085/api/v1/dashboard/stats

# Get agents list
curl http://localhost:8085/api/v1/agents/list

# Get recent activities
curl http://localhost:8085/api/v1/activities
```