# ADR-036: Backup Retention System

## Status
**Accepted** - Implemented in v2.34.4

## Context
EntityDB's WAL integrity system was creating routine backups every 5 minutes without any cleanup mechanism. This led to:
- **Disk Space Exhaustion**: 288 backups per day accumulating indefinitely
- **No Retention Policy**: All backups kept forever regardless of age or relevance
- **Hardcoded Intervals**: 5-minute backup frequency not configurable
- **No Size Limits**: Total backup size could grow without bounds

## Decision
Implement a comprehensive backup retention system with:

1. **Configurable Backup Intervals**
   - `ENTITYDB_BACKUP_INTERVAL` (default: 1 hour)
   - Replaced hardcoded `BACKUP_INTERVAL = 5 * time.Minute`

2. **Time-Based Retention Policies**
   - `ENTITYDB_BACKUP_RETENTION_HOURS` (default: 24)
   - `ENTITYDB_BACKUP_RETENTION_DAYS` (default: 7)
   - `ENTITYDB_BACKUP_RETENTION_WEEKS` (default: 4)
   - Smart algorithm keeps representative backups from each period

3. **Size-Based Limits**
   - `ENTITYDB_BACKUP_MAX_SIZE_MB` (default: 1000)
   - Automatically removes oldest backups when limit exceeded

4. **Emergency Backup Protection**
   - Backups created during corruption events preserved for 24 hours
   - Ensures critical recovery data isn't deleted

## Implementation

### Configuration Integration
```go
// Added to config/config.go
BackupInterval       time.Duration  // How often to create backups
BackupRetentionHours int           // Keep last N hourly backups
BackupRetentionDays  int           // Keep last N daily backups
BackupRetentionWeeks int           // Keep last N weekly backups
BackupMaxSizeMB      int64         // Maximum total backup size
```

### Backup Cleanup Logic
```go
// wal_integrity_system.go - cleanupOldBackups()
func (w *WALIntegritySystem) cleanupOldBackups() error {
    // 1. List and sort backup files by modification time
    // 2. Apply retention policies (hourly/daily/weekly)
    // 3. Check size limits
    // 4. Delete marked backups
    // 5. Preserve emergency backups
}
```

### Single Source of Truth
- WAL integrity system updated to accept config parameter
- No duplicate backup management code
- All backup operations centralized in `wal_integrity_system.go`

## Consequences

### Positive
- **Predictable Disk Usage**: Maximum ~35 backup files instead of unlimited
- **Production Ready**: Configurable for different deployment scenarios
- **Intelligent Coverage**: Keeps recent hourly + daily + weekly backups
- **Emergency Protection**: Critical backups preserved during corruption
- **Zero Downtime**: Cleanup happens automatically after backup creation

### Negative
- Slightly more complex configuration (5 new parameters)
- Backup cleanup adds minor CPU overhead during backup creation

### Neutral
- Default 1-hour interval reduces backup frequency from 5 minutes
- Older backups beyond retention period are permanently deleted

## Testing Results
Successfully tested with:
- Reduced 6 backups to 2 based on 2-hour retention policy
- Cleanup logged: "Cleaned up 6 old backup files"
- No performance impact observed
- Configuration changes take effect on restart

## Related
- ADR-028: WAL Corruption Prevention System (introduced backups)
- ADR-029: Memory Optimization Architecture (bounded resource usage)
- ADR-031: Bar-Raising Metrics Retention (similar retention concepts)