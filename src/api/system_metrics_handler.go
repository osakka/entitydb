package api

import (
	"entitydb/models"
	"entitydb/logger"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// SystemMetricsHandler handles EntityDB-specific system metrics requests
type SystemMetricsHandler struct {
	entityRepo *models.RepositoryQueryWrapper
	startTime  time.Time
}

// NewSystemMetricsHandler creates a new system metrics handler
func NewSystemMetricsHandler(entityRepo models.EntityRepository) *SystemMetricsHandler {
	return &SystemMetricsHandler{
		entityRepo: models.NewRepositoryQueryWrapper(entityRepo),
		startTime:  time.Now(),
	}
}

// SystemMetricsResponse represents comprehensive system metrics
type SystemMetricsResponse struct {
	System         SystemInfo         `json:"system"`
	Database       DatabaseMetrics    `json:"database"`
	Performance    PerformanceMetrics `json:"performance"`
	Memory         MemoryMetrics      `json:"memory"`
	Storage        StorageMetrics     `json:"storage"`
	Temporal       TemporalMetrics    `json:"temporal"`
	EntityStats    EntityStats        `json:"entity_stats"`
	ActivityStats  ActivityStats      `json:"activity_stats"`
}

type SystemInfo struct {
	Version       string        `json:"version"`
	GoVersion     string        `json:"go_version"`
	Uptime        time.Duration `json:"uptime"`
	UptimeSeconds float64       `json:"uptime_seconds"`
	StartTime     time.Time     `json:"start_time"`
	CurrentTime   time.Time     `json:"current_time"`
	NumCPU        int           `json:"num_cpu"`
	NumGoroutines int           `json:"num_goroutines"`
}

type DatabaseMetrics struct {
	TotalEntities    int                    `json:"total_entities"`
	EntitiesByType   map[string]int         `json:"entities_by_type"`
	EntitiesByStatus map[string]int         `json:"entities_by_status"`
	TagsTotal        int                    `json:"tags_total"`
	TagsUnique       int                    `json:"tags_unique"`
	AvgTagsPerEntity float64               `json:"avg_tags_per_entity"`
}

type PerformanceMetrics struct {
	GCRuns          uint32        `json:"gc_runs"`
	LastGCPause     time.Duration `json:"last_gc_pause_ns"`
	TotalGCPause    time.Duration `json:"total_gc_pause_ns"`
	QueryCacheHits  int           `json:"query_cache_hits"`
	QueryCacheMiss  int           `json:"query_cache_miss"`
	IndexLookups    int64         `json:"index_lookups"`
}

type MemoryMetrics struct {
	AllocBytes      uint64 `json:"alloc_bytes"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`
	SysBytes        uint64 `json:"sys_bytes"`
	HeapAllocBytes  uint64 `json:"heap_alloc_bytes"`
	HeapSysBytes    uint64 `json:"heap_sys_bytes"`
	HeapIdleBytes   uint64 `json:"heap_idle_bytes"`
	HeapInUseBytes  uint64 `json:"heap_in_use_bytes"`
	StackInUseBytes uint64 `json:"stack_in_use_bytes"`
}

type StorageMetrics struct {
	DatabaseSizeBytes    int64 `json:"database_size_bytes"`
	WALSizeBytes        int64 `json:"wal_size_bytes"`
	IndexSizeBytes      int64 `json:"index_size_bytes"`
	TotalStorageBytes   int64 `json:"total_storage_bytes"`
	CompressionRatio    float64 `json:"compression_ratio"`
}

type TemporalMetrics struct {
	TemporalTagsCount     int     `json:"temporal_tags_count"`
	NonTemporalTagsCount  int     `json:"non_temporal_tags_count"`
	TemporalTagsRatio     float64 `json:"temporal_tags_ratio"`
	TimeRangeStart        *time.Time `json:"time_range_start"`
	TimeRangeEnd          *time.Time `json:"time_range_end"`
	AverageTimestampAge   float64 `json:"average_timestamp_age_hours"`
}

type EntityStats struct {
	CreatedToday     int            `json:"created_today"`
	CreatedThisWeek  int            `json:"created_this_week"`
	CreatedThisMonth int            `json:"created_this_month"`
	UpdatedToday     int            `json:"updated_today"`
	LargestEntity    EntitySizeInfo `json:"largest_entity"`
	SmallestEntity   EntitySizeInfo `json:"smallest_entity"`
}

type EntitySizeInfo struct {
	ID          string `json:"id"`
	SizeBytes   int    `json:"size_bytes"`
	TagCount    int    `json:"tag_count"`
	EntityType  string `json:"entity_type"`
}

type ActivityStats struct {
	RecentOperations     []OperationStat    `json:"recent_operations"`
	OperationsPerHour    map[string]int     `json:"operations_per_hour"`
	ErrorRate            float64            `json:"error_rate"`
	AverageResponseTime  float64            `json:"average_response_time_ms"`
}

type OperationStat struct {
	Timestamp   time.Time `json:"timestamp"`
	Operation   string    `json:"operation"`
	EntityID    string    `json:"entity_id"`
	EntityType  string    `json:"entity_type"`
	Duration    float64   `json:"duration_ms"`
}

// SystemMetrics returns comprehensive EntityDB system metrics
// @Summary EntityDB system metrics
// @Description Get comprehensive system metrics specific to EntityDB
// @Tags metrics
// @Accept json
// @Produce json
// @Success 200 {object} SystemMetricsResponse
// @Router /api/v1/system/metrics [get]
func (h *SystemMetricsHandler) SystemMetrics(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Handling system metrics request")
	
	startTime := time.Now()
	
	// Gather all entities for analysis
	allEntities, err := h.entityRepo.Query().Execute()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to gather entity data")
		return
	}
	
	// System info
	uptime := time.Since(h.startTime)
	systemInfo := SystemInfo{
		Version:       "2.14.0+",
		GoVersion:     runtime.Version(),
		Uptime:        uptime,
		UptimeSeconds: uptime.Seconds(),
		StartTime:     h.startTime,
		CurrentTime:   time.Now(),
		NumCPU:        runtime.NumCPU(),
		NumGoroutines: runtime.NumGoroutine(),
	}
	
	// Database metrics
	dbMetrics := h.calculateDatabaseMetrics(allEntities)
	
	// Performance metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	perfMetrics := PerformanceMetrics{
		GCRuns:       memStats.NumGC,
		LastGCPause:  time.Duration(memStats.PauseNs[(memStats.NumGC+255)%256]),
		TotalGCPause: time.Duration(memStats.PauseTotalNs),
		// Note: Query cache metrics would need to be implemented in the repository
		QueryCacheHits: 0, // Placeholder
		QueryCacheMiss: 0, // Placeholder
		IndexLookups:   0, // Placeholder
	}
	
	// Memory metrics
	memoryMetrics := MemoryMetrics{
		AllocBytes:      memStats.Alloc,
		TotalAllocBytes: memStats.TotalAlloc,
		SysBytes:        memStats.Sys,
		HeapAllocBytes:  memStats.HeapAlloc,
		HeapSysBytes:    memStats.HeapSys,
		HeapIdleBytes:   memStats.HeapIdle,
		HeapInUseBytes:  memStats.HeapInuse,
		StackInUseBytes: memStats.StackInuse,
	}
	
	// Storage metrics
	storageMetrics := h.calculateStorageMetrics()
	
	// Temporal metrics
	temporalMetrics := h.calculateTemporalMetrics(allEntities)
	
	// Entity stats
	entityStats := h.calculateEntityStats(allEntities)
	
	// Activity stats (simplified for development focus)
	activityStats := ActivityStats{
		RecentOperations:    []OperationStat{}, // Placeholder
		OperationsPerHour:   make(map[string]int),
		ErrorRate:           0.0, // Placeholder
		AverageResponseTime: 0.0, // Placeholder
	}
	
	// Build response
	response := SystemMetricsResponse{
		System:        systemInfo,
		Database:      dbMetrics,
		Performance:   perfMetrics,
		Memory:        memoryMetrics,
		Storage:       storageMetrics,
		Temporal:      temporalMetrics,
		EntityStats:   entityStats,
		ActivityStats: activityStats,
	}
	
	logger.Debug("System metrics collected in %v", time.Since(startTime))
	RespondJSON(w, http.StatusOK, response)
}

