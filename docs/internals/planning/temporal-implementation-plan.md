# Temporal Implementation Plan

## Phase 1: Modify Core Models

### Update Entity Model
```go
// entity.go - Add helper methods
func (e *Entity) AddTag(tag string) {
    timestamp := time.Now().Format(time.RFC3339Nano)
    e.Tags = append(e.Tags, fmt.Sprintf("%s.%s", timestamp, tag))
}

func (e *Entity) GetTagsWithoutTimestamp() []string {
    result := []string{}
    for _, tag := range e.Tags {
        parts := strings.SplitN(tag, ".", 2)
        if len(parts) == 2 {
            result = append(result, parts[1])
        } else {
            result = append(result, tag)
        }
    }
    return result
}
```

### Update Relationship Model
```go
// entity_relationship.go
type EntityRelationship struct {
    ID               string    `json:"id"`
    SourceID         string    `json:"source_id"`
    RelationshipType string    `json:"relationship_type"`
    TargetID         string    `json:"target_id"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}
```

## Phase 2: Update Storage Layer

### Binary Storage Updates
```go
// entity_repository.go
func (r *EntityRepository) Create(entity *Entity) error {
    // Ensure all tags have timestamps
    timestampedTags := []string{}
    timestamp := time.Now().Format(time.RFC3339Nano)
    
    for _, tag := range entity.Tags {
        if !strings.Contains(tag, ".") {
            timestampedTags = append(timestampedTags, fmt.Sprintf("%s.%s", timestamp, tag))
        } else {
            timestampedTags = append(timestampedTags, tag)
        }
    }
    entity.Tags = timestampedTags
    
    // Continue with storage...
}
```

## Phase 3: Update API Layer

### Transparent Timestamp Handling
```go
// entity_handler.go
func (h *EntityHandler) Create(w http.ResponseWriter, r *http.Request) {
    var entity Entity
    // ... decode request ...
    
    // Tags come in without timestamps
    // Storage layer adds them automatically
    
    err := h.repo.Create(&entity)
    
    // Response strips timestamps by default
    response := entity
    response.Tags = entity.GetTagsWithoutTimestamp()
    
    json.NewEncoder(w).Encode(response)
}
```

### Add Temporal Query Options
```go
// Add query parameter handling
includeTimestamps := r.URL.Query().Get("include_timestamps") == "true"

if !includeTimestamps {
    entity.Tags = entity.GetTagsWithoutTimestamp()
}
```

## Phase 4: Migration Tool

### Create Migration Script
```go
// share/utilities/migrate_temporal.go
func migrateEntity(entity *Entity) *Entity {
    migrated := *entity
    timestamp := entity.CreatedAt
    
    newTags := []string{}
    for _, tag := range entity.Tags {
        if !strings.Contains(tag, ".") {
            newTags = append(newTags, fmt.Sprintf("%s.%s", timestamp, tag))
        } else {
            newTags = append(newTags, tag)
        }
    }
    migrated.Tags = newTags
    
    return &migrated
}
```

## Phase 5: Testing

### Test Scenarios
1. Create entity without timestamps → stored with timestamps
2. Query entity → returns without timestamps by default
3. Query with include_timestamps → returns with timestamps
4. Temporal queries work correctly
5. Migration tool handles existing data

## Rollout Strategy

1. **Backward Compatible**: Old clients continue to work
2. **Gradual Migration**: Run migration tool on existing data
3. **Feature Flag**: Enable temporal features gradually
4. **Documentation**: Update API docs with temporal options

This implementation maintains full backward compatibility while adding comprehensive temporal capabilities to EntityDB.