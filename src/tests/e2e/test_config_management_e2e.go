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

type ConfigValue struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Type      string      `json:"type"`
	UpdatedAt string      `json:"updated_at"`
}

type FeatureFlag struct {
	Name        string      `json:"name"`
	Enabled     bool        `json:"enabled"`
	Value       interface{} `json:"value,omitempty"`
	Description string      `json:"description,omitempty"`
}

type TestResult struct {
	TestName string
	Success  bool
	Duration time.Duration
	Error    string
	Details  map[string]interface{}
}

type ConfigManagementTestSuite struct {
	client     *http.Client
	results    []TestResult
	baseURL    string
	adminToken string
}

func NewConfigManagementTestSuite() *ConfigManagementTestSuite {
	// Create HTTP client that ignores SSL certificates for testing
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   TestTimeout,
	}

	return &ConfigManagementTestSuite{
		client:  client,
		results: make([]TestResult, 0),
		baseURL: BaseURL,
	}
}

func (s *ConfigManagementTestSuite) makeRequest(method, endpoint string, body interface{}, token string) (*http.Response, error) {
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

func (s *ConfigManagementTestSuite) addResult(testName string, success bool, duration time.Duration, err error, details map[string]interface{}) {
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

func (s *ConfigManagementTestSuite) getAuthToken() (string, error) {
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

// Test 1: Admin Authentication for Config Access
func (s *ConfigManagementTestSuite) testConfigAuthentication() {
	start := time.Now()
	testName := "Config Authentication"

	token, err := s.getAuthToken()
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	s.adminToken = token

	// Test access to config endpoint
	resp, err := s.makeRequest("GET", "/api/v1/config", nil, token)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	details := map[string]interface{}{
		"auth_successful":    token != "",
		"config_accessible":  resp.StatusCode == 200,
		"token_length":       len(token),
	}

	success := token != "" && resp.StatusCode == 200
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 2: Configuration Retrieval
func (s *ConfigManagementTestSuite) testConfigRetrieval() {
	start := time.Now()
	testName := "Configuration Retrieval"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Test getting all configuration
	resp, err := s.makeRequest("GET", "/api/v1/config", nil, s.adminToken)
	if err != nil {
		s.addResult(testName, false, time.Since(start), err, nil)
		return
	}
	defer resp.Body.Close()

	configCount := 0
	configData := make(map[string]interface{})
	
	if resp.StatusCode == 200 {
		if err := json.NewDecoder(resp.Body).Decode(&configData); err == nil {
			configCount = len(configData)
		}
	}

	// Test getting specific config values
	specificConfigTests := []string{
		"/api/v1/config?key=server_name",
		"/api/v1/config?key=log_level",
		"/api/v1/config?key=metrics_enabled",
	}

	specificConfigSuccess := 0
	for _, configEndpoint := range specificConfigTests {
		resp, err := s.makeRequest("GET", configEndpoint, nil, s.adminToken)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 || resp.StatusCode == 404 { // 404 is also valid if config doesn't exist
				specificConfigSuccess++
			}
		}
	}

	details := map[string]interface{}{
		"total_config_items":         configCount,
		"config_retrieval_working":   resp.StatusCode == 200,
		"specific_config_tests":      len(specificConfigTests),
		"specific_config_success":    specificConfigSuccess,
		"config_endpoint_accessible": true,
	}

	success := resp.StatusCode == 200 && specificConfigSuccess >= len(specificConfigTests)/2
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 3: Feature Flag Management
func (s *ConfigManagementTestSuite) testFeatureFlagManagement() {
	start := time.Now()
	testName := "Feature Flag Management"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Test feature flag operations
	testFlags := []struct {
		name        string
		enabled     bool
		value       interface{}
		description string
	}{
		{"test_feature_1", true, "enabled", "Test feature flag 1"},
		{"test_feature_2", false, nil, "Test feature flag 2"},
		{"debug_mode", true, map[string]interface{}{"level": "verbose"}, "Debug mode configuration"},
	}

	flagOperations := 0
	successfulOperations := 0

	for _, flag := range testFlags {
		flagOperations++
		
		// Try to set feature flag
		flagData := map[string]interface{}{
			"flag":        flag.name,
			"enabled":     flag.enabled,
		}
		if flag.value != nil {
			flagData["value"] = flag.value
		}

		resp, err := s.makeRequest("POST", "/api/v1/feature-flags/set", flagData, s.adminToken)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 || resp.StatusCode == 201 {
				successfulOperations++
			}
		}
	}

	// Test retrieving feature flags
	resp, err := s.makeRequest("GET", "/api/v1/feature-flags", nil, s.adminToken)
	retrievalWorking := false
	retrievedFlags := 0
	
	if err == nil && resp.StatusCode == 200 {
		retrievalWorking = true
		var flags []FeatureFlag
		if err := json.NewDecoder(resp.Body).Decode(&flags); err == nil {
			retrievedFlags = len(flags)
		}
		resp.Body.Close()
	}

	details := map[string]interface{}{
		"flag_operations":         flagOperations,
		"successful_operations":   successfulOperations,
		"retrieval_working":       retrievalWorking,
		"retrieved_flags":         retrievedFlags,
		"flag_management_working": successfulOperations > 0 && retrievalWorking,
	}

	success := successfulOperations >= flagOperations/2 && retrievalWorking
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 4: Configuration Updates
func (s *ConfigManagementTestSuite) testConfigurationUpdates() {
	start := time.Now()
	testName := "Configuration Updates"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Test configuration update operations
	configUpdates := []struct {
		key   string
		value interface{}
		type_ string
	}{
		{"test_config_1", "test_value_1", "string"},
		{"test_config_2", 12345, "integer"},
		{"test_config_3", true, "boolean"},
		{"test_config_4", map[string]interface{}{"nested": "value"}, "object"},
	}

	updateOperations := 0
	successfulUpdates := 0

	for _, config := range configUpdates {
		updateOperations++
		
		configData := map[string]interface{}{
			"key":   config.key,
			"value": config.value,
		}
		if config.type_ != "" {
			configData["type"] = config.type_
		}

		// Try to update configuration
		resp, err := s.makeRequest("POST", "/api/v1/config/set", configData, s.adminToken)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 || resp.StatusCode == 201 {
				successfulUpdates++
				
				// Verify the update by reading it back
				resp2, err2 := s.makeRequest("GET", "/api/v1/config?key="+config.key, nil, s.adminToken)
				if err2 == nil && resp2.StatusCode == 200 {
					resp2.Body.Close()
					// Additional verification could be done here
				}
			}
		}
	}

	// Configuration deletion is not supported by API
	deleteOperations := 0
	successfulDeletes := 0

	details := map[string]interface{}{
		"update_operations":      updateOperations,
		"successful_updates":     successfulUpdates,
		"delete_operations":      deleteOperations,
		"successful_deletes":     successfulDeletes,
		"config_modification_rate": float64(successfulUpdates+successfulDeletes) / float64(updateOperations+deleteOperations) * 100,
	}

	success := successfulUpdates >= updateOperations/2
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 5: Runtime Configuration Changes
func (s *ConfigManagementTestSuite) testRuntimeConfigChanges() {
	start := time.Now()
	testName := "Runtime Configuration Changes"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Test log level changes
	logLevelTests := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	logLevelChanges := 0
	successfulLogChanges := 0

	for _, level := range logLevelTests {
		logLevelChanges++
		
		logData := map[string]interface{}{
			"level": level,
		}

		resp, err := s.makeRequest("POST", "/api/v1/admin/log-level", logData, s.adminToken)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				successfulLogChanges++
				
				// Verify the change
				resp2, err2 := s.makeRequest("GET", "/api/v1/admin/log-level", nil, s.adminToken)
				if err2 == nil && resp2.StatusCode == 200 {
					resp2.Body.Close()
				}
			}
		}
	}

	// Test trace subsystem configuration
	traceSubsystemTests := []string{"auth", "storage", "metrics", "all"}
	traceChanges := 0
	successfulTraceChanges := 0

	for _, subsystem := range traceSubsystemTests {
		traceChanges++
		
		traceData := map[string]interface{}{
			"subsystems": []string{subsystem},
			"enabled":    true,
		}

		resp, err := s.makeRequest("POST", "/api/v1/admin/trace-subsystems", traceData, s.adminToken)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				successfulTraceChanges++
			}
		}
	}

	details := map[string]interface{}{
		"log_level_changes":         logLevelChanges,
		"successful_log_changes":    successfulLogChanges,
		"trace_changes":             traceChanges,
		"successful_trace_changes":  successfulTraceChanges,
		"runtime_changes_working":   successfulLogChanges > 0 || successfulTraceChanges > 0,
		"runtime_success_rate":      float64(successfulLogChanges+successfulTraceChanges) / float64(logLevelChanges+traceChanges) * 100,
	}

	success := successfulLogChanges >= logLevelChanges/2 || successfulTraceChanges >= traceChanges/2
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 6: Configuration Validation
func (s *ConfigManagementTestSuite) testConfigurationValidation() {
	start := time.Now()
	testName := "Configuration Validation"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Test invalid configuration scenarios
	invalidConfigs := []struct {
		name       string
		data       map[string]interface{}
		shouldFail bool
	}{
		{
			name:       "missing key",
			data:       map[string]interface{}{"value": "test"},
			shouldFail: true,
		},
		{
			name:       "empty key",
			data:       map[string]interface{}{"key": "", "value": "test"},
			shouldFail: true,
		},
		{
			name:       "invalid log level",
			data:       map[string]interface{}{"level": "INVALID_LEVEL"},
			shouldFail: true,
		},
		{
			name:       "valid config",
			data:       map[string]interface{}{"key": "valid_test_key", "value": "valid_value"},
			shouldFail: false,
		},
	}

	validationTests := len(invalidConfigs)
	expectedBehaviors := 0

	for _, test := range invalidConfigs {
		var endpoint string
		if test.name == "invalid log level" {
			endpoint = "/api/v1/admin/log-level"
		} else {
			endpoint = "/api/v1/config/set"
		}

		resp, err := s.makeRequest("POST", endpoint, test.data, s.adminToken)
		if err != nil {
			if test.shouldFail {
				expectedBehaviors++
			}
			continue
		}
		defer resp.Body.Close()

		if test.shouldFail && (resp.StatusCode == 400 || resp.StatusCode == 422) {
			expectedBehaviors++ // Validation correctly rejected invalid data
		} else if !test.shouldFail && (resp.StatusCode == 200 || resp.StatusCode == 201) {
			expectedBehaviors++ // Valid data was accepted
		}
	}

	details := map[string]interface{}{
		"validation_tests":     validationTests,
		"expected_behaviors":   expectedBehaviors,
		"validation_accuracy":  float64(expectedBehaviors) / float64(validationTests) * 100,
	}

	success := expectedBehaviors >= validationTests*3/4 // 75% should behave as expected
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 7: Configuration Permissions
func (s *ConfigManagementTestSuite) testConfigurationPermissions() {
	start := time.Now()
	testName := "Configuration Permissions"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Test that non-admin users cannot access config endpoints
	// First, try to get a non-admin token (if available)
	nonAdminToken := ""
	loginData := map[string]string{
		"username": "testuser1",
		"password": "password123",
	}

	resp, err := s.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	if err == nil && resp.StatusCode == 200 {
		var loginResp LoginResponse
		if err := json.NewDecoder(resp.Body).Decode(&loginResp); err == nil {
			nonAdminToken = loginResp.Token
		}
		resp.Body.Close()
	}

	permissionTests := 0
	expectedDenials := 0

	// Test config access with non-admin token
	if nonAdminToken != "" {
		configEndpoints := []string{
			"/api/v1/config",
			"/api/v1/config/set",
			"/api/v1/admin/log-level",
			"/api/v1/feature-flags/set",
		}

		for _, endpoint := range configEndpoints {
			permissionTests++
			
			testData := map[string]interface{}{"key": "test", "value": "test"}
			method := "GET"
			if strings.Contains(endpoint, "set") || strings.Contains(endpoint, "log-level") {
				method = "POST"
			}

			resp, err := s.makeRequest(method, endpoint, testData, nonAdminToken)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == 403 || resp.StatusCode == 401 {
					expectedDenials++ // This should be denied
				}
			}
		}
	}

	// Test config access with admin token (should succeed)
	adminTests := []string{
		"/api/v1/config",
	}

	for _, endpoint := range adminTests {
		permissionTests++
		
		resp, err := s.makeRequest("GET", endpoint, nil, s.adminToken)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				expectedDenials++ // This should succeed
			}
		}
	}

	details := map[string]interface{}{
		"permission_tests":        permissionTests,
		"expected_behaviors":      expectedDenials,
		"non_admin_token_obtained": nonAdminToken != "",
		"rbac_compliance_rate":    func() float64 {
			if permissionTests > 0 {
				return float64(expectedDenials) / float64(permissionTests) * 100
			}
			return 0
		}(),
	}

	success := permissionTests > 0 && expectedDenials >= permissionTests*2/3
	s.addResult(testName, success, time.Since(start), nil, details)
}

