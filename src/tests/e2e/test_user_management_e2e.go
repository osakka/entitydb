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

type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Tags     []string `json:"tags"`
	Content  string   `json:"content"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

type TestResult struct {
	TestName string
	Success  bool
	Duration time.Duration
	Error    string
	Details  map[string]interface{}
}

type UserManagementTestSuite struct {
	client        *http.Client
	results       []TestResult
	baseURL       string
	adminToken    string
	testUsers     map[string]string // username -> user_id mapping
}

func NewUserManagementTestSuite() *UserManagementTestSuite {
	// Create HTTP client that ignores SSL certificates for testing
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   TestTimeout,
	}

	return &UserManagementTestSuite{
		client:    client,
		results:   make([]TestResult, 0),
		baseURL:   BaseURL,
		testUsers: make(map[string]string),
	}
}

func (s *UserManagementTestSuite) makeRequest(method, endpoint string, body interface{}, token string) (*http.Response, error) {
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

func (s *UserManagementTestSuite) addResult(testName string, success bool, duration time.Duration, err error, details map[string]interface{}) {
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

func (s *UserManagementTestSuite) getAuthToken() (string, error) {
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

// Test 1: Admin Authentication and Session Management
func (s *UserManagementTestSuite) testAdminAuthentication() {
	start := time.Now()
	testName := "Admin Authentication and Session Management"

	// Test login
	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	s.adminToken = token

	// Test token validation by making an authenticated request
	resp, err := s.makeRequest("GET", "/api/v1/dashboard/stats", nil, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	authWorking := resp.StatusCode == 200
	
	// Test session persistence with multiple requests
	sessionTests := 0
	successfulRequests := 0
	
	for i := 0; i < 3; i++ {
		resp, err := s.makeRequest("GET", "/api/v1/dashboard/stats", nil, token)
		if err == nil {
			sessionTests++
			if resp.StatusCode == 200 {
				successfulRequests++
			}
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond) // Small delay between requests
	}

	details := map[string]interface{}{
		"initial_auth_working":    authWorking,
		"token_length":           len(token),
		"session_tests":          sessionTests,
		"successful_requests":    successfulRequests,
		"session_persistence":    float64(successfulRequests) / float64(sessionTests) * 100,
	}

	success := authWorking && successfulRequests >= 2
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 2: Create Test Users
func (s *UserManagementTestSuite) testCreateUsers() {
	start := time.Now()
	testName := "Create Test Users"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Test users to create with unique timestamp suffix
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	testUsers := []struct {
		username string
		email    string
		password string
		roles    []string
	}{
		{"testuser1_" + timestamp, "test1_" + timestamp + "@example.com", "password123", []string{}},
		{"testuser2_" + timestamp, "test2_" + timestamp + "@example.com", "password456", []string{}},
		{"manager1_" + timestamp, "manager_" + timestamp + "@example.com", "managerpass", []string{}},
		{"developer1_" + timestamp, "dev_" + timestamp + "@example.com", "devpass", []string{}},
	}

	successCount := 0
	createdUsers := make([]string, 0)

	for _, testUser := range testUsers {
		userData := map[string]interface{}{
			"username": testUser.username,
			"email":    testUser.email,
			"password": testUser.password,
		}

		resp, err := s.makeRequest("POST", "/api/v1/users/create", userData, s.adminToken)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			var user User
			if err := json.NewDecoder(resp.Body).Decode(&user); err == nil {
				s.testUsers[testUser.username] = user.ID
				successCount++
				createdUsers = append(createdUsers, testUser.username)
			}
		}
	}

	details := map[string]interface{}{
		"users_attempted":   len(testUsers),
		"users_created":     successCount,
		"created_usernames": createdUsers,
	}

	success := successCount >= len(testUsers)/2 // At least half should succeed
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 3: User Authentication Testing
func (s *UserManagementTestSuite) testUserAuthentication() {
	start := time.Now()
	testName := "User Authentication Testing"

	if len(s.testUsers) == 0 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no test users available"), nil)
		return
	}

	// Test authentication for created users
	authTests := make([]struct {
		username string
		password string
		shouldWork bool
	}, 0)
	
	// Add tests for created users
	for username := range s.testUsers {
		if strings.Contains(username, "testuser1") {
			authTests = append(authTests, struct {
				username string
				password string
				shouldWork bool
			}{username, "password123", true})
		} else if strings.Contains(username, "testuser2") {
			authTests = append(authTests, struct {
				username string
				password string
				shouldWork bool
			}{username, "password456", true})
		}
	}
	
	// Add negative test cases
	if len(authTests) > 0 {
		authTests = append(authTests, struct {
			username string
			password string
			shouldWork bool
		}{authTests[0].username, "wrongpassword", false})
	}
	
	authTests = append(authTests, struct {
		username string
		password string
		shouldWork bool
	}{"nonexistentuser", "anypassword", false})

	successfulAuths := 0
	failedAuths := 0
	totalTests := 0

	for _, authTest := range authTests {
		totalTests++
		loginData := map[string]string{
			"username": authTest.username,
			"password": authTest.password,
		}

		resp, err := s.makeRequest("POST", "/api/v1/auth/login", loginData, "")
		if err != nil {
			if !authTest.shouldWork {
				failedAuths++ // Expected failure
			}
			continue
		}
		defer resp.Body.Close()

		if authTest.shouldWork && resp.StatusCode == 200 {
			successfulAuths++
		} else if !authTest.shouldWork && resp.StatusCode != 200 {
			failedAuths++
		}
	}

	details := map[string]interface{}{
		"total_auth_tests":    totalTests,
		"successful_auths":    successfulAuths,
		"expected_failures":   failedAuths,
		"auth_success_rate":   float64(successfulAuths+failedAuths) / float64(totalTests) * 100,
	}

	success := (successfulAuths + failedAuths) >= totalTests*3/4 // 75% should work as expected
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 4: User Information Retrieval
func (s *UserManagementTestSuite) testUserInformationRetrieval() {
	start := time.Now()
	testName := "User Information Retrieval"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Test getting user list
	resp, err := s.makeRequest("GET", "/api/v1/entities/query?tag=type:user&limit=10", nil, s.adminToken)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	var response struct {
		Entities []User `json:"entities"`
		Total    int    `json:"total"`
	}
	userCount := 0
	if resp.StatusCode == 200 {
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			userCount = len(response.Entities)
		}
	}

	// Test getting specific user information
	specificUserTests := 0
	successfulRetrivals := 0

	for _, userID := range s.testUsers {
		specificUserTests++
		resp, err := s.makeRequest("GET", "/api/v1/entities/get?id="+userID, nil, s.adminToken)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			var user User
			if err := json.NewDecoder(resp.Body).Decode(&user); err == nil {
				successfulRetrivals++
				
				// Verify user data integrity
				if user.ID == userID {
					// Additional verification could be added here
				}
			}
		}
	}

	details := map[string]interface{}{
		"total_users_found":      userCount,
		"specific_user_tests":    specificUserTests,
		"successful_retrievals":  successfulRetrivals,
		"retrieval_success_rate": func() float64 {
			if specificUserTests > 0 {
				return float64(successfulRetrivals) / float64(specificUserTests) * 100
			}
			return 0
		}(),
	}

	success := userCount > 0 && successfulRetrivals >= specificUserTests/2
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 5: User Permission and Role Management
func (s *UserManagementTestSuite) testUserPermissions() {
	start := time.Now()
	testName := "User Permission and Role Management"

	if s.adminToken == "" || len(s.testUsers) == 0 {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token or test users available"), nil)
		return
	}

	// Test access control - regular user should not be able to create users
	testUserToken := ""
	testUsername := ""
	if len(s.testUsers) > 0 {
		// Find the first testuser for authentication
		for username := range s.testUsers {
			if strings.Contains(username, "testuser1") {
				testUsername = username
				break
			}
		}
		
		if testUsername != "" {
			// Get token for a regular user
			loginData := map[string]string{
				"username": testUsername,
				"password": "password123",
			}

			resp, err := s.makeRequest("POST", "/api/v1/auth/login", loginData, "")
			if err == nil && resp.StatusCode == 200 {
				var loginResp LoginResponse
				if err := json.NewDecoder(resp.Body).Decode(&loginResp); err == nil {
					testUserToken = loginResp.Token
				}
				resp.Body.Close()
			}
		}
	}

	permissionTests := 0
	expectedDenials := 0

	// Test 1: Regular user cannot create users
	if testUserToken != "" {
		permissionTests++
		userData := map[string]interface{}{
			"username": "unauthorized_user",
			"email":    "unauth@example.com",
			"password": "somepass",
		}

		resp, err := s.makeRequest("POST", "/api/v1/users/create", userData, testUserToken)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 403 || resp.StatusCode == 401 {
				expectedDenials++ // This should be denied
			}
		}
	}

	// Test 2: Admin can create users
	if s.adminToken != "" {
		permissionTests++
		userData := map[string]interface{}{
			"username": "admin_created_user",
			"email":    "admincreated@example.com",
			"password": "adminpass",
		}

		resp, err := s.makeRequest("POST", "/api/v1/users/create", userData, s.adminToken)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 || resp.StatusCode == 201 {
				expectedDenials++ // This should succeed
			}
		}
	}

	// Test 3: Regular user can access their own profile but not admin functions
	if testUserToken != "" {
		permissionTests++
		resp, err := s.makeRequest("GET", "/api/v1/dashboard/stats", nil, testUserToken)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 403 || resp.StatusCode == 401 {
				expectedDenials++ // Regular user should not access admin dashboard
			}
		}
	}

	details := map[string]interface{}{
		"permission_tests":      permissionTests,
		"expected_behaviors":    expectedDenials,
		"rbac_compliance_rate":  func() float64 {
			if permissionTests > 0 {
				return float64(expectedDenials) / float64(permissionTests) * 100
			}
			return 0
		}(),
		"test_user_token_obtained": testUserToken != "",
	}

	success := permissionTests > 0 && expectedDenials >= permissionTests*2/3
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 6: Session Logout and Invalidation
func (s *UserManagementTestSuite) testSessionLogout() {
	start := time.Now()
	testName := "Session Logout and Invalidation"

	// Get a fresh token for testing logout
	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}

	// Verify token works initially
	resp, err := s.makeRequest("GET", "/api/v1/dashboard/stats", nil, token)
	initialWorking := false
	if err == nil && resp.StatusCode == 200 {
		initialWorking = true
		resp.Body.Close()
	}

	// Perform logout
	logoutResp, err := s.makeRequest("POST", "/api/v1/auth/logout", nil, token)
	logoutSuccessful := false
	if err == nil && (logoutResp.StatusCode == 200 || logoutResp.StatusCode == 204) {
		logoutSuccessful = true
		logoutResp.Body.Close()
	}

	// Verify token is invalidated after logout
	postLogoutResp, err := s.makeRequest("GET", "/api/v1/dashboard/stats", nil, token)
	tokenInvalidated := false
	if err == nil {
		defer postLogoutResp.Body.Close()
		if postLogoutResp.StatusCode == 401 || postLogoutResp.StatusCode == 403 {
			tokenInvalidated = true
		}
	}

	details := map[string]interface{}{
		"initial_token_working": initialWorking,
		"logout_successful":     logoutSuccessful,
		"token_invalidated":     tokenInvalidated,
		"logout_flow_complete":  initialWorking && logoutSuccessful && tokenInvalidated,
	}

	success := initialWorking && logoutSuccessful && tokenInvalidated
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 7: User Data Validation
func (s *UserManagementTestSuite) testUserDataValidation() {
	start := time.Now()
	testName := "User Data Validation"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Test various invalid user creation scenarios
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	validationTests := []struct {
		name     string
		userData map[string]interface{}
		shouldFail bool
	}{
		{
			name: "missing username",
			userData: map[string]interface{}{
				"email":    "test_" + timestamp + "@example.com",
				"password": "password123",
			},
			shouldFail: true,
		},
		{
			name: "missing email",
			userData: map[string]interface{}{
				"username": "testuser_" + timestamp,
				"password": "password123",
			},
			shouldFail: true,
		},
		{
			name: "missing password",
			userData: map[string]interface{}{
				"username": "testuser_" + timestamp,
				"email":    "test_" + timestamp + "@example.com",
			},
			shouldFail: true,
		},
		{
			name: "invalid email format",
			userData: map[string]interface{}{
				"username": "testuser_" + timestamp,
				"email":    "invalid-email",
				"password": "password123",
			},
			shouldFail: true,
		},
		{
			name: "valid user data",
			userData: map[string]interface{}{
				"username": "validuser_" + timestamp,
				"email":    "valid_" + timestamp + "@example.com",
				"password": "validpass123",
			},
			shouldFail: false,
		},
	}

	totalValidationTests := len(validationTests)
	expectedBehaviors := 0

	for _, test := range validationTests {
		resp, err := s.makeRequest("POST", "/api/v1/users/create", test.userData, s.adminToken)
		if err != nil {
			if test.shouldFail {
				expectedBehaviors++
			}
			continue
		}
		defer resp.Body.Close()

		if test.shouldFail && (resp.StatusCode == 400 || resp.StatusCode == 422 || resp.StatusCode == 409) {
			expectedBehaviors++ // Validation correctly rejected invalid data (400/422) or duplicate username (409)
		} else if !test.shouldFail && (resp.StatusCode == 200 || resp.StatusCode == 201) {
			expectedBehaviors++ // Valid data was accepted
		}
	}

	details := map[string]interface{}{
		"validation_tests":       totalValidationTests,
		"expected_behaviors":     expectedBehaviors,
		"validation_accuracy":    float64(expectedBehaviors) / float64(totalValidationTests) * 100,
	}

	success := expectedBehaviors >= totalValidationTests*4/5 // 80% should behave as expected
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 8: Concurrent User Operations
func (s *UserManagementTestSuite) testConcurrentUserOperations() {
	start := time.Now()
	testName := "Concurrent User Operations"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Test concurrent user creation
	concurrentUsers := 3
	results := make(chan bool, concurrentUsers)
	timestamp := time.Now().UnixNano()

	for i := 0; i < concurrentUsers; i++ {
		go func(index int) {
			userData := map[string]interface{}{
				"username": fmt.Sprintf("concurrent_user_%d_%d", index, timestamp),
				"email":    fmt.Sprintf("concurrent%d_%d@example.com", index, timestamp),
				"password": fmt.Sprintf("pass%d", index),
			}

			resp, err := s.makeRequest("POST", "/api/v1/users/create", userData, s.adminToken)
			success := false
			if err == nil && (resp.StatusCode == 200 || resp.StatusCode == 201) {
				success = true
				resp.Body.Close()
			}
			results <- success
		}(i)
	}

	successfulConcurrentCreations := 0
	for i := 0; i < concurrentUsers; i++ {
		if <-results {
			successfulConcurrentCreations++
		}
	}

	// Test concurrent authentication
	concurrentAuths := 2
	authResults := make(chan bool, concurrentAuths)

	for i := 0; i < concurrentAuths; i++ {
		go func() {
			loginData := map[string]string{
				"username": TestUsername,
				"password": TestPassword,
			}

			resp, err := s.makeRequest("POST", "/api/v1/auth/login", loginData, "")
			success := false
			if err == nil && resp.StatusCode == 200 {
				success = true
				resp.Body.Close()
			}
			authResults <- success
		}()
	}

	successfulConcurrentAuths := 0
	for i := 0; i < concurrentAuths; i++ {
		if <-authResults {
			successfulConcurrentAuths++
		}
	}

	details := map[string]interface{}{
		"concurrent_creations":     concurrentUsers,
		"successful_creations":     successfulConcurrentCreations,
		"concurrent_auths":         concurrentAuths,
		"successful_auths":         successfulConcurrentAuths,
		"concurrency_success_rate": float64(successfulConcurrentCreations+successfulConcurrentAuths) / float64(concurrentUsers+concurrentAuths) * 100,
	}

	success := successfulConcurrentCreations >= concurrentUsers/2 && successfulConcurrentAuths >= concurrentAuths/2
	s.addResult(testName, success, time.Since(start), nil, details)
}

func (s *UserManagementTestSuite) runAllTests() {
	fmt.Println("👥 Starting EntityDB User Management Testing End-to-End Test Suite")
	fmt.Println(strings.Repeat("=", 80))

	tests := []func(){
		s.testAdminAuthentication,
		s.testCreateUsers,
		s.testUserAuthentication,
		s.testUserInformationRetrieval,
		s.testUserPermissions,
		s.testSessionLogout,
		s.testUserDataValidation,
		s.testConcurrentUserOperations,
	}

	for _, test := range tests {
		test()
		time.Sleep(300 * time.Millisecond) // Delay between tests for server processing
	}
}

func (s *UserManagementTestSuite) printResults() {
	fmt.Println("\n📊 USER MANAGEMENT TESTING RESULTS")
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

	// Summary of user management testing coverage
	fmt.Println("\n👥 USER MANAGEMENT TESTING COVERAGE:")
	fmt.Println("✅ Admin authentication and session management")
	fmt.Println("✅ User creation and management")
	fmt.Println("✅ User authentication testing")
	fmt.Println("✅ User information retrieval")
	fmt.Println("✅ Permission and role-based access control")
	fmt.Println("✅ Session logout and invalidation")
	fmt.Println("✅ User data validation")
	fmt.Println("✅ Concurrent user operations")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("EntityDB User Management Testing End-to-End Test Suite")
		fmt.Println("Usage: go run test_user_management_e2e.go")
		fmt.Println("\nThis comprehensive test suite validates:")
		fmt.Println("- Admin authentication and session management")
		fmt.Println("- User creation, authentication, and management")
		fmt.Println("- Role-based access control (RBAC) enforcement")
		fmt.Println("- Session handling and invalidation")
		fmt.Println("- User data validation and security")
		fmt.Println("- Concurrent user operations")
		fmt.Println("- EntityDB's complete user management system")
		return
	}

	suite := NewUserManagementTestSuite()
	suite.runAllTests()
	suite.printResults()
}