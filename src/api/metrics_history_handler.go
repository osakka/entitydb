package api

import (
	"entitydb/models"
	"entitydb/storage/binary"
	"entitydb/logger"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MetricsHistoryHandler handles historical metrics queries
type MetricsHistoryHandler struct {
	repo models.EntityRepository
}

// NewMetricsHistoryHandler creates a new metrics history handler
func NewMetricsHistoryHandler(repo models.EntityRepository) *MetricsHistoryHandler {
	return &MetricsHistoryHandler{repo: repo}
}

// MetricDataPoint represents a single metric value at a point in time
type MetricDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// MetricHistoryResponse represents the response for metric history queries
type MetricHistoryResponse struct {
	MetricName string            `json:"metric_name"`
	Unit       string            `json:"unit"`
	DataPoints []MetricDataPoint `json:"data_points"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Count      int               `json:"count"`
}

// GetMetricHistory retrieves historical values for a specific metric
// @Summary Get metric history
// @Description Retrieve historical values for a specific metric
// @Tags metrics
// @Accept json
// @Produce json
// @Param metric_name query string true "Metric name (e.g., memory_alloc, entity_count_total)"
// @Param hours query int false "Number of hours to look back (default: 24)"
// @Param limit query int false "Maximum number of data points (default: 100)"
// @Param aggregation query string false "Aggregation level: raw, 1min, 1hour, 1day (default: raw)"
// @Success 200 {object} MetricHistoryResponse
// @Router /api/v1/metrics/history [get]
func (h *MetricsHistoryHandler) GetMetricHistory(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	metricName := r.URL.Query().Get("metric_name")
	if metricName == "" {
		RespondError(w, http.StatusBadRequest, "metric_name is required")
		return
	}
	
	// Parse hours parameter (default to 24)
	hours := 24
	if hoursStr := r.URL.Query().Get("hours"); hoursStr != "" {
		if parsed, err := strconv.Atoi(hoursStr); err == nil && parsed > 0 {
			hours = parsed
		}
	}
	
	// Parse limit parameter (default to 100)
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	
	// Parse aggregation level
	aggregation := r.URL.Query().Get("aggregation")
	if aggregation == "" {
		aggregation = "raw"
	}
	
	// Calculate time range
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)
	
	logger.Debug("Fetching metric history: name=%s, hours=%d, limit=%d, aggregation=%s", metricName, hours, limit, aggregation)
	
	// Get the metric entity - check for aggregated version first if requested
	metricID := fmt.Sprintf("metric_%s", metricName)
	if aggregation != "raw" {
		metricID = fmt.Sprintf("metric_%s_agg_%s", metricName, aggregation)
	}
	entity, err := h.repo.GetByID(metricID)
	if err != nil {
		logger.Warn("Metric entity not found: %s", metricID)
		RespondError(w, http.StatusNotFound, fmt.Sprintf("Metric '%s' not found", metricName))
		return
	}
	
	// Get entity repository - handle wrapped repositories
	// All temporal functionality is now merged into the base EntityRepository
	var entityRepo *binary.EntityRepository
	switch repo := h.repo.(type) {
	case *binary.EntityRepository:
		entityRepo = repo
	case *binary.CachedRepository:
		// CachedRepository wraps another repository
		if er, ok := repo.EntityRepository.(*binary.EntityRepository); ok {
			entityRepo = er
		}
	}
	
	if entityRepo == nil {
		RespondError(w, http.StatusInternalServerError, "Temporal features not available")
		return
	}
	
	// Get entity history
	history, err := entityRepo.GetEntityHistory(metricID, limit*2) // Get more to filter by time
	if err != nil {
		logger.Error("Failed to get metric history for %s: %v", metricID, err)
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve metric history")
		return
	}
	
	// Extract metric values from history
	dataPoints := []MetricDataPoint{}
	unit := ""
	
	// Find unit from current entity
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if strings.HasPrefix(tag, "unit:") {
			unit = strings.TrimPrefix(tag, "unit:")
			break
		}
	}
	
	// Process each historical version
	for _, change := range history {
		// Check if within time range
		changeTime := time.Unix(0, change.Timestamp)
		if changeTime.Before(startTime) || changeTime.After(endTime) {
			continue
		}
		
		// EntityChange doesn't have tags - we need to parse from entity's current tags
		// Skip processing history for now and just use entity's temporal tags
		
		// Limit results
		if len(dataPoints) >= limit {
			break
		}
	}
	
	// Extract metric values from entity's temporal tags
	for _, tag := range entity.Tags {
		// Handle temporal tags
		actualTag := tag
		tagTime := time.Now()
		
		if idx := strings.LastIndex(tag, "|"); idx != -1 {
			// Extract timestamp and tag
			parts := strings.SplitN(tag, "|", 2)
			if len(parts) == 2 {
				if ts, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					actualTag = parts[1]
					tagTime = time.Unix(0, ts)
				}
			}
		}
		
		// Check if within time range
		if tagTime.Before(startTime) || tagTime.After(endTime) {
			continue
		}
		
		// Look for value tags
		if strings.HasPrefix(actualTag, "value:") {
			valueStr := strings.TrimPrefix(actualTag, "value:")
			
			// Handle aggregated values (format: value:avg:min:max:count)
			var value float64
			if strings.Contains(valueStr, ":") {
				// Extract average value from aggregated format
				parts := strings.Split(valueStr, ":")
				if len(parts) >= 2 {
					valueStr = parts[0] // Use the first value (average)
				}
			}
			
			if v, err := strconv.ParseFloat(valueStr, 64); err == nil {
				value = v
				dataPoints = append(dataPoints, MetricDataPoint{
					Timestamp: tagTime,
					Value:     value,
				})
			}
		}
	}
	
	// If no historical data, get current value
	if len(dataPoints) == 0 {
		// Extract current value from entity
		for _, tag := range entity.GetTagsWithoutTimestamp() {
			if strings.HasPrefix(tag, "value:") {
				valueStr := strings.TrimPrefix(tag, "value:")
				if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
					dataPoints = append(dataPoints, MetricDataPoint{
						Timestamp: time.Now(),
						Value:     value,
					})
					break
				}
			}
		}
	}
	
	// Sort data points by timestamp (oldest first)
	for i := 0; i < len(dataPoints)-1; i++ {
		for j := i + 1; j < len(dataPoints); j++ {
			if dataPoints[i].Timestamp.After(dataPoints[j].Timestamp) {
				dataPoints[i], dataPoints[j] = dataPoints[j], dataPoints[i]
			}
		}
	}
	
	// Create response
	response := MetricHistoryResponse{
		MetricName: metricName,
		Unit:       unit,
		DataPoints: dataPoints,
		StartTime:  startTime,
		EndTime:    endTime,
		Count:      len(dataPoints),
	}
	
	logger.Debug("Returning %d data points for metric %s", len(dataPoints), metricName)
	RespondJSON(w, http.StatusOK, response)
}

// GetAvailableMetrics returns a list of all available metrics
// @Summary List available metrics
// @Description Get a list of all metrics being collected
// @Tags metrics
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Router /api/v1/metrics/available [get]
func (h *MetricsHistoryHandler) GetAvailableMetrics(w http.ResponseWriter, r *http.Request) {
	// Get all metric entities
	metrics, err := h.repo.ListByTag("type:metric")
	if err != nil {
		logger.Error("Failed to list metrics: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to list metrics")
		return
	}
	
	// Extract metric names
	metricNames := []string{}
	seen := make(map[string]bool)
	
	for _, metric := range metrics {
		for _, tag := range metric.GetTagsWithoutTimestamp() {
			if strings.HasPrefix(tag, "name:") {
				name := strings.TrimPrefix(tag, "name:")
				if !seen[name] {
					metricNames = append(metricNames, name)
					seen[name] = true
				}
				break
			}
		}
	}
	
	// Sort metric names
	for i := 0; i < len(metricNames)-1; i++ {
		for j := i + 1; j < len(metricNames); j++ {
			if metricNames[i] > metricNames[j] {
				metricNames[i], metricNames[j] = metricNames[j], metricNames[i]
			}
		}
	}
	
	RespondJSON(w, http.StatusOK, metricNames)
}