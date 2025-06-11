package api

import (
	"encoding/json"
	"math"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
)

// ComprehensiveMetricsHandler provides all metrics for 70T scale monitoring
type ComprehensiveMetricsHandler struct {
	repo            models.EntityRepository
	startTime       time.Time
	
	// Operation metrics
	opMetrics       *OperationMetricsCollector
	
	// Storage metrics  
	storageCollector *StorageMetricsCollector
	
	// Cache metrics
	cacheCollector  *CacheMetricsCollector
	
	// Index metrics
	indexCollector  *IndexMetricsCollector
	
	// RBAC metrics
	rbacCollector   *RBACMetricsCollector
	
	// Temporal metrics
	temporalCollector *TemporalMetricsCollector
}

// OperationMetricsCollector tracks all entity operations
type OperationMetricsCollector struct {
	creates   *OperationTracker
	reads     *OperationTracker
	updates   *OperationTracker
	deletes   *OperationTracker
	queries   *OperationTracker
}

// OperationTracker tracks metrics for a specific operation
type OperationTracker struct {
	count      int64
	errors     int64
	totalTime  int64
	minTime    int64
	maxTime    int64
	
	mu         sync.Mutex
	samples    []int64
	sampleIdx  int
}

// StorageMetricsCollector tracks storage layer metrics
type StorageMetricsCollector struct {
	totalEntities    int64
	totalSize        int64
	compressionRatio float64
	walSize          int64
	walOps           int64
	checkpoints      int64
	lastCheckpoint   time.Time
	diskReads        int64
	diskWrites       int64
	
	mu              sync.RWMutex
}

// CacheMetricsCollector tracks cache performance
type CacheMetricsCollector struct {
	entityHits      int64
	entityMisses    int64
	entitySize      int64
	entityEvictions int64
	
	queryHits       int64
	queryMisses     int64
	querySize       int64
	
	bufferHits      int64
	bufferMisses    int64
	bufferSize      int64
}

// IndexMetricsCollector tracks index performance
type IndexMetricsCollector struct {
	btreeDepth      int
	btreeNodes      int64
	btreeSearches   int64
	btreeInserts    int64
	btreeDeletes    int64
	
	bloomSize       int64
	bloomHits       int64
	bloomMisses     int64
	bloomFPRate     float64
	
	skiplistLevels  int
	skiplistNodes   int64
	skiplistTime    int64
	
	tagIndexSize    int64
	tagIndexEntries int64
	namespaces      int64
}

// RBACMetricsCollector tracks auth and permissions
type RBACMetricsCollector struct {
	activeSessions   int64
	totalLogins      int64
	failedLogins     int64
	permChecks       int64
	permDenials      int64
	permCacheHits    int64
	
	mu              sync.RWMutex
	sessionsByRole  map[string]int64
	topUsers        []UserActivity
	securityEvents  []SecurityEvent
}

// TemporalMetricsCollector tracks temporal query metrics
type TemporalMetricsCollector struct {
	asOfQueries     int64
	historyQueries  int64
	diffQueries     int64
	changesQueries  int64
	
	timelineDepth   int64
	temporalDensity float64
	oldestTime      int64
	newestTime      int64
	totalTempTags   int64
	
	retentionDeleted int64
	retentionReclaimed int64
	
	mu             sync.RWMutex
	hotRanges      []HotTimeRange
}

// UserActivity represents user activity metrics
type UserActivity struct {
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	Operations int64  `json:"operations"`
	LastActive int64  `json:"last_active"`
}

// HotTimeRange represents frequently accessed time ranges
type HotTimeRange struct {
	Start      int64  `json:"start"`
	End        int64  `json:"end"`
	Queries    int64  `json:"queries"`
	Label      string `json:"label"`
}

// NewComprehensiveMetricsHandler creates a new comprehensive metrics handler
func NewComprehensiveMetricsHandler(repo models.EntityRepository) *ComprehensiveMetricsHandler {
	return &ComprehensiveMetricsHandler{
		repo:      repo,
		startTime: time.Now(),
		
		opMetrics: &OperationMetricsCollector{
			creates: newOperationTracker(),
			reads:   newOperationTracker(),
			updates: newOperationTracker(),
			deletes: newOperationTracker(),
			queries: newOperationTracker(),
		},
		
		storageCollector:  &StorageMetricsCollector{},
		cacheCollector:    &CacheMetricsCollector{},
		indexCollector:    &IndexMetricsCollector{},
		rbacCollector:     &RBACMetricsCollector{
			sessionsByRole: make(map[string]int64),
			topUsers:       make([]UserActivity, 0),
			securityEvents: make([]SecurityEvent, 0),
		},
		temporalCollector: &TemporalMetricsCollector{
			hotRanges: make([]HotTimeRange, 0),
		},
	}
}

