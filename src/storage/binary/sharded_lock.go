package binary

import (
	"hash/fnv"
	"sync"
)

// ShardedLock provides a sharded locking mechanism to reduce contention
type ShardedLock struct {
	shards    []sync.RWMutex
	shardMask uint32
}

// NewShardedLock creates a new sharded lock with the specified number of shards
// shardCount must be a power of 2 for efficient masking
func NewShardedLock(shardCount int) *ShardedLock {
	// Ensure shardCount is power of 2
	if shardCount <= 0 || (shardCount&(shardCount-1)) != 0 {
		shardCount = 16 // Default to 16 shards
	}
	
	return &ShardedLock{
		shards:    make([]sync.RWMutex, shardCount),
		shardMask: uint32(shardCount - 1),
	}
}

// getShard returns the shard index for a given key
func (sl *ShardedLock) getShard(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32() & sl.shardMask
}

// Lock locks the shard for the given key
func (sl *ShardedLock) Lock(key string) {
	shard := sl.getShard(key)
	sl.shards[shard].Lock()
}

// Unlock unlocks the shard for the given key
func (sl *ShardedLock) Unlock(key string) {
	shard := sl.getShard(key)
	sl.shards[shard].Unlock()
}

// RLock read-locks the shard for the given key
func (sl *ShardedLock) RLock(key string) {
	shard := sl.getShard(key)
	sl.shards[shard].RLock()
}

// RUnlock read-unlocks the shard for the given key
func (sl *ShardedLock) RUnlock(key string) {
	shard := sl.getShard(key)
	sl.shards[shard].RUnlock()
}

// WithLock executes a function while holding the write lock for the key
func (sl *ShardedLock) WithLock(key string, fn func()) {
	sl.Lock(key)
	defer sl.Unlock(key)
	fn()
}

// WithRLock executes a function while holding the read lock for the key
func (sl *ShardedLock) WithRLock(key string, fn func()) {
	sl.RLock(key)
	defer sl.RUnlock(key)
	fn()
}

// ShardedMap provides a thread-safe map with sharded locking
type ShardedMap struct {
	locks  *ShardedLock
	shards []map[string]interface{}
}

// NewShardedMap creates a new sharded map
func NewShardedMap(shardCount int) *ShardedMap {
	if shardCount <= 0 || (shardCount&(shardCount-1)) != 0 {
		shardCount = 16
	}
	
	shards := make([]map[string]interface{}, shardCount)
	for i := range shards {
		shards[i] = make(map[string]interface{})
	}
	
	return &ShardedMap{
		locks:  NewShardedLock(shardCount),
		shards: shards,
	}
}

// Get retrieves a value from the map
func (sm *ShardedMap) Get(key string) (interface{}, bool) {
	sm.locks.RLock(key)
	defer sm.locks.RUnlock(key)
	
	shard := sm.locks.getShard(key)
	val, ok := sm.shards[shard][key]
	return val, ok
}

// Set stores a value in the map
func (sm *ShardedMap) Set(key string, value interface{}) {
	sm.locks.Lock(key)
	defer sm.locks.Unlock(key)
	
	shard := sm.locks.getShard(key)
	sm.shards[shard][key] = value
}

// Delete removes a value from the map
func (sm *ShardedMap) Delete(key string) {
	sm.locks.Lock(key)
	defer sm.locks.Unlock(key)
	
	shard := sm.locks.getShard(key)
	delete(sm.shards[shard], key)
}

// Len returns the total number of items across all shards
func (sm *ShardedMap) Len() int {
	total := 0
	for i := range sm.shards {
		sm.locks.shards[i].RLock()
		total += len(sm.shards[i])
		sm.locks.shards[i].RUnlock()
	}
	return total
}