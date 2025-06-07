package binary

import (
	"sync"
	"entitydb/logger"
)

// TracedRWMutex wraps sync.RWMutex with tracing
type TracedRWMutex struct {
	mu   sync.RWMutex
	name string
}

// NewTracedRWMutex creates a new traced mutex
func NewTracedRWMutex(name string) *TracedRWMutex {
	return &TracedRWMutex{name: name}
}

// Lock acquires write lock with tracing
func (t *TracedRWMutex) Lock(traceID string) {
	logger.LogLockOperation(traceID, "RWMutex", t.name, "lock_acquire")
	t.mu.Lock()
	logger.LogLockOperation(traceID, "RWMutex", t.name, "lock_acquired")
}

// Unlock releases write lock with tracing
func (t *TracedRWMutex) Unlock(traceID string) {
	logger.LogLockOperation(traceID, "RWMutex", t.name, "unlock")
	t.mu.Unlock()
}

// RLock acquires read lock with tracing
func (t *TracedRWMutex) RLock(traceID string) {
	logger.LogLockOperation(traceID, "RWMutex", t.name, "rlock_acquire")
	t.mu.RLock()
	logger.LogLockOperation(traceID, "RWMutex", t.name, "rlock_acquired")
}

// RUnlock releases read lock with tracing
func (t *TracedRWMutex) RUnlock(traceID string) {
	logger.LogLockOperation(traceID, "RWMutex", t.name, "runlock")
	t.mu.RUnlock()
}

// TracedMutex wraps sync.Mutex with tracing
type TracedMutex struct {
	mu   sync.Mutex
	name string
}

// NewTracedMutex creates a new traced mutex
func NewTracedMutex(name string) *TracedMutex {
	return &TracedMutex{name: name}
}

// Lock acquires lock with tracing
func (t *TracedMutex) Lock(traceID string) {
	logger.LogLockOperation(traceID, "Mutex", t.name, "lock_acquire")
	t.mu.Lock()
	logger.LogLockOperation(traceID, "Mutex", t.name, "lock_acquired")
}

// Unlock releases lock with tracing
func (t *TracedMutex) Unlock(traceID string) {
	logger.LogLockOperation(traceID, "Mutex", t.name, "unlock")
	t.mu.Unlock()
}