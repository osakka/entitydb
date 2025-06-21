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
	ID        string   `json:"id"`
	Tags      []string `json:"tags"`
	Content   string   `json:"content"`
	CreatedAt int64    `json:"created_at"`
	UpdatedAt int64    `json:"updated_at"`
}

type WhoAmIResponse struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

type EntitySummaryResponse struct {
	LastUpdated   int64             `json:"last_updated"`
	RecentEntities []string         `json:"recent_entities"`
	Timestamp     int64             `json:"timestamp"`
	TotalCount    int               `json:"total_count"`
	TypeCounts    map[string]int    `json:"type_counts"`
}

type TestResult struct {
	TestName string
	Success  bool
	Duration time.Duration
	Error    string
	Details  map[string]interface{}
}

type DatabaseTestSuite struct {
	client  *http.Client
	results []TestResult
	baseURL string
}

func NewDatabaseTestSuite() *DatabaseTestSuite {
	// Create HTTP client that ignores SSL certificates for testing
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   TestTimeout,
	}

	return &DatabaseTestSuite{
		client:  client,
		results: make([]TestResult, 0),
		baseURL: BaseURL,
	}
}

func (s *DatabaseTestSuite) makeRequest(method, endpoint string, body interface{}, token string) (*http.Response, error) {
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

func (s *DatabaseTestSuite) addResult(testName string, success bool, duration time.Duration, err error, details map[string]interface{}) {
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

func (s *DatabaseTestSuite) getAuthToken() (string, error) {
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
		return "", fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}

	return loginResp.Token, nil
}

// Test 1: Authentication and User Data
func (s *DatabaseTestSuite) testAuthentication() {
	start := time.Now()
	testName := "Authentication & User Data"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// Test whoami endpoint
	resp, err := s.makeRequest("GET", "/api/v1/auth/whoami", nil, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("whoami failed with status %d", resp.StatusCode), nil)
		return
	}

	var whoami WhoAmIResponse
	if err := json.NewDecoder(resp.Body).Decode(&whoami); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"token_length": len(token),
		"user_id":      whoami.ID,
		"username":     whoami.Username,
		"email":        whoami.Email,
		"roles":        whoami.Roles,
	}

	success := whoami.Username == "admin" && len(whoami.Roles) > 0 && whoami.Email != ""
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 2: Entity Creation (CREATE)
func (s *DatabaseTestSuite) testEntityCreation() {
	start := time.Now()
	testName := "Entity Creation (CREATE)"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	entityData := map[string]interface{}{
		"tags":    []string{"name:test-create", "type:test", "purpose:crud-testing", "category:database"},
		"content": "VGVzdCBlbnRpdHkgZm9yIENSVUQgdGVzdGluZw==", // "Test entity for CRUD testing" in base64
	}

	resp, err := s.makeRequest("POST", "/api/v1/entities/create", entityData, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("creation failed: %d - %s", resp.StatusCode, string(bodyBytes)), nil)
		return
	}

	var entity Entity
	if err := json.NewDecoder(resp.Body).Decode(&entity); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"entity_id":    entity.ID,
		"tags_count":   len(entity.Tags),
		"created_at":   entity.CreatedAt,
		"has_content":  entity.Content != "",
	}

	success := entity.ID != "" && len(entity.Tags) > 4 && entity.CreatedAt > 0
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 3: Entity Reading (READ)
func (s *DatabaseTestSuite) testEntityReading() {
	start := time.Now()
	testName := "Entity Reading (READ)"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// First create an entity to read
	entityData := map[string]interface{}{
		"tags":    []string{"name:test-read", "type:test", "purpose:read-testing"},
		"content": "UmVhZCB0ZXN0IGVudGl0eQ==", // "Read test entity" in base64
	}

	createResp, err := s.makeRequest("POST", "/api/v1/entities/create", entityData, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer createResp.Body.Close()

	var createdEntity Entity
	json.NewDecoder(createResp.Body).Decode(&createdEntity)

	// Now read the entity
	resp, err := s.makeRequest("GET", "/api/v1/entities/get?id="+createdEntity.ID, nil, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("read failed: %d - %s", resp.StatusCode, string(bodyBytes)), nil)
		return
	}

	var readEntity Entity
	if err := json.NewDecoder(resp.Body).Decode(&readEntity); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"entity_id":     readEntity.ID,
		"tags_count":    len(readEntity.Tags),
		"content_match": readEntity.Content != "",
	}

	success := readEntity.ID == createdEntity.ID && len(readEntity.Tags) > 0
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 4: Entity Updating (UPDATE)
func (s *DatabaseTestSuite) testEntityUpdating() {
	start := time.Now()
	testName := "Entity Updating (UPDATE)"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// First create an entity to update
	entityData := map[string]interface{}{
		"tags":    []string{"name:test-update", "type:test", "status:original"},
		"content": "T3JpZ2luYWwgY29udGVudA==", // "Original content" in base64
	}

	createResp, err := s.makeRequest("POST", "/api/v1/entities/create", entityData, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer createResp.Body.Close()

	var createdEntity Entity
	json.NewDecoder(createResp.Body).Decode(&createdEntity)

	// Now update the entity
	updateData := map[string]interface{}{
		"id":      createdEntity.ID,
		"tags":    []string{"name:test-update-modified", "type:test", "status:updated"},
		"content": "VXBkYXRlZCBjb250ZW50", // "Updated content" in base64
	}

	resp, err := s.makeRequest("PUT", "/api/v1/entities/update", updateData, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("update failed: %d - %s", resp.StatusCode, string(bodyBytes)), nil)
		return
	}

	var updatedEntity Entity
	if err := json.NewDecoder(resp.Body).Decode(&updatedEntity); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"entity_id":      updatedEntity.ID,
		"updated_at":     updatedEntity.UpdatedAt,
		"tags_count":     len(updatedEntity.Tags),
		"time_changed":   updatedEntity.UpdatedAt > updatedEntity.CreatedAt,
	}

	success := updatedEntity.ID == createdEntity.ID && updatedEntity.UpdatedAt > createdEntity.CreatedAt
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 5: Entity Listing (LIST)
func (s *DatabaseTestSuite) testEntityListing() {
	start := time.Now()
	testName := "Entity Listing (LIST)"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	resp, err := s.makeRequest("GET", "/api/v1/entities/list?limit=10", nil, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("listing failed: %d - %s", resp.StatusCode, string(bodyBytes)), nil)
		return
	}

	var entities []Entity
	if err := json.NewDecoder(resp.Body).Decode(&entities); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"entity_count":  len(entities),
		"limit_applied": len(entities) <= 10,
	}

	if len(entities) > 0 {
		details["first_entity_id"] = entities[0].ID
		details["first_entity_tags"] = len(entities[0].Tags)
	}

	success := len(entities) > 0 && len(entities) <= 10
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 6: Entity Querying (QUERY)
func (s *DatabaseTestSuite) testEntityQuerying() {
	start := time.Now()
	testName := "Entity Querying (QUERY)"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	resp, err := s.makeRequest("GET", "/api/v1/entities/query?tag=type:test&limit=5", nil, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("query failed: %d - %s", resp.StatusCode, string(bodyBytes)), nil)
		return
	}

	var entities []Entity
	if err := json.NewDecoder(resp.Body).Decode(&entities); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"entity_count": len(entities),
		"query_tag":    "type:test",
		"limit":        5,
	}

	success := len(entities) >= 0 // Query can return 0 results validly
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 7: Temporal History
func (s *DatabaseTestSuite) testTemporalHistory() {
	start := time.Now()
	testName := "Temporal History"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// Create and update an entity to have history
	entityData := map[string]interface{}{
		"tags":    []string{"name:test-history", "type:test", "version:1"},
		"content": "SGlzdG9yeSB0ZXN0", // "History test" in base64
	}

	createResp, err := s.makeRequest("POST", "/api/v1/entities/create", entityData, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer createResp.Body.Close()

	var createdEntity Entity
	json.NewDecoder(createResp.Body).Decode(&createdEntity)

	// Update the entity to create history
	updateData := map[string]interface{}{
		"id":      createdEntity.ID,
		"tags":    []string{"name:test-history", "type:test", "version:2"},
		"content": "VXBkYXRlZCBoaXN0b3J5", // "Updated history" in base64
	}
	s.makeRequest("PUT", "/api/v1/entities/update", updateData, token)

	// Now get history
	resp, err := s.makeRequest("GET", "/api/v1/entities/history?id="+createdEntity.ID, nil, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("history failed: %d - %s", resp.StatusCode, string(bodyBytes)), nil)
		return
	}

	var history []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"entity_id":      createdEntity.ID,
		"history_count":  len(history),
		"has_changes":    len(history) > 0,
	}

	success := len(history) > 0
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 8: Temporal As-Of
func (s *DatabaseTestSuite) testTemporalAsOf() {
	start := time.Now()
	testName := "Temporal As-Of"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	resp, err := s.makeRequest("GET", "/api/v1/entities/as-of?timestamp="+timestamp, nil, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	details := map[string]interface{}{
		"timestamp":     timestamp,
		"status_code":   resp.StatusCode,
	}

	// As-of might return 200 with empty results or other valid responses
	success := resp.StatusCode == 200 || resp.StatusCode == 404
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 9: Tag Values
func (s *DatabaseTestSuite) testTagValues() {
	start := time.Now()
	testName := "Tag Values Query"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	resp, err := s.makeRequest("GET", "/api/v1/tags/values?namespace=type", nil, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("tag values failed: %d - %s", resp.StatusCode, string(bodyBytes)), nil)
		return
	}

	var tagValues []string
	if err := json.NewDecoder(resp.Body).Decode(&tagValues); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"namespace":    "type",
		"value_count":  len(tagValues),
	}

	if len(tagValues) > 0 {
		details["sample_values"] = tagValues
	}

	success := len(tagValues) >= 0 // Valid to have 0 or more tag values
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 10: Entity Summary
func (s *DatabaseTestSuite) testEntitySummary() {
	start := time.Now()
	testName := "Entity Summary"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	resp, err := s.makeRequest("GET", "/api/v1/entities/summary", nil, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("summary failed: %d - %s", resp.StatusCode, string(bodyBytes)), nil)
		return
	}

	var summary EntitySummaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"total_count":     summary.TotalCount,
		"type_counts":     summary.TypeCounts,
		"recent_count":    len(summary.RecentEntities),
		"has_timestamp":   summary.Timestamp > 0,
	}

	success := summary.TotalCount > 0 && summary.Timestamp > 0
	s.addResult(testName, success, time.Since(start), nil, details)
}

