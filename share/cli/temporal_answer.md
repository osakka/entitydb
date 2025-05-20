# Temporal System in EntityDB - How It Works

## Short Answer

**NO**, you don't need to add temporal handling to your API queries/actions!

The temporal system is **completely transparent** to API users:

1. **Storage**: All tags automatically get timestamps internally
2. **API Usage**: You work with tags normally (without timestamps)
3. **Optional**: Use `include_timestamps=true` only if you need to see timestamps

## How It Works

### Creating Entities (No special handling needed)
```bash
# You send tags WITHOUT timestamps
curl -X POST /api/v1/entities/create \
  -d '{
    "tags": ["type:ticket", "status:open", "priority:high"]
  }'

# Storage automatically adds timestamps:
# 2025-05-18T22:38:48.807532034+01:00|type:ticket
# 2025-05-18T22:38:48.807532035+01:00|status:open
# 2025-05-18T22:38:48.807532036+01:00|priority:high
```

### Querying Entities (Tags returned without timestamps)
```bash
# Query normally
curl -X GET /api/v1/entities/list?tag=type:ticket

# Response shows tags WITHOUT timestamps:
{
  "tags": ["type:ticket", "status:open", "priority:high"]
}
```

### Seeing Timestamps (Optional)
```bash
# Only if you need temporal data
curl -X GET /api/v1/entities/get?id=XXX&include_timestamps=true

# Response with timestamps:
{
  "tags": [
    "2025-05-18T22:38:48.807532034+01:00|type:ticket",
    "2025-05-18T22:38:48.807532035+01:00|status:open"
  ]
}
```

### Temporal Features (Always show timestamps)
```bash
# History queries show timestamps
curl -X GET /api/v1/entities/history?id=XXX

# Time-travel queries
curl -X GET /api/v1/entities/as-of?id=XXX&as_of=2025-05-18T10:00:00Z
```

## Summary

- **Normal Operations**: Work with tags as usual, no timestamps needed
- **Storage Layer**: Automatically handles all temporal tracking
- **API Layer**: Strips timestamps by default for easier use
- **Temporal Queries**: Use history/as-of endpoints when you need time-based data
- **Optional**: Use `include_timestamps=true` only when you need raw temporal data

The temporal system is designed to be invisible during normal use while providing full temporal capabilities when needed!