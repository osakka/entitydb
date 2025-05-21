package binary

import (
	"entitydb/models"
	"entitydb/logger"
	"sync"
	"time"
	"sort"
	"runtime"
	"sync/atomic"
	"fmt"
	"strconv"
	"strings"
	"errors"
)

// TemporalRepository extends HighPerformanceRepository with temporal features
type TemporalRepository struct {
	*HighPerformanceRepository
	
	// Temporal indexes for ultra-fast time queries
	timelineIndex    *TemporalBTree           // B-tree for ordered timeline
	bucketIndex      map[int64]*EntitySet     // Time buckets for range queries
	entityTimelines  map[string]*Timeline     // Per-entity timeline
	
	// Advanced caching
	temporalCache    *TemporalCache
	
	// Statistics
	temporalStats    *TemporalStats
	
	// Configuration
	bucketSize       int64 // Default to hour buckets
	
	// Locks
	timelineMu       sync.RWMutex
	bucketMu         sync.RWMutex
}

// EntitySet is a thread-safe set of entity IDs
type EntitySet struct {
	mu    sync.RWMutex
	items map[string]bool
}

// Timeline tracks all timestamps for an entity
type Timeline struct {
	mu         sync.RWMutex
	entityID   string
	timestamps []int64 // Sorted timestamps
	tags       map[int64][]string // Timestamp -> tags at that time
}

// TemporalCache provides high-speed temporal query caching
type TemporalCache struct {
	mu        sync.RWMutex
	asOfCache map[string]*models.Entity // "entityID:timestamp" -> entity
	maxSize   int
}

// TemporalStats tracks performance metrics
type TemporalStats struct {
	asOfQueries      uint64
	rangeQueries     uint64
	cacheHits        uint64
	cacheMisses      uint64
	avgAsOfLatency   int64 // nanoseconds
	avgRangeLatency  int64 // nanoseconds
}

// NewTemporalRepository creates a temporal-optimized repository
func NewTemporalRepository(dataPath string) (*TemporalRepository, error) {
	// Create base high-performance repository
	highPerfRepo, err := NewHighPerformanceRepository(dataPath)
	if err != nil {
		return nil, err
	}
	
	// Initialize temporal extensions
	repo := &TemporalRepository{
		HighPerformanceRepository: highPerfRepo,
		timelineIndex:        NewTemporalBTree(32), // degree 32 for balanced performance
		bucketIndex:          make(map[int64]*EntitySet),
		entityTimelines:      make(map[string]*Timeline),
		temporalCache:        &TemporalCache{
			asOfCache: make(map[string]*models.Entity),
			maxSize:   10000,
		},
		temporalStats: &TemporalStats{},
		bucketSize:    HourBucket.BucketSize,
	}
	
	// Build temporal indexes in parallel
	go repo.buildTemporalIndexes()
	
	return repo, nil
}

// buildTemporalIndexes builds all temporal indexes
func (r *TemporalRepository) buildTemporalIndexes() {
	logger.Info("Building temporal indexes...")
	start := time.Now()
	
	// Get all entities
	entities, err := r.HighPerformanceRepository.ListByTag("")
	if err != nil {
		logger.Error("Failed to get entities for temporal indexing: %v", err)
		return
	}
	
	// Use parallel processing
	numWorkers := runtime.NumCPU()
	workChan := make(chan *models.Entity, len(entities))
	var wg sync.WaitGroup
	
	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entity := range workChan {
				r.indexEntityTemporal(entity)
			}
		}()
	}
	
	// Queue work
	for _, entity := range entities {
		workChan <- entity
	}
	close(workChan)
	
	// Wait for completion
	wg.Wait()
	
	logger.Info("Built temporal indexes in %v", time.Since(start))
}

