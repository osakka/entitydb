// Package binary provides bounded entity caching with LRU eviction.
package binary

import (
	"container/list"
	"entitydb/models"
	"sync"
	"sync/atomic"
	"time"
)

// cacheEntry represents an entry in the entity cache
type cacheEntry struct {
	entity      *models.Entity
	size        int64
	accessTime  int64
	accessCount int64
	listElement *list.Element
}

// BoundedEntityCache provides a size-limited entity cache with LRU eviction
// and adaptive sizing based on memory pressure.
type BoundedEntityCache struct {
	mu          sync.RWMutex
	entries     map[string]*cacheEntry
	lru         *list.List
	maxSize     int
	currentSize int
	
	// Memory tracking
	memoryUsed  int64
	memoryLimit int64
	
	// Statistics
	hits      int64
	misses    int64
	evictions int64
	
	// Adaptive sizing
	lastResize   time.Time
	resizeFunc   func() int
	evictionFunc func(entityID string, entity *models.Entity)
}

// NewBoundedEntityCache creates a new bounded entity cache
func NewBoundedEntityCache(maxSize int, memoryLimit int64) *BoundedEntityCache {
	return &BoundedEntityCache{
		entries:     make(map[string]*cacheEntry),
		lru:         list.New(),
		maxSize:     maxSize,
		memoryLimit: memoryLimit,
		lastResize:  time.Now(),
	}
}

// Get retrieves an entity from the cache
func (c *BoundedEntityCache) Get(entityID string) (*models.Entity, bool) {
	c.mu.RLock()
	entry, ok := c.entries[entityID]
	if !ok {
		c.mu.RUnlock()
		atomic.AddInt64(&c.misses, 1)
		return nil, false
	}
	
	// Update access stats
	atomic.StoreInt64(&entry.accessTime, time.Now().UnixNano())
	atomic.AddInt64(&entry.accessCount, 1)
	atomic.AddInt64(&c.hits, 1)
	
	// Move to front of LRU (upgrade to write lock)
	c.mu.RUnlock()
	c.mu.Lock()
	c.lru.MoveToFront(entry.listElement)
	c.mu.Unlock()
	
	return entry.entity, true
}

