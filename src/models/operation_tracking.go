package models

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
	
	"entitydb/logger"
)

// OperationID represents a unique identifier for tracking operations
type OperationID string

// OperationType represents the type of operation being performed
type OperationType string

const (
	// Operation types
	OpTypeRead         OperationType = "READ"
	OpTypeWrite        OperationType = "WRITE"
	OpTypeUpdate       OperationType = "UPDATE"
	OpTypeDelete       OperationType = "DELETE"
	OpTypeIndex        OperationType = "INDEX"
	OpTypeWAL          OperationType = "WAL"
	OpTypeTransaction  OperationType = "TRANSACTION"
	OpTypeVerification OperationType = "VERIFICATION"
	OpTypeRecovery     OperationType = "RECOVERY"
)

// OperationContext holds operation tracking information
type OperationContext struct {
	ID        OperationID
	Type      OperationType
	EntityID  string
	StartTime time.Time
	EndTime   time.Time
	Status    string
	Error     error
	Metadata  map[string]interface{}
	mu        sync.RWMutex
}

// Global operation tracker
var operationTracker = &OperationTracker{
	operations: make(map[OperationID]*OperationContext),
}

// OperationTracker manages all operations
type OperationTracker struct {
	operations map[OperationID]*OperationContext
	mu         sync.RWMutex
}

// GenerateOperationID creates a new unique operation ID
func GenerateOperationID() OperationID {
	// Generate 8 random bytes
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp if random fails
		return OperationID(fmt.Sprintf("op_%d", time.Now().UnixNano()))
	}
	
	// Format as hex string with prefix
	return OperationID(fmt.Sprintf("op_%s_%d", hex.EncodeToString(b), time.Now().UnixNano()))
}

// StartOperation begins tracking a new operation
func StartOperation(opType OperationType, entityID string, metadata map[string]interface{}) *OperationContext {
	op := &OperationContext{
		ID:        GenerateOperationID(),
		Type:      opType,
		EntityID:  entityID,
		StartTime: time.Now(),
		Status:    "started",
		Metadata:  metadata,
	}
	
	if op.Metadata == nil {
		op.Metadata = make(map[string]interface{})
	}
	
	// Store in tracker
	operationTracker.mu.Lock()
	operationTracker.operations[op.ID] = op
	operationTracker.mu.Unlock()
	
	// Log operation start
	logger.Debug("Started %s operation %s for entity %s", opType, op.ID, entityID)
	logger.Trace("%s metadata: %+v", op.ID, op.Metadata)
	
	return op
}

// CompleteOperation marks an operation as completed
func (op *OperationContext) Complete() {
	op.mu.Lock()
	defer op.mu.Unlock()
	
	op.EndTime = time.Now()
	op.Status = "completed"
	duration := op.EndTime.Sub(op.StartTime)
	
	logger.Debug("Completed %s operation %s for entity %s (duration: %v)", 
		op.Type, op.ID, op.EntityID, duration)
}

// FailOperation marks an operation as failed
func (op *OperationContext) Fail(err error) {
	op.mu.Lock()
	defer op.mu.Unlock()
	
	op.EndTime = time.Now()
	op.Status = "failed"
	op.Error = err
	duration := op.EndTime.Sub(op.StartTime)
	
	logger.Error("Failed %s operation %s for entity %s (duration: %v): %v", 
		op.Type, op.ID, op.EntityID, duration, err)
}

// SetMetadata adds or updates metadata for an operation
func (op *OperationContext) SetMetadata(key string, value interface{}) {
	op.mu.Lock()
	defer op.mu.Unlock()
	
	if op.Metadata == nil {
		op.Metadata = make(map[string]interface{})
	}
	op.Metadata[key] = value
	
	logger.Trace("%s updated metadata %s = %v", op.ID, key, value)
}

// GetMetadata retrieves metadata value
func (op *OperationContext) GetMetadata(key string) (interface{}, bool) {
	op.mu.RLock()
	defer op.mu.RUnlock()
	
	val, ok := op.Metadata[key]
	return val, ok
}

// Log adds a log entry for this operation
func (op *OperationContext) Log(level, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fullMessage := fmt.Sprintf("Operation %s: %s", op.ID, message)
	
	switch level {
	case "trace":
		logger.Trace(fullMessage)
	case "debug":
		logger.Debug(fullMessage)
	case "info":
		logger.Info(fullMessage)
	case "warn":
		logger.Warn(fullMessage)
	case "error":
		logger.Error(fullMessage)
	default:
		logger.Info(fullMessage)
	}
}

