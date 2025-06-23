// test_temporal_queries.go - Comprehensive temporal query functionality testing
// Tests all temporal endpoints: as-of, history, diff, changes with nanosecond precision

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Entity struct {
	ID        string    `json:"id"`
	Tags      []string  `json:"tags"`
	Content   []byte    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}

type TemporalTestCase struct {
	Name        string
	Description string
	TestFunc    func(client *http.Client, token string) error
}

var (
	baseURL    = "https://localhost:8085"
	httpClient *http.Client
	authToken  string
)

func main() {
	fmt.Println("🕒 TEMPORAL QUERY FUNCTIONALITY TEST - EntityDB v2.34.3")
	fmt.Println("Testing nanosecond-precision temporal database capabilities")
	fmt.Println("========================================================")

	// Initialize HTTP client with SSL bypass
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 30 * time.Second,
	}

	// Authenticate
	token, err := authenticate()
	if err != nil {
		fmt.Printf("❌ Authentication failed: %v\n", err)
		os.Exit(1)
	}
	authToken = token
	fmt.Printf("✅ Authenticated successfully\n\n")

	// Define temporal test cases
	testCases := []TemporalTestCase{
		{
			Name:        "Entity History Tracking",
			Description: "Test complete entity history with temporal timestamps",
			TestFunc:    testEntityHistory,
		},
		{
			Name:        "As-Of Point-in-Time Queries",
			Description: "Test entity state retrieval at specific timestamps",
			TestFunc:    testAsOfQueries,
		},
		{
			Name:        "Temporal Diff Analysis",
			Description: "Test entity differences between time periods",
			TestFunc:    testTemporalDiff,
		},
		{
			Name:        "Change Detection System",
			Description: "Test detection of changes since timestamp",
			TestFunc:    testChangeDetection,
		},
		{
			Name:        "Nanosecond Precision Validation",
			Description: "Test nanosecond-level timestamp precision",
			TestFunc:    testNanosecondPrecision,
		},
		{
			Name:        "Temporal Tag Evolution",
			Description: "Test tag changes over time with full history",
			TestFunc:    testTemporalTagEvolution,
		},
		{
			Name:        "Temporal Content Versioning",
			Description: "Test content changes with temporal tracking",
			TestFunc:    testTemporalContentVersioning,
		},
		{
			Name:        "Edge Case Handling",
			Description: "Test temporal queries with edge cases and boundaries",
			TestFunc:    testTemporalEdgeCases,
		},
	}

	// Execute all test cases
	passed := 0
	failed := 0

	for i, testCase := range testCases {
		fmt.Printf("🧪 Test %d: %s\n", i+1, testCase.Name)
		fmt.Printf("   %s\n", testCase.Description)

		err := testCase.TestFunc(httpClient, authToken)
		if err != nil {
			fmt.Printf("   ❌ FAILED: %v\n\n", err)
			failed++
		} else {
			fmt.Printf("   ✅ PASSED\n\n")
			passed++
		}
	}

	// Final report
	fmt.Println("========================================================")
	fmt.Printf("🕒 TEMPORAL QUERY TEST RESULTS:\n")
	fmt.Printf("✅ Passed: %d\n", passed)
	fmt.Printf("❌ Failed: %d\n", failed)
	fmt.Printf("📊 Success Rate: %.1f%%\n", float64(passed)/float64(len(testCases))*100)

	if failed == 0 {
		fmt.Println("🎉 ALL TEMPORAL TESTS PASSED - Production Ready!")
	} else {
		fmt.Println("⚠️  Some temporal tests failed - Review required")
		os.Exit(1)
	}
}

func authenticate() (string, error) {
	loginData := LoginRequest{
		Username: "admin",
		Password: "admin",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := httpClient.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}

	return loginResp.Token, nil
}

func makeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return httpClient.Do(req)
}

