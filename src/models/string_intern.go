package models

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

// internEntry represents an entry in the string intern cache
type internEntry struct {
	value      string
	accessTime int64 // nanoseconds for atomic access
	accessCount int64 // frequency tracking
	listElement *list.Element
}

// StringIntern provides a global string interning pool to reduce memory usage
// for frequently repeated strings like tags. Now with bounded size and LRU eviction.
type StringIntern struct {
	mu          sync.RWMutex
	strings     map[string]*internEntry
	lru         *list.List
	maxSize     int
	currentSize int
	
	// Memory tracking
	memoryUsed  int64
	memoryLimit int64
	
	// Statistics
	hits        int64
	misses      int64
	evictions   int64
	
	// Adaptive sizing
	lastResize  time.Time
	resizeFunc  func() int
}

// Default configuration from environment/config
var (
	defaultMaxSize     = 100000 // 100k strings
	defaultMemoryLimit = int64(100 * 1024 * 1024) // 100MB
)

// defaultStringInterner is the singleton instance for string interning
var defaultStringInterner *StringIntern

func init() {
	defaultStringInterner = &StringIntern{
		strings:     make(map[string]*internEntry),
		lru:         list.New(),
		maxSize:     defaultMaxSize,
		memoryLimit: defaultMemoryLimit,
		lastResize:  time.Now(),
	}
	// Set resize function after initialization to avoid cycle
	defaultStringInterner.resizeFunc = adaptiveSizeFunc
}

// Intern returns an interned version of the string with bounded memory usage
// If the string already exists in the pool, it returns the pooled version
// Otherwise, it adds the string to the pool with LRU eviction if needed
func Intern(s string) string {
	// Empty strings don't need interning
	if len(s) == 0 {
		return ""
	}
	
	// Fast path - check if already interned
	defaultStringInterner.mu.RLock()
	if entry, ok := defaultStringInterner.strings[s]; ok {
		// Update access time and count atomically
		atomic.StoreInt64(&entry.accessTime, time.Now().UnixNano())
		atomic.AddInt64(&entry.accessCount, 1)
		atomic.AddInt64(&defaultStringInterner.hits, 1)
		defaultStringInterner.mu.RUnlock()
		return entry.value
	}
	defaultStringInterner.mu.RUnlock()
	
	// Slow path - add to pool
	return defaultStringInterner.internSlow(s)
}

// internSlow handles adding new strings with eviction if needed
func (si *StringIntern) internSlow(s string) string {
	si.mu.Lock()
	defer si.mu.Unlock()
	
	// Double check in case another goroutine added it
	if entry, ok := si.strings[s]; ok {
		atomic.StoreInt64(&entry.accessTime, time.Now().UnixNano())
		atomic.AddInt64(&entry.accessCount, 1)
		atomic.AddInt64(&si.hits, 1)
		// Move to front of LRU
		si.lru.MoveToFront(entry.listElement)
		return entry.value
	}
	
	atomic.AddInt64(&si.misses, 1)
	
	// Check if we need to evict
	stringSize := int64(len(s) + 48) // String overhead estimate
	si.evictIfNeeded(stringSize)
	
	// Create new entry
	entry := &internEntry{
		value:       s,
		accessTime:  time.Now().UnixNano(),
		accessCount: 1,
	}
	
	// Add to LRU and map
	entry.listElement = si.lru.PushFront(s)
	si.strings[s] = entry
	si.currentSize++
	atomic.AddInt64(&si.memoryUsed, stringSize)
	
	return s
}

// evictIfNeeded removes least recently used entries if limits are exceeded
func (si *StringIntern) evictIfNeeded(newSize int64) {
	// Check both count and memory limits
	for (si.currentSize >= si.maxSize || atomic.LoadInt64(&si.memoryUsed)+newSize > si.memoryLimit) && si.lru.Len() > 0 {
		// Get least recently used
		elem := si.lru.Back()
		if elem == nil {
			break
		}
		
		key := elem.Value.(string)
		if entry, ok := si.strings[key]; ok {
			// Remove from map and list
			delete(si.strings, key)
			si.lru.Remove(elem)
			si.currentSize--
			
			// Update memory tracking
			evictedSize := int64(len(key) + 48)
			atomic.AddInt64(&si.memoryUsed, -evictedSize)
			atomic.AddInt64(&si.evictions, 1)
			
			// Don't evict frequently accessed items
			if atomic.LoadInt64(&entry.accessCount) > 100 {
				// Move to front instead of evicting
				si.lru.MoveToFront(elem)
				continue
			}
		}
	}
	
	// Check if we should adapt size based on memory pressure
	if time.Since(si.lastResize) > 30*time.Second && si.resizeFunc != nil {
		newMaxSize := si.resizeFunc()
		if newMaxSize != si.maxSize {
			si.maxSize = newMaxSize
			si.lastResize = time.Now()
		}
	}
}

