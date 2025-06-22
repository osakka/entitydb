# Entity API Usage Guide

This guide provides practical examples of using the Entity API for common tasks in the EntityDB system.

## Introduction

The EntityDB system is built on a pure entity-based architecture that consolidates all operations under a unified entity API. This guide will help you understand how to effectively use this API for various common tasks.

## Key Concepts

### Entity-Based Architecture

In EntityDB, everything is an entity. Instead of specialized API endpoints for different object types (issues, workspaces, agents), all objects are represented as entities with different tag sets that define their type and characteristics.

Advantages of this approach:
- **Uniformity**: Single API pattern for all operations
- **Flexibility**: New object types can be added without API changes
- **Extensibility**: Objects can have custom properties without schema changes
- **Consistency**: Objects share the same query and filtering mechanisms

### Entities and Relationships

The two key concepts in the entity architecture are:

1. **Entities**: Objects with tags, content, and properties
2. **Relationships**: Connections between entities with types and properties

## Basic Entity Operations

### Creating Different Entity Types

#### Creating a Workspace (Project)

**Using the HTTP API:**

```bash
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:workspace", "status:active"],
    "content": [
      {
        "key": "name",
        "value": "Project Alpha"
      },
      {
        "key": "description",
        "value": "Strategic initiative for Q2"
      }
    ]
  }'
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity create \
  --type=workspace \
  --title="Project Alpha" \
  --description="Strategic initiative for Q2" \
  --tags="status:active"
```

#### Creating an Issue

**Using the HTTP API:**

```bash
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:issue", "status:pending", "priority:high", "workspace:alpha"],
    "content": [
      {
        "key": "title",
        "value": "Implement User Authentication"
      },
      {
        "key": "description",
        "value": "Add JWT-based authentication flow for users"
      }
    ]
  }'
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity create \
  --type=issue \
  --title="Implement User Authentication" \
  --description="Add JWT-based authentication flow for users" \
  --tags="status:pending,priority:high,workspace:alpha"
```

#### Creating an Epic

**Using the HTTP API:**

```bash
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:epic", "status:pending", "workspace:alpha"],
    "content": [
      {
        "key": "title",
        "value": "User Management System"
      },
      {
        "key": "description",
        "value": "Complete user management system including authentication, profiles, and permissions"
      }
    ]
  }'
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity create \
  --type=epic \
  --title="User Management System" \
  --description="Complete user management system including authentication, profiles, and permissions" \
  --tags="status:pending,workspace:alpha"
```

### Querying Entities

#### Listing All Issues in a Workspace

**Using the HTTP API:**

```bash
curl -X GET "http://localhost:8085/api/v1/entities?tag=type:issue&tag=workspace:alpha" \
  -H "Authorization: Bearer <your_token>"
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity list --type=issue --tags="workspace:alpha"
```

#### Finding High Priority Issues

**Using the HTTP API:**

```bash
curl -X GET "http://localhost:8085/api/v1/entities?tag=type:issue&tag=priority:high" \
  -H "Authorization: Bearer <your_token>"
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity list --type=issue --tags="priority:high"
```

#### Finding Blocked Issues

**Using the HTTP API:**

```bash
curl -X GET "http://localhost:8085/api/v1/entities?tag=type:issue&tag=status:blocked" \
  -H "Authorization: Bearer <your_token>"
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity list --type=issue --tags="status:blocked"
```

#### Retrieving a Specific Entity

**Using the HTTP API:**

```bash
curl -X GET "http://localhost:8085/api/v1/entities/get?id=entity_12345" \
  -H "Authorization: Bearer <your_token>"
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity get --id=entity_12345
```

### Updating Entities

#### Changing Issue Status

**Using the HTTP API:**

```bash
# First, get the current entity
curl -X GET "http://localhost:8085/api/v1/entities/get?id=entity_12345" \
  -H "Authorization: Bearer <your_token>"

# Then update it with modified tags
curl -X PUT http://localhost:8085/api/v1/entities \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "entity_12345",
    "tags": ["type:issue", "status:in_progress", "priority:high", "workspace:alpha"],
    "content": [
      {
        "key": "title",
        "value": "Implement User Authentication"
      },
      {
        "key": "description",
        "value": "Add JWT-based authentication flow for users"
      }
    ]
  }'
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity update \
  --id=entity_12345 \
  --tags="type:issue,status:in_progress,priority:high,workspace:alpha"
```

