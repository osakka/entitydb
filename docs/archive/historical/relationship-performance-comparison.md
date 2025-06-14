# Relationship Implementation Performance Comparison

## Current Implementation Performance

### Storage Overhead

**Relationship as Entity**:
- Each relationship = 1 full entity (~500-1000 bytes)
- Typical relationship entity:
  - ID: 36 bytes (UUID)
  - Tags: 4-6 tags × ~30 bytes = 120-180 bytes
  - Content (JSON): ~100-200 bytes
  - Metadata overhead: ~100 bytes
  - **Total: ~450-500 bytes per relationship**

**Example**: User with 10 relationships = 10 × 500 = 5KB additional storage

### Query Performance

#### Authentication Flow (Current)
```
1. Find user by username tag
   - ListByTag("identity:username:john") 
   - O(1) with tag index
   - 1 disk read

2. Get user's credential
   - GetRelationshipsBySource(userID)
     - ListByTag("_source:user_123")
     - O(1) with tag index
     - Returns relationship entities (1-5 typically)
   - For each relationship:
     - GetByID(relationshipID) - 1 disk read
     - Parse relationship to get targetID
     - GetByID(credentialID) - 1 disk read
   
Total: 3-7 disk reads for authentication
```

#### Finding All Related Entities
```
GetRelationshipsBySource(entityID):
1. ListByTag("_source:entity_123") - O(1)
2. For each relationship (N):
   - Parse relationship entity
   - Extract targetID
   
Total: 1 + N disk reads
```

### Index Overhead

Current tag indexes maintained:
- `type:relationship` → [all relationship IDs]
- `_source:X` → [relationship IDs where source=X]
- `_target:X` → [relationship IDs where target=X]
- `_relationship:Y` → [relationship IDs of type Y]

For 10,000 entities with avg 5 relationships each:
- 50,000 relationship entities
- 200,000+ index entries (4 indexes per relationship)

## Tag-Based Implementation Performance

### Storage Overhead

**Relationship as Tag**:
- Each relationship = 1 tag (~50 bytes)
- Tag format: `relationship:type:target_id`
- With timestamp: `2025-01-06T12:00:00.000000000Z|relationship:has_credential:cred_123`
- **Total: ~80 bytes per relationship**

**Storage Savings**: 450 bytes → 80 bytes = **82% reduction**

### Query Performance

#### Authentication Flow (Tag-Based)
```
1. Find user by username tag
   - ListByTag("identity:username:john")
   - O(1) with tag index
   - 1 disk read

2. Get user's credential
   - User entity already loaded with all tags
   - Extract credential IDs from tags: O(T) where T = number of tags
   - GetByID(credentialID) - 1 disk read
   
Total: 2 disk reads (66% reduction)
```

#### Finding All Related Entities
```
GetRelationships(entity):
1. Entity already loaded with tags
2. Filter tags by prefix "relationship:" - O(T)
3. Extract target IDs

Total: 0 additional disk reads (already have entity)
```

### Index Overhead

Tag indexes maintained:
- `relationship:has_credential:cred_123` → [user_123]
- `relationship:member_of:group_456` → [user_123, user_789]
- etc.

For 10,000 entities with avg 5 relationships:
- 0 additional entities
- 50,000 index entries (1 per relationship)
- **75% reduction in index entries**

## Performance Comparison Summary

| Metric | Current (Entity-Based) | Proposed (Tag-Based) | Improvement |
|--------|------------------------|---------------------|-------------|
| **Storage per relationship** | 450-500 bytes | 80 bytes | 82% less |
| **Auth flow disk reads** | 3-7 reads | 2 reads | 66% less |
| **Get relationships** | 1 + N reads | 0 reads | 100% less |
| **Index entries** | 200,000+ | 50,000 | 75% less |
| **Create relationship** | 1 write | 1-2 writes | Similar |
| **Delete relationship** | 1 read + 1 write | 1-2 writes | Similar |

## Benchmarking Scenarios

### Scenario 1: User Authentication (1M users)
- Current: 3M-7M disk reads total
- Tag-based: 2M disk reads total
- **Performance gain: 33-71%**

### Scenario 2: Permission Checks (RBAC)
- Current: Check user → Get relationships → Get roles → Get permissions
  - 4+ sequential disk reads
- Tag-based: User entity has all relationship tags
  - 1 disk read + in-memory filtering
- **Performance gain: 75%+**

### Scenario 3: Dataset Entity Listing
- Current: List entities → For each, check relationships
  - N + (N × R) reads where R = avg relationships
- Tag-based: List entities with tag filter
  - N reads (relationships in tags)
- **Performance gain: R× improvement**

## Memory Impact

### Current Implementation
- Relationship entities cached separately
- 50,000 relationships × 500 bytes = 25MB

### Tag-Based Implementation  
- Tags part of entity (already cached)
- No additional memory for relationships
- **Memory savings: 25MB**

## Query Pattern Analysis

### Queries That Improve
1. "Get all relationships for entity" - 100% improvement
2. "Check if relationship exists" - No disk read needed
3. "Get entities with relationship to X" - Direct tag query
4. "Count relationships" - In-memory from tags

### Queries That May Degrade
1. "Get all relationships of type Y" - Must scan more entities
   - Mitigation: Maintain type-specific indexes
2. "Get relationship metadata" - Stored in value tags
   - Mitigation: Efficient tag parsing

## Conclusion

The tag-based approach offers significant performance improvements for most common use cases:

- **82% storage reduction**
- **66-75% fewer disk reads for common operations**
- **75% reduction in index overhead**
- **Eliminates need for relationship repository**

The main tradeoff is slightly more complex queries for finding all relationships of a specific type across all entities. However, this can be mitigated with proper indexing strategies.

For EntityDB's typical use cases (auth, RBAC, entity associations), the tag-based approach would provide substantial performance benefits while simplifying the architecture.