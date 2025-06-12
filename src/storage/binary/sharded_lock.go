// Package binary provides high-performance concurrent data structures for the
// EntityDB storage layer. The sharded lock implementation distributes lock
// contention across multiple shards to improve scalability under high concurrency.
package binary

import (
	"hash/fnv"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	// NumShards is the number of lock shards (must be power of 2).
	// Higher values reduce contention but increase memory usage.
	// 256 shards provides good balance for most workloads.
	NumShards = 256
	
	// MaxReadersBatch controls reader/writer fairness.
	// After this many readers, a waiting writer gets priority.
	// Prevents writer starvation under heavy read loads.
	MaxReadersBatch = 10
	
	// NumLockShards for general-purpose ShardedLock.
	// Smaller than tag index shards as it's used for coarser locking.
	NumLockShards = 64
)

// ShardedTagIndex implements a high-performance tag index using sharding
// to distribute lock contention. Tags are distributed across shards using
// consistent hashing, allowing concurrent operations on different tags.
//
// Performance characteristics:
//   - O(1) tag lookups within a shard
//   - Concurrent operations on different shards
//   - Fair reader/writer scheduling prevents starvation
//   - Memory usage: O(unique_tags * avg_entities_per_tag)
//
// Example usage:
//   index := NewShardedTagIndex()
//   index.AddTag("type:user", "user-123")
//   entities := index.GetEntitiesForTag("type:user")
type ShardedTagIndex struct {
	shards [NumShards]*TagIndexShard
}

// TagIndexShard represents a single shard of the tag index.
// Each shard maintains its own lock and tag mappings, allowing
// concurrent access to different shards.
type TagIndexShard struct {
	mu       sync.RWMutex        // Protects tag map access
	tags     map[string][]string // tag -> entity IDs mapping
	queue    *FairQueue          // Ensures fair reader/writer access
}

// FairQueue implements fair scheduling between readers and writers
// to prevent starvation. It ensures writers eventually get access
// even under heavy read load.
//
// Algorithm:
//   1. Readers can proceed if no writer is waiting
//   2. After MaxReadersBatch reads, writers get priority
//   3. Writers process one at a time
//   4. Queued operations maintain FIFO order
type FairQueue struct {
	mu            sync.Mutex      // Protects queue state
	readerQueue   []chan struct{} // Blocked readers
	writerQueue   []chan struct{} // Blocked writers
	activeReaders int32           // Current active reader count
	readerCount   int32           // Reads since last write
	writerWaiting bool            // Writer is waiting for access
}

// NewShardedTagIndex creates a new sharded tag index
func NewShardedTagIndex() *ShardedTagIndex {
	index := &ShardedTagIndex{}
	for i := 0; i < NumShards; i++ {
		index.shards[i] = &TagIndexShard{
			tags:  make(map[string][]string),
			queue: NewFairQueue(),
		}
	}
	return index
}

// NewFairQueue creates a new fair queue
func NewFairQueue() *FairQueue {
	return &FairQueue{
		readerQueue: make([]chan struct{}, 0),
		writerQueue: make([]chan struct{}, 0),
	}
}

// getShard determines which shard owns a given tag using consistent hashing.
// The FNV-1a hash provides good distribution across shards with low collision rate.
//
// Shard selection algorithm:
//   1. Hash the tag string using FNV-1a (fast, non-cryptographic)
//   2. Use bitwise AND with (NumShards-1) for modulo operation
//   3. This works because NumShards is power of 2
//
// Performance: O(len(tag)) for hash calculation
func (s *ShardedTagIndex) getShard(tag string) *TagIndexShard {
	h := fnv.New32a()
	h.Write([]byte(tag))
	shardIdx := h.Sum32() & (NumShards - 1)
	return s.shards[shardIdx]
}

