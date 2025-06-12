# EntityDB Content V3 Migration

## Overview

EntityDB v3 introduces a new content model:
- Single content artifact per entity
- MIME types instead of custom types
- Relationships for multiple artifacts

## Changes

### Before (V2)
```json
{
  "id": "uuid",
  "tags": ["type:document"],
  "content": [
    {"type": "title", "value": "My Document"},
    {"type": "body", "value": "Content here"},
    {"type": "author", "value": "John Doe"}
  ]
}
```

### After (V3)
```json
{
  "id": "uuid",
  "tags": ["type:document", "title:My Document", "author:John Doe"],
  "content": {
    "mime_type": "text/plain",
    "data": "Content here",
    "size": 12,
    "checksum": "sha256...",
    "created_at": "2025-05-19T16:00:00Z",
    "updated_at": "2025-05-19T16:00:00Z"
  }
}
```

## Migration Strategy

### 1. Simple Content
- Single text content → Keep as is with `text/plain` MIME type
- Multiple text fields → Move metadata to tags, keep main content

### 2. Complex Content
For entities with multiple content items:
1. Create parent entity with primary content
2. Create child entities for additional content
3. Link with relationships

Example:
```
Document Entity
├── Content: Main document text (text/plain)
├── Tags: ["type:document", "title:My Doc"]
└── Relationships:
    ├── Attachment 1 (PDF)
    ├── Attachment 2 (Image)
    └── Metadata (JSON)
```

### 3. Special Cases

#### User Entities
- Username → Tag: `id:username:john`
- Password hash → Tag: `auth:password_hash:...`
- Email → Tag: `contact:email:john@example.com`
- Profile → Content: JSON with profile data

#### Configuration
- Config values → Content: JSON with all settings
- Individual settings → Tags for quick access

## Implementation Steps

1. **Create V3 models** ✓
2. **Add migration functions**
3. **Update API handlers**
4. **Create migration tool**
5. **Update documentation**

## Benefits

1. **Cleaner API**: Single content per entity is simpler
2. **Better Performance**: Less data to parse per entity
3. **Standard Types**: MIME types are universal
4. **Flexibility**: Relationships handle complex structures

## Backwards Compatibility

During transition:
- Support both V2 and V3 formats
- Auto-migrate on read
- Deprecation warnings for V2 API
- Migration tool for bulk conversion