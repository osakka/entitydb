package api

import (
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// AsyncMetricsReader reads metrics from the new UUID-based async metrics system
type AsyncMetricsReader struct {
	repo models.EntityRepository
}

// NewAsyncMetricsReader creates a new async metrics reader
func NewAsyncMetricsReader(repo models.EntityRepository) *AsyncMetricsReader {
	return &AsyncMetricsReader{repo: repo}
}

// GetMetricValue retrieves the latest value for a metric by name from async system
func (r *AsyncMetricsReader) GetMetricValue(metricName string) float64 {
	// Find metric entities by name tag
	entities, err := r.repo.ListByTag(fmt.Sprintf("name:%s", metricName))
	if err != nil {
		logger.Debug("AsyncMetricsReader: no entities found for metric %s: %v", metricName, err)
		return 0.0
	}

	if len(entities) == 0 {
		logger.Debug("AsyncMetricsReader: no metric entities found for name %s", metricName)
		return 0.0
	}

	// Get the first metric entity (they should all have the same name)
	metricEntity := entities[0]

	// Find the most recent value tag
	var latestValue float64
	var latestTimestamp int64

	for _, tag := range metricEntity.GetTagsWithoutTimestamp() {
		// Parse temporal tags to find value tags
		if strings.HasPrefix(tag, "value:") {
			valueStr := strings.TrimPrefix(tag, "value:")
			if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
				// Extract timestamp from the original tag with timestamp
				for _, originalTag := range metricEntity.Tags {
					if strings.Contains(originalTag, tag) {
						timestamp := r.extractTimestampFromTag(originalTag)
						if timestamp > latestTimestamp {
							latestTimestamp = timestamp
							latestValue = value
						}
						break
					}
				}
			}
		}
	}

	logger.Debug("AsyncMetricsReader: retrieved value %.2f for metric %s", latestValue, metricName)
	return latestValue
}

// GetMetricHistory retrieves historical values for a metric
func (r *AsyncMetricsReader) GetMetricHistory(metricName string, since time.Time) []MetricDataPoint {
	entities, err := r.repo.ListByTag(fmt.Sprintf("name:%s", metricName))
	if err != nil || len(entities) == 0 {
		return []MetricDataPoint{}
	}

	metricEntity := entities[0]
	var dataPoints []MetricDataPoint

	sinceNanos := since.UnixNano()

	for _, tag := range metricEntity.Tags {
		if strings.Contains(tag, "value:") {
			timestamp := r.extractTimestampFromTag(tag)
			if timestamp >= sinceNanos {
				// Extract value
				parts := strings.Split(tag, "|")
				if len(parts) >= 2 && strings.HasPrefix(parts[1], "value:") {
					valueStr := strings.TrimPrefix(parts[1], "value:")
					if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
						dataPoints = append(dataPoints, MetricDataPoint{
							Timestamp: time.Unix(0, timestamp),
							Value:     value,
						})
					}
				}
			}
		}
	}

	return dataPoints
}

// GetAvailableMetrics returns all available metrics in the async system
func (r *AsyncMetricsReader) GetAvailableMetrics() []string {
	entities, err := r.repo.ListByTag("type:metric")
	if err != nil {
		return []string{}
	}

	var metrics []string
	for _, entity := range entities {
		for _, tag := range entity.GetTagsWithoutTimestamp() {
			if strings.HasPrefix(tag, "name:") {
				metricName := strings.TrimPrefix(tag, "name:")
				metrics = append(metrics, metricName)
				break
			}
		}
	}

	return metrics
}

// extractTimestampFromTag extracts nanosecond timestamp from a temporal tag
func (r *AsyncMetricsReader) extractTimestampFromTag(tag string) int64 {
	parts := strings.Split(tag, "|")
	if len(parts) >= 1 {
		if timestamp, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
			return timestamp
		}
	}
	return 0
}

// Note: MetricDataPoint is already defined in metrics_history_handler.go