package binary

import (
	"entitydb/logger"
	"sync"
	"time"
)

// UpdateCircuitBreaker prevents cascading failures from high-frequency entity updates
// Implements rate limiting and failure tracking to maintain system stability
type UpdateCircuitBreaker struct {
	// Rate limiting
	lastUpdate    map[string]time.Time
	updateCounter map[string]int
	
	// Failure tracking  
	failureCount  map[string]int
	lastFailure   map[string]time.Time
	
	// Circuit state
	circuitOpen   map[string]bool
	
	mu sync.RWMutex
	
	// Configuration
	maxUpdatesPerSecond int           // Max updates per entity per second
	maxFailures         int           // Max failures before opening circuit
	circuitTimeout      time.Duration // How long to keep circuit open
	resetWindow         time.Duration // Time window for rate limiting
}

// NewUpdateCircuitBreaker creates a circuit breaker for entity updates
func NewUpdateCircuitBreaker() *UpdateCircuitBreaker {
	return &UpdateCircuitBreaker{
		lastUpdate:          make(map[string]time.Time),
		updateCounter:       make(map[string]int),
		failureCount:        make(map[string]int),
		lastFailure:         make(map[string]time.Time),
		circuitOpen:         make(map[string]bool),
		maxUpdatesPerSecond: 10,                // 10 updates/second max per entity
		maxFailures:         5,                 // 5 failures before circuit opens
		circuitTimeout:      30 * time.Second, // 30 second circuit timeout
		resetWindow:         1 * time.Second,  // 1 second rate limit window
	}
}

// CanUpdate checks if an entity update is allowed
func (cb *UpdateCircuitBreaker) CanUpdate(entityID string) (bool, string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	now := time.Now()
	
	// Check if circuit is open for this entity
	if cb.circuitOpen[entityID] {
		if now.Sub(cb.lastFailure[entityID]) > cb.circuitTimeout {
			// Reset circuit after timeout
			logger.Debug("Circuit breaker reset for entity %s", entityID)
			cb.circuitOpen[entityID] = false
			cb.failureCount[entityID] = 0
		} else {
			return false, "circuit breaker open - too many recent failures"
		}
	}
	
	// Rate limiting check
	lastUpdate := cb.lastUpdate[entityID]
	if !lastUpdate.IsZero() {
		timeSinceUpdate := now.Sub(lastUpdate)
		
		// Reset counter if outside window
		if timeSinceUpdate > cb.resetWindow {
			cb.updateCounter[entityID] = 0
		}
		
		// Check rate limit
		if timeSinceUpdate < cb.resetWindow {
			if cb.updateCounter[entityID] >= cb.maxUpdatesPerSecond {
				return false, "rate limit exceeded - max 10 updates/second per entity"
			}
		}
	}
	
	// Update tracking
	cb.lastUpdate[entityID] = now
	cb.updateCounter[entityID]++
	
	return true, ""
}

// RecordSuccess records a successful update
func (cb *UpdateCircuitBreaker) RecordSuccess(entityID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	// Reset failure count on success
	if cb.failureCount[entityID] > 0 {
		cb.failureCount[entityID] = 0
		logger.Debug("Reset failure count for entity %s after successful update", entityID)
	}
}

// RecordFailure records a failed update and may open the circuit
func (cb *UpdateCircuitBreaker) RecordFailure(entityID string, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	now := time.Now()
	cb.failureCount[entityID]++
	cb.lastFailure[entityID] = now
	
	logger.Debug("Update failure for entity %s: %v (failure count: %d)", 
		entityID, err, cb.failureCount[entityID])
	
	// Open circuit if too many failures
	if cb.failureCount[entityID] >= cb.maxFailures {
		cb.circuitOpen[entityID] = true
		logger.Warn("Circuit breaker opened for entity %s after %d failures", 
			entityID, cb.failureCount[entityID])
	}
}

// GetStats returns circuit breaker statistics
func (cb *UpdateCircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	openCircuits := 0
	totalFailures := 0
	activeEntities := len(cb.lastUpdate)
	
	for _, isOpen := range cb.circuitOpen {
		if isOpen {
			openCircuits++
		}
	}
	
	for _, failures := range cb.failureCount {
		totalFailures += failures
	}
	
	return map[string]interface{}{
		"active_entities":      activeEntities,
		"open_circuits":        openCircuits,
		"total_failures":       totalFailures,
		"max_updates_per_sec":  cb.maxUpdatesPerSecond,
		"circuit_timeout_sec":  cb.circuitTimeout.Seconds(),
	}
}

// Cleanup removes old entries to prevent memory growth
func (cb *UpdateCircuitBreaker) Cleanup() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	now := time.Now()
	cleanupThreshold := 5 * time.Minute // Remove entries older than 5 minutes
	
	for entityID, lastUpdate := range cb.lastUpdate {
		if now.Sub(lastUpdate) > cleanupThreshold {
			delete(cb.lastUpdate, entityID)
			delete(cb.updateCounter, entityID)
			delete(cb.failureCount, entityID)
			delete(cb.lastFailure, entityID)
			delete(cb.circuitOpen, entityID)
		}
	}
}

// SetConfiguration allows runtime configuration updates
func (cb *UpdateCircuitBreaker) SetConfiguration(maxUpdatesPerSecond int, maxFailures int, circuitTimeoutSeconds int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.maxUpdatesPerSecond = maxUpdatesPerSecond
	cb.maxFailures = maxFailures
	cb.circuitTimeout = time.Duration(circuitTimeoutSeconds) * time.Second
	
	logger.Info("Circuit breaker configuration updated: %d ops/sec, %d max failures, %d sec timeout",
		maxUpdatesPerSecond, maxFailures, circuitTimeoutSeconds)
}