// indexEntityTemporal indexes a single entity's temporal data
func (r *TemporalRepository) indexEntityTemporal(entity *models.Entity) {
	timeline := &Timeline{
		entityID:   entity.ID,
		timestamps: make([]int64, 0),
		tags:       make(map[int64][]string),
	}
	
	// Parse all temporal tags
	for _, tag := range entity.Tags {
		timestamp, cleanTag, err := ParseTemporalTagImproved(tag)
		if err != nil {
			continue // Skip non-temporal tags
		}
		
		timeline.timestamps = append(timeline.timestamps, timestamp)
		timeline.tags[timestamp] = append(timeline.tags[timestamp], cleanTag)
		
		// Update timeline index
		r.timelineMu.Lock()
		r.timelineIndex.Put(timestamp, entity.ID)
		r.timelineMu.Unlock()
		
		// Update bucket index
		bucket := TimeBucket{r.bucketSize}.GetBucket(timestamp)
		r.bucketMu.Lock()
		if r.bucketIndex[bucket] == nil {
			r.bucketIndex[bucket] = &EntitySet{
				items: make(map[string]bool),
			}
		}
		r.bucketIndex[bucket].mu.Lock()
		r.bucketIndex[bucket].items[entity.ID] = true
		r.bucketIndex[bucket].mu.Unlock()
		r.bucketMu.Unlock()
	}
	
	// Sort timestamps
	sort.Slice(timeline.timestamps, func(i, j int) bool {
		return timeline.timestamps[i] < timeline.timestamps[j]
	})
	
	// Store timeline
	r.entityTimelines[entity.ID] = timeline
}

// GetByID implements entity retrieval by delegating to base repository
func (r *TemporalRepository) GetByID(id string) (*models.Entity, error) {
	// Debug log
	logger.Debug("TemporalRepository.GetByID: Fetching entity with ID %s", id)
	
	// Use the embedded HighPerformanceRepository's GetByID
	entity, err := r.HighPerformanceRepository.GetByID(id)
	
	if err != nil {
		logger.Error("TemporalRepository.GetByID: Failed to get entity %s: %v", id, err)
	} else if entity != nil {
		logger.Debug("TemporalRepository.GetByID: Found entity %s with %d tags and %d bytes content", 
			id, len(entity.Tags), len(entity.Content))
	}
	
	return entity, err
}

// ParseTemporalTagImproved parses a tag with timestamp prefix (enhanced version)
// Format can be either:
// 1. "2025-05-20T20:02:48.098692124+01:00|type:test" (pipe format)
// 2. "2025-05-20T20:02:48.098692124.type:test" (dot format, deprecated)
func ParseTemporalTagImproved(tag string) (int64, string, error) {
	// Try pipe format first (current standard)
	parts := strings.SplitN(tag, "|", 2)
	if len(parts) == 2 {
		// Parse timestamp
		t, err := time.Parse(time.RFC3339Nano, parts[0])
		if err != nil {
			return 0, "", fmt.Errorf("invalid timestamp in tag: %v", err)
		}
		return t.UnixNano(), parts[1], nil
	}
	
	// Try dot format (legacy)
	parts = strings.SplitN(tag, ".", 2)
	if len(parts) == 2 {
		// Parse timestamp
		t, err := time.Parse("2006-01-02T15:04:05.999999999", parts[0])
		if err != nil {
			return 0, "", fmt.Errorf("invalid timestamp in tag: %v", err)
		}
		return t.UnixNano(), parts[1], nil
	}
	
	// Special case: Unix timestamp format (numeric only)
	if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
		if ts, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
			return ts, parts[1], nil
		}
	}
	
	return 0, "", errors.New("not a temporal tag")
}

// FormatTagWithTimestampImproved formats a tag with its timestamp (enhanced version)
func FormatTagWithTimestampImproved(tag string, timestamp int64) string {
	// Convert nanosecond timestamp to RFC3339Nano format
	t := time.Unix(0, timestamp)
	timeStr := t.Format(time.RFC3339Nano)
	return timeStr + "|" + tag
}

