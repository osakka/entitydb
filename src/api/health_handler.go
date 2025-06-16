package api

import (
	"entitydb/models"
	"entitydb/logger"
	"net/http"
	"os"
	"time"
	"runtime"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	entityRepo *models.RepositoryQueryWrapper
	startTime  time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(entityRepo models.EntityRepository) *HealthHandler {
	return &HealthHandler{
		entityRepo: models.NewRepositoryQueryWrapper(entityRepo),
		startTime:  time.Now(),
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string            `json:"status"`
	Timestamp   time.Time         `json:"timestamp"`
	Uptime      string            `json:"uptime"`
	Version     string            `json:"version"`
	Checks      map[string]string `json:"checks"`
	Metrics     HealthMetrics     `json:"metrics"`
}

// HealthMetrics contains system health metrics
// @Description System health and performance metrics
type HealthMetrics struct {
	EntityCount    int           `json:"entity_count" example:"100"`
	UserCount      int           `json:"user_count" example:"5"`
	DatabaseSize   int64         `json:"database_size_bytes" example:"1048576"`
	MemoryUsage    MemoryMetrics `json:"memory_usage"`
	GoRoutines     int           `json:"goroutines" example:"25"`
}

// MemoryMetrics contains memory usage information
// @Description Memory usage statistics
type MemoryMetrics struct {
	Alloc      uint64 `json:"alloc_bytes" example:"10485760"`
	TotalAlloc uint64 `json:"total_alloc_bytes" example:"20971520"`
	Sys        uint64 `json:"sys_bytes" example:"73400320"`
	NumGC      uint32 `json:"num_gc" example:"5"`
}

// Health returns the health status of the system
// @Summary Health check
// @Description Get system health status and basic metrics
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Handling health check request")
	
	// Calculate uptime
	uptime := time.Since(h.startTime)
	
	// Perform health checks
	checks := make(map[string]string)
	status := "healthy"
	
	// Check database connectivity
	_, err := h.entityRepo.Query().Limit(1).Execute()
	if err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		status = "unhealthy"
	} else {
		checks["database"] = "healthy"
	}
	
	// Count entities
	allEntities, err := h.entityRepo.Query().Execute()
	entityCount := 0
	if err == nil {
		entityCount = len(allEntities)
	}
	
	// Count users
	users, err := h.entityRepo.Query().HasTag("type:user").Execute()
	userCount := 0
	if err == nil {
		userCount = len(users)
	}
	
	// Get database file size (including WAL and index files)
	var dbSize int64
	if stat, err := os.Stat("/opt/entitydb/var/entities.db"); err == nil {
		dbSize += stat.Size()
	}
	if stat, err := os.Stat("/opt/entitydb/var/entitydb.wal"); err == nil {
		dbSize += stat.Size()
	}
	if stat, err := os.Stat("/opt/entitydb/var/entities.db.idx"); err == nil {
		dbSize += stat.Size()
	}
	
	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	// Build response
	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
		Version:   "2.14.0+",
		Checks:    checks,
		Metrics: HealthMetrics{
			EntityCount:  entityCount,
			UserCount:    userCount,
			DatabaseSize: dbSize,
			MemoryUsage: MemoryMetrics{
				Alloc:      memStats.Alloc,
				TotalAlloc: memStats.TotalAlloc,
				Sys:        memStats.Sys,
				NumGC:      memStats.NumGC,
			},
			GoRoutines:   runtime.NumGoroutine(),
		},
	}
	
	// Set appropriate HTTP status
	statusCode := http.StatusOK
	if status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}
	
	RespondJSON(w, statusCode, response)
}