# ADR-030: Unified Temporal Deletion Architecture

**Date**: 2025-06-22  
**Status**: Accepted  
**Context**: EntityDB v2.34.2 Deletion System Architecture

## Context

EntityDB currently lacks deletion functionality while maintaining its temporal database principles. Traditional hard deletion would break the temporal model and audit trail capabilities. We need a deletion system that:

1. Preserves temporal consistency and audit trails
2. Enables undelete/recovery capabilities  
3. Provides policy-driven data lifecycle management
4. Maintains unified file format architecture
5. Supports compliance requirements (GDPR, etc.)

## Decision

We implement a **Unified Temporal Deletion Architecture** with three-tier lifecycle management:

```
ACTIVE → SOFT_DELETED → ARCHIVED → PURGED
   ↓         ↓            ↓          ↓
Live     Recoverable   Cold Store  Gone
```

## Architecture Design

### **Core Principles**

1. **Temporal Consistency**: All operations preserve temporal tag model
2. **Single Source of Truth**: Unified .edb file with logical sections
3. **Policy-Driven**: Configurable retention and lifecycle policies
4. **Audit Complete**: Full trail of all deletion and recovery operations
5. **Zero Regressions**: No impact on existing functionality

### **Entity Lifecycle States**

```go
type EntityLifecycleState string

const (
    StateActive      EntityLifecycleState = "active"
    StateSoftDeleted EntityLifecycleState = "soft_deleted"  
    StateArchived    EntityLifecycleState = "archived"
    StatePurged      EntityLifecycleState = "purged"
)

type LifecycleTransition struct {
    FromState EntityLifecycleState
    ToState   EntityLifecycleState
    Timestamp time.Time
    UserID    string
    Reason    string
    Policy    string
}
```

### **Status Tag Architecture**

**Active Entity**:
```
type:user
status:active
created_at:1234567890
updated_at:1234567891
```

**Soft Deleted Entity**:
```
type:user
status:soft_deleted|1234567892
deleted_by:admin-user-id
delete_reason:user_requested
deletion_policy:user_data_policy
retention_until:1234567892+policy.soft_retention
```

**Archived Entity**:
```
type:user
status:archived|1234567893
archived_by:system-collector
archive_reason:retention_policy
archive_policy:user_data_policy
purge_after:1234567893+policy.purge_after
```

### **Unified File Format Enhancement**

```
┌─────────────────────────────────────────────────────────────┐
│                    UNIFIED .EDB FILE                        │
├─────────┬─────────┬─────────┬─────────┬─────────┬───────────┤
│ HEADER  │ ACTIVE  │ DELETED │ ARCHIVE │   WAL   │   INDEX   │
│         │  DATA   │  DATA   │  DATA   │         │           │
├─────────┼─────────┼─────────┼─────────┼─────────┼───────────┤
│ Format  │ Live    │ Soft    │ Cold    │ Durable │ Fast      │
│ Version │ Entities│ Deleted │ Storage │ Logging │ Lookups   │
│ Offsets │ Full    │ Limited │ Audit   │ Recovery│ Temporal  │
│ Config  │ Index   │ Index   │ Only    │ Replay  │ Tags      │
└─────────┴─────────┴─────────┴─────────┴─────────┴───────────┘
```

**Enhanced Header Structure**:
```go
type UnifiedHeader struct {
    // Existing fields
    Magic           [4]byte
    Version         uint32
    DataOffset      uint64
    DataSize        uint64
    WALOffset       uint64
    WALSize         uint64
    IndexOffset     uint64
    IndexSize       uint64
    
    // New deletion sections
    DeletedOffset   uint64    // Soft deleted entities
    DeletedSize     uint64
    ArchiveOffset   uint64    // Archived entities  
    ArchiveSize     uint64
    
    // Deletion metadata
    DeletionPolicies map[string]RetentionPolicy
    LastCollectionRun time.Time
}
```

### **Retention Policy Framework**

```go
type RetentionPolicy struct {
    Name              string        `json:"name"`
    SoftRetention     time.Duration `json:"soft_retention"`     // How long in soft_deleted
    ArchiveAfter      time.Duration `json:"archive_after"`      // Move to archive
    PurgeAfter        time.Duration `json:"purge_after"`        // Complete removal
    RequiresApproval  bool          `json:"requires_approval"`  // Manual approval needed
    AuditLevel        AuditLevel    `json:"audit_level"`        // How much to log
    CascadeRules      []CascadeRule `json:"cascade_rules"`      // Dependent entities
}

type AuditLevel string
const (
    AuditMinimal AuditLevel = "minimal"  // Basic deletion log
    AuditFull    AuditLevel = "full"     // Complete operation history
    AuditForensic AuditLevel = "forensic" // Maximum detail for compliance
)

type CascadeRule struct {
    ParentPattern string      `json:"parent_pattern"`   // e.g., "type:user"
    ChildPattern  string      `json:"child_pattern"`    // e.g., "created_by:${parent.id}"
    Action        CascadeAction `json:"action"`         // What to do with children
}

type CascadeAction string
const (
    CascadeDelete  CascadeAction = "delete"   // Delete children too
    CascadeOrphan  CascadeAction = "orphan"   // Remove parent reference
    CascadeBlock   CascadeAction = "block"    // Prevent deletion if children exist
)
```