// GetEntityAsOf implements temporal query interface with improved error handling and timestamp parsing
func (r *TemporalRepository) GetEntityAsOf(entityID string, asOf time.Time) (*models.Entity, error) {
	logger.Debug("GetEntityAsOf: Finding entity %s as of %v", entityID, asOf)
	
	// Track stats
	queryStart := time.Now()
	atomic.AddUint64(&r.temporalStats.asOfQueries, 1)
	defer func() {
		latency := time.Since(queryStart).Nanoseconds()
		atomic.StoreInt64(&r.temporalStats.avgAsOfLatency, latency)
	}()
	
	asOfNanos := asOf.UnixNano()
	cacheKey := entityID + ":" + strconv.FormatInt(asOfNanos, 10)
	
	// Check cache
	r.temporalCache.mu.RLock()
	if cached, ok := r.temporalCache.asOfCache[cacheKey]; ok {
		r.temporalCache.mu.RUnlock()
		atomic.AddUint64(&r.temporalStats.cacheHits, 1)
		logger.Debug("GetEntityAsOf: Cache hit for %s at %v", entityID, asOf)
		return cached, nil
	}
	r.temporalCache.mu.RUnlock()
	
	atomic.AddUint64(&r.temporalStats.cacheMisses, 1)
	logger.Debug("GetEntityAsOf: Cache miss for %s at %v", entityID, asOf)
	
	// First get current entity to get the content (which isn't temporal)
	currentEntity, err := r.GetByID(entityID)
	if err != nil {
		logger.Error("GetEntityAsOf: Failed to get current entity: %v", err)
		return nil, fmt.Errorf("entity not found: %s", err)
	}
	
	// Get entity timeline
	timeline, ok := r.entityTimelines[entityID]
	if !ok {
		logger.Error("GetEntityAsOf: No timeline found for entity %s", entityID)
		return nil, fmt.Errorf("entity timeline not found: %s", entityID)
	}
	
	timeline.mu.RLock()
	defer timeline.mu.RUnlock()
	
	logger.Debug("GetEntityAsOf: Found timeline with %d timestamps for entity %s", 
		len(timeline.timestamps), entityID)
	
	// Binary search for the right timestamp
	idx := sort.Search(len(timeline.timestamps), func(i int) bool {
		return timeline.timestamps[i] > asOfNanos
	})
	
	if idx == 0 && len(timeline.timestamps) > 0 {
		// No data before this time, but entity exists
		logger.Debug("GetEntityAsOf: No data for entity %s before %v, earliest timestamp is %v", 
			entityID, asOf, time.Unix(0, timeline.timestamps[0]))
		return nil, fmt.Errorf("entity did not exist at %v", asOf)
	}
	
	// Reconstruct entity at this time
	result := &models.Entity{
		ID:        entityID,
		Tags:      make([]string, 0),
		Content:   currentEntity.Content, // Content isn't temporal, use current content
		CreatedAt: currentEntity.CreatedAt,
		UpdatedAt: currentEntity.UpdatedAt,
	}
	
	// Collect all tags up to asOf time
	for i := 0; i < idx; i++ {
		timestamp := timeline.timestamps[i]
		for _, tag := range timeline.tags[timestamp] {
			// Format with timestamp for full temporal tag
			temporalTag := FormatTagWithTimestampImproved(tag, timestamp)
			result.Tags = append(result.Tags, temporalTag)
			logger.Debug("GetEntityAsOf: Added tag %s from timestamp %v", 
				tag, time.Unix(0, timestamp))
		}
	}
	
	logger.Debug("GetEntityAsOf: Constructed historical entity with %d tags", len(result.Tags))
	
	// Cache result
	r.temporalCache.mu.Lock()
	if len(r.temporalCache.asOfCache) < r.temporalCache.maxSize {
		r.temporalCache.asOfCache[cacheKey] = result
		logger.Debug("GetEntityAsOf: Cached result for future use")
	}
	r.temporalCache.mu.Unlock()
	
	return result, nil
}

