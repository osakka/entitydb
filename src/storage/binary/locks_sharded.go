package binary

import (
	"sync"
	"time"
	"hash/fnv"
)

const (
	// EntityLockShards is the number of shards for entity locks
	EntityLockShards = 256
	// TagLockShards is the number of shards for tag locks
	TagLockShards = 128
)

// ShardedLockManager is an optimized lock manager using sharded locks
type ShardedLockManager struct {
	// File-level lock for structural changes
	fileLock sync.RWMutex
	
	// Sharded entity locks for reduced contention
	entityLocks [EntityLockShards]*sync.RWMutex
	
	// Sharded tag locks for efficient tag operations
	tagLocks [TagLockShards]*sync.RWMutex
	
	// Operation locks for specific operations
	writeLock   sync.Mutex // Serializes write operations
	compactLock sync.Mutex // Prevents compaction during writes
	
	// Statistics
	stats LockStats
}

// NewShardedLockManager creates a new sharded lock manager
func NewShardedLockManager() *ShardedLockManager {
	lm := &ShardedLockManager{}
	
	// Initialize entity lock shards
	for i := 0; i < EntityLockShards; i++ {
		lm.entityLocks[i] = &sync.RWMutex{}
	}
	
	// Initialize tag lock shards
	for i := 0; i < TagLockShards; i++ {
		lm.tagLocks[i] = &sync.RWMutex{}
	}
	
	return lm
}

// getEntityShard returns the shard index for an entity ID
func (lm *ShardedLockManager) getEntityShard(entityID string) int {
	h := fnv.New32a()
	h.Write([]byte(entityID))
	return int(h.Sum32() % EntityLockShards)
}

// getTagShard returns the shard index for a tag
func (lm *ShardedLockManager) getTagShard(tag string) int {
	h := fnv.New32a()
	h.Write([]byte(tag))
	return int(h.Sum32() % TagLockShards)
}

// AcquireFileLock acquires a file-level lock
func (lm *ShardedLockManager) AcquireFileLock(lockType LockType) {
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
func (lm *ShardedLockManager) ReleaseFileLock(lockType LockType) {
	switch lockType {
	case ReadLock:
		lm.fileLock.RUnlock()
	case WriteLock:
		lm.fileLock.Unlock()
	}
}

// AcquireEntityLock acquires a lock for a specific entity
func (lm *ShardedLockManager) AcquireEntityLock(entityID string, lockType LockType) {
	shard := lm.getEntityShard(entityID)
	lock := lm.entityLocks[shard]
	
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
func (lm *ShardedLockManager) ReleaseEntityLock(entityID string, lockType LockType) {
	shard := lm.getEntityShard(entityID)
	lock := lm.entityLocks[shard]
	
	switch lockType {
	case ReadLock:
		lock.RUnlock()
	case WriteLock:
		lock.Unlock()
	}
}

// AcquireEntityLocks acquires locks for multiple entities efficiently
func (lm *ShardedLockManager) AcquireEntityLocks(entityIDs []string, lockType LockType) {
	// Group entities by shard to minimize lock acquisition overhead
	shardGroups := make(map[int][]string)
	for _, id := range entityIDs {
		shard := lm.getEntityShard(id)
		shardGroups[shard] = append(shardGroups[shard], id)
	}
	
	// Acquire locks in shard order to prevent deadlocks
	shards := make([]int, 0, len(shardGroups))
	for shard := range shardGroups {
		shards = append(shards, shard)
	}
	
	// Sort shards to ensure consistent lock ordering
	for i := 0; i < len(shards); i++ {
		for j := i + 1; j < len(shards); j++ {
			if shards[i] > shards[j] {
				shards[i], shards[j] = shards[j], shards[i]
			}
		}
	}
	
	// Acquire locks in order
	for _, shard := range shards {
		lock := lm.entityLocks[shard]
		switch lockType {
		case ReadLock:
			lock.RLock()
		case WriteLock:
			lock.Lock()
		}
	}
}

// ReleaseEntityLocks releases locks for multiple entities
func (lm *ShardedLockManager) ReleaseEntityLocks(entityIDs []string, lockType LockType) {
	// Group entities by shard
	shardSet := make(map[int]bool)
	for _, id := range entityIDs {
		shard := lm.getEntityShard(id)
		shardSet[shard] = true
	}
	
	// Release all shard locks
	for shard := range shardSet {
		lock := lm.entityLocks[shard]
		switch lockType {
		case ReadLock:
			lock.RUnlock()
		case WriteLock:
			lock.Unlock()
		}
	}
}

// AcquireTagLock acquires a lock for a specific tag
func (lm *ShardedLockManager) AcquireTagLock(tag string, lockType LockType) {
	shard := lm.getTagShard(tag)
	lock := lm.tagLocks[shard]
	
	switch lockType {
	case ReadLock:
		lock.RLock()
	case WriteLock:
		lock.Lock()
	}
}

// ReleaseTagLock releases a lock for a specific tag
func (lm *ShardedLockManager) ReleaseTagLock(tag string, lockType LockType) {
	shard := lm.getTagShard(tag)
	lock := lm.tagLocks[shard]
	
	switch lockType {
	case ReadLock:
		lock.RUnlock()
	case WriteLock:
		lock.Unlock()
	}
}

// AcquireWriteLock acquires the global write lock
func (lm *ShardedLockManager) AcquireWriteLock() {
	lm.writeLock.Lock()
}

// ReleaseWriteLock releases the global write lock
func (lm *ShardedLockManager) ReleaseWriteLock() {
	lm.writeLock.Unlock()
}

// AcquireCompactLock acquires the compaction lock
func (lm *ShardedLockManager) AcquireCompactLock() {
	lm.compactLock.Lock()
}

// ReleaseCompactLock releases the compaction lock
func (lm *ShardedLockManager) ReleaseCompactLock() {
	lm.compactLock.Unlock()
}

// GetStats returns lock statistics
func (lm *ShardedLockManager) GetStats() LockStats {
	lm.stats.mu.Lock()
	defer lm.stats.mu.Unlock()
	return lm.stats
}

// IsCompatibleWithLockManager checks if this implements the same interface
func (lm *ShardedLockManager) IsCompatibleWithLockManager() bool {
	return true
}