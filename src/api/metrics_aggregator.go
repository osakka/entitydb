package api

import (
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MetricsAggregator aggregates labeled metrics into simple metrics for UI display
type MetricsAggregator struct {
	repo     models.EntityRepository
	mu       sync.Mutex
	interval time.Duration
	ctx      chan struct{}
	wg       sync.WaitGroup
}

// NewMetricsAggregator creates a new metrics aggregator
func NewMetricsAggregator(repo models.EntityRepository, interval time.Duration) *MetricsAggregator {
	return &MetricsAggregator{
		repo:     repo,
		interval: interval,
		ctx:      make(chan struct{}),
	}
}

// Start begins the aggregation process
func (a *MetricsAggregator) Start() {
	logger.Info("Starting metrics aggregator with interval: %v", a.interval)
	
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		
		// Initial aggregation after a short delay
		select {
		case <-time.After(10 * time.Second):
			a.aggregate()
		case <-a.ctx:
			return
		}
		
		// Periodic aggregation
		ticker := time.NewTicker(a.interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				a.aggregate()
			case <-a.ctx:
				return
			}
		}
	}()
}

// Stop stops the aggregation process
func (a *MetricsAggregator) Stop() {
	close(a.ctx)
	a.wg.Wait()
}

// aggregate performs the aggregation of labeled metrics
func (a *MetricsAggregator) aggregate() {
	logger.Info("Starting metrics aggregation...")
	
	// Define metrics to aggregate
	metricsToAggregate := []struct {
		name        string
		unit        string
		aggregation string // "sum", "avg", "max", "min", "last"
		description string
	}{
		{"http_request_duration_ms", "milliseconds", "avg", "Average HTTP request duration"},
		{"query_execution_time_ms", "milliseconds", "avg", "Average query execution time"},
		{"storage_read_duration_ms", "milliseconds", "avg", "Average storage read duration"},
		{"storage_write_duration_ms", "milliseconds", "avg", "Average storage write duration"},
		{"error_count", "count", "sum", "Total error count"},
		{"http_errors_total", "count", "sum", "Total HTTP errors"},
		{"http_requests_total", "count", "sum", "Total HTTP requests"},
		{"query_error_count", "count", "sum", "Total query errors"},
	}
	
	for _, metric := range metricsToAggregate {
		a.aggregateMetric(metric.name, metric.unit, metric.aggregation, metric.description)
	}
	
	logger.Info("Metrics aggregation completed")
}