func newOperationTracker() *OperationTracker {
	return &OperationTracker{
		minTime: int64(^uint64(0) >> 1), // MaxInt64
		samples: make([]int64, 1000),
	}
}

// ServeHTTP handles the comprehensive metrics endpoint
func (h *ComprehensiveMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics := h.collectAllMetrics()
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		logger.Error("Failed to encode metrics: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// collectAllMetrics gathers all system metrics
func (h *ComprehensiveMetricsHandler) collectAllMetrics() map[string]interface{} {
	// Get runtime metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Calculate uptime
	uptime := time.Since(h.startTime)
	
	// Get repository stats if it's a binary repository
	var repoStats map[string]interface{}
	if binaryRepo, ok := h.repo.(*binary.EntityRepository); ok {
		repoStats = h.collectRepositoryStats(binaryRepo)
	}
	
	return map[string]interface{}{
		"timestamp": time.Now().UnixNano(),
		"uptime_ns": uptime.Nanoseconds(),
		
		// Operation metrics
		"operations": map[string]interface{}{
			"create": h.getOperationStats(h.opMetrics.creates),
			"read":   h.getOperationStats(h.opMetrics.reads),
			"update": h.getOperationStats(h.opMetrics.updates),
			"delete": h.getOperationStats(h.opMetrics.deletes),
			"query":  h.getOperationStats(h.opMetrics.queries),
			"total_rate_per_sec": h.getTotalOperationRate(),
		},
		
		// Storage metrics
		"storage": h.getStorageMetrics(repoStats),
		
		// Cache metrics
		"cache": h.getCacheMetrics(),
		
		// Index metrics
		"indexing": h.getIndexMetrics(repoStats),
		
		// RBAC metrics
		"rbac": h.getRBACMetrics(),
		
		// Temporal metrics
		"temporal": h.getTemporalMetrics(),
		
		// System metrics
		"system": map[string]interface{}{
			"goroutines":     runtime.NumGoroutine(),
			"cpu_cores":      runtime.NumCPU(),
			"memory_alloc":   m.Alloc,
			"memory_sys":     m.Sys,
			"memory_heap":    m.HeapAlloc,
			"gc_runs":        m.NumGC,
			"gc_pause_ns":    m.PauseTotalNs,
		},
		
		// Health indicators
		"health": h.getHealthIndicators(),
	}
}

// getOperationStats returns stats for an operation
func (h *ComprehensiveMetricsHandler) getOperationStats(op *OperationTracker) map[string]interface{} {
	op.mu.Lock()
	defer op.mu.Unlock()
	
	count := atomic.LoadInt64(&op.count)
	errors := atomic.LoadInt64(&op.errors)
	
	avgTime := int64(0)
	if count > 0 {
		avgTime = atomic.LoadInt64(&op.totalTime) / count
	}
	
	// Calculate percentiles
	percentiles := h.calculatePercentiles(op.samples, op.sampleIdx)
	
	return map[string]interface{}{
		"count":       count,
		"errors":      errors,
		"error_rate":  float64(errors) / float64(maxInt64(count, 1)),
		"avg_time_ns": avgTime,
		"min_time_ns": atomic.LoadInt64(&op.minTime),
		"max_time_ns": atomic.LoadInt64(&op.maxTime),
		"p50_ns":      percentiles["p50"],
		"p95_ns":      percentiles["p95"],
		"p99_ns":      percentiles["p99"],
		"p999_ns":     percentiles["p999"],
	}
}

// calculatePercentiles calculates percentiles from samples
func (h *ComprehensiveMetricsHandler) calculatePercentiles(samples []int64, idx int) map[string]int64 {
	// Copy valid samples
	valid := make([]int64, 0, len(samples))
	for i := 0; i < len(samples); i++ {
		if samples[i] > 0 {
			valid = append(valid, samples[i])
		}
	}
	
	if len(valid) == 0 {
		return map[string]int64{
			"p50": 0, "p95": 0, "p99": 0, "p999": 0,
		}
	}
	
	sort.Slice(valid, func(i, j int) bool {
		return valid[i] < valid[j]
	})
	
	return map[string]int64{
		"p50":  valid[len(valid)*50/100],
		"p95":  valid[len(valid)*95/100],
		"p99":  valid[len(valid)*99/100],
		"p999": valid[minInt(len(valid)*999/1000, len(valid)-1)],
	}
}

// getTotalOperationRate calculates total operations per second
func (h *ComprehensiveMetricsHandler) getTotalOperationRate() float64 {
	uptime := time.Since(h.startTime).Seconds()
	if uptime == 0 {
		return 0
	}
	
	total := atomic.LoadInt64(&h.opMetrics.creates.count) +
		atomic.LoadInt64(&h.opMetrics.reads.count) +
		atomic.LoadInt64(&h.opMetrics.updates.count) +
		atomic.LoadInt64(&h.opMetrics.deletes.count) +
		atomic.LoadInt64(&h.opMetrics.queries.count)
	
	return float64(total) / uptime
}

// getStorageMetrics returns storage layer metrics
func (h *ComprehensiveMetricsHandler) getStorageMetrics(repoStats map[string]interface{}) map[string]interface{} {
	h.storageCollector.mu.RLock()
	defer h.storageCollector.mu.RUnlock()
	
	metrics := map[string]interface{}{
		"total_entities":     h.storageCollector.totalEntities,
		"total_size_bytes":   h.storageCollector.totalSize,
		"compression_ratio":  h.storageCollector.compressionRatio,
		"wal_size_bytes":     h.storageCollector.walSize,
		"wal_operations":     h.storageCollector.walOps,
		"checkpoints":        h.storageCollector.checkpoints,
		"last_checkpoint":    h.storageCollector.lastCheckpoint.UnixNano(),
		"disk_read_bytes":    h.storageCollector.diskReads,
		"disk_write_bytes":   h.storageCollector.diskWrites,
	}
	
	// Add repository-specific stats
	if repoStats != nil {
		for k, v := range repoStats {
			metrics[k] = v
		}
	}
	
	return metrics
}

// getCacheMetrics returns cache performance metrics
func (h *ComprehensiveMetricsHandler) getCacheMetrics() map[string]interface{} {
	entityHits := atomic.LoadInt64(&h.cacheCollector.entityHits)
	entityMisses := atomic.LoadInt64(&h.cacheCollector.entityMisses)
	entityTotal := entityHits + entityMisses
	
	queryHits := atomic.LoadInt64(&h.cacheCollector.queryHits)
	queryMisses := atomic.LoadInt64(&h.cacheCollector.queryMisses)
	queryTotal := queryHits + queryMisses
	
	return map[string]interface{}{
		"entity_cache": map[string]interface{}{
			"hits":       entityHits,
			"misses":     entityMisses,
			"hit_rate":   float64(entityHits) / float64(maxInt64(entityTotal, 1)),
			"size":       atomic.LoadInt64(&h.cacheCollector.entitySize),
			"evictions":  atomic.LoadInt64(&h.cacheCollector.entityEvictions),
		},
		"query_cache": map[string]interface{}{
			"hits":       queryHits,
			"misses":     queryMisses,
			"hit_rate":   float64(queryHits) / float64(maxInt64(queryTotal, 1)),
			"size":       atomic.LoadInt64(&h.cacheCollector.querySize),
		},
		"buffer_pool": map[string]interface{}{
			"hits":       atomic.LoadInt64(&h.cacheCollector.bufferHits),
			"misses":     atomic.LoadInt64(&h.cacheCollector.bufferMisses),
			"size_bytes": atomic.LoadInt64(&h.cacheCollector.bufferSize),
		},
	}
}

// getIndexMetrics returns index performance metrics
func (h *ComprehensiveMetricsHandler) getIndexMetrics(repoStats map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"btree": map[string]interface{}{
			"depth":    h.indexCollector.btreeDepth,
			"nodes":    atomic.LoadInt64(&h.indexCollector.btreeNodes),
			"searches": atomic.LoadInt64(&h.indexCollector.btreeSearches),
			"inserts":  atomic.LoadInt64(&h.indexCollector.btreeInserts),
			"deletes":  atomic.LoadInt64(&h.indexCollector.btreeDeletes),
		},
		"bloom_filter": map[string]interface{}{
			"size_bytes":          atomic.LoadInt64(&h.indexCollector.bloomSize),
			"hits":                atomic.LoadInt64(&h.indexCollector.bloomHits),
			"misses":              atomic.LoadInt64(&h.indexCollector.bloomMisses),
			"false_positive_rate": h.indexCollector.bloomFPRate,
		},
		"skiplist": map[string]interface{}{
			"levels":          h.indexCollector.skiplistLevels,
			"nodes":           atomic.LoadInt64(&h.indexCollector.skiplistNodes),
			"avg_search_time": atomic.LoadInt64(&h.indexCollector.skiplistTime),
		},
		"tag_index": map[string]interface{}{
			"size":       atomic.LoadInt64(&h.indexCollector.tagIndexSize),
			"entries":    atomic.LoadInt64(&h.indexCollector.tagIndexEntries),
			"namespaces": atomic.LoadInt64(&h.indexCollector.namespaces),
		},
	}
}

// getRBACMetrics returns RBAC and security metrics
func (h *ComprehensiveMetricsHandler) getRBACMetrics() map[string]interface{} {
	h.rbacCollector.mu.RLock()
	defer h.rbacCollector.mu.RUnlock()
	
	return map[string]interface{}{
		"sessions": map[string]interface{}{
			"active":       atomic.LoadInt64(&h.rbacCollector.activeSessions),
			"total_logins": atomic.LoadInt64(&h.rbacCollector.totalLogins),
			"failed_logins": atomic.LoadInt64(&h.rbacCollector.failedLogins),
			"by_role":      h.rbacCollector.sessionsByRole,
		},
		"permissions": map[string]interface{}{
			"checks":     atomic.LoadInt64(&h.rbacCollector.permChecks),
			"denials":    atomic.LoadInt64(&h.rbacCollector.permDenials),
			"cache_hits": atomic.LoadInt64(&h.rbacCollector.permCacheHits),
		},
		"top_users":       h.rbacCollector.topUsers,
		"security_events": h.rbacCollector.securityEvents,
	}
}

// getTemporalMetrics returns temporal query metrics
func (h *ComprehensiveMetricsHandler) getTemporalMetrics() map[string]interface{} {
	h.temporalCollector.mu.RLock()
	defer h.temporalCollector.mu.RUnlock()
	
	return map[string]interface{}{
		"queries": map[string]interface{}{
			"as_of":   atomic.LoadInt64(&h.temporalCollector.asOfQueries),
			"history": atomic.LoadInt64(&h.temporalCollector.historyQueries),
			"diff":    atomic.LoadInt64(&h.temporalCollector.diffQueries),
			"changes": atomic.LoadInt64(&h.temporalCollector.changesQueries),
		},
		"timeline": map[string]interface{}{
			"avg_depth":       atomic.LoadInt64(&h.temporalCollector.timelineDepth),
			"temporal_density": h.temporalCollector.temporalDensity,
			"oldest_ns":       atomic.LoadInt64(&h.temporalCollector.oldestTime),
			"newest_ns":       atomic.LoadInt64(&h.temporalCollector.newestTime),
			"total_tags":      atomic.LoadInt64(&h.temporalCollector.totalTempTags),
		},
		"retention": map[string]interface{}{
			"deleted_tags":     atomic.LoadInt64(&h.temporalCollector.retentionDeleted),
			"reclaimed_bytes":  atomic.LoadInt64(&h.temporalCollector.retentionReclaimed),
		},
		"hot_ranges": h.temporalCollector.hotRanges,
	}
}

// getHealthIndicators returns system health indicators
func (h *ComprehensiveMetricsHandler) getHealthIndicators() map[string]interface{} {
	// Calculate health scores
	opsHealth := h.calculateOpsHealth()
	cacheHealth := h.calculateCacheHealth()
	storageHealth := h.calculateStorageHealth()
	
	overallHealth := (opsHealth + cacheHealth + storageHealth) / 3
	
	return map[string]interface{}{
		"overall_score":     overallHealth,
		"operations_score":  opsHealth,
		"cache_score":       cacheHealth,
		"storage_score":     storageHealth,
		"status":            h.getHealthStatus(overallHealth),
		"recommendations":   h.getHealthRecommendations(overallHealth, opsHealth, cacheHealth, storageHealth),
	}
}

// calculateOpsHealth calculates operations health score (0-100)
func (h *ComprehensiveMetricsHandler) calculateOpsHealth() float64 {
	errorRate := 0.0
	totalOps := float64(0)
	
	for _, op := range []*OperationTracker{
		h.opMetrics.creates,
		h.opMetrics.reads,
		h.opMetrics.updates,
		h.opMetrics.deletes,
		h.opMetrics.queries,
	} {
		count := float64(atomic.LoadInt64(&op.count))
		errors := float64(atomic.LoadInt64(&op.errors))
		totalOps += count
		if count > 0 {
			errorRate += (errors / count) * count
		}
	}
	
	if totalOps > 0 {
		errorRate = errorRate / totalOps
	}
	
	// Score based on error rate (100 = 0% errors, 0 = 100% errors)
	return math.Max(0, 100*(1-errorRate))
}

// calculateCacheHealth calculates cache health score (0-100)
func (h *ComprehensiveMetricsHandler) calculateCacheHealth() float64 {
	entityHits := float64(atomic.LoadInt64(&h.cacheCollector.entityHits))
	entityMisses := float64(atomic.LoadInt64(&h.cacheCollector.entityMisses))
	
	if entityHits+entityMisses == 0 {
		return 100 // No cache usage yet
	}
	
	hitRate := entityHits / (entityHits + entityMisses)
	return hitRate * 100
}

// calculateStorageHealth calculates storage health score (0-100)
func (h *ComprehensiveMetricsHandler) calculateStorageHealth() float64 {
	// Simple health based on compression ratio
	ratio := h.storageCollector.compressionRatio
	if ratio > 4 {
		return 100
	} else if ratio > 3 {
		return 90
	} else if ratio > 2 {
		return 80
	} else if ratio > 1 {
		return 70
	}
	return 60
}

// getHealthStatus returns health status string
func (h *ComprehensiveMetricsHandler) getHealthStatus(score float64) string {
	if score >= 95 {
		return "excellent"
	} else if score >= 85 {
		return "good"
	} else if score >= 70 {
		return "fair"
	} else if score >= 50 {
		return "poor"
	}
	return "critical"
}

// getHealthRecommendations returns health recommendations
func (h *ComprehensiveMetricsHandler) getHealthRecommendations(overall, ops, cache, storage float64) []string {
	recommendations := make([]string, 0)
	
	if ops < 80 {
		recommendations = append(recommendations, "High error rate detected - investigate failed operations")
	}
	
	if cache < 70 {
		recommendations = append(recommendations, "Low cache hit rate - consider increasing cache size")
	}
	
	if storage < 80 {
		recommendations = append(recommendations, "Low compression ratio - review entity content patterns")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "System operating within normal parameters")
	}
	
	return recommendations
}

