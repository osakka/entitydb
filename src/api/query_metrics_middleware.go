package api

import (
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// QueryMetricsCollector collects metrics for query operations
type QueryMetricsCollector struct {
	repo models.EntityRepository
}

// NewQueryMetricsCollector creates a new query metrics collector
func NewQueryMetricsCollector(repo models.EntityRepository) *QueryMetricsCollector {
	return &QueryMetricsCollector{repo: repo}
}

// TrackQuery tracks metrics for a query operation
func (q *QueryMetricsCollector) TrackQuery(queryType string, tags []string, startTime time.Time, resultCount int, err error) {
	duration := time.Since(startTime)
	
	// Calculate query complexity score based on number of tags and operators
	complexity := q.calculateComplexity(tags)
	
	// Store query execution time
	q.storeMetric("query_execution_time_ms", 
		float64(duration.Milliseconds()), 
		"milliseconds",
		"Query execution time",
		map[string]string{
			"query_type": queryType,
			"complexity": strconv.Itoa(complexity),
			"success": strconv.FormatBool(err == nil),
		})
	
	// Store result count
	if err == nil {
		q.storeMetric("query_result_count",
			float64(resultCount),
			"count",
			"Number of query results",
			map[string]string{
				"query_type": queryType,
			})
	}
	
	// Track slow queries (> 100ms)
	if duration > 100*time.Millisecond {
		q.storeMetric("slow_query_count",
			1,
			"count",
			"Slow query count",
			map[string]string{
				"query_type": queryType,
				"duration_bucket": q.getDurationBucket(duration),
			})
		
		logger.Warn("Slow query detected: type=%s, duration=%v, complexity=%d, results=%d",
			queryType, duration, complexity, resultCount)
	}
	
	// Track errors
	if err != nil {
		q.storeMetric("query_error_count",
			1,
			"count",
			"Query error count",
			map[string]string{
				"query_type": queryType,
				"error_type": q.categorizeError(err),
			})
	}
}

// TrackCacheOperation tracks cache hit/miss metrics
func (q *QueryMetricsCollector) TrackCacheOperation(operation string, hit bool) {
	if hit {
		q.storeMetric("query_cache_hits",
			1,
			"count",
			"Query cache hits",
			map[string]string{
				"operation": operation,
			})
	} else {
		q.storeMetric("query_cache_misses",
			1,
			"count",
			"Query cache misses",
			map[string]string{
				"operation": operation,
			})
	}
}

// TrackIndexLookup tracks index lookup performance
func (q *QueryMetricsCollector) TrackIndexLookup(indexType string, operation string, duration time.Duration, found bool) {
	q.storeMetric("index_lookup_duration_ms",
		float64(duration.Milliseconds()),
		"milliseconds",
		"Index lookup duration",
		map[string]string{
			"index_type": indexType,
			"operation": operation,
			"found": strconv.FormatBool(found),
		})
}

// calculateComplexity calculates a complexity score for a query
func (q *QueryMetricsCollector) calculateComplexity(tags []string) int {
	complexity := len(tags)
	
	// Add complexity for special operators
	for _, tag := range tags {
		if strings.Contains(tag, "*") {
			complexity += 2 // Wildcards add complexity
		}
		if strings.Contains(tag, "|") {
			complexity += 3 // OR operations add more complexity
		}
		if strings.HasPrefix(tag, "!") {
			complexity += 2 // Negations add complexity
		}
	}
	
	return complexity
}

// getDurationBucket returns a bucket label for the duration
func (q *QueryMetricsCollector) getDurationBucket(duration time.Duration) string {
	switch {
	case duration < 100*time.Millisecond:
		return "0-100ms"
	case duration < 500*time.Millisecond:
		return "100-500ms"
	case duration < 1*time.Second:
		return "500ms-1s"
	case duration < 5*time.Second:
		return "1s-5s"
	default:
		return ">5s"
	}
}

// categorizeError categorizes the error type
func (q *QueryMetricsCollector) categorizeError(err error) string {
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "not found"):
		return "not_found"
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "permission"):
		return "permission_denied"
	case strings.Contains(errStr, "invalid"):
		return "invalid_query"
	default:
		return "internal_error"
	}
}

