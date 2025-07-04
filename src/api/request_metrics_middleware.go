package api

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
)

// RequestMetricsMiddleware tracks HTTP request metrics
type RequestMetricsMiddleware struct {
	repo       models.EntityRepository
	workerPool *MetricsWorkerPool
	lastValues map[string]float64 // Track last values for change detection
	mu         sync.RWMutex       // Protect lastValues map
	
	// BAR-RAISING SOLUTION: Circuit breaker to prevent feedback loops
	failureCount    int           // Count of consecutive failures
	circuitOpen     bool          // True when circuit is open (metrics collection disabled)
	lastFailure     time.Time     // Time of last failure
	circuitMu       sync.RWMutex  // Protect circuit breaker state
}

// NewRequestMetricsMiddleware creates a new request metrics middleware
func NewRequestMetricsMiddleware(repo models.EntityRepository) *RequestMetricsMiddleware {
	// Create worker pool with 10 workers and queue size of 1000
	workerPool := NewMetricsWorkerPool(10, 1000)
	return &RequestMetricsMiddleware{
		repo:       repo,
		workerPool: workerPool,
		lastValues: make(map[string]float64),
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

// Middleware returns the HTTP middleware function
func (m *RequestMetricsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// BAR-RAISING SOLUTION: Check circuit breaker first
		if m.isCircuitOpen() {
			logger.Trace("Request metrics circuit is open - skipping metrics collection to prevent feedback loops")
			next.ServeHTTP(w, r)
			return
		}
		
		// Skip metrics endpoint to avoid recursion
		if strings.HasPrefix(r.URL.Path, "/api/v1/metrics") || 
		   strings.HasPrefix(r.URL.Path, "/api/v1/system/metrics") ||
		   r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		
		// TEMPORARY: Skip auth endpoints to avoid potential deadlock
		if strings.HasPrefix(r.URL.Path, "/api/v1/auth/") {
			logger.Debug("Skipping metrics for auth endpoint: %s", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}
		
		logger.Debug("RequestMetricsMiddleware: Processing request %s %s", r.Method, r.URL.Path)
		
		start := time.Now()
		
		// Wrap response writer
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // default
		}
		
		// Get request size
		requestSize := r.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}
		
		// Process request
		next.ServeHTTP(wrapped, r)
		
		// Calculate duration
		duration := time.Since(start)
		
		// Normalize path for metrics (remove IDs)
		path := normalizePath(r.URL.Path)
		
		// Skip storing if it's a static file
		if isStaticFile(path) {
			return
		}
		
		// Submit metrics task to worker pool
		method := r.Method
		statusCode := wrapped.statusCode
		respSize := int64(wrapped.size)
		
		submitted := m.workerPool.Submit(func() {
			m.storeRequestMetrics(method, path, statusCode, duration, requestSize, respSize)
		})
		
		if !submitted {
			logger.Warn("Failed to submit metrics task for %s %s (queue full)", method, path)
		}
	})
}

// normalizePath removes dynamic parts from paths for better grouping
func normalizePath(path string) string {
	// Remove trailing slash
	path = strings.TrimSuffix(path, "/")
	
	// Common patterns to normalize
	patterns := []struct {
		prefix string
		replacement string
	}{
		{"/api/v1/entities/", "/api/v1/entities/:id"},
		{"/api/v1/users/", "/api/v1/users/:id"},
		{"/api/v1/datasets/", "/api/v1/datasets/:id"},
		{"/api/v1/entity-relationships/", "/api/v1/entity-relationships/:id"},
	}
	
	for _, pattern := range patterns {
		if strings.HasPrefix(path, pattern.prefix) && len(path) > len(pattern.prefix) {
			return pattern.replacement
		}
	}
	
	return path
}

