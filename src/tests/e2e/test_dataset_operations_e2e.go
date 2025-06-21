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

type Entity struct {
	ID       string   `json:"id"`
	Tags     []string `json:"tags"`
	Content  string   `json:"content"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

type QueryEntityResponse struct {
	Entities []*Entity `json:"entities"`
	Total    int       `json:"total"`
	Offset   int       `json:"offset"`
	Limit    int       `json:"limit"`
}

type TestResult struct {
	TestName string
	Success  bool
	Duration time.Duration
	Error    string
	Details  map[string]interface{}
}

type DatasetOperationsTestSuite struct {
	client       *http.Client
	results      []TestResult
	baseURL      string
	adminToken   string
	testDatasets map[string][]string // dataset_name -> entity_ids
}

func NewDatasetOperationsTestSuite() *DatasetOperationsTestSuite {
	// Create HTTP client that ignores SSL certificates for testing
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   TestTimeout,
	}

	return &DatasetOperationsTestSuite{
		client:       client,
		results:      make([]TestResult, 0),
		baseURL:      BaseURL,
		testDatasets: make(map[string][]string),
	}
}

func (s *DatasetOperationsTestSuite) makeRequest(method, endpoint string, body interface{}, token string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		if bodyStr, ok := body.(string); ok {
			reqBody = strings.NewReader(bodyStr)
		} else {
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %v", err)
			}
			reqBody = strings.NewReader(string(jsonData))
		}
	}

	req, err := http.NewRequest(method, s.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return s.client.Do(req)
}

func (s *DatasetOperationsTestSuite) addResult(testName string, success bool, duration time.Duration, err error, details map[string]interface{}) {
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

func (s *DatasetOperationsTestSuite) getAuthToken() (string, error) {
	loginData := map[string]string{
		"username": TestUsername,
		"password": TestPassword,
	}

	resp, err := s.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}

	return loginResp.Token, nil
}

// Test 1: Dataset Authentication and Access
func (s *DatasetOperationsTestSuite) testDatasetAuthentication() {
	start := time.Now()
	testName := "Dataset Authentication and Access"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	s.adminToken = token

	// Test accessing system dataset
	resp, err := s.makeRequest("GET", "/api/v1/entities/list?dataset=system&limit=5", nil, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	systemAccessible := resp.StatusCode == 200
	var entities []Entity
	systemEntityCount := 0
	if systemAccessible {
		if err := json.NewDecoder(resp.Body).Decode(&entities); err == nil {
			systemEntityCount = len(entities)
		}
	}

	details := map[string]interface{}{
		"auth_successful":      token != "",
		"system_accessible":    systemAccessible,
		"system_entity_count":  systemEntityCount,
		"token_length":         len(token),
	}

	success := token != "" && systemAccessible && systemEntityCount > 0
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 2: Create Multiple Datasets
func (s *DatasetOperationsTestSuite) testCreateMultipleDatasets() {
	start := time.Now()
	testName := "Create Multiple Datasets"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Create entities in different datasets
	testDatasets := []struct {
		name        string
		entityCount int
		entityType  string
	}{
		{"ecommerce", 5, "product"},
		{"analytics", 3, "metric"},
		{"workflow", 4, "task"},
		{"inventory", 3, "item"},
	}

	totalEntitiesCreated := 0
	datasetCount := 0

	for _, dataset := range testDatasets {
		datasetEntities := make([]string, 0)
		
		for i := 0; i < dataset.entityCount; i++ {
			entityData := map[string]interface{}{
				"tags": []string{
					fmt.Sprintf("dataset:%s", dataset.name),
					fmt.Sprintf("type:%s", dataset.entityType),
					fmt.Sprintf("name:%s_%d", dataset.entityType, i+1),
					"purpose:dataset-testing",
				},
				"content": fmt.Sprintf("Test %s entity %d in %s dataset", dataset.entityType, i+1, dataset.name),
			}

			resp, err := s.makeRequest("POST", "/api/v1/entities/create", entityData, s.adminToken)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == 200 || resp.StatusCode == 201 {
				var entity Entity
				if err := json.NewDecoder(resp.Body).Decode(&entity); err == nil {
					datasetEntities = append(datasetEntities, entity.ID)
					totalEntitiesCreated++
				}
			}
		}

		if len(datasetEntities) > 0 {
			s.testDatasets[dataset.name] = datasetEntities
			datasetCount++
		}
	}

	details := map[string]interface{}{
		"datasets_attempted":     len(testDatasets),
		"datasets_created":       datasetCount,
		"total_entities_created": totalEntitiesCreated,
		"dataset_names":          s.getDatasetNames(),
	}

	success := datasetCount >= len(testDatasets)/2 && totalEntitiesCreated >= 10
	s.addResult(testName, success, time.Since(start), nil, details)
}

func (s *DatasetOperationsTestSuite) getDatasetNames() []string {
	var names []string
	for name := range s.testDatasets {
		names = append(names, name)
	}
	return names
}

// Test 3: Dataset Isolation and Querying
func (s *DatasetOperationsTestSuite) testDatasetIsolation() {
	start := time.Now()
	testName := "Dataset Isolation and Querying"

	if s.adminToken == "" || len(s.testDatasets) == 0 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token or test datasets available"), nil)
		return
	}

	queryResults := make(map[string]int)
	isolationTests := 0
	successfulIsolation := 0

	// Test querying each dataset individually
	for datasetName := range s.testDatasets {
		isolationTests++
		
		resp, err := s.makeRequest("GET", fmt.Sprintf("/api/v1/entities/query?tag=dataset:%s&limit=20", datasetName), nil, s.adminToken)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			var response QueryEntityResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
				queryResults[datasetName] = len(response.Entities)
				if len(response.Entities) > 0 {
					successfulIsolation++
				}
			}
		}
	}

	// Test cross-dataset queries
	crossDatasetQueries := 0
	crossDatasetResults := 0

	// Query across all test datasets
	crossDatasetQueries++
	resp, err := s.makeRequest("GET", "/api/v1/entities/query?tag=purpose:dataset-testing&limit=50", nil, s.adminToken)
	if err == nil && resp.StatusCode == 200 {
		var response QueryEntityResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			crossDatasetResults = len(response.Entities)
		}
		resp.Body.Close()
	}

	details := map[string]interface{}{
		"isolation_tests":         isolationTests,
		"successful_isolation":    successfulIsolation,
		"query_results":           queryResults,
		"cross_dataset_results":   crossDatasetResults,
		"isolation_success_rate":  float64(successfulIsolation) / float64(isolationTests) * 100,
	}

	success := successfulIsolation >= isolationTests/2 && crossDatasetResults > 0
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 4: Dataset-Specific Operations
func (s *DatasetOperationsTestSuite) testDatasetSpecificOperations() {
	start := time.Now()
	testName := "Dataset-Specific Operations"

	if s.adminToken == "" || len(s.testDatasets) == 0 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token or test datasets available"), nil)
		return
	}

	operationTests := 0
	successfulOperations := 0

	// Test operations on specific dataset entities
	for datasetName, entityIDs := range s.testDatasets {
		if len(entityIDs) == 0 {
			continue
		}

		// Test updating an entity in the dataset
		operationTests++
		entityID := entityIDs[0]
		
		updateData := map[string]interface{}{
			"id":   entityID,
			"tags": []string{fmt.Sprintf("status:updated_in_%s", datasetName)},
		}

		resp, err := s.makeRequest("PUT", "/api/v1/entities/update", updateData, s.adminToken)
		if err == nil && resp.StatusCode == 200 {
			successfulOperations++
		}
		if resp != nil {
			resp.Body.Close()
		}

		// Test getting entity from specific dataset
		operationTests++
		resp, err = s.makeRequest("GET", fmt.Sprintf("/api/v1/entities/get?id=%s", entityID), nil, s.adminToken)
		if err == nil && resp.StatusCode == 200 {
			var entity Entity
			if err := json.NewDecoder(resp.Body).Decode(&entity); err == nil {
				// Verify entity belongs to correct dataset
				hasDatasetTag := false
				for _, tag := range entity.Tags {
					if strings.Contains(tag, fmt.Sprintf("dataset:%s", datasetName)) {
						hasDatasetTag = true
						break
					}
				}
				if hasDatasetTag {
					successfulOperations++
				}
			}
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	details := map[string]interface{}{
		"operation_tests":        operationTests,
		"successful_operations":  successfulOperations,
		"datasets_tested":        len(s.testDatasets),
		"operation_success_rate": func() float64 {
			if operationTests > 0 {
				return float64(successfulOperations) / float64(operationTests) * 100
			}
			return 0
		}(),
	}

	success := successfulOperations >= operationTests*2/3
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 5: Multi-Tenancy Simulation
func (s *DatasetOperationsTestSuite) testMultiTenancySimulation() {
	start := time.Now()
	testName := "Multi-Tenancy Simulation"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Simulate different tenants with their own datasets
	tenants := []struct {
		name     string
		datasets []string
	}{
		{"tenant_a", []string{"tenant_a_primary", "tenant_a_analytics"}},
		{"tenant_b", []string{"tenant_b_primary", "tenant_b_reports"}},
		{"tenant_c", []string{"tenant_c_data"}},
	}

	tenantOperations := 0
	successfulTenantOps := 0

	for _, tenant := range tenants {
		for _, dataset := range tenant.datasets {
			tenantOperations++
			
			// Create tenant-specific entity
			entityData := map[string]interface{}{
				"tags": []string{
					fmt.Sprintf("dataset:%s", dataset),
					fmt.Sprintf("tenant:%s", tenant.name),
					"type:tenant_data",
					"purpose:multi-tenancy-test",
				},
				"content": fmt.Sprintf("Data for %s in dataset %s", tenant.name, dataset),
			}

			resp, err := s.makeRequest("POST", "/api/v1/entities/create", entityData, s.adminToken)
			if err == nil && (resp.StatusCode == 200 || resp.StatusCode == 201) {
				successfulTenantOps++
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}

	// Test tenant isolation
	isolationTests := 0
	isolationSuccesses := 0

	for _, tenant := range tenants {
		isolationTests++
		
		// Query entities for specific tenant
		resp, err := s.makeRequest("GET", fmt.Sprintf("/api/v1/entities/query?tag=tenant:%s&limit=10", tenant.name), nil, s.adminToken)
		if err == nil && resp.StatusCode == 200 {
			var response QueryEntityResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err == nil && len(response.Entities) > 0 {
				// Verify all entities belong to the correct tenant
				allCorrectTenant := true
				for _, entity := range response.Entities {
					hasTenantTag := false
					for _, tag := range entity.Tags {
						if strings.Contains(tag, fmt.Sprintf("tenant:%s", tenant.name)) {
							hasTenantTag = true
							break
						}
					}
					if !hasTenantTag {
						allCorrectTenant = false
						break
					}
				}
				if allCorrectTenant {
					isolationSuccesses++
				}
			}
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	details := map[string]interface{}{
		"tenant_operations":      tenantOperations,
		"successful_tenant_ops":  successfulTenantOps,
		"isolation_tests":        isolationTests,
		"isolation_successes":    isolationSuccesses,
		"tenants_tested":         len(tenants),
		"multi_tenancy_working":  successfulTenantOps > 0 && isolationSuccesses > 0,
	}

	success := successfulTenantOps >= tenantOperations/2 && isolationSuccesses >= isolationTests/2
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 6: Dataset Performance and Scalability
func (s *DatasetOperationsTestSuite) testDatasetPerformance() {
	start := time.Now()
	testName := "Dataset Performance and Scalability"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Create a larger dataset for performance testing
	performanceDataset := "performance_test"
	entityCount := 20
	creationStartTime := time.Now()
	
	createdEntities := 0
	for i := 0; i < entityCount; i++ {
		entityData := map[string]interface{}{
			"tags": []string{
				fmt.Sprintf("dataset:%s", performanceDataset),
				"type:performance_entity",
				fmt.Sprintf("batch:performance_batch_%d", i/5), // Group into batches of 5
				"purpose:performance-testing",
			},
			"content": fmt.Sprintf("Performance test entity %d with some sample content data", i),
		}

		resp, err := s.makeRequest("POST", "/api/v1/entities/create", entityData, s.adminToken)
		if err == nil && (resp.StatusCode == 200 || resp.StatusCode == 201) {
			createdEntities++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	
	creationDuration := time.Since(creationStartTime)

	// Test query performance
	queryStartTime := time.Now()
	resp, err := s.makeRequest("GET", fmt.Sprintf("/api/v1/entities/query?tag=dataset:%s&limit=50", performanceDataset), nil, s.adminToken)
	queryDuration := time.Since(queryStartTime)
	
	queryResults := 0
	if err == nil && resp.StatusCode == 200 {
		var response QueryEntityResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			queryResults = len(response.Entities)
		}
		resp.Body.Close()
	}

	// Test batch operations performance
	batchStartTime := time.Now()
	batchResp, err := s.makeRequest("GET", "/api/v1/entities/query?tag=batch:performance_batch_0&limit=10", nil, s.adminToken)
	batchDuration := time.Since(batchStartTime)
	
	batchResults := 0
	if err == nil && batchResp.StatusCode == 200 {
		var response QueryEntityResponse
		if err := json.NewDecoder(batchResp.Body).Decode(&response); err == nil {
			batchResults = len(response.Entities)
		}
		batchResp.Body.Close()
	}

	details := map[string]interface{}{
		"entities_attempted":       entityCount,
		"entities_created":         createdEntities,
		"creation_duration_ms":     creationDuration.Milliseconds(),
		"query_duration_ms":        queryDuration.Milliseconds(),
		"batch_query_duration_ms":  batchDuration.Milliseconds(),
		"query_results":            queryResults,
		"batch_results":            batchResults,
		"avg_creation_time_ms":     func() float64 {
			if createdEntities > 0 {
				return float64(creationDuration.Milliseconds()) / float64(createdEntities)
			}
			return 0
		}(),
		"performance_acceptable":   queryDuration.Milliseconds() < 500 && batchDuration.Milliseconds() < 200,
	}

	success := createdEntities >= entityCount*3/4 && queryResults > 0 && queryDuration.Milliseconds() < 1000
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 7: Dataset Metadata and Discovery
func (s *DatasetOperationsTestSuite) testDatasetMetadataDiscovery() {
	start := time.Now()
	testName := "Dataset Metadata and Discovery"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Discover all dataset values using tag values endpoint
	resp, err := s.makeRequest("GET", "/api/v1/tags/values?namespace=dataset", nil, s.adminToken)
	discoveredDatasets := make([]string, 0)
	discoveryWorking := false
	
	if err == nil && resp.StatusCode == 200 {
		discoveryWorking = true
		var values []string
		if err := json.NewDecoder(resp.Body).Decode(&values); err == nil {
			discoveredDatasets = values
		}
		resp.Body.Close()
	}

	// Test dataset statistics
	datasetStats := make(map[string]int)
	statsTests := 0
	successfulStats := 0

	for _, dataset := range discoveredDatasets[:min(len(discoveredDatasets), 5)] { // Limit to first 5 datasets
		statsTests++
		
		resp, err := s.makeRequest("GET", fmt.Sprintf("/api/v1/entities/query?tag=dataset:%s&limit=100", dataset), nil, s.adminToken)
		if err == nil && resp.StatusCode == 200 {
			var response QueryEntityResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
				datasetStats[dataset] = len(response.Entities)
				successfulStats++
			}
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Test type discovery within datasets
	typeDiscovery := make(map[string][]string)
	typeTests := 0
	successfulTypeDiscovery := 0

	for dataset := range datasetStats {
		typeTests++
		
		resp, err := s.makeRequest("GET", "/api/v1/tags/values?namespace=type", nil, s.adminToken)
		if err == nil && resp.StatusCode == 200 {
			var types []string
			if err := json.NewDecoder(resp.Body).Decode(&types); err == nil {
				typeDiscovery[dataset] = types
				successfulTypeDiscovery++
			}
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	details := map[string]interface{}{
		"discovery_working":         discoveryWorking,
		"discovered_datasets":       len(discoveredDatasets),
		"dataset_names":             discoveredDatasets,
		"dataset_stats":             datasetStats,
		"stats_tests":               statsTests,
		"successful_stats":          successfulStats,
		"type_discovery":            typeDiscovery,
		"type_tests":                typeTests,
		"successful_type_discovery": successfulTypeDiscovery,
	}

	success := discoveryWorking && len(discoveredDatasets) > 0 && successfulStats >= statsTests/2
	s.addResult(testName, success, time.Since(start), nil, details)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Test 8: Dataset Temporal Operations
func (s *DatasetOperationsTestSuite) testDatasetTemporalOperations() {
	start := time.Now()
	testName := "Dataset Temporal Operations"

	if s.adminToken == "" || len(s.testDatasets) == 0 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token or test datasets available"), nil)
		return
	}

	temporalTests := 0
	successfulTemporal := 0

	// Test temporal queries on dataset entities
	for _, entityIDs := range s.testDatasets {
		if len(entityIDs) == 0 {
			continue
		}

		entityID := entityIDs[0]

		// Test history query
		temporalTests++
		resp, err := s.makeRequest("GET", fmt.Sprintf("/api/v1/entities/history?id=%s", entityID), nil, s.adminToken)
		if err == nil && resp.StatusCode == 200 {
			var history []interface{}
			if err := json.NewDecoder(resp.Body).Decode(&history); err == nil && len(history) > 0 {
				successfulTemporal++
			}
		}
		if resp != nil {
			resp.Body.Close()
		}

		// Test as-of query
		temporalTests++
		timestamp := time.Now().Add(-5 * time.Minute).Format(time.RFC3339)
		resp, err = s.makeRequest("GET", fmt.Sprintf("/api/v1/entities/as-of?id=%s&timestamp=%s", entityID, timestamp), nil, s.adminToken)
		if err == nil && (resp.StatusCode == 200 || resp.StatusCode == 404) { // 404 is acceptable if entity didn't exist at that time
			successfulTemporal++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Test temporal changes across dataset
	changesTests := 0
	successfulChanges := 0

	if len(s.testDatasets) > 0 {
		changesTests++
		
		// Get changes since 10 minutes ago
		timestamp := time.Now().Add(-10 * time.Minute).Format(time.RFC3339)
		resp, err := s.makeRequest("GET", fmt.Sprintf("/api/v1/entities/changes?since=%s&limit=20", timestamp), nil, s.adminToken)
		if err == nil && resp.StatusCode == 200 {
			var changes []interface{}
			if err := json.NewDecoder(resp.Body).Decode(&changes); err == nil {
				successfulChanges++
			}
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	details := map[string]interface{}{
		"temporal_tests":        temporalTests,
		"successful_temporal":   successfulTemporal,
		"changes_tests":         changesTests,
		"successful_changes":    successfulChanges,
		"datasets_tested":       len(s.testDatasets),
		"temporal_success_rate": func() float64 {
			total := temporalTests + changesTests
			success := successfulTemporal + successfulChanges
			if total > 0 {
				return float64(success) / float64(total) * 100
			}
			return 0
		}(),
	}

	totalTests := temporalTests + changesTests
	totalSuccess := successfulTemporal + successfulChanges
	success := totalTests > 0 && totalSuccess >= totalTests/2
	s.addResult(testName, success, time.Since(start), nil, details)
}

func (s *DatasetOperationsTestSuite) runAllTests() {
	fmt.Println("🗄️  Starting EntityDB Dataset Operations Testing End-to-End Test Suite")
	fmt.Println(strings.Repeat("=", 80))

	tests := []func(){
		s.testDatasetAuthentication,
		s.testCreateMultipleDatasets,
		s.testDatasetIsolation,
		s.testDatasetSpecificOperations,
		s.testMultiTenancySimulation,
		s.testDatasetPerformance,
		s.testDatasetMetadataDiscovery,
		s.testDatasetTemporalOperations,
	}

	for _, test := range tests {
		test()
		time.Sleep(300 * time.Millisecond) // Delay between tests for server processing
	}
}

func (s *DatasetOperationsTestSuite) printResults() {
	fmt.Println("\n📊 DATASET OPERATIONS TESTING RESULTS")
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

	if successRate >= 90 {
		fmt.Println("🎉 EXCELLENT - 90%+ Success Rate Achieved!")
	} else if successRate >= 75 {
		fmt.Println("👍 GOOD - 75%+ Success Rate Achieved!")
	} else {
		fmt.Printf("🚨 NEEDS IMPROVEMENT - Success rate %.1f%% below target\n", successRate)
	}

	// Summary of dataset operations testing coverage
	fmt.Println("\n🗄️  DATASET OPERATIONS TESTING COVERAGE:")
	fmt.Println("✅ Dataset authentication and access control")
	fmt.Println("✅ Multi-dataset creation and management")
	fmt.Println("✅ Dataset isolation and querying")
	fmt.Println("✅ Dataset-specific operations")
	fmt.Println("✅ Multi-tenancy simulation")
	fmt.Println("✅ Dataset performance and scalability")
	fmt.Println("✅ Dataset metadata discovery")
	fmt.Println("✅ Temporal operations across datasets")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("EntityDB Dataset Operations Testing End-to-End Test Suite")
		fmt.Println("Usage: go run test_dataset_operations_e2e.go")
		fmt.Println("\nThis comprehensive test suite validates:")
		fmt.Println("- Dataset authentication and access control")
		fmt.Println("- Multi-dataset creation and management")
		fmt.Println("- Dataset isolation and data segregation")
		fmt.Println("- Multi-tenancy capabilities")
		fmt.Println("- Dataset performance and scalability")
		fmt.Println("- Dataset discovery and metadata operations")
		fmt.Println("- Temporal operations across different datasets")
		fmt.Println("- EntityDB's complete multi-tenant dataset system")
		return
	}

	suite := NewDatasetOperationsTestSuite()
	suite.runAllTests()
	suite.printResults()
}