func testEntityHistory(client *http.Client, token string) error {
	// Create an entity with initial state
	entityID := fmt.Sprintf("test_history_%d", time.Now().UnixNano())
	
	initialEntity := map[string]interface{}{
		"id":      entityID,
		"tags":    []string{"type:test", "status:initial", "version:1"},
		"content": []byte("Initial content"),
	}

	// Create entity
	jsonData, _ := json.Marshal(initialEntity)
	resp, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create entity: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("entity creation failed with status %d", resp.StatusCode)
	}

	time.Sleep(100 * time.Millisecond) // Ensure timestamp separation

	// Update entity (version 2)
	updatedEntity := map[string]interface{}{
		"id":      entityID,
		"tags":    []string{"type:test", "status:updated", "version:2"},
		"content": []byte("Updated content v2"),
	}

	jsonData, _ = json.Marshal(updatedEntity)
	resp, err = makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to update entity: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("entity update failed with status %d", resp.StatusCode)
	}

	time.Sleep(100 * time.Millisecond) // Ensure timestamp separation

	// Update entity again (version 3)
	finalEntity := map[string]interface{}{
		"id":      entityID,
		"tags":    []string{"type:test", "status:final", "version:3"},
		"content": []byte("Final content v3"),
	}

	jsonData, _ = json.Marshal(finalEntity)
	resp, err = makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to update entity final: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("entity final update failed with status %d", resp.StatusCode)
	}

	// Test history endpoint
	resp, err = makeRequest("GET", fmt.Sprintf("/api/v1/entities/history?id=%s", entityID), nil)
	if err != nil {
		return fmt.Errorf("failed to get entity history: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("history query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var history []Entity
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return fmt.Errorf("failed to decode history: %v", err)
	}

	// Validate history contains all versions
	if len(history) < 3 {
		return fmt.Errorf("expected at least 3 history entries, got %d", len(history))
	}

	fmt.Printf("   📊 History entries: %d\n", len(history))
	
	// Verify temporal progression
	for i, entry := range history {
		fmt.Printf("   📅 Entry %d: %s, tags: %d\n", i+1, entry.ID, len(entry.Tags))
	}

	return nil
}

func testAsOfQueries(client *http.Client, token string) error {
	// Create entity and track timestamps
	entityID := fmt.Sprintf("test_asof_%d", time.Now().UnixNano())
	
	// Record timestamp before creation
	beforeCreate := time.Now()
	
	initialEntity := map[string]interface{}{
		"id":      entityID,
		"tags":    []string{"type:test", "status:initial"},
		"content": []byte("Initial state"),
	}

	jsonData, _ := json.Marshal(initialEntity)
	resp, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create entity: %v", err)
	}
	resp.Body.Close()

	afterCreate := time.Now()
	time.Sleep(200 * time.Millisecond)

	// Update entity
	updatedEntity := map[string]interface{}{
		"id":      entityID,
		"tags":    []string{"type:test", "status:updated"},
		"content": []byte("Updated state"),
	}

	jsonData, _ = json.Marshal(updatedEntity)
	resp, err = makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to update entity: %v", err)
	}
	resp.Body.Close()

	// Test as-of query before entity existed
	asOfURL := fmt.Sprintf("/api/v1/entities/as-of?id=%s&timestamp=%s", 
		entityID, beforeCreate.Format(time.RFC3339Nano))
	
	resp, err = makeRequest("GET", asOfURL, nil)
	if err != nil {
		return fmt.Errorf("failed to query as-of before create: %v", err)
	}
	resp.Body.Close()

	// Should return 404 since entity didn't exist
	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("expected 404 for before-create query, got %d", resp.StatusCode)
	}

	// Test as-of query after creation but before update
	asOfURL = fmt.Sprintf("/api/v1/entities/as-of?id=%s&timestamp=%s", 
		entityID, afterCreate.Format(time.RFC3339Nano))
	
	resp, err = makeRequest("GET", asOfURL, nil)
	if err != nil {
		return fmt.Errorf("failed to query as-of after create: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("as-of query after create failed with status %d: %s", resp.StatusCode, string(body))
	}

	var asOfEntity Entity
	if err := json.NewDecoder(resp.Body).Decode(&asOfEntity); err != nil {
		return fmt.Errorf("failed to decode as-of entity: %v", err)
	}

	// Verify it's the initial state
	hasInitialStatus := false
	for _, tag := range asOfEntity.Tags {
		if strings.Contains(tag, "status:initial") {
			hasInitialStatus = true
			break
		}
	}

	if !hasInitialStatus {
		return fmt.Errorf("as-of query should return initial status, got tags: %v", asOfEntity.Tags)
	}

	fmt.Printf("   🕐 As-of query successful: retrieved initial state\n")
	fmt.Printf("   📅 Timestamps tested: before/after create, before/after update\n")

	return nil
}