// isStaticFile checks if the path is for a static file
func isStaticFile(path string) bool {
	staticExtensions := []string{".js", ".css", ".html", ".png", ".jpg", ".svg", ".ico"}
	for _, ext := range staticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

// storeRequestMetrics stores the collected metrics
func (m *RequestMetricsMiddleware) storeRequestMetrics(method, path string, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	// Mark this goroutine as performing metrics operations to prevent recursion
	binary.SetMetricsOperation(true)
	defer binary.SetMetricsOperation(false)
	
	// Add panic recovery to ensure goroutine doesn't crash
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in storeRequestMetrics: %v", r)
		}
	}()
	
	logger.Debug("Storing request metrics: method=%s, path=%s, status=%d, duration=%v", method, path, statusCode, duration)
	
	// Store multiple metrics for comprehensive monitoring
	
	// 1. Request count by endpoint
	logger.Debug("Storing metric 1: http_requests_total")
	m.storeMetric("http_requests_total", 1, "count", 
		"Total HTTP requests",
		map[string]string{
			"method": method,
			"path": path,
			"status": strconv.Itoa(statusCode),
		})
	logger.Debug("Completed storing metric 1")
	
	// 2. Request duration
	logger.Debug("Storing metric 2: http_request_duration_ms = %v ms", duration.Milliseconds())
	m.storeMetric("http_request_duration_ms", float64(duration.Milliseconds()), "milliseconds",
		"HTTP request duration",
		map[string]string{
			"method": method,
			"path": path,
		})
	logger.Debug("Completed storing metric 2")
	
	// 3. Request size
	if requestSize > 0 {
		logger.Debug("Storing metric 3: http_request_size_bytes = %v bytes", requestSize)
		m.storeMetric("http_request_size_bytes", float64(requestSize), "bytes",
			"HTTP request size",
			map[string]string{
				"method": method,
				"path": path,
			})
		logger.Debug("Completed storing metric 3")
	}
	
	// 4. Response size
	logger.Debug("Storing metric 4: http_response_size_bytes = %v bytes", responseSize)
	m.storeMetric("http_response_size_bytes", float64(responseSize), "bytes",
		"HTTP response size",
		map[string]string{
			"method": method,
			"path": path,
			"status": strconv.Itoa(statusCode),
		})
	logger.Debug("Completed storing metric 4")
	
	// 5. Error count (4xx and 5xx)
	if statusCode >= 400 {
		errorType := "client_error"
		if statusCode >= 500 {
			errorType = "server_error"
		}
		m.storeMetric("http_errors_total", 1, "count",
			"Total HTTP errors",
			map[string]string{
				"method": method,
				"path": path,
				"type": errorType,
				"status": strconv.Itoa(statusCode),
			})
	}
	
	// 6. Slow requests (> 1 second)
	if duration > time.Second {
		m.storeMetric("http_slow_requests_total", 1, "count",
			"Slow HTTP requests (>1s)",
			map[string]string{
				"method": method,
				"path": path,
			})
	}
}

