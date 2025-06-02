package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"entitydb/models"
	"entitydb/logger"
)

// ApplicationMetricsHandler handles generic application metrics requests
// This is a generic handler that applications can use to retrieve their metrics
type ApplicationMetricsHandler struct {
	repo models.EntityRepository
}

// NewApplicationMetricsHandler creates a new application metrics handler
func NewApplicationMetricsHandler(repo models.EntityRepository) *ApplicationMetricsHandler {
	return &ApplicationMetricsHandler{repo: repo}
}

// ApplicationMetricSeries represents a time series for a metric
type ApplicationMetricSeries struct {
	Name   string                 `json:"name"`
	Labels map[string]string      `json:"labels"`
	Data   []ApplicationDataPoint `json:"data"`
}

// ApplicationDataPoint represents a single data point
type ApplicationDataPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

// GetApplicationMetrics returns metrics filtered by application namespace
func (h *ApplicationMetricsHandler) GetApplicationMetrics(w http.ResponseWriter, r *http.Request) {
	// Get application namespace from query parameter
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = r.URL.Query().Get("app") // Support both 'namespace' and 'app' parameters
	}
	
	logger.Info("Fetching application metrics for namespace: %s", namespace)

	// Build tag filter
	tagFilter := "type:metric"
	if namespace != "" {
		// Application can store metrics with app:<namespace> tag
		tagFilter = "app:" + namespace
	}

	// Get metric entities
	entities, err := h.repo.ListByTag(tagFilter)
	if err != nil {
		logger.Error("Failed to fetch metrics: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to fetch metrics")
		return
	}

	// Group metrics by name and labels
	metricsMap := make(map[string]*ApplicationMetricSeries)
	
	for _, entity := range entities {
		// Only process metric entities
		hasMetricType := false
		for _, tag := range entity.Tags {
			cleanTag := tag
			if idx := strings.LastIndex(tag, "|"); idx > 0 {
				cleanTag = tag[idx+1:]
			}
			if cleanTag == "type:metric" {
				hasMetricType = true
				break
			}
		}
		
		if !hasMetricType {
			continue
		}

		metricName := ""
		labels := make(map[string]string)
		
		// Parse tags to extract metric info
		for _, tag := range entity.Tags {
			// Strip timestamp if present
			cleanTag := tag
			if idx := strings.LastIndex(tag, "|"); idx > 0 {
				cleanTag = tag[idx+1:]
			}
			
			if strings.HasPrefix(cleanTag, "name:") {
				metricName = strings.TrimPrefix(cleanTag, "name:")
			} else if strings.HasPrefix(cleanTag, "label:") {
				// Parse labels like label:key:value
				parts := strings.SplitN(strings.TrimPrefix(cleanTag, "label:"), ":", 2)
				if len(parts) == 2 {
					labels[parts[0]] = parts[1]
				}
			}
		}
		
		// Skip if no metric name
		if metricName == "" {
			continue
		}
		
		// Create a unique key for this metric series
		seriesKey := metricName
		for k, v := range labels {
			seriesKey += "_" + k + ":" + v
		}
		
		// Initialize series if not exists
		if _, exists := metricsMap[seriesKey]; !exists {
			metricsMap[seriesKey] = &ApplicationMetricSeries{
				Name:   metricName,
				Labels: labels,
				Data:   []ApplicationDataPoint{},
			}
		}
		
		// Extract temporal values
		for _, tag := range entity.Tags {
			if strings.Contains(tag, "|") && strings.Contains(tag, "value:") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					timestampStr := parts[0]
					valuePart := parts[1]
					
					if strings.HasPrefix(valuePart, "value:") {
						valueStr := strings.TrimPrefix(valuePart, "value:")
						
						var value float64
						if err := json.Unmarshal([]byte(valueStr), &value); err == nil {
							// Parse timestamp (it's in nanoseconds)
							if timestampNanos, err := json.Number(timestampStr).Int64(); err == nil {
								ts := time.Unix(0, timestampNanos)
								metricsMap[seriesKey].Data = append(metricsMap[seriesKey].Data, ApplicationDataPoint{
									Timestamp: ts.Format(time.RFC3339),
									Value:     value,
								})
							}
						}
					}
				}
			}
		}
	}
	
	// Convert map to slice
	var metricsList []ApplicationMetricSeries
	for _, series := range metricsMap {
		if len(series.Data) > 0 {
			metricsList = append(metricsList, *series)
		}
	}
	
	// Generic response that applications can process as needed
	response := map[string]interface{}{
		"metrics": metricsList,
		"summary": map[string]interface{}{
			"totalMetrics": len(metricsList),
			"namespace":    namespace,
			"lastUpdated":  time.Now().Format(time.RFC3339),
		},
	}
	
	RespondJSON(w, http.StatusOK, response)
}