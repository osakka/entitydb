# EntityDB Ticketing System Example

## Overview

This example demonstrates how to build a complete ticketing system using EntityDB's temporal database. The system uses entities and tags to represent tickets, projects, comments, and labels.

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│    Projects     │     │     Tickets     │     │    Comments     │
├─────────────────┤     ├─────────────────┤     ├─────────────────┤
│ type:project    │     │ type:ticket     │     │ type:comment    │
│ id:code:PROD    │────>│ project:PROD    │<────│ ticket:PROD-1   │
│ name:Production │     │ id:ticket:PROD-1│     │ author:admin    │
│ status:active   │     │ status:open     │     │                 │
└─────────────────┘     │ priority:high   │     └─────────────────┘
                        │ assigned_to:john│
                        └─────────────────┘
                                ↓
                        ┌─────────────────┐
                        │     Labels      │
                        ├─────────────────┤
                        │ type:label      │
                        │ category:priority│
                        │ name:High       │
                        │ color:red       │
                        └─────────────────┘
```

## Entity Types

### 1. Project Entities
Projects organize tickets into logical groups.

```json
{
  "tags": [
    "type:project",
    "id:code:HELPDESK",
    "name:Help Desk",
    "status:active"
  ],
  "content": [
    {"type": "description", "value": "IT Help Desk ticketing"}
  ]
}
```

### 2. Ticket Entities
Tickets represent issues or tasks to be resolved.

```json
{
  "tags": [
    "type:ticket",
    "id:ticket:HD-001",
    "project:HELPDESK",
    "status:open",
    "priority:critical",
    "category:network",
    "assigned_to:tech_john",
    "created_by:user123"
  ],
  "content": [
    {"type": "title", "value": "Network outage in Building A"},
    {"type": "description", "value": "Complete network connectivity loss"}
  ]
}
```

### 3. Comment Entities
Comments provide updates and discussion on tickets.

```json
{
  "tags": [
    "type:comment",
    "ticket:HD-001",
    "author:tech_john"
  ],
  "content": [
    {"type": "text", "value": "Investigating issue. Main switch has failed."}
  ]
}
```

### 4. Label Entities
Labels categorize tickets for easy filtering.

```json
{
  "tags": [
    "type:label",
    "category:priority",
    "name:Critical",
    "color:red"
  ],
  "content": []
}
```

## Tag Structure

### Common Tags
- `type:` - Entity type (ticket, project, comment, label)
- `id:` - Unique identifier
- `status:` - Current state (open, closed, in_progress)
- `priority:` - Urgency level (critical, high, medium, low)

### Ticket-Specific Tags
- `project:PROJECT_CODE` - Links ticket to project
- `assigned_to:username` - Current assignee
- `created_by:username` - Ticket creator
- `category:` - Type of issue (network, hardware, software)
- `label:` - Additional categorization

### Temporal Benefits
- All changes are automatically timestamped with nanosecond precision
- Complete audit trail of modifications
- Query ticket state at any point in time
- Track progression and history

## Usage Examples

### Create a Ticket
```bash
curl -X POST $BASE_URL/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:ticket",
      "id:ticket:HD-001",
      "project:HELPDESK",
      "status:open",
      "priority:high"
    ],
    "content": [
      {"type": "title", "value": "Printer not working"},
      {"type": "description", "value": "Main office printer showing error"}
    ]
  }'
```

### Query Open Tickets
```bash
curl -X GET "$BASE_URL/entities/list?tag=type:ticket" \
  -H "Authorization: Bearer $TOKEN" | \
jq '.[] | select(.tags[] | contains("status:open"))'
```

### Add Comment to Ticket
```bash
curl -X POST $BASE_URL/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:comment",
      "ticket:HD-001",
      "author:support_agent"
    ],
    "content": [
      {"type": "text", "value": "Printer has been replaced"}
    ]
  }'
```

### Update Ticket Status
```bash
# First get the ticket
TICKET=$(curl -X GET "$BASE_URL/entities/get?id=$TICKET_ID" \
  -H "Authorization: Bearer $TOKEN")

# Update status tag
# (In a real implementation, you'd modify the tags array)
```

## Temporal Features

### View Ticket History
```bash
curl -X GET "$BASE_URL/entities/history?id=$TICKET_ID" \
  -H "Authorization: Bearer $TOKEN"
```

### Get Ticket State at Specific Time
```bash
curl -X GET "$BASE_URL/entities/as-of?id=$TICKET_ID&as_of=2024-05-18T10:00:00Z" \
  -H "Authorization: Bearer $TOKEN"
```

## Relationships

You can create explicit relationships between entities:

```bash
# Link a ticket to a parent ticket
curl -X POST $BASE_URL/entity-relationships \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "source_id": "CHILD_TICKET_ID",
    "relationship_type": "child_of",
    "target_id": "PARENT_TICKET_ID"
  }'
```

## Advanced Queries

### Find High Priority Tickets
```bash
curl -X GET "$BASE_URL/entities/query?filter=tag:priority&operator=eq&value=high" \
  -H "Authorization: Bearer $TOKEN"
```

### Get Tickets Updated in Last Hour
```bash
curl -X GET "$BASE_URL/entities/changes?since=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)" \
  -H "Authorization: Bearer $TOKEN"
```

## Benefits of Using EntityDB

1. **Temporal Tracking**: Every change is automatically timestamped
2. **Flexibility**: Add new fields/tags without schema changes
3. **Audit Trail**: Complete history of all modifications
4. **Simple API**: RESTful interface for all operations
5. **Relationships**: Link tickets, create hierarchies
6. **Performance**: Binary storage format for efficiency

## Real-World Use Cases

- IT Help Desk Systems
- Bug Tracking
- Project Management
- Customer Support
- Incident Management
- Change Request Tracking

## Summary

This ticketing system demonstrates EntityDB's flexibility and temporal capabilities. The tag-based system allows for easy extension without schema changes, while temporal features provide complete audit trails and time-travel queries.