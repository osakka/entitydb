package api

import (
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ErrorMetricsCollector collects error metrics across the system
type ErrorMetricsCollector struct {
	repo          models.EntityRepository
	mu            sync.Mutex
	errorPatterns map[string]int // Track error patterns for categorization
	errorChan     chan errorEvent
	stopChan      chan struct{}
}

// errorEvent represents an error to be tracked asynchronously
type errorEvent struct {
	component string
	err       error
	severity  string
}

// NewErrorMetricsCollector creates a new error metrics collector
func NewErrorMetricsCollector(repo models.EntityRepository) *ErrorMetricsCollector {
	c := &ErrorMetricsCollector{
		repo:          repo,
		errorPatterns: make(map[string]int),
		errorChan:     make(chan errorEvent, 1000), // Buffered channel to prevent blocking
		stopChan:      make(chan struct{}),
	}
	
	// Start background goroutine to process errors
	go c.processErrors()
	
	return c
}

// TrackError tracks an error occurrence asynchronously
func (c *ErrorMetricsCollector) TrackError(component string, err error, severity string) {
	if err == nil {
		return
	}
	
	// Send error to channel for async processing
	select {
	case c.errorChan <- errorEvent{component: component, err: err, severity: severity}:
		// Successfully queued
	default:
		// Channel is full, log but don't block
		logger.Warn("Error tracking channel full, dropping error event for %s: %v", component, err)
	}
}

// processErrors processes errors in the background
func (c *ErrorMetricsCollector) processErrors() {
	for {
		select {
		case event := <-c.errorChan:
			c.processError(event)
		case <-c.stopChan:
			return
		}
	}
}

// processError processes a single error event
func (c *ErrorMetricsCollector) processError(event errorEvent) {
	errorType := c.categorizeError(event.err)
	errorMsg := event.err.Error()
	
	// Track error count
	c.storeMetric("error_count", 1, "count", "Total error count",
		map[string]string{
			"component": event.component,
			"type":      errorType,
			"severity":  event.severity,
		})
	
	// Track error patterns
	c.mu.Lock()
	pattern := c.extractErrorPattern(errorMsg)
	c.errorPatterns[pattern]++
	patternCount := c.errorPatterns[pattern]
	c.mu.Unlock()
	
	// Track frequent error patterns
	if patternCount > 10 {
		c.storeMetric("frequent_error_patterns", float64(patternCount), "count",
			"Frequently occurring error patterns",
			map[string]string{
				"pattern": pattern,
			})
	}
	
	// Log errors with appropriate severity
	switch event.severity {
	case "critical":
		logger.Error("[%s] Critical error: %v", event.component, event.err)
	case "error":
		logger.Error("[%s] Error: %v", event.component, event.err)
	case "warning":
		logger.Warn("[%s] Warning: %v", event.component, event.err)
	default:
		logger.Debug("[%s] Error: %v", event.component, event.err)
	}
}

// TrackPanic tracks panic occurrences with stack trace
func (c *ErrorMetricsCollector) TrackPanic(component string) {
	if r := recover(); r != nil {
		// Get stack trace
		stack := string(debug.Stack())
		
		// Track panic count
		c.storeMetric("panic_count", 1, "count", "Total panic count",
			map[string]string{
				"component": component,
			})
		
		// Log panic with stack trace
		logger.Error("[%s] PANIC: %v\nStack trace:\n%s", component, r, stack)
		
		// Re-panic to maintain normal panic behavior
		panic(r)
	}
}

// TrackRecovery tracks error recovery attempts
func (c *ErrorMetricsCollector) TrackRecovery(component string, errorType string, duration time.Duration, success bool) {
	// Track recovery time
	c.storeMetric("error_recovery_time_ms", 
		float64(duration.Milliseconds()), 
		"milliseconds",
		"Time to recover from errors",
		map[string]string{
			"component": component,
			"error_type": errorType,
			"success": strconv.FormatBool(success),
		})
	
	// Track recovery attempts
	c.storeMetric("recovery_attempts", 1, "count", "Error recovery attempts",
		map[string]string{
			"component": component,
			"error_type": errorType,
			"success": strconv.FormatBool(success),
		})
}

// categorizeError categorizes the error type
func (c *ErrorMetricsCollector) categorizeError(err error) string {
	errStr := strings.ToLower(err.Error())
	
	switch {
	case strings.Contains(errStr, "not found"):
		return "not_found"
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "permission") || strings.Contains(errStr, "unauthorized"):
		return "permission_denied"
	case strings.Contains(errStr, "invalid"):
		return "invalid_input"
	case strings.Contains(errStr, "connection") || strings.Contains(errStr, "network"):
		return "network_error"
	case strings.Contains(errStr, "disk") || strings.Contains(errStr, "storage"):
		return "storage_error"
	case strings.Contains(errStr, "memory"):
		return "memory_error"
	case strings.Contains(errStr, "corrupt"):
		return "corruption_error"
	default:
		return "internal_error"
	}
}

