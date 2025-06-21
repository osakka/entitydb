package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Test configuration
const (
	BaseURL = "https://localhost:8085"
	TestUsername = "admin"
	TestPassword = "admin"
	TestTimeout = 30 * time.Second
)

// Response structures
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id"`
	User      struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		Email    string   `json:"email"`
		Roles    []string `json:"roles"`
	} `json:"user"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
	Version   string `json:"version"`
	Checks    struct {
		Database string `json:"database"`
	} `json:"checks"`
	Metrics struct {
		EntityCount int `json:"entity_count"`
		UserCount   int `json:"user_count"`
		DatabaseSizeBytes int64 `json:"database_size_bytes"`
		MemoryUsage struct {
			AllocBytes      int64 `json:"alloc_bytes"`
			TotalAllocBytes int64 `json:"total_alloc_bytes"`
			SysBytes        int64 `json:"sys_bytes"`
			NumGC           int   `json:"num_gc"`
		} `json:"memory_usage"`
		Goroutines int `json:"goroutines"`
	} `json:"metrics"`
}

type SystemMetricsResponse struct {
	System struct {
		Version     string  `json:"version"`
		GoVersion   string  `json:"go_version"`
		Uptime      int64   `json:"uptime"`
		UptimeSeconds float64 `json:"uptime_seconds"`
		NumCPU      int     `json:"num_cpu"`
		NumGoroutines int   `json:"num_goroutines"`
	} `json:"system"`
	Database struct {
		TotalEntities int            `json:"total_entities"`
		EntitiesByType map[string]int `json:"entities_by_type"`
		TagsTotal     int            `json:"tags_total"`
	} `json:"database"`
	Memory struct {
		AllocBytes     int64 `json:"alloc_bytes"`
		TotalAllocBytes int64 `json:"total_alloc_bytes"`
		SysBytes       int64 `json:"sys_bytes"`
	} `json:"memory"`
	Storage struct {
		DatabaseSizeBytes int64 `json:"database_size_bytes"`
		WALSizeBytes      int64 `json:"wal_size_bytes"`
		ReadOperations    int   `json:"read_operations"`
		WriteOperations   int   `json:"write_operations"`
	} `json:"storage"`
}

type MetricHistoryResponse struct {
	MetricName  string `json:"metric_name"`
	Unit        string `json:"unit"`
	DataPoints  []struct {
		Timestamp string  `json:"timestamp"`
		Value     float64 `json:"value"`
	} `json:"data_points"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Count     int    `json:"count"`
}

type RBACMetricsResponse struct {
	Timestamp string `json:"timestamp"`
	Users     struct {
		TotalUsers int `json:"total_users"`
		AdminCount int `json:"admin_count"`
	} `json:"users"`
	Auth struct {
		SuccessfulLogins int     `json:"successful_logins"`
		FailedLogins     int     `json:"failed_logins"`
		SuccessRate      float64 `json:"success_rate"`
	} `json:"auth"`
	Sessions struct {
		ActiveCount    int `json:"active_count"`
		TotalToday     int `json:"total_today"`
		AvgDurationMS  int `json:"avg_duration_ms"`
	} `json:"sessions"`
}

type TestResult struct {
	TestName string
	Success  bool
	Duration time.Duration
	Error    string
	Details  map[string]interface{}
}

type MetricsTestSuite struct {
	client  *http.Client
	results []TestResult
	baseURL string
}

func NewMetricsTestSuite() *MetricsTestSuite {
	// Create HTTP client that ignores SSL certificates for testing
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   TestTimeout,
	}

	return &MetricsTestSuite{
		client:  client,
		results: make([]TestResult, 0),
		baseURL: BaseURL,
	}
}