// AddTag associates an entity ID with a tag in the index.
// If the entity is already associated with the tag, this is a no-op.
//
// Concurrency behavior:
//   - Acquires write lock on the tag's shard only
//   - Other shards remain accessible for concurrent operations
//   - Uses fair queuing to prevent writer starvation
//
// Parameters:
//   - tag: The tag to index (e.g., "type:user")
//   - entityID: The entity to associate with the tag
//
// Thread Safety:
//   Safe for concurrent use. Multiple AddTag calls for different
//   shards can proceed in parallel.
func (s *ShardedTagIndex) AddTag(tag string, entityID string) {
	shard := s.getShard(tag)
	
	// Use fair queue for write access
	shard.queue.AcquireWrite()
	defer shard.queue.ReleaseWrite()
	
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	if shard.tags[tag] == nil {
		shard.tags[tag] = make([]string, 0, 1)
	}
	
	// Check if entity already exists for this tag
	for _, id := range shard.tags[tag] {
		if id == entityID {
			return
		}
	}
	
	shard.tags[tag] = append(shard.tags[tag], entityID)
}

// GetEntitiesForTag returns all entity IDs for a given tag
func (s *ShardedTagIndex) GetEntitiesForTag(tag string) []string {
	shard := s.getShard(tag)
	
	// Use fair queue for read access
	shard.queue.AcquireRead()
	defer shard.queue.ReleaseRead()
	
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	
	entities := shard.tags[tag]
	if entities == nil {
		return []string{}
	}
	
	// Return a copy to avoid race conditions
	result := make([]string, len(entities))
	copy(result, entities)
	return result
}

// RemoveTag removes an entity from a tag in the sharded index
func (s *ShardedTagIndex) RemoveTag(tag string, entityID string) {
	shard := s.getShard(tag)
	
	// Use fair queue for write access
	shard.queue.AcquireWrite()
	defer shard.queue.ReleaseWrite()
	
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	entities := shard.tags[tag]
	if entities == nil {
		return // Tag doesn't exist
	}
	
	// Remove entityID from the slice
	newEntities := make([]string, 0, len(entities))
	for _, id := range entities {
		if id != entityID {
			newEntities = append(newEntities, id)
		}
	}
	
	if len(newEntities) > 0 {
		shard.tags[tag] = newEntities
	} else {
		// Remove the tag entirely if no entities left
		delete(shard.tags, tag)
	}
}

// GetAllTags returns all tags in the index (for backward compatibility)
// This should be avoided in hot paths as it locks all shards
func (s *ShardedTagIndex) GetAllTags() map[string][]string {
	result := make(map[string][]string)
	
	// Lock all shards for reading
	for _, shard := range s.shards {
		shard.queue.AcquireRead()
		shard.mu.RLock()
	}
	
	// Copy all data
	for _, shard := range s.shards {
		for tag, entities := range shard.tags {
			result[tag] = append([]string(nil), entities...)
		}
	}
	
	// Unlock all shards
	for _, shard := range s.shards {
		shard.mu.RUnlock()
		shard.queue.ReleaseRead()
	}
	
	return result
}

// AcquireRead acquires a read lock using fair scheduling to prevent writer starvation.
// Readers can proceed concurrently unless a writer is waiting and the reader
// batch limit has been reached.
//
// Fair scheduling algorithm:
//   1. Check if writer is waiting AND reader batch limit reached
//   2. If so, queue this reader and block until signaled
//   3. Otherwise, increment active reader count and proceed
//   4. Track total reads since last write for fairness
//
// This ensures writers eventually get access even under heavy read load,
// preventing writer starvation while still allowing read concurrency.
//
// Thread Safety:
//   Safe for concurrent use. Multiple readers can call simultaneously.
func (q *FairQueue) AcquireRead() {
	q.mu.Lock()
	
	// If a writer is waiting and we've had enough reads, wait
	if q.writerWaiting && atomic.LoadInt32(&q.readerCount) >= MaxReadersBatch {
		ch := make(chan struct{})
		q.readerQueue = append(q.readerQueue, ch)
		q.mu.Unlock()
		<-ch // Block until writer completes
		q.mu.Lock()
	}
	
	atomic.AddInt32(&q.activeReaders, 1)
	atomic.AddInt32(&q.readerCount, 1)
	q.mu.Unlock()
}

