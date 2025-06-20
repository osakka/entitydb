package api

import (
	"encoding/json"
	"entitydb/models"
	"entitydb/logger"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MetricsCollector handles temporal metrics collection
type MetricsCollector struct {
	repo models.EntityRepository
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(repo models.EntityRepository) *MetricsCollector {
	return &MetricsCollector{repo: repo}
}

// MetricUpdate represents a metric value update
type MetricUpdate struct {
	MetricName string  `json:"metric_name" binding:"required"` // e.g., "cpu_usage", "memory_used"
	Value      float64 `json:"value" binding:"required"`
	Unit       string  `json:"unit" binding:"required"`       // e.g., "percent", "bytes", "requests"
	Instance   string  `json:"instance"`                       // e.g., "server1", "app2"
	Labels     map[string]string `json:"labels"`           // Additional labels
}

// CollectMetric updates a metric entity with a new temporal value
func (c *MetricsCollector) CollectMetric(w http.ResponseWriter, r *http.Request) {
	var update MetricUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid metric update: " + err.Error())
		return
	}

	// Try to find existing metric entity by searching for name and instance tags
	var entity *models.Entity
	var metricID string
	
	// Search for existing entity by metric name and instance
	nameTagEntities, err := c.repo.ListByTag(fmt.Sprintf("metric:name:%s", update.MetricName))
	var found bool
	
	if err == nil {
		// Look for entity with matching instance
		for _, e := range nameTagEntities {
			cleanTags := e.GetTagsWithoutTimestamp()
			instanceTag := fmt.Sprintf("metric:instance:%s", update.Instance)
			
			for _, tag := range cleanTags {
				if tag == instanceTag {
					entity = e
					metricID = e.ID
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}
	
	if !found {
		// Create new metric entity using UUID architecture
		additionalTags := []string{
			fmt.Sprintf("metric:name:%s", update.MetricName),
			fmt.Sprintf("metric:instance:%s", update.Instance),
		}
		
		// Add label tags
		for k, v := range update.Labels {
			additionalTags = append(additionalTags, fmt.Sprintf("metric:label:%s:%s", k, v))
		}
		
		newEntity, err := models.NewEntityWithMandatoryTags(
			"metric",                    // entityType
			"system",                    // dataset
			models.SystemUserID,         // createdBy (system user)
			additionalTags,             // additional tags
		)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to create metric entity: " + err.Error())
			return
		}
		
		newEntity.Content = c.createMetricContent(update)
		
		if err := c.repo.Create(newEntity); err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to store metric entity: " + err.Error())
			return
		}
		
		entity = newEntity
		metricID = newEntity.ID
		logger.Info("Created new metric entity with UUID: %s", metricID)
	}
	
	// ATOMIC TAG FIX: Create all new tags with same timestamp to prevent temporal explosion
	valueTag := fmt.Sprintf("metric:value:%.2f:%s", update.Value, update.Unit)
	snapshotTag := fmt.Sprintf("metric:current:%s:%.2f:%s", update.MetricName, update.Value, update.Unit)
	
	// Remove old snapshot tag if exists (handle temporal tags)
	filteredTags := []string{}
	for _, tag := range entity.Tags {
		// Extract actual tag from temporal format
		actualTag := tag
		if idx := strings.LastIndex(tag, "|"); idx != -1 {
			actualTag = tag[idx+1:]
		}
		
		// Keep all tags except old snapshot and value tags
		if !strings.HasPrefix(actualTag, "metric:current:") && !strings.HasPrefix(actualTag, "metric:value:") {
			filteredTags = append(filteredTags, tag)
		}
	}
	
	// CRITICAL FIX: Use same timestamp for all new tags to prevent temporal explosion
	now := time.Now().UnixNano()
	timestampedValueTag := fmt.Sprintf("%d|%s", now, valueTag)
	timestampedSnapshotTag := fmt.Sprintf("%d|%s", now, snapshotTag)
	
	// Set all tags atomically with same timestamp
	entity.Tags = append(filteredTags, timestampedValueTag, timestampedSnapshotTag)
	entity.Content = c.createMetricContent(update)
	
	// Use direct update to avoid AddTag() temporal explosion
	if err := c.repo.Update(entity); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update metric: " + err.Error())
		return
	}
	
	response := map[string]interface{}{
		"metric_id": metricID,
		"metric":    update.MetricName,
		"value":     update.Value,
		"unit":      update.Unit,
		"timestamp": time.Now().Format(time.RFC3339),
		"message":   "Metric value recorded with temporal tag",
	}
	
	RespondJSON(w, http.StatusOK, response)
}

// GetMetricHistory retrieves the temporal history of a metric
func (c *MetricsCollector) GetMetricHistory(w http.ResponseWriter, r *http.Request) {
	metricName := r.URL.Query().Get("metric")
	instance := r.URL.Query().Get("instance")
	since := r.URL.Query().Get("since") // RFC3339 timestamp
	until := r.URL.Query().Get("until")   // RFC3339 timestamp
	
	if metricName == "" {
		RespondError(w, http.StatusBadRequest, "metric parameter is required")
		return
	}
	
	// Find metric entity by searching for name and instance tags
	nameTagEntities, err := c.repo.ListByTag(fmt.Sprintf("metric:name:%s", metricName))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to search for metric")
		return
	}
	
	var entity *models.Entity
	var found bool
	
	if instance == "" {
		instance = "default"
	}
	
	// Look for entity with matching instance
	for _, e := range nameTagEntities {
		cleanTags := e.GetTagsWithoutTimestamp()
		instanceTag := fmt.Sprintf("metric:instance:%s", instance)
		
		for _, tag := range cleanTags {
			if tag == instanceTag {
				entity = e
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	
	if !found {
		RespondError(w, http.StatusNotFound, "Metric not found")
		return
	}
	
	// Parse time range
	var sinceTime, untilTime time.Time
	if since != "" {
		sinceTime, _ = time.Parse(time.RFC3339, since)
	} else {
		sinceTime = time.Now().Add(-24 * time.Hour) // Default: last 24 hours
	}
	if until != "" {
		untilTime, _ = time.Parse(time.RFC3339, until)
	} else {
		untilTime = time.Now()
	}
	
	// Extract metric values from temporal tags
	values := c.extractMetricValues(entity, sinceTime, untilTime)
	
	response := map[string]interface{}{
		"metric":    metricName,
		"instance":  instance,
		"since":     sinceTime.Format(time.RFC3339),
		"until":     untilTime.Format(time.RFC3339),
		"values":    values,
		"count":     len(values),
	}
	
	RespondJSON(w, http.StatusOK, response)
}

// GetCurrentMetrics retrieves the current value for all metrics
func (c *MetricsCollector) GetCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	// Get all metric entities
	metricEntities, err := c.repo.ListByTag("type:metric")
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch metrics")
		return
	}
	
	metrics := make(map[string]interface{})
	
	for _, entity := range metricEntities {
		// Extract metric name and current value
		var metricName string
		var currentValue float64
		var unit string
		
		for _, tag := range entity.GetTagsWithoutTimestamp() {
			if strings.HasPrefix(tag, "metric:name:") {
				metricName = strings.TrimPrefix(tag, "metric:name:")
			} else if strings.HasPrefix(tag, "metric:current:") {
				// Parse current value from snapshot tag
				parts := strings.Split(strings.TrimPrefix(tag, "metric:current:"), ":")
				if len(parts) >= 2 {
					currentValue, _ = strconv.ParseFloat(parts[1], 64)
					if len(parts) >= 3 {
						unit = parts[2]
					}
				}
			}
		}
		
		if metricName != "" {
			metrics[metricName] = map[string]interface{}{
				"value": currentValue,
				"unit":  unit,
			}
		}
	}
	
	RespondJSON(w, http.StatusOK, metrics)
}

// Helper functions

func (c *MetricsCollector) createMetricContent(update MetricUpdate) []byte {
	content := map[string]interface{}{
		"metric":     update.MetricName,
		"value":      update.Value,
		"unit":       update.Unit,
		"instance":   update.Instance,
		"labels":     update.Labels,
		"updated_at": time.Now().Format(time.RFC3339),
	}
	data, _ := json.Marshal(content)
	return data
}

// Helper method to extract metric values from temporal tags
func (c *MetricsCollector) extractMetricValues(entity *models.Entity, since, until time.Time) []map[string]interface{} {
	values := []map[string]interface{}{}
	
	// This would need the temporal repository to get historical tags
	// For now, return current value
	for _, tag := range entity.Tags {
		// Look for temporal value tags
		if strings.Contains(tag, "|metric:value:") {
			parts := strings.Split(tag, "|")
			if len(parts) == 2 {
				// Extract timestamp
				timestampNano, err := strconv.ParseInt(parts[0], 10, 64)
				if err != nil {
					continue
				}
				timestamp := time.Unix(0, timestampNano)
				
				// Check if within time range
				if timestamp.Before(since) || timestamp.After(until) {
					continue
				}
				
				// Extract value
				valueParts := strings.Split(parts[1], ":")
				if len(valueParts) >= 3 {
					value, err := strconv.ParseFloat(valueParts[2], 64)
					if err == nil {
						values = append(values, map[string]interface{}{
							"timestamp": timestamp.Format(time.RFC3339),
							"value":     value,
						})
					}
				}
			}
		}
	}
	
	return values
}