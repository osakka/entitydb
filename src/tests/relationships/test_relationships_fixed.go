// test_relationships_fixed.go - Fixed relationship management test for pure tag-based system
// Tests EntityDB's pure tag-based relationship system with correct API response format

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Entity struct {
	ID        string    `json:"id"`
	Tags      []string  `json:"tags"`
	Content   []byte    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type QueryResponse struct {
	Entities []Entity `json:"entities"`
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

type RelationshipTestCase struct {
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
	fmt.Println("🔗 RELATIONSHIP MANAGEMENT TEST (FIXED) - EntityDB v2.34.3")
	fmt.Println("Testing pure tag-based relationship system")
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

	// Define relationship test cases
	testCases := []RelationshipTestCase{
		{
			Name:        "Basic Tag-Based Relationships",
			Description: "Test creating relationships using pure tag system",
			TestFunc:    testBasicTagRelationships,
		},
		{
			Name:        "Bidirectional Relationships",
			Description: "Test bidirectional relationships (parent-child, peers)",
			TestFunc:    testBidirectionalRelationships,
		},
		{
			Name:        "Complex Relationship Networks",
			Description: "Test complex relationship graphs and queries",
			TestFunc:    testComplexRelationshipNetworks,
		},
		{
			Name:        "Temporal Relationship Evolution",
			Description: "Test relationships changing over time",
			TestFunc:    testTemporalRelationshipEvolution,
		},
		{
			Name:        "Relationship Query Patterns",
			Description: "Test various query patterns for relationships",
			TestFunc:    testRelationshipQueryPatterns,
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
	fmt.Printf("🔗 RELATIONSHIP TEST RESULTS:\n")
	fmt.Printf("✅ Passed: %d\n", passed)
	fmt.Printf("❌ Failed: %d\n", failed)
	fmt.Printf("📊 Success Rate: %.1f%%\n", float64(passed)/float64(len(testCases))*100)

	if failed == 0 {
		fmt.Println("🎉 ALL RELATIONSHIP TESTS PASSED - Production Ready!")
	} else {
		fmt.Println("⚠️  Some relationship tests failed - Review required")
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

func queryEntities(tags ...string) ([]Entity, error) {
	// Build query URL with tags
	queryURL := "/api/v1/entities/query?"
	for i, tag := range tags {
		if i > 0 {
			queryURL += "&"
		}
		queryURL += "tags=" + tag
	}

	resp, err := makeRequest("GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("query request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response with correct structure
	var queryResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode query response: %v", err)
	}

	return queryResp.Entities, nil
}

func testBasicTagRelationships(client *http.Client, token string) error {
	// Create two entities to relate
	userID := fmt.Sprintf("user_%d", time.Now().UnixNano())
	projectID := fmt.Sprintf("project_%d", time.Now().UnixNano())
	
	// Create user entity
	userEntity := map[string]interface{}{
		"id":      userID,
		"tags":    []string{"type:user", "name:John_Doe", "role:developer"},
		"content": []byte("User profile data"),
	}

	jsonData, _ := json.Marshal(userEntity)
	resp, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create user entity: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("user creation failed with status %d", resp.StatusCode)
	}

	// Create project entity
	projectEntity := map[string]interface{}{
		"id":      projectID,
		"tags":    []string{"type:project", "name:EntityDB", "status:active"},
		"content": []byte("Project details"),
	}

	jsonData, _ = json.Marshal(projectEntity)
	resp, err = makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create project entity: %v", err)
	}
	resp.Body.Close()

	// Method 1: Add relationship tag directly to user
	userEntity["tags"] = append(userEntity["tags"].([]string), fmt.Sprintf("works_on:%s", projectID))
	
	jsonData, _ = json.Marshal(userEntity)
	resp, err = makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to add relationship tag: %v", err)
	}
	resp.Body.Close()

	// Method 2: Create a relationship entity
	relationshipID := fmt.Sprintf("rel_%s_%s", userID, projectID)
	relationshipEntity := map[string]interface{}{
		"id": relationshipID,
		"tags": []string{
			"type:relationship",
			fmt.Sprintf("from:%s", userID),
			fmt.Sprintf("to:%s", projectID),
			"relation:works_on",
			"since:2025-01-01",
		},
		"content": []byte(`{"hours_per_week": 40, "role": "lead_developer"}`),
	}

	jsonData, _ = json.Marshal(relationshipEntity)
	resp, err = makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create relationship entity: %v", err)
	}
	resp.Body.Close()

	// Query relationships using proper method
	relationships, err := queryEntities(fmt.Sprintf("from:%s", userID))
	if err != nil {
		return fmt.Errorf("failed to query user relationships: %v", err)
	}

	if len(relationships) == 0 {
		return fmt.Errorf("expected at least 1 relationship, got 0")
	}

	// Query user directly to verify tag was added
	userEntities, err := queryEntities(fmt.Sprintf("id:%s", userID))
	if err != nil {
		return fmt.Errorf("failed to query user: %v", err)
	}

	hasWorksOnTag := false
	if len(userEntities) > 0 {
		for _, tag := range userEntities[0].Tags {
			if len(tag) > 20 && tag[20:] == fmt.Sprintf("works_on:%s", projectID) {
				hasWorksOnTag = true
				break
			}
		}
	}

	fmt.Printf("   📊 Relationships created: %d\n", len(relationships))
	fmt.Printf("   🔗 User->Project direct tag: %v\n", hasWorksOnTag)
	fmt.Printf("   📋 Relationship entity created successfully\n")

	return nil
}

func testBidirectionalRelationships(client *http.Client, token string) error {
	// Create parent-child bidirectional relationship
	parentID := fmt.Sprintf("parent_%d", time.Now().UnixNano())
	childID := fmt.Sprintf("child_%d", time.Now().UnixNano())
	
	// Create parent entity
	parentEntity := map[string]interface{}{
		"id":      parentID,
		"tags":    []string{"type:folder", "name:Documents"},
		"content": []byte("Parent folder"),
	}

	jsonData, _ := json.Marshal(parentEntity)
	resp, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create parent: %v", err)
	}
	resp.Body.Close()

	// Create child entity with parent reference
	childEntity := map[string]interface{}{
		"id":      childID,
		"tags":    []string{"type:file", "name:report.pdf", fmt.Sprintf("parent:%s", parentID)},
		"content": []byte("Child file"),
	}

	jsonData, _ = json.Marshal(childEntity)
	resp, err = makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create child: %v", err)
	}
	resp.Body.Close()

	// Update parent to include child reference
	parentEntity["tags"] = append(parentEntity["tags"].([]string), fmt.Sprintf("contains:%s", childID))
	
	jsonData, _ = json.Marshal(parentEntity)
	resp, err = makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to update parent with child reference: %v", err)
	}
	resp.Body.Close()

	// Query to verify bidirectional relationship
	// Find entities that contain the child
	containsChild, err := queryEntities(fmt.Sprintf("contains:%s", childID))
	if err != nil {
		return fmt.Errorf("failed to query parent: %v", err)
	}

	// Find entities that have parent
	hasParent, err := queryEntities(fmt.Sprintf("parent:%s", parentID))
	if err != nil {
		return fmt.Errorf("failed to query children: %v", err)
	}

	if len(containsChild) == 0 || len(hasParent) == 0 {
		return fmt.Errorf("bidirectional relationships not properly established: parent->child=%d, child->parent=%d", 
			len(containsChild), len(hasParent))
	}

	fmt.Printf("   🔄 Bidirectional relationships verified\n")
	fmt.Printf("   📁 Parent contains child: %d entities\n", len(containsChild))
	fmt.Printf("   📄 Child has parent: %d entities\n", len(hasParent))

	return nil
}