**Built-in Policies**:
```go
var DefaultPolicies = map[string]RetentionPolicy{
    "user_data": {
        Name:             "User Data Protection",
        SoftRetention:    30 * 24 * time.Hour,   // 30 days
        ArchiveAfter:     90 * 24 * time.Hour,   // 90 days
        PurgeAfter:      365 * 24 * time.Hour,   // 1 year
        RequiresApproval: true,
        AuditLevel:       AuditForensic,
        CascadeRules: []CascadeRule{
            {
                ParentPattern: "type:user",
                ChildPattern:  "created_by:${parent.id}",
                Action:        CascadeDelete,
            },
        },
    },
    "system_metrics": {
        Name:             "System Metrics Cleanup",
        SoftRetention:    7 * 24 * time.Hour,    // 7 days
        ArchiveAfter:     30 * 24 * time.Hour,   // 30 days
        PurgeAfter:       90 * 24 * time.Hour,   // 90 days
        RequiresApproval: false,
        AuditLevel:       AuditMinimal,
    },
    "temporary_data": {
        Name:             "Temporary Data Cleanup",
        SoftRetention:    24 * time.Hour,        // 1 day
        ArchiveAfter:     7 * 24 * time.Hour,    // 7 days
        PurgeAfter:       30 * 24 * time.Hour,   // 30 days
        RequiresApproval: false,
        AuditLevel:       AuditMinimal,
    },
}
```

### **Deletion Collector Architecture**

```go
type DeletionCollector struct {
    repository    Repository
    policies      map[string]RetentionPolicy
    config        CollectorConfig
    metrics       DeletionMetrics
    auditLogger   AuditLogger
}

type CollectorConfig struct {
    RunInterval       time.Duration
    BatchSize         int
    MaxConcurrency    int
    SafetyChecks      bool
    DryRun           bool
}

type DeletionMetrics struct {
    SoftDeleted       int64
    Archived          int64  
    Purged            int64
    RecoveredEntities int64
    PolicyViolations  int64
    CollectionRuns    int64
}

func (dc *DeletionCollector) ProcessDeletions() error {
    // Three-phase collection
    
    // Phase 1: Process soft deleted entities for archival
    softDeleted := dc.repository.ListByTag("status:soft_deleted")
    for _, entity := range softDeleted {
        policy := dc.getPolicyForEntity(entity)
        if dc.shouldArchive(entity, policy) {
            dc.archiveEntity(entity, policy)
        }
    }
    
    // Phase 2: Process archived entities for purging
    archived := dc.repository.ListByTag("status:archived")
    for _, entity := range archived {
        policy := dc.getPolicyForEntity(entity)
        if dc.shouldPurge(entity, policy) {
            dc.purgeEntity(entity, policy)
        }
    }
    
    // Phase 3: Update metrics and audit logs
    dc.updateMetrics()
    dc.auditCollectionRun()
    
    return nil
}
```

### **API Design Specification**

**Core Deletion Operations**:
```
DELETE /api/v1/entities/{id}                    # Soft delete
POST   /api/v1/entities/{id}/undelete           # Recover from soft delete
POST   /api/v1/entities/{id}/archive            # Force archive (admin)
POST   /api/v1/entities/{id}/purge              # Force purge (admin)
```

**Query Operations**:
```
GET /api/v1/entities/deleted                    # List soft deleted
GET /api/v1/entities/archived                   # List archived  
GET /api/v1/entities/{id}/history?include_deleted=true  # Full history
```

**Policy Management**:
```
GET    /api/v1/admin/deletion/policies          # List policies
POST   /api/v1/admin/deletion/policies          # Create policy
PUT    /api/v1/admin/deletion/policies/{name}   # Update policy
DELETE /api/v1/admin/deletion/policies/{name}   # Remove policy
```

**Collection Management**:
```
GET  /api/v1/admin/deletion/collector/status    # Collector status
POST /api/v1/admin/deletion/collector/run       # Force collection run
GET  /api/v1/admin/deletion/pending             # Entities pending action
POST /api/v1/admin/deletion/approve/{id}        # Approve deletion
```

### **RBAC Integration**

**Required Permissions**:
```go
const (
    PermEntityDelete      = "entity:delete"       // Soft delete entities
    PermEntityUndelete    = "entity:undelete"     // Recover entities
    PermEntityPurge       = "entity:purge"        // Force purge (admin)
    PermDeletionView      = "deletion:view"       # View deleted entities
    PermDeletionAdmin     = "deletion:admin"      # Manage policies
    PermDeletionCollector = "deletion:collector"  # Control collector
)
```

