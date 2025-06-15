# EntityDB Admin Interface Guide

## Overview

The EntityDB Admin Interface provides a comprehensive web-based control panel for managing and monitoring the EntityDB platform. It offers a unified interface for administrators to handle entities, relationships, users, and sessions within the entity-based architecture.

## Accessing the Admin Interface

There are several ways to access the Admin Interface:

1. From the main dashboard at `http://localhost:8085/`, click on the "Admin" link in the navigation bar
2. Direct URL access: `http://localhost:8085/admin/`
3. Using the test script: `/opt/entitydb/bin/test_admin_ui.sh`

> **Note**: The admin interface is now accessible directly via the `/admin/` path.

## Features

### Dashboard

The dashboard provides a high-level overview of the system with:

- Entity count cards
- Relationship statistics
- User management
- Session monitoring
- System status information

### Entity Management

The Entity Management section allows you to:

- View all entities in the system
- Filter entities by type, status, and tags
- Create new entities with complete property and tag support
- View detailed entity information
- Edit entity properties and metadata
- Delete entities when they're no longer needed

Entities are the foundation of the EntityDB system, representing various objects like:

- Issues
- Agents
- Workspaces
- Sessions
- Users
- And any other custom entity types

### Relationship Management

The Relationship Management section enables you to:

- View existing relationships between entities
- Create new relationships between any two entities
- Edit relationship types and properties
- Delete relationships

Relationships define how entities connect to each other with types such as:

- Parent/Child
- Depends On/Required By
- Assigned To
- Member Of
- And other custom relationship types

### User Management

The User Management section provides tools to:

- View all users in the system
- Create new user accounts
- Assign roles and permissions
- Edit user information
- Change user status (active/inactive)
- Delete users

### Session Management

The Session Management section allows you to:

- Monitor active sessions
- View session details and progress
- Create new sessions
- Edit session properties
- End or delete sessions

## Technical Implementation

The Admin Interface adheres to the pure entity-based architecture of EntityDB:

1. All operations are performed through the entity API
2. The interface uses the same authentication and authorization system as the rest of EntityDB
3. Changes made in the Admin Interface are immediately reflected throughout the system

### Components

- **HTML/CSS**: Modern, responsive design with light/dark theme support
- **JavaScript**: Client-side functionality for interacting with the EntityDB API
- **API Integration**: Complete integration with the entity and relationship APIs
- **Debug Tools**: Built-in debug panel for troubleshooting

## Troubleshooting

If you encounter issues with the Admin Interface:

### Server Connectivity

If the admin interface isn't loading:

1. Check if the EntityDB server is running:
   ```
   /opt/entitydb/bin/entitydbd.sh status
   ```

2. Try the standalone test server:
   ```
   /opt/entitydb/bin/test_admin_ui.sh
   ```

3. To stop the standalone server:
   ```
   /opt/entitydb/bin/stop_admin_ui.sh
   ```

### Authentication Issues

If you're having trouble logging in:

1. Ensure you have a valid token:
   - The admin interface will redirect to login if no token is found
   - Your token might be expired

2. Check your permissions:
   - Admin features require admin role permissions
   - Some operations might be restricted based on your role

### Debug Mode

The Admin Interface includes a debug panel that can help diagnose issues:

1. The debug panel appears as a bar at the bottom of the screen
2. Click on it to expand the full debug information
3. It shows:
   - API connection status
   - Authentication status
   - Data sample information
   - Console logs

## Best Practices

1. **Entity Creation**: When creating entities, always include appropriate tags to ensure proper categorization
2. **Relationships**: Maintain clean relationship structures to avoid orphaned entities
3. **User Management**: Follow the principle of least privilege when assigning roles
4. **Regular Monitoring**: Use the Admin Interface to regularly monitor system health and activity

## Development and Extension

The Admin Interface follows a modular design that allows for extension:

1. New entity types can be added without modifying the interface
2. Additional relationship types are automatically supported
3. The debug module provides extension points for custom troubleshooting

## Security Considerations

The Admin Interface implements several security measures:

1. All API calls use JWT authentication
2. Role-based access control for administrative functions
3. Input validation on all forms
4. Confirmation prompts for destructive operations

## Further Information

For additional details on the EntityDB system's entity-based architecture, refer to:

- `/opt/entitydb/CLAUDE.md`: Overall system documentation
- `/opt/entitydb/docs/entity_architecture_guide.md`: Entity architecture documentation
- `/opt/entitydb/docs/entity_api_reference.md`: API reference documentation