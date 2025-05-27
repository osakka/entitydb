# Data Integrity Implementation

## Overview

EntityDB now includes comprehensive data integrity features to ensure reliability and consistency across all operations. This implementation was added to address database corruption issues and provide end-to-end data verification.

## Key Components

### 1. Operation Tracking
- **Location**: `/opt/entitydb/src/models/operation_tracking.go`
- **Purpose**: Track every data operation with unique IDs
- **Features**:
  - Unique operation IDs for traceability
  - Operation lifecycle tracking (start, progress, complete, failed)
  - Metadata storage for debugging
  - Thread-safe concurrent access

### 2. Checksum Verification
- **Integrated Into**: Writer and Reader components
- **Algorithm**: SHA256
- **Storage**: Checksums stored as temporal tags
- **Verification**: Automatic verification on read operations

### 3. Transaction Management
- **Location**: `/opt/entitydb/src/storage/binary/transaction_manager.go`
- **Features**:
  - Two-phase commit protocol
  - Atomic operations across multiple entities
  - Automatic rollback on failure
  - Transaction state persistence

### 4. Recovery System
- **Location**: `/opt/entitydb/src/storage/binary/recovery.go`
- **Capabilities**:
  - WAL-based recovery
  - Cache recovery
  - Index repair
  - Corrupted entity recovery
  - Multiple recovery sources (WAL, cache, backups)

### 5. Real-time Monitoring
- **Endpoint**: `/api/v1/integrity/metrics`
- **Dashboard**: `/integrity.html`
- **Metrics**:
  - Operation success/failure rates
  - Checksum verification statistics
  - Recovery operations
  - Transaction states

## Implementation Details

### Tag Index Fix
The tag indexing system was fixed to properly handle temporal tags:
- Non-timestamped versions of temporal tags are now indexed for easier searching
- Timestamps are parsed as Unix nanoseconds
- Critical for authentication and tag-based lookups

### Authentication Flow
1. User lookup by tag: `identity:username:USERNAME`
2. Credential retrieval via `has_credential` relationship
3. Password verification: `bcrypt(password + salt)`
4. Salt stored as credential tag: `salt:SALT_VALUE`

### Relationship Storage
Entity relationships require both fields:
- `Type`: Used by the application layer
- `RelationshipType`: Used by the storage layer
- Both must be set for proper functionality

## Configuration

### Environment Variables
```bash
# Enable integrity features
ENTITYDB_ENABLE_CHECKSUMS=true
ENTITYDB_ENABLE_TRANSACTIONS=true
ENTITYDB_RECOVERY_MODE=auto

# Operation tracking
ENTITYDB_OPERATION_LOG_LEVEL=INFO
ENTITYDB_OPERATION_RETENTION_DAYS=7
```

### Feature Flags
```json
{
  "integrity:checksums": true,
  "integrity:transactions": true,
  "integrity:recovery": true,
  "integrity:monitoring": true
}
```

## Usage

### Creating Entities with Integrity
```go
// Operations are automatically tracked
entity := &models.Entity{
    ID: "test_entity",
    Tags: []string{"type:test"},
    Content: []byte("test content"),
}

// Checksum automatically calculated and stored
err := repo.Create(entity)
```

### Transaction Example
```go
tx := tm.Begin()
defer tx.Rollback()

// Multiple operations
tx.Create(entity1)
tx.Update(entity2)
tx.Delete(entity3)

// Atomic commit
err := tx.Commit()
```

### Recovery Example
```go
// Automatic recovery on corruption
entity, err := repo.GetByID("corrupted_entity")
if err != nil {
    // Recovery manager automatically attempts recovery
    // from WAL, cache, or backups
}
```

## Monitoring

### Health Check
```bash
curl https://localhost:8085/health
```

### Integrity Metrics
```bash
curl https://localhost:8085/api/v1/integrity/metrics
```

### Dashboard
Access the integrity dashboard at: https://localhost:8085/integrity.html

## Best Practices

1. **Always use transactions** for multi-entity operations
2. **Monitor checksum failures** - they indicate potential corruption
3. **Regular WAL rotation** to prevent unbounded growth
4. **Backup before recovery** operations
5. **Enable all integrity features** in production

## Troubleshooting

### Common Issues

1. **"Checksum mismatch" errors**
   - Indicates data corruption
   - Recovery will attempt automatically
   - Check logs for root cause

2. **"Transaction timeout" errors**
   - Increase transaction timeout
   - Check for deadlocks
   - Review operation complexity

3. **"Recovery failed" errors**
   - Check WAL integrity
   - Verify backup availability
   - Manual intervention may be required

### Debug Commands

```bash
# Check entity integrity
./bin/entitydb verify --entity-id=ENTITY_ID

# Rebuild tag index
./bin/entitydb rebuild-index --data-path=./var

# Verify WAL integrity
./bin/entitydb verify-wal --wal-path=./var/entitydb.wal
```

## Performance Impact

- Checksum calculation: ~5% overhead on writes
- Transaction management: ~10% overhead for multi-entity operations
- Recovery: Minimal impact (only on corruption)
- Monitoring: <1% overhead

## Future Enhancements

1. Merkle tree for efficient verification
2. Distributed transaction support
3. Point-in-time recovery
4. Automated corruption prevention
5. Machine learning for anomaly detection