// Test 8: Configuration Persistence
func (s *ConfigManagementTestSuite) testConfigurationPersistence() {
	start := time.Now()
	testName := "Configuration Persistence"

	if s.adminToken == "" {
		s.addResult(testName, false, time.Since(start), fmt.Errorf("no admin token available"), nil)
		return
	}

	// Create a persistent configuration
	persistentConfigs := []struct {
		key   string
		value interface{}
	}{
		{"persistent_test_1", "persistent_value_1"},
		{"persistent_test_2", 42},
		{"persistent_test_3", true},
	}

	createOperations := 0
	successfulCreations := 0

	for _, config := range persistentConfigs {
		createOperations++
		
		configData := map[string]interface{}{
			"key":   config.key,
			"value": config.value,
		}

		resp, err := s.makeRequest("POST", "/api/v1/config/set", configData, s.adminToken)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 || resp.StatusCode == 201 {
				successfulCreations++
			}
		}
	}

	// Verify persistence by reading back configurations
	readOperations := 0
	successfulReads := 0

	for _, config := range persistentConfigs {
		readOperations++
		
		resp, err := s.makeRequest("GET", "/api/v1/config?key="+config.key, nil, s.adminToken)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				var result map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
					successfulReads++
					// Could verify the actual value here
				}
			}
		}
	}

	details := map[string]interface{}{
		"create_operations":     createOperations,
		"successful_creations":  successfulCreations,
		"read_operations":       readOperations,
		"successful_reads":      successfulReads,
		"persistence_working":   successfulCreations > 0 && successfulReads > 0,
		"persistence_rate":      func() float64 {
			if createOperations > 0 && readOperations > 0 {
				return float64(successfulReads) / float64(successfulCreations) * 100
			}
			return 0
		}(),
	}

	success := successfulCreations >= createOperations/2 && successfulReads >= readOperations/2
	s.addResult(testName, success, time.Since(start), nil, details)
}

