package api

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	
	"entitydb/logger"
	"entitydb/models"
)

// RequestMetricsMiddleware tracks HTTP request metrics
type RequestMetricsMiddleware struct {
	repo       models.EntityRepository
	workerPool *MetricsWorkerPool
}

// NewRequestMetricsMiddleware creates a new request metrics middleware
func NewRequestMetricsMiddleware(repo models.EntityRepository) *RequestMetricsMiddleware {
	// Create worker pool with 10 workers and queue size of 1000
	workerPool := NewMetricsWorkerPool(10, 1000)
	return &RequestMetricsMiddleware{
		repo:       repo,
		workerPool: workerPool,
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
		// Skip metrics endpoint to avoid recursion
		if strings.HasPrefix(r.URL.Path, "/api/v1/metrics") || r.URL.Path == "/metrics" {
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
	// Build metric ID with labels in sorted order for consistency
	metricID := "metric_" + name
	
	// Sort label keys for consistent ID generation
	var keys []string
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	for _, k := range keys {
		metricID += "_" + k + "_" + labels[k]
	}
	
	logger.Debug("Storing metric: id=%s, value=%.2f", metricID, value)
	
	// Check if metric exists
	entity, err := m.repo.GetByID(metricID)
	if err != nil {
		// Create new metric entity
		tags := []string{
			"type:metric",
			"dataset:system",
			"name:" + name,
			"unit:" + unit,
			"description:" + description,
		}
		
		// Add label tags
		for k, v := range labels {
			tags = append(tags, "label:"+k+":"+v)
		}
		
		// Don't add static value tag - we'll use AddTag for temporal values
		// tags = append(tags, "value:"+strconv.FormatFloat(value, 'f', 2, 64))
		
		// Retention: keep request metrics for 1 hour with 1000 data points max
		tags = append(tags, "retention:count:1000", "retention:period:3600")
		
		newEntity := &models.Entity{
			ID:      metricID,
			Tags:    tags,
			Content: []byte{},
		}
		
		if err := m.repo.Create(newEntity); err != nil {
			logger.Error("Failed to create request metric %s: %v", metricID, err)
			// Don't return - entity might already exist, continue to add temporal value
		}
	} else {
		// Entity exists - for counters, we need to increment the current value
		if unit == "count" {
			// Get current value
			currentValue := 0.0
			for _, tag := range entity.GetTagsWithoutTimestamp() {
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
	
	// Add temporal value tag
	valueTag := "value:" + strconv.FormatFloat(value, 'f', 2, 64)
	if err := m.repo.AddTag(metricID, valueTag); err != nil {
		logger.Error("Failed to update request metric %s: %v", metricID, err)
	}
}
// Shutdown gracefully stops the metrics worker pool
func (m *RequestMetricsMiddleware) Shutdown() {
	if m.workerPool != nil {
		logger.Info("Shutting down metrics worker pool...")
		m.workerPool.Shutdown()
	}
}