// aggregateMetric aggregates a specific metric
func (a *MetricsAggregator) aggregateMetric(metricName, unit, aggregation, description string) {
	// Find all metrics with this name
	metrics, err := a.repo.ListByTag(fmt.Sprintf("name:%s", metricName))
	if err != nil {
		logger.Error("Failed to list metrics for %s: %v", metricName, err)
		return
	}
	
	if len(metrics) == 0 {
		logger.Info("No metrics found for %s", metricName)
		return
	}
	
	logger.Info("Found %d metrics with name %s", len(metrics), metricName)
	
	// Collect recent values from all labeled metrics
	now := time.Now()
	cutoff := now.Add(-24 * time.Hour) // Look at last 24 hours for better coverage
	logger.Info("Aggregating %s - now=%s, cutoff=%s", metricName, now.Format(time.RFC3339), cutoff.Format(time.RFC3339))
	
	var values []float64
	for _, metric := range metrics {
		// Get entity with full temporal tags by re-fetching it
		// (since ListByTag returns entities without timestamps)
		fullEntity, err := a.repo.GetByID(metric.ID)
		if err != nil {
			logger.Debug("Failed to get full entity for metric %s: %v", metric.ID, err)
			continue
		}
		
		// Find the most recent value within the cutoff time
		var latestValue float64
		var latestTime time.Time
		foundValue := false
		
		logger.Debug("Processing metric %s with %d tags", fullEntity.ID, len(fullEntity.Tags))
		
		// Log first few tags to debug
		for i, tag := range fullEntity.Tags {
			if i < 3 {
				logger.Debug("Tag %d: %s", i, tag)
			}
		}
		
		for _, tag := range fullEntity.Tags {
			// Temporal Tag Processing Algorithm:
			// 1. Default to current time for non-temporal tags
			// 2. Detect temporal format by looking for pipe delimiter (|)
			// 3. Extract timestamp portion (before |) and tag portion (after |)
			// 4. Parse timestamp as epoch nanoseconds 
			// 5. Skip malformed timestamps but continue processing other tags
			actualTag := tag
			tagTime := now
			
			// Check for temporal tag format: "TIMESTAMP|tag"
			if idx := strings.Index(tag, "|"); idx != -1 {
				// Split into timestamp and tag components
				timestampStr := tag[:idx]    // Everything before the pipe
				actualTag = tag[idx+1:]      // Everything after the pipe
				
				// Parse timestamp as epoch nanoseconds (EntityDB standard format)
				if ts, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
					tagTime = time.Unix(0, ts)  // Convert nanoseconds to time.Time
					logger.Trace("Parsed temporal tag: timestamp=%d (%s), tag=%s", ts, tagTime.Format(time.RFC3339), actualTag)
				} else {
					// Malformed timestamp - log and skip this tag
					logger.Debug("Failed to parse timestamp from tag: %s, error: %v", tag, err)
					continue
				}
			}
			
			// Look for value tags
			if strings.HasPrefix(actualTag, "value:") {
				valueStr := strings.TrimPrefix(actualTag, "value:")
				if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
					// Skip if outside cutoff time
					if tagTime.After(cutoff) {
						if !foundValue || tagTime.After(latestTime) {
							latestValue = value
							latestTime = tagTime
							foundValue = true
							logger.Debug("Found value %.2f at time %s", value, tagTime.Format(time.RFC3339))
						}
					} else {
						logger.Trace("Skipping value %.2f at time %s (before cutoff %s)", value, tagTime.Format(time.RFC3339), cutoff.Format(time.RFC3339))
					}
				}
			}
		}
		
		if foundValue {
			values = append(values, latestValue)
			logger.Debug("Added value %.2f from metric %s", latestValue, metric.ID)
		} else {
			logger.Debug("No recent value found in metric %s", metric.ID)
		}
	}
	
	// Calculate aggregated value
	var aggregatedValue float64
	
	if len(values) == 0 {
		logger.Info("No recent values found for %s (cutoff: %s)", metricName, cutoff.Format(time.RFC3339))
		// Even if no recent values, create/update the metric with a zero value for the UI
		aggregatedValue = 0.0
	} else {
		logger.Info("Found %d recent values for %s", len(values), metricName)
		
		switch aggregation {
	case "sum":
		for _, v := range values {
			aggregatedValue += v
		}
	case "avg":
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		aggregatedValue = sum / float64(len(values))
	case "max":
		aggregatedValue = values[0]
		for _, v := range values {
			if v > aggregatedValue {
				aggregatedValue = v
			}
		}
	case "min":
		aggregatedValue = values[0]
		for _, v := range values {
			if v < aggregatedValue {
				aggregatedValue = v
			}
		}
	case "last":
		aggregatedValue = values[len(values)-1]
	}
	}
	
	// Store aggregated value under simple metric name
	simpleMetricID := fmt.Sprintf("metric_%s", metricName)
	
	// Check if simple metric exists
	_, err = a.repo.GetByID(simpleMetricID)
	if err != nil {
		// Create new simple metric entity
		tags := []string{
			"type:metric",
			"dataset:system",
			fmt.Sprintf("name:%s", metricName),
			fmt.Sprintf("unit:%s", unit),
			fmt.Sprintf("description:%s", description),
			fmt.Sprintf("aggregation:%s", aggregation),
			// Don't add static value tag - we'll add temporal value tag below
			// Retention settings for UI metrics
			"retention:count:500",   // Keep 500 data points
			"retention:period:7200", // Retain for 2 hours
		}
		
		newEntity := &models.Entity{
			ID:      simpleMetricID,
			Tags:    tags,
			Content: []byte{},
		}
		
		if err := a.repo.Create(newEntity); err != nil {
			logger.Error("Failed to create aggregated metric %s: %v", simpleMetricID, err)
			return
		}
		logger.Info("Created aggregated metric: %s", simpleMetricID)
		// Don't return - continue to add temporal value tag
	}
	
	// Always add new temporal value tag
	valueTag := fmt.Sprintf("value:%.2f", aggregatedValue)
	logger.Debug("Calling AddTag for %s with tag: %s", simpleMetricID, valueTag)
	if err := a.repo.AddTag(simpleMetricID, valueTag); err != nil {
		logger.Error("Failed to update aggregated metric %s: %v", simpleMetricID, err)
		return
	}
	logger.Debug("Successfully called AddTag for %s", simpleMetricID)
	
	logger.Info("Updated aggregated metric %s with value %.2f from %d sources", metricName, aggregatedValue, len(values))
}

// Global instance
var metricsAggregator *MetricsAggregator

// InitMetricsAggregator initializes the global metrics aggregator
func InitMetricsAggregator(repo models.EntityRepository, interval time.Duration) {
	metricsAggregator = NewMetricsAggregator(repo, interval)
}

// GetMetricsAggregator returns the global metrics aggregator
func GetMetricsAggregator() *MetricsAggregator {
	return metricsAggregator
}