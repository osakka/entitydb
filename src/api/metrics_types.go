package api

import (
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"   // Monotonically increasing value
	MetricTypeGauge     MetricType = "gauge"     // Current value that can go up or down
	MetricTypeHistogram MetricType = "histogram" // Distribution of values
	MetricTypeSummary   MetricType = "summary"   // Statistical summary
)

// MetricValue represents a metric value with metadata
type MetricValue struct {
	Value     float64
	Timestamp time.Time
	Labels    map[string]string
}

// HistogramValue represents a histogram data point
type HistogramValue struct {
	Count   int64              // Total number of observations
	Sum     float64            // Sum of all observations
	Buckets map[float64]int64  // Bucket upper bounds -> counts
}

// SummaryValue represents a summary data point
type SummaryValue struct {
	Count      int64              // Total number of observations
	Sum        float64            // Sum of all observations
	Quantiles  map[float64]float64 // Quantile -> value
}

// MetricsTypeManager manages different metric types
type MetricsTypeManager struct {
	repo            models.EntityRepository
	histogramBuckets []float64
	mu              sync.RWMutex
}

// NewMetricsTypeManager creates a new metrics type manager
func NewMetricsTypeManager(repo models.EntityRepository, histogramBuckets []float64) *MetricsTypeManager {
	return &MetricsTypeManager{
		repo:             repo,
		histogramBuckets: histogramBuckets,
	}
}

// RecordCounter records a counter metric (always increments)
func (m *MetricsTypeManager) RecordCounter(name string, value float64, labels map[string]string, help string) error {
	if value < 0 {
		return fmt.Errorf("counter values must be non-negative")
	}
	
	metricID := m.buildMetricID(name, labels)
	
	// Get or create metric entity
	entity, err := m.repo.GetByID(metricID)
	if err != nil {
		// Create new counter metric
		tags := []string{
			"type:metric",
			"metric:type:counter",
			"dataset:system",
			fmt.Sprintf("name:%s", name),
			fmt.Sprintf("description:%s", help),
			"unit:count",
		}
		
		// Add label tags
		for k, v := range labels {
			tags = append(tags, fmt.Sprintf("label:%s:%s", k, v))
		}
		
		entity = &models.Entity{
			ID:      metricID,
			Tags:    tags,
			Content: []byte{},
		}
		
		if err := m.repo.Create(entity); err != nil {
			return fmt.Errorf("failed to create counter metric: %w", err)
		}
	}
	
	// For counters, we need to get the last value and add to it
	lastValue := m.getLastValue(entity)
	newValue := lastValue + value
	
	// ATOMIC TAG FIX: Add temporal value tag with explicit timestamp
	valueTag := fmt.Sprintf("value:%.6f", newValue)
	nowNano := time.Now().UnixNano()
	timestampedValueTag := fmt.Sprintf("%d|%s", nowNano, valueTag)
	
	// Get entity and update atomically
	entity, getErr := m.repo.GetByID(metricID)
	if getErr != nil {
		return fmt.Errorf("failed to get counter entity: %w", getErr)
	}
	entity.Tags = append(entity.Tags, timestampedValueTag)
	if updateErr := m.repo.Update(entity); updateErr != nil {
		return fmt.Errorf("failed to update counter: %w", updateErr)
	}
	
	logger.Trace("Recorded counter %s = %.6f (increment: %.6f)", name, newValue, value)
	return nil
}

