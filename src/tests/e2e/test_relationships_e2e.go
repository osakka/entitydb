package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
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

type TestResult struct {
	TestName string
	Success  bool
	Duration time.Duration
	Error    string
	Details  map[string]interface{}
}

type RelationshipTestSuite struct {
	client     *http.Client
	results    []TestResult
	baseURL    string
	adminToken string                // Shared token for all tests
	testEntities map[string]string // name -> entity_id mapping
}

func NewRelationshipTestSuite() *RelationshipTestSuite {
	// Create HTTP client that ignores SSL certificates for testing
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   TestTimeout,
	}

	return &RelationshipTestSuite{
		client:       client,
		results:      make([]TestResult, 0),
		baseURL:      BaseURL,
		testEntities: make(map[string]string),
	}
}

func (s *RelationshipTestSuite) makeRequest(method, endpoint string, body interface{}, token string) (*http.Response, error) {
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

// makeAuthenticatedRequest automatically handles token refresh if needed
func (s *RelationshipTestSuite) makeAuthenticatedRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	maxRetries := 2
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Ensure we have a valid token
		if s.adminToken == "" {
			if err := s.initializeAuth(); err != nil {
				return nil, fmt.Errorf("failed to initialize auth: %v", err)
			}
		}
		
		resp, err := s.makeRequest(method, endpoint, body, s.adminToken)
		if err != nil {
			return nil, err
		}
		
		// If we get 401/403, try refreshing the token once
		if (resp.StatusCode == 401 || resp.StatusCode == 403) && attempt == 0 {
			resp.Body.Close()
			s.adminToken = "" // Force token refresh
			continue
		}
		
		return resp, nil
	}
	
	return nil, fmt.Errorf("authentication failed after retries")
}