// Put adds or updates an entity in the cache
func (c *BoundedEntityCache) Put(entityID string, entity *models.Entity) {
	if entity == nil {
		return
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Calculate entity size
	entitySize := c.calculateEntitySize(entity)
	
	// Check if already exists
	if entry, ok := c.entries[entityID]; ok {
		// Update existing entry
		oldSize := entry.size
		entry.entity = entity
		entry.size = entitySize
		entry.accessTime = time.Now().UnixNano()
		atomic.AddInt64(&entry.accessCount, 1)
		
		// Update memory tracking
		atomic.AddInt64(&c.memoryUsed, entitySize-oldSize)
		
		// Move to front
		c.lru.MoveToFront(entry.listElement)
		return
	}
	
	// Check if we need to evict
	c.evictIfNeeded(entitySize)
	
	// Create new entry
	entry := &cacheEntry{
		entity:      entity,
		size:        entitySize,
		accessTime:  time.Now().UnixNano(),
		accessCount: 1,
	}
	
	// Add to LRU and map
	entry.listElement = c.lru.PushFront(entityID)
	c.entries[entityID] = entry
	c.currentSize++
	atomic.AddInt64(&c.memoryUsed, entitySize)
}

// Delete removes an entity from the cache
func (c *BoundedEntityCache) Delete(entityID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if entry, ok := c.entries[entityID]; ok {
		delete(c.entries, entityID)
		c.lru.Remove(entry.listElement)
		c.currentSize--
		atomic.AddInt64(&c.memoryUsed, -entry.size)
		
		if c.evictionFunc != nil {
			c.evictionFunc(entityID, entry.entity)
		}
	}
}

// evictIfNeeded removes least recently used entries if limits are exceeded
func (c *BoundedEntityCache) evictIfNeeded(newSize int64) {
	for (c.currentSize >= c.maxSize || atomic.LoadInt64(&c.memoryUsed)+newSize > c.memoryLimit) && c.lru.Len() > 0 {
		elem := c.lru.Back()
		if elem == nil {
			break
		}
		
		entityID := elem.Value.(string)
		if entry, ok := c.entries[entityID]; ok {
			// Don't evict frequently accessed items
			if atomic.LoadInt64(&entry.accessCount) > 100 {
				// Move to middle instead of evicting
				c.lru.MoveAfter(elem, c.lru.Front())
				continue
			}
			
			// Remove from cache
			delete(c.entries, entityID)
			c.lru.Remove(elem)
			c.currentSize--
			atomic.AddInt64(&c.memoryUsed, -entry.size)
			atomic.AddInt64(&c.evictions, 1)
			
			if c.evictionFunc != nil {
				c.evictionFunc(entityID, entry.entity)
			}
		}
	}
	
	// Adaptive sizing
	if time.Since(c.lastResize) > 30*time.Second && c.resizeFunc != nil {
		newMaxSize := c.resizeFunc()
		if newMaxSize != c.maxSize {
			c.maxSize = newMaxSize
			c.lastResize = time.Now()
		}
	}
}

// calculateEntitySize estimates the memory size of an entity
func (c *BoundedEntityCache) calculateEntitySize(entity *models.Entity) int64 {
	if entity == nil {
		return 0
	}
	
	// Base struct size
	size := int64(200) // Approximate base struct overhead
	
	// ID size
	size += int64(len(entity.ID))
	
	// Tags size
	for _, tag := range entity.Tags {
		size += int64(len(tag) + 24) // String + overhead
	}
	
	// Content size
	if entity.Content != nil {
		size += int64(len(entity.Content))
	}
	
	return size
}

// Clear removes all entries from the cache
func (c *BoundedEntityCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.entries = make(map[string]*cacheEntry)
	c.lru = list.New()
	c.currentSize = 0
	atomic.StoreInt64(&c.memoryUsed, 0)
}

// Stats returns cache statistics
type CacheStats struct {
	Size        int
	MemoryUsed  int64
	Hits        int64
	Misses      int64
	Evictions   int64
	HitRate     float64
	MaxSize     int
	MemoryLimit int64
}

func (c *BoundedEntityCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	hits := atomic.LoadInt64(&c.hits)
	misses := atomic.LoadInt64(&c.misses)
	total := hits + misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}
	
	return CacheStats{
		Size:        c.currentSize,
		MemoryUsed:  atomic.LoadInt64(&c.memoryUsed),
		Hits:        hits,
		Misses:      misses,
		Evictions:   atomic.LoadInt64(&c.evictions),
		HitRate:     hitRate,
		MaxSize:     c.maxSize,
		MemoryLimit: c.memoryLimit,
	}
}

// SetMaxSize updates the maximum cache size
func (c *BoundedEntityCache) SetMaxSize(size int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxSize = size
}

// SetMemoryLimit updates the memory limit
func (c *BoundedEntityCache) SetMemoryLimit(limit int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.memoryLimit = limit
}

// SetEvictionCallback sets a callback for when entities are evicted
func (c *BoundedEntityCache) SetEvictionCallback(fn func(entityID string, entity *models.Entity)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evictionFunc = fn
}

// TriggerPressureCleanup performs aggressive cache cleanup under memory pressure
func (c *BoundedEntityCache) TriggerPressureCleanup(pressure float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Calculate how much to evict based on pressure
	targetEviction := int(float64(c.currentSize) * pressure * 0.4) // Evict up to 40% of entries under high pressure
	if targetEviction < 5 {
		targetEviction = 5 // Minimum cleanup
	}
	
	evicted := 0
	// Evict from the end of LRU (least recently used)
	for evicted < targetEviction && c.lru.Len() > 0 {
		elem := c.lru.Back()
		if elem != nil {
			entityID := elem.Value.(string)
			if entry, ok := c.entries[entityID]; ok {
				// Remove from cache regardless of access count under pressure
				delete(c.entries, entityID)
				c.lru.Remove(elem)
				c.currentSize--
				atomic.AddInt64(&c.memoryUsed, -entry.size)
				atomic.AddInt64(&c.evictions, 1)
				
				if c.evictionFunc != nil {
					c.evictionFunc(entityID, entry.entity)
				}
				
				evicted++
			}
		} else {
			break
		}
	}
	
	// If pressure is critical, also reduce max size temporarily
	if pressure > 0.9 {
		c.maxSize = int(float64(c.maxSize) * 0.7) // Reduce by 30%
		if c.maxSize < 100 {
			c.maxSize = 100 // Minimum cache size
		}
		
		// Also reduce memory limit temporarily
		c.memoryLimit = int64(float64(c.memoryLimit) * 0.7)
		if c.memoryLimit < 10*1024*1024 {
			c.memoryLimit = 10 * 1024 * 1024 // Minimum 10MB
		}
	}
}