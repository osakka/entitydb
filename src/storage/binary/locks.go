package binary

import (
	"sync"
	"time"
)

// LockType defines the type of lock
type LockType int

const (
	ReadLock LockType = iota
	WriteLock
)

// LockManager handles granular locks for entities and files
type LockManager struct {
	// File-level lock for structural changes
	fileLock sync.RWMutex
	
	// Entity-level locks for fine-grained access
	entityLocks map[string]*sync.RWMutex
	entityMu    sync.Mutex // Protects the entityLocks map
	
	// Tag index locks for efficient tag operations
	tagLocks map[string]*sync.RWMutex
	tagMu    sync.Mutex // Protects the tagLocks map
	
	// Operation locks for specific operations
	writeLock   sync.Mutex // Serializes write operations
	compactLock sync.Mutex // Prevents compaction during writes
	
	// Statistics
	stats LockStats
}

// LockStats tracks locking statistics
type LockStats struct {
	mu            sync.Mutex
	ReadLocks     int64
	WriteLocks    int64
	WaitTime      time.Duration
	HeldTime      time.Duration
	Contentions   int64
}

// NewLockManager creates a new lock manager
func NewLockManager() *LockManager {
	return &LockManager{
		entityLocks: make(map[string]*sync.RWMutex),
		tagLocks:    make(map[string]*sync.RWMutex),
	}
}

// AcquireFileLock acquires a file-level lock
func (lm *LockManager) AcquireFileLock(lockType LockType) {
	start := time.Now()
	
	switch lockType {
	case ReadLock:
		lm.fileLock.RLock()
		lm.stats.ReadLocks++
	case WriteLock:
		lm.fileLock.Lock()
		lm.stats.WriteLocks++
	}
	
	lm.stats.mu.Lock()
	lm.stats.WaitTime += time.Since(start)
	lm.stats.mu.Unlock()
}

// ReleaseFileLock releases a file-level lock
func (lm *LockManager) ReleaseFileLock(lockType LockType) {
	switch lockType {
	case ReadLock:
		lm.fileLock.RUnlock()
	case WriteLock:
		lm.fileLock.Unlock()
	}
}

// AcquireEntityLock acquires a lock for a specific entity
func (lm *LockManager) AcquireEntityLock(entityID string, lockType LockType) {
	lm.entityMu.Lock()
	lock, exists := lm.entityLocks[entityID]
	if !exists {
		lock = &sync.RWMutex{}
		lm.entityLocks[entityID] = lock
	}
	lm.entityMu.Unlock()
	
	start := time.Now()
	
	switch lockType {
	case ReadLock:
		lock.RLock()
		lm.stats.ReadLocks++
	case WriteLock:
		lock.Lock()
		lm.stats.WriteLocks++
	}
	
	lm.stats.mu.Lock()
	lm.stats.WaitTime += time.Since(start)
	lm.stats.mu.Unlock()
}

// ReleaseEntityLock releases a lock for a specific entity
func (lm *LockManager) ReleaseEntityLock(entityID string, lockType LockType) {
	lm.entityMu.Lock()
	lock, exists := lm.entityLocks[entityID]
	lm.entityMu.Unlock()
	
	if !exists {
		return
	}
	
	switch lockType {
	case ReadLock:
		lock.RUnlock()
	case WriteLock:
		lock.Unlock()
	}
}

// AcquireTagLock acquires a lock for a specific tag
func (lm *LockManager) AcquireTagLock(tag string, lockType LockType) {
	lm.tagMu.Lock()
	lock, exists := lm.tagLocks[tag]
	if !exists {
		lock = &sync.RWMutex{}
		lm.tagLocks[tag] = lock
	}
	lm.tagMu.Unlock()
	
	start := time.Now()
	
	switch lockType {
	case ReadLock:
		lock.RLock()
		lm.stats.ReadLocks++
	case WriteLock:
		lock.Lock()
		lm.stats.WriteLocks++
	}
	
	lm.stats.mu.Lock()
	lm.stats.WaitTime += time.Since(start)
	lm.stats.mu.Unlock()
}

// ReleaseTagLock releases a lock for a specific tag
func (lm *LockManager) ReleaseTagLock(tag string, lockType LockType) {
	lm.tagMu.Lock()
	lock, exists := lm.tagLocks[tag]
	lm.tagMu.Unlock()
	
	if !exists {
		return
	}
	
	switch lockType {
	case ReadLock:
		lock.RUnlock()
	case WriteLock:
		lock.Unlock()
	}
}

// AcquireWriteLock serializes write operations
func (lm *LockManager) AcquireWriteLock() {
	lm.writeLock.Lock()
}

// ReleaseWriteLock releases the write serialization lock
func (lm *LockManager) ReleaseWriteLock() {
	lm.writeLock.Unlock()
}

// GetStats returns locking statistics
func (lm *LockManager) GetStats() LockStats {
	lm.stats.mu.Lock()
	defer lm.stats.mu.Unlock()
	return lm.stats
}

// CleanupOldLocks removes locks for entities that haven't been accessed recently
func (lm *LockManager) CleanupOldLocks() {
	// This would be called periodically to prevent memory leaks
	// For now, we'll keep all locks in memory
}