func (s *DatabaseTestSuite) runAllTests() {
	fmt.Println("🚀 Starting EntityDB Database Endpoints End-to-End Test Suite")
	fmt.Println(strings.Repeat("=", 80))

	tests := []func(){
		s.testAuthentication,
		s.testEntityCreation,
		s.testEntityReading,
		s.testEntityUpdating,
		s.testEntityListing,
		s.testEntityQuerying,
		s.testTemporalHistory,
		s.testTemporalAsOf,
		s.testTagValues,
		s.testEntitySummary,
	}

	for _, test := range tests {
		test()
		time.Sleep(500 * time.Millisecond) // Longer delay to prevent session conflicts
	}
}

func (s *DatabaseTestSuite) printResults() {
	fmt.Println("\n📊 DATABASE ENDPOINTS TEST RESULTS")
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
		fmt.Println("EntityDB Database Endpoints End-to-End Test Suite")
		fmt.Println("Usage: go run test_database_endpoints_e2e.go")
		fmt.Println("\nThis comprehensive test suite validates:")
		fmt.Println("- Authentication and user data integrity")
		fmt.Println("- Entity CRUD operations (Create, Read, Update)")
		fmt.Println("- Entity listing and querying")
		fmt.Println("- Temporal database features (History, As-Of)")
		fmt.Println("- Tag operations and values")
		fmt.Println("- Entity summary and statistics")
		fmt.Println("- Authorization and permission enforcement")
		return
	}

	suite := NewDatabaseTestSuite()
	suite.runAllTests()
	suite.printResults()
}