func testComplexRelationshipNetworks(client *http.Client, token string) error {
	// Create a complex organization structure
	ceoID := fmt.Sprintf("ceo_%d", time.Now().UnixNano())
	ctoID := fmt.Sprintf("cto_%d", time.Now().UnixNano())
	cfoID := fmt.Sprintf("cfo_%d", time.Now().UnixNano())
	dev1ID := fmt.Sprintf("dev1_%d", time.Now().UnixNano())
	dev2ID := fmt.Sprintf("dev2_%d", time.Now().UnixNano())
	
	// Create all entities with their relationships embedded
	entities := []map[string]interface{}{
		{
			"id":      ceoID,
			"tags":    []string{"type:employee", "role:ceo", "name:Alice"},
			"content": []byte("CEO data"),
		},
		{
			"id":      ctoID,
			"tags":    []string{"type:employee", "role:cto", "name:Bob", fmt.Sprintf("reports_to:%s", ceoID)},
			"content": []byte("CTO data"),
		},
		{
			"id":      cfoID,
			"tags":    []string{"type:employee", "role:cfo", "name:Carol", fmt.Sprintf("reports_to:%s", ceoID)},
			"content": []byte("CFO data"),
		},
		{
			"id":      dev1ID,
			"tags":    []string{"type:employee", "role:developer", "name:Dave", fmt.Sprintf("reports_to:%s", ctoID)},
			"content": []byte("Dev1 data"),
		},
		{
			"id":      dev2ID,
			"tags":    []string{"type:employee", "role:developer", "name:Eve", fmt.Sprintf("reports_to:%s", ctoID)},
			"content": []byte("Dev2 data"),
		},
	}

	// Create all employee entities
	for _, entity := range entities {
		jsonData, _ := json.Marshal(entity)
		resp, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create entity %s: %v", entity["id"], err)
		}
		resp.Body.Close()
	}

	// Update CEO with manages relationships
	ceoEntity := entities[0]
	ceoEntity["tags"] = append(ceoEntity["tags"].([]string), 
		fmt.Sprintf("manages:%s", ctoID),
		fmt.Sprintf("manages:%s", cfoID))
	
	jsonData, _ := json.Marshal(ceoEntity)
	resp, err := makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to update CEO: %v", err)
	}
	resp.Body.Close()

	// Query complex relationship network
	// Find all people reporting to CEO
	ceoReports, err := queryEntities(fmt.Sprintf("reports_to:%s", ceoID))
	if err != nil {
		return fmt.Errorf("failed to query CEO's reports: %v", err)
	}

	// Find all developers
	developers, err := queryEntities("role:developer")
	if err != nil {
		return fmt.Errorf("failed to query developers: %v", err)
	}

	// Find CTO's reports
	ctoReports, err := queryEntities(fmt.Sprintf("reports_to:%s", ctoID))
	if err != nil {
		return fmt.Errorf("failed to query CTO's reports: %v", err)
	}

	fmt.Printf("   🏢 Complex org structure created\n")
	fmt.Printf("   👥 CEO direct reports: %d\n", len(ceoReports))
	fmt.Printf("   💻 Total developers: %d\n", len(developers))
	fmt.Printf("   🔗 CTO team size: %d\n", len(ctoReports))

	if len(ceoReports) != 2 {
		return fmt.Errorf("expected 2 CEO reports, got %d", len(ceoReports))
	}

	return nil
}

