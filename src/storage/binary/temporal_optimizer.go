package binary

import (
	"entitydb/models"
	"sync"
	"time"
	"strings"
	"sort"
)

// TemporalOptimizer provides high-performance temporal query optimizations
type TemporalOptimizer struct {
	mu sync.RWMutex
	
	// Timeline index: timestamp -> entity IDs
	timeline map[int64][]string
	
	// Entity temporal index: entity ID -> sorted timestamps
	entityTimeline map[string][]int64
	
	// Time-bucketed index for range queries (hourly buckets)
	hourlyBuckets map[int64]map[string]bool
	
	// Cache for as-of queries
	asOfCache map[string]*models.Entity
	cacheMu   sync.RWMutex
}

// NewTemporalOptimizer creates a new temporal optimizer
func NewTemporalOptimizer() *TemporalOptimizer {
	return &TemporalOptimizer{
		timeline:       make(map[int64][]string),
		entityTimeline: make(map[string][]int64),
		hourlyBuckets:  make(map[int64]map[string]bool),
		asOfCache:      make(map[string]*models.Entity),
	}
}

// IndexEntity indexes an entity's temporal data for fast queries
func (t *TemporalOptimizer) IndexEntity(entity *models.Entity) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	timestamps := []int64{}
	
	// Extract timestamps from tags
	for _, tag := range entity.Tags {
		parts := strings.SplitN(tag, "|", 2)
		if len(parts) == 2 {
			// Parse timestamp
			ts, err := time.Parse(time.RFC3339Nano, parts[0])
			if err == nil {
				nano := ts.UnixNano()
				timestamps = append(timestamps, nano)
				
				// Update timeline index
				t.timeline[nano] = append(t.timeline[nano], entity.ID)
				
				// Update hourly bucket
				hourBucket := nano / (3600 * 1e9) * (3600 * 1e9) // Round to hour
				if t.hourlyBuckets[hourBucket] == nil {
					t.hourlyBuckets[hourBucket] = make(map[string]bool)
				}
				t.hourlyBuckets[hourBucket][entity.ID] = true
			}
		}
	}
	
	// Sort timestamps for this entity
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] < timestamps[j]
	})
	
	// Store entity timeline
	t.entityTimeline[entity.ID] = timestamps
}

// FindEntitiesAsOf finds entities that existed at a specific time
func (t *TemporalOptimizer) FindEntitiesAsOf(asOf time.Time, filter func(*models.Entity) bool) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	asOfNano := asOf.UnixNano()
	results := make(map[string]bool)
	
	// Use hourly buckets for efficient lookup
	hourBucket := asOfNano / (3600 * 1e9) * (3600 * 1e9)
	
	// Check current hour and previous hour (to catch entities created near boundary)
	for _, bucket := range []int64{hourBucket - 3600*1e9, hourBucket} {
		if entities, ok := t.hourlyBuckets[bucket]; ok {
			for entityID := range entities {
				// Check if entity existed at asOf time
				if timestamps, ok := t.entityTimeline[entityID]; ok && len(timestamps) > 0 {
					// Binary search for the right timestamp
					idx := sort.Search(len(timestamps), func(i int) bool {
						return timestamps[i] > asOfNano
					})
					
					// If idx > 0, entity existed at asOf time
					if idx > 0 {
						results[entityID] = true
					}
				}
			}
		}
	}
	
	// Convert to slice
	entityIDs := make([]string, 0, len(results))
	for id := range results {
		entityIDs = append(entityIDs, id)
	}
	
	return entityIDs
}

// GetEntityAsOf optimized version using cache
func (t *TemporalOptimizer) GetEntityAsOf(repo *EntityRepository, entityID string, asOf time.Time) (*models.Entity, error) {
	// Check cache first
	cacheKey := entityID + ":" + asOf.Format(time.RFC3339Nano)
	
	t.cacheMu.RLock()
	if cached, ok := t.asOfCache[cacheKey]; ok {
		t.cacheMu.RUnlock()
		return cached, nil
	}
	t.cacheMu.RUnlock()
	
	// Get entity from repo
	entity, err := repo.GetByID(entityID)
	if err != nil {
		return nil, err
	}
	
	// Filter tags by timestamp
	asOfNano := asOf.UnixNano()
	filteredTags := []string{}
	
	for _, tag := range entity.Tags {
		parts := strings.SplitN(tag, "|", 2)
		if len(parts) == 2 {
			ts, err := time.Parse(time.RFC3339Nano, parts[0])
			if err == nil && ts.UnixNano() <= asOfNano {
				filteredTags = append(filteredTags, tag)
			}
		}
	}
	
	// Create historical view
	historicalEntity := &models.Entity{
		ID:   entity.ID,
		Tags: filteredTags,
		Content: []byte{}, // TODO: Filter content by timestamp
	}
	
	// Cache the result
	t.cacheMu.Lock()
	t.asOfCache[cacheKey] = historicalEntity
	t.cacheMu.Unlock()
	
	return historicalEntity, nil
}

// FindEntitiesInRange finds entities modified within a time range
func (t *TemporalOptimizer) FindEntitiesInRange(start, end time.Time) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	startNano := start.UnixNano()
	endNano := end.UnixNano()
	results := make(map[string]bool)
	
	// Use hourly buckets for efficient range query
	startBucket := startNano / (3600 * 1e9) * (3600 * 1e9)
	endBucket := endNano / (3600 * 1e9) * (3600 * 1e9)
	
	for bucket := startBucket; bucket <= endBucket; bucket += 3600 * 1e9 {
		if entities, ok := t.hourlyBuckets[bucket]; ok {
			for entityID := range entities {
				// Check if entity has changes in the time range
				if timestamps, ok := t.entityTimeline[entityID]; ok {
					for _, ts := range timestamps {
						if ts >= startNano && ts <= endNano {
							results[entityID] = true
							break
						}
					}
				}
			}
		}
	}
	
	// Convert to slice
	entityIDs := make([]string, 0, len(results))
	for id := range results {
		entityIDs = append(entityIDs, id)
	}
	
	return entityIDs
}

// GetTemporalStats returns statistics about temporal data
func (t *TemporalOptimizer) GetTemporalStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	totalTimestamps := 0
	for _, timestamps := range t.entityTimeline {
		totalTimestamps += len(timestamps)
	}
	
	return map[string]interface{}{
		"entities":         len(t.entityTimeline),
		"totalTimestamps":  totalTimestamps,
		"timelineEntries":  len(t.timeline),
		"hourlyBuckets":    len(t.hourlyBuckets),
		"cacheSize":        len(t.asOfCache),
	}
}