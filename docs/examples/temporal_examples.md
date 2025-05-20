# EntityDB Temporal Examples

## Basic Tag Creation with Timestamps

When you create an entity with tags, EntityDB automatically adds timestamps:

```bash
# Create entity
curl -X POST http://localhost:8085/api/v1/entities \
  -H "Content-Type: application/json" \
  -d '{
    "type": "user",
    "title": "John Doe",
    "tags": ["status:active", "role:developer"]
  }'
```

Response shows both temporal and simple tags:
```json
{
  "id": "ent_04a144bec506",
  "tags": [
    "2025-05-17T18:01:12.207179095.status=active",
    "status:active",
    "2025-05-17T18:01:12.207192454.role=developer",
    "role:developer"
  ]
}
```

## Time-Travel Queries (Future API)

```bash
# Get entity as it was on January 1st
entitydb-cli entity get --id=ent_123 --as-of="2025-01-01T00:00:00"

# See entity history for past week
entitydb-cli entity history --id=ent_123 --since="7d"

# Find all changes in January
entitydb-cli entity changes --from="2025-01-01" --to="2025-01-31"
```

## Tracking Changes Over Time

```go
// Example: Track status changes
entity, _ := repo.GetByID("ent_123")

statusHistory := []StatusChange{}
for _, tag := range entity.Tags {
    if strings.Contains(tag, ".status=") {
        timestamp, _ := extractTimestamp(tag)
        value := extractValue(tag)
        statusHistory = append(statusHistory, StatusChange{
            Time:   timestamp,
            Status: value,
        })
    }
}

// Result: Complete status history
// 2025-01-01T10:00:00 - status:pending
// 2025-01-15T14:30:00 - status:active
// 2025-02-01T09:15:00 - status:completed
```

## Audit Trail Example

```go
// See who changed what when
func GetAuditTrail(entityID string) []AuditEntry {
    entity, _ := repo.GetByID(entityID)
    
    entries := []AuditEntry{}
    for _, tag := range entity.Tags {
        timestamp, _ := extractTimestamp(tag)
        namespace, value := parseTag(tag)
        
        entries = append(entries, AuditEntry{
            Time:      timestamp,
            User:      entity.UpdatedBy, // From JWT token
            Action:    "tag_added",
            Namespace: namespace,
            Value:     value,
        })
    }
    
    return entries
}
```

## Temporal Relationships

```json
{
  "id": "ent_task_123",
  "tags": [
    "2025-01-01T10:00:00.000000000.rel:assigned_to=user_456",
    "2025-01-15T14:30:00.000000000.rel:assigned_to=user_789",
    "2025-02-01T09:15:00.000000000.status=completed"
  ]
}
```

This shows:
- Task initially assigned to user_456 on Jan 1
- Reassigned to user_789 on Jan 15
- Completed on Feb 1

## Compliance Reporting

```bash
# Generate compliance report for Q1
entitydb-cli report compliance \
  --from="2025-01-01" \
  --to="2025-03-31" \
  --entity-type="transaction"

# Output: Complete audit trail of all transactions
```

## Performance Analysis

```bash
# Find when performance degraded
entitydb-cli analyze performance \
  --metric="response_time" \
  --threshold="500ms" \
  --range="30d"

# Shows exactly when response times exceeded 500ms
```

## Configuration Rollback

```bash
# See config at specific time
entitydb-cli entity get \
  --id="config_production" \
  --as-of="2025-01-15T10:00:00"

# Restore to that state
entitydb-cli entity restore \
  --id="config_production" \
  --to="2025-01-15T10:00:00"
```

## Benefits Over Traditional Approaches

### Traditional Audit Table
```sql
-- Requires separate audit table
CREATE TABLE user_audit (
    id INT PRIMARY KEY,
    user_id INT,
    field_name VARCHAR(50),
    old_value TEXT,
    new_value TEXT,
    changed_by INT,
    changed_at TIMESTAMP
);

-- Complex queries for history
SELECT * FROM user_audit 
WHERE user_id = 123 
AND changed_at <= '2025-01-15'
ORDER BY changed_at;
```

### EntityDB Approach
```bash
# History is built-in
entitydb-cli entity get --id=user_123 --as-of="2025-01-15"

# No extra tables, triggers, or complex queries needed
```

## Real-World Use Cases

### 1. Debugging Production Issues
```bash
# When did the issue start?
entitydb-cli entity changes --id=system_config --last="24h"

# What was the state before the issue?
entitydb-cli entity get --id=system_config --as-of="2025-01-15T09:00:00"
```

### 2. Security Incident Investigation
```bash
# Who accessed what when?
entitydb-cli audit trail --entity=secure_document --last="7d"

# What permissions did user have at time of access?
entitydb-cli entity get --id=user_456 --as-of="2025-01-15T14:30:00" | grep "rbac:perm"
```

### 3. Compliance Demonstration
```bash
# Prove state at audit date
entitydb-cli entity snapshot --type=financial --date="2024-12-31"

# Show complete change history
entitydb-cli entity history --type=financial --year="2024"
```

## Implementation Notes

The temporal features are already working! Every tag you create is timestamped. The query features shown above are the next logical enhancement to fully leverage this temporal foundation.