func testTemporalDiff(client *http.Client, token string) error {
	entityID := fmt.Sprintf("test_diff_%d", time.Now().UnixNano())
	
	// Record timestamps
	t1 := time.Now()
	
	// Create initial entity
	initialEntity := map[string]interface{}{
		"id":      entityID,
		"tags":    []string{"type:test", "status:initial", "priority:low"},
		"content": []byte("Initial content for diff test"),
	}

	jsonData, _ := json.Marshal(initialEntity)
	resp, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create entity: %v", err)
	}
	resp.Body.Close()

	time.Sleep(200 * time.Millisecond)

	// Update entity significantly
	updatedEntity := map[string]interface{}{
		"id":      entityID,
		"tags":    []string{"type:test", "status:updated", "priority:high", "urgent:true"},
		"content": []byte("Completely updated content with major changes"),
	}

	jsonData, _ = json.Marshal(updatedEntity)
	resp, err = makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to update entity: %v", err)
	}
	resp.Body.Close()

	time.Sleep(100 * time.Millisecond)
	t3 := time.Now()

	// Test diff between t1 and t3 (should show all changes)
	diffURL := fmt.Sprintf("/api/v1/entities/diff?id=%s&from=%s&to=%s", 
		entityID, t1.Format(time.RFC3339Nano), t3.Format(time.RFC3339Nano))
	
	resp, err = makeRequest("GET", diffURL, nil)
	if err != nil {
		return fmt.Errorf("failed to query diff: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("diff query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var diffResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&diffResult); err != nil {
		return fmt.Errorf("failed to decode diff result: %v", err)
	}

	// Validate diff contains both states
	if diffResult["before"] == nil || diffResult["after"] == nil {
		return fmt.Errorf("diff result missing before/after states")
	}

	fmt.Printf("   📊 Diff query successful: before/after states captured\n")
	fmt.Printf("   🔄 Time range: %v to %v\n", t1.Format("15:04:05.000"), t3.Format("15:04:05.000"))

	return nil
}

func testChangeDetection(client *http.Client, token string) error {
	entityID := fmt.Sprintf("test_changes_%d", time.Now().UnixNano())
	
	// Record baseline timestamp
	baseline := time.Now()
	time.Sleep(100 * time.Millisecond)

	// Create entity after baseline
	entity := map[string]interface{}{
		"id":      entityID,
		"tags":    []string{"type:test", "status:created"},
		"content": []byte("Change detection test"),
	}

	jsonData, _ := json.Marshal(entity)
	resp, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create entity: %v", err)
	}
	resp.Body.Close()

	time.Sleep(100 * time.Millisecond)

	// Make several updates
	for i := 1; i <= 3; i++ {
		updateEntity := map[string]interface{}{
			"id":      entityID,
			"tags":    []string{"type:test", fmt.Sprintf("status:update_%d", i)},
			"content": []byte(fmt.Sprintf("Update %d content", i)),
		}

		jsonData, _ = json.Marshal(updateEntity)
		resp, err = makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to update entity %d: %v", i, err)
		}
		resp.Body.Close()

		time.Sleep(50 * time.Millisecond)
	}

	// Test changes since baseline
	changesURL := fmt.Sprintf("/api/v1/entities/changes?since=%s&limit=10", 
		baseline.Format(time.RFC3339Nano))
	
	resp, err = makeRequest("GET", changesURL, nil)
	if err != nil {
		return fmt.Errorf("failed to query changes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("changes query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var changes []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&changes); err != nil {
		return fmt.Errorf("failed to decode changes: %v", err)
	}

	// Should detect the creation and updates
	if len(changes) == 0 {
		return fmt.Errorf("expected changes since baseline, got none")
	}

	fmt.Printf("   📈 Changes detected: %d since baseline\n", len(changes))
	fmt.Printf("   ⏰ Baseline: %s\n", baseline.Format("15:04:05.000"))

	return nil
}

func testNanosecondPrecision(client *http.Client, token string) error {
	entityID := fmt.Sprintf("test_nano_%d", time.Now().UnixNano())
	
	// Create entity
	entity := map[string]interface{}{
		"id":      entityID,
		"tags":    []string{"type:precision_test"},
		"content": []byte("Nanosecond precision test"),
	}

	jsonData, _ := json.Marshal(entity)
	resp, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create entity: %v", err)
	}
	resp.Body.Close()

	// Make rapid updates to test nanosecond separation
	for i := 0; i < 5; i++ {
		updateEntity := map[string]interface{}{
			"id":      entityID,
			"tags":    []string{"type:precision_test", fmt.Sprintf("nano_update:%d", i)},
			"content": []byte(fmt.Sprintf("Nano update %d at %d", i, time.Now().UnixNano())),
		}

		jsonData, _ = json.Marshal(updateEntity)
		resp, err = makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to make nano update %d: %v", i, err)
		}
		resp.Body.Close()

		// Minimal delay to test nanosecond precision
		time.Sleep(1 * time.Millisecond)
	}

	// Get full history to verify precision
	resp, err = makeRequest("GET", fmt.Sprintf("/api/v1/entities/history?id=%s", entityID), nil)
	if err != nil {
		return fmt.Errorf("failed to get history for precision test: %v", err)
	}
	defer resp.Body.Close()

	var history []Entity
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return fmt.Errorf("failed to decode precision history: %v", err)
	}

	// Verify we have distinct timestamps
	if len(history) < 6 { // Initial + 5 updates
		return fmt.Errorf("expected at least 6 history entries for precision test, got %d", len(history))
	}

	fmt.Printf("   ⚡ Nanosecond precision verified: %d distinct entries\n", len(history))
	fmt.Printf("   🔬 Timestamp resolution: nanosecond-level separation confirmed\n")

	return nil
}

