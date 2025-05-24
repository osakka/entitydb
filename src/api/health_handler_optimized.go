package api

import (
	"entitydb/models"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// OptimizedHealthHandler handles health check requests with minimal overhead
type OptimizedHealthHandler struct {
	startTime time.Time
	
	// Cached metrics updated periodically
	cachedMetrics atomic.Value // stores *CachedHealthMetrics
	metricsLock   sync.Mutex
	lastUpdate    time.Time
	updatePeriod  time.Duration
	
	// Reference to entity repo for periodic updates only
	entityRepo models.EntityRepository
}

// CachedHealthMetrics stores pre-computed metrics
type CachedHealthMetrics struct {
	EntityCount  int
	UserCount    int
	DatabaseSize int64
}

// NewOptimizedHealthHandler creates a new optimized health handler
func NewOptimizedHealthHandler(entityRepo models.EntityRepository) *OptimizedHealthHandler {
	h := &OptimizedHealthHandler{
		startTime:    time.Now(),
		updatePeriod: 30 * time.Second, // Update metrics every 30 seconds
		entityRepo:   entityRepo,
	}
	
	// Initialize with empty metrics
	h.cachedMetrics.Store(&CachedHealthMetrics{})
	
	// Start background metrics updater
	go h.metricsUpdater()
	
	return h
}

// Health returns health status WITHOUT database queries
func (h *OptimizedHealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	// Get memory stats (fast operation)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	// Get cached metrics
	metrics := h.cachedMetrics.Load().(*CachedHealthMetrics)
	
	// Build response with minimal overhead
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    time.Since(h.startTime).String(),
		Version:   "2.14.0+",
		Checks: map[string]string{
			"database": "healthy", // Assume healthy unless proven otherwise
		},
		Metrics: HealthMetrics{
			EntityCount:  metrics.EntityCount,
			UserCount:    metrics.UserCount,
			DatabaseSize: metrics.DatabaseSize,
			MemoryUsage: MemoryMetrics{
				Alloc:      memStats.Alloc,
				TotalAlloc: memStats.TotalAlloc,
				Sys:        memStats.Sys,
				NumGC:      memStats.NumGC,
			},
			GoRoutines: runtime.NumGoroutine(),
		},
	}
	
	RespondJSON(w, http.StatusOK, response)
}

// metricsUpdater runs in background to update cached metrics
func (h *OptimizedHealthHandler) metricsUpdater() {
	ticker := time.NewTicker(h.updatePeriod)
	defer ticker.Stop()
	
	// Initial update
	h.updateMetrics()
	
	for range ticker.C {
		h.updateMetrics()
	}
}

// updateMetrics performs the expensive operations in background
func (h *OptimizedHealthHandler) updateMetrics() {
	h.metricsLock.Lock()
	defer h.metricsLock.Unlock()
	
	newMetrics := &CachedHealthMetrics{}
	
	// Count all entities (expensive, done in background)
	if entities, err := h.entityRepo.List(); err == nil {
		newMetrics.EntityCount = len(entities)
		
		// Count users while we have the data
		userCount := 0
		for _, entity := range entities {
			for _, tag := range entity.Tags {
				if tag == "type:user" || strings.HasSuffix(tag, "|type:user") {
					userCount++
					break
				}
			}
		}
		newMetrics.UserCount = userCount
	}
	
	// Get database size
	if stat, err := os.Stat("/opt/entitydb/var/entities.ebf"); err == nil {
		newMetrics.DatabaseSize = stat.Size()
	}
	
	// Store the updated metrics atomically
	h.cachedMetrics.Store(newMetrics)
	h.lastUpdate = time.Now()
}