// FindEntitiesInRange finds all entities modified within a time range
func (r *TemporalRepository) FindEntitiesInRange(start, end time.Time) ([]*models.Entity, error) {
	// Track stats
	queryStart := time.Now()
	atomic.AddUint64(&r.temporalStats.rangeQueries, 1)
	defer func() {
		latency := time.Since(queryStart).Nanoseconds()
		atomic.StoreInt64(&r.temporalStats.avgRangeLatency, latency)
	}()
	
	startNanos := start.UnixNano()
	endNanos := end.UnixNano()
	
	// Find relevant buckets
	startBucket := TimeBucket{r.bucketSize}.GetBucket(startNanos)
	endBucket := TimeBucket{r.bucketSize}.GetBucket(endNanos)
	
	entityIDs := make(map[string]bool)
	
	r.bucketMu.RLock()
	defer r.bucketMu.RUnlock()
	
	// Check each bucket in range
	for bucket := startBucket; bucket <= endBucket; bucket += r.bucketSize {
		if entitySet, ok := r.bucketIndex[bucket]; ok {
			entitySet.mu.RLock()
			for entityID := range entitySet.items {
				// Verify entity actually has changes in range
				if timeline, ok := r.entityTimelines[entityID]; ok {
					timeline.mu.RLock()
					for _, ts := range timeline.timestamps {
						if ts >= startNanos && ts <= endNanos {
							entityIDs[entityID] = true
							break
						}
					}
					timeline.mu.RUnlock()
				}
			}
			entitySet.mu.RUnlock()
		}
	}
	
	// Get full entities
	results := make([]*models.Entity, 0, len(entityIDs))
	for entityID := range entityIDs {
		entity, err := r.GetByID(entityID)
		if err == nil {
			results = append(results, entity)
		}
	}
	
	return results, nil
}

// GetEntityHistory implements the interface for getting entity history with improved implementation
func (r *TemporalRepository) GetEntityHistory(entityID string, limit int) ([]*models.EntityChange, error) {
	logger.Debug("GetEntityHistory: Getting history for entity %s with limit %d", entityID, limit)
	
	// First verify entity exists
	entity, err := r.GetByID(entityID)
	if err != nil {
		logger.Error("GetEntityHistory: Entity %s not found: %v", entityID, err)
		return nil, fmt.Errorf("entity not found: %s", entityID)
	}
	
	logger.Debug("GetEntityHistory: Found entity %s with %d tags", entityID, len(entity.Tags))
	
	timeline, ok := r.entityTimelines[entityID]
	if !ok {
		logger.Error("GetEntityHistory: No timeline found for entity %s", entityID)
		return nil, fmt.Errorf("entity timeline not found: %s", entityID)
	}
	
	timeline.mu.RLock()
	defer timeline.mu.RUnlock()
	
	logger.Debug("GetEntityHistory: Found timeline with %d timestamps", len(timeline.timestamps))
	
	// Extract all timestamps from tags
	type TimestampedTag struct {
		Timestamp int64
		Tag       string
		Original  string
	}
	
	// Extract and collect all temporal tags
	tagHistory := make([]TimestampedTag, 0)
	for _, tag := range entity.Tags {
		timestamp, cleanTag, err := ParseTemporalTagImproved(tag)
		if err == nil {
			tagHistory = append(tagHistory, TimestampedTag{
				Timestamp: timestamp,
				Tag:       cleanTag,
				Original:  tag,
			})
		}
	}
	
	// Sort by timestamp (newest first)
	for i := 0; i < len(tagHistory); i++ {
		for j := i + 1; j < len(tagHistory); j++ {
			if tagHistory[i].Timestamp < tagHistory[j].Timestamp {
				tagHistory[i], tagHistory[j] = tagHistory[j], tagHistory[i]
			}
		}
	}
	
	// Create change records
	changes := make([]*models.EntityChange, 0)
	for i := 0; i < len(tagHistory) && (limit <= 0 || i < limit); i++ {
		entry := tagHistory[i]
		change := &models.EntityChange{
			Type:      "tag_added",
			Timestamp: time.Unix(0, entry.Timestamp),
			NewValue:  entry.Tag,
			EntityID:  entityID,
		}
		changes = append(changes, change)
	}
	
	logger.Debug("GetEntityHistory: Returning %d changes for entity %s", len(changes), entityID)
	return changes, nil
}