#### Adding Content to an Entity

**Using the HTTP API:**

```bash
# First, get the current entity
curl -X GET "http://localhost:8085/api/v1/entities/get?id=entity_12345" \
  -H "Authorization: Bearer <your_token>"

# Then update it with additional content
curl -X PUT http://localhost:8085/api/v1/entities \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "entity_12345",
    "tags": ["type:issue", "status:in_progress", "priority:high", "workspace:alpha"],
    "content": [
      {
        "key": "title",
        "value": "Implement User Authentication"
      },
      {
        "key": "description",
        "value": "Add JWT-based authentication flow for users"
      },
      {
        "key": "assigned_to",
        "value": "agent_claude"
      },
      {
        "key": "progress",
        "value": "50"
      }
    ]
  }'
```

## Working with Relationships

### Creating Relationships

#### Adding a Dependency Relationship

**Using the HTTP API:**

```bash
curl -X POST http://localhost:8085/api/v1/entity-relationships \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "source_id": "entity_12345",
    "target_id": "entity_67890",
    "relationship_type": "depends_on",
    "metadata": "{\"priority\": \"high\", \"reason\": \"Needs database schema update first\"}"
  }'
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity relationship create \
  --source=entity_12345 \
  --target=entity_67890 \
  --type=depends_on \
  --metadata='{"priority": "high", "reason": "Needs database schema update first"}'
```

#### Creating a Parent-Child Relationship

**Using the HTTP API:**

```bash
curl -X POST http://localhost:8085/api/v1/entity-relationships \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "source_id": "entity_epic123",
    "target_id": "entity_issue456",
    "relationship_type": "parent_of"
  }'
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity relationship create \
  --source=entity_epic123 \
  --target=entity_issue456 \
  --type=parent_of
```

#### Assigning an Issue to an Agent

**Using the HTTP API:**

```bash
curl -X POST http://localhost:8085/api/v1/entity-relationships \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "source_id": "entity_issue123",
    "target_id": "entity_agent456",
    "relationship_type": "assigned_to"
  }'
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity relationship create \
  --source=entity_issue123 \
  --target=entity_agent456 \
  --type=assigned_to
```

### Querying Relationships

#### Finding All Dependencies of an Issue

**Using the HTTP API:**

```bash
curl -X GET "http://localhost:8085/api/v1/entity-relationships/source?id=entity_12345&type=depends_on" \
  -H "Authorization: Bearer <your_token>"
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity relationship list --source=entity_12345 --type=depends_on
```

#### Finding All Issues Assigned to an Agent

**Using the HTTP API:**

```bash
curl -X GET "http://localhost:8085/api/v1/entity-relationships/target?id=entity_agent456&type=assigned_to" \
  -H "Authorization: Bearer <your_token>"
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity relationship list --target=entity_agent456 --type=assigned_to
```

#### Finding All Child Issues of an Epic

**Using the HTTP API:**

```bash
curl -X GET "http://localhost:8085/api/v1/entity-relationships/source?id=entity_epic123&type=parent_of" \
  -H "Authorization: Bearer <your_token>"
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity relationship list --source=entity_epic123 --type=parent_of
```

### Deleting Relationships

**Using the HTTP API:**

```bash
curl -X DELETE "http://localhost:8085/api/v1/entity-relationships?source_id=entity_12345&target_id=entity_67890&type=depends_on" \
  -H "Authorization: Bearer <your_token>"
```

**Using the command-line tool:**

```bash
./bin/entitydbc.sh entity relationship delete \
  --source=entity_12345 \
  --target=entity_67890 \
  --type=depends_on
```

## Common Patterns and Workflows

### Issue Workflow Management

```bash
# Create a new issue
./bin/entitydbc.sh entity create \
  --type=issue \
  --title="Implement login form" \
  --tags="status:pending,priority:high,workspace:entitydb"

# Assign the issue
ISSUE_ID=$(./bin/entitydbc.sh entity list --type=issue --tags="title:Implement login form" --format=id)
./bin/entitydbc.sh entity relationship create --source=$ISSUE_ID --target=agent_claude --type=assigned_to

# Start working on the issue
./bin/entitydbc.sh entity update --id=$ISSUE_ID --tags="type:issue,status:in_progress,priority:high,workspace:entitydb"

# Complete the issue
./bin/entitydbc.sh entity update --id=$ISSUE_ID --tags="type:issue,status:completed,priority:high,workspace:entitydb"
```