// extractErrorPattern extracts a generalized pattern from error message
func (c *ErrorMetricsCollector) extractErrorPattern(errorMsg string) string {
	// Remove specific IDs, numbers, and paths
	pattern := errorMsg
	
	// Replace UUIDs
	pattern = strings.ReplaceAll(pattern, `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`, "UUID")
	
	// Replace numbers
	pattern = strings.ReplaceAll(pattern, `\d+`, "N")
	
	// Replace file paths
	pattern = strings.ReplaceAll(pattern, `\/[^\s]+`, "/PATH")
	
	// Truncate to reasonable length
	if len(pattern) > 100 {
		pattern = pattern[:100] + "..."
	}
	
	return pattern
}

// storeMetric stores a metric value with labels
func (c *ErrorMetricsCollector) storeMetric(name string, value float64, unit string, description string, labels map[string]string) {
	// Build metric ID with labels
	metricID := "metric_" + name
	for k, v := range labels {
		metricID += "_" + k + "_" + v
	}
	
	// Check if metric exists
	entity, err := c.repo.GetByID(metricID)
	if err != nil {
		// Create new metric entity
		tags := []string{
			"type:metric",
			"dataspace:system",
			"name:" + name,
			"unit:" + unit,
			"description:" + description,
		}
		
		// Add label tags
		for k, v := range labels {
			tags = append(tags, fmt.Sprintf("label:%s:%s", k, v))
		}
		
		// Initial value
		tags = append(tags, fmt.Sprintf("value:%.2f", value))
		
		// Retention for error metrics: 24 hours, 2000 data points
		tags = append(tags, "retention:count:2000", "retention:period:86400")
		
		newEntity := &models.Entity{
			ID:      metricID,
			Tags:    tags,
			Content: []byte{},
		}
		
		if err := c.repo.Create(newEntity); err != nil {
			// Don't log to avoid recursion
			return
		}
		return
	}
	
	// For counters, increment the current value
	if unit == "count" {
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
	
	// Add temporal value tag
	valueTag := fmt.Sprintf("value:%.2f", value)
	if err := c.repo.AddTag(metricID, valueTag); err != nil {
		// Don't log to avoid recursion
	}
}

// Global instance
var errorMetrics *ErrorMetricsCollector

// InitErrorMetrics initializes the global error metrics collector
func InitErrorMetrics(repo models.EntityRepository) {
	errorMetrics = NewErrorMetricsCollector(repo)
}

// GetErrorMetrics returns the global error metrics instance
func GetErrorMetrics() *ErrorMetricsCollector {
	return errorMetrics
}

// Stop gracefully shuts down the error collector
func (c *ErrorMetricsCollector) Stop() {
	close(c.stopChan)
}

// TrackHTTPError is a convenience function for tracking HTTP errors
func TrackHTTPError(component string, statusCode int, err error) {
	if errorMetrics == nil {
		return
	}
	
	severity := "error"
	if statusCode >= 500 {
		severity = "critical"
	} else if statusCode >= 400 {
		severity = "warning"
	}
	
	errorMetrics.TrackError(component, err, severity)
}