// getChangesAtTimestamp determines what changed at a specific timestamp
func (r *TemporalRepository) getChangesAtTimestamp(timeline *Timeline, index int) []string {
	if index >= len(timeline.timestamps) {
		return nil
	}
	
	timestamp := timeline.timestamps[index]
	return timeline.tags[timestamp]
}

// GetTemporalStats returns performance statistics
func (r *TemporalRepository) GetTemporalStats() map[string]interface{} {
	return map[string]interface{}{
		"asOfQueries":     atomic.LoadUint64(&r.temporalStats.asOfQueries),
		"rangeQueries":    atomic.LoadUint64(&r.temporalStats.rangeQueries),
		"cacheHits":       atomic.LoadUint64(&r.temporalStats.cacheHits),
		"cacheMisses":     atomic.LoadUint64(&r.temporalStats.cacheMisses),
		"avgAsOfLatency":  time.Duration(atomic.LoadInt64(&r.temporalStats.avgAsOfLatency)),
		"avgRangeLatency": time.Duration(atomic.LoadInt64(&r.temporalStats.avgRangeLatency)),
		"timelineEntries": len(r.entityTimelines),
		"buckets":         len(r.bucketIndex),
		"cacheSize":       len(r.temporalCache.asOfCache),
	}
}

// OptimizeForTimeRange pre-loads data for a specific time range
func (r *TemporalRepository) OptimizeForTimeRange(start, end time.Time) {
	// Pre-populate cache for this range
	entities, err := r.FindEntitiesInRange(start, end)
	if err != nil {
		return
	}
	
	// Cache multiple time points for each entity
	timePoints := []time.Time{
		start,
		start.Add((end.Sub(start)) / 2), // midpoint
		end,
	}
	
	for _, entity := range entities {
		for _, t := range timePoints {
			r.GetEntityAsOf(entity.ID, t) // This will cache the result
		}
	}
}

// GetRecentChanges returns recent changes to entities with improved implementation
func (r *TemporalRepository) GetRecentChanges(limit int) ([]*models.EntityChange, error) {
	logger.Debug("GetRecentChanges: Finding recent changes with limit %d", limit)
	
	// Collect all recent timestamps across all timelines
	type TimestampedChange struct {
		EntityID   string
		Timestamp  int64
		Tags       []string
	}
	
	allChanges := make([]TimestampedChange, 0)
	
	// Scan all timelines
	for entityID, timeline := range r.entityTimelines {
		timeline.mu.RLock()
		
		// Only look at the most recent few timestamps for efficiency
		count := 10 // Arbitrary number of recent timestamps to check per entity
		if count > len(timeline.timestamps) {
			count = len(timeline.timestamps)
		}
		
		for i := len(timeline.timestamps) - 1; i >= len(timeline.timestamps) - count && i >= 0; i-- {
			timestamp := timeline.timestamps[i]
			tags := timeline.tags[timestamp]
			
			allChanges = append(allChanges, TimestampedChange{
				EntityID:   entityID,
				Timestamp:  timestamp,
				Tags:       tags,
			})
		}
		
		timeline.mu.RUnlock()
	}
	
	// Sort changes by timestamp (newest first)
	sort.Slice(allChanges, func(i, j int) bool {
		return allChanges[i].Timestamp > allChanges[j].Timestamp
	})
	
	// Convert to EntityChange objects
	changes := make([]*models.EntityChange, 0)
	for i := 0; i < len(allChanges) && (limit <= 0 || i < limit); i++ {
		change := allChanges[i]
		
		for _, tag := range change.Tags {
			changes = append(changes, &models.EntityChange{
				Type:      "tag_added",
				Timestamp: time.Unix(0, change.Timestamp),
				NewValue:  tag,
				EntityID:  change.EntityID,
			})
			
			if limit > 0 && len(changes) >= limit {
				break
			}
		}
	}
	
	logger.Debug("GetRecentChanges: Found %d recent changes", len(changes))
	return changes, nil
}