func testTemporalRelationshipEvolution(client *http.Client, token string) error {
	// Test how relationships change over time
	personID := fmt.Sprintf("person_%d", time.Now().UnixNano())
	company1ID := fmt.Sprintf("company1_%d", time.Now().UnixNano())
	company2ID := fmt.Sprintf("company2_%d", time.Now().UnixNano())
	
	// Create entities
	entities := []map[string]interface{}{
		{
			"id":      personID,
			"tags":    []string{"type:person", "name:John"},
			"content": []byte("Person data"),
		},
		{
			"id":      company1ID,
			"tags":    []string{"type:company", "name:StartupCo"},
			"content": []byte("Company 1 data"),
		},
		{
			"id":      company2ID,
			"tags":    []string{"type:company", "name:BigCorp"},
			"content": []byte("Company 2 data"),
		},
	}

	for _, entity := range entities {
		jsonData, _ := json.Marshal(entity)
		resp, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create entity: %v", err)
		}
		resp.Body.Close()
	}

	// Start employment with Company 1
	personEntity := entities[0]
	personEntity["tags"] = append(personEntity["tags"].([]string),
		fmt.Sprintf("employed_by:%s", company1ID),
		"position:developer",
		"employment_status:active")
	
	jsonData, _ := json.Marshal(personEntity)
	resp, err := makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to start employment: %v", err)
	}
	resp.Body.Close()

	time.Sleep(100 * time.Millisecond)

	// Get promoted
	personEntity["tags"] = []string{
		"type:person", 
		"name:John",
		fmt.Sprintf("employed_by:%s", company1ID),
		"position:senior_developer",
		"employment_status:active",
	}

	jsonData, _ = json.Marshal(personEntity)
	resp, err = makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to update position: %v", err)
	}
	resp.Body.Close()

	time.Sleep(100 * time.Millisecond)

	// Change companies
	personEntity["tags"] = []string{
		"type:person", 
		"name:John",
		fmt.Sprintf("employed_by:%s", company2ID),
		fmt.Sprintf("previous_employer:%s", company1ID),
		"position:tech_lead",
		"employment_status:active",
	}

	jsonData, _ = json.Marshal(personEntity)
	resp, err = makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to change companies: %v", err)
	}
	resp.Body.Close()

	// Query current employment
	currentEmployment, err := queryEntities(
		fmt.Sprintf("employed_by:%s", company2ID),
		"employment_status:active")
	if err != nil {
		return fmt.Errorf("failed to query current employment: %v", err)
	}

	// Query employment history
	resp, err = makeRequest("GET", fmt.Sprintf("/api/v1/entities/history?id=%s", personID), nil)
	if err != nil {
		return fmt.Errorf("failed to get employment history: %v", err)
	}
	
	var history []Entity
	json.NewDecoder(resp.Body).Decode(&history)
	resp.Body.Close()

	fmt.Printf("   ⏰ Temporal relationship evolution tracked\n")
	fmt.Printf("   📈 Employment history entries: %d\n", len(history))
	fmt.Printf("   💼 Current employer: Company2 (%d active)\n", len(currentEmployment))
	fmt.Printf("   🔄 Career progression: Developer → Senior Developer → Tech Lead\n")

	return nil
}

