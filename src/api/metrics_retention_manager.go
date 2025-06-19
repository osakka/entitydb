package api

import (
	"context"
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MetricsRetentionManager manages retention policies for metrics
type MetricsRetentionManager struct {
	repo                  models.EntityRepository
	retentionRaw         time.Duration
	retention1Min        time.Duration
	retention1Hour       time.Duration
	retention1Day        time.Duration
	ctx                  context.Context
	cancel               context.CancelFunc
	mu                   sync.RWMutex
	aggregationRunning   bool
}

// NewMetricsRetentionManager creates a new retention manager
func NewMetricsRetentionManager(repo models.EntityRepository, rawRetention, min1Retention, hour1Retention, day1Retention time.Duration) *MetricsRetentionManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &MetricsRetentionManager{
		repo:           repo,
		retentionRaw:   rawRetention,
		retention1Min:  min1Retention,
		retention1Hour: hour1Retention,
		retention1Day:  day1Retention,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start begins the retention management process
func (m *MetricsRetentionManager) Start() {
	logger.Info("Starting metrics retention manager - Raw: %v, 1min: %v, 1hour: %v, 1day: %v", 
		m.retentionRaw, m.retention1Min, m.retention1Hour, m.retention1Day)
	
	// Run retention cleanup every hour
	go func() {
		// Initial delay to let system stabilize
		select {
		case <-time.After(5 * time.Minute):
		case <-m.ctx.Done():
			return
		}
		
		// Run immediately on start
		m.enforceRetention()
		
		// Then run periodically
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				m.enforceRetention()
			case <-m.ctx.Done():
				logger.Info("Metrics retention manager stopped")
				return
			}
		}
	}()
	
	// Run aggregation every 5 minutes
	go func() {
		// Initial delay to let system stabilize, but shorter than retention delay
		select {
		case <-time.After(2 * time.Minute):
		case <-m.ctx.Done():
			return
		}
		
		// Run immediately on start (before retention kicks in)
		m.performAggregation()
		
		// Then run periodically
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				m.performAggregation()
			case <-m.ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the retention manager
func (m *MetricsRetentionManager) Stop() {
	m.cancel()
}

// enforceRetention removes old metric values based on retention policies
func (m *MetricsRetentionManager) enforceRetention() {
	logger.Info("Starting metrics retention enforcement")
	startTime := time.Now()
	
	// Mark this goroutine as performing metrics operations to prevent recursion
	binary.SetMetricsOperation(true)
	defer binary.SetMetricsOperation(false)
	
	// Get all metric entities
	metrics, err := m.repo.ListByTag("type:metric")
	if err != nil {
		logger.Error("Failed to list metrics for retention: %v", err)
		return
	}
	
	logger.Debug("Processing retention for %d metrics", len(metrics))
	
	cleanedCount := 0
	totalTagsRemoved := 0
	
	for _, metric := range metrics {
		// Skip aggregated metrics (they have their own retention)
		isAggregated := false
		for _, tag := range metric.GetTagsWithoutTimestamp() {
			if strings.HasPrefix(tag, "aggregation:") {
				isAggregated = true
				break
			}
		}
		
		retention := m.retentionRaw
		if isAggregated {
			// Determine retention based on aggregation level
			for _, tag := range metric.GetTagsWithoutTimestamp() {
				if tag == "aggregation:1min" {
					retention = m.retention1Min
				} else if tag == "aggregation:1hour" {
					retention = m.retention1Hour
				} else if tag == "aggregation:1day" {
					retention = m.retention1Day
				}
			}
		}
		
		// Process each tag to find old values
		cutoffTime := time.Now().Add(-retention)
		cutoffNanos := cutoffTime.UnixNano()
		
		tagsToRemove := []string{}
		
		for _, tag := range metric.Tags {
			// Parse temporal tags
			if idx := strings.Index(tag, "|"); idx != -1 {
				timestampStr := tag[:idx]
				actualTag := tag[idx+1:]
				
				// Only process value tags
				if strings.HasPrefix(actualTag, "value:") {
					if timestamp, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
						if timestamp < cutoffNanos {
							tagsToRemove = append(tagsToRemove, tag)
						}
					}
				}
			}
		}
		
		// Remove old tags if any found
		if len(tagsToRemove) > 0 {
			logger.Debug("Removing %d old value tags from metric %s", len(tagsToRemove), metric.ID)
			
			// EntityDB doesn't have a RemoveTag method, so we need to update the entity
			// with filtered tags
			newTags := []string{}
			for _, tag := range metric.Tags {
				found := false
				for _, removeTag := range tagsToRemove {
					if tag == removeTag {
						found = true
						break
					}
				}
				if !found {
					newTags = append(newTags, tag)
				}
			}
			
			metric.Tags = newTags
			if err := m.repo.Update(metric); err != nil {
				logger.Error("Failed to update metric %s for retention: %v", metric.ID, err)
			} else {
				cleanedCount++
				totalTagsRemoved += len(tagsToRemove)
			}
		}
	}
	
	duration := time.Since(startTime)
	logger.Info("Retention enforcement complete: %d metrics cleaned, %d tags removed in %v", 
		cleanedCount, totalTagsRemoved, duration)
}

// performAggregation creates aggregated metrics from raw data
func (m *MetricsRetentionManager) performAggregation() {
	m.mu.Lock()
	if m.aggregationRunning {
		m.mu.Unlock()
		logger.Debug("Aggregation already running, skipping")
		return
	}
	m.aggregationRunning = true
	m.mu.Unlock()
	
	// Mark this goroutine as performing metrics operations to prevent recursion
	binary.SetMetricsOperation(true)
	defer binary.SetMetricsOperation(false)
	
	defer func() {
		m.mu.Lock()
		m.aggregationRunning = false
		m.mu.Unlock()
	}()
	
	logger.Info("Starting metrics aggregation")
	startTime := time.Now()
	
	// Get all raw metrics (non-aggregated)
	metrics, err := m.repo.ListByTag("type:metric")
	if err != nil {
		logger.Error("Failed to list metrics for aggregation: %v", err)
		return
	}
	
	logger.Info("Found %d total metrics for aggregation processing", len(metrics))
	
	aggregatedCount := 0
	rawMetricsCount := 0
	
	for _, metric := range metrics {
		// Skip if already aggregated
		isAggregated := false
		for _, tag := range metric.GetTagsWithoutTimestamp() {
			if strings.HasPrefix(tag, "aggregation:") {
				isAggregated = true
				break
			}
		}
		
		if isAggregated {
			logger.Debug("Skipping aggregated metric: %s", metric.ID)
			continue
		}
		
		rawMetricsCount++
		logger.Debug("Processing raw metric: %s", metric.ID)
		
		// Get metric name
		metricName := ""
		for _, tag := range metric.GetTagsWithoutTimestamp() {
			if strings.HasPrefix(tag, "name:") {
				metricName = strings.TrimPrefix(tag, "name:")
				break
			}
		}
		
		if metricName == "" {
			continue
		}
		
		// Aggregate at different intervals
		if err := m.aggregateMetric(metric, metricName, 1*time.Minute, "1min"); err != nil {
			logger.Warn("Failed to aggregate %s at 1min: %v", metricName, err)
		} else {
			aggregatedCount++
		}
		
		if err := m.aggregateMetric(metric, metricName, 1*time.Hour, "1hour"); err != nil {
			logger.Warn("Failed to aggregate %s at 1hour: %v", metricName, err)
		} else {
			aggregatedCount++
		}
		
		if err := m.aggregateMetric(metric, metricName, 24*time.Hour, "1day"); err != nil {
			logger.Warn("Failed to aggregate %s at 1day: %v", metricName, err)
		} else {
			aggregatedCount++
		}
	}
	
	duration := time.Since(startTime)
	logger.Info("Aggregation complete: %d raw metrics processed, %d aggregations performed in %v", rawMetricsCount, aggregatedCount, duration)
}

// aggregateMetric creates an aggregated version of a metric
func (m *MetricsRetentionManager) aggregateMetric(metric *models.Entity, metricName string, interval time.Duration, intervalName string) error {
	// Mark this operation as metrics-related to prevent recursion
	binary.SetMetricsOperation(true)
	defer binary.SetMetricsOperation(false)
	
	// Create aggregated metric ID
	aggMetricID := fmt.Sprintf("metric_%s_agg_%s", metricName, intervalName)
	
	// Check if aggregated metric exists
	aggMetric, err := m.repo.GetByID(aggMetricID)
	if err != nil {
		// Create new aggregated metric
		aggMetric = &models.Entity{
			ID: aggMetricID,
			Tags: []string{
				"type:metric",
				"dataset:system",
				fmt.Sprintf("name:%s", metricName),
				fmt.Sprintf("aggregation:%s", intervalName),
				fmt.Sprintf("source:%s", metric.ID),
			},
			Content: []byte{},
		}
		
		// Copy other metadata tags
		for _, tag := range metric.GetTagsWithoutTimestamp() {
			if strings.HasPrefix(tag, "unit:") || strings.HasPrefix(tag, "description:") {
				aggMetric.Tags = append(aggMetric.Tags, tag)
			}
		}
		
		if err := m.repo.Create(aggMetric); err != nil {
			return fmt.Errorf("failed to create aggregated metric: %w", err)
		}
	}
	
	// Get latest aggregation timestamp
	latestAggTime := time.Time{}
	for _, tag := range aggMetric.Tags {
		if idx := strings.Index(tag, "|"); idx != -1 {
			if strings.HasPrefix(tag[idx+1:], "value:") {
				if timestamp, err := strconv.ParseInt(tag[:idx], 10, 64); err == nil {
					t := time.Unix(0, timestamp)
					if t.After(latestAggTime) {
						latestAggTime = t
					}
				}
			}
		}
	}
	
	// Collect values to aggregate since last aggregation
	now := time.Now()
	buckets := make(map[time.Time][]float64)
	
	for _, tag := range metric.Tags {
		if idx := strings.Index(tag, "|"); idx != -1 {
			timestampStr := tag[:idx]
			actualTag := tag[idx+1:]
			
			if strings.HasPrefix(actualTag, "value:") {
				if timestamp, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
					t := time.Unix(0, timestamp)
					
					// Skip if before last aggregation
					if !latestAggTime.IsZero() && t.Before(latestAggTime) {
						continue
					}
					
					// Skip if too recent for complete bucket
					if t.After(now.Add(-interval)) {
						continue
					}
					
					// Parse value
					valueStr := strings.TrimPrefix(actualTag, "value:")
					if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
						// Round down to interval
						bucket := t.Truncate(interval)
						buckets[bucket] = append(buckets[bucket], value)
					}
				}
			}
		}
	}
	
	// Create aggregated values
	for bucket, values := range buckets {
		if len(values) == 0 {
			continue
		}
		
		// Calculate aggregates
		sum := 0.0
		min := values[0]
		max := values[0]
		
		for _, v := range values {
			sum += v
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		
		avg := sum / float64(len(values))
		
		// Add aggregated value tag with bucket information
		aggTag := fmt.Sprintf("value:%.2f:avg:%.2f:min:%.2f:max:%.2f:count:%d:bucket:%d", 
			avg, avg, min, max, len(values), bucket.Unix())
		
		// Add the aggregated value (will get current timestamp, but includes bucket time in value)
		if err := m.repo.AddTag(aggMetricID, aggTag); err != nil {
			logger.Warn("Failed to add aggregated tag for %s: %v", aggMetricID, err)
		}
	}
	
	return nil
}