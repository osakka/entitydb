package binary

import (
	"strings"
	"sync/atomic"
)

// MetricsGuard prevents metrics collection feedback loops by detecting
// and blocking recursive metrics operations
type MetricsGuard struct {
	// Use atomic for thread-safe operation counting
	activeOperations int64
}

var globalMetricsGuard = &MetricsGuard{}

// WithMetricsGuard executes a function while preventing metrics collection
// This is used to break feedback loops where metrics create entities
// which trigger more metrics
func WithMetricsGuard(fn func() error) error {
	// Increment active operations
	atomic.AddInt64(&globalMetricsGuard.activeOperations, 1)
	defer atomic.AddInt64(&globalMetricsGuard.activeOperations, -1)
	
	// Execute the function
	return fn()
}

// IsMetricsEntity checks if an entity ID indicates it's a metrics entity
func IsMetricsEntity(entityID string) bool {
	return strings.HasPrefix(entityID, "metric_") || 
	       strings.Contains(entityID, "_metric_") ||
	       strings.Contains(entityID, "_metrics_")
}

// ShouldSkipMetrics returns true if metrics collection should be skipped
// to prevent feedback loops
func ShouldSkipMetrics(entityID string) bool {
	// Skip if we're already in a metrics operation
	if atomic.LoadInt64(&globalMetricsGuard.activeOperations) > 0 {
		return true
	}
	
	// Skip if this is a metrics entity
	if IsMetricsEntity(entityID) {
		return true
	}
	
	// Skip if global metrics operations are in progress
	if IsMetricsOperation() {
		return true
	}
	
	return false
}