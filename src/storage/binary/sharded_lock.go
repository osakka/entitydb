package binary

import (
	"hash/fnv"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	// NumShards is the number of lock shards (must be power of 2)
	NumShards = 256
	// MaxReadersBatch is the number of readers before a writer gets priority
	MaxReadersBatch = 10
	// NumLockShards for ShardedLock
	NumLockShards = 64
)

// ShardedTagIndex implements a sharded tag index with fair locking
type ShardedTagIndex struct {
	shards [NumShards]*TagIndexShard
}

// TagIndexShard represents a single shard of the tag index
type TagIndexShard struct {
	mu       sync.RWMutex
	tags     map[string][]string // tag -> entity IDs
	queue    *FairQueue
}

// FairQueue implements fair queuing for readers and writers
type FairQueue struct {
	mu            sync.Mutex
	readerQueue   []chan struct{}
	writerQueue   []chan struct{}
	activeReaders int32
	readerCount   int32 // Count of reads since last write
	writerWaiting bool
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

// getShard returns the shard for a given tag
func (s *ShardedTagIndex) getShard(tag string) *TagIndexShard {
	h := fnv.New32a()
	h.Write([]byte(tag))
	shardIdx := h.Sum32() & (NumShards - 1)
	return s.shards[shardIdx]
}

// AddTag adds an entity ID to a tag's index
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

// AcquireRead acquires a read lock with fair queuing
func (q *FairQueue) AcquireRead() {
	q.mu.Lock()
	
	// If a writer is waiting and we've had enough reads, wait
	if q.writerWaiting && atomic.LoadInt32(&q.readerCount) >= MaxReadersBatch {
		ch := make(chan struct{})
		q.readerQueue = append(q.readerQueue, ch)
		q.mu.Unlock()
		<-ch
		q.mu.Lock()
	}
	
	atomic.AddInt32(&q.activeReaders, 1)
	atomic.AddInt32(&q.readerCount, 1)
	q.mu.Unlock()
}

// ReleaseRead releases a read lock
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

// ShardedLock provides fine-grained locking with multiple shards
// This is used by TemporalRepository for entity and bucket locks
type ShardedLock struct {
	locks    []*sync.RWMutex
	numShards int
}

// NewShardedLock creates a new sharded lock with the specified number of shards
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

// WithLock executes a function while holding a write lock for the given key
func (sl *ShardedLock) WithLock(key string, fn func()) {
	sl.Lock(key)
	defer sl.Unlock(key)
	fn()
}

// WithRLock executes a function while holding a read lock for the given key
func (sl *ShardedLock) WithRLock(key string, fn func()) {
	sl.RLock(key)
	defer sl.RUnlock(key)
	fn()
}