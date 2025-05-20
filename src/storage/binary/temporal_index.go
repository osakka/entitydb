package binary

import (
	"sync"
	"time"
	"sort"
	"strings"
)

// TemporalIndex provides efficient temporal queries
type TemporalIndex struct {
	mu              sync.RWMutex
	timestampIndex  map[string][]TemporalEntry // entityID -> sorted temporal entries
	timeRangeIndex  map[int64][]string         // timestamp bucket -> entity IDs
	bucketSize      int64                      // bucket size in seconds (3600 = 1 hour)
}

type TemporalEntry struct {
	EntityID  string
	Timestamp time.Time
	Tag       string
}

// NewTemporalIndex creates a new temporal index
func NewTemporalIndex() *TemporalIndex {
	return &TemporalIndex{
		timestampIndex: make(map[string][]TemporalEntry),
		timeRangeIndex: make(map[int64][]string),
		bucketSize:     3600, // 1 hour buckets
	}
}

// AddEntry adds a temporal entry to the index
func (ti *TemporalIndex) AddEntry(entityID string, tag string, timestamp time.Time) {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	
	// Extract timestamp from temporal tag if not provided
	if timestamp.IsZero() && strings.Contains(tag, "|") {
		parts := strings.SplitN(tag, "|", 2)
		if len(parts) == 2 {
			// Try to parse as Unix timestamp
			if t, err := time.Parse("2006-01-02T15:04:05.999999999", parts[0]); err == nil {
				timestamp = t
			}
		}
	}
	
	if timestamp.IsZero() {
		return // Skip non-temporal tags
	}
	
	entry := TemporalEntry{
		EntityID:  entityID,
		Timestamp: timestamp,
		Tag:       tag,
	}
	
	// Add to entity timestamp index
	ti.timestampIndex[entityID] = append(ti.timestampIndex[entityID], entry)
	
	// Sort entries for this entity by timestamp
	sort.Slice(ti.timestampIndex[entityID], func(i, j int) bool {
		return ti.timestampIndex[entityID][i].Timestamp.Before(ti.timestampIndex[entityID][j].Timestamp)
	})
	
	// Add to time range index (bucketed)
	bucket := timestamp.Unix() / ti.bucketSize
	ti.timeRangeIndex[bucket] = append(ti.timeRangeIndex[bucket], entityID)
}

// RemoveEntity removes all entries for an entity
func (ti *TemporalIndex) RemoveEntity(entityID string) {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	
	// Remove from timestamp index
	delete(ti.timestampIndex, entityID)
	
	// Remove from time range index
	for bucket, entities := range ti.timeRangeIndex {
		filtered := make([]string, 0)
		for _, id := range entities {
			if id != entityID {
				filtered = append(filtered, id)
			}
		}
		if len(filtered) == 0 {
			delete(ti.timeRangeIndex, bucket)
		} else {
			ti.timeRangeIndex[bucket] = filtered
		}
	}
}

// GetEntityAsOf returns the entity state at a specific timestamp
func (ti *TemporalIndex) GetEntityAsOf(entityID string, timestamp time.Time) []string {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	
	entries, exists := ti.timestampIndex[entityID]
	if !exists {
		return nil
	}
	
	// Binary search for entries up to the timestamp
	validTags := make([]string, 0)
	namespaceLatest := make(map[string]string)
	
	for _, entry := range entries {
		if entry.Timestamp.After(timestamp) {
			break
		}
		
		// Extract namespace from tag
		tag := entry.Tag
		if idx := strings.Index(tag, "|"); idx != -1 {
			tag = tag[idx+1:]
		}
		
		namespace := ""
		if idx := strings.Index(tag, ":"); idx != -1 {
			namespace = tag[:idx]
		} else if idx := strings.Index(tag, "="); idx != -1 {
			namespace = tag[:idx]
		}
		
		if namespace != "" {
			namespaceLatest[namespace] = tag
		}
	}
	
	// Convert map to slice
	for _, tag := range namespaceLatest {
		validTags = append(validTags, tag)
	}
	
	return validTags
}

// GetChangesInRange returns entities that changed within a time range
func (ti *TemporalIndex) GetChangesInRange(from, to time.Time) []string {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	
	entitySet := make(map[string]bool)
	
	// Calculate bucket range
	fromBucket := from.Unix() / ti.bucketSize
	toBucket := to.Unix() / ti.bucketSize
	
	// Check all buckets in range
	for bucket := fromBucket; bucket <= toBucket; bucket++ {
		entities, exists := ti.timeRangeIndex[bucket]
		if !exists {
			continue
		}
		
		// Add entities from this bucket
		for _, entityID := range entities {
			// Fine-grained check for entities in this bucket
			entries, exists := ti.timestampIndex[entityID]
			if !exists {
				continue
			}
			
			for _, entry := range entries {
				if entry.Timestamp.After(from) && entry.Timestamp.Before(to) {
					entitySet[entityID] = true
					break
				}
			}
		}
	}
	
	// Convert set to slice
	result := make([]string, 0, len(entitySet))
	for entityID := range entitySet {
		result = append(result, entityID)
	}
	
	return result
}

// GetRecentChanges returns entities changed since a timestamp
func (ti *TemporalIndex) GetRecentChanges(since time.Time) []string {
	return ti.GetChangesInRange(since, time.Now())
}

// GetEntityHistory returns all temporal entries for an entity
func (ti *TemporalIndex) GetEntityHistory(entityID string, from, to time.Time) []TemporalEntry {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	
	entries, exists := ti.timestampIndex[entityID]
	if !exists {
		return nil
	}
	
	result := make([]TemporalEntry, 0)
	for _, entry := range entries {
		if entry.Timestamp.After(from) && entry.Timestamp.Before(to) {
			result = append(result, entry)
		}
	}
	
	return result
}