// storeMetric stores a metric value with labels using async collection
func (q *QueryMetricsCollector) storeMetric(name string, value float64, unit string, description string, labels map[string]string) {
	// Route to async metrics system to prevent deadlocks and use UUIDs
	if globalAsyncCollector := binary.GetGlobalAsyncCollector(); globalAsyncCollector != nil {
		globalAsyncCollector.CollectMetric(name, value, unit, description, labels)
		logger.Trace("QueryMetrics.storeMetric: queued metric %s = %.2f via async collector", name, value)
		return
	}
	
	// Fallback: Legacy system (deprecated but maintained for zero regression)
	metricID := "metric_" + name
	
	// Check if metric exists
	entity, err := q.repo.GetByID(metricID)
	if err != nil {
		// Create new metric entity
		tags := []string{
			"type:metric",
			"dataset:system",
			"name:" + name,
			"unit:" + unit,
			"description:" + description,
			"retention:count:500", 
			"retention:period:21600",
		}
		
		newEntity := &models.Entity{
			ID:      metricID,
			Tags:    tags,
			Content: []byte{},
		}
		
		if err := q.repo.Create(newEntity); err != nil {
			logger.Error("Failed to create query metric %s: %v", metricID, err)
			return
		}
	}
	
	// Build value tag with labels embedded
	valueTag := fmt.Sprintf("value:%.2f", value)
	
	// Add sorted labels to value tag for dimensional data
	if len(labels) > 0 {
		var sortedKeys []string
		for k := range labels {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		
		var labelParts []string
		for _, k := range sortedKeys {
			labelParts = append(labelParts, fmt.Sprintf("%s=%s", k, labels[k]))
		}
		valueTag += ":" + strings.Join(labelParts, ":")
	}
	
	// For counters, we need special handling
	if unit == "count" {
		// For counters with labels, we need to track the current value for this label combination
		// Get current value for this specific label combination
		currentValue := 0.0
		targetLabelString := ""
		var labelParts []string
		
		if len(labels) > 0 {
			var sortedKeys []string
			for k := range labels {
				sortedKeys = append(sortedKeys, k)
			}
			sort.Strings(sortedKeys)
			
			for _, k := range sortedKeys {
				labelParts = append(labelParts, fmt.Sprintf("%s=%s", k, labels[k]))
			}
			targetLabelString = ":" + strings.Join(labelParts, ":")
		}
		
		// Look for existing value with same labels
		if entity != nil {
			for _, tag := range entity.GetTagsWithoutTimestamp() {
				if strings.HasPrefix(tag, "value:") && strings.Contains(tag, targetLabelString) {
					valueStr := strings.TrimPrefix(tag, "value:")
					if colonIdx := strings.Index(valueStr, ":"); colonIdx > 0 {
						valueStr = valueStr[:colonIdx]
					}
					if val, err := strconv.ParseFloat(valueStr, 64); err == nil {
						currentValue = val
						break
					}
				}
			}
		}
		
		// Update value tag with incremented value
		valueTag = fmt.Sprintf("value:%.2f", currentValue + value)
		if len(labels) > 0 {
			valueTag += ":" + strings.Join(labelParts, ":")
		}
	}
	
	// ATOMIC TAG FIX: Add temporal value tag with explicit timestamp
	nowNano := time.Now().UnixNano()
	timestampedValueTag := fmt.Sprintf("%d|%s", nowNano, valueTag)
	
	// Get entity and update atomically
	entity, getErr := q.repo.GetByID(metricID)
	if getErr != nil {
		logger.Error("Failed to get query metric entity %s: %v", metricID, getErr)
		return
	}
	entity.Tags = append(entity.Tags, timestampedValueTag)
	if updateErr := q.repo.Update(entity); updateErr != nil {
		logger.Error("Failed to update query metric %s: %v", metricID, updateErr)
	}
}

// Global instance for use in handlers
var queryMetrics *QueryMetricsCollector

// InitQueryMetrics initializes the global query metrics collector
func InitQueryMetrics(repo models.EntityRepository) {
	queryMetrics = NewQueryMetricsCollector(repo)
}

// GetQueryMetrics returns the global query metrics collector
func GetQueryMetrics() *QueryMetricsCollector {
	return queryMetrics
}