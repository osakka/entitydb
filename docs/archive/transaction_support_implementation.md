# Transaction Support Implementation

## Overview

This document describes the implementation of transaction support in the EntityDB platform. Transaction support allows for atomic operations across multiple database actions, ensuring data integrity even in complex operations like creating an entity and its relationships in a single atomic operation.

## Implementation

The transaction support is implemented at two levels:

1. **Database Level**: Direct transaction support through SQLite's transaction mechanisms.
2. **Repository Level**: High-level transaction support through the TransactionEntityRepository.

## Components

### 1. TransactionManager

The `TransactionManager` provides centralized transaction management for the EntityDB system. It ensures atomic operations across multiple repositories.

**File**: `/opt/entitydb/src/models/sqlite/transaction_manager.go`

**Key Features**:
- Transaction creation, commit, and rollback
- Comprehensive error handling
- Transaction context tracking
- Helper methods for SQL operations within transactions

### 2. TransactionEntityRepository

The `TransactionEntityRepository` extends the standard EntityRepository to support atomic operations.

**File**: `/opt/entitydb/src/models/sqlite/transaction_entity_repository.go`

**Key Features**:
- Transaction-aware CRUD operations
- Atomic operations involving multiple entities
- Atomic operations involving both entities and relationships
- High-level operations like entity merging, cloning, and ownership transfer

### 3. Factory Integration

The transaction-aware repository is integrated into the existing factory system to make it easily accessible throughout the application.

**Files**:
- `/opt/entitydb/src/models/entity_factory.go` - Factory registration functions
- `/opt/entitydb/src/models/repository_factory.go` - Repository factory implementation
- `/opt/entitydb/src/models/sqlite/factory.go` - SQLite-specific factory implementation

## Usage Examples

### Basic Transaction Usage

```go
// Get transaction-aware repository
repoFactory := models.NewRepositoryFactory()
entityRepo, err := repoFactory.CreateTransactionEntityRepository(models.SQLiteRepository, "file:entitydb.db")
if err != nil {
    log.Fatalf("Failed to create transaction entity repository: %v", err)
}

// Use repository with transaction
txRepo := entityRepo.(*sqlite.TransactionEntityRepository)

// Begin transaction
ctx, err := txRepo.Begin()
if err != nil {
    log.Fatalf("Failed to begin transaction: %v", err)
}

// Perform operations within transaction
entity, err := txRepo.CreateWithTx(ctx, myEntity)
if err != nil {
    txRepo.Rollback(ctx) // Roll back on error
    log.Fatalf("Failed to create entity: %v", err)
}

// Create relationship within same transaction
relationship := models.NewEntityRelationship(entity.ID, "parent_of", childID)
err = txRepo.CreateRelationshipWithTx(ctx, relationship)
if err != nil {
    txRepo.Rollback(ctx) // Roll back on error
    log.Fatalf("Failed to create relationship: %v", err)
}

// Commit transaction
err = txRepo.Commit(ctx)
if err != nil {
    log.Fatalf("Failed to commit transaction: %v", err)
}
```

### Using WithTransaction Helper

```go
// Use WithTransaction for simpler transaction handling
err := txRepo.WithTransaction(func(ctx *sqlite.TransactionContext) error {
    // Create entity
    entity, err := txRepo.CreateWithTx(ctx, myEntity)
    if err != nil {
        return fmt.Errorf("failed to create entity: %w", err)
    }
    
    // Create relationship
    relationship := models.NewEntityRelationship(entity.ID, "parent_of", childID)
    if err := txRepo.CreateRelationshipWithTx(ctx, relationship); err != nil {
        return fmt.Errorf("failed to create relationship: %w", err)
    }
    
    return nil // Will commit if no errors
})

if err != nil {
    log.Fatalf("Transaction failed: %v", err)
}
```

### High-Level Atomic Operations

```go
// Create entity with relationships in a single atomic operation
entity := &models.Entity{
    Title: "Parent Entity",
    Tags: []string{"type:parent", "status:active"},
}

relationships := []*models.EntityRelationship{
    models.NewEntityRelationship("", "parent_of", "entity_child_1"),
    models.NewEntityRelationship("", "parent_of", "entity_child_2"),
}

createdEntity, err := txRepo.CreateEntityWithRelationships(entity, relationships)
if err != nil {
    log.Fatalf("Failed to create entity with relationships: %v", err)
}
```

## Benefits

1. **Data Integrity**: Ensures that complex operations either complete fully or not at all.
2. **Simplified Error Handling**: Automatic rollback on error simplifies error handling logic.
3. **Performance**: Batch database operations in single transactions for better performance.
4. **Consistency**: Ensures the database remains in a consistent state even during complex operations.

## Supported Operations

The transaction-aware repository supports the following atomic operations:

1. **CreateEntityWithRelationships**: Create an entity and its relationships in a single transaction.
2. **UpdateEntityWithRelationships**: Update an entity and modify its relationships in a single transaction.
3. **DeleteEntityWithRelationships**: Delete an entity and all its relationships in a single transaction.
4. **TransferEntityOwnership**: Transfer all relationships from one entity to another.
5. **MergeEntities**: Combine two entities into one, merging their tags, content, and relationships.
6. **CloneEntity**: Create a copy of an entity with all its data.

## Implementation Notes

1. **Thread Safety**: The transaction manager is thread-safe using a mutex to synchronize access.
2. **Resource Management**: Transactions are properly closed even in error cases.
3. **Error Handling**: Comprehensive error handling with detailed error messages.
4. **Logging**: Detailed logging of transaction operations for debugging and auditing.

## Conclusion

The transaction support implementation provides a robust foundation for atomic operations in the EntityDB platform. It ensures data integrity while simplifying the implementation of complex operations involving multiple database changes.