func (s *MetricsTestSuite) makeRequest(method, endpoint string, token string) (*http.Response, error) {
	req, err := http.NewRequest(method, s.baseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return s.client.Do(req)
}

func (s *MetricsTestSuite) addResult(testName string, success bool, duration time.Duration, err error, details map[string]interface{}) {
	result := TestResult{
		TestName: testName,
		Success:  success,
		Duration: duration,
		Details:  details,
	}
	if err != nil {
		result.Error = err.Error()
	}
	s.results = append(s.results, result)
}

func (s *MetricsTestSuite) getAuthToken() (string, error) {
	loginData := `{"username":"` + TestUsername + `","password":"` + TestPassword + `"}`
	
	// Use the same client with SSL bypass for authentication
	req, err := http.NewRequest("POST", s.baseURL+"/api/v1/auth/login", strings.NewReader(loginData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}

	return loginResp.Token, nil
}

// Test 1: Health Endpoint
func (s *MetricsTestSuite) testHealthEndpoint() {
	start := time.Now()
	testName := "Health Endpoint"

	resp, err := s.makeRequest("GET", "/health", "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("expected 200, got %d", resp.StatusCode), nil)
		return
	}

	var healthResp HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"status":       healthResp.Status,
		"entity_count": healthResp.Metrics.EntityCount,
		"user_count":   healthResp.Metrics.UserCount,
		"database_size": healthResp.Metrics.DatabaseSizeBytes,
		"memory_alloc": healthResp.Metrics.MemoryUsage.AllocBytes,
		"goroutines":   healthResp.Metrics.Goroutines,
	}

	success := healthResp.Status == "healthy" && healthResp.Metrics.EntityCount > 0
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 2: Prometheus Metrics Endpoint
func (s *MetricsTestSuite) testPrometheusMetrics() {
	start := time.Now()
	testName := "Prometheus Metrics"

	resp, err := s.makeRequest("GET", "/metrics", "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("expected 200, got %d", resp.StatusCode), nil)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	content := string(body)
	details := map[string]interface{}{
		"content_length": len(content),
		"has_help_text":  strings.Contains(content, "# HELP"),
		"has_type_text":  strings.Contains(content, "# TYPE"),
		"has_entities":   strings.Contains(content, "entitydb_entities_total"),
		"has_uptime":     strings.Contains(content, "entitydb_uptime_seconds"),
	}

	success := details["has_help_text"].(bool) && details["has_entities"].(bool)
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 3: System Metrics Endpoint
func (s *MetricsTestSuite) testSystemMetrics() {
	start := time.Now()
	testName := "System Metrics"

	resp, err := s.makeRequest("GET", "/api/v1/system/metrics", "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("expected 200, got %d", resp.StatusCode), nil)
		return
	}

	var sysResp SystemMetricsResponse
	if err := json.NewDecoder(resp.Body).Decode(&sysResp); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"version":        sysResp.System.Version,
		"go_version":     sysResp.System.GoVersion,
		"total_entities": sysResp.Database.TotalEntities,
		"cpu_count":      sysResp.System.NumCPU,
		"memory_alloc":   sysResp.Memory.AllocBytes,
		"database_size":  sysResp.Storage.DatabaseSizeBytes,
	}

	success := sysResp.Database.TotalEntities > 0 && sysResp.System.NumCPU > 0
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 4: Available Metrics
func (s *MetricsTestSuite) testAvailableMetrics() {
	start := time.Now()
	testName := "Available Metrics"

	resp, err := s.makeRequest("GET", "/api/v1/metrics/available", "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("expected 200, got %d", resp.StatusCode), nil)
		return
	}

	var metrics []string
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"metric_count": len(metrics),
		"has_memory":   false,
		"has_entity":   false,
		"has_gc":       false,
	}

	for _, metric := range metrics {
		if strings.Contains(metric, "memory") {
			details["has_memory"] = true
		}
		if strings.Contains(metric, "entity") {
			details["has_entity"] = true
		}
		if strings.Contains(metric, "gc") {
			details["has_gc"] = true
		}
	}

	success := len(metrics) > 10 && details["has_memory"].(bool) && details["has_entity"].(bool)
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 5: Metric History
func (s *MetricsTestSuite) testMetricHistory() {
	start := time.Now()
	testName := "Metric History"

	resp, err := s.makeRequest("GET", "/api/v1/metrics/history?metric_name=memory_alloc&hours=1&limit=10", "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("expected 200, got %d", resp.StatusCode), nil)
		return
	}

	var histResp MetricHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&histResp); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"metric_name":   histResp.MetricName,
		"unit":          histResp.Unit,
		"data_points":   histResp.Count,
		"has_data":      len(histResp.DataPoints) > 0,
	}

	if len(histResp.DataPoints) > 0 {
		details["latest_value"] = histResp.DataPoints[0].Value
	}

	success := histResp.MetricName == "memory_alloc" && histResp.Unit == "bytes"
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 6: RBAC Metrics (requires auth)
func (s *MetricsTestSuite) testRBACMetrics() {
	start := time.Now()
	testName := "RBAC Metrics"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	resp, err := s.makeRequest("GET", "/api/v1/rbac/metrics", token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("expected 200, got %d", resp.StatusCode), nil)
		return
	}

	var rbacResp RBACMetricsResponse
	if err := json.NewDecoder(resp.Body).Decode(&rbacResp); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"total_users":  rbacResp.Users.TotalUsers,
		"admin_count":  rbacResp.Users.AdminCount,
		"active_sessions": rbacResp.Sessions.ActiveCount,
	}

	success := rbacResp.Users.TotalUsers >= 1 // At least admin user
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 7: Public RBAC Metrics (no auth)
func (s *MetricsTestSuite) testPublicRBACMetrics() {
	start := time.Now()
	testName := "Public RBAC Metrics"

	resp, err := s.makeRequest("GET", "/api/v1/rbac/metrics/public", "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("expected 200, got %d", resp.StatusCode), nil)
		return
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"has_timestamp": response["timestamp"] != nil,
		"has_auth":      response["auth"] != nil,
		"has_sessions":  response["sessions"] != nil,
	}

	success := details["has_timestamp"].(bool) && details["has_auth"].(bool)
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 8: Application Metrics (requires auth)
func (s *MetricsTestSuite) testApplicationMetrics() {
	start := time.Now()
	testName := "Application Metrics"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	resp, err := s.makeRequest("GET", "/api/v1/application/metrics", token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("expected 200, got %d", resp.StatusCode), nil)
		return
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"has_metrics": response["metrics"] != nil,
		"has_summary": response["summary"] != nil,
	}

	if metrics, ok := response["metrics"].([]interface{}); ok {
		details["metric_count"] = len(metrics)
	}

	success := details["has_metrics"].(bool) && details["has_summary"].(bool)
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 9: Comprehensive Metrics
func (s *MetricsTestSuite) testComprehensiveMetrics() {
	start := time.Now()
	testName := "Comprehensive Metrics"

	resp, err := s.makeRequest("GET", "/api/v1/metrics/comprehensive", "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("expected 200, got %d", resp.StatusCode), nil)
		return
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	requiredSections := []string{"system", "storage", "operations", "cache", "temporal", "rbac"}
	details := map[string]interface{}{
		"sections_found": 0,
		"has_timestamp": response["timestamp"] != nil,
	}

	for _, section := range requiredSections {
		if response[section] != nil {
			details["sections_found"] = details["sections_found"].(int) + 1
		}
	}

	success := details["sections_found"].(int) >= len(requiredSections)-1 // Allow 1 missing section
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 10: Error Handling
func (s *MetricsTestSuite) testErrorHandling() {
	start := time.Now()
	testName := "Error Handling"

	// Test invalid metric name
	resp, err := s.makeRequest("GET", "/api/v1/metrics/history?metric_name=invalid_metric_name", "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	var errorResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&errorResp)

	// Test unauthorized access
	resp2, _ := s.makeRequest("GET", "/api/v1/rbac/metrics", "invalid_token")
	defer resp2.Body.Close()

	details := map[string]interface{}{
		"invalid_metric_status":    resp.StatusCode,
		"invalid_metric_has_error": errorResp["error"] != nil,
		"unauthorized_status":      resp2.StatusCode,
	}

	success := (resp.StatusCode == 404 || resp.StatusCode == 400) && resp2.StatusCode == 401
	s.addResult(testName, success, time.Since(start), nil, details)
}

func (s *MetricsTestSuite) runAllTests() {
	fmt.Println("🚀 Starting EntityDB Metrics Endpoints End-to-End Test Suite")
	fmt.Println(strings.Repeat("=", 80))

	tests := []func(){
		s.testHealthEndpoint,
		s.testPrometheusMetrics,
		s.testSystemMetrics,
		s.testAvailableMetrics,
		s.testMetricHistory,
		s.testRBACMetrics,
		s.testPublicRBACMetrics,
		s.testApplicationMetrics,
		s.testComprehensiveMetrics,
		s.testErrorHandling,
	}

	for _, test := range tests {
		test()
		time.Sleep(100 * time.Millisecond) // Small delay between tests
	}
}

func (s *MetricsTestSuite) printResults() {
	fmt.Println("\n📊 METRICS ENDPOINTS TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	totalTests := len(s.results)
	passedTests := 0
	totalDuration := time.Duration(0)

	for i, result := range s.results {
		status := "❌ FAIL"
		if result.Success {
			status = "✅ PASS"
			passedTests++
		}

		fmt.Printf("[%d/%d] %s - %s (%.2fms)\n", 
			i+1, totalTests, status, result.TestName, float64(result.Duration.Nanoseconds())/1000000)

		if !result.Success {
			fmt.Printf("      Error: %s\n", result.Error)
		}

		if result.Details != nil && len(result.Details) > 0 {
			fmt.Printf("      Details: ")
			for k, v := range result.Details {
				fmt.Printf("%s=%v ", k, v)
			}
			fmt.Println()
		}

		totalDuration += result.Duration
		fmt.Println()
	}

	successRate := float64(passedTests) / float64(totalTests) * 100
	
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("🎯 SUMMARY: %d/%d tests passed (%.1f%% success rate)\n", passedTests, totalTests, successRate)
	fmt.Printf("⏱️  Total execution time: %.2fms\n", float64(totalDuration.Nanoseconds())/1000000)
	fmt.Printf("📈 Average test time: %.2fms\n", float64(totalDuration.Nanoseconds()/int64(totalTests))/1000000)

	if successRate < 100 {
		fmt.Printf("🚨 ISSUES DETECTED - Success rate %.1f%% below 100%% target\n", successRate)
		fmt.Println("   Failures require investigation")
	} else {
		fmt.Println("🎉 ALL TESTS PASSED - 100% Success Rate Achieved!")
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("EntityDB Metrics Endpoints End-to-End Test Suite")
		fmt.Println("Usage: go run test_metrics_endpoints_e2e.go")
		fmt.Println("\nThis comprehensive test suite validates:")
		fmt.Println("- Health endpoint functionality")
		fmt.Println("- Prometheus metrics format")
		fmt.Println("- System metrics endpoint")
		fmt.Println("- Available metrics listing")
		fmt.Println("- Metric history functionality")
		fmt.Println("- RBAC metrics (authenticated)")
		fmt.Println("- Public RBAC metrics")
		fmt.Println("- Application metrics (authenticated)")
		fmt.Println("- Comprehensive metrics endpoint")
		fmt.Println("- Error handling and authentication")
		return
	}

	suite := NewMetricsTestSuite()
	suite.runAllTests()
	suite.printResults()
}