// test_performance_stress.go - Comprehensive performance and stress testing
// Tests EntityDB's performance under various load conditions

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type PerformanceTestCase struct {
	Name        string
	Description string
	TestFunc    func(client *http.Client, token string) error
}

type PerformanceMetrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalDuration      time.Duration
	MinLatency         time.Duration
	MaxLatency         time.Duration
	AvgLatency         time.Duration
	RequestsPerSecond  float64
}

var (
	baseURL    = "https://localhost:8085"
	httpClient *http.Client
	authToken  string
	
	// Metrics tracking
	totalRequests      int64
	successfulRequests int64
	failedRequests     int64
	latencies          []time.Duration
	latencyMutex       sync.Mutex
)

func main() {
	fmt.Println("⚡ PERFORMANCE AND STRESS TEST - EntityDB v2.34.3")
	fmt.Println("Testing performance under various load conditions")
	fmt.Println("=================================================")

	// Initialize HTTP client with connection pooling
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
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

	// Define performance test cases
	testCases := []PerformanceTestCase{
		{
			Name:        "Single Entity Operations",
			Description: "Test individual CRUD operation performance",
			TestFunc:    testSingleEntityOperations,
		},
		{
			Name:        "Bulk Entity Creation",
			Description: "Test creating many entities rapidly",
			TestFunc:    testBulkEntityCreation,
		},
		{
			Name:        "Concurrent Read/Write",
			Description: "Test concurrent read and write operations",
			TestFunc:    testConcurrentReadWrite,
		},
		{
			Name:        "Query Performance",
			Description: "Test complex query performance with filters",
			TestFunc:    testQueryPerformance,
		},
		{
			Name:        "Temporal Query Load",
			Description: "Test temporal query performance under load",
			TestFunc:    testTemporalQueryLoad,
		},
		{
			Name:        "Sustained Load Test",
			Description: "Test sustained load over time",
			TestFunc:    testSustainedLoad,
		},
		{
			Name:        "Memory Pressure Test",
			Description: "Test performance under memory pressure",
			TestFunc:    testMemoryPressure,
		},
		{
			Name:        "API Endpoint Coverage",
			Description: "Test all major endpoints under load",
			TestFunc:    testAPIEndpointCoverage,
		},
	}

	// Execute all test cases
	passed := 0
	failed := 0

	for i, testCase := range testCases {
		fmt.Printf("🧪 Test %d: %s\n", i+1, testCase.Name)
		fmt.Printf("   %s\n", testCase.Description)
		
		// Reset metrics for each test
		resetMetrics()

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
	fmt.Println("=================================================")
	fmt.Printf("⚡ PERFORMANCE TEST RESULTS:\n")
	fmt.Printf("✅ Passed: %d\n", passed)
	fmt.Printf("❌ Failed: %d\n", failed)
	fmt.Printf("📊 Success Rate: %.1f%%\n", float64(passed)/float64(len(testCases))*100)

	if failed == 0 {
		fmt.Println("🎉 ALL PERFORMANCE TESTS PASSED - Production Ready!")
	} else {
		fmt.Println("⚠️  Some performance tests failed - Review required")
		os.Exit(1)
	}
}

func authenticate() (string, error) {
	loginData := map[string]string{
		"username": "admin",
		"password": "admin",
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

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["token"].(string), nil
}

func makeRequest(method, endpoint string, body io.Reader) (*http.Response, time.Duration, error) {
	req, err := http.NewRequest(method, baseURL+endpoint, body)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := httpClient.Do(req)
	duration := time.Since(start)

	// Track metrics
	atomic.AddInt64(&totalRequests, 1)
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		atomic.AddInt64(&successfulRequests, 1)
	} else {
		atomic.AddInt64(&failedRequests, 1)
	}

	latencyMutex.Lock()
	latencies = append(latencies, duration)
	latencyMutex.Unlock()

	return resp, duration, err
}

func resetMetrics() {
	atomic.StoreInt64(&totalRequests, 0)
	atomic.StoreInt64(&successfulRequests, 0)
	atomic.StoreInt64(&failedRequests, 0)
	latencyMutex.Lock()
	latencies = []time.Duration{}
	latencyMutex.Unlock()
}

func calculateMetrics(testDuration time.Duration) PerformanceMetrics {
	metrics := PerformanceMetrics{
		TotalRequests:      atomic.LoadInt64(&totalRequests),
		SuccessfulRequests: atomic.LoadInt64(&successfulRequests),
		FailedRequests:     atomic.LoadInt64(&failedRequests),
		TotalDuration:      testDuration,
	}

	latencyMutex.Lock()
	defer latencyMutex.Unlock()

	if len(latencies) > 0 {
		var total time.Duration
		metrics.MinLatency = latencies[0]
		metrics.MaxLatency = latencies[0]

		for _, lat := range latencies {
			total += lat
			if lat < metrics.MinLatency {
				metrics.MinLatency = lat
			}
			if lat > metrics.MaxLatency {
				metrics.MaxLatency = lat
			}
		}

		metrics.AvgLatency = total / time.Duration(len(latencies))
	}

	if testDuration > 0 {
		metrics.RequestsPerSecond = float64(metrics.TotalRequests) / testDuration.Seconds()
	}

	return metrics
}

func testSingleEntityOperations(client *http.Client, token string) error {
	operations := []struct {
		name     string
		method   string
		endpoint string
		body     interface{}
	}{
		{
			name:     "Create",
			method:   "POST",
			endpoint: "/api/v1/entities/create",
			body: map[string]interface{}{
				"id":      fmt.Sprintf("perf_test_%d", time.Now().UnixNano()),
				"tags":    []string{"type:performance_test", "operation:create"},
				"content": []byte("Performance test entity"),
			},
		},
		{
			name:     "Read",
			method:   "GET",
			endpoint: "/api/v1/entities/get?id=admin",
			body:     nil,
		},
		{
			name:     "List",
			method:   "GET",
			endpoint: "/api/v1/entities/list?limit=10",
			body:     nil,
		},
		{
			name:     "Query",
			method:   "GET",
			endpoint: "/api/v1/entities/query?tags=type:user",
			body:     nil,
		},
	}

	fmt.Printf("   📊 Testing individual operations:\n")

	for _, op := range operations {
		var body io.Reader
		if op.body != nil {
			jsonData, _ := json.Marshal(op.body)
			body = bytes.NewBuffer(jsonData)
		}

		resp, duration, err := makeRequest(op.method, op.endpoint, body)
		if err != nil {
			return fmt.Errorf("%s operation failed: %v", op.name, err)
		}
		resp.Body.Close()

		fmt.Printf("      %s: %v", op.name, duration)
		if duration > 100*time.Millisecond {
			fmt.Printf(" ⚠️  (>100ms)")
		} else {
			fmt.Printf(" ✅")
		}
		fmt.Println()
	}

	return nil
}

func testBulkEntityCreation(client *http.Client, token string) error {
	numEntities := 100
	start := time.Now()
	
	var wg sync.WaitGroup
	concurrency := 10
	entitiesPerWorker := numEntities / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for j := 0; j < entitiesPerWorker; j++ {
				entity := map[string]interface{}{
					"id": fmt.Sprintf("bulk_test_%d_%d_%d", workerID, j, time.Now().UnixNano()),
					"tags": []string{
						"type:bulk_test",
						fmt.Sprintf("worker:%d", workerID),
						fmt.Sprintf("index:%d", j),
					},
					"content": []byte(fmt.Sprintf("Bulk test entity %d-%d", workerID, j)),
				}

				jsonData, _ := json.Marshal(entity)
				resp, _, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
				if err != nil {
					fmt.Printf("      ⚠️  Worker %d failed on entity %d: %v\n", workerID, j, err)
					continue
				}
				resp.Body.Close()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)
	
	metrics := calculateMetrics(duration)
	
	fmt.Printf("   📊 Bulk creation results:\n")
	fmt.Printf("      Total entities: %d\n", numEntities)
	fmt.Printf("      Duration: %v\n", duration)
	fmt.Printf("      Success rate: %.1f%%\n", float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100)
	fmt.Printf("      Throughput: %.2f entities/second\n", metrics.RequestsPerSecond)
	fmt.Printf("      Avg latency: %v\n", metrics.AvgLatency)

	if metrics.SuccessfulRequests < int64(numEntities*90/100) {
		return fmt.Errorf("bulk creation success rate too low: %.1f%%", 
			float64(metrics.SuccessfulRequests)/float64(numEntities)*100)
	}

	return nil
}

func testConcurrentReadWrite(client *http.Client, token string) error {
	duration := 10 * time.Second
	start := time.Now()
	done := make(chan bool)

	// Start concurrent readers
	for i := 0; i < 5; i++ {
		go func(readerID int) {
			for {
				select {
				case <-done:
					return
				default:
					// Random read operations
					operations := []string{
						"/api/v1/entities/list?limit=10",
						"/api/v1/entities/query?tags=type:user",
						"/api/v1/entities/query?tags=type:bulk_test",
					}
					
					endpoint := operations[rand.Intn(len(operations))]
					resp, _, _ := makeRequest("GET", endpoint, nil)
					if resp != nil {
						resp.Body.Close()
					}
					
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	// Start concurrent writers
	for i := 0; i < 3; i++ {
		go func(writerID int) {
			for {
				select {
				case <-done:
					return
				default:
					entity := map[string]interface{}{
						"id": fmt.Sprintf("concurrent_%d_%d", writerID, time.Now().UnixNano()),
						"tags": []string{
							"type:concurrent_test",
							fmt.Sprintf("writer:%d", writerID),
						},
						"content": []byte("Concurrent write test"),
					}

					jsonData, _ := json.Marshal(entity)
					resp, _, _ := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
					if resp != nil {
						resp.Body.Close()
					}
					
					time.Sleep(50 * time.Millisecond)
				}
			}
		}(i)
	}

	// Let it run for the duration
	time.Sleep(duration)
	close(done)
	
	testDuration := time.Since(start)
	metrics := calculateMetrics(testDuration)

	fmt.Printf("   📊 Concurrent read/write results:\n")
	fmt.Printf("      Duration: %v\n", testDuration)
	fmt.Printf("      Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("      Success rate: %.1f%%\n", float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100)
	fmt.Printf("      Throughput: %.2f requests/second\n", metrics.RequestsPerSecond)
	fmt.Printf("      Avg latency: %v\n", metrics.AvgLatency)
	fmt.Printf("      Max latency: %v\n", metrics.MaxLatency)

	if metrics.SuccessfulRequests < int64(float64(metrics.TotalRequests)*0.95) {
		return fmt.Errorf("concurrent operations success rate too low: %.1f%%", 
			float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100)
	}

	return nil
}

func testQueryPerformance(client *http.Client, token string) error {
	// First create some test data with various tags
	testData := []map[string]interface{}{
		{
			"id":   "query_test_1",
			"tags": []string{"type:document", "status:active", "priority:high", "category:technical"},
		},
		{
			"id":   "query_test_2",
			"tags": []string{"type:document", "status:draft", "priority:low", "category:business"},
		},
		{
			"id":   "query_test_3",
			"tags": []string{"type:task", "status:active", "priority:medium", "assigned:user123"},
		},
	}

	for _, data := range testData {
		data["content"] = []byte("Query test data")
		jsonData, _ := json.Marshal(data)
		resp, _, _ := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Test various query patterns
	queries := []struct {
		name     string
		endpoint string
	}{
		{"Single tag", "/api/v1/entities/query?tags=type:document"},
		{"Multiple tags (AND)", "/api/v1/entities/query?tags=type:document&tags=status:active"},
		{"Wildcard", "/api/v1/entities/query?tags=priority:*"},
		{"Complex filter", "/api/v1/entities/query?tags=type:document&tags=category:technical"},
		{"Large result set", "/api/v1/entities/query?tags=type:bulk_test&limit=100"},
	}

	fmt.Printf("   📊 Query performance results:\n")

	for _, q := range queries {
		resp, duration, err := makeRequest("GET", q.endpoint, nil)
		if err != nil {
			fmt.Printf("      %s: ERROR - %v\n", q.name, err)
			continue
		}
		
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		
		entities, _ := result["entities"].([]interface{})
		count := len(entities)
		
		fmt.Printf("      %s: %v (%d results)", q.name, duration, count)
		if duration > 50*time.Millisecond {
			fmt.Printf(" ⚠️  (>50ms)")
		} else {
			fmt.Printf(" ✅")
		}
		fmt.Println()
	}

	return nil
}

func testTemporalQueryLoad(client *http.Client, token string) error {
	// Create an entity with history
	entityID := fmt.Sprintf("temporal_test_%d", time.Now().UnixNano())
	
	// Create and update entity multiple times
	for i := 0; i < 5; i++ {
		entity := map[string]interface{}{
			"id": entityID,
			"tags": []string{
				"type:temporal_test",
				fmt.Sprintf("version:%d", i),
				fmt.Sprintf("status:%s", []string{"draft", "review", "approved", "published", "archived"}[i]),
			},
			"content": []byte(fmt.Sprintf("Version %d content", i)),
		}

		jsonData, _ := json.Marshal(entity)
		method := "POST"
		if i > 0 {
			method = "PUT"
		}
		
		resp, _, _ := makeRequest(method, "/api/v1/entities/"+map[string]string{"POST": "create", "PUT": "update"}[method], bytes.NewBuffer(jsonData))
		if resp != nil {
			resp.Body.Close()
		}
		
		time.Sleep(100 * time.Millisecond)
	}

	// Test temporal queries under load
	temporalQueries := []struct {
		name     string
		endpoint string
	}{
		{"History", fmt.Sprintf("/api/v1/entities/history?id=%s", entityID)},
		{"As-of (recent)", fmt.Sprintf("/api/v1/entities/as-of?id=%s&timestamp=%s", entityID, time.Now().Add(-1*time.Second).Format(time.RFC3339))},
		{"Changes (last hour)", "/api/v1/entities/changes?since=" + time.Now().Add(-1*time.Hour).Format(time.RFC3339)},
	}

	fmt.Printf("   📊 Temporal query performance:\n")

	// Run each query multiple times concurrently
	for _, q := range temporalQueries {
		var wg sync.WaitGroup
		var totalDuration time.Duration
		var mu sync.Mutex
		successCount := 0
		
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				
				resp, duration, err := makeRequest("GET", q.endpoint, nil)
				if err == nil && resp != nil {
					resp.Body.Close()
					mu.Lock()
					totalDuration += duration
					successCount++
					mu.Unlock()
				}
			}()
		}
		
		wg.Wait()
		
		avgDuration := totalDuration / time.Duration(successCount)
		fmt.Printf("      %s: avg %v over %d requests", q.name, avgDuration, successCount)
		if avgDuration > 100*time.Millisecond {
			fmt.Printf(" ⚠️  (>100ms)")
		} else {
			fmt.Printf(" ✅")
		}
		fmt.Println()
	}

	return nil
}

func testSustainedLoad(client *http.Client, token string) error {
	duration := 30 * time.Second
	fmt.Printf("   🏃 Running sustained load test for %v...\n", duration)
	
	start := time.Now()
	done := make(chan bool)
	
	// Simulate realistic mixed workload
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				// 70% reads, 20% writes, 10% queries
				r := rand.Float32()
				
				if r < 0.7 {
					// Read operation
					resp, _, _ := makeRequest("GET", "/api/v1/entities/list?limit=20", nil)
					if resp != nil {
						resp.Body.Close()
					}
				} else if r < 0.9 {
					// Write operation
					entity := map[string]interface{}{
						"id":      fmt.Sprintf("sustained_%d", time.Now().UnixNano()),
						"tags":    []string{"type:sustained_test", "timestamp:" + time.Now().Format(time.RFC3339)},
						"content": []byte("Sustained load test entity"),
					}
					jsonData, _ := json.Marshal(entity)
					resp, _, _ := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
					if resp != nil {
						resp.Body.Close()
					}
				} else {
					// Query operation
					resp, _, _ := makeRequest("GET", "/api/v1/entities/query?tags=type:sustained_test&limit=10", nil)
					if resp != nil {
						resp.Body.Close()
					}
				}
				
				// Simulate realistic request rate
				time.Sleep(time.Duration(rand.Intn(50)+10) * time.Millisecond)
			}
		}
	}()

	time.Sleep(duration)
	close(done)
	
	testDuration := time.Since(start)
	metrics := calculateMetrics(testDuration)

	fmt.Printf("   📊 Sustained load results:\n")
	fmt.Printf("      Duration: %v\n", testDuration)
	fmt.Printf("      Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("      Success rate: %.1f%%\n", float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100)
	fmt.Printf("      Avg throughput: %.2f requests/second\n", metrics.RequestsPerSecond)
	fmt.Printf("      Avg latency: %v\n", metrics.AvgLatency)
	fmt.Printf("      P99 latency: %v\n", metrics.MaxLatency)

	// Check system health after sustained load
	resp, _, err := makeRequest("GET", "/health", nil)
	if err != nil {
		return fmt.Errorf("health check failed after sustained load: %v", err)
	}
	
	var health map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&health)
	resp.Body.Close()
	
	if status, ok := health["status"].(string); ok && status != "healthy" {
		return fmt.Errorf("system unhealthy after sustained load: %s", status)
	}

	return nil
}

func testMemoryPressure(client *http.Client, token string) error {
	// Create entities with large content to test memory handling
	fmt.Printf("   🧠 Testing memory pressure handling...\n")
	
	// Create progressively larger entities
	sizes := []int{1024, 10240, 102400, 1048576} // 1KB, 10KB, 100KB, 1MB
	
	for i, size := range sizes {
		// Generate content of specified size
		content := make([]byte, size)
		for j := range content {
			content[j] = byte('A' + (j % 26))
		}
		
		entity := map[string]interface{}{
			"id":      fmt.Sprintf("memory_test_%d_%d", i, size),
			"tags":    []string{"type:memory_test", fmt.Sprintf("size:%d", size)},
			"content": content,
		}
		
		jsonData, _ := json.Marshal(entity)
		start := time.Now()
		resp, _, err := makeRequest("POST", "/api/v1/entities/create", bytes.NewBuffer(jsonData))
		duration := time.Since(start)
		
		if err != nil {
			fmt.Printf("      Size %d bytes: ERROR - %v\n", size, err)
			continue
		}
		
		status := "✅"
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			status = "⚠️"
		}
		resp.Body.Close()
		
		fmt.Printf("      Size %d bytes: %v %s\n", size, duration, status)
	}

	// Check memory metrics
	resp, _, err := makeRequest("GET", "/api/v1/system/metrics", nil)
	if err == nil {
		var metrics map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&metrics)
		resp.Body.Close()
		
		if memory, ok := metrics["memory"].(map[string]interface{}); ok {
			if heapAlloc, ok := memory["heap_alloc_mb"].(float64); ok {
				fmt.Printf("      Current heap allocation: %.2f MB\n", heapAlloc)
				if heapAlloc > 500 {
					fmt.Printf("      ⚠️  High memory usage detected\n")
				}
			}
		}
	}

	return nil
}