// ReleaseRead releases a read lock and potentially unblocks waiting writers.
// Must be called exactly once for each successful AcquireRead.
//
// Cleanup process:
//   1. Decrement active reader count
//   2. If no readers remain AND writer is queued, signal it
//   3. Reset reader count when writer proceeds (fairness reset)
//
// Thread Safety:
//   Safe for concurrent use. Matches with AcquireRead calls.
func (q *FairQueue) ReleaseRead() {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	atomic.AddInt32(&q.activeReaders, -1)
	
	// If no active readers and a writer is waiting, signal it
	if atomic.LoadInt32(&q.activeReaders) == 0 && len(q.writerQueue) > 0 {
		ch := q.writerQueue[0]
		q.writerQueue = q.writerQueue[1:]
		close(ch)
		atomic.StoreInt32(&q.readerCount, 0) // Reset reader count
	}
}

// AcquireWrite acquires a write lock with fair queuing
func (q *FairQueue) AcquireWrite() {
	q.mu.Lock()
	
	// If there are active readers or other writers, wait
	if atomic.LoadInt32(&q.activeReaders) > 0 || q.writerWaiting {
		ch := make(chan struct{})
		q.writerQueue = append(q.writerQueue, ch)
		q.writerWaiting = true
		q.mu.Unlock()
		<-ch
		return
	}
	
	q.writerWaiting = true
	q.mu.Unlock()
}

// ReleaseWrite releases a write lock
func (q *FairQueue) ReleaseWrite() {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	q.writerWaiting = false
	
	// Signal waiting readers if any
	for _, ch := range q.readerQueue {
		close(ch)
	}
	q.readerQueue = q.readerQueue[:0]
	
	// Or signal next writer
	if len(q.readerQueue) == 0 && len(q.writerQueue) > 0 {
		ch := q.writerQueue[0]
		q.writerQueue = q.writerQueue[1:]
		close(ch)
	}
}

// OptimizedListByTag performs an optimized tag lookup using sharding
func (s *ShardedTagIndex) OptimizedListByTag(tag string, fullScan bool) []string {
	if !fullScan {
		// Direct lookup for exact tag match
		return s.GetEntitiesForTag(tag)
	}
	
	// For pattern matching, we still need to scan, but we can parallelize
	results := make(chan []string, NumShards)
	var wg sync.WaitGroup
	
	// Search all shards in parallel
	for _, shard := range s.shards {
		wg.Add(1)
		go func(sh *TagIndexShard) {
			defer wg.Done()
			
			sh.queue.AcquireRead()
			sh.mu.RLock()
			
			localResults := []string{}
			for t, entities := range sh.tags {
				if t == tag || matchesPattern(t, tag) {
					localResults = append(localResults, entities...)
				}
			}
			
			sh.mu.RUnlock()
			sh.queue.ReleaseRead()
			
			if len(localResults) > 0 {
				results <- localResults
			}
		}(shard)
	}
	
	// Wait for all shards to complete
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect results
	allResults := []string{}
	seen := make(map[string]bool)
	
	for shardResults := range results {
		for _, entityID := range shardResults {
			if !seen[entityID] {
				seen[entityID] = true
				allResults = append(allResults, entityID)
			}
		}
	}
	
	return allResults
}

// matchesPattern checks if a tag matches a pattern (simple implementation)
func matchesPattern(tag, pattern string) bool {
	// Check if this is a temporal tag (has timestamp)
	if strings.Contains(tag, "|") {
		parts := strings.SplitN(tag, "|", 2)
		if len(parts) == 2 {
			// Compare the actual tag part with the pattern
			return parts[1] == pattern
		}
	}
	
	// For internal tags like _source:, _target:
	if len(pattern) > 0 && pattern[0] == '_' {
		return strings.HasPrefix(tag, pattern)
	}
	
	return false
}

// GetShardStats returns statistics about shard distribution
func (s *ShardedTagIndex) GetShardStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	totalTags := 0
	maxTags := 0
	minTags := int(^uint(0) >> 1) // Max int
	distribution := make([]int, NumShards)
	
	for i, shard := range s.shards {
		shard.mu.RLock()
		count := len(shard.tags)
		shard.mu.RUnlock()
		
		distribution[i] = count
		totalTags += count
		if count > maxTags {
			maxTags = count
		}
		if count < minTags {
			minTags = count
		}
	}
	
	stats["total_tags"] = totalTags
	stats["num_shards"] = NumShards
	stats["max_tags_per_shard"] = maxTags
	stats["min_tags_per_shard"] = minTags
	stats["avg_tags_per_shard"] = totalTags / NumShards
	
	return stats
}

