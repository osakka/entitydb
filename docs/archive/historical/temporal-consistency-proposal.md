# Temporal Consistency Proposal for EntityDB

## Current State (Inconsistent)

Currently, EntityDB has mixed temporal storage:

1. **Some tags with timestamps**: `2025-05-18T19:56:58.499388918.type=user`
2. **Some tags without**: `type:user`
3. **Content with timestamps**: In ContentItem struct
4. **Relationships**: Currently no temporal tracking

## Proposed Unified Approach

### Core Principle
Everything in the database should have temporal information, but this should be transparent to users.

### Storage Format
```
Internal Storage: <timestamp>.<tag>
User API:        <tag>
```

### Implementation

1. **Tags**
   - Storage: `2025-05-18T19:56:58.499388918.type:user`
   - API Input: `type:user`
   - API Output: `type:user` (default) or with timestamp if requested

2. **Content** 
   - Already has timestamps in ContentItem
   - Keep current structure

3. **Relationships**
   - Add timestamp to EntityRelationship struct
   - Storage: Track when relationships were created/modified
   - API: Transparent to users

### Benefits

1. **Consistent Temporal Queries**
   - All data has timestamps
   - Can query any aspect at any point in time
   - Enables full temporal features

2. **Transparent to Users**
   - API accepts tags without timestamps
   - System automatically adds current timestamp
   - Temporal queries available when needed

3. **Backward Compatible**
   - Existing APIs continue to work
   - New temporal features are additive

### Implementation Steps

1. Modify storage layer to always add timestamps
2. Update API handlers to:
   - Strip timestamps from output (default behavior)
   - Add timestamps on input
   - Provide temporal query options
3. Update relationship model with timestamps
4. Create migration tool for existing data

### Example API Usage

```bash
# User sends (no timestamp)
POST /api/v1/entities/create
{
  "tags": ["type:user", "status:active"],
  "content": [{"type": "name", "value": "John"}]
}

# System stores
{
  "tags": [
    "2025-05-18T21:00:00.123456789.type:user",
    "2025-05-18T21:00:00.123456789.status:active"
  ],
  "content": [
    {
      "timestamp": "2025-05-18T21:00:00.123456789",
      "type": "name", 
      "value": "John"
    }
  ]
}

# User receives (default - no timestamp)
{
  "tags": ["type:user", "status:active"],
  "content": [{"type": "name", "value": "John"}]
}

# User can request with timestamps
GET /api/v1/entities/get?id=123&include_timestamps=true
```

### Temporal Query Examples

```bash
# Get entity as it was at specific time
GET /api/v1/entities/as-of?id=123&timestamp=2025-05-18T20:00:00

# Get history of specific tag
GET /api/v1/entities/tag-history?id=123&tag=status

# Get relationship history
GET /api/v1/relationships/history?entity=123
```

This approach provides full temporal capabilities while keeping the API simple for users who don't need temporal features.