// storeMetric stores a metric value with labels
func (m *RequestMetricsMiddleware) storeMetric(name string, value float64, unit string, description string, labels map[string]string) {
	// Build metric identifier with labels in sorted order for consistent lookup
	metricKey := name
	
	// Sort label keys for consistent key generation
	var keys []string
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	for _, k := range keys {
		metricKey += "_" + k + "_" + labels[k]
	}
	
	// Check if value has changed using change detection (CRITICAL FIX)
	m.mu.RLock()
	lastValue, exists := m.lastValues[metricKey]
	m.mu.RUnlock()
	
	// For counters, always increment, but for gauges only store if changed
	skipStorage := false
	if unit != "count" && exists && lastValue == value {
		logger.Trace("Request metric %s unchanged (%.2f), skipping storage", metricKey, value)
		skipStorage = true
	}
	
	if !skipStorage {
		// Update last value
		m.mu.Lock()
		m.lastValues[metricKey] = value
		m.mu.Unlock()
		
		logger.Debug("Request metric %s changed from %.2f to %.2f, storing", metricKey, lastValue, value)
	} else {
		logger.Debug("Skipping unchanged metric: key=%s, value=%.2f", metricKey, value)
		return
	}
	
	// Try to find existing metric entity by searching for name and label tags
	searchTags := []string{
		"name:" + name,
		"type:metric",
	}
	for k, v := range labels {
		searchTags = append(searchTags, "label:"+k+":"+v)
	}
	
	var metricEntity *models.Entity
	var metricID string
	
	// Search for existing entity by name tag (simplified approach)
	nameTagEntities, err := m.repo.ListByTag(fmt.Sprintf("name:%s", name))
	var found bool
	
	if err == nil {
		// Look for entity with matching labels
		for _, entity := range nameTagEntities {
			cleanTags := entity.GetTagsWithoutTimestamp()
			matches := 0
			required := len(labels) + 1 // +1 for type:metric
			
			for _, tag := range cleanTags {
				if tag == "type:metric" {
					matches++
				}
				for k, v := range labels {
					if tag == "label:"+k+":"+v {
						matches++
					}
				}
			}
			
			if matches == required {
				metricEntity = entity
				metricID = entity.ID
				found = true
				logger.Trace("Found existing metric entity: %s for metric %s", metricID, metricKey)
				break
			}
		}
	}
	
	if !found {
		// Create new metric entity using UUID architecture
		additionalTags := []string{
			"name:" + name,
			"unit:" + unit,
			"description:" + description,
			"retention:count:1000", // Keep request metrics for high volume
			"retention:period:3600", // Keep for 1 hour
		}
		
		// Add label tags
		for k, v := range labels {
			additionalTags = append(additionalTags, "label:"+k+":"+v)
		}
		
		newEntity, err := models.NewEntityWithMandatoryTags(
			"metric",                    // entityType
			"system",                    // dataset
			models.SystemUserID,         // createdBy (system user)
			additionalTags,             // additional tags
		)
		if err != nil {
			logger.Error("Failed to create request metric entity for %s: %v", metricKey, err)
			return
		}
		
		if err := m.repo.Create(newEntity); err != nil {
			logger.Error("Failed to store request metric entity %s: %v", newEntity.ID, err)
			// Don't return - entity might already exist, continue to add temporal value
		}
		
		metricEntity = newEntity
		metricID = newEntity.ID
		logger.Debug("Created request metric entity with UUID: %s for metric %s", metricID, metricKey)
	} else {
		// Entity exists - for counters, we need to increment the current value
		if unit == "count" {
			// Get current value from most recent value tag
			currentValue := 0.0
			for _, tag := range metricEntity.GetTagsWithoutTimestamp() {
				if strings.HasPrefix(tag, "value:") {
					if val, err := strconv.ParseFloat(strings.TrimPrefix(tag, "value:"), 64); err == nil {
						currentValue = val
						break
					}
				}
			}
			value = currentValue + value
		}
	}
	
	// ATOMIC TAG FIX: Add temporal value tag with explicit timestamp
	valueTag := "value:" + strconv.FormatFloat(value, 'f', 2, 64)
	nowNano := time.Now().UnixNano()
	timestampedValueTag := fmt.Sprintf("%d|%s", nowNano, valueTag)
	
	// Get entity and update atomically
	entity, getErr := m.repo.GetByID(metricID)
	if getErr != nil {
		logger.Error("Failed to get request metric entity %s: %v", metricID, getErr)
		return
	}
	entity.Tags = append(entity.Tags, timestampedValueTag)
	if updateErr := m.repo.Update(entity); updateErr != nil {
		logger.Error("Failed to update request metric %s: %v", metricID, updateErr)
		m.recordFailure() // BAR-RAISING: Track failures for circuit breaker
		return
	}
	
	// BAR-RAISING: Record successful operation to reset failure count
	m.recordSuccess()
	logger.Trace("Stored request metric %s with value: %.2f %s (entity: %s)", metricKey, value, unit, metricID)
}
// Shutdown gracefully stops the metrics worker pool
func (m *RequestMetricsMiddleware) Shutdown() {
	if m.workerPool != nil {
		logger.Info("Shutting down metrics worker pool...")
		m.workerPool.Shutdown()
	}
}

// BAR-RAISING SOLUTION: Circuit breaker methods to prevent feedback loops

// isCircuitOpen checks if the circuit breaker is open (metrics collection disabled)
func (m *RequestMetricsMiddleware) isCircuitOpen() bool {
	m.circuitMu.RLock()
	defer m.circuitMu.RUnlock()
	
	// Circuit is open if we have too many failures
	if m.failureCount >= 5 {
		// Auto-recovery after 5 minutes
		if time.Since(m.lastFailure) > 5*time.Minute {
			m.circuitMu.RUnlock()
			m.circuitMu.Lock()
			m.failureCount = 0
			m.circuitOpen = false
			m.circuitMu.Unlock()
			logger.Info("Request metrics circuit breaker auto-recovery: reopening after 5 minutes")
			m.circuitMu.RLock()
		}
	}
	
	return m.circuitOpen
}

// recordFailure increments failure count and may open the circuit
func (m *RequestMetricsMiddleware) recordFailure() {
	m.circuitMu.Lock()
	defer m.circuitMu.Unlock()
	
	m.failureCount++
	m.lastFailure = time.Now()
	
	if m.failureCount >= 5 && !m.circuitOpen {
		m.circuitOpen = true
		logger.Warn("REQUEST METRICS CIRCUIT BREAKER OPENED: Disabling request metrics after %d consecutive failures", m.failureCount)
	}
}

// recordSuccess resets failure count and closes circuit if open
func (m *RequestMetricsMiddleware) recordSuccess() {
	m.circuitMu.Lock()
	defer m.circuitMu.Unlock()
	
	if m.failureCount > 0 || m.circuitOpen {
		logger.Info("Request metrics circuit breaker: Successful operation - resetting failure count (was %d)", m.failureCount)
		m.failureCount = 0
		m.circuitOpen = false
	}
}
