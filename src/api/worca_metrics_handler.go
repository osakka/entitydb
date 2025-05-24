package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"entitydb/models"
	"entitydb/logger"
)

// WorcaMetricsHandler handles metrics requests for the Worca dashboard
type WorcaMetricsHandler struct {
	repo models.EntityRepository
}

// NewWorcaMetricsHandler creates a new metrics handler
func NewWorcaMetricsHandler(repo models.EntityRepository) *WorcaMetricsHandler {
	return &WorcaMetricsHandler{repo: repo}
}

// MetricSeries represents a time series for a metric
type MetricSeries struct {
	Name   string                 `json:"name"`
	Labels map[string]string      `json:"labels"`
	Data   []MetricDataPoint      `json:"data"`
}

// MetricDataPoint represents a single data point
type MetricDataPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

// GetDashboardMetrics returns aggregated metrics for the Worca dashboard
func (h *WorcaMetricsHandler) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	logger.Info("[WorcaMetrics] Fetching dashboard metrics")

	// Get all metric entities
	entities, err := h.repo.ListByTag("type:metric")
	if err != nil {
		logger.Error("[WorcaMetrics] Failed to fetch metrics: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to fetch metrics")
		return
	}

	// Group metrics by name and instance
	metricsMap := make(map[string]*MetricSeries)
	
	for _, entity := range entities {
		metricName := ""
		instance := ""
		labels := make(map[string]string)
		
		// Parse tags to extract metric info
		for _, tag := range entity.Tags {
			// Strip timestamp if present
			cleanTag := tag
			if idx := strings.LastIndex(tag, "|"); idx > 0 {
				cleanTag = tag[idx+1:]
			}
			
			if strings.HasPrefix(cleanTag, "metric:name:") {
				metricName = strings.TrimPrefix(cleanTag, "metric:name:")
			} else if strings.HasPrefix(cleanTag, "metric:instance:") {
				instance = strings.TrimPrefix(cleanTag, "metric:instance:")
			} else if strings.HasPrefix(cleanTag, "metric:label:") {
				// Parse labels like metric:label:key:value
				parts := strings.SplitN(strings.TrimPrefix(cleanTag, "metric:label:"), ":", 2)
				if len(parts) == 2 {
					labels[parts[0]] = parts[1]
				}
			}
		}
		
		// Create a unique key for this metric series
		seriesKey := metricName + "_" + instance
		for k, v := range labels {
			seriesKey += "_" + k + ":" + v
		}
		
		// Initialize series if not exists
		if _, exists := metricsMap[seriesKey]; !exists {
			metricsMap[seriesKey] = &MetricSeries{
				Name:   metricName,
				Labels: labels,
				Data:   []MetricDataPoint{},
			}
			metricsMap[seriesKey].Labels["instance"] = instance
		}
		
		// Extract temporal values
		for _, tag := range entity.Tags {
			if strings.Contains(tag, "|") && strings.Contains(tag, "metric:value:") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					timestamp := parts[0]
					valuePart := parts[1]
					
					if strings.HasPrefix(valuePart, "metric:value:") {
						valueStr := strings.TrimPrefix(valuePart, "metric:value:")
						// Remove unit suffix if present
						if colonIdx := strings.Index(valueStr, ":"); colonIdx > 0 {
							valueStr = valueStr[:colonIdx]
						}
						
						var value float64
						if err := json.Unmarshal([]byte(valueStr), &value); err == nil {
							// Convert timestamp to ISO format
							if ts, err := time.Parse(time.RFC3339Nano, timestamp); err == nil {
								metricsMap[seriesKey].Data = append(metricsMap[seriesKey].Data, MetricDataPoint{
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
	var metricsList []MetricSeries
	for _, series := range metricsMap {
		metricsList = append(metricsList, *series)
	}
	
	// Prepare dashboard-specific aggregations
	response := map[string]interface{}{
		"metrics": metricsList,
		"summary": map[string]interface{}{
			"totalMetrics": len(metricsList),
			"lastUpdated":  time.Now().Format(time.RFC3339),
		},
		// Add specific dashboard metrics
		"dashboardData": h.prepareDashboardData(metricsList),
	}
	
	RespondJSON(w, http.StatusOK, response)
}

// prepareDashboardData prepares specific metrics for dashboard charts
func (h *WorcaMetricsHandler) prepareDashboardData(metrics []MetricSeries) map[string]interface{} {
	dashboardData := make(map[string]interface{})
	
	// Task completion trend
	taskCompletionData := []map[string]interface{}{}
	for _, m := range metrics {
		if m.Name == "worca_tasks_completed" {
			for _, dp := range m.Data {
				taskCompletionData = append(taskCompletionData, map[string]interface{}{
					"x": dp.Timestamp,
					"y": dp.Value,
				})
			}
		}
	}
	dashboardData["taskCompletion"] = taskCompletionData
	
	// Team productivity
	productivityData := []map[string]interface{}{}
	for _, m := range metrics {
		if m.Name == "worca_team_productivity" {
			for _, dp := range m.Data {
				productivityData = append(productivityData, map[string]interface{}{
					"x": dp.Timestamp,
					"y": dp.Value,
				})
			}
		}
	}
	dashboardData["teamProductivity"] = productivityData
	
	// Sprint velocity
	velocityData := []map[string]interface{}{}
	for _, m := range metrics {
		if m.Name == "worca_sprint_velocity" {
			sprint := m.Labels["sprint"]
			if len(m.Data) > 0 {
				velocityData = append(velocityData, map[string]interface{}{
					"sprint": sprint,
					"velocity": m.Data[len(m.Data)-1].Value, // Latest value
				})
			}
		}
	}
	dashboardData["sprintVelocity"] = velocityData
	
	// Task status distribution
	statusData := map[string]float64{}
	for _, m := range metrics {
		if m.Name == "worca_task_status" {
			status := m.Labels["status"]
			if len(m.Data) > 0 {
				statusData[status] = m.Data[len(m.Data)-1].Value
			}
		}
	}
	dashboardData["taskStatus"] = statusData
	
	// Member utilization
	utilizationData := []map[string]interface{}{}
	for _, m := range metrics {
		if m.Name == "worca_member_utilization" {
			member := m.Labels["member"]
			if len(m.Data) > 0 {
				utilizationData = append(utilizationData, map[string]interface{}{
					"member": member,
					"utilization": m.Data[len(m.Data)-1].Value,
				})
			}
		}
	}
	dashboardData["memberUtilization"] = utilizationData
	
	// Epic progress
	epicData := []map[string]interface{}{}
	for _, m := range metrics {
		if m.Name == "worca_epic_progress" {
			epic := m.Labels["epic"]
			if len(m.Data) > 0 {
				epicData = append(epicData, map[string]interface{}{
					"epic": epic,
					"progress": m.Data[len(m.Data)-1].Value,
				})
			}
		}
	}
	dashboardData["epicProgress"] = epicData
	
	// Burndown chart
	burndownData := []map[string]interface{}{}
	for _, m := range metrics {
		if m.Name == "worca_sprint_burndown" {
			for _, dp := range m.Data {
				day := m.Labels["day"]
				burndownData = append(burndownData, map[string]interface{}{
					"day": day,
					"remaining": dp.Value,
				})
			}
		}
	}
	dashboardData["burndown"] = burndownData
	
	return dashboardData
}