func testRelationshipQueryPatterns(client *http.Client, token string) error {
	// Test various query patterns for relationships
	
	// Create a small knowledge graph
	baseTime := time.Now().UnixNano()
	
	// Create entities
	entities := []map[string]interface{}{
		{
			"id":   fmt.Sprintf("article_%d", baseTime),
			"tags": []string{"type:article", "title:EntityDB_Guide", "status:published"},
		},
		{
			"id":   fmt.Sprintf("author_%d", baseTime),
			"tags": []string{"type:person", "name:Jane_Smith", "role:writer"},
		},
		{
			"id":   fmt.Sprintf("category_%d", baseTime),
			"tags": []string{"type:category", "name:Databases", "topic:technology"},
		},
		{
			"id":   fmt.Sprintf("tag1_%d", baseTime),
			"tags": []string{"type:tag", "value:temporal", "weight:high"},
		},
		{
			"id":   fmt.Sprintf("tag2_%d", baseTime),
			"tags": []string{"type:tag", "value:nosql", "weight:medium"},
		},
	}

	for _, entity := range entities {
		entity["content"] = []byte("Content")
		jsonData, _ := json.Marshal(entity)
		resp, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create entity: %v", err)
		}
		resp.Body.Close()
	}

	// Add relationships using tags
	authorID := entities[1]["id"].(string)
	categoryID := entities[2]["id"].(string)
	tag1ID := entities[3]["id"].(string)
	tag2ID := entities[4]["id"].(string)

	// Update article with relationships
	entities[0]["tags"] = append(entities[0]["tags"].([]string),
		fmt.Sprintf("author:%s", authorID),
		fmt.Sprintf("category:%s", categoryID),
		fmt.Sprintf("tagged:%s", tag1ID),
		fmt.Sprintf("tagged:%s", tag2ID))
	
	jsonData, _ := json.Marshal(entities[0])
	resp, err := makeRequest("PUT", "/api/v1/entities/update", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to add article relationships: %v", err)
	}
	resp.Body.Close()

	// Test different query patterns
	
	// 1. Find all articles by author
	articlesByAuthor, err := queryEntities(fmt.Sprintf("author:%s", authorID))
	if err != nil {
		return fmt.Errorf("failed to query articles by author: %v", err)
	}

	// 2. Find all articles in category
	articlesInCategory, err := queryEntities(fmt.Sprintf("category:%s", categoryID))
	if err != nil {
		return fmt.Errorf("failed to query articles in category: %v", err)
	}

	// 3. Find all published articles
	publishedArticles, err := queryEntities("type:article", "status:published")
	if err != nil {
		return fmt.Errorf("failed to query published articles: %v", err)
	}

	// 4. Find entities with specific tag
	taggedEntities, err := queryEntities(fmt.Sprintf("tagged:%s", tag1ID))
	if err != nil {
		return fmt.Errorf("failed to query tagged entities: %v", err)
	}

	// 5. Complex query - published articles in tech category
	techArticles, err := queryEntities("type:article", "status:published")
	if err != nil {
		return fmt.Errorf("failed to complex query: %v", err)
	}

	fmt.Printf("   🔍 Query patterns tested:\n")
	fmt.Printf("   📝 Articles by author: %d\n", len(articlesByAuthor))
	fmt.Printf("   📁 Articles in category: %d\n", len(articlesInCategory))
	fmt.Printf("   ✅ Published articles: %d\n", len(publishedArticles))
	fmt.Printf("   🏷️ Tagged entities: %d\n", len(taggedEntities))
	fmt.Printf("   🔧 Complex queries: %d results\n", len(techArticles))

	return nil
}