// collectRepositoryStats collects stats from binary repository
func (h *ComprehensiveMetricsHandler) collectRepositoryStats(repo *binary.EntityRepository) map[string]interface{} {
	// This would interface with the actual binary repository
	// For now, return placeholder stats
	return map[string]interface{}{
		"journal_size":     1024 * 1024 * 100, // 100MB
		"index_size":       1024 * 1024 * 50,  // 50MB
		"data_files":       42,
		"fragmentation":    0.15,
	}
}

// RecordOperation records an operation for metrics
func (h *ComprehensiveMetricsHandler) RecordOperation(opType string, duration time.Duration, err error) {
	var tracker *OperationTracker
	
	switch opType {
	case "create":
		tracker = h.opMetrics.creates
	case "read":
		tracker = h.opMetrics.reads
	case "update":
		tracker = h.opMetrics.updates
	case "delete":
		tracker = h.opMetrics.deletes
	case "query":
		tracker = h.opMetrics.queries
	default:
		return
	}
	
	durationNs := duration.Nanoseconds()
	
	// Update counters
	atomic.AddInt64(&tracker.count, 1)
	if err != nil {
		atomic.AddInt64(&tracker.errors, 1)
	}
	
	// Update times
	atomic.AddInt64(&tracker.totalTime, durationNs)
	
	// Update min/max
	for {
		min := atomic.LoadInt64(&tracker.minTime)
		if durationNs >= min || atomic.CompareAndSwapInt64(&tracker.minTime, min, durationNs) {
			break
		}
	}
	
	for {
		max := atomic.LoadInt64(&tracker.maxTime)
		if durationNs <= max || atomic.CompareAndSwapInt64(&tracker.maxTime, max, durationNs) {
			break
		}
	}
	
	// Store sample
	tracker.mu.Lock()
	tracker.samples[tracker.sampleIdx] = durationNs
	tracker.sampleIdx = (tracker.sampleIdx + 1) % len(tracker.samples)
	tracker.mu.Unlock()
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}