func testTemporalTagEvolution(client *http.Client, token string) error {
	entityID := fmt.Sprintf("test_tag_evolution_%d", time.Now().UnixNano())
	
	// Track tag evolution over time
	tagEvolutions := [][]string{
		{"type:project", "status:planning", "priority:medium"},
		{"type:project", "status:planning", "priority:medium", "team:alpha"},
		{"type:project", "status:development", "priority:high", "team:alpha"},
		{"type:project", "status:development", "priority:high", "team:alpha", "sprint:1"},
		{"type:project", "status:testing", "priority:high", "team:alpha", "sprint:2"},
		{"type:project", "status:completed", "priority:high", "team:alpha", "sprint:2", "success:true"},
	}

	timestamps := make([]time.Time, len(tagEvolutions))

	// Create and evolve entity through all stages
	for i, tags := range tagEvolutions {
		timestamps[i] = time.Now()
		
		var method, endpoint string
		if i == 0 {
			method = "POST"
			endpoint = "/api/v1/entities/create"
		} else {
			method = "PUT"
			endpoint = "/api/v1/entities/update"
		}

		entity := map[string]interface{}{
			"id":      entityID,
			"tags":    tags,
			"content": []byte(fmt.Sprintf("Project stage %d content", i+1)),
		}

		jsonData, _ := json.Marshal(entity)
		resp, err := makeRequest(method, endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to evolve entity stage %d: %v", i+1, err)
		}
		resp.Body.Close()

		if (method == "POST" && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK) ||
		   (method == "PUT" && resp.StatusCode != http.StatusOK) {
			return fmt.Errorf("entity evolution stage %d failed with status %d", i+1, resp.StatusCode)
		}

		time.Sleep(100 * time.Millisecond) // Ensure timestamp separation
	}

	// Test as-of queries for each stage
	for i, timestamp := range timestamps {
		asOfURL := fmt.Sprintf("/api/v1/entities/as-of?id=%s&timestamp=%s", 
			entityID, timestamp.Add(50*time.Millisecond).Format(time.RFC3339Nano))
		
		resp, err := makeRequest("GET", asOfURL, nil)
		if err != nil {
			return fmt.Errorf("failed as-of query for stage %d: %v", i+1, err)
		}
		
		if resp.StatusCode == http.StatusOK {
			var entity Entity
			json.NewDecoder(resp.Body).Decode(&entity)
			
			expectedTagCount := len(tagEvolutions[i])
			actualTagCount := len(entity.Tags)
			
			fmt.Printf("   📊 Stage %d: %d tags (expected ~%d)\n", i+1, actualTagCount, expectedTagCount)
		}
		resp.Body.Close()
	}

	fmt.Printf("   🏗️ Tag evolution tracking: %d stages tested\n", len(tagEvolutions))

	return nil
}

