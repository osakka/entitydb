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

type HealthMetrics struct {
	EntityCount    int               `json:"entity_count"`
	UserCount      int               `json:"user_count"`
	DatabaseSize   int64             `json:"database_size_bytes"`
	MemoryUsage    runtime.MemStats  `json:"memory_usage"`
	GoRoutines     int               `json:"goroutines"`
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
	
	// Get database file size
	var dbSize int64
	if stat, err := os.Stat("/opt/entitydb/var/entities.ebf"); err == nil {
		dbSize = stat.Size()
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
			MemoryUsage:  memStats,
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