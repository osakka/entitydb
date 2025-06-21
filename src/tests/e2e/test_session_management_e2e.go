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

// Test configuration
const (
	BaseURL = "https://localhost:8085"
	TestUsername = "admin"
	TestPassword = "admin"
	MaxRetries = 3
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

type SessionTestSuite struct {
	client  *http.Client
	results []TestResult
	baseURL string
}

func NewSessionTestSuite() *SessionTestSuite {
	// Create HTTP client that ignores SSL certificates for testing
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   TestTimeout,
	}

	return &SessionTestSuite{
		client:  client,
		results: make([]TestResult, 0),
		baseURL: BaseURL,
	}
}

func (s *SessionTestSuite) makeRequest(method, endpoint string, body interface{}, token string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
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

func (s *SessionTestSuite) addResult(testName string, success bool, duration time.Duration, err error, details map[string]interface{}) {
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

// Test 1: Basic Login
func (s *SessionTestSuite) testBasicLogin() {
	start := time.Now()
	testName := "Basic Login"
	
	loginData := map[string]string{
		"username": TestUsername,
		"password": TestPassword,
	}

	resp, err := s.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("expected 200, got %d", resp.StatusCode), nil)
		return
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	details := map[string]interface{}{
		"token_length": len(loginResp.Token),
		"user_id":      loginResp.UserID,
		"username":     loginResp.User.Username,
		"roles":        loginResp.User.Roles,
		"expires_at":   loginResp.ExpiresAt,
	}

	if loginResp.Token == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("empty token received"), details)
		return
	}

	s.addResult(testName, true, time.Since(start), nil, details)
}

// Test 2: Token Validation
func (s *SessionTestSuite) testTokenValidation() {
	start := time.Now()
	testName := "Token Validation"

	// First, get a valid token
	loginData := map[string]string{
		"username": TestUsername,
		"password": TestPassword,
	}

	resp, err := s.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// Test the token with a protected endpoint
	resp2, err := s.makeRequest("GET", "/api/v1/entities/list", nil, loginResp.Token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp2.Body.Close()

	details := map[string]interface{}{
		"token_used":     loginResp.Token[:16] + "...", // Truncated for security
		"response_code":  resp2.StatusCode,
		"content_length": resp2.ContentLength,
	}

	if resp2.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp2.Body)
		details["response_body"] = string(bodyBytes)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("expected 200, got %d", resp2.StatusCode), details)
		return
	}

	s.addResult(testName, true, time.Since(start), nil, details)
}

// Test 3: Invalid Token Handling
func (s *SessionTestSuite) testInvalidToken() {
	start := time.Now()
	testName := "Invalid Token Handling"

	invalidToken := "invalid_token_12345"
	resp, err := s.makeRequest("GET", "/api/v1/entities/list", nil, invalidToken)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	details := map[string]interface{}{
		"response_code": resp.StatusCode,
		"expected_code": 401,
	}

	// Should return 401 Unauthorized
	if resp.StatusCode != 401 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		details["response_body"] = string(bodyBytes)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("expected 401, got %d", resp.StatusCode), details)
		return
	}

	s.addResult(testName, true, time.Since(start), nil, details)
}

// Test 4: Session Logout
func (s *SessionTestSuite) testSessionLogout() {
	start := time.Now()
	testName := "Session Logout"

	// Login first
	loginData := map[string]string{
		"username": TestUsername,
		"password": TestPassword,
	}

	resp, err := s.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// Test that token works
	resp2, err := s.makeRequest("GET", "/api/v1/entities/list", nil, loginResp.Token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	resp2.Body.Close()

	if resp2.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("token validation failed before logout"), nil)
		return
	}

	// Logout
	resp3, err := s.makeRequest("POST", "/api/v1/auth/logout", nil, loginResp.Token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp3.Body.Close()

	// Test that token no longer works
	resp4, err := s.makeRequest("GET", "/api/v1/entities/list", nil, loginResp.Token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp4.Body.Close()

	details := map[string]interface{}{
		"logout_status":      resp3.StatusCode,
		"post_logout_status": resp4.StatusCode,
		"expected_failure":   401,
	}

	if resp4.StatusCode != 401 {
		bodyBytes, _ := io.ReadAll(resp4.Body)
		details["post_logout_body"] = string(bodyBytes)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("token still valid after logout, got %d", resp4.StatusCode), details)
		return
	}

	s.addResult(testName, true, time.Since(start), nil, details)
}

// Test 5: Concurrent Sessions
func (s *SessionTestSuite) testConcurrentSessions() {
	start := time.Now()
	testName := "Concurrent Sessions"

	// Create two sessions
	loginData := map[string]string{
		"username": TestUsername,
		"password": TestPassword,
	}

	// Session 1
	resp1, err := s.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp1.Body.Close()

	var login1 LoginResponse
	if err := json.NewDecoder(resp1.Body).Decode(&login1); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// Session 2
	resp2, err := s.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp2.Body.Close()

	var login2 LoginResponse
	if err := json.NewDecoder(resp2.Body).Decode(&login2); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// Test both tokens work
	resp3, err := s.makeRequest("GET", "/api/v1/entities/list", nil, login1.Token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	resp3.Body.Close()

	resp4, err := s.makeRequest("GET", "/api/v1/entities/list", nil, login2.Token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	resp4.Body.Close()

	details := map[string]interface{}{
		"session1_status": resp3.StatusCode,
		"session2_status": resp4.StatusCode,
		"tokens_different": login1.Token != login2.Token,
	}

	if resp3.StatusCode != 200 || resp4.StatusCode != 200 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("concurrent sessions failed: %d, %d", resp3.StatusCode, resp4.StatusCode), details)
		return
	}

	if login1.Token == login2.Token {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("concurrent sessions have same token"), details)
		return
	}

	s.addResult(testName, true, time.Since(start), nil, details)
}