// GetOperation retrieves an operation by ID
func GetOperation(id OperationID) (*OperationContext, bool) {
	operationTracker.mu.RLock()
	defer operationTracker.mu.RUnlock()
	
	op, ok := operationTracker.operations[id]
	return op, ok
}

// CleanupOldOperations removes completed operations older than specified duration
func CleanupOldOperations(maxAge time.Duration) int {
	operationTracker.mu.Lock()
	defer operationTracker.mu.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	removed := 0
	
	for id, op := range operationTracker.operations {
		if op.Status != "started" && op.EndTime.Before(cutoff) {
			delete(operationTracker.operations, id)
			removed++
		}
	}
	
	if removed > 0 {
		logger.Debug("Cleaned up %d old operations", removed)
	}
	
	return removed
}

// GetActiveOperations returns all currently active operations
func GetActiveOperations() []*OperationContext {
	operationTracker.mu.RLock()
	defer operationTracker.mu.RUnlock()
	
	var active []*OperationContext
	for _, op := range operationTracker.operations {
		if op.Status == "started" {
			active = append(active, op)
		}
	}
	
	return active
}

// Context keys for operation tracking
type contextKey string

const operationContextKey contextKey = "operation"

// WithOperation adds an operation to a context
func WithOperation(ctx context.Context, op *OperationContext) context.Context {
	return context.WithValue(ctx, operationContextKey, op)
}

// OperationFromContext retrieves an operation from context
func OperationFromContext(ctx context.Context) (*OperationContext, bool) {
	op, ok := ctx.Value(operationContextKey).(*OperationContext)
	return op, ok
}

// OperationSummary provides a summary of operations
type OperationSummary struct {
	Total      int
	Active     int
	Completed  int
	Failed     int
	ByType     map[OperationType]int
	RecentOps  []*OperationContext
}

// GetOperationSummary returns a summary of all operations
func GetOperationSummary() *OperationSummary {
	operationTracker.mu.RLock()
	defer operationTracker.mu.RUnlock()
	
	summary := &OperationSummary{
		Total:  len(operationTracker.operations),
		ByType: make(map[OperationType]int),
	}
	
	// Collect recent operations (last 10)
	recent := make([]*OperationContext, 0, 10)
	
	for _, op := range operationTracker.operations {
		// Count by status
		switch op.Status {
		case "started":
			summary.Active++
		case "completed":
			summary.Completed++
		case "failed":
			summary.Failed++
		}
		
		// Count by type
		summary.ByType[op.Type]++
		
		// Add to recent if within last 10
		if len(recent) < 10 {
			recent = append(recent, op)
		}
	}
	
	summary.RecentOps = recent
	return summary
}

// OperationStats holds aggregated operation statistics
type OperationStats struct {
	TotalOperations      int64
	SuccessfulOperations int64
	FailedOperations     int64
	ActiveOperations     int
	ByType               map[OperationType]int64
}

// RecoveryStats holds recovery operation statistics
type RecoveryStats struct {
	TotalAttempts    int
	Successful       int
	Failed           int
	LastRecoveryTime time.Time
}

// Global stats trackers
var (
	globalOpStats = &OperationStats{
		ByType: make(map[OperationType]int64),
	}
	globalRecoveryStats = &RecoveryStats{}
	statsMu sync.RWMutex
)

// GetOperationStats returns global operation statistics
func GetOperationStats() OperationStats {
	statsMu.RLock()
	defer statsMu.RUnlock()
	
	// Calculate from current operations
	operationTracker.mu.RLock()
	defer operationTracker.mu.RUnlock()
	
	stats := OperationStats{
		ByType: make(map[OperationType]int64),
	}
	
	for _, op := range operationTracker.operations {
		stats.TotalOperations++
		stats.ByType[op.Type]++
		
		switch op.Status {
		case "started":
			stats.ActiveOperations++
		case "completed":
			stats.SuccessfulOperations++
		case "failed":
			stats.FailedOperations++
		}
	}
	
	return stats
}

// GetRecoveryStats returns global recovery statistics
func GetRecoveryStats() RecoveryStats {
	statsMu.RLock()
	defer statsMu.RUnlock()
	
	// Count recovery operations
	operationTracker.mu.RLock()
	defer operationTracker.mu.RUnlock()
	
	stats := RecoveryStats{}
	
	for _, op := range operationTracker.operations {
		if op.Type == OpTypeRecovery {
			stats.TotalAttempts++
			if op.Status == "completed" {
				stats.Successful++
				if op.EndTime.After(stats.LastRecoveryTime) {
					stats.LastRecoveryTime = op.EndTime
				}
			} else if op.Status == "failed" {
				stats.Failed++
			}
		}
	}
	
	return stats
}
