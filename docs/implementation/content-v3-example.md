# EntityDB Content V3 Example

## Current Content Model (V2)
Multiple content items per entity:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tags": ["type:document", "status:draft"],
  "content": [
    {"type": "title", "value": "Project Report"},
    {"type": "body", "value": "This is the main content..."},
    {"type": "author", "value": "Jane Smith"},
    {"type": "metadata", "value": "{\"version\": 1, \"format\": \"markdown\"}"}
  ]
}
```

## New Content Model (V3)
Single content artifact per entity with MIME type:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tags": [
    "type:document",
    "status:draft",
    "title:Project Report",
    "author:Jane Smith",
    "format:markdown",
    "version:1"
  ],
  "content": {
    "mime_type": "text/markdown",
    "data": "# Project Report\n\nThis is the main content...",
    "size": 42,
    "checksum": "sha256:abcd1234...",
    "created_at": "2025-05-19T16:30:00Z",
    "updated_at": "2025-05-19T16:30:00Z"
  }
}
```

## Complex Documents with Relationships

For documents with multiple artifacts:

### Parent Document
```json
{
  "id": "doc-001",
  "tags": ["type:document", "title:Annual Report"],
  "content": {
    "mime_type": "text/markdown",
    "data": "# Annual Report 2025\n\n## Summary..."
  }
}
```

### Attachment 1 (Image)
```json
{
  "id": "img-001",
  "tags": ["type:attachment", "parent:doc-001", "name:chart.png"],
  "content": {
    "mime_type": "image/png",
    "data": "iVBORw0KGgoAAAANSUhEUgA..." // Base64 encoded
  }
}
```

### Attachment 2 (Spreadsheet)
```json
{
  "id": "file-001",
  "tags": ["type:attachment", "parent:doc-001", "name:financial_data.csv"],
  "content": {
    "mime_type": "text/csv",
    "data": "Date,Revenue,Expenses\n2025-01-01,50000,30000..."
  }
}
```

### Relationships
```json
{
  "source": "doc-001",
  "target": "img-001",
  "type": "has_attachment"
}
```
```json
{
  "source": "doc-001",
  "target": "file-001",
  "type": "has_attachment"
}
```

## Benefits

1. **Clarity**: Each entity has a clear, single purpose
2. **Standard Types**: MIME types are universally understood
3. **Flexibility**: Relationships handle complex structures
4. **Performance**: Less data per entity means faster queries
5. **Compatibility**: Works with standard HTTP content negotiation