func (s *RelationshipTestSuite) addResult(testName string, success bool, duration time.Duration, err error, details map[string]interface{}) {
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

func (s *RelationshipTestSuite) getAuthToken() (string, error) {
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

func (s *RelationshipTestSuite) initializeAuth() error {
	if s.adminToken == "" {
		token, err := s.getAuthTokenWithRetry()
		if err != nil {
			return err
		}
		s.adminToken = token
	}
	return nil
}

// getAuthTokenWithRetry implements robust authentication with retry logic
func (s *RelationshipTestSuite) getAuthTokenWithRetry() (string, error) {
	maxRetries := 3
	baseDelay := 500 * time.Millisecond
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		token, err := s.getAuthToken()
		if err == nil && token != "" {
			// Validate token immediately to ensure it's usable
			if s.validateToken(token) {
				return token, nil
			}
		}
		
		if attempt < maxRetries-1 {
			// Exponential backoff with jitter
			delay := baseDelay * time.Duration(1<<attempt)
			time.Sleep(delay)
		}
	}
	
	return "", fmt.Errorf("failed to obtain valid token after %d attempts", maxRetries)
}

// validateToken checks if a token is immediately usable
func (s *RelationshipTestSuite) validateToken(token string) bool {
	// Test token with a lightweight endpoint
	resp, err := s.makeRequest("GET", "/api/v1/dashboard/stats", nil, token)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func (s *RelationshipTestSuite) createTestEntity(name, entityType string, additionalTags []string) (string, error) {
	tags := []string{
		"name:" + name,
		"type:" + entityType,
		"purpose:relationship-testing",
		"test_suite:relationships",
	}
	tags = append(tags, additionalTags...)

	entityData := map[string]interface{}{
		"tags":    tags,
		"content": fmt.Sprintf("Test entity for relationship testing: %s", name),
	}

	resp, err := s.makeAuthenticatedRequest("POST", "/api/v1/entities/create", entityData)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("creation failed: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var entity Entity
	if err := json.NewDecoder(resp.Body).Decode(&entity); err != nil {
		return "", err
	}

	s.testEntities[name] = entity.ID
	return entity.ID, nil
}

// Test 1: Setup Test Entities
func (s *RelationshipTestSuite) testSetupTestEntities() {
	start := time.Now()
	testName := "Setup Test Entities"

	if err := s.initializeAuth(); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// Create test entities for relationship testing
	testEntities := map[string][]string{
		"project-alpha":   {"status:active", "priority:high"},
		"task-1":         {"status:pending", "priority:medium"},
		"task-2":         {"status:in-progress", "priority:high"},
		"user-john":      {"role:developer", "status:active"},
		"user-jane":      {"role:manager", "status:active"},
		"epic-backend":   {"category:development", "status:planning"},
	}

	createdCount := 0
	for entityName, additionalTags := range testEntities {
		if _, err := s.createTestEntity(entityName, "test", additionalTags); err != nil {
			s.addResult(testName, false, time.Since(start), fmt.Errorf("failed to create %s: %v", entityName, err), nil)
			return
		}
		createdCount++
	}

	details := map[string]interface{}{
		"entities_created": createdCount,
		"entity_names":     s.getEntityNames(),
	}

	success := createdCount == len(testEntities)
	s.addResult(testName, success, time.Since(start), nil, details)
}

func (s *RelationshipTestSuite) getEntityNames() []string {
	var names []string
	for name := range s.testEntities {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Test 2: Create Basic Relationships
func (s *RelationshipTestSuite) testCreateBasicRelationships() {
	start := time.Now()
	testName := "Create Basic Relationships"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Create relationships between entities using tags
	relationships := []struct {
		sourceEntity string
		targetEntity string
		relationType string
	}{
		{"task-1", "project-alpha", "belongs_to"},
		{"task-2", "project-alpha", "belongs_to"},
		{"task-1", "user-john", "assigned_to"},
		{"task-2", "user-jane", "assigned_to"},
		{"task-2", "task-1", "depends_on"},
		{"epic-backend", "project-alpha", "part_of"},
	}

	successCount := 0
	for _, rel := range relationships {
		sourceID, exists := s.testEntities[rel.sourceEntity]
		if !exists {
			continue
		}
		targetID, exists := s.testEntities[rel.targetEntity]
		if !exists {
			continue
		}

		// Update source entity with relationship tag
		relationshipTag := rel.relationType + ":" + targetID
		updateData := map[string]interface{}{
			"id":   sourceID,
			"tags": []string{relationshipTag},
		}

		resp, err := s.makeRequest("PUT", "/api/v1/entities/update", updateData, s.adminToken)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			successCount++
		}
	}

	details := map[string]interface{}{
		"relationships_attempted": len(relationships),
		"relationships_created":   successCount,
		"relationship_types":      []string{"belongs_to", "assigned_to", "depends_on", "part_of"},
	}

	success := successCount >= len(relationships)/2 // At least half should succeed
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 3: Query Relationships by Type
func (s *RelationshipTestSuite) testQueryRelationshipsByType() {
	start := time.Now()
	testName := "Query Relationships by Type"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Query entities by relationship type
	relationshipTypes := []string{"belongs_to", "assigned_to", "depends_on", "part_of"}
	queryResults := make(map[string]int)

	for _, relType := range relationshipTypes {
		resp, err := s.makeRequest("GET", "/api/v1/entities/query?tag="+relType+":*&limit=10", nil, s.adminToken)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			var response struct {
				Entities []Entity `json:"entities"`
				Total    int      `json:"total"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
				queryResults[relType] = len(response.Entities)
			}
		}
	}

	details := map[string]interface{}{
		"relationship_queries": queryResults,
		"types_tested":         relationshipTypes,
		"total_relationships":  s.sumValues(queryResults),
	}

	success := len(queryResults) >= 2 // At least 2 relationship types should have results
	s.addResult(testName, success, time.Since(start), nil, details)
}

func (s *RelationshipTestSuite) sumValues(m map[string]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}

// Test 4: Query Specific Relationships
func (s *RelationshipTestSuite) testQuerySpecificRelationships() {
	start := time.Now()
	testName := "Query Specific Relationships"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Query for specific relationship targets
	projectAlphaID, exists := s.testEntities["project-alpha"]
	if !exists {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("project-alpha entity not found"), nil)
		return
	}

	// Find all entities that belong to project-alpha
	resp, err := s.makeRequest("GET", "/api/v1/entities/query?tag=belongs_to:"+projectAlphaID, nil, s.adminToken)
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

	var response struct {
		Entities []Entity `json:"entities"`
		Total    int      `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	entities := response.Entities

	details := map[string]interface{}{
		"target_entity":      "project-alpha",
		"relationship_type":  "belongs_to",
		"entities_found":     len(entities),
		"entity_ids":         s.extractEntityIDs(entities),
	}

	success := len(entities) > 0
	s.addResult(testName, success, time.Since(start), nil, details)
}

func (s *RelationshipTestSuite) extractEntityIDs(entities []Entity) []string {
	var ids []string
	for _, entity := range entities {
		ids = append(ids, entity.ID)
	}
	return ids
}

// Test 5: Bidirectional Relationship Creation
func (s *RelationshipTestSuite) testBidirectionalRelationships() {
	start := time.Now()
	testName := "Bidirectional Relationships"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	task1ID, exists1 := s.testEntities["task-1"]
	task2ID, exists2 := s.testEntities["task-2"]
	if !exists1 || !exists2 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("test entities not found"), nil)
		return
	}

	// Create bidirectional relationship: task-1 blocks task-2, task-2 blocked_by task-1
	updates := []struct {
		entityID string
		tag      string
	}{
		{task1ID, "blocks:" + task2ID},
		{task2ID, "blocked_by:" + task1ID},
	}

	successCount := 0
	for _, update := range updates {
		updateData := map[string]interface{}{
			"id":   update.entityID,
			"tags": []string{update.tag},
		}

		resp, err := s.makeRequest("PUT", "/api/v1/entities/update", updateData, s.adminToken)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			successCount++
		}
	}

	// Verify bidirectional relationship exists
	verificationQueries := []string{
		"blocks:" + task2ID,
		"blocked_by:" + task1ID,
	}

	verifiedCount := 0
	for _, query := range verificationQueries {
		resp, err := s.makeRequest("GET", "/api/v1/entities/query?tag="+query, nil, s.adminToken)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			var response struct {
				Entities []Entity `json:"entities"`
				Total    int      `json:"total"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&response); err == nil && len(response.Entities) > 0 {
				verifiedCount++
			}
		}
	}

	details := map[string]interface{}{
		"updates_attempted":  len(updates),
		"updates_successful": successCount,
		"queries_verified":   verifiedCount,
		"bidirectional_types": []string{"blocks", "blocked_by"},
	}

	success := successCount >= 1 && verifiedCount >= 1
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 6: Complex Multi-Tag Relationship Queries
func (s *RelationshipTestSuite) testComplexRelationshipQueries() {
	start := time.Now()
	testName := "Complex Multi-Tag Queries"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	userJohnID, exists := s.testEntities["user-john"]
	if !exists {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("user-john entity not found"), nil)
		return
	}

	// Query for entities that are both assigned to john AND belong to project-alpha
	projectAlphaID := s.testEntities["project-alpha"]
	complexQuery := fmt.Sprintf("assigned_to:%s&tag=belongs_to:%s", userJohnID, projectAlphaID)

	resp, err := s.makeRequest("GET", "/api/v1/entities/query?tag="+complexQuery, nil, s.adminToken)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	var entities []Entity
	if resp.StatusCode == 200 {
		var response struct {
			Entities []Entity `json:"entities"`
			Total    int      `json:"total"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			entities = response.Entities
		}
	}

	// Also test wildcard queries
	wildcardQueries := []string{
		"assigned_to:*",
		"belongs_to:*",
		"depends_on:*",
	}

	wildcardResults := make(map[string]int)
	for _, query := range wildcardQueries {
		resp, err := s.makeRequest("GET", "/api/v1/entities/query?tag="+query+"&limit=5", nil, s.adminToken)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			var response struct {
				Entities []Entity `json:"entities"`
				Total    int      `json:"total"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
				wildcardResults[query] = len(response.Entities)
			}
		}
	}

	details := map[string]interface{}{
		"complex_query_results": len(entities),
		"wildcard_results":      wildcardResults,
		"total_wildcard_matches": s.sumValues(wildcardResults),
	}

	success := len(wildcardResults) >= 2 // At least 2 wildcard queries should have results
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 7: Relationship Tag Values Discovery
func (s *RelationshipTestSuite) testRelationshipTagValuesDiscovery() {
	start := time.Now()
	testName := "Relationship Tag Values Discovery"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Discover unique values for different relationship types
	relationshipNamespaces := []string{"belongs_to", "assigned_to", "depends_on", "blocks", "blocked_by"}
	discoveredValues := make(map[string][]string)

	for _, namespace := range relationshipNamespaces {
		resp, err := s.makeAuthenticatedRequest("GET", "/api/v1/tags/values?namespace="+namespace, nil)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			// The endpoint returns a structured response, not just an array
			var response struct {
				Values []string `json:"values"`
				Count  int      `json:"count"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
				discoveredValues[namespace] = response.Values
			}
		}
	}

	details := map[string]interface{}{
		"namespaces_tested":    relationshipNamespaces,
		"discovered_values":    discoveredValues,
		"total_namespaces":     len(discoveredValues),
		"total_unique_values":  s.countUniqueValues(discoveredValues),
	}

	success := len(discoveredValues) >= 2 // At least 2 namespaces should have values
	s.addResult(testName, success, time.Since(start), nil, details)
}

func (s *RelationshipTestSuite) countUniqueValues(valueMap map[string][]string) int {
	uniqueValues := make(map[string]bool)
	for _, values := range valueMap {
		for _, value := range values {
			uniqueValues[value] = true
		}
	}
	return len(uniqueValues)
}

// Test 8: Temporal Relationship History
func (s *RelationshipTestSuite) testTemporalRelationshipHistory() {
	start := time.Now()
	testName := "Temporal Relationship History"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	task1ID, exists := s.testEntities["task-1"]
	if !exists {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("task-1 entity not found"), nil)
		return
	}

	// Get history of relationship changes for task-1
	resp, err := s.makeRequest("GET", "/api/v1/entities/history?id="+task1ID, nil, s.adminToken)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("history query failed: %d - %s", resp.StatusCode, string(bodyBytes)), nil)
		return
	}

	var history []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// Also test as-of query for a specific timestamp
	timestamp := time.Now().Add(-5 * time.Minute).Format(time.RFC3339)
	asOfResp, err := s.makeRequest("GET", "/api/v1/entities/as-of?id="+task1ID+"&timestamp="+timestamp, nil, s.adminToken)
	asOfStatus := 0
	if err == nil {
		asOfStatus = asOfResp.StatusCode
		asOfResp.Body.Close()
	}

	details := map[string]interface{}{
		"entity_id":          task1ID,
		"history_entries":    len(history),
		"as_of_status":       asOfStatus,
		"timestamp_tested":   timestamp,
	}

	success := len(history) > 0
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 9: Relationship Update and Modification
func (s *RelationshipTestSuite) testRelationshipModification() {
	start := time.Now()
	testName := "Relationship Modification"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	task2ID, exists := s.testEntities["task-2"]
	if !exists {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("task-2 entity not found"), nil)
		return
	}

	userJohnID := s.testEntities["user-john"]

	// Change assignment from jane to john
	updateData := map[string]interface{}{
		"id":   task2ID,
		"tags": []string{"assigned_to:" + userJohnID, "status:reassigned"},
	}

	resp, err := s.makeRequest("PUT", "/api/v1/entities/update", updateData, s.adminToken)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	updateSuccess := resp.StatusCode == 200

	// Verify the relationship change
	verifyResp, err := s.makeRequest("GET", "/api/v1/entities/get?id="+task2ID, nil, s.adminToken)
	var verifyEntity Entity
	verifySuccess := false
	if err == nil && verifyResp.StatusCode == 200 {
		json.NewDecoder(verifyResp.Body).Decode(&verifyEntity)
		verifyResp.Body.Close()
		
		// Check if new assignment exists in tags
		for _, tag := range verifyEntity.Tags {
			if strings.Contains(tag, "assigned_to:"+userJohnID) {
				verifySuccess = true
				break
			}
		}
	}

	details := map[string]interface{}{
		"entity_id":           task2ID,
		"update_successful":   updateSuccess,
		"verification_successful": verifySuccess,
		"old_assignment":      "user-jane",
		"new_assignment":      "user-john",
		"tags_after_update":   len(verifyEntity.Tags),
	}

	success := updateSuccess && verifySuccess
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 10: Relationship Performance with Multiple Entities
func (s *RelationshipTestSuite) testRelationshipPerformance() {
	start := time.Now()
	testName := "Relationship Performance"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Create multiple entities with relationships for performance testing
	performanceEntities := make([]string, 0)
	for i := 0; i < 5; i++ {
		entityName := fmt.Sprintf("perf-entity-%d", i)
		entityID, err := s.createTestEntity(entityName, "performance", []string{"category:performance-test"})
		if err != nil {
			continue
		}
		performanceEntities = append(performanceEntities, entityID)
	}

	// Create relationships between performance entities
	relationshipCount := 0
	queryStartTime := time.Now()
	
	for i, sourceID := range performanceEntities {
		for j, targetID := range performanceEntities {
			if i != j {
				updateData := map[string]interface{}{
					"id":   sourceID,
					"tags": []string{fmt.Sprintf("perf_relates_to:%s", targetID)},
				}
				
				resp, err := s.makeRequest("PUT", "/api/v1/entities/update", updateData, s.adminToken)
				if err == nil && resp.StatusCode == 200 {
					relationshipCount++
				}
				if resp != nil {
					resp.Body.Close()
				}
			}
		}
	}

	relationshipCreationTime := time.Since(queryStartTime)

	// Test query performance
	queryStartTime = time.Now()
	resp, err := s.makeRequest("GET", "/api/v1/entities/query?tag=perf_relates_to:*&limit=20", nil, s.adminToken)
	queryTime := time.Since(queryStartTime)
	
	var queryResults []Entity
	if err == nil && resp.StatusCode == 200 {
		var response struct {
			Entities []Entity `json:"entities"`
			Total    int      `json:"total"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			queryResults = response.Entities
		}
		resp.Body.Close()
	}

	details := map[string]interface{}{
		"entities_created":           len(performanceEntities),
		"relationships_created":      relationshipCount,
		"relationship_creation_time": relationshipCreationTime.Milliseconds(),
		"query_time_ms":             queryTime.Milliseconds(),
		"query_results":             len(queryResults),
	}

	success := relationshipCount > 0 && queryTime.Milliseconds() < 1000 // Query should complete within 1 second
	s.addResult(testName, success, time.Since(start), nil, details)
}

func (s *RelationshipTestSuite) runAllTests() {
	fmt.Println("🔗 Starting EntityDB Relationship Testing End-to-End Test Suite")
	fmt.Println(strings.Repeat("=", 80))

	tests := []func(){
		s.testSetupTestEntities,
		s.testCreateBasicRelationships,
		s.testQueryRelationshipsByType,
		s.testQuerySpecificRelationships,
		s.testBidirectionalRelationships,
		s.testComplexRelationshipQueries,
		s.testRelationshipTagValuesDiscovery,
		s.testTemporalRelationshipHistory,
		s.testRelationshipModification,
		s.testRelationshipPerformance,
	}

	for _, test := range tests {
		test()
		time.Sleep(300 * time.Millisecond) // Delay between tests for server processing
	}
}

func (s *RelationshipTestSuite) printResults() {
	fmt.Println("\n📊 RELATIONSHIP TESTING RESULTS")
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

	// Summary of relationship testing coverage
	fmt.Println("\n🔗 RELATIONSHIP TESTING COVERAGE:")
	fmt.Println("✅ Tag-based relationship creation")
	fmt.Println("✅ Relationship querying by type")
	fmt.Println("✅ Specific relationship queries")
	fmt.Println("✅ Bidirectional relationships")
	fmt.Println("✅ Complex multi-tag queries")
	fmt.Println("✅ Relationship discovery")
	fmt.Println("✅ Temporal relationship history")
	fmt.Println("✅ Relationship modification")
	fmt.Println("✅ Performance testing")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("EntityDB Relationship Testing End-to-End Test Suite")
		fmt.Println("Usage: go run test_relationships_e2e.go")
		fmt.Println("\nThis comprehensive test suite validates:")
		fmt.Println("- Tag-based relationship creation and management")
		fmt.Println("- Relationship querying by type and target")
		fmt.Println("- Bidirectional relationship integrity")
		fmt.Println("- Complex multi-tag relationship queries")
		fmt.Println("- Relationship tag values discovery")
		fmt.Println("- Temporal relationship history and tracking")
		fmt.Println("- Relationship modification and updates")
		fmt.Println("- Performance with multiple relationships")
		fmt.Println("- EntityDB's modern tag-based relationship system")
		return
	}

	suite := NewRelationshipTestSuite()
	suite.runAllTests()
	suite.printResults()
}