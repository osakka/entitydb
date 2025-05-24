package api

import (
	"encoding/json"
	"entitydb/models"
	"entitydb/logger"
	"fmt"
	"net/http"
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

	// Generate metric entity ID based on metric name and instance
	metricID := c.generateMetricID(update.MetricName, update.Instance)
	
	// Try to get existing metric entity
	entity, err := c.repo.GetByID(metricID)
	if err != nil {
		// Create new metric entity if it doesn't exist
		entity = &models.Entity{
			ID: metricID,
			Tags: []string{
				"type:metric",
				fmt.Sprintf("metric:name:%s", update.MetricName),
				fmt.Sprintf("metric:instance:%s", update.Instance),
			},
			Content: c.createMetricContent(update),
		}
		
		// Add label tags
		for k, v := range update.Labels {
			entity.Tags = append(entity.Tags, fmt.Sprintf("metric:label:%s:%s", k, v))
		}
		
		if err := c.repo.Create(entity); err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to create metric entity: " + err.Error())
			return
		}
		
		logger.Info("Created new metric entity: %s", metricID)
	}
	
	// Add new temporal value tag
	// This is the KEY - we're adding a NEW tag with the current timestamp
	// EntityDB will automatically timestamp this tag!
	valueTag := fmt.Sprintf("metric:value:%.2f:%s", update.Value, update.Unit)
	
	// Also add a snapshot tag for easy querying of current value
	snapshotTag := fmt.Sprintf("metric:current:%s:%.2f:%s", update.MetricName, update.Value, update.Unit)
	
	// Remove old snapshot tag if exists (handle temporal tags)
	filteredTags := []string{}
	for _, tag := range entity.Tags {
		// Extract actual tag from temporal format
		actualTag := tag
		if idx := strings.LastIndex(tag, "|"); idx != -1 {
			actualTag = tag[idx+1:]
		}
		
		// Keep all tags except old snapshot
		if !strings.HasPrefix(actualTag, "metric:current:") {
			filteredTags = append(filteredTags, tag)
		}
	}
	entity.Tags = append(filteredTags, valueTag, snapshotTag)
	
	// Update content with latest value
	entity.Content = c.createMetricContent(update)
	
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
	
	metricID := c.generateMetricID(metricName, instance)
	
	// Get the metric entity
	entity, err := c.repo.GetByID(metricID)
	if err != nil {
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
		"metric_id": metricID,
		"metric":    metricName,
		"instance":  instance,
		"since":     sinceTime.Format(time.RFC3339),
		"until":     untilTime.Format(time.RFC3339),
		"values":    values,
		"count":     len(values),
	}
	
	RespondJSON(w, http.StatusOK, response)
}

// GetCurrentMetrics returns current values for all metrics
func (c *MetricsCollector) GetCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	// Find all metric entities
	entities, err := c.repo.ListByTag("type:metric")
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to list metrics")
		return
	}
	
	metrics := []map[string]interface{}{}
	
	for _, entity := range entities {
		// Extract current value from snapshot tag
		var metricName, currentValue, unit string
		var value float64
		
		for _, tag := range entity.Tags {
			// Handle temporal tags
			actualTag := tag
			if idx := strings.LastIndex(tag, "|"); idx != -1 {
				actualTag = tag[idx+1:]
			}
			
			if strings.HasPrefix(actualTag, "metric:name:") {
				metricName = strings.TrimPrefix(actualTag, "metric:name:")
			}
			if strings.HasPrefix(actualTag, "metric:current:") {
				// Parse: metric:current:NAME:VALUE:UNIT
				parts := strings.Split(actualTag, ":")
				if len(parts) >= 5 {
					currentValue = parts[3]
					unit = parts[4]
					fmt.Sscanf(currentValue, "%f", &value)
				}
			}
		}
		
		if metricName != "" && currentValue != "" {
			metric := map[string]interface{}{
				"id":        entity.ID,
				"name":      metricName,
				"value":     value,
				"unit":      unit,
				"updated":   entity.UpdatedAt,
			}
			
			// Add instance if present
			for _, tag := range entity.Tags {
				if strings.HasPrefix(tag, "metric:instance:") {
					metric["instance"] = strings.TrimPrefix(tag, "metric:instance:")
					break
				}
			}
			
			metrics = append(metrics, metric)
		}
	}
	
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"metrics": metrics,
		"count":   len(metrics),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Helper functions

func (c *MetricsCollector) generateMetricID(metricName, instance string) string {
	if instance == "" {
		instance = "default"
	}
	return fmt.Sprintf("metric_%s_%s", metricName, instance)
}

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

func (c *MetricsCollector) extractMetricValues(entity *models.Entity, since, until time.Time) []map[string]interface{} {
	values := []map[string]interface{}{}
	
	// Parse temporal tags to extract metric values
	for _, tag := range entity.Tags {
		// Check if it's a value tag
		if strings.HasPrefix(tag, "metric:value:") {
			// Extract timestamp from temporal tag
			parts := strings.SplitN(tag, "|", 2)
			if len(parts) == 2 {
				// Parse timestamp
				timestamp, err := time.Parse(time.RFC3339, parts[0])
				if err != nil {
					continue
				}
				
				// Check time range
				if timestamp.Before(since) || timestamp.After(until) {
					continue
				}
				
				// Parse value and unit from tag
				valueParts := strings.Split(parts[1], ":")
				if len(valueParts) >= 4 { // metric:value:NUMBER:UNIT
					var value float64
					fmt.Sscanf(valueParts[2], "%f", &value)
					
					values = append(values, map[string]interface{}{
						"timestamp": timestamp.Format(time.RFC3339),
						"value":     value,
						"unit":      valueParts[3],
					})
				}
			}
		}
	}
	
	return values
}