// adaptiveSizeFunc adjusts cache size based on memory pressure
func adaptiveSizeFunc() int {
	// This would check runtime.MemStats in production
	// For now, return current size
	return defaultStringInterner.maxSize
}

// InternSlice interns all strings in a slice
func InternSlice(strings []string) []string {
	for i, s := range strings {
		strings[i] = Intern(s)
	}
	return strings
}

// Size returns the number of interned strings
func Size() int {
	defaultStringInterner.mu.RLock()
	defer defaultStringInterner.mu.RUnlock()
	return defaultStringInterner.currentSize
}

// MemoryUsed returns the approximate memory used by interned strings
func MemoryUsed() int64 {
	return atomic.LoadInt64(&defaultStringInterner.memoryUsed)
}

// StringInternStats returns interning statistics for the simple interner
type StringInternStats struct {
	Size        int
	MemoryUsed  int64
	Hits        int64
	Misses      int64
	Evictions   int64
	HitRate     float64
	MaxSize     int
	MemoryLimit int64
}

func Stats() StringInternStats {
	defaultStringInterner.mu.RLock()
	defer defaultStringInterner.mu.RUnlock()
	
	hits := atomic.LoadInt64(&defaultStringInterner.hits)
	misses := atomic.LoadInt64(&defaultStringInterner.misses)
	total := hits + misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}
	
	return StringInternStats{
		Size:        defaultStringInterner.currentSize,
		MemoryUsed:  atomic.LoadInt64(&defaultStringInterner.memoryUsed),
		Hits:        hits,
		Misses:      misses,
		Evictions:   atomic.LoadInt64(&defaultStringInterner.evictions),
		HitRate:     hitRate,
		MaxSize:     defaultStringInterner.maxSize,
		MemoryLimit: defaultStringInterner.memoryLimit,
	}
}

// SetMaxSize updates the maximum number of strings that can be interned
func SetMaxSize(size int) {
	defaultStringInterner.mu.Lock()
	defer defaultStringInterner.mu.Unlock()
	defaultStringInterner.maxSize = size
}

// SetMemoryLimit updates the memory limit for interned strings
func SetMemoryLimit(limit int64) {
	defaultStringInterner.mu.Lock()
	defer defaultStringInterner.mu.Unlock()
	defaultStringInterner.memoryLimit = limit
}

// Clear removes all interned strings (use with caution)
func Clear() {
	defaultStringInterner.mu.Lock()
	defer defaultStringInterner.mu.Unlock()
	defaultStringInterner.strings = make(map[string]*internEntry)
	defaultStringInterner.lru = list.New()
	defaultStringInterner.currentSize = 0
	atomic.StoreInt64(&defaultStringInterner.memoryUsed, 0)
	atomic.StoreInt64(&defaultStringInterner.hits, 0)
	atomic.StoreInt64(&defaultStringInterner.misses, 0)
	atomic.StoreInt64(&defaultStringInterner.evictions, 0)
}

// GetDefaultStringInterner returns the default string interner instance
func GetDefaultStringInterner() *StringIntern {
	return defaultStringInterner
}

// TriggerPressureCleanup performs aggressive cleanup under memory pressure
func (si *StringIntern) TriggerPressureCleanup(pressure float64) {
	si.mu.Lock()
	defer si.mu.Unlock()
	
	// Calculate how much to evict based on pressure
	targetEviction := int(float64(si.currentSize) * pressure * 0.3) // Evict up to 30% of entries under high pressure
	if targetEviction < 10 {
		targetEviction = 10 // Minimum cleanup
	}
	
	evicted := 0
	// Evict from the end of LRU (least recently used)
	for evicted < targetEviction && si.lru.Len() > 0 {
		element := si.lru.Back()
		if element != nil {
			// The LRU stores the string itself
			str := element.Value.(string)
			
			// Remove from map and list
			delete(si.strings, str)
			si.lru.Remove(element)
			
			// Update memory usage
			atomic.AddInt64(&si.memoryUsed, -int64(len(str)))
			si.currentSize--
			atomic.AddInt64(&si.evictions, 1)
			
			evicted++
		} else {
			break
		}
	}
	
	// If pressure is critical, also reduce max size temporarily
	if pressure > 0.9 {
		si.maxSize = int(float64(si.maxSize) * 0.8) // Reduce by 20%
		if si.maxSize < 10000 {
			si.maxSize = 10000 // Minimum size
		}
	}
}