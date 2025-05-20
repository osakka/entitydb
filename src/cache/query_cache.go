package cache

import (
	"sync"
	"time"
)

// CacheEntry represents a cached query result
type CacheEntry struct {
	Result      interface{}
	Timestamp   time.Time
	AccessCount int
}

// QueryCache implements a simple in-memory cache for query results
type QueryCache struct {
	mu       sync.RWMutex
	entries  map[string]*CacheEntry
	maxSize  int
	ttl      time.Duration
}

// NewQueryCache creates a new query cache
func NewQueryCache(maxSize int, ttl time.Duration) *QueryCache {
	cache := &QueryCache{
		entries:  make(map[string]*CacheEntry),
		maxSize:  maxSize,
		ttl:      ttl,
	}
	
	// Start cleanup goroutine
	go cache.cleanupLoop()
	
	return cache
}

// Get retrieves a cached result
func (c *QueryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()
	
	if !exists {
		return nil, false
	}
	
	// Check if entry is expired
	if time.Since(entry.Timestamp) > c.ttl {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil, false
	}
	
	// Update access count
	c.mu.Lock()
	entry.AccessCount++
	c.mu.Unlock()
	
	return entry.Result, true
}

// Set stores a result in the cache
func (c *QueryCache) Set(key string, result interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Check if we need to evict entries
	if len(c.entries) >= c.maxSize {
		c.evictLRU()
	}
	
	c.entries[key] = &CacheEntry{
		Result:      result,
		Timestamp:   time.Now(),
		AccessCount: 1,
	}
}

// evictLRU removes the least recently used entry
func (c *QueryCache) evictLRU() {
	var lruKey string
	var lruEntry *CacheEntry
	
	for key, entry := range c.entries {
		if lruEntry == nil || entry.AccessCount < lruEntry.AccessCount {
			lruKey = key
			lruEntry = entry
		}
	}
	
	if lruKey != "" {
		delete(c.entries, lruKey)
	}
}

// cleanupLoop periodically removes expired entries
func (c *QueryCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.Sub(entry.Timestamp) > c.ttl {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// Clear removes all entries from the cache
func (c *QueryCache) Clear() {
	c.mu.Lock()
	c.entries = make(map[string]*CacheEntry)
	c.mu.Unlock()
}

// Invalidate removes specific entries from the cache
func (c *QueryCache) Invalidate(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Simple pattern matching - could be enhanced
	for key := range c.entries {
		if len(pattern) == 0 || key[:len(pattern)] == pattern {
			delete(c.entries, key)
		}
	}
}

// GetStats returns cache statistics
func (c *QueryCache) GetStats() (int, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	total := len(c.entries)
	var hits int
	for _, entry := range c.entries {
		if entry.AccessCount > 1 {
			hits++
		}
	}
	
	return total, hits
}