**Permission Matrix**:
```
Operation           | User | Admin | Super Admin
--------------------|------|-------|------------
Soft Delete Own     |  ✓   |   ✓   |     ✓
Soft Delete Any     |  ✗   |   ✓   |     ✓  
Undelete Own        |  ✓   |   ✓   |     ✓
Undelete Any        |  ✗   |   ✓   |     ✓
View Deleted Own    |  ✓   |   ✓   |     ✓
View Deleted Any    |  ✗   |   ✓   |     ✓
Force Archive       |  ✗   |   ✗   |     ✓
Force Purge         |  ✗   |   ✗   |     ✓
Manage Policies     |  ✗   |   ✗   |     ✓
Control Collector   |  ✗   |   ✗   |     ✓
```

## Implementation Strategy

### **Phase 1: Core Infrastructure** 
1. Enhanced entity lifecycle status management
2. Retention policy framework  
3. Deletion metadata tracking

### **Phase 2: File Format Enhancement**
1. Extended unified header with deletion sections
2. Section-aware readers and writers
3. Migration tools for existing data

### **Phase 3: Collection System**
1. Background deletion collector
2. Policy engine implementation
3. Metrics and monitoring

### **Phase 4: API Integration**
1. RESTful deletion endpoints
2. RBAC permission integration
3. Comprehensive error handling

### **Phase 5: Testing & Validation**
1. Unit test suite for all components
2. Integration testing with existing systems
3. Performance validation and optimization

## Consequences

### **Positive**
- **Temporal Consistency**: Full preservation of audit trails
- **Policy Flexibility**: Configurable retention for different data types
- **Compliance Ready**: GDPR and other regulatory requirements supported
- **Undelete Capability**: Recovery from accidental deletions
- **Performance Optimized**: Separate sections reduce query overhead
- **Zero Regressions**: Existing functionality completely preserved

### **Neutral**
- **Additional Complexity**: More sophisticated deletion workflow
- **Storage Overhead**: Retention of deleted data until purge
- **Background Processing**: Collector runs periodically

### **Negative**
- **None Identified**: All functionality is additive with no breaking changes

## Monitoring and Observability

### **Metrics to Track**
```go
type DeletionMetrics struct {
    // Operational metrics
    EntitiesSoftDeleted    Counter
    EntitiesArchived       Counter
    EntitiesPurged         Counter
    EntitiesRecovered      Counter
    
    // Performance metrics  
    CollectionDuration     Histogram
    PolicyEvaluationTime   Histogram
    ArchivalLatency        Histogram
    
    // Error metrics
    PolicyViolations       Counter
    CollectionErrors       Counter
    RecoveryFailures       Counter
    
    // Resource metrics
    DeletedSectionSize     Gauge
    ArchiveSectionSize     Gauge
    OrphanedEntities       Gauge
}
```

### **Alerting Rules**
- Collection failures exceeding threshold
- Policy violations requiring attention
- Archive section growth beyond limits
- Recovery operation failures

## Migration Strategy

### **Backward Compatibility**
- All existing entities remain unchanged
- No modification to existing API endpoints
- Current query behavior preserved

### **Data Migration**
- No migration required for existing data
- New installations get enhanced format
- Gradual adoption of deletion policies

### **Rollback Plan**
- Deletion sections can be ignored by older versions
- Policy configurations stored separately
- Feature can be disabled via configuration

## Future Enhancements

### **Advanced Features**
1. **Intelligent Archival**: ML-based prediction of deletion candidates
2. **Cross-Reference Protection**: Prevent deletion of referenced entities
3. **Batch Operations**: Bulk deletion with progress tracking
4. **Export/Import**: Archive to external storage systems
5. **Compliance Automation**: Automatic GDPR compliance workflows

### **Integration Opportunities**
1. **Backup Systems**: Integration with backup and recovery tools
2. **Analytics**: Deletion pattern analysis and optimization
3. **External Storage**: Cold archive to S3/cloud storage
4. **Workflow Systems**: Integration with approval workflows

## Related ADRs

- ADR-027: Database File Unification (Foundation)
- ADR-028: WAL Corruption Prevention (Security)
- ADR-029: Memory Optimization (Performance)

## Decision Outcome

**Status**: ✅ **ACCEPTED**

This Unified Temporal Deletion Architecture provides EntityDB with production-grade deletion capabilities while preserving all temporal database benefits and maintaining zero regressions to existing functionality.

**Key Benefits**:
- **Temporal Consistency**: Complete audit trail preservation
- **Policy Flexibility**: Configurable retention for compliance
- **Recovery Capability**: Full undelete functionality
- **Performance Optimized**: Efficient section-based organization
- **Production Ready**: Comprehensive monitoring and error handling

**Implementation Timeline**: 6 systematic steps with surgical precision and bar-raising quality at each phase.