### Managing Hierarchies

```bash
# Create a workspace
./bin/entitydbc.sh entity create --type=workspace --title="Project Alpha"
WORKSPACE_ID=$(./bin/entitydbc.sh entity list --type=workspace --tags="title:Project Alpha" --format=id)

# Create an epic in the workspace
./bin/entitydbc.sh entity create --type=epic --title="User Authentication" --tags="workspace:$WORKSPACE_ID"
EPIC_ID=$(./bin/entitydbc.sh entity list --type=epic --tags="title:User Authentication" --format=id)
./bin/entitydbc.sh entity relationship create --source=$WORKSPACE_ID --target=$EPIC_ID --type=parent

# Create a story in the epic
./bin/entitydbc.sh entity create --type=story --title="Login Flow" --tags="workspace:$WORKSPACE_ID,epic:$EPIC_ID"
STORY_ID=$(./bin/entitydbc.sh entity list --type=story --tags="title:Login Flow" --format=id)
./bin/entitydbc.sh entity relationship create --source=$EPIC_ID --target=$STORY_ID --type=parent

# Create an issue in the story
./bin/entitydbc.sh entity create --type=issue --title="Implement login form" --tags="workspace:$WORKSPACE_ID,epic:$EPIC_ID,story:$STORY_ID"
ISSUE_ID=$(./bin/entitydbc.sh entity list --type=issue --tags="title:Implement login form" --format=id)
./bin/entitydbc.sh entity relationship create --source=$STORY_ID --target=$ISSUE_ID --type=parent
```

## Advanced Usage Patterns

### Using Custom Tags for Categorization

You can add any custom tags to entities for advanced categorization and filtering:

```bash
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:issue", 
      "status:pending", 
      "priority:high", 
      "workspace:alpha",
      "component:frontend",
      "difficulty:medium",
      "estimated_hours:4",
      "sprint:2023-Q2-1"
    ],
    "content": [
      {
        "key": "title",
        "value": "Implement Dashboard UI"
      },
      {
        "key": "description",
        "value": "Create the responsive dashboard UI based on design mockups"
      }
    ]
  }'
```

Then filter by these custom tags:

```bash
curl -X GET "http://localhost:8085/api/v1/entities?tag=type:issue&tag=component:frontend&tag=sprint:2023-Q2-1" \
  -H "Authorization: Bearer <your_token>"
```

### Creating Complex Relationship Networks

You can create complex networks of entities by combining multiple relationship types:

1. Create an epic for a feature
2. Create multiple stories under the epic
3. Create tasks for each story
4. Establish dependencies between tasks
5. Assign tasks to agents

Here's an example of how these relationships would be structured:

```
Epic: "User Authentication"
│
├── Story: "User Registration"
│   │
│   ├── Task: "Create registration form"
│   │   └── assigned_to: Agent A
│   │
│   ├── Task: "Implement form validation"
│   │   └── assigned_to: Agent B
│   │   └── depends_on: "Create registration form"
│   │
│   └── Task: "Implement backend registration endpoint"
│       └── assigned_to: Agent C
│
└── Story: "User Login"
    │
    ├── Task: "Create login form" 
    │   └── assigned_to: Agent A
    │
    └── Task: "Implement JWT authentication"
        └── assigned_to: Agent C
        └── depends_on: "Implement backend registration endpoint"
```

### Building a Query System with Entity Tags

You can implement a powerful query system using tag combinations:

```bash
# Find all high priority frontend issues assigned to Agent A that are blocked
curl -X GET "http://localhost:8085/api/v1/entities?tag=type:issue&tag=priority:high&tag=component:frontend&tag=status:blocked" \
  -H "Authorization: Bearer <your_token>"

# Then for each result, find what's blocking it
for entity_id in $results; do
  curl -X GET "http://localhost:8085/api/v1/entity-relationships/source?id=$entity_id&type=depends_on" \
    -H "Authorization: Bearer <your_token>"
done
```