func (h *SystemMetricsHandler) calculateDatabaseMetrics(entities []*models.Entity) DatabaseMetrics {
	entitiesByType := make(map[string]int)
	entitiesByStatus := make(map[string]int)
	totalTags := 0
	uniqueTags := make(map[string]bool)
	
	for _, entity := range entities {
		// Count tags
		totalTags += len(entity.Tags)
		
		// Track unique tags and analyze
		entityType := "unknown"
		entityStatus := "unknown"
		
		for _, tag := range entity.Tags {
			uniqueTags[tag] = true
			
			// Extract actual tag from temporal format
			tagPart := tag
			if strings.Contains(tag, "|") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					tagPart = parts[1]
				}
			}
			
			if strings.HasPrefix(tagPart, "type:") {
				entityType = strings.TrimPrefix(tagPart, "type:")
			} else if strings.HasPrefix(tagPart, "status:") {
				entityStatus = strings.TrimPrefix(tagPart, "status:")
			}
		}
		
		entitiesByType[entityType]++
		entitiesByStatus[entityStatus]++
	}
	
	avgTags := 0.0
	if len(entities) > 0 {
		avgTags = float64(totalTags) / float64(len(entities))
	}
	
	return DatabaseMetrics{
		TotalEntities:    len(entities),
		EntitiesByType:   entitiesByType,
		EntitiesByStatus: entitiesByStatus,
		TagsTotal:        totalTags,
		TagsUnique:       len(uniqueTags),
		AvgTagsPerEntity: avgTags,
	}
}