// RecordGauge records a gauge metric (can go up or down)
func (m *MetricsTypeManager) RecordGauge(name string, value float64, labels map[string]string, help string) error {
	metricID := m.buildMetricID(name, labels)
	
	// Get or create metric entity
	entity, err := m.repo.GetByID(metricID)
	if err != nil {
		// Create new gauge metric
		tags := []string{
			"type:metric",
			"metric:type:gauge",
			"dataset:system",
			fmt.Sprintf("name:%s", name),
			fmt.Sprintf("description:%s", help),
			"unit:value",
		}
		
		// Add label tags
		for k, v := range labels {
			tags = append(tags, fmt.Sprintf("label:%s:%s", k, v))
		}
		
		entity = &models.Entity{
			ID:      metricID,
			Tags:    tags,
			Content: []byte{},
		}
		
		if err := m.repo.Create(entity); err != nil {
			return fmt.Errorf("failed to create gauge metric: %w", err)
		}
	}
	
	// ATOMIC TAG FIX: For gauges, set current value with explicit timestamp
	valueTag := fmt.Sprintf("value:%.6f", value)
	nowNano := time.Now().UnixNano()
	timestampedValueTag := fmt.Sprintf("%d|%s", nowNano, valueTag)
	
	// Update entity atomically
	entity.Tags = append(entity.Tags, timestampedValueTag)
	if updateErr := m.repo.Update(entity); updateErr != nil {
		return fmt.Errorf("failed to update gauge: %w", updateErr)
	}
	
	logger.Trace("Recorded gauge %s = %.6f", name, value)
	return nil
}

// RecordHistogram records a histogram observation
func (m *MetricsTypeManager) RecordHistogram(name string, value float64, labels map[string]string, help string) error {
	metricID := m.buildMetricID(name, labels)
	
	// Get or create metric entity
	entity, err := m.repo.GetByID(metricID)
	if err != nil {
		// Create new histogram metric
		tags := []string{
			"type:metric",
			"metric:type:histogram",
			"dataset:system",
			fmt.Sprintf("name:%s", name),
			fmt.Sprintf("description:%s", help),
			"unit:value",
		}
		
		// Add label tags
		for k, v := range labels {
			tags = append(tags, fmt.Sprintf("label:%s:%s", k, v))
		}
		
		// Add bucket configuration
		bucketStr := ""
		for i, b := range m.histogramBuckets {
			if i > 0 {
				bucketStr += ","
			}
			bucketStr += fmt.Sprintf("%.3f", b)
		}
		tags = append(tags, fmt.Sprintf("histogram:buckets:%s", bucketStr))
		
		entity = &models.Entity{
			ID:      metricID,
			Tags:    tags,
			Content: []byte{},
		}
		
		if err := m.repo.Create(entity); err != nil {
			return fmt.Errorf("failed to create histogram metric: %w", err)
		}
	}
	
	// ATOMIC TAG FIX: For histograms, store individual observations with explicit timestamp
	// The aggregation will calculate buckets and percentiles
	obsTag := fmt.Sprintf("observation:%.6f", value)
	nowNano := time.Now().UnixNano()
	timestampedObsTag := fmt.Sprintf("%d|%s", nowNano, obsTag)
	
	// Get entity and update atomically
	entity, getErr := m.repo.GetByID(metricID)
	if getErr != nil {
		return fmt.Errorf("failed to get histogram entity: %w", getErr)
	}
	entity.Tags = append(entity.Tags, timestampedObsTag)
	if updateErr := m.repo.Update(entity); updateErr != nil {
		return fmt.Errorf("failed to record histogram observation: %w", updateErr)
	}
	
	logger.Trace("Recorded histogram observation %s = %.6f", name, value)
	return nil
}