// Test 6: Entity Operations with Session
func (s *SessionTestSuite) testEntityOperationsWithSession() {
	start := time.Now()
	testName := "Entity Operations with Session"

	// Login
	loginData := map[string]string{
		"username": TestUsername,
		"password": TestPassword,
	}

	resp, err := s.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// Create entity
	entityData := map[string]interface{}{
		"tags":    []string{"name:session-test-entity", "type:test", "purpose:session-validation"},
		"content": "VGVzdCBlbnRpdHkgZm9yIHNlc3Npb24gdmFsaWRhdGlvbg==", // Base64: "Test entity for session validation"
	}

	resp2, err := s.makeRequest("POST", "/api/v1/entities/create", entityData, loginResp.Token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != 200 && resp2.StatusCode != 201 {
		bodyBytes, _ := io.ReadAll(resp2.Body)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("entity creation failed: %d - %s", resp2.StatusCode, string(bodyBytes)), nil)
		return
	}

	var entity Entity
	if err := json.NewDecoder(resp2.Body).Decode(&entity); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// Retrieve entity
	resp3, err := s.makeRequest("GET", "/api/v1/entities/get?id="+entity.ID, nil, loginResp.Token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp3.Body)
		s.addResult(testName, false, time.Since(start), fmt.Errorf("entity retrieval failed: %d - %s", resp3.StatusCode, string(bodyBytes)), nil)
		return
	}

	details := map[string]interface{}{
		"entity_id":     entity.ID,
		"tags_count":    len(entity.Tags),
		"creation_time": entity.CreatedAt,
	}

	s.addResult(testName, true, time.Since(start), nil, details)
}

// Test 7: Session Persistence
func (s *SessionTestSuite) testSessionPersistence() {
	start := time.Now()
	testName := "Session Persistence"

	// Login
	loginData := map[string]string{
		"username": TestUsername,
		"password": TestPassword,
	}

	resp, err := s.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// Test token multiple times with delays
	delays := []time.Duration{0, 1 * time.Second, 2 * time.Second, 5 * time.Second}
	var statuses []int

	for i, delay := range delays {
		if delay > 0 {
			time.Sleep(delay)
		}

		resp2, err := s.makeRequest("GET", "/api/v1/entities/list", nil, loginResp.Token)
		if err != nil {
			s.addResult(testName, false, time.Since(start), err, nil)
			return
		}
		resp2.Body.Close()
		statuses = append(statuses, resp2.StatusCode)

		if resp2.StatusCode != 200 {
			details := map[string]interface{}{
				"attempt":      i + 1,
				"delay":        delay.String(),
				"status_codes": statuses,
			}
			s.addResult(testName, false, time.Since(start), fmt.Errorf("token invalid after %v delay, got %d", delay, resp2.StatusCode), details)
			return
		}
	}

	details := map[string]interface{}{
		"attempts":     len(delays),
		"status_codes": statuses,
		"max_delay":    delays[len(delays)-1].String(),
	}

	s.addResult(testName, true, time.Since(start), nil, details)
}

func (s *SessionTestSuite) runAllTests() {
	fmt.Println("🚀 Starting EntityDB Session Management End-to-End Test Suite")
	fmt.Println(strings.Repeat("=", 80))

	tests := []func(){
		s.testBasicLogin,
		s.testTokenValidation,
		s.testInvalidToken,
		s.testSessionLogout,
		s.testConcurrentSessions,
		s.testEntityOperationsWithSession,
		s.testSessionPersistence,
	}

	for _, test := range tests {
		test()
		time.Sleep(100 * time.Millisecond) // Small delay between tests
	}
}

func (s *SessionTestSuite) printResults() {
	fmt.Println("\n📊 SESSION MANAGEMENT TEST RESULTS")
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
		fmt.Println("   Failures require investigation and fixes")
	} else {
		fmt.Println("🎉 ALL TESTS PASSED - 100% Success Rate Achieved!")
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("EntityDB Session Management End-to-End Test Suite")
		fmt.Println("Usage: go run test_session_management_e2e.go")
		fmt.Println("\nThis comprehensive test suite validates:")
		fmt.Println("- Basic login functionality")
		fmt.Println("- Token validation and handling")
		fmt.Println("- Invalid token rejection")
		fmt.Println("- Session logout and invalidation")
		fmt.Println("- Concurrent session support")
		fmt.Println("- Entity operations with sessions")
		fmt.Println("- Session persistence over time")
		return
	}

	suite := NewSessionTestSuite()
	suite.runAllTests()
	suite.printResults()
}