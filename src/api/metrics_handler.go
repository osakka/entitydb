package api

import (
	"entitydb/models"
	"entitydb/logger"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// MetricsHandler handles Prometheus-style metrics requests
type MetricsHandler struct {
	entityRepo *models.RepositoryQueryWrapper
	startTime  time.Time
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(entityRepo models.EntityRepository) *MetricsHandler {
	return &MetricsHandler{
		entityRepo: models.NewRepositoryQueryWrapper(entityRepo),
		startTime:  time.Now(),
	}
}

// PrometheusMetrics returns Prometheus-compatible metrics
// @Summary Prometheus metrics
// @Description Get system metrics in Prometheus format
// @Tags metrics
// @Produce text/plain
// @Success 200 {string} string "Prometheus metrics"
// @Router /metrics [get]
func (h *MetricsHandler) PrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Handling Prometheus metrics request")
	
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	
	var metrics strings.Builder
	
	// Uptime metric
	uptime := time.Since(h.startTime).Seconds()
	metrics.WriteString("# HELP entitydb_uptime_seconds Time since server started\n")
	metrics.WriteString("# TYPE entitydb_uptime_seconds counter\n")
	metrics.WriteString(fmt.Sprintf("entitydb_uptime_seconds %.2f\n", uptime))
	metrics.WriteString("\n")
	
	// Entity count metrics
	allEntities, err := h.entityRepo.Query().Execute()
	entityCount := 0
	if err == nil {
		entityCount = len(allEntities)
	}
	
	metrics.WriteString("# HELP entitydb_entities_total Total number of entities\n")
	metrics.WriteString("# TYPE entitydb_entities_total gauge\n")
	metrics.WriteString(fmt.Sprintf("entitydb_entities_total %d\n", entityCount))
	metrics.WriteString("\n")
	
	// Entity count by type
	entityByType := make(map[string]int)
	for _, entity := range allEntities {
		entityType := "unknown"
		for _, tag := range entity.Tags {
			// Extract type from temporal tags (handle timestamp|tag format)
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
		entityByType[entityType]++
	}
	
	metrics.WriteString("# HELP entitydb_entities_by_type_total Number of entities by type\n")
	metrics.WriteString("# TYPE entitydb_entities_by_type_total gauge\n")
	for entityType, count := range entityByType {
		metrics.WriteString(fmt.Sprintf("entitydb_entities_by_type_total{type=\"%s\"} %d\n", entityType, count))
	}
	metrics.WriteString("\n")
	
	// Database size
	var dbSize int64
	if stat, err := os.Stat("/opt/entitydb/var/entities.ebf"); err == nil {
		dbSize = stat.Size()
	}
	
	metrics.WriteString("# HELP entitydb_database_size_bytes Size of database file in bytes\n")
	metrics.WriteString("# TYPE entitydb_database_size_bytes gauge\n")
	metrics.WriteString(fmt.Sprintf("entitydb_database_size_bytes %d\n", dbSize))
	metrics.WriteString("\n")
	
	// Memory metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	metrics.WriteString("# HELP entitydb_memory_alloc_bytes Currently allocated memory in bytes\n")
	metrics.WriteString("# TYPE entitydb_memory_alloc_bytes gauge\n")
	metrics.WriteString(fmt.Sprintf("entitydb_memory_alloc_bytes %d\n", memStats.Alloc))
	metrics.WriteString("\n")
	
	metrics.WriteString("# HELP entitydb_memory_sys_bytes Total memory obtained from OS in bytes\n")
	metrics.WriteString("# TYPE entitydb_memory_sys_bytes gauge\n")
	metrics.WriteString(fmt.Sprintf("entitydb_memory_sys_bytes %d\n", memStats.Sys))
	metrics.WriteString("\n")
	
	metrics.WriteString("# HELP entitydb_memory_heap_alloc_bytes Heap memory allocated in bytes\n")
	metrics.WriteString("# TYPE entitydb_memory_heap_alloc_bytes gauge\n")
	metrics.WriteString(fmt.Sprintf("entitydb_memory_heap_alloc_bytes %d\n", memStats.HeapAlloc))
	metrics.WriteString("\n")
	
	// Goroutine count
	metrics.WriteString("# HELP entitydb_goroutines_total Number of active goroutines\n")
	metrics.WriteString("# TYPE entitydb_goroutines_total gauge\n")
	metrics.WriteString(fmt.Sprintf("entitydb_goroutines_total %d\n", runtime.NumGoroutine()))
	metrics.WriteString("\n")
	
	// GC metrics
	metrics.WriteString("# HELP entitydb_gc_runs_total Total number of GC runs\n")
	metrics.WriteString("# TYPE entitydb_gc_runs_total counter\n")
	metrics.WriteString(fmt.Sprintf("entitydb_gc_runs_total %d\n", memStats.NumGC))
	metrics.WriteString("\n")
	
	// WAL file size (if exists)
	var walSize int64
	if stat, err := os.Stat("/opt/entitydb/var/entitydb.wal"); err == nil {
		walSize = stat.Size()
	}
	
	metrics.WriteString("# HELP entitydb_wal_size_bytes Size of WAL file in bytes\n")
	metrics.WriteString("# TYPE entitydb_wal_size_bytes gauge\n")
	metrics.WriteString(fmt.Sprintf("entitydb_wal_size_bytes %d\n", walSize))
	metrics.WriteString("\n")
	
	// Version info
	metrics.WriteString("# HELP entitydb_info Information about EntityDB server\n")
	metrics.WriteString("# TYPE entitydb_info gauge\n")
	metrics.WriteString("entitydb_info{version=\"2.14.0+\",go_version=\"" + runtime.Version() + "\"} 1\n")
	metrics.WriteString("\n")
	
	w.Write([]byte(metrics.String()))
}