// ToMap converts the sharded index to a regular map for persistence
func (idx *ShardedTagIndex) ToMap() map[string][]string {
	result := make(map[string][]string)
	
	// Collect all tags from all shards
	for _, shard := range idx.shards {
		shard.mu.RLock()
		for tag, entities := range shard.tags {
			result[tag] = append(result[tag], entities...)
		}
		shard.mu.RUnlock()
	}
	
	return result
}

// ShardedLock provides fine-grained locking by distributing locks across
// multiple shards. This reduces contention when locking different keys,
// as operations on different shards can proceed concurrently.
//
// Used extensively in TemporalRepository for:
//   - Entity-level locking (prevent concurrent modifications)
//   - Timeline bucket locking (temporal index updates)
//   - Tag index locking (consistent tag operations)
//
// Performance characteristics:
//   - O(1) lock acquisition
//   - Concurrent operations on ~1/numShards of keyspace
//   - Memory usage: O(numShards) - typically small
//
// Example usage:
//   lock := NewShardedLock(64)
//   lock.Lock("entity-123")
//   defer lock.Unlock("entity-123")
//   // ... critical section ...
type ShardedLock struct {
	locks     []*sync.RWMutex // Array of locks, one per shard
	numShards int             // Number of shards (not required to be power of 2)
}

// NewShardedLock creates a new sharded lock with the specified number of shards.
// More shards reduce contention but increase memory usage.
//
// Recommended shard counts:
//   - 16-32: Low concurrency applications
//   - 64-128: Medium concurrency
//   - 256+: High concurrency with many unique keys
//
// Parameters:
//   - numShards: Number of lock shards to create
//
// Returns:
//   - *ShardedLock: Initialized lock manager
func NewShardedLock(numShards int) *ShardedLock {
	sl := &ShardedLock{
		locks:     make([]*sync.RWMutex, numShards),
		numShards: numShards,
	}
	for i := 0; i < numShards; i++ {
		sl.locks[i] = &sync.RWMutex{}
	}
	return sl
}

// Lock acquires a write lock for the given key
func (sl *ShardedLock) Lock(key string) {
	shard := sl.getShard(key)
	sl.locks[shard].Lock()
}

// Unlock releases a write lock for the given key
func (sl *ShardedLock) Unlock(key string) {
	shard := sl.getShard(key)
	sl.locks[shard].Unlock()
}

// RLock acquires a read lock for the given key
func (sl *ShardedLock) RLock(key string) {
	shard := sl.getShard(key)
	sl.locks[shard].RLock()
}

// RUnlock releases a read lock for the given key
func (sl *ShardedLock) RUnlock(key string) {
	shard := sl.getShard(key)
	sl.locks[shard].RUnlock()
}

// getShard returns the shard index for a given key
func (sl *ShardedLock) getShard(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() % uint32(sl.numShards))
}

// WithLock executes a function while holding a write lock for the given key.
// This is a convenience method that ensures proper lock release via defer.
//
// Deadlock Prevention:
//   When locking multiple keys, always acquire locks in a consistent order
//   (e.g., alphabetical) to prevent circular wait conditions.
//
// Parameters:
//   - key: The key to lock (hashed to determine shard)
//   - fn: Function to execute while holding the lock
//
// Example:
//   sl.WithLock("user-123", func() {
//       // Critical section - exclusive access to "user-123"
//       updateUser(...)
//   })
func (sl *ShardedLock) WithLock(key string, fn func()) {
	sl.Lock(key)
	defer sl.Unlock(key)
	fn()
}

// WithRLock executes a function while holding a read lock for the given key.
// Multiple readers can hold the lock simultaneously for the same key.
//
// Use for read-only operations that must not see partial updates.
//
// Parameters:
//   - key: The key to lock (hashed to determine shard)
//   - fn: Function to execute while holding the read lock
//
// Example:
//   var user User
//   sl.WithRLock("user-123", func() {
//       user = getUser("user-123")
//   })
func (sl *ShardedLock) WithRLock(key string, fn func()) {
	sl.RLock(key)
	defer sl.RUnlock(key)
	fn()
}