func (h *SystemMetricsHandler) calculateStorageMetrics() StorageMetrics {
	var dbSize, walSize, indexSize int64
	
	// Database file size
	if stat, err := os.Stat("/opt/entitydb/var/entities.ebf"); err == nil {
		dbSize = stat.Size()
	}
	
	// WAL file size
	if stat, err := os.Stat("/opt/entitydb/var/entitydb.wal"); err == nil {
		walSize = stat.Size()
	}
	
	// Index size (estimated)
	indexSize = dbSize / 10 // Rough estimate for development
	
	total := dbSize + walSize + indexSize
	compressionRatio := 1.0 // Placeholder
	
	return StorageMetrics{
		DatabaseSizeBytes: dbSize,
		WALSizeBytes:     walSize,
		IndexSizeBytes:   indexSize,
		TotalStorageBytes: total,
		CompressionRatio: compressionRatio,
	}
}

func (h *SystemMetricsHandler) calculateTemporalMetrics(entities []*models.Entity) TemporalMetrics {
	temporalCount := 0
	nonTemporalCount := 0
	var earliest, latest *time.Time
	var totalAge time.Duration
	ageCount := 0
	
	for _, entity := range entities {
		for _, tag := range entity.Tags {
			if strings.Contains(tag, "|") {
				// Temporal tag
				temporalCount++
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					if timestamp, err := time.Parse(time.RFC3339Nano, parts[0]); err == nil {
						if earliest == nil || timestamp.Before(*earliest) {
							earliest = &timestamp
						}
						if latest == nil || timestamp.After(*latest) {
							latest = &timestamp
						}
						totalAge += time.Since(timestamp)
						ageCount++
					}
				}
			} else {
				// Non-temporal tag
				nonTemporalCount++
			}
		}
	}
	
	ratio := 0.0
	if temporalCount + nonTemporalCount > 0 {
		ratio = float64(temporalCount) / float64(temporalCount + nonTemporalCount)
	}
	
	avgAge := 0.0
	if ageCount > 0 {
		avgAge = totalAge.Hours() / float64(ageCount)
	}
	
	return TemporalMetrics{
		TemporalTagsCount:    temporalCount,
		NonTemporalTagsCount: nonTemporalCount,
		TemporalTagsRatio:    ratio,
		TimeRangeStart:       earliest,
		TimeRangeEnd:         latest,
		AverageTimestampAge:  avgAge,
	}
}

func (h *SystemMetricsHandler) calculateEntityStats(entities []*models.Entity) EntityStats {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	weekAgo := today.Add(-7 * 24 * time.Hour)
	monthAgo := today.Add(-30 * 24 * time.Hour)
	
	var createdToday, createdWeek, createdMonth, updatedToday int
	var largest, smallest *EntitySizeInfo
	
	for _, entity := range entities {
		// Parse creation time
		if createdAt, err := time.Parse(time.RFC3339, entity.CreatedAt); err == nil {
			if createdAt.After(today) {
				createdToday++
			}
			if createdAt.After(weekAgo) {
				createdWeek++
			}
			if createdAt.After(monthAgo) {
				createdMonth++
			}
		}
		
		// Parse update time
		if updatedAt, err := time.Parse(time.RFC3339, entity.UpdatedAt); err == nil {
			if updatedAt.After(today) {
				updatedToday++
			}
		}
		
		// Calculate entity size
		entitySize := len(entity.Content) + len(strings.Join(entity.Tags, ""))
		entityType := "unknown"
		for _, tag := range entity.Tags {
			tagPart := tag
			if strings.Contains(tag, "|") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					tagPart = parts[1]
				}
			}
			if strings.HasPrefix(tagPart, "type:") {
				entityType = strings.TrimPrefix(tagPart, "type:")
				break
			}
		}
		
		entityInfo := EntitySizeInfo{
			ID:         entity.ID,
			SizeBytes:  entitySize,
			TagCount:   len(entity.Tags),
			EntityType: entityType,
		}
		
		if largest == nil || entitySize > largest.SizeBytes {
			largest = &entityInfo
		}
		if smallest == nil || entitySize < smallest.SizeBytes {
			smallest = &entityInfo
		}
	}
	
	stats := EntityStats{
		CreatedToday:     createdToday,
		CreatedThisWeek:  createdWeek,
		CreatedThisMonth: createdMonth,
		UpdatedToday:     updatedToday,
	}
	
	if largest != nil {
		stats.LargestEntity = *largest
	}
	if smallest != nil {
		stats.SmallestEntity = *smallest
	}
	
	return stats
}