## Best Practices

### Tag Naming Conventions

- Use lowercase for all tag keys and values
- Use colon (`:`) as the separator between key and value
- Use singular nouns for tag keys (e.g., `type` not `types`)
- Use standard tag keys where possible:
  - `type`: The entity type
  - `status`: The entity status
  - `priority`: The entity priority
  - `workspace`: The workspace identifier

### Content Organization

- Use consistent key names for common attributes
- Store complex data as JSON strings in the value field
- Avoid duplicating information in both tags and content
  - Tags should be used for filtering and categorization
  - Content should be used for display and details

### Relationship Management

- Use the appropriate relationship type for the semantic meaning
- Add meaningful metadata to relationships when relevant
- Keep relationship chains shallow where possible
- For complex hierarchies, use `parent_of` relationships

### Performance Optimization

1. **Filter Effectively**

Always use the most specific filters possible to reduce the result set:

```bash
# Bad: Fetches all entities, then filters client-side
./bin/entitydbc.sh entity list

# Good: Uses server-side filtering
./bin/entitydbc.sh entity list --type=issue --tags="priority:high,status:in_progress"
```

2. **Batching Operations**

Batch related operations when possible:

```bash
# Create an issue and link it to a workspace in one session
./bin/entitydbc.sh entity create --type=issue --title="Fix bug" --tags="workspace:entitydb"
ISSUE_ID=$(./bin/entitydbc.sh entity list --type=issue --tags="title:Fix bug" --format=id)
./bin/entitydbc.sh entity relationship create --source=workspace_entitydb --target=$ISSUE_ID --type=parent
```

3. **Use Entity Embedding**

For complex hierarchies, consider embedding the paths in tags:

```bash
# Embed the path for faster queries
./bin/entitydbc.sh entity create \
  --type=issue \
  --title="Fix sub-task" \
  --tags="workspace:entitydb,epic:epic_123,story:story_456,path:/entitydb/epic_123/story_456"
```

## Troubleshooting

### Common Issues and Solutions

1. **Permission Denied (403) Responses**
   - Check that your token has the required permissions
   - Verify that you have access to the entities involved

2. **Entity Not Found (404) Responses**
   - Double-check the entity ID
   - Verify that the entity hasn't been deleted

3. **Invalid Request (400) Responses**
   - Check your request format against the API documentation
   - Ensure all required fields are present

4. **Relationship Already Exists (409) Responses**
   - Check if the relationship already exists with the same type
   - You may need to delete the existing relationship first

## Entity Schema Migration

If you're migrating from the legacy issue-based system to the entity-based system, you can use the migration tool:

```bash
./bin/migrate_issues_to_entities
```

This will:
1. Convert all issues to entities with appropriate tags
2. Convert all issue dependencies to entity relationships
3. Convert workspace memberships to entity relationships
4. Preserve all existing issue metadata as entity content

## Common Entity Tags

The entity system uses standardized tags for consistency. Here are the commonly used tags:

### Type Tags

- `type:issue`: Represents an issue or task
- `type:workspace`: Represents a workspace (formerly project)
- `type:epic`: Represents an epic (collection of stories)
- `type:story`: Represents a user story
- `type:agent`: Represents an agent profile

### Status Tags

- `status:pending`: Entity is created but not started
- `status:in_progress`: Entity is actively being worked on
- `status:blocked`: Entity is blocked by dependencies
- `status:completed`: Entity has been completed
- `status:archived`: Entity is archived

### Priority Tags

- `priority:low`: Low priority
- `priority:medium`: Medium priority
- `priority:high`: High priority
- `priority:critical`: Critical priority

## Common Relationship Types

The following relationship types are recognized by the system:

- `parent_of`: Hierarchical parent-child relationship
- `depends_on`: Dependency relationship (source depends on target)
- `blocks`: Dependency relationship (source blocks target)
- `assigned_to`: Assignment relationship (source is assigned to target)
- `related_to`: General relationship between entities
- `created_by`: Creator relationship

## Conclusion

The entity-based architecture provides a flexible, powerful, and consistent way to manage all objects in the EntityDB system. By understanding and following these patterns and best practices, you can effectively use the Entity API to build and manage complex workflows.

For detailed API reference documentation, see the [Entity API Reference](entity_api_reference.md) document.