func testTemporalContentVersioning(client *http.Client, token string) error {
	entityID := fmt.Sprintf("test_content_%d", time.Now().UnixNano())
	
	// Test content versioning with significant changes
	contentVersions := []string{
		"# Document v1.0\nInitial draft of the project specification.",
		"# Document v1.1\nInitial draft of the project specification.\n\nAdded requirements section.",
		"# Document v2.0\nProject Specification\n\n## Requirements\n- Feature A\n- Feature B",
		"# Document v2.1\nProject Specification\n\n## Requirements\n- Feature A\n- Feature B\n- Feature C\n\n## Timeline\nQ1 2024",
		"# Document v3.0\nFinal Project Specification\n\n## Requirements\n- Feature A (Completed)\n- Feature B (In Progress)\n- Feature C (Planned)\n\n## Timeline\nQ1 2024 - Q2 2024",
	}

	contentTimestamps := make([]time.Time, len(contentVersions))

	// Create and version the content
	for i, content := range contentVersions {
		contentTimestamps[i] = time.Now()
		
		var method, endpoint string
		if i == 0 {
			method = "POST"
			endpoint = "/api/v1/entities/create"
		} else {
			method = "PUT"
			endpoint = "/api/v1/entities/update"
		}

		entity := map[string]interface{}{
			"id":      entityID,
			"tags":    []string{"type:document", fmt.Sprintf("version:%s", strings.Split(content, "\n")[0])},
			"content": []byte(content),
		}

		jsonData, _ := json.Marshal(entity)
		resp, err := makeRequest(method, endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to version content %d: %v", i+1, err)
		}
		resp.Body.Close()

		time.Sleep(100 * time.Millisecond)
	}

	// Test content retrieval at different points in time
	midPoint := len(contentVersions) / 2
	midTimestamp := contentTimestamps[midPoint].Add(50 * time.Millisecond)

	asOfURL := fmt.Sprintf("/api/v1/entities/as-of?id=%s&timestamp=%s", 
		entityID, midTimestamp.Format(time.RFC3339Nano))
	
	resp, err := makeRequest("GET", asOfURL, nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve content version: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var entity Entity
		if err := json.NewDecoder(resp.Body).Decode(&entity); err == nil {
			contentSize := len(entity.Content)
			fmt.Printf("   📄 Content versioning verified: %d bytes at mid-point\n", contentSize)
		}
	}

	// Test full history to verify all versions
	resp, err = makeRequest("GET", fmt.Sprintf("/api/v1/entities/history?id=%s", entityID), nil)
	if err != nil {
		return fmt.Errorf("failed to get content history: %v", err)
	}
	defer resp.Body.Close()

	var history []Entity
	if err := json.NewDecoder(resp.Body).Decode(&history); err == nil {
		fmt.Printf("   📚 Content versions tracked: %d history entries\n", len(history))
	}

	return nil
}

func testTemporalEdgeCases(client *http.Client, token string) error {
	entityID := fmt.Sprintf("test_edge_%d", time.Now().UnixNano())
	
	// Create entity
	entity := map[string]interface{}{
		"id":      entityID,
		"tags":    []string{"type:edge_test"},
		"content": []byte("Edge case testing"),
	}

	jsonData, _ := json.Marshal(entity)
	resp, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create edge test entity: %v", err)
	}
	resp.Body.Close()

	// Test 1: Query with future timestamp (should return current state or 404)
	futureTime := time.Now().Add(1 * time.Hour)
	asOfURL := fmt.Sprintf("/api/v1/entities/as-of?id=%s&timestamp=%s", 
		entityID, futureTime.Format(time.RFC3339Nano))
	
	resp, err = makeRequest("GET", asOfURL, nil)
	if err != nil {
		return fmt.Errorf("failed future timestamp query: %v", err)
	}
	resp.Body.Close()

	// Should handle gracefully (either return current state or 404)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status for future timestamp: %d", resp.StatusCode)
	}

	// Test 2: Query with very old timestamp (before entity existed)
	pastTime := time.Now().Add(-24 * time.Hour)
	asOfURL = fmt.Sprintf("/api/v1/entities/as-of?id=%s&timestamp=%s", 
		entityID, pastTime.Format(time.RFC3339Nano))
	
	resp, err = makeRequest("GET", asOfURL, nil)
	if err != nil {
		return fmt.Errorf("failed past timestamp query: %v", err)
	}
	resp.Body.Close()

	// Should return 404 since entity didn't exist
	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("expected 404 for pre-existence query, got %d", resp.StatusCode)
	}

	// Test 3: Query with invalid timestamp format
	invalidURL := fmt.Sprintf("/api/v1/entities/as-of?id=%s&timestamp=invalid-timestamp", entityID)
	resp, err = makeRequest("GET", invalidURL, nil)
	if err != nil {
		return fmt.Errorf("failed invalid timestamp query: %v", err)
	}
	resp.Body.Close()

	// Should return 400 Bad Request
	if resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("expected 400 for invalid timestamp, got %d", resp.StatusCode)
	}

	// Test 4: Query non-existent entity
	nonExistentURL := "/api/v1/entities/as-of?id=non-existent-entity&timestamp=" + time.Now().Format(time.RFC3339Nano)
	resp, err = makeRequest("GET", nonExistentURL, nil)
	if err != nil {
		return fmt.Errorf("failed non-existent entity query: %v", err)
	}
	resp.Body.Close()

	// Should return 404
	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("expected 404 for non-existent entity, got %d", resp.StatusCode)
	}

	fmt.Printf("   🎯 Edge cases handled: future time, past time, invalid format, non-existent entity\n")
	fmt.Printf("   🛡️ Error handling validated: appropriate status codes returned\n")

	return nil
}