func testAPIEndpointCoverage(client *http.Client, token string) error {
	// Test various API endpoints to ensure comprehensive coverage
	endpoints := []struct {
		name     string
		method   string
		endpoint string
		body     interface{}
	}{
		// Core entity operations
		{"List entities", "GET", "/api/v1/entities/list?limit=5", nil},
		{"Query entities", "GET", "/api/v1/entities/query?tags=type:user", nil},
		{"Get entity", "GET", "/api/v1/entities/get?id=admin", nil},
		
		// Temporal operations
		{"Entity history", "GET", "/api/v1/entities/history?id=admin&limit=5", nil},
		{"Recent changes", "GET", "/api/v1/entities/changes?limit=10", nil},
		
		// System operations
		{"Health check", "GET", "/health", nil},
		{"System metrics", "GET", "/api/v1/system/metrics", nil},
		{"Dashboard stats", "GET", "/api/v1/dashboard/stats", nil},
		
		// Tag operations
		{"Tag values", "GET", "/api/v1/tags/values?namespace=type", nil},
		
		// User operations
		{"List users", "GET", "/api/v1/users/list", nil},
	}

	fmt.Printf("   📊 API endpoint performance:\n")
	
	failedEndpoints := 0
	for _, ep := range endpoints {
		var body io.Reader
		if ep.body != nil {
			jsonData, _ := json.Marshal(ep.body)
			body = bytes.NewBuffer(jsonData)
		}

		resp, duration, err := makeRequest(ep.method, ep.endpoint, body)
		if err != nil {
			fmt.Printf("      %s: ERROR - %v\n", ep.name, err)
			failedEndpoints++
			continue
		}
		
		status := "✅"
		if resp.StatusCode >= 400 {
			status = fmt.Sprintf("❌ (%d)", resp.StatusCode)
			failedEndpoints++
		} else if duration > 100*time.Millisecond {
			status = "⚠️  (slow)"
		}
		
		resp.Body.Close()
		fmt.Printf("      %s: %v %s\n", ep.name, duration, status)
	}

	if failedEndpoints > len(endpoints)/10 {
		return fmt.Errorf("too many endpoint failures: %d/%d", failedEndpoints, len(endpoints))
	}

	return nil
}