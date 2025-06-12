# SQLite Transaction Support in EntityDB

This document describes the transaction support implemented in the EntityDB system SQLite database layer.

## Overview

The EntityDB system uses SQLite as its primary persistence layer. To ensure data integrity and consistency, the system implements a robust transaction management system that allows for atomic operations across multiple database statements.

## Transaction Patterns

The system supports two main transaction patterns:

### 1. Function-Based Transaction (WithTransaction)

The function-based transaction pattern is the preferred approach for most use cases. It provides a clean, safe way to execute multiple database operations within a transaction context.

```go
err := db.WithTransaction(func(tx *sql.Tx) error {
    // Execute multiple operations within the transaction
    _, err := tx.Exec("INSERT INTO entities (id, tags, content) VALUES (?, ?, ?)", 
        entity.ID, tagsJSON, contentJSON)
    if err != nil {
        return err // This will automatically trigger a rollback
    }
    
    // Do more operations...
    
    return nil // This will trigger a commit
})
```

Key advantages:
- Automatic rollback on error
- Clean error handling
- Simplified code structure
- Protection against forgetting to commit or rollback

### 2. Explicit Transaction API

For more complex scenarios where more control is needed, the system provides an explicit transaction API:

```go
// Begin a transaction
tx, err := db.BeginTransaction()
if err != nil {
    return err
}

// Use defer to ensure the transaction is rolled back if not committed
defer func() {
    if !tx.finished {
        tx.Rollback()
    }
}()

// Execute operations within the transaction
_, err = tx.Exec("INSERT INTO entities (id, tags, content) VALUES (?, ?, ?)", 
    entity.ID, tagsJSON, contentJSON)
if err != nil {
    return err
}

// More operations...

// Commit the transaction
if err := tx.Commit(); err != nil {
    return err
}
```

Key advantages:
- More explicit control over the transaction lifecycle
- Ability to conditionally commit or rollback based on complex logic
- Can span multiple functions or methods

## Implementation Details

### Transaction Safety

The transaction implementation includes several safety features:

1. **Mutex Locking**: Each transaction acquires a mutex lock to prevent concurrent modifications to the same database connection.

2. **Automatic Rollback**: Transactions that fail or panic are automatically rolled back.

3. **State Tracking**: The transaction state (committed/rolled back) is tracked to prevent double commits or rollbacks.

4. **Error Propagation**: All errors are properly propagated with context added to help with debugging.

### Transaction Methods

The transaction API provides the following methods:

- **BeginTransaction()**: Starts a new transaction
- **Commit()**: Commits the transaction
- **Rollback()**: Rolls back the transaction
- **Exec()**: Executes a SQL statement within the transaction
- **Query()**: Executes a query that returns multiple rows within the transaction
- **QueryRow()**: Executes a query that returns a single row within the transaction

### Usage in Repositories

All repository implementations in the system should use transactions for operations that modify data or require consistency across multiple operations. Examples include:

- Creating an entity with tags
- Updating an entity and its relationships
- Deleting an entity and all its references

## Best Practices

1. **Use WithTransaction When Possible**: The function-based pattern is safer and cleaner for most use cases.

2. **Keep Transactions Short**: Transactions should be as short as possible to minimize lock contention.

3. **Handle Errors Properly**: Always check for errors and ensure proper rollback/commit.

4. **Be Careful with Nested Transactions**: SQLite doesn't support true nested transactions. Use savepoints instead if needed.

5. **Use Transactions for Multi-Step Operations**: Any operation that requires multiple database changes should use a transaction.

## Example: Entity Creation with Tags

Here's an example of creating an entity with tags using the transaction system:

```go
func (r *EntityRepository) Create(entity *models.Entity) error {
    return r.db.WithTransaction(func(tx *sql.Tx) error {
        // Marshal tags and content to JSON
        tagsJSON, err := json.Marshal(entity.Tags)
        if err != nil {
            return err
        }

        contentJSON, err := json.Marshal(entity.Content)
        if err != nil {
            return err
        }

        // Insert entity into database
        _, err = tx.Exec(
            "INSERT INTO entities (id, tags, content) VALUES (?, ?, ?)",
            entity.ID,
            string(tagsJSON),
            string(contentJSON),
        )
        if err != nil {
            return err
        }

        return nil
    })
}
```

This ensures that the entity creation either succeeds completely or fails completely, with no partial updates.

## Transaction Isolation

SQLite provides different transaction modes:

- **DEFERRED** (default): Locks are acquired when needed
- **IMMEDIATE**: Write lock acquired at the start of the transaction
- **EXCLUSIVE**: Both read and write locks acquired at the start

The current implementation uses the default DEFERRED mode, which is suitable for most use cases.

## Conclusion

The transaction support in EntityDB's SQLite layer provides a robust foundation for data integrity and consistency. By following the patterns and best practices outlined in this document, developers can ensure that database operations are atomic, consistent, isolated, and durable (ACID compliant).