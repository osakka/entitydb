# EntityDB Advanced Query Implementation

## Summary

This document describes the advanced query functionality implemented for EntityDB.

## Features Implemented

### 1. Query Builder Pattern

Created an `EntityQuery` struct in `/opt/entitydb/src/models/entity_query.go` that provides a fluent API for building complex queries:

```go
type EntityQuery struct {
    repo       EntityRepository
    tags       []string
    wildcards  []string
    content    string
    namespace  string
    limit      int
    offset     int
    orderBy    string
    orderDir   string
    filters    []Filter
    operators  []string  // AND/OR operators for filters
}
```

### 2. Advanced Filtering

Implemented support for filtering by:
- Field comparisons (eq, ne, gt, lt, gte, lte, like, in)
- Tag namespaces (e.g., `tag:type` filters)
- Content types and values
- Created/Updated timestamps
- Tag count

Example:
```go
query.AddFilter("tag_count", "gt", 2)
query.AddFilter("content_type", "eq", "title")
query.AddFilter("created_at", "gte", "2025-01-01T00:00:00Z")
```

### 3. Sorting Capabilities

Added sorting support for:
- Created/Updated timestamps
- Entity ID
- Tag count

Both ascending and descending order are supported.

### 4. Pagination

Full pagination support with:
- Limit: Maximum number of results
- Offset: Starting position for results

### 5. API Endpoint

Created `/api/v1/entities/query` endpoint that accepts query parameters:
- `filter`: Field to filter on
- `operator`: Comparison operator
- `value`: Value to compare against
- `sort`: Field to sort by
- `order`: Sort direction (asc/desc)
- `limit`: Result limit
- `offset`: Result offset

### 6. Authentication Fix

Fixed RBAC middleware to properly use session authentication by:
- Updating `EntityHandlerRBAC` to use `RequirePermission` instead of `RBACMiddleware`
- Adding session manager to RBAC handlers
- Properly chaining authentication and authorization middleware

## Example Usage

### Basic Query
```bash
curl -X GET "http://localhost:8085/api/v1/entities/query?limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

### Filter by Tag
```bash
curl -X GET "http://localhost:8085/api/v1/entities/query?filter=tag:type&operator=eq&value=document" \
  -H "Authorization: Bearer $TOKEN"
```

### Sort by Creation Date
```bash
curl -X GET "http://localhost:8085/api/v1/entities/query?sort=created_at&order=desc&limit=5" \
  -H "Authorization: Bearer $TOKEN"
```

### Filter by Tag Count
```bash
curl -X GET "http://localhost:8085/api/v1/entities/query?filter=tag_count&operator=gt&value=2" \
  -H "Authorization: Bearer $TOKEN"
```

## Implementation Details

### Files Modified

1. `/opt/entitydb/src/models/entity_query.go` - Core query builder implementation
2. `/opt/entitydb/src/api/entity_handler.go` - Added QueryEntities handler
3. `/opt/entitydb/src/api/swagger_models.go` - Added QueryEntityResponse type
4. `/opt/entitydb/src/api/entity_handler_rbac.go` - Updated to use proper auth
5. `/opt/entitydb/src/main.go` - Added query route and updated handler creation
6. `/opt/entitydb/src/storage/binary/entity_repository.go` - Added Query() method
7. `/opt/entitydb/src/docs/swagger.json` - Added query endpoint documentation

### Key Algorithms

1. **Filtering**: Iterates through entities and evaluates each against filters
2. **Sorting**: Uses bubble sort for simplicity (can be optimized with quicksort)
3. **Pagination**: Applied after filtering and sorting
4. **Time Comparison**: Parses string timestamps to compare temporal data

### Performance Considerations

- All queries load entities into memory for filtering
- Sorting uses O(n²) bubble sort algorithm
- Could be optimized with:
  - Index-based filtering
  - More efficient sorting algorithms
  - Lazy loading of entity data

## Testing

A comprehensive test script was created at `/opt/entitydb/share/tests/test_advanced_query.sh` that tests:
- Authentication
- Entity creation
- Query by tag
- Sorting
- Tag count filtering
- Content type filtering
- Pagination

## Future Enhancements

1. **Composite Filters**: Support for complex AND/OR combinations
2. **Aggregations**: COUNT, SUM, AVG operations
3. **Full-text Search**: Better content searching
4. **Performance**: Index-based queries, caching
5. **Query Language**: SQL-like or GraphQL interface