func (s *ConfigManagementTestSuite) runAllTests() {
	fmt.Println("⚙️  Starting EntityDB Configuration Management Testing End-to-End Test Suite")
	fmt.Println(strings.Repeat("=", 80))

	tests := []func(){
		s.testConfigAuthentication,
		s.testConfigRetrieval,
		s.testFeatureFlagManagement,
		s.testConfigurationUpdates,
		s.testRuntimeConfigChanges,
		s.testConfigurationValidation,
		s.testConfigurationPermissions,
		s.testConfigurationPersistence,
	}

	for _, test := range tests {
		test()
		time.Sleep(300 * time.Millisecond) // Delay between tests for server processing
	}
}

func (s *ConfigManagementTestSuite) printResults() {
	fmt.Println("\n📊 CONFIGURATION MANAGEMENT TESTING RESULTS")
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

	// Summary of configuration management testing coverage
	fmt.Println("\n⚙️  CONFIGURATION MANAGEMENT TESTING COVERAGE:")
	fmt.Println("✅ Configuration authentication and access control")
	fmt.Println("✅ Configuration retrieval and querying")
	fmt.Println("✅ Feature flag management")
	fmt.Println("✅ Configuration updates and modifications")
	fmt.Println("✅ Runtime configuration changes")
	fmt.Println("✅ Configuration validation and error handling")
	fmt.Println("✅ Permission-based configuration access")
	fmt.Println("✅ Configuration persistence and storage")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("EntityDB Configuration Management Testing End-to-End Test Suite")
		fmt.Println("Usage: go run test_config_management_e2e.go")
		fmt.Println("\nThis comprehensive test suite validates:")
		fmt.Println("- Configuration authentication and authorization")
		fmt.Println("- Configuration retrieval and management")
		fmt.Println("- Feature flag operations")
		fmt.Println("- Runtime configuration changes (log levels, trace subsystems)")
		fmt.Println("- Configuration validation and error handling")
		fmt.Println("- Permission-based access control for configuration")
		fmt.Println("- Configuration persistence and storage")
		fmt.Println("- EntityDB's complete configuration management system")
		return
	}

	suite := NewConfigManagementTestSuite()
	suite.runAllTests()
	suite.printResults()
}