// GetEntityDiff returns changes between two time points with improved implementation
func (r *TemporalRepository) GetEntityDiff(entityID string, t1, t2 time.Time) (*models.Entity, *models.Entity, error) {
	logger.Debug("GetEntityDiff: Comparing entity %s between %v and %v", entityID, t1, t2)
	
	// Get entity at both times
	entity1, err1 := r.GetEntityAsOf(entityID, t1)
	entity2, err2 := r.GetEntityAsOf(entityID, t2)
	
	// Log the results of the lookups
	if err1 != nil {
		logger.Debug("GetEntityDiff: Entity %s not found at %v: %v", entityID, t1, err1)
	} else {
		logger.Debug("GetEntityDiff: Entity %s at %v has %d tags", 
			entityID, t1, len(entity1.Tags))
	}
	
	if err2 != nil {
		logger.Debug("GetEntityDiff: Entity %s not found at %v: %v", entityID, t2, err2)
	} else {
		logger.Debug("GetEntityDiff: Entity %s at %v has %d tags", 
			entityID, t2, len(entity2.Tags))
	}
	
	// Handle cases where entity doesn't exist at one time
	if err1 != nil && err2 == nil {
		// Entity created between t1 and t2
		logger.Debug("GetEntityDiff: Entity %s was created between %v and %v", 
			entityID, t1, t2)
		return nil, entity2, nil
	} else if err1 == nil && err2 != nil {
		// Entity deleted between t1 and t2
		logger.Debug("GetEntityDiff: Entity %s was deleted between %v and %v", 
			entityID, t1, t2)
		return entity1, nil, nil
	} else if err1 != nil && err2 != nil {
		logger.Error("GetEntityDiff: Entity %s not found at either time point", entityID)
		return nil, nil, fmt.Errorf("entity not found at either time")
	}
	
	// Compute tag differences for logging
	tagsAdded := 0
	tagsRemoved := 0
	
	// Create maps of tags without timestamps
	tagsAt1 := make(map[string]bool)
	tagsAt2 := make(map[string]bool)
	
	for _, tag := range entity1.GetTagsWithoutTimestamp() {
		tagsAt1[tag] = true
	}
	
	for _, tag := range entity2.GetTagsWithoutTimestamp() {
		tagsAt2[tag] = true
		if !tagsAt1[tag] {
			tagsAdded++
		}
	}
	
	for tag := range tagsAt1 {
		if !tagsAt2[tag] {
			tagsRemoved++
		}
	}
	
	logger.Debug("GetEntityDiff: Found %d tags added and %d tags removed", 
		tagsAdded, tagsRemoved)
	
	return entity1, entity2, nil
}

// Create implements models.EntityRepository with temporal indexing
func (r *TemporalRepository) Create(entity *models.Entity) error {
	// Debug log
	logger.Debug("TemporalRepository.Create: Creating entity with ID %s, %d tags, content size: %d bytes", 
		entity.ID, len(entity.Tags), len(entity.Content))
	
	// Create in base repository
	err := r.HighPerformanceRepository.Create(entity)
	if err != nil {
		logger.Error("TemporalRepository.Create: Failed to create in base repository: %v", err)
		return err
	}
	
	// Index the new entity temporally
	r.indexEntityTemporal(entity)
	
	// Verify entity was created correctly
	storedEntity, err := r.GetByID(entity.ID)
	if err != nil {
		logger.Error("TemporalRepository.Create: Entity was created but cannot be retrieved: %v", err)
	} else {
		logger.Debug("TemporalRepository.Create: Entity created and verified with %d tags and %d bytes content", 
			len(storedEntity.Tags), len(storedEntity.Content))
	}
	
	return nil
}

// Update implements models.EntityRepository with temporal indexing
func (r *TemporalRepository) Update(entity *models.Entity) error {
	// Update in base repository
	err := r.HighPerformanceRepository.Update(entity)
	if err != nil {
		return err
	}
	
	// Re-index the entity temporally
	r.indexEntityTemporal(entity)
	
	return nil
}

// ListByTag implements listing with temporal support
func (r *TemporalRepository) ListByTag(tag string) ([]*models.Entity, error) {
	// Use the base repository for now
	return r.HighPerformanceRepository.ListByTag(tag)
}