// GetCounterRate calculates the rate of change for a counter over a time period
func (m *MetricsTypeManager) GetCounterRate(name string, labels map[string]string, duration time.Duration) (float64, error) {
	metricID := m.buildMetricID(name, labels)
	
	entity, err := m.repo.GetByID(metricID)
	if err != nil {
		return 0, fmt.Errorf("metric not found: %w", err)
	}
	
	// Check if it's a counter
	isCounter := false
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if tag == "metric:type:counter" {
			isCounter = true
			break
		}
	}
	
	if !isCounter {
		return 0, fmt.Errorf("metric %s is not a counter", name)
	}
	
	// Get values within the time period
	now := time.Now()
	startTime := now.Add(-duration)
	
	values := []MetricValue{}
	for _, tag := range entity.Tags {
		if idx := strings.Index(tag, "|"); idx != -1 {
			timestampStr := tag[:idx]
			actualTag := tag[idx+1:]
			
			if strings.HasPrefix(actualTag, "value:") {
				if timestamp, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
					t := time.Unix(0, timestamp)
					if t.After(startTime) && t.Before(now) {
						if value, err := strconv.ParseFloat(strings.TrimPrefix(actualTag, "value:"), 64); err == nil {
							values = append(values, MetricValue{
								Value:     value,
								Timestamp: t,
							})
						}
					}
				}
			}
		}
	}
	
	if len(values) < 2 {
		return 0, fmt.Errorf("insufficient data points for rate calculation")
	}
	
	// Sort by timestamp
	sort.Slice(values, func(i, j int) bool {
		return values[i].Timestamp.Before(values[j].Timestamp)
	})
	
	// Calculate rate (change per second)
	firstValue := values[0]
	lastValue := values[len(values)-1]
	
	timeDiff := lastValue.Timestamp.Sub(firstValue.Timestamp).Seconds()
	if timeDiff <= 0 {
		return 0, fmt.Errorf("invalid time difference")
	}
	
	valueDiff := lastValue.Value - firstValue.Value
	rate := valueDiff / timeDiff
	
	return rate, nil
}

// GetHistogramPercentiles calculates percentiles for a histogram
func (m *MetricsTypeManager) GetHistogramPercentiles(name string, labels map[string]string, percentiles []float64, duration time.Duration) (map[float64]float64, error) {
	metricID := m.buildMetricID(name, labels)
	
	entity, err := m.repo.GetByID(metricID)
	if err != nil {
		return nil, fmt.Errorf("metric not found: %w", err)
	}
	
	// Collect observations within the time period
	now := time.Now()
	startTime := now.Add(-duration)
	
	observations := []float64{}
	for _, tag := range entity.Tags {
		if idx := strings.Index(tag, "|"); idx != -1 {
			timestampStr := tag[:idx]
			actualTag := tag[idx+1:]
			
			if strings.HasPrefix(actualTag, "observation:") {
				if timestamp, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
					t := time.Unix(0, timestamp)
					if t.After(startTime) && t.Before(now) {
						if value, err := strconv.ParseFloat(strings.TrimPrefix(actualTag, "observation:"), 64); err == nil {
							observations = append(observations, value)
						}
					}
				}
			}
		}
	}
	
	if len(observations) == 0 {
		return nil, fmt.Errorf("no observations found")
	}
	
	// Sort observations
	sort.Float64s(observations)
	
	// Calculate percentiles
	result := make(map[float64]float64)
	for _, p := range percentiles {
		if p < 0 || p > 100 {
			continue
		}
		
		index := int(math.Ceil(float64(len(observations)) * p / 100.0)) - 1
		if index < 0 {
			index = 0
		}
		if index >= len(observations) {
			index = len(observations) - 1
		}
		
		result[p] = observations[index]
	}
	
	return result, nil
}

// buildMetricID creates a consistent metric ID from name and labels
func (m *MetricsTypeManager) buildMetricID(name string, labels map[string]string) string {
	id := "metric_" + name
	
	// Sort label keys for consistent ID
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	for _, k := range keys {
		id += "_" + k + "_" + labels[k]
	}
	
	return id
}

// getLastValue gets the most recent value from a metric entity
func (m *MetricsTypeManager) getLastValue(entity *models.Entity) float64 {
	var lastValue float64
	var lastTimestamp int64
	
	for _, tag := range entity.Tags {
		if idx := strings.Index(tag, "|"); idx != -1 {
			timestampStr := tag[:idx]
			actualTag := tag[idx+1:]
			
			if strings.HasPrefix(actualTag, "value:") {
				if timestamp, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
					if timestamp > lastTimestamp {
						if value, err := strconv.ParseFloat(strings.TrimPrefix(actualTag, "value:"), 64); err == nil {
							lastValue = value
							lastTimestamp = timestamp
						}